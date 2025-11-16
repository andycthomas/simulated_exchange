# Performance Tiers - Quick Start Guide

This guide explains how to easily run the Simulated Exchange at different performance tiers without manual reconfiguration.

## Overview

The system supports three performance tiers, each optimized for different use cases:

| Tier | Orders/Sec | Latency | CPU Cores | Memory | Use Case |
|------|-----------|---------|-----------|---------|----------|
| **Light** | 1,000 | <20ms | 4 | 2-4 GB | Development, initial deployment |
| **Average** | 5,000 | <50ms | 8 | 8-12 GB | Standard production workload |
| **Maximum** | 12,000 | <150ms | 16 | 20-24 GB | Peak events, stress testing |

## Quick Start

### Method 1: Using Makefile (Recommended)

The easiest way to run at different tiers:

```bash
# Run with light tier (1,000 ops/sec)
make run-light

# Run with average tier (5,000 ops/sec) - DEFAULT
make run-average
make run  # Same as run-average

# Run with maximum tier (12,000 ops/sec)
make run-maximum
```

### Method 2: Using Shell Script

For more control and information:

```bash
# Run with specific tier
./scripts/run-tier.sh light
./scripts/run-tier.sh average
./scripts/run-tier.sh maximum

# Show tier information
./scripts/run-tier.sh info
./scripts/run-tier.sh info light

# Compare all tiers
./scripts/run-tier.sh compare
```

### Method 3: Manual Configuration

Load environment variables directly:

```bash
# Light tier
export $(cat .env.light | grep -v '^#' | xargs)
export GOMAXPROCS=4
go run cmd/api/main.go

# Average tier
export $(cat .env.average | grep -v '^#' | xargs)
export GOMAXPROCS=8
go run cmd/api/main.go

# Maximum tier
export $(cat .env.maximum | grep -v '^#' | xargs)
export GOMAXPROCS=16
go run cmd/api/main.go
```

## Configuration Files

Three environment files are pre-configured:

- **`.env.light`** - Tier 1: Light/Minimum configuration
- **`.env.average`** - Tier 2: Standard/Average configuration
- **`.env.maximum`** - Tier 3: Peak/Maximum configuration

### Key Configuration Differences

| Parameter | Light | Average | Maximum |
|-----------|-------|---------|---------|
| `SIMULATION_ORDERS_PER_SEC` | 100 | 500 | 1000 |
| `SIMULATION_WORKERS` | 4 | 8 | 16 |
| `DB_MAX_CONNECTIONS` | 25 | 50 | 100 |
| `GOMAXPROCS` | 4 | 8 | 16 |
| `SIMULATION_SYMBOLS` | 3 symbols | 5 symbols | 8 symbols |
| `LOG_LEVEL` | info | warn | warn |

## Performance Characteristics

### Tier 1: Light (Minimum)

**Target Capacity:**
- 500-1,000 orders/second
- 250-750 trades/second
- 100-250 concurrent users
- <20ms P99 latency

**Resource Usage:**
- CPU: 4 cores at 25% utilization
- Memory: 2-4 GB
- ~500 goroutines
- 25 database connections

**Use Cases:**
- Initial production rollout
- Development and staging environments
- Cost-sensitive deployments
- Low-volume trading pairs

**Key Benefits:**
- Rock-solid stability
- Sub-20ms latency guaranteed
- 99.99% uptime SLA achievable
- 4x capacity buffer for spikes

---

### Tier 2: Average (Standard Production)

**Target Capacity:**
- 5,000 orders/second
- 3,750 trades/second
- 1,000 concurrent users
- <50ms P99 latency
- **432 million orders/day**

**Resource Usage:**
- CPU: 8 cores at 50% utilization
- Memory: 8-12 GB
- ~2,500 goroutines
- 50 database connections

**Use Cases:**
- Primary production deployment
- Multi-symbol trading (5-10 active pairs)
- Standard institutional trading
- 24/7 continuous operations

**Economic Impact:**
- Daily Trading Volume: $2.16 billion
- Annual Trading Volume: $788 billion
- Revenue Potential: $788M - $2.36B (0.1-0.3% fee)

**Key Benefits:**
- High performance with stability
- Excellent cost-to-performance ratio
- 2x capacity headroom for bursts
- 99.95% availability target

---

### Tier 3: Maximum (Peak Performance)

**Target Capacity:**
- 12,000 orders/second
- 9,000 trades/second
- 2,500 concurrent users
- <150ms P99 latency
- **1.04 billion orders/day**
- 15,000 ops/sec burst for 30 seconds

**Resource Usage:**
- CPU: 16 cores at 85% utilization
- Memory: 20-24 GB
- ~6,000 goroutines
- 100 database connections

**Use Cases:**
- Major market events (earnings, Fed announcements)
- Flash trading competitions
- High-frequency trading (HFT) scenarios
- Stress testing and chaos engineering
- Product launches, token sales

**Economic Impact:**
- Daily Trading Volume: $5.2 billion
- Annual Trading Volume: $1.89 trillion
- Revenue Potential: $1.89B - $5.67B (0.1-0.3% fee)
- Peak Hour Capacity: $216M in trading volume

**Key Benefits:**
- Maximum throughput on single instance
- Handles flash trading events
- Graceful degradation under overload
- Auto-scaling triggers enabled

## Switching Between Tiers

### During Development

Simply stop the current instance (Ctrl+C) and start with a different tier:

```bash
# Start with light tier for testing
make run-light

# Stop (Ctrl+C), then start with maximum tier for stress testing
make run-maximum
```

### In Production

For zero-downtime switching:

1. Start new instance with desired tier configuration
2. Update load balancer to route traffic to new instance
3. Gracefully shutdown old instance

## Verification

### Check Current Performance

After starting the server, verify the tier is working correctly:

```bash
# Check metrics endpoint
curl http://localhost:8080/api/metrics | jq

# Expected metrics for each tier:
# Light:   orders_per_sec ~100,  avg_latency <20ms
# Average: orders_per_sec ~500,  avg_latency <50ms
# Maximum: orders_per_sec ~1000, avg_latency <150ms
```

### Monitor System Resources

```bash
# CPU usage (should match tier allocation)
top -bn1 | grep "Cpu(s)"

# Memory usage
free -h

# Goroutine count
curl http://localhost:8080/api/health | jq .goroutines
```

### Run Load Tests

Use the demo system to validate performance:

```bash
# Light tier validation
curl -X POST http://localhost:8080/demo/load-test \
  -H "Content-Type: application/json" \
  -d '{"intensity": "light", "duration": 60, "orders_per_second": 1000}'

# Average tier validation
curl -X POST http://localhost:8080/demo/load-test \
  -H "Content-Type: application/json" \
  -d '{"intensity": "medium", "duration": 120, "orders_per_second": 5000}'

# Maximum tier validation
curl -X POST http://localhost:8080/demo/load-test \
  -H "Content-Type: application/json" \
  -d '{"intensity": "extreme", "duration": 90, "orders_per_second": 12000}'
```

## Customization

### Modifying Tier Configurations

Edit the environment files to adjust parameters:

```bash
# Edit light tier configuration
nano .env.light

# Edit average tier configuration
nano .env.average

# Edit maximum tier configuration
nano .env.maximum
```

### Creating Custom Tiers

Create your own tier configuration:

```bash
# Copy existing tier as template
cp .env.average .env.custom

# Edit as needed
nano .env.custom

# Run with custom configuration
export $(cat .env.custom | grep -v '^#' | xargs)
export GOMAXPROCS=12
go run cmd/api/main.go
```

## Troubleshooting

### Performance Not Meeting Expectations

**Problem:** Actual throughput is lower than tier target

**Solutions:**
1. Check CPU availability: `lscpu` (ensure you have enough cores)
2. Check memory: `free -h` (ensure sufficient RAM)
3. Verify GOMAXPROCS: `echo $GOMAXPROCS`
4. Check for resource contention: `top` or `htop`

### High Latency

**Problem:** Latency exceeds tier target (e.g., >50ms on average tier)

**Solutions:**
1. Reduce concurrent load
2. Check database connection pool settings
3. Verify no background processes consuming resources
4. Consider scaling to higher tier

### Out of Memory

**Problem:** Application crashes with OOM error

**Solutions:**
1. Reduce to lower tier (e.g., maximum → average)
2. Decrease `SIMULATION_WORKERS`
3. Reduce `DB_MAX_CONNECTIONS`
4. Set `GOMEMLIMIT` (e.g., `export GOMEMLIMIT=28GB`)

### Port Already in Use

**Problem:** "address already in use" error

**Solutions:**
```bash
# Find process using port 8080
sudo lsof -i :8080

# Kill existing process
kill -9 <PID>

# Or use different port
export SERVER_PORT=8081
make run-average
```

## Best Practices

### Development Workflow

```bash
# Start with light tier for development
make run-light

# Test with average tier before deploying
make run-average

# Stress test with maximum tier
make run-maximum
```

### Production Deployment

1. **Start Conservative:** Begin with light tier
2. **Monitor & Scale:** Watch metrics and scale to average tier when needed
3. **Peak Planning:** Switch to maximum tier for expected high-traffic events
4. **Load Testing:** Always test configuration changes in staging first

### Performance Testing

```bash
# 1. Compare tier performance
./scripts/run-tier.sh compare

# 2. Start desired tier
make run-average

# 3. Run load test
curl -X POST http://localhost:8080/demo/load-test \
  -d '{"intensity": "medium", "duration": 300, "orders_per_second": 5000}'

# 4. Monitor metrics
watch -n 2 'curl -s http://localhost:8080/api/metrics | jq ".data | {orders_per_sec, avg_latency, p99_latency}"'
```

## Additional Resources

- **Full Scalability Analysis:** See `docs/SCALABILITY_ANALYSIS.md`
- **Architecture Details:** See `docs/ARCHITECTURE.md`
- **SRE Runbook:** See `docs/SRE_RUNBOOK.md`
- **API Documentation:** See `README.md`

## Quick Reference

```bash
# View all available make targets
make help

# Run with specific tier
make run-light    # 1,000 ops/sec
make run-average  # 5,000 ops/sec (default)
make run-maximum  # 12,000 ops/sec

# Compare tiers
./scripts/run-tier.sh compare

# Check tier configuration
./scripts/run-tier.sh info light
./scripts/run-tier.sh info average
./scripts/run-tier.sh info maximum

# Monitor performance
curl http://localhost:8080/api/metrics | jq
curl http://localhost:8080/api/health | jq
```

---

**Last Updated:** 2025-10-26
**Version:** 1.0
