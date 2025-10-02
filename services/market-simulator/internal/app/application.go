package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"simulated_exchange/pkg/cache"
	"simulated_exchange/pkg/config"
	"simulated_exchange/pkg/messaging"
	"simulated_exchange/pkg/shared"
	"simulated_exchange/services/market-simulator/internal/domain"
	"simulated_exchange/services/market-simulator/internal/handlers"
	"simulated_exchange/services/market-simulator/internal/server"
)

// Application represents the Market Simulator microservice
type Application struct {
	config *config.Config
	logger *slog.Logger

	// Infrastructure
	cache    *cache.RedisClient
	eventBus *messaging.RedisEventBus

	// Services
	priceService       shared.PriceService
	simulatorService   shared.SimulatorService
	priceGenerator     *domain.PriceGenerator
	marketDataService  *domain.MarketDataService

	// HTTP Server (for health checks and metrics)
	server *server.Server

	// Lifecycle management
	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup

	// State tracking
	startTime time.Time
	isRunning bool
	mutex     sync.RWMutex
}

// NewApplication creates a new Market Simulator application instance
func NewApplication() (*Application, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Override service name for market simulator
	cfg.Service.Name = "market-simulator"

	// Initialize structured logging
	logger, err := initializeLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())

	app := &Application{
		config: cfg,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize infrastructure components
	if err := app.initializeInfrastructure(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize services
	if err := app.initializeServices(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize server
	if err := app.initializeServer(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	logger.Info("Market Simulator application initialized successfully",
		"service", cfg.Service.Name,
		"version", cfg.Service.Version,
		"environment", cfg.Service.Environment,
	)

	return app, nil
}

// Start starts the Market Simulator application
func (a *Application) Start(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isRunning {
		return fmt.Errorf("market simulator is already running")
	}

	a.logger.Info("Starting Market Simulator application")
	a.startTime = time.Now()

	// Start HTTP server for health checks
	if err := a.startServer(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Start market simulation
	if err := a.startSimulation(); err != nil {
		return fmt.Errorf("failed to start simulation: %w", err)
	}

	a.isRunning = true
	a.logger.Info("Market Simulator application started successfully",
		"startup_duration", time.Since(a.startTime),
	)

	return nil
}

// Stop gracefully stops the Market Simulator application
func (a *Application) Stop() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isRunning {
		return fmt.Errorf("market simulator is not running")
	}

	a.logger.Info("Stopping Market Simulator application")
	stopStart := time.Now()

	// Cancel context to signal all goroutines to stop
	a.cancel()

	// Stop simulation
	if a.simulatorService != nil {
		if err := a.simulatorService.Stop(context.Background()); err != nil {
			a.logger.Warn("Error stopping simulation", "error", err)
		}
	}

	// Close event bus
	if a.eventBus != nil {
		if err := a.eventBus.Close(); err != nil {
			a.logger.Warn("Error closing event bus", "error", err)
		}
	}

	// Close cache connection
	if a.cache != nil {
		if err := a.cache.Close(); err != nil {
			a.logger.Warn("Error closing cache connection", "error", err)
		}
	}

	// Wait for all goroutines to finish
	a.waitGroup.Wait()

	a.isRunning = false
	uptime := time.Since(a.startTime)
	shutdownDuration := time.Since(stopStart)

	a.logger.Info("Market Simulator application stopped",
		"uptime", uptime,
		"shutdown_duration", shutdownDuration,
	)

	return nil
}

// initializeInfrastructure sets up cache and messaging
func (a *Application) initializeInfrastructure() error {
	a.logger.Info("Initializing infrastructure components")

	// Initialize Redis cache
	redisClient := redis.NewClient(&redis.Options{
		Addr:     a.config.GetRedisAddress(),
		Password: a.config.Redis.Password,
		DB:       a.config.Redis.Database,
	})

	// Test Redis connection
	if err := redisClient.Ping(a.ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	a.cache = cache.NewRedisClient(a.config.GetRedisAddress(), a.config.Redis.Password, a.config.Redis.Database)

	// Initialize event bus
	a.eventBus = messaging.NewRedisEventBus(redisClient)

	a.logger.Info("Infrastructure components initialized successfully")
	return nil
}

// initializeServices sets up business logic services
func (a *Application) initializeServices() error {
	a.logger.Info("Initializing services")

	// Initialize price generator
	priceConfig := domain.PriceGeneratorConfig{
		BaseVolatility:    0.02,
		VolatilityDecay:   0.95,
		SpreadPercentage:  0.001,
		PriceStepSize:     0.01,
		TrendPersistence:  0.7,
		MeanReversion:     0.3,
		HistorySize:       1000,
		RandomSeed:        time.Now().UnixNano(),
	}
	a.priceGenerator = domain.NewPriceGenerator(priceConfig, a.logger)

	// Initialize market data service
	a.marketDataService = domain.NewMarketDataService(a.cache, a.eventBus, a.logger)

	// Initialize price service
	a.priceService = domain.NewPriceService(a.priceGenerator, a.marketDataService, a.logger)

	// Initialize simulator service
	a.simulatorService = domain.NewSimulatorService(
		a.priceService.(*domain.PriceService),
		a.marketDataService,
		a.eventBus,
		a.logger,
	)

	a.logger.Info("Services initialized successfully")
	return nil
}

// initializeServer sets up the HTTP server for health checks
func (a *Application) initializeServer() error {
	a.logger.Info("Initializing HTTP server")

	// Create handlers
	healthHandler := handlers.NewHealthHandler(a.cache, a.logger)
	simulatorHandler := handlers.NewSimulatorHandler(a.simulatorService, a.priceService.(*domain.PriceService), a.logger)

	// Create server
	a.server = server.NewServer(
		a.config,
		healthHandler,
		simulatorHandler,
		a.logger,
	)

	a.logger.Info("HTTP server initialized successfully")
	return nil
}

// startServer starts the HTTP server
func (a *Application) startServer() error {
	a.logger.Info("Starting HTTP server")

	a.waitGroup.Add(1)
	go func() {
		defer a.waitGroup.Done()

		if err := a.server.Start(); err != nil {
			a.logger.Error("HTTP server failed", "error", err)
		}
	}()

	return nil
}

// startSimulation starts the market simulation
func (a *Application) startSimulation() error {
	a.logger.Info("Starting market simulation")

	// Set initial prices
	initialPrices := map[string]float64{
		"BTCUSD": 50000.0,
		"ETHUSD": 3000.0,
		"ADAUSD": 1.5,
	}

	for symbol, price := range initialPrices {
		if err := a.priceGenerator.SetBasePrice(symbol, price); err != nil {
			a.logger.Warn("Failed to set initial price", "symbol", symbol, "price", price, "error", err)
		}
	}

	// Start simulation service
	if err := a.simulatorService.Start(a.ctx); err != nil {
		return fmt.Errorf("failed to start simulation service: %w", err)
	}

	a.logger.Info("Market simulation started successfully")
	return nil
}

// initializeLogger sets up structured logging based on configuration
func initializeLogger(cfg config.LoggingConfig) (*slog.Logger, error) {
	var handler slog.Handler

	// Configure log level
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	// Configure output destination
	var output *os.File = os.Stdout
	if cfg.OutputFile != "" {
		file, err := os.OpenFile(cfg.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	}

	// Configure handler based on format
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	case "text":
		handler = slog.NewTextHandler(output, opts)
	default:
		handler = slog.NewJSONHandler(output, opts)
	}

	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	return logger, nil
}