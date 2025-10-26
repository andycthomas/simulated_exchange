package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"simulated_exchange/internal/api/dto"
	"simulated_exchange/internal/demo"
)

// DemoHandler defines HTTP endpoints for demo operations
type DemoHandler interface {
	StartLoadTest(c *gin.Context)
	GetLoadTestStatus(c *gin.Context)
	GetLoadTestResults(c *gin.Context)
	StopLoadTest(c *gin.Context)
	StartChaosTest(c *gin.Context)
	GetChaosTestStatus(c *gin.Context)
	GetChaosTestResults(c *gin.Context)
	StopChaosTest(c *gin.Context)
	ResetSystem(c *gin.Context)
	GetSystemStatus(c *gin.Context)
}

// DemoHandlerImpl implements the DemoHandler interface
type DemoHandlerImpl struct {
	demoController DemoController
}

// DemoController interface for demo operations
type DemoController interface {
	StartLoadTest(ctx context.Context, scenario demo.LoadTestScenario) error
	StopLoadTest(ctx context.Context) error
	GetLoadTestStatus(ctx context.Context) (*demo.LoadTestStatus, error)
	TriggerChaosTest(ctx context.Context, scenario demo.ChaosTestScenario) error
	StopChaosTest(ctx context.Context) error
	GetChaosTestStatus(ctx context.Context) (*demo.ChaosTestStatus, error)
	ResetSystem(ctx context.Context) error
	GetSystemStatus(ctx context.Context) (*demo.DemoSystemStatus, error)
}

// NewDemoHandler creates a new demo handler
func NewDemoHandler(demoController DemoController) DemoHandler {
	return &DemoHandlerImpl{
		demoController: demoController,
	}
}

// LoadTestRequest represents the request for starting a load test
type LoadTestRequest struct {
	Intensity       string   `json:"intensity"`
	Duration        int      `json:"duration"` // seconds
	OrdersPerSecond int      `json:"orders_per_second"`
	ConcurrentUsers int      `json:"concurrent_users"`
	Symbols         []string `json:"symbols"`
}

// StartLoadTest handles POST /demo/load-test
func (h *DemoHandlerImpl) StartLoadTest(c *gin.Context) {
	var req LoadTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	// Set defaults
	if req.Duration == 0 {
		req.Duration = 60
	}
	if req.OrdersPerSecond == 0 {
		req.OrdersPerSecond = 25
	}
	if req.ConcurrentUsers == 0 {
		req.ConcurrentUsers = 10
	}
	if len(req.Symbols) == 0 {
		req.Symbols = []string{"AAPL", "GOOGL", "MSFT", "TSLA"}
	}
	if req.Intensity == "" {
		req.Intensity = "light"
	}

	// Create scenario
	scenario := demo.LoadTestScenario{
		Name:            "API Load Test",
		Description:     "Load test triggered via API",
		Intensity:       demo.LoadIntensity(req.Intensity),
		Duration:        time.Duration(req.Duration) * time.Second,
		OrdersPerSecond: req.OrdersPerSecond,
		ConcurrentUsers: req.ConcurrentUsers,
		Symbols:         req.Symbols,
		OrderTypes:      []string{"limit", "market"},
		PriceVariation:  0.05,
		VolumeRange: demo.VolumeRange{
			Min: 1,
			Max: 100,
		},
		UserBehaviorPattern: demo.UserBehaviorPattern{
			BuyRatio:         0.5,
			MarketOrderRatio: 0.3,
		},
		RampUp: demo.RampUpConfig{
			Enabled:  true,
			Duration: time.Duration(req.Duration/10) * time.Second,
		},
	}

	// Start load test
	if err := h.demoController.StartLoadTest(c.Request.Context(), scenario); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "LOAD_TEST_FAILED",
				Message: "Failed to start load test",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message":            "Load test started successfully",
			"intensity":          req.Intensity,
			"duration_seconds":   req.Duration,
			"orders_per_second":  req.OrdersPerSecond,
			"concurrent_users":   req.ConcurrentUsers,
		},
	})
}

// GetLoadTestStatus handles GET /demo/load-test/status
func (h *DemoHandlerImpl) GetLoadTestStatus(c *gin.Context) {
	status, err := h.demoController.GetLoadTestStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "STATUS_FETCH_FAILED",
				Message: "Failed to get load test status",
				Details: err.Error(),
			},
		})
		return
	}

	responseData := map[string]interface{}{
		"is_running": status.IsRunning,
		"phase":      status.Phase,
		"progress":   status.Progress,
	}

	if status.Scenario != nil {
		responseData["scenario"] = map[string]interface{}{
			"name":              status.Scenario.Name,
			"intensity":         status.Scenario.Intensity,
			"orders_per_second": status.Scenario.OrdersPerSecond,
		}
	}

	if status.CurrentMetrics != nil {
		responseData["current_metrics"] = map[string]interface{}{
			"orders_per_second": status.CurrentMetrics.OrdersPerSecond,
			"average_latency":   status.CurrentMetrics.AverageLatency,
			"error_rate":        status.CurrentMetrics.ErrorRate,
		}
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    responseData,
	})
}

// GetLoadTestResults handles GET /demo/load-test/results
func (h *DemoHandlerImpl) GetLoadTestResults(c *gin.Context) {
	status, err := h.demoController.GetLoadTestStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "RESULTS_FETCH_FAILED",
				Message: "Failed to get load test results",
				Details: err.Error(),
			},
		})
		return
	}

	responseData := map[string]interface{}{
		"completed":        !status.IsRunning && status.Phase == demo.LoadPhaseCompleted,
		"start_time":       status.StartTime,
		"elapsed_time":     status.ElapsedTime.String(),
		"remaining_time":   status.RemainingTime.String(),
		"completed_orders": status.CompletedOrders,
		"failed_orders":    status.FailedOrders,
		"active_orders":    status.ActiveOrders,
	}

	if status.CurrentMetrics != nil {
		responseData["final_metrics"] = map[string]interface{}{
			"orders_per_second": status.CurrentMetrics.OrdersPerSecond,
			"average_latency":   status.CurrentMetrics.AverageLatency,
			"p95_latency":       status.CurrentMetrics.P95Latency,
			"p99_latency":       status.CurrentMetrics.P99Latency,
			"throughput":        status.CurrentMetrics.Throughput,
			"error_rate":        status.CurrentMetrics.ErrorRate,
		}
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    responseData,
	})
}

// StopLoadTest handles DELETE /demo/load-test
func (h *DemoHandlerImpl) StopLoadTest(c *gin.Context) {
	if err := h.demoController.StopLoadTest(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "STOP_FAILED",
				Message: "Failed to stop load test",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Load test stopped successfully",
		},
	})
}

// ChaosTestRequest represents the request for starting a chaos test
type ChaosTestRequest struct {
	Type       string                 `json:"type"`
	Duration   int                    `json:"duration"` // seconds, but can also accept string like "60s"
	Severity   string                 `json:"severity"`
	Target     map[string]interface{} `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
	Recovery   map[string]interface{} `json:"recovery"`
}

// StartChaosTest handles POST /demo/chaos-test
func (h *DemoHandlerImpl) StartChaosTest(c *gin.Context) {
	var req ChaosTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: err.Error(),
			},
		})
		return
	}

	// Set defaults
	if req.Duration == 0 {
		req.Duration = 60
	}
	if req.Severity == "" {
		req.Severity = "medium"
	}

	// Create scenario
	scenario := demo.ChaosTestScenario{
		Name:        "API Chaos Test - " + req.Type,
		Description: "Chaos test triggered via API",
		Type:        demo.ChaosType(req.Type),
		Duration:    time.Duration(req.Duration) * time.Second,
		Severity:    demo.ChaosSeverity(req.Severity),
		Target: demo.ChaosTarget{
			Component: getStringFromMap(req.Target, "component", "trading_engine"),
			Percentage: getFloatFromMap(req.Target, "percentage", 50),
		},
		Parameters: demo.ChaosParams{
			LatencyMs:       int(getFloatFromMap(req.Parameters, "latency_ms", 100)),
			ErrorRate:       getFloatFromMap(req.Parameters, "error_rate", 0.1),
			CPULimitPercent: getFloatFromMap(req.Parameters, "cpu_percentage", 80),
			MemoryLimitMB:   int(getFloatFromMap(req.Parameters, "memory_limit_mb", 1024)),
		},
		Recovery: demo.RecoveryConfig{
			AutoRecover:     getBoolFromMap(req.Recovery, "auto_recover", true),
			GracefulRecover: getBoolFromMap(req.Recovery, "graceful_recover", true),
		},
	}

	// Start chaos test
	if err := h.demoController.TriggerChaosTest(c.Request.Context(), scenario); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "CHAOS_TEST_FAILED",
				Message: "Failed to start chaos test",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message":        "Chaos test started successfully",
			"type":           req.Type,
			"duration":       req.Duration,
			"severity":       req.Severity,
		},
	})
}

// GetChaosTestStatus handles GET /demo/chaos-test/status
func (h *DemoHandlerImpl) GetChaosTestStatus(c *gin.Context) {
	status, err := h.demoController.GetChaosTestStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "STATUS_FETCH_FAILED",
				Message: "Failed to get chaos test status",
				Details: err.Error(),
			},
		})
		return
	}

	responseData := map[string]interface{}{
		"is_running": status.IsRunning,
		"phase":      status.Phase,
	}

	if status.Scenario != nil {
		responseData["scenario"] = map[string]interface{}{
			"type":     status.Scenario.Type,
			"severity": status.Scenario.Severity,
			"duration": status.Scenario.Duration.Seconds(),
		}
	}

	if status.Metrics != nil {
		responseData["metrics"] = map[string]interface{}{
			"service_degradation": status.Metrics.ServiceDegradation,
			"resilience_score":    status.Metrics.ResilienceScore,
			"errors_generated":    status.Metrics.ErrorsGenerated,
		}
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    responseData,
	})
}

// GetChaosTestResults handles GET /demo/chaos-test/results
func (h *DemoHandlerImpl) GetChaosTestResults(c *gin.Context) {
	status, err := h.demoController.GetChaosTestStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "RESULTS_FETCH_FAILED",
				Message: "Failed to get chaos test results",
				Details: err.Error(),
			},
		})
		return
	}

	responseData := map[string]interface{}{
		"completed":  !status.IsRunning && status.Phase == demo.ChaosPhaseCompleted,
		"start_time": status.StartTime,
		"phase":      status.Phase,
	}

	if status.Scenario != nil {
		responseData["scenario"] = map[string]interface{}{
			"type":     status.Scenario.Type,
			"severity": status.Scenario.Severity,
			"duration": status.Scenario.Duration.Seconds(),
			"target":   status.Scenario.Target,
		}
	}

	if status.Metrics != nil {
		responseData["impact"] = map[string]interface{}{
			"service_degradation": status.Metrics.ServiceDegradation,
			"resilience_score":    status.Metrics.ResilienceScore,
			"errors_generated":    status.Metrics.ErrorsGenerated,
		}
	}

	responseData["affected_targets_count"] = len(status.AffectedTargets)
	responseData["errors_count"] = len(status.Errors)

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    responseData,
	})
}

// StopChaosTest handles DELETE /demo/chaos-test
func (h *DemoHandlerImpl) StopChaosTest(c *gin.Context) {
	if err := h.demoController.StopChaosTest(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "STOP_FAILED",
				Message: "Failed to stop chaos test",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Chaos test stopped successfully",
		},
	})
}

// Helper functions to extract values from map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string, defaultVal string) string {
	if m == nil {
		return defaultVal
	}
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultVal
}

func getFloatFromMap(m map[string]interface{}, key string, defaultVal float64) float64 {
	if m == nil {
		return defaultVal
	}
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return defaultVal
}

func getBoolFromMap(m map[string]interface{}, key string, defaultVal bool) bool {
	if m == nil {
		return defaultVal
	}
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultVal
}

// ResetSystem handles POST /demo/reset
func (h *DemoHandlerImpl) ResetSystem(c *gin.Context) {
	if err := h.demoController.ResetSystem(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "RESET_FAILED",
				Message: "Failed to reset system",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "System reset successfully",
		},
	})
}

// GetSystemStatus handles GET /demo/status
func (h *DemoHandlerImpl) GetSystemStatus(c *gin.Context) {
	status, err := h.demoController.GetSystemStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{
			Success: false,
			Error: &dto.APIError{
				Code:    "STATUS_FETCH_FAILED",
				Message: "Failed to get system status",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"overall":          status.Overall,
			"trading_engine":   status.TradingEngine,
			"order_service":    status.OrderService,
			"metrics_service":  status.MetricsService,
			"database":         status.Database,
			"active_scenarios": status.ActiveScenarios,
			"alerts":           status.Alerts,
		},
	})
}
