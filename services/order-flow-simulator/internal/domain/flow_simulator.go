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

// FlowSimulatorStatus represents the current status of the flow simulator
type FlowSimulatorStatus struct {
	IsRunning       bool                       `json:"is_running"`
	StartTime       time.Time                  `json:"start_time"`
	OrdersGenerated int64                      `json:"orders_generated"`
	OrdersSubmitted int64                      `json:"orders_submitted"`
	OrdersFailed    int64                      `json:"orders_failed"`
	ActiveUsers     int                        `json:"active_users"`
	SymbolStats     map[string]SymbolFlowStats `json:"symbol_stats"`
	LastUpdate      time.Time                  `json:"last_update"`
}

// SymbolFlowStats tracks statistics for a specific trading symbol
type SymbolFlowStats struct {
	Symbol          string  `json:"symbol"`
	OrdersGenerated int64   `json:"orders_generated"`
	OrdersSubmitted int64   `json:"orders_submitted"`
	OrderRate       float64 `json:"current_order_rate"`
	LastOrderTime   time.Time `json:"last_order_time"`
}

// FlowSimulator orchestrates the overall order flow simulation
type FlowSimulator struct {
	orderGenerator   *OrderGenerator
	userSimulator    *UserSimulator
	tradingAPIClient *TradingAPIClient
	eventBus         *messaging.RedisEventBus
	logger           *slog.Logger
	adaptiveThrottle *AdaptiveThrottle

	// State management
	isRunning     bool
	startTime     time.Time
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	statusMutex   sync.RWMutex

	// Statistics
	stats         FlowSimulatorStatus
	symbolStats   map[string]*SymbolFlowStats
	statsMutex    sync.RWMutex

	// Configuration
	orderSubmissionInterval time.Duration
	statisticsUpdateInterval time.Duration
	maxOrdersPerSecond      float64
}

// NewFlowSimulator creates a new flow simulator
func NewFlowSimulator(
	orderGenerator *OrderGenerator,
	userSimulator *UserSimulator,
	tradingAPIClient *TradingAPIClient,
	eventBus *messaging.RedisEventBus,
	logger *slog.Logger,
) *FlowSimulator {
	// Create adaptive throttle with EXTREMELY conservative settings
	adaptiveThrottle := NewAdaptiveThrottle(0.01, logger) // Start with 0.01 orders/second (1 order per 100 seconds)

	return &FlowSimulator{
		orderGenerator:           orderGenerator,
		userSimulator:            userSimulator,
		tradingAPIClient:         tradingAPIClient,
		eventBus:                 eventBus,
		logger:                   logger,
		adaptiveThrottle:         adaptiveThrottle,
		symbolStats:              make(map[string]*SymbolFlowStats),
		orderSubmissionInterval:  30 * time.Second, // MUCH slower: every 30 seconds
		statisticsUpdateInterval: 60 * time.Second,
		maxOrdersPerSecond:       0.02, // EXTREMELY conservative: 0.02 orders per second max
	}
}

// Start begins the order flow simulation
func (fs *FlowSimulator) Start(ctx context.Context) error {
	fs.statusMutex.Lock()
	defer fs.statusMutex.Unlock()

	if fs.isRunning {
		return fmt.Errorf("flow simulator is already running")
	}

	fs.logger.Info("Starting order flow simulation")

	// Create context for the simulation
	fs.ctx, fs.cancel = context.WithCancel(ctx)
	fs.isRunning = true
	fs.startTime = time.Now()

	// Initialize statistics
	fs.initializeStats()

	// Start user behavior simulation
	if err := fs.userSimulator.StartSimulation(fs.ctx); err != nil {
		fs.isRunning = false
		return fmt.Errorf("failed to start user simulation: %w", err)
	}

	// Start order generation and submission
	fs.wg.Add(1)
	go fs.runOrderGenerationLoop()

	// Start statistics collection
	fs.wg.Add(1)
	go fs.runStatisticsLoop()

	// Start event publishing
	fs.wg.Add(1)
	go fs.runEventPublishingLoop()

	fs.logger.Info("Order flow simulation started successfully")
	return nil
}

// Stop gracefully stops the order flow simulation
func (fs *FlowSimulator) Stop() error {
	fs.statusMutex.Lock()
	defer fs.statusMutex.Unlock()

	if !fs.isRunning {
		return fmt.Errorf("flow simulator is not running")
	}

	fs.logger.Info("Stopping order flow simulation")

	// Cancel context to stop all goroutines
	fs.cancel()

	// Wait for all goroutines to finish
	fs.wg.Wait()

	fs.isRunning = false
	fs.logger.Info("Order flow simulation stopped successfully")
	return nil
}

// GetStatus returns the current status of the flow simulator
func (fs *FlowSimulator) GetStatus() FlowSimulatorStatus {
	fs.statsMutex.RLock()
	defer fs.statsMutex.RUnlock()

	// Create a copy of the stats
	status := fs.stats
	status.ActiveUsers = fs.userSimulator.GetActiveUserCount()
	status.LastUpdate = time.Now()

	// Copy symbol stats
	status.SymbolStats = make(map[string]SymbolFlowStats)
	for symbol, stats := range fs.symbolStats {
		status.SymbolStats[symbol] = *stats
	}

	return status
}

// SetOrderRate adjusts the order generation rate for a symbol
func (fs *FlowSimulator) SetOrderRate(symbol string, rate float64) error {
	if rate < 0 || rate > fs.maxOrdersPerSecond {
		return fmt.Errorf("invalid order rate: %f (must be between 0 and %f)", rate, fs.maxOrdersPerSecond)
	}

	fs.orderGenerator.AdjustRateForSymbol(symbol, rate/fs.orderGenerator.config.BaseOrderRate)

	fs.logger.Info("Order rate adjusted",
		"symbol", symbol,
		"new_rate", rate,
	)

	return nil
}

// SetVolatilityMode enables or disables high volatility mode
func (fs *FlowSimulator) SetVolatilityMode(enabled bool) {
	fs.orderGenerator.SetVolatilityMode(enabled)
	fs.logger.Info("Volatility mode changed", "enabled", enabled)
}

// Private methods

func (fs *FlowSimulator) initializeStats() {
	symbols := fs.orderGenerator.GetSupportedSymbols()

	fs.statsMutex.Lock()
	defer fs.statsMutex.Unlock()

	fs.stats = FlowSimulatorStatus{
		IsRunning:       true,
		StartTime:       fs.startTime,
		OrdersGenerated: 0,
		OrdersSubmitted: 0,
		OrdersFailed:    0,
		ActiveUsers:     0,
		SymbolStats:     make(map[string]SymbolFlowStats),
		LastUpdate:      time.Now(),
	}

	for _, symbol := range symbols {
		fs.symbolStats[symbol] = &SymbolFlowStats{
			Symbol:          symbol,
			OrdersGenerated: 0,
			OrdersSubmitted: 0,
			OrderRate:       fs.orderGenerator.GetOrderRate(symbol),
			LastOrderTime:   time.Time{},
		}
	}
}

func (fs *FlowSimulator) runOrderGenerationLoop() {
	defer fs.wg.Done()

	ticker := time.NewTicker(fs.orderSubmissionInterval)
	defer ticker.Stop()

	symbols := fs.orderGenerator.GetSupportedSymbols()

	for {
		select {
		case <-fs.ctx.Done():
			return
		case <-ticker.C:
			// Generate orders for each symbol based on their rates
			for _, symbol := range symbols {
				fs.processSymbolOrders(symbol)
			}
		}
	}
}

func (fs *FlowSimulator) processSymbolOrders(symbol string) {
	// Check adaptive throttling first
	if fs.adaptiveThrottle.ShouldThrottle() {
		fs.logger.Debug("Order generation throttled by adaptive throttle")
		return
	}

	// Check if we can generate orders based on rate limiting
	if !fs.orderGenerator.CanGenerateOrder() {
		fs.logger.Debug("Order generation rate limited")
		return
	}

	// Check system health before generating orders
	if !fs.adaptiveThrottle.IsHealthy() {
		fs.logger.Warn("System health check failed, skipping order generation")
		return
	}

	orderRate := fs.adaptiveThrottle.GetCurrentRate() // Use throttled rate instead of base rate

	// Calculate probability of generating an order this tick
	ticksPerSecond := 1.0 / fs.orderSubmissionInterval.Seconds()
	probability := orderRate / ticksPerSecond

	if probability > 1.0 {
		probability = 1.0
	}

	// Random chance to generate an order (using more conservative probability)
	if fs.shouldGenerateOrder(probability * 0.5) { // Reduce probability by 50%
		fs.generateAndSubmitOrder(symbol)
	}
}

func (fs *FlowSimulator) shouldGenerateOrder(probability float64) bool {
	// Simple probability check
	// In a more sophisticated implementation, we might use Poisson distribution
	return probability > 0.5 // Simplified for this example
}

func (fs *FlowSimulator) generateAndSubmitOrder(symbol string) {
	// Get current price from user simulator's market state
	marketState, exists := fs.userSimulator.GetMarketState(symbol)
	currentPrice := 100.0 // Default price
	if exists {
		currentPrice = marketState.CurrentPrice
	}

	// Select user behavior type (this would normally come from active users)
	userTypes := []string{"conservative", "aggressive", "momentum"}
	userType := userTypes[len(userTypes)%3] // Simple rotation

	// Generate order
	order, err := fs.orderGenerator.GenerateOrder(fs.ctx, userType, symbol, currentPrice)
	if err != nil {
		fs.logger.Error("Failed to generate order", "error", err, "symbol", symbol)
		fs.adaptiveThrottle.RecordError() // Record generation error
		fs.incrementStat("orders_failed")
		return
	}

	fs.incrementStat("orders_generated")
	fs.incrementSymbolStat(symbol, "orders_generated")

	// Add to buffer for batch processing
	fs.orderGenerator.AddToBuffer(order)

	// Check if buffer should be flushed
	if fs.orderGenerator.ShouldFlushBuffer() {
		bufferedOrders := fs.orderGenerator.FlushBuffer()
		if len(bufferedOrders) > 0 {
			// Submit batch of orders
			go fs.submitOrderBatch(bufferedOrders)
		}
	}
}

func (fs *FlowSimulator) submitOrder(order *shared.Order) {
	err := fs.tradingAPIClient.SubmitOrder(fs.ctx, order)
	if err != nil {
		fs.logger.Error("Failed to submit order",
			"error", err,
			"order_id", order.ID,
			"symbol", order.Symbol,
		)
		fs.adaptiveThrottle.RecordError() // Record submission error
		fs.incrementStat("orders_failed")
		return
	}

	fs.adaptiveThrottle.RecordSuccess() // Record successful submission
	fs.incrementStat("orders_submitted")
	fs.incrementSymbolStat(order.Symbol, "orders_submitted")

	fs.logger.Debug("Order submitted successfully",
		"order_id", order.ID,
		"symbol", order.Symbol,
		"type", order.Type,
		"side", order.Side,
		"quantity", order.Quantity,
		"price", order.Price,
	)
}

// submitOrderBatch submits multiple orders in a batch for better performance
func (fs *FlowSimulator) submitOrderBatch(orders []*shared.Order) {
	for _, order := range orders {
		// Add small delay between order submissions in batch
		time.Sleep(fs.adaptiveThrottle.GetThrottleDelay())
		fs.submitOrder(order)
	}
}

func (fs *FlowSimulator) runStatisticsLoop() {
	defer fs.wg.Done()

	ticker := time.NewTicker(fs.statisticsUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fs.ctx.Done():
			return
		case <-ticker.C:
			fs.updateStatistics()
		}
	}
}

func (fs *FlowSimulator) updateStatistics() {
	fs.statsMutex.Lock()
	defer fs.statsMutex.Unlock()

	fs.stats.LastUpdate = time.Now()
	fs.stats.ActiveUsers = fs.userSimulator.GetActiveUserCount()

	// Update order rates for symbols
	for symbol, stats := range fs.symbolStats {
		stats.OrderRate = fs.orderGenerator.GetOrderRate(symbol)
	}

	fs.logger.Debug("Statistics updated",
		"orders_generated", fs.stats.OrdersGenerated,
		"orders_submitted", fs.stats.OrdersSubmitted,
		"active_users", fs.stats.ActiveUsers,
	)
}

func (fs *FlowSimulator) runEventPublishingLoop() {
	defer fs.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-fs.ctx.Done():
			return
		case <-ticker.C:
			fs.publishStatusEvent()
		}
	}
}

func (fs *FlowSimulator) publishStatusEvent() {
	status := fs.GetStatus()

	event := &shared.Event{
		ID:        "flow-status-" + fmt.Sprintf("%d", time.Now().Unix()),
		Type:      "flow_simulator_status",
		Source:    "order-flow-simulator",
		Data: map[string]interface{}{
			"is_running":        status.IsRunning,
			"orders_generated":  status.OrdersGenerated,
			"orders_submitted":  status.OrdersSubmitted,
			"orders_failed":     status.OrdersFailed,
			"active_users":      status.ActiveUsers,
		},
		Timestamp: time.Now(),
	}

	if err := fs.eventBus.Publish(fs.ctx, event); err != nil {
		fs.logger.Error("Failed to publish status event", "error", err)
	}
}

func (fs *FlowSimulator) incrementStat(statName string) {
	fs.statsMutex.Lock()
	defer fs.statsMutex.Unlock()

	switch statName {
	case "orders_generated":
		fs.stats.OrdersGenerated++
	case "orders_submitted":
		fs.stats.OrdersSubmitted++
	case "orders_failed":
		fs.stats.OrdersFailed++
	}
}

func (fs *FlowSimulator) incrementSymbolStat(symbol string, statName string) {
	fs.statsMutex.Lock()
	defer fs.statsMutex.Unlock()

	stats, exists := fs.symbolStats[symbol]
	if !exists {
		return
	}

	switch statName {
	case "orders_generated":
		stats.OrdersGenerated++
	case "orders_submitted":
		stats.OrdersSubmitted++
		stats.LastOrderTime = time.Now()
	}
}