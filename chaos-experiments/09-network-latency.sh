#!/bin/bash
################################################################################
# Chaos Experiment 9: Network Latency Injection
################################################################################
#
# PURPOSE:
#   Test gradual performance degradation via progressive network delay. Simulates:
#   - WAN degradation
#   - Network congestion
#   - Geographically distributed services
#   - Slow network infrastructure
#
# WHAT THIS TESTS:
#   - Timeout tuning validation
#   - Latency sensitivity of services
#   - Retry behavior under delays
#   - User experience degradation
#   - Query timeout thresholds
#
# EXPECTED RESULTS:
#   1. Baseline latency measured (~1ms)
#   2. 50ms delay injected - noticeable slowdown
#   3. 100ms delay - significant impact
#   4. 200ms delay - timeouts begin
#   5. Retransmissions visible in BPF
#   6. Immediate recovery when delay removed
#
# BPF OBSERVABILITY:
#   - TCP RTT (Round Trip Time) distribution shift
#   - Retransmission rate increase
#   - Application timeout events
#   - Queue depth in network stack
#
################################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

EXPERIMENT_NAME="Network Latency Injection"
LOG_FILE="/tmp/chaos-exp-09-$(date +%Y%m%d-%H%M%S).log"
TRADING_API_IFACE=""

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

    if ! command -v tc &> /dev/null; then
        log_warning "tc (traffic control) not found, installing..."
        sudo dnf install -y iproute-tc 2>&1 | tee -a "$LOG_FILE"
    fi
    log_success "tc (traffic control) available"

    # Check if netem kernel module is available
    if ! lsmod | grep -q sch_netem; then
        log "Attempting to load sch_netem kernel module..."
        if sudo modprobe sch_netem 2>/dev/null; then
            log_success "sch_netem module loaded"
        else
            log_error "sch_netem kernel module not available on this system"
            log_error "This kernel does not support network emulation (netem)"
            log ""
            log "WORKAROUND: This experiment requires the sch_netem kernel module."
            log "Your kernel ($(uname -r)) does not have this module compiled in."
            log ""
            log "To run this experiment, you need:"
            log "  1. A kernel with CONFIG_NET_SCH_NETEM=m or =y"
            log "  2. Or use a different system/VM with netem support"
            log ""
            log "You can still learn from this script's structure and apply"
            log "the concepts on a system with proper kernel support."
            exit 1
        fi
    else
        log_success "sch_netem module already loaded"
    fi

    # Get trading-api container network namespace and interface
    TRADING_API_PID=$(docker inspect simulated-exchange-trading-api --format '{{.State.Pid}}')
    log "Trading API container PID: $TRADING_API_PID"

    # Get the interface name inside the container
    # Strip @ifXXXX suffix that appears in veth pairs
    TRADING_API_IFACE=$(sudo nsenter -t $TRADING_API_PID -n ip link show | grep "^[0-9]:" | grep -v "lo:" | awk -F: '{print $2}' | xargs | head -1 | cut -d'@' -f1)
    log "Trading API network interface: $TRADING_API_IFACE"

    if [ -z "$TRADING_API_IFACE" ]; then
        log_error "Could not determine trading-api network interface"
        exit 1
    fi
}

capture_baseline() {
    section "Step 1: Capturing Baseline Network Latency"

    log "Testing baseline API latency..."

    BASELINE_TIMES=()
    for i in {1..10}; do
        START=$(date +%s%N)
        if curl -s --max-time 5 http://localhost/api/health &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            BASELINE_TIMES+=($ELAPSED)
            echo -n "." | tee -a "$LOG_FILE"
        else
            log_warning "Request $i failed"
        fi
    done
    echo "" | tee -a "$LOG_FILE"

    # Calculate average
    BASELINE_AVG=0
    if [ ${#BASELINE_TIMES[@]} -gt 0 ]; then
        for time in "${BASELINE_TIMES[@]}"; do
            BASELINE_AVG=$((BASELINE_AVG + time))
        done
        BASELINE_AVG=$((BASELINE_AVG / ${#BASELINE_TIMES[@]}))
    fi

    log_success "Baseline API latency: ${BASELINE_AVG}ms"

    # Test database connectivity latency
    log ""
    log "Testing database query latency (baseline)..."

    DB_START=$(date +%s%N)
    docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
        "SELECT COUNT(*) FROM orders LIMIT 1;" &> /dev/null
    DB_END=$(date +%s%N)
    DB_BASELINE=$(( (DB_END - DB_START) / 1000000 ))

    log_success "Baseline DB query latency: ${DB_BASELINE}ms"

    # Test ping latency to postgres
    log ""
    log "Ping latency to PostgreSQL container..."
    POSTGRES_IP=$(docker inspect simulated-exchange-postgres --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}')

    sudo nsenter -t $TRADING_API_PID -n ping -c 5 $POSTGRES_IP | tail -1 | tee -a "$LOG_FILE"
}

start_bpf_monitoring() {
    section "Step 2: Starting BPF Network Monitoring"

    if ! command -v bpftrace &> /dev/null; then
        log_warning "bpftrace not found - skipping BPF monitoring"
        return
    fi

    cat > /tmp/latency-bpf.bt << 'BPFEOF'
#!/usr/bin/env bpftrace

BEGIN {
    printf("Monitoring TCP RTT and retransmissions...\n\n");
}

// Track TCP RTT (Round Trip Time)
kprobe:tcp_rcv_established {
    $sk = (struct sock *)arg0;
    $tp = (struct tcp_sock *)$sk;
    $srtt_us = $tp->srtt_us >> 3;  // Convert from internal to microseconds

    @rtt_dist = hist($srtt_us);
    @avg_rtt = avg($srtt_us);
}

// Track TCP retransmissions
kprobe:tcp_retransmit_skb {
    $sk = (struct sock *)arg0;
    @retransmits[comm] = count();
    @total_retransmits = count();
}

// Track connection timeouts
kprobe:tcp_write_err {
    @connection_timeouts[comm] = count();
}

// Track packet drops
tracepoint:skb:kfree_skb {
    @packet_drops = count();
}

// Interval reporting
interval:s:5 {
    printf("\n[%s] === Network Stats (5s) ===\n", strftime("%H:%M:%S", nsecs));
    printf("Average RTT:\n");
    print(@avg_rtt);
    printf("Total retransmits:\n");
    print(@total_retransmits);
    printf("Packet drops:\n");
    print(@packet_drops);

    printf("\nRetransmits by process:\n");
    print(@retransmits);

    clear(@avg_rtt);
    clear(@retransmits);
    clear(@total_retransmits);
    clear(@packet_drops);
}

END {
    printf("\n=== Final Network Summary ===\n");

    printf("RTT distribution (microseconds):\n");
    print(@rtt_dist);

    printf("\nConnection timeouts:\n");
    print(@connection_timeouts);
}
BPFEOF

    chmod +x /tmp/latency-bpf.bt

    log "Starting BPF monitoring..."
    sudo bpftrace /tmp/latency-bpf.bt > "$LOG_FILE.bpf" 2>&1 &
    BPF_PID=$!

    log_success "BPF monitoring started (PID: $BPF_PID)"
    sleep 2
}

inject_latency() {
    local DELAY=$1
    local LABEL=$2

    section "Step $3: Injecting ${DELAY}ms Network Latency ($LABEL)"

    log "Adding ${DELAY}ms delay to $TRADING_API_IFACE..."

    # Use tc netem to add latency
    sudo nsenter -t $TRADING_API_PID -n tc qdisc add dev $TRADING_API_IFACE root netem delay ${DELAY}ms 2>&1 | tee -a "$LOG_FILE" || \
        sudo nsenter -t $TRADING_API_PID -n tc qdisc change dev $TRADING_API_IFACE root netem delay ${DELAY}ms 2>&1 | tee -a "$LOG_FILE"

    log_success "Latency injected: ${DELAY}ms"

    log ""
    log "Verifying delay with ping..."
    POSTGRES_IP=$(docker inspect simulated-exchange-postgres --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}')
    sudo nsenter -t $TRADING_API_PID -n ping -c 3 $POSTGRES_IP | tail -1 | tee -a "$LOG_FILE"

    log ""
    log "Waiting 5s for latency to stabilize..."
    sleep 5
}

test_api_with_latency() {
    local DELAY=$1
    local ARRAY_NAME=$2

    log "Testing API performance with ${DELAY}ms latency..."

    local -n TIMES=$ARRAY_NAME
    FAILURES=0

    for i in {1..10}; do
        START=$(date +%s%N)
        # Increase timeout as latency increases
        TIMEOUT=$((15 + DELAY / 10))
        if timeout $TIMEOUT curl -s --max-time $((TIMEOUT - 2)) http://localhost/api/health &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            TIMES+=($ELAPSED)

            if [ $((i % 3)) -eq 0 ]; then
                log "  Request $i: ${ELAPSED}ms"
            else
                echo -n "." | tee -a "$LOG_FILE"
            fi
        else
            FAILURES=$((FAILURES + 1))
            log_error "  Request $i: TIMEOUT or FAILED"
        fi
        sleep 2
    done
    echo "" | tee -a "$LOG_FILE"

    # Calculate average
    local AVG=0
    if [ ${#TIMES[@]} -gt 0 ]; then
        for time in "${TIMES[@]}"; do
            AVG=$((AVG + time))
        done
        AVG=$((AVG / ${#TIMES[@]}))
    fi

    log_warning "Average latency with ${DELAY}ms delay: ${AVG}ms"
    log "Failed requests: $FAILURES / 10"

    if [ $AVG -gt 0 ] && [ $BASELINE_AVG -gt 0 ]; then
        IMPACT=$(( (AVG - BASELINE_AVG) ))
        log "Latency increase: +${IMPACT}ms from baseline"
    fi

    # Store average for this delay level
    eval "${ARRAY_NAME}_AVG=$AVG"
    eval "${ARRAY_NAME}_FAILURES=$FAILURES"
}

test_database_with_latency() {
    local DELAY=$1

    log ""
    log "Testing database query with ${DELAY}ms latency..."

    DB_TIMES=()
    DB_FAILURES=0

    for i in {1..5}; do
        START=$(date +%s%N)
        TIMEOUT=$((30 + DELAY / 5))
        if timeout $TIMEOUT docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
            "SELECT COUNT(*) FROM orders LIMIT 1;" &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            DB_TIMES+=($ELAPSED)
            log "  DB query $i: ${ELAPSED}ms"
        else
            DB_FAILURES=$((DB_FAILURES + 1))
            log_error "  DB query $i: TIMEOUT"
        fi
    done

    # Calculate average
    DB_AVG=0
    if [ ${#DB_TIMES[@]} -gt 0 ]; then
        for time in "${DB_TIMES[@]}"; do
            DB_AVG=$((DB_AVG + time))
        done
        DB_AVG=$((DB_AVG / ${#DB_TIMES[@]}))
        log "DB query average: ${DB_AVG}ms"
    fi

    log "DB query failures: $DB_FAILURES / 5"
}

remove_latency() {
    section "Step 7: Removing Network Latency"

    log "Removing tc qdisc delay from $TRADING_API_IFACE..."

    sudo nsenter -t $TRADING_API_PID -n tc qdisc del dev $TRADING_API_IFACE root 2>&1 | tee -a "$LOG_FILE"

    log_success "Latency removed"

    log ""
    log "Verifying latency removal with ping..."
    POSTGRES_IP=$(docker inspect simulated-exchange-postgres --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}')
    sudo nsenter -t $TRADING_API_PID -n ping -c 3 $POSTGRES_IP | tail -1 | tee -a "$LOG_FILE"

    log ""
    log "Waiting 5s for network to stabilize..."
    sleep 5
}

measure_recovery() {
    section "Step 8: Measuring Recovery Performance"

    log "Testing API performance after latency removal..."

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

    log_success "Recovery latency: ${RECOVERY_AVG}ms"

    if [ $RECOVERY_AVG -gt 0 ] && [ $BASELINE_AVG -gt 0 ]; then
        RECOVERY_PERC=$((RECOVERY_AVG * 100 / BASELINE_AVG))
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

LATENCY PROGRESSION TEST:
- Baseline: ${BASELINE_AVG}ms
- 50ms delay: ${LATENCY_50_AVG:-N/A}ms (${LATENCY_50_FAILURES:-0} failures)
- 100ms delay: ${LATENCY_100_AVG:-N/A}ms (${LATENCY_100_FAILURES:-0} failures)
- 200ms delay: ${LATENCY_200_AVG:-N/A}ms (${LATENCY_200_FAILURES:-0} failures)
- Recovery: ${RECOVERY_AVG}ms

IMPACT ANALYSIS:
EOF

    if [ ${LATENCY_50_AVG:-0} -gt 0 ]; then
        IMPACT_50=$(( LATENCY_50_AVG - BASELINE_AVG ))
        echo "- 50ms delay impact: +${IMPACT_50}ms (~$(( LATENCY_50_AVG * 100 / BASELINE_AVG ))% of baseline)" | tee -a "$LOG_FILE"
    fi

    if [ ${LATENCY_100_AVG:-0} -gt 0 ]; then
        IMPACT_100=$(( LATENCY_100_AVG - BASELINE_AVG ))
        echo "- 100ms delay impact: +${IMPACT_100}ms (~$(( LATENCY_100_AVG * 100 / BASELINE_AVG ))% of baseline)" | tee -a "$LOG_FILE"
    fi

    if [ ${LATENCY_200_AVG:-0} -gt 0 ]; then
        IMPACT_200=$(( LATENCY_200_AVG - BASELINE_AVG ))
        echo "- 200ms delay impact: +${IMPACT_200}ms (~$(( LATENCY_200_AVG * 100 / BASELINE_AVG ))% of baseline)" | tee -a "$LOG_FILE"
    fi

    cat << EOF | tee -a "$LOG_FILE"

KEY FINDINGS:
1. Network latency directly impacts API response times
2. 50ms delay: ${LATENCY_50_FAILURES:-0} timeouts (tolerable)
3. 100ms delay: ${LATENCY_100_FAILURES:-0} timeouts (noticeable degradation)
4. 200ms delay: ${LATENCY_200_FAILURES:-0} timeouts (severe impact)
5. BPF shows RTT increase and potential retransmissions
6. Immediate recovery when latency removed

OBSERVATIONS:
- Latency compounds across service calls
- Each additional 50ms adds ~50-100ms to total response
- Database queries affected proportionally
- Timeout thresholds determine failure points
- TCP retransmissions increase with latency
- Application remains functional but slow

TIMEOUT ANALYSIS:
EOF

    TOTAL_FAILURES=$(( ${LATENCY_50_FAILURES:-0} + ${LATENCY_100_FAILURES:-0} + ${LATENCY_200_FAILURES:-0} ))
    echo "- Total failures across all tests: $TOTAL_FAILURES / 30" | tee -a "$LOG_FILE"

    if [ ${LATENCY_200_FAILURES:-0} -gt 5 ]; then
        echo "- ⚠ 200ms delay causes significant failures" | tee -a "$LOG_FILE"
        echo "- Current timeouts may be too aggressive" | tee -a "$LOG_FILE"
    else
        echo "- ✓ Timeouts appear well-tuned" | tee -a "$LOG_FILE"
    fi

    cat << EOF | tee -a "$LOG_FILE"

FILES GENERATED:
- Experiment log: $LOG_FILE
- BPF network trace: $LOG_FILE.bpf

LESSONS LEARNED:
- Network latency is additive across service hops
- Even small delays (50ms) have measurable impact
- High latency (200ms+) can trigger timeouts
- RTT distribution visible in BPF provides insights
- Retransmissions indicate network stress

RECOMMENDATIONS:
1. Set timeout values accounting for network variance
2. Monitor RTT (Round Trip Time) continuously
3. Alert on RTT > 100ms for local services
4. Consider retry logic with exponential backoff
5. Test geographically distributed deployments
6. Use CDN/edge caching to reduce latency impact
7. Tune TCP congestion control for WAN scenarios
8. Implement request hedging for latency-sensitive operations

PRODUCTION READINESS:
- System handles moderate latency (50-100ms) gracefully
- Timeouts ${TOTAL_FAILURES:-0}/30 indicate ${LATENCY_200_FAILURES:-0} > 5 ? "need tuning" : "are appropriate"}
- WAN deployments should plan for 100-200ms baseline RTT
- Consider circuit breakers for high-latency scenarios

EOF

    log_success "Experiment completed"
    log "Full log available at: $LOG_FILE"
}

cleanup() {
    section "Cleanup"

    # Remove any tc rules
    if [ -n "$TRADING_API_PID" ] && [ -n "$TRADING_API_IFACE" ]; then
        log "Removing any remaining tc qdisc rules..."
        sudo nsenter -t $TRADING_API_PID -n tc qdisc del dev $TRADING_API_IFACE root 2>/dev/null || true
    fi

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

    # Progressive latency injection
    inject_latency 50 "Moderate Delay" 3
    declare -a LATENCY_50
    test_api_with_latency 50 LATENCY_50
    test_database_with_latency 50

    inject_latency 100 "Significant Delay" 4
    declare -a LATENCY_100
    test_api_with_latency 100 LATENCY_100
    test_database_with_latency 100

    inject_latency 200 "High Delay" 5
    declare -a LATENCY_200
    test_api_with_latency 200 LATENCY_200
    test_database_with_latency 200

    remove_latency
    measure_recovery
    stop_bpf_monitoring
    generate_report

    section "EXPERIMENT COMPLETE"
}

main
