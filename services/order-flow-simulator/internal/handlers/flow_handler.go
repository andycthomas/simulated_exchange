package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"simulated_exchange/services/order-flow-simulator/internal/domain"
)

// FlowHandler handles flow simulation related HTTP requests
type FlowHandler struct {
	flowSimulator *domain.FlowSimulator
	userSimulator *domain.UserSimulator
	logger        *slog.Logger
}

// NewFlowHandler creates a new flow handler
func NewFlowHandler(
	flowSimulator *domain.FlowSimulator,
	userSimulator *domain.UserSimulator,
	logger *slog.Logger,
) *FlowHandler {
	return &FlowHandler{
		flowSimulator: flowSimulator,
		userSimulator: userSimulator,
		logger:        logger,
	}
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

// SetOrderRateRequest represents the request for setting order rate
type SetOrderRateRequest struct {
	Symbol string  `json:"symbol" binding:"required"`
	Rate   float64 `json:"rate" binding:"required,min=0"`
}

// SetVolatilityRequest represents the request for setting volatility mode
type SetVolatilityRequest struct {
	Enabled bool `json:"enabled"`
}

// GetStatus handles GET /api/status
func (h *FlowHandler) GetStatus(c *gin.Context) {
	status := h.flowSimulator.GetStatus()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    status,
	})
}

// SetOrderRate handles POST /api/order-rate
func (h *FlowHandler) SetOrderRate(c *gin.Context) {
	var req SetOrderRateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid order rate request", "error", err)
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

	err := h.flowSimulator.SetOrderRate(req.Symbol, req.Rate)
	if err != nil {
		h.logger.Error("Failed to set order rate", "error", err, "symbol", req.Symbol, "rate", req.Rate)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "ORDER_RATE_SET_FAILED",
				Message: "Failed to set order rate",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"symbol":  req.Symbol,
			"rate":    req.Rate,
			"message": "Order rate updated successfully",
		},
	})

	h.logger.Info("Order rate updated",
		"symbol", req.Symbol,
		"rate", req.Rate,
	)
}

// SetVolatility handles POST /api/volatility
func (h *FlowHandler) SetVolatility(c *gin.Context) {
	var req SetVolatilityRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid volatility request", "error", err)
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

	h.flowSimulator.SetVolatilityMode(req.Enabled)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"volatility_enabled": req.Enabled,
			"message":           "Volatility mode updated successfully",
		},
	})

	h.logger.Info("Volatility mode updated", "enabled", req.Enabled)
}

// GetUserSessions handles GET /api/users
func (h *FlowHandler) GetUserSessions(c *gin.Context) {
	activeUserCount := h.userSimulator.GetActiveUserCount()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"active_users": activeUserCount,
		},
	})
}

// GetMarketState handles GET /api/market/:symbol
func (h *FlowHandler) GetMarketState(c *gin.Context) {
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

	marketState, exists := h.userSimulator.GetMarketState(symbol)
	if !exists {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "MARKET_STATE_NOT_FOUND",
				Message: "Market state not found for symbol",
				Details: symbol,
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    marketState,
	})
}

// GetSimulationMetrics handles GET /api/metrics
func (h *FlowHandler) GetSimulationMetrics(c *gin.Context) {
	// Get query parameters for time range
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	status := h.flowSimulator.GetStatus()

	// Create simplified metrics response
	metrics := map[string]interface{}{
		"current_status":      status.IsRunning,
		"orders_generated":    status.OrdersGenerated,
		"orders_submitted":    status.OrdersSubmitted,
		"orders_failed":       status.OrdersFailed,
		"active_users":        status.ActiveUsers,
		"symbol_statistics":   status.SymbolStats,
		"uptime_seconds":      status.LastUpdate.Sub(status.StartTime).Seconds(),
		"last_updated":        status.LastUpdate,
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    metrics,
	})
}

// GetSymbolStats handles GET /api/symbols/:symbol/stats
func (h *FlowHandler) GetSymbolStats(c *gin.Context) {
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

	status := h.flowSimulator.GetStatus()
	symbolStats, exists := status.SymbolStats[symbol]

	if !exists {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "SYMBOL_STATS_NOT_FOUND",
				Message: "Statistics not found for symbol",
				Details: symbol,
			},
		})
		return
	}

	// Get market state for additional context
	marketState, _ := h.userSimulator.GetMarketState(symbol)

	response := map[string]interface{}{
		"symbol":            symbolStats.Symbol,
		"orders_generated":  symbolStats.OrdersGenerated,
		"orders_submitted":  symbolStats.OrdersSubmitted,
		"current_order_rate": symbolStats.OrderRate,
		"last_order_time":   symbolStats.LastOrderTime,
	}

	if marketState != nil {
		response["market_state"] = marketState
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}