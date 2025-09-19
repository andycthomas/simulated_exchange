package ai

import (
	"math"
	"testing"
	"time"

	"simulated_exchange/internal/metrics"
)

// MockMetricsCollector simulates the real metrics collection system
type MockMetricsCollector struct {
	snapshots []metrics.MetricsSnapshot
	interval  time.Duration
}

func NewMockMetricsCollector(interval time.Duration) *MockMetricsCollector {
	return &MockMetricsCollector{
		snapshots: make([]metrics.MetricsSnapshot, 0),
		interval:  interval,
	}
}

func (m *MockMetricsCollector) CollectMetrics() metrics.MetricsSnapshot {
	timestamp := time.Now()

	// Simulate realistic trading system metrics with some variance
	baseLatency := 50.0 + math.Sin(float64(len(m.snapshots))/10)*5
	baseThroughput := 1000.0 + math.Cos(float64(len(m.snapshots))/8)*50
	// baseCPU := 35.0 + math.Sin(float64(len(m.snapshots))/5)*3
	// baseMemory := 45.0 + math.Cos(float64(len(m.snapshots))/7)*2
	// baseErrorRate := 0.001 + math.Sin(float64(len(m.snapshots))/15)*0.0005

	snapshot := metrics.MetricsSnapshot{
		WindowStart:      timestamp,
		WindowEnd:        timestamp.Add(time.Minute),
		AvgLatency:       time.Duration(baseLatency * float64(time.Millisecond)),
		MaxLatency:       time.Duration(baseLatency * 2.5 * float64(time.Millisecond)),
		MinLatency:       time.Duration(baseLatency * 0.5 * float64(time.Millisecond)),
		OrdersPerSec:     baseThroughput,
		TradesPerSec:     baseThroughput * 0.75,
		OrderCount:       int64(baseThroughput * 60),
		TradeCount:       int64(baseThroughput * 0.75 * 60),
		TotalVolume:      baseThroughput * 100,
		VolumePerSec:     baseThroughput * 100,
	}

	m.snapshots = append(m.snapshots, snapshot)
	return snapshot
}

func (m *MockMetricsCollector) GetRecentSnapshots(duration time.Duration) []metrics.MetricsSnapshot {
	cutoff := time.Now().Add(-duration)
	recent := make([]metrics.MetricsSnapshot, 0)

	for _, snapshot := range m.snapshots {
		if snapshot.WindowStart.After(cutoff) {
			recent = append(recent, snapshot)
		}
	}

	return recent
}

func (m *MockMetricsCollector) SimulatePerformanceDegradation() {
	// Simulate gradual performance degradation
	for i := 0; i < 20; i++ {
		timestamp := time.Now().Add(time.Duration(i) * m.interval)

		// Gradually degrading metrics
		degradationFactor := float64(i) * 0.1
		latency := 50.0 + degradationFactor*10
		throughput := 1000.0 - degradationFactor*50
		// cpu := 35.0 + degradationFactor*2
		// memory := 45.0 + degradationFactor*1.5
		// errorRate := 0.001 + degradationFactor*0.002

		snapshot := metrics.MetricsSnapshot{
			WindowStart:      timestamp,
			WindowEnd:        timestamp.Add(time.Minute),
			AvgLatency:       time.Duration(latency * float64(time.Millisecond)),
			MaxLatency:       time.Duration(latency * 2.5 * float64(time.Millisecond)),
			MinLatency:       time.Duration(latency * 0.5 * float64(time.Millisecond)),
			OrdersPerSec:     throughput,
			TradesPerSec:     throughput * 0.75,
			OrderCount:       int64(throughput * 60),
			TradeCount:       int64(throughput * 0.75 * 60),
			TotalVolume:      throughput * 100,
			VolumePerSec:     throughput * 100,
		}

		m.snapshots = append(m.snapshots, snapshot)
	}
}

func (m *MockMetricsCollector) SimulateTrafficSpike() {
	// Simulate sudden traffic spike
	for i := 0; i < 10; i++ {
		timestamp := time.Now().Add(time.Duration(i) * m.interval)

		var latency, throughput float64
		// var cpu, memory, errorRate float64

		if i >= 3 && i <= 6 { // Spike period
			latency = 200.0
			throughput = 2500.0 // High throughput during spike
			// cpu = 85.0
			// memory = 78.0
			// errorRate = 0.015
		} else {
			latency = 50.0
			throughput = 1000.0
			// cpu = 35.0
			// memory = 45.0
			// errorRate = 0.001
		}

		snapshot := metrics.MetricsSnapshot{
			WindowStart:      timestamp,
			WindowEnd:        timestamp.Add(time.Minute),
			AvgLatency:       time.Duration(latency * float64(time.Millisecond)),
			MaxLatency:       time.Duration(latency * 2.5 * float64(time.Millisecond)),
			MinLatency:       time.Duration(latency * 0.5 * float64(time.Millisecond)),
			OrdersPerSec:     throughput,
			TradesPerSec:     throughput * 0.75,
			OrderCount:       int64(throughput * 60),
			TradeCount:       int64(throughput * 0.75 * 60),
			TotalVolume:      throughput * 100,
			VolumePerSec:     throughput * 100,
		}

		m.snapshots = append(m.snapshots, snapshot)
	}
}

// PerformanceMonitoringSystem integrates metrics collection with AI analysis
type PerformanceMonitoringSystem struct {
	metricsCollector *MockMetricsCollector
	analyzer         PerformanceAI
	businessCalc     BusinessImpactCalculator
	reportGenerator  ReportGenerator
	analysisHistory  []PerformanceAnalysis
}

func NewPerformanceMonitoringSystem() *PerformanceMonitoringSystem {
	return &PerformanceMonitoringSystem{
		metricsCollector: NewMockMetricsCollector(time.Minute),
		analyzer:         NewIntelligentAnalyzer(DefaultMLAnalysisConfig()),
		businessCalc:     NewROICalculator(DefaultROIConfig(), DefaultMarketParameters(), DefaultCostModel()),
		reportGenerator:  NewExecutiveReportGenerator(DefaultReportConfig()),
		analysisHistory:  make([]PerformanceAnalysis, 0),
	}
}

func (pms *PerformanceMonitoringSystem) RunAnalysisCycle() PerformanceAnalysis {
	// Collect recent metrics
	snapshots := pms.metricsCollector.GetRecentSnapshots(30 * time.Minute)

	if len(snapshots) < 5 {
		// Not enough data for analysis
		return PerformanceAnalysis{}
	}

	// Perform AI analysis
	bottlenecks := pms.analyzer.AnalyzeBottlenecks(snapshots)
	capacityPrediction := pms.analyzer.PredictCapacity(snapshots, 2*time.Hour)

	// Calculate performance score based on recent metrics
	performanceScore := pms.calculatePerformanceScore(snapshots)
	healthStatus := pms.determineHealthStatus(performanceScore, bottlenecks)

	// Create comprehensive analysis
	analysis := PerformanceAnalysis{
		ID:                 generateAnalysisID(),
		Timestamp:          time.Now(),
		TimeRange:          TimeRange{Start: snapshots[0].WindowStart, End: snapshots[len(snapshots)-1].WindowEnd},
		Bottlenecks:        bottlenecks,
		CapacityPrediction: capacityPrediction,
		PerformanceScore:   performanceScore,
		HealthStatus:       healthStatus,
		Confidence:         pms.calculateOverallConfidence(bottlenecks, capacityPrediction),
	}

	// Generate recommendations
	analysis.Recommendations = pms.analyzer.GenerateRecommendations(analysis)

	// Calculate business impact for recommendations
	if len(snapshots) > 0 {
		currentMetrics := snapshots[len(snapshots)-1]
		for i, rec := range analysis.Recommendations {
			roiAnalysis := pms.businessCalc.CalculateROI(rec, currentMetrics)
			analysis.Recommendations[i].Impact = BusinessImpact{
				Revenue:      roiAnalysis.AnnualSavings * 0.3,
				Cost:         roiAnalysis.InitialInvestment,
				OverallScore: math.Min(roiAnalysis.ROIPercentage/100, 1.0),
			}
		}
	}

	pms.analysisHistory = append(pms.analysisHistory, analysis)
	return analysis
}

func (pms *PerformanceMonitoringSystem) calculatePerformanceScore(snapshots []metrics.MetricsSnapshot) float64 {
	if len(snapshots) == 0 {
		return 0.0
	}

	recent := snapshots[len(snapshots)-1]

	// Score based on key metrics (0.0 to 1.0)
	latencyMs := float64(recent.AvgLatency.Milliseconds())
	latencyScore := math.Max(0, (150-latencyMs)/150) // Good under 150ms
	throughputScore := math.Min(1, recent.OrdersPerSec/1500)  // Target 1500 TPS
	// Note: ErrorRate and CPUUsage are not in MetricsSnapshot, using simplified scoring

	// Weighted average (simplified without unavailable metrics)
	return (latencyScore*0.5 + throughputScore*0.5)
}

func (pms *PerformanceMonitoringSystem) determineHealthStatus(score float64, bottlenecks []Bottleneck) HealthStatus {
	criticalBottlenecks := 0
	for _, b := range bottlenecks {
		if b.Severity > 0.8 {
			criticalBottlenecks++
		}
	}

	if criticalBottlenecks > 2 || score < 0.3 {
		return HealthCritical
	} else if criticalBottlenecks > 0 || score < 0.5 {
		return HealthPoor
	} else if score < 0.7 {
		return HealthFair
	} else if score < 0.85 {
		return HealthGood
	}
	return HealthExcellent
}

func (pms *PerformanceMonitoringSystem) calculateOverallConfidence(bottlenecks []Bottleneck, prediction CapacityPrediction) float64 {
	if len(bottlenecks) == 0 {
		return 0.5 // Medium confidence when no issues detected
	}

	totalConfidence := 0.0
	for _, b := range bottlenecks {
		totalConfidence += b.Confidence
	}

	avgBottleneckConfidence := totalConfidence / float64(len(bottlenecks))
	predictionConfidence := prediction.ConfidenceInterval.Confidence

	return (avgBottleneckConfidence + predictionConfidence) / 2.0
}

func generateAnalysisID() string {
	return time.Now().Format("analysis-20060102-150405")
}

// Integration Tests

func TestIntegration_EndToEndPerformanceMonitoring(t *testing.T) {
	system := NewPerformanceMonitoringSystem()

	// Simulate normal operation
	for i := 0; i < 10; i++ {
		system.metricsCollector.CollectMetrics()
		time.Sleep(10 * time.Millisecond) // Simulate time passage
	}

	// Run analysis
	analysis := system.RunAnalysisCycle()

	// Verify analysis was generated
	if analysis.ID == "" {
		t.Error("Expected analysis to have an ID")
	}

	if analysis.Timestamp.IsZero() {
		t.Error("Expected analysis to have a timestamp")
	}

	// Should have reasonable performance score for normal operation
	if analysis.PerformanceScore < 0.5 {
		t.Errorf("Expected decent performance score for normal operation, got %f", analysis.PerformanceScore)
	}

	// Should detect as healthy system
	if analysis.HealthStatus == HealthCritical || analysis.HealthStatus == HealthPoor {
		t.Errorf("Expected healthy status for normal operation, got %s", analysis.HealthStatus)
	}
}

func TestIntegration_PerformanceDegradationDetection(t *testing.T) {
	system := NewPerformanceMonitoringSystem()

	// Simulate performance degradation
	system.metricsCollector.SimulatePerformanceDegradation()

	// Run analysis
	analysis := system.RunAnalysisCycle()

	// Should detect bottlenecks
	if len(analysis.Bottlenecks) == 0 {
		t.Error("Expected bottlenecks to be detected during performance degradation")
	}

	// Should generate recommendations
	if len(analysis.Recommendations) == 0 {
		t.Error("Expected recommendations to be generated for degrading system")
	}

	// Performance score should be lower
	if analysis.PerformanceScore > 0.7 {
		t.Errorf("Expected lower performance score for degrading system, got %f", analysis.PerformanceScore)
	}

	// Health status should indicate issues
	if analysis.HealthStatus == HealthExcellent || analysis.HealthStatus == HealthGood {
		t.Errorf("Expected poor health status for degrading system, got %s", analysis.HealthStatus)
	}

	// Should have high-priority recommendations
	foundHighPriority := false
	for _, rec := range analysis.Recommendations {
		if rec.Priority == PriorityHigh || rec.Priority == PriorityCritical {
			foundHighPriority = true
			break
		}
	}

	if !foundHighPriority {
		t.Error("Expected high-priority recommendations for degrading system")
	}
}

func TestIntegration_TrafficSpikeHandling(t *testing.T) {
	system := NewPerformanceMonitoringSystem()

	// Simulate traffic spike
	system.metricsCollector.SimulateTrafficSpike()

	// Run analysis
	analysis := system.RunAnalysisCycle()

	// Should detect high-severity bottlenecks
	highSeverityBottlenecks := 0
	for _, bottleneck := range analysis.Bottlenecks {
		if bottleneck.Severity > 0.7 {
			highSeverityBottlenecks++
		}
	}

	if highSeverityBottlenecks == 0 {
		t.Error("Expected high-severity bottlenecks during traffic spike")
	}

	// Capacity prediction should recommend scaling
	if analysis.CapacityPrediction.RecommendedCapacity.ComputeUnits <= 1 {
		t.Error("Expected capacity scaling recommendation during traffic spike")
	}

	// Should have scaling recommendations
	foundScalingRec := false
	for _, rec := range analysis.Recommendations {
		if rec.Type == RecommendationTypeScaling {
			foundScalingRec = true
			break
		}
	}

	if !foundScalingRec {
		t.Error("Expected scaling recommendation during traffic spike")
	}
}

func TestIntegration_BusinessImpactCalculation(t *testing.T) {
	system := NewPerformanceMonitoringSystem()

	// Simulate degrading performance
	system.metricsCollector.SimulatePerformanceDegradation()

	// Run analysis
	analysis := system.RunAnalysisCycle()

	// Verify business impact is calculated for recommendations
	for _, rec := range analysis.Recommendations {
		if rec.Impact.OverallScore == 0.0 {
			t.Error("Expected business impact to be calculated for recommendations")
		}

		// ROI should be reasonable
		if rec.Impact.OverallScore < 0.0 || rec.Impact.OverallScore > 1.0 {
			t.Errorf("Invalid business impact score: %f", rec.Impact.OverallScore)
		}
	}

	// Calculate detailed ROI for top recommendation
	if len(analysis.Recommendations) > 0 {
		topRec := analysis.Recommendations[0]
		snapshots := system.metricsCollector.GetRecentSnapshots(30 * time.Minute)

		if len(snapshots) > 0 {
			currentMetrics := snapshots[len(snapshots)-1]
			roi := system.businessCalc.CalculateROI(topRec, currentMetrics)

			if roi.ROIPercentage <= 0 {
				t.Error("Expected positive ROI for optimization recommendation")
			}

			if roi.PaybackPeriod <= 0 {
				t.Error("Expected positive payback period")
			}
		}
	}
}

func TestIntegration_ExecutiveReportGeneration(t *testing.T) {
	system := NewPerformanceMonitoringSystem()

	// Generate some performance issues
	system.metricsCollector.SimulatePerformanceDegradation()

	// Run analysis
	analysis := system.RunAnalysisCycle()

	// Generate executive report in multiple formats
	jsonReport, err := system.reportGenerator.GenerateExecutiveReport(analysis, ReportFormatJSON)
	if err != nil {
		t.Errorf("Failed to generate JSON report: %v", err)
	}

	if len(jsonReport) == 0 {
		t.Error("Expected non-empty JSON report")
	}

	textReport, err := system.reportGenerator.GenerateExecutiveReport(analysis, ReportFormatText)
	if err != nil {
		t.Errorf("Failed to generate text report: %v", err)
	}

	if len(textReport) == 0 {
		t.Error("Expected non-empty text report")
	}

	markdownReport, err := system.reportGenerator.GenerateExecutiveReport(analysis, ReportFormatMarkdown)
	if err != nil {
		t.Errorf("Failed to generate markdown report: %v", err)
	}

	if len(markdownReport) == 0 {
		t.Error("Expected non-empty markdown report")
	}

	// Generate executive summary
	summary := system.reportGenerator.GenerateSummaryReport(analysis)

	if summary.OverallHealth == "" {
		t.Error("Expected executive summary to have health status")
	}

	if summary.PerformanceScore == 0.0 {
		t.Error("Expected executive summary to have performance score")
	}

	if len(summary.KeyRecommendations) == 0 {
		t.Error("Expected executive summary to have key recommendations")
	}
}

func TestIntegration_ContinuousMonitoring(t *testing.T) {
	system := NewPerformanceMonitoringSystem()

	// Simulate continuous monitoring over time
	analyses := make([]PerformanceAnalysis, 0)

	// Phase 1: Normal operation
	for i := 0; i < 5; i++ {
		system.metricsCollector.CollectMetrics()
	}
	analyses = append(analyses, system.RunAnalysisCycle())

	// Phase 2: Performance degradation
	system.metricsCollector.SimulatePerformanceDegradation()
	analyses = append(analyses, system.RunAnalysisCycle())

	// Phase 3: Traffic spike
	system.metricsCollector.SimulateTrafficSpike()
	analyses = append(analyses, system.RunAnalysisCycle())

	// Verify we have multiple analyses
	if len(analyses) != 3 {
		t.Errorf("Expected 3 analyses, got %d", len(analyses))
	}

	// Performance should degrade over time
	if analyses[1].PerformanceScore >= analyses[0].PerformanceScore {
		t.Error("Expected performance score to decrease during degradation")
	}

	// Should detect different types of issues
	foundLatencyBottleneck := false
	foundThroughputBottleneck := false

	for _, analysis := range analyses {
		for _, bottleneck := range analysis.Bottlenecks {
			if bottleneck.Type == BottleneckTypeLatency {
				foundLatencyBottleneck = true
			}
			if bottleneck.Type == BottleneckTypeThroughput {
				foundThroughputBottleneck = true
			}
		}
	}

	if !foundLatencyBottleneck {
		t.Error("Expected to detect latency bottlenecks over monitoring period")
	}

	if !foundThroughputBottleneck {
		t.Error("Expected to detect throughput bottlenecks over monitoring period")
	}

	// Verify analysis history is maintained
	if len(system.analysisHistory) != 3 {
		t.Errorf("Expected analysis history to contain 3 entries, got %d", len(system.analysisHistory))
	}
}

func TestIntegration_RealTimeAlertConditions(t *testing.T) {
	system := NewPerformanceMonitoringSystem()

	// Simulate critical performance issue
	for i := 0; i < 5; i++ {
		timestamp := time.Now().Add(time.Duration(i) * time.Minute)

		// Critical metrics
		snapshot := metrics.MetricsSnapshot{
			WindowStart:      timestamp,
			WindowEnd:        timestamp.Add(time.Minute),
			AvgLatency:       time.Duration(300 * time.Millisecond), // Very high
			MaxLatency:       time.Duration(800 * time.Millisecond),
			MinLatency:       time.Duration(100 * time.Millisecond),
			OrdersPerSec:     200.0, // Very low
			TradesPerSec:     150.0,
			OrderCount:       int64(200 * 60),
			TradeCount:       int64(150 * 60),
			TotalVolume:      200 * 100,
			VolumePerSec:     200 * 100,
		}

		system.metricsCollector.snapshots = append(system.metricsCollector.snapshots, snapshot)
	}

	analysis := system.RunAnalysisCycle()

	// Should detect critical health status
	if analysis.HealthStatus != HealthCritical {
		t.Errorf("Expected critical health status for severe issues, got %s", analysis.HealthStatus)
	}

	// Should have critical priority recommendations
	criticalRecommendations := 0
	for _, rec := range analysis.Recommendations {
		if rec.Priority == PriorityCritical {
			criticalRecommendations++
		}
	}

	if criticalRecommendations == 0 {
		t.Error("Expected critical priority recommendations for severe performance issues")
	}

	// Performance score should be very low
	if analysis.PerformanceScore > 0.3 {
		t.Errorf("Expected very low performance score for critical issues, got %f", analysis.PerformanceScore)
	}
}

func BenchmarkIntegration_FullAnalysisCycle(b *testing.B) {
	system := NewPerformanceMonitoringSystem()

	// Prepare test data
	system.metricsCollector.SimulatePerformanceDegradation()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.RunAnalysisCycle()
	}
}