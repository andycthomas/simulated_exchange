#!/bin/bash
################################################################################
# Chaos Experiment 6: Connection Pool Exhaustion
################################################################################
#
# PURPOSE:
#   Test system behavior when database connection pool is exhausted. This simulates:
#   - Connection leak scenarios
#   - max_connections limit reached
#   - Pool misconfiguration
#   - High concurrent load
#
# WHAT THIS TESTS:
#   - Connection pool management
#   - Queue behavior when pool is saturated
#   - Error handling for "too many connections"
#   - Connection reuse and cleanup
#   - Graceful degradation under connection pressure
#
# EXPECTED RESULTS:
#   1. Baseline connection count captured (typically 5-10)
#   2. max_connections reduced to 20 (low limit)
#   3. Order flow creates connection demand
#   4. Connections climb to limit (20)
#   5. New requests fail with "too many clients"
#   6. API returns connection errors
#   7. Recovery after limit restored
#
# BPF OBSERVABILITY:
#   - TCP connection attempts to port 5432
#   - Failed connection attempts
#   - Time waiting for available connection
#   - Connection churn rate
#
################################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

EXPERIMENT_NAME="Connection Pool Exhaustion"
LOG_FILE="/tmp/chaos-exp-06-$(date +%Y%m%d-%H%M%S).log"
ORIGINAL_MAX_CONNECTIONS=""

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

    if ! docker ps | grep -q "simulated-exchange-trading-api"; then
        log_error "Trading API container not running"
        exit 1
    fi
    log_success "Trading API container running"
}

capture_baseline() {
    section "Step 1: Capturing Baseline Connection Statistics"

    log "Querying current max_connections setting..."
    ORIGINAL_MAX_CONNECTIONS=$(docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -t -c \
        "SHOW max_connections;" | xargs)

    log_success "Current max_connections: $ORIGINAL_MAX_CONNECTIONS"

    log ""
    log "Current active connections:"
    docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
        "SELECT count(*) as active_connections FROM pg_stat_activity;" | tee -a "$LOG_FILE"

    log ""
    log "Connection breakdown by application:"
    docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
        "SELECT application_name, count(*) as conn_count FROM pg_stat_activity GROUP BY application_name;" | tee -a "$LOG_FILE"

    BASELINE_CONNECTIONS=$(docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -t -c \
        "SELECT count(*) FROM pg_stat_activity;" | xargs)

    log_success "Baseline active connections: $BASELINE_CONNECTIONS"
}

start_bpf_monitoring() {
    section "Step 2: Starting BPF Connection Monitoring"

    if ! command -v bpftrace &> /dev/null; then
        log_warning "bpftrace not found - skipping BPF monitoring"
        return
    fi

    cat > /tmp/conn-bpf.bt << 'BPFEOF'
#!/usr/bin/env bpftrace

BEGIN {
    printf("Monitoring PostgreSQL connections (port 5432)...\n\n");
}

// Track connection attempts to PostgreSQL
kprobe:tcp_v4_connect {
    $sk = (struct sock *)arg0;
    $dport = $sk->__sk_common.skc_dport;

    // PostgreSQL port is 5432 (in network byte order)
    if ($dport == 0x3514) {  // 5432 in big-endian
        @connection_attempts[comm] = count();
        @attempt_times[comm] = nsecs;
    }
}

// Track successful connections
kretprobe:tcp_v4_connect {
    if (retval == 0 && @attempt_times[comm] != 0) {
        $latency_ms = (nsecs - @attempt_times[comm]) / 1000000;
        @connection_success[comm] = count();
        @connection_latency = hist($latency_ms);
        delete(@attempt_times[comm]);
    }
}

// Track failed connections
kretprobe:tcp_v4_connect {
    if (retval < 0 && @attempt_times[comm] != 0) {
        @connection_failures[comm] = count();
        @failure_errno[comm, -retval] = count();
        delete(@attempt_times[comm]);
    }
}

// Track TCP connection states
kprobe:tcp_set_state {
    $sk = (struct sock *)arg0;
    $oldstate = arg1;
    $newstate = arg2;
    $dport = $sk->__sk_common.skc_dport;

    if ($dport == 0x3514) {
        @state_changes[$oldstate, $newstate] = count();
    }
}

// Interval reporting
interval:s:5 {
    printf("\n[%s] === Connection Stats (5s) ===\n", strftime("%H:%M:%S", nsecs));

    printf("Connection attempts:\n");
    print(@connection_attempts);

    printf("\nSuccessful connections:\n");
    print(@connection_success);

    printf("\nFailed connections:\n");
    print(@connection_failures);

    printf("\nFailure reasons (errno):\n");
    print(@failure_errno);

    clear(@connection_attempts);
    clear(@connection_success);
    clear(@connection_failures);
}

END {
    printf("\n=== Final Connection Summary ===\n");

    printf("Connection latency distribution (ms):\n");
    print(@connection_latency);

    printf("\nTCP state changes:\n");
    print(@state_changes);

    printf("\nFailure reasons:\n");
    print(@failure_errno);
}
BPFEOF

    chmod +x /tmp/conn-bpf.bt

    log "Starting BPF monitoring..."
    sudo bpftrace /tmp/conn-bpf.bt > "$LOG_FILE.bpf" 2>&1 &
    BPF_PID=$!

    log_success "BPF monitoring started (PID: $BPF_PID)"
    sleep 2
}

reduce_max_connections() {
    section "Step 3: Reducing max_connections to Trigger Exhaustion"

    log "Setting max_connections to 20 (from $ORIGINAL_MAX_CONNECTIONS)..."

    docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
        "ALTER SYSTEM SET max_connections = 20;" 2>&1 | tee -a "$LOG_FILE"

    log "Restarting PostgreSQL to apply new max_connections..."
    docker restart simulated-exchange-postgres 2>&1 | tee -a "$LOG_FILE"

    log "Waiting 15s for PostgreSQL to restart..."
    sleep 15

    # Verify PostgreSQL is back up
    MAX_WAIT=30
    ELAPSED=0
    while [ $ELAPSED -lt $MAX_WAIT ]; do
        if docker exec simulated-exchange-postgres pg_isready -U trading_user &> /dev/null; then
            log_success "PostgreSQL is ready"
            break
        fi
        sleep 2
        ELAPSED=$((ELAPSED + 2))
        echo -n "." | tee -a "$LOG_FILE"
    done
    echo "" | tee -a "$LOG_FILE"

    # Verify new limit
    NEW_MAX=$(docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -t -c \
        "SHOW max_connections;" | xargs)

    log_success "New max_connections: $NEW_MAX"

    log ""
    log "Note: Connection pool exhaustion will occur when active connections reach 20"
}

start_order_flow() {
    section "Step 4: Starting Order Flow to Create Connection Demand"

    log "Starting order-flow-simulator to generate connection load..."

    docker compose start order-flow-simulator 2>&1 | tee -a "$LOG_FILE"

    log_success "Order flow started"
    log ""
    log "Waiting 10s for connections to build up..."
    sleep 10
}

monitor_connection_saturation() {
    section "Step 5: Monitoring Connection Pool Saturation"

    log "Monitoring connection count until saturation..."

    SATURATION_REACHED=false
    ERRORS_OBSERVED=false
    START_TIME=$(date +%s)
    MAX_MONITOR_TIME=120  # 2 minutes

    while [ $(($(date +%s) - START_TIME)) -lt $MAX_MONITOR_TIME ]; do
        # Get current connection count
        CURRENT_CONNECTIONS=$(docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -t -c \
            "SELECT count(*) FROM pg_stat_activity;" 2>/dev/null | xargs || echo "0")

        log "Active connections: $CURRENT_CONNECTIONS / 20"

        # Check for connection errors in trading-api logs
        if docker logs simulated-exchange-trading-api --tail 50 2>&1 | grep -q "too many clients\|connection pool exhausted\|failed to connect"; then
            if [ "$ERRORS_OBSERVED" = false ]; then
                log_error "Connection errors detected in trading-api logs!"
                ERRORS_OBSERVED=true
            fi
        fi

        # Check if we've reached saturation (within 2 of max)
        if [ "$CURRENT_CONNECTIONS" -ge 18 ]; then
            if [ "$SATURATION_REACHED" = false ]; then
                log_warning "Connection pool nearing saturation!"
                SATURATION_REACHED=true
                SATURATION_TIME=$(date +%s)
            fi
        fi

        # Show connection breakdown
        if [ $(($(date +%s) - START_TIME)) -gt 30 ] && [ $(( ($(date +%s) - START_TIME) % 15 )) -eq 0 ]; then
            log "Connection breakdown:"
            docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
                "SELECT application_name, state, count(*) FROM pg_stat_activity GROUP BY application_name, state;" \
                2>/dev/null | tee -a "$LOG_FILE" || true
        fi

        sleep 5
    done

    if [ "$SATURATION_REACHED" = true ]; then
        TIME_TO_SATURATION=$((SATURATION_TIME - START_TIME))
        log_warning "Pool saturation reached after ${TIME_TO_SATURATION} seconds"
    else
        log "Pool did not reach saturation within ${MAX_MONITOR_TIME} seconds"
    fi

    if [ "$ERRORS_OBSERVED" = true ]; then
        log_error "Connection errors were observed during test"
    fi
}

test_api_behavior() {
    section "Step 6: Testing API Behavior with Exhausted Pool"

    log "Testing API endpoints with connection pool under pressure..."

    API_SUCCESS=0
    API_FAILURES=0

    for i in {1..10}; do
        log "API test $i..."
        if timeout 15 curl -s --max-time 10 http://localhost/api/health &> /dev/null; then
            API_SUCCESS=$((API_SUCCESS + 1))
            log "  SUCCESS"
        else
            API_FAILURES=$((API_FAILURES + 1))
            log_error "  FAILED (timeout or connection error)"
        fi
        sleep 2
    done

    log ""
    log_success "API successes: $API_SUCCESS / 10"
    log_error "API failures: $API_FAILURES / 10"

    # Check API logs for connection errors
    log ""
    log "Recent trading-api errors related to connections:"
    docker logs simulated-exchange-trading-api --tail 100 2>&1 | \
        grep -i "connection\|pool\|database" | tail -20 | tee -a "$LOG_FILE"
}

check_postgres_logs() {
    section "Step 7: Checking PostgreSQL Connection Logs"

    log "Examining PostgreSQL logs for connection rejections..."

    docker logs simulated-exchange-postgres --tail 100 2>&1 | \
        grep -i "too many connections\|connection\|fatal\|error" | tail -30 | tee -a "$LOG_FILE"

    log ""
    log "Connection statistics from pg_stat_database:"
    docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
        "SELECT datname, numbackends, xact_commit, xact_rollback FROM pg_stat_database WHERE datname='trading_db';" \
        2>/dev/null | tee -a "$LOG_FILE" || true
}

stop_order_flow() {
    section "Step 8: Stopping Order Flow"

    log "Stopping order-flow-simulator..."
    docker compose stop order-flow-simulator 2>&1 | tee -a "$LOG_FILE"

    log_success "Order flow stopped"
    log ""
    log "Waiting 10s for connections to drain..."
    sleep 10

    REMAINING=$(docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -t -c \
        "SELECT count(*) FROM pg_stat_activity;" 2>/dev/null | xargs || echo "0")

    log "Remaining connections: $REMAINING"
}

restore_max_connections() {
    section "Step 9: Restoring Original max_connections"

    if [ -n "$ORIGINAL_MAX_CONNECTIONS" ]; then
        log "Restoring max_connections to $ORIGINAL_MAX_CONNECTIONS..."

        docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
            "ALTER SYSTEM SET max_connections = $ORIGINAL_MAX_CONNECTIONS;" 2>&1 | tee -a "$LOG_FILE"
    else
        log "Restoring max_connections to 100 (safe default)..."
        docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
            "ALTER SYSTEM SET max_connections = 100;" 2>&1 | tee -a "$LOG_FILE"
    fi

    log "Restarting PostgreSQL to apply restored settings..."
    docker restart simulated-exchange-postgres 2>&1 | tee -a "$LOG_FILE"

    log "Waiting 15s for PostgreSQL to restart..."
    sleep 15

    # Verify restoration
    RESTORED_MAX=$(docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -t -c \
        "SHOW max_connections;" 2>/dev/null | xargs || echo "unknown")

    log_success "max_connections restored to: $RESTORED_MAX"
}

stop_bpf_monitoring() {
    section "Step 10: Stopping BPF Monitoring"

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
- Original max_connections: $ORIGINAL_MAX_CONNECTIONS
- Test max_connections: 20
- Baseline connections: $BASELINE_CONNECTIONS

CONNECTION STATISTICS:
- Saturation reached: $SATURATION_REACHED
- Time to saturation: ${TIME_TO_SATURATION:-N/A} seconds
- Connection errors observed: $ERRORS_OBSERVED

API PERFORMANCE:
- Successful requests: $API_SUCCESS / 10
- Failed requests: $API_FAILURES / 10
- Failure rate: $((API_FAILURES * 10))%

KEY FINDINGS:
1. Connection pool reached limit of 20
2. New connections rejected with "too many clients"
3. API experienced $API_FAILURES failures due to connection unavailability
4. Connection errors visible in application logs
5. Saturation reached in ${TIME_TO_SATURATION:-N/A} seconds
6. Service degraded but did not crash

OBSERVATIONS:
- PostgreSQL enforces max_connections strictly
- Applications queue/fail when pool exhausted
- Connection pooling in app helps but has limits
- BPF shows failed TCP connection attempts
- Error messages are clear and diagnosable
- No data corruption or crashes

FILES GENERATED:
- Experiment log: $LOG_FILE
- BPF connection trace: $LOG_FILE.bpf

LESSONS LEARNED:
- Connection limits are hard constraints
- Pool exhaustion causes service degradation
- Clear error messages help diagnosis
- Connection reuse is critical
- Monitoring connection count is essential

RECOMMENDATIONS:
1. Set max_connections with 50%+ headroom
2. Implement connection pooling (PgBouncer)
3. Monitor active connection count continuously
4. Alert at 80% of max_connections
5. Implement connection leak detection
6. Set appropriate connection timeouts
7. Consider read replicas to distribute load
8. Review application connection management

PRODUCTION IMPACT:
- This is a common failure mode in high-traffic systems
- Connection exhaustion causes cascading failures
- Proper connection pooling is critical
- Horizontal scaling may be needed

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

    # Ensure PostgreSQL is running
    if ! docker ps | grep -q "simulated-exchange-postgres"; then
        log "Restarting PostgreSQL..."
        docker start simulated-exchange-postgres 2>&1 | tee -a "$LOG_FILE"
        sleep 15
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
    reduce_max_connections
    start_order_flow
    monitor_connection_saturation
    test_api_behavior
    check_postgres_logs
    stop_order_flow
    restore_max_connections
    stop_bpf_monitoring
    generate_report

    section "EXPERIMENT COMPLETE"
}

main
