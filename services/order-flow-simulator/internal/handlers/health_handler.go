package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"simulated_exchange/pkg/cache"
	"simulated_exchange/services/order-flow-simulator/internal/domain"
)

// HealthHandler handles health check requests for order flow simulator
type HealthHandler struct {
	cache            *cache.RedisClient
	tradingAPIClient *domain.TradingAPIClient
	logger           *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cache *cache.RedisClient, tradingAPIClient *domain.TradingAPIClient, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		cache:            cache,
		tradingAPIClient: tradingAPIClient,
		logger:           logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// GetHealth handles GET /health
func (h *HealthHandler) GetHealth(c *gin.Context) {
	startTime := time.Now()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Service:   "order-flow-simulator",
		Version:   "1.0.0",
		Checks:    make(map[string]HealthCheck),
	}

	// Check Redis connectivity
	cacheStatus := h.checkCache()
	response.Checks["cache"] = cacheStatus
	if cacheStatus.Status != "healthy" {
		response.Status = "unhealthy"
	}

	// Check Trading API connectivity
	tradingAPIStatus := h.checkTradingAPI()
	response.Checks["trading_api"] = tradingAPIStatus
	if tradingAPIStatus.Status != "healthy" {
		response.Status = "degraded" // Service can still function without trading API
	}

	// Add response time
	response.Checks["response_time"] = HealthCheck{
		Status:  "healthy",
		Message: time.Since(startTime).String(),
	}

	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
		h.logger.Warn("Health check failed", "checks", response.Checks)
	} else if response.Status == "degraded" {
		h.logger.Warn("Health check degraded", "checks", response.Checks)
	}

	c.JSON(statusCode, response)
}

// GetReadiness handles GET /ready
func (h *HealthHandler) GetReadiness(c *gin.Context) {
	cacheStatus := h.checkCache()
	tradingAPIStatus := h.checkTradingAPI()

	ready := cacheStatus.Status == "healthy" && tradingAPIStatus.Status == "healthy"

	response := map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now().Format(time.RFC3339),
		"checks": map[string]HealthCheck{
			"cache":       cacheStatus,
			"trading_api": tradingAPIStatus,
		},
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// GetLiveness handles GET /live
func (h *HealthHandler) GetLiveness(c *gin.Context) {
	response := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "order-flow-simulator",
	}

	c.JSON(http.StatusOK, response)
}

// checkCache checks Redis connectivity
func (h *HealthHandler) checkCache() HealthCheck {
	if h.cache == nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "cache connection not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.cache.Ping(ctx); err != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	return HealthCheck{
		Status: "healthy",
	}
}

// checkTradingAPI checks trading API connectivity
func (h *HealthHandler) checkTradingAPI() HealthCheck {
	if h.tradingAPIClient == nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "trading API client not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.tradingAPIClient.HealthCheck(ctx); err != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	return HealthCheck{
		Status: "healthy",
	}
}