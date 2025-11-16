# Performance Review Demo Cheat Sheet

**Duration:** 5 minutes
**Audience:** Technical leadership, SRE teams
**Objective:** Demonstrate automated incident response platform

---

## 🎯 The Hook (30 seconds)

**SAY:**
- "Who's been paged at 3am? How long for root cause?"
- "I built a platform that does it in 90 seconds. Let me show you."

---

## 📋 Demo Flow

### Step 1: Prove It Works (15 seconds)

**Command:**
```bash
curl -s http://localhost:8080/health | jq '.status'
```

**SAY:**
- "Trading system operational, processing orders..."

---

### Step 2: The Magic Command (2 minutes)

**Command:**
```bash
make chaos-flamegraph
```

**This ONE command demonstrates ALL FOUR technologies:**
- ✅ Live Trading System
- ✅ Chaos Engineering (DB kill)
- ✅ eBPF Kernel Tracing
- ✅ Flamegraphs + AI Analysis

**NARRATE while it runs:**
- "Capturing baseline metrics..."
- "Starting eBPF kernel-level tracing..."
- "Injecting chaos - killing the database..."
- "Profiling under stress with pprof..."
- "Generating visual flamegraphs..."
- "AI analyzing bottlenecks..."

---

### Step 3: Show eBPF Kernel Tracing (30 seconds)

**Command:**
```bash
tail -30 chaos-results/flamegraph-profiling/experiment.log.bpf
```

**SAY:**
- "eBPF caught everything at kernel level - TCP failures, retries, timing"
- "Zero overhead, no application changes needed"

---

### Step 4: Show Visual Flamegraph (45 seconds)

**Command:**
```bash
firefox chaos-results/flamegraph-profiling/flamegraphs/cpu_load.svg &
```

**DEMONSTRATE:**
- Width = CPU time (wider = hotspot)
- Click boxes to zoom
- Ctrl+F to search
- Hover for details

**SAY:**
- "See this wide bar? That's our hotspot. Width shows CPU time."

---

### Step 5: Show AI-Powered Analysis (30 seconds)

**Command:**
```bash
head -40 flamegraphs/cpu_load_analysis.md
```

**HIGHLIGHT:**
- Top hotspot automatically identified
- Priority ranking (HIGH/MEDIUM/LOW)
- Specific code locations
- Actionable recommendations

---

## 💥 FINALE: The Impact (15 seconds)

**TRADITIONAL ROOT CAUSE: 60-90 minutes**
- SSH to servers.................. 2 min
- Attach profiler................. 5 min
- Capture data.................... 5 min
- Download & analyze.............. 15 min
- Research optimization........... 30 min
- Document findings............... 10 min

**THIS PLATFORM: 90 seconds (fully automated)**

**ROI: 40-60x faster incident response**

**SAY:**
- "From hours to 90 seconds."
- "That's the difference between being up at 3am..."
- "...and sleeping through the night."

---

## 🎬 Pre-Demo Checklist

### 15 Minutes Before
- [ ] Run `make prep-demo-quick` (30 sec) or `make prep-demo` (3 min)
- [ ] Verify services: `docker compose ps`
- [ ] Test health check: `curl http://localhost:8080/health`

### 5 Minutes Before
- [ ] Increase terminal font size (huge!)
- [ ] Open this cheat sheet in another window
- [ ] Open browser for flamegraph viewing
- [ ] Close unnecessary windows/tabs

### During Demo
- [ ] Speak slowly and clearly
- [ ] Keep narrating during long commands
- [ ] Point at the screen while explaining flamegraphs
- [ ] Maintain eye contact with audience

---

## 🔧 Backup Plan

If live demo fails, use backup artifacts:

```bash
# Show pre-generated flamegraphs
ls -lh chaos-results-BACKUP-*/flamegraph-profiling/flamegraphs/

# Open backup flamegraph
firefox chaos-results-BACKUP-*/flamegraph-profiling/flamegraphs/cpu_load*.svg &

# Show backup analysis
cat flamegraphs-BACKUP-*/cpu_*_analysis.md | head -40
```

**SAY:**
- "I have pre-generated results from an earlier run..."
- (Then continue with Steps 3-5)

---

## 🎯 Key Messages to Reinforce

1. **One Command = Complete Diagnostics**
   - "Notice: ONE command, full root cause analysis"

2. **Production-Ready**
   - "This is a real trading system, 41,821 lines of Go"
   - "Chaos experiments simulate real production failures"

3. **Zero Overhead**
   - "eBPF runs at kernel level with negligible overhead"
   - "No application changes required"

4. **Automated Intelligence**
   - "AI automatically prioritizes issues"
   - "Goes straight from incident to actionable recommendations"

5. **Business Impact**
   - "Mean time to resolution: 90 seconds vs 60-90 minutes"
   - "Reduced on-call burden, faster incident response"

---

## 📞 Common Questions & Answers

**Q: Does this work in production?**
- A: "Yes. eBPF is used in production by Netflix, Facebook, Google."

**Q: What about the overhead?**
- A: "eBPF: <1% overhead. Profiling: 5% during capture. Chaos: dev/staging only."

**Q: Can this integrate with our existing monitoring?**
- A: "Absolutely. Exports to Prometheus, Grafana, standard observability stack."

**Q: How hard is this to deploy?**
- A: "Docker Compose for dev. Kubernetes manifests for production. 30 min setup."

**Q: What languages does this support?**
- A: "Go natively. Python, Java, C++ with respective profiling tools."

---

## 🚀 Quick Command Reference

```bash
# Preparation
make prep-demo-quick    # Use existing artifacts
make prep-demo          # Fresh run

# Demo modes
./scripts/demo-guide.sh # Interactive with prompts
make run-demo           # Quick reference (no prompts)

# Service management
docker compose ps       # Check services
docker compose up -d    # Start services
docker compose restart  # Restart if needed

# Manual profiling
make profile-cpu        # Just CPU profile
make chaos-flamegraph   # Full chaos + profiling

# View results
make profile-view       # Open in browser
make chaos-results      # Show recent experiments
```

---

## 💡 Pro Tips

1. **Terminal Setup**
   - Use large font (18-24pt)
   - High contrast theme
   - Full screen terminal

2. **Timing**
   - Practice the narration
   - Don't rush the hook
   - Use the 90-second wait to explain technologies

3. **Audience Engagement**
   - Ask about 3am pages (gets nods)
   - Point at flamegraph bars
   - Emphasize time savings (60min → 90sec)

4. **Technical Depth**
   - Adjust based on audience
   - Execs: Focus on ROI
   - Engineers: Show eBPF details
   - SRE: Emphasize automation

---

## 📚 Supporting Materials

- **Full Demo Script:** `make run-demo`
- **Interactive Guide:** `./scripts/demo-guide.sh`
- **Architecture Docs:** `docs/ARCHITECTURE.md`
- **SRE Runbook:** `docs/SRE_RUNBOOK.md`
- **Business Value:** `docs/BUSINESS_VALUE.md`

---

## 🎓 Practice Runs

**Before the actual demo:**

1. Run through once silently
2. Run through with narration
3. Time yourself (aim for 4:30-5:00)
4. Practice backup plan
5. Test all URLs/paths

**Rehearsal checklist:**
- [ ] Hook feels natural
- [ ] Commands execute correctly
- [ ] Narration timing is good
- [ ] Flamegraph displays properly
- [ ] Finale has impact

---

## 🏆 Success Criteria

You nailed it if:
- ✅ Hook grabbed attention in first 30 seconds
- ✅ All commands executed successfully
- ✅ Audience saw the flamegraph
- ✅ Time savings message landed
- ✅ Got questions afterward

---

**Remember:** This is about the business value, not just the tech.
**Focus:** From incident to resolution in 90 seconds vs 60-90 minutes.
**Impact:** Sleep through the night instead of being paged at 3am.

🔥 **Good luck!** 🔥
