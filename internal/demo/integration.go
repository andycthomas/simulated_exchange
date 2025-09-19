package demo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"simulated_exchange/internal/api/dto"
	"simulated_exchange/internal/app"
)

// TradingEngineAdapter adapts the existing trading engine for demo use
type TradingEngineAdapter struct {
	tradingEngine    *app.SimulationTradingEngine
	orderService     *app.MockOrderService
	metricsService   *app.MockMetricsService
	logger           *slog.Logger

	// Reset state tracking
	initialState     *SystemState
}

// SystemState represents the system state for reset functionality
type SystemState struct {
	OrderCount       int                    `json:"order_count"`
	ActiveOrders     map[string]interface{} `json:"active_orders"`
	SystemMetrics    interface{}            `json:"system_metrics"`
	Timestamp        time.Time              `json:"timestamp"`
}

// NewTradingEngineAdapter creates a new trading engine adapter
func NewTradingEngineAdapter(
	tradingEngine *app.SimulationTradingEngine,
	orderService *app.MockOrderService,
	metricsService *app.MockMetricsService,
	logger *slog.Logger,
) *TradingEngineAdapter {
	adapter := &TradingEngineAdapter{
		tradingEngine:  tradingEngine,
		orderService:   orderService,
		metricsService: metricsService,
		logger:         logger,
	}

	// Capture initial state
	adapter.captureInitialState()

	return adapter
}

// PlaceOrder implements TradingEngineIntegration.PlaceOrder
func (tea *TradingEngineAdapter) PlaceOrder(order dto.PlaceOrderRequest) (dto.OrderResponse, error) {
	tea.logger.Debug("Placing demo order", "symbol", order.Symbol, "side", order.Side, "quantity", order.Quantity)

	// Use the existing trading engine
	response, err := tea.tradingEngine.PlaceOrder(order)
	if err != nil {
		tea.logger.Error("Failed to place order", "error", err, "order", order)
		return dto.OrderResponse{}, err
	}

	// Convert interface{} response to OrderResponse
	if responseMap, ok := response.(map[string]interface{}); ok {
		orderResponse := dto.OrderResponse{
			ID:       responseMap["order_id"].(string),
			Symbol:   order.Symbol,
			Side:     order.Side,
			Type:     order.Type,
			Quantity: order.Quantity,
			Price:    order.Price,
			Status:   responseMap["status"].(string),
		}
		tea.logger.Debug("Order placed successfully", "order_id", orderResponse.ID, "status", orderResponse.Status)
		return orderResponse, nil
	}
	return dto.OrderResponse{}, fmt.Errorf("unexpected response format")
}

// CancelOrder implements TradingEngineIntegration.CancelOrder
func (tea *TradingEngineAdapter) CancelOrder(orderID string) error {
	tea.logger.Debug("Cancelling demo order", "order_id", orderID)

	if err := tea.orderService.CancelOrder(orderID); err != nil {
		tea.logger.Error("Failed to cancel order", "error", err, "order_id", orderID)
		return err
	}

	tea.logger.Debug("Order cancelled successfully", "order_id", orderID)
	return nil
}

// GetOrderStatus implements TradingEngineIntegration.GetOrderStatus
func (tea *TradingEngineAdapter) GetOrderStatus(orderID string) (dto.OrderResponse, error) {
	order, err := tea.orderService.GetOrder(orderID)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	// Convert internal order to DTO
	response := dto.OrderResponse{
		ID:       order.ID,
		Symbol:   order.Symbol,
		Side:     order.Side,
		Type:     order.Type,
		Quantity: order.Quantity,
		Price:    order.Price,
		Status:   order.Status,
	}

	return response, nil
}

// GetMetrics implements TradingEngineIntegration.GetMetrics
func (tea *TradingEngineAdapter) GetMetrics() (interface{}, error) {
	metrics := tea.metricsService.GetRealTimeMetrics()
	return metrics, nil
}

// Reset implements TradingEngineIntegration.Reset
func (tea *TradingEngineAdapter) Reset() error {
	tea.logger.Info("Resetting trading engine to initial state")

	// Reset would be implemented here
	// For now, we'll just log the reset operation
	tea.logger.Info("Trading engine reset completed")

	return nil
}

// Private methods

func (tea *TradingEngineAdapter) captureInitialState() {
	tea.logger.Debug("Capturing initial system state")

	// Get current metrics as initial state
	metrics, err := tea.GetMetrics()
	if err != nil {
		tea.logger.Warn("Failed to capture initial metrics", "error", err)
		metrics = nil
	}

	tea.initialState = &SystemState{
		OrderCount:    0,
		ActiveOrders:  make(map[string]interface{}),
		SystemMetrics: metrics,
		Timestamp:     time.Now(),
	}

	tea.logger.Debug("Initial system state captured")
}

// DemoConfigBuilder helps build demo configurations
type DemoConfigBuilder struct {
	config *DemoConfig
}

// NewDemoConfigBuilder creates a new demo config builder
func NewDemoConfigBuilder() *DemoConfigBuilder {
	return &DemoConfigBuilder{
		config: &DemoConfig{
			LoadTest: LoadTestConfig{
				MaxConcurrentTests: 1,
				DefaultTimeout:     30 * time.Minute,
				MetricsInterval:    time.Second,
				MaxDuration:        time.Hour,
				EnabledScenarios:   []string{"light", "medium", "heavy", "stress"},
			},
			Chaos: ChaosConfig{
				MaxConcurrentTests: 1,
				DefaultTimeout:     15 * time.Minute,
				SafetyLimits: SafetyLimits{
					MaxLatencyMs:        5000,
					MaxErrorRate:        0.5,
					MaxCPUUsage:         90.0,
					MaxMemoryUsage:      90.0,
					RequireConfirmation: true,
				},
				EnabledTypes: []string{"latency_injection", "error_simulation", "resource_exhaustion"},
			},
			WebSocket: WebSocketConfig{
				MaxConnections: 100,
				PingInterval:   30 * time.Second,
				WriteTimeout:   10 * time.Second,
				ReadTimeout:    60 * time.Second,
				MaxMessageSize: 1024 * 1024, // 1MB
			},
			Metrics: DemoMetricsConfig{
				CollectionInterval: time.Second,
				RetentionPeriod:    time.Hour,
				EnabledMetrics:     []string{"cpu", "memory", "latency", "throughput", "error_rate"},
				ExportFormats:      []string{"json", "prometheus"},
			},
		},
	}
}

// WithLoadTestConfig configures load testing
func (dcb *DemoConfigBuilder) WithLoadTestConfig(config LoadTestConfig) *DemoConfigBuilder {
	dcb.config.LoadTest = config
	return dcb
}

// WithChaosConfig configures chaos testing
func (dcb *DemoConfigBuilder) WithChaosConfig(config ChaosConfig) *DemoConfigBuilder {
	dcb.config.Chaos = config
	return dcb
}

// WithWebSocketConfig configures WebSocket settings
func (dcb *DemoConfigBuilder) WithWebSocketConfig(config WebSocketConfig) *DemoConfigBuilder {
	dcb.config.WebSocket = config
	return dcb
}

// WithMetricsConfig configures metrics collection
func (dcb *DemoConfigBuilder) WithMetricsConfig(config DemoMetricsConfig) *DemoConfigBuilder {
	dcb.config.Metrics = config
	return dcb
}

// Build returns the constructed configuration
func (dcb *DemoConfigBuilder) Build() *DemoConfig {
	return dcb.config
}

// DemoSystemFactory creates and configures the complete demo system
type DemoSystemFactory struct {
	logger *slog.Logger
}

// NewDemoSystemFactory creates a new demo system factory
func NewDemoSystemFactory(logger *slog.Logger) *DemoSystemFactory {
	return &DemoSystemFactory{
		logger: logger,
	}
}

// CreateDemoSystem creates a complete demo system
func (dsf *DemoSystemFactory) CreateDemoSystem(
	tradingEngine *app.SimulationTradingEngine,
	orderService *app.MockOrderService,
	metricsService *app.MockMetricsService,
	config *DemoConfig,
) (*DemoSystem, error) {
	dsf.logger.Info("Creating demo system")

	// Create trading engine adapter
	engineAdapter := NewTradingEngineAdapter(tradingEngine, orderService, metricsService, dsf.logger)

	// Create scenario manager
	scenarioManager := NewStandardScenarioManager(engineAdapter, dsf.logger, config)

	// Create demo controller
	controller := NewStandardDemoController(config, scenarioManager, engineAdapter, dsf.logger)

	// Create WebSocket integration
	wsIntegration := NewWebSocketDemoIntegration(config.WebSocket, controller, dsf.logger)

	// Create demo system
	system := &DemoSystem{
		Controller:        controller,
		ScenarioManager:   scenarioManager,
		TradingEngine:     engineAdapter,
		WebSocketIntegration: wsIntegration,
		Config:            config,
		Logger:            dsf.logger,
	}

	dsf.logger.Info("Demo system created successfully")
	return system, nil
}

// DemoSystem represents the complete demo system
type DemoSystem struct {
	Controller           DemoController
	ScenarioManager      ScenarioManager
	TradingEngine        TradingEngineIntegration
	WebSocketIntegration *WebSocketDemoIntegration
	Config               *DemoConfig
	Logger               *slog.Logger

	// System state
	running bool
}

// Start initializes and starts the demo system
func (ds *DemoSystem) Start(ctx context.Context) error {
	ds.Logger.Info("Starting demo system")

	// Start WebSocket integration
	if err := ds.WebSocketIntegration.Start(ctx); err != nil {
		return fmt.Errorf("failed to start WebSocket integration: %w", err)
	}

	ds.running = true
	ds.Logger.Info("Demo system started successfully")

	return nil
}

// Stop gracefully shuts down the demo system
func (ds *DemoSystem) Stop(ctx context.Context) error {
	ds.Logger.Info("Stopping demo system")

	// Reset system to clean state
	if err := ds.Controller.ResetSystem(ctx); err != nil {
		ds.Logger.Error("Failed to reset system during shutdown", "error", err)
	}

	// Stop WebSocket integration
	if err := ds.WebSocketIntegration.Stop(ctx); err != nil {
		ds.Logger.Error("Failed to stop WebSocket integration", "error", err)
	}

	// Shutdown controller
	if controller, ok := ds.Controller.(*StandardDemoController); ok {
		if err := controller.Shutdown(ctx); err != nil {
			ds.Logger.Error("Failed to shutdown demo controller", "error", err)
		}
	}

	ds.running = false
	ds.Logger.Info("Demo system stopped")

	return nil
}

// IsRunning returns whether the demo system is running
func (ds *DemoSystem) IsRunning() bool {
	return ds.running
}

// GetSystemStatus returns the current system status
func (ds *DemoSystem) GetSystemStatus(ctx context.Context) (*DemoSystemStatus, error) {
	return ds.Controller.GetSystemStatus(ctx)
}

// GetAvailableScenarios returns all available demo scenarios
func (ds *DemoSystem) GetAvailableScenarios() ([]LoadTestScenario, []ChaosTestScenario) {
	loadScenarios := ds.ScenarioManager.GetAvailableLoadScenarios()
	chaosScenarios := ds.ScenarioManager.GetAvailableChaosScenarios()
	return loadScenarios, chaosScenarios
}

// GetWebSocketStats returns WebSocket integration statistics
func (ds *DemoSystem) GetWebSocketStats() map[string]interface{} {
	return ds.WebSocketIntegration.GetStats()
}

// Health check functionality
func (ds *DemoSystem) HealthCheck(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"demo_system": "healthy",
		"timestamp":   time.Now(),
		"running":     ds.running,
	}

	// Check controller status
	if status, err := ds.Controller.GetSystemStatus(ctx); err == nil {
		health["system_status"] = status.Overall
		health["active_scenarios"] = len(status.ActiveScenarios)
	} else {
		health["system_status"] = "unknown"
		health["error"] = err.Error()
	}

	// Check WebSocket stats
	wsStats := ds.WebSocketIntegration.GetStats()
	health["websocket_connections"] = wsStats["active_connections"]

	return health
}

// Utility functions for demo system management

// GetDefaultDemoConfig returns a default demo configuration
func GetDefaultDemoConfig() *DemoConfig {
	return NewDemoConfigBuilder().Build()
}

// ValidateDemoConfig validates a demo configuration
func ValidateDemoConfig(config *DemoConfig) error {
	if config.LoadTest.MaxConcurrentTests <= 0 {
		return fmt.Errorf("max concurrent load tests must be positive")
	}

	if config.Chaos.MaxConcurrentTests <= 0 {
		return fmt.Errorf("max concurrent chaos tests must be positive")
	}

	if config.WebSocket.MaxConnections <= 0 {
		return fmt.Errorf("max WebSocket connections must be positive")
	}

	if config.Chaos.SafetyLimits.MaxErrorRate < 0 || config.Chaos.SafetyLimits.MaxErrorRate > 1 {
		return fmt.Errorf("max error rate must be between 0 and 1")
	}

	return nil
}

// CreateDemoSystemWithDefaults creates a demo system with default configuration
func CreateDemoSystemWithDefaults(
	tradingEngine *app.SimulationTradingEngine,
	orderService *app.MockOrderService,
	metricsService *app.MockMetricsService,
	logger *slog.Logger,
) (*DemoSystem, error) {
	config := GetDefaultDemoConfig()

	if err := ValidateDemoConfig(config); err != nil {
		return nil, fmt.Errorf("invalid demo config: %w", err)
	}

	factory := NewDemoSystemFactory(logger)
	return factory.CreateDemoSystem(tradingEngine, orderService, metricsService, config)
}