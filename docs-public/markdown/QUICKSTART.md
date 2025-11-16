# Chaos Engineering Quick Start Guide

## 🎯 Goal
Test the resilience of your trading exchange system using production-ready chaos engineering experiments with BPF observability.

## ⚡ 5-Minute Quick Start

### 1. Verify Prerequisites
```bash
# Check you're in the right directory
cd /home/acthomas/1_projects/simulated_exchange/chaos-experiments

# Verify services are running
docker compose ps

# Should see: postgres, redis, trading-api, market-simulator, nginx all "Up"
```

### 2. Run Your First Experiment (No Root Required)
```bash
# Database Sudden Death - easiest experiment to start with
./01-database-sudden-death.sh
```

**What happens**:
1. Script captures baseline metrics
2. Kills PostgreSQL database
3. Observes trading-api behavior
4. PostgreSQL auto-restarts
5. Measures recovery time
6. Generates detailed report

**Expected output**:
```
========================================================================
CHAOS EXPERIMENT: Database Sudden Death
========================================================================

[2025-10-06 14:30:15] ✓ Docker found
[2025-10-06 14:30:15] ✓ PostgreSQL container running
...
[2025-10-06 14:35:22] ✓ Experiment completed successfully
```

### 3. Review Results
```bash
# View the most recent experiment log
ls -lt /tmp/chaos-exp-*.log | head -1 | awk '{print $NF}' | xargs cat

# Or just the summary
ls -lt /tmp/chaos-exp-*.log | head -1 | awk '{print $NF}' | xargs tail -30
```

---

## 📊 What Each Experiment Tests

| When to Use | Experiment | What You Learn |
|------------|------------|----------------|
| **START HERE** | 01-database-sudden-death.sh | Connection resilience, retry logic |
| Database issues | 06-connection-pool-exhaustion.sh | Connection limits |
| Network problems | 02-network-partition.sh | Network failure handling |
| Slow performance | 04-cpu-throttling.sh | CPU constraints impact |
| Memory issues | 03-memory-oom-kill.sh | Memory limits, OOM killer |
| Cache failures | 07-redis-cache-failure.sh | Cache fallback behavior |
| Everything fails | 10-multi-service-failure.sh | Full system recovery |

---

## 🔧 System Requirements

### Minimal (Required)
- Docker & Docker Compose running
- Services: postgres, trading-api, redis all healthy
- Disk space: 500MB for logs

### Optimal (For Full BPF Insights)
- Root/sudo access
- bpftrace installed: `sudo dnf install bpftrace`
- jq installed: `sudo dnf install jq`
- Kernel 4.9+ (you have 5.14 ✓)

**Check your system**:
```bash
# Quick system check
echo "Docker: $(docker --version)"
echo "Compose: $(docker compose version)"
echo "bpftrace: $(bpftrace --version 2>/dev/null || echo 'NOT INSTALLED (optional)')"
echo "jq: $(jq --version 2>/dev/null || echo 'NOT INSTALLED (optional)')"
echo "Kernel: $(uname -r)"
```

---

## 🚀 Recommended Learning Path

### Day 1: Get Familiar (30 minutes)
```bash
# Run the easiest 3 experiments
./01-database-sudden-death.sh    # 5 min
./07-redis-cache-failure.sh      # 5 min  (create this simple one)
# Review logs and understand the pattern
```

### Day 2: Test Resources (45 minutes)
```bash
# Test resource limits
./04-cpu-throttling.sh                    # 10 min (create this)
sudo ./03-memory-oom-kill.sh              # 15 min
sudo ./06-connection-pool-exhaustion.sh   # 10 min (create this)
```

### Day 3: Network Chaos (60 minutes)
```bash
# Test network resilience
sudo ./02-network-partition.sh      # 10 min
sudo ./09-network-latency.sh        # 15 min (create this)
```

### Day 4: Advanced (90 minutes)
```bash
# Complex scenarios
sudo ./05-disk-io-starvation.sh     # 15 min (create this)
./10-multi-service-failure.sh       # 10 min (create this)

# Analyze all BPF traces
find /tmp -name "chaos-exp-*.log.bpf" -mtime -1 -exec echo "=== {} ===" \; -exec tail -20 {} \;
```

---

## 📁 Output Files Explained

After running an experiment, you'll find:

```
/tmp/chaos-exp-01-20251006-143015.log       # Main log (human-readable)
/tmp/chaos-exp-01-20251006-143015.log.bpf   # BPF trace (system-level)
```

### Reading the Main Log
```bash
# Success indicators (look for ✓)
grep "✓" /tmp/chaos-exp-*.log | tail -10

# Warnings (look for ⚠)
grep "⚠" /tmp/chaos-exp-*.log | tail -10

# Errors (look for ✗)
grep "✗" /tmp/chaos-exp-*.log | tail -10

# Summary report (always at the end)
tail -50 /tmp/chaos-exp-*.log
```

### Reading BPF Traces
```bash
# Connection attempts
grep "CONN_ATTEMPT" /tmp/chaos-exp-*.log.bpf

# Failures
grep "FAILED\|ERROR" /tmp/chaos-exp-*.log.bpf

# Statistics
grep "Statistics" -A 10 /tmp/chaos-exp-*.log.bpf
```

---

## 🎓 Understanding BPF Output

BPF (Berkeley Packet Filter) gives you kernel-level visibility. Here's what to look for:

### Example: Database Experiment BPF Output
```
[14:30:45] CONN_ATTEMPT       trading-api (PID: 1234)
[14:30:45] CONN_FAILED        trading-api (PID: 1234, errno: 111)
[14:30:46] TCP_RETRANSMIT     trading-api
[14:30:47] CONN_ATTEMPT       trading-api (PID: 1234)
[14:30:47] CONN_FAILED        trading-api (PID: 1234, errno: 111)
```

**What this tells you**:
- errno 111 = ECONNREFUSED (database is down)
- Retry happening every ~1 second
- TCP retransmissions occurring
- This is NORMAL and EXPECTED behavior

### Common Error Codes
- **111**: ECONNREFUSED - Service not listening
- **110**: ETIMEDOUT - Connection timed out
- **113**: EHOSTUNREACH - No route to host
- **104**: ECONNRESET - Connection reset by peer

---

## 🔍 What "Good" Looks Like

### Successful Experiment Output
```
KEY FINDINGS:
1. Database failure detection: Immediate (< 1s)          ✓
2. Service degradation: Graceful fallback to cache       ✓
3. Error handling: Clear error messages logged           ✓
4. Recovery time: 15s from restart to operational        ✓
5. No manual intervention required                       ✓
```

### Red Flags to Investigate
```
KEY FINDINGS:
1. Database failure detection: 30s (slow!)               🚨
2. Service crashed instead of degrading                  🚨
3. No error logging (silent failure)                     🚨
4. Recovery time: 5 minutes (too long)                   🚨
5. Required manual restart                               🚨
```

---

## 🛠️ Troubleshooting

### Experiment Fails to Start
```bash
# Check all services are running
docker compose ps

# If any are down, restart
docker compose restart

# Ensure order-flow-simulator is STOPPED
docker compose stop order-flow-simulator
```

### "Permission Denied" Errors
```bash
# Some experiments need root
sudo ./02-network-partition.sh

# Check which ones need root
grep "Root Required: Yes" README.md
```

### "bpftrace: command not found"
```bash
# Install bpftrace
sudo dnf install bpftrace

# Or continue without it (experiments still work, less visibility)
./01-database-sudden-death.sh  # Will warn but continue
```

### Services Don't Recover
```bash
# Manual recovery
docker compose restart postgres redis trading-api

# Check logs
docker logs simulated-exchange-trading-api --tail 50
```

### Experiment Hangs
```bash
# Press Ctrl+C (cleanup runs automatically)
# Then manually verify:
docker compose ps   # All should be "Up"
```

---

## 🎯 Real-World Scenarios

### Scenario 1: "Our database keeps going down in production"
**Run**: `./01-database-sudden-death.sh`
**Learn**: How fast do we detect? How do we retry? Do we fallback gracefully?

### Scenario 2: "Orders are timing out during peak hours"
**Run**: `sudo ./09-network-latency.sh`
**Learn**: What latency causes timeouts? Are our timeouts tuned correctly?

### Scenario 3: "Trading API keeps crashing with OOM"
**Run**: `sudo ./03-memory-oom-kill.sh`
**Learn**: When do we hit OOM? Is our memory limit appropriate? Do we leak?

### Scenario 4: "Everything crashed during a deployment"
**Run**: `./10-multi-service-failure.sh`
**Learn**: Do services restart in correct order? How long is total recovery?

---

## 📈 Measuring Success

After each experiment, ask:

1. **Did the service recover automatically?** (Should be YES)
2. **How long did recovery take?** (Should be < 60s)
3. **Was there data loss?** (Should be NO)
4. **Did users see errors or just slowness?** (Slowness is better)
5. **Are error messages clear?** (Should be YES)

---

## 🔄 Next Steps After Running Experiments

1. **Review all experiment logs**
   ```bash
   ls -lt /tmp/chaos-exp-*.log | head -10
   ```

2. **Identify improvement areas**
   - Are timeouts too aggressive?
   - Is fallback behavior correct?
   - Do we need circuit breakers?

3. **Update monitoring**
   - Alert on conditions observed in experiments
   - Dashboard metrics for failure scenarios

4. **Create runbooks**
   - Document recovery procedures
   - Note expected vs unexpected behaviors

5. **Schedule regular chaos**
   - Weekly: Run 1-2 experiments
   - Monthly: Full suite
   - Pre-deployment: Verify resilience

---

## 📚 Additional Resources

- **README.md**: Full experiment descriptions
- **EXPERIMENTS_SUMMARY.md**: Detailed reference guide
- **Individual scripts**: Read the header comments for specifics

### Getting Help

1. Check script comments: `head -50 01-database-sudden-death.sh`
2. Review experiment logs: `cat /tmp/chaos-exp-*.log`
3. Examine BPF traces: `cat /tmp/chaos-exp-*.log.bpf`

---

## 🎉 You're Ready!

Start with:
```bash
./01-database-sudden-death.sh
```

Then explore the others based on your interests and production issues.

Remember: **The goal is to learn, not to break things.** These experiments help you understand how your system behaves under failure so you can make it more resilient.

Happy chaos engineering! 🔥
