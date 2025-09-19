package simulation

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"simulated_exchange/internal/api/dto"
)

// RealisticSimulator implements MarketSimulator interface
type RealisticSimulator struct {
	// Core components
	priceGenerator PriceGenerator
	orderGenerator OrderGenerator
	eventGenerator EventGenerator
	tradingEngine  TradingEngine

	// State management
	config           SimulationConfig
	status           SimulationStatus
	isRunning        bool
	stopChan         chan struct{}
	doneChan         chan struct{}
	mu               sync.RWMutex
	orderSemaphore   chan struct{} // Limits concurrent order operations

	// Pattern management
	activePatterns   map[string]*SimulationPattern
	patternHistory   []PatternEvent

	// Statistics tracking
	stats            SimulationStatistics
	lastStatsUpdate  time.Time

	// Background workers
	priceWorker      *Worker
	orderWorker      *Worker
	eventWorker      *Worker
	patternWorker    *Worker
}

// Worker represents a background worker goroutine
type Worker struct {
	name     string
	interval time.Duration
	stopChan chan struct{}
	doneChan chan struct{}
	workFunc func() error
}

// PatternEvent records pattern activation/deactivation
type PatternEvent struct {
	PatternName string    `json:"pattern_name"`
	EventType   string    `json:"event_type"` // "activated" or "deactivated"
	Timestamp   time.Time `json:"timestamp"`
	Duration    time.Duration `json:"duration,omitempty"`
}

// SimulationStatistics tracks simulation performance metrics
type SimulationStatistics struct {
	StartTime           time.Time          `json:"start_time"`
	TotalRuntime        time.Duration      `json:"total_runtime"`
	OrdersGenerated     int64              `json:"orders_generated"`
	PriceUpdates        int64              `json:"price_updates"`
	EventsTriggered     int64              `json:"events_triggered"`
	PatternsActivated   int64              `json:"patterns_activated"`
	ErrorCount          int64              `json:"error_count"`
	LastError           error              `json:"last_error,omitempty"`
	PerformanceMetrics  PerformanceMetrics `json:"performance_metrics"`
}

// PerformanceMetrics tracks system performance
type PerformanceMetrics struct {
	AvgOrderGenerationTime time.Duration `json:"avg_order_generation_time"`
	AvgPriceUpdateTime     time.Duration `json:"avg_price_update_time"`
	MemoryUsage           int64         `json:"memory_usage"`
	GoroutineCount        int           `json:"goroutine_count"`
}

// NewRealisticSimulator creates a new market simulator instance
func NewRealisticSimulator(
	priceGen PriceGenerator,
	orderGen OrderGenerator,
	eventGen EventGenerator,
	tradingEngine TradingEngine,
) *RealisticSimulator {
	return &RealisticSimulator{
		priceGenerator: priceGen,
		orderGenerator: orderGen,
		eventGenerator: eventGen,
		tradingEngine:  tradingEngine,
		activePatterns: make(map[string]*SimulationPattern),
		patternHistory: make([]PatternEvent, 0),
		orderSemaphore: make(chan struct{}, 10), // Limit to 10 concurrent orders
		status: SimulationStatus{
			IsRunning:     false,
			CurrentPrices: make(map[string]float64),
			ActivePatterns: make([]string, 0),
		},
		stats: SimulationStatistics{
			PerformanceMetrics: PerformanceMetrics{},
		},
	}
}

// StartSimulation begins the market simulation
func (rs *RealisticSimulator) StartSimulation(ctx context.Context, config SimulationConfig) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.isRunning {
		return SimulationError{
			Code:    ErrorCodeSimulationFailed,
			Message: "Simulation is already running",
		}
	}

	// Validate configuration
	if err := rs.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	rs.config = config
	rs.isRunning = true
	rs.stopChan = make(chan struct{})
	rs.doneChan = make(chan struct{})

	// Initialize status
	rs.status = SimulationStatus{
		IsRunning:       true,
		StartTime:       time.Now(),
		RunningDuration: 0,
		CurrentPrices:   make(map[string]float64),
		ActivePatterns:  make([]string, 0),
		MarketCondition: config.MarketCondition,
	}

	// Initialize statistics
	rs.stats = SimulationStatistics{
		StartTime:         time.Now(),
		PerformanceMetrics: PerformanceMetrics{},
	}

	// Copy initial prices
	for symbol, price := range config.InitialPrices {
		rs.status.CurrentPrices[symbol] = price
		rs.priceGenerator.SetBasePrice(symbol, price)
	}

	// Configure generators
	rs.orderGenerator.SetUserProfiles(config.UserProfiles)
	rs.orderGenerator.UpdateMarketSentiment(config.MarketSentiment)

	// Start background workers
	rs.startWorkers(ctx)

	log.Printf("Market simulation started with %d symbols for %v",
		len(config.Symbols), config.SimulationDuration)

	return nil
}

// StopSimulation gracefully stops the simulation
func (rs *RealisticSimulator) StopSimulation() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if !rs.isRunning {
		return SimulationError{
			Code:    ErrorCodeSimulationFailed,
			Message: "Simulation is not running",
		}
	}

	log.Println("Stopping market simulation...")

	// Signal all workers to stop
	close(rs.stopChan)

	// Wait for workers to finish (with timeout)
	select {
	case <-rs.doneChan:
		log.Println("All workers stopped gracefully")
	case <-time.After(5 * time.Second):
		log.Println("Warning: Workers did not stop within timeout")
	}

	// Update final status
	rs.isRunning = false
	rs.status.IsRunning = false
	rs.stats.TotalRuntime = time.Since(rs.stats.StartTime)

	// Reset generators
	rs.priceGenerator.Reset()

	log.Printf("Market simulation stopped. Total runtime: %v, Orders generated: %d",
		rs.stats.TotalRuntime, rs.stats.OrdersGenerated)

	return nil
}

// InjectVolatility introduces volatility into the market
func (rs *RealisticSimulator) InjectVolatility(pattern VolatilityPattern, duration time.Duration) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if !rs.isRunning {
		return SimulationError{
			Code:    ErrorCodeSimulationFailed,
			Message: "Cannot inject volatility when simulation is not running",
		}
	}

	log.Printf("Injecting volatility pattern: %s for %v", pattern, duration)

	// Calculate intensity based on pattern
	intensity := rs.calculateVolatilityIntensity(pattern)

	// Apply volatility to price generator
	rs.priceGenerator.SimulateVolatility(pattern, intensity)

	// Schedule volatility removal
	go func() {
		time.Sleep(duration)
		rs.priceGenerator.SimulateVolatility(VolatilityDecay, 0.1)
	}()

	rs.stats.EventsTriggered++
	return nil
}

// GetSimulationStatus returns current simulation state
func (rs *RealisticSimulator) GetSimulationStatus() SimulationStatus {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	// Update running duration
	if rs.isRunning {
		rs.status.RunningDuration = time.Since(rs.status.StartTime)
	}

	// Copy active patterns
	activePatterns := make([]string, 0, len(rs.activePatterns))
	for name := range rs.activePatterns {
		activePatterns = append(activePatterns, name)
	}
	rs.status.ActivePatterns = activePatterns

	// Copy statistics
	rs.status.OrdersGenerated = rs.stats.OrdersGenerated
	rs.status.PriceUpdates = rs.stats.PriceUpdates
	rs.status.LastError = rs.stats.LastError

	return rs.status
}

// UpdateConfig updates simulation parameters during runtime
func (rs *RealisticSimulator) UpdateConfig(config SimulationConfig) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if err := rs.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration update: %w", err)
	}

	rs.config = config

	// Update generators with new config
	rs.orderGenerator.SetUserProfiles(config.UserProfiles)
	rs.orderGenerator.UpdateMarketSentiment(config.MarketSentiment)

	// Update status
	rs.status.MarketCondition = config.MarketCondition

	log.Println("Simulation configuration updated during runtime")
	return nil
}

// Private helper methods

func (rs *RealisticSimulator) validateConfig(config SimulationConfig) error {
	if len(config.Symbols) == 0 {
		return SimulationError{
			Code:    ErrorCodeInvalidConfig,
			Message: "At least one symbol must be specified",
		}
	}

	if config.SimulationDuration <= 0 {
		return SimulationError{
			Code:    ErrorCodeInvalidConfig,
			Message: "Simulation duration must be positive",
		}
	}

	if config.TickInterval <= 0 {
		return SimulationError{
			Code:    ErrorCodeInvalidConfig,
			Message: "Tick interval must be positive",
		}
	}

	if config.OrderGenerationRate < 0 {
		return SimulationError{
			Code:    ErrorCodeInvalidConfig,
			Message: "Order generation rate cannot be negative",
		}
	}

	if len(config.InitialPrices) == 0 {
		return SimulationError{
			Code:    ErrorCodeInvalidConfig,
			Message: "Initial prices must be specified for all symbols",
		}
	}

	for _, symbol := range config.Symbols {
		if _, exists := config.InitialPrices[symbol]; !exists {
			return SimulationError{
				Code:    ErrorCodeInvalidConfig,
				Message: fmt.Sprintf("Initial price not specified for symbol: %s", symbol),
			}
		}
	}

	return nil
}

func (rs *RealisticSimulator) startWorkers(ctx context.Context) {

	// Price update worker
	rs.priceWorker = &Worker{
		name:     "price_worker",
		interval: rs.config.PriceUpdateFrequency,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		workFunc: rs.updatePrices,
	}

	// Order generation worker
	rs.orderWorker = &Worker{
		name:     "order_worker",
		interval: rs.config.TickInterval,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		workFunc: rs.generateOrders,
	}

	// Event generation worker
	rs.eventWorker = &Worker{
		name:     "event_worker",
		interval: rs.config.NewsEventFrequency,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		workFunc: rs.processEvents,
	}

	// Pattern management worker
	rs.patternWorker = &Worker{
		name:     "pattern_worker",
		interval: 5 * time.Second, // Check patterns every 5 seconds
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		workFunc: rs.managePatterns,
	}

	workers := []*Worker{rs.priceWorker, rs.orderWorker, rs.eventWorker, rs.patternWorker}

	// Start all workers
	for _, worker := range workers {
		go rs.runWorker(worker)
	}

	// Supervisor goroutine
	go func() {
		defer close(rs.doneChan)

		// Wait for stop signal or simulation duration
		select {
		case <-rs.stopChan:
			log.Println("Stop signal received")
		case <-time.After(rs.config.SimulationDuration):
			log.Println("Simulation duration completed")
		case <-ctx.Done():
			log.Println("Context cancelled")
		}

		// Stop all workers
		for _, worker := range workers {
			close(worker.stopChan)
		}

		// Wait for all workers to finish
		for _, worker := range workers {
			<-worker.doneChan
		}
	}()
}

func (rs *RealisticSimulator) runWorker(worker *Worker) {
	defer close(worker.doneChan)

	ticker := time.NewTicker(worker.interval)
	defer ticker.Stop()

	log.Printf("Starting worker: %s", worker.name)

	for {
		select {
		case <-worker.stopChan:
			log.Printf("Worker %s received stop signal", worker.name)
			return
		case <-ticker.C:
			start := time.Now()
			if err := worker.workFunc(); err != nil {
				rs.mu.Lock()
				rs.stats.ErrorCount++
				rs.stats.LastError = err
				rs.mu.Unlock()
				log.Printf("Worker %s error: %v", worker.name, err)
			}

			// Update performance metrics
			rs.updatePerformanceMetrics(worker.name, time.Since(start))
		}
	}
}

func (rs *RealisticSimulator) updatePrices() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	timeElapsed := time.Since(rs.stats.StartTime)

	for _, symbol := range rs.config.Symbols {
		currentPrice := rs.status.CurrentPrices[symbol]
		newPrice := rs.priceGenerator.GeneratePrice(symbol, currentPrice, timeElapsed)

		rs.status.CurrentPrices[symbol] = newPrice
		rs.stats.PriceUpdates++
	}

	return nil
}

func (rs *RealisticSimulator) generateOrders() error {
	rs.mu.RLock()
	config := rs.config
	currentPrices := make(map[string]float64)
	for k, v := range rs.status.CurrentPrices {
		currentPrices[k] = v
	}
	rs.mu.RUnlock()

	// Generate orders for each symbol
	for _, symbol := range config.Symbols {
		currentPrice := currentPrices[symbol]

		// Generate realistic orders based on market conditions
		orders := rs.orderGenerator.GenerateRealisticOrders(
			symbol,
			currentPrice,
			config.MarketCondition,
		)

		// Submit orders to trading engine
		for _, orderReq := range orders {
			if rs.tradingEngine != nil {
				_, err := rs.tradingEngine.PlaceOrder(orderReq)
				if err != nil {
					log.Printf("Failed to place simulated order: %v", err)
				}
			}
		}

		rs.mu.Lock()
		rs.stats.OrdersGenerated += int64(len(orders))
		rs.mu.Unlock()
	}

	return nil
}

func (rs *RealisticSimulator) processEvents() error {
	if rs.eventGenerator == nil {
		return nil
	}

	// Generate random market event
	if rand.Float64() < 0.3 { // 30% chance of event per interval
		event := rs.eventGenerator.GenerateMarketEvent()

		if err := rs.eventGenerator.InjectEvent(event); err != nil {
			return fmt.Errorf("failed to inject event: %w", err)
		}

		// Apply event effects
		rs.applyEventEffects(event)

		rs.mu.Lock()
		rs.stats.EventsTriggered++
		rs.mu.Unlock()

		log.Printf("Market event triggered: %s (%s)", event.Description, event.Type)
	}

	return nil
}

func (rs *RealisticSimulator) managePatterns() error {
	if !rs.config.EnablePatterns {
		return nil
	}

	// Check for pattern triggers based on market conditions
	rs.checkPatternTriggers()

	// Update active patterns
	rs.updateActivePatterns()

	return nil
}

func (rs *RealisticSimulator) calculateVolatilityIntensity(pattern VolatilityPattern) float64 {
	switch pattern {
	case VolatilitySpike:
		return 0.8 + rand.Float64()*0.2 // 0.8 to 1.0
	case VolatilityDecay:
		return 0.1 + rand.Float64()*0.2 // 0.1 to 0.3
	case VolatilityOscillate:
		return 0.4 + rand.Float64()*0.4 // 0.4 to 0.8
	case VolatilityRandom:
		return rand.Float64() // 0.0 to 1.0
	case VolatilityNews:
		return 0.5 + rand.Float64()*0.3 // 0.5 to 0.8
	default:
		return 0.3
	}
}

func (rs *RealisticSimulator) applyEventEffects(event MarketEvent) {
	// Apply price impact
	for _, symbol := range event.AffectedSymbols {
		if currentPrice, exists := rs.status.CurrentPrices[symbol]; exists {
			priceChange := currentPrice * (event.PriceImpact / 100.0)

			// Determine direction based on event type and severity
			direction := rs.determineEventDirection(event)
			newPrice := currentPrice + (priceChange * direction)

			if newPrice > 0 {
				rs.status.CurrentPrices[symbol] = newPrice
			}
		}
	}

	// Trigger user behavior changes
	behaviorPattern := rs.mapEventToBehavior(event)
	intensity := rs.mapSeverityToIntensity(event.Severity)

	orders := rs.orderGenerator.SimulateUserBehavior(behaviorPattern, intensity)

	// Submit behavior-driven orders
	for _, orderReq := range orders {
		if rs.tradingEngine != nil {
			go func(req dto.PlaceOrderRequest) {
				// Acquire semaphore to limit concurrent operations
				rs.orderSemaphore <- struct{}{}
				defer func() { <-rs.orderSemaphore }() // Release semaphore

				_, err := rs.tradingEngine.PlaceOrder(req)
				if err != nil {
					log.Printf("Failed to place event-driven order: %v", err)
				}
			}(orderReq)
		}
	}
}

func (rs *RealisticSimulator) determineEventDirection(event MarketEvent) float64 {
	switch event.Type {
	case EventEarnings:
		// Random direction for earnings
		if rand.Float64() < 0.6 {
			return 1.0 // Positive earnings
		}
		return -1.0
	case EventRegulatory:
		return -0.5 // Usually negative
	case EventEconomic:
		// Context-dependent
		return (rand.Float64() - 0.5) * 2
	default:
		return (rand.Float64() - 0.5) * 2 // Random direction
	}
}

func (rs *RealisticSimulator) mapEventToBehavior(event MarketEvent) UserBehaviorPattern {
	switch event.Severity {
	case SeverityCrisis:
		return BehaviorPanic
	case SeverityHigh:
		if event.PriceImpact > 0 {
			return BehaviorFOMO
		}
		return BehaviorPanic
	case SeverityMedium:
		return BehaviorMomentum
	default:
		return BehaviorConservative
	}
}

func (rs *RealisticSimulator) mapSeverityToIntensity(severity EventSeverity) float64 {
	switch severity {
	case SeverityCrisis:
		return 1.0
	case SeverityHigh:
		return 0.8
	case SeverityMedium:
		return 0.5
	case SeverityLow:
		return 0.2
	default:
		return 0.1
	}
}

func (rs *RealisticSimulator) checkPatternTriggers() {
	// Implementation for pattern trigger detection
	// This would check various market conditions and activate patterns
	// For now, implement basic random pattern activation

	for patternName, probability := range rs.config.PatternProbabilities {
		if rand.Float64() < probability {
			rs.activatePattern(patternName)
		}
	}
}

func (rs *RealisticSimulator) activatePattern(patternName string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if _, exists := rs.activePatterns[patternName]; exists {
		return // Pattern already active
	}

	// Create and activate pattern
	pattern := rs.createPattern(patternName)
	if pattern != nil {
		rs.activePatterns[patternName] = pattern
		rs.patternHistory = append(rs.patternHistory, PatternEvent{
			PatternName: patternName,
			EventType:   "activated",
			Timestamp:   time.Now(),
		})
		rs.stats.PatternsActivated++

		log.Printf("Pattern activated: %s", patternName)
	}
}

func (rs *RealisticSimulator) createPattern(patternName string) *SimulationPattern {
	// Create pattern based on name
	// This is a simplified implementation
	switch patternName {
	case "flash_crash":
		return &SimulationPattern{
			Name:        "Flash Crash",
			Description: "Sudden dramatic price decline",
			Duration:    2 * time.Minute,
		}
	case "fomo_spike":
		return &SimulationPattern{
			Name:        "FOMO Spike",
			Description: "Fear of missing out price spike",
			Duration:    5 * time.Minute,
		}
	case "whale_dump":
		return &SimulationPattern{
			Name:        "Whale Dump",
			Description: "Large holder selling pressure",
			Duration:    10 * time.Minute,
		}
	default:
		return nil
	}
}

func (rs *RealisticSimulator) updateActivePatterns() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Remove expired patterns
	for name, pattern := range rs.activePatterns {
		// Find activation time from history
		var activationTime time.Time
		for i := len(rs.patternHistory) - 1; i >= 0; i-- {
			if rs.patternHistory[i].PatternName == name && rs.patternHistory[i].EventType == "activated" {
				activationTime = rs.patternHistory[i].Timestamp
				break
			}
		}

		if time.Since(activationTime) > pattern.Duration {
			delete(rs.activePatterns, name)
			rs.patternHistory = append(rs.patternHistory, PatternEvent{
				PatternName: name,
				EventType:   "deactivated",
				Timestamp:   time.Now(),
				Duration:    time.Since(activationTime),
			})

			log.Printf("Pattern deactivated: %s (duration: %v)", name, time.Since(activationTime))
		}
	}
}

func (rs *RealisticSimulator) updatePerformanceMetrics(workerName string, duration time.Duration) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// Update specific metrics based on worker type
	switch workerName {
	case "order_worker":
		// Exponential moving average for order generation time
		if rs.stats.PerformanceMetrics.AvgOrderGenerationTime == 0 {
			rs.stats.PerformanceMetrics.AvgOrderGenerationTime = duration
		} else {
			alpha := 0.1 // Smoothing factor
			rs.stats.PerformanceMetrics.AvgOrderGenerationTime =
				time.Duration(float64(rs.stats.PerformanceMetrics.AvgOrderGenerationTime)*(1-alpha) +
							 float64(duration)*alpha)
		}
	case "price_worker":
		// Exponential moving average for price update time
		if rs.stats.PerformanceMetrics.AvgPriceUpdateTime == 0 {
			rs.stats.PerformanceMetrics.AvgPriceUpdateTime = duration
		} else {
			alpha := 0.1
			rs.stats.PerformanceMetrics.AvgPriceUpdateTime =
				time.Duration(float64(rs.stats.PerformanceMetrics.AvgPriceUpdateTime)*(1-alpha) +
							 float64(duration)*alpha)
		}
	}
}

// GetStatistics returns detailed simulation statistics
func (rs *RealisticSimulator) GetStatistics() SimulationStatistics {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	stats := rs.stats
	if rs.isRunning {
		stats.TotalRuntime = time.Since(rs.stats.StartTime)
	}

	return stats
}

// GetPatternHistory returns the history of pattern activations
func (rs *RealisticSimulator) GetPatternHistory() []PatternEvent {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	// Return a copy
	history := make([]PatternEvent, len(rs.patternHistory))
	copy(history, rs.patternHistory)
	return history
}