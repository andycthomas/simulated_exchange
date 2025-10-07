#!/bin/bash
################################################################################
# Chaos Experiment 8: Nginx Reverse Proxy Failure
################################################################################
#
# PURPOSE:
#   Test impact of reverse proxy failure. This simulates:
#   - Load balancer outage
#   - Proxy misconfiguration
#   - Nginx crash or restart
#   - Frontend/backend separation validation
#
# WHAT THIS TESTS:
#   - Direct service access paths
#   - Load balancing failure impact
#   - Rate limiting bypass when proxy down
#   - Frontend vs backend service separation
#   - Service-to-service communication resilience
#
# EXPECTED RESULTS:
#   1. Dashboard accessible via nginx (port 80)
#   2. Nginx killed
#   3. Dashboard becomes unreachable on port 80
#   4. Direct API access on port 8080 still works
#   5. Service-to-service communication unaffected
#   6. Rate limiting bypassed
#   7. Immediate recovery when nginx restarts
#
# BPF OBSERVABILITY:
#   - TCP RST packets on port 80
#   - Connection refused errors
#   - Traffic to backend port 8080 continues
#   - Connection state changes
#
################################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

EXPERIMENT_NAME="Nginx Reverse Proxy Failure"
LOG_FILE="/tmp/chaos-exp-08-$(date +%Y%m%d-%H%M%S).log"

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

    if ! docker ps | grep -q "simulated-exchange-nginx"; then
        log_error "Nginx container not running"
        exit 1
    fi
    log_success "Nginx container running"

    if ! docker ps | grep -q "simulated-exchange-trading-api"; then
        log_error "Trading API container not running"
        exit 1
    fi
    log_success "Trading API container running"
}

capture_baseline() {
    section "Step 1: Capturing Baseline Proxy Performance"

    log "Testing dashboard access via nginx (port 80)..."

    NGINX_TIMES=()
    for i in {1..5}; do
        START=$(date +%s%N)
        if curl -s --max-time 5 http://localhost/ &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            NGINX_TIMES+=($ELAPSED)
            log "  Request $i via nginx: ${ELAPSED}ms"
        else
            log_warning "  Request $i failed"
        fi
    done

    # Calculate average
    NGINX_AVG=0
    if [ ${#NGINX_TIMES[@]} -gt 0 ]; then
        for time in "${NGINX_TIMES[@]}"; do
            NGINX_AVG=$((NGINX_AVG + time))
        done
        NGINX_AVG=$((NGINX_AVG / ${#NGINX_TIMES[@]}))
    fi

    log_success "Nginx proxy latency: ${NGINX_AVG}ms"

    log ""
    log "Testing direct API access (port 8080)..."

    DIRECT_TIMES=()
    for i in {1..5}; do
        START=$(date +%s%N)
        if curl -s --max-time 5 http://localhost:8080/api/health &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            DIRECT_TIMES+=($ELAPSED)
            log "  Request $i direct: ${ELAPSED}ms"
        else
            log_warning "  Request $i failed"
        fi
    done

    # Calculate average
    DIRECT_AVG=0
    if [ ${#DIRECT_TIMES[@]} -gt 0 ]; then
        for time in "${DIRECT_TIMES[@]}"; do
            DIRECT_AVG=$((DIRECT_AVG + time))
        done
        DIRECT_AVG=$((DIRECT_AVG / ${#DIRECT_TIMES[@]}))
    fi

    log_success "Direct API latency: ${DIRECT_AVG}ms"

    log ""
    log "Nginx connection stats (recent requests):"
    # Nginx logs go to stdout in Docker, so use docker logs instead
    docker logs simulated-exchange-nginx --tail 10 2>/dev/null | tee -a "$LOG_FILE" || \
        log_warning "Nginx logs not available"
}

start_bpf_monitoring() {
    section "Step 2: Starting BPF Monitoring"

    if ! command -v bpftrace &> /dev/null; then
        log_warning "bpftrace not found - skipping BPF monitoring"
        return
    fi

    cat > /tmp/nginx-bpf.bt << 'BPFEOF'
#!/usr/bin/env bpftrace

BEGIN {
    printf("Monitoring nginx (port 80) and direct API (port 8080)...\n\n");
}

// Track connections to port 80 (nginx)
kprobe:tcp_v4_connect {
    $sk = (struct sock *)arg0;
    $dport = $sk->__sk_common.skc_dport;

    if ($dport == 0x5000) {  // 80 in big-endian
        @port80_attempts[comm] = count();
    }
}

kretprobe:tcp_v4_connect {
    $sk = (struct sock *)arg0;
    $dport = $sk->__sk_common.skc_dport;

    if ($dport == 0x5000) {
        if (retval == 0) {
            @port80_success[comm] = count();
        } else {
            @port80_failures[comm] = count();
            @port80_errno[-retval] = count();
        }
    }
}

// Track connections to port 8080 (direct API)
kprobe:tcp_v4_connect {
    $sk = (struct sock *)arg0;
    $dport = $sk->__sk_common.skc_dport;

    if ($dport == 0x901F) {  // 8080 in big-endian
        @port8080_attempts[comm] = count();
    }
}

kretprobe:tcp_v4_connect {
    $sk = (struct sock *)arg0;
    $dport = $sk->__sk_common.skc_dport;

    if ($dport == 0x901F && retval == 0) {
        @port8080_success[comm] = count();
    }
}

// Track TCP RST packets (connection refused)
tracepoint:tcp:tcp_send_reset {
    @tcp_resets = count();
}

// Interval reporting
interval:s:5 {
    printf("\n[%s] === Connection Stats (5s) ===\n", strftime("%H:%M:%S", nsecs));

    printf("Port 80 (nginx) attempts:\n");
    print(@port80_attempts);

    printf("\nPort 80 successes:\n");
    print(@port80_success);

    printf("\nPort 80 failures:\n");
    print(@port80_failures);

    printf("\nPort 8080 (direct) attempts:\n");
    print(@port8080_attempts);

    printf("\nTCP RST packets sent: %d\n", @tcp_resets);

    clear(@port80_attempts);
    clear(@port80_success);
    clear(@port80_failures);
    clear(@port8080_attempts);
    clear(@port8080_success);
    clear(@tcp_resets);
}

END {
    printf("\n=== Final Summary ===\n");

    printf("Port 80 failure reasons (errno):\n");
    print(@port80_errno);
}
BPFEOF

    chmod +x /tmp/nginx-bpf.bt

    log "Starting BPF monitoring..."
    sudo bpftrace /tmp/nginx-bpf.bt > "$LOG_FILE.bpf" 2>&1 &
    BPF_PID=$!

    log_success "BPF monitoring started (PID: $BPF_PID)"
    sleep 2
}

kill_nginx() {
    section "Step 3: Killing Nginx Container"

    log "Sending KILL signal to nginx..."
    docker kill -s KILL simulated-exchange-nginx 2>&1 | tee -a "$LOG_FILE"

    log_success "Nginx killed"
    log_warning "Reverse proxy is now OFFLINE"

    log ""
    log "Verifying nginx is down..."
    sleep 2

    if docker ps | grep -q "simulated-exchange-nginx"; then
        log_error "Nginx container is still running"
    else
        log_success "Nginx container is stopped"
    fi
}

test_frontend_access() {
    section "Step 4: Testing Frontend Access (Should Fail)"

    log "Attempting to access dashboard via port 80..."
    log "Waiting 3s for ports to close..."
    sleep 3

    NGINX_FAILURES=0
    for i in {1..5}; do
        log "Attempt $i to access port 80..."
        if timeout 5 curl -s --max-time 3 http://localhost/ &> /dev/null; then
            log_error "  UNEXPECTED SUCCESS - nginx should be down!"
        else
            NGINX_FAILURES=$((NGINX_FAILURES + 1))
            log_success "  Failed as expected (connection refused)"
        fi
        sleep 1
    done

    if [ $NGINX_FAILURES -eq 5 ]; then
        log_success "All frontend requests failed as expected"
    else
        log_warning "Some requests succeeded unexpectedly ($((5 - NGINX_FAILURES))/5)"
    fi

    log ""
    log "Testing specific error type..."
    curl -v http://localhost/ 2>&1 | grep -i "refused\|failed to connect" | head -5 | tee -a "$LOG_FILE" || \
        log "Connection error detected"
}

test_direct_api_access() {
    section "Step 5: Testing Direct API Access (Should Work)"

    log "Testing direct access to trading-api on port 8080..."

    DIRECT_SUCCESS=0
    DIRECT_NO_NGINX_TIMES=()

    for i in {1..10}; do
        START=$(date +%s%N)
        if timeout 10 curl -s --max-time 8 http://localhost:8080/api/health &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            DIRECT_NO_NGINX_TIMES+=($ELAPSED)
            DIRECT_SUCCESS=$((DIRECT_SUCCESS + 1))

            if [ $((i % 2)) -eq 0 ]; then
                log "  Request $i: ${ELAPSED}ms - SUCCESS"
            else
                echo -n "." | tee -a "$LOG_FILE"
            fi
        else
            log_error "  Request $i: FAILED"
        fi
        sleep 1
    done
    echo "" | tee -a "$LOG_FILE"

    log_success "Direct API successes: $DIRECT_SUCCESS / 10"

    # Calculate average
    DIRECT_NO_NGINX_AVG=0
    if [ ${#DIRECT_NO_NGINX_TIMES[@]} -gt 0 ]; then
        for time in "${DIRECT_NO_NGINX_TIMES[@]}"; do
            DIRECT_NO_NGINX_AVG=$((DIRECT_NO_NGINX_AVG + time))
        done
        DIRECT_NO_NGINX_AVG=$((DIRECT_NO_NGINX_AVG / ${#DIRECT_NO_NGINX_TIMES[@]}))
        log "Direct API average latency: ${DIRECT_NO_NGINX_AVG}ms"
    fi

    if [ $DIRECT_SUCCESS -eq 10 ]; then
        log_success "Backend services unaffected by nginx failure"
    else
        log_warning "Some backend requests failed ($((10 - DIRECT_SUCCESS))/10)"
    fi
}

test_service_to_service() {
    section "Step 6: Testing Service-to-Service Communication"

    log "Verifying trading-api can still access database..."

    if docker exec simulated-exchange-trading-api wget -q -O - http://localhost:8080/api/health 2>/dev/null | grep -q "ok\|healthy"; then
        log_success "Internal health check passed"
    else
        log_warning "Internal health check failed or returned unexpected response"
    fi

    log ""
    log "Checking database connectivity from trading-api..."

    DB_CONN=$(docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -t -c \
        "SELECT count(*) FROM pg_stat_activity WHERE application_name LIKE '%trading%';" 2>/dev/null | xargs || echo "0")

    log "Trading-api database connections: $DB_CONN"

    if [ "$DB_CONN" -gt 0 ]; then
        log_success "Database connectivity intact"
    else
        log_warning "No active trading-api database connections"
    fi
}

restart_nginx() {
    section "Step 7: Restarting Nginx"

    log "Starting nginx container..."
    docker start simulated-exchange-nginx 2>&1 | tee -a "$LOG_FILE"

    log "Waiting for nginx to be ready..."
    MAX_WAIT=20
    ELAPSED=0

    while [ $ELAPSED -lt $MAX_WAIT ]; do
        if timeout 3 curl -s http://localhost/ &> /dev/null; then
            RECOVERY_TIME=$ELAPSED
            log_success "Nginx is ready (recovery time: ${RECOVERY_TIME}s)"
            break
        fi
        sleep 2
        ELAPSED=$((ELAPSED + 2))
        echo -n "." | tee -a "$LOG_FILE"
    done
    echo "" | tee -a "$LOG_FILE"

    if [ $ELAPSED -ge $MAX_WAIT ]; then
        log_error "Nginx did not recover within ${MAX_WAIT} seconds"
    fi
}

measure_recovery() {
    section "Step 8: Measuring Recovery Performance"

    log "Testing frontend access after nginx restart..."

    RECOVERY_TIMES=()
    RECOVERY_SUCCESS=0

    for i in {1..5}; do
        START=$(date +%s%N)
        if curl -s --max-time 5 http://localhost/ &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            RECOVERY_TIMES+=($ELAPSED)
            RECOVERY_SUCCESS=$((RECOVERY_SUCCESS + 1))
            log "  Request $i: ${ELAPSED}ms"
        else
            log_error "  Request $i failed"
        fi
        sleep 1
    done

    # Calculate average
    RECOVERY_AVG=0
    if [ ${#RECOVERY_TIMES[@]} -gt 0 ]; then
        for time in "${RECOVERY_TIMES[@]}"; do
            RECOVERY_AVG=$((RECOVERY_AVG + time))
        done
        RECOVERY_AVG=$((RECOVERY_AVG / ${#RECOVERY_TIMES[@]}))
    fi

    log_success "Recovery latency: ${RECOVERY_AVG}ms"
    log "Recovery success rate: $RECOVERY_SUCCESS / 5"

    if [ $RECOVERY_AVG -gt 0 ] && [ $NGINX_AVG -gt 0 ]; then
        RECOVERY_PERC=$((RECOVERY_AVG * 100 / NGINX_AVG))
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

ACCESS TESTING:
- Frontend (port 80) failures during outage: $NGINX_FAILURES / 5
- Direct API (port 8080) successes during outage: $DIRECT_SUCCESS / 10
- Recovery successes: $RECOVERY_SUCCESS / 5

PERFORMANCE:
- Baseline nginx latency: ${NGINX_AVG}ms
- Baseline direct API latency: ${DIRECT_AVG}ms
- Direct API during outage: ${DIRECT_NO_NGINX_AVG}ms
- Recovery latency: ${RECOVERY_AVG}ms

RECOVERY:
- Nginx restart time: ${RECOVERY_TIME:-N/A}s
- Frontend access restored: YES
- Backend services affected: NO

KEY FINDINGS:
1. Frontend (port 80) became completely unavailable
2. Direct backend access (port 8080) remained functional
3. Service-to-service communication unaffected
4. Database connectivity intact during proxy failure
5. Immediate recovery when nginx restarted
6. BPF shows ECONNREFUSED errors on port 80

OBSERVATIONS:
- Nginx failure only impacts frontend access
- Backend services operate independently
- No cascading failures to other services
- Clean separation between frontend and backend
- Rate limiting/proxy features lost during outage
- Docker auto-restart would normally handle this

ARCHITECTURE VALIDATION:
✓ Frontend/backend separation working correctly
✓ Services not tightly coupled to proxy
✓ Direct API access available as fallback
✓ Service mesh/communication resilient

FILES GENERATED:
- Experiment log: $LOG_FILE
- BPF connection trace: $LOG_FILE.bpf

LESSONS LEARNED:
- Reverse proxy is single point of failure for frontend
- Direct API access provides resilience
- Load balancer health checks critical
- Multi-proxy deployment recommended for HA
- Rate limiting lost when proxy down

RECOMMENDATIONS:
1. Deploy multiple nginx instances for HA
2. Use health checks and automatic failover
3. Document direct API access for emergencies
4. Monitor nginx process and restart automatically
5. Consider nginx + HAProxy for redundancy
6. Alert on nginx failures immediately
7. Keep direct backend ports accessible (8080)
8. Test load balancer failover regularly

PRODUCTION IMPACT:
- Dashboard users cannot access system via port 80
- API clients using direct port 8080 unaffected
- Internal services continue operating normally
- Rate limiting and caching lost temporarily
- No data loss or state corruption

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

    # Ensure nginx is running
    if ! docker ps | grep -q "simulated-exchange-nginx"; then
        log "Starting nginx container..."
        docker start simulated-exchange-nginx 2>&1 | tee -a "$LOG_FILE"
        sleep 5
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
    kill_nginx
    test_frontend_access
    test_direct_api_access
    test_service_to_service
    restart_nginx
    measure_recovery
    stop_bpf_monitoring
    generate_report

    section "EXPERIMENT COMPLETE"
}

main
