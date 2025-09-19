package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAIAnalyzer(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		analyzer := NewAIAnalyzer()
		assert.NotNil(t, analyzer)
		assert.Equal(t, 100*time.Millisecond, analyzer.latencyThreshold)
		assert.Equal(t, 1000.0, analyzer.throughputThreshold)
		assert.Equal(t, 5, analyzer.trendWindowSize)
		assert.Equal(t, 0.7, analyzer.bottleneckThreshold)
	})

	t.Run("custom configuration", func(t *testing.T) {
		analyzer := NewAIAnalyzerWithConfig(
			50*time.Millisecond,
			500.0,
			3,
			0.8,
		)
		assert.NotNil(t, analyzer)
		assert.Equal(t, 50*time.Millisecond, analyzer.latencyThreshold)
		assert.Equal(t, 500.0, analyzer.throughputThreshold)
		assert.Equal(t, 3, analyzer.trendWindowSize)
		assert.Equal(t, 0.8, analyzer.bottleneckThreshold)
	})
}

func TestAIAnalyzer_AnalyzeLatency(t *testing.T) {
	analyzer := NewAIAnalyzer()

	t.Run("empty snapshots", func(t *testing.T) {
		analysis := analyzer.AnalyzeLatency([]MetricsSnapshot{})
		assert.Equal(t, TrendDirection(""), analysis.Trend)
		assert.Equal(t, time.Duration(0), analysis.CurrentLatency)
	})

	t.Run("single snapshot", func(t *testing.T) {
		snapshots := []MetricsSnapshot{
			{
				AvgLatency: 50 * time.Millisecond,
			},
		}

		analysis := analyzer.AnalyzeLatency(snapshots)
		assert.Equal(t, TrendFlat, analysis.Trend)
		assert.Equal(t, 50*time.Millisecond, analysis.CurrentLatency)
	})

	t.Run("increasing latency trend", func(t *testing.T) {
		snapshots := []MetricsSnapshot{
			{AvgLatency: 10 * time.Millisecond},
			{AvgLatency: 20 * time.Millisecond},
			{AvgLatency: 30 * time.Millisecond},
			{AvgLatency: 40 * time.Millisecond},
			{AvgLatency: 50 * time.Millisecond},
		}

		analysis := analyzer.AnalyzeLatency(snapshots)
		assert.Equal(t, TrendUp, analysis.Trend)
		assert.Equal(t, 50*time.Millisecond, analysis.CurrentLatency)
		assert.Greater(t, analysis.PredictedLatency, analysis.CurrentLatency)
	})

	t.Run("decreasing latency trend", func(t *testing.T) {
		snapshots := []MetricsSnapshot{
			{AvgLatency: 100 * time.Millisecond},
			{AvgLatency: 80 * time.Millisecond},
			{AvgLatency: 60 * time.Millisecond},
			{AvgLatency: 40 * time.Millisecond},
			{AvgLatency: 20 * time.Millisecond},
		}

		analysis := analyzer.AnalyzeLatency(snapshots)
		assert.Equal(t, TrendDown, analysis.Trend)
		assert.Equal(t, 20*time.Millisecond, analysis.CurrentLatency)
		assert.Less(t, analysis.PredictedLatency, analysis.CurrentLatency)
	})

	t.Run("flat latency trend", func(t *testing.T) {
		snapshots := []MetricsSnapshot{
			{AvgLatency: 50 * time.Millisecond},
			{AvgLatency: 51 * time.Millisecond},
			{AvgLatency: 49 * time.Millisecond},
			{AvgLatency: 50 * time.Millisecond},
			{AvgLatency: 52 * time.Millisecond},
		}

		analysis := analyzer.AnalyzeLatency(snapshots)
		assert.Equal(t, TrendFlat, analysis.Trend)
		assert.Equal(t, 52*time.Millisecond, analysis.CurrentLatency)
	})
}

func TestAIAnalyzer_PredictThroughput(t *testing.T) {
	analyzer := NewAIAnalyzer()

	t.Run("empty snapshots", func(t *testing.T) {
		prediction := analyzer.PredictThroughput([]MetricsSnapshot{})
		assert.Equal(t, ThroughputPrediction{}, prediction)
	})

	t.Run("increasing throughput trend", func(t *testing.T) {
		snapshots := []MetricsSnapshot{
			{OrdersPerSec: 100, TradesPerSec: 50},  // Total: 150
			{OrdersPerSec: 120, TradesPerSec: 60},  // Total: 180
			{OrdersPerSec: 140, TradesPerSec: 70},  // Total: 210
			{OrdersPerSec: 160, TradesPerSec: 80},  // Total: 240
			{OrdersPerSec: 180, TradesPerSec: 90},  // Total: 270
		}

		prediction := analyzer.PredictThroughput(snapshots)
		assert.Equal(t, TrendUp, prediction.Trend)
		assert.Equal(t, 270.0, prediction.CurrentThroughput)
		assert.Equal(t, 270.0, prediction.MaxObservedThroughput)
		assert.Greater(t, prediction.PredictedThroughput, prediction.CurrentThroughput)
		assert.Greater(t, prediction.ConfidenceLevel, 0.0)
		assert.LessOrEqual(t, prediction.ConfidenceLevel, 1.0)
	})

	t.Run("decreasing throughput trend", func(t *testing.T) {
		snapshots := []MetricsSnapshot{
			{OrdersPerSec: 500, TradesPerSec: 200}, // Total: 700
			{OrdersPerSec: 450, TradesPerSec: 180}, // Total: 630
			{OrdersPerSec: 400, TradesPerSec: 160}, // Total: 560
			{OrdersPerSec: 350, TradesPerSec: 140}, // Total: 490
			{OrdersPerSec: 300, TradesPerSec: 120}, // Total: 420
		}

		prediction := analyzer.PredictThroughput(snapshots)
		assert.Equal(t, TrendDown, prediction.Trend)
		assert.Equal(t, 420.0, prediction.CurrentThroughput)
		assert.Equal(t, 700.0, prediction.MaxObservedThroughput)
		assert.Less(t, prediction.PredictedThroughput, prediction.CurrentThroughput)
	})

	t.Run("flat throughput trend", func(t *testing.T) {
		snapshots := []MetricsSnapshot{
			{OrdersPerSec: 200, TradesPerSec: 100},
			{OrdersPerSec: 205, TradesPerSec: 98},
			{OrdersPerSec: 198, TradesPerSec: 102},
			{OrdersPerSec: 202, TradesPerSec: 99},
			{OrdersPerSec: 199, TradesPerSec: 101},
		}

		prediction := analyzer.PredictThroughput(snapshots)
		assert.Equal(t, TrendFlat, prediction.Trend)
		assert.Equal(t, 300.0, prediction.CurrentThroughput) // 199 + 101
	})
}

func TestAIAnalyzer_DetectBottlenecks(t *testing.T) {
	analyzer := NewAIAnalyzer()

	t.Run("high latency bottleneck", func(t *testing.T) {
		snapshot := MetricsSnapshot{
			AvgLatency:   200 * time.Millisecond, // Above 100ms threshold
			OrdersPerSec: 1500,                   // Good throughput
			TradesPerSec: 800,
			OrderCount:   1000,
			TradeCount:   500,
		}

		bottlenecks := analyzer.DetectBottlenecks(snapshot)
		require.Len(t, bottlenecks, 1)

		bottleneck := bottlenecks[0]
		assert.Equal(t, "HIGH_LATENCY", bottleneck.Type)
		assert.Greater(t, bottleneck.Severity, 0.0)
		assert.LessOrEqual(t, bottleneck.Severity, 1.0)
		assert.Equal(t, "PROCESSING_ENGINE", bottleneck.Component)
		assert.Contains(t, bottleneck.Description, "latency")
	})

	t.Run("low throughput bottleneck", func(t *testing.T) {
		snapshot := MetricsSnapshot{
			AvgLatency:   50 * time.Millisecond, // Good latency
			OrdersPerSec: 300,                   // Low throughput (below 70% of 1000)
			TradesPerSec: 200,                   // Total: 500, threshold: 700
			OrderCount:   100,
			TradeCount:   50,
		}

		bottlenecks := analyzer.DetectBottlenecks(snapshot)
		require.Len(t, bottlenecks, 1)

		bottleneck := bottlenecks[0]
		assert.Equal(t, "LOW_THROUGHPUT", bottleneck.Type)
		assert.Greater(t, bottleneck.Severity, 0.0)
		assert.Equal(t, "ORDER_PROCESSOR", bottleneck.Component)
	})

	t.Run("memory pressure bottleneck", func(t *testing.T) {
		snapshot := MetricsSnapshot{
			AvgLatency:   50 * time.Millisecond,
			OrdersPerSec: 1500,
			TradesPerSec: 800,
			OrderCount:   8000,  // High event count
			TradeCount:   5000,  // Total: 13000 > 10000 threshold
		}

		bottlenecks := analyzer.DetectBottlenecks(snapshot)

		// Find memory pressure bottleneck
		var memoryBottleneck *Bottleneck
		for _, b := range bottlenecks {
			if b.Type == "MEMORY_PRESSURE" {
				memoryBottleneck = &b
				break
			}
		}

		require.NotNil(t, memoryBottleneck)
		assert.Equal(t, "MEMORY_PRESSURE", memoryBottleneck.Type)
		assert.Equal(t, "METRICS_COLLECTOR", memoryBottleneck.Component)
	})

	t.Run("latency variance bottleneck", func(t *testing.T) {
		snapshot := MetricsSnapshot{
			AvgLatency:   50 * time.Millisecond,
			MaxLatency:   300 * time.Millisecond, // High variance
			MinLatency:   10 * time.Millisecond,  // Variance: 290ms > 100ms threshold
			OrdersPerSec: 1500,
			TradesPerSec: 800,
			OrderCount:   1000,
			TradeCount:   500,
		}

		bottlenecks := analyzer.DetectBottlenecks(snapshot)

		// Find latency variance bottleneck
		var varianceBottleneck *Bottleneck
		for _, b := range bottlenecks {
			if b.Type == "LATENCY_VARIANCE" {
				varianceBottleneck = &b
				break
			}
		}

		require.NotNil(t, varianceBottleneck)
		assert.Equal(t, "LATENCY_VARIANCE", varianceBottleneck.Type)
		assert.Equal(t, "SYSTEM_LOAD", varianceBottleneck.Component)
	})

	t.Run("no bottlenecks", func(t *testing.T) {
		snapshot := MetricsSnapshot{
			AvgLatency:   50 * time.Millisecond,  // Good latency
			MaxLatency:   60 * time.Millisecond,  // Low variance
			MinLatency:   40 * time.Millisecond,
			OrdersPerSec: 800,                    // Good throughput
			TradesPerSec: 600,                    // Total: 1400 > 1000
			OrderCount:   1000,                   // Normal event count
			TradeCount:   500,
		}

		bottlenecks := analyzer.DetectBottlenecks(snapshot)
		assert.Len(t, bottlenecks, 0)
	})

	t.Run("multiple bottlenecks", func(t *testing.T) {
		snapshot := MetricsSnapshot{
			AvgLatency:   200 * time.Millisecond, // High latency
			MaxLatency:   400 * time.Millisecond, // High variance
			MinLatency:   50 * time.Millisecond,
			OrdersPerSec: 200,                    // Low throughput
			TradesPerSec: 100,                    // Total: 300 < 700
			OrderCount:   8000,                   // Memory pressure
			TradeCount:   5000,
		}

		bottlenecks := analyzer.DetectBottlenecks(snapshot)
		assert.Greater(t, len(bottlenecks), 1)

		// Verify we have different types of bottlenecks
		types := make(map[string]bool)
		for _, b := range bottlenecks {
			types[b.Type] = true
		}
		assert.Greater(t, len(types), 1)
	})
}

func TestAIAnalyzer_GenerateRecommendations(t *testing.T) {
	analyzer := NewAIAnalyzer()

	t.Run("latency trend recommendations", func(t *testing.T) {
		analysis := PerformanceAnalysis{
			LatencyTrend:    TrendUp,
			ThroughputTrend: TrendFlat,
			Bottlenecks:     []Bottleneck{},
		}

		recommendations := analyzer.GenerateRecommendations(analysis)
		assert.Contains(t, recommendations, "Consider optimizing order processing algorithms")
		assert.Contains(t, recommendations, "Review system resource allocation and scaling")
	})

	t.Run("throughput trend recommendations", func(t *testing.T) {
		analysis := PerformanceAnalysis{
			LatencyTrend:    TrendFlat,
			ThroughputTrend: TrendDown,
			Bottlenecks:     []Bottleneck{},
		}

		recommendations := analyzer.GenerateRecommendations(analysis)
		assert.Contains(t, recommendations, "Implement parallel processing for order matching")
		assert.Contains(t, recommendations, "Consider using more efficient data structures")
	})

	t.Run("high latency bottleneck recommendations", func(t *testing.T) {
		analysis := PerformanceAnalysis{
			LatencyTrend:    TrendFlat,
			ThroughputTrend: TrendFlat,
			Bottlenecks: []Bottleneck{
				{
					Type:     "HIGH_LATENCY",
					Severity: 0.9, // High severity
				},
			},
		}

		recommendations := analyzer.GenerateRecommendations(analysis)
		assert.Contains(t, recommendations, "URGENT: Implement latency optimization measures")
		assert.Contains(t, recommendations, "Review database query performance")
		assert.Contains(t, recommendations, "Consider implementing caching mechanisms")
	})

	t.Run("low throughput bottleneck recommendations", func(t *testing.T) {
		analysis := PerformanceAnalysis{
			Bottlenecks: []Bottleneck{
				{
					Type:     "LOW_THROUGHPUT",
					Severity: 0.5,
				},
			},
		}

		recommendations := analyzer.GenerateRecommendations(analysis)
		assert.Contains(t, recommendations, "Scale horizontally by adding more processing nodes")
		assert.Contains(t, recommendations, "Optimize critical path execution")
	})

	t.Run("memory pressure bottleneck recommendations", func(t *testing.T) {
		analysis := PerformanceAnalysis{
			Bottlenecks: []Bottleneck{
				{
					Type:     "MEMORY_PRESSURE",
					Severity: 0.6,
				},
			},
		}

		recommendations := analyzer.GenerateRecommendations(analysis)
		assert.Contains(t, recommendations, "Implement more aggressive event cleanup")
		assert.Contains(t, recommendations, "Consider streaming metrics to external storage")
	})

	t.Run("latency variance bottleneck recommendations", func(t *testing.T) {
		analysis := PerformanceAnalysis{
			Bottlenecks: []Bottleneck{
				{
					Type:     "LATENCY_VARIANCE",
					Severity: 0.4,
				},
			},
		}

		recommendations := analyzer.GenerateRecommendations(analysis)
		assert.Contains(t, recommendations, "Investigate and reduce system jitter")
		assert.Contains(t, recommendations, "Implement consistent resource allocation")
	})

	t.Run("optimal performance recommendations", func(t *testing.T) {
		analysis := PerformanceAnalysis{
			LatencyTrend:    TrendFlat,
			ThroughputTrend: TrendFlat,
			Bottlenecks:     []Bottleneck{},
		}

		recommendations := analyzer.GenerateRecommendations(analysis)
		assert.Contains(t, recommendations, "System performance is optimal")
		assert.Contains(t, recommendations, "Continue monitoring for early detection of issues")
	})
}

func TestAIAnalyzer_TrendCalculation(t *testing.T) {
	analyzer := NewAIAnalyzer()

	t.Run("calculate slope", func(t *testing.T) {
		// Test with known linear data
		values := []float64{1, 2, 3, 4, 5} // Slope should be 1
		slope := analyzer.calculateSlope(values)
		assert.InDelta(t, 1.0, slope, 0.001)

		// Test with decreasing data
		decreasingValues := []float64{5, 4, 3, 2, 1} // Slope should be -1
		slope = analyzer.calculateSlope(decreasingValues)
		assert.InDelta(t, -1.0, slope, 0.001)

		// Test with flat data
		flatValues := []float64{3, 3, 3, 3, 3} // Slope should be 0
		slope = analyzer.calculateSlope(flatValues)
		assert.InDelta(t, 0.0, slope, 0.001)
	})

	t.Run("exponential smoothing", func(t *testing.T) {
		values := []float64{10, 12, 11, 13, 15}
		alpha := 0.3

		prediction := analyzer.exponentialSmoothing(values, alpha)

		// Prediction should be between the last value and the smoothed average
		assert.Greater(t, prediction, 10.0)
		assert.Less(t, prediction, 20.0)

		// With alpha=0.3, prediction should be closer to recent values
		assert.Greater(t, prediction, 12.0)
	})

	t.Run("trend confidence calculation", func(t *testing.T) {
		// Consistent increasing trend should have high confidence
		consistentValues := []float64{1, 2, 3, 4, 5, 6, 7}
		confidence := analyzer.calculateTrendConfidence(consistentValues)
		assert.Greater(t, confidence, 0.5)

		// Random values should have lower confidence
		randomValues := []float64{1, 5, 2, 8, 3, 9, 1}
		confidence = analyzer.calculateTrendConfidence(randomValues)
		assert.LessOrEqual(t, confidence, 0.8)

		// Insufficient data should return default confidence
		shortValues := []float64{1, 2}
		confidence = analyzer.calculateTrendConfidence(shortValues)
		assert.Equal(t, 0.5, confidence)
	})
}

func TestAIAnalyzer_LatencyPercentiles(t *testing.T) {
	analyzer := NewAIAnalyzer()

	t.Run("calculate percentiles", func(t *testing.T) {
		snapshots := []MetricsSnapshot{
			{AvgLatency: 10 * time.Millisecond},
			{AvgLatency: 20 * time.Millisecond},
			{AvgLatency: 30 * time.Millisecond},
			{AvgLatency: 40 * time.Millisecond},
			{AvgLatency: 50 * time.Millisecond},
			{AvgLatency: 60 * time.Millisecond},
			{AvgLatency: 70 * time.Millisecond},
			{AvgLatency: 80 * time.Millisecond},
			{AvgLatency: 90 * time.Millisecond},
			{AvgLatency: 100 * time.Millisecond},
		}

		percentiles := analyzer.calculateLatencyPercentiles(snapshots)

		// P50 should be around the middle
		assert.GreaterOrEqual(t, percentiles[0], 40*time.Millisecond)
		assert.LessOrEqual(t, percentiles[0], 60*time.Millisecond)

		// P95 should be near the top
		assert.GreaterOrEqual(t, percentiles[1], 80*time.Millisecond)

		// P99 should be at or near the maximum
		assert.GreaterOrEqual(t, percentiles[2], 90*time.Millisecond)

		// Percentiles should be in increasing order
		assert.LessOrEqual(t, percentiles[0], percentiles[1])
		assert.LessOrEqual(t, percentiles[1], percentiles[2])
	})

	t.Run("empty snapshots", func(t *testing.T) {
		percentiles := analyzer.calculateLatencyPercentiles([]MetricsSnapshot{})
		assert.Equal(t, time.Duration(0), percentiles[0])
		assert.Equal(t, time.Duration(0), percentiles[1])
		assert.Equal(t, time.Duration(0), percentiles[2])
	})
}

func TestAIAnalyzer_EdgeCases(t *testing.T) {
	analyzer := NewAIAnalyzer()

	t.Run("single value calculations", func(t *testing.T) {
		singleValue := []float64{42.0}

		slope := analyzer.calculateSlope(singleValue)
		assert.Equal(t, 0.0, slope)

		trend := analyzer.calculateTrend(singleValue)
		assert.Equal(t, TrendFlat, trend)

		prediction := analyzer.exponentialSmoothing(singleValue, 0.3)
		assert.Equal(t, 42.0, prediction)
	})

	t.Run("zero and negative values", func(t *testing.T) {
		values := []float64{0, -1, 2, -3, 4}

		slope := analyzer.calculateSlope(values)
		assert.NotEqual(t, 0.0, slope) // Should still calculate a slope

		trend := analyzer.calculateTrend(values)
		assert.Contains(t, []TrendDirection{TrendUp, TrendDown, TrendFlat}, trend)
	})

	t.Run("identical values", func(t *testing.T) {
		identicalValues := []float64{5, 5, 5, 5, 5}

		slope := analyzer.calculateSlope(identicalValues)
		assert.Equal(t, 0.0, slope)

		trend := analyzer.calculateTrend(identicalValues)
		assert.Equal(t, TrendFlat, trend)

		threshold := analyzer.calculateSlopeThreshold(identicalValues)
		assert.Equal(t, 0.0, threshold)
	})
}