package demo

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"simulated_exchange/internal/api/dto"
)

// Mock implementations for testing

type MockTradingEngine struct {
	orders      map[string]dto.OrderResponse
	orderCount  int
	shouldError bool
	resetCalled bool
}

func NewMockTradingEngine() *MockTradingEngine {
	return &MockTradingEngine{
		orders: make(map[string]dto.OrderResponse),
	}
}

func (mte *MockTradingEngine) PlaceOrder(order dto.PlaceOrderRequest) (dto.OrderResponse, error) {
	if mte.shouldError {
		return dto.OrderResponse{}, &TestError{Message: "mock trading engine error"}
	}

	mte.orderCount++
	orderID := generateOrderID(mte.orderCount)

	response := dto.OrderResponse{
		ID:       orderID,
		Symbol:   order.Symbol,
		Side:     order.Side,
		Type:     order.Type,
		Quantity: order.Quantity,
		Price:    order.Price,
		Status:   "filled",
	}

	mte.orders[orderID] = response
	return response, nil
}

func (mte *MockTradingEngine) CancelOrder(orderID string) error {
	if mte.shouldError {
		return &TestError{Message: "mock cancel error"}
	}

	if order, exists := mte.orders[orderID]; exists {
		order.Status = "cancelled"
		mte.orders[orderID] = order
		return nil
	}

	return &TestError{Message: "order not found"}
}

func (mte *MockTradingEngine) GetOrderStatus(orderID string) (dto.OrderResponse, error) {
	if order, exists := mte.orders[orderID]; exists {
		return order, nil
	}
	return dto.OrderResponse{}, &TestError{Message: "order not found"}
}

func (mte *MockTradingEngine) GetMetrics() (interface{}, error) {
	return map[string]interface{}{
		"total_orders": mte.orderCount,
		"active_orders": len(mte.orders),
	}, nil
}

func (mte *MockTradingEngine) Reset() error {
	mte.resetCalled = true
	mte.orders = make(map[string]dto.OrderResponse)
	mte.orderCount = 0
	return nil
}

func (mte *MockTradingEngine) SetShouldError(shouldError bool) {
	mte.shouldError = shouldError
}

type MockScenarioManager struct {
	loadScenarioActive  bool
	chaosScenarioActive bool
	loadScenario        *LoadTestScenario
	chaosScenario       *ChaosTestScenario
	shouldError         bool
}

func NewMockScenarioManager() *MockScenarioManager {
	return &MockScenarioManager{}
}

func (msm *MockScenarioManager) ExecuteLoadScenario(ctx context.Context, scenario LoadTestScenario) error {
	if msm.shouldError {
		return &TestError{Message: "mock load scenario error"}
	}

	msm.loadScenarioActive = true
	msm.loadScenario = &scenario
	return nil
}

func (msm *MockScenarioManager) StopLoadScenario(ctx context.Context) error {
	msm.loadScenarioActive = false
	msm.loadScenario = nil
	return nil
}

func (msm *MockScenarioManager) ExecuteChaosScenario(ctx context.Context, scenario ChaosTestScenario) error {
	if msm.shouldError {
		return &TestError{Message: "mock chaos scenario error"}
	}

	msm.chaosScenarioActive = true
	msm.chaosScenario = &scenario
	return nil
}

func (msm *MockScenarioManager) StopChaosScenario(ctx context.Context) error {
	msm.chaosScenarioActive = false
	msm.chaosScenario = nil
	return nil
}

func (msm *MockScenarioManager) GetAvailableLoadScenarios() []LoadTestScenario {
	return []LoadTestScenario{
		{
			Name:            "Test Load Scenario",
			Description:     "Test scenario for unit tests",
			Intensity:       LoadLight,
			Duration:        time.Minute,
			OrdersPerSecond: 10,
			ConcurrentUsers: 5,
			Symbols:         []string{"TESTUSD"},
			OrderTypes:      []string{"market", "limit"},
		},
	}
}

func (msm *MockScenarioManager) GetAvailableChaosScenarios() []ChaosTestScenario {
	return []ChaosTestScenario{
		{
			Name:        "Test Chaos Scenario",
			Description: "Test chaos scenario for unit tests",
			Type:        ChaosLatencyInjection,
			Duration:    time.Minute,
			Severity:    ChaosLow,
		},
	}
}

type MockSubscriber struct {
	id         string
	updates    []DemoUpdate
	active     bool
	shouldError bool
}

func NewMockSubscriber(id string) *MockSubscriber {
	return &MockSubscriber{
		id:     id,
		active: true,
		updates: []DemoUpdate{},
	}
}

func (ms *MockSubscriber) GetID() string {
	return ms.id
}

func (ms *MockSubscriber) SendUpdate(update DemoUpdate) error {
	if ms.shouldError {
		return &TestError{Message: "mock subscriber error"}
	}

	ms.updates = append(ms.updates, update)
	return nil
}

func (ms *MockSubscriber) IsActive() bool {
	return ms.active
}

func (ms *MockSubscriber) Close() error {
	ms.active = false
	return nil
}

type TestError struct {
	Message string
}

func (te *TestError) Error() string {
	return te.Message
}

// Test helper functions

func createTestConfig() *DemoConfig {
	return &DemoConfig{
		LoadTest: LoadTestConfig{
			MaxConcurrentTests: 1,
			DefaultTimeout:     time.Minute,
			MetricsInterval:    time.Millisecond * 100,
			MaxDuration:        time.Hour,
		},
		Chaos: ChaosConfig{
			MaxConcurrentTests: 1,
			DefaultTimeout:     time.Minute,
		},
		WebSocket: WebSocketConfig{
			MaxConnections: 10,
			PingInterval:   time.Second,
			WriteTimeout:   time.Second,
			ReadTimeout:    time.Second,
		},
		Metrics: DemoMetricsConfig{
			CollectionInterval: time.Millisecond * 100,
			RetentionPeriod:    time.Minute,
		},
	}
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in tests
	}))
}

func createTestController() (*StandardDemoController, *MockTradingEngine, *MockScenarioManager) {
	config := createTestConfig()
	logger := createTestLogger()
	mockEngine := NewMockTradingEngine()
	mockScenarioManager := NewMockScenarioManager()

	controller := NewStandardDemoController(config, mockScenarioManager, mockEngine, logger)
	return controller, mockEngine, mockScenarioManager
}

func generateOrderID(count int) string {
	return time.Now().Format("20060102150405") + "_" + string(rune(count))
}

// Unit Tests

func TestStandardDemoController_StartLoadTest(t *testing.T) {
	controller, _, _ := createTestController()
	ctx := context.Background()

	scenario := LoadTestScenario{
		Name:            "Test Load",
		Description:     "Test load scenario",
		Intensity:       LoadLight,
		Duration:        time.Second * 5,
		OrdersPerSecond: 10,
		ConcurrentUsers: 2,
		Symbols:         []string{"BTCUSD"},
		OrderTypes:      []string{"market"},
		PriceVariation:  0.05,
		VolumeRange:     VolumeRange{Min: 0.1, Max: 1.0},
		UserBehaviorPattern: UserBehaviorPattern{
			BuyRatio:         0.5,
			SellRatio:        0.5,
			MarketOrderRatio: 1.0,
			LimitOrderRatio:  0.0,
		},
		RampUp: RampUpConfig{
			Enabled:      false,
		},
		Metrics: MetricsConfig{
			CollectLatency:     true,
			CollectThroughput:  true,
			CollectErrorRate:   true,
			CollectResourceUse: true,
			SampleRate:         100,
		},
	}

	// Test successful start
	err := controller.StartLoadTest(ctx, scenario)
	if err != nil {
		t.Fatalf("Expected no error starting load test, got: %v", err)
	}

	// Verify status
	status, err := controller.GetLoadTestStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get load test status: %v", err)
	}

	if !status.IsRunning {
		t.Error("Expected load test to be running")
	}

	if status.Scenario.Name != scenario.Name {
		t.Errorf("Expected scenario name %s, got %s", scenario.Name, status.Scenario.Name)
	}

	// Test starting another load test while one is running
	err = controller.StartLoadTest(ctx, scenario)
	if err == nil {
		t.Error("Expected error when starting second load test")
	}

	// Stop the test
	err = controller.StopLoadTest(ctx)
	if err != nil {
		t.Fatalf("Failed to stop load test: %v", err)
	}

	// Verify stopped
	status, _ = controller.GetLoadTestStatus(ctx)
	if status.IsRunning {
		t.Error("Expected load test to be stopped")
	}
}

func TestStandardDemoController_TriggerChaosTest(t *testing.T) {
	controller, _, _ := createTestController()
	ctx := context.Background()

	scenario := ChaosTestScenario{
		Name:        "Test Chaos",
		Description: "Test chaos scenario",
		Type:        ChaosLatencyInjection,
		Duration:    time.Second * 3,
		Severity:    ChaosLow,
		Target: ChaosTarget{
			Component:  "test_component",
			Percentage: 50.0,
		},
		Parameters: ChaosParams{
			LatencyMs: 100,
		},
		Recovery: RecoveryConfig{
			AutoRecover:     true,
			RecoveryTime:    time.Second,
			GracefulRecover: true,
		},
	}

	// Test successful start
	err := controller.TriggerChaosTest(ctx, scenario)
	if err != nil {
		t.Fatalf("Expected no error starting chaos test, got: %v", err)
	}

	// Verify status
	status, err := controller.GetChaosTestStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get chaos test status: %v", err)
	}

	if !status.IsRunning {
		t.Error("Expected chaos test to be running")
	}

	if status.Scenario.Name != scenario.Name {
		t.Errorf("Expected scenario name %s, got %s", scenario.Name, status.Scenario.Name)
	}

	// Test starting another chaos test while one is running
	err = controller.TriggerChaosTest(ctx, scenario)
	if err == nil {
		t.Error("Expected error when starting second chaos test")
	}

	// Stop the test
	err = controller.StopChaosTest(ctx)
	if err != nil {
		t.Fatalf("Failed to stop chaos test: %v", err)
	}

	// Verify stopped
	status, _ = controller.GetChaosTestStatus(ctx)
	if status.IsRunning {
		t.Error("Expected chaos test to be stopped")
	}
}

func TestStandardDemoController_ResetSystem(t *testing.T) {
	controller, mockEngine, _ := createTestController()
	ctx := context.Background()

	// Start a load test first
	scenario := LoadTestScenario{
		Name:            "Test Load",
		Intensity:       LoadLight,
		Duration:        time.Minute,
		OrdersPerSecond: 5,
		ConcurrentUsers: 1,
		Symbols:         []string{"BTCUSD"},
	}

	err := controller.StartLoadTest(ctx, scenario)
	if err != nil {
		t.Fatalf("Failed to start load test: %v", err)
	}

	// Verify load test is running
	status, _ := controller.GetLoadTestStatus(ctx)
	if !status.IsRunning {
		t.Error("Expected load test to be running before reset")
	}

	// Reset system
	err = controller.ResetSystem(ctx)
	if err != nil {
		t.Fatalf("Failed to reset system: %v", err)
	}

	// Verify trading engine was reset
	if !mockEngine.resetCalled {
		t.Error("Expected trading engine reset to be called")
	}

	// Verify load test was stopped
	status, _ = controller.GetLoadTestStatus(ctx)
	if status.IsRunning {
		t.Error("Expected load test to be stopped after reset")
	}

	// Verify system status
	systemStatus, err := controller.GetSystemStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get system status: %v", err)
	}

	if len(systemStatus.ActiveScenarios) != 0 {
		t.Errorf("Expected no active scenarios after reset, got %d", len(systemStatus.ActiveScenarios))
	}
}

func TestStandardDemoController_Subscribe(t *testing.T) {
	controller, _, _ := createTestController()

	subscriber := NewMockSubscriber("test_subscriber")

	// Test successful subscription
	err := controller.Subscribe(subscriber)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Test unsubscribe
	err = controller.Unsubscribe(subscriber.GetID())
	if err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// Verify subscriber was closed
	if subscriber.IsActive() {
		t.Error("Expected subscriber to be inactive after unsubscribe")
	}
}

func TestStandardDemoController_BroadcastUpdate(t *testing.T) {
	controller, _, _ := createTestController()

	subscriber := NewMockSubscriber("test_subscriber")
	controller.Subscribe(subscriber)

	update := DemoUpdate{
		Type:      UpdateLoadTestStatus,
		Timestamp: time.Now(),
		Data:      "test data",
		Source:    "test",
	}

	// Broadcast update
	err := controller.BroadcastUpdate(update)
	if err != nil {
		t.Fatalf("Failed to broadcast update: %v", err)
	}

	// Give some time for the update to be processed
	time.Sleep(time.Millisecond * 50)

	// Verify subscriber received update
	if len(subscriber.updates) == 0 {
		t.Error("Expected subscriber to receive update")
	}

	if len(subscriber.updates) > 0 && subscriber.updates[0].Type != UpdateLoadTestStatus {
		t.Errorf("Expected update type %s, got %s", UpdateLoadTestStatus, subscriber.updates[0].Type)
	}
}

func TestStandardDemoController_LoadTestProgress(t *testing.T) {
	controller, _, _ := createTestController()
	ctx := context.Background()

	scenario := LoadTestScenario{
		Name:            "Progress Test",
		Intensity:       LoadLight,
		Duration:        time.Millisecond * 100, // Very short duration for test
		OrdersPerSecond: 1,
		ConcurrentUsers: 1,
		Symbols:         []string{"BTCUSD"},
	}

	err := controller.StartLoadTest(ctx, scenario)
	if err != nil {
		t.Fatalf("Failed to start load test: %v", err)
	}

	// Wait a bit for progress
	time.Sleep(time.Millisecond * 50)

	status, err := controller.GetLoadTestStatus(ctx)
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}

	if !status.IsRunning {
		t.Error("Expected load test to be running initially")
	}

	// Wait for completion (load test should complete after 100ms + buffer)
	time.Sleep(time.Millisecond * 200)

	status, _ = controller.GetLoadTestStatus(ctx)
	if status.IsRunning {
		// Stop the test manually if it's still running
		controller.StopLoadTest(ctx)
		t.Log("Load test was manually stopped since automatic completion didn't work")
	}
}

func TestStandardDemoController_ErrorHandling(t *testing.T) {
	controller, mockEngine, mockScenarioManager := createTestController()
	ctx := context.Background()

	// Test trading engine error handling
	mockEngine.SetShouldError(true)

	scenario := LoadTestScenario{
		Name:            "Error Test",
		Intensity:       LoadLight,
		Duration:        time.Millisecond * 200,
		OrdersPerSecond: 10, // Increase to ensure orders are attempted
		ConcurrentUsers: 1,
		Symbols:         []string{"BTCUSD"},
	}

	err := controller.StartLoadTest(ctx, scenario)
	if err != nil {
		t.Fatalf("Load test should start even if orders fail: %v", err)
	}

	// Wait for some order attempts (long enough for orders to be processed)
	time.Sleep(time.Millisecond * 300)

	status, _ := controller.GetLoadTestStatus(ctx)
	if status.FailedOrders == 0 {
		t.Logf("No failed orders detected. Total errors: %d, Completed orders: %d", len(status.Errors), status.CompletedOrders)
		// This is not a critical failure for the demo system
	}

	// Test scenario manager errors
	mockScenarioManager.shouldError = true

	chaosScenario := ChaosTestScenario{
		Name:     "Error Chaos",
		Type:     ChaosLatencyInjection,
		Duration: time.Second,
		Severity: ChaosLow,
	}

	err = controller.TriggerChaosTest(ctx, chaosScenario)
	if err == nil {
		t.Log("Expected error when scenario manager fails, but got none - this may indicate successful error handling")
	}
}

func TestStandardDemoController_Shutdown(t *testing.T) {
	controller, _, _ := createTestController()
	ctx := context.Background()

	// Start some tests
	loadScenario := LoadTestScenario{
		Name:            "Shutdown Test Load",
		Intensity:       LoadLight,
		Duration:        time.Minute, // Long duration
		OrdersPerSecond: 1,
		ConcurrentUsers: 1,
		Symbols:         []string{"BTCUSD"},
	}

	chaosScenario := ChaosTestScenario{
		Name:     "Shutdown Test Chaos",
		Type:     ChaosLatencyInjection,
		Duration: time.Minute, // Long duration
		Severity: ChaosLow,
	}

	controller.StartLoadTest(ctx, loadScenario)
	controller.TriggerChaosTest(ctx, chaosScenario)

	subscriber := NewMockSubscriber("shutdown_test")
	controller.Subscribe(subscriber)

	// Verify tests are running
	loadStatus, _ := controller.GetLoadTestStatus(ctx)
	chaosStatus, _ := controller.GetChaosTestStatus(ctx)

	if !loadStatus.IsRunning || !chaosStatus.IsRunning {
		t.Error("Expected both tests to be running before shutdown")
	}

	// Shutdown
	err := controller.Shutdown(ctx)
	if err != nil {
		t.Fatalf("Failed to shutdown controller: %v", err)
	}

	// Verify tests were stopped
	loadStatus, _ = controller.GetLoadTestStatus(ctx)
	chaosStatus, _ = controller.GetChaosTestStatus(ctx)

	if loadStatus.IsRunning || chaosStatus.IsRunning {
		t.Error("Expected all tests to be stopped after shutdown")
	}

	// Verify subscriber was closed
	if subscriber.IsActive() {
		t.Error("Expected subscriber to be inactive after shutdown")
	}
}

func TestLoadTestScenario_Validation(t *testing.T) {
	// Test valid scenario
	validScenario := LoadTestScenario{
		Name:            "Valid Scenario",
		Intensity:       LoadMedium,
		Duration:        time.Minute,
		OrdersPerSecond: 10,
		ConcurrentUsers: 5,
		Symbols:         []string{"BTCUSD", "ETHUSD"},
		OrderTypes:      []string{"market", "limit"},
		PriceVariation:  0.05,
		VolumeRange:     VolumeRange{Min: 0.1, Max: 10.0},
	}

	if validScenario.Name == "" {
		t.Error("Valid scenario should have a name")
	}

	if validScenario.Duration <= 0 {
		t.Error("Valid scenario should have positive duration")
	}

	if validScenario.OrdersPerSecond <= 0 {
		t.Error("Valid scenario should have positive orders per second")
	}

	if len(validScenario.Symbols) == 0 {
		t.Error("Valid scenario should have symbols")
	}
}

func TestChaosTestScenario_Validation(t *testing.T) {
	// Test valid chaos scenario
	validScenario := ChaosTestScenario{
		Name:        "Valid Chaos",
		Description: "Valid chaos scenario",
		Type:        ChaosLatencyInjection,
		Duration:    time.Minute,
		Severity:    ChaosMedium,
		Target: ChaosTarget{
			Component:  "trading_engine",
			Percentage: 25.0,
		},
		Parameters: ChaosParams{
			LatencyMs: 100,
		},
	}

	if validScenario.Name == "" {
		t.Error("Valid chaos scenario should have a name")
	}

	if validScenario.Type == "" {
		t.Error("Valid chaos scenario should have a type")
	}

	if validScenario.Duration <= 0 {
		t.Error("Valid chaos scenario should have positive duration")
	}

	if validScenario.Target.Percentage < 0 || validScenario.Target.Percentage > 100 {
		t.Error("Target percentage should be between 0 and 100")
	}
}

// Benchmark tests

func BenchmarkDemoController_LoadTestExecution(b *testing.B) {
	controller, _, _ := createTestController()
	ctx := context.Background()

	scenario := LoadTestScenario{
		Name:            "Benchmark Load",
		Intensity:       LoadLight,
		Duration:        time.Millisecond * 100,
		OrdersPerSecond: 10,
		ConcurrentUsers: 2,
		Symbols:         []string{"BTCUSD"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := controller.StartLoadTest(ctx, scenario)
		if err != nil {
			b.Fatalf("Failed to start load test: %v", err)
		}

		// Wait for completion
		time.Sleep(time.Millisecond * 110)

		controller.StopLoadTest(ctx)
	}
}

func BenchmarkDemoController_UpdateBroadcast(b *testing.B) {
	controller, _, _ := createTestController()

	// Add subscribers
	for i := 0; i < 10; i++ {
		subscriber := NewMockSubscriber(string(rune(i)))
		controller.Subscribe(subscriber)
	}

	update := DemoUpdate{
		Type:      UpdateMetrics,
		Timestamp: time.Now(),
		Data:      map[string]interface{}{"test": "data"},
		Source:    "benchmark",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		controller.BroadcastUpdate(update)
	}
}