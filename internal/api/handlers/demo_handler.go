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

// StartChaosTest handles POST /demo/chaos-test
func (h *DemoHandlerImpl) StartChaosTest(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, dto.APIResponse{
		Success: false,
		Error: &dto.APIError{
			Code:    "NOT_IMPLEMENTED",
			Message: "Chaos testing not yet implemented",
		},
	})
}

// GetChaosTestStatus handles GET /demo/chaos-test/status
func (h *DemoHandlerImpl) GetChaosTestStatus(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, dto.APIResponse{
		Success: false,
		Error: &dto.APIError{
			Code:    "NOT_IMPLEMENTED",
			Message: "Chaos testing not yet implemented",
		},
	})
}

// StopChaosTest handles DELETE /demo/chaos-test
func (h *DemoHandlerImpl) StopChaosTest(c *gin.Context) {
	// Implementation placeholder
	c.JSON(http.StatusNotImplemented, dto.APIResponse{
		Success: false,
		Error: &dto.APIError{
			Code:    "NOT_IMPLEMENTED",
			Message: "Chaos testing not yet implemented",
		},
	})
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
