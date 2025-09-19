package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"simulated_exchange/internal/api"
	"simulated_exchange/internal/api/dto"
	"simulated_exchange/internal/api/handlers"
	"simulated_exchange/internal/config"
	"simulated_exchange/internal/metrics"
	"simulated_exchange/internal/simulation"
	"simulated_exchange/internal/types"
)

// Container manages all application dependencies following dependency injection principles
type Container struct {
	config *config.Config
	logger *slog.Logger

	// Core services
	orderService   OrderService
	metricsService MetricsService
	healthService  HealthServiceInterface

	// Simulation components
	marketSimulator MarketSimulator
	priceGenerator  PriceGenerator
	orderGenerator  OrderGenerator

	// API components
	server *api.Server
	deps   *api.DependencyContainer
}

// Service interfaces for dependency inversion
type OrderService interface {
	PlaceOrder(orderID, symbol string, side, orderType string, quantity, price float64) error
	GetOrder(orderID string) (handlers.Order, error)
	CancelOrder(orderID string) error
	GetOrderBook(symbol string) (handlers.OrderBook, error)
}

type MetricsService interface {
	GetRealTimeMetrics() handlers.MetricsSnapshot
	GetPerformanceAnalysis() handlers.PerformanceAnalysis
	IsHealthy() bool
}

type HealthServiceInterface interface {
	Check(ctx context.Context) HealthStatus
	RegisterCheck(name string, check HealthCheck)
	GetStatus() HealthStatus
}

// Use the actual simulation interfaces
type MarketSimulator = simulation.MarketSimulator
type PriceGenerator = simulation.PriceGenerator
type OrderGenerator = simulation.OrderGenerator

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Duration    time.Duration     `json:"duration"`
	Checks      map[string]string `json:"checks"`
	ErrorCount  int               `json:"error_count"`
	LastError   string            `json:"last_error,omitempty"`
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config, logger *slog.Logger) (*Container, error) {
	container := &Container{
		config: cfg,
		logger: logger,
	}

	if err := container.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	return container, nil
}

// initializeServices initializes all services in the correct order
func (c *Container) initializeServices() error {
	// Initialize core services
	if err := c.initializeCoreServices(); err != nil {
		return fmt.Errorf("failed to initialize core services: %w", err)
	}

	// Initialize simulation components if enabled
	if c.config.Simulation.Enabled {
		if err := c.initializeSimulationComponents(); err != nil {
			return fmt.Errorf("failed to initialize simulation components: %w", err)
		}
	}

	// Initialize health service
	if err := c.initializeHealthService(); err != nil {
		return fmt.Errorf("failed to initialize health service: %w", err)
	}

	// Initialize API components
	if err := c.initializeAPIComponents(); err != nil {
		return fmt.Errorf("failed to initialize API components: %w", err)
	}

	return nil
}

// initializeCoreServices initializes the core business services
func (c *Container) initializeCoreServices() error {
	c.logger.Info("Initializing core services")

	// Initialize order service
	c.orderService = NewMockOrderService(c.logger)

	// Initialize metrics service
	c.metricsService = NewMockMetricsService(c.config.Metrics, c.logger)

	c.logger.Info("Core services initialized successfully")
	return nil
}

// initializeSimulationComponents initializes market simulation components
func (c *Container) initializeSimulationComponents() error {
	c.logger.Info("Initializing simulation components")

	// Create price generator with config
	priceConfig := simulation.PriceGeneratorConfig{
		BaseVolatility:    0.02,
		VolatilityDecay:   0.95,
		SpreadPercentage:  0.001,
		PriceStepSize:     0.01,
		TrendPersistence:  0.7,
		MeanReversion:     0.3,
		HistorySize:       1000,
		RandomSeed:        time.Now().UnixNano(),
	}
	c.priceGenerator = simulation.NewRealisticPriceGenerator(priceConfig)

	// Create order generator with config
	orderConfig := simulation.OrderGeneratorConfig{
		BaseOrderRate:     10.0,
		MarketHoursBoost:  1.5,
		VolatilityBoost:   2.0,
		NewsEventBoost:    3.0,
		UserTypeMix:       map[string]float64{"conservative": 0.4, "aggressive": 0.3, "momentum": 0.3},
		RandomSeed:        time.Now().UnixNano(),
		RealtimeMode:      true,
	}
	c.orderGenerator = simulation.NewRealisticOrderGenerator(orderConfig)

	// Create event generator with config
	eventConfig := simulation.EventGeneratorConfig{
		BaseEventRate:      0.1,
		EventProbabilities: map[simulation.EventType]float64{},
		RandomSeed:         time.Now().UnixNano(),
	}
	eventGenerator := simulation.NewPatternEventGenerator(eventConfig)

	// Create trading engine interface for simulation
	tradingEngine := NewSimulationTradingEngine(c.orderService, c.logger)

	// Create market simulator
	c.marketSimulator = simulation.NewRealisticSimulator(
		c.priceGenerator,
		c.orderGenerator,
		eventGenerator,
		tradingEngine,
	)

	c.logger.Info("Simulation components initialized successfully")
	return nil
}

// initializeHealthService initializes the health check service
func (c *Container) initializeHealthService() error {
	c.logger.Info("Initializing health service")

	c.healthService = NewHealthService(c.config.Health, c.logger)

	// Register health checks for all services
	c.healthService.RegisterCheck("metrics", func(ctx context.Context) error {
		if !c.metricsService.IsHealthy() {
			return fmt.Errorf("metrics service is unhealthy")
		}
		return nil
	})

	if c.config.Simulation.Enabled && c.marketSimulator != nil {
		c.healthService.RegisterCheck("simulation", func(ctx context.Context) error {
			status := c.marketSimulator.GetSimulationStatus()
			if status.LastError != nil {
				return fmt.Errorf("simulation error: %s", status.LastError.Error())
			}
			return nil
		})
	}

	c.logger.Info("Health service initialized successfully")
	return nil
}

// initializeAPIComponents initializes the API server and routing
func (c *Container) initializeAPIComponents() error {
	c.logger.Info("Initializing API components")

	// Create dependency container for API
	deps := api.NewDependencyContainer(c.orderService, c.metricsService)

	// Create server configuration
	serverConfig := &api.Config{
		Port:         c.config.Server.Port,
		Environment:  c.config.Server.Environment,
		ReadTimeout:  c.config.Server.ReadTimeout,
		WriteTimeout: c.config.Server.WriteTimeout,
	}

	// Create server
	c.server = api.NewServer(deps, serverConfig)

	// Add health check endpoint to server
	if c.config.Health.Enabled {
		c.addHealthEndpoint()
	}

	c.logger.Info("API components initialized successfully")
	return nil
}

// addHealthEndpoint adds health check endpoint to the server
func (c *Container) addHealthEndpoint() {
	// This would typically be done through the server's router
	// For now, we'll track that the health service is available
	c.logger.Info("Health endpoint configured", "endpoint", c.config.Health.Endpoint)
}

// GetOrderService returns the order service
func (c *Container) GetOrderService() OrderService {
	return c.orderService
}

// GetMetricsService returns the metrics service
func (c *Container) GetMetricsService() MetricsService {
	return c.metricsService
}

// GetHealthService returns the health service
func (c *Container) GetHealthService() HealthServiceInterface {
	return c.healthService
}

// GetMarketSimulator returns the market simulator
func (c *Container) GetMarketSimulator() MarketSimulator {
	return c.marketSimulator
}

// GetServer returns the API server
func (c *Container) GetServer() *api.Server {
	return c.server
}

// GetConfig returns the application configuration
func (c *Container) GetConfig() *config.Config {
	return c.config
}

// Shutdown gracefully shuts down all services in reverse order
func (c *Container) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down container services")

	var errors []error

	// Stop simulation first
	if c.marketSimulator != nil {
		if err := c.marketSimulator.StopSimulation(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop simulation: %w", err))
		}
	}

	// Stop server
	if c.server != nil {
		// The server shutdown is handled by the application layer
		c.logger.Info("Server shutdown will be handled by application layer")
	}

	// Stop health service
	if c.healthService != nil {
		// Health service cleanup if needed
		c.logger.Info("Health service cleanup completed")
	}

	// Stop metrics service
	if c.metricsService != nil {
		// Metrics service cleanup if needed
		c.logger.Info("Metrics service cleanup completed")
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	c.logger.Info("Container shutdown completed successfully")
	return nil
}

// MockOrderService implements OrderService interface
type MockOrderService struct {
	orders map[string]handlers.Order
	mutex  sync.RWMutex
	logger *slog.Logger
}

func NewMockOrderService(logger *slog.Logger) *MockOrderService {
	return &MockOrderService{
		orders: make(map[string]handlers.Order),
		logger: logger,
	}
}

func (m *MockOrderService) PlaceOrder(orderID, symbol string, side, orderType string, quantity, price float64) error {
	order := handlers.Order{
		ID:       orderID,
		Symbol:   symbol,
		Side:     side,
		Type:     orderType,
		Quantity: quantity,
		Price:    price,
		Status:   "active",
	}

	m.mutex.Lock()
	m.orders[orderID] = order
	m.mutex.Unlock()

	m.logger.Info("Order placed", "order_id", orderID, "symbol", symbol, "side", side, "quantity", quantity)
	return nil
}

func (m *MockOrderService) GetOrder(orderID string) (handlers.Order, error) {
	m.mutex.RLock()
	order, exists := m.orders[orderID]
	m.mutex.RUnlock()

	if exists {
		return order, nil
	}
	return handlers.Order{}, fmt.Errorf("order not found: %s", orderID)
}

func (m *MockOrderService) CancelOrder(orderID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if order, exists := m.orders[orderID]; exists {
		order.Status = "cancelled"
		m.orders[orderID] = order
		m.logger.Info("Order cancelled", "order_id", orderID)
		return nil
	}
	return fmt.Errorf("order not found: %s", orderID)
}

func (m *MockOrderService) GetOrderBook(symbol string) (handlers.OrderBook, error) {
	return handlers.OrderBook{
		Symbol: symbol,
		Bids:   []handlers.OrderBookEntry{},
		Asks:   []handlers.OrderBookEntry{},
	}, nil
}

// MockMetricsService implements MetricsService interface
type MockMetricsService struct {
	collector *metrics.RealTimeMetrics
	analyzer  *metrics.AIAnalyzer
	isHealthy bool
	logger    *slog.Logger
}

func NewMockMetricsService(cfg config.MetricsConfig, logger *slog.Logger) *MockMetricsService {
	collector := metrics.NewRealTimeMetrics(cfg.CollectionTime)
	analyzer := metrics.NewAIAnalyzer()

	service := &MockMetricsService{
		collector: collector,
		analyzer:  analyzer,
		isHealthy: true,
		logger:    logger,
	}

	// Start background data generation if metrics are enabled
	if cfg.Enabled {
		go service.generateMockData()
	}

	return service
}

func (m *MockMetricsService) generateMockData() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	orderCount := int64(0)
	for range ticker.C {
		orderCount++

		// Generate mock order event
		m.collector.RecordOrder(metrics.OrderEvent{
			OrderID:   fmt.Sprintf("order_%d", orderCount),
			Symbol:    m.getRandomSymbol(),
			Side:      m.getRandomSide(),
			Type:      "limit",
			Quantity:  m.getRandomQuantity(),
			Price:     m.getRandomPrice(),
			Timestamp: time.Now(),
			Latency:   m.getRandomLatency(),
		})

		// Generate mock trade event occasionally
		if orderCount%3 == 0 {
			m.collector.RecordTrade(metrics.TradeEvent{
				TradeID:     fmt.Sprintf("trade_%d", orderCount/3),
				Symbol:      m.getRandomSymbol(),
				Quantity:    m.getRandomQuantity(),
				Price:       m.getRandomPrice(),
				Timestamp:   time.Now(),
				Latency:     m.getRandomLatency(),
				BuyOrderID:  fmt.Sprintf("buy_order_%d", orderCount),
				SellOrderID: fmt.Sprintf("sell_order_%d", orderCount+1),
			})
		}
	}
}

func (m *MockMetricsService) GetRealTimeMetrics() handlers.MetricsSnapshot {
	snapshot := m.collector.CalculateMetrics(60 * time.Second)

	symbolMetrics := make(map[string]handlers.SymbolMetrics)
	for symbol, sm := range snapshot.SymbolMetrics {
		symbolMetrics[symbol] = handlers.SymbolMetrics{
			OrderCount: sm.OrderCount,
			TradeCount: sm.TradeCount,
			Volume:     sm.Volume,
			AvgPrice:   sm.AvgPrice,
		}
	}

	return handlers.MetricsSnapshot{
		OrderCount:    snapshot.OrderCount,
		TradeCount:    snapshot.TradeCount,
		TotalVolume:   snapshot.TotalVolume,
		AvgLatency:    snapshot.AvgLatency.String(),
		OrdersPerSec:  snapshot.OrdersPerSec,
		TradesPerSec:  snapshot.TradesPerSec,
		SymbolMetrics: symbolMetrics,
	}
}

func (m *MockMetricsService) GetPerformanceAnalysis() handlers.PerformanceAnalysis {
	snapshots := []metrics.MetricsSnapshot{m.collector.CalculateMetrics(60 * time.Second)}
	latencyAnalysis := m.analyzer.AnalyzeLatency(snapshots)
	bottlenecks := m.analyzer.DetectBottlenecks(snapshots[len(snapshots)-1])

	var bottleneckDTOs []handlers.Bottleneck
	for _, b := range bottlenecks {
		bottleneckDTOs = append(bottleneckDTOs, handlers.Bottleneck{
			Type:        b.Type,
			Severity:    b.Severity,
			Description: b.Description,
		})
	}

	recommendations := []string{
		"System performance is within normal parameters",
		"Consider scaling if traffic increases significantly",
		"Monitor latency during peak hours",
	}

	return handlers.PerformanceAnalysis{
		Timestamp:       time.Now().Format(time.RFC3339),
		TrendDirection:  string(latencyAnalysis.Trend),
		Bottlenecks:     bottleneckDTOs,
		Recommendations: recommendations,
	}
}

func (m *MockMetricsService) IsHealthy() bool {
	return m.isHealthy
}

// Helper methods for generating mock data
func (m *MockMetricsService) getRandomSymbol() string {
	symbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA", "AMZN", "META", "NVDA"}
	return symbols[time.Now().UnixNano()%int64(len(symbols))]
}

func (m *MockMetricsService) getRandomSide() types.OrderSide {
	if time.Now().UnixNano()%2 == 0 {
		return types.Buy
	}
	return types.Sell
}

func (m *MockMetricsService) getRandomQuantity() float64 {
	return float64(time.Now().UnixNano()%1000 + 10)
}

func (m *MockMetricsService) getRandomPrice() float64 {
	return float64(time.Now().UnixNano()%50000+10000) / 100.0
}

func (m *MockMetricsService) getRandomLatency() time.Duration {
	ms := time.Now().UnixNano()%50 + 1
	return time.Duration(ms) * time.Millisecond
}

// SimulationTradingEngine adapts OrderService for simulation use
type SimulationTradingEngine struct {
	orderService OrderService
	logger       *slog.Logger
}

func NewSimulationTradingEngine(orderService OrderService, logger *slog.Logger) *SimulationTradingEngine {
	return &SimulationTradingEngine{
		orderService: orderService,
		logger:       logger,
	}
}

func (s *SimulationTradingEngine) PlaceOrder(order dto.PlaceOrderRequest) (interface{}, error) {
	// Generate a unique order ID for simulation
	orderID := fmt.Sprintf("sim_%d_%s", time.Now().UnixNano(), order.Symbol)

	err := s.orderService.PlaceOrder(
		orderID,
		order.Symbol,
		order.Side,
		order.Type,
		order.Quantity,
		order.Price,
	)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"order_id": orderID,
		"status":   "placed",
	}, nil
}