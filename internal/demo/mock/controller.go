package mock

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"simulated_exchange/internal/demo"
)

// Controller implements demo.DemoController interface for testing
type Controller struct {
	logger          *slog.Logger
	loadTestStatus  *demo.LoadTestStatus
	chaosTestStatus *demo.ChaosTestStatus
	systemStatus    *demo.DemoSystemStatus
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewController creates a new mock demo controller
func NewController(logger *slog.Logger) *Controller {
	ctx, cancel := context.WithCancel(context.Background())

	return &Controller{
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		loadTestStatus: &demo.LoadTestStatus{
			IsRunning: false,
			Phase:     demo.LoadPhaseCompleted,
			Progress:  0,
		},
		chaosTestStatus: &demo.ChaosTestStatus{
			IsRunning: false,
			Phase:     demo.ChaosPhaseCompleted,
		},
		systemStatus: &demo.DemoSystemStatus{
			Overall: demo.HealthHealthy,
			TradingEngine: demo.ComponentHealth{
				Status:       demo.HealthHealthy,
				ResponseTime: 5 * time.Millisecond,
			},
			OrderService: demo.ComponentHealth{
				Status:       demo.HealthHealthy,
				ResponseTime: 3 * time.Millisecond,
			},
			MetricsService: demo.ComponentHealth{
				Status:       demo.HealthHealthy,
				ResponseTime: 2 * time.Millisecond,
			},
			Database: demo.ComponentHealth{
				Status:       demo.HealthHealthy,
				ResponseTime: 10 * time.Millisecond,
			},
			ActiveScenarios: []string{},
			Alerts:          []demo.SystemAlert{},
		},
	}
}

func (m *Controller) StartLoadTest(ctx context.Context, scenario demo.LoadTestScenario) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.loadTestStatus.IsRunning {
		return fmt.Errorf("load test already running")
	}

	m.logger.Info("Starting load test", "scenario", scenario.Name, "intensity", scenario.Intensity)

	m.loadTestStatus = &demo.LoadTestStatus{
		IsRunning:  true,
		Phase:      demo.LoadPhaseRampUp,
		Progress:   0,
		Scenario:   &scenario,
		StartTime:  time.Now(),
		CurrentMetrics: &demo.LoadTestMetrics{
			Timestamp:       time.Now(),
			OrdersPerSecond: 0,
			AverageLatency:  0,
			ErrorRate:       0,
		},
		HistoricalMetrics: []demo.LoadTestMetrics{},
		ActiveOrders:      0,
		CompletedOrders:   0,
		FailedOrders:      0,
		CurrentUsers:      scenario.ConcurrentUsers,
		Errors:            []demo.LoadTestError{},
	}

	// Start background load test simulation with internal context
	// Don't use the request context as it will be cancelled when the HTTP request completes
	go m.simulateLoadTest(m.ctx, scenario)

	return nil
}

func (m *Controller) simulateLoadTest(ctx context.Context, scenario demo.LoadTestScenario) {
	defer func() {
		m.mutex.Lock()
		m.loadTestStatus.IsRunning = false
		m.loadTestStatus.Phase = demo.LoadPhaseCompleted
		m.loadTestStatus.Progress = 100
		m.mutex.Unlock()
		m.logger.Info("Load test completed")
	}()

	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			if elapsed >= scenario.Duration {
				return
			}

			m.mutex.Lock()
			progress := float64(elapsed) / float64(scenario.Duration) * 100
			m.loadTestStatus.Progress = progress
			m.loadTestStatus.ElapsedTime = elapsed
			m.loadTestStatus.RemainingTime = scenario.Duration - elapsed

			// Update phase based on progress
			if progress < 10 {
				m.loadTestStatus.Phase = demo.LoadPhaseRampUp
			} else if progress < 90 {
				m.loadTestStatus.Phase = demo.LoadPhaseSustained
			} else {
				m.loadTestStatus.Phase = demo.LoadPhaseRampDown
			}

			// Simulate metrics
			if m.loadTestStatus.CurrentMetrics != nil {
				m.loadTestStatus.CurrentMetrics.Timestamp = time.Now()
				m.loadTestStatus.CurrentMetrics.OrdersPerSecond = float64(scenario.OrdersPerSecond)
				m.loadTestStatus.CurrentMetrics.AverageLatency = 20 + float64(time.Now().UnixNano()%30)
				m.loadTestStatus.CurrentMetrics.P95Latency = m.loadTestStatus.CurrentMetrics.AverageLatency + 10
				m.loadTestStatus.CurrentMetrics.P99Latency = m.loadTestStatus.CurrentMetrics.AverageLatency + 20
				m.loadTestStatus.CurrentMetrics.ErrorRate = 0.01
				m.loadTestStatus.CurrentMetrics.Throughput = float64(scenario.OrdersPerSecond)
			}

			// Update counts
			m.loadTestStatus.CompletedOrders += scenario.OrdersPerSecond
			m.loadTestStatus.FailedOrders += int(float64(scenario.OrdersPerSecond) * 0.01)

			m.mutex.Unlock()
		}
	}
}

func (m *Controller) StopLoadTest(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.loadTestStatus.IsRunning {
		return fmt.Errorf("no load test is currently running")
	}

	m.logger.Info("Stopping load test")
	m.loadTestStatus.IsRunning = false
	m.loadTestStatus.Phase = demo.LoadPhaseCompleted

	return nil
}

func (m *Controller) GetLoadTestStatus(ctx context.Context) (*demo.LoadTestStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to avoid race conditions
	statusCopy := *m.loadTestStatus
	return &statusCopy, nil
}

func (m *Controller) TriggerChaosTest(ctx context.Context, scenario demo.ChaosTestScenario) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.chaosTestStatus.IsRunning {
		return fmt.Errorf("chaos test already running")
	}

	m.logger.Info("Starting chaos test", "type", scenario.Type, "severity", scenario.Severity)

	m.chaosTestStatus = &demo.ChaosTestStatus{
		IsRunning:       true,
		Phase:           demo.ChaosPhaseInjection,
		Scenario:        &scenario,
		StartTime:       time.Now(),
		Metrics: &demo.ChaosTestMetrics{
			ServiceDegradation: 0,
			ResilienceScore:    1.0,
			ErrorsGenerated:    0,
		},
		AffectedTargets: []string{},
		Errors:          []demo.ChaosTestError{},
	}

	// Start background chaos test simulation
	go m.simulateChaosTest(m.ctx, scenario)

	return nil
}

func (m *Controller) simulateChaosTest(ctx context.Context, scenario demo.ChaosTestScenario) {
	defer func() {
		m.mutex.Lock()
		m.chaosTestStatus.IsRunning = false
		m.chaosTestStatus.Phase = demo.ChaosPhaseCompleted
		m.mutex.Unlock()
		m.logger.Info("Chaos test completed")
	}()

	startTime := time.Now()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			if elapsed >= scenario.Duration {
				return
			}

			m.mutex.Lock()

			// Update phase based on time
			if elapsed < scenario.Duration/10 {
				m.chaosTestStatus.Phase = demo.ChaosPhaseInjection
			} else if elapsed < scenario.Duration*9/10 {
				m.chaosTestStatus.Phase = demo.ChaosPhaseSustained
			} else {
				m.chaosTestStatus.Phase = demo.ChaosPhaseRecovery
			}

			// Simulate impact metrics based on chaos type
			if m.chaosTestStatus.Metrics != nil {
				progress := float64(elapsed) / float64(scenario.Duration)

				switch scenario.Type {
				case demo.ChaosLatencyInjection:
					m.chaosTestStatus.Metrics.ServiceDegradation = 0.3 * progress
					m.chaosTestStatus.Metrics.ResilienceScore = 1.0 - (0.2 * progress)
				case demo.ChaosErrorSimulation:
					m.chaosTestStatus.Metrics.ServiceDegradation = 0.5 * progress
					m.chaosTestStatus.Metrics.ResilienceScore = 1.0 - (0.4 * progress)
					m.chaosTestStatus.Metrics.ErrorsGenerated = int(progress * 100)
				case demo.ChaosResourceExhaustion:
					m.chaosTestStatus.Metrics.ServiceDegradation = 0.7 * progress
					m.chaosTestStatus.Metrics.ResilienceScore = 1.0 - (0.6 * progress)
				default:
					m.chaosTestStatus.Metrics.ServiceDegradation = 0.2 * progress
					m.chaosTestStatus.Metrics.ResilienceScore = 1.0 - (0.1 * progress)
				}

				m.chaosTestStatus.Metrics.Timestamp = time.Now()
			}

			m.mutex.Unlock()
		}
	}
}

func (m *Controller) StopChaosTest(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.chaosTestStatus.IsRunning {
		return fmt.Errorf("no chaos test is currently running")
	}

	m.logger.Info("Stopping chaos test")
	m.chaosTestStatus.IsRunning = false
	m.chaosTestStatus.Phase = demo.ChaosPhaseCompleted

	return nil
}

func (m *Controller) GetChaosTestStatus(ctx context.Context) (*demo.ChaosTestStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	statusCopy := *m.chaosTestStatus
	return &statusCopy, nil
}

func (m *Controller) ResetSystem(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.logger.Info("Resetting demo system")

	m.loadTestStatus = &demo.LoadTestStatus{
		IsRunning: false,
		Phase:     demo.LoadPhaseCompleted,
		Progress:  0,
	}

	m.chaosTestStatus = &demo.ChaosTestStatus{
		IsRunning: false,
		Phase:     demo.ChaosPhaseCompleted,
	}

	return nil
}

func (m *Controller) GetSystemStatus(ctx context.Context) (*demo.DemoSystemStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	statusCopy := *m.systemStatus
	return &statusCopy, nil
}

func (m *Controller) Subscribe(subscriber demo.DemoSubscriber) error {
	return fmt.Errorf("not implemented")
}

func (m *Controller) Unsubscribe(subscriberID string) error {
	return fmt.Errorf("not implemented")
}

func (m *Controller) BroadcastUpdate(update demo.DemoUpdate) error {
	return fmt.Errorf("not implemented")
}

func (m *Controller) Close() error {
	m.cancel()
	return nil
}
