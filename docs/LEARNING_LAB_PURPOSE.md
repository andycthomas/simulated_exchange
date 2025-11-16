# 🎓 Simulated Exchange: SRE Learning Lab

## Purpose & Vision

The **Simulated Exchange** is a **safe, standalone learning environment** designed specifically for SRE team members to develop and practice critical reliability engineering skills without risk to production or test systems.

### The Learning Lab Concept

Think of this as your **personal SRE training ground** - a fully functional trading exchange that you can:
- **Break safely** - Experiment with chaos engineering without fear
- **Monitor deeply** - Learn eBPF, flamegraphs, and profiling tools
- **Optimize freely** - Test performance tuning techniques
- **Recover quickly** - Practice incident response workflows

## Relationship to Production Systems

### Complementary, Not Competitive

```
┌─────────────────────────────────────────────────────────┐
│           Production Trading Infrastructure              │
│                  (DO NOT TOUCH)                          │
│  • Real money, real customers, real risk                │
│  • 17 Linux servers in production-like config           │
│  • Critical for business operations                     │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ Learn skills here first
                          │
┌─────────────────────────────────────────────────────────┐
│        17-Server Test Environment                       │
│              (LIMITED ACCESS)                            │
│  • Production-like configuration                        │
│  • Used for BIOS changes, config testing               │
│  • Shared by developers & operations                   │
│  • Still has risk/impact on team testing               │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ Practice here first
                          │
┌─────────────────────────────────────────────────────────┐
│        Simulated Exchange Learning Lab                  │
│            (YOUR SAFE SANDBOX) ⭐                        │
│  • Zero risk to real systems                           │
│  • Break it, crash it, chaos it - no consequences      │
│  • Learn eBPF, flamegraphs, chaos engineering          │
│  • Individual or team learning environment             │
│  • Docker-based, runs on single machine/laptop        │
└─────────────────────────────────────────────────────────┘
```

### Why Both?

| System | Purpose | When to Use | Risk Level |
|--------|---------|-------------|------------|
| **Production** | Revenue generation, customer service | Never for learning | ⛔ CRITICAL |
| **17-Server Test** | Validate config/BIOS changes, integration testing | When changes are ready for pre-prod validation | ⚠️ HIGH - Shared resource |
| **Simulated Exchange** | Learn, experiment, break things | Always - before touching real systems | ✅ ZERO - Isolated sandbox |

## What Makes This a Great Learning Lab

### 1. **Risk-Free Experimentation**

```bash
# You can literally do this without consequences:
docker compose down  # Destroy everything
rm -rf data/         # Delete all data
# Start fresh in 30 seconds
docker compose up -d
```

**Contrast with Test Environment:**
- Breaking test env affects entire team
- Recovery requires coordination
- Mistakes impact developer productivity
- Limited availability for learning

### 2. **Realistic Trading Exchange Behavior**

Not a toy application - real components you'll see in production:
- ✅ Microservices architecture (3 Go services)
- ✅ PostgreSQL database with real queries
- ✅ Redis for caching and pub/sub
- ✅ Prometheus metrics collection
- ✅ Grafana dashboards
- ✅ NGINX load balancing
- ✅ Real performance characteristics (2,500+ orders/sec)

### 3. **Built-In Chaos Engineering**

9 pre-configured chaos experiments you can safely run:

```bash
# Learn chaos engineering safely
./chaos-experiments/01-service-failure.sh      # Kill services, watch recovery
./chaos-experiments/04-cpu-throttling.sh       # Starve CPU, observe impact
./chaos-experiments/06-network-partition.sh    # Break network, test resilience
```

**Why this matters:** Before you use these techniques on the 17-server test env, understand them here first.

### 4. **eBPF & Performance Profiling Sandbox**

Learn modern observability tools with zero risk:

```bash
# Generate flamegraphs safely
./scripts/generate-flamegraph.sh cpu 30 8080

# Experiment with eBPF tools
sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* { @[probe] = count(); }'

# Profile without fear of impact
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
```

**Why this matters:** These tools can impact performance. Learn how to use them without affecting real work.

### 5. **Individual Learning Pace**

- **Run on your laptop** - No shared resource conflicts
- **Work offline** - Learn during commute, travel, evenings
- **Repeat experiments** - Destroy and rebuild as many times as needed
- **Document findings** - Save your learnings for team knowledge sharing

## Learning Paths for SRE Team Members

### 🎯 Path 1: Observability Foundations (Week 1-2)

**Goal:** Master metrics, logs, and profiling

1. **Day 1-2: Metrics & Dashboards**
   ```bash
   make monitoring-up
   # Explore Grafana dashboards
   # Learn Prometheus queries
   # Understand RED metrics (Rate, Errors, Duration)
   ```

2. **Day 3-4: Performance Profiling**
   ```bash
   ./scripts/generate-flamegraph.sh cpu 30 8080
   # Read the AI analysis
   # Understand hotspots
   # Learn pprof commands
   ```

3. **Day 5: Log Analysis**
   ```bash
   docker logs -f simulated-exchange-trading-api
   # Understand structured logging
   # Filter and query logs
   ```

**Deliverable:** Present 3 performance insights to team

### 🎯 Path 2: Chaos Engineering (Week 3-4)

**Goal:** Understand failure modes and resilience

1. **Week 3: Basic Chaos**
   - Run all 9 chaos experiments
   - Document observed behaviors
   - Learn recovery patterns
   - Practice incident response

2. **Week 4: Custom Chaos**
   - Design your own chaos scenarios
   - Test hypotheses about system behavior
   - Measure blast radius
   - Validate monitoring alerts

**Deliverable:** Create 2 new chaos experiments

### 🎯 Path 3: eBPF & Advanced Profiling (Week 5-6)

**Goal:** Master low-level system observability

1. **Learn BPF Tools**
   ```bash
   # Network latency analysis
   sudo bpftrace scripts/network-latency.bt

   # System call tracing
   sudo bpftrace scripts/syscall-count.bt
   ```

2. **CPU Profiling Deep Dive**
   - Generate profiles under different loads
   - Compare baseline vs chaos vs load
   - Identify optimization opportunities

**Deliverable:** BPF script that solves a real observability problem

### 🎯 Path 4: Performance Optimization (Week 7-8)

**Goal:** Apply SRE techniques to improve performance

1. **Baseline & Benchmark**
   ```bash
   make load-test-light
   # Record baseline metrics
   ```

2. **Optimize**
   - Identify bottlenecks from flamegraphs
   - Apply optimization techniques
   - Re-measure performance

3. **Document**
   - What you changed
   - Why you changed it
   - Impact on performance

**Deliverable:** Performance optimization report

## Transitioning Skills to Real Systems

### The Learning Ladder

```
1. LEARN in Simulated Exchange (this lab)
   ↓ Gain confidence, make mistakes safely

2. VALIDATE in Simulated Exchange
   ↓ Prove your approach works

3. DOCUMENT your findings
   ↓ Share with team, get review

4. APPLY to 17-server test environment
   ↓ With team coordination & approval

5. REFINE based on test environment learnings
   ↓ Iterate safely

6. PROPOSE for production
   ↓ With runbook, rollback plan, metrics
```

### Example: Learning eBPF

**❌ DON'T:** Run unfamiliar eBPF scripts on test servers
- Risk: Unknown performance impact
- Risk: Potential to crash services
- Risk: Affecting team testing

**✅ DO:** Master eBPF in Simulated Exchange first

```bash
# Safe learning progression:
1. Run eBPF scripts in Simulated Exchange
2. Understand overhead and safety
3. Create your own scripts
4. Test impact on system performance
5. Document findings and safe usage
6. THEN propose for test environment with:
   - Clear purpose
   - Expected overhead
   - Rollback plan
   - Time window request
```

## Team Benefits

### Knowledge Sharing

**Share Your Experiments:**
```bash
# Save your work
git branch my-chaos-experiments
git add chaos-experiments/custom/
git commit -m "New latency injection experiment"

# Share with team
git push origin my-chaos-experiments
# Create PR for team review
```

### Safe Mentoring

**Senior SREs can:**
- Demo techniques live without risk
- Let juniors try dangerous things safely
- Review approaches before production use
- Build team runbooks from learnings

### Continuous Learning

**Use for:**
- Weekly chaos game days
- Lunch & learn sessions
- Interview candidate assessments
- New tool evaluation
- Conference technique testing

## Cost-Benefit Analysis

### Simulated Exchange Learning Lab

**Costs:**
- 10 minutes setup time
- Laptop/desktop resources while running
- Your learning time investment

**Benefits:**
- Zero risk to production/test systems
- Unlimited experimentation freedom
- Individual learning pace
- Reusable for entire team
- Safe failure environment
- Knowledge sharing platform

### Without Learning Lab

**Costs:**
- Risk to shared test environment
- Team productivity impact when things break
- Coordination overhead
- Limited learning opportunities
- Fear of experimentation

**Benefits:**
- More realistic environment (only marginally over simulated)

## Success Metrics

Track your learning progress:

- [ ] **Week 1-2:** Can read and interpret Grafana dashboards
- [ ] **Week 3-4:** Successfully run all 9 chaos experiments
- [ ] **Week 5-6:** Created 3 custom eBPF scripts
- [ ] **Week 7-8:** Optimized performance by 20%+
- [ ] **Month 3:** Contributed chaos experiment to team
- [ ] **Month 4:** Applied learning to test environment
- [ ] **Month 6:** Led team learning session

## Getting Started

### For Individual Learners

```bash
# 1. Clone and start
git clone <repo>
cd simulated_exchange
make docker-up

# 2. Pick a learning path above
# 3. Document your learnings
# 4. Share with team
```

### For Team Leads

```bash
# 1. Set up team learning sessions
#    - Weekly chaos engineering games
#    - Monthly performance optimization challenges
#
# 2. Create team challenges
#    - "Find 5 performance bottlenecks"
#    - "Design chaos for worst case scenario"
#    - "Build custom monitoring dashboard"
#
# 3. Track team progress
#    - Learning path completion
#    - Skills developed
#    - Insights shared
```

## FAQ

### Q: Why not just use the 17-server test environment?

**A:** Multiple reasons:
1. **Risk** - Breaking test env affects entire team
2. **Access** - Shared resource, can't monopolize for learning
3. **Repeatability** - Hard to reset to clean state
4. **Availability** - Busy with actual testing work
5. **Overhead** - Requires coordination and scheduling

The simulated exchange is **always available, always safe, always yours**.

### Q: Is this realistic enough to learn from?

**A:** Yes! It has:
- Real microservices architecture
- Real databases (PostgreSQL, Redis)
- Real metrics (Prometheus, Grafana)
- Real performance characteristics
- Real failure modes
- Real observability tools

What it **intentionally doesn't have:**
- 17 physical servers (use Docker instead)
- Production data (use simulated trading data)
- Production load (but can simulate 2,500+ ops/sec)

### Q: When should I use test environment vs learning lab?

**Learning Lab (Simulated Exchange):**
- Learning new tools/techniques
- Chaos engineering practice
- Performance optimization experiments
- eBPF script development
- Breaking things intentionally
- **Any time you're not 100% confident**

**Test Environment:**
- Validating production-ready changes
- BIOS/firmware testing
- Multi-server configuration changes
- Integration testing with real components
- **After you've proven it works in learning lab**

### Q: Can multiple team members use it simultaneously?

**A:** Yes! Two approaches:

1. **Individual instances** (recommended for learning)
   - Each person runs on their laptop
   - Zero conflicts
   - Learn at your own pace

2. **Shared demo instance** (for team sessions)
   - One instance for demos/presentations
   - Coordinate who's using when
   - Great for team learning days

### Q: What if I break it?

**A:** That's the point! 🎉

```bash
# Completely broken it?
docker compose down
docker system prune -af  # Nuclear option
docker compose up -d     # Fresh start in 30 seconds
```

You **cannot** break it in a way that matters. That's the whole idea.

## Resources

- **Quick Start:** [README.md](README.md)
- **SRE Runbook:** [SRE_RUNBOOK.md](SRE_RUNBOOK.md)
- **Chaos Experiments:** [chaos-experiments/](../chaos-experiments/)
- **Flamegraph Guide:** [FLAMEGRAPH_GUIDE.md](FLAMEGRAPH_GUIDE.md)
- **Architecture:** [ARCHITECTURE.md](ARCHITECTURE.md)

## Conclusion

The Simulated Exchange learning lab is your **risk-free sandbox** for developing SRE skills that you'll apply to real systems. It's not competing with the test environment - it's **preparing you** to use it more effectively.

**Remember:**
- Production = **DO NOT TOUCH for learning**
- Test Environment = **After you've learned and validated**
- Simulated Exchange = **Learn, break, experiment freely**

Now go break something! 🚀

---

**Questions?** Add them to team wiki or discuss in SRE channel.
**Improvements?** PR this document with your learnings!
