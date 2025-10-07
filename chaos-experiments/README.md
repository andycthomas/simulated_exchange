# Chaos Engineering Experiments for Trading Exchange System

This directory contains 10 comprehensive chaos engineering experiments designed to test the resilience, fault tolerance, and recovery mechanisms of the simulated trading exchange system.

## Overview

Each experiment is a standalone shell script that:
- **Documents the purpose** and expected outcomes
- **Sets up monitoring** with BPF (Berkeley Packet Filter) tracing
- **Injects a specific failure** in a controlled manner
- **Observes system behavior** during the failure
- **Captures metrics** before, during, and after the failure
- **Generates detailed reports** with findings and recommendations
- **Cleans up** automatically, even if interrupted

## Prerequisites

- **Root access** (sudo) for most experiments
- **Docker & Docker Compose** installed and running
- **bpftrace** installed: `sudo dnf install bpftrace`
- **jq** for JSON parsing: `sudo dnf install jq`
- All trading exchange services running

## Quick Start

```bash
# Make all scripts executable
chmod +x *.sh

# Run a single experiment
sudo ./01-database-sudden-death.sh

# View the results
cat /tmp/chaos-exp-01-*.log
```

## Experiment Catalog

### 1. Database Sudden Death (`01-database-sudden-death.sh`)
**Difficulty**: ⭐ Easy
**Duration**: ~5 minutes
**Root Required**: No

**Purpose**: Test how the system handles catastrophic database failure.

**What It Tests**:
- Connection retry logic
- Fallback mechanisms
- Service resilience without database
- Recovery time and orchestration

**Expected Results**:
- Trading-API loses database connection immediately
- Metrics switch to fallback/cached data
- Order submissions fail with database errors
- Automatic recovery within 30-60 seconds after restart
- Dashboard shows degraded service status

**BPF Insights**:
- TCP connection attempts to port 5432
- Failed connection syscalls
- Retry patterns and exponential backoff
- Time to successful reconnection

**Run Command**:
```bash
./01-database-sudden-death.sh
```

---

### 2. Network Partition (`02-network-partition.sh`)
**Difficulty**: ⭐⭐ Medium
**Duration**: ~10 minutes
**Root Required**: Yes (iptables)

**Purpose**: Test system behavior when network communication between trading-api and PostgreSQL is blocked.

**What It Tests**:
- TCP connection timeout behavior
- Retry and exponential backoff
- Graceful degradation to cached data
- Network-level resilience patterns

**Expected Results**:
- Trading-API cannot reach PostgreSQL (packets dropped)
- TCP retransmissions visible
- New connection attempts fail (EHOSTUNREACH)
- Metrics fall back to cached data
- Immediate recovery after partition heals

**BPF Insights**:
- TCP SYN packets being sent
- Retransmission attempts
- Connection timeout events
- RTT increasing before failure

**Run Command**:
```bash
sudo ./02-network-partition.sh
```

---

### 3. Memory Pressure & OOM Kill (`03-memory-oom-kill.sh`)
**Difficulty**: ⭐⭐ Medium
**Duration**: ~15 minutes
**Root Required**: Yes (BPF monitoring)

**Purpose**: Test system behavior when trading-api runs out of memory.

**What It Tests**:
- Container memory limits enforcement
- OOM killer behavior
- Service restart mechanisms
- Memory allocation patterns

**Expected Results**:
- Memory usage climbs to limit (512MB)
- Kernel OOM killer terminates process
- Container auto-restarts
- Brief service outage (30-60s)
- Dashboard shows service offline then recovers

**BPF Insights**:
- Memory allocation patterns
- OOM killer victim selection
- Process termination signals
- Memory pressure events

**Run Command**:
```bash
sudo ./03-memory-oom-kill.sh
```

---

### 4. CPU Throttling (`04-cpu-throttling.sh`)
**Difficulty**: ⭐ Easy
**Duration**: ~10 minutes
**Root Required**: No

**Purpose**: Test gradual performance degradation by limiting CPU resources.

**What It Tests**:
- CPU scheduler behavior
- Performance under resource constraints
- Queue buildup patterns
- Graceful degradation vs. hard failure

**Expected Results**:
- Latency increases from 400ms → 2000ms+
- Orders queue waiting for CPU
- Dashboard updates slow down
- No hard failures, just degradation

**BPF Insights**:
- CPU scheduler delays
- Context switch frequency
- Time spent in runqueue
- Hot code paths under contention

---

### 5. Disk I/O Starvation (`05-disk-io-starvation.sh`)
**Difficulty**: ⭐⭐⭐ Hard
**Duration**: ~15 minutes
**Root Required**: Yes

**Purpose**: Test system behavior when disk I/O is saturated.

**What It Tests**:
- Query performance under I/O pressure
- Checkpoint delay handling
- Database write throughput
- Filesystem cache effectiveness

**Expected Results**:
- Disk latency spikes from <1ms → 100ms+
- PostgreSQL checkpoint delays
- Query timeouts increase
- Write throughput drops dramatically

**BPF Insights**:
- I/O queue depth
- Read vs write latency distribution
- Disk wait time per process
- Filesystem cache hit rate

---

### 6. Connection Pool Exhaustion (`06-connection-pool-exhaustion.sh`)
**Difficulty**: ⭐⭐ Medium
**Duration**: ~10 minutes
**Root Required**: Yes (BPF)

**Purpose**: Test system behavior when all database connections are exhausted.

**What It Tests**:
- Connection pool management
- Queue behavior when pool is full
- Error handling for connection failures
- Connection reuse patterns

**Expected Results**:
- Connections reach max limit
- New requests wait indefinitely
- Error: "FATAL: too many connections"
- Dashboard stops updating
- Service health degrades

**BPF Insights**:
- Connection attempt patterns
- Time spent waiting for connection
- Connection reuse statistics
- Failed connection error distribution

---

### 7. Redis Cache Failure (`07-redis-cache-failure.sh`)
**Difficulty**: ⭐ Easy
**Duration**: ~5 minutes
**Root Required**: No

**Purpose**: Test cache fallback behavior when Redis fails.

**What It Tests**:
- Cache miss handling
- Database load increase
- Event bus resilience
- Graceful degradation

**Expected Results**:
- Cache misses force database queries
- Database CPU usage increases
- Latency spikes for cached endpoints
- Services remain functional (degraded)

**BPF Insights**:
- Redis connection attempts
- Database query increase
- Network traffic shift to postgres
- TCP connection state changes

---

### 8. Nginx Reverse Proxy Failure (`08-nginx-failure.sh`)
**Difficulty**: ⭐ Easy
**Duration**: ~5 minutes
**Root Required**: No

**Purpose**: Test impact of reverse proxy failure.

**What It Tests**:
- Direct service access paths
- Load balancing failure
- Rate limiting bypass
- Frontend vs backend separation

**Expected Results**:
- Dashboard becomes inaccessible
- Direct service ports still functional
- Rate limiting removed
- Browser shows connection refused

**BPF Insights**:
- TCP RST packets on port 80
- Connection timeout patterns
- Service-to-service traffic unaffected

---

### 9. Network Latency Injection (`09-network-latency.sh`)
**Difficulty**: ⭐⭐ Medium
**Duration**: ~15 minutes
**Root Required**: Yes (tc netem)

**Purpose**: Test gradual performance degradation via network delay.

**What It Tests**:
- Timeout tuning validation
- Latency sensitivity
- Retry behavior under delays
- User experience degradation

**Expected Results**:
- Latency increases progressively
- Queries timeout at high delays
- Retransmissions visible
- Dashboard shows query failures

**BPF Insights**:
- TCP RTT distribution shift
- Retransmission rate increase
- Application-level timeout behavior
- Queue buildup in kernel

---

### 10. Multi-Service Cascade Failure (`10-multi-service-failure.sh`)
**Difficulty**: ⭐⭐⭐ Hard
**Duration**: ~10 minutes
**Root Required**: No

**Purpose**: Test recovery from simultaneous failure of multiple services.

**What It Tests**:
- Dependency management
- Startup ordering
- Health check coordination
- Recovery time objectives

**Expected Results**:
- Cascade failure: postgres → redis → trading-api
- Health checks fail sequentially
- Restart order: postgres → redis → trading-api
- Recovery time: 30-60 seconds

**BPF Insights**:
- Process termination sequence
- Connection re-establishment patterns
- Service discovery attempts
- Time to first successful health check

---

## Running All Experiments

```bash
#!/bin/bash
# Run all experiments in sequence

for i in {01..10}; do
    script=$(ls ${i}-*.sh 2>/dev/null | head -1)
    if [ -f "$script" ]; then
        echo "Running experiment $i: $script"
        sudo ./$script
        echo "Waiting 30s before next experiment..."
        sleep 30
    fi
done
```

## Understanding the Output

Each experiment generates:

1. **Console output**: Real-time progress and findings
2. **Log file**: `/tmp/chaos-exp-XX-YYYYMMDD-HHMMSS.log`
3. **BPF trace**: `/tmp/chaos-exp-XX-YYYYMMDD-HHMMSS.log.bpf`
4. **Summary report**: Included at end of log file

### Sample Output Structure

```
========================================================================
CHAOS EXPERIMENT: Database Sudden Death
========================================================================

[2025-10-06 14:30:15] Checking Prerequisites
[2025-10-06 14:30:15] ✓ Docker found
[2025-10-06 14:30:15] ✓ PostgreSQL container running
...

========================================================================
Step 3: Injecting Failure - Killing PostgreSQL
========================================================================

[2025-10-06 14:30:45] Sending KILL signal to PostgreSQL container...
[2025-10-06 14:30:45] ✓ PostgreSQL killed
[2025-10-06 14:30:45] ⚠ Database is now OFFLINE
...

========================================================================
Experiment Summary Report
========================================================================

KEY FINDINGS:
1. Database failure detection: Immediate (< 1s)
2. Recovery time: 15s from restart to operational
...
```

## Interpreting BPF Traces

BPF traces provide low-level system observability. Key patterns to look for:

- **Connection attempts**: Frequent retries indicate resilience logic
- **Retransmissions**: Network issues or partitions
- **OOM events**: Memory pressure and limits
- **Timeouts**: Application-level vs network-level
- **Error codes**: ECONNREFUSED (111), ETIMEDOUT (110), etc.

## Safety Notes

- ⚠️ These experiments **intentionally break things**
- ✅ All experiments include **automatic cleanup**
- ✅ Services **auto-recover** via Docker restart policies
- ⚠️ Run on **test/dev environments only**
- ✅ Each experiment is **idempotent** and can be safely re-run

## Troubleshooting

### Experiment doesn't complete
- Check logs in `/tmp/chaos-exp-*.log`
- Manually run cleanup: `docker compose restart`
- Ensure sufficient privileges (sudo)

### BPF monitoring fails
- Install bpftrace: `sudo dnf install bpftrace`
- Check kernel version: `uname -r` (need 4.9+)
- Experiments still run without BPF (reduced insights)

### Services don't recover
- Check Docker: `docker compose ps`
- Restart manually: `docker compose restart`
- Check logs: `docker logs simulated-exchange-trading-api`

## Next Steps

After running experiments:

1. **Review BPF traces** for low-level insights
2. **Analyze recovery times** and identify bottlenecks
3. **Tune timeouts** based on observed behavior
4. **Implement improvements** (circuit breakers, retries, etc.)
5. **Set up alerts** for conditions observed during chaos

## References

- [Principles of Chaos Engineering](https://principlesofchaos.org/)
- [BPF Performance Tools](http://www.brendangregg.com/bpf-performance-tools-book.html)
- [Netflix Chaos Monkey](https://netflix.github.io/chaosmonkey/)

## Contributing

To add new experiments:
1. Copy an existing script as template
2. Follow the naming convention: `NN-experiment-name.sh`
3. Include comprehensive documentation in header
4. Test thoroughly before committing
5. Update this README with experiment details
