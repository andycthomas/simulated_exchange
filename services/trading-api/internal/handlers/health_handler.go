package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"simulated_exchange/pkg/cache"
	"simulated_exchange/pkg/database"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db     *database.PostgresDB
	cache  *cache.RedisClient
	logger *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.PostgresDB, cache *cache.RedisClient, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		cache:  cache,
		logger: logger,
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
		Service:   "trading-api",
		Version:   "1.0.0",
		Checks:    make(map[string]HealthCheck),
	}

	// Check database connectivity
	dbStatus := h.checkDatabase()
	response.Checks["database"] = dbStatus
	if dbStatus.Status != "healthy" {
		response.Status = "unhealthy"
	}

	// Check Redis connectivity
	cacheStatus := h.checkCache()
	response.Checks["cache"] = cacheStatus
	if cacheStatus.Status != "healthy" {
		response.Status = "unhealthy"
	}

	// Add response time
	response.Checks["response_time"] = HealthCheck{
		Status:  "healthy",
		Message: time.Since(startTime).String(),
	}

	// Always return 200 OK for dashboard connectivity, but include status in response
	// The dashboard can check the actual health status from the response body
	if response.Status == "unhealthy" {
		h.logger.Warn("Health check failed", "checks", response.Checks)
	}

	c.JSON(http.StatusOK, response)
}

// GetReadiness handles GET /ready
func (h *HealthHandler) GetReadiness(c *gin.Context) {
	// Check if service is ready to handle requests
	dbStatus := h.checkDatabase()
	cacheStatus := h.checkCache()

	ready := dbStatus.Status == "healthy" && cacheStatus.Status == "healthy"

	response := map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now().Format(time.RFC3339),
		"checks": map[string]HealthCheck{
			"database": dbStatus,
			"cache":    cacheStatus,
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
	// Simple liveness check - service is running
	response := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "trading-api",
	}

	c.JSON(http.StatusOK, response)
}

// checkDatabase checks database connectivity
func (h *HealthHandler) checkDatabase() HealthCheck {
	if h.db == nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "database connection not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: err.Error(),
		}
	}

	return HealthCheck{
		Status: "healthy",
	}
}

// checkCache checks Redis connectivity
func (h *HealthHandler) checkCache() HealthCheck {
	if h.cache == nil {
		return HealthCheck{
			Status:  "unhealthy",
			Message: "cache connection not initialized",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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