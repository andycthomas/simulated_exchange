# Known Limitations and Compatibility Notes

## System-Specific Limitations

### Your Current System
- **Kernel:** 5.14.0-362.8.1.el9_3.x86_64
- **OS:** RHEL/CentOS 9.3

## Experiment Compatibility Matrix

| Experiment | Core Functionality | BPF Monitoring | Notes |
|------------|-------------------|----------------|-------|
| 01 - Database Sudden Death | ✅ Works | ⚠️  May vary | Core experiment works without BPF |
| 02 - Network Partition | ✅ Works | ⚠️  May vary | iptables works, BPF optional |
| 03 - Memory OOM Kill | ✅ Works | ⚠️  Simplified | OOM detection via dmesg |
| 04 - CPU Throttling | ✅ Works | ⚠️  May vary | CPU limits work without BPF |
| 05 - Disk I/O Starvation | ✅ Works | ⚠️  May vary | iostat provides metrics |
| 06 - Connection Pool Exhaustion | ✅ Works | ⚠️  May vary | Postgres queries work |
| 07 - Redis Cache Failure | ✅ Works | ⚠️  May vary | Core functionality intact |
| 08 - Nginx Failure | ✅ Works | ⚠️  May vary | Uses docker logs instead |
| 09 - Network Latency | ❌ Requires netem | ❌ N/A | Kernel module not available |
| 10 - Multi-Service Failure | ✅ Works | ⚠️  May vary | Process monitoring works |

**Legend:**
- ✅ Works - Fully functional on your system
- ⚠️  May vary - BPF tracepoints may differ by kernel version
- ❌ Requires - Missing kernel feature

---

## BPF Compatibility Issues

### What's Affected

**BPF tracepoint structures are NOT stable across kernel versions.**

Your kernel (5.14.0) may have different tracepoint field definitions than what the scripts expect.

### Common BPF Errors

```bash
ERROR: Struct/union does not contain a field named 'comm'
ERROR: Struct/union does not contain a field named 'totalpages'
ERROR: The args builtin can only be used with tracepoint probes
ERROR: printf: %d specifier expects a value of type int
```

### Why This Happens

**Kernel Evolution:**
- Kernel 5.10: OOM tracepoint has fields A, B, C
- Kernel 5.14: OOM tracepoint has fields A, D, E (removed B,C, added D,E)
- Kernel 5.15: OOM tracepoint has fields A, D, E, F (added F)

**Your scripts were written for one kernel version but run on another.**

### How Scripts Handle This

**Graceful Degradation Strategy:**

1. **Simplified BPF scripts** - Use only basic, widely-available tracepoints
2. **Compile-time validation** - Test if script compiles before running
3. **Fallback mechanisms** - Continue experiment without BPF
4. **Alternative metrics** - Use dmesg, docker stats, proc files

**Example:**
```bash
# If BPF fails:
⚠️  BPF script has compatibility issues on this kernel
  Experiment will continue without BPF monitoring
  Using alternative detection: dmesg + docker stats
```

---

## What You Lose Without BPF

### With BPF (When Working):
- ✅ Kernel-level visibility
- ✅ TCP retransmission counts per process
- ✅ Memory allocation patterns
- ✅ CPU scheduler delays
- ✅ Real-time tracing with nanosecond precision

### Without BPF (Fallback):
- ✅ Experiment still works
- ✅ High-level metrics (docker stats, iostat)
- ✅ Kernel logs (dmesg)
- ✅ Application logs
- ❌ Lost: per-process kernel-level details
- ❌ Lost: real-time low-latency tracing

**Bottom Line:** You still learn about chaos engineering and system resilience. BPF adds deep insights but isn't required for the core value.

---

## Specific Kernel Requirements

### Experiment 09: Network Latency

**Requires:** `CONFIG_NET_SCH_NETEM=m` or `=y`

**Check:**
```bash
# Check if module exists
lsmod | grep sch_netem

# Try to load
sudo modprobe sch_netem

# Search for module file
find /lib/modules/$(uname -r) -name '*netem*'
```

**Your System:** ❌ Module not compiled in kernel

**Workarounds:**
1. Run on different system (Ubuntu, Fedora usually have it)
2. Compile custom kernel with netem
3. Skip this experiment (8 others work fine)

---

## How to Make BPF More Portable

### For Learning (Current Approach)
✅ Simplified scripts with fallbacks
✅ Graceful error messages
✅ Alternative metrics

### For Production (Advanced)
- **Use BTF (BPF Type Format)** - Kernel type information
- **Use CO-RE (Compile Once, Run Everywhere)** - Portable BPF
- **Use libbpf** instead of bpftrace
- **Version-specific scripts** - Detect kernel, load appropriate BPF

**Example: Production-Grade Approach**
```bash
KERNEL_VERSION=$(uname -r | cut -d. -f1-2)

case $KERNEL_VERSION in
    "5.10")
        bpftrace /bpf/oom_5.10.bt
        ;;
    "5.14")
        bpftrace /bpf/oom_5.14.bt
        ;;
    *)
        echo "Using generic fallback"
        # Use dmesg instead
        ;;
esac
```

---

## Testing Recommendations

### Before Running Experiments

**Quick BPF Compatibility Test:**
```bash
# Test if basic tracepoints work
sudo bpftrace -e 'BEGIN { printf("BPF works!\n"); exit(); }'

# Test if tracepoint exists
sudo bpftrace -l 'tracepoint:sched:sched_switch'

# Test if kprobe works
sudo bpftrace -e 'kprobe:tcp_v4_connect { printf("TCP connect\n"); exit(); }' &
sleep 2 && curl -s http://google.com && sudo pkill bpftrace
```

**If those work:** BPF is functional, tracepoint specifics may vary
**If those fail:** BPF has deeper issues on your system

### Run Experiments Anyway

**Even without BPF, you learn:**
- ✅ How to inject failures
- ✅ How services respond to chaos
- ✅ How to measure recovery
- ✅ What metrics to monitor
- ✅ How to automate chaos testing

**BPF adds depth, but chaos engineering fundamentals still apply.**

---

## Production vs Learning Environment

### This Learning Environment
- Simplified Docker containers
- Single-host deployment
- Kernel limitations acceptable
- BPF optional (nice-to-have)

### Real Production Trading System
- Bare metal servers
- Multi-datacenter setup
- Full-featured kernels
- Production-grade monitoring (Datadog, New Relic, etc.)
- BPF one tool among many

**Key Insight:** The chaos experiments teach you **what to test** and **how to test it**. The specific tooling (BPF vs other) matters less than the methodology.

---

## Summary

### ✅ What Works on Your System
- 9 out of 10 experiments fully functional
- Core chaos engineering concepts demonstrated
- Automated failure injection and recovery testing
- Clear metrics and reporting

### ⚠️ What's Limited
- BPF tracepoints may need kernel-specific adjustments
- Network latency experiment requires netem module
- Deep kernel observability reduced without BPF

### 💡 What You Learn
- **Regardless of BPF:** Chaos engineering methodology
- **With BPF:** Deep kernel-level insights
- **Without BPF:** Still valuable chaos testing

**Your chaos engineering suite is production-ready for learning. For real trading systems, you'd run these on production-grade infrastructure with full kernel support.**

---

## Next Steps

1. ✅ **Run the experiments** - They work even with limitations
2. ✅ **Learn the patterns** - Failure injection, measurement, recovery
3. ✅ **Document findings** - What breaks, how long to recover
4. 📝 **Apply to production** - Use these techniques on real systems
5. 📝 **Adapt BPF scripts** - Make kernel-specific versions as needed

**The goal isn't perfect BPF traces. The goal is understanding system resilience under failure.**
