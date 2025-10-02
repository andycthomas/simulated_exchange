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
	"simulated_exchange/services/order-flow-simulator/internal/domain"
	"simulated_exchange/services/order-flow-simulator/internal/handlers"
	"simulated_exchange/services/order-flow-simulator/internal/server"
)

// Application represents the Order Flow Simulator microservice
type Application struct {
	config *config.Config
	logger *slog.Logger

	// Infrastructure
	cache    *cache.RedisClient
	eventBus *messaging.RedisEventBus

	// Services
	orderGenerator     *domain.OrderGenerator
	userSimulator      *domain.UserSimulator
	flowSimulator      *domain.FlowSimulator
	tradingAPIClient   *domain.TradingAPIClient

	// HTTP Server (for health checks and control)
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

// NewApplication creates a new Order Flow Simulator application instance
func NewApplication() (*Application, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Override service name for order flow simulator
	cfg.Service.Name = "order-flow-simulator"

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

	logger.Info("Order Flow Simulator application initialized successfully",
		"service", cfg.Service.Name,
		"version", cfg.Service.Version,
		"environment", cfg.Service.Environment,
	)

	return app, nil
}

// Start starts the Order Flow Simulator application
func (a *Application) Start(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isRunning {
		return fmt.Errorf("order flow simulator is already running")
	}

	a.logger.Info("Starting Order Flow Simulator application")
	a.startTime = time.Now()

	// Subscribe to events
	if err := a.subscribeToEvents(); err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// Start HTTP server for health checks and control
	if err := a.startServer(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Start order flow simulation
	if err := a.startSimulation(); err != nil {
		return fmt.Errorf("failed to start simulation: %w", err)
	}

	a.isRunning = true
	a.logger.Info("Order Flow Simulator application started successfully",
		"startup_duration", time.Since(a.startTime),
	)

	return nil
}

// Stop gracefully stops the Order Flow Simulator application
func (a *Application) Stop() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isRunning {
		return fmt.Errorf("order flow simulator is not running")
	}

	a.logger.Info("Stopping Order Flow Simulator application")
	stopStart := time.Now()

	// Cancel context to signal all goroutines to stop
	a.cancel()

	// Stop simulation
	if a.flowSimulator != nil {
		if err := a.flowSimulator.Stop(); err != nil {
			a.logger.Warn("Error stopping flow simulation", "error", err)
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

	a.logger.Info("Order Flow Simulator application stopped",
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

	// Initialize trading API client
	tradingAPIURL := "http://trading-api:8080" // Docker service name
	a.tradingAPIClient = domain.NewTradingAPIClient(tradingAPIURL, a.logger)

	// Initialize order generator
	orderConfig := domain.OrderGeneratorConfig{
		BaseOrderRate:   10.0,
		VolatilityBoost: 2.0,
		UserTypeMix: map[string]float64{
			"conservative": 0.4,
			"aggressive":   0.3,
			"momentum":     0.3,
		},
		RandomSeed: time.Now().UnixNano(),
	}
	a.orderGenerator = domain.NewOrderGenerator(orderConfig, a.logger)

	// Initialize user simulator
	a.userSimulator = domain.NewUserSimulator(a.orderGenerator, a.logger)

	// Initialize flow simulator
	a.flowSimulator = domain.NewFlowSimulator(
		a.orderGenerator,
		a.userSimulator,
		a.tradingAPIClient,
		a.eventBus,
		a.logger,
	)

	a.logger.Info("Services initialized successfully")
	return nil
}

// initializeServer sets up the HTTP server for health checks and control
func (a *Application) initializeServer() error {
	a.logger.Info("Initializing HTTP server")

	// Create handlers
	healthHandler := handlers.NewHealthHandler(a.cache, a.tradingAPIClient, a.logger)
	flowHandler := handlers.NewFlowHandler(a.flowSimulator, a.userSimulator, a.logger)

	// Create server
	a.server = server.NewServer(
		a.config,
		healthHandler,
		flowHandler,
		a.logger,
	)

	a.logger.Info("HTTP server initialized successfully")
	return nil
}

// subscribeToEvents subscribes to relevant events from other services
func (a *Application) subscribeToEvents() error {
	a.logger.Info("Subscribing to events")

	// Subscribe to price updates to react to market changes
	err := a.eventBus.Subscribe(a.ctx, shared.EventTypePriceUpdate, a.handlePriceUpdate)
	if err != nil {
		return fmt.Errorf("failed to subscribe to price updates: %w", err)
	}

	// Subscribe to trade executions to adjust user behavior
	err = a.eventBus.Subscribe(a.ctx, shared.EventTypeTradeExecuted, a.handleTradeExecuted)
	if err != nil {
		return fmt.Errorf("failed to subscribe to trade executions: %w", err)
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

// startSimulation starts the order flow simulation
func (a *Application) startSimulation() error {
	a.logger.Info("Starting order flow simulation")

	// Start flow simulation service
	if err := a.flowSimulator.Start(a.ctx); err != nil {
		return fmt.Errorf("failed to start flow simulation service: %w", err)
	}

	a.logger.Info("Order flow simulation started successfully")
	return nil
}

// Event handlers

func (a *Application) handlePriceUpdate(ctx context.Context, event *shared.Event) error {
	// React to price changes by adjusting order generation behavior
	symbol, ok := event.Data["symbol"].(string)
	if !ok {
		return fmt.Errorf("invalid symbol in price update event")
	}

	price, ok := event.Data["price"].(float64)
	if !ok {
		return fmt.Errorf("invalid price in price update event")
	}

	// Notify user simulator about price change
	a.userSimulator.OnPriceUpdate(symbol, price)

	a.logger.Debug("Processed price update", "symbol", symbol, "price", price)
	return nil
}

func (a *Application) handleTradeExecuted(ctx context.Context, event *shared.Event) error {
	// React to trade executions by adjusting user behavior
	symbol, ok := event.Data["symbol"].(string)
	if !ok {
		return fmt.Errorf("invalid symbol in trade executed event")
	}

	price, ok := event.Data["price"].(float64)
	if !ok {
		return fmt.Errorf("invalid price in trade executed event")
	}

	quantity, ok := event.Data["quantity"].(float64)
	if !ok {
		return fmt.Errorf("invalid quantity in trade executed event")
	}

	// Notify user simulator about trade execution
	a.userSimulator.OnTradeExecuted(symbol, price, quantity)

	a.logger.Debug("Processed trade execution", "symbol", symbol, "price", price, "quantity", quantity)
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