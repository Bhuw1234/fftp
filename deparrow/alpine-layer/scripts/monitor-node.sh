#!/bin/bash
# DEparrow Node Monitoring Script
# Monitors node health, resources, and network connectivity

set -e

LOG_FILE="/var/log/deparrow/monitor.log"
METRICS_FILE="/var/lib/deparrow/metrics.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

check_bacalhau_health() {
    if ! pgrep -f "bacalhau serve" > /dev/null; then
        log_message "ERROR: Bacalhau process not running"
        return 1
    fi
    
    # Check if Bacalhau is responding to API calls
    if ! curl -sf http://localhost:1234/api/v1/health > /dev/null 2>&1; then
        log_message "ERROR: Bacalhau API not responding"
        return 1
    fi
    
    return 0
}

check_docker_health() {
    if ! docker ps > /dev/null 2>&1; then
        log_message "ERROR: Docker daemon not responding"
        return 1
    fi
    
    # Check running containers
    local running_containers=$(docker ps --filter "status=running" --format "{{.Names}}" | wc -l)
    log_message "INFO: Docker running with $running_containers containers"
    
    return 0
}

check_resource_usage() {
    # CPU usage
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
    log_message "INFO: CPU usage: ${cpu_usage}%"
    
    # Memory usage
    local mem_info=$(free | grep Mem)
    local mem_total=$(echo $mem_info | awk '{print $2}')
    local mem_used=$(echo $mem_info | awk '{print $3}')
    local mem_percent=$(( mem_used * 100 / mem_total ))
    log_message "INFO: Memory usage: ${mem_percent}%"
    
    # Disk usage
    local disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    log_message "INFO: Disk usage: ${disk_usage}%"
    
    # Check thresholds
    if [ "$cpu_usage" -gt 90 ]; then
        log_message "WARNING: High CPU usage: ${cpu_usage}%"
    fi
    
    if [ "$mem_percent" -gt 90 ]; then
        log_message "WARNING: High memory usage: ${mem_percent}%"
    fi
    
    if [ "$disk_usage" -gt 90 ]; then
        log_message "WARNING: High disk usage: ${disk_usage}%"
        return 1
    fi
    
    return 0
}

check_network_connectivity() {
    # Check bootstrap connectivity
    if [ -n "$DEPARROW_BOOTSTRAP" ]; then
        if curl -sf --max-time 10 "$DEPARROW_BOOTSTRAP/api/v1/health" > /dev/null 2>&1; then
            log_message "INFO: Bootstrap connectivity OK"
        else
            log_message "WARNING: Cannot reach DEparrow bootstrap"
        fi
    fi
    
    # Check NATS connectivity
    if [ -n "$DEPARROW_ORCHESTRATOR_HOST" ]; then
        if nc -z "$DEPARROW_ORCHESTRATOR_HOST" 4222 2>/dev/null; then
            log_message "INFO: NATS connectivity OK"
        else
            log_message "WARNING: Cannot reach NATS server at $DEPARROW_ORCHESTRATOR_HOST:4222"
        fi
    fi
    
    return 0
}

collect_metrics() {
    local timestamp=$(date -Iseconds)
    
    # System metrics
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
    local mem_info=$(free | grep Mem)
    local mem_total=$(echo $mem_info | awk '{print $2}')
    local mem_used=$(echo $mem_info | awk '{print $3}')
    local disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    # Docker metrics
    local running_containers=$(docker ps --filter "status=running" --format "{{.Names}}" | wc -l)
    local total_containers=$(docker ps --all --format "{{.Names}}" | wc -l)
    
    # Network metrics
    local network_rx=$(cat /proc/net/dev | grep eth0 | awk '{print $2}' | head -1)
    local network_tx=$(cat /proc/net/dev | grep eth0 | awk '{print $10}' | head -1)
    
    # Create metrics JSON
    cat > "$METRICS_FILE" << EOF
{
  "timestamp": "$timestamp",
  "node_id": "${NODE_ID:-unknown}",
  "system": {
    "cpu_usage": ${cpu_usage:-0},
    "memory_total": ${mem_total:-0},
    "memory_used": ${mem_used:-0},
    "disk_usage_percent": ${disk_usage:-0}
  },
  "docker": {
    "running_containers": ${running_containers:-0},
    "total_containers": ${total_containers:-0}
  },
  "network": {
    "rx_bytes": ${network_rx:-0},
    "tx_bytes": ${network_tx:-0}
  },
  "bacalhau": {
    "status": "$(pgrep -f 'bacalhau serve' > /dev/null && echo 'running' || echo 'stopped')",
    "jobs_executed": 0,
    "credits_earned": 0.0
  }
}
EOF
}

send_heartbeat() {
    if [ -n "$DEPARROW_API_KEY" ] && [ -n "$DEPARROW_BOOTSTRAP" ]; then
        local node_id="${NODE_ID:-unknown}"
        
        curl -X POST \
            -H "Authorization: Bearer $DEPARROW_API_KEY" \
            -H "Content-Type: application/json" \
            -d "{\"node_id\":\"$node_id\",\"status\":\"online\",\"metrics\":$(cat $METRICS_FILE)}" \
            "$DEPARROW_BOOTSTRAP/api/v1/nodes/$node_id/heartbeat" \
            > /dev/null 2>&1 || log_message "WARNING: Failed to send heartbeat"
    fi
}

cleanup_old_logs() {
    # Rotate logs if they get too large
    if [ -f "$LOG_FILE" ]; then
        local log_size=$(stat -c%s "$LOG_FILE" 2>/dev/null || echo 0)
        if [ "$log_size" -gt 104857600 ]; then  # 100MB
            mv "$LOG_FILE" "${LOG_FILE}.old"
            touch "$LOG_FILE"
            log_message "INFO: Log file rotated"
        fi
    fi
}

main() {
    log_message "INFO: Starting node monitoring"
    
    # Check system health
    local health_ok=true
    
    if ! check_bacalhau_health; then
        health_ok=false
    fi
    
    if ! check_docker_health; then
        health_ok=false
    fi
    
    if ! check_resource_usage; then
        health_ok=false
    fi
    
    check_network_connectivity
    
    # Collect and send metrics
    collect_metrics
    send_heartbeat
    
    # Cleanup
    cleanup_old_logs
    
    if [ "$health_ok" = true ]; then
        log_message "INFO: Node health check passed"
        return 0
    else
        log_message "ERROR: Node health check failed"
        return 1
    fi
}

# Run main function if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi
