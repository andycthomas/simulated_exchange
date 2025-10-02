package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"simulated_exchange/pkg/shared"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	tradingService shared.TradingService
	logger         *slog.Logger
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(tradingService shared.TradingService, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{
		tradingService: tradingService,
		logger:         logger,
	}
}

// PlaceOrderRequest represents the request body for placing an order
type PlaceOrderRequest struct {
	UserID   string  `json:"user_id" binding:"required"`
	Symbol   string  `json:"symbol" binding:"required,min=1,max=10"`
	Side     string  `json:"side" binding:"required,oneof=BUY SELL"`
	Type     string  `json:"type" binding:"required,oneof=MARKET LIMIT"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Price    float64 `json:"price" binding:"omitempty,gt=0"`
}

// PlaceOrderResponse represents the response after placing an order
type PlaceOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// OrderResponse represents order information in API responses
type OrderResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Type      string  `json:"type"`
	Quantity  float64 `json:"quantity"`
	Price     float64 `json:"price"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// OrderBookResponse represents order book information
type OrderBookResponse struct {
	Symbol    string              `json:"symbol"`
	Bids      []OrderBookEntry    `json:"bids"`
	Asks      []OrderBookEntry    `json:"asks"`
	UpdatedAt string              `json:"updated_at"`
}

// OrderBookEntry represents a single order book entry
type OrderBookEntry struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	Orders   int     `json:"orders"`
}

// APIResponse provides a consistent structure for all API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents error information in API responses
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PlaceOrder handles POST /api/orders
func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req PlaceOrderRequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid place order request", "error", err)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request format",
				Details: err.Error(),
			},
		})
		return
	}

	// Additional validation for limit orders
	if req.Type == "LIMIT" && req.Price <= 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Price is required for limit orders",
			},
		})
		return
	}

	// Convert to domain model
	order := &shared.Order{
		UserID:   req.UserID,
		Symbol:   req.Symbol,
		Side:     shared.OrderSide(req.Side),
		Type:     shared.OrderType(req.Type),
		Quantity: req.Quantity,
		Price:    req.Price,
	}

	// Call service layer
	placedOrder, err := h.tradingService.PlaceOrder(c.Request.Context(), order)
	if err != nil {
		h.logger.Error("Failed to place order", "error", err, "user_id", req.UserID)

		// Handle different error types
		var apiError *APIError
		switch e := err.(type) {
		case *shared.ValidationError:
			apiError = &APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
				Details: e.Error(),
			}
		case *shared.BusinessError:
			apiError = &APIError{
				Code:    e.Code,
				Message: e.Message,
				Details: e.Details,
			}
		default:
			apiError = &APIError{
				Code:    "ORDER_PLACEMENT_FAILED",
				Message: "Failed to place order",
				Details: err.Error(),
			}
		}

		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   apiError,
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data: PlaceOrderResponse{
			OrderID: placedOrder.ID,
			Status:  string(placedOrder.Status),
			Message: "Order placed successfully",
		},
	})

	h.logger.Info("Order placed successfully",
		"order_id", placedOrder.ID,
		"user_id", req.UserID,
		"symbol", req.Symbol,
	)
}

// GetOrder handles GET /api/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	if orderID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "MISSING_ORDER_ID",
				Message: "Order ID is required",
			},
		})
		return
	}

	// Call service layer
	order, err := h.tradingService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		h.logger.Warn("Failed to get order", "error", err, "order_id", orderID)

		status := http.StatusInternalServerError
		if err == shared.ErrOrderNotFound {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "ORDER_NOT_FOUND",
				Message: "Order not found",
				Details: err.Error(),
			},
		})
		return
	}

	// Convert to response model
	orderResponse := OrderResponse{
		ID:        order.ID,
		UserID:    order.UserID,
		Symbol:    order.Symbol,
		Side:      string(order.Side),
		Type:      string(order.Type),
		Quantity:  order.Quantity,
		Price:     order.Price,
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: order.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    orderResponse,
	})
}

// GetOrders handles GET /api/orders
func (h *OrderHandler) GetOrders(c *gin.Context) {
	// Optional query parameters
	limit := 50 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	// Call service layer to get recent orders
	orders, err := h.tradingService.GetRecentOrders(c.Request.Context(), limit)
	if err != nil {
		h.logger.Warn("Failed to get orders", "error", err)

		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "ORDER_RETRIEVAL_FAILED",
				Message: "Failed to retrieve orders",
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    orders,
	})
}

// CancelOrder handles DELETE /api/orders/:id
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	if orderID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "MISSING_ORDER_ID",
				Message: "Order ID is required",
			},
		})
		return
	}

	// Call service layer
	err := h.tradingService.CancelOrder(c.Request.Context(), orderID)
	if err != nil {
		h.logger.Warn("Failed to cancel order", "error", err, "order_id", orderID)

		status := http.StatusInternalServerError
		if err == shared.ErrOrderNotFound {
			status = http.StatusNotFound
		}

		c.JSON(status, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "ORDER_CANCELLATION_FAILED",
				Message: "Failed to cancel order",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"order_id": orderID,
			"status":   "CANCELLED",
			"message":  "Order cancelled successfully",
		},
	})

	h.logger.Info("Order cancelled successfully", "order_id", orderID)
}

// GetOrderBook handles GET /api/orderbook/:symbol
func (h *OrderHandler) GetOrderBook(c *gin.Context) {
	symbol := c.Param("symbol")

	if symbol == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "MISSING_SYMBOL",
				Message: "Symbol is required",
			},
		})
		return
	}

	// Call service layer
	orderBook, err := h.tradingService.GetOrderBook(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Error("Failed to get order book", "error", err, "symbol", symbol)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "ORDER_BOOK_ERROR",
				Message: "Failed to get order book",
				Details: err.Error(),
			},
		})
		return
	}

	// Convert to response model
	response := OrderBookResponse{
		Symbol:    orderBook.Symbol,
		UpdatedAt: orderBook.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Bids:      h.aggregateOrderBook(orderBook.Bids),
		Asks:      h.aggregateOrderBook(orderBook.Asks),
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetUserOrders handles GET /api/users/:user_id/orders
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID := c.Param("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "MISSING_USER_ID",
				Message: "User ID is required",
			},
		})
		return
	}

	// Call service layer
	orders, err := h.tradingService.GetUserOrders(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user orders", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "USER_ORDERS_ERROR",
				Message: "Failed to get user orders",
				Details: err.Error(),
			},
		})
		return
	}

	// Convert to response model
	var orderResponses []OrderResponse
	for _, order := range orders {
		orderResponses = append(orderResponses, OrderResponse{
			ID:        order.ID,
			UserID:    order.UserID,
			Symbol:    order.Symbol,
			Side:      string(order.Side),
			Type:      string(order.Type),
			Quantity:  order.Quantity,
			Price:     order.Price,
			Status:    string(order.Status),
			CreatedAt: order.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: order.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    orderResponses,
	})
}

// aggregateOrderBook aggregates orders by price level
func (h *OrderHandler) aggregateOrderBook(orders []shared.Order) []OrderBookEntry {
	priceMap := make(map[float64]*OrderBookEntry)

	for _, order := range orders {
		if entry, exists := priceMap[order.Price]; exists {
			entry.Quantity += order.Quantity
			entry.Orders++
		} else {
			priceMap[order.Price] = &OrderBookEntry{
				Price:    order.Price,
				Quantity: order.Quantity,
				Orders:   1,
			}
		}
	}

	var entries []OrderBookEntry
	for _, entry := range priceMap {
		entries = append(entries, *entry)
	}

	return entries
}