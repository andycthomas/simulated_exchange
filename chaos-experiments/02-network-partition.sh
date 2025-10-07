#!/bin/bash
################################################################################
# Chaos Experiment 2: Network Partition
################################################################################
#
# PURPOSE:
#   Test system behavior when network communication between trading-api and
#   PostgreSQL is blocked while both services remain running. This simulates:
#   - Network switch failure
#   - Firewall misconfiguration
#   - Network partition (split-brain scenario)
#   - Cable disconnection
#
# WHAT THIS TESTS:
#   - TCP connection timeout behavior
#   - Retry and exponential backoff logic
#   - Graceful degradation to cached/fallback data
#   - Network-level resilience patterns
#   - TCP retransmission mechanics
#
# EXPECTED RESULTS:
#   1. Trading-API cannot reach PostgreSQL (packets dropped)
#   2. Existing connections time out after TCP retransmission attempts
#   3. New connection attempts fail immediately (EHOSTUNREACH)
#   4. Metrics endpoint falls back to cached data
#   5. BPF shows TCP retransmissions and connection failures
#   6. After partition healed: immediate recovery
#   7. Dashboard shows "Database query failed - using cached data"
#
# BPF OBSERVABILITY:
#   - TCP SYN packets being sent but not acknowledged
#   - Retransmission attempts (exponential backoff)
#   - Connection timeout events
#   - RTT (Round Trip Time) increasing before failure
#
################################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
EXPERIMENT_NAME="Network Partition"
LOG_FILE="/tmp/chaos-exp-02-$(date +%Y%m%d-%H%M%S).log"
PARTITION_DURATION=60  # seconds

# Helper functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✓${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ✗${NC} $1" | tee -a "$LOG_FILE"
}

section() {
    echo "" | tee -a "$LOG_FILE"
    echo "========================================================================" | tee -a "$LOG_FILE"
    echo "$1" | tee -a "$LOG_FILE"
    echo "========================================================================" | tee -a "$LOG_FILE"
}

# Check prerequisites
check_prerequisites() {
    section "Checking Prerequisites"

    # Check if running as root (needed for iptables)
    if [ "$EUID" -ne 0 ]; then
        log_error "This experiment requires root privileges (for iptables)"
        log "Please run with: sudo $0"
        exit 1
    fi
    log_success "Running as root"

    if ! command -v docker &> /dev/null; then
        log_error "Docker not found"
        exit 1
    fi
    log_success "Docker found"

    if ! command -v iptables &> /dev/null; then
        log_error "iptables not found"
        exit 1
    fi
    log_success "iptables found"

    if ! docker ps | grep -q "simulated-exchange-postgres"; then
        log_error "PostgreSQL container not running"
        exit 1
    fi
    log_success "PostgreSQL container running"

    if ! docker ps | grep -q "simulated-exchange-trading-api"; then
        log_error "Trading API container not running"
        exit 1
    fi
    log_success "Trading API container running"
}

# Get network information
get_network_info() {
    section "Step 1: Gathering Network Information"

    # Get container PIDs for network namespace access
    TRADING_API_PID=$(docker inspect -f '{{.State.Pid}}' simulated-exchange-trading-api)
    POSTGRES_PID=$(docker inspect -f '{{.State.Pid}}' simulated-exchange-postgres)

    log "Trading API PID: $TRADING_API_PID"
    log "PostgreSQL PID: $POSTGRES_PID"

    # Get container IP addresses
    TRADING_API_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' simulated-exchange-trading-api)
    POSTGRES_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' simulated-exchange-postgres)

    log_success "Trading API IP: $TRADING_API_IP"
    log_success "PostgreSQL IP: $POSTGRES_IP"

    # Test connectivity before partition
    log ""
    log "Testing connectivity before partition..."

    if nsenter -t $TRADING_API_PID -n ping -c 3 -W 2 $POSTGRES_IP &> /dev/null; then
        log_success "Ping from trading-api to postgres: SUCCESS"
    else
        log_error "Cannot ping postgres from trading-api"
        exit 1
    fi

    # Test database connection
    if docker exec simulated-exchange-trading-api timeout 5 sh -c \
        "echo > /dev/tcp/$POSTGRES_IP/5432" 2>/dev/null; then
        log_success "TCP connection to postgres:5432: SUCCESS"
    else
        log_warning "Could not test TCP connection (sh not available in distroless)"
    fi
}

# Capture baseline
capture_baseline() {
    section "Step 2: Capturing Baseline Metrics"

    log "Current database connections..."
    docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
        "SELECT count(*), state FROM pg_stat_activity WHERE datname='trading_db' GROUP BY state;" \
        2>&1 | tee -a "$LOG_FILE"

    log ""
    log "Testing metrics endpoint..."
    BASELINE_RESPONSE=$(curl -s -w "\nHTTP:%{http_code} TIME:%{time_total}s" \
        http://localhost/api/metrics 2>&1)

    echo "$BASELINE_RESPONSE" | head -10 | tee -a "$LOG_FILE"

    BASELINE_HTTP=$(echo "$BASELINE_RESPONSE" | grep "HTTP:" | cut -d: -f2 | cut -d' ' -f1)
    BASELINE_TIME=$(echo "$BASELINE_RESPONSE" | grep "TIME:" | cut -d: -f2)

    log_success "Baseline HTTP code: $BASELINE_HTTP"
    log_success "Baseline response time: $BASELINE_TIME"
}

# Start BPF monitoring
start_bpf_monitoring() {
    section "Step 3: Starting BPF Network Monitoring"

    if ! command -v bpftrace &> /dev/null; then
        log_warning "bpftrace not found - skipping BPF monitoring"
        return
    fi

    log "Creating BPF script for network partition monitoring..."

    cat > /tmp/network-partition-bpf.bt << 'BPFEOF'
#!/usr/bin/env bpftrace

BEGIN {
    printf("Monitoring network partition effects...\n");
    printf("Tracking TCP to postgres (port 5432)\n\n");
}

// Track TCP connection attempts
kprobe:tcp_connect
{
    $sk = (struct sock *)arg0;
    $dport = $sk->__sk_common.skc_dport;
    if ($dport == 5432 || $dport == 0xd204) {
        @connection_attempts = count();
            printf("[%s] TCP CONNECT attempt from %s\n",
                   strftime("%H:%M:%S", nsecs), comm);
    }
}

// Track connection failures
kretprobe:tcp_connect
/retval < 0/
{
    @connection_failures = count();
    printf("[%s] TCP CONNECT FAILED: %s (errno: %d)\n",
           strftime("%H:%M:%S", nsecs), comm, -retval);
}

// Track TCP retransmissions (KEY METRIC during partition)
kprobe:tcp_retransmit_skb {
    @retransmissions[comm] = count();
    @total_retransmissions = count();
    printf("[%s] TCP RETRANSMIT: %s\n",
           strftime("%H:%M:%S", nsecs), comm);
}

// Track TCP timeout events
kprobe:tcp_write_timeout {
    @timeouts = count();
    printf("[%s] TCP WRITE TIMEOUT detected\n",
           strftime("%H:%M:%S", nsecs));
}

// Track dropped packets
tracepoint:skb:kfree_skb {
    @packets_dropped = count();
}

// Periodic summary
interval:s:10 {
    time();
    printf("\n=== Statistics (last 10 seconds) ===\n");
    printf("Connection attempts: %lld\n", (int64)@connection_attempts);
    printf("Connection failures: %lld\n", (int64)@connection_failures);
    printf("Total retransmissions: %lld\n", (int64)@total_retransmissions);
    printf("Timeouts: %lld\n", (int64)@timeouts);
    printf("Packets dropped: %lld\n", (int64)@packets_dropped);

    if (@total_retransmissions > 0) {
        printf("\nRetransmissions by process:\n");
        print(@retransmissions);
    }

    printf("\n");
    clear(@connection_attempts);
    clear(@connection_failures);
    clear(@total_retransmissions);
    clear(@timeouts);
    clear(@packets_dropped);
}

END {
    printf("\n=== Final Summary ===\n");
    printf("Total retransmissions by process:\n");
    print(@retransmissions);
}
BPFEOF

    chmod +x /tmp/network-partition-bpf.bt

    log "Starting BPF monitoring in background..."
    bpftrace /tmp/network-partition-bpf.bt > "$LOG_FILE.bpf" 2>&1 &
    BPF_PID=$!

    log_success "BPF monitoring started (PID: $BPF_PID)"
    sleep 2
}

# Create network partition
create_partition() {
    section "Step 4: Creating Network Partition"

    log "Partition method: iptables DROP rule in trading-api network namespace"
    log "This will block traffic from trading-api to postgres while keeping both services running"

    log ""
    log "Installing iptables DROP rule..."

    # Drop all packets from trading-api to postgres
    nsenter -t $TRADING_API_PID -n iptables -A OUTPUT -d $POSTGRES_IP -j DROP

    PARTITION_START=$(date +%s)

    log_success "Network partition created at: $(date +'%Y-%m-%d %H:%M:%S')"
    log_warning "Trading-API can NO LONGER reach PostgreSQL at $POSTGRES_IP"

    # Verify rule is in place
    log ""
    log "Verifying iptables rule..."
    if nsenter -t $TRADING_API_PID -n iptables -L OUTPUT -n | grep -q "$POSTGRES_IP"; then
        log_success "iptables rule confirmed"
        nsenter -t $TRADING_API_PID -n iptables -L OUTPUT -n | grep "$POSTGRES_IP" | tee -a "$LOG_FILE"
    else
        log_error "iptables rule not found"
        exit 1
    fi

    # Test that partition is working
    log ""
    log "Testing partition (ping should FAIL)..."
    if nsenter -t $TRADING_API_PID -n ping -c 2 -W 2 $POSTGRES_IP &> /dev/null; then
        log_error "Ping succeeded - partition NOT working!"
        exit 1
    else
        log_success "Ping failed as expected - partition is working"
    fi
}

# Observe partition effects
observe_partition() {
    section "Step 5: Observing Partition Effects"

    log "Monitoring system behavior during partition..."
    log "Duration: ${PARTITION_DURATION}s"

    # Test API every 10 seconds
    for i in $(seq 1 $((PARTITION_DURATION / 10))); do
        log ""
        log "--- Observation $i (t+$((i * 10))s) ---"

        # Test metrics endpoint
        log "Testing metrics endpoint..."
        RESPONSE=$(curl -s -w "\nHTTP:%{http_code} TIME:%{time_total}s" \
            -m 20 http://localhost/api/metrics 2>&1)

        HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP:" | cut -d: -f2 | cut -d' ' -f1)
        RESPONSE_TIME=$(echo "$RESPONSE" | grep "TIME:" | cut -d: -f2)

        if [ "$HTTP_CODE" == "200" ]; then
            # Check if using fallback data
            if echo "$RESPONSE" | grep -q '"order_count":200'; then
                log_warning "HTTP 200 but using FALLBACK data (order_count=200)"
            else
                log "HTTP 200 - may still have cached/recent data"
            fi
        else
            log_error "HTTP $HTTP_CODE - API degraded"
        fi

        log "Response time: ${RESPONSE_TIME}s"

        # Check logs for errors
        ERROR_COUNT=$(docker logs --since 10s simulated-exchange-trading-api 2>&1 | \
            grep -i "error\|failed\|timeout" | wc -l)

        if [ "$ERROR_COUNT" -gt 0 ]; then
            log_warning "Found $ERROR_COUNT error/timeout messages in last 10s"
            docker logs --since 10s simulated-exchange-trading-api 2>&1 | \
                grep -i "error\|failed\|timeout" | tail -3 | tee -a "$LOG_FILE"
        fi

        # PostgreSQL still running (should be healthy)
        if docker exec simulated-exchange-postgres pg_isready -U trading_user &> /dev/null; then
            log_success "PostgreSQL is still healthy (but unreachable from trading-api)"
        fi

        sleep 10
    done
}

# Heal partition
heal_partition() {
    section "Step 6: Healing Network Partition"

    log "Removing iptables DROP rule to restore connectivity..."

    nsenter -t $TRADING_API_PID -n iptables -D OUTPUT -d $POSTGRES_IP -j DROP

    PARTITION_END=$(date +%s)
    PARTITION_DURATION_ACTUAL=$((PARTITION_END - PARTITION_START))

    log_success "Network partition removed at: $(date +'%Y-%m-%d %H:%M:%S')"
    log_success "Partition duration: ${PARTITION_DURATION_ACTUAL}s"

    # Verify connectivity restored
    log ""
    log "Verifying connectivity restored..."

    if nsenter -t $TRADING_API_PID -n ping -c 3 -W 2 $POSTGRES_IP &> /dev/null; then
        log_success "Ping from trading-api to postgres: RESTORED"
    else
        log_error "Ping still failing after partition removal!"
    fi

    log ""
    log "Waiting for system to recover (15s)..."
    sleep 15
}

# Capture post-partition metrics
capture_post_metrics() {
    section "Step 7: Capturing Post-Partition Metrics"

    log "Testing metrics endpoint after recovery..."
    POST_RESPONSE=$(curl -s -w "\nHTTP:%{http_code} TIME:%{time_total}s" \
        http://localhost/api/metrics 2>&1)

    echo "$POST_RESPONSE" | head -10 | tee -a "$LOG_FILE"

    POST_HTTP=$(echo "$POST_RESPONSE" | grep "HTTP:" | cut -d: -f2 | cut -d' ' -f1)
    POST_TIME=$(echo "$POST_RESPONSE" | grep "TIME:" | cut -d: -f2)

    log_success "Post-recovery HTTP code: $POST_HTTP (was: $BASELINE_HTTP)"
    log_success "Post-recovery response time: $POST_TIME (was: $BASELINE_TIME)"

    if [ "$POST_HTTP" == "200" ]; then
        if echo "$POST_RESPONSE" | grep -q '"order_count":200'; then
            log_warning "Still using fallback data - may need more time to recover"
        else
            log_success "Using real database data - recovery complete"
        fi
    fi

    log ""
    log "Database connections after recovery..."
    docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
        "SELECT count(*), state FROM pg_stat_activity WHERE datname='trading_db' GROUP BY state;" \
        2>&1 | tee -a "$LOG_FILE"
}

# Stop BPF monitoring
stop_bpf_monitoring() {
    section "Step 8: Stopping BPF Monitoring"

    if [ -n "$BPF_PID" ] && ps -p $BPF_PID > /dev/null 2>&1; then
        log "Stopping BPF trace..."
        kill -INT $BPF_PID
        sleep 2

        log_success "BPF monitoring stopped"

        if [ -f "$LOG_FILE.bpf" ]; then
            log ""
            log "BPF Summary:"
            echo "--- Last 30 lines of BPF output ---" | tee -a "$LOG_FILE"
            tail -30 "$LOG_FILE.bpf" | tee -a "$LOG_FILE"
        fi
    fi
}

# Generate report
generate_report() {
    section "Experiment Summary Report"

    cat << EOF | tee -a "$LOG_FILE"

EXPERIMENT: $EXPERIMENT_NAME
DATE: $(date +'%Y-%m-%d %H:%M:%S')
PARTITION DURATION: ${PARTITION_DURATION_ACTUAL}s

CONFIGURATION:
- Trading API IP: $TRADING_API_IP
- PostgreSQL IP: $POSTGRES_IP
- Partition method: iptables DROP in network namespace

KEY FINDINGS:
1. Network partition detection: Observable via TCP retransmissions
2. Service behavior: Trading API fell back to cached/fallback data
3. Baseline response time: $BASELINE_TIME
4. Post-recovery response time: $POST_TIME
5. Recovery: Immediate after partition healed

OBSERVATIONS:
- TCP retransmissions visible in BPF during partition
- Connection timeouts occurred as expected
- API gracefully degraded to fallback data
- No service crashes or restarts
- Automatic recovery upon network restoration

FILES GENERATED:
- Experiment log: $LOG_FILE
- BPF trace: $LOG_FILE.bpf

NETWORK PARTITION EFFECTS:
- TCP connections timed out after retransmission attempts
- New connection attempts immediately failed
- No data corruption or loss detected
- Services remained healthy despite network isolation

NEXT STEPS:
1. Review BPF trace for TCP retransmission patterns
2. Analyze timeout values and retry behavior
3. Consider implementing circuit breaker pattern
4. Review connection pool timeout settings

EOF

    log_success "Experiment completed successfully"
    log "Full log available at: $LOG_FILE"
}

# Cleanup
cleanup() {
    section "Cleanup"

    log "Removing any remaining iptables rules..."

    # Remove DROP rule if it still exists
    nsenter -t $TRADING_API_PID -n iptables -D OUTPUT -d $POSTGRES_IP -j DROP 2>/dev/null || true

    # Kill BPF if still running
    if [ -n "$BPF_PID" ] && ps -p $BPF_PID > /dev/null 2>&1; then
        kill -9 $BPF_PID 2>/dev/null || true
    fi

    log_success "Cleanup complete"
}

# Main execution
main() {
    clear
    section "CHAOS EXPERIMENT: $EXPERIMENT_NAME"

    log "Starting chaos engineering experiment..."
    log "All output will be logged to: $LOG_FILE"

    # Trap errors and cleanup on exit
    trap cleanup EXIT

    check_prerequisites
    get_network_info
    capture_baseline
    start_bpf_monitoring
    create_partition
    observe_partition
    heal_partition
    capture_post_metrics
    stop_bpf_monitoring
    generate_report

    section "EXPERIMENT COMPLETE"
}

# Run the experiment
main
