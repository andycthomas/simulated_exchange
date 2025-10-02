package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"simulated_exchange/pkg/shared"
	"simulated_exchange/services/market-simulator/internal/domain"
)

// SimulatorHandler handles simulator-related HTTP requests
type SimulatorHandler struct {
	simulatorService shared.SimulatorService
	priceService     *domain.PriceService
	logger           *slog.Logger
}

// NewSimulatorHandler creates a new simulator handler
func NewSimulatorHandler(
	simulatorService shared.SimulatorService,
	priceService *domain.PriceService,
	logger *slog.Logger,
) *SimulatorHandler {
	return &SimulatorHandler{
		simulatorService: simulatorService,
		priceService:     priceService,
		logger:           logger,
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

// VolatilityRequest represents the request body for injecting volatility
type VolatilityRequest struct {
	Pattern   string  `json:"pattern" binding:"required,oneof=spike decay oscillate random"`
	Intensity float64 `json:"intensity" binding:"required,min=0.1,max=1.0"`
}

// GetStatus handles GET /api/status
func (h *SimulatorHandler) GetStatus(c *gin.Context) {
	status, err := h.simulatorService.GetStatus(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get simulator status", "error", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "STATUS_ERROR",
				Message: "Failed to get simulator status",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    status,
	})
}

// InjectVolatility handles POST /api/volatility
func (h *SimulatorHandler) InjectVolatility(c *gin.Context) {
	var req VolatilityRequest

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

	err := h.simulatorService.InjectVolatility(c.Request.Context(), req.Pattern, req.Intensity)
	if err != nil {
		h.logger.Error("Failed to inject volatility", "error", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "VOLATILITY_INJECTION_FAILED",
				Message: "Failed to inject volatility",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"pattern":   req.Pattern,
			"intensity": req.Intensity,
			"message":   "Volatility injected successfully",
		},
	})

	h.logger.Info("Volatility injected",
		"pattern", req.Pattern,
		"intensity", req.Intensity,
	)
}

// GetMarketData handles GET /api/market/:symbol
func (h *SimulatorHandler) GetMarketData(c *gin.Context) {
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

	marketData, err := h.priceService.GetMarketData(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Warn("Failed to get market data", "symbol", symbol, "error", err)
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "MARKET_DATA_NOT_FOUND",
				Message: "Market data not found",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    marketData,
	})
}

// GetCurrentPrice handles GET /api/price/:symbol
func (h *SimulatorHandler) GetCurrentPrice(c *gin.Context) {
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

	price, err := h.priceService.GetCurrentPrice(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Warn("Failed to get current price", "symbol", symbol, "error", err)
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "PRICE_NOT_FOUND",
				Message: "Price not found",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"symbol": symbol,
			"price":  price,
		},
	})
}

// GetPriceHistory handles GET /api/history/:symbol
func (h *SimulatorHandler) GetPriceHistory(c *gin.Context) {
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

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	// For simplicity, get recent history
	// In production, this would parse 'from' and 'to' query parameters
	history, err := h.priceService.GetPriceHistory(c.Request.Context(), symbol,
		time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		h.logger.Warn("Failed to get price history", "symbol", symbol, "error", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "HISTORY_ERROR",
				Message: "Failed to get price history",
				Details: err.Error(),
			},
		})
		return
	}

	// Apply limit
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"symbol":  symbol,
			"history": history,
			"count":   len(history),
		},
	})
}

// GetAllSymbols handles GET /api/symbols
func (h *SimulatorHandler) GetAllSymbols(c *gin.Context) {
	symbols := h.priceService.GetAllSymbols()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"symbols": symbols,
			"count":   len(symbols),
		},
	})
}