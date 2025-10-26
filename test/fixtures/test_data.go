package fixtures

import (
	"time"

	"simulated_exchange/internal/api/dto"
	"simulated_exchange/internal/demo"
)

// TestData provides comprehensive test data for E2E testing
type TestData struct {
	Orders          []dto.PlaceOrderRequest
	LoadScenarios   []demo.LoadTestScenario
	ChaosScenarios  []demo.ChaosTestScenario
	Symbols         []string
	TestUsers       []TestUser
}

// TestUser represents a test user configuration
type TestUser struct {
	ID              string
	Name            string
	TradingStrategy string
	OrdersPerMinute int
	MaxPosition     float64
}

// GetTestData returns comprehensive test data for E2E testing
func GetTestData() *TestData {
	return &TestData{
		Orders:          GetTestOrders(),
		LoadScenarios:   GetLoadTestScenarios(),
		ChaosScenarios:  GetChaosTestScenarios(),
		Symbols:         GetTestSymbols(),
		TestUsers:       GetTestUsers(),
	}
}

// GetTestOrders returns various order configurations for testing
func GetTestOrders() []dto.PlaceOrderRequest {
	return []dto.PlaceOrderRequest{
		// Market orders
		{
			Symbol:   "BTCUSD",
			Side:     "buy",
			Type:     "market",
			Quantity: 1.0,
			Price:    50000.0,
		},
		{
			Symbol:   "ETHUSD",
			Side:     "sell",
			Type:     "market",
			Quantity: 10.0,
			Price:    3000.0,
		},
		// Limit orders
		{
			Symbol:   "BTCUSD",
			Side:     "buy",
			Type:     "limit",
			Quantity: 0.5,
			Price:    49000.0,
		},
		{
			Symbol:   "ETHUSD",
			Side:     "sell",
			Type:     "limit",
			Quantity: 5.0,
			Price:    3100.0,
		},
		// Large orders for stress testing
		{
			Symbol:   "BTCUSD",
			Side:     "buy",
			Type:     "limit",
			Quantity: 100.0,
			Price:    48000.0,
		},
		{
			Symbol:   "ADAUSD",
			Side:     "sell",
			Type:     "market",
			Quantity: 10000.0,
			Price:    0.50,
		},
		// Edge case orders
		{
			Symbol:   "DOTUSD",
			Side:     "buy",
			Type:     "limit",
			Quantity: 0.001,
			Price:    25.0,
		},
		{
			Symbol:   "SOLUSD",
			Side:     "sell",
			Type:     "limit",
			Quantity: 999.999,
			Price:    100.0,
		},
	}
}

// GetLoadTestScenarios returns predefined load test scenarios
func GetLoadTestScenarios() []demo.LoadTestScenario {
	return []demo.LoadTestScenario{
		{
			Name:            "E2E Light Load",
			Description:     "Light load test for E2E validation",
			Intensity:       demo.LoadLight,
			Duration:        30 * time.Second,
			OrdersPerSecond: 5,
			ConcurrentUsers: 10,
			Symbols:         []string{"BTCUSD", "ETHUSD"},
			OrderTypes:      []string{"market", "limit"},
		},
		{
			Name:            "E2E Medium Load",
			Description:     "Medium load test for performance validation",
			Intensity:       demo.LoadMedium,
			Duration:        60 * time.Second,
			OrdersPerSecond: 25,
			ConcurrentUsers: 50,
			Symbols:         []string{"BTCUSD", "ETHUSD", "ADAUSD", "DOTUSD"},
			OrderTypes:      []string{"market", "limit"},
		},
		{
			Name:            "E2E Heavy Load",
			Description:     "Heavy load test for stress validation",
			Intensity:       demo.LoadHeavy,
			Duration:        120 * time.Second,
			OrdersPerSecond: 100,
			ConcurrentUsers: 200,
			Symbols:         []string{"BTCUSD", "ETHUSD", "ADAUSD", "DOTUSD", "SOLUSD"},
			OrderTypes:      []string{"market", "limit"},
		},
		{
			Name:            "E2E Demo Presentation",
			Description:     "Demo scenario for live presentation",
			Intensity:       demo.LoadLight,
			Duration:        5 * time.Minute,
			OrdersPerSecond: 10,
			ConcurrentUsers: 25,
			Symbols:         []string{"BTCUSD", "ETHUSD", "ADAUSD"},
			OrderTypes:      []string{"market", "limit"},
		},
	}
}

// GetChaosTestScenarios returns predefined chaos test scenarios
func GetChaosTestScenarios() []demo.ChaosTestScenario {
	return []demo.ChaosTestScenario{
		{
			Name:        "E2E Latency Injection",
			Description: "Inject network latency for resilience testing",
			Type:        demo.ChaosLatencyInjection,
			Duration:    30 * time.Second,
			Severity:    demo.ChaosLow,
			Parameters: demo.ChaosParams{
				LatencyMs: 100,
			},
		},
		{
			Name:        "E2E Error Simulation",
			Description: "Simulate random errors for fault tolerance testing",
			Type:        demo.ChaosErrorSimulation,
			Duration:    45 * time.Second,
			Severity:    demo.ChaosMedium,
			Parameters: demo.ChaosParams{
				ErrorRate: 0.15,
			},
		},
		{
			Name:        "E2E Resource Exhaustion",
			Description: "Simulate resource constraints",
			Type:        demo.ChaosResourceExhaustion,
			Duration:    60 * time.Second,
			Severity:    demo.ChaosHigh,
			Parameters: demo.ChaosParams{
				CPULimitPercent: 80.0,
				MemoryLimitMB:   768,
			},
		},
		{
			Name:        "E2E Network Partition",
			Description: "Simulate network partition scenarios",
			Type:        demo.ChaosNetworkPartition,
			Duration:    90 * time.Second,
			Severity:    demo.ChaosHigh,
			Parameters: demo.ChaosParams{
				NetworkDelayMs:    100,
				PacketLossPercent: 20.0,
			},
		},
	}
}

// GetTestSymbols returns trading symbols for testing
func GetTestSymbols() []string {
	return []string{
		"BTCUSD",
		"ETHUSD",
		"ADAUSD",
		"DOTUSD",
		"SOLUSD",
		"LINKUSD",
		"MATICUSD",
		"AVAXUSD",
	}
}

// GetTestUsers returns test user configurations
func GetTestUsers() []TestUser {
	return []TestUser{
		{
			ID:              "user_001",
			Name:            "Conservative Trader",
			TradingStrategy: "conservative",
			OrdersPerMinute: 2,
			MaxPosition:     10000.0,
		},
		{
			ID:              "user_002",
			Name:            "Aggressive Trader",
			TradingStrategy: "aggressive",
			OrdersPerMinute: 15,
			MaxPosition:     100000.0,
		},
		{
			ID:              "user_003",
			Name:            "High Frequency Trader",
			TradingStrategy: "hft",
			OrdersPerMinute: 60,
			MaxPosition:     1000000.0,
		},
		{
			ID:              "user_004",
			Name:            "Momentum Trader",
			TradingStrategy: "momentum",
			OrdersPerMinute: 8,
			MaxPosition:     50000.0,
		},
		{
			ID:              "user_005",
			Name:            "Arbitrage Trader",
			TradingStrategy: "arbitrage",
			OrdersPerMinute: 25,
			MaxPosition:     200000.0,
		},
	}
}

// Performance test data configurations
type PerformanceTestConfig struct {
	MaxLatencyMs     int
	MinThroughput    float64
	MaxErrorRate     float64
	TargetCPUUsage   float64
	TargetMemoryUsage float64
}

// GetPerformanceTargets returns performance benchmark targets
func GetPerformanceTargets() PerformanceTestConfig {
	return PerformanceTestConfig{
		MaxLatencyMs:      100,
		MinThroughput:     1000.0, // orders per second
		MaxErrorRate:      0.01,   // 1%
		TargetCPUUsage:    70.0,   // 70%
		TargetMemoryUsage: 80.0,   // 80%
	}
}

// Test environment configurations
type TestEnvironment struct {
	Name            string
	DatabaseURL     string
	APIPort         int
	MetricsPort     int
	DemoWSPort      int
	LogLevel        string
	TestTimeout     time.Duration
}

// GetTestEnvironments returns different test environment configurations
func GetTestEnvironments() map[string]TestEnvironment {
	return map[string]TestEnvironment{
		"local": {
			Name:        "Local Development",
			DatabaseURL: "memory://test.db",
			APIPort:     8080,
			MetricsPort: 9090,
			DemoWSPort:  8081,
			LogLevel:    "debug",
			TestTimeout: 5 * time.Minute,
		},
		"ci": {
			Name:        "CI/CD Pipeline",
			DatabaseURL: "memory://ci_test.db",
			APIPort:     18080,
			MetricsPort: 19090,
			DemoWSPort:  18081,
			LogLevel:    "info",
			TestTimeout: 10 * time.Minute,
		},
		"docker": {
			Name:        "Docker Container",
			DatabaseURL: "memory://docker_test.db",
			APIPort:     8080,
			MetricsPort: 9090,
			DemoWSPort:  8081,
			LogLevel:    "warn",
			TestTimeout: 15 * time.Minute,
		},
	}
}

// ExpectedResults defines expected outcomes for various test scenarios
type ExpectedResults struct {
	OrderPlacement  OrderPlacementExpectations
	LoadTesting     LoadTestingExpectations
	ChaosTesting    ChaosTestingExpectations
	Performance     PerformanceExpectations
}

type OrderPlacementExpectations struct {
	SuccessRate        float64
	MaxLatencyMs       int
	OrderBookUpdates   bool
	MetricsRecording   bool
}

type LoadTestingExpectations struct {
	CompletionRate     float64
	ThroughputAccuracy float64
	ProgressTracking   bool
	RealTimeUpdates    bool
}

type ChaosTestingExpectations struct {
	RecoveryTime       time.Duration
	DataConsistency    bool
	AlertGeneration    bool
	SystemStability    bool
}

type PerformanceExpectations struct {
	LatencyP99         time.Duration
	ThroughputTarget   float64
	ErrorRateMax       float64
	ResourceUsageMax   float64
}

// GetExpectedResults returns expected test outcomes
func GetExpectedResults() ExpectedResults {
	return ExpectedResults{
		OrderPlacement: OrderPlacementExpectations{
			SuccessRate:      0.99,
			MaxLatencyMs:     50,
			OrderBookUpdates: true,
			MetricsRecording: true,
		},
		LoadTesting: LoadTestingExpectations{
			CompletionRate:     0.95,
			ThroughputAccuracy: 0.90,
			ProgressTracking:   true,
			RealTimeUpdates:    true,
		},
		ChaosTesting: ChaosTestingExpectations{
			RecoveryTime:    30 * time.Second,
			DataConsistency: true,
			AlertGeneration: true,
			SystemStability: true,
		},
		Performance: PerformanceExpectations{
			LatencyP99:       100 * time.Millisecond,
			ThroughputTarget: 1000.0,
			ErrorRateMax:     0.01,
			ResourceUsageMax: 80.0,
		},
	}
}