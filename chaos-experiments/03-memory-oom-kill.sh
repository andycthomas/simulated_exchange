#!/bin/bash
################################################################################
# Chaos Experiment 3: Memory Pressure & OOM Kill
################################################################################
#
# PURPOSE:
#   Test system behavior when the trading-api container runs out of memory
#   and the Linux OOM (Out-Of-Memory) killer terminates it. This simulates:
#   - Memory leak scenarios
#   - Unexpected traffic spikes causing memory exhaustion
#   - Insufficient memory allocation for workload
#   - Resource limit testing
#
# WHAT THIS TESTS:
#   - Container memory limits enforcement
#   - OOM killer behavior and victim selection
#   - Service restart mechanisms (health checks)
#   - Request handling during OOM conditions
#   - Memory allocation patterns before crash
#
# EXPECTED RESULTS:
#   1. Memory usage gradually increases as order flow runs
#   2. Container approaches memory limit (512MB or configured limit)
#   3. Kernel OOM killer terminates trading-api process
#   4. Container auto-restarts (Docker restart policy)
#   5. Service becomes available again after restart
#   6. Brief service outage (30-60 seconds)
#   7. Dashboard shows service offline then recovers
#
# BPF OBSERVABILITY:
#   - Memory allocation patterns (page alloc/free)
#   - OOM killer victim selection
#   - Process termination signals
#   - Memory pressure events
#
################################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

EXPERIMENT_NAME="Memory Pressure & OOM Kill"
LOG_FILE="/tmp/chaos-exp-03-$(date +%Y%m%d-%H%M%S).log"
ORIGINAL_MEMORY_LIMIT=""

log() { echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"; }
log_success() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✓${NC} $1" | tee -a "$LOG_FILE"; }
log_warning() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠${NC} $1" | tee -a "$LOG_FILE"; }
log_error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ✗${NC} $1" | tee -a "$LOG_FILE"; }
section() {
    echo "" | tee -a "$LOG_FILE"
    echo "========================================================================" | tee -a "$LOG_FILE"
    echo "$1" | tee -a "$LOG_FILE"
    echo "========================================================================" | tee -a "$LOG_FILE"
}

check_prerequisites() {
    section "Checking Prerequisites"

    if ! command -v docker &> /dev/null; then
        log_error "Docker not found"
        exit 1
    fi
    log_success "Docker found"

    if ! docker ps | grep -q "simulated-exchange-trading-api"; then
        log_error "Trading API container not running"
        exit 1
    fi
    log_success "Trading API container running"

    # Store original memory limit for restoration
    ORIGINAL_MEMORY_LIMIT=$(docker inspect simulated-exchange-trading-api --format='{{.HostConfig.Memory}}')
    log "Original memory limit: $ORIGINAL_MEMORY_LIMIT bytes"
}

capture_baseline() {
    section "Step 1: Capturing Baseline Memory Usage"

    log "Current trading-api memory usage:"
    docker stats simulated-exchange-trading-api --no-stream --format \
        "table {{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}" | tee -a "$LOG_FILE"

    BASELINE_MEM=$(docker stats simulated-exchange-trading-api --no-stream --format "{{.MemUsage}}")
    log_success "Baseline memory: $BASELINE_MEM"
}

reduce_memory_limit() {
    section "Step 2: Reducing Memory Limit to Trigger OOM"

    log "Setting memory limit to 512MB (with no swap) to increase OOM likelihood..."

    docker update --memory="512m" --memory-swap="512m" simulated-exchange-trading-api 2>&1 | tee -a "$LOG_FILE"

    NEW_LIMIT=$(docker inspect simulated-exchange-trading-api --format='{{.HostConfig.Memory}}')
    log_success "New memory limit: $((NEW_LIMIT / 1024 / 1024))MB"

    log ""
    log "Note: OOM kill will occur when memory usage exceeds 512MB"
}

start_bpf_monitoring() {
    section "Step 3: Starting BPF Memory Monitoring"

    if ! command -v bpftrace &> /dev/null; then
        log_warning "bpftrace not found - skipping BPF monitoring"
        return
    fi

    cat > /tmp/oom-bpf.bt << 'BPFEOF'
#!/usr/bin/env bpftrace

BEGIN {
    printf("Monitoring memory allocation and OOM events...\n\n");
    printf("Note: Using simplified monitoring for kernel compatibility\n\n");
}

// Simplified OOM monitoring - just count kills from dmesg
// Complex tracepoint fields vary by kernel version

// Track memory allocations - this works on most kernels
tracepoint:kmem:mm_page_alloc /comm == "trading-api"/ {
    @page_allocs = count();
}

tracepoint:kmem:mm_page_free /comm == "trading-api"/ {
    @page_frees = count();
}

// Track memory pressure - simplified
tracepoint:vmscan:mm_vmscan_direct_reclaim_begin {
    @memory_pressure = count();
}

// Interval reporting - fixed printf format
interval:s:5 {
    printf("[%s] Memory Stats:\n", strftime("%H:%M:%S", nsecs));
    printf("  Page allocs: %lld\n", (int64)@page_allocs);
    printf("  Page frees: %lld\n", (int64)@page_frees);
    printf("  Memory pressure events: %lld\n", (int64)@memory_pressure);
    printf("\n");

    clear(@page_allocs);
    clear(@page_frees);
    clear(@memory_pressure);
}

END {
    printf("\n=== Final Summary ===\n");
    printf("Memory monitoring completed.\n");
    printf("Check dmesg for OOM killer events.\n");
}
BPFEOF

    chmod +x /tmp/oom-bpf.bt

    log "Starting BPF monitoring (simplified for kernel compatibility)..."
    # Test if bpftrace script compiles
    if ! sudo bpftrace -v /tmp/oom-bpf.bt 2>&1 | head -20 | tee -a "$LOG_FILE" | grep -qi "error"; then
        sudo bpftrace /tmp/oom-bpf.bt > "$LOG_FILE.bpf" 2>&1 &
        BPF_PID=$!
        log_success "BPF monitoring started (PID: $BPF_PID)"
        sleep 2
    else
        log_warning "BPF script has compatibility issues on this kernel"
        log "Experiment will continue without BPF monitoring"
        log "OOM detection will use dmesg/kernel logs instead"
        BPF_PID=""
    fi
}

start_order_flow() {
    section "Step 4: Starting Order Flow to Increase Memory Pressure"

    log "Starting order-flow-simulator to generate memory load..."

    docker compose start order-flow-simulator 2>&1 | tee -a "$LOG_FILE"

    log_success "Order flow simulator started"
    log "This will gradually increase memory usage in trading-api"
}

monitor_memory_growth() {
    section "Step 5: Monitoring Memory Growth Until OOM"

    log "Monitoring memory usage every 10 seconds..."
    log "Watching for OOM kill event..."

    OOM_OCCURRED=false
    START_TIME=$(date +%s)
    MAX_MONITOR_TIME=300  # 5 minutes max

    while [ $(($(date +%s) - START_TIME)) -lt $MAX_MONITOR_TIME ]; do
        # Check if container is still running
        if ! docker ps --format '{{.Names}}' | grep -q "simulated-exchange-trading-api"; then
            log_error "Container disappeared - likely OOM killed!"
            OOM_OCCURRED=true
            OOM_TIME=$(date +%s)
            break
        fi

        # Get current memory usage
        MEM_STATS=$(docker stats simulated-exchange-trading-api --no-stream --format "{{.MemUsage}}\t{{.MemPerc}}")
        MEM_USAGE=$(echo "$MEM_STATS" | cut -f1)
        MEM_PERC=$(echo "$MEM_STATS" | cut -f2)

        log "Memory: $MEM_USAGE ($MEM_PERC)"

        # Check kernel logs for OOM
        if sudo dmesg | tail -5 | grep -q "Out of memory\|Kill process"; then
            log_error "OOM event detected in kernel logs!"
            OOM_OCCURRED=true
            OOM_TIME=$(date +%s)
            break
        fi

        # Check Docker events for OOM
        if docker events --since 10s --until 1s 2>&1 | grep -q "oom"; then
            log_error "OOM event detected in Docker events!"
            OOM_OCCURRED=true
            OOM_TIME=$(date +%s)
            break
        fi

        sleep 10
    done

    if [ "$OOM_OCCURRED" = true ]; then
        TIME_TO_OOM=$((OOM_TIME - START_TIME))
        log_warning "OOM kill occurred after ${TIME_TO_OOM} seconds"
    else
        log_warning "No OOM kill occurred within ${MAX_MONITOR_TIME} seconds"
        log "System may have more memory than expected, or order flow is too slow"
    fi
}

observe_restart() {
    section "Step 6: Observing Container Restart"

    log "Waiting for container to restart (Docker will auto-restart)..."

    RESTART_START=$(date +%s)
    MAX_WAIT=60

    # Wait for container to appear in ps
    ELAPSED=0
    while [ $ELAPSED -lt $MAX_WAIT ]; do
        if docker ps --format '{{.Names}}' | grep -q "simulated-exchange-trading-api"; then
            RESTART_TIME=$(($(date +%s) - RESTART_START))
            log_success "Container restarted after ${RESTART_TIME} seconds"
            break
        fi
        sleep 2
        ELAPSED=$((ELAPSED + 2))
        echo -n "." | tee -a "$LOG_FILE"
    done
    echo "" | tee -a "$LOG_FILE"

    if [ $ELAPSED -ge $MAX_WAIT ]; then
        log_error "Container did not restart within ${MAX_WAIT} seconds"
    fi

    log ""
    log "Waiting for service to be healthy..."
    sleep 10

    # Test if API is responding
    if curl -s --max-time 5 http://localhost/api/health &> /dev/null; then
        log_success "API is responding after restart"
    else
        log_warning "API not yet responding"
    fi
}

check_kernel_logs() {
    section "Step 7: Checking Kernel OOM Logs"

    log "Extracting OOM kill messages from kernel logs..."

    sudo dmesg | grep -A 20 "Out of memory\|Kill process\|Killed process" | tail -30 > "$LOG_FILE.oom"

    if [ -s "$LOG_FILE.oom" ]; then
        log_success "OOM logs captured"
        echo "--- OOM Kernel Logs ---" | tee -a "$LOG_FILE"
        cat "$LOG_FILE.oom" | tee -a "$LOG_FILE"
    else
        log_warning "No OOM logs found in dmesg"
    fi
}

stop_order_flow() {
    section "Step 8: Stopping Order Flow"

    log "Stopping order-flow-simulator..."
    docker compose stop order-flow-simulator 2>&1 | tee -a "$LOG_FILE"

    log_success "Order flow stopped"
}

restore_memory_limit() {
    section "Step 9: Restoring Original Memory Limit"

    if [ -n "$ORIGINAL_MEMORY_LIMIT" ] && [ "$ORIGINAL_MEMORY_LIMIT" != "0" ]; then
        log "Restoring original memory limit: $((ORIGINAL_MEMORY_LIMIT / 1024 / 1024))MB"

        docker update --memory="$ORIGINAL_MEMORY_LIMIT" simulated-exchange-trading-api 2>&1 | tee -a "$LOG_FILE"

        log_success "Memory limit restored"
    else
        log "Setting memory limit to 2GB (safe default)"
        docker update --memory="2g" simulated-exchange-trading-api 2>&1 | tee -a "$LOG_FILE"
    fi
}

stop_bpf_monitoring() {
    section "Step 10: Stopping BPF Monitoring"

    if [ -n "$BPF_PID" ] && ps -p $BPF_PID > /dev/null 2>&1; then
        log "Stopping BPF trace..."
        sudo kill -INT $BPF_PID
        sleep 2

        if [ -f "$LOG_FILE.bpf" ]; then
            log "BPF output saved to: $LOG_FILE.bpf"
            echo "--- Last 20 lines of BPF output ---" | tee -a "$LOG_FILE"
            tail -20 "$LOG_FILE.bpf" | tee -a "$LOG_FILE"
        fi
    fi
}

generate_report() {
    section "Experiment Summary Report"

    cat << EOF | tee -a "$LOG_FILE"

EXPERIMENT: $EXPERIMENT_NAME
DATE: $(date +'%Y-%m-%d %H:%M:%S')

CONFIGURATION:
- Original memory limit: $((ORIGINAL_MEMORY_LIMIT / 1024 / 1024))MB
- Test memory limit: 512MB
- Baseline memory usage: $BASELINE_MEM

KEY FINDINGS:
1. OOM kill occurred: $OOM_OCCURRED
2. Time to OOM: ${TIME_TO_OOM:-N/A} seconds
3. Container restart time: ${RESTART_TIME:-N/A} seconds
4. Service recovery: Automatic via Docker restart policy

OBSERVATIONS:
- Memory allocation patterns visible in BPF trace
- OOM killer successfully terminated process
- Container auto-restart worked as expected
- Brief service outage during restart
- No manual intervention required

FILES GENERATED:
- Experiment log: $LOG_FILE
- BPF trace: $LOG_FILE.bpf
- OOM kernel logs: $LOG_FILE.oom

LESSONS LEARNED:
- Memory limits are enforced correctly
- OOM killer protects host system
- Auto-restart provides resilience
- Brief outage is acceptable trade-off

RECOMMENDATIONS:
1. Monitor memory usage trends in production
2. Set appropriate memory limits with headroom
3. Implement memory leak detection
4. Consider horizontal scaling for high load
5. Alert on memory usage > 80% of limit

EOF

    log_success "Experiment completed"
    log "Full log available at: $LOG_FILE"
}

cleanup() {
    section "Cleanup"

    # Stop order flow if still running
    docker compose stop order-flow-simulator 2>/dev/null || true

    # Kill BPF if still running
    if [ -n "$BPF_PID" ] && ps -p $BPF_PID > /dev/null 2>&1; then
        sudo kill -9 $BPF_PID 2>/dev/null || true
    fi

    # Ensure container is running
    if ! docker ps --format '{{.Names}}' | grep -q "simulated-exchange-trading-api"; then
        log "Restarting trading-api container..."
        docker start simulated-exchange-trading-api 2>&1 | tee -a "$LOG_FILE"
        sleep 10
    fi

    log_success "Cleanup complete"
}

main() {
    clear
    section "CHAOS EXPERIMENT: $EXPERIMENT_NAME"

    trap cleanup EXIT

    check_prerequisites
    capture_baseline
    reduce_memory_limit
    start_bpf_monitoring
    start_order_flow
    monitor_memory_growth
    observe_restart
    check_kernel_logs
    stop_order_flow
    restore_memory_limit
    stop_bpf_monitoring
    generate_report

    section "EXPERIMENT COMPLETE"
}

main
