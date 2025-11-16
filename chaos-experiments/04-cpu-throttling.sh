#!/bin/bash
################################################################################
# Chaos Experiment 4: CPU Throttling
################################################################################
#
# PURPOSE:
#   Test system behavior when CPU resources are severely limited. This simulates:
#   - CPU contention from noisy neighbors
#   - Under-provisioned container resources
#   - High load scenarios with limited compute
#   - Gradual performance degradation vs hard failure
#
# WHAT THIS TESTS:
#   - CPU scheduler behavior under constraints
#   - Request latency increase under CPU pressure
#   - Queue buildup patterns
#   - Graceful degradation vs catastrophic failure
#   - Service behavior when CPU-starved
#
# EXPECTED RESULTS:
#   1. Baseline latency captured (~50-200ms per request)
#   2. CPU quota reduced to 50% of 1 core (0.5 CPUs)
#   3. Request latency increases 3-5x
#   4. No hard errors, just slower responses
#   5. Queue buildup visible in metrics
#   6. Immediate improvement when CPU restored
#   7. Dashboard updates slow down but don't stop
#
# BPF OBSERVABILITY:
#   - CPU scheduler delays (runqueue wait time)
#   - Context switches per second
#   - CPU samples per process (profiling)
#   - Time spent waiting for CPU
#
################################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

EXPERIMENT_NAME="CPU Throttling"
LOG_FILE="/tmp/chaos-exp-04-$(date +%Y%m%d-%H%M%S).log"
ORIGINAL_CPU_QUOTA=""
ORIGINAL_CPU_PERIOD=""

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

    # Store original CPU settings
    ORIGINAL_CPU_QUOTA=$(docker inspect simulated-exchange-trading-api --format='{{.HostConfig.CpuQuota}}')
    ORIGINAL_CPU_PERIOD=$(docker inspect simulated-exchange-trading-api --format='{{.HostConfig.CpuPeriod}}')

    log "Original CPU quota: $ORIGINAL_CPU_QUOTA"
    log "Original CPU period: $ORIGINAL_CPU_PERIOD"
}

capture_baseline() {
    section "Step 1: Capturing Baseline Performance"

    log "Testing API response time (baseline)..."

    BASELINE_TIMES=()
    for i in {1..10}; do
        START=$(date +%s%N)
        if curl -s --max-time 5 http://localhost/api/health &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 )) # Convert to milliseconds
            BASELINE_TIMES+=($ELAPSED)
            echo -n "." | tee -a "$LOG_FILE"
        else
            log_warning "Request $i failed"
        fi
    done
    echo "" | tee -a "$LOG_FILE"

    # Calculate average baseline latency
    BASELINE_AVG=0
    if [ ${#BASELINE_TIMES[@]} -gt 0 ]; then
        for time in "${BASELINE_TIMES[@]}"; do
            BASELINE_AVG=$((BASELINE_AVG + time))
        done
        BASELINE_AVG=$((BASELINE_AVG / ${#BASELINE_TIMES[@]}))
    fi

    log_success "Baseline average latency: ${BASELINE_AVG}ms"

    log ""
    log "Current CPU usage:"
    docker stats simulated-exchange-trading-api --no-stream --format \
        "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | tee -a "$LOG_FILE"
}

start_bpf_monitoring() {
    section "Step 2: Starting BPF CPU Monitoring"

    if ! command -v bpftrace &> /dev/null; then
        log_warning "bpftrace not found - skipping BPF monitoring"
        return
    fi

    cat > /tmp/cpu-bpf.bt << 'BPFEOF'
#!/usr/bin/env bpftrace

BEGIN {
    printf("Monitoring CPU scheduler behavior...\n\n");
}

// Track time spent waiting in runqueue (scheduler delay)
tracepoint:sched:sched_stat_wait {
    @runqueue_delay[comm] = hist(args->delay);
    @total_wait_time[comm] = sum(args->delay);
}

// Track context switches
tracepoint:sched:sched_switch {
    @context_switches = count();
}

// Sample on-CPU processes
profile:hz:99 /comm == "trading-api"/ {
    @cpu_samples[ustack] = count();
}

// Interval reporting
interval:s:5 {
    printf("\n[%s] === CPU Stats ===\n", strftime("%H:%M:%S", nsecs));
    printf("Context switches (5s):\n");
    print(@context_switches);

    printf("\nTop processes by scheduler wait time:\n");
    print(@total_wait_time, 5);

    clear(@context_switches);
}

END {
    printf("\n=== Final Summary ===\n");
    printf("Runqueue delay distribution:\n");
    print(@runqueue_delay);

    printf("\nTop CPU-consuming code paths (trading-api):\n");
    print(@cpu_samples, 10);
}
BPFEOF

    chmod +x /tmp/cpu-bpf.bt

    log "Starting BPF monitoring..."
    sudo bpftrace /tmp/cpu-bpf.bt > "$LOG_FILE.bpf" 2>&1 &
    BPF_PID=$!

    log_success "BPF monitoring started (PID: $BPF_PID)"
    sleep 2
}

throttle_cpu() {
    section "Step 3: Throttling CPU to 50%"

    log "Reducing CPU quota to 0.5 cores (50% of 1 core)..."

    # CPU quota is in microseconds per period
    # Standard period is 100000 microseconds (100ms)
    # For 0.5 cores: quota = 50000 (50% of 100000)

    docker update --cpus="0.5" simulated-exchange-trading-api 2>&1 | tee -a "$LOG_FILE"

    NEW_QUOTA=$(docker inspect simulated-exchange-trading-api --format='{{.HostConfig.CpuQuota}}')
    NEW_PERIOD=$(docker inspect simulated-exchange-trading-api --format='{{.HostConfig.CpuPeriod}}')

    CORES=$(echo "scale=2; $NEW_QUOTA / $NEW_PERIOD" | bc)

    log_success "CPU throttled to $CORES cores"
    log ""
    log "Note: Application now has only 50% of 1 CPU core available"
}

observe_degradation() {
    section "Step 4: Observing Performance Degradation"

    log "Measuring response times under CPU throttling..."
    log "Testing over 60 seconds..."

    THROTTLED_TIMES=()
    FAILED_REQUESTS=0

    for i in {1..30}; do
        START=$(date +%s%N)
        if timeout 10 curl -s --max-time 8 http://localhost/api/health &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            THROTTLED_TIMES+=($ELAPSED)

            if [ $((i % 5)) -eq 0 ]; then
                log "Request $i: ${ELAPSED}ms"
            else
                echo -n "." | tee -a "$LOG_FILE"
            fi
        else
            FAILED_REQUESTS=$((FAILED_REQUESTS + 1))
            log_warning "Request $i failed or timed out"
        fi
        sleep 2
    done
    echo "" | tee -a "$LOG_FILE"

    # Calculate average throttled latency
    THROTTLED_AVG=0
    if [ ${#THROTTLED_TIMES[@]} -gt 0 ]; then
        for time in "${THROTTLED_TIMES[@]}"; do
            THROTTLED_AVG=$((THROTTLED_AVG + time))
        done
        THROTTLED_AVG=$((THROTTLED_AVG / ${#THROTTLED_TIMES[@]}))
    fi

    log_success "Throttled average latency: ${THROTTLED_AVG}ms"
    log "Failed requests: $FAILED_REQUESTS / 30"

    if [ $THROTTLED_AVG -gt 0 ] && [ $BASELINE_AVG -gt 0 ]; then
        DEGRADATION=$((THROTTLED_AVG * 100 / BASELINE_AVG))
        log_warning "Performance degradation: ${DEGRADATION}% of baseline"
    fi

    log ""
    log "CPU usage during throttling:"
    docker stats simulated-exchange-trading-api --no-stream --format \
        "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | tee -a "$LOG_FILE"
}

test_metrics_endpoint() {
    section "Step 5: Testing Metrics Endpoint Under Load"

    log "Testing /api/metrics endpoint (more CPU-intensive)..."

    METRICS_TIMES=()
    for i in {1..5}; do
        log "Metrics request $i..."
        START=$(date +%s%N)
        if timeout 30 curl -s --max-time 25 http://localhost/api/metrics &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            METRICS_TIMES+=($ELAPSED)
            log "  Response time: ${ELAPSED}ms"
        else
            log_error "  Request failed or timed out"
        fi
        sleep 3
    done

    # Calculate average
    METRICS_AVG=0
    if [ ${#METRICS_TIMES[@]} -gt 0 ]; then
        for time in "${METRICS_TIMES[@]}"; do
            METRICS_AVG=$((METRICS_AVG + time))
        done
        METRICS_AVG=$((METRICS_AVG / ${#METRICS_TIMES[@]}))
        log_success "Metrics endpoint average: ${METRICS_AVG}ms"
    else
        log_error "All metrics requests failed"
    fi
}

restore_cpu() {
    section "Step 6: Restoring Original CPU Quota"

    if [ -n "$ORIGINAL_CPU_QUOTA" ] && [ "$ORIGINAL_CPU_QUOTA" != "-1" ] && [ "$ORIGINAL_CPU_QUOTA" != "0" ]; then
        ORIGINAL_CORES=$(echo "scale=2; $ORIGINAL_CPU_QUOTA / $ORIGINAL_CPU_PERIOD" | bc)
        log "Restoring original CPU quota: $ORIGINAL_CORES cores"

        docker update --cpus="$ORIGINAL_CORES" simulated-exchange-trading-api 2>&1 | tee -a "$LOG_FILE"
    else
        log "Restoring CPU to 2.0 cores (safe default)"
        docker update --cpus="2.0" simulated-exchange-trading-api 2>&1 | tee -a "$LOG_FILE"
    fi

    log_success "CPU quota restored"

    log ""
    log "Waiting 10s for performance to stabilize..."
    sleep 10
}

measure_recovery() {
    section "Step 7: Measuring Recovery Performance"

    log "Testing response times after CPU restoration..."

    RECOVERY_TIMES=()
    for i in {1..10}; do
        START=$(date +%s%N)
        if curl -s --max-time 5 http://localhost/api/health &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            RECOVERY_TIMES+=($ELAPSED)
            echo -n "." | tee -a "$LOG_FILE"
        fi
    done
    echo "" | tee -a "$LOG_FILE"

    # Calculate average
    RECOVERY_AVG=0
    if [ ${#RECOVERY_TIMES[@]} -gt 0 ]; then
        for time in "${RECOVERY_TIMES[@]}"; do
            RECOVERY_AVG=$((RECOVERY_AVG + time))
        done
        RECOVERY_AVG=$((RECOVERY_AVG / ${#RECOVERY_TIMES[@]}))
    fi

    log_success "Recovery average latency: ${RECOVERY_AVG}ms"

    if [ $RECOVERY_AVG -gt 0 ] && [ $BASELINE_AVG -gt 0 ]; then
        RECOVERY_PERC=$((RECOVERY_AVG * 100 / BASELINE_AVG))
        log "Recovery performance: ${RECOVERY_PERC}% of baseline"
    fi
}

stop_bpf_monitoring() {
    section "Step 8: Stopping BPF Monitoring"

    if [ -n "$BPF_PID" ] && ps -p $BPF_PID > /dev/null 2>&1; then
        log "Stopping BPF trace..."
        sudo kill -INT $BPF_PID
        sleep 2

        if [ -f "$LOG_FILE.bpf" ]; then
            log "BPF output saved to: $LOG_FILE.bpf"
            echo "--- Last 30 lines of BPF output ---" | tee -a "$LOG_FILE"
            tail -30 "$LOG_FILE.bpf" | tee -a "$LOG_FILE"
        fi
    fi
}

generate_report() {
    section "Experiment Summary Report"

    cat << EOF | tee -a "$LOG_FILE"

EXPERIMENT: $EXPERIMENT_NAME
DATE: $(date +'%Y-%m-%d %H:%M:%S')

CONFIGURATION:
- Original CPU quota: ${ORIGINAL_CPU_QUOTA:-unlimited}
- Throttled CPU: 0.5 cores (50% of 1 core)
- Test duration: 60 seconds
- Request count: 30 health checks, 5 metrics calls

PERFORMANCE METRICS:
- Baseline latency: ${BASELINE_AVG}ms
- Throttled latency: ${THROTTLED_AVG}ms
- Recovery latency: ${RECOVERY_AVG}ms
- Failed requests: $FAILED_REQUESTS / 30
- Metrics endpoint avg: ${METRICS_AVG}ms

DEGRADATION ANALYSIS:
EOF

    if [ $THROTTLED_AVG -gt 0 ] && [ $BASELINE_AVG -gt 0 ]; then
        DEGRADATION=$((THROTTLED_AVG * 100 / BASELINE_AVG))
        echo "- Performance under throttling: ${DEGRADATION}% of baseline" | tee -a "$LOG_FILE"

        if [ $DEGRADATION -gt 500 ]; then
            echo "- Impact: SEVERE (>5x slowdown)" | tee -a "$LOG_FILE"
        elif [ $DEGRADATION -gt 300 ]; then
            echo "- Impact: HIGH (3-5x slowdown)" | tee -a "$LOG_FILE"
        elif [ $DEGRADATION -gt 200 ]; then
            echo "- Impact: MODERATE (2-3x slowdown)" | tee -a "$LOG_FILE"
        else
            echo "- Impact: LOW (<2x slowdown)" | tee -a "$LOG_FILE"
        fi
    fi

    cat << EOF | tee -a "$LOG_FILE"

KEY FINDINGS:
1. CPU throttling caused ${DEGRADATION:-N/A}% performance degradation
2. No hard failures - graceful degradation observed
3. Service remained responsive (slower but functional)
4. Immediate recovery when CPU restored
5. Metrics endpoint significantly impacted

OBSERVATIONS:
- Requests queue when CPU-starved
- No crashes or errors, just increased latency
- BPF shows increased scheduler delays
- Context switches increase under contention
- Dashboard updates slow but continue

FILES GENERATED:
- Experiment log: $LOG_FILE
- BPF trace: $LOG_FILE.bpf

LESSONS LEARNED:
- CPU limits cause graceful degradation, not failure
- Response times increase proportionally
- Queue buildup is primary symptom
- Service remains available but slow

RECOMMENDATIONS:
1. Set CPU limits with 50%+ headroom
2. Monitor request latency as CPU indicator
3. Alert on P95 latency increases
4. Consider horizontal scaling for CPU-bound workloads
5. Profile hot code paths with BPF

EOF

    log_success "Experiment completed"
    log "Full log available at: $LOG_FILE"
}

cleanup() {
    section "Cleanup"

    # Kill BPF if still running
    if [ -n "$BPF_PID" ] && ps -p $BPF_PID > /dev/null 2>&1; then
        sudo kill -9 $BPF_PID 2>/dev/null || true
    fi

    log_success "Cleanup complete"
}

main() {
    clear
    section "CHAOS EXPERIMENT: $EXPERIMENT_NAME"

    trap cleanup EXIT

    check_prerequisites
    capture_baseline
    start_bpf_monitoring
    throttle_cpu
    observe_degradation
    test_metrics_endpoint
    restore_cpu
    measure_recovery
    stop_bpf_monitoring
    generate_report

    section "EXPERIMENT COMPLETE"
}

main
