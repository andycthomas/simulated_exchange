package domain

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"simulated_exchange/pkg/messaging"
	"simulated_exchange/pkg/shared"
)

// SimulatorService implements shared.SimulatorService interface
type SimulatorService struct {
	priceService      *PriceService
	marketDataService *MarketDataService
	eventBus          *messaging.RedisEventBus
	logger            *slog.Logger

	// Simulation state
	isRunning   bool
	ctx         context.Context
	cancel      context.CancelFunc
	waitGroup   sync.WaitGroup
	mutex       sync.RWMutex
	startTime   time.Time

	// Configuration
	symbols           []string
	updateInterval    time.Duration
	volatilityEvents  map[string]time.Time
}

// NewSimulatorService creates a new simulator service
func NewSimulatorService(
	priceService *PriceService,
	marketDataService *MarketDataService,
	eventBus *messaging.RedisEventBus,
	logger *slog.Logger,
) *SimulatorService {
	return &SimulatorService{
		priceService:      priceService,
		marketDataService: marketDataService,
		eventBus:          eventBus,
		logger:            logger,
		symbols:           []string{"BTCUSD", "ETHUSD", "ADAUSD"},
		updateInterval:    1 * time.Second,
		volatilityEvents:  make(map[string]time.Time),
	}
}

// Start starts the market simulation
func (ss *SimulatorService) Start(ctx context.Context) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	if ss.isRunning {
		return fmt.Errorf("simulator is already running")
	}

	ss.ctx, ss.cancel = context.WithCancel(ctx)
	ss.isRunning = true
	ss.startTime = time.Now()

	// Start price generation workers
	for _, symbol := range ss.symbols {
		ss.startPriceWorker(symbol)
	}

	// Start volatility event worker
	ss.startVolatilityWorker()

	// Start status reporter
	ss.startStatusReporter()

	ss.logger.Info("Market simulation started",
		"symbols", ss.symbols,
		"update_interval", ss.updateInterval,
	)

	return nil
}

// Stop stops the market simulation
func (ss *SimulatorService) Stop(ctx context.Context) error {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	if !ss.isRunning {
		return fmt.Errorf("simulator is not running")
	}

	ss.logger.Info("Stopping market simulation")

	// Cancel context to stop all workers
	ss.cancel()

	// Wait for all workers to finish
	ss.waitGroup.Wait()

	ss.isRunning = false

	ss.logger.Info("Market simulation stopped",
		"runtime", time.Since(ss.startTime),
	)

	return nil
}

// IsRunning returns true if the simulator is running
func (ss *SimulatorService) IsRunning() bool {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()
	return ss.isRunning
}

// GetStatus returns the current simulation status
func (ss *SimulatorService) GetStatus(ctx context.Context) (*shared.ServiceInfo, error) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	status := &shared.ServiceInfo{
		Name:      "market-simulator",
		Version:   "1.0.0",
		Status:    "stopped",
		Timestamp: time.Now(),
		Uptime:    0,
		Metadata: map[string]string{
			"symbols":         fmt.Sprintf("%v", ss.symbols),
			"update_interval": ss.updateInterval.String(),
		},
	}

	if ss.isRunning {
		status.Status = "running"
		status.Uptime = time.Since(ss.startTime)
	}

	return status, nil
}

// InjectVolatility injects volatility into the market
func (ss *SimulatorService) InjectVolatility(ctx context.Context, pattern string, intensity float64) error {
	ss.logger.Info("Injecting volatility",
		"pattern", pattern,
		"intensity", intensity,
	)

	// Apply volatility to all symbols
	for _, symbol := range ss.symbols {
		if err := ss.priceService.SimulateVolatility(symbol, pattern, intensity); err != nil {
			ss.logger.Warn("Failed to apply volatility",
				"symbol", symbol,
				"error", err,
			)
		}
	}

	// Publish volatility event
	event := &shared.Event{
		Type:   "volatility.injected",
		Source: "market-simulator",
		Data: map[string]interface{}{
			"pattern":   pattern,
			"intensity": intensity,
			"symbols":   ss.symbols,
		},
	}

	return ss.eventBus.Publish(ctx, event)
}

// Worker functions

func (ss *SimulatorService) startPriceWorker(symbol string) {
	ss.waitGroup.Add(1)

	go func() {
		defer ss.waitGroup.Done()

		ticker := time.NewTicker(ss.updateInterval)
		defer ticker.Stop()

		ss.logger.Debug("Started price worker", "symbol", symbol)

		for {
			select {
			case <-ss.ctx.Done():
				ss.logger.Debug("Price worker stopped", "symbol", symbol)
				return
			case <-ticker.C:
				if err := ss.priceService.GenerateAndPublishPrice(ss.ctx, symbol); err != nil {
					ss.logger.Warn("Failed to generate price",
						"symbol", symbol,
						"error", err,
					)
				}
			}
		}
	}()
}

func (ss *SimulatorService) startVolatilityWorker() {
	ss.waitGroup.Add(1)

	go func() {
		defer ss.waitGroup.Done()

		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
		defer ticker.Stop()

		ss.logger.Debug("Started volatility worker")

		for {
			select {
			case <-ss.ctx.Done():
				ss.logger.Debug("Volatility worker stopped")
				return
			case <-ticker.C:
				ss.maybeInjectRandomVolatility()
			}
		}
	}()
}

func (ss *SimulatorService) startStatusReporter() {
	ss.waitGroup.Add(1)

	go func() {
		defer ss.waitGroup.Done()

		ticker := time.NewTicker(60 * time.Second) // Report every minute
		defer ticker.Stop()

		ss.logger.Debug("Started status reporter")

		for {
			select {
			case <-ss.ctx.Done():
				ss.logger.Debug("Status reporter stopped")
				return
			case <-ticker.C:
				ss.publishSystemStatus()
			}
		}
	}()
}

func (ss *SimulatorService) maybeInjectRandomVolatility() {
	// 10% chance of random volatility event
	if time.Now().Unix()%10 != 0 {
		return
	}

	patterns := []string{"spike", "decay", "oscillate", "random"}
	pattern := patterns[time.Now().Unix()%int64(len(patterns))]
	intensity := 0.1 + (float64(time.Now().Unix()%5) * 0.1) // 0.1 to 0.5

	if err := ss.InjectVolatility(ss.ctx, pattern, intensity); err != nil {
		ss.logger.Warn("Failed to inject random volatility", "error", err)
	} else {
		ss.logger.Info("Random volatility injected",
			"pattern", pattern,
			"intensity", intensity,
		)
	}
}

func (ss *SimulatorService) publishSystemStatus() {
	status := map[string]interface{}{
		"service":     "market-simulator",
		"status":      "running",
		"uptime":      time.Since(ss.startTime).String(),
		"symbols":     ss.symbols,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if err := ss.marketDataService.PublishSystemStatus(ss.ctx, status); err != nil {
		ss.logger.Warn("Failed to publish system status", "error", err)
	}
}