# 🚀 SRE Team Member Onboarding - Learning Lab

## Welcome to Your Safe Learning Environment!

This guide gets you up and running with the Simulated Exchange learning lab in **30 minutes**. By the end, you'll have broken and recovered a trading system safely!

## What is This?

A **safe sandbox** where you can:
- 💥 Break things without consequences
- 🔍 Learn eBPF and profiling tools
- 🎯 Practice chaos engineering
- 📊 Master observability

**Important:** This complements (doesn't replace) our 17-server test environment. Learn here first, apply there later.

## Quick Start (30 Minutes)

### Prerequisites

- Docker installed and running
- 8GB+ RAM available
- Terminal access
- Coffee ☕ (optional but recommended)

### Step 1: Get It Running (5 minutes)

```bash
# Clone the repo
cd ~/projects  # or wherever you keep code
git clone <repo-url> simulated_exchange
cd simulated_exchange

# Start everything
make docker-up

# Verify it's running
docker ps  # Should see 6+ containers

# Check health
curl http://localhost:8080/health
```

**Expected output:**
```json
{"status":"healthy","timestamp":"...","version":"1.0.0"}
```

### Step 2: Explore the Dashboard (5 minutes)

Open in your browser:
- **Trading Dashboard:** http://localhost
- **Grafana:** http://localhost:3000 (admin/admin123)
- **Prometheus:** http://localhost:9090

**What to look at:**
1. Grafana → "Trading Exchange Overview" dashboard
2. Watch orders/sec, latency, throughput
3. See real-time metrics updating

### Step 3: Generate Some Load (5 minutes)

```bash
# Light load test
make load-test-light

# Watch the metrics in Grafana
# You should see:
# - Orders/sec increase to ~25
# - Latency staying under 50ms
# - Throughput climbing
```

### Step 4: Break Something! (5 minutes)

**Your first chaos experiment:**

```bash
# Kill the trading API
./chaos-experiments/01-service-failure.sh

# Watch what happens in Grafana:
# - Error rate spikes
# - Order processing stops
# - Service restarts automatically
# - System recovers
```

**This is your first lesson:** Systems fail, and that's okay if they recover well.

### Step 5: Profile Performance (5 minutes)

```bash
# Generate a CPU flamegraph
./scripts/generate-flamegraph.sh cpu 30 8080

# This creates:
# - flamegraphs/cpu_TIMESTAMP.svg (interactive visualization)
# - flamegraphs/cpu_profile_TIMESTAMP.pb_analysis.md (AI analysis)

# View the analysis
cat flamegraphs/cpu_profile_*_analysis.md | head -50
```

**What to look for:**
- Top CPU hotspots
- Optimization recommendations
- System call overhead

### Step 6: Access Web Documentation (5 minutes)

If running with Caddy (external access enabled):
- **Docs:** https://docs.andythomas-sre.com
- **Flamegraphs:** https://flamegraph.andythomas-sre.com

Or locally browse:
```bash
firefox docs-public/index.html
```

**Explore:**
- SRE Runbook
- Chaos Experiments Guide
- Flamegraph Analysis Center

## Your First Week Learning Plan

### Day 1: Get Comfortable

**Morning (2 hours):**
```bash
# 1. Read the learning lab purpose
cat docs/LEARNING_LAB_PURPOSE.md

# 2. Explore all dashboards
#    - Trading dashboard
#    - Grafana metrics
#    - Prometheus targets

# 3. Run each chaos experiment once
ls chaos-experiments/*.sh
./chaos-experiments/01-service-failure.sh
# ... run them all
```

**Afternoon (2 hours):**
- Read [SRE_RUNBOOK.md](SRE_RUNBOOK.md)
- Create your own notes document
- Document what you learned

### Day 2: Metrics & Observability

**Tasks:**
1. Create a custom Grafana dashboard
2. Write 5 useful Prometheus queries
3. Generate flamegraphs for each service:
   ```bash
   ./scripts/generate-flamegraph.sh cpu 30 8080  # trading-api
   ./scripts/generate-flamegraph.sh cpu 30 8081  # market-simulator
   ./scripts/generate-flamegraph.sh cpu 30 8082  # order-flow
   ```
4. Compare the profiles - what's different?

### Day 3: Chaos Engineering

**Morning:**
- Run each chaos experiment
- Document observed behavior
- Note recovery time for each

**Afternoon:**
- Design your own chaos scenario
- Implement it as a new script
- Test your hypothesis

**Example custom chaos:**
```bash
# Create: chaos-experiments/custom/memory-leak.sh
#!/bin/bash
# Simulate memory leak by running memory-hungry container
docker run --rm -d --name memory-hog \
  --network simulated_exchange_trading_network \
  progrium/stress --vm 1 --vm-bytes 512M --timeout 60s
```

### Day 4: Performance Analysis

**Goals:**
1. Establish baseline performance
2. Run load test while profiling
3. Identify top 3 bottlenecks
4. Document findings

```bash
# Baseline
make load-test-light &
./scripts/generate-flamegraph.sh cpu 30 8080

# Medium load
make load-test-medium &
./scripts/generate-flamegraph.sh cpu 30 8080

# Compare the two profiles
```

### Day 5: Deep Dive - Pick One

Choose your adventure:

**Option A: eBPF Exploration**
```bash
# Learn eBPF tools
sudo apt-get install bpfcc-tools  # if not installed

# Trace system calls
sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* { @[probe] = count(); }'

# Network latency
sudo tcplife

# Block I/O
sudo biolatency
```

**Option B: Database Optimization**
```bash
# Connect to PostgreSQL
docker exec -it simulated-exchange-postgres psql -U trading_user -d trading_db

# Analyze queries
EXPLAIN ANALYZE SELECT * FROM orders WHERE status = 'pending';

# Check indexes
\d orders

# Monitor query performance
SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;
```

**Option C: Custom Monitoring**
```bash
# Create custom metric exporter
# Build dashboard
# Set up alerts
```

## Common Tasks Cheat Sheet

### Starting & Stopping

```bash
# Start everything
make docker-up

# Stop everything
make docker-down

# Restart single service
docker restart simulated-exchange-trading-api

# View logs
docker logs -f simulated-exchange-trading-api

# Completely reset
make clean-docker
make docker-up
```

### Load Testing

```bash
# Light (25 orders/sec)
make load-test-light

# Medium (100 orders/sec)
make load-test-medium

# Custom load
./scripts/load-generator.sh --rate 50 --duration 60
```

### Profiling

```bash
# CPU profile (30 seconds)
./scripts/generate-flamegraph.sh cpu 30 8080

# Heap profile (snapshot)
./scripts/generate-flamegraph.sh heap 0 8080

# Goroutine profile
./scripts/generate-flamegraph.sh goroutine 0 8080

# Interactive pprof
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
```

### Chaos Engineering

```bash
# List all experiments
ls chaos-experiments/*.sh

# Run specific experiment
./chaos-experiments/04-cpu-throttling.sh

# Create custom experiment
cp chaos-experiments/01-service-failure.sh chaos-experiments/custom/my-chaos.sh
# Edit and run
```

### Monitoring

```bash
# Check service health
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health

# Get metrics
curl http://localhost:8080/metrics

# Query Prometheus
curl 'http://localhost:9090/api/v1/query?query=trading_orders_total'
```

## Learning Resources

### In This Repository

| Document | Purpose | When to Read |
|----------|---------|--------------|
| [LEARNING_LAB_PURPOSE.md](LEARNING_LAB_PURPOSE.md) | Understand the "why" | Day 1, first thing |
| [README.md](README.md) | Quick start & overview | Day 1 |
| [SRE_RUNBOOK.md](SRE_RUNBOOK.md) | Complete operational guide | Day 2-3 |
| [SRE_DEMO_GUIDE.md](SRE_DEMO_GUIDE.md) | Demo scenarios | Week 2 |
| [FLAMEGRAPH_GUIDE.md](FLAMEGRAPH_GUIDE.md) | Performance profiling | Week 2 |
| [ARCHITECTURE.md](ARCHITECTURE.md) | System design | As needed |

### External Resources

**Chaos Engineering:**
- [Principles of Chaos](https://principlesofchaos.org/)
- [Chaos Engineering Book](https://www.oreilly.com/library/view/chaos-engineering/9781491988459/)

**eBPF:**
- [BPF Performance Tools](http://www.brendangregg.com/bpf-performance-tools-book.html)
- [eBPF.io](https://ebpf.io/)

**Observability:**
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [Grafana Tutorials](https://grafana.com/tutorials/)

**Performance:**
- [Brendan Gregg's Blog](http://www.brendangregg.com/blog/index.html)
- [Go Performance](https://go.dev/doc/diagnostics)

## Troubleshooting

### Port Already in Use

```bash
# Find what's using the port
sudo lsof -i :8080

# Kill it or change port
export SERVER_PORT=8081
make docker-up
```

### Out of Memory

```bash
# Stop some services
docker stop simulated-exchange-order-flow-simulator

# Or increase Docker memory
# Docker Desktop → Settings → Resources → Memory → 8GB+
```

### Container Won't Start

```bash
# Check logs
docker logs simulated-exchange-trading-api

# Common fixes:
1. Check port conflicts
2. Verify dependencies started (postgres, redis)
3. Check disk space
4. Restart Docker
```

### Metrics Not Showing

```bash
# Verify Prometheus is scraping
http://localhost:9090/targets

# Check service /metrics endpoint
curl http://localhost:8080/metrics

# Restart Prometheus
docker restart simulated-exchange-prometheus
```

## Next Steps

After your first week:

1. **Share your learnings**
   - Write up interesting findings
   - Demo to team what you learned
   - Add to team wiki

2. **Contribute back**
   - Create new chaos experiments
   - Improve documentation
   - Add useful scripts

3. **Apply to real systems**
   - Review learnings with senior SRE
   - Plan test environment validation
   - Document production readiness

## Getting Help

- **Documentation Issues:** Create issue in repo
- **Learning Questions:** Ask in SRE channel
- **Found a Bug:** PR with fix + test
- **Want to Share:** Weekly SRE learning demo

## Success Checklist

After onboarding, you should be able to:

- [ ] Start/stop the simulated exchange
- [ ] Navigate all Grafana dashboards
- [ ] Run all 9 chaos experiments
- [ ] Generate and interpret flamegraphs
- [ ] Create custom Prometheus queries
- [ ] Understand the architecture
- [ ] Know when to use learning lab vs test env
- [ ] Feel confident breaking things safely

**Welcome to your SRE learning journey!** 🎓

Remember: The best way to learn is to break things. And this is the safest place to do it.

---

*Questions? Improvements? Add them to this doc via PR!*
