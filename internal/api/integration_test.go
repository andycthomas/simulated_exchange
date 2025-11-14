package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"simulated_exchange/internal/api/dto"
	"simulated_exchange/internal/api/handlers"
	"simulated_exchange/internal/demo"
)

// MockOrderService for integration testing
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) PlaceOrder(orderID, symbol string, side, orderType string, quantity, price float64) error {
	args := m.Called(orderID, symbol, side, orderType, quantity, price)
	return args.Error(0)
}

func (m *MockOrderService) GetOrder(orderID string) (handlers.Order, error) {
	args := m.Called(orderID)
	return args.Get(0).(handlers.Order), args.Error(1)
}

func (m *MockOrderService) CancelOrder(orderID string) error {
	args := m.Called(orderID)
	return args.Error(0)
}

func (m *MockOrderService) GetOrderBook(symbol string) (handlers.OrderBook, error) {
	args := m.Called(symbol)
	return args.Get(0).(handlers.OrderBook), args.Error(1)
}

// MockMetricsService for integration testing
type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) GetRealTimeMetrics() handlers.MetricsSnapshot {
	args := m.Called()
	return args.Get(0).(handlers.MetricsSnapshot)
}

func (m *MockMetricsService) GetPerformanceAnalysis() handlers.PerformanceAnalysis {
	args := m.Called()
	return args.Get(0).(handlers.PerformanceAnalysis)
}

func (m *MockMetricsService) IsHealthy() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockDemoController for integration testing
type MockDemoController struct {
	mock.Mock
}

func (m *MockDemoController) StartLoadTest(ctx context.Context, scenario demo.LoadTestScenario) error {
	args := m.Called(ctx, scenario)
	return args.Error(0)
}

func (m *MockDemoController) StopLoadTest(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDemoController) GetLoadTestStatus(ctx context.Context) (*demo.LoadTestStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*demo.LoadTestStatus), args.Error(1)
}

func (m *MockDemoController) TriggerChaosTest(ctx context.Context, scenario demo.ChaosTestScenario) error {
	args := m.Called(ctx, scenario)
	return args.Error(0)
}

func (m *MockDemoController) StopChaosTest(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDemoController) GetChaosTestStatus(ctx context.Context) (*demo.ChaosTestStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*demo.ChaosTestStatus), args.Error(1)
}

func (m *MockDemoController) ResetSystem(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDemoController) GetSystemStatus(ctx context.Context) (*demo.DemoSystemStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*demo.DemoSystemStatus), args.Error(1)
}

func TestAPI_IntegrationTests(t *testing.T) {
	// Setup mock services
	mockOrderService := new(MockOrderService)
	mockMetricsService := new(MockMetricsService)
	mockDemoController := new(MockDemoController)

	// Create dependency container
	deps := NewDependencyContainer(mockOrderService, mockMetricsService, mockDemoController)

	// Create server with test configuration
	config := &Config{
		Port:         "8080",
		Environment:  "test",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	server := NewServer(deps, config)
	router := server.GetRouter()

	t.Run("POST /api/orders - Place Order Integration", func(t *testing.T) {
		// Setup mock expectation
		mockOrderService.On("PlaceOrder", mock.AnythingOfType("string"), "AAPL", "buy", "limit", 100.0, 150.0).Return(nil)

		// Prepare request
		orderRequest := dto.PlaceOrderRequest{
			Symbol:   "AAPL",
			Side:     "buy",
			Type:     "limit",
			Quantity: 100,
			Price:    150.0,
		}

		body, err := json.Marshal(orderRequest)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusCreated, w.Code)

		var response dto.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)

		// Verify response data structure
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "PLACED", data["status"])
		assert.NotEmpty(t, data["order_id"])

		mockOrderService.AssertExpectations(t)
	})

	t.Run("GET /api/orders/:id - Get Order Integration", func(t *testing.T) {
		// Setup mock expectation
		expectedOrder := handlers.Order{
			ID:       "test-order-123",
			Symbol:   "AAPL",
			Side:     "buy",
			Type:     "limit",
			Quantity: 100,
			Price:    150.0,
			Status:   "active",
		}
		mockOrderService.On("GetOrder", "test-order-123").Return(expectedOrder, nil)

		// Execute request
		req, err := http.NewRequest("GET", "/api/orders/test-order-123", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)

		// Verify response data structure
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "test-order-123", data["id"])
		assert.Equal(t, "AAPL", data["symbol"])
		assert.Equal(t, "buy", data["side"])
		assert.Equal(t, float64(100), data["quantity"])
		assert.Equal(t, float64(150), data["price"])

		mockOrderService.AssertExpectations(t)
	})

	t.Run("DELETE /api/orders/:id - Cancel Order Integration", func(t *testing.T) {
		// Setup mock expectation
		mockOrderService.On("CancelOrder", "test-order-123").Return(nil)

		// Execute request
		req, err := http.NewRequest("DELETE", "/api/orders/test-order-123", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)

		// Verify response data structure
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "CANCELLED", data["status"])
		assert.Equal(t, "test-order-123", data["order_id"])

		mockOrderService.AssertExpectations(t)
	})

	t.Run("GET /api/metrics - Get Metrics Integration", func(t *testing.T) {
		// Setup mock expectations
		expectedMetrics := handlers.MetricsSnapshot{
			OrderCount:   100,
			TradeCount:   50,
			TotalVolume:  10000.0,
			AvgLatency:   "5ms",
			OrdersPerSec: 10.5,
			TradesPerSec: 5.2,
			SymbolMetrics: map[string]handlers.SymbolMetrics{
				"AAPL": {
					OrderCount: 50,
					TradeCount: 25,
					Volume:     5000.0,
					AvgPrice:   150.0,
				},
			},
		}

		expectedAnalysis := handlers.PerformanceAnalysis{
			Timestamp:      time.Now().Format(time.RFC3339),
			TrendDirection: "upward",
			Bottlenecks: []handlers.Bottleneck{
				{
					Type:        "latency",
					Severity:    0.6,
					Description: "Moderate latency increase detected",
				},
			},
			Recommendations: []string{"Consider increasing processing capacity"},
		}

		mockMetricsService.On("GetRealTimeMetrics").Return(expectedMetrics)
		mockMetricsService.On("GetPerformanceAnalysis").Return(expectedAnalysis)

		// Execute request
		req, err := http.NewRequest("GET", "/api/metrics", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)

		// Verify response data structure
		data := response.Data.(map[string]interface{})
		assert.Equal(t, float64(100), data["order_count"])
		assert.Equal(t, float64(50), data["trade_count"])
		assert.Equal(t, float64(10000), data["total_volume"])
		assert.NotNil(t, data["symbol_metrics"])
		assert.NotNil(t, data["analysis"])

		mockMetricsService.AssertExpectations(t)
	})

	t.Run("GET /api/health - Health Check Integration", func(t *testing.T) {
		// Setup mock expectation
		mockMetricsService.On("IsHealthy").Return(true)

		// Execute request
		req, err := http.NewRequest("GET", "/api/health", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Data)

		// Verify response data structure
		data := response.Data.(map[string]interface{})
		assert.Equal(t, "healthy", data["status"])
		assert.NotNil(t, data["timestamp"])
		assert.NotNil(t, data["services"])
		assert.Equal(t, "1.0.0", data["version"])

		mockMetricsService.AssertExpectations(t)
	})

	t.Run("GET / - Root Endpoint Integration", func(t *testing.T) {
		// Execute request
		req, err := http.NewRequest("GET", "/", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "simulated-exchange-api", response["service"])
		assert.Equal(t, "1.0.0", response["version"])
		assert.Equal(t, "running", response["status"])
	})
}

func TestAPI_ErrorHandling_Integration(t *testing.T) {
	// Setup mock services
	mockOrderService := new(MockOrderService)
	mockMetricsService := new(MockMetricsService)
	mockDemoController := new(MockDemoController)

	// Create dependency container
	deps := NewDependencyContainer(mockOrderService, mockMetricsService, mockDemoController)

	// Create server
	config := DefaultConfig()
	config.Environment = "test"
	server := NewServer(deps, config)
	router := server.GetRouter()

	t.Run("Invalid JSON Content Type", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/orders", bytes.NewBuffer([]byte(`{"symbol":"AAPL"}`)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, w.Code)

		var response dto.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "UNSUPPORTED_MEDIA_TYPE", response.Error.Code)
	})

	t.Run("Invalid JSON Format", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/api/orders", bytes.NewBuffer([]byte(`{invalid json}`)))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dto.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		orderRequest := map[string]interface{}{
			"symbol": "AAPL",
			// Missing required fields
		}

		body, err := json.Marshal(orderRequest)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response dto.APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
	})

	t.Run("404 Not Found", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/nonexistent", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAPI_CORS_Integration(t *testing.T) {
	// Setup mock services
	mockOrderService := new(MockOrderService)
	mockMetricsService := new(MockMetricsService)
	mockDemoController := new(MockDemoController)

	// Create dependency container
	deps := NewDependencyContainer(mockOrderService, mockMetricsService, mockDemoController)

	// Create server
	config := DefaultConfig()
	config.Environment = "test"
	server := NewServer(deps, config)
	router := server.GetRouter()

	t.Run("CORS Preflight Request", func(t *testing.T) {
		req, err := http.NewRequest("OPTIONS", "/api/orders", nil)
		assert.NoError(t, err)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	})

	t.Run("CORS Actual Request", func(t *testing.T) {
		mockMetricsService.On("IsHealthy").Return(true)

		req, err := http.NewRequest("GET", "/api/health", nil)
		assert.NoError(t, err)
		req.Header.Set("Origin", "http://localhost:3000")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

		mockMetricsService.AssertExpectations(t)
	})
}

func TestAPI_SecurityHeaders_Integration(t *testing.T) {
	// Setup mock services
	mockOrderService := new(MockOrderService)
	mockMetricsService := new(MockMetricsService)
	mockDemoController := new(MockDemoController)

	// Create dependency container
	deps := NewDependencyContainer(mockOrderService, mockMetricsService, mockDemoController)

	// Create server
	config := DefaultConfig()
	config.Environment = "test"
	server := NewServer(deps, config)
	router := server.GetRouter()

	t.Run("Security Headers Present", func(t *testing.T) {
		mockMetricsService.On("IsHealthy").Return(true)

		req, err := http.NewRequest("GET", "/api/health", nil)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")

		mockMetricsService.AssertExpectations(t)
	})
}