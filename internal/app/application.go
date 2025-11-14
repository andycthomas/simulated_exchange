package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"simulated_exchange/internal/config"
	"simulated_exchange/internal/simulation"
)

// Application represents the main application orchestrator
type Application struct {
	config    *config.Config
	logger    *slog.Logger
	container *Container

	// Lifecycle management
	ctx       context.Context
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup

	// State tracking
	startTime time.Time
	isRunning bool
	mutex     sync.RWMutex
}

// NewApplication creates a new application instance
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

	// Initialize dependency container
	container, err := NewContainer(cfg, logger)
	if err != nil {
		cancel() // Clean up context
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	app.container = container

	logger.Info("Application initialized successfully",
		"environment", cfg.Server.Environment,
		"port", cfg.Server.Port,
		"simulation_enabled", cfg.Simulation.Enabled,
		"metrics_enabled", cfg.Metrics.Enabled,
		"health_enabled", cfg.Health.Enabled,
	)

	return app, nil
}

// Start starts the application and all its components
func (a *Application) Start() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.isRunning {
		return fmt.Errorf("application is already running")
	}

	a.logger.Info("Starting application")
	a.startTime = time.Now()

	// Start components in dependency order
	if err := a.startComponents(); err != nil {
		return fmt.Errorf("failed to start components: %w", err)
	}

	a.isRunning = true
	a.logger.Info("Application started successfully",
		"startup_duration", time.Since(a.startTime),
	)

	return nil
}

// startComponents starts all application components in the correct order
func (a *Application) startComponents() error {
	// Start health checks first
	if a.config.Health.Enabled {
		a.logger.Info("Health checks enabled")
		// Health service is already started in container initialization
	}

	// Start market simulation if enabled
	if a.config.Simulation.Enabled {
		if err := a.startSimulation(); err != nil {
			return fmt.Errorf("failed to start simulation: %w", err)
		}
	}

	// Start HTTP server
	if err := a.startServer(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// startSimulation starts the market simulation
func (a *Application) startSimulation() error {
	a.logger.Info("Starting market simulation")

	simulator := a.container.GetMarketSimulator()
	if simulator == nil {
		return fmt.Errorf("market simulator not available")
	}

	// Create simulation configuration
	simConfig := simulation.SimulationConfig{
		Symbols:              a.config.Simulation.Symbols,
		SimulationDuration:   24 * time.Hour, // Run continuously
		TickInterval:         100 * time.Millisecond,
		OrderGenerationRate:  a.config.Simulation.OrdersPerSecond,
		PriceUpdateFrequency: 1 * time.Second,
		BaseVolatility:       a.config.Simulation.VolatilityFactor,
		InitialPrices:        map[string]float64{
			"BTCUSD":  50000.0,
			"ETHUSD":  3000.0,
			"ADAUSD":  1.5,
			"DOTUSD":  25.0,
			"SOLUSD":  100.0,
			"LINKUSD": 15.0,
			"AVAXUSD": 35.0,
			"MATICUSD": 1.2,
		},
		MarketCondition:      simulation.MarketCondition("STEADY"),
		EnablePatterns:       a.config.Simulation.EnableVolatility,
		NewsEventFrequency:   a.config.Simulation.PatternInterval,
	}

	// Start simulation in background
	a.waitGroup.Add(1)
	go func() {
		defer a.waitGroup.Done()
		if err := simulator.StartSimulation(a.ctx, simConfig); err != nil {
			a.logger.Error("Simulation failed", "error", err)
		}
	}()

	a.logger.Info("Market simulation started",
		"symbols", a.config.Simulation.Symbols,
		"orders_per_second", a.config.Simulation.OrdersPerSecond,
		"volatility_factor", a.config.Simulation.VolatilityFactor,
	)

	return nil
}

// startServer starts the HTTP server
func (a *Application) startServer() error {
	a.logger.Info("Starting HTTP server")

	server := a.container.GetServer()
	if server == nil {
		return fmt.Errorf("server not available")
	}

	// Start server in background
	a.waitGroup.Add(1)
	go func() {
		defer a.waitGroup.Done()

		a.logger.Info("HTTP server listening",
			"host", a.config.Server.Host,
			"port", a.config.Server.Port,
			"environment", a.config.Server.Environment,
		)

		if err := server.StartWithContext(a.ctx); err != nil {
			// Only log error if it's not due to context cancellation
			select {
			case <-a.ctx.Done():
				a.logger.Info("HTTP server stopped due to context cancellation")
			default:
				a.logger.Error("HTTP server failed", "error", err)
			}
		}
	}()

	return nil
}

// Stop gracefully stops the application and all its components
func (a *Application) Stop() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.isRunning {
		return fmt.Errorf("application is not running")
	}

	a.logger.Info("Stopping application")
	stopStart := time.Now()

	// Cancel context to signal all goroutines to stop
	a.cancel()

	// Create timeout context for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop container services
	if err := a.container.Shutdown(shutdownCtx); err != nil {
		a.logger.Warn("Container shutdown completed with errors", "error", err)
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		a.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("All goroutines stopped gracefully")
	case <-shutdownCtx.Done():
		a.logger.Warn("Shutdown timeout reached, some goroutines may still be running")
	}

	a.isRunning = false
	uptime := time.Since(a.startTime)
	shutdownDuration := time.Since(stopStart)

	a.logger.Info("Application stopped",
		"uptime", uptime,
		"shutdown_duration", shutdownDuration,
	)

	return nil
}

// IsRunning returns true if the application is currently running
func (a *Application) IsRunning() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.isRunning
}

// GetConfig returns the application configuration
func (a *Application) GetConfig() *config.Config {
	return a.config
}

// GetContainer returns the dependency injection container
func (a *Application) GetContainer() *Container {
	return a.container
}

// GetUptime returns the application uptime
func (a *Application) GetUptime() time.Duration {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if !a.isRunning {
		return 0
	}
	return time.Since(a.startTime)
}

// GetHealthStatus returns the current health status
func (a *Application) GetHealthStatus() HealthStatus {
	if a.container.GetHealthService() == nil {
		return HealthStatus{
			Status:    "unhealthy",
			Timestamp: time.Now(),
			Checks:    map[string]string{"health_service": "not available"},
		}
	}
	return a.container.GetHealthService().GetStatus()
}

// InjectVolatility injects volatility into the market simulation
func (a *Application) InjectVolatility(pattern simulation.VolatilityPattern, duration time.Duration) error {
	if !a.config.Simulation.Enabled {
		return fmt.Errorf("simulation is not enabled")
	}

	simulator := a.container.GetMarketSimulator()
	if simulator == nil {
		return fmt.Errorf("market simulator not available")
	}

	return simulator.InjectVolatility(pattern, duration)
}

// GetSimulationStatus returns the current simulation status
func (a *Application) GetSimulationStatus() (simulation.SimulationStatus, error) {
	if !a.config.Simulation.Enabled {
		return simulation.SimulationStatus{}, fmt.Errorf("simulation is not enabled")
	}

	simulator := a.container.GetMarketSimulator()
	if simulator == nil {
		return simulation.SimulationStatus{}, fmt.Errorf("market simulator not available")
	}

	return simulator.GetSimulationStatus(), nil
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