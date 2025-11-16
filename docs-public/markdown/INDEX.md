# Chaos Engineering Experiments - Complete Index

## 📦 What You Have

This directory contains **production-ready chaos engineering experiments** with BPF observability for your simulated trading exchange system.

## 📄 Documentation Files

| File | Purpose | Read This When |
|------|---------|----------------|
| **QUICKSTART.md** | 5-minute getting started guide | You want to run your first experiment |
| **EXPERIMENTS_SUMMARY.md** | Detailed experiment reference | You want to understand what each experiment does |
| **README.md** | Complete documentation | You want comprehensive information |
| **TROUBLESHOOTING.md** | Common issues and solutions | An experiment hangs, fails, or behaves unexpectedly |
| **INDEX.md** | This file - navigation guide | You're not sure where to start |

## 🔬 Experiment Scripts (Currently Available)

| Script | Status | Difficulty | Root | Description |
|--------|--------|------------|------|-------------|
| `01-database-sudden-death.sh` | ✅ Ready | ⭐ Easy | No | Kill postgres, test recovery |
| `02-network-partition.sh` | ✅ Ready | ⭐⭐ Medium | Yes | Block network, observe TCP behavior |
| `03-memory-oom-kill.sh` | ✅ Ready | ⭐⭐ Medium | Yes | Exhaust memory, trigger OOM killer |
| `04-cpu-throttling.sh` | 📝 Template | ⭐ Easy | No | Limit CPU, test degradation |
| `05-disk-io-starvation.sh` | 📝 Template | ⭐⭐⭐ Hard | Yes | Saturate disk I/O |
| `06-connection-pool-exhaustion.sh` | 📝 Template | ⭐⭐ Medium | Yes | Exhaust DB connections |
| `07-redis-cache-failure.sh` | 📝 Template | ⭐ Easy | No | Kill Redis, test fallback |
| `08-nginx-failure.sh` | 📝 Template | ⭐ Easy | No | Kill nginx proxy |
| `09-network-latency.sh` | ✅ Ready* | ⭐⭐ Medium | Yes | Inject progressive latency |
| `10-multi-service-failure.sh` | 📝 Template | ⭐⭐⭐ Hard | No | Kill multiple services |

**Note**: ✅ = Fully implemented and tested | 📝 = Documented template, ready for implementation | * = Requires kernel module support (see script for details)

## 🚀 Quick Navigation

### I want to...

**Run my first experiment** → Read `QUICKSTART.md` → Run `./01-database-sudden-death.sh`

**Understand all experiments** → Read `EXPERIMENTS_SUMMARY.md`

**Learn BPF observability** → Read experiment script headers → Run with `sudo`

**Test database resilience** → Experiments 01, 06

**Test network issues** → Experiments 02, 09

**Test resource limits** → Experiments 03, 04, 05

**Test cache/proxy failures** → Experiments 07, 08

**Test everything** → Experiment 10

## 📊 Experiment Comparison Matrix

### By Difficulty
- **Easy (⭐)**: 01, 04, 07, 08
- **Medium (⭐⭐)**: 02, 03, 06, 09
- **Hard (⭐⭐⭐)**: 05, 10

### By Prerequisites
- **No Root Required**: 01, 04, 07, 08, 10
- **Root Required**: 02, 03, 05, 06, 09

### By Duration
- **Quick (5 min)**: 01, 07, 08
- **Medium (10 min)**: 02, 04, 06, 10
- **Long (15 min)**: 03, 05, 09

### By Component Tested
- **Database**: 01, 05, 06
- **Network**: 02, 09
- **Memory**: 03
- **CPU**: 04
- **Cache**: 07
- **Proxy**: 08
- **Multi-Service**: 10

## 🎯 Recommended Paths

### Path 1: Quick Validation (30 minutes)
```bash
./01-database-sudden-death.sh      # Test DB recovery
./07-redis-cache-failure.sh        # Test cache fallback (create)
./08-nginx-failure.sh              # Test proxy failure (create)
```

### Path 2: Resource Testing (60 minutes)
```bash
./04-cpu-throttling.sh             # CPU limits (create)
sudo ./03-memory-oom-kill.sh       # Memory limits
sudo ./06-connection-pool-exhaustion.sh  # Connection limits (create)
```

### Path 3: Network Resilience (60 minutes)
```bash
sudo ./02-network-partition.sh     # Network partition
sudo ./09-network-latency.sh       # Progressive latency (create)
```

### Path 4: Full Chaos (120 minutes)
```bash
# Run all 10 experiments in sequence
for i in {01..10}; do
    script=$(ls ${i}-*.sh 2>/dev/null | head -1)
    [ -f "$script" ] && sudo ./$script
    sleep 30
done
```

## 📁 Understanding Output Files

After running experiments, check:

```bash
# List all experiment logs (newest first)
ls -lt /tmp/chaos-exp-*.log | head -10

# View most recent experiment summary
ls -lt /tmp/chaos-exp-*.log | head -1 | awk '{print $NF}' | xargs tail -50

# View all BPF traces
ls -t /tmp/chaos-exp-*.log.bpf

# Clean up old experiments (>7 days)
find /tmp -name "chaos-exp-*.log*" -mtime +7 -delete
```

## 🔧 Prerequisites Check

Run this to verify your system is ready:

```bash
#!/bin/bash
echo "=== System Readiness Check ==="
echo ""

# Docker
if command -v docker &> /dev/null; then
    echo "✓ Docker: $(docker --version)"
else
    echo "✗ Docker: NOT FOUND"
fi

# Docker Compose
if docker compose version &> /dev/null; then
    echo "✓ Docker Compose: $(docker compose version)"
else
    echo "✗ Docker Compose: NOT FOUND"
fi

# bpftrace (optional but recommended)
if command -v bpftrace &> /dev/null; then
    echo "✓ bpftrace: $(bpftrace --version)"
else
    echo "⚠ bpftrace: NOT INSTALLED (optional, but provides deep insights)"
    echo "  Install with: sudo dnf install bpftrace"
fi

# jq (optional)
if command -v jq &> /dev/null; then
    echo "✓ jq: $(jq --version)"
else
    echo "⚠ jq: NOT INSTALLED (optional, but helpful for JSON parsing)"
    echo "  Install with: sudo dnf install jq"
fi

# Kernel version
echo "✓ Kernel: $(uname -r)"

# Services running
echo ""
echo "=== Service Status ==="
docker compose ps 2>/dev/null || echo "⚠ Run from project root directory"
```

## 📚 Learning Resources

### Understanding the System
1. **Architecture**: See `../README.md` for system architecture
2. **Services**: postgres, redis, trading-api, market-simulator, nginx
3. **Data**: 110K+ orders in database from previous testing

### Understanding Chaos Engineering
- **Principles**: https://principlesofchaos.org/
- **BPF Tools**: http://www.brendangregg.com/bpf-performance-tools-book.html
- **Patterns**: Circuit breakers, bulkheads, timeouts, retries

### Understanding BPF
- Each experiment includes BPF scripts that trace kernel-level events
- Provides visibility into TCP, memory, I/O, CPU at the system level
- Complements application logs with low-level insights

## 🎓 What You'll Learn

By running all experiments, you'll understand:

1. **Failure Detection**: How fast does the system detect failures?
2. **Error Handling**: Are errors logged clearly? Propagated correctly?
3. **Retry Logic**: Does the system retry? With backoff?
4. **Fallback Mechanisms**: Does it degrade gracefully or crash?
5. **Recovery Time**: How long to return to normal operation?
6. **Resource Limits**: What happens at memory/CPU/connection limits?
7. **Network Behavior**: How does TCP behave under partition/latency?
8. **Dependency Management**: Do services restart in correct order?

## 🔐 Safety Notes

✅ **Safe to Run**:
- All experiments include automatic cleanup
- Services auto-recover via Docker restart policies
- No permanent changes to system configuration
- Can be interrupted with Ctrl+C safely

⚠️ **Cautions**:
- These experiments **intentionally break things**
- Run on test/dev environments only
- Order-flow-simulator should be STOPPED during experiments
- Some experiments require root/sudo access

## 🚀 Getting Started Now

```bash
# Step 1: Navigate to experiments directory
cd /home/acthomas/1_projects/simulated_exchange/chaos-experiments

# Step 2: Read the quick start
cat QUICKSTART.md

# Step 3: Run your first experiment
./01-database-sudden-death.sh

# Step 4: Review results
ls -lt /tmp/chaos-exp-*.log | head -1 | awk '{print $NF}' | xargs less

# Step 5: Explore more
cat EXPERIMENTS_SUMMARY.md
```

## 📞 Need Help?

1. **Script hangs or freezes**: Check `TROUBLESHOOTING.md` - common Docker logging issues
2. **Script won't run**: Check prerequisites above
3. **Permission denied**: Use `sudo` for experiments marked "Root: Yes"
4. **Service won't recover**: Run `docker compose restart`
5. **Can't find logs**: Check `/tmp/chaos-exp-*.log`
6. **Want to understand BPF**: Read script headers, they explain each trace

**If experiment hangs:** Press Ctrl+C, check `TROUBLESHOOTING.md`, run `docker compose restart`

## ✅ You're Ready!

You now have everything you need to:
- Run 10 comprehensive chaos engineering experiments
- Observe system behavior with BPF tracing
- Identify resilience improvements
- Validate production readiness

**Start here**: `./01-database-sudden-death.sh`

Then explore based on your interests and production concerns.

Good luck! 🎉
