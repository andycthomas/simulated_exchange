package demo

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"simulated_exchange/internal/api/dto"
)

// StandardDemoController implements the DemoController interface
type StandardDemoController struct {
	config           *DemoConfig
	scenarioManager  ScenarioManager
	tradingEngine    TradingEngineIntegration
	logger           *slog.Logger

	// State management
	mu                sync.RWMutex
	loadTestStatus    *LoadTestStatus
	chaosTestStatus   *ChaosTestStatus
	systemStatus      *DemoSystemStatus

	// Real-time updates
	subscribers       map[string]DemoSubscriber
	subscribersMu     sync.RWMutex
	updateChan        chan DemoUpdate

	// Control channels
	stopLoadTest      chan struct{}
	stopChaosTest     chan struct{}
	stopUpdates       chan struct{}

	// Metrics collection
	metricsCollector  *MetricsCollector
	isRunning         bool
}

// NewStandardDemoController creates a new demo controller instance
func NewStandardDemoController(
	config *DemoConfig,
	scenarioManager ScenarioManager,
	tradingEngine TradingEngineIntegration,
	logger *slog.Logger,
) *StandardDemoController {
	controller := &StandardDemoController{
		config:          config,
		scenarioManager: scenarioManager,
		tradingEngine:   tradingEngine,
		logger:          logger,
		subscribers:     make(map[string]DemoSubscriber),
		updateChan:      make(chan DemoUpdate, 1000),
		stopLoadTest:    make(chan struct{}),
		stopChaosTest:   make(chan struct{}),
		stopUpdates:     make(chan struct{}),
		loadTestStatus: &LoadTestStatus{
			IsRunning: false,
			Phase:     LoadPhaseCompleted,
		},
		chaosTestStatus: &ChaosTestStatus{
			IsRunning: false,
			Phase:     ChaosPhaseCompleted,
		},
		systemStatus: &DemoSystemStatus{
			Overall: HealthHealthy,
			TradingEngine: ComponentHealth{Status: HealthHealthy},
			OrderService:  ComponentHealth{Status: HealthHealthy},
			MetricsService: ComponentHealth{Status: HealthHealthy},
			Database:      ComponentHealth{Status: HealthHealthy},
			ActiveScenarios: []string{},
			Alerts:        []SystemAlert{},
		},
		metricsCollector: NewMetricsCollector(config.Metrics, logger),
		isRunning:        true,
	}

	// Start background processes
	go controller.processUpdates()
	go controller.collectSystemMetrics()

	return controller
}

// StartLoadTest begins a load testing scenario
func (dc *StandardDemoController) StartLoadTest(ctx context.Context, scenario LoadTestScenario) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if dc.loadTestStatus.IsRunning {
		return fmt.Errorf("load test already running: %s", dc.loadTestStatus.Scenario.Name)
	}

	dc.logger.Info("Starting load test scenario", "scenario", scenario.Name, "intensity", scenario.Intensity)

	// Initialize load test status
	dc.loadTestStatus = &LoadTestStatus{
		IsRunning:       true,
		Scenario:        &scenario,
		StartTime:       time.Now(),
		Progress:        0.0,
		Phase:           LoadPhaseRampUp,
		CurrentMetrics:  &LoadTestMetrics{},
		HistoricalMetrics: []LoadTestMetrics{},
		ActiveOrders:    0,
		CompletedOrders: 0,
		FailedOrders:    0,
		CurrentUsers:    0,
		Errors:          []LoadTestError{},
	}

	// Update system status
	dc.systemStatus.ActiveScenarios = append(dc.systemStatus.ActiveScenarios, scenario.Name)

	// Start the load test execution
	go dc.executeLoadTest(ctx, scenario)

	// Broadcast update
	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateLoadTestStatus,
		Timestamp: time.Now(),
		Data:      dc.loadTestStatus,
		Source:    "load_controller",
	})

	return nil
}

// StopLoadTest stops the current load test
func (dc *StandardDemoController) StopLoadTest(ctx context.Context) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if !dc.loadTestStatus.IsRunning {
		return fmt.Errorf("no load test currently running")
	}

	dc.logger.Info("Stopping load test", "scenario", dc.loadTestStatus.Scenario.Name)

	// Signal stop
	select {
	case dc.stopLoadTest <- struct{}{}:
	default:
	}

	// Update status
	dc.loadTestStatus.IsRunning = false
	dc.loadTestStatus.Phase = LoadPhaseCompleted

	// Update system status
	dc.removeActiveScenario(dc.loadTestStatus.Scenario.Name)

	// Broadcast update
	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateLoadTestStatus,
		Timestamp: time.Now(),
		Data:      dc.loadTestStatus,
		Source:    "load_controller",
	})

	return nil
}

// GetLoadTestStatus returns the current load test status
func (dc *StandardDemoController) GetLoadTestStatus(ctx context.Context) (*LoadTestStatus, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	// Create a copy to avoid race conditions
	status := *dc.loadTestStatus
	if dc.loadTestStatus.IsRunning {
		status.ElapsedTime = time.Since(dc.loadTestStatus.StartTime)
		status.RemainingTime = dc.loadTestStatus.Scenario.Duration - status.ElapsedTime
		if dc.loadTestStatus.Scenario.Duration > 0 {
			status.Progress = float64(status.ElapsedTime) / float64(dc.loadTestStatus.Scenario.Duration) * 100
		}
	}

	return &status, nil
}

// TriggerChaosTest starts a chaos engineering scenario
func (dc *StandardDemoController) TriggerChaosTest(ctx context.Context, scenario ChaosTestScenario) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if dc.chaosTestStatus.IsRunning {
		return fmt.Errorf("chaos test already running: %s", dc.chaosTestStatus.Scenario.Name)
	}

	dc.logger.Info("Starting chaos test scenario", "scenario", scenario.Name, "type", scenario.Type)

	// Initialize chaos test status
	dc.chaosTestStatus = &ChaosTestStatus{
		IsRunning:     true,
		Scenario:      &scenario,
		StartTime:     time.Now(),
		Progress:      0.0,
		Phase:         ChaosPhaseInjection,
		AffectedTargets: []string{},
		Metrics:       &ChaosTestMetrics{},
		Errors:        []ChaosTestError{},
	}

	// Update system status
	dc.systemStatus.ActiveScenarios = append(dc.systemStatus.ActiveScenarios, scenario.Name)

	// Start the chaos test execution
	go dc.executeChaosTest(ctx, scenario)

	// Broadcast update
	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateChaosTestStatus,
		Timestamp: time.Now(),
		Data:      dc.chaosTestStatus,
		Source:    "chaos_controller",
	})

	return nil
}

// StopChaosTest stops the current chaos test
func (dc *StandardDemoController) StopChaosTest(ctx context.Context) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if !dc.chaosTestStatus.IsRunning {
		return fmt.Errorf("no chaos test currently running")
	}

	dc.logger.Info("Stopping chaos test", "scenario", dc.chaosTestStatus.Scenario.Name)

	// Signal stop
	select {
	case dc.stopChaosTest <- struct{}{}:
	default:
	}

	// Update status
	dc.chaosTestStatus.IsRunning = false
	dc.chaosTestStatus.Phase = ChaosPhaseCompleted

	// Update system status
	dc.removeActiveScenario(dc.chaosTestStatus.Scenario.Name)

	// Broadcast update
	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateChaosTestStatus,
		Timestamp: time.Now(),
		Data:      dc.chaosTestStatus,
		Source:    "chaos_controller",
	})

	return nil
}

// GetChaosTestStatus returns the current chaos test status
func (dc *StandardDemoController) GetChaosTestStatus(ctx context.Context) (*ChaosTestStatus, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	// Create a copy to avoid race conditions
	status := *dc.chaosTestStatus
	if dc.chaosTestStatus.IsRunning {
		status.ElapsedTime = time.Since(dc.chaosTestStatus.StartTime)
		status.RemainingTime = dc.chaosTestStatus.Scenario.Duration - status.ElapsedTime
		if dc.chaosTestStatus.Scenario.Duration > 0 {
			status.Progress = float64(status.ElapsedTime) / float64(dc.chaosTestStatus.Scenario.Duration) * 100
		}
	}

	return &status, nil
}

// ResetSystem resets the demo system to a clean state
func (dc *StandardDemoController) ResetSystem(ctx context.Context) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.logger.Info("Resetting demo system")

	// Stop any running tests
	if dc.loadTestStatus.IsRunning {
		select {
		case dc.stopLoadTest <- struct{}{}:
		default:
		}
	}

	if dc.chaosTestStatus.IsRunning {
		select {
		case dc.stopChaosTest <- struct{}{}:
		default:
		}
	}

	// Reset trading engine
	if err := dc.tradingEngine.Reset(); err != nil {
		dc.logger.Error("Failed to reset trading engine", "error", err)
		return fmt.Errorf("failed to reset trading engine: %w", err)
	}

	// Reset status
	dc.loadTestStatus = &LoadTestStatus{
		IsRunning: false,
		Phase:     LoadPhaseCompleted,
	}

	dc.chaosTestStatus = &ChaosTestStatus{
		IsRunning: false,
		Phase:     ChaosPhaseCompleted,
	}

	dc.systemStatus = &DemoSystemStatus{
		Overall: HealthHealthy,
		TradingEngine: ComponentHealth{Status: HealthHealthy},
		OrderService:  ComponentHealth{Status: HealthHealthy},
		MetricsService: ComponentHealth{Status: HealthHealthy},
		Database:      ComponentHealth{Status: HealthHealthy},
		ActiveScenarios: []string{},
		Alerts:        []SystemAlert{},
	}

	// Reset metrics collector
	dc.metricsCollector.Reset()

	// Broadcast reset update
	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateSystemStatus,
		Timestamp: time.Now(),
		Data:      dc.systemStatus,
		Source:    "system_controller",
	})

	dc.logger.Info("Demo system reset completed")
	return nil
}

// GetSystemStatus returns the current system status
func (dc *StandardDemoController) GetSystemStatus(ctx context.Context) (*DemoSystemStatus, error) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	// Update timestamp and return copy
	status := *dc.systemStatus
	status.Timestamp = time.Now()
	status.SystemMetrics = dc.metricsCollector.GetCurrentMetrics()

	return &status, nil
}

// Subscribe adds a subscriber for real-time updates
func (dc *StandardDemoController) Subscribe(subscriber DemoSubscriber) error {
	dc.subscribersMu.Lock()
	defer dc.subscribersMu.Unlock()

	if len(dc.subscribers) >= dc.config.WebSocket.MaxConnections {
		return fmt.Errorf("maximum connections reached")
	}

	dc.subscribers[subscriber.GetID()] = subscriber
	dc.logger.Info("Subscriber added", "id", subscriber.GetID())

	return nil
}

// Unsubscribe removes a subscriber
func (dc *StandardDemoController) Unsubscribe(subscriberID string) error {
	dc.subscribersMu.Lock()
	defer dc.subscribersMu.Unlock()

	if subscriber, exists := dc.subscribers[subscriberID]; exists {
		subscriber.Close()
		delete(dc.subscribers, subscriberID)
		dc.logger.Info("Subscriber removed", "id", subscriberID)
	}

	return nil
}

// BroadcastUpdate sends an update to all subscribers
func (dc *StandardDemoController) BroadcastUpdate(update DemoUpdate) error {
	select {
	case dc.updateChan <- update:
		return nil
	default:
		dc.logger.Warn("Update channel full, dropping update", "type", update.Type)
		return fmt.Errorf("update channel full")
	}
}

// Private methods

func (dc *StandardDemoController) executeLoadTest(ctx context.Context, scenario LoadTestScenario) {
	dc.logger.Info("Executing load test", "scenario", scenario.Name)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	orderCounter := 0

	for {
		select {
		case <-dc.stopLoadTest:
			dc.logger.Info("Load test stopped by request")
			return
		case <-ctx.Done():
			dc.logger.Info("Load test stopped by context cancellation")
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			if elapsed >= scenario.Duration {
				dc.logger.Info("Load test completed", "duration", elapsed)
				dc.completeLoadTest()
				return
			}

			// Execute load test logic
			dc.executeLoadTestIteration(scenario, &orderCounter)

			// Update progress
			dc.updateLoadTestProgress(elapsed, scenario.Duration)
		}
	}
}

func (dc *StandardDemoController) executeChaosTest(ctx context.Context, scenario ChaosTestScenario) {
	dc.logger.Info("Executing chaos test", "scenario", scenario.Name)

	// Injection phase
	dc.updateChaosTestPhase(ChaosPhaseInjection)
	if err := dc.scenarioManager.ExecuteChaosScenario(ctx, scenario); err != nil {
		dc.addChaosTestError("injection_failed", err.Error(), "")
		return
	}

	// Sustained phase
	dc.updateChaosTestPhase(ChaosPhaseSustained)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	for {
		select {
		case <-dc.stopChaosTest:
			dc.logger.Info("Chaos test stopped by request")
			dc.recoverFromChaos(scenario)
			return
		case <-ctx.Done():
			dc.logger.Info("Chaos test stopped by context cancellation")
			dc.recoverFromChaos(scenario)
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			if elapsed >= scenario.Duration {
				dc.logger.Info("Chaos test completed", "duration", elapsed)
				dc.recoverFromChaos(scenario)
				return
			}

			// Update chaos test metrics
			dc.updateChaosTestMetrics()

			// Update progress
			dc.updateChaosTestProgress(elapsed, scenario.Duration)
		}
	}
}

func (dc *StandardDemoController) executeLoadTestIteration(scenario LoadTestScenario, orderCounter *int) {
	// Calculate current load based on ramp-up configuration
	currentLoad := dc.calculateCurrentLoad(scenario)

	// Generate orders based on current load
	for i := 0; i < currentLoad; i++ {
		order := dc.generateTestOrder(scenario, *orderCounter)
		*orderCounter++

		// Place order through trading engine
		go func(orderReq dto.PlaceOrderRequest, orderNum int) {
			start := time.Now()
			_, err := dc.tradingEngine.PlaceOrder(orderReq)
			latency := time.Since(start)

			dc.updateLoadTestMetrics(orderNum, latency, err)
		}(order, *orderCounter)
	}
}

func (dc *StandardDemoController) calculateCurrentLoad(scenario LoadTestScenario) int {
	if !scenario.RampUp.Enabled {
		return scenario.OrdersPerSecond
	}

	elapsed := time.Since(dc.loadTestStatus.StartTime)
	rampUpProgress := float64(elapsed) / float64(scenario.RampUp.Duration)

	if rampUpProgress >= 1.0 {
		return scenario.OrdersPerSecond
	}

	startLoad := float64(scenario.OrdersPerSecond) * scenario.RampUp.StartPercent / 100
	endLoad := float64(scenario.OrdersPerSecond) * scenario.RampUp.EndPercent / 100
	currentLoad := startLoad + (endLoad-startLoad)*rampUpProgress

	return int(currentLoad)
}

func (dc *StandardDemoController) generateTestOrder(scenario LoadTestScenario, orderNum int) dto.PlaceOrderRequest {
	// Select random symbol
	symbol := scenario.Symbols[orderNum%len(scenario.Symbols)]

	// Determine order side based on behavior pattern
	side := "buy"
	if float64(orderNum%100) >= scenario.UserBehaviorPattern.BuyRatio*100 {
		side = "sell"
	}

	// Determine order type
	orderType := "limit"
	if float64(orderNum%100) < scenario.UserBehaviorPattern.MarketOrderRatio*100 {
		orderType = "market"
	}

	// Generate quantity within range
	quantity := scenario.VolumeRange.Min +
		float64(orderNum%100)/100*(scenario.VolumeRange.Max-scenario.VolumeRange.Min)

	// Generate price with variation
	basePrice := 100.0 // Base price for demo
	price := basePrice * (1 + (float64(orderNum%20)-10)/100*scenario.PriceVariation)

	return dto.PlaceOrderRequest{
		Symbol:   symbol,
		Side:     side,
		Type:     orderType,
		Quantity: quantity,
		Price:    price,
	}
}

func (dc *StandardDemoController) updateLoadTestMetrics(orderNum int, latency time.Duration, err error) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Check if load test is still active
	if dc.loadTestStatus == nil || !dc.loadTestStatus.IsRunning {
		return
	}

	if err != nil {
		dc.loadTestStatus.FailedOrders++
		dc.addLoadTestError("order_failed", err.Error(), fmt.Sprintf("order_%d", orderNum))
	} else {
		dc.loadTestStatus.CompletedOrders++
	}

	// Update current metrics (ensure CurrentMetrics is not nil)
	if dc.loadTestStatus.CurrentMetrics == nil {
		dc.loadTestStatus.CurrentMetrics = &LoadTestMetrics{}
	}
	dc.loadTestStatus.CurrentMetrics.Timestamp = time.Now()
	dc.loadTestStatus.CurrentMetrics.AverageLatency = float64(latency.Milliseconds())

	// Broadcast metrics update
	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateLoadTestStatus,
		Timestamp: time.Now(),
		Data:      dc.loadTestStatus,
		Source:    "load_metrics",
	})
}

func (dc *StandardDemoController) updateChaosTestMetrics() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.chaosTestStatus.Metrics.Timestamp = time.Now()
	// Update chaos-specific metrics here

	// Broadcast metrics update
	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateChaosTestStatus,
		Timestamp: time.Now(),
		Data:      dc.chaosTestStatus,
		Source:    "chaos_metrics",
	})
}

func (dc *StandardDemoController) updateLoadTestProgress(elapsed, duration time.Duration) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if duration > 0 {
		dc.loadTestStatus.Progress = float64(elapsed) / float64(duration) * 100
		dc.loadTestStatus.ElapsedTime = elapsed
		dc.loadTestStatus.RemainingTime = duration - elapsed
	}
}

func (dc *StandardDemoController) updateChaosTestProgress(elapsed, duration time.Duration) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	if duration > 0 {
		dc.chaosTestStatus.Progress = float64(elapsed) / float64(duration) * 100
		dc.chaosTestStatus.ElapsedTime = elapsed
		dc.chaosTestStatus.RemainingTime = duration - elapsed
	}
}

func (dc *StandardDemoController) updateChaosTestPhase(phase ChaosTestPhase) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.chaosTestStatus.Phase = phase
	dc.logger.Info("Chaos test phase updated", "phase", phase)
}

func (dc *StandardDemoController) completeLoadTest() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.loadTestStatus.IsRunning = false
	dc.loadTestStatus.Phase = LoadPhaseCompleted
	dc.loadTestStatus.Progress = 100.0

	dc.removeActiveScenario(dc.loadTestStatus.Scenario.Name)

	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateScenarioComplete,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"type":     "load_test",
			"scenario": dc.loadTestStatus.Scenario.Name,
			"status":   dc.loadTestStatus,
		},
		Source: "load_controller",
	})
}

func (dc *StandardDemoController) recoverFromChaos(scenario ChaosTestScenario) {
	dc.updateChaosTestPhase(ChaosPhaseRecovery)

	// Stop chaos injection
	ctx := context.Background()
	if err := dc.scenarioManager.StopChaosScenario(ctx); err != nil {
		dc.addChaosTestError("recovery_failed", err.Error(), "")
	}

	// Mark as completed
	dc.mu.Lock()
	dc.chaosTestStatus.IsRunning = false
	dc.chaosTestStatus.Phase = ChaosPhaseCompleted
	dc.chaosTestStatus.Progress = 100.0
	dc.removeActiveScenario(scenario.Name)
	dc.mu.Unlock()

	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateScenarioComplete,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"type":     "chaos_test",
			"scenario": scenario.Name,
			"status":   dc.chaosTestStatus,
		},
		Source: "chaos_controller",
	})
}

func (dc *StandardDemoController) addLoadTestError(errorType, message, orderID string) {
	error := LoadTestError{
		Timestamp: time.Now(),
		Type:      errorType,
		Message:   message,
		OrderID:   orderID,
		Severity:  "error",
	}

	dc.loadTestStatus.Errors = append(dc.loadTestStatus.Errors, error)

	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateError,
		Timestamp: time.Now(),
		Data:      error,
		Source:    "load_controller",
	})
}

func (dc *StandardDemoController) addChaosTestError(errorType, message, target string) {
	error := ChaosTestError{
		Timestamp: time.Now(),
		Type:      errorType,
		Message:   message,
		Target:    target,
		Severity:  "error",
	}

	dc.chaosTestStatus.Errors = append(dc.chaosTestStatus.Errors, error)

	dc.broadcastUpdate(DemoUpdate{
		Type:      UpdateError,
		Timestamp: time.Now(),
		Data:      error,
		Source:    "chaos_controller",
	})
}

func (dc *StandardDemoController) removeActiveScenario(scenarioName string) {
	newScenarios := []string{}
	for _, s := range dc.systemStatus.ActiveScenarios {
		if s != scenarioName {
			newScenarios = append(newScenarios, s)
		}
	}
	dc.systemStatus.ActiveScenarios = newScenarios
}

func (dc *StandardDemoController) processUpdates() {
	for {
		select {
		case <-dc.stopUpdates:
			return
		case update := <-dc.updateChan:
			dc.broadcastToSubscribers(update)
		}
	}
}

func (dc *StandardDemoController) broadcastToSubscribers(update DemoUpdate) {
	dc.subscribersMu.RLock()
	defer dc.subscribersMu.RUnlock()

	for id, subscriber := range dc.subscribers {
		if !subscriber.IsActive() {
			continue
		}

		if err := subscriber.SendUpdate(update); err != nil {
			dc.logger.Warn("Failed to send update to subscriber", "id", id, "error", err)
			// Remove inactive subscriber
			go func(subID string) {
				dc.Unsubscribe(subID)
			}(id)
		}
	}
}

func (dc *StandardDemoController) broadcastUpdate(update DemoUpdate) {
	select {
	case dc.updateChan <- update:
	default:
		dc.logger.Warn("Update channel full, dropping update", "type", update.Type)
	}
}

func (dc *StandardDemoController) collectSystemMetrics() {
	ticker := time.NewTicker(dc.config.Metrics.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dc.stopUpdates:
			return
		case <-ticker.C:
			metrics := dc.metricsCollector.CollectSystemMetrics()

			dc.mu.Lock()
			dc.systemStatus.SystemMetrics = metrics
			dc.systemStatus.Timestamp = time.Now()
			dc.mu.Unlock()

			dc.broadcastUpdate(DemoUpdate{
				Type:      UpdateMetrics,
				Timestamp: time.Now(),
				Data:      metrics,
				Source:    "metrics_collector",
			})
		}
	}
}

// Shutdown gracefully shuts down the demo controller
func (dc *StandardDemoController) Shutdown(ctx context.Context) error {
	dc.logger.Info("Shutting down demo controller")

	// Stop all running tests
	dc.ResetSystem(ctx)

	// Stop background processes
	close(dc.stopUpdates)

	// Close all subscribers
	dc.subscribersMu.Lock()
	for _, subscriber := range dc.subscribers {
		subscriber.Close()
	}
	dc.subscribersMu.Unlock()

	dc.isRunning = false
	dc.logger.Info("Demo controller shutdown completed")

	return nil
}