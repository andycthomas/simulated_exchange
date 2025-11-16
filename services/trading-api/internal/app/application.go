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
	"simulated_exchange/pkg/database"
	"simulated_exchange/pkg/messaging"
	"simulated_exchange/pkg/monitoring"
	"simulated_exchange/pkg/repository"
	"simulated_exchange/pkg/shared"
	"simulated_exchange/services/trading-api/internal/domain"
	"simulated_exchange/services/trading-api/internal/handlers"
	"simulated_exchange/services/trading-api/internal/server"
)

// Application represents the Trading API microservice
type Application struct {
	config *config.Config
	logger *slog.Logger

	// Infrastructure
	db       *database.PostgresDB
	cache    *cache.RedisClient
	eventBus *messaging.RedisEventBus

	// Repositories
	orderRepo shared.OrderRepository
	tradeRepo shared.TradeRepository
	userRepo  shared.UserRepository

	// Services
	tradingService shared.TradingService
	orderMatcher   shared.OrderMatcher

	// HTTP Server
	server *server.Server

	// Metrics
	metricsCollector *monitoring.MetricsCollector
	metricsUpdater   *monitoring.PeriodicMetricsUpdater

	// Lifecycle management
	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup

	// State tracking
	startTime time.Time
	isRunning bool
	mutex     sync.RWMutex
}

// NewApplication creates a new Trading API application instance
func NewApplication() (*Application, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

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

	// Initialize repositories
	if err := app.initializeRepositories(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Initialize services
	if err := app.initializeServices(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	// Initialize metrics
	if err := app.initializeMetrics(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize server
	if err := app.initializeServer(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize server: %w", err)
	}

	logger.Info("Trading API application initialized successfully",
		"service", cfg.Service.Name,
		"version", cfg.Service.Version,
		"environment", cfg.Service.Environment,
		"port", cfg.Server.Port,
	)

	return app, nil
}

// Start starts the Trading API application
func (a *Application) Start(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isRunning {
		return fmt.Errorf("trading API is already running")
	}

	a.logger.Info("Starting Trading API application")
	a.startTime = time.Now()

	// Start periodic metrics updates
	a.metricsUpdater.Start("trading-api")

	// Set initial service health
	a.metricsCollector.SetServiceHealth("trading-api", true)

	// Subscribe to events
	if err := a.subscribeToEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// Start HTTP server
	if err := a.startServer(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	a.isRunning = true
	a.logger.Info("Trading API application started successfully",
		"startup_duration", time.Since(a.startTime),
	)

	return nil
}

// Stop gracefully stops the Trading API application
func (a *Application) Stop() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isRunning {
		return fmt.Errorf("trading API is not running")
	}

	a.logger.Info("Stopping Trading API application")
	stopStart := time.Now()

	// Stop periodic metrics updates
	if a.metricsUpdater != nil {
		a.metricsUpdater.Stop()
	}

	// Cancel context to signal all goroutines to stop
	a.cancel()

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

	// Close database connection
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			a.logger.Warn("Error closing database connection", "error", err)
		}
	}

	// Wait for all goroutines to finish
	a.waitGroup.Wait()

	a.isRunning = false
	uptime := time.Since(a.startTime)
	shutdownDuration := time.Since(stopStart)

	a.logger.Info("Trading API application stopped",
		"uptime", uptime,
		"shutdown_duration", shutdownDuration,
	)

	return nil
}

// initializeInfrastructure sets up database, cache, and messaging
func (a *Application) initializeInfrastructure() error {
	a.logger.Info("Initializing infrastructure components")

	// Initialize PostgreSQL database
	db, err := database.NewPostgresDB(a.config.GetDatabaseConnectionString())
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	a.db = db

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

// initializeRepositories sets up all repositories
func (a *Application) initializeRepositories() error {
	a.logger.Info("Initializing repositories")

	a.orderRepo = repository.NewPostgresOrderRepository(a.db.GetDB())
	a.tradeRepo = repository.NewPostgresTradeRepository(a.db.GetDB())
	a.userRepo = repository.NewPostgresUserRepository(a.db.GetDB())

	a.logger.Info("Repositories initialized successfully")
	return nil
}

// initializeServices sets up business logic services
func (a *Application) initializeServices() error {
	a.logger.Info("Initializing services")

	// Initialize order matcher
	a.orderMatcher = domain.NewOrderMatcher(a.logger)

	// Initialize trading service
	a.tradingService = domain.NewTradingService(
		a.orderRepo,
		a.tradeRepo,
		a.cache,
		a.eventBus,
		a.orderMatcher,
		a.logger,
	)

	a.logger.Info("Services initialized successfully")
	return nil
}

// initializeMetrics sets up metrics collection
func (a *Application) initializeMetrics() error {
	a.logger.Info("Initializing metrics")

	// Create metrics collector
	a.metricsCollector = monitoring.NewMetricsCollector(a.logger)

	// Create periodic metrics updater
	a.metricsUpdater = monitoring.NewPeriodicMetricsUpdater(a.metricsCollector, a.logger)

	a.logger.Info("Metrics initialized successfully")
	return nil
}

// initializeServer sets up the HTTP server
func (a *Application) initializeServer() error {
	a.logger.Info("Initializing HTTP server")

	// Create handlers
	orderHandler := handlers.NewOrderHandler(a.tradingService, a.metricsCollector, a.logger)
	healthHandler := handlers.NewHealthHandler(a.db, a.cache, a.logger)
	metricsHandler := handlers.NewMetricsHandler(a.tradingService, a.logger, time.Now())

	// Create server (pass our metrics collector so it's exposed via /metrics)
	a.server = server.NewServer(
		a.config,
		orderHandler,
		healthHandler,
		metricsHandler,
		a.metricsCollector,
		a.logger,
	)

	a.logger.Info("HTTP server initialized successfully")
	return nil
}

// subscribeToEvents subscribes to relevant events from other services
func (a *Application) subscribeToEvents() error {
	a.logger.Info("Subscribing to events")

	// Subscribe to price updates from market simulator
	err := a.eventBus.Subscribe(a.ctx, shared.EventTypePriceUpdate, a.handlePriceUpdate)
	if err != nil {
		return fmt.Errorf("failed to subscribe to price updates: %w", err)
	}

	// Subscribe to market data from market simulator
	err = a.eventBus.Subscribe(a.ctx, shared.EventTypeMarketData, a.handleMarketData)
	if err != nil {
		return fmt.Errorf("failed to subscribe to market data: %w", err)
	}

	a.logger.Info("Successfully subscribed to events")
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

// Event handlers

func (a *Application) handlePriceUpdate(ctx context.Context, event *shared.Event) error {
	// Update price in cache for fast access
	symbol, ok := event.Data["symbol"].(string)
	if !ok {
		return fmt.Errorf("invalid symbol in price update event")
	}

	price, ok := event.Data["price"].(float64)
	if !ok {
		return fmt.Errorf("invalid price in price update event")
	}

	err := a.cache.SetPrice(ctx, symbol, price)
	if err != nil {
		a.logger.Warn("Failed to cache price update", "error", err, "symbol", symbol)
	}

	a.logger.Debug("Processed price update", "symbol", symbol, "price", price)
	return nil
}

func (a *Application) handleMarketData(ctx context.Context, event *shared.Event) error {
	// Cache market data for quick access
	symbol, ok := event.Data["symbol"].(string)
	if !ok {
		return fmt.Errorf("invalid symbol in market data event")
	}

	marketData := &shared.MarketData{
		Symbol:           symbol,
		CurrentPrice:     getFloat64FromEvent(event.Data, "current_price"),
		PreviousPrice:    getFloat64FromEvent(event.Data, "previous_price"),
		DailyHigh:        getFloat64FromEvent(event.Data, "daily_high"),
		DailyLow:         getFloat64FromEvent(event.Data, "daily_low"),
		DailyVolume:      getFloat64FromEvent(event.Data, "daily_volume"),
		PriceChange:      getFloat64FromEvent(event.Data, "price_change"),
		PriceChangePerc:  getFloat64FromEvent(event.Data, "price_change_percent"),
		Timestamp:        time.Now(),
	}

	err := a.cache.SetMarketData(ctx, symbol, marketData)
	if err != nil {
		a.logger.Warn("Failed to cache market data", "error", err, "symbol", symbol)
	}

	a.logger.Debug("Processed market data", "symbol", symbol)
	return nil
}

// Helper functions

func getFloat64FromEvent(data map[string]interface{}, key string) float64 {
	if value, ok := data[key].(float64); ok {
		return value
	}
	return 0
}

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