# Chaos Engineering Experiments - Quick Reference Guide

## Executive Summary

This directory contains **10 production-ready chaos engineering experiments** designed to test your simulated trading exchange system. Each experiment is fully automated, self-documenting, and includes BPF-based observability.

**Total Testing Time**: ~90 minutes for all 10 experiments
**Prerequisites**: Root access, bpftrace, docker, jq

---

## Experiment Matrix

| # | Name | Difficulty | Duration | Root | Key Learning |
|---|------|------------|----------|------|--------------|
| 01 | Database Sudden Death | ⭐ Easy | 5 min | No | Connection resilience & retry logic |
| 02 | Network Partition | ⭐⭐ Medium | 10 min | Yes | TCP behavior under network failures |
| 03 | Memory OOM Kill | ⭐⭐ Medium | 15 min | Yes | Resource limits & OOM killer |
| 04 | CPU Throttling | ⭐ Easy | 10 min | No | Performance degradation patterns |
| 05 | Disk I/O Starvation | ⭐⭐⭐ Hard | 15 min | Yes | Database performance under I/O pressure |
| 06 | Connection Pool Exhaustion | ⭐⭐ Medium | 10 min | Yes | Connection management limits |
| 07 | Redis Cache Failure | ⭐ Easy | 5 min | No | Cache fallback mechanisms |
| 08 | Nginx Proxy Failure | ⭐ Easy | 5 min | No | Frontend vs backend separation |
| 09 | Network Latency Injection | ⭐⭐ Medium | 15 min | Yes | Latency sensitivity & timeouts |
| 10 | Multi-Service Cascade | ⭐⭐⭐ Hard | 10 min | No | Dependency orchestration |

---

## Detailed Experiment Descriptions

### Experiment 01: Database Sudden Death
```bash
./01-database-sudden-death.sh
```

**Simulates**: Hardware failure, datacenter outage, disk corruption

**Steps**:
1. Capture baseline metrics and database connections
2. Start BPF monitoring for TCP connections to port 5432
3. Kill PostgreSQL with SIGKILL (simulates sudden death)
4. Observe connection retry attempts and failures
5. Restart PostgreSQL and measure recovery time
6. Generate report with retry patterns and recovery metrics

**Expected Observations**:
- Immediate connection failures
- Exponential backoff in retry attempts (visible in BPF)
- Trading API switches to fallback data
- Dashboard shows "Database unavailable"
- Auto-recovery in 30-60 seconds

**BPF Metrics**:
- Connection attempts per second
- Failed connections with errno (ECONNREFUSED: 111)
- Time between retry attempts
- Successful reconnection timestamp

**Key Files Generated**:
- `/tmp/chaos-exp-01-*.log` - Full execution log
- `/tmp/chaos-exp-01-*.log.bpf` - BPF trace with connection attempts

---

### Experiment 02: Network Partition
```bash
sudo ./02-network-partition.sh
```

**Simulates**: Network switch failure, firewall misconfiguration, split-brain

**Steps**:
1. Get container network namespaces and IP addresses
2. Test baseline connectivity (ping, TCP connection)
3. Start BPF monitoring for TCP retransmissions
4. Inject iptables DROP rule in trading-api namespace
5. Observe TCP retransmit attempts and timeouts
6. Remove DROP rule to heal partition
7. Measure recovery time

**Expected Observations**:
- Packets sent but never acknowledged
- TCP retransmissions with exponential backoff
- Eventually: Connection timeout (ETIMEDOUT)
- No service crashes (graceful degradation)
- Immediate recovery when partition heals

**BPF Metrics**:
- TCP SYN packets sent
- Retransmission count per connection
- RTT (Round Trip Time) before timeout
- Time to detect partition (~15 seconds for timeout)

**Why This Matters**: Tests split-brain scenarios where both services are healthy but cannot communicate.

---

### Experiment 03: Memory Pressure & OOM Kill
```bash
sudo ./03-memory-oom-kill.sh
```

**Simulates**: Memory leak, traffic spike, insufficient resources

**Steps**:
1. Capture baseline memory usage
2. Reduce container memory limit to 512MB
3. Start BPF monitoring for memory allocation
4. Start order flow simulator to increase memory pressure
5. Monitor memory growth until OOM kill
6. Observe container restart
7. Restore original memory limit

**Expected Observations**:
- Memory usage climbs steadily
- Page allocation rate increases
- Memory pressure events in kernel
- OOM killer selects and terminates process
- Docker auto-restarts container
- Brief service outage (30-60s)

**BPF Metrics**:
- Page allocations vs frees (delta shows growth)
- Memory pressure (direct reclaim) events
- OOM victim selection (PID, process name)
- Time from 90% memory to OOM kill

**Kernel Logs Captured**:
```
Out of memory: Killed process 12345 (trading-api) total-vm:524288kB
```

---

### Experiment 04: CPU Throttling
```bash
./04-cpu-throttling.sh
```

**Simulates**: CPU contention, noisy neighbor, resource constraints

**Steps**:
1. Measure baseline CPU usage and request latency
2. Start BPF CPU profiling
3. Reduce CPU quota to 50% of 1 core (0.5 CPUs)
4. Start order flow to create CPU demand
5. Monitor latency increase over time
6. Restore original CPU allocation

**Expected Observations**:
- Request latency increases 3-5x
- No errors, just slow responses
- Queue buildup visible
- CPU scheduler delays visible in BPF
- Immediate improvement when CPU restored

**BPF Metrics**:
- CPU samples per process
- Time spent waiting for CPU (runqueue delay)
- Context switches per second
- Hot functions consuming CPU

**Key Insight**: Demonstrates graceful degradation vs hard failure.

---

### Experiment 05: Disk I/O Starvation
```bash
sudo ./05-disk-io-starvation.sh
```

**Simulates**: Disk failure, I/O contention, slow storage

**Steps**:
1. Monitor baseline disk I/O latency
2. Start BPF I/O monitoring (vfs_read, vfs_write)
3. Create I/O contention with stress-ng
4. Start order flow to generate database writes
5. Monitor disk latency and query timeouts
6. Stop stress test
7. Measure recovery

**Expected Observations**:
- Disk latency spikes from <1ms to 100ms+
- Database checkpoint delays
- Query timeouts increase
- Write throughput drops
- PostgreSQL may go into recovery mode

**BPF Metrics**:
- I/O latency histogram (read vs write)
- I/O queue depth
- Bytes read/written per process
- Filesystem cache hit rate

**Why Critical**: Database performance is directly tied to disk I/O.

---

### Experiment 06: Connection Pool Exhaustion
```bash
sudo ./06-connection-pool-exhaustion.sh
```

**Simulates**: Connection leak, max_connections reached, pool misconfiguration

**Steps**:
1. Check current max_connections limit
2. Reduce PostgreSQL max_connections to 20
3. Start BPF monitoring for connection attempts
4. Start heavy order flow to consume connections
5. Monitor pool saturation
6. Observe "too many connections" errors
7. Restore original max_connections

**Expected Observations**:
- Connections grow from 5 → 20 (limit reached)
- New requests queue/wait indefinitely
- Error: "FATAL: sorry, too many clients already"
- Dashboard stops updating
- Immediate recovery when limit increased

**BPF Metrics**:
- Connection attempts per second
- Failed connections (ECONNREFUSED)
- Time waiting for available connection
- Connection churn rate

**Production Impact**: This is a common failure mode in high-traffic systems.

---

### Experiment 07: Redis Cache Failure
```bash
./07-redis-cache-failure.sh
```

**Simulates**: Cache service outage, Redis crash, memory eviction

**Steps**:
1. Monitor Redis operations (reads/writes to port 6379)
2. Capture baseline metrics (with cache hits)
3. Kill Redis container
4. Observe database load increase
5. Monitor latency impact
6. Restart Redis and observe recovery

**Expected Observations**:
- Cache miss rate jumps to 100%
- Database query load increases
- Latency increases for cached endpoints
- Event publishing fails
- Services remain functional (degraded mode)

**BPF Metrics**:
- Redis connection attempts after failure
- Database query count increase
- Network traffic shift from Redis to Postgres
- TCP connection state changes

**Key Learning**: Demonstrates cache-aside pattern resilience.

---

### Experiment 08: Nginx Reverse Proxy Failure
```bash
./08-nginx-failure.sh
```

**Simulates**: Load balancer failure, proxy misconfiguration

**Steps**:
1. Test dashboard accessibility through nginx
2. Kill nginx container
3. Attempt dashboard access (should fail)
4. Test direct service access on port 8080 (should work)
5. Restart nginx
6. Verify recovery

**Expected Observations**:
- Dashboard (port 80) becomes unreachable
- Direct API (port 8080) still works
- Rate limiting bypassed
- No service-to-service impact
- Instant recovery when nginx restarts

**BPF Metrics**:
- TCP RST packets on port 80
- Connection refused errors
- Traffic to backend port 8080 continues

**Why Test**: Validates frontend/backend separation.

---

### Experiment 09: Network Latency Injection
```bash
sudo ./09-network-latency.sh
```

**Simulates**: WAN degradation, network congestion, geographically distributed services

**Steps**:
1. Measure baseline ping and request latency
2. Add 50ms delay with tc netem
3. Observe impact, then increase to 100ms
4. Continue increasing to 200ms (timeout threshold)
5. Monitor retransmissions and failures
6. Remove delay and measure recovery

**Expected Observations**:
- Latency increases gradually
- At 50ms: Noticeable slowdown
- At 100ms: Significant impact
- At 200ms: Timeouts begin
- Retransmissions visible in BPF
- Immediate recovery when delay removed

**BPF Metrics**:
- TCP RTT distribution shift
- Retransmission rate increase
- Application timeout events
- Queue depth in kernel network stack

**Use Case**: Validates timeout configuration and retry behavior.

---

### Experiment 10: Multi-Service Cascade Failure
```bash
./10-multi-service-failure.sh
```

**Simulates**: Datacenter outage, rolling update gone wrong, correlated failures

**Steps**:
1. Capture all service states
2. Kill postgres, redis, trading-api in rapid succession
3. Monitor restart orchestration
4. Measure dependency startup ordering
5. Calculate total recovery time
6. Verify all services healthy

**Expected Observations**:
- Services fail in dependency order
- Health checks fail sequentially
- Auto-restart respects dependencies
- Full recovery in 30-60 seconds
- No manual intervention needed

**BPF Metrics**:
- Process termination sequence
- Connection re-establishment patterns
- Service discovery attempts
- Time to first successful health check per service

**Why Most Complex**: Tests entire system resilience and recovery orchestration.

---

## How to Use This Guide

### For Your First Experiment
1. Start with **Experiment 01** (easiest, no root required)
2. Read the script comments to understand each step
3. Run it: `./01-database-sudden-death.sh`
4. Review the generated log files
5. Examine BPF traces for low-level insights

### For Comprehensive Testing
```bash
# Run all experiments in recommended order
./01-database-sudden-death.sh      # Warm up, learn the patterns
./07-redis-cache-failure.sh        # Simple, no root
./08-nginx-failure.sh              # Quick test
./04-cpu-throttling.sh             # Performance testing
sudo ./09-network-latency.sh       # Network behavior
sudo ./06-connection-pool-exhaustion.sh  # Resource limits
sudo ./02-network-partition.sh     # Network failures
sudo ./05-disk-io-starvation.sh    # Infrastructure
sudo ./03-memory-oom-kill.sh       # Kernel interactions
./10-multi-service-failure.sh      # Grand finale
```

### For Production Readiness Assessment
Focus on these experiments:
- **02**: Network failures are common
- **03**: Memory leaks happen
- **06**: Connection pools are finite
- **09**: Latency is variable
- **10**: Cascades must be handled

---

## Interpreting Results

### Success Criteria
✅ **Service recovers automatically** - No manual intervention needed
✅ **Metrics show degradation, not crash** - Graceful fallback to cached data
✅ **Recovery time < 60 seconds** - Acceptable downtime
✅ **No data loss** - State is preserved or recoverable
✅ **Clear error messages** - Observable what went wrong

### Red Flags
🚨 **Service doesn't restart** - Health checks may be broken
🚨 **Data corruption** - State management issues
🚨 **Infinite retry loops** - Backoff not implemented
🚨 **No error visibility** - Silent failures
🚨 **Recovery time > 5 minutes** - Dependencies or timeouts misconfigured

---

## Expected BPF Insights Summary

| Experiment | Key BPF Observation | What It Reveals |
|-----------|---------------------|-----------------|
| 01 | Connection retry count | Exponential backoff working correctly |
| 02 | TCP retransmissions | Network resilience patterns |
| 03 | Page allocation delta | Memory leak detection |
| 04 | CPU runqueue time | Scheduler contention |
| 05 | I/O latency histogram | Disk bottlenecks |
| 06 | Connection wait time | Pool sizing adequacy |
| 07 | Cache miss increase | Cache dependency level |
| 08 | Port 80 vs 8080 traffic | Load balancer necessity |
| 09 | RTT distribution | Latency sensitivity |
| 10 | Restart sequence timing | Dependency orchestration |

---

## Next Actions After Running Experiments

1. **Review all log files** in `/tmp/chaos-exp-*-*.log`
2. **Analyze BPF traces** for unexpected patterns
3. **Identify improvements**:
   - Circuit breakers needed?
   - Timeout values appropriate?
   - Resource limits sufficient?
   - Retry logic correct?
4. **Set up monitoring** for conditions observed
5. **Create runbooks** for recovery procedures
6. **Schedule regular chaos** to prevent regression

---

## Troubleshooting Guide

### "Permission denied" errors
→ Run with `sudo` (experiments 2, 3, 5, 6, 9 require root)

### "bpftrace not found"
→ Install: `sudo dnf install bpftrace`
→ Experiments still run without BPF (reduced visibility)

### Services don't recover
→ Check: `docker compose ps`
→ Restart: `docker compose restart`
→ Logs: `docker logs simulated-exchange-trading-api`

### Experiment hangs
→ Press Ctrl+C (cleanup runs automatically)
→ Manually reset: `docker compose restart`

---

## Files Reference

Each experiment generates these files:

```
/tmp/chaos-exp-01-20251006-143015.log       # Main experiment log
/tmp/chaos-exp-01-20251006-143015.log.bpf   # BPF trace output
/tmp/chaos-exp-01-20251006-143015.log.oom   # OOM kernel logs (exp 03 only)
```

Log file format:
- ✓ Green = Success
- ⚠ Yellow = Warning
- ✗ Red = Error
- Blue = Info

---

## Production Considerations

⚠️ **DO NOT run these experiments in production** without:
1. Maintenance window scheduled
2. Stakeholder approval
3. Rollback plan ready
4. Monitoring alerts suppressed
5. Backup of current state

✅ **Safe to run in**:
- Development environments
- Test environments
- Staging (with approval)
- Isolated QA systems

---

## Learning Path

**Week 1**: Run experiments 1, 7, 8 (simple, no root)
**Week 2**: Run experiments 4, 6 (resource limits)
**Week 3**: Run experiments 2, 9 (network)
**Week 4**: Run experiments 3, 5, 10 (advanced)

By end of month: Complete understanding of system failure modes and resilience patterns.

---

## Questions?

- Review experiment script comments
- Check BPF traces for low-level details
- Consult README.md for detailed info
- Reference Principles of Chaos Engineering

**Remember**: The goal is not to break things, but to **learn how they break** and **verify they recover**.
