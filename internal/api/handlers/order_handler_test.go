package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"simulated_exchange/internal/api/dto"
)

// MockOrderService implements OrderService for testing
type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) PlaceOrder(orderID, symbol string, side, orderType string, quantity, price float64) error {
	args := m.Called(orderID, symbol, side, orderType, quantity, price)
	return args.Error(0)
}

func (m *MockOrderService) GetOrder(orderID string) (Order, error) {
	args := m.Called(orderID)
	return args.Get(0).(Order), args.Error(1)
}

func (m *MockOrderService) CancelOrder(orderID string) error {
	args := m.Called(orderID)
	return args.Error(0)
}

func (m *MockOrderService) GetOrderBook(symbol string) (OrderBook, error) {
	args := m.Called(symbol)
	return args.Get(0).(OrderBook), args.Error(1)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestOrderHandlerImpl_PlaceOrder(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockOrderService)
		expectedStatus int
		expectedError  *string
		expectedData   bool
	}{
		{
			name: "successful order placement",
			requestBody: dto.PlaceOrderRequest{
				Symbol:   "AAPL",
				Side:     "buy",
				Type:     "limit",
				Quantity: 100,
				Price:    150.0,
			},
			mockSetup: func(m *MockOrderService) {
				m.On("PlaceOrder", mock.AnythingOfType("string"), "AAPL", "buy", "limit", 100.0, 150.0).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedData:   true,
		},
		{
			name: "invalid request - missing symbol",
			requestBody: dto.PlaceOrderRequest{
				Side:     "buy",
				Type:     "limit",
				Quantity: 100,
				Price:    150.0,
			},
			mockSetup: func(m *MockOrderService) {
				// No mock setup needed as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("INVALID_REQUEST"),
		},
		{
			name: "invalid request - negative quantity",
			requestBody: dto.PlaceOrderRequest{
				Symbol:   "AAPL",
				Side:     "buy",
				Type:     "limit",
				Quantity: -100,
				Price:    150.0,
			},
			mockSetup: func(m *MockOrderService) {
				// No mock setup needed as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("INVALID_REQUEST"),
		},
		{
			name: "invalid request - invalid side",
			requestBody: dto.PlaceOrderRequest{
				Symbol:   "AAPL",
				Side:     "invalid",
				Type:     "limit",
				Quantity: 100,
				Price:    150.0,
			},
			mockSetup: func(m *MockOrderService) {
				// No mock setup needed as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("INVALID_REQUEST"),
		},
		{
			name: "service error",
			requestBody: dto.PlaceOrderRequest{
				Symbol:   "AAPL",
				Side:     "buy",
				Type:     "limit",
				Quantity: 100,
				Price:    150.0,
			},
			mockSetup: func(m *MockOrderService) {
				m.On("PlaceOrder", mock.AnythingOfType("string"), "AAPL", "buy", "limit", 100.0, 150.0).Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  stringPtr("ORDER_PLACEMENT_FAILED"),
		},
		{
			name:        "invalid JSON",
			requestBody: "invalid json",
			mockSetup: func(m *MockOrderService) {
				// No mock setup needed as JSON parsing should fail
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  stringPtr("INVALID_REQUEST"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockOrderService)
			tt.mockSetup(mockService)

			handler := NewOrderHandler(mockService)
			router := setupTestRouter()
			router.POST("/api/orders", handler.PlaceOrder)

			// Prepare request
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Execute
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response dto.APIResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != nil {
				assert.False(t, response.Success)
				assert.NotNil(t, response.Error)
				assert.Equal(t, *tt.expectedError, response.Error.Code)
			} else {
				assert.True(t, response.Success)
				assert.Nil(t, response.Error)
			}

			if tt.expectedData {
				assert.NotNil(t, response.Data)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestOrderHandlerImpl_GetOrder(t *testing.T) {
	tests := []struct {
		name           string
		orderID        string
		mockSetup      func(*MockOrderService)
		expectedStatus int
		expectedError  *string
		expectedData   bool
		skipJSONCheck  bool
	}{
		{
			name:    "successful order retrieval",
			orderID: "order-123",
			mockSetup: func(m *MockOrderService) {
				order := Order{
					ID:       "order-123",
					Symbol:   "AAPL",
					Side:     "buy",
					Type:     "limit",
					Quantity: 100,
					Price:    150.0,
					Status:   "active",
				}
				m.On("GetOrder", "order-123").Return(order, nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:    "order not found",
			orderID: "nonexistent",
			mockSetup: func(m *MockOrderService) {
				m.On("GetOrder", "nonexistent").Return(Order{}, errors.New("order not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("ORDER_NOT_FOUND"),
		},
		{
			name:    "empty order ID returns 404 from router",
			orderID: "",
			mockSetup: func(m *MockOrderService) {
				// No mock setup needed as this will result in 404 from router
			},
			expectedStatus: http.StatusNotFound,
			skipJSONCheck:  true, // Router 404 doesn't return JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockOrderService)
			tt.mockSetup(mockService)

			handler := NewOrderHandler(mockService)
			router := setupTestRouter()
			router.GET("/api/orders/:id", handler.GetOrder)

			// Execute
			url := "/api/orders/" + tt.orderID
			req, err := http.NewRequest("GET", url, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.skipJSONCheck {
				var response dto.APIResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				if tt.expectedError != nil {
					assert.False(t, response.Success)
					assert.NotNil(t, response.Error)
					assert.Equal(t, *tt.expectedError, response.Error.Code)
				} else {
					assert.True(t, response.Success)
					assert.Nil(t, response.Error)
				}

				if tt.expectedData {
					assert.NotNil(t, response.Data)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestOrderHandlerImpl_CancelOrder(t *testing.T) {
	tests := []struct {
		name           string
		orderID        string
		mockSetup      func(*MockOrderService)
		expectedStatus int
		expectedError  *string
		expectedData   bool
		skipJSONCheck  bool
	}{
		{
			name:    "successful order cancellation",
			orderID: "order-123",
			mockSetup: func(m *MockOrderService) {
				m.On("CancelOrder", "order-123").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedData:   true,
		},
		{
			name:    "order not found for cancellation",
			orderID: "nonexistent",
			mockSetup: func(m *MockOrderService) {
				m.On("CancelOrder", "nonexistent").Return(errors.New("order not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  stringPtr("ORDER_CANCELLATION_FAILED"),
		},
		{
			name:    "empty order ID returns 404 from router",
			orderID: "",
			mockSetup: func(m *MockOrderService) {
				// No mock setup needed as this will result in 404 from router
			},
			expectedStatus: http.StatusNotFound,
			skipJSONCheck:  true, // Router 404 doesn't return JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockOrderService)
			tt.mockSetup(mockService)

			handler := NewOrderHandler(mockService)
			router := setupTestRouter()
			router.DELETE("/api/orders/:id", handler.CancelOrder)

			// Execute
			url := "/api/orders/" + tt.orderID
			req, err := http.NewRequest("DELETE", url, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.skipJSONCheck {
				var response dto.APIResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				if tt.expectedError != nil {
					assert.False(t, response.Success)
					assert.NotNil(t, response.Error)
					assert.Equal(t, *tt.expectedError, response.Error.Code)
				} else {
					assert.True(t, response.Success)
					assert.Nil(t, response.Error)
				}

				if tt.expectedData {
					assert.NotNil(t, response.Data)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}