package engine

import (
	"fmt"
	"time"

	"simulated_exchange/internal/metrics"
	"simulated_exchange/internal/types"
)

// MetricsTradingEngine is a decorator that adds metrics collection to any TradingEngine
type MetricsTradingEngine struct {
	engine         OrderProcessor
	metricsService MetricsService
}

// MetricsService interface for collecting metrics
type MetricsService interface {
	RecordOrderEvent(event metrics.OrderEvent)
	RecordTradeEvent(event metrics.TradeEvent)
	IsHealthy() bool
}

// NewMetricsTradingEngine creates a new metrics-enabled trading engine
func NewMetricsTradingEngine(engine OrderProcessor, metricsService MetricsService) *MetricsTradingEngine {
	return &MetricsTradingEngine{
		engine:         engine,
		metricsService: metricsService,
	}
}

// PlaceOrder places an order and records metrics
func (mte *MetricsTradingEngine) PlaceOrder(order types.Order) error {
	startTime := time.Now()

	// Execute the order
	err := mte.engine.PlaceOrder(order)

	// Calculate latency
	latency := time.Since(startTime)

	// Record metrics if service is healthy
	if mte.metricsService.IsHealthy() {
		event := metrics.OrderEvent{
			OrderID:   order.ID,
			Symbol:    order.Symbol,
			Side:      order.Side,
			Type:      order.Type,
			Quantity:  order.Quantity,
			Price:     order.Price,
			Timestamp: time.Now(),
			Latency:   latency,
		}
		mte.metricsService.RecordOrderEvent(event)
	}

	return err
}

// CancelOrder cancels an order and records metrics
func (mte *MetricsTradingEngine) CancelOrder(id string) error {
	startTime := time.Now()

	// Execute the cancellation
	err := mte.engine.CancelOrder(id)

	// Calculate latency
	latency := time.Since(startTime)

	// Record metrics if service is healthy (for cancellation operations)
	if mte.metricsService.IsHealthy() {
		event := metrics.OrderEvent{
			OrderID:   id,
			Symbol:    "CANCEL", // Special symbol for cancellation events
			Side:      "CANCEL",
			Type:      "CANCEL",
			Quantity:  0,
			Price:     0,
			Timestamp: time.Now(),
			Latency:   latency,
		}
		mte.metricsService.RecordOrderEvent(event)
	}

	return err
}

// GetOrderBook retrieves the order book (passthrough)
func (mte *MetricsTradingEngine) GetOrderBook(symbol string) (types.OrderBook, error) {
	return mte.engine.GetOrderBook(symbol)
}

// MetricsTradeExecutor is a decorator that adds metrics collection to trade execution
type MetricsTradeExecutor struct {
	executor       TradeExecutor
	metricsService MetricsService
}

// NewMetricsTradeExecutor creates a new metrics-enabled trade executor
func NewMetricsTradeExecutor(executor TradeExecutor, metricsService MetricsService) *MetricsTradeExecutor {
	return &MetricsTradeExecutor{
		executor:       executor,
		metricsService: metricsService,
	}
}

// ExecuteTrade executes a trade and records metrics
func (mte *MetricsTradeExecutor) ExecuteTrade(buyOrder types.Order, sellOrder types.Order, quantity float64, price float64) (types.Trade, error) {
	startTime := time.Now()

	// Execute the trade
	trade, err := mte.executor.ExecuteTrade(buyOrder, sellOrder, quantity, price)

	// Calculate latency
	latency := time.Since(startTime)

	// Record metrics if successful and service is healthy
	if err == nil && mte.metricsService.IsHealthy() {
		event := metrics.TradeEvent{
			TradeID:     trade.ID,
			Symbol:      trade.Symbol,
			Quantity:    trade.Quantity,
			Price:       trade.Price,
			Timestamp:   time.Now(),
			Latency:     latency,
			BuyOrderID:  trade.BuyOrderID,
			SellOrderID: trade.SellOrderID,
		}
		mte.metricsService.RecordTradeEvent(event)
	}

	return trade, err
}

// TradingEngineWithMetrics creates a complete trading engine with metrics integration
func TradingEngineWithMetrics(
	orderRepo OrderRepository,
	tradeRepo TradeRepository,
	matcher OrderMatcher,
	executor TradeExecutor,
	metricsService MetricsService,
) OrderProcessor {
	// Create metrics-enabled trade executor
	metricsExecutor := NewMetricsTradeExecutor(executor, metricsService)

	// Create base trading engine with metrics executor
	baseEngine := NewTradingEngine(orderRepo, tradeRepo, matcher, metricsExecutor)

	// Wrap with metrics collection for orders
	return NewMetricsTradingEngine(baseEngine, metricsService)
}

// PerformanceMonitoredTradingEngine combines trading engine with performance monitoring
type PerformanceMonitoredTradingEngine struct {
	engine                *MetricsTradingEngine
	metricsService        *metrics.RealTimeMetricsService
	performanceThresholds PerformanceThresholds
}

// PerformanceThresholds defines acceptable performance limits
type PerformanceThresholds struct {
	MaxLatency            time.Duration
	MinThroughput         float64
	MaxBottleneckSeverity float64
}

// NewPerformanceMonitoredTradingEngine creates a fully monitored trading engine
func NewPerformanceMonitoredTradingEngine(
	orderRepo OrderRepository,
	tradeRepo TradeRepository,
	matcher OrderMatcher,
	executor TradeExecutor,
	thresholds PerformanceThresholds,
) (*PerformanceMonitoredTradingEngine, error) {
	// Create metrics components
	collector := metrics.NewRealTimeMetrics(60 * time.Second)
	analyzer := metrics.NewAIAnalyzer()
	metricsService := metrics.NewRealTimeMetricsService(collector, analyzer)

	// Start metrics collection
	if err := metricsService.Start(); err != nil {
		return nil, fmt.Errorf("failed to start metrics service: %w", err)
	}

	// Create metrics-enabled trading engine
	tradingEngine := TradingEngineWithMetrics(
		orderRepo,
		tradeRepo,
		matcher,
		executor,
		metricsService,
	).(*MetricsTradingEngine)

	return &PerformanceMonitoredTradingEngine{
		engine:                tradingEngine,
		metricsService:        metricsService,
		performanceThresholds: thresholds,
	}, nil
}

// PlaceOrder places an order with performance monitoring
func (pmte *PerformanceMonitoredTradingEngine) PlaceOrder(order types.Order) error {
	return pmte.engine.PlaceOrder(order)
}

// CancelOrder cancels an order with performance monitoring
func (pmte *PerformanceMonitoredTradingEngine) CancelOrder(id string) error {
	return pmte.engine.CancelOrder(id)
}

// GetOrderBook retrieves the order book
func (pmte *PerformanceMonitoredTradingEngine) GetOrderBook(symbol string) (types.OrderBook, error) {
	return pmte.engine.GetOrderBook(symbol)
}

// GetPerformanceMetrics returns current performance metrics
func (pmte *PerformanceMonitoredTradingEngine) GetPerformanceMetrics() metrics.MetricsSnapshot {
	return pmte.metricsService.GetRealTimeMetrics()
}

// GetPerformanceAnalysis returns current performance analysis
func (pmte *PerformanceMonitoredTradingEngine) GetPerformanceAnalysis() metrics.PerformanceAnalysis {
	return pmte.metricsService.GetPerformanceAnalysis()
}

// IsPerformanceHealthy checks if performance is within acceptable thresholds
func (pmte *PerformanceMonitoredTradingEngine) IsPerformanceHealthy() bool {
	if !pmte.metricsService.IsHealthy() {
		return false
	}

	metricsSnapshot := pmte.GetPerformanceMetrics()
	analysis := pmte.GetPerformanceAnalysis()

	// Check latency threshold
	if metricsSnapshot.AvgLatency > pmte.performanceThresholds.MaxLatency {
		return false
	}

	// Check throughput threshold
	totalThroughput := metricsSnapshot.OrdersPerSec + metricsSnapshot.TradesPerSec
	if totalThroughput < pmte.performanceThresholds.MinThroughput {
		return false
	}

	// Check bottleneck severity
	for _, bottleneck := range analysis.Bottlenecks {
		if bottleneck.Severity > pmte.performanceThresholds.MaxBottleneckSeverity {
			return false
		}
	}

	return true
}

// GetHealthStatus returns detailed health status
func (pmte *PerformanceMonitoredTradingEngine) GetHealthStatus() HealthStatus {
	isHealthy := pmte.IsPerformanceHealthy()
	metricsSnapshot := pmte.GetPerformanceMetrics()
	analysis := pmte.GetPerformanceAnalysis()

	status := HealthStatus{
		IsHealthy:         isHealthy,
		MetricsHealthy:    pmte.metricsService.IsHealthy(),
		CurrentLatency:    metricsSnapshot.AvgLatency,
		CurrentThroughput: metricsSnapshot.OrdersPerSec + metricsSnapshot.TradesPerSec,
		BottleneckCount:   len(analysis.Bottlenecks),
		Recommendations:   analysis.Recommendations,
	}

	// Add threshold violations
	if metricsSnapshot.AvgLatency > pmte.performanceThresholds.MaxLatency {
		status.ThresholdViolations = append(status.ThresholdViolations,
			fmt.Sprintf("Latency %v exceeds threshold %v", metricsSnapshot.AvgLatency, pmte.performanceThresholds.MaxLatency))
	}

	totalThroughput := metricsSnapshot.OrdersPerSec + metricsSnapshot.TradesPerSec
	if totalThroughput < pmte.performanceThresholds.MinThroughput {
		status.ThresholdViolations = append(status.ThresholdViolations,
			fmt.Sprintf("Throughput %.2f below threshold %.2f", totalThroughput, pmte.performanceThresholds.MinThroughput))
	}

	return status
}

// Stop stops the performance monitoring
func (pmte *PerformanceMonitoredTradingEngine) Stop() error {
	return pmte.metricsService.Stop()
}

// HealthStatus represents the health status of the trading engine
type HealthStatus struct {
	IsHealthy             bool
	MetricsHealthy        bool
	CurrentLatency        time.Duration
	CurrentThroughput     float64
	BottleneckCount       int
	ThresholdViolations   []string
	Recommendations       []string
}

// DefaultPerformanceThresholds returns sensible default performance thresholds
func DefaultPerformanceThresholds() PerformanceThresholds {
	return PerformanceThresholds{
		MaxLatency:            100 * time.Millisecond,
		MinThroughput:         100.0, // orders+trades per second
		MaxBottleneckSeverity: 0.8,   // 80% severity threshold
	}
}