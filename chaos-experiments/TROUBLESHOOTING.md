# Chaos Experiments Troubleshooting Guide

## Common Issues and Solutions

### Issue 0: BPF Scripts Fail with Tracepoint Errors

**Symptom:**
```bash
ERROR: Struct/union does not contain a field named 'comm'
ERROR: Struct/union does not contain a field named 'totalpages'
ERROR: The args builtin can only be used with tracepoint probes
ERROR: printf: %d specifier expects a value of type int
```

**Root Cause:**
- BPF tracepoint structures vary by kernel version
- What works on kernel 5.15 may not work on 5.14
- Field names, types, and availability differ

**Solution:**
The scripts have been updated to handle this gracefully:

1. **Simplified BPF scripts** - Use only universally available tracepoints
2. **Compile-time checking** - Scripts test if BPF compiles before running
3. **Graceful degradation** - Experiments continue without BPF if it fails

**What You'll See:**
```bash
⚠ BPF script has compatibility issues on this kernel
  Experiment will continue without BPF monitoring
  OOM detection will use dmesg/kernel logs instead
```

**Impact:**
- ✅ Experiments still run and work correctly
- ⚠️  You lose kernel-level observability (BPF traces)
- ✅ Alternative detection methods are used (dmesg, docker stats, etc.)

**Why This Happens:**
BPF tracepoints are **not stable across kernel versions**. To be truly portable, you need to:
- Use only basic tracepoints
- Check kernel version and adapt
- Use BTF (BPF Type Format) and CO-RE (Compile Once, Run Everywhere)

**For Production:**
This is why companies like Facebook use **BTF/CO-RE** or **kernel-specific BPF programs**. For learning purposes, these experiments work fine without BPF - the chaos itself is what matters.

---

### Issue 1: Script Hangs When Accessing Container Logs

**Symptom:**
```bash
# Script hangs at "Nginx connection stats:" and never completes
# Must press Ctrl+C to interrupt
```

**Root Cause:**
Many Docker containers redirect logs to `/dev/stdout` and `/dev/stderr` instead of actual log files.

```bash
# Check if log files are symlinks:
docker exec container ls -la /var/log/nginx/
# Output:
# lrwxrwxrwx 1 root root 11 Aug 13 18:52 access.log -> /dev/stdout
```

When you try to `cat /dev/stdout`, it hangs waiting for input.

**Solution:**
Use `docker logs` instead of trying to `cat` log files inside containers:

```bash
# ❌ WRONG - Will hang:
docker exec nginx cat /var/log/nginx/access.log

# ✅ CORRECT - Use docker logs:
docker logs nginx --tail 10
```

**Fixed in:** 08-nginx-failure.sh (line 139-140)

---

### Issue 2: BPF Script Fails with "Cannot attach to probe"

**Symptom:**
```bash
ERROR: Cannot attach kprobe: No such file or directory
```

**Root Cause:**
The kernel function or tracepoint doesn't exist on your kernel version.

**Solution:**
1. Check available probes:
```bash
# List all available kprobes
bpftrace -l 'kprobe:*' | grep tcp_v4_connect

# List all tracepoints
bpftrace -l 'tracepoint:*' | grep tcp
```

2. Use tracepoints (more stable) instead of kprobes when possible:
```bash
# Instead of:
kprobe:tcp_v4_connect { ... }

# Use:
tracepoint:tcp:tcp_connect { ... }
```

---

### Issue 3: "Permission Denied" When Running Experiments

**Symptom:**
```bash
./01-database-sudden-death.sh
ERROR: Permission denied
```

**Root Cause:**
Some experiments require root access for:
- BPF tracing (bpftrace)
- Network manipulation (tc, iptables)
- Container inspection

**Solution:**
Run experiments marked "Root Required: Yes" with sudo:

```bash
# Check if root is required:
grep "Root Required" README.md

# Run with sudo:
sudo ./02-network-partition.sh
```

**No root required:** 01, 04, 07, 08, 10
**Root required:** 02, 03, 05, 06, 09

---

### Issue 4: Experiment Fails Because Service Not Running

**Symptom:**
```bash
[2025-10-06 14:30:15] ✗ PostgreSQL container not running
```

**Root Cause:**
Required services are not started.

**Solution:**
1. Check all services are running:
```bash
docker compose ps
```

2. Start missing services:
```bash
docker compose up -d
```

3. Verify specific service:
```bash
docker ps | grep simulated-exchange
```

**Required services for experiments:**
- postgres (all experiments)
- redis (experiments 07, 10)
- trading-api (all experiments)
- nginx (experiment 08)

---

### Issue 5: Order Flow Simulator Should Be Stopped

**Symptom:**
Experiments show inconsistent results or system becomes overloaded during chaos testing.

**Root Cause:**
Order flow simulator is generating continuous load, interfering with experiment measurements.

**Solution:**
Stop order flow simulator before running experiments:

```bash
# Stop order flow
docker compose stop order-flow-simulator

# Verify it's stopped
docker ps | grep order-flow-simulator  # Should return nothing

# Run experiments
./01-database-sudden-death.sh

# Restart order flow when done (if needed)
docker compose start order-flow-simulator
```

**Note:** With 110K+ existing orders in the database, you don't need continuous order generation for realistic chaos testing.

---

### Issue 6: bpftrace Not Found

**Symptom:**
```bash
[2025-10-06 14:30:15] ⚠ bpftrace not found - skipping BPF monitoring
```

**Root Cause:**
bpftrace is not installed on the system.

**Solution:**
Install bpftrace:

```bash
# RHEL/CentOS/Fedora
sudo dnf install bpftrace

# Ubuntu/Debian
sudo apt install bpftrace

# Verify installation
bpftrace --version
```

**Note:** Experiments will still run without bpftrace, but you'll lose kernel-level observability insights.

---

### Issue 7: Docker Containers Don't Auto-Restart

**Symptom:**
After killing a container in an experiment, it stays dead instead of auto-restarting.

**Root Cause:**
Container restart policy is not set or is set to "no".

**Solution:**
Check and update restart policy in `docker-compose.yml`:

```yaml
services:
  postgres:
    restart: always  # or "unless-stopped"
```

Verify current restart policy:
```bash
docker inspect simulated-exchange-postgres | grep -i restart
```

---

### Issue 8: Experiment Times Out Waiting for Service

**Symptom:**
```bash
[2025-10-06 14:35:00] ⚠ PostgreSQL did not recover within 60 seconds
```

**Root Cause:**
Service is taking longer than expected to restart, or there's an actual problem.

**Solution:**
1. Check container status:
```bash
docker ps -a | grep postgres
```

2. Check container logs:
```bash
docker logs simulated-exchange-postgres --tail 50
```

3. Manually restart if needed:
```bash
docker restart simulated-exchange-postgres
```

4. Increase timeout in script if service legitimately needs more time.

---

### Issue 9: Network Experiments Fail (tc/iptables errors)

**Symptom:**
```bash
ERROR: Cannot find device "eth0@if10713"
ERROR: RTNETLINK answers: Operation not permitted
ERROR: Specified qdisc kind is unknown (netem)
```

**Root Cause:**
- Not running as root
- Container network interface name includes @ifXXXX suffix
- Network namespace issues
- sch_netem kernel module not available

**Solution:**
1. Run with sudo:
```bash
sudo ./02-network-partition.sh
sudo ./09-network-latency.sh
```

2. For "Cannot find device eth0@if10713" error:
   - **FIXED** in script version - interface name now correctly stripped to `eth0`
   - The script now uses `cut -d'@' -f1` to remove the @ifXXXX suffix

3. For "netem" errors (Experiment 09):
```bash
# Check if netem module is available:
lsmod | grep sch_netem

# Try to load it:
sudo modprobe sch_netem

# If module not found:
# Your kernel doesn't have CONFIG_NET_SCH_NETEM compiled in
# You need a different kernel or system with netem support
```

4. Install required tools:
```bash
sudo dnf install iproute-tc iptables
```

**Note on Experiment 09 (Network Latency):**
This experiment requires the `sch_netem` kernel module. Some kernels (especially minimal or hardened kernels) don't include this module. If your kernel doesn't support it, you can:
- Use a different system/VM with full kernel support
- Boot a different kernel with netem support
- Skip this experiment (you can still learn the concepts from the script)

---

### Issue 10: Cleanup Doesn't Run After Ctrl+C

**Symptom:**
After interrupting experiment with Ctrl+C, system is left in broken state.

**Root Cause:**
The `trap cleanup EXIT` is not catching the interrupt signal.

**Solution:**
1. Each script has cleanup function that runs automatically on exit
2. If cleanup didn't run, manually run it:

```bash
# For network experiments (02, 09):
# Remove tc rules
docker exec trading-api tc qdisc del dev eth0 root 2>/dev/null || true

# For iptables rules (02):
# Flush iptables
docker exec trading-api iptables -F 2>/dev/null || true

# For all experiments:
# Restart all services
docker compose restart
```

---

## Getting Help

If you encounter an issue not listed here:

1. **Check experiment logs:**
```bash
ls -lt /tmp/chaos-exp-*.log | head -1
cat $(ls -lt /tmp/chaos-exp-*.log | head -1 | awk '{print $NF}')
```

2. **Check BPF traces (if applicable):**
```bash
cat /tmp/chaos-exp-*.log.bpf
```

3. **Check Docker logs:**
```bash
docker logs simulated-exchange-trading-api --tail 100
docker logs simulated-exchange-postgres --tail 100
```

4. **Check system state:**
```bash
docker compose ps
docker ps -a
```

5. **Reset everything:**
```bash
docker compose restart
# Wait 30 seconds for services to stabilize
sleep 30
```

---

## Best Practices

### Before Running Experiments

1. ✅ Ensure all required services are running
2. ✅ Stop order-flow-simulator
3. ✅ Check you have sudo access (for experiments that need it)
4. ✅ Have at least 1GB free disk space for logs
5. ✅ Note the current system state (baseline)

### During Experiments

1. ✅ Let the script run to completion
2. ✅ Don't run multiple experiments simultaneously
3. ✅ Monitor the output for unexpected errors
4. ✅ If you must interrupt, use Ctrl+C (not kill)

### After Experiments

1. ✅ Review the generated logs
2. ✅ Check all services recovered
3. ✅ Verify no lingering tc/iptables rules
4. ✅ Archive logs for later analysis
5. ✅ Document any unexpected findings

---

## Quick Diagnostic Script

Create this script to quickly check system readiness:

```bash
#!/bin/bash
# chaos-preflight-check.sh

echo "=== Chaos Experiments Preflight Check ==="
echo ""

# Check Docker
if command -v docker &> /dev/null; then
    echo "✓ Docker installed: $(docker --version | cut -d' ' -f3)"
else
    echo "✗ Docker NOT installed"
fi

# Check services
echo ""
echo "Service Status:"
for svc in postgres redis trading-api nginx; do
    if docker ps | grep -q "simulated-exchange-$svc"; then
        echo "  ✓ $svc: Running"
    else
        echo "  ✗ $svc: NOT running"
    fi
done

# Check order flow simulator
echo ""
if docker ps | grep -q "order-flow-simulator"; then
    echo "⚠ order-flow-simulator is RUNNING (should be stopped)"
else
    echo "✓ order-flow-simulator is stopped"
fi

# Check bpftrace
echo ""
if command -v bpftrace &> /dev/null; then
    echo "✓ bpftrace installed: $(bpftrace --version 2>&1 | head -1)"
else
    echo "⚠ bpftrace NOT installed (optional but recommended)"
fi

# Check sudo access
echo ""
if sudo -n true 2>/dev/null; then
    echo "✓ sudo access available"
else
    echo "⚠ sudo access may require password"
fi

# Check disk space
echo ""
AVAILABLE=$(df /tmp | tail -1 | awk '{print $4}')
if [ $AVAILABLE -gt 1000000 ]; then
    echo "✓ Disk space: ${AVAILABLE}K available"
else
    echo "⚠ Low disk space: ${AVAILABLE}K available"
fi

echo ""
echo "=== Preflight Check Complete ==="
```

Save as `/usr/local/bin/chaos-preflight-check` and run before experiments.

---

## FAQ

**Q: Can I run experiments in production?**
A: NO! These experiments intentionally break things. Only run in dev/test/staging.

**Q: Why do some experiments need root?**
A: BPF tracing, network manipulation (tc, iptables), and some container operations require root privileges.

**Q: How long do experiments take?**
A: 5-15 minutes each. Plan for 2 hours to run all 10 experiments sequentially.

**Q: What if an experiment leaves the system broken?**
A: Each script has automatic cleanup. If that fails, run `docker compose restart`.

**Q: Can I customize the experiments?**
A: Yes! Each script is well-commented. Modify timeouts, delays, and parameters as needed.

**Q: Do I need all 110K orders in the database?**
A: No. Experiments work with any amount of data. Having realistic data helps validate performance under load.
