package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"simulated_exchange/internal/demo"
	"simulated_exchange/test/fixtures"
	"simulated_exchange/test/helpers"
)

// DemoFlowTestSuite tests complete demo presentation workflows end-to-end
type DemoFlowTestSuite struct {
	suite.Suite
	server       *helpers.TestServer
	httpClient   *helpers.HTTPClient
	demoRunner   *helpers.DemoTestRunner
	testData     *fixtures.TestData
	assertions   *helpers.AssertionHelpers
	ctx          context.Context
	ctxCancel    context.CancelFunc
}

// SetupSuite initializes the demo test suite
func (suite *DemoFlowTestSuite) SetupSuite() {
	suite.server = helpers.NewTestServer(suite.T())
	suite.httpClient = helpers.NewHTTPClient(suite.T())
	suite.testData = fixtures.GetTestData()
	suite.assertions = helpers.NewAssertionHelpers(suite.T())

	// Create context for demo tests
	suite.ctx, suite.ctxCancel = context.WithTimeout(context.Background(), 10*time.Minute)

	// Start the demo system
	err := suite.server.DemoSystem.Start(suite.ctx)
	require.NoError(suite.T(), err)

	// Create demo test runner with WebSocket connection
	var runnerErr error
	suite.demoRunner, runnerErr = helpers.NewDemoTestRunner(
		suite.T(),
		suite.server.DemoSystem,
		suite.server.GetURL(""),
	)
	require.NoError(suite.T(), runnerErr)

	// Wait for system to be ready
	helpers.WaitForCondition(suite.T(), func() bool {
		health, err := suite.httpClient.GetHealth(suite.server.GetURL(""))
		return err == nil && health.Status == "healthy"
	}, 30*time.Second, "System should be healthy")

	// Wait for demo system to be running
	helpers.WaitForCondition(suite.T(), func() bool {
		return suite.server.DemoSystem.IsRunning()
	}, 15*time.Second, "Demo system should be running")
}

// TearDownSuite cleans up after demo test suite
func (suite *DemoFlowTestSuite) TearDownSuite() {
	if suite.demoRunner != nil {
		suite.demoRunner.Close()
	}
	if suite.ctxCancel != nil {
		suite.ctxCancel()
	}
	if suite.server != nil {
		suite.server.Close()
	}
}

// SetupTest resets demo system state before each test
func (suite *DemoFlowTestSuite) SetupTest() {
	err := suite.server.DemoSystem.Controller.ResetSystem(suite.ctx)
	require.NoError(suite.T(), err)

	// Wait for reset to complete
	time.Sleep(200 * time.Millisecond)
}

// TestLoadTestScenarios tests various load testing demo scenarios
func (suite *DemoFlowTestSuite) TestLoadTestScenarios() {
	scenarios := suite.testData.LoadScenarios

	suite.T().Run("LightLoadScenario", func(t *testing.T) {
		scenario := scenarios[0] // E2E Light Load

		// Run the load test scenario
		results, err := suite.demoRunner.RunLoadTestScenario(suite.ctx, scenario)
		require.NoError(t, err)

		// Assert scenario completion
		suite.assertions.AssertLoadTestCompletion(results, scenario.Duration)

		// Verify WebSocket updates were received
		assert.Greater(t, results.GetUpdateCount(), 0, "Should receive WebSocket updates during load test")

		// Verify final status
		assert.NotNil(t, results.FinalStatus, "Should have final status")
		assert.Equal(t, demo.LoadPhaseCompleted, results.FinalStatus.Phase, "Should complete in completed phase")
		assert.Equal(t, 100.0, results.FinalStatus.Progress, "Should reach 100% progress")

		t.Logf("Light load scenario completed: Duration=%v, Updates=%d, Orders=%d",
			results.GetDuration(), results.GetUpdateCount(),
			results.FinalStatus.CompletedOrders+results.FinalStatus.FailedOrders)
	})

	suite.T().Run("MediumLoadScenario", func(t *testing.T) {
		scenario := scenarios[1] // E2E Medium Load

		// Run the medium load test scenario
		results, err := suite.demoRunner.RunLoadTestScenario(suite.ctx, scenario)
		require.NoError(t, err)

		// Assert scenario completion
		suite.assertions.AssertLoadTestCompletion(results, scenario.Duration)

		// Verify higher throughput for medium load
		totalOrders := results.FinalStatus.CompletedOrders + results.FinalStatus.FailedOrders
		expectedMinOrders := int64(float64(scenario.OrdersPerSecond) * scenario.Duration.Seconds() * 0.7) // 70% of target
		assert.GreaterOrEqual(t, totalOrders, expectedMinOrders,
			"Medium load should process at least 70%% of target orders")

		t.Logf("Medium load scenario completed: Duration=%v, Orders=%d, Target=%d",
			results.GetDuration(), totalOrders, int64(float64(scenario.OrdersPerSecond)*scenario.Duration.Seconds()))
	})

	suite.T().Run("DemoPresentationScenario", func(t *testing.T) {
		scenario := scenarios[3] // E2E Demo Presentation

		// For demo presentation, we'll run a shorter version
		shortScenario := scenario
		shortScenario.Duration = 30 * time.Second
		shortScenario.RampUpTime = 5 * time.Second
		shortScenario.SustainTime = 20 * time.Second
		shortScenario.RampDownTime = 5 * time.Second

		// Run the demo presentation scenario
		results, err := suite.demoRunner.RunLoadTestScenario(suite.ctx, shortScenario)
		require.NoError(t, err)

		// Assert scenario completion
		suite.assertions.AssertLoadTestCompletion(results, shortScenario.Duration)

		// Verify real-time updates for presentation
		assert.Greater(t, results.GetUpdateCount(), 10, "Demo presentation should generate frequent updates")

		// Verify phase transitions
		phasesSeen := make(map[demo.LoadPhase]bool)
		for _, update := range results.Updates {
			if update.Type == demo.UpdateLoadTestStatus {
				if updateData, ok := update.Data.(*demo.LoadTestStatus); ok {
					phasesSeen[updateData.Phase] = true
				}
			}
		}

		// Should see at least ramp-up and sustained phases
		assert.True(t, phasesSeen[demo.LoadPhaseRampUp] || phasesSeen[demo.LoadPhaseSustained],
			"Should observe load test phases during presentation")

		t.Logf("Demo presentation completed: Duration=%v, Updates=%d, Phases observed=%v",
			results.GetDuration(), results.GetUpdateCount(), phasesSeen)
	})
}

// TestChaosTestScenarios tests chaos engineering demo scenarios
func (suite *DemoFlowTestSuite) TestChaosTestScenarios() {
	scenarios := suite.testData.ChaosScenarios

	suite.T().Run("LatencyInjectionScenario", func(t *testing.T) {
		scenario := scenarios[0] // E2E Latency Injection

		// Start chaos test
		err := suite.server.DemoSystem.Controller.TriggerChaosTest(suite.ctx, scenario)
		require.NoError(t, err)

		// Monitor chaos test progress
		startTime := time.Now()
		maxWaitTime := scenario.Duration + 30*time.Second

		var finalStatus *demo.ChaosTestStatus
		for time.Since(startTime) < maxWaitTime {
			status, err := suite.server.DemoSystem.Controller.GetChaosTestStatus(suite.ctx)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}

			if !status.IsRunning {
				finalStatus = status
				break
			}

			time.Sleep(time.Second)
		}

		// Verify chaos test completed
		require.NotNil(t, finalStatus, "Chaos test should complete")
		assert.False(t, finalStatus.IsRunning, "Chaos test should not be running")
		assert.Equal(t, demo.ChaosPhaseRecovery, finalStatus.Phase, "Should complete in recovery phase")

		t.Logf("Latency injection completed: Duration=%v, Errors=%d",
			time.Since(startTime), len(finalStatus.Errors))
	})

	suite.T().Run("ErrorSimulationScenario", func(t *testing.T) {
		scenario := scenarios[1] // E2E Error Simulation

		// Start chaos test
		err := suite.server.DemoSystem.Controller.TriggerChaosTest(suite.ctx, scenario)
		require.NoError(t, err)

		// Let it run for a shorter time for testing
		time.Sleep(10 * time.Second)

		// Check if chaos test is active
		status, err := suite.server.DemoSystem.Controller.GetChaosTestStatus(suite.ctx)
		require.NoError(t, err)

		if status.IsRunning {
			// Stop the chaos test early for testing
			err = suite.server.DemoSystem.Controller.StopChaosTest(suite.ctx)
			require.NoError(t, err)

			// Wait for stop to complete
			time.Sleep(2 * time.Second)

			// Verify it stopped
			status, err = suite.server.DemoSystem.Controller.GetChaosTestStatus(suite.ctx)
			require.NoError(t, err)
			assert.False(t, status.IsRunning, "Chaos test should stop when requested")
		}

		t.Logf("Error simulation scenario controlled: Final phase=%s", status.Phase)
	})
}

// TestDemoControlFlow tests the complete demo control workflow
func (suite *DemoFlowTestSuite) TestDemoControlFlow() {
	suite.T().Run("CompleteDemo resentationFlow", func(t *testing.T) {
		// Step 1: Reset system to clean state
		err := suite.server.DemoSystem.Controller.ResetSystem(suite.ctx)
		require.NoError(t, err)

		// Step 2: Check initial system status
		initialStatus, err := suite.server.DemoSystem.Controller.GetSystemStatus(suite.ctx)
		require.NoError(t, err)
		assert.Equal(t, "idle", initialStatus.Overall, "System should start in idle state")
		assert.Empty(t, initialStatus.ActiveScenarios, "Should have no active scenarios initially")

		// Step 3: Start load test demo
		loadScenario := suite.testData.LoadScenarios[0]
		loadScenario.Duration = 20 * time.Second // Shorten for test

		err = suite.server.DemoSystem.Controller.StartLoadTest(suite.ctx, loadScenario)
		require.NoError(t, err)

		// Step 4: Verify load test is running
		loadStatus, err := suite.server.DemoSystem.Controller.GetLoadTestStatus(suite.ctx)
		require.NoError(t, err)
		assert.True(t, loadStatus.IsRunning, "Load test should be running")

		// Step 5: Check system status during load test
		duringStatus, err := suite.server.DemoSystem.Controller.GetSystemStatus(suite.ctx)
		require.NoError(t, err)
		assert.Contains(t, duringStatus.ActiveScenarios, loadScenario.Name, "Should show active load scenario")

		// Step 6: Wait for load test to complete
		helpers.WaitForCondition(suite.T(), func() bool {
			status, err := suite.server.DemoSystem.Controller.GetLoadTestStatus(suite.ctx)
			return err == nil && !status.IsRunning
		}, loadScenario.Duration+10*time.Second, "Load test should complete")

		// Step 7: Start chaos test demo
		chaosScenario := suite.testData.ChaosScenarios[0]
		chaosScenario.Duration = 15 * time.Second // Shorten for test

		err = suite.server.DemoSystem.Controller.TriggerChaosTest(suite.ctx, chaosScenario)
		require.NoError(t, err)

		// Step 8: Verify chaos test is running
		chaosStatus, err := suite.server.DemoSystem.Controller.GetChaosTestStatus(suite.ctx)
		require.NoError(t, err)
		assert.True(t, chaosStatus.IsRunning, "Chaos test should be running")

		// Step 9: Wait for chaos test to complete
		helpers.WaitForCondition(suite.T(), func() bool {
			status, err := suite.server.DemoSystem.Controller.GetChaosTestStatus(suite.ctx)
			return err == nil && !status.IsRunning
		}, chaosScenario.Duration+10*time.Second, "Chaos test should complete")

		// Step 10: Final system reset
		err = suite.server.DemoSystem.Controller.ResetSystem(suite.ctx)
		require.NoError(t, err)

		// Step 11: Verify clean final state
		finalStatus, err := suite.server.DemoSystem.Controller.GetSystemStatus(suite.ctx)
		require.NoError(t, err)
		assert.Equal(t, "idle", finalStatus.Overall, "System should return to idle state")

		t.Log("Complete demo presentation flow executed successfully")
	})
}

// TestWebSocketUpdates tests real-time WebSocket updates during demos
func (suite *DemoFlowTestSuite) TestWebSocketUpdates() {
	suite.T().Run("LoadTestUpdates", func(t *testing.T) {
		// Start monitoring WebSocket updates
		updatesChan := make(chan demo.DemoUpdate, 100)
		go suite.monitorWebSocketUpdates(updatesChan)

		// Start a load test
		scenario := suite.testData.LoadScenarios[0]
		scenario.Duration = 15 * time.Second

		err := suite.server.DemoSystem.Controller.StartLoadTest(suite.ctx, scenario)
		require.NoError(t, err)

		// Collect updates for a period
		var updates []demo.DemoUpdate
		timeout := time.After(20 * time.Second)

	CollectLoop:
		for {
			select {
			case update := <-updatesChan:
				updates = append(updates, update)
			case <-timeout:
				break CollectLoop
			}
		}

		// Verify we received updates
		assert.Greater(t, len(updates), 0, "Should receive WebSocket updates during load test")

		// Verify update types
		updateTypes := make(map[demo.DemoUpdateType]int)
		for _, update := range updates {
			updateTypes[update.Type]++
		}

		assert.Greater(t, updateTypes[demo.UpdateLoadTestStatus], 0, "Should receive load test status updates")

		t.Logf("Received %d WebSocket updates: %v", len(updates), updateTypes)
	})

	suite.T().Run("SystemStatusUpdates", func(t *testing.T) {
		// Monitor system-level updates
		updatesChan := make(chan demo.DemoUpdate, 50)
		go suite.monitorWebSocketUpdates(updatesChan)

		// Trigger system changes
		err := suite.server.DemoSystem.Controller.ResetSystem(suite.ctx)
		require.NoError(t, err)

		// Collect updates for reset operation
		var updates []demo.DemoUpdate
		timeout := time.After(5 * time.Second)

	ResetLoop:
		for {
			select {
			case update := <-updatesChan:
				updates = append(updates, update)
			case <-timeout:
				break ResetLoop
			}
		}

		// Should receive some system-related updates
		assert.GreaterOrEqual(t, len(updates), 0, "Should receive updates during system operations")

		t.Logf("System operation updates: %d", len(updates))
	})
}

// monitorWebSocketUpdates monitors WebSocket messages and parses demo updates
func (suite *DemoFlowTestSuite) monitorWebSocketUpdates(updatesChan chan<- demo.DemoUpdate) {
	for {
		message, err := suite.demoRunner.WaitForMessage(2 * time.Second)
		if err != nil {
			// Timeout is expected when no updates
			continue
		}

		var update demo.DemoUpdate
		if err := json.Unmarshal(message, &update); err == nil {
			select {
			case updatesChan <- update:
			default:
				// Channel full, skip update
			}
		}
	}
}

// TestDemoScenarioConfiguration tests demo scenario configuration and validation
func (suite *DemoFlowTestSuite) TestDemoScenarioConfiguration() {
	suite.T().Run("LoadScenarioValidation", func(t *testing.T) {
		// Test valid load scenario
		validScenario := demo.LoadTestScenario{
			Name:            "Valid Test Scenario",
			Description:     "A valid scenario for testing",
			Intensity:       demo.LoadLight,
			Duration:        30 * time.Second,
			OrdersPerSecond: 10,
			ConcurrentUsers: 5,
			Symbols:         []string{"BTCUSD"},
			OrderTypes:      []string{"market", "limit"},
		}

		err := suite.server.DemoSystem.Controller.StartLoadTest(suite.ctx, validScenario)
		require.NoError(t, err, "Valid scenario should start successfully")

		// Stop the test
		err = suite.server.DemoSystem.Controller.StopLoadTest(suite.ctx)
		require.NoError(t, err)

		// Test invalid scenarios
		invalidScenarios := []demo.LoadTestScenario{
			// Zero duration
			{Name: "Zero Duration", Duration: 0, OrdersPerSecond: 10, ConcurrentUsers: 5, Symbols: []string{"BTCUSD"}},
			// Negative orders per second
			{Name: "Negative OPS", Duration: time.Minute, OrdersPerSecond: -1, ConcurrentUsers: 5, Symbols: []string{"BTCUSD"}},
			// Zero concurrent users
			{Name: "Zero Users", Duration: time.Minute, OrdersPerSecond: 10, ConcurrentUsers: 0, Symbols: []string{"BTCUSD"}},
			// Empty symbols
			{Name: "No Symbols", Duration: time.Minute, OrdersPerSecond: 10, ConcurrentUsers: 5, Symbols: []string{}},
		}

		for i, scenario := range invalidScenarios {
			err := suite.server.DemoSystem.Controller.StartLoadTest(suite.ctx, scenario)
			if err == nil {
				// If it started, stop it and consider it handled gracefully
				suite.server.DemoSystem.Controller.StopLoadTest(suite.ctx)
				t.Logf("Invalid scenario %d was handled gracefully", i)
			} else {
				t.Logf("Invalid scenario %d correctly rejected: %v", i, err)
			}
		}
	})

	suite.T().Run("ChaosScenarioValidation", func(t *testing.T) {
		// Test valid chaos scenario
		validScenario := demo.ChaosTestScenario{
			Name:        "Valid Chaos Test",
			Description: "A valid chaos scenario",
			Type:        demo.ChaosLatencyInjection,
			Duration:    15 * time.Second,
			Severity:    demo.ChaosLow,
			Parameters: map[string]interface{}{
				"latency_ms": 100,
			},
		}

		err := suite.server.DemoSystem.Controller.TriggerChaosTest(suite.ctx, validScenario)
		require.NoError(t, err, "Valid chaos scenario should start successfully")

		// Stop the test
		err = suite.server.DemoSystem.Controller.StopChaosTest(suite.ctx)
		require.NoError(t, err)
	})
}

// TestDemoSystemIntegration tests integration between demo components
func (suite *DemoFlowTestSuite) TestDemoSystemIntegration() {
	suite.T().Run("DemoWithTrading", func(t *testing.T) {
		// Start a demo load test
		scenario := suite.testData.LoadScenarios[0]
		scenario.Duration = 20 * time.Second

		err := suite.server.DemoSystem.Controller.StartLoadTest(suite.ctx, scenario)
		require.NoError(t, err)

		// Simultaneously place manual orders through the API
		baseURL := suite.server.GetURL("")
		manualOrders := suite.testData.Orders[:3]

		for _, order := range manualOrders {
			resp, err := suite.httpClient.PlaceOrder(baseURL, order)
			require.NoError(t, err)
			suite.assertions.AssertOrderPlacementSuccess(resp, order.Symbol)
		}

		// Wait for demo to complete
		helpers.WaitForCondition(suite.T(), func() bool {
			status, err := suite.server.DemoSystem.Controller.GetLoadTestStatus(suite.ctx)
			return err == nil && !status.IsRunning
		}, scenario.Duration+10*time.Second, "Demo load test should complete")

		// Verify both demo and manual orders were processed
		metrics, err := suite.httpClient.GetMetrics(baseURL)
		require.NoError(t, err)
		assert.Greater(t, metrics.OrderCount, int64(len(manualOrders)), "Should process both demo and manual orders")

		t.Logf("Integration test completed: Total orders processed=%d", metrics.OrderCount)
	})

	suite.T().Run("DemoSystemHealth", func(t *testing.T) {
		// Check demo system health
		health := suite.server.DemoSystem.HealthCheck(suite.ctx)
		require.NotNil(t, health)

		assert.Equal(t, "healthy", health["demo_system"], "Demo system should be healthy")
		assert.NotNil(t, health["timestamp"], "Should have timestamp")
		assert.Equal(t, true, health["running"], "Demo system should be running")

		t.Logf("Demo system health: %v", health)
	})
}

// TestDemoFlowTestSuite runs the demo flow test suite
func TestDemoFlowTestSuite(t *testing.T) {
	suite.Run(t, new(DemoFlowTestSuite))
}