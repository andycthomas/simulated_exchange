# 🔧 Troubleshooting Guide
**Simulated Exchange Platform - Issue Resolution & Support**

*Comprehensive troubleshooting guide for common issues, performance problems, and system recovery*

---

## 📋 Quick Issue Resolution

### 🚨 Emergency Response Checklist

When experiencing system issues, follow this priority checklist:

1. **Check System Health**: `curl http://localhost:8080/health`
2. **Verify Services**: `docker-compose ps` or `kubectl get pods`
3. **Review Logs**: `docker-compose logs -f` or application logs
4. **Check Resources**: CPU, memory, disk space availability
5. **Validate Network**: Connectivity and port accessibility
6. **Reset if Needed**: Clean restart with data preservation

### 🎯 Common Issue Categories

| Issue Type | Symptoms | Quick Fix | Section |
|------------|----------|-----------|---------|
| **Startup Problems** | Service won't start | [Startup Issues](#startup-issues) | §1 |
| **Performance Degradation** | Slow responses | [Performance Issues](#performance-issues) | §2 |
| **Connection Problems** | API/WebSocket failures | [Connectivity Issues](#connectivity-issues) | §3 |
| **Demo System Issues** | Load tests fail | [Demo System Problems](#demo-system-problems) | §4 |
| **Docker Problems** | Container issues | [Docker Issues](#docker-issues) | §5 |
| **Testing Failures** | Tests not passing | [Testing Issues](#testing-issues) | §6 |

---

## 🚀 Section 1: Startup Issues

### Issue 1.1: Application Won't Start

#### Symptoms
- `go run cmd/api/main.go` fails
- "Address already in use" errors
- Configuration loading failures

#### Diagnosis Steps
```bash
# Check if port is already in use
lsof -i :8080
netstat -tulpn | grep :8080

# Verify Go environment
go version
go env GOOS GOARCH

# Check for missing dependencies
go mod tidy
go mod download
```

#### Solutions

**Port Conflict Resolution:**
```bash
# Kill process using port 8080
sudo kill $(lsof -t -i:8080)

# Or use alternative port
export PORT=8081
go run cmd/api/main.go
```

**Environment Configuration:**
```bash
# Set required environment variables
export GO_ENV=development
export LOG_LEVEL=info
export DATABASE_URL=memory://simulated.db

# Verify configuration
go run cmd/api/main.go --config-check
```

**Dependency Issues:**
```bash
# Clean and rebuild
go clean -cache
go mod tidy
go mod download
go build ./...
```

### Issue 1.2: Docker Startup Failures

#### Symptoms
- `docker-compose up` fails
- Container exits immediately
- Image build failures

#### Diagnosis Steps
```bash
# Check Docker status
docker --version
docker-compose --version

# Inspect container logs
docker-compose logs app

# Check image build
docker-compose build --no-cache
```

#### Solutions

**Container Build Issues:**
```bash
# Rebuild from scratch
docker-compose down -v
docker system prune -f
docker-compose build --no-cache
docker-compose up -d
```

**Permission Problems:**
```bash
# Fix file permissions
chmod +x scripts/*.sh
chmod 644 config/*.yml

# Rebuild and restart
docker-compose up --build
```

**Resource Constraints:**
```bash
# Check Docker resources
docker system df
docker stats

# Increase Docker memory/CPU if needed
# Docker Desktop: Settings → Resources
```

### Issue 1.3: Configuration Problems

#### Symptoms
- "Config file not found" errors
- Invalid configuration values
- Environment variable issues

#### Diagnosis Steps
```bash
# Verify config file existence
ls -la config/
cat config/config.yaml

# Check environment variables
env | grep -E "(PORT|LOG_LEVEL|DATABASE_URL)"

# Validate configuration
go run cmd/api/main.go --validate-config
```

#### Solutions

**Missing Configuration:**
```bash
# Create default config
mkdir -p config
cat > config/config.yaml << 'EOF'
server:
  port: 8080
  environment: development
demo:
  enabled: true
  websocket_port: 8081
metrics:
  enabled: true
  collection_interval: 1s
EOF
```

**Environment Variables:**
```bash
# Set all required variables
export PORT=8080
export LOG_LEVEL=info
export GO_ENV=development
export DATABASE_URL=memory://simulated.db
export DEMO_ENABLED=true
export WEBSOCKET_PORT=8081
```

---

## ⚡ Section 2: Performance Issues

### Issue 2.1: High Latency

#### Symptoms
- API responses > 100ms
- Slow order processing
- WebSocket connection delays

#### Diagnosis Steps
```bash
# Check current performance
curl -w "@curl-format.txt" http://localhost:8080/api/orders

# Create curl timing format file
cat > curl-format.txt << 'EOF'
     time_namelookup:  %{time_namelookup}\n
        time_connect:  %{time_connect}\n
     time_appconnect:  %{time_appconnect}\n
    time_pretransfer:  %{time_pretransfer}\n
       time_redirect:  %{time_redirect}\n
  time_starttransfer:  %{time_starttransfer}\n
                     ----------\n
          time_total:  %{time_total}\n
EOF

# Monitor system resources
top -p $(pgrep -f "cmd/api/main.go")
iostat -x 1 5
```

#### Solutions

**Application-Level Optimization:**
```bash
# Enable Go profiling
go run cmd/api/main.go -cpuprofile=cpu.prof

# Analyze CPU profile
go tool pprof cpu.prof
# Commands: top, list, web

# Memory profiling
go run cmd/api/main.go -memprofile=mem.prof
go tool pprof mem.prof
```

**Configuration Tuning:**
```bash
# Adjust Go runtime settings
export GOGC=100
export GOMAXPROCS=4

# Database optimization
export MAX_CONCURRENT_ORDERS=500
export ORDER_TIMEOUT=15s
```

**System-Level Optimization:**
```bash
# Increase file descriptor limits
ulimit -n 65536

# Optimize TCP settings
echo 'net.core.somaxconn = 65535' | sudo tee -a /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

### Issue 2.2: Memory Leaks

#### Symptoms
- Gradually increasing memory usage
- System becomes unresponsive
- Out of memory errors

#### Diagnosis Steps
```bash
# Monitor memory usage over time
while true; do
  ps aux | grep "cmd/api/main.go" | grep -v grep
  free -h
  sleep 30
done

# Go memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap
# Commands: top, list, svg
```

#### Solutions

**Memory Leak Detection:**
```bash
# Enable memory profiling
go run cmd/api/main.go -memprofile=mem.prof

# Compare memory profiles over time
go tool pprof -base=mem1.prof mem2.prof
```

**Resource Management:**
```bash
# Implement graceful shutdown
# Add to main.go:
c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)
go func() {
    <-c
    // Cleanup code here
    os.Exit(0)
}()
```

**Garbage Collection Tuning:**
```bash
# Adjust GC settings
export GOGC=50  # More frequent GC
export GODEBUG=gctrace=1  # GC tracing

# Force garbage collection
curl http://localhost:8080/debug/pprof/heap?gc=1
```

### Issue 2.3: CPU Usage Problems

#### Symptoms
- High CPU utilization (>80%)
- System sluggishness
- Increased response times

#### Diagnosis Steps
```bash
# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# System CPU analysis
top -H -p $(pgrep -f "cmd/api/main.go")
perf top -p $(pgrep -f "cmd/api/main.go")
```

#### Solutions

**CPU Optimization:**
```bash
# Analyze hotspots
go tool pprof -web cpu.prof

# Check goroutine usage
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

**Concurrency Tuning:**
```bash
# Limit concurrent operations
export MAX_CONCURRENT_ORDERS=100
export GOMAXPROCS=2

# Implement rate limiting
# Add to API handlers:
rateLimiter := rate.NewLimiter(100, 200) // 100 req/sec, burst 200
```

---

## 🌐 Section 3: Connectivity Issues

### Issue 3.1: API Connection Failures

#### Symptoms
- "Connection refused" errors
- Timeouts on API calls
- Intermittent connectivity

#### Diagnosis Steps
```bash
# Test basic connectivity
curl -v http://localhost:8080/health
telnet localhost 8080

# Check network configuration
netstat -tulpn | grep :8080
ss -tulpn | grep :8080

# Firewall status
sudo ufw status
sudo iptables -L
```

#### Solutions

**Port and Firewall Configuration:**
```bash
# Open required ports
sudo ufw allow 8080/tcp
sudo ufw allow 8081/tcp
sudo ufw allow 9090/tcp

# Check Docker network
docker network ls
docker network inspect simulated_exchange_default
```

**Service Binding Issues:**
```bash
# Bind to all interfaces
export HOST=0.0.0.0
export PORT=8080

# Verify binding
netstat -tulpn | grep :8080
```

**DNS and Routing:**
```bash
# Test localhost resolution
nslookup localhost
ping localhost

# Check routing
ip route show
```

### Issue 3.2: WebSocket Connection Problems

#### Symptoms
- WebSocket connection drops
- Real-time updates not working
- Browser console errors

#### Diagnosis Steps
```bash
# Test WebSocket connection
wscat -c ws://localhost:8081/ws/demo

# Check WebSocket logs
docker-compose logs websocket-service

# Browser developer tools
# Check Network tab for WebSocket status
```

#### Solutions

**WebSocket Server Configuration:**
```bash
# Verify WebSocket service
curl http://localhost:8081/health

# Check CORS settings
# Add to WebSocket server:
c.Header("Access-Control-Allow-Origin", "*")
c.Header("Access-Control-Allow-Methods", "GET")
c.Header("Access-Control-Allow-Headers", "Origin, Content-Type")
```

**Browser Configuration:**
```bash
# Test with different client
npm install -g wscat
wscat -c ws://localhost:8081/ws/demo

# Check browser console for errors
# Disable browser extensions if needed
```

**Proxy and Load Balancer Issues:**
```bash
# Direct connection test
curl --upgrade ws://localhost:8081/ws/demo

# Check for proxy interference
unset http_proxy https_proxy
```

### Issue 3.3: Database Connection Issues

#### Symptoms
- Database connection errors
- Query timeouts
- Transaction failures

#### Diagnosis Steps
```bash
# Check database status
curl http://localhost:8080/api/metrics | jq '.database_status'

# Test database connectivity
# Add to application:
db.Ping()
```

#### Solutions

**In-Memory Database Issues:**
```bash
# Reset database
curl -X POST http://localhost:8080/demo/reset

# Check memory usage
free -h
```

**Connection Pool Configuration:**
```go
// Optimize connection settings
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## 🎭 Section 4: Demo System Problems

### Issue 4.1: Load Test Failures

#### Symptoms
- Load tests won't start
- Tests stop unexpectedly
- Incorrect metrics reporting

#### Diagnosis Steps
```bash
# Check demo system status
curl http://localhost:8080/demo/status | jq

# Verify load test configuration
curl http://localhost:8080/demo/scenarios | jq

# Monitor during test
curl http://localhost:8080/api/metrics | jq '.load_test'
```

#### Solutions

**Load Test Configuration:**
```bash
# Reset demo system
curl -X POST http://localhost:8080/demo/reset

# Start with light load
curl -X POST http://localhost:8080/demo/load-test \
  -H "Content-Type: application/json" \
  -d '{
    "scenario": "light",
    "duration": "30s",
    "orders_per_second": 5
  }'
```

**Resource Constraints:**
```bash
# Check system resources
free -h
df -h
uptime

# Reduce load test intensity
curl -X POST http://localhost:8080/demo/load-test \
  -H "Content-Type: application/json" \
  -d '{
    "scenario": "minimal",
    "duration": "15s",
    "orders_per_second": 1
  }'
```

**Safety Limits:**
```bash
# Check safety limits
curl http://localhost:8080/demo/config | jq '.safety_limits'

# Adjust if needed
export DEMO_MAX_ERROR_RATE=0.1
export DEMO_MAX_CPU_USAGE=80.0
```

### Issue 4.2: Chaos Engineering Problems

#### Symptoms
- Chaos tests won't start
- System doesn't recover
- Excessive performance impact

#### Diagnosis Steps
```bash
# Check chaos test status
curl http://localhost:8080/demo/chaos-status | jq

# Monitor system during chaos
top -p $(pgrep -f "cmd/api/main.go")
```

#### Solutions

**Safe Chaos Testing:**
```bash
# Start with minimal chaos
curl -X POST http://localhost:8080/demo/chaos-test \
  -H "Content-Type: application/json" \
  -d '{
    "type": "latency_injection",
    "duration": "10s",
    "severity": "low"
  }'

# Stop chaos immediately if needed
curl -X DELETE http://localhost:8080/demo/chaos-test
```

**Recovery Issues:**
```bash
# Force system reset
curl -X POST http://localhost:8080/demo/reset

# Restart if needed
docker-compose restart app
```

### Issue 4.3: WebSocket Demo Updates

#### Symptoms
- Real-time updates not showing
- WebSocket connection errors
- Delayed or missing data

#### Diagnosis Steps
```bash
# Test WebSocket connection
wscat -c ws://localhost:8081/ws/demo

# Check message flow
curl http://localhost:8080/api/metrics | jq '.websocket_connections'
```

#### Solutions

**WebSocket Connectivity:**
```bash
# Restart WebSocket service
docker-compose restart websocket-service

# Test with curl
curl --upgrade ws://localhost:8081/ws/demo
```

**Browser Issues:**
```javascript
// Test WebSocket in browser console
const ws = new WebSocket('ws://localhost:8081/ws/demo');
ws.onopen = () => console.log('Connected');
ws.onmessage = (event) => console.log('Message:', event.data);
ws.onerror = (error) => console.log('Error:', error);
```

---

## 🐳 Section 5: Docker Issues

### Issue 5.1: Container Build Failures

#### Symptoms
- Docker build fails
- Image build takes too long
- Out of disk space errors

#### Diagnosis Steps
```bash
# Check Docker status
docker version
docker info

# Inspect build process
docker-compose build --progress=plain --no-cache

# Check disk space
docker system df
df -h
```

#### Solutions

**Build Optimization:**
```bash
# Clean Docker system
docker system prune -a -f
docker volume prune -f

# Rebuild with no cache
docker-compose build --no-cache --parallel

# Multi-stage build optimization
# Ensure Dockerfile uses:
FROM golang:1.21-alpine AS builder
# ... build steps ...
FROM alpine:latest AS runtime
```

**Resource Management:**
```bash
# Increase Docker resources
# Docker Desktop: Settings → Resources
# Recommended: 4GB RAM, 2 CPUs

# Clean up unused images
docker image prune -a -f
```

### Issue 5.2: Container Runtime Issues

#### Symptoms
- Containers exit unexpectedly
- Service dependencies fail
- Health checks failing

#### Diagnosis Steps
```bash
# Check container status
docker-compose ps
docker-compose logs --tail=50

# Inspect specific container
docker inspect <container_id>
docker logs <container_id>
```

#### Solutions

**Service Dependencies:**
```bash
# Restart in correct order
docker-compose down
docker-compose up -d app
sleep 10
docker-compose up -d

# Check dependency order in docker-compose.yml
depends_on:
  - app
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
```

**Health Check Configuration:**
```yaml
# Optimize health checks
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### Issue 5.3: Network and Volume Issues

#### Symptoms
- Services can't communicate
- Data persistence problems
- Port binding failures

#### Diagnosis Steps
```bash
# Check Docker networks
docker network ls
docker network inspect simulated_exchange_default

# Inspect volumes
docker volume ls
docker volume inspect simulated_exchange_data
```

#### Solutions

**Network Configuration:**
```bash
# Recreate network
docker-compose down
docker network prune -f
docker-compose up -d

# Manual network creation
docker network create simulated_exchange_network
```

**Volume Management:**
```bash
# Reset volumes
docker-compose down -v
docker volume prune -f
docker-compose up -d

# Backup important data first
docker cp container_name:/data ./backup/
```

---

## 🧪 Section 6: Testing Issues

### Issue 6.1: Unit Test Failures

#### Symptoms
- Tests fail unexpectedly
- Race condition errors
- Coverage drops

#### Diagnosis Steps
```bash
# Run tests with verbose output
go test -v ./...

# Check for race conditions
go test -race ./...

# Run specific failing test
go test -v -run TestSpecificFunction ./path/to/package
```

#### Solutions

**Race Condition Fixes:**
```bash
# Run with race detector
go test -race ./...

# Common fixes:
# - Use sync.RWMutex for shared data
# - Avoid shared global variables
# - Use channels for coordination
```

**Test Environment:**
```bash
# Set test environment
export GO_ENV=test
export LOG_LEVEL=error

# Clean test cache
go clean -testcache

# Run tests with timeout
go test -timeout=30s ./...
```

### Issue 6.2: E2E Test Problems

#### Symptoms
- E2E tests timing out
- Service dependencies failing
- Docker test environment issues

#### Diagnosis Steps
```bash
# Check test environment
make test-e2e-debug

# Monitor test execution
docker-compose -f docker-compose.test.yml logs -f

# Check test data
ls -la test/fixtures/
```

#### Solutions

**Test Environment Setup:**
```bash
# Reset test environment
make clean-test
make test-setup

# Run individual test suites
go test -v ./test/e2e/trading_flow_test.go
go test -v ./test/e2e/demo_flow_test.go
```

**Timing Issues:**
```bash
# Increase timeouts
export TEST_TIMEOUT=60s
export API_TIMEOUT=30s

# Add retry logic
# Implement exponential backoff in tests
```

### Issue 6.3: Performance Test Issues

#### Symptoms
- Benchmarks failing SLA targets
- Inconsistent performance results
- Resource exhaustion during tests

#### Diagnosis Steps
```bash
# Run performance tests
make test-perf

# Monitor resources during tests
top &
go test -bench=. -benchmem ./test/e2e/performance_test.go
```

#### Solutions

**Performance Optimization:**
```bash
# Optimize test environment
export GOMAXPROCS=4
export GOGC=100

# Run benchmarks multiple times
go test -bench=. -count=5 -benchmem ./test/e2e/
```

**Resource Management:**
```bash
# Limit concurrent tests
export MAX_CONCURRENT_TESTS=2

# Monitor during execution
go test -bench=. -cpuprofile=bench.prof ./test/e2e/
go tool pprof bench.prof
```

---

## 🚨 Section 7: Emergency Procedures

### Complete System Reset

When all else fails, perform a complete system reset:

```bash
#!/bin/bash
# Emergency reset script

echo "🚨 Starting emergency system reset..."

# Stop all services
docker-compose down -v
killall -9 go || true

# Clean Docker completely
docker system prune -a -f
docker volume prune -f
docker network prune -f

# Reset application state
rm -f *.prof *.log
rm -rf tmp/

# Rebuild from scratch
make clean
go clean -cache -modcache
go mod download

# Restart system
echo "🚀 Restarting system..."
docker-compose up --build -d

# Wait for health check
sleep 30
curl -f http://localhost:8080/health || echo "❌ Health check failed"

echo "✅ Emergency reset complete"
```

### Data Recovery

If data needs to be preserved during reset:

```bash
# Backup current state
mkdir -p backup/$(date +%Y%m%d_%H%M%S)
docker cp app_container:/data backup/$(date +%Y%m%d_%H%M%S)/
docker logs app_container > backup/$(date +%Y%m%d_%H%M%S)/app.log

# Reset and restore
docker-compose down -v
# ... reset steps ...
docker-compose up -d
docker cp backup/latest/data app_container:/
```

### Emergency Contacts

- **Engineering Team**: engineering-team@company.com
- **On-Call Support**: +1-555-SUPPORT
- **Emergency Escalation**: cto@company.com
- **Documentation**: https://docs.company.com/simulated-exchange

---

## 📊 Monitoring and Alerting

### Health Check Monitoring

Set up automated monitoring for critical endpoints:

```bash
#!/bin/bash
# health-monitor.sh
while true; do
  if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "🚨 ALERT: Health check failed at $(date)"
    # Send alert notification
  fi
  sleep 60
done
```

### Performance Monitoring

Monitor key performance metrics:

```bash
# performance-monitor.sh
while true; do
  LATENCY=$(curl -w '%{time_total}' -s -o /dev/null http://localhost:8080/api/orders)
  if (( $(echo "$LATENCY > 0.1" | bc -l) )); then
    echo "🐌 Performance degraded: ${LATENCY}s latency"
  fi
  sleep 30
done
```

### Log Analysis

Set up log monitoring for error patterns:

```bash
# log-monitor.sh
tail -f app.log | while read line; do
  if echo "$line" | grep -qi "error\|panic\|fatal"; then
    echo "🚨 Error detected: $line"
    # Process error
  fi
done
```

---

## 📞 Getting Help

### Self-Service Resources

1. **Documentation**: Check [README.md](../README.md) and [API.md](API.md)
2. **Architecture**: Review [ARCHITECTURE.md](ARCHITECTURE.md)
3. **Demo Guide**: Follow [DEMO_SCRIPT.md](DEMO_SCRIPT.md)
4. **Business Value**: See [BUSINESS_VALUE.md](BUSINESS_VALUE.md)

### Support Channels

1. **GitHub Issues**: Report bugs and feature requests
2. **Team Chat**: Internal team communication
3. **Email Support**: technical-support@company.com
4. **Emergency Hotline**: For critical production issues

### Escalation Path

1. **Level 1**: Team lead or senior developer
2. **Level 2**: Engineering manager
3. **Level 3**: Principal engineer or architect
4. **Level 4**: CTO or VP Engineering

---

## 🎯 Prevention Best Practices

### Proactive Monitoring

- Set up comprehensive health checks
- Monitor key performance metrics
- Implement automated alerting
- Regular performance benchmarking

### Code Quality

- Maintain 80%+ test coverage
- Run linting and security scans
- Perform regular code reviews
- Use race detection in testing

### Deployment Safety

- Use blue-green deployments
- Implement feature flags
- Maintain rollback procedures
- Test disaster recovery plans

### Documentation Maintenance

- Keep troubleshooting guide updated
- Document new issues and solutions
- Share knowledge across team
- Regular documentation reviews

---

**Remember**: When in doubt, don't hesitate to ask for help. It's better to escalate early than to risk system stability. 🚀

*This troubleshooting guide is maintained by the Performance Engineering Team. Last updated: 2024-01-15*