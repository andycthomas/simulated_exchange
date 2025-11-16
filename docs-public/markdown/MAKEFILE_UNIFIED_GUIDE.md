# Unified Makefile Guide

**Date:** 2025-10-26
**Changes:** Merged `Makefile` and `Makefile.demo` into a single unified `Makefile`

## What Changed?

### Before (Two Makefiles)
- **`Makefile`**: Development, testing, CI/CD, foreground run commands
- **`Makefile.demo`**: Demo orchestration, chaos experiments, background management

**Problems:**
- Confusing - two entry points for same tasks
- `make run` vs `make -f Makefile.demo start-background`
- Log watching didn't work unless started with exact right command
- Multiple instances running from different start methods

### After (Unified Makefile)
- **One `Makefile`** with all functionality organized in 10 clear sections
- Consistent command interface
- All commands use `make <target>` (no more `-f Makefile.demo`)
- Fixed logging and monitoring

## Quick Start

### 1. Start the Application
```bash
# Start with average performance tier (recommended)
make start-average

# Or choose a different tier
make start-light      # 1k ops/sec, 4 cores
make start-maximum    # 12k ops/sec, 16 cores
```

### 2. Monitor
```bash
# Check status
make status

# Watch live metrics (Ctrl+C to stop)
make watch-metrics

# Watch logs (Ctrl+C to stop)
make watch-logs

# Check health
make health
```

### 3. Stop
```bash
make stop
```

## Common Workflows

### Development Workflow
```bash
# Setup dependencies
make deps

# Run tests
make test-unit
make test-e2e
make test-all

# Start app for development (foreground)
make run              # or run-average-fg, run-light-fg, run-maximum-fg

# In another terminal, test endpoints
curl http://localhost:8080/health
```

### Demo/Presentation Workflow
```bash
# Start app in background
make start-average

# Run a quick demo
make demo-trading
make demo-metrics

# Run complete automated demo (30 min)
make demo-full

# Stop when done
make stop
```

### Load Testing Workflow
```bash
# Start app
make start-maximum

# Run load tests
make demo-loadtest-light    # 25 ops/sec
make demo-loadtest-medium   # 100 ops/sec
make demo-loadtest-heavy    # 200 ops/sec

# Monitor metrics
make watch-metrics
```

### Chaos Engineering Workflow
```bash
# Start app
make start-average

# Run individual chaos experiments
make chaos-latency
make chaos-db-death
make chaos-network-partition

# Or run all experiments (90-120 min)
make chaos-all

# View chaos logs
make chaos-logs
```

## All Available Commands

### Application Management
| Command | Description |
|---------|-------------|
| `make start-light` | Start with light tier (1k ops/sec, background) |
| `make start-average` | Start with average tier (5k ops/sec, background) |
| `make start-maximum` | Start with maximum tier (12k ops/sec, background) |
| `make stop` | Stop the application gracefully |
| `make restart-light` | Restart with light tier |
| `make restart-average` | Restart with average tier |
| `make restart-maximum` | Restart with maximum tier |
| `make status` | Check if app is running |
| `make health` | Check system health |

### Monitoring
| Command | Description |
|---------|-------------|
| `make watch-metrics` | Live metrics updates (every 5s) |
| `make watch-health` | Live health check (every 2s) |
| `make watch-logs` | Tail application logs |
| `make dashboard` | Open web dashboard in browser |
| `make metrics-snapshot` | Save current metrics to file |

### Demos
| Command | Description |
|---------|-------------|
| `make demo-full` | Complete automated demo (30 min) |
| `make demo-interactive` | Interactive guided demo |
| `make demo-presentation` | Presentation-ready demo (45 min) |
| `make demo-trading` | Basic trading operations |
| `make demo-metrics` | Real-time metrics display |
| `make demo-loadtest-light` | Light load test (25 ops/sec) |
| `make demo-loadtest-medium` | Medium load test (100 ops/sec) |
| `make demo-loadtest-heavy` | Heavy load test (200 ops/sec) |
| `make demo-chaos-latency` | Chaos: latency injection |
| `make demo-chaos-errors` | Chaos: error simulation |

### Chaos Experiments
| Command | Description |
|---------|-------------|
| `make chaos-all` | Run all 10 chaos experiments |
| `make chaos-db-death` | Database sudden death |
| `make chaos-network-partition` | Network partition |
| `make chaos-oom-kill` | Memory OOM kill |
| `make chaos-cpu-throttle` | CPU throttling |
| `make chaos-disk-io` | Disk I/O starvation |
| `make chaos-conn-pool` | Connection pool exhaustion |
| `make chaos-redis` | Redis cache failure |
| `make chaos-nginx` | Nginx proxy failure |
| `make chaos-latency` | Network latency injection |
| `make chaos-cascade` | Multi-service cascade |
| `make chaos-logs` | View chaos experiment logs |

### Testing & Development
| Command | Description |
|---------|-------------|
| `make test-all` | Run all tests (lint, unit, e2e) |
| `make test-unit` | Run unit tests with coverage |
| `make test-e2e` | Run E2E tests |
| `make test-demo` | Run demo system tests |
| `make test-perf` | Run performance tests |
| `make benchmark` | Run benchmarks |
| `make lint` | Run code linter |
| `make fmt` | Format code |
| `make vet` | Run go vet |
| `make security` | Run security checks |
| `make coverage` | Generate coverage report |

### Foreground Run (Development)
| Command | Description |
|---------|-------------|
| `make run` | Run with average tier (foreground) |
| `make run-light-fg` | Run with light tier (foreground) |
| `make run-average-fg` | Run with average tier (foreground) |
| `make run-maximum-fg` | Run with maximum tier (foreground) |

### Docker & CI/CD
| Command | Description |
|---------|-------------|
| `make docker-build` | Build Docker test image |
| `make docker-test-unit` | Run unit tests in Docker |
| `make docker-test-e2e` | Run E2E tests with Docker Compose |
| `make docker-test-all` | Run comprehensive Docker test suite |
| `make ci-test` | Run CI test suite |
| `make ci-e2e` | Run CI E2E tests |
| `make ci-full` | Run full CI suite |

### Utilities
| Command | Description |
|---------|-------------|
| `make clean` | Clean temporary files and logs |
| `make clean-chaos` | Clean chaos experiment artifacts |
| `make clean-docker` | Clean Docker images and containers |
| `make deps` | Download and verify dependencies |
| `make install-deps` | Install system dependencies |
| `make install-bpf` | Install BPF/eBPF tools |
| `make check-prereqs` | Check all prerequisites |
| `make dev-setup` | Set up development environment |
| `make version` | Show version information |
| `make help` | Show comprehensive help |

## Performance Tiers

### Light Tier
- **Target:** 1,000 orders/sec
- **Latency:** <20ms P99
- **Resources:** 4 cores, 2-4GB RAM
- **Use case:** Development, testing, demos

### Average Tier (Default)
- **Target:** 5,000 orders/sec
- **Latency:** <50ms P99
- **Resources:** 8 cores, 8-12GB RAM
- **Daily Volume:** 432M orders, $2.16B trading
- **Use case:** Standard production, most demos

### Maximum Tier
- **Target:** 12,000 orders/sec
- **Latency:** <150ms P99
- **Resources:** 16 cores, 20-24GB RAM
- **Daily Volume:** 1.04B orders, $5.2B trading
- **Use case:** Peak load, stress testing

## File Locations

- **Application logs:** `/tmp/simulated-exchange.log`
- **PID file:** `/tmp/simulated-exchange.pid`
- **Metrics snapshots:** `/tmp/metrics-snapshot-*.json`
- **Chaos logs:** `/tmp/chaos-logs/`
- **Test results:** `test-output/`

## Backward Compatibility

The unified Makefile maintains backward compatibility:

- `make start-background` → still works (alias for `make start-average`)
- `make start` → still works (alias for `make start-average`)
- All old `Makefile` test targets still work
- All old `Makefile.demo` targets still work

## Migration Notes

### Old Makefiles Backed Up
- `Makefile.backup.20251026-133000` - Original Makefile
- `Makefile.demo.backup.20251026-133000` - Original Makefile.demo
- `Makefile.demo.old` - Renamed from Makefile.demo

### If You Need to Rollback
```bash
# Restore original Makefile
mv Makefile Makefile.unified
cp Makefile.backup.20251026-133000 Makefile
cp Makefile.demo.backup.20251026-133000 Makefile.demo
```

## Troubleshooting

### App won't start
```bash
# Check if port 8080 is in use
lsof -i :8080

# Check logs
tail -f /tmp/simulated-exchange.log

# Verify .env files exist
ls -la .env.*
```

### Watch-logs not working
```bash
# Make sure app was started with make start-* (not make run)
make stop
make start-average

# Now logs will be at /tmp/simulated-exchange.log
make watch-logs
```

### Can't find Makefile.demo commands
All commands are now in the main `Makefile`. Remove `-f Makefile.demo` from your commands:
```bash
# Old way
make -f Makefile.demo start-background
make -f Makefile.demo watch-metrics

# New way
make start-average
make watch-metrics
```

## Tips

1. **Always use `make help`** to see all available commands
2. **Use background mode** (`make start-*`) for demos and monitoring
3. **Use foreground mode** (`make run-*-fg`) for development with visible logs
4. **Check status often** with `make status` and `make health`
5. **Clean up** with `make stop` when done

## Examples

### Quick Demo
```bash
make start-average
make demo-trading
make demo-metrics
make stop
```

### Load Test
```bash
make start-maximum
make demo-loadtest-medium &
make watch-metrics
# Ctrl+C to stop watching
make stop
```

### Development
```bash
make test-unit
make run  # foreground, see logs directly
# Ctrl+C to stop
```

### Full Testing Suite
```bash
make clean
make test-all
make coverage
make benchmark
```

---

**For more help:** Run `make help` to see all commands with descriptions
