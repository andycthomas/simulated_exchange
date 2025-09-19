package demo

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"simulated_exchange/internal/api/dto"
)

// StandardScenarioManager implements the ScenarioManager interface
type StandardScenarioManager struct {
	tradingEngine    TradingEngineIntegration
	logger           *slog.Logger
	config           *DemoConfig

	// Scenario state
	mu                 sync.RWMutex
	activeLoadScenario *LoadTestScenario
	activeChaosScenario *ChaosTestScenario

	// Chaos injection state
	chaosInjectors     map[string]ChaosInjector
	injectionActive    bool

	// Load generation state
	loadGenerators     []*LoadGenerator
	loadActive         bool
	stopLoadChan       chan struct{}
	stopChaosChan      chan struct{}
}

// ChaosInjector interface for different chaos types
type ChaosInjector interface {
	Inject(ctx context.Context, params ChaosParams) error
	Stop(ctx context.Context) error
	GetStatus() ChaosInjectionStatus
}

// LoadGenerator generates load for testing
type LoadGenerator struct {
	ID           string
	Scenario     LoadTestScenario
	TradingEngine TradingEngineIntegration
	Logger       *slog.Logger

	OrderCounter int
	ErrorCount   int
	StopChan     chan struct{}
	Active       bool
}

// ChaosInjectionStatus represents the status of chaos injection
type ChaosInjectionStatus struct {
	Type            ChaosType `json:"type"`
	Active          bool      `json:"active"`
	StartTime       time.Time `json:"start_time"`
	TargetsAffected int       `json:"targets_affected"`
	Metrics         map[string]float64 `json:"metrics"`
}

// MetricsCollector collects system and application metrics
type MetricsCollector struct {
	config           DemoMetricsConfig
	logger           *slog.Logger
	startTime        time.Time
	lastGCPause      time.Duration
	lastCollectionTime time.Time
}

// NewStandardScenarioManager creates a new scenario manager
func NewStandardScenarioManager(
	tradingEngine TradingEngineIntegration,
	logger *slog.Logger,
	config *DemoConfig,
) *StandardScenarioManager {
	manager := &StandardScenarioManager{
		tradingEngine:   tradingEngine,
		logger:          logger,
		config:          config,
		chaosInjectors:  make(map[string]ChaosInjector),
		loadGenerators:  []*LoadGenerator{},
		stopLoadChan:    make(chan struct{}),
		stopChaosChan:   make(chan struct{}),
	}

	// Initialize chaos injectors
	manager.initializeChaosInjectors()

	return manager
}

// ExecuteLoadScenario starts executing a load test scenario
func (sm *StandardScenarioManager) ExecuteLoadScenario(ctx context.Context, scenario LoadTestScenario) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.loadActive {
		return fmt.Errorf("load scenario already active")
	}

	sm.logger.Info("Starting load scenario execution", "scenario", scenario.Name)

	sm.activeLoadScenario = &scenario
	sm.loadActive = true

	// Create load generators based on concurrent users
	sm.loadGenerators = make([]*LoadGenerator, scenario.ConcurrentUsers)
	for i := 0; i < scenario.ConcurrentUsers; i++ {
		generator := &LoadGenerator{
			ID:            fmt.Sprintf("generator_%d", i),
			Scenario:      scenario,
			TradingEngine: sm.tradingEngine,
			Logger:        sm.logger,
			StopChan:      make(chan struct{}),
			Active:        true,
		}
		sm.loadGenerators[i] = generator

		// Start load generation in goroutine
		go sm.runLoadGenerator(ctx, generator)
	}

	return nil
}

// StopLoadScenario stops the current load scenario
func (sm *StandardScenarioManager) StopLoadScenario(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.loadActive {
		return fmt.Errorf("no load scenario currently active")
	}

	sm.logger.Info("Stopping load scenario", "scenario", sm.activeLoadScenario.Name)

	// Stop all load generators
	for _, generator := range sm.loadGenerators {
		if generator.Active {
			close(generator.StopChan)
			generator.Active = false
		}
	}

	sm.loadActive = false
	sm.activeLoadScenario = nil
	sm.loadGenerators = []*LoadGenerator{}

	return nil
}

// ExecuteChaosScenario starts executing a chaos test scenario
func (sm *StandardScenarioManager) ExecuteChaosScenario(ctx context.Context, scenario ChaosTestScenario) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.injectionActive {
		return fmt.Errorf("chaos scenario already active")
	}

	sm.logger.Info("Starting chaos scenario execution", "scenario", scenario.Name, "type", scenario.Type)

	injector, exists := sm.chaosInjectors[string(scenario.Type)]
	if !exists {
		return fmt.Errorf("chaos injector not found for type: %s", scenario.Type)
	}

	sm.activeChaosScenario = &scenario
	sm.injectionActive = true

	// Start chaos injection
	if err := injector.Inject(ctx, scenario.Parameters); err != nil {
		sm.injectionActive = false
		sm.activeChaosScenario = nil
		return fmt.Errorf("failed to start chaos injection: %w", err)
	}

	return nil
}

// StopChaosScenario stops the current chaos scenario
func (sm *StandardScenarioManager) StopChaosScenario(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.injectionActive {
		return fmt.Errorf("no chaos scenario currently active")
	}

	sm.logger.Info("Stopping chaos scenario", "scenario", sm.activeChaosScenario.Name)

	// Stop chaos injection
	injector, exists := sm.chaosInjectors[string(sm.activeChaosScenario.Type)]
	if exists {
		if err := injector.Stop(ctx); err != nil {
			sm.logger.Error("Failed to stop chaos injector", "type", sm.activeChaosScenario.Type, "error", err)
		}
	}

	sm.injectionActive = false
	sm.activeChaosScenario = nil

	return nil
}

// GetAvailableLoadScenarios returns predefined load test scenarios
func (sm *StandardScenarioManager) GetAvailableLoadScenarios() []LoadTestScenario {
	return []LoadTestScenario{
		{
			Name:            "Light Load Test",
			Description:     "Low intensity load test for baseline performance",
			Intensity:       LoadLight,
			Duration:        2 * time.Minute,
			OrdersPerSecond: 10,
			ConcurrentUsers: 5,
			Symbols:         []string{"BTCUSD", "ETHUSD"},
			OrderTypes:      []string{"market", "limit"},
			PriceVariation:  0.05,
			VolumeRange:     VolumeRange{Min: 0.1, Max: 1.0},
			UserBehaviorPattern: UserBehaviorPattern{
				BuyRatio:         0.5,
				SellRatio:        0.5,
				MarketOrderRatio: 0.3,
				LimitOrderRatio:  0.7,
				CancelRatio:      0.1,
			},
			RampUp: RampUpConfig{
				Enabled:      true,
				Duration:     30 * time.Second,
				StartPercent: 10,
				EndPercent:   100,
			},
			Metrics: MetricsConfig{
				CollectLatency:     true,
				CollectThroughput:  true,
				CollectErrorRate:   true,
				CollectResourceUse: true,
				SampleRate:         100,
			},
		},
		{
			Name:            "Medium Load Test",
			Description:     "Medium intensity load test for typical usage patterns",
			Intensity:       LoadMedium,
			Duration:        5 * time.Minute,
			OrdersPerSecond: 50,
			ConcurrentUsers: 25,
			Symbols:         []string{"BTCUSD", "ETHUSD", "ADAUSD"},
			OrderTypes:      []string{"market", "limit", "stop"},
			PriceVariation:  0.1,
			VolumeRange:     VolumeRange{Min: 0.5, Max: 5.0},
			UserBehaviorPattern: UserBehaviorPattern{
				BuyRatio:         0.55,
				SellRatio:        0.45,
				MarketOrderRatio: 0.4,
				LimitOrderRatio:  0.5,
				CancelRatio:      0.15,
			},
			RampUp: RampUpConfig{
				Enabled:      true,
				Duration:     60 * time.Second,
				StartPercent: 20,
				EndPercent:   100,
			},
			Metrics: MetricsConfig{
				CollectLatency:     true,
				CollectThroughput:  true,
				CollectErrorRate:   true,
				CollectResourceUse: true,
				SampleRate:         100,
			},
		},
		{
			Name:            "Heavy Load Test",
			Description:     "High intensity load test for peak performance evaluation",
			Intensity:       LoadHeavy,
			Duration:        10 * time.Minute,
			OrdersPerSecond: 200,
			ConcurrentUsers: 100,
			Symbols:         []string{"BTCUSD", "ETHUSD", "ADAUSD", "DOTUSD", "SOLUSD"},
			OrderTypes:      []string{"market", "limit", "stop", "stop_limit"},
			PriceVariation:  0.15,
			VolumeRange:     VolumeRange{Min: 1.0, Max: 10.0},
			UserBehaviorPattern: UserBehaviorPattern{
				BuyRatio:         0.6,
				SellRatio:        0.4,
				MarketOrderRatio: 0.5,
				LimitOrderRatio:  0.4,
				CancelRatio:      0.2,
			},
			RampUp: RampUpConfig{
				Enabled:      true,
				Duration:     120 * time.Second,
				StartPercent: 30,
				EndPercent:   100,
			},
			Metrics: MetricsConfig{
				CollectLatency:     true,
				CollectThroughput:  true,
				CollectErrorRate:   true,
				CollectResourceUse: true,
				SampleRate:         100,
			},
		},
		{
			Name:            "Stress Test",
			Description:     "Maximum intensity stress test to find system limits",
			Intensity:       LoadStress,
			Duration:        3 * time.Minute,
			OrdersPerSecond: 500,
			ConcurrentUsers: 200,
			Symbols:         []string{"BTCUSD", "ETHUSD", "ADAUSD", "DOTUSD", "SOLUSD", "LINKUSD"},
			OrderTypes:      []string{"market", "limit", "stop", "stop_limit"},
			PriceVariation:  0.2,
			VolumeRange:     VolumeRange{Min: 0.1, Max: 20.0},
			UserBehaviorPattern: UserBehaviorPattern{
				BuyRatio:         0.5,
				SellRatio:        0.5,
				MarketOrderRatio: 0.7,
				LimitOrderRatio:  0.25,
				CancelRatio:      0.3,
			},
			RampUp: RampUpConfig{
				Enabled:      true,
				Duration:     60 * time.Second,
				StartPercent: 50,
				EndPercent:   100,
			},
			Metrics: MetricsConfig{
				CollectLatency:     true,
				CollectThroughput:  true,
				CollectErrorRate:   true,
				CollectResourceUse: true,
				SampleRate:         100,
			},
		},
	}
}

// GetAvailableChaosScenarios returns predefined chaos test scenarios
func (sm *StandardScenarioManager) GetAvailableChaosScenarios() []ChaosTestScenario {
	return []ChaosTestScenario{
		{
			Name:        "Latency Injection - Mild",
			Description: "Inject mild latency to test system resilience",
			Type:        ChaosLatencyInjection,
			Duration:    2 * time.Minute,
			Severity:    ChaosLow,
			Target: ChaosTarget{
				Component:  "trading_engine",
				Services:   []string{"order_service"},
				Endpoints:  []string{"/api/orders"},
				Percentage: 25.0,
			},
			Parameters: ChaosParams{
				LatencyMs: 100,
			},
			Recovery: RecoveryConfig{
				AutoRecover:     true,
				RecoveryTime:    30 * time.Second,
				GracefulRecover: true,
			},
		},
		{
			Name:        "Error Simulation - Moderate",
			Description: "Simulate random errors to test error handling",
			Type:        ChaosErrorSimulation,
			Duration:    3 * time.Minute,
			Severity:    ChaosMedium,
			Target: ChaosTarget{
				Component:  "order_service",
				Services:   []string{"order_placement", "order_cancellation"},
				Endpoints:  []string{"/api/orders", "/api/orders/:id"},
				Percentage: 10.0,
			},
			Parameters: ChaosParams{
				ErrorRate: 0.05,
			},
			Recovery: RecoveryConfig{
				AutoRecover:     true,
				RecoveryTime:    15 * time.Second,
				GracefulRecover: true,
			},
		},
		{
			Name:        "Resource Exhaustion - CPU",
			Description: "Simulate high CPU usage to test resource management",
			Type:        ChaosResourceExhaustion,
			Duration:    90 * time.Second,
			Severity:    ChaosHigh,
			Target: ChaosTarget{
				Component:  "system",
				Services:   []string{"all"},
				Percentage: 100.0,
			},
			Parameters: ChaosParams{
				CPULimitPercent: 80.0,
			},
			Recovery: RecoveryConfig{
				AutoRecover:     true,
				RecoveryTime:    30 * time.Second,
				GracefulRecover: false,
			},
		},
		{
			Name:        "Memory Pressure",
			Description: "Simulate memory pressure to test memory management",
			Type:        ChaosResourceExhaustion,
			Duration:    2 * time.Minute,
			Severity:    ChaosHigh,
			Target: ChaosTarget{
				Component:  "system",
				Services:   []string{"all"},
				Percentage: 100.0,
			},
			Parameters: ChaosParams{
				MemoryLimitMB: 100,
			},
			Recovery: RecoveryConfig{
				AutoRecover:     true,
				RecoveryTime:    45 * time.Second,
				GracefulRecover: true,
			},
		},
	}
}

// Private methods

func (sm *StandardScenarioManager) initializeChaosInjectors() {
	sm.chaosInjectors[string(ChaosLatencyInjection)] = NewLatencyInjector(sm.logger)
	sm.chaosInjectors[string(ChaosErrorSimulation)] = NewErrorInjector(sm.logger)
	sm.chaosInjectors[string(ChaosResourceExhaustion)] = NewResourceInjector(sm.logger)
}

func (sm *StandardScenarioManager) runLoadGenerator(ctx context.Context, generator *LoadGenerator) {
	sm.logger.Info("Starting load generator", "id", generator.ID)

	ticker := time.NewTicker(time.Second / time.Duration(generator.Scenario.OrdersPerSecond/generator.Scenario.ConcurrentUsers))
	defer ticker.Stop()

	for {
		select {
		case <-generator.StopChan:
			sm.logger.Info("Load generator stopped", "id", generator.ID)
			return
		case <-ctx.Done():
			sm.logger.Info("Load generator stopped by context", "id", generator.ID)
			return
		case <-ticker.C:
			if err := sm.generateOrder(generator); err != nil {
				generator.ErrorCount++
				sm.logger.Warn("Failed to generate order", "generator", generator.ID, "error", err)
			} else {
				generator.OrderCounter++
			}
		}
	}
}

func (sm *StandardScenarioManager) generateOrder(generator *LoadGenerator) error {
	scenario := generator.Scenario
	orderNum := generator.OrderCounter

	// Generate random order based on scenario parameters
	symbol := scenario.Symbols[rand.Intn(len(scenario.Symbols))]

	// Determine side based on behavior pattern
	side := "buy"
	if rand.Float64() > scenario.UserBehaviorPattern.BuyRatio {
		side = "sell"
	}

	// Determine order type
	orderType := "limit"
	if rand.Float64() < scenario.UserBehaviorPattern.MarketOrderRatio {
		orderType = "market"
	}

	// Generate quantity
	quantity := scenario.VolumeRange.Min +
		rand.Float64()*(scenario.VolumeRange.Max-scenario.VolumeRange.Min)

	// Generate price with variation
	basePrice := 100.0 + float64(orderNum%1000) // Simulate price movement
	priceVariation := (rand.Float64()*2 - 1) * scenario.PriceVariation
	price := basePrice * (1 + priceVariation)

	order := dto.PlaceOrderRequest{
		Symbol:   symbol,
		Side:     side,
		Type:     orderType,
		Quantity: quantity,
		Price:    price,
	}

	_, err := generator.TradingEngine.PlaceOrder(order)
	return err
}

// Chaos Injectors Implementation

// LatencyInjector implements latency chaos injection
type LatencyInjector struct {
	logger       *slog.Logger
	active       bool
	startTime    time.Time
	latencyMs    int
	stopChan     chan struct{}
}

func NewLatencyInjector(logger *slog.Logger) *LatencyInjector {
	return &LatencyInjector{
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

func (li *LatencyInjector) Inject(ctx context.Context, params ChaosParams) error {
	li.logger.Info("Starting latency injection", "latency_ms", params.LatencyMs)

	li.active = true
	li.startTime = time.Now()
	li.latencyMs = params.LatencyMs

	// Simulate latency injection
	go func() {
		for {
			select {
			case <-li.stopChan:
				return
			case <-time.After(time.Duration(li.latencyMs) * time.Millisecond):
				// Simulate latency by sleeping
				time.Sleep(time.Duration(rand.Intn(li.latencyMs)) * time.Millisecond)
			}
		}
	}()

	return nil
}

func (li *LatencyInjector) Stop(ctx context.Context) error {
	li.logger.Info("Stopping latency injection")

	li.active = false
	close(li.stopChan)
	li.stopChan = make(chan struct{})

	return nil
}

func (li *LatencyInjector) GetStatus() ChaosInjectionStatus {
	return ChaosInjectionStatus{
		Type:      ChaosLatencyInjection,
		Active:    li.active,
		StartTime: li.startTime,
		Metrics: map[string]float64{
			"injected_latency_ms": float64(li.latencyMs),
		},
	}
}

// ErrorInjector implements error chaos injection
type ErrorInjector struct {
	logger       *slog.Logger
	active       bool
	startTime    time.Time
	errorRate    float64
	errorCount   int
	stopChan     chan struct{}
}

func NewErrorInjector(logger *slog.Logger) *ErrorInjector {
	return &ErrorInjector{
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

func (ei *ErrorInjector) Inject(ctx context.Context, params ChaosParams) error {
	ei.logger.Info("Starting error injection", "error_rate", params.ErrorRate)

	ei.active = true
	ei.startTime = time.Now()
	ei.errorRate = params.ErrorRate
	ei.errorCount = 0

	// Simulate error injection
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ei.stopChan:
				return
			case <-ticker.C:
				if rand.Float64() < ei.errorRate {
					ei.errorCount++
					ei.logger.Debug("Chaos error injected", "total_errors", ei.errorCount)
				}
			}
		}
	}()

	return nil
}

func (ei *ErrorInjector) Stop(ctx context.Context) error {
	ei.logger.Info("Stopping error injection", "total_errors_injected", ei.errorCount)

	ei.active = false
	close(ei.stopChan)
	ei.stopChan = make(chan struct{})

	return nil
}

func (ei *ErrorInjector) GetStatus() ChaosInjectionStatus {
	return ChaosInjectionStatus{
		Type:      ChaosErrorSimulation,
		Active:    ei.active,
		StartTime: ei.startTime,
		Metrics: map[string]float64{
			"error_rate":      ei.errorRate,
			"errors_injected": float64(ei.errorCount),
		},
	}
}

// ResourceInjector implements resource exhaustion chaos injection
type ResourceInjector struct {
	logger       *slog.Logger
	active       bool
	startTime    time.Time
	cpuWorkers   []chan struct{}
	memoryBlocks [][]byte
	stopChan     chan struct{}
}

func NewResourceInjector(logger *slog.Logger) *ResourceInjector {
	return &ResourceInjector{
		logger:       logger,
		cpuWorkers:   []chan struct{}{},
		memoryBlocks: [][]byte{},
		stopChan:     make(chan struct{}),
	}
}

func (ri *ResourceInjector) Inject(ctx context.Context, params ChaosParams) error {
	ri.logger.Info("Starting resource exhaustion", "cpu_limit", params.CPULimitPercent, "memory_limit", params.MemoryLimitMB)

	ri.active = true
	ri.startTime = time.Now()

	// CPU exhaustion
	if params.CPULimitPercent > 0 {
		numWorkers := runtime.NumCPU()
		ri.cpuWorkers = make([]chan struct{}, numWorkers)

		for i := 0; i < numWorkers; i++ {
			worker := make(chan struct{})
			ri.cpuWorkers[i] = worker

			go func(workerID int, stop chan struct{}) {
				for {
					select {
					case <-stop:
						return
					case <-ri.stopChan:
						return
					default:
						// Consume CPU cycles
						for j := 0; j < 1000000; j++ {
							_ = j * j
						}
						time.Sleep(time.Microsecond * time.Duration(100-params.CPULimitPercent))
					}
				}
			}(i, worker)
		}
	}

	// Memory exhaustion
	if params.MemoryLimitMB > 0 {
		blockSize := 1024 * 1024 // 1MB blocks
		numBlocks := params.MemoryLimitMB

		for i := 0; i < numBlocks; i++ {
			block := make([]byte, blockSize)
			// Fill with random data to prevent optimization
			for j := range block {
				block[j] = byte(rand.Intn(256))
			}
			ri.memoryBlocks = append(ri.memoryBlocks, block)

			// Small delay to allow monitoring
			time.Sleep(10 * time.Millisecond)
		}
	}

	return nil
}

func (ri *ResourceInjector) Stop(ctx context.Context) error {
	ri.logger.Info("Stopping resource exhaustion")

	ri.active = false

	// Stop CPU workers
	for _, worker := range ri.cpuWorkers {
		close(worker)
	}
	ri.cpuWorkers = []chan struct{}{}

	// Release memory blocks
	ri.memoryBlocks = [][]byte{}
	runtime.GC() // Force garbage collection

	close(ri.stopChan)
	ri.stopChan = make(chan struct{})

	return nil
}

func (ri *ResourceInjector) GetStatus() ChaosInjectionStatus {
	var memUsageMB float64
	if len(ri.memoryBlocks) > 0 {
		memUsageMB = float64(len(ri.memoryBlocks))
	}

	return ChaosInjectionStatus{
		Type:            ChaosResourceExhaustion,
		Active:          ri.active,
		StartTime:       ri.startTime,
		TargetsAffected: len(ri.cpuWorkers),
		Metrics: map[string]float64{
			"cpu_workers":    float64(len(ri.cpuWorkers)),
			"memory_usage_mb": memUsageMB,
		},
	}
}

// MetricsCollector Implementation

func NewMetricsCollector(config DemoMetricsConfig, logger *slog.Logger) *MetricsCollector {
	return &MetricsCollector{
		config:    config,
		logger:    logger,
		startTime: time.Now(),
	}
}

func (mc *MetricsCollector) CollectSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate GC pause time
	gcPause := time.Duration(m.PauseNs[(m.NumGC+255)%256])
	gcPauseDiff := gcPause - mc.lastGCPause
	mc.lastGCPause = gcPause

	metrics := SystemMetrics{
		CPU:        mc.getCPUUsage(),
		Memory:     float64(m.Alloc) / 1024 / 1024, // MB
		DiskIO:     mc.getDiskIOUsage(),
		NetworkIO:  mc.getNetworkIOUsage(),
		Goroutines: runtime.NumGoroutine(),
		HeapSize:   int64(m.HeapAlloc),
		GCPauses:   float64(gcPauseDiff.Nanoseconds()) / 1000000, // ms
	}

	mc.lastCollectionTime = time.Now()
	return metrics
}

func (mc *MetricsCollector) GetCurrentMetrics() SystemMetrics {
	return mc.CollectSystemMetrics()
}

func (mc *MetricsCollector) Reset() {
	mc.startTime = time.Now()
	mc.lastGCPause = 0
	mc.lastCollectionTime = time.Time{}
}

// Simplified metric collection methods
func (mc *MetricsCollector) getCPUUsage() float64 {
	// Simplified CPU usage calculation
	// In a real implementation, this would use system calls or external libraries
	return rand.Float64() * 100
}

func (mc *MetricsCollector) getDiskIOUsage() float64 {
	// Simplified disk I/O calculation
	return rand.Float64() * 100
}

func (mc *MetricsCollector) getNetworkIOUsage() float64 {
	// Simplified network I/O calculation
	return rand.Float64() * 100
}