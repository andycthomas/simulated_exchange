package metrics

import (
	"math"
	"sort"
	"time"
)

// AIAnalyzer implements PerformanceAnalyzer with basic ML-style analysis
type AIAnalyzer struct {
	// Configuration parameters
	latencyThreshold    time.Duration
	throughputThreshold float64
	trendWindowSize     int
	bottleneckThreshold float64
}

// NewAIAnalyzer creates a new AIAnalyzer with default configuration
func NewAIAnalyzer() *AIAnalyzer {
	return &AIAnalyzer{
		latencyThreshold:    100 * time.Millisecond,
		throughputThreshold: 1000.0, // orders/trades per second
		trendWindowSize:     5,       // number of snapshots to analyze for trends
		bottleneckThreshold: 0.7,     // 70% threshold for bottleneck detection
	}
}

// NewAIAnalyzerWithConfig creates a new AIAnalyzer with custom configuration
func NewAIAnalyzerWithConfig(latencyThreshold time.Duration, throughputThreshold float64, trendWindowSize int, bottleneckThreshold float64) *AIAnalyzer {
	return &AIAnalyzer{
		latencyThreshold:    latencyThreshold,
		throughputThreshold: throughputThreshold,
		trendWindowSize:     trendWindowSize,
		bottleneckThreshold: bottleneckThreshold,
	}
}

// AnalyzeLatency performs latency trend analysis using linear regression
func (ai *AIAnalyzer) AnalyzeLatency(snapshots []MetricsSnapshot) LatencyAnalysis {
	if len(snapshots) == 0 {
		return LatencyAnalysis{}
	}

	// Extract latency data points
	latencies := make([]float64, len(snapshots))
	for i, snapshot := range snapshots {
		latencies[i] = float64(snapshot.AvgLatency.Nanoseconds())
	}

	// Calculate trend using linear regression
	trend := ai.calculateTrend(latencies)

	// Get current latency
	currentLatency := snapshots[len(snapshots)-1].AvgLatency

	// Predict future latency based on trend
	predictedLatency := ai.predictNextValue(latencies, trend)

	// Calculate percentiles
	percentiles := ai.calculateLatencyPercentiles(snapshots)

	return LatencyAnalysis{
		Trend:            trend,
		CurrentLatency:   currentLatency,
		PredictedLatency: time.Duration(predictedLatency),
		PercentileP50:    percentiles[0],
		PercentileP95:    percentiles[1],
		PercentileP99:    percentiles[2],
	}
}

// PredictThroughput predicts future throughput using exponential smoothing
func (ai *AIAnalyzer) PredictThroughput(snapshots []MetricsSnapshot) ThroughputPrediction {
	if len(snapshots) == 0 {
		return ThroughputPrediction{}
	}

	// Extract throughput data (orders + trades per second)
	throughputs := make([]float64, len(snapshots))
	maxThroughput := 0.0

	for i, snapshot := range snapshots {
		throughput := snapshot.OrdersPerSec + snapshot.TradesPerSec
		throughputs[i] = throughput
		if throughput > maxThroughput {
			maxThroughput = throughput
		}
	}

	// Calculate trend
	trend := ai.calculateTrend(throughputs)

	// Current throughput
	currentThroughput := throughputs[len(throughputs)-1]

	// Predict using exponential smoothing
	predictedThroughput := ai.exponentialSmoothing(throughputs, 0.3) // Alpha = 0.3

	// Calculate confidence based on trend consistency
	confidence := ai.calculateTrendConfidence(throughputs)

	return ThroughputPrediction{
		Trend:                 trend,
		CurrentThroughput:     currentThroughput,
		PredictedThroughput:   predictedThroughput,
		MaxObservedThroughput: maxThroughput,
		ConfidenceLevel:       confidence,
	}
}

// DetectBottlenecks analyzes metrics to detect performance bottlenecks
func (ai *AIAnalyzer) DetectBottlenecks(snapshot MetricsSnapshot) []Bottleneck {
	var bottlenecks []Bottleneck

	// High latency bottleneck
	if snapshot.AvgLatency > ai.latencyThreshold {
		severity := float64(snapshot.AvgLatency) / float64(ai.latencyThreshold*2) // Normalize
		if severity > 1.0 {
			severity = 1.0
		}

		bottlenecks = append(bottlenecks, Bottleneck{
			Type:        "HIGH_LATENCY",
			Severity:    severity,
			Description: "Average processing latency is above threshold",
			Component:   "PROCESSING_ENGINE",
		})
	}

	// Low throughput bottleneck
	totalThroughput := snapshot.OrdersPerSec + snapshot.TradesPerSec
	if totalThroughput < ai.throughputThreshold*ai.bottleneckThreshold {
		severity := 1.0 - (totalThroughput / ai.throughputThreshold)
		if severity > 1.0 {
			severity = 1.0
		}

		bottlenecks = append(bottlenecks, Bottleneck{
			Type:        "LOW_THROUGHPUT",
			Severity:    severity,
			Description: "System throughput is below expected levels",
			Component:   "ORDER_PROCESSOR",
		})
	}

	// Memory pressure bottleneck (based on event accumulation)
	if snapshot.OrderCount+snapshot.TradeCount > 10000 { // Threshold for memory pressure
		severity := float64(snapshot.OrderCount+snapshot.TradeCount) / 20000.0
		if severity > 1.0 {
			severity = 1.0
		}

		bottlenecks = append(bottlenecks, Bottleneck{
			Type:        "MEMORY_PRESSURE",
			Severity:    severity,
			Description: "High number of events may indicate memory pressure",
			Component:   "METRICS_COLLECTOR",
		})
	}

	// Latency variance bottleneck
	latencyVariance := float64(snapshot.MaxLatency - snapshot.MinLatency)
	if latencyVariance > float64(ai.latencyThreshold) {
		severity := latencyVariance / float64(ai.latencyThreshold*2)
		if severity > 1.0 {
			severity = 1.0
		}

		bottlenecks = append(bottlenecks, Bottleneck{
			Type:        "LATENCY_VARIANCE",
			Severity:    severity,
			Description: "High latency variance indicates inconsistent performance",
			Component:   "SYSTEM_LOAD",
		})
	}

	return bottlenecks
}

// GenerateRecommendations generates performance optimization recommendations
func (ai *AIAnalyzer) GenerateRecommendations(analysis PerformanceAnalysis) []string {
	var recommendations []string

	// Latency-based recommendations
	if analysis.LatencyTrend == TrendUp {
		recommendations = append(recommendations, "Consider optimizing order processing algorithms")
		recommendations = append(recommendations, "Review system resource allocation and scaling")
	}

	// Throughput-based recommendations
	if analysis.ThroughputTrend == TrendDown {
		recommendations = append(recommendations, "Implement parallel processing for order matching")
		recommendations = append(recommendations, "Consider using more efficient data structures")
	}

	// Bottleneck-specific recommendations
	for _, bottleneck := range analysis.Bottlenecks {
		switch bottleneck.Type {
		case "HIGH_LATENCY":
			if bottleneck.Severity > 0.8 {
				recommendations = append(recommendations, "URGENT: Implement latency optimization measures")
			}
			recommendations = append(recommendations, "Review database query performance")
			recommendations = append(recommendations, "Consider implementing caching mechanisms")

		case "LOW_THROUGHPUT":
			recommendations = append(recommendations, "Scale horizontally by adding more processing nodes")
			recommendations = append(recommendations, "Optimize critical path execution")

		case "MEMORY_PRESSURE":
			recommendations = append(recommendations, "Implement more aggressive event cleanup")
			recommendations = append(recommendations, "Consider streaming metrics to external storage")

		case "LATENCY_VARIANCE":
			recommendations = append(recommendations, "Investigate and reduce system jitter")
			recommendations = append(recommendations, "Implement consistent resource allocation")
		}
	}

	// General recommendations
	if len(analysis.Bottlenecks) == 0 {
		recommendations = append(recommendations, "System performance is optimal")
		recommendations = append(recommendations, "Continue monitoring for early detection of issues")
	}

	return recommendations
}

// calculateTrend determines the trend direction using linear regression
func (ai *AIAnalyzer) calculateTrend(values []float64) TrendDirection {
	if len(values) < 2 {
		return TrendFlat
	}

	// Use last N values for trend calculation
	windowSize := ai.trendWindowSize
	if len(values) < windowSize {
		windowSize = len(values)
	}

	recentValues := values[len(values)-windowSize:]
	slope := ai.calculateSlope(recentValues)

	// Determine trend based on slope magnitude
	threshold := ai.calculateSlopeThreshold(recentValues)

	if slope > threshold {
		return TrendUp
	} else if slope < -threshold {
		return TrendDown
	}
	return TrendFlat
}

// calculateSlope calculates the slope of a linear regression line
func (ai *AIAnalyzer) calculateSlope(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2 float64

	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope: (n*ΣXY - ΣX*ΣY) / (n*ΣX² - (ΣX)²)
	numerator := n*sumXY - sumX*sumY
	denominator := n*sumX2 - sumX*sumX

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// calculateSlopeThreshold calculates threshold for trend detection based on data variance
func (ai *AIAnalyzer) calculateSlopeThreshold(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	// Calculate standard deviation
	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))

	stdDev := math.Sqrt(variance)

	// Threshold is 10% of one standard deviation per data point
	return stdDev * 0.1
}

// predictNextValue predicts the next value based on trend
func (ai *AIAnalyzer) predictNextValue(values []float64, trend TrendDirection) float64 {
	if len(values) == 0 {
		return 0
	}

	current := values[len(values)-1]

	if trend == TrendFlat {
		return current
	}

	// Simple linear extrapolation
	slope := ai.calculateSlope(values)
	return current + slope
}

// exponentialSmoothing applies exponential smoothing for prediction
func (ai *AIAnalyzer) exponentialSmoothing(values []float64, alpha float64) float64 {
	if len(values) == 0 {
		return 0
	}

	if len(values) == 1 {
		return values[0]
	}

	// Start with first value
	smoothed := values[0]

	// Apply exponential smoothing
	for i := 1; i < len(values); i++ {
		smoothed = alpha*values[i] + (1-alpha)*smoothed
	}

	// Predict next value
	return alpha*values[len(values)-1] + (1-alpha)*smoothed
}

// calculateTrendConfidence calculates confidence level based on trend consistency
func (ai *AIAnalyzer) calculateTrendConfidence(values []float64) float64 {
	if len(values) < 3 {
		return 0.5 // Low confidence with insufficient data
	}

	// Calculate how consistent the trend is
	slopes := make([]float64, 0)
	windowSize := 3

	for i := 0; i <= len(values)-windowSize; i++ {
		window := values[i : i+windowSize]
		slope := ai.calculateSlope(window)
		slopes = append(slopes, slope)
	}

	// Calculate coefficient of variation for slopes
	if len(slopes) == 0 {
		return 0.5
	}

	mean := 0.0
	for _, slope := range slopes {
		mean += slope
	}
	mean /= float64(len(slopes))

	variance := 0.0
	for _, slope := range slopes {
		variance += (slope - mean) * (slope - mean)
	}
	variance /= float64(len(slopes))

	stdDev := math.Sqrt(variance)

	// Lower coefficient of variation = higher confidence
	if mean == 0 {
		return 0.5
	}

	cv := math.Abs(stdDev / mean)
	confidence := 1.0 / (1.0 + cv) // Inverse relationship

	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// calculateLatencyPercentiles calculates P50, P95, P99 latency percentiles
func (ai *AIAnalyzer) calculateLatencyPercentiles(snapshots []MetricsSnapshot) [3]time.Duration {
	if len(snapshots) == 0 {
		return [3]time.Duration{0, 0, 0}
	}

	// Collect all latency values
	var allLatencies []time.Duration
	for _, snapshot := range snapshots {
		// For simplicity, we'll use the average latency from each snapshot
		// In a real implementation, you'd want to collect all individual latency measurements
		allLatencies = append(allLatencies, snapshot.AvgLatency)
	}

	// Sort latencies
	sort.Slice(allLatencies, func(i, j int) bool {
		return allLatencies[i] < allLatencies[j]
	})

	n := len(allLatencies)
	p50Index := int(float64(n) * 0.50)
	p95Index := int(float64(n) * 0.95)
	p99Index := int(float64(n) * 0.99)

	// Ensure indices are within bounds
	if p50Index >= n {
		p50Index = n - 1
	}
	if p95Index >= n {
		p95Index = n - 1
	}
	if p99Index >= n {
		p99Index = n - 1
	}

	return [3]time.Duration{
		allLatencies[p50Index],
		allLatencies[p95Index],
		allLatencies[p99Index],
	}
}