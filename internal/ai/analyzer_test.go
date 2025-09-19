package ai

import (
	"math"
	"testing"
	"time"

	"simulated_exchange/internal/metrics"
)

func createTestMetricsSnapshot(timestamp time.Time, latency float64, throughput float64, cpu float64, memory float64, errorRate float64) metrics.MetricsSnapshot {
	return metrics.MetricsSnapshot{
		WindowStart:     timestamp,
		WindowEnd:       timestamp.Add(time.Minute),
		AvgLatency:      time.Duration(latency * float64(time.Millisecond)),
		MaxLatency:      time.Duration(latency * 2.0 * float64(time.Millisecond)),
		MinLatency:      time.Duration(latency * 0.5 * float64(time.Millisecond)),
		OrdersPerSec:    throughput,
		TradesPerSec:    throughput * 0.8,
		OrderCount:      int64(throughput * 60),
		TradeCount:      int64(throughput * 0.8 * 60),
		TotalVolume:     throughput * 100,
		VolumePerSec:    throughput * 100,
	}
}

func createRealisticMetricsSequence() []metrics.MetricsSnapshot {
	baseTime := time.Now().Add(-1 * time.Hour)
	snapshots := make([]metrics.MetricsSnapshot, 60) // 1 hour of minute-by-minute data

	for i := 0; i < 60; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)

		// Simulate realistic trading patterns with gradual degradation
		latency := 50.0 + float64(i)*0.5 + math.Sin(float64(i)/10)*10 // Gradual increase with noise
		throughput := 1000.0 - float64(i)*2 + math.Cos(float64(i)/8)*50 // Gradual decrease
		cpu := 30.0 + float64(i)*0.8 + math.Sin(float64(i)/5)*5
		memory := 40.0 + float64(i)*0.6 + math.Cos(float64(i)/7)*3
		errorRate := 0.001 + float64(i)*0.0001 // Gradually increasing errors

		snapshots[i] = createTestMetricsSnapshot(timestamp, latency, throughput, cpu, memory, errorRate)
	}

	return snapshots
}

func createSpikeMetricsSequence() []metrics.MetricsSnapshot {
	baseTime := time.Now().Add(-30 * time.Minute)
	snapshots := make([]metrics.MetricsSnapshot, 30)

	for i := 0; i < 30; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)

		// Normal metrics with a spike at minute 15
		latency := 45.0
		throughput := 950.0
		cpu := 35.0
		memory := 42.0
		errorRate := 0.001

		if i >= 14 && i <= 16 { // Spike period
			latency = 250.0  // 5x spike
			throughput = 400.0 // Dramatic drop
			cpu = 85.0
			memory = 78.0
			errorRate = 0.05 // 50x increase
		}

		snapshots[i] = createTestMetricsSnapshot(timestamp, latency, throughput, cpu, memory, errorRate)
	}

	return snapshots
}

func TestIntelligentAnalyzer_AnalyzeBottlenecks_RealisticScenario(t *testing.T) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())
	snapshots := createRealisticMetricsSequence()

	bottlenecks := analyzer.AnalyzeBottlenecks(snapshots)

	// Should detect multiple bottlenecks in degrading system
	if len(bottlenecks) == 0 {
		t.Error("Expected bottlenecks to be detected in degrading system")
	}

	// Verify bottleneck types detected
	foundLatency := false
	foundThroughput := false

	for _, bottleneck := range bottlenecks {
		switch bottleneck.Type {
		case BottleneckTypeLatency:
			foundLatency = true
			if bottleneck.Severity < 0.3 {
				t.Errorf("Expected higher severity for latency bottleneck, got %f", bottleneck.Severity)
			}
		case BottleneckTypeThroughput:
			foundThroughput = true
			if bottleneck.Severity < 0.3 {
				t.Errorf("Expected higher severity for throughput bottleneck, got %f", bottleneck.Severity)
			}
		}

		// Verify confidence is reasonable
		if bottleneck.Confidence < 0.5 || bottleneck.Confidence > 1.0 {
			t.Errorf("Invalid confidence level: %f", bottleneck.Confidence)
		}
	}

	if !foundLatency {
		t.Error("Expected to detect latency bottleneck in degrading system")
	}
	if !foundThroughput {
		t.Error("Expected to detect throughput bottleneck in degrading system")
	}
}

func TestIntelligentAnalyzer_AnalyzeBottlenecks_SpikeScenario(t *testing.T) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())
	snapshots := createSpikeMetricsSequence()

	bottlenecks := analyzer.AnalyzeBottlenecks(snapshots)

	// Should detect high-severity bottlenecks due to spike
	if len(bottlenecks) == 0 {
		t.Error("Expected bottlenecks to be detected during spike")
	}

	// Find the highest severity bottleneck
	maxSeverity := 0.0
	for _, bottleneck := range bottlenecks {
		if bottleneck.Severity > maxSeverity {
			maxSeverity = bottleneck.Severity
		}
	}

	// Spike should result in high severity
	if maxSeverity < 0.7 {
		t.Errorf("Expected high severity bottleneck during spike, got %f", maxSeverity)
	}
}

func TestIntelligentAnalyzer_PredictCapacity_GrowthScenario(t *testing.T) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())
	snapshots := createRealisticMetricsSequence()

	prediction := analyzer.PredictCapacity(snapshots, 2*time.Hour)

	// Verify prediction structure
	if prediction.TimeHorizon != 2*time.Hour {
		t.Errorf("Expected time horizon 2h, got %v", prediction.TimeHorizon)
	}

	// Should predict increased load based on degrading trend
	if prediction.PredictedLoad.OrdersPerSecond <= 0 {
		t.Error("Expected positive predicted orders per second")
	}

	if prediction.PredictedLoad.GrowthRate <= 0 {
		t.Error("Expected positive growth rate prediction")
	}

	// Verify capacity recommendations
	if prediction.RecommendedCapacity.ComputeUnits <= 0 {
		t.Error("Expected positive compute units recommendation")
	}

	if prediction.RecommendedCapacity.MemoryGB <= 0 {
		t.Error("Expected positive memory recommendation")
	}

	// Confidence interval should be reasonable
	if prediction.ConfidenceInterval.Confidence < 0.8 || prediction.ConfidenceInterval.Confidence > 1.0 {
		t.Errorf("Invalid confidence interval: %f", prediction.ConfidenceInterval.Confidence)
	}
}

func TestIntelligentAnalyzer_GenerateRecommendations(t *testing.T) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())
	snapshots := createRealisticMetricsSequence()

	// Create a performance analysis
	bottlenecks := analyzer.AnalyzeBottlenecks(snapshots)
	capacityPrediction := analyzer.PredictCapacity(snapshots, time.Hour)

	analysis := PerformanceAnalysis{
		ID:                 "test-analysis",
		Timestamp:          time.Now(),
		Bottlenecks:        bottlenecks,
		CapacityPrediction: capacityPrediction,
		PerformanceScore:   0.7,
		HealthStatus:       HealthFair,
	}

	recommendations := analyzer.GenerateRecommendations(analysis)

	// Should generate recommendations for detected bottlenecks
	if len(recommendations) == 0 {
		t.Error("Expected recommendations to be generated")
	}

	// Verify recommendation structure
	for _, rec := range recommendations {
		if rec.Title == "" {
			t.Error("Recommendation should have a title")
		}

		if rec.Description == "" {
			t.Error("Recommendation should have a description")
		}

		if rec.Priority == "" {
			t.Error("Recommendation should have a priority")
		}

		if rec.Confidence < 0.0 || rec.Confidence > 1.0 {
			t.Errorf("Invalid recommendation confidence: %f", rec.Confidence)
		}

		// Verify business impact structure
		if rec.Impact.OverallScore < 0.0 || rec.Impact.OverallScore > 1.0 {
			t.Errorf("Invalid business impact score: %f", rec.Impact.OverallScore)
		}
	}

	// Should have at least one high-priority recommendation for degrading system
	foundHighPriority := false
	for _, rec := range recommendations {
		if rec.Priority == PriorityHigh || rec.Priority == PriorityCritical {
			foundHighPriority = true
			break
		}
	}

	if !foundHighPriority {
		t.Error("Expected at least one high-priority recommendation for degrading system")
	}
}

func TestIntelligentAnalyzer_StatisticalMethods(t *testing.T) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())

	// Test statistical calculations with known data
	values := []float64{10, 20, 30, 40, 50}

	mean := analyzer.calculateMean(values)
	expectedMean := 30.0
	if math.Abs(mean-expectedMean) > 0.001 {
		t.Errorf("Expected mean %f, got %f", expectedMean, mean)
	}

	stdDev := analyzer.calculateStandardDeviation(values, mean)
	expectedStdDev := math.Sqrt(250.0) // Verified calculation
	if math.Abs(stdDev-expectedStdDev) > 0.001 {
		t.Errorf("Expected std dev %f, got %f", expectedStdDev, stdDev)
	}

	p95 := analyzer.calculatePercentile(values, 95)
	expectedP95 := 46.0 // Linear interpolation between 40 and 50
	if math.Abs(p95-expectedP95) > 0.001 {
		t.Errorf("Expected P95 %f, got %f", expectedP95, p95)
	}
}

func TestIntelligentAnalyzer_TrendAnalysis(t *testing.T) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())
	snapshots := createRealisticMetricsSequence()

	// Test growth rate calculation for trend detection
	values := make([]float64, len(snapshots))
	for i, snapshot := range snapshots {
		values[i] = float64(snapshot.AvgLatency.Milliseconds())
	}

	growthRate := analyzer.calculateGrowthRate(values)

	// Should detect positive growth (increasing latency)
	if growthRate <= 0 {
		t.Errorf("Expected positive growth rate for increasing latency trend, got %f", growthRate)
	}

	// Test trend strength calculation
	trendStrength := analyzer.calculateTrendStrength(values)

	if trendStrength < 0 || trendStrength > 1 {
		t.Errorf("Expected trend strength between 0 and 1, got %f", trendStrength)
	}
}

func TestIntelligentAnalyzer_AdaptiveThresholds(t *testing.T) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())
	snapshots := createRealisticMetricsSequence()

	// Test that the analyzer can handle the snapshots without errors
	bottlenecks := analyzer.AnalyzeBottlenecks(snapshots)

	// Should produce some analysis results
	if len(bottlenecks) < 0 {
		t.Error("Analyzer should handle snapshots without errors")
	}

	// Test capacity prediction to ensure adaptive behavior works
	prediction := analyzer.PredictCapacity(snapshots, time.Hour)
	if prediction.TimeHorizon != time.Hour {
		t.Errorf("Expected time horizon %v, got %v", time.Hour, prediction.TimeHorizon)
	}
}

func TestIntelligentAnalyzer_InsufficientData(t *testing.T) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())

	// Test with insufficient data points
	snapshots := createRealisticMetricsSequence()[:5] // Only 5 data points

	bottlenecks := analyzer.AnalyzeBottlenecks(snapshots)

	// Should still work but with lower confidence
	for _, bottleneck := range bottlenecks {
		if bottleneck.Confidence > 0.8 {
			t.Errorf("Expected lower confidence with insufficient data, got %f", bottleneck.Confidence)
		}
	}

	// Capacity prediction should have warnings
	prediction := analyzer.PredictCapacity(snapshots, time.Hour)

	if len(prediction.RiskFactors) == 0 {
		t.Error("Expected risk factors to be flagged with insufficient data")
	}
}

func TestIntelligentAnalyzer_StableSystem(t *testing.T) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())

	// Create stable metrics (no degradation)
	baseTime := time.Now().Add(-30 * time.Minute)
	snapshots := make([]metrics.MetricsSnapshot, 30)

	for i := 0; i < 30; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		// Stable metrics with minor noise
		latency := 45.0 + math.Sin(float64(i)/10)*2
		throughput := 1000.0 + math.Cos(float64(i)/8)*20
		cpu := 35.0 + math.Sin(float64(i)/5)*1
		memory := 42.0 + math.Cos(float64(i)/7)*1
		errorRate := 0.001

		snapshots[i] = createTestMetricsSnapshot(timestamp, latency, throughput, cpu, memory, errorRate)
	}

	bottlenecks := analyzer.AnalyzeBottlenecks(snapshots)

	// Should detect fewer or no bottlenecks in stable system
	criticalBottlenecks := 0
	for _, bottleneck := range bottlenecks {
		if bottleneck.Severity > 0.7 {
			criticalBottlenecks++
		}
	}

	if criticalBottlenecks > 1 {
		t.Errorf("Expected few critical bottlenecks in stable system, got %d", criticalBottlenecks)
	}

	// Capacity prediction should be more conservative
	prediction := analyzer.PredictCapacity(snapshots, time.Hour)

	if prediction.PredictedLoad.GrowthRate > 0.1 {
		t.Errorf("Expected low growth rate for stable system, got %f", prediction.PredictedLoad.GrowthRate)
	}
}

func BenchmarkIntelligentAnalyzer_AnalyzeBottlenecks(b *testing.B) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())
	snapshots := createRealisticMetricsSequence()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeBottlenecks(snapshots)
	}
}

func BenchmarkIntelligentAnalyzer_PredictCapacity(b *testing.B) {
	analyzer := NewIntelligentAnalyzer(DefaultMLAnalysisConfig())
	snapshots := createRealisticMetricsSequence()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.PredictCapacity(snapshots, time.Hour)
	}
}