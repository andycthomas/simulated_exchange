package metrics

import (
	"context"
	"sync"
	"time"
)

// RealTimeMetricsService implements MetricsService for orchestrating metrics collection and analysis
type RealTimeMetricsService struct {
	collector         MetricsCollector
	analyzer          PerformanceAnalyzer
	analysisHistory   []MetricsSnapshot
	maxHistorySize    int
	analysisInterval  time.Duration
	windowSize        time.Duration

	// Control
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mutex     sync.RWMutex
	running   bool
	healthy   bool

	// Latest analysis results
	lastAnalysis PerformanceAnalysis
}

// NewRealTimeMetricsService creates a new metrics service
func NewRealTimeMetricsService(collector MetricsCollector, analyzer PerformanceAnalyzer) *RealTimeMetricsService {
	return &RealTimeMetricsService{
		collector:        collector,
		analyzer:         analyzer,
		analysisHistory:  make([]MetricsSnapshot, 0),
		maxHistorySize:   100, // Keep last 100 snapshots
		analysisInterval: 10 * time.Second,
		windowSize:       60 * time.Second,
		healthy:          true,
	}
}

// NewRealTimeMetricsServiceWithConfig creates a new metrics service with custom configuration
func NewRealTimeMetricsServiceWithConfig(
	collector MetricsCollector,
	analyzer PerformanceAnalyzer,
	maxHistorySize int,
	analysisInterval time.Duration,
	windowSize time.Duration,
) *RealTimeMetricsService {
	return &RealTimeMetricsService{
		collector:        collector,
		analyzer:         analyzer,
		analysisHistory:  make([]MetricsSnapshot, 0),
		maxHistorySize:   maxHistorySize,
		analysisInterval: analysisInterval,
		windowSize:       windowSize,
		healthy:          true,
	}
}

// Start begins the metrics collection and analysis process
func (rms *RealTimeMetricsService) Start() error {
	rms.mutex.Lock()
	defer rms.mutex.Unlock()

	if rms.running {
		return nil // Already running
	}

	rms.ctx, rms.cancel = context.WithCancel(context.Background())
	rms.running = true
	rms.healthy = true

	// Start background analysis routine
	rms.wg.Add(1)
	go rms.analysisRoutine()

	return nil
}

// Stop stops the metrics collection and analysis process
func (rms *RealTimeMetricsService) Stop() error {
	rms.mutex.Lock()
	defer rms.mutex.Unlock()

	if !rms.running {
		return nil // Already stopped
	}

	rms.cancel()
	rms.running = false

	rms.mutex.Unlock()
	rms.wg.Wait() // Wait for background routines to finish
	rms.mutex.Lock()

	return nil
}

// GetRealTimeMetrics returns the current metrics snapshot
func (rms *RealTimeMetricsService) GetRealTimeMetrics() MetricsSnapshot {
	return rms.collector.GetCurrentMetrics()
}

// GetPerformanceAnalysis returns the latest performance analysis
func (rms *RealTimeMetricsService) GetPerformanceAnalysis() PerformanceAnalysis {
	rms.mutex.RLock()
	defer rms.mutex.RUnlock()

	return rms.lastAnalysis
}

// IsHealthy returns whether the metrics service is healthy
func (rms *RealTimeMetricsService) IsHealthy() bool {
	rms.mutex.RLock()
	defer rms.mutex.RUnlock()

	return rms.healthy && rms.running
}

// RecordOrderEvent records an order event through the collector
func (rms *RealTimeMetricsService) RecordOrderEvent(event OrderEvent) {
	if rms.IsHealthy() {
		rms.collector.RecordOrder(event)
	}
}

// RecordTradeEvent records a trade event through the collector
func (rms *RealTimeMetricsService) RecordTradeEvent(event TradeEvent) {
	if rms.IsHealthy() {
		rms.collector.RecordTrade(event)
	}
}

// GetHistoricalSnapshots returns historical metrics snapshots
func (rms *RealTimeMetricsService) GetHistoricalSnapshots() []MetricsSnapshot {
	rms.mutex.RLock()
	defer rms.mutex.RUnlock()

	// Return a copy to avoid race conditions
	history := make([]MetricsSnapshot, len(rms.analysisHistory))
	copy(history, rms.analysisHistory)
	return history
}

// ResetMetrics resets all collected metrics
func (rms *RealTimeMetricsService) ResetMetrics() {
	rms.mutex.Lock()
	defer rms.mutex.Unlock()

	rms.collector.Reset()
	rms.analysisHistory = rms.analysisHistory[:0]
	rms.lastAnalysis = PerformanceAnalysis{}
}

// analysisRoutine runs periodic performance analysis
func (rms *RealTimeMetricsService) analysisRoutine() {
	defer rms.wg.Done()

	ticker := time.NewTicker(rms.analysisInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rms.ctx.Done():
			return
		case <-ticker.C:
			rms.performAnalysis()
		}
	}
}

// performAnalysis conducts performance analysis and updates results
func (rms *RealTimeMetricsService) performAnalysis() {
	defer func() {
		// Recover from any panics to maintain service health
		if r := recover(); r != nil {
			rms.mutex.Lock()
			rms.healthy = false
			rms.mutex.Unlock()
		}
	}()

	// Get current metrics
	currentMetrics := rms.collector.CalculateMetrics(rms.windowSize)

	rms.mutex.Lock()
	defer rms.mutex.Unlock()

	// Add to history
	rms.analysisHistory = append(rms.analysisHistory, currentMetrics)

	// Trim history if it exceeds max size
	if len(rms.analysisHistory) > rms.maxHistorySize {
		// Keep only the most recent entries
		excess := len(rms.analysisHistory) - rms.maxHistorySize
		rms.analysisHistory = rms.analysisHistory[excess:]
	}

	// Perform analysis if we have sufficient history
	if len(rms.analysisHistory) >= 2 {
		latencyAnalysis := rms.analyzer.AnalyzeLatency(rms.analysisHistory)
		throughputPrediction := rms.analyzer.PredictThroughput(rms.analysisHistory)
		bottlenecks := rms.analyzer.DetectBottlenecks(currentMetrics)

		// Create performance analysis
		analysis := PerformanceAnalysis{
			Timestamp:           time.Now(),
			LatencyTrend:        latencyAnalysis.Trend,
			ThroughputTrend:     throughputPrediction.Trend,
			PredictedThroughput: throughputPrediction.PredictedThroughput,
			Bottlenecks:         bottlenecks,
		}

		// Generate recommendations
		analysis.Recommendations = rms.analyzer.GenerateRecommendations(analysis)

		rms.lastAnalysis = analysis
	}

	// Update health status
	rms.healthy = true
}

// GetAnalysisConfig returns the current analysis configuration
func (rms *RealTimeMetricsService) GetAnalysisConfig() AnalysisConfig {
	rms.mutex.RLock()
	defer rms.mutex.RUnlock()

	return AnalysisConfig{
		MaxHistorySize:   rms.maxHistorySize,
		AnalysisInterval: rms.analysisInterval,
		WindowSize:       rms.windowSize,
	}
}

// UpdateAnalysisConfig updates the analysis configuration
func (rms *RealTimeMetricsService) UpdateAnalysisConfig(config AnalysisConfig) {
	rms.mutex.Lock()
	defer rms.mutex.Unlock()

	if config.MaxHistorySize > 0 {
		rms.maxHistorySize = config.MaxHistorySize
	}
	if config.AnalysisInterval > 0 {
		rms.analysisInterval = config.AnalysisInterval
	}
	if config.WindowSize > 0 {
		rms.windowSize = config.WindowSize
	}
}

// AnalysisConfig represents configuration for the analysis process
type AnalysisConfig struct {
	MaxHistorySize   int
	AnalysisInterval time.Duration
	WindowSize       time.Duration
}

// GetDetailedHealthStatus returns detailed health information
func (rms *RealTimeMetricsService) GetDetailedHealthStatus() HealthStatus {
	rms.mutex.RLock()
	defer rms.mutex.RUnlock()

	status := HealthStatus{
		IsRunning:     rms.running,
		IsHealthy:     rms.healthy,
		LastAnalysis:  rms.lastAnalysis.Timestamp,
		HistorySize:   len(rms.analysisHistory),
		MaxHistory:    rms.maxHistorySize,
	}

	// Calculate uptime if running
	if rms.running && rms.ctx != nil {
		// Note: For a proper uptime calculation, you'd want to store the start time
		status.Uptime = time.Since(rms.lastAnalysis.Timestamp)
	}

	return status
}

// HealthStatus represents the health status of the metrics service
type HealthStatus struct {
	IsRunning    bool
	IsHealthy    bool
	LastAnalysis time.Time
	HistorySize  int
	MaxHistory   int
	Uptime       time.Duration
}

// GetMetricsSummary returns a summary of current metrics and analysis
func (rms *RealTimeMetricsService) GetMetricsSummary() MetricsSummary {
	currentMetrics := rms.GetRealTimeMetrics()
	analysis := rms.GetPerformanceAnalysis()
	healthStatus := rms.GetDetailedHealthStatus()

	return MetricsSummary{
		Timestamp:        time.Now(),
		CurrentMetrics:   currentMetrics,
		Analysis:         analysis,
		HealthStatus:     healthStatus,
		ActiveBottlenecks: len(analysis.Bottlenecks),
	}
}

// MetricsSummary represents a comprehensive summary of metrics and analysis
type MetricsSummary struct {
	Timestamp         time.Time
	CurrentMetrics    MetricsSnapshot
	Analysis          PerformanceAnalysis
	HealthStatus      HealthStatus
	ActiveBottlenecks int
}