package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"simulated_exchange/internal/metrics"
	"simulated_exchange/internal/repository"
	"simulated_exchange/internal/types"
)

func TestMetricsTradingEngine_Integration(t *testing.T) {
	// Create trading engine components
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	// Create metrics components
	collector := metrics.NewRealTimeMetrics(60 * time.Second)
	analyzer := metrics.NewAIAnalyzer()
	metricsService := metrics.NewRealTimeMetricsService(collector, analyzer)

	// Start metrics service
	err := metricsService.Start()
	require.NoError(t, err)
	defer metricsService.Stop()

	// Create metrics-enabled trading engine
	engine := TradingEngineWithMetrics(orderRepo, tradeRepo, matcher, executor, metricsService)

	t.Run("order metrics collection", func(t *testing.T) {
		order := types.Order{
			ID:       "test_order_1",
			Symbol:   "AAPL",
			Side:     types.Buy,
			Type:     types.Limit,
			Quantity: 100,
			Price:    150.0,
		}

		err := engine.PlaceOrder(order)
		require.NoError(t, err)

		// Wait a moment for metrics to be processed
		time.Sleep(100 * time.Millisecond)

		// Verify metrics were collected
		currentMetrics := metricsService.GetRealTimeMetrics()
		assert.Equal(t, int64(1), currentMetrics.OrderCount)
		assert.Greater(t, currentMetrics.AvgLatency, time.Duration(0))

		// Verify symbol-specific metrics
		symbolMetrics, exists := currentMetrics.SymbolMetrics["AAPL"]
		assert.True(t, exists)
		assert.Equal(t, int64(1), symbolMetrics.OrderCount)
	})

	t.Run("trade metrics collection", func(t *testing.T) {
		// Reset metrics
		metricsService.ResetMetrics()

		// Place a sell order first
		sellOrder := types.Order{
			ID:       "sell_order_1",
			Symbol:   "AAPL",
			Side:     types.Sell,
			Type:     types.Limit,
			Quantity: 50,
			Price:    149.0,
		}

		err := engine.PlaceOrder(sellOrder)
		require.NoError(t, err)

		// Place a buy order that should match
		buyOrder := types.Order{
			ID:       "buy_order_1",
			Symbol:   "AAPL",
			Side:     types.Buy,
			Type:     types.Limit,
			Quantity: 50,
			Price:    149.0,
		}

		err = engine.PlaceOrder(buyOrder)
		require.NoError(t, err)

		// Wait for metrics processing
		time.Sleep(100 * time.Millisecond)

		// Verify both order and trade metrics
		currentMetrics := metricsService.GetRealTimeMetrics()
		assert.Equal(t, int64(2), currentMetrics.OrderCount) // Two orders placed
		assert.Equal(t, int64(1), currentMetrics.TradeCount) // One trade executed
		assert.Equal(t, 50.0, currentMetrics.TotalVolume)

		// Verify symbol metrics include trade data
		symbolMetrics, exists := currentMetrics.SymbolMetrics["AAPL"]
		assert.True(t, exists)
		assert.Equal(t, int64(2), symbolMetrics.OrderCount)
		assert.Equal(t, int64(1), symbolMetrics.TradeCount)
		assert.Equal(t, 50.0, symbolMetrics.Volume)
		assert.Equal(t, 149.0, symbolMetrics.AvgPrice)
	})

	t.Run("cancellation metrics", func(t *testing.T) {
		// Reset metrics
		metricsService.ResetMetrics()

		// Place an order
		order := types.Order{
			ID:       "cancel_test_order",
			Symbol:   "GOOGL",
			Side:     types.Buy,
			Type:     types.Limit,
			Quantity: 100,
			Price:    2800.0,
		}

		err := engine.PlaceOrder(order)
		require.NoError(t, err)

		// Cancel the order
		err = engine.CancelOrder("cancel_test_order")
		require.NoError(t, err)

		// Wait for metrics processing
		time.Sleep(100 * time.Millisecond)

		// Verify metrics include both placement and cancellation
		currentMetrics := metricsService.GetRealTimeMetrics()
		assert.GreaterOrEqual(t, currentMetrics.OrderCount, int64(1)) // At least the placement
		assert.Greater(t, currentMetrics.AvgLatency, time.Duration(0))
	})
}

func TestPerformanceMonitoredTradingEngine(t *testing.T) {
	// Create repositories
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	// Create performance thresholds
	thresholds := PerformanceThresholds{
		MaxLatency:            50 * time.Millisecond,
		MinThroughput:         10.0, // Low threshold for testing
		MaxBottleneckSeverity: 0.8,
	}

	// Create performance monitored engine
	engine, err := NewPerformanceMonitoredTradingEngine(
		orderRepo,
		tradeRepo,
		matcher,
		executor,
		thresholds,
	)
	require.NoError(t, err)
	defer engine.Stop()

	t.Run("basic functionality", func(t *testing.T) {
		order := types.Order{
			ID:       "perf_test_order",
			Symbol:   "AAPL",
			Side:     types.Buy,
			Type:     types.Limit,
			Quantity: 100,
			Price:    150.0,
		}

		err := engine.PlaceOrder(order)
		require.NoError(t, err)

		// Verify engine functionality
		orderBook, err := engine.GetOrderBook("AAPL")
		require.NoError(t, err)
		assert.Len(t, orderBook.Bids, 1)
	})

	t.Run("performance metrics retrieval", func(t *testing.T) {
		// Generate some activity
		for i := 0; i < 5; i++ {
			order := types.Order{
				Symbol:   "AAPL",
				Side:     types.Buy,
				Type:     types.Limit,
				Quantity: 10,
				Price:    150.0,
			}
			engine.PlaceOrder(order)
		}

		// Wait for metrics processing
		time.Sleep(100 * time.Millisecond)

		// Get performance metrics
		metricsSnapshot := engine.GetPerformanceMetrics()
		assert.Greater(t, metricsSnapshot.OrderCount, int64(0))
		assert.Greater(t, metricsSnapshot.AvgLatency, time.Duration(0))

		// Get performance analysis
		analysis := engine.GetPerformanceAnalysis()
		assert.NotEmpty(t, analysis.Timestamp)
	})

	t.Run("health monitoring", func(t *testing.T) {
		// Check initial health
		isHealthy := engine.IsPerformanceHealthy()
		assert.True(t, isHealthy)

		// Get detailed health status
		healthStatus := engine.GetHealthStatus()
		assert.True(t, healthStatus.IsHealthy)
		assert.True(t, healthStatus.MetricsHealthy)
		assert.GreaterOrEqual(t, healthStatus.CurrentLatency, time.Duration(0))
		assert.GreaterOrEqual(t, healthStatus.CurrentThroughput, 0.0)
	})

	t.Run("performance analysis over time", func(t *testing.T) {
		// Generate consistent activity to trigger analysis
		for i := 0; i < 10; i++ {
			sellOrder := types.Order{
				Symbol:   "PERF",
				Side:     types.Sell,
				Type:     types.Limit,
				Quantity: 10,
				Price:    100.0,
			}
			engine.PlaceOrder(sellOrder)

			buyOrder := types.Order{
				Symbol:   "PERF",
				Side:     types.Buy,
				Type:     types.Limit,
				Quantity: 10,
				Price:    100.0,
			}
			engine.PlaceOrder(buyOrder)

			// Small delay to spread events over time
			time.Sleep(10 * time.Millisecond)
		}

		// Wait for analysis
		time.Sleep(200 * time.Millisecond)

		// Check that analysis has been performed
		analysis := engine.GetPerformanceAnalysis()
		assert.NotEmpty(t, analysis.Timestamp)

		// Should have some recommendations
		assert.NotEmpty(t, analysis.Recommendations)

		// Get historical data
		historicalSnapshots := engine.metricsService.GetHistoricalSnapshots()
		assert.Greater(t, len(historicalSnapshots), 0)
	})
}

func TestMetricsTradeExecutor(t *testing.T) {
	// Create components
	tradeRepo := repository.NewMemoryTradeRepository()
	baseExecutor := NewSimpleTradeExecutor(tradeRepo)

	collector := metrics.NewRealTimeMetrics(60 * time.Second)
	analyzer := metrics.NewAIAnalyzer()
	metricsService := metrics.NewRealTimeMetricsService(collector, analyzer)

	err := metricsService.Start()
	require.NoError(t, err)
	defer metricsService.Stop()

	// Create metrics-enabled executor
	executor := NewMetricsTradeExecutor(baseExecutor, metricsService)

	t.Run("trade execution with metrics", func(t *testing.T) {
		buyOrder := types.Order{
			ID:       "buy_1",
			Symbol:   "AAPL",
			Side:     types.Buy,
			Quantity: 100,
		}

		sellOrder := types.Order{
			ID:       "sell_1",
			Symbol:   "AAPL",
			Side:     types.Sell,
			Quantity: 100,
		}

		trade, err := executor.ExecuteTrade(buyOrder, sellOrder, 50, 150.0)
		require.NoError(t, err)

		// Verify trade was created
		assert.Equal(t, "AAPL", trade.Symbol)
		assert.Equal(t, 50.0, trade.Quantity)
		assert.Equal(t, 150.0, trade.Price)

		// Wait for metrics processing
		time.Sleep(100 * time.Millisecond)

		// Verify metrics were recorded
		currentMetrics := metricsService.GetRealTimeMetrics()
		assert.Equal(t, int64(1), currentMetrics.TradeCount)
		assert.Equal(t, 50.0, currentMetrics.TotalVolume)
		assert.Greater(t, currentMetrics.AvgLatency, time.Duration(0))
	})
}

func TestMetricsIntegration_HighLoad(t *testing.T) {
	// Create components for high-load testing
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	collector := metrics.NewRealTimeMetrics(30 * time.Second) // Shorter window for testing
	analyzer := metrics.NewAIAnalyzer()
	metricsService := metrics.NewRealTimeMetricsService(collector, analyzer)

	err := metricsService.Start()
	require.NoError(t, err)
	defer metricsService.Stop()

	engine := TradingEngineWithMetrics(orderRepo, tradeRepo, matcher, executor, metricsService)

	t.Run("high load metrics collection", func(t *testing.T) {
		const numOrders = 100

		// Generate high load
		for i := 0; i < numOrders; i++ {
			order := types.Order{
				Symbol:   "LOAD",
				Side:     types.Buy,
				Type:     types.Limit,
				Quantity: 10,
				Price:    100.0,
			}

			err := engine.PlaceOrder(order)
			assert.NoError(t, err)
		}

		// Wait for metrics processing
		time.Sleep(500 * time.Millisecond)

		// Verify metrics captured high load
		currentMetrics := metricsService.GetRealTimeMetrics()
		assert.Equal(t, int64(numOrders), currentMetrics.OrderCount)
		assert.Greater(t, currentMetrics.OrdersPerSec, 0.0)

		// Verify symbol metrics
		symbolMetrics, exists := currentMetrics.SymbolMetrics["LOAD"]
		assert.True(t, exists)
		assert.Equal(t, int64(numOrders), symbolMetrics.OrderCount)
	})
}

func TestDefaultPerformanceThresholds(t *testing.T) {
	thresholds := DefaultPerformanceThresholds()

	assert.Equal(t, 100*time.Millisecond, thresholds.MaxLatency)
	assert.Equal(t, 100.0, thresholds.MinThroughput)
	assert.Equal(t, 0.8, thresholds.MaxBottleneckSeverity)
}

func TestHealthStatus(t *testing.T) {
	// Create engine with strict thresholds for testing violations
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	strictThresholds := PerformanceThresholds{
		MaxLatency:            1 * time.Nanosecond, // Impossible to meet
		MinThroughput:         1000000.0,           // Very high threshold
		MaxBottleneckSeverity: 0.01,                // Very low threshold
	}

	engine, err := NewPerformanceMonitoredTradingEngine(
		orderRepo,
		tradeRepo,
		matcher,
		executor,
		strictThresholds,
	)
	require.NoError(t, err)
	defer engine.Stop()

	// Generate some activity
	order := types.Order{
		Symbol:   "TEST",
		Side:     types.Buy,
		Type:     types.Limit,
		Quantity: 100,
		Price:    150.0,
	}
	engine.PlaceOrder(order)

	// Wait for metrics
	time.Sleep(200 * time.Millisecond)

	// Check health status with strict thresholds
	healthStatus := engine.GetHealthStatus()

	// Should detect threshold violations
	assert.False(t, healthStatus.IsHealthy)
	assert.Greater(t, len(healthStatus.ThresholdViolations), 0)
	assert.NotEmpty(t, healthStatus.Recommendations)
}