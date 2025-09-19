package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"simulated_exchange/internal/api/dto"
	"simulated_exchange/test/fixtures"
	"simulated_exchange/test/helpers"
)

// TradingFlowTestSuite tests complete trading workflows end-to-end
type TradingFlowTestSuite struct {
	suite.Suite
	server     *helpers.TestServer
	httpClient *helpers.HTTPClient
	testData   *fixtures.TestData
	assertions *helpers.AssertionHelpers
}

// SetupSuite initializes the test suite
func (suite *TradingFlowTestSuite) SetupSuite() {
	suite.server = helpers.NewTestServer(suite.T())
	suite.httpClient = helpers.NewHTTPClient(suite.T())
	suite.testData = fixtures.GetTestData()
	suite.assertions = helpers.NewAssertionHelpers(suite.T())

	// Start the demo system
	ctx := context.Background()
	err := suite.server.DemoSystem.Start(ctx)
	require.NoError(suite.T(), err)

	// Wait for system to be ready
	helpers.WaitForCondition(suite.T(), func() bool {
		health, err := suite.httpClient.GetHealth(suite.server.GetURL(""))
		return err == nil && health.Status == "healthy"
	}, 30*time.Second, "System should be healthy")
}

// TearDownSuite cleans up after test suite
func (suite *TradingFlowTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// SetupTest resets system state before each test
func (suite *TradingFlowTestSuite) SetupTest() {
	ctx := context.Background()
	err := suite.server.DemoSystem.Controller.ResetSystem(ctx)
	require.NoError(suite.T(), err)

	// Wait for reset to complete
	time.Sleep(100 * time.Millisecond)
}

// TestOrderPlacementFlow tests the complete order placement workflow
func (suite *TradingFlowTestSuite) TestOrderPlacementFlow() {
	baseURL := suite.server.GetURL("")
	orders := suite.testData.Orders

	suite.T().Run("SingleOrderPlacement", func(t *testing.T) {
		order := orders[0] // Market buy order

		// Place order
		resp, err := suite.httpClient.PlaceOrder(baseURL, order)
		require.NoError(t, err)
		suite.assertions.AssertOrderPlacementSuccess(resp, order.Symbol)

		// Extract order ID from response
		orderData, ok := resp.Data.(map[string]interface{})
		require.True(t, ok, "Response data should be a map")
		orderID, ok := orderData["order_id"].(string)
		require.True(t, ok, "Should have order_id in response")
		require.NotEmpty(t, orderID, "Order ID should not be empty")

		// Verify order can be retrieved
		getResp, err := suite.httpClient.GetOrder(baseURL, orderID)
		require.NoError(t, err)
		assert.True(t, getResp.Success, "Should successfully retrieve order")
	})

	suite.T().Run("MultipleOrderTypes", func(t *testing.T) {
		// Test different order types
		testCases := []struct {
			name  string
			order dto.PlaceOrderRequest
		}{
			{"MarketBuy", orders[0]},
			{"MarketSell", orders[1]},
			{"LimitBuy", orders[2]},
			{"LimitSell", orders[3]},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resp, err := suite.httpClient.PlaceOrder(baseURL, tc.order)
				require.NoError(t, err)
				suite.assertions.AssertOrderPlacementSuccess(resp, tc.order.Symbol)
			})
		}
	})

	suite.T().Run("OrderCancellation", func(t *testing.T) {
		order := orders[2] // Limit order for cancellation

		// Place order
		resp, err := suite.httpClient.PlaceOrder(baseURL, order)
		require.NoError(t, err)
		suite.assertions.AssertOrderPlacementSuccess(resp, order.Symbol)

		// Extract order ID
		orderData := resp.Data.(map[string]interface{})
		orderID := orderData["order_id"].(string)

		// Cancel order
		cancelResp, err := suite.httpClient.CancelOrder(baseURL, orderID)
		require.NoError(t, err)
		assert.True(t, cancelResp.Success, "Order cancellation should succeed")

		// Verify order status is cancelled
		getResp, err := suite.httpClient.GetOrder(baseURL, orderID)
		require.NoError(t, err)
		if getResp.Success {
			orderInfo := getResp.Data.(map[string]interface{})
			status, exists := orderInfo["status"]
			if exists {
				assert.Equal(t, "cancelled", status, "Order status should be cancelled")
			}
		}
	})
}

// TestOrderValidation tests order validation scenarios
func (suite *TradingFlowTestSuite) TestOrderValidation() {
	baseURL := suite.server.GetURL("")

	suite.T().Run("InvalidOrderData", func(t *testing.T) {
		invalidOrders := []dto.PlaceOrderRequest{
			// Invalid symbol
			{Symbol: "", Side: "buy", Type: "market", Quantity: 1.0, Price: 100.0},
			// Invalid side
			{Symbol: "BTCUSD", Side: "invalid", Type: "market", Quantity: 1.0, Price: 100.0},
			// Invalid type
			{Symbol: "BTCUSD", Side: "buy", Type: "invalid", Quantity: 1.0, Price: 100.0},
			// Zero quantity
			{Symbol: "BTCUSD", Side: "buy", Type: "market", Quantity: 0, Price: 100.0},
			// Negative quantity
			{Symbol: "BTCUSD", Side: "buy", Type: "market", Quantity: -1.0, Price: 100.0},
			// Zero price for limit order
			{Symbol: "BTCUSD", Side: "buy", Type: "limit", Quantity: 1.0, Price: 0},
		}

		for i, order := range invalidOrders {
			t.Run(assert.CallerInfo()[0]+":"+assert.ObjectsAreEqual, func(t *testing.T) {
				resp, err := suite.httpClient.PlaceOrder(baseURL, order)
				// Either HTTP error or API error response is acceptable
				if err != nil {
					// HTTP-level error (400, 422, etc.)
					t.Logf("Order %d correctly rejected with HTTP error: %v", i, err)
				} else {
					// API-level error
					assert.False(t, resp.Success, "Invalid order %d should be rejected", i)
					assert.NotNil(t, resp.Error, "Should have error details for invalid order %d", i)
				}
			})
		}
	})

	suite.T().Run("EdgeCaseOrders", func(t *testing.T) {
		edgeCaseOrders := []dto.PlaceOrderRequest{
			// Very small quantity
			{Symbol: "BTCUSD", Side: "buy", Type: "limit", Quantity: 0.001, Price: 50000.0},
			// Very large quantity
			{Symbol: "BTCUSD", Side: "sell", Type: "limit", Quantity: 999999.0, Price: 50000.0},
			// Very high price
			{Symbol: "BTCUSD", Side: "buy", Type: "limit", Quantity: 1.0, Price: 999999.0},
			// Very low price
			{Symbol: "BTCUSD", Side: "sell", Type: "limit", Quantity: 1.0, Price: 0.01},
		}

		for i, order := range edgeCaseOrders {
			t.Run(assert.CallerInfo()[0]+":"+assert.ObjectsAreEqual, func(t *testing.T) {
				resp, err := suite.httpClient.PlaceOrder(baseURL, order)
				require.NoError(t, err, "Edge case order %d should not cause HTTP error", i)
				// These orders might be accepted or rejected based on business rules
				t.Logf("Edge case order %d result: success=%v", i, resp.Success)
			})
		}
	})
}

// TestOrderBookIntegration tests order book integration
func (suite *TradingFlowTestSuite) TestOrderBookIntegration() {
	baseURL := suite.server.GetURL("")

	suite.T().Run("OrderBookUpdates", func(t *testing.T) {
		symbol := "BTCUSD"

		// Place multiple orders to populate order book
		orders := []dto.PlaceOrderRequest{
			{Symbol: symbol, Side: "buy", Type: "limit", Quantity: 1.0, Price: 49000.0},
			{Symbol: symbol, Side: "buy", Type: "limit", Quantity: 2.0, Price: 48900.0},
			{Symbol: symbol, Side: "sell", Type: "limit", Quantity: 1.5, Price: 51000.0},
			{Symbol: symbol, Side: "sell", Type: "limit", Quantity: 2.5, Price: 51100.0},
		}

		for _, order := range orders {
			resp, err := suite.httpClient.PlaceOrder(baseURL, order)
			require.NoError(t, err)
			suite.assertions.AssertOrderPlacementSuccess(resp, order.Symbol)
		}

		// Wait for order book to update
		time.Sleep(500 * time.Millisecond)

		// Verify metrics reflect the orders
		metrics, err := suite.httpClient.GetMetrics(baseURL)
		require.NoError(t, err)
		assert.Greater(t, metrics.OrderCount, int64(0), "Should have processed orders")

		// Check symbol-specific metrics
		if symbolMetrics, exists := metrics.SymbolMetrics[symbol]; exists {
			assert.Greater(t, symbolMetrics.OrderCount, int64(0), "Symbol should have order count")
		}
	})
}

// TestConcurrentOrders tests concurrent order processing
func (suite *TradingFlowTestSuite) TestConcurrentOrders() {
	baseURL := suite.server.GetURL("")
	loadRunner := helpers.NewLoadTestRunner(suite.T(), baseURL)

	suite.T().Run("LowConcurrency", func(t *testing.T) {
		// Generate orders for concurrent execution
		orders := make([]dto.PlaceOrderRequest, 20)
		for i := 0; i < len(orders); i++ {
			orders[i] = helpers.RandomOrder(suite.testData.Symbols)
		}

		// Run with low concurrency
		results, err := loadRunner.RunConcurrentOrders(orders, 5)
		require.NoError(t, err)

		// Verify results
		assert.Equal(t, len(orders), results.TotalOrders, "Should process all orders")
		assert.Greater(t, results.GetSuccessRate(), 0.8, "Should have high success rate")
		assert.Greater(t, results.GetThroughput(), 10.0, "Should achieve reasonable throughput")

		// Check latency is reasonable
		p95Latency := results.GetLatencyPercentile(95)
		assert.Less(t, p95Latency, 200*time.Millisecond, "P95 latency should be reasonable")
	})

	suite.T().Run("MediumConcurrency", func(t *testing.T) {
		// Generate more orders for medium concurrency test
		orders := make([]dto.PlaceOrderRequest, 50)
		for i := 0; i < len(orders); i++ {
			orders[i] = helpers.RandomOrder(suite.testData.Symbols)
		}

		// Run with medium concurrency
		results, err := loadRunner.RunConcurrentOrders(orders, 15)
		require.NoError(t, err)

		// Verify results
		assert.Equal(t, len(orders), results.TotalOrders, "Should process all orders")
		assert.Greater(t, results.GetSuccessRate(), 0.7, "Should maintain good success rate")
		assert.Greater(t, results.GetThroughput(), 20.0, "Should achieve higher throughput")

		// Log performance metrics
		t.Logf("Medium concurrency results: Success Rate: %.2f%%, Throughput: %.2f ops/s, P95 Latency: %v",
			results.GetSuccessRate()*100, results.GetThroughput(), results.GetLatencyPercentile(95))
	})
}

// TestMetricsCollection tests metrics collection during trading
func (suite *TradingFlowTestSuite) TestMetricsCollection() {
	baseURL := suite.server.GetURL("")

	suite.T().Run("MetricsAfterTrading", func(t *testing.T) {
		// Get initial metrics
		initialMetrics, err := suite.httpClient.GetMetrics(baseURL)
		require.NoError(t, err)
		initialOrderCount := initialMetrics.OrderCount

		// Place several orders
		orders := suite.testData.Orders[:4] // Use first 4 orders
		for _, order := range orders {
			resp, err := suite.httpClient.PlaceOrder(baseURL, order)
			require.NoError(t, err)
			suite.assertions.AssertOrderPlacementSuccess(resp, order.Symbol)
		}

		// Wait for metrics to update
		time.Sleep(2 * time.Second)

		// Get updated metrics
		updatedMetrics, err := suite.httpClient.GetMetrics(baseURL)
		require.NoError(t, err)

		// Verify metrics have been updated
		assert.GreaterOrEqual(t, updatedMetrics.OrderCount, initialOrderCount+int64(len(orders)),
			"Order count should increase after placing orders")

		// Verify metrics structure
		assert.NotEmpty(t, updatedMetrics.AvgLatency, "Should have average latency")
		assert.GreaterOrEqual(t, updatedMetrics.OrdersPerSec, 0.0, "Should have orders per second metric")

		// Verify symbol metrics
		assert.NotEmpty(t, updatedMetrics.SymbolMetrics, "Should have symbol-specific metrics")
		for symbol, metrics := range updatedMetrics.SymbolMetrics {
			assert.NotEmpty(t, symbol, "Symbol should not be empty")
			assert.GreaterOrEqual(t, metrics.OrderCount, int64(0), "Symbol order count should be non-negative")
		}
	})

	suite.T().Run("RealTimeMetricsUpdates", func(t *testing.T) {
		// Monitor metrics over time while placing orders
		metricsHistory := make([]*dto.MetricsResponse, 0)

		// Collect initial metrics
		metrics, err := suite.httpClient.GetMetrics(baseURL)
		require.NoError(t, err)
		metricsHistory = append(metricsHistory, metrics)

		// Place orders with intervals and collect metrics
		for i := 0; i < 5; i++ {
			order := helpers.RandomOrder(suite.testData.Symbols)
			resp, err := suite.httpClient.PlaceOrder(baseURL, order)
			require.NoError(t, err)
			suite.assertions.AssertOrderPlacementSuccess(resp, order.Symbol)

			// Wait and collect metrics
			time.Sleep(500 * time.Millisecond)
			metrics, err := suite.httpClient.GetMetrics(baseURL)
			require.NoError(t, err)
			metricsHistory = append(metricsHistory, metrics)
		}

		// Verify metrics are progressing
		assert.Len(t, metricsHistory, 6, "Should have collected 6 metric snapshots")

		// Check that order count is generally increasing
		for i := 1; i < len(metricsHistory); i++ {
			current := metricsHistory[i]
			previous := metricsHistory[i-1]
			assert.GreaterOrEqual(t, current.OrderCount, previous.OrderCount,
				"Order count should not decrease over time")
		}
	})
}

// TestErrorHandling tests error handling in trading flows
func (suite *TradingFlowTestSuite) TestErrorHandling() {
	baseURL := suite.server.GetURL("")

	suite.T().Run("ServerErrorRecovery", func(t *testing.T) {
		// This test simulates server error conditions and recovery
		// Note: In a real scenario, you might inject faults or use chaos engineering

		// Try to place an order with invalid data to trigger error handling
		invalidOrder := dto.PlaceOrderRequest{
			Symbol:   "INVALID_SYMBOL_THAT_SHOULD_NOT_EXIST",
			Side:     "buy",
			Type:     "market",
			Quantity: 1.0,
			Price:    100.0,
		}

		resp, err := suite.httpClient.PlaceOrder(baseURL, invalidOrder)
		// System should handle this gracefully
		if err != nil {
			t.Logf("Invalid order correctly rejected with error: %v", err)
		} else {
			assert.False(t, resp.Success, "Invalid order should be rejected")
		}

		// Verify system is still responsive after error
		validOrder := suite.testData.Orders[0]
		resp, err = suite.httpClient.PlaceOrder(baseURL, validOrder)
		require.NoError(t, err)
		suite.assertions.AssertOrderPlacementSuccess(resp, validOrder.Symbol)
	})

	suite.T().Run("HealthCheckAfterErrors", func(t *testing.T) {
		// Check system health after error conditions
		health, err := suite.httpClient.GetHealth(baseURL)
		require.NoError(t, err)
		assert.Equal(t, "healthy", health.Status, "System should remain healthy after errors")
		assert.NotEmpty(t, health.Services, "Should report service statuses")
	})
}

// TestEndToEndScenarios tests complete end-to-end trading scenarios
func (suite *TradingFlowTestSuite) TestEndToEndScenarios() {
	baseURL := suite.server.GetURL("")

	suite.T().Run("TradingSessionSimulation", func(t *testing.T) {
		// Simulate a complete trading session
		session := struct {
			users  []fixtures.TestUser
			orders []dto.PlaceOrderRequest
		}{
			users:  suite.testData.TestUsers[:3], // Use first 3 test users
			orders: make([]dto.PlaceOrderRequest, 0),
		}

		// Generate orders for each user
		for _, user := range session.users {
			userOrders := suite.generateUserOrders(user, 5) // 5 orders per user
			session.orders = append(session.orders, userOrders...)
		}

		// Execute trading session
		loadRunner := helpers.NewLoadTestRunner(suite.T(), baseURL)
		results, err := loadRunner.RunConcurrentOrders(session.orders, 10)
		require.NoError(t, err)

		// Verify session results
		assert.Greater(t, results.GetSuccessRate(), 0.8, "Trading session should have high success rate")
		assert.Greater(t, results.GetThroughput(), 15.0, "Should achieve good throughput during session")

		// Verify final metrics
		finalMetrics, err := suite.httpClient.GetMetrics(baseURL)
		require.NoError(t, err)
		assert.Greater(t, finalMetrics.OrderCount, int64(len(session.orders)*0.8),
			"Should have processed most orders")

		t.Logf("Trading session completed: %d orders, %.2f%% success rate, %.2f ops/s throughput",
			len(session.orders), results.GetSuccessRate()*100, results.GetThroughput())
	})
}

// generateUserOrders generates orders based on user trading strategy
func (suite *TradingFlowTestSuite) generateUserOrders(user fixtures.TestUser, count int) []dto.PlaceOrderRequest {
	orders := make([]dto.PlaceOrderRequest, count)
	symbols := suite.testData.Symbols

	for i := 0; i < count; i++ {
		symbol := symbols[i%len(symbols)]
		side := "buy"
		if i%2 == 1 {
			side = "sell"
		}

		// Adjust order parameters based on user strategy
		var quantity, price float64
		switch user.TradingStrategy {
		case "conservative":
			quantity = 1.0 + float64(i)*0.5
			price = 50000.0 + float64(i)*100.0
		case "aggressive":
			quantity = 10.0 + float64(i)*5.0
			price = 45000.0 + float64(i)*500.0
		case "hft":
			quantity = 0.1 + float64(i)*0.1
			price = 50000.0 + float64(i)*50.0
		default:
			quantity = 5.0 + float64(i)*1.0
			price = 50000.0 + float64(i)*200.0
		}

		orders[i] = dto.PlaceOrderRequest{
			Symbol:   symbol,
			Side:     side,
			Type:     "limit",
			Quantity: quantity,
			Price:    price,
		}
	}

	return orders
}

// TestTradingFlowTestSuite runs the trading flow test suite
func TestTradingFlowTestSuite(t *testing.T) {
	suite.Run(t, new(TradingFlowTestSuite))
}