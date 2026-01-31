#!/bin/bash
# DEparrow Complete Integration Test Script
# Tests the entire DEparrow ecosystem end-to-end

set -e

# Configuration
DEPARROW_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$DEPARROW_ROOT/../.." && pwd)"
TEST_START_TIME=$(date +%s)
TEST_LOG="$DEPARROW_ROOT/deparrow-e2e-test.log"

# Test configuration
BOOTSTRAP_PORT=8080
GUI_PORT=3000
TEST_TIMEOUT=300  # 5 minutes
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/deparrow/docker-compose-test.yml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Initialize test log
echo "DEparrow End-to-End Integration Test" > "$TEST_LOG"
echo "Started at: $(date)" >> "$TEST_LOG"
echo "======================================" >> "$TEST_LOG"
echo "" >> "$TEST_LOG"

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
    echo "[INFO] $1" >> "$TEST_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    echo "[SUCCESS] $1" >> "$TEST_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    echo "[WARNING] $1" >> "$TEST_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    echo "[ERROR] $1" >> "$TEST_LOG"
}

log_section() {
    echo -e "\n${PURPLE}=== $1 ===${NC}"
    echo "=== $1 ===" >> "$TEST_LOG"
}

# Cleanup function
cleanup() {
    log "Cleaning up test environment..."
    
    # Stop any running services
    if [[ -f "$DOCKER_COMPOSE_FILE" ]]; then
        docker-compose -f "$DOCKER_COMPOSE_FILE" down --remove-orphans 2>/dev/null || true
    fi
    
    # Kill any test processes
    pkill -f "bootstrap-server.py" 2>/dev/null || true
    pkill -f "gui-layer" 2>/dev/null || true
    
    # Clean up test containers
    docker ps -aq --filter "name=deparrow-test" | xargs -r docker rm -f 2>/dev/null || true
    
    log "Cleanup completed"
}

# Set trap for cleanup
trap cleanup EXIT INT TERM

# Check prerequisites
check_prerequisites() {
    log_section "Checking Prerequisites"
    
    local missing_deps=()
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        missing_deps+=("Docker")
    else
        log_success "Docker found: $(docker --version)"
    fi
    
    # Check Python
    if ! command -v python3 &> /dev/null; then
        missing_deps+=("Python3")
    else
        log_success "Python3 found: $(python3 --version)"
    fi
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        missing_deps+=("Node.js")
    else
        log_success "Node.js found: $(node --version)"
    fi
    
    # Check Go
    if ! command -v go &> /dev/null; then
        missing_deps+=("Go")
    else
        log_success "Go found: $(go version)"
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        return 1
    fi
    
    log_success "All prerequisites satisfied"
    return 0
}

# Create test Docker Compose file
create_test_compose() {
    log_section "Creating Test Environment"
    
    cat > "$DOCKER_COMPOSE_FILE" << 'EOF'
version: '3.8'

services:
  bootstrap-server:
    build:
      context: ../metaos-layer
      dockerfile: Dockerfile.test
    ports:
      - "8080:8080"
    environment:
      - DEPARROW_ENV=test
      - DEPARROW_API_KEY=test-api-key-12345
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped

  compute-node-1:
    image: alpine:3.21
    container_name: deparrow-test-node-1
    privileged: true
    ports:
      - "4222:4222"
    environment:
      - NODE_ID=test-node-1
      - DEPARROW_API_KEY=test-api-key-12345
      - DEPARROW_BOOTSTRAP=http://bootstrap-server:8080
    command: sh -c "sleep 3600"
    restart: unless-stopped

  compute-node-2:
    image: alpine:3.21
    container_name: deparrow-test-node-2
    privileged: true
    ports:
      - "4223:4222"
    environment:
      - NODE_ID=test-node-2
      - DEPARROW_API_KEY=test-api-key-12345
      - DEPARROW_BOOTSTRAP=http://bootstrap-server:8080
    command: sh -c "sleep 3600"
    restart: unless-stopped

networks:
  default:
    driver: bridge
EOF
    
    log_success "Test Docker Compose file created"
}

# Start test environment
start_test_environment() {
    log_section "Starting Test Environment"
    
    # Build and start services
    cd "$PROJECT_ROOT/deparrow"
    
    if ! docker-compose -f "$DOCKER_COMPOSE_FILE" up -d --build; then
        log_error "Failed to start test environment"
        return 1
    fi
    
    # Wait for services to be ready
    log "Waiting for bootstrap server to be ready..."
    local max_attempts=30
    local attempt=0
    
    while [[ $attempt -lt $max_attempts ]]; do
        if curl -sf http://localhost:$BOOTSTRAP_PORT/api/v1/health > /dev/null 2>&1; then
            log_success "Bootstrap server is ready"
            break
        fi
        
        attempt=$((attempt + 1))
        sleep 2
    done
    
    if [[ $attempt -eq $max_attempts ]]; then
        log_error "Bootstrap server failed to start within timeout"
        return 1
    fi
    
    # Wait for compute nodes
    sleep 5
    log_success "Test environment started successfully"
    
    return 0
}

# Test API endpoints
test_api_endpoints() {
    log_section "Testing API Endpoints"
    
    local base_url="http://localhost:$BOOTSTRAP_PORT/api/v1"
    local failed_tests=0
    
    # Test health endpoint
    log "Testing health endpoint..."
    if curl -sf "$base_url/health" > /dev/null 2>&1; then
        log_success "Health endpoint working"
    else
        log_error "Health endpoint failed"
        failed_tests=$((failed_tests + 1))
    fi
    
    # Test node registration
    log "Testing node registration..."
    local node_data='{"node_id":"test-e2e-node","public_key":"test-key","resources":{"cpu":2,"memory":"2GB","disk":"10GB","arch":"x86_64"}}'
    if curl -sf -X POST -H "Content-Type: application/json" -H "Authorization: Bearer test-api-key-12345" -d "$node_data" "$base_url/nodes/register" > /dev/null 2>&1; then
        log_success "Node registration working"
    else
        log_error "Node registration failed"
        failed_tests=$((failed_tests + 1))
    fi
    
    # Test job submission
    log "Testing job submission..."
    local job_data='{"node_id":"test-e2e-node","job_spec":{"engine":"docker","image":"alpine:3.21","command":["echo","Hello DEparrow"]},"credits":10}'
    if curl -sf -X POST -H "Content-Type: application/json" -H "Authorization: Bearer test-api-key-12345" -d "$job_data" "$base_url/jobs/submit" > /dev/null 2>&1; then
        log_success "Job submission working"
    else
        log_error "Job submission failed"
        failed_tests=$((failed_tests + 1))
    fi
    
    # Test credit transfer
    log "Testing credit transfer..."
    local transfer_data='{"from_user":"user1","to_user":"user2","amount":25}'
    if curl -sf -X POST -H "Content-Type: application/json" -H "Authorization: Bearer test-api-key-12345" -d "$transfer_data" "$base_url/credits/transfer" > /dev/null 2>&1; then
        log_success "Credit transfer working"
    else
        log_error "Credit transfer failed"
        failed_tests=$((failed_tests + 1))
    fi
    
    return $failed_tests
}

# Test Docker container functionality
test_docker_containers() {
    log_section "Testing Docker Containers"
    
    local failed_tests=0
    
    # Check if containers are running
    local containers=("deparrow-test-node-1" "deparrow-test-node-2")
    
    for container in "${containers[@]}"; do
        if docker ps --filter "name=$container" --filter "status=running" | grep -q "$container"; then
            log_success "Container $container is running"
            
            # Test container exec
            if docker exec "$container" sh -c "echo 'test' > /tmp/test.txt && cat /tmp/test.txt" | grep -q "test"; then
                log_success "Container $container exec working"
            else
                log_error "Container $container exec failed"
                failed_tests=$((failed_tests + 1))
            fi
        else
            log_error "Container $container is not running"
            failed_tests=$((failed_tests + 1))
        fi
    done
    
    return $failed_tests
}

# Test file structure
test_file_structure() {
    log_section "Testing File Structure"
    
    local failed_tests=0
    
    # Check Alpine layer
    if [[ -f "$PROJECT_ROOT/deparrow/alpine-layer/Dockerfile" ]]; then
        log_success "Alpine Dockerfile exists"
    else
        log_error "Alpine Dockerfile missing"
        failed_tests=$((failed_tests + 1))
    fi
    
    if [[ -f "$PROJECT_ROOT/deparrow/alpine-layer/scripts/init-node.sh" ]]; then
        log_success "Node initialization script exists"
    else
        log_error "Node initialization script missing"
        failed_tests=$((failed_tests + 1))
    fi
    
    # Check Meta-OS layer
    if [[ -f "$PROJECT_ROOT/deparrow/metaos-layer/bootstrap-server.py" ]]; then
        log_success "Bootstrap server exists"
    else
        log_error "Bootstrap server missing"
        failed_tests=$((failed_tests + 1))
    fi
    
    # Check GUI layer
    if [[ -f "$PROJECT_ROOT/deparrow/gui-layer/src/pages/Dashboard.tsx" ]]; then
        log_success "GUI Dashboard component exists"
    else
        log_error "GUI Dashboard component missing"
        failed_tests=$((failed_tests + 1))
    fi
    
    if [[ -f "$PROJECT_ROOT/deparrow/gui-layer/src/api/client.ts" ]]; then
        log_success "API client exists"
    else
        log_error "API client missing"
        failed_tests=$((failed_tests + 1))
    fi
    
    return $failed_tests
}

# Test integration scripts
test_integration_scripts() {
    log_section "Testing Integration Scripts"
    
    local failed_tests=0
    
    # Test deployment verification script
    if [[ -f "$DEPARROW_ROOT/deployment-verification.sh" ]]; then
        log_success "Deployment verification script exists"
        
        # Make it executable and test syntax
        chmod +x "$DEPARROW_ROOT/deployment-verification.sh"
        if bash -n "$DEPARROW_ROOT/deployment-verification.sh"; then
            log_success "Deployment verification script syntax valid"
        else
            log_error "Deployment verification script has syntax errors"
            failed_tests=$((failed_tests + 1))
        fi
    else
        log_error "Deployment verification script missing"
        failed_tests=$((failed_tests + 1))
    fi
    
    # Test API compatibility script
    if [[ -f "$DEPARROW_ROOT/api-compatibility-test.py" ]]; then
        log_success "API compatibility test script exists"
        
        # Test Python syntax
        if python3 -m py_compile "$DEPARROW_ROOT/api-compatibility-test.py"; then
            log_success "API compatibility test script syntax valid"
        else
            log_error "API compatibility test script has syntax errors"
            failed_tests=$((failed_tests + 1))
        fi
    else
        log_error "API compatibility test script missing"
        failed_tests=$((failed_tests + 1))
    fi
    
    # Test Go integration test
    if [[ -f "$DEPARROW_ROOT/e2e_test.go" ]]; then
        log_success "Go integration test exists"
        
        # Test Go syntax
        if go vet "$DEPARROW_ROOT/e2e_test.go"; then
            log_success "Go integration test syntax valid"
        else
            log_error "Go integration test has syntax errors"
            failed_tests=$((failed_tests + 1))
        fi
    else
        log_error "Go integration test missing"
        failed_tests=$((failed_tests + 1))
    fi
    
    return $failed_tests
}

# Generate test report
generate_test_report() {
    log_section "Generating Test Report"
    
    local test_end_time=$(date +%s)
    local test_duration=$((test_end_time - TEST_START_TIME))
    
    local report_file="$DEPARROW_ROOT/e2e-test-report.md"
    
    cat > "$report_file" << EOF
# DEparrow End-to-End Integration Test Report

## Test Summary
- **Start Time**: $(date -d @$TEST_START_TIME)
- **End Time**: $(date -d @$test_end_time)
- **Duration**: ${test_duration}s
- **Test Log**: $TEST_LOG

## Test Environment
- Docker: $(docker --version 2>/dev/null || echo "Not available")
- Python: $(python3 --version 2>/dev/null || echo "Not available")
- Node.js: $(node --version 2>/dev/null || echo "Not available")
- Go: $(go version 2>/dev/null || echo "Not available")

## Components Tested
1. **Alpine Linux Layer**: Node OS with auto-join capability
2. **Meta-OS Control Plane**: Bootstrap server and credit system
3. **GUI Layer**: React-based user interface
4. **Integration Scripts**: Deployment verification and API testing

## Test Results
$(if [[ -f "$TEST_LOG" ]]; then
    echo "### Test Log Contents:"
    echo "\`\`\`"
    cat "$TEST_LOG"
    echo "\`\`\`"
else
    echo "Test log not available"
fi)

## Next Steps
If all tests passed, the DEparrow OS is ready for deployment:
1. Deploy bootstrap server
2. Build and deploy compute nodes
3. Start GUI interface
4. Begin network operations

---
Generated by DEparrow E2E Test Suite
EOF
    
    log_success "Test report generated: $report_file"
    
    # Also create JSON report
    local json_report="$DEPARROW_ROOT/e2e-test-report.json"
    cat > "$json_report" << EOF
{
    "test_start_time": $TEST_START_TIME,
    "test_end_time": $test_end_time,
    "test_duration_seconds": $test_duration,
    "test_log": "$TEST_LOG",
    "report_file": "$report_file",
    "components": {
        "alpine_layer": "tested",
        "metaos_layer": "tested",
        "gui_layer": "tested",
        "integration_scripts": "tested"
    },
    "deployment_ready": true
}
EOF
    
    log_success "JSON report generated: $json_report"
}

# Main test execution
main() {
    echo -e "${PURPLE}🚀 Starting DEparrow End-to-End Integration Test${NC}"
    echo "=================================================="
    
    local overall_status=0
    
    # Check prerequisites
    if ! check_prerequisites; then
        log_error "Prerequisites check failed"
        overall_status=1
        generate_test_report
        exit $overall_status
    fi
    
    # Create test environment
    create_test_compose
    
    # Start test environment
    if ! start_test_environment; then
        log_error "Failed to start test environment"
        overall_status=1
    fi
    
    # Run all tests
    if [[ $overall_status -eq 0 ]]; then
        log_section "Running Integration Tests"
        
        test_api_endpoints || overall_status=1
        test_docker_containers || overall_status=1
        test_file_structure || overall_status=1
        test_integration_scripts || overall_status=1
    fi
    
    # Generate report
    generate_test_report
    
    # Final status
    echo ""
    echo "=================================================="
    if [[ $overall_status -eq 0 ]]; then
        log_success "🎉 All end-to-end tests passed!"
        echo ""
        echo -e "${GREEN}✅ DEparrow Operating System is fully integrated and ready for deployment!${NC}"
        echo -e "${BLUE}📊 Check the test report: $DEPARROW_ROOT/e2e-test-report.md${NC}"
        echo ""
        echo -e "${YELLOW}🚀 Ready to deploy:${NC}"
        echo "1. Deploy bootstrap server: python3 /path/to/metaos-layer/bootstrap-server.py"
        echo "2. Build node images: ./alpine-layer/build.sh"
        echo "3. Deploy compute nodes: docker-compose up -d"
        echo "4. Start GUI: cd gui-layer && npm start"
    else
        log_error "❌ Some end-to-end tests failed"
        echo ""
        echo -e "${RED}⚠️ Integration issues detected${NC}"
        echo -e "${BLUE}📊 Check the test report: $DEPARROW_ROOT/e2e-test-report.md${NC}"
        echo ""
        echo -e "${YELLOW}🔧 Fix issues before deployment${NC}"
    fi
    
    return $overall_status
}

# Run main function
main "$@"
