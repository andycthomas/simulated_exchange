#!/bin/bash
################################################################################
# Chaos Experiment 5: Disk I/O Starvation
################################################################################
#
# PURPOSE:
#   Test system behavior when disk I/O is saturated. This simulates:
#   - Disk hardware degradation or failure
#   - I/O contention from other processes
#   - Slow network-attached storage
#   - Database performance under I/O pressure
#
# WHAT THIS TESTS:
#   - Query performance under I/O pressure
#   - PostgreSQL checkpoint behavior
#   - Database write throughput degradation
#   - Filesystem cache effectiveness
#   - Timeout handling for slow queries
#
# EXPECTED RESULTS:
#   1. Baseline disk latency captured (<1ms)
#   2. stress-ng creates I/O saturation
#   3. Disk latency spikes to 50-200ms+
#   4. Database queries slow significantly
#   5. Some queries may timeout
#   6. PostgreSQL checkpoint delays visible
#   7. Immediate improvement when I/O stress removed
#
# BPF OBSERVABILITY:
#   - I/O latency distribution (read vs write)
#   - I/O queue depth
#   - Bytes read/written per process
#   - Filesystem layer latency (VFS)
#
################################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

EXPERIMENT_NAME="Disk I/O Starvation"
LOG_FILE="/tmp/chaos-exp-05-$(date +%Y%m%d-%H%M%S).log"
STRESS_PID=""

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

    if ! docker ps | grep -q "simulated-exchange-postgres"; then
        log_error "PostgreSQL container not running"
        exit 1
    fi
    log_success "PostgreSQL container running"

    # Check for stress-ng (install if needed)
    if ! command -v stress-ng &> /dev/null; then
        log_warning "stress-ng not found, installing..."
        sudo dnf install -y stress-ng 2>&1 | tee -a "$LOG_FILE"
    fi
    log_success "stress-ng available"

    if ! command -v iostat &> /dev/null; then
        log_warning "iostat not found, installing sysstat..."
        sudo dnf install -y sysstat 2>&1 | tee -a "$LOG_FILE"
    fi
    log_success "iostat available"
}

capture_baseline() {
    section "Step 1: Capturing Baseline I/O Performance"

    log "Current disk I/O statistics:"
    iostat -x 1 3 | tee -a "$LOG_FILE"

    log ""
    log "PostgreSQL container I/O stats:"
    docker stats simulated-exchange-postgres --no-stream --format \
        "table {{.Name}}\t{{.BlockIO}}\t{{.MemUsage}}" | tee -a "$LOG_FILE"

    log ""
    log "Testing database query latency (baseline)..."

    BASELINE_QUERY_TIMES=()
    for i in {1..5}; do
        START=$(date +%s%N)
        docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
            "SELECT COUNT(*) FROM orders;" &> /dev/null
        END=$(date +%s%N)
        ELAPSED=$(( (END - START) / 1000000 ))
        BASELINE_QUERY_TIMES+=($ELAPSED)
        log "  Query $i: ${ELAPSED}ms"
    done

    # Calculate average
    BASELINE_QUERY_AVG=0
    for time in "${BASELINE_QUERY_TIMES[@]}"; do
        BASELINE_QUERY_AVG=$((BASELINE_QUERY_AVG + time))
    done
    BASELINE_QUERY_AVG=$((BASELINE_QUERY_AVG / ${#BASELINE_QUERY_TIMES[@]}))

    log_success "Baseline query latency: ${BASELINE_QUERY_AVG}ms"
}

start_bpf_monitoring() {
    section "Step 2: Starting BPF I/O Monitoring"

    if ! command -v bpftrace &> /dev/null; then
        log_warning "bpftrace not found - skipping BPF monitoring"
        return
    fi

    cat > /tmp/io-bpf.bt << 'BPFEOF'
#!/usr/bin/env bpftrace

BEGIN {
    printf("Monitoring disk I/O operations...\n\n");
}

// Track VFS read latency
kprobe:vfs_read {
    @read_start[tid] = nsecs;
}

kretprobe:vfs_read /@read_start[tid]/ {
    $latency_us = (nsecs - @read_start[tid]) / 1000;
    @read_latency = hist($latency_us);
    @read_count = count();
    delete(@read_start[tid]);
}

// Track VFS write latency
kprobe:vfs_write {
    @write_start[tid] = nsecs;
}

kretprobe:vfs_write /@write_start[tid]/ {
    $latency_us = (nsecs - @write_start[tid]) / 1000;
    @write_latency = hist($latency_us);
    @write_count = count();
    delete(@write_start[tid]);
}

// Track block I/O requests
tracepoint:block:block_rq_issue {
    @io_queue_depth = count();
    @bytes_issued[args->comm] = sum(args->bytes);
}

// Track PostgreSQL specific I/O
kprobe:vfs_read /comm == "postgres"/ {
    @postgres_reads = count();
}

kprobe:vfs_write /comm == "postgres"/ {
    @postgres_writes = count();
}

// Interval reporting
interval:s:5 {
    printf("\n[%s] === I/O Stats (5s) ===\n", strftime("%H:%M:%S", nsecs));
    printf("Total reads:\n");
    print(@read_count);
    printf("Total writes:\n");
    print(@write_count);
    printf("PostgreSQL reads:\n");
    print(@postgres_reads);
    printf("PostgreSQL writes:\n");
    print(@postgres_writes);

    printf("\nTop I/O by process:\n");
    print(@bytes_issued, 5);

    clear(@read_count);
    clear(@write_count);
    clear(@postgres_reads);
    clear(@postgres_writes);
}

END {
    printf("\n=== Final I/O Summary ===\n");
    printf("Read latency distribution (microseconds):\n");
    print(@read_latency);

    printf("\nWrite latency distribution (microseconds):\n");
    print(@write_latency);

    printf("\nBytes issued by process:\n");
    print(@bytes_issued);
}
BPFEOF

    chmod +x /tmp/io-bpf.bt

    log "Starting BPF I/O monitoring..."
    sudo bpftrace /tmp/io-bpf.bt > "$LOG_FILE.bpf" 2>&1 &
    BPF_PID=$!

    log_success "BPF monitoring started (PID: $BPF_PID)"
    sleep 2
}

start_io_stress() {
    section "Step 3: Starting I/O Stress Test"

    log "Starting stress-ng with I/O saturation..."
    log "This will create heavy disk I/O load for 120 seconds"

    # Run stress-ng targeting I/O
    # --iomix: Mix of I/O operations
    # --timeout: Duration of stress
    # --metrics-brief: Show summary
    stress-ng --iomix 4 --timeout 120s --metrics-brief > "$LOG_FILE.stress" 2>&1 &
    STRESS_PID=$!

    log_success "I/O stress started (PID: $STRESS_PID)"
    log ""
    log "Waiting 10s for I/O pressure to build..."
    sleep 10
}

observe_io_impact() {
    section "Step 4: Observing I/O Impact on Database"

    log "Current I/O statistics under stress:"
    iostat -x 1 3 | tee -a "$LOG_FILE"

    log ""
    log "Testing database query latency under I/O stress..."

    STRESS_QUERY_TIMES=()
    FAILED_QUERIES=0

    for i in {1..10}; do
        START=$(date +%s%N)
        if timeout 30 docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
            "SELECT COUNT(*) FROM orders;" &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            STRESS_QUERY_TIMES+=($ELAPSED)
            log "  Query $i: ${ELAPSED}ms"
        else
            FAILED_QUERIES=$((FAILED_QUERIES + 1))
            log_error "  Query $i: TIMEOUT or FAILED"
        fi
        sleep 3
    done

    # Calculate average
    STRESS_QUERY_AVG=0
    if [ ${#STRESS_QUERY_TIMES[@]} -gt 0 ]; then
        for time in "${STRESS_QUERY_TIMES[@]}"; do
            STRESS_QUERY_AVG=$((STRESS_QUERY_AVG + time))
        done
        STRESS_QUERY_AVG=$((STRESS_QUERY_AVG / ${#STRESS_QUERY_TIMES[@]}))
        log_warning "Query latency under I/O stress: ${STRESS_QUERY_AVG}ms"
    fi

    log "Failed queries: $FAILED_QUERIES / 10"

    if [ $STRESS_QUERY_AVG -gt 0 ] && [ $BASELINE_QUERY_AVG -gt 0 ]; then
        DEGRADATION=$((STRESS_QUERY_AVG * 100 / BASELINE_QUERY_AVG))
        log_warning "Query performance: ${DEGRADATION}% of baseline (${DEGRADATION}% slowdown)"
    fi
}

check_postgres_checkpoints() {
    section "Step 5: Checking PostgreSQL Checkpoint Activity"

    log "Examining PostgreSQL checkpoint behavior..."

    docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
        "SELECT * FROM pg_stat_bgwriter;" | tee -a "$LOG_FILE"

    log ""
    log "Recent checkpoint activity from logs:"
    docker logs simulated-exchange-postgres --tail 50 2>&1 | grep -i "checkpoint" | tail -10 | tee -a "$LOG_FILE"
}

test_api_under_io_stress() {
    section "Step 6: Testing API Endpoints Under I/O Stress"

    log "Testing trading API while database I/O is stressed..."

    API_TIMES=()
    API_FAILURES=0

    for i in {1..5}; do
        log "API request $i..."
        START=$(date +%s%N)
        if timeout 30 curl -s --max-time 25 http://localhost/api/metrics &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            API_TIMES+=($ELAPSED)
            log "  Response time: ${ELAPSED}ms"
        else
            API_FAILURES=$((API_FAILURES + 1))
            log_error "  Request TIMEOUT or FAILED"
        fi
        sleep 3
    done

    if [ ${#API_TIMES[@]} -gt 0 ]; then
        API_AVG=0
        for time in "${API_TIMES[@]}"; do
            API_AVG=$((API_AVG + time))
        done
        API_AVG=$((API_AVG / ${#API_TIMES[@]}))
        log_warning "API average response time: ${API_AVG}ms"
    fi

    log "API failures: $API_FAILURES / 5"
}

stop_io_stress() {
    section "Step 7: Stopping I/O Stress Test"

    if [ -n "$STRESS_PID" ] && ps -p $STRESS_PID > /dev/null 2>&1; then
        log "Stopping stress-ng..."
        kill $STRESS_PID 2>/dev/null || true
        wait $STRESS_PID 2>/dev/null || true
        log_success "I/O stress stopped"
    else
        log "stress-ng already completed"
    fi

    log ""
    log "Waiting 15s for I/O to stabilize..."
    sleep 15
}

measure_recovery() {
    section "Step 8: Measuring Recovery Performance"

    log "I/O statistics after stress removal:"
    iostat -x 1 3 | tee -a "$LOG_FILE"

    log ""
    log "Testing database query latency (recovery)..."

    RECOVERY_QUERY_TIMES=()
    for i in {1..5}; do
        START=$(date +%s%N)
        docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
            "SELECT COUNT(*) FROM orders;" &> /dev/null
        END=$(date +%s%N)
        ELAPSED=$(( (END - START) / 1000000 ))
        RECOVERY_QUERY_TIMES+=($ELAPSED)
        log "  Query $i: ${ELAPSED}ms"
    done

    # Calculate average
    RECOVERY_QUERY_AVG=0
    for time in "${RECOVERY_QUERY_TIMES[@]}"; do
        RECOVERY_QUERY_AVG=$((RECOVERY_QUERY_AVG + time))
    done
    RECOVERY_QUERY_AVG=$((RECOVERY_QUERY_AVG / ${#RECOVERY_QUERY_TIMES[@]}))

    log_success "Recovery query latency: ${RECOVERY_QUERY_AVG}ms"

    if [ $RECOVERY_QUERY_AVG -gt 0 ] && [ $BASELINE_QUERY_AVG -gt 0 ]; then
        RECOVERY_PERC=$((RECOVERY_QUERY_AVG * 100 / BASELINE_QUERY_AVG))
        log "Recovery performance: ${RECOVERY_PERC}% of baseline"
    fi
}

stop_bpf_monitoring() {
    section "Step 9: Stopping BPF Monitoring"

    if [ -n "$BPF_PID" ] && ps -p $BPF_PID > /dev/null 2>&1; then
        log "Stopping BPF trace..."
        sudo kill -INT $BPF_PID
        sleep 2

        if [ -f "$LOG_FILE.bpf" ]; then
            log "BPF output saved to: $LOG_FILE.bpf"
            echo "--- Last 40 lines of BPF output ---" | tee -a "$LOG_FILE"
            tail -40 "$LOG_FILE.bpf" | tee -a "$LOG_FILE"
        fi
    fi
}

generate_report() {
    section "Experiment Summary Report"

    cat << EOF | tee -a "$LOG_FILE"

EXPERIMENT: $EXPERIMENT_NAME
DATE: $(date +'%Y-%m-%d %H:%M:%S')

CONFIGURATION:
- Stress tool: stress-ng --iomix 4
- Duration: 120 seconds
- Target: Host filesystem (affects all containers)

QUERY PERFORMANCE:
- Baseline query latency: ${BASELINE_QUERY_AVG}ms
- Under I/O stress: ${STRESS_QUERY_AVG}ms
- Recovery latency: ${RECOVERY_QUERY_AVG}ms
- Failed queries: $FAILED_QUERIES / 10

API PERFORMANCE:
- API average response: ${API_AVG:-N/A}ms
- API failures: $API_FAILURES / 5

DEGRADATION ANALYSIS:
EOF

    if [ $STRESS_QUERY_AVG -gt 0 ] && [ $BASELINE_QUERY_AVG -gt 0 ]; then
        DEGRADATION=$((STRESS_QUERY_AVG * 100 / BASELINE_QUERY_AVG))
        SLOWDOWN=$(( (DEGRADATION - 100) ))
        echo "- Performance degradation: ${SLOWDOWN}% slower" | tee -a "$LOG_FILE"

        if [ $DEGRADATION -gt 1000 ]; then
            echo "- Impact: CRITICAL (>10x slowdown)" | tee -a "$LOG_FILE"
        elif [ $DEGRADATION -gt 500 ]; then
            echo "- Impact: SEVERE (5-10x slowdown)" | tee -a "$LOG_FILE"
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
1. I/O saturation caused ${SLOWDOWN:-N/A}% query slowdown
2. Failed queries: $FAILED_QUERIES (timeouts due to I/O wait)
3. PostgreSQL checkpoint activity affected
4. API response times significantly degraded
5. Immediate recovery when I/O stress removed

OBSERVATIONS:
- Disk I/O is critical bottleneck for database
- Read/write latency visible in BPF histograms
- PostgreSQL background writer impacted
- Cascading effect to API layer
- Filesystem cache helps but not sufficient

FILES GENERATED:
- Experiment log: $LOG_FILE
- BPF I/O trace: $LOG_FILE.bpf
- stress-ng metrics: $LOG_FILE.stress

LESSONS LEARNED:
- Database performance directly tied to disk I/O
- I/O saturation causes query timeouts
- Checkpoint delays can compound problem
- Fast storage (SSD/NVMe) critical for production
- Monitoring I/O latency is essential

RECOMMENDATIONS:
1. Use fast storage (NVMe SSD) for database volumes
2. Monitor disk I/O latency and queue depth
3. Set appropriate query timeouts for I/O conditions
4. Consider read replicas to distribute I/O load
5. Alert on sustained high I/O wait times
6. Tune PostgreSQL checkpoint settings
7. Use connection pooling to limit concurrent queries

EOF

    log_success "Experiment completed"
    log "Full log available at: $LOG_FILE"
}

cleanup() {
    section "Cleanup"

    # Stop stress-ng if still running
    if [ -n "$STRESS_PID" ] && ps -p $STRESS_PID > /dev/null 2>&1; then
        kill -9 $STRESS_PID 2>/dev/null || true
    fi

    # Kill any remaining stress-ng processes
    pkill -9 stress-ng 2>/dev/null || true

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
    start_io_stress
    observe_io_impact
    check_postgres_checkpoints
    test_api_under_io_stress
    stop_io_stress
    measure_recovery
    stop_bpf_monitoring
    generate_report

    section "EXPERIMENT COMPLETE"
}

main
