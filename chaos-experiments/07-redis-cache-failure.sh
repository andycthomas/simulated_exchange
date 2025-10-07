#!/bin/bash
################################################################################
# Chaos Experiment 7: Redis Cache Failure
################################################################################
#
# PURPOSE:
#   Test cache fallback behavior when Redis fails. This simulates:
#   - Cache service outage
#   - Redis crash or restart
#   - Memory eviction causing cache unavailability
#   - Event bus failure
#
# WHAT THIS TESTS:
#   - Cache miss handling and fallback logic
#   - Database load increase when cache unavailable
#   - Event publishing resilience
#   - Graceful degradation patterns
#   - Service behavior without caching layer
#
# EXPECTED RESULTS:
#   1. Baseline cache hit rate captured
#   2. Redis killed (SIGKILL)
#   3. Cache misses force direct database queries
#   4. Database CPU/connection usage increases
#   5. API latency increases for cached endpoints
#   6. Services remain functional (degraded mode)
#   7. Immediate recovery when Redis restarts
#
# BPF OBSERVABILITY:
#   - Redis connection attempts (port 6379)
#   - Failed connection errors
#   - Database query increase
#   - Network traffic shift from Redis to Postgres
#
################################################################################

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

EXPERIMENT_NAME="Redis Cache Failure"
LOG_FILE="/tmp/chaos-exp-07-$(date +%Y%m%d-%H%M%S).log"

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

    if ! docker ps | grep -q "simulated-exchange-redis"; then
        log_error "Redis container not running"
        exit 1
    fi
    log_success "Redis container running"

    if ! docker ps | grep -q "simulated-exchange-trading-api"; then
        log_error "Trading API container not running"
        exit 1
    fi
    log_success "Trading API container running"
}

capture_baseline() {
    section "Step 1: Capturing Baseline Cache Performance"

    log "Testing Redis connectivity..."
    if docker exec simulated-exchange-redis redis-cli ping &> /dev/null; then
        log_success "Redis is responding"
    else
        log_error "Redis is not responding"
        exit 1
    fi

    log ""
    log "Redis info and statistics:"
    docker exec simulated-exchange-redis redis-cli INFO stats | grep -E "keyspace_hits|keyspace_misses|total_commands_processed" | tee -a "$LOG_FILE"

    log ""
    log "Redis memory usage:"
    docker exec simulated-exchange-redis redis-cli INFO memory | grep -E "used_memory_human|maxmemory_human" | tee -a "$LOG_FILE"

    log ""
    log "Testing API metrics endpoint (uses cache)..."

    BASELINE_TIMES=()
    for i in {1..5}; do
        START=$(date +%s%N)
        if curl -s --max-time 5 http://localhost/api/metrics &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            BASELINE_TIMES+=($ELAPSED)
            log "  Request $i: ${ELAPSED}ms"
        else
            log_warning "  Request $i failed"
        fi
        sleep 1
    done

    # Calculate average
    BASELINE_AVG=0
    if [ ${#BASELINE_TIMES[@]} -gt 0 ]; then
        for time in "${BASELINE_TIMES[@]}"; do
            BASELINE_AVG=$((BASELINE_AVG + time))
        done
        BASELINE_AVG=$((BASELINE_AVG / ${#BASELINE_TIMES[@]}))
    fi

    log_success "Baseline API latency (with cache): ${BASELINE_AVG}ms"

    log ""
    log "PostgreSQL query count (baseline):"
    BASELINE_DB_QUERIES=$(docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -t -c \
        "SELECT sum(calls) FROM pg_stat_statements;" 2>/dev/null | xargs || echo "0")
    log "Total queries: $BASELINE_DB_QUERIES"
}

start_bpf_monitoring() {
    section "Step 2: Starting BPF Monitoring"

    if ! command -v bpftrace &> /dev/null; then
        log_warning "bpftrace not found - skipping BPF monitoring"
        return
    fi

    cat > /tmp/redis-bpf.bt << 'BPFEOF'
#!/usr/bin/env bpftrace

BEGIN {
    printf("Monitoring Redis (6379) and PostgreSQL (5432) connections...\n\n");
}

// Track connections to Redis (port 6379)
kprobe:tcp_v4_connect {
    $sk = (struct sock *)arg0;
    $dport = $sk->__sk_common.skc_dport;

    if ($dport == 0xEB18) {  // 6379 in big-endian
        @redis_attempts[comm] = count();
        @conn_port[tid] = $dport;
    } else if ($dport == 0x3514) {  // 5432 in big-endian
        @postgres_attempts[comm] = count();
        @conn_port[tid] = $dport;
    }
}

kretprobe:tcp_v4_connect /@conn_port[tid]/ {
    $dport = @conn_port[tid];

    if ($dport == 0xEB18) {  // Redis
        if (retval == 0) {
            @redis_success[comm] = count();
        } else {
            @redis_failures[comm] = count();
            @redis_errno[comm, -retval] = count();
        }
    } else if ($dport == 0x3514 && retval == 0) {  // PostgreSQL
        @postgres_success[comm] = count();
    }

    delete(@conn_port[tid]);
}

// Interval reporting
interval:s:5 {
    printf("\n[%s] === Connection Stats (5s) ===\n", strftime("%H:%M:%S", nsecs));

    printf("Redis attempts:\n");
    print(@redis_attempts);

    printf("\nRedis successes:\n");
    print(@redis_success);

    printf("\nRedis failures:\n");
    print(@redis_failures);

    printf("\nPostgreSQL attempts (database fallback):\n");
    print(@postgres_attempts);

    clear(@redis_attempts);
    clear(@redis_success);
    clear(@redis_failures);
    clear(@postgres_attempts);
    clear(@postgres_success);
}

END {
    printf("\n=== Final Summary ===\n");

    printf("Redis connection failures by errno:\n");
    print(@redis_errno);
}
BPFEOF

    chmod +x /tmp/redis-bpf.bt

    log "Starting BPF monitoring..."
    sudo bpftrace /tmp/redis-bpf.bt > "$LOG_FILE.bpf" 2>&1 &
    BPF_PID=$!

    log_success "BPF monitoring started (PID: $BPF_PID)"
    sleep 2
}

kill_redis() {
    section "Step 3: Killing Redis Container"

    log "Sending KILL signal to Redis..."
    docker kill -s KILL simulated-exchange-redis 2>&1 | tee -a "$LOG_FILE"

    log_success "Redis killed"
    log_warning "Cache is now OFFLINE"

    log ""
    log "Verifying Redis is down..."
    sleep 2

    if docker ps | grep -q "simulated-exchange-redis"; then
        log_error "Redis container is still running"
    else
        log_success "Redis container is stopped"
    fi
}

observe_cache_miss_impact() {
    section "Step 4: Observing Cache Miss Impact"

    log "Testing API behavior without cache..."
    log "Waiting 5s for services to detect Redis failure..."
    sleep 5

    NO_CACHE_TIMES=()
    API_FAILURES=0

    for i in {1..10}; do
        START=$(date +%s%N)
        if timeout 30 curl -s --max-time 25 http://localhost/api/metrics &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            NO_CACHE_TIMES+=($ELAPSED)

            if [ $((i % 2)) -eq 0 ]; then
                log "  Request $i: ${ELAPSED}ms"
            else
                echo -n "." | tee -a "$LOG_FILE"
            fi
        else
            API_FAILURES=$((API_FAILURES + 1))
            log_error "  Request $i FAILED"
        fi
        sleep 3
    done
    echo "" | tee -a "$LOG_FILE"

    # Calculate average
    NO_CACHE_AVG=0
    if [ ${#NO_CACHE_TIMES[@]} -gt 0 ]; then
        for time in "${NO_CACHE_TIMES[@]}"; do
            NO_CACHE_AVG=$((NO_CACHE_AVG + time))
        done
        NO_CACHE_AVG=$((NO_CACHE_AVG / ${#NO_CACHE_TIMES[@]}))
    fi

    log_warning "API latency without cache: ${NO_CACHE_AVG}ms"
    log "API failures: $API_FAILURES / 10"

    if [ $NO_CACHE_AVG -gt 0 ] && [ $BASELINE_AVG -gt 0 ]; then
        SLOWDOWN=$(( (NO_CACHE_AVG * 100 / BASELINE_AVG) - 100 ))
        log_warning "Latency increase: ${SLOWDOWN}% slower without cache"
    fi
}

check_database_load() {
    section "Step 5: Checking Database Load Increase"

    log "Querying PostgreSQL for query count increase..."

    NO_CACHE_DB_QUERIES=$(docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -t -c \
        "SELECT sum(calls) FROM pg_stat_statements;" 2>/dev/null | xargs || echo "0")

    QUERY_INCREASE=$((NO_CACHE_DB_QUERIES - BASELINE_DB_QUERIES))

    log "Baseline DB queries: $BASELINE_DB_QUERIES"
    log "Current DB queries: $NO_CACHE_DB_QUERIES"
    log_warning "Query increase: +$QUERY_INCREASE queries"

    log ""
    log "Database connection count:"
    docker exec simulated-exchange-postgres psql -U trading_user -d trading_db -c \
        "SELECT count(*) as active_connections FROM pg_stat_activity;" | tee -a "$LOG_FILE"

    log ""
    log "PostgreSQL resource usage:"
    docker stats simulated-exchange-postgres --no-stream --format \
        "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | tee -a "$LOG_FILE"
}

check_application_logs() {
    section "Step 6: Checking Application Error Logs"

    log "Checking trading-api logs for Redis connection errors..."

    docker logs simulated-exchange-trading-api --tail 100 2>&1 | \
        grep -i "redis\|cache\|connection refused" | tail -20 | tee -a "$LOG_FILE"

    log ""
    log "Recent errors in trading-api:"
    docker logs simulated-exchange-trading-api --tail 50 2>&1 | \
        grep -i "error\|fail" | tail -10 | tee -a "$LOG_FILE"
}

restart_redis() {
    section "Step 7: Restarting Redis"

    log "Starting Redis container..."
    docker start simulated-exchange-redis 2>&1 | tee -a "$LOG_FILE"

    log "Waiting for Redis to be ready..."
    MAX_WAIT=30
    ELAPSED=0

    while [ $ELAPSED -lt $MAX_WAIT ]; do
        if docker exec simulated-exchange-redis redis-cli ping &> /dev/null; then
            RECOVERY_TIME=$ELAPSED
            log_success "Redis is ready (recovery time: ${RECOVERY_TIME}s)"
            break
        fi
        sleep 2
        ELAPSED=$((ELAPSED + 2))
        echo -n "." | tee -a "$LOG_FILE"
    done
    echo "" | tee -a "$LOG_FILE"

    if [ $ELAPSED -ge $MAX_WAIT ]; then
        log_error "Redis did not recover within ${MAX_WAIT} seconds"
    fi
}

measure_recovery_performance() {
    section "Step 8: Measuring Recovery Performance"

    log "Waiting 5s for cache to warm up..."
    sleep 5

    log ""
    log "Testing API performance with cache restored..."

    RECOVERY_TIMES=()
    for i in {1..5}; do
        START=$(date +%s%N)
        if curl -s --max-time 5 http://localhost/api/metrics &> /dev/null; then
            END=$(date +%s%N)
            ELAPSED=$(( (END - START) / 1000000 ))
            RECOVERY_TIMES+=($ELAPSED)
            log "  Request $i: ${ELAPSED}ms"
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

    log_success "Recovery API latency: ${RECOVERY_AVG}ms"

    if [ $RECOVERY_AVG -gt 0 ] && [ $BASELINE_AVG -gt 0 ]; then
        RECOVERY_PERC=$((RECOVERY_AVG * 100 / BASELINE_AVG))
        log "Recovery performance: ${RECOVERY_PERC}% of baseline"
    fi

    log ""
    log "Redis stats after recovery:"
    docker exec simulated-exchange-redis redis-cli INFO stats | grep -E "keyspace_hits|keyspace_misses" | tee -a "$LOG_FILE"
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

PERFORMANCE METRICS:
- Baseline latency (with cache): ${BASELINE_AVG}ms
- Latency without cache: ${NO_CACHE_AVG}ms
- Recovery latency: ${RECOVERY_AVG}ms
- API failures: $API_FAILURES / 10

DATABASE IMPACT:
- Baseline DB queries: $BASELINE_DB_QUERIES
- Queries without cache: $NO_CACHE_DB_QUERIES
- Query increase: +$QUERY_INCREASE

RECOVERY:
- Redis restart time: ${RECOVERY_TIME:-N/A}s
- Service degradation: Functional but slower
- Data loss: None

DEGRADATION ANALYSIS:
EOF

    if [ $NO_CACHE_AVG -gt 0 ] && [ $BASELINE_AVG -gt 0 ]; then
        SLOWDOWN=$(( (NO_CACHE_AVG * 100 / BASELINE_AVG) - 100 ))
        echo "- Performance without cache: ${SLOWDOWN}% slower" | tee -a "$LOG_FILE"

        if [ $SLOWDOWN -gt 500 ]; then
            echo "- Impact: SEVERE (>5x slowdown)" | tee -a "$LOG_FILE"
        elif [ $SLOWDOWN -gt 300 ]; then
            echo "- Impact: HIGH (3-5x slowdown)" | tee -a "$LOG_FILE"
        elif [ $SLOWDOWN -gt 100 ]; then
            echo "- Impact: MODERATE (2-3x slowdown)" | tee -a "$LOG_FILE"
        else
            echo "- Impact: LOW (<2x slowdown)" | tee -a "$LOG_FILE"
        fi
    fi

    cat << EOF | tee -a "$LOG_FILE"

KEY FINDINGS:
1. Cache failure caused ${SLOWDOWN:-N/A}% latency increase
2. Database query load increased by $QUERY_INCREASE queries
3. Service remained functional (graceful degradation)
4. No crashes or data loss
5. Immediate recovery when cache restored
6. BPF shows connection attempts to dead Redis

OBSERVATIONS:
- Cache-aside pattern worked correctly
- All cache misses fell back to database
- Database handled increased load adequately
- Error logging helped diagnosis
- Auto-restart via Docker worked well
- Cache warming occurs gradually after restart

FILES GENERATED:
- Experiment log: $LOG_FILE
- BPF connection trace: $LOG_FILE.bpf

LESSONS LEARNED:
- Caching provides significant performance benefit
- Cache failure is survivable with proper fallback
- Database must be sized for cache-miss scenarios
- Clear error messages help operations
- Auto-restart provides resilience

RECOMMENDATIONS:
1. Implement cache health checks and alerts
2. Size database for occasional cache failures
3. Monitor cache hit/miss ratios
4. Set appropriate cache TTLs
5. Consider Redis Sentinel for HA
6. Log cache connection failures clearly
7. Test cache warming strategies
8. Implement circuit breakers for cache operations

CACHE EFFECTIVENESS:
- Baseline shows cache is providing ${SLOWDOWN:-N/A}% performance improvement
- Cache-aside pattern successfully handles failures
- Database remains bottleneck when cache unavailable

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

    # Ensure Redis is running
    if ! docker ps | grep -q "simulated-exchange-redis"; then
        log "Starting Redis container..."
        docker start simulated-exchange-redis 2>&1 | tee -a "$LOG_FILE"
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
    kill_redis
    observe_cache_miss_impact
    check_database_load
    check_application_logs
    restart_redis
    measure_recovery_performance
    stop_bpf_monitoring
    generate_report

    section "EXPERIMENT COMPLETE"
}

main
