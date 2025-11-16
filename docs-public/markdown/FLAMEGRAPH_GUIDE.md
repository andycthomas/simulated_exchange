# Flamegraph Guide

Complete guide to generating, analyzing, and interpreting flamegraphs for the Simulated Exchange platform.

---

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Start](#quick-start)
3. [Profile Types](#profile-types)
4. [Generating Flamegraphs](#generating-flamegraphs)
5. [Reading Flamegraphs](#reading-flamegraphs)
6. [Common Patterns](#common-patterns)
7. [AI-Powered Analysis](#ai-powered-analysis)
8. [Advanced Usage](#advanced-usage)
9. [Troubleshooting](#troubleshooting)

---

## Introduction

### What are Flamegraphs?

Flamegraphs are visual representations of profiling data that show:
- **Width:** Time spent in each function (wider = more time)
- **Height:** Call stack depth (deeper = more nested calls)
- **Color:** Different shades for visual distinction (no semantic meaning)

### Why Use Flamegraphs?

- **Visual Hotspot Identification:** Quickly spot performance bottlenecks
- **Stack Trace Context:** See entire call paths, not just individual functions
- **Interactive Exploration:** Click to zoom, search functions, see percentages
- **Language Agnostic:** Works with any profiling data

### When to Profile?

- 🔴 **During Incidents:** Identify root cause of performance degradation
- 🟡 **Regular Monitoring:** Baseline normal behavior for comparison
- 🟢 **Before Optimization:** Measure to guide where to focus efforts
- 🔵 **After Changes:** Verify optimizations had desired effect

---

## Quick Start

### 30-Second Flamegraph

```bash
# Generate CPU profile (recommended starting point)
make profile-cpu

# View in browser
make profile-view

# That's it! Flamegraph is at: flamegraphs/cpu_profile_*.svg
```

### 3-Minute Comprehensive Analysis

```bash
# Generate all profile types
make profile-all

# View all flamegraphs
firefox flamegraphs/*.svg &

# Or serve via HTTP for remote viewing
make profile-serve
# Open: http://localhost:8888
```

### Full Chaos + Profiling Demo

```bash
# Complete chaos engineering + profiling
make chaos-flamegraph

# This generates:
# - Baseline profile
# - Under-load profile
# - Chaos recovery profile
# - All profile types (heap, goroutine, etc.)
# - AI analysis report
```

---

## Profile Types

### 1. CPU Profile (`profile-cpu`)

**What it shows:** Where CPU time is spent
**Duration:** 30 seconds (configurable)
**Use when:** Performance is slow, CPU usage is high

```bash
make profile-cpu        # 30 second capture
make profile-cpu-long   # 60 second capture
```

**Look for:**
- Wide bars = CPU hotspots
- Repeated patterns = loops or recursive calls
- GC functions = memory pressure

### 2. Heap Profile (`profile-heap`)

**What it shows:** Memory allocations and usage
**Duration:** Snapshot (instant)
**Use when:** High memory usage, suspected memory leaks

```bash
make profile-heap
```

**Look for:**
- Large allocations
- Unexpected allocation sources
- Growing allocations over time

### 3. Goroutine Profile (`profile-goroutine`)

**What it shows:** Active goroutines and their states
**Duration:** Snapshot (instant)
**Use when:** Goroutine leaks, deadlocks, concurrency issues

```bash
make profile-goroutine
```

**Look for:**
- Too many goroutines (leak)
- Blocked goroutines
- Channel operations

### 4. Allocations Profile (`profile-allocs`)

**What it shows:** Memory allocation patterns over time
**Duration:** 30 seconds
**Use when:** GC pressure, allocation churn

```bash
make profile-allocs
```

**Look for:**
- Frequent small allocations
- Short-lived objects
- Allocation hotspots

### 5. Block Profile (`profile-block`)

**What it shows:** Where goroutines block on synchronization
**Duration:** Snapshot (instant)
**Use when:** Suspected lock contention, blocking I/O

```bash
make profile-block
```

**Look for:**
- Mutex locks
- Channel operations
- I/O wait times

### 6. Mutex Profile (`profile-mutex`)

**What it shows:** Lock contention and mutex wait times
**Duration:** Snapshot (instant)
**Use when:** Suspected lock contention

```bash
make profile-mutex
```

**Look for:**
- Highly contended locks
- Long wait times
- Lock ordering issues

---

## Generating Flamegraphs

### Using Make Targets (Recommended)

```bash
# Individual profiles
make profile-cpu
make profile-heap
make profile-goroutine

# All at once
make profile-all

# Clean up old profiles
make profile-clean
```

### Using the Script Directly

```bash
# Basic usage
./scripts/generate-flamegraph.sh cpu 30 8080

# Syntax
./scripts/generate-flamegraph.sh [type] [duration] [port]

# Examples
./scripts/generate-flamegraph.sh cpu 60 8080      # 60-sec CPU profile
./scripts/generate-flamegraph.sh heap 0 8081      # Heap snapshot, port 8081
./scripts/generate-flamegraph.sh goroutine 0 8080 # Goroutine snapshot
```

### Manual pprof Usage

```bash
# Capture profile
curl -o cpu.pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Generate flamegraph
go tool pprof -svg cpu.pprof > cpu.svg

# Interactive analysis
go tool pprof cpu.pprof
# In pprof: top, list <function>, web
```

### Profiling Different Services

```bash
# Trading API (main service)
./scripts/generate-flamegraph.sh cpu 30 8080

# Market Simulator
./scripts/generate-flamegraph.sh cpu 30 8081

# Order Flow Simulator
./scripts/generate-flamegraph.sh cpu 30 8082
```

---

## Reading Flamegraphs

### Visual Interpretation

```
┌─────────────────────────────────────────────────────────┐
│ main                                                    │  ← Bottom: Entry point
│ ├─────────────────────────────────┐                    │
│ │ handleRequest                   │                    │  ← Middle: Call stack
│ │ ├───────────┐ ├──────┐         │                    │
│ │ │ parseJSON │ │ query│         │                    │  ← Top: Leaf functions
│ │ └───────────┘ └──────┘         │                    │
│ └─────────────────────────────────┘                    │
└─────────────────────────────────────────────────────────┘
   ▲                                   ▲
   │                                   │
   Wider = More CPU Time         Narrower = Less CPU Time
```

### Key Principles

1. **Width = Time**
   - Wider bars consumed more CPU time
   - Focus on the widest bars first

2. **Height = Call Depth**
   - Each layer is a function call
   - Read bottom-to-top to follow execution

3. **Color = Visual Aid**
   - Colors are for visual distinction only
   - No semantic meaning (not hot vs cold)

4. **Flat vs Cumulative**
   - **Flat:** Time in that function directly
   - **Cumulative:** Time in function + callees

### Interactive Features

**Click a Box:**
- Zooms to show that function and its children
- Resets view to focus on subtree

**Ctrl+F (Search):**
- Highlights matching function names
- Great for finding specific code paths

**Hover:**
- Shows function name
- Displays time and percentage
- Shows sample counts

**Reset View:**
- Click the top bar
- Or refresh the page

---

## Common Patterns

### 1. Garbage Collection Overhead

**Pattern:**
```
runtime.gcBgMarkWorker
runtime.gcDrain
├─ Wide bars in GC functions
```

**Meaning:** High GC pressure, too many allocations

**Solutions:**
- Use object pooling (`sync.Pool`)
- Pre-allocate slices/maps
- Reduce pointer usage
- Increase GOGC if memory permits

---

### 2. Lock Contention

**Pattern:**
```
sync.(*Mutex).Lock
├─ Many thin bars stacked
├─ Long vertical columns
```

**Meaning:** Threads waiting for locks

**Solutions:**
- Reduce lock granularity
- Use `sync.RWMutex` for read-heavy workloads
- Replace locks with channels
- Consider lock-free data structures

---

### 3. JSON Serialization

**Pattern:**
```
encoding/json.Marshal
encoding/json.Unmarshal
├─ Wide bars in JSON functions
```

**Meaning:** High serialization cost

**Solutions:**
- Use `json.RawMessage` for delayed parsing
- Consider faster alternatives (easyjson, jsoniter)
- Batch operations
- Stream large payloads

---

### 4. Database Query Overhead

**Pattern:**
```
database/sql.(*DB).Query
├─ Wide bars in database calls
```

**Meaning:** Slow or frequent queries

**Solutions:**
- Add indexes
- Use prepared statements
- Implement caching
- Batch operations
- Optimize queries

---

### 5. Regular Expressions

**Pattern:**
```
regexp.(*Regexp).FindString
├─ Visible bars in regex functions
```

**Meaning:** Regex compilation or execution cost

**Solutions:**
- Pre-compile regexes (`regexp.MustCompile`)
- Cache compiled regexes
- Consider string operations instead
- Simplify complex patterns

---

### 6. Reflection

**Pattern:**
```
reflect.Value.Call
reflect.Type.Method
├─ Bars in reflect package
```

**Meaning:** Runtime reflection overhead

**Solutions:**
- Use code generation instead
- Cache reflection results
- Avoid reflection in hot paths
- Use concrete types

---

## AI-Powered Analysis

### Automatic Analysis

```bash
# Generate profile with AI analysis
make chaos-flamegraph

# Or analyze existing profile
python3 scripts/ai-flamegraph-analyzer.py flamegraphs/cpu_profile_*.pb.gz
```

### Analysis Output

The AI analyzer generates a markdown report with:

1. **Executive Summary**
   - Top hotspot identified
   - Critical issue count
   - Overall assessment

2. **Hotspot List**
   - Priority ranking (HIGH/MEDIUM/LOW)
   - CPU percentages
   - Issue descriptions

3. **Optimization Recommendations**
   - Categorized by type (Memory, Concurrency, etc.)
   - Specific actionable suggestions
   - Priority guidance

4. **Next Steps**
   - Commands to run
   - Additional profiles to capture
   - Investigation paths

### Example Analysis Output

```markdown
## 🎯 Executive Summary

**Top Hotspot:** `internal/matching.(*Engine).ProcessOrder`
- **CPU Usage:** 32.5% (direct)
- **Priority:** HIGH

**Critical Issues:** 3 high-priority hotspots found

## 🔥 Top Performance Hotspots

### 1. 🔴 HIGH Priority

**Function:** `internal/matching.(*Engine).ProcessOrder`

**Metrics:**
- Direct CPU: 32.5%
- Total CPU (with callees): 45.2%

**Issues Identified:**
- Consumes 32.5% CPU directly
- Call tree consumes 45.2% total CPU

## 💡 Optimization Recommendations

### 1. 🔴 Concurrency

**Priority:** HIGH
**Issue:** Lock contention: 12.3% CPU

**Suggested Actions:**
- Reduce lock granularity (use more fine-grained locks)
- Replace mutexes with channels where appropriate
- Use sync.RWMutex for read-heavy workloads
- Consider lock-free data structures
```

---

## Advanced Usage

### Comparing Profiles

```bash
# Capture before optimization
make profile-cpu
mv flamegraphs/cpu_profile_*.pb.gz before.pb.gz

# Make optimizations
# ...

# Capture after optimization
make profile-cpu
mv flamegraphs/cpu_profile_*.pb.gz after.pb.gz

# Compare
go tool pprof -base before.pb.gz after.pb.gz
```

### Differential Flamegraphs

```bash
# Show difference between two profiles
go tool pprof -base baseline.pb.gz -svg current.pb.gz > diff.svg
```

### Profiling Under Load

```bash
# Start load generator
./scripts/load-generator.sh &

# Capture profile while under load
make profile-cpu

# Stop load
kill %1
```

### Custom Profile Duration

```bash
# 5-minute CPU profile for rare code paths
./scripts/generate-flamegraph.sh cpu 300 8080

# Very quick 10-second profile
./scripts/generate-flamegraph.sh cpu 10 8080
```

### Remote Profiling

```bash
# Serve flamegraphs over HTTP
make profile-serve  # Starts on port 8888

# Access from anywhere
# Open: http://<your-host>:8888

# Custom port
./scripts/serve-flamegraphs.sh 9000
```

### Continuous Profiling

```bash
# Profile every 5 minutes
while true; do
    make profile-cpu
    sleep 300
done

# Or use a cron job
*/5 * * * * cd /path/to/project && make profile-cpu
```

---

## Troubleshooting

### Profile Capture Fails

**Problem:** `curl` returns error or empty file

**Solutions:**
```bash
# Check if service is running
curl http://localhost:8080/health

# Verify pprof endpoint is available
curl http://localhost:8080/debug/pprof/

# Check for port conflicts
netstat -tuln | grep 8080

# Restart service
docker compose restart trading-api
```

### Flamegraph Not Generated

**Problem:** SVG file not created

**Solutions:**
```bash
# Check if Go is installed
go version

# Verify pprof tool is available
go tool pprof -h

# Try manual generation
go tool pprof -svg profile.pb.gz > profile.svg

# Check for errors in profile file
file profile.pb.gz  # Should say "gzip compressed data"
```

### Empty or Tiny Flamegraph

**Problem:** Flamegraph generated but shows little data

**Solutions:**
```bash
# System is idle - generate load first
./scripts/load-generator.sh

# Profile duration too short
./scripts/generate-flamegraph.sh cpu 60 8080  # Increase duration

# Wrong profile type
make profile-cpu  # CPU profile shows execution time
make profile-heap # Heap profile shows allocations
```

### AI Analysis Doesn't Run

**Problem:** No `*_analysis.md` file generated

**Solutions:**
```bash
# Check if Python 3 is installed
python3 --version

# Verify AI analyzer exists
ls scripts/ai-flamegraph-analyzer.py

# Run manually
python3 scripts/ai-flamegraph-analyzer.py flamegraphs/cpu_*.pb.gz

# Check for script errors
python3 scripts/ai-flamegraph-analyzer.py flamegraphs/cpu_*.pb.gz --verbose
```

### Browser Won't Open

**Problem:** `make profile-view` doesn't open browser

**Solutions:**
```bash
# Open manually
firefox flamegraphs/*.svg

# Or use HTTP server
make profile-serve
# Then open: http://localhost:8888

# Check which browsers are available
which firefox chromium google-chrome
```

---

## Best Practices

### 1. Establish Baselines

```bash
# Regular profiling under normal conditions
# Schedule weekly: Sundays at 2am
0 2 * * 0 cd /app && make profile-cpu
```

### 2. Profile Before Optimizing

> "Premature optimization is the root of all evil" - Donald Knuth

Always profile first to identify actual bottlenecks, not assumed ones.

### 3. Profile in Production-Like Environments

Development environments often hide performance issues. Profile with:
- Production-like data volumes
- Realistic concurrency
- Similar hardware specs

### 4. Keep Historical Data

```bash
# Archive profiles with timestamps
cp flamegraphs/*.svg archive/$(date +%Y%m%d)/
```

### 5. Correlate with Metrics

Compare flamegraphs with:
- Prometheus metrics
- Error rates
- Response times
- Resource usage

### 6. Focus on the Biggest Wins

Optimize functions that:
- Have high flat % (direct CPU time)
- Are called frequently
- Are in hot code paths
- Are fixable (not in stdlib)

### 7. Verify Optimizations

```bash
# Before
make profile-cpu
# Optimize
# After
make profile-cpu
# Compare improvements
```

---

## Examples

### Example 1: Finding a Slow Function

```bash
# Generate CPU profile
make profile-cpu

# Open flamegraph
firefox flamegraphs/cpu_profile_*.svg

# Look for widest bars
# Click to zoom in
# See that processOrder is 40% of CPU

# Check AI analysis
cat flamegraphs/cpu_*_analysis.md | head -40

# AI confirms: processOrder is top hotspot
# Recommendation: Reduce lock contention

# View function details
go tool pprof flamegraphs/cpu_profile_*.pb.gz
(pprof) list processOrder
```

### Example 2: Diagnosing Memory Issues

```bash
# High memory usage reported
docker stats

# Capture heap profile
make profile-heap

# View flamegraph
firefox flamegraphs/heap_*.svg

# Wide bar in cacheManager.Load
# AI analysis suggests:
# "Large allocations in cache, consider TTL limits"

# Implement fix: Add cache size limit
# Verify:
make profile-heap
# Memory usage reduced!
```

### Example 3: Goroutine Leak

```bash
# Goroutine count increasing
curl http://localhost:8080/debug/pprof/goroutine?debug=1

# Capture goroutine profile
make profile-goroutine

# View flamegraph
firefox flamegraphs/goroutine_*.svg

# See many goroutines blocked in websocket.Read
# Find missing connection cleanup

# Fix and verify
make profile-goroutine
# Goroutine count back to normal
```

---

## Resources

### Internal Documentation
- **Architecture:** `docs/ARCHITECTURE.md`
- **SRE Runbook:** `docs/SRE_RUNBOOK.md`
- **Demo Guide:** `PERFORMANCE_REVIEW_DEMO_CHEAT_SHEET.md`

### External Resources
- [Brendan Gregg's Flamegraphs](http://www.brendangregg.com/flamegraphs.html)
- [Go pprof Documentation](https://pkg.go.dev/net/http/pprof)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [Uber's Go Profiling Guide](https://eng.uber.com/go-geom/)

### Tools
- [go-torch](https://github.com/uber/go-torch) - Flamegraph generator
- [pprof](https://github.com/google/pprof) - Profiling tool
- [graphviz](https://graphviz.org/) - Graph visualization

---

## Quick Reference Card

```bash
# GENERATION
make profile-cpu        # CPU profile (30s)
make profile-heap       # Heap snapshot
make profile-all        # All profile types
make chaos-flamegraph   # Full chaos + profiling

# VIEWING
make profile-view       # Open in browser
make profile-serve      # HTTP server on :8888
firefox flamegraphs/*.svg

# ANALYSIS
cat flamegraphs/*_analysis.md
go tool pprof flamegraphs/*.pb.gz

# MANAGEMENT
make profile-clean      # Delete old profiles
ls -lh flamegraphs/     # List profiles

# HELP
make profile-help       # Show all profiling targets
```

---

**Happy Profiling! 🔥**

*Remember: Profile first, optimize second, measure always.*
