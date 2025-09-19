package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"simulated_exchange/internal/api/dto"
)

// OrderHandlerImpl implements the OrderHandler interface
type OrderHandlerImpl struct {
	orderService OrderService
}

// NewOrderHandler creates a new order handler with dependency injection
func NewOrderHandler(orderService OrderService) OrderHandler {
	return &OrderHandlerImpl{
		orderService: orderService,
	}
}

// PlaceOrder handles POST /api/orders
func (h *OrderHandlerImpl) PlaceOrder(c *gin.Context) {
	var req dto.PlaceOrderRequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request format",
				Details: err.Error(),
			},
		})
		return
	}

	// Additional business logic validation
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
				Details: err.Error(),
			},
		})
		return
	}

	// Generate unique order ID
	orderID := uuid.New().String()

	// Call service layer
	err := h.orderService.PlaceOrder(
		orderID,
		req.Symbol,
		req.Side,
		req.Type,
		req.Quantity,
		req.Price,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "ORDER_PLACEMENT_FAILED",
				Message: "Failed to place order",
				Details: err.Error(),
			},
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Data: dto.PlaceOrderResponse{
			OrderID: orderID,
			Status:  "PLACED",
			Message: "Order placed successfully",
		},
	})
}

// GetOrder handles GET /api/orders/:id
func (h *OrderHandlerImpl) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	if orderID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "MISSING_ORDER_ID",
				Message: "Order ID is required",
			},
		})
		return
	}

	// Call service layer
	order, err := h.orderService.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "ORDER_NOT_FOUND",
				Message: "Order not found",
				Details: err.Error(),
			},
		})
		return
	}

	// Convert domain model to DTO
	orderResponse := dto.OrderResponse{
		ID:       order.ID,
		Symbol:   order.Symbol,
		Side:     order.Side,
		Type:     order.Type,
		Quantity: order.Quantity,
		Price:    order.Price,
		Status:   order.Status,
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    orderResponse,
	})
}

// CancelOrder handles DELETE /api/orders/:id
func (h *OrderHandlerImpl) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	if orderID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "MISSING_ORDER_ID",
				Message: "Order ID is required",
			},
		})
		return
	}

	// Call service layer
	err := h.orderService.CancelOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "ORDER_CANCELLATION_FAILED",
				Message: "Failed to cancel order",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: dto.CancelOrderResponse{
			OrderID: orderID,
			Status:  "CANCELLED",
			Message: "Order cancelled successfully",
		},
	})
}