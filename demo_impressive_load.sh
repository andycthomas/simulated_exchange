#!/bin/bash

# Impressive Demo Load Test Script
# Optimized for 16-core, 30GB RAM system
# Safe, impressive, and maintains dashboard responsiveness

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080}"
DEMO_PROFILE="${1:-recommended}"

# Function to print colored messages
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_header() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  $1${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
    echo ""
}

# Function to check if API is running
check_api() {
    print_info "Checking API availability..."
    if curl -s -f "${API_URL}/health" > /dev/null 2>&1; then
        print_success "API is running at ${API_URL}"
        return 0
    else
        print_error "API is not running at ${API_URL}"
        print_info "Start the API with: go run cmd/api/main.go"
        exit 1
    fi
}

# Function to get system resources
show_system_resources() {
    print_header "System Resources"

    CORES=$(nproc)
    TOTAL_MEM=$(free -h | awk '/^Mem:/ {print $2}')
    AVAIL_MEM=$(free -h | awk '/^Mem:/ {print $7}')

    echo "  CPU Cores:       ${CORES}"
    echo "  Total Memory:    ${TOTAL_MEM}"
    echo "  Available Memory: ${AVAIL_MEM}"
    echo ""
}

# Function to start load test
start_load_test() {
    local profile=$1
    local name=$2
    local duration=$3
    local ops=$4
    local users=$5
    local symbols=$6

    print_header "Starting Load Test: ${name}"

    echo "  Configuration:"
    echo "    Duration:           ${duration}s ($(($duration / 60)) minutes)"
    echo "    Orders/second:      ${ops}"
    echo "    Concurrent users:   ${users}"
    echo "    Trading symbols:    ${symbols}"
    echo ""

    print_warning "Expected results:"
    echo "    Total orders:       ~$((ops * duration))"
    echo "    Total trades:       ~$((ops * duration * 60 / 100))"
    echo "    Memory usage:       ~$((ops * duration / 360))MB"
    echo ""

    print_info "Sending request to API..."

    RESPONSE=$(curl -s -X POST "${API_URL}/demo/load-test" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"${name}\",
            \"intensity\": \"${profile}\",
            \"duration\": ${duration},
            \"orders_per_second\": ${ops},
            \"concurrent_users\": ${users},
            \"symbols\": [${symbols}]
        }")

    if [ $? -eq 0 ]; then
        print_success "Load test started successfully!"
        echo ""
        echo "Response:"
        echo "${RESPONSE}" | jq '.' 2>/dev/null || echo "${RESPONSE}"
        echo ""
        print_info "Monitor progress:"
        echo "  📊 Dashboard:  ${API_URL}/dashboard"
        echo "  📈 Status API: ${API_URL}/demo/load-test/status"
        echo "  📉 Metrics:    ${API_URL}/api/metrics"
        echo ""
        print_info "To monitor in real-time, run:"
        echo "  watch -n 1 'curl -s ${API_URL}/demo/load-test/status | jq .'"
    else
        print_error "Failed to start load test"
        exit 1
    fi
}

# Function to monitor load test
monitor_load_test() {
    print_header "Monitoring Load Test"
    print_info "Fetching status every 5 seconds (Ctrl+C to stop monitoring)"
    echo ""

    while true; do
        STATUS=$(curl -s "${API_URL}/demo/load-test/status")

        if [ $? -eq 0 ]; then
            IS_RUNNING=$(echo "${STATUS}" | jq -r '.data.is_running' 2>/dev/null)

            if [ "${IS_RUNNING}" = "true" ]; then
                PROGRESS=$(echo "${STATUS}" | jq -r '.data.progress' 2>/dev/null)
                PHASE=$(echo "${STATUS}" | jq -r '.data.phase' 2>/dev/null)
                COMPLETED=$(echo "${STATUS}" | jq -r '.data.completed_orders' 2>/dev/null)
                FAILED=$(echo "${STATUS}" | jq -r '.data.failed_orders' 2>/dev/null)

                clear
                print_header "Load Test Status"
                echo "  Progress:        ${PROGRESS}%"
                echo "  Phase:           ${PHASE}"
                echo "  Orders Completed: ${COMPLETED}"
                echo "  Orders Failed:    ${FAILED}"
                echo ""
                echo "  Press Ctrl+C to stop monitoring (test will continue)"
                echo ""
            else
                clear
                print_success "Load test completed!"
                echo ""
                echo "Final Results:"
                echo "${STATUS}" | jq '.data' 2>/dev/null || echo "${STATUS}"
                break
            fi
        fi

        sleep 5
    done
}

# Function to show help
show_help() {
    cat << EOF
Impressive Demo Load Test Script

Usage: $0 [PROFILE] [COMMAND]

PROFILES:
  safe         - 1,000 ops/sec for 10 min (safe, good baseline)
  recommended  - 3,000 ops/sec for 15 min (RECOMMENDED for demos)
  aggressive   - 5,000 ops/sec for 10 min (impressive, near limits)
  marathon     - 2,000 ops/sec for 60 min (sustained capacity test)

COMMANDS:
  start        - Start the load test (default)
  monitor      - Monitor running load test
  stop         - Stop current load test
  status       - Show current status
  reset        - Reset system to clean state

EXAMPLES:
  $0 recommended          # Start recommended 3K ops/sec demo
  $0 aggressive start     # Start aggressive 5K ops/sec test
  $0 monitor              # Monitor currently running test
  $0 stop                 # Stop current test

ENVIRONMENT VARIABLES:
  API_URL      - API endpoint (default: http://localhost:8080)

For detailed information, see: PERFORMANCE_SCALING_RECOMMENDATIONS.md
EOF
}

# Main execution
print_header "Impressive Demo Load Test"

# Check if help requested
if [ "$1" = "help" ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

# Check API
check_api

# Show system resources
show_system_resources

# Handle commands
COMMAND="${2:-start}"

case "${COMMAND}" in
    start)
        case "${DEMO_PROFILE}" in
            safe)
                start_load_test "medium" "Safe Impressive Demo" 600 1000 50 \
                    '"BTCUSD", "ETHUSD", "ADAUSD", "DOTUSD"'
                ;;
            recommended)
                start_load_test "high" "Recommended Demo (3K ops/sec)" 900 3000 150 \
                    '"BTCUSD", "ETHUSD", "ADAUSD", "DOTUSD", "SOLUSD", "LINKUSD"'
                ;;
            aggressive)
                start_load_test "maximum" "Aggressive Demo (5K ops/sec)" 600 5000 250 \
                    '"BTCUSD", "ETHUSD", "ADAUSD", "DOTUSD", "SOLUSD", "LINKUSD", "AVAXUSD", "MATICUSD"'
                ;;
            marathon)
                start_load_test "sustained" "Marathon Test (1 hour)" 3600 2000 100 \
                    '"BTCUSD", "ETHUSD", "ADAUSD", "DOTUSD"'
                ;;
            *)
                print_error "Unknown profile: ${DEMO_PROFILE}"
                echo ""
                show_help
                exit 1
                ;;
        esac

        echo ""
        read -p "Would you like to monitor the test progress? (y/N) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            monitor_load_test
        fi
        ;;

    monitor)
        monitor_load_test
        ;;

    stop)
        print_info "Stopping load test..."
        RESPONSE=$(curl -s -X POST "${API_URL}/demo/load-test/stop")
        if [ $? -eq 0 ]; then
            print_success "Load test stopped"
            echo "${RESPONSE}" | jq '.' 2>/dev/null || echo "${RESPONSE}"
        else
            print_error "Failed to stop load test"
        fi
        ;;

    status)
        print_info "Fetching current status..."
        curl -s "${API_URL}/demo/load-test/status" | jq '.' 2>/dev/null
        ;;

    reset)
        print_warning "Resetting system to clean state..."
        RESPONSE=$(curl -s -X POST "${API_URL}/demo/reset")
        if [ $? -eq 0 ]; then
            print_success "System reset complete"
            echo "${RESPONSE}" | jq '.' 2>/dev/null || echo "${RESPONSE}"
        else
            print_error "Failed to reset system"
        fi
        ;;

    *)
        print_error "Unknown command: ${COMMAND}"
        echo ""
        show_help
        exit 1
        ;;
esac

print_success "Done!"
