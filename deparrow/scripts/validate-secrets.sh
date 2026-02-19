#!/bin/bash
# =============================================================================
# DEparrow Environment Validation Script
# =============================================================================
#
# Validates that required environment variables are set before starting
# the production stack. Run this before docker-compose up.
#
# Usage:
#   ./scripts/validate-secrets.sh
#   ./scripts/validate-secrets.sh --strict  # Also check for default values
#
# Exit codes:
#   0 - All validations passed
#   1 - Missing required variables
#   2 - Insecure default values detected (with --strict)
#
# =============================================================================

set -uo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track validation status
ERRORS=0
WARNINGS=0

# Parse arguments
STRICT_MODE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --strict|-s)
            STRICT_MODE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--strict|-s]"
            echo ""
            echo "Validates DEparrow environment configuration."
            echo ""
            echo "Options:"
            echo "  --strict, -s    Also check for insecure default values"
            echo "  --help, -h      Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "=========================================="
echo "DEparrow Environment Validation"
echo "=========================================="
echo ""

# Check if .env file exists
if [[ ! -f ".env" ]]; then
    echo -e "${RED}ERROR: .env file not found!${NC}"
    echo "Please copy .env.example to .env and configure your secrets:"
    echo "  cp .env.example .env"
    echo "  # Edit .env with your secure values"
    exit 1
fi

echo -e "${GREEN}✓${NC} .env file found"

# Source the .env file to load variables
set -a
source .env
set +a

echo ""

# -----------------------------------------------------------------------------
# Required secrets - MUST be set
# -----------------------------------------------------------------------------
echo "Checking required secrets..."
echo ""

REQUIRED_VARS=(
    "DEPARROW_SECRET_KEY:JWT signing key for Meta-OS"
    "POSTGRES_PASSWORD:PostgreSQL database password"
    "GRAFANA_ADMIN_PASSWORD:Grafana admin dashboard password"
)

for var_spec in "${REQUIRED_VARS[@]}"; do
    IFS=':' read -r var_name var_desc <<< "$var_spec"
    
    # Check if variable is set
    if [[ -z "${!var_name:-}" ]]; then
        echo -e "${RED}✗${NC} $var_name is NOT SET"
        echo "  Description: $var_desc"
        echo "  Fix: Set $var_name in your .env file"
        echo ""
        ((ERRORS++))
    else
        echo -e "${GREEN}✓${NC} $var_name is set"
        
        # Strict mode: check for insecure defaults
        if [[ "$STRICT_MODE" == "true" ]]; then
            value="${!var_name}"
            
            # Check for common insecure patterns (case-insensitive)
            insecure_patterns=(
                "change-me"
                "change_me"
                "changeme"
                "your-"
                "password"
                "secret"
                "admin"
                "default"
                "12345"
                "test"
                "secure-random"
                "secure_password"
            )
            
            for pattern in "${insecure_patterns[@]}"; do
                if [[ "${value,,}" == *"${pattern,,}"* ]]; then
                    echo -e "  ${YELLOW}⚠${NC} Contains potentially insecure value: '$pattern'"
                    ((WARNINGS++))
                fi
            done
            
            # Check minimum length
            if [[ ${#value} -lt 16 ]]; then
                echo -e "  ${YELLOW}⚠${NC} Value is shorter than 16 characters (current: ${#value})"
                ((WARNINGS++))
            fi
        fi
    fi
done

echo ""

# -----------------------------------------------------------------------------
# Optional checks
# -----------------------------------------------------------------------------
echo "Checking optional configuration..."
echo ""

# Check Redis password (optional but recommended)
if [[ -z "${REDIS_PASSWORD:-}" ]]; then
    echo -e "${YELLOW}⚠${NC} REDIS_PASSWORD is not set (Redis will run without authentication)"
    echo "  Recommendation: Set REDIS_PASSWORD for production deployments"
    ((WARNINGS++))
else
    echo -e "${GREEN}✓${NC} REDIS_PASSWORD is set"
fi

# Check if we're using latest images (potentially unstable)
if [[ "${BACALHAU_VERSION:-latest}" == "latest" ]]; then
    echo -e "${YELLOW}⚠${NC} Using 'latest' Bacalhau image (may cause unexpected updates)"
    echo "  Recommendation: Pin BACALHAU_VERSION to a specific version"
fi

if [[ "${PROMETHEUS_VERSION:-latest}" == "latest" ]]; then
    echo -e "${YELLOW}⚠${NC} Using 'latest' Prometheus image"
fi

if [[ "${GRAFANA_VERSION:-latest}" == "latest" ]]; then
    echo -e "${YELLOW}⚠${NC} Using 'latest' Grafana image"
fi

echo ""

# -----------------------------------------------------------------------------
# File permissions check
# -----------------------------------------------------------------------------
echo "Checking file permissions..."
echo ""

if [[ -f ".env" ]]; then
    env_perms=$(stat -c "%a" .env 2>/dev/null || stat -f "%Lp" .env 2>/dev/null || echo "unknown")
    if [[ "$env_perms" != "600" && "$env_perms" != "400" ]]; then
        echo -e "${YELLOW}⚠${NC} .env file has permissive permissions: $env_perms"
        echo "  Recommendation: Run 'chmod 600 .env' to restrict access"
        ((WARNINGS++))
    else
        echo -e "${GREEN}✓${NC} .env file has secure permissions: $env_perms"
    fi
fi

echo ""

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
echo "=========================================="
echo "Validation Summary"
echo "=========================================="
echo ""

if [[ $ERRORS -gt 0 ]]; then
    echo -e "${RED}✗ Found $ERRORS error(s)${NC}"
    echo ""
    echo "Cannot start DEparrow with missing required secrets."
    echo "Please fix the errors above and run this script again."
    exit 1
fi

if [[ $WARNINGS -gt 0 ]]; then
    echo -e "${YELLOW}⚠ Found $WARNINGS warning(s)${NC}"
    echo ""
    echo "Warnings indicate potential security issues."
    echo "Review the warnings above before deploying to production."
fi

echo -e "${GREEN}✓ All required secrets are configured${NC}"
echo ""
echo "You can now start DEparrow:"
echo "  docker-compose -f docker-compose.prod.yml up -d"
echo ""

exit 0
