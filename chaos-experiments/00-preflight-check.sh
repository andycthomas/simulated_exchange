#!/bin/bash
################################################################################
# Chaos Engineering Preflight Check
#
# This script verifies that the system is ready for chaos experiments
################################################################################

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "╔═══════════════════════════════════════════════════════════════════════════╗"
echo "║                  CHAOS ENGINEERING PREFLIGHT CHECK                        ║"
echo "╚═══════════════════════════════════════════════════════════════════════════╝"
echo ""

CHECKS_PASSED=0
CHECKS_FAILED=0
CHECKS_WARNING=0

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((CHECKS_PASSED++))
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    ((CHECKS_FAILED++))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((CHECKS_WARNING++))
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. System Requirements"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    check_pass "Operating System: Linux"
else
    check_warn "Operating System: $OSTYPE (some features may not work)"
fi

# Check if running as root (for certain experiments)
if [[ $EUID -eq 0 ]]; then
    check_pass "Running as root (all experiments available)"
else
    check_warn "Not running as root (some experiments will require sudo)"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. Required Tools"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check Docker
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | tr -d ',')
    check_pass "Docker installed: $DOCKER_VERSION"
else
    check_fail "Docker not installed"
fi

# Check Docker Compose
if docker compose version &> /dev/null; then
    COMPOSE_VERSION=$(docker compose version | cut -d' ' -f4)
    check_pass "Docker Compose installed: $COMPOSE_VERSION"
elif command -v docker-compose &> /dev/null; then
    COMPOSE_VERSION=$(docker-compose --version | cut -d' ' -f4 | tr -d ',')
    check_pass "Docker Compose installed: $COMPOSE_VERSION"
else
    check_fail "Docker Compose not installed"
fi

# Check Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    check_pass "Go installed: $GO_VERSION"
else
    check_fail "Go not installed"
fi

# Check curl
if command -v curl &> /dev/null; then
    check_pass "curl installed"
else
    check_fail "curl not installed"
fi

# Check jq
if command -v jq &> /dev/null; then
    check_pass "jq installed"
else
    check_warn "jq not installed (recommended for JSON parsing)"
fi

# Check pprof (Go tool)
if go tool pprof -h &> /dev/null; then
    check_pass "pprof available"
else
    check_warn "pprof not available (profiling features limited)"
fi

# Check bpftrace (optional)
if command -v bpftrace &> /dev/null; then
    check_pass "bpftrace installed (eBPF tracing enabled)"
else
    check_warn "bpftrace not installed (eBPF tracing disabled)"
fi

# Check tc (traffic control)
if command -v tc &> /dev/null; then
    check_pass "tc (traffic control) available"
else
    check_warn "tc not available (network experiments limited)"
fi

# Check stress-ng (optional)
if command -v stress-ng &> /dev/null; then
    check_pass "stress-ng installed"
else
    check_warn "stress-ng not installed (CPU/memory stress tests limited)"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. Docker Services"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

cd "$PROJECT_ROOT"

# Check if docker-compose.yml exists
if [[ -f "docker-compose.yml" ]]; then
    check_pass "docker-compose.yml found"
else
    check_fail "docker-compose.yml not found"
fi

# Check if services are running
if docker compose ps 2>/dev/null | grep -q "Up\|running"; then
    check_pass "Docker services are running"

    # Check specific services
    SERVICES=("trading-api" "market-simulator" "postgres" "redis")
    for service in "${SERVICES[@]}"; do
        if docker compose ps "$service" 2>/dev/null | grep -q "Up\|running"; then
            check_pass "  └─ $service is running"
        else
            check_warn "  └─ $service not running"
        fi
    done
else
    check_warn "Docker services not running (start with: docker compose up -d)"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. API Endpoints"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check Trading API
if curl -s -f http://localhost:8080/health &> /dev/null; then
    check_pass "Trading API responding on :8080"
else
    check_warn "Trading API not responding on :8080"
fi

# Check Market Simulator
if curl -s -f http://localhost:8081/health &> /dev/null; then
    check_pass "Market Simulator responding on :8081"
else
    check_warn "Market Simulator not responding on :8081"
fi

# Check Order Flow Simulator
if curl -s -f http://localhost:8082/health &> /dev/null; then
    check_pass "Order Flow Simulator responding on :8082"
else
    check_warn "Order Flow Simulator not responding on :8082"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. Experiment Scripts"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

EXPERIMENT_SCRIPTS=(
    "01-database-sudden-death.sh"
    "02-network-partition.sh"
    "03-memory-oom-kill.sh"
    "09-network-latency.sh"
)

for script in "${EXPERIMENT_SCRIPTS[@]}"; do
    if [[ -f "$SCRIPT_DIR/$script" && -x "$SCRIPT_DIR/$script" ]]; then
        check_pass "$script found and executable"
    elif [[ -f "$SCRIPT_DIR/$script" ]]; then
        check_warn "$script found but not executable (run: chmod +x $script)"
    else
        check_fail "$script not found"
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. Directories"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check writable /tmp
if [[ -w /tmp ]]; then
    check_pass "/tmp is writable (experiment logs will be stored here)"
else
    check_fail "/tmp is not writable"
fi

# Create necessary directories
DIRS=("test-output" "flamegraphs" "chaos-results")
for dir in "${DIRS[@]}"; do
    if [[ -d "$PROJECT_ROOT/$dir" ]]; then
        check_pass "$dir/ directory exists"
    else
        mkdir -p "$PROJECT_ROOT/$dir" 2>/dev/null && check_pass "$dir/ directory created" || check_warn "Could not create $dir/ directory"
    fi
done

echo ""
echo "╔═══════════════════════════════════════════════════════════════════════════╗"
echo "║                           PREFLIGHT SUMMARY                               ║"
echo "╚═══════════════════════════════════════════════════════════════════════════╝"
echo ""
echo -e "Checks Passed:  ${GREEN}$CHECKS_PASSED${NC}"
echo -e "Checks Warning: ${YELLOW}$CHECKS_WARNING${NC}"
echo -e "Checks Failed:  ${RED}$CHECKS_FAILED${NC}"
echo ""

if [[ $CHECKS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}✅ System is ready for chaos experiments!${NC}"
    echo ""
    echo "Quick start:"
    echo "  make chaos-db-death         # Easy experiment (no root required)"
    echo "  make chaos-network-latency  # Medium experiment (requires root)"
    echo "  make chaos-all              # Run all experiments"
    echo ""
    exit 0
elif [[ $CHECKS_FAILED -gt 0 && $CHECKS_WARNING -gt 0 ]]; then
    echo -e "${YELLOW}⚠️  System has some issues but basic experiments can run${NC}"
    echo ""
    echo "You can still run:"
    echo "  make chaos-db-death  # This experiment should work"
    echo ""
    echo "To fix issues, install missing tools and start Docker services."
    exit 1
else
    echo -e "${RED}❌ System is not ready for chaos experiments${NC}"
    echo ""
    echo "Please fix the failed checks above before proceeding."
    exit 1
fi
