# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Simulated Exchange Platform is a high-performance, cloud-native trading exchange simulation system written in Go. It demonstrates performance testing, chaos engineering, real-time metrics collection, and market simulation capabilities. The system processes 2,500+ orders/second with sub-50ms P99 latency and 99.95% availability.

## Essential Commands

### Development & Testing

```bash
# Development setup
make deps                    # Download and verify dependencies
make dev-setup              # Install dev tools (golangci-lint, gosec)

# Quick development cycle
make dev-test               # Fast unit tests only
go run cmd/api/main.go      # Start the application locally

# Testing
make test-unit              # Unit tests with coverage
make test-e2e               # End-to-end tests (15min timeout)
make test-demo              # Demo system tests
make test-perf              # Performance tests
make test-all               # All tests except Docker (lint + unit + demo + e2e)

# Coverage
make coverage               # Generate HTML coverage report (test-output/coverage.html)

# Quality checks
make lint                   # Run golangci-lint
make fmt                    # Format code (go fmt + gofmt)
make vet                    # Run go vet
make security               # Run gosec security scanner
```

### Docker & Load Testing

```bash
# Docker
docker-compose up --build          # Start full system with Docker
make docker-test-all               # Comprehensive Docker test suite

# Load testing
make load-test-light              # Light load (25 orders/sec)
make load-test-medium             # Medium load (100 orders/sec)
make load-test-heavy              # Heavy load (requires Docker)

# Demo testing
make demo-test-scenarios          # Test demo load scenarios
make demo-test-chaos              # Test chaos engineering scenarios
```

### Running Specific Tests

```bash
# Run a single test
go test -v ./internal/engine -run TestTradingEngine

# Run tests with race detection
go test -race ./internal/...

# Run benchmarks
make benchmark
go test -bench=. -benchmem ./test/e2e/...
```

## Architecture Overview

### Core Design Principles

1. **Dependency Injection**: The entire application uses constructor injection via `internal/app/container.go`. All services are wired together in the container and injected through interfaces.

2. **Interface-Based Design**: All major components are defined by interfaces (see `interfaces.go` files). This enables easy mocking for tests and swappable implementations.

3. **Layered Architecture**:
   - **Domain Layer** (`internal/domain/`, `internal/types/`): Core business entities (Order, Trade, OrderBook)
   - **Application Layer** (`internal/app/`): Orchestration, DI container, lifecycle management
   - **Business Logic** (`internal/engine/`, `internal/simulation/`): Trading engine, market simulator
   - **Infrastructure** (`internal/api/`, `internal/repository/`): HTTP handlers, data persistence
   - **Cross-Cutting** (`internal/metrics/`, `internal/demo/`): Metrics, chaos engineering

4. **Graceful Shutdown**: The application handles SIGINT/SIGTERM with context cancellation propagating through all goroutines. 30-second shutdown timeout is enforced.

### Key Components

**Application Container** (`internal/app/container.go`)
- Central DI container that wires all services together
- Initialization order: Config → Services → Simulation → API Server → Health
- Manages component lifecycle and graceful shutdown

**Trading Engine** (`internal/engine/`)
- Core order matching using price-time priority algorithm
- Thread-safe with RWMutex for concurrent access
- Supports market and limit orders (buy/sell)
- Automatic trade execution when orders match
- `MetricsTradingEngine` wrapper adds instrumentation

**Market Simulator** (`internal/simulation/`)
- Generates realistic market activity with configurable patterns
- Price generation: Brownian motion with drift, volatility clustering, mean reversion
- Order generation: Multiple user profiles (retail, institutional, HFT) with realistic behavior
- Event simulation: News events, flash crashes, volatility spikes
- Runs in background goroutine with configurable tick interval (default 100ms)

**Demo System** (`internal/demo/`)
- Orchestrates load testing and chaos engineering scenarios
- Load test intensities: light (10-50 ops/s), medium (50-200), heavy (200-500), stress (500-1000)
- Chaos types: latency injection, error simulation, resource exhaustion, network partition
- WebSocket integration for real-time updates during demos
- Safety limits prevent runaway chaos tests

**Metrics System** (`internal/metrics/`)
- Real-time metrics collection with time-windowed buffers
- Tracks: order count, trade count, latency (avg/P95/P99), throughput, error rate
- AI-powered performance analyzer detects bottlenecks and trends
- Symbol-specific metrics breakdown
- Periodic calculation (1s intervals)

**API Server** (`internal/api/`)
- Gin web framework with structured middleware stack
- Middleware order: Error Handler → Logging → Security → CORS → Validation → Rate Limiting
- RESTful endpoints for orders, metrics, health checks
- Dashboard endpoint serves web UI

### Data Flow: Order Placement

```
Client → API Handler → Validation Middleware → Order Service → Trading Engine
  ↓
Trading Engine:
  1. Validate order
  2. Generate order ID (UUID)
  3. Save to repository (in-memory)
  4. Fetch matching orders for symbol
  5. Find matches using price-time priority
  6. Execute trades
  7. Update order quantities
  8. Remove filled orders
  ↓
Metrics Collector: Record order event + trade events
  ↓
Response to Client
```

### Critical Implementation Details

**Order Matching Algorithm** (`internal/engine/order_matcher.go`)
- Price-time priority: Best price first, then earliest timestamp
- Bids sorted DESC by price (highest first)
- Asks sorted ASC by price (lowest first)
- Matches execute at the resting order's price (maker price)
- Partial fills supported (order quantity reduced)

**Concurrency Safety**
- `TradingEngine` uses `sync.RWMutex` for thread-safe order book access
- Metrics collector uses `sync.Mutex` for event recording
- Market simulator runs in separate goroutine, respects context cancellation
- All background tasks properly handle shutdown via context

**Configuration** (`internal/config/config.go`)
- Environment variable based configuration
- Key vars: `SERVER_PORT`, `SIMULATION_ENABLED`, `METRICS_ENABLED`, `LOG_LEVEL`
- Simulation config: `SIMULATION_SYMBOLS`, `SIMULATION_ORDERS_PER_SEC`, `SIMULATION_VOLATILITY`
- Defaults ensure system works out-of-box

**Repository Pattern** (`internal/repository/`)
- In-memory implementations for development/testing
- Interface-based design allows swapping to PostgreSQL (`pkg/repository/`)
- Order repo: CRUD operations, query by symbol/status/user
- Trade repo: Store executed trades with full audit trail

## Development Guidelines

### Adding New Features

1. **Define the interface** in appropriate `interfaces.go` file
2. **Implement the interface** in a new file with unit tests
3. **Wire it in the container** (`internal/app/container.go`)
4. **Add API endpoints** if needed (`internal/api/handlers/`)
5. **Write integration tests** (`test/e2e/`)
6. **Update documentation** (README.md, ARCHITECTURE.md as needed)

### Testing Strategy

- **Unit tests**: Every package should have `*_test.go` files alongside implementation
- **Integration tests**: `*_integration_test.go` files test component interactions
- **E2E tests**: `test/e2e/` contains full system tests
- **Coverage target**: 80%+ (enforced in CI)
- **Use testify**: `assert` and `require` for test assertions
- **Mock interfaces**: Use `testify/mock` or manual mocks

### Code Quality Standards

- Run `make fmt` before committing (enforced by pre-commit hooks)
- All public functions/types must have godoc comments
- Complex logic requires inline comments explaining "why" not "what"
- Error handling: Always check errors, use meaningful error messages
- Logging: Use structured logging with contextual fields
- No naked returns in functions longer than 5 lines

### Performance Considerations

- The trading engine is the hot path - optimize for latency
- Use object pooling for frequently allocated objects (see metrics collector)
- Avoid allocations in tight loops
- Profile with `go test -bench` before optimizing
- Load test changes with `make load-test-medium`

### Common Patterns

**Service Creation Pattern:**
```go
type MyService struct {
    dependency1 Dependency1Interface
    dependency2 Dependency2Interface
    mu          sync.RWMutex
}

func NewMyService(dep1 Dependency1Interface, dep2 Dependency2Interface) *MyService {
    return &MyService{
        dependency1: dep1,
        dependency2: dep2,
    }
}
```

**Background Worker Pattern:**
```go
func (s *Service) Start(ctx context.Context) error {
    go func() {
        ticker := time.NewTicker(s.interval)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                s.doWork()
            }
        }
    }()
    return nil
}
```

**Error Wrapping Pattern:**
```go
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

## Important Caveats

### In-Memory Storage
- Default repository is in-memory - data lost on restart
- For persistence, swap to PostgreSQL repos in `pkg/repository/`
- In-memory is intentional for demo/testing scenarios

### Simulation vs Production
- Market simulator generates fake but realistic data
- Not suitable for actual trading without significant modifications
- Price movements are simulated, not connected to real markets

### Scalability Limits
- Current single-instance limit: ~5,000 orders/second
- Horizontal scaling requires distributed order book (planned feature)
- For now, vertical scaling (bigger instance) is recommended

### Testing in CI/CD
- E2E tests require Docker for full environment
- GitHub Actions workflow: `make ci-test` (unit) → `make ci-e2e` → benchmarks
- Full Docker test suite takes ~10 minutes
- Use `make test-unit` for quick feedback during development

## File Organization

```
cmd/api/main.go              # Application entry point
internal/
  app/                       # Application orchestration & DI
    application.go           # Main lifecycle manager
    container.go             # Dependency injection container
    health.go                # Health check service
  engine/                    # Trading engine
    trading_engine.go        # Core matching engine
    order_matcher.go         # Price-time priority algorithm
    trade_executor.go        # Trade execution logic
    metrics_trading_engine.go # Instrumented wrapper
  simulation/                # Market simulation
    market_simulator.go      # Main simulator orchestrator
    price_generator.go       # Realistic price movements
    order_generator.go       # User behavior simulation
    patterns.go              # Market patterns (bull/bear/sideways)
  demo/                      # Demo & chaos engineering
    controller.go            # Demo orchestration
    scenarios.go             # Predefined test scenarios
    websocket.go             # Real-time updates
  metrics/                   # Metrics & analytics
    collector.go             # Real-time metrics collection
    analyzer.go              # AI-powered analysis
    service.go               # Metrics service orchestration
  api/                       # HTTP API layer
    server.go                # Server setup & middleware
    handlers/                # Request handlers
    middleware/              # HTTP middleware
    dto/                     # Data transfer objects
  domain/                    # Domain models
  repository/                # In-memory data storage
  config/                    # Configuration management
test/
  e2e/                       # End-to-end tests
    demo_flow_test.go        # Demo system tests
    performance_test.go      # Load & performance tests
    sla_validation_test.go   # SLA compliance tests
  fixtures/                  # Test data
  helpers/                   # Test utilities
```

## Troubleshooting

### Port Already in Use
If `go run cmd/api/main.go` fails with "address already in use":
```bash
# Find process using port 8080
sudo lsof -i :8080
# Kill it
kill -9 <PID>
# Or use different port
export SERVER_PORT=8081
go run cmd/api/main.go
```

### Tests Failing Randomly
- Race conditions: Run with `go test -race` to detect
- Timing issues: Increase timeouts in test config
- Docker issues: Clean up with `make clean-docker`

### Simulation Not Generating Orders
Check environment variable: `export SIMULATION_ENABLED=true`
Verify in logs: Look for "Market simulation started" message

### High Memory Usage
- Metrics collector retains 60s of events by default
- Adjust `METRICS_COLLECTION_TIME` env var to reduce retention
- Profile with `go test -memprofile=mem.out`

## Additional Resources

- **README.md**: Quick start, features, deployment guide
- **ARCHITECTURE.md**: Detailed architecture documentation with diagrams
- **SRE_RUNBOOK.md**: Complete operational guide for demonstrations
- **API docs**: See handler files for endpoint documentation
- **Design decisions**: Check git history and PR discussions for context
