package e2e

import (
	"context"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"simulated_exchange/internal/api/dto"
	"simulated_exchange/test/fixtures"
	"simulated_exchange/test/helpers"
)

// PerformanceTestSuite tests performance benchmarks and load characteristics
type PerformanceTestSuite struct {
	suite.Suite
	server         *helpers.TestServer
	httpClient     *helpers.HTTPClient
	loadRunner     *helpers.LoadTestRunner
	testData       *fixtures.TestData
	perfTargets    fixtures.PerformanceTestConfig
	assertions     *helpers.AssertionHelpers
}

// SetupSuite initializes the performance test suite
func (suite *PerformanceTestSuite) SetupSuite() {
	suite.server = helpers.NewTestServer(suite.T())
	suite.httpClient = helpers.NewHTTPClient(suite.T())
	suite.loadRunner = helpers.NewLoadTestRunner(suite.T(), suite.server.GetURL(""))
	suite.testData = fixtures.GetTestData()
	suite.perfTargets = fixtures.GetPerformanceTargets()
	suite.assertions = helpers.NewAssertionHelpers(suite.T())

	// Start the demo system
	ctx := context.Background()
	err := suite.server.DemoSystem.Start(ctx)
	require.NoError(suite.T(), err)

	// Wait for system to be ready
	helpers.WaitForCondition(suite.T(), func() bool {
		health, err := suite.httpClient.GetHealth(suite.server.GetURL(""))
		return err == nil && health.Status == "healthy"
	}, 30*time.Second, "System should be healthy")
}

// TearDownSuite cleans up after performance test suite
func (suite *PerformanceTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// SetupTest resets system state before each performance test
func (suite *PerformanceTestSuite) SetupTest() {
	ctx := context.Background()
	err := suite.server.DemoSystem.Controller.ResetSystem(ctx)
	require.NoError(suite.T(), err)
	time.Sleep(100 * time.Millisecond)
}

// TestOrderPlacementPerformance tests order placement performance under various loads
func (suite *PerformanceTestSuite) TestOrderPlacementPerformance() {
	suite.T().Run("SingleOrderLatency", func(t *testing.T) {
		order := suite.testData.Orders[0]
		iterations := 10
		latencies := make([]time.Duration, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			resp, err := suite.httpClient.PlaceOrder(suite.server.GetURL(""), order)
			latency := time.Since(start)

			require.NoError(t, err)
			require.True(t, resp.Success, "Order should succeed")
			latencies[i] = latency
		}

		// Calculate average latency
		var totalLatency time.Duration
		for _, lat := range latencies {
			totalLatency += lat
		}
		avgLatency := totalLatency / time.Duration(iterations)

		// Assert performance targets
		assert.Less(t, avgLatency, time.Duration(suite.perfTargets.MaxLatencyMs)*time.Millisecond,
			"Average latency should be under %dms, got %v", suite.perfTargets.MaxLatencyMs, avgLatency)

		t.Logf("Single order average latency: %v", avgLatency)
	})

	suite.T().Run("ConcurrentOrderThroughput", func(t *testing.T) {
		// Test different concurrency levels
		concurrencyLevels := []int{10, 25, 50, 100}
		ordersPerLevel := 100

		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
				// Generate orders
				orders := make([]dto.PlaceOrderRequest, ordersPerLevel)
				for i := 0; i < ordersPerLevel; i++ {
					orders[i] = helpers.RandomOrder(suite.testData.Symbols)
				}

				// Run concurrent orders
				results, err := suite.loadRunner.RunConcurrentOrders(orders, concurrency)
				require.NoError(t, err)

				// Assert performance targets
				suite.assertions.AssertPerformanceTargets(results, suite.perfTargets)

				t.Logf("Concurrency %d: Throughput=%.2f ops/s, Success Rate=%.2f%%, P95 Latency=%v",
					concurrency, results.GetThroughput(), results.GetSuccessRate()*100,
					results.GetLatencyPercentile(95))
			})
		}
	})

	suite.T().Run("SustainedLoad", func(t *testing.T) {
		// Run sustained load for longer period
		duration := 30 * time.Second
		targetTPS := 50 // transactions per second
		concurrency := 20

		// Calculate total orders needed
		totalOrders := int(duration.Seconds()) * targetTPS
		orders := make([]dto.PlaceOrderRequest, totalOrders)
		for i := 0; i < totalOrders; i++ {
			orders[i] = helpers.RandomOrder(suite.testData.Symbols)
		}

		start := time.Now()
		results, err := suite.loadRunner.RunConcurrentOrders(orders, concurrency)
		actualDuration := time.Since(start)

		require.NoError(t, err)

		// Verify sustained performance
		actualTPS := results.GetThroughput()
		assert.Greater(t, actualTPS, float64(targetTPS)*0.8, "Should achieve at least 80%% of target TPS")
		assert.Less(t, actualDuration, duration*2, "Should not take more than 2x expected time")

		// Assert performance targets
		suite.assertions.AssertPerformanceTargets(results, suite.perfTargets)

		t.Logf("Sustained load: Target=%d TPS, Actual=%.2f TPS, Duration=%v, Success Rate=%.2f%%",
			targetTPS, actualTPS, actualDuration, results.GetSuccessRate()*100)
	})
}

// TestDemoPerformance tests demo system performance characteristics
func (suite *PerformanceTestSuite) TestDemoPerformance() {
	ctx := context.Background()

	suite.T().Run("LoadTestScenarioPerformance", func(t *testing.T) {
		scenario := suite.testData.LoadScenarios[1] // Medium load scenario
		scenario.Duration = 45 * time.Second // Adjust for performance test

		// Start load test and measure performance
		start := time.Now()
		err := suite.server.DemoSystem.Controller.StartLoadTest(ctx, scenario)
		require.NoError(t, err)

		// Monitor performance during load test
		performanceData := suite.monitorPerformanceDuringLoadTest(ctx, scenario.Duration)

		// Wait for completion
		helpers.WaitForCondition(suite.T(), func() bool {
			status, err := suite.server.DemoSystem.Controller.GetLoadTestStatus(ctx)
			return err == nil && !status.IsRunning
		}, scenario.Duration+30*time.Second, "Load test should complete")

		duration := time.Since(start)

		// Analyze performance data
		avgLatency := suite.calculateAverageLatency(performanceData.Latencies)
		maxLatency := suite.calculateMaxLatency(performanceData.Latencies)

		// Assert performance targets
		assert.Less(t, avgLatency, time.Duration(suite.perfTargets.MaxLatencyMs)*time.Millisecond,
			"Average latency during load test should be acceptable")
		assert.Less(t, maxLatency, time.Duration(suite.perfTargets.MaxLatencyMs*3)*time.Millisecond,
			"Max latency should be reasonable")

		t.Logf("Load test performance: Duration=%v, Avg Latency=%v, Max Latency=%v, Samples=%d",
			duration, avgLatency, maxLatency, len(performanceData.Latencies))
	})

	suite.T().Run("ChaosTestPerformance", func(t *testing.T) {
		scenario := suite.testData.ChaosScenarios[0] // Latency injection
		scenario.Duration = 30 * time.Second

		// Measure baseline performance
		baselineLatency := suite.measureBaselineLatency()

		// Start chaos test
		err := suite.server.DemoSystem.Controller.TriggerChaosTest(ctx, scenario)
		require.NoError(t, err)

		// Measure performance during chaos
		chaosLatency := suite.measureLatencyDuringChaos(15 * time.Second)

		// Stop chaos test
		err = suite.server.DemoSystem.Controller.StopChaosTest(ctx)
		require.NoError(t, err)

		// Measure recovery performance
		time.Sleep(5 * time.Second) // Allow recovery
		recoveryLatency := suite.measureBaselineLatency()

		// Analyze chaos impact
		chaosImpact := float64(chaosLatency) / float64(baselineLatency)
		recoveryFactor := float64(recoveryLatency) / float64(baselineLatency)

		// Assert chaos engineering goals
		assert.Less(t, chaosImpact, 5.0, "Chaos should not increase latency more than 5x")
		assert.Less(t, recoveryFactor, 1.5, "Recovery should bring latency within 1.5x of baseline")

		t.Logf("Chaos performance: Baseline=%v, During Chaos=%v (%.2fx), Recovery=%v (%.2fx)",
			baselineLatency, chaosLatency, chaosImpact, recoveryLatency, recoveryFactor)
	})
}

// TestScalabilityCharacteristics tests system scalability under increasing load
func (suite *PerformanceTestSuite) TestScalabilityCharacteristics() {
	suite.T().Run("VerticalScaling", func(t *testing.T) {
		// Test how performance scales with increasing load on single instance
		loadLevels := []struct {
			name        string
			ordersCount int
			concurrency int
		}{
			{"Light", 50, 5},
			{"Medium", 200, 15},
			{"Heavy", 500, 30},
			{"Stress", 1000, 50},
		}

		results := make(map[string]*helpers.LoadTestResults)

		for _, level := range loadLevels {
			t.Run(level.name, func(t *testing.T) {
				orders := make([]dto.PlaceOrderRequest, level.ordersCount)
				for i := 0; i < level.ordersCount; i++ {
					orders[i] = helpers.RandomOrder(suite.testData.Symbols)
				}

				result, err := suite.loadRunner.RunConcurrentOrders(orders, level.concurrency)
				require.NoError(t, err)

				results[level.name] = result

				t.Logf("%s Load: %d orders, %.2f ops/s, %.2f%% success, P95 latency=%v",
					level.name, level.ordersCount, result.GetThroughput(),
					result.GetSuccessRate()*100, result.GetLatencyPercentile(95))
			})
		}

		// Analyze scalability characteristics
		suite.analyzeScalabilityTrends(results)
	})

	suite.T().Run("MemoryUsageUnderLoad", func(t *testing.T) {
		// Monitor memory usage during increasing load
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		initialMemory := memStats.Alloc

		// Run increasing load and monitor memory
		loadSizes := []int{100, 300, 500, 800, 1000}
		memoryUsage := make([]uint64, len(loadSizes))

		for i, size := range loadSizes {
			orders := make([]dto.PlaceOrderRequest, size)
			for j := 0; j < size; j++ {
				orders[j] = helpers.RandomOrder(suite.testData.Symbols)
			}

			_, err := suite.loadRunner.RunConcurrentOrders(orders, 25)
			require.NoError(t, err)

			runtime.ReadMemStats(&memStats)
			memoryUsage[i] = memStats.Alloc

			t.Logf("Load size %d: Memory usage %d KB", size, (memStats.Alloc-initialMemory)/1024)
		}

		// Memory should not grow excessively
		maxMemoryIncrease := memoryUsage[len(memoryUsage)-1] - initialMemory
		assert.Less(t, maxMemoryIncrease, uint64(100*1024*1024), // 100MB
			"Memory usage should not exceed 100MB under load")
	})
}

// TestResourceUtilization tests CPU and memory utilization patterns
func (suite *PerformanceTestSuite) TestResourceUtilization() {
	suite.T().Run("CPUUtilizationPatterns", func(t *testing.T) {
		// This is a simplified CPU monitoring test
		// In a real scenario, you'd use more sophisticated monitoring

		// Run a sustained load
		orders := make([]dto.PlaceOrderRequest, 500)
		for i := 0; i < len(orders); i++ {
			orders[i] = helpers.RandomOrder(suite.testData.Symbols)
		}

		// Monitor goroutines as a proxy for resource usage
		initialGoroutines := runtime.NumGoroutine()

		start := time.Now()
		_, err := suite.loadRunner.RunConcurrentOrders(orders, 30)
		duration := time.Since(start)

		require.NoError(t, err)

		finalGoroutines := runtime.NumGoroutine()
		goroutineIncrease := finalGoroutines - initialGoroutines

		// Goroutines should return to reasonable levels
		time.Sleep(2 * time.Second) // Allow cleanup
		cleanupGoroutines := runtime.NumGoroutine()

		assert.Less(t, goroutineIncrease, 200, "Goroutine increase should be reasonable")
		assert.Less(t, cleanupGoroutines-initialGoroutines, 50, "Goroutines should cleanup properly")

		t.Logf("Resource utilization: Duration=%v, Goroutines: Initial=%d, Peak=%d, Final=%d",
			duration, initialGoroutines, finalGoroutines, cleanupGoroutines)
	})

	suite.T().Run("GarbageCollectionImpact", func(t *testing.T) {
		var gcStats runtime.MemStats
		runtime.ReadMemStats(&gcStats)
		initialGCs := gcStats.NumGC

		// Run load that should trigger GC
		orders := make([]dto.PlaceOrderRequest, 1000)
		for i := 0; i < len(orders); i++ {
			orders[i] = helpers.RandomOrder(suite.testData.Symbols)
		}

		start := time.Now()
		result, err := suite.loadRunner.RunConcurrentOrders(orders, 40)
		duration := time.Since(start)

		require.NoError(t, err)

		runtime.ReadMemStats(&gcStats)
		totalGCs := gcStats.NumGC - initialGCs

		// Verify performance remains acceptable even with GC
		assert.Greater(t, result.GetThroughput(), 50.0, "Throughput should remain high despite GC")
		assert.Less(t, result.GetLatencyPercentile(95), 500*time.Millisecond, "P95 latency should be reasonable")

		t.Logf("GC Impact: Duration=%v, GCs triggered=%d, Throughput=%.2f ops/s",
			duration, totalGCs, result.GetThroughput())
	})
}

// Benchmark functions for precise performance measurements

// BenchmarkOrderPlacement benchmarks single order placement
func BenchmarkOrderPlacement(b *testing.B) {
	server := helpers.NewTestServer(&testing.T{})
	defer server.Close()

	httpClient := helpers.NewHTTPClient(&testing.T{})
	order := dto.PlaceOrderRequest{
		Symbol:   "BTCUSD",
		Side:     "buy",
		Type:     "market",
		Quantity: 1.0,
		Price:    50000.0,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := httpClient.PlaceOrder(server.GetURL(""), order)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkConcurrentOrders benchmarks concurrent order processing
func BenchmarkConcurrentOrders(b *testing.B) {
	server := helpers.NewTestServer(&testing.T{})
	defer server.Close()

	loadRunner := helpers.NewLoadTestRunner(&testing.T{}, server.GetURL(""))
	symbols := []string{"BTCUSD", "ETHUSD", "ADAUSD"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orders := make([]dto.PlaceOrderRequest, 10)
		for j := 0; j < 10; j++ {
			orders[j] = helpers.RandomOrder(symbols)
		}

		_, err := loadRunner.RunConcurrentOrders(orders, 5)
		if err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkMetricsCollection benchmarks metrics collection performance
func BenchmarkMetricsCollection(b *testing.B) {
	server := helpers.NewTestServer(&testing.T{})
	defer server.Close()

	httpClient := helpers.NewHTTPClient(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := httpClient.GetMetrics(server.GetURL(""))
		if err != nil {
			b.Error(err)
		}
	}
}

// BenchmarkDemoSystemOperations benchmarks demo system operations
func BenchmarkDemoSystemOperations(b *testing.B) {
	server := helpers.NewTestServer(&testing.T{})
	defer server.Close()

	ctx := context.Background()
	server.DemoSystem.Start(ctx)
	defer server.DemoSystem.Stop(ctx)

	scenario := fixtures.GetLoadTestScenarios()[0]
	scenario.Duration = time.Second // Very short for benchmark

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := server.DemoSystem.Controller.StartLoadTest(ctx, scenario)
		if err != nil {
			b.Error(err)
		}

		// Wait briefly and stop
		time.Sleep(100 * time.Millisecond)
		server.DemoSystem.Controller.StopLoadTest(ctx)
	}
}

// Helper methods for performance analysis

// PerformanceData holds performance monitoring data
type PerformanceData struct {
	Latencies []time.Duration
	Timestamps []time.Time
	ErrorCount int64
}

// monitorPerformanceDuringLoadTest monitors performance during a load test
func (suite *PerformanceTestSuite) monitorPerformanceDuringLoadTest(ctx context.Context, duration time.Duration) *PerformanceData {
	data := &PerformanceData{
		Latencies:  make([]time.Duration, 0),
		Timestamps: make([]time.Time, 0),
	}

	stopTime := time.Now().Add(duration)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for time.Now().Before(stopTime) {
		select {
		case <-ticker.C:
			// Measure a sample operation latency
			start := time.Now()
			_, err := suite.httpClient.GetMetrics(suite.server.GetURL(""))
			latency := time.Since(start)

			if err != nil {
				atomic.AddInt64(&data.ErrorCount, 1)
			} else {
				data.Latencies = append(data.Latencies, latency)
				data.Timestamps = append(data.Timestamps, time.Now())
			}
		case <-ctx.Done():
			return data
		}
	}

	return data
}

// measureBaselineLatency measures baseline operation latency
func (suite *PerformanceTestSuite) measureBaselineLatency() time.Duration {
	samples := 10
	var totalLatency time.Duration

	for i := 0; i < samples; i++ {
		start := time.Now()
		_, err := suite.httpClient.GetMetrics(suite.server.GetURL(""))
		if err == nil {
			totalLatency += time.Since(start)
		}
		time.Sleep(100 * time.Millisecond)
	}

	return totalLatency / time.Duration(samples)
}

// measureLatencyDuringChaos measures latency during chaos injection
func (suite *PerformanceTestSuite) measureLatencyDuringChaos(duration time.Duration) time.Duration {
	samples := 0
	var totalLatency time.Duration
	stopTime := time.Now().Add(duration)

	for time.Now().Before(stopTime) {
		start := time.Now()
		_, err := suite.httpClient.GetMetrics(suite.server.GetURL(""))
		if err == nil {
			totalLatency += time.Since(start)
			samples++
		}
		time.Sleep(200 * time.Millisecond)
	}

	if samples == 0 {
		return time.Hour // Return high value if no successful samples
	}
	return totalLatency / time.Duration(samples)
}

// calculateAverageLatency calculates average latency from samples
func (suite *PerformanceTestSuite) calculateAverageLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	var total time.Duration
	for _, lat := range latencies {
		total += lat
	}
	return total / time.Duration(len(latencies))
}

// calculateMaxLatency finds maximum latency from samples
func (suite *PerformanceTestSuite) calculateMaxLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	max := latencies[0]
	for _, lat := range latencies[1:] {
		if lat > max {
			max = lat
		}
	}
	return max
}

// analyzeScalabilityTrends analyzes scalability trends from test results
func (suite *PerformanceTestSuite) analyzeScalabilityTrends(results map[string]*helpers.LoadTestResults) {
	suite.T().Logf("Scalability Analysis:")
	suite.T().Logf("%-10s | %-12s | %-12s | %-12s | %-12s", "Load", "Throughput", "Success Rate", "P95 Latency", "Efficiency")

	loadOrder := []string{"Light", "Medium", "Heavy", "Stress"}
	for _, load := range loadOrder {
		if result, exists := results[load]; exists {
			efficiency := result.GetSuccessRate() * result.GetThroughput() / 100.0 // Simple efficiency metric
			suite.T().Logf("%-10s | %-12.2f | %-12.2f | %-12v | %-12.2f",
				load,
				result.GetThroughput(),
				result.GetSuccessRate()*100,
				result.GetLatencyPercentile(95),
				efficiency)
		}
	}
}

// TestPerformanceTestSuite runs the performance test suite
func TestPerformanceTestSuite(t *testing.T) {
	suite.Run(t, new(PerformanceTestSuite))
}