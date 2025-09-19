package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"simulated_exchange/internal/api/dto"
)

// MetricsHandlerImpl implements the MetricsHandler interface
type MetricsHandlerImpl struct {
	metricsService MetricsService
}

// NewMetricsHandler creates a new metrics handler with dependency injection
func NewMetricsHandler(metricsService MetricsService) MetricsHandler {
	return &MetricsHandlerImpl{
		metricsService: metricsService,
	}
}

// GetMetrics handles GET /api/metrics
func (h *MetricsHandlerImpl) GetMetrics(c *gin.Context) {
	// Get real-time metrics from service
	metrics := h.metricsService.GetRealTimeMetrics()

	// Get performance analysis
	analysis := h.metricsService.GetPerformanceAnalysis()

	// Convert domain models to DTOs
	symbolMetricsDTO := make(map[string]dto.SymbolMetricsDTO)
	for symbol, symbolMetrics := range metrics.SymbolMetrics {
		symbolMetricsDTO[symbol] = dto.SymbolMetricsDTO{
			OrderCount: symbolMetrics.OrderCount,
			TradeCount: symbolMetrics.TradeCount,
			Volume:     symbolMetrics.Volume,
			AvgPrice:   symbolMetrics.AvgPrice,
		}
	}

	// Convert bottlenecks
	var bottlenecksDTO []dto.BottleneckDTO
	for _, bottleneck := range analysis.Bottlenecks {
		bottlenecksDTO = append(bottlenecksDTO, dto.BottleneckDTO{
			Type:        bottleneck.Type,
			Severity:    bottleneck.Severity,
			Description: bottleneck.Description,
		})
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, analysis.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	response := dto.MetricsResponse{
		OrderCount:    metrics.OrderCount,
		TradeCount:    metrics.TradeCount,
		TotalVolume:   metrics.TotalVolume,
		AvgLatency:    metrics.AvgLatency,
		OrdersPerSec:  metrics.OrdersPerSec,
		TradesPerSec:  metrics.TradesPerSec,
		SymbolMetrics: symbolMetricsDTO,
		Analysis: &dto.PerformanceAnalysisDTO{
			Timestamp:       timestamp,
			TrendDirection:  analysis.TrendDirection,
			Bottlenecks:     bottlenecksDTO,
			Recommendations: analysis.Recommendations,
		},
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetHealth handles GET /api/health
func (h *MetricsHandlerImpl) GetHealth(c *gin.Context) {
	// Check service health
	isHealthy := h.metricsService.IsHealthy()

	// Prepare service status map
	services := map[string]string{
		"metrics_service": "healthy",
		"trading_engine":  "healthy",
		"database":        "healthy",
	}

	// Update service status based on health checks
	if !isHealthy {
		services["metrics_service"] = "unhealthy"
	}

	// Determine overall status
	status := "healthy"
	if !isHealthy {
		status = "unhealthy"
	}

	response := dto.HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  services,
		Version:   "1.0.0",
	}

	// Return appropriate HTTP status code
	httpStatus := http.StatusOK
	if !isHealthy {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, dto.APIResponse{
		Success: isHealthy,
		Data:    response,
	})
}