package ai

import (
	"fmt"
	"math"
	"sort"
	"time"

	"simulated_exchange/internal/metrics"
)

// IntelligentAnalyzer implements PerformanceAI with machine learning algorithms
type IntelligentAnalyzer struct {
	config            MLAnalysisConfig
	historicalData    []metrics.MetricsSnapshot
	baselineMetrics   *metrics.MetricsSnapshot
	learningEnabled   bool
	adaptiveThresholds map[string]float64
}

// NewIntelligentAnalyzer creates a new AI-powered performance analyzer
func NewIntelligentAnalyzer(config MLAnalysisConfig) *IntelligentAnalyzer {
	return &IntelligentAnalyzer{
		config:             config,
		historicalData:     make([]metrics.MetricsSnapshot, 0),
		learningEnabled:    true,
		adaptiveThresholds: make(map[string]float64),
	}
}

// NewDefaultIntelligentAnalyzer creates analyzer with default configuration
func NewDefaultIntelligentAnalyzer() *IntelligentAnalyzer {
	return NewIntelligentAnalyzer(DefaultMLAnalysisConfig())
}

// SetBaseline establishes baseline metrics for comparison
func (ia *IntelligentAnalyzer) SetBaseline(baseline metrics.MetricsSnapshot) {
	ia.baselineMetrics = &baseline
	ia.updateAdaptiveThresholds(baseline)
}

// UpdateHistoricalData adds new metrics data for learning
func (ia *IntelligentAnalyzer) UpdateHistoricalData(snapshot metrics.MetricsSnapshot) {
	ia.historicalData = append(ia.historicalData, snapshot)

	// Maintain reasonable history size
	maxHistorySize := 1000
	if len(ia.historicalData) > maxHistorySize {
		ia.historicalData = ia.historicalData[len(ia.historicalData)-maxHistorySize:]
	}

	// Update adaptive thresholds with new data
	if ia.learningEnabled {
		ia.updateAdaptiveThresholds(snapshot)
	}
}

// AnalyzeBottlenecks identifies performance bottlenecks using ML algorithms
func (ia *IntelligentAnalyzer) AnalyzeBottlenecks(snapshots []metrics.MetricsSnapshot) []Bottleneck {
	if len(snapshots) < ia.config.MinDataPoints {
		return nil
	}

	var bottlenecks []Bottleneck

	// Analyze latency bottlenecks
	latencyBottlenecks := ia.detectLatencyBottlenecks(snapshots)
	bottlenecks = append(bottlenecks, latencyBottlenecks...)

	// Analyze throughput bottlenecks
	throughputBottlenecks := ia.detectThroughputBottlenecks(snapshots)
	bottlenecks = append(bottlenecks, throughputBottlenecks...)

	// Analyze capacity bottlenecks
	capacityBottlenecks := ia.detectCapacityBottlenecks(snapshots)
	bottlenecks = append(bottlenecks, capacityBottlenecks...)

	// Analyze variance bottlenecks (instability)
	varianceBottlenecks := ia.detectVarianceBottlenecks(snapshots)
	bottlenecks = append(bottlenecks, varianceBottlenecks...)

	// Sort by severity (highest first)
	sort.Slice(bottlenecks, func(i, j int) bool {
		return bottlenecks[i].Severity > bottlenecks[j].Severity
	})

	return bottlenecks
}

// PredictCapacity forecasts system capacity needs using ML prediction
func (ia *IntelligentAnalyzer) PredictCapacity(snapshots []metrics.MetricsSnapshot, timeHorizon time.Duration) CapacityPrediction {
	if len(snapshots) < ia.config.MinDataPoints {
		return CapacityPrediction{
			TimeHorizon: timeHorizon,
			RiskFactors: []string{"Insufficient historical data for prediction"},
		}
	}

	// Perform trend analysis for capacity prediction
	trends := ia.analyzeTrends(snapshots)

	// Predict future load using linear regression and exponential smoothing
	loadPrediction := ia.predictLoad(snapshots, timeHorizon)

	// Calculate recommended capacity with safety margins
	capacity := ia.calculateRecommendedCapacity(loadPrediction, trends)

	// Calculate confidence intervals
	confidence := ia.calculatePredictionConfidence(snapshots, loadPrediction)

	return CapacityPrediction{
		TimeHorizon:         timeHorizon,
		PredictedLoad:       loadPrediction,
		RecommendedCapacity: capacity,
		ConfidenceInterval:  confidence,
		Assumptions: []string{
			"Current growth trends continue",
			"No major architectural changes",
			"Similar usage patterns maintained",
		},
		RiskFactors: ia.identifyCapacityRiskFactors(trends),
		CreatedAt:   time.Now(),
	}
}

// GenerateRecommendations creates AI-powered optimization recommendations
func (ia *IntelligentAnalyzer) GenerateRecommendations(analysis PerformanceAnalysis) []Recommendation {
	var recommendations []Recommendation

	// Generate recommendations based on bottlenecks
	for _, bottleneck := range analysis.Bottlenecks {
		recs := ia.generateBottleneckRecommendations(bottleneck)
		recommendations = append(recommendations, recs...)
	}

	// Generate proactive recommendations based on trends
	trendRecs := ia.generateTrendBasedRecommendations(analysis.TrendAnalysis)
	recommendations = append(recommendations, trendRecs...)

	// Generate capacity recommendations
	capacityRecs := ia.generateCapacityRecommendations(analysis.CapacityPrediction)
	recommendations = append(recommendations, capacityRecs...)

	// Prioritize and deduplicate recommendations
	recommendations = ia.prioritizeRecommendations(recommendations)

	return recommendations
}

// Private methods for ML algorithms and analysis

// detectLatencyBottlenecks uses statistical analysis to identify latency issues
func (ia *IntelligentAnalyzer) detectLatencyBottlenecks(snapshots []metrics.MetricsSnapshot) []Bottleneck {
	var bottlenecks []Bottleneck

	latencies := ia.extractLatencyValues(snapshots)
	if len(latencies) == 0 {
		return bottlenecks
	}

	// Calculate statistical measures
	mean := ia.calculateMean(latencies)
	stdDev := ia.calculateStandardDeviation(latencies, mean)
	p95 := ia.calculatePercentile(latencies, 0.95)

	// Adaptive threshold based on historical data
	threshold := ia.getAdaptiveThreshold("latency", mean*1.5)

	// Detect anomalies using statistical methods
	if p95 > threshold {
		severity := ia.calculateLatencySeverity(p95, threshold, stdDev)
		impact := ia.calculateLatencyImpact(p95, mean)

		bottleneck := Bottleneck{
			Type:        BottleneckTypeLatency,
			Component:   "Trading Engine",
			Severity:    severity,
			Impact:      impact,
			Description: fmt.Sprintf("High latency detected: P95=%.2fms (threshold=%.2fms)", p95, threshold),
			DetectedAt:  time.Now(),
			AffectedMetrics: []string{"avg_latency", "p95_latency"},
			Confidence:  ia.calculateConfidence(len(latencies), stdDev/mean),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	// Detect latency variance issues
	if stdDev/mean > 0.5 { // High coefficient of variation
		severity := math.Min(0.8, (stdDev/mean)*1.6)

		bottleneck := Bottleneck{
			Type:        BottleneckTypeLatency,
			Component:   "System Stability",
			Severity:    severity,
			Impact:      ia.calculateVarianceImpact(stdDev/mean),
			Description: fmt.Sprintf("High latency variance detected: CV=%.2f", stdDev/mean),
			DetectedAt:  time.Now(),
			AffectedMetrics: []string{"latency_variance"},
			Confidence:  0.85,
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// detectThroughputBottlenecks analyzes throughput patterns for bottlenecks
func (ia *IntelligentAnalyzer) detectThroughputBottlenecks(snapshots []metrics.MetricsSnapshot) []Bottleneck {
	var bottlenecks []Bottleneck

	ordersPerSec := ia.extractOrdersPerSecond(snapshots)
	tradesPerSec := ia.extractTradesPerSecond(snapshots)

	if len(ordersPerSec) == 0 {
		return bottlenecks
	}

	// Analyze order processing throughput
	orderMean := ia.calculateMean(ordersPerSec)
	expectedThroughput := ia.getAdaptiveThreshold("orders_throughput", 100.0)

	if orderMean < expectedThroughput*0.7 { // 30% below expected
		severity := (expectedThroughput - orderMean) / expectedThroughput
		impact := ia.calculateThroughputImpact(orderMean, expectedThroughput)

		bottleneck := Bottleneck{
			Type:        BottleneckTypeThroughput,
			Component:   "Order Processing",
			Severity:    severity,
			Impact:      impact,
			Description: fmt.Sprintf("Low order throughput: %.2f ops/sec (expected: %.2f)", orderMean, expectedThroughput),
			DetectedAt:  time.Now(),
			AffectedMetrics: []string{"orders_per_sec"},
			Confidence:  ia.calculateConfidence(len(ordersPerSec), 0.1),
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	// Analyze trade execution efficiency
	if len(tradesPerSec) > 0 {
		tradeMean := ia.calculateMean(tradesPerSec)
		expectedTradeRate := orderMean * 0.3 // Expect 30% of orders to result in trades

		if tradeMean < expectedTradeRate*0.5 { // Significantly below expected
			severity := (expectedTradeRate - tradeMean) / expectedTradeRate
			impact := ia.calculateTradeEfficiencyImpact(tradeMean, orderMean)

			bottleneck := Bottleneck{
				Type:        BottleneckTypeThroughput,
				Component:   "Trade Execution",
				Severity:    severity,
				Impact:      impact,
				Description: fmt.Sprintf("Low trade efficiency: %.2f trades/sec vs %.2f orders/sec", tradeMean, orderMean),
				DetectedAt:  time.Now(),
				AffectedMetrics: []string{"trades_per_sec", "trade_ratio"},
				Confidence:  0.75,
			}

			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	return bottlenecks
}

// detectCapacityBottlenecks identifies resource capacity constraints
func (ia *IntelligentAnalyzer) detectCapacityBottlenecks(snapshots []metrics.MetricsSnapshot) []Bottleneck {
	var bottlenecks []Bottleneck

	// Analyze order volume trends for capacity planning
	volumes := ia.extractTotalVolumes(snapshots)
	if len(volumes) < 5 {
		return bottlenecks
	}

	// Calculate growth rate using linear regression
	growthRate := ia.calculateGrowthRate(volumes)

	// Memory pressure analysis (simulated based on volume)
	currentVolume := volumes[len(volumes)-1]
	estimatedMemoryUsage := currentVolume * 0.001 // Simplified calculation

	if estimatedMemoryUsage > 1000 { // Arbitrary threshold for demo
		severity := math.Min(1.0, estimatedMemoryUsage/2000)
		impact := ia.calculateMemoryImpact(estimatedMemoryUsage)

		bottleneck := Bottleneck{
			Type:        BottleneckTypeMemory,
			Component:   "Memory Subsystem",
			Severity:    severity,
			Impact:      impact,
			Description: fmt.Sprintf("High memory pressure detected: estimated %.2f MB usage", estimatedMemoryUsage),
			DetectedAt:  time.Now(),
			AffectedMetrics: []string{"total_volume", "memory_usage"},
			Confidence:  0.7,
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	// Rapid growth detection
	if growthRate > 0.1 { // 10% growth rate threshold
		severity := math.Min(0.9, growthRate*5)
		impact := ia.calculateGrowthImpact(growthRate)

		bottleneck := Bottleneck{
			Type:        BottleneckTypeCapacity,
			Component:   "System Capacity",
			Severity:    severity,
			Impact:      impact,
			Description: fmt.Sprintf("Rapid growth detected: %.2f%% growth rate", growthRate*100),
			DetectedAt:  time.Now(),
			AffectedMetrics: []string{"total_volume", "growth_rate"},
			Confidence:  0.8,
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// detectVarianceBottlenecks identifies system instability issues
func (ia *IntelligentAnalyzer) detectVarianceBottlenecks(snapshots []metrics.MetricsSnapshot) []Bottleneck {
	var bottlenecks []Bottleneck

	ordersPerSec := ia.extractOrdersPerSecond(snapshots)
	if len(ordersPerSec) < ia.config.MinDataPoints {
		return bottlenecks
	}

	// Calculate coefficient of variation for stability analysis
	mean := ia.calculateMean(ordersPerSec)
	stdDev := ia.calculateStandardDeviation(ordersPerSec, mean)
	cv := stdDev / mean

	if cv > 0.7 { // High variability threshold
		severity := math.Min(1.0, cv*1.2)
		impact := ia.calculateVarianceImpact(cv)

		bottleneck := Bottleneck{
			Type:        BottleneckTypeIO,
			Component:   "System Stability",
			Severity:    severity,
			Impact:      impact,
			Description: fmt.Sprintf("High system variability: CV=%.2f", cv),
			DetectedAt:  time.Now(),
			AffectedMetrics: []string{"orders_per_sec_variance"},
			Confidence:  0.75,
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// analyzeTrends performs comprehensive trend analysis
func (ia *IntelligentAnalyzer) analyzeTrends(snapshots []metrics.MetricsSnapshot) TrendAnalysis {
	latencies := ia.extractLatencyValues(snapshots)
	throughputs := ia.extractOrdersPerSecond(snapshots)
	volumes := ia.extractTotalVolumes(snapshots)

	return TrendAnalysis{
		LatencyTrend:    ia.determineTrendDirection(latencies),
		ThroughputTrend: ia.determineTrendDirection(throughputs),
		VolumeTrend:     ia.determineTrendDirection(volumes),
		ErrorRateTrend:  TrendStable, // Simplified for demo
		TrendStrength:   ia.calculateTrendStrength(throughputs),
		Seasonality:     ia.detectSeasonality(snapshots),
	}
}

// Machine Learning utility methods

// extractLatencyValues extracts latency values from snapshots
func (ia *IntelligentAnalyzer) extractLatencyValues(snapshots []metrics.MetricsSnapshot) []float64 {
	values := make([]float64, 0, len(snapshots))
	for _, snapshot := range snapshots {
		if latencyMs := snapshot.AvgLatency.Milliseconds(); latencyMs > 0 {
			values = append(values, float64(latencyMs))
		}
	}
	return values
}

// extractOrdersPerSecond extracts orders per second from snapshots
func (ia *IntelligentAnalyzer) extractOrdersPerSecond(snapshots []metrics.MetricsSnapshot) []float64 {
	values := make([]float64, 0, len(snapshots))
	for _, snapshot := range snapshots {
		values = append(values, snapshot.OrdersPerSec)
	}
	return values
}

// extractTradesPerSecond extracts trades per second from snapshots
func (ia *IntelligentAnalyzer) extractTradesPerSecond(snapshots []metrics.MetricsSnapshot) []float64 {
	values := make([]float64, 0, len(snapshots))
	for _, snapshot := range snapshots {
		values = append(values, snapshot.TradesPerSec)
	}
	return values
}

// extractTotalVolumes extracts total volumes from snapshots
func (ia *IntelligentAnalyzer) extractTotalVolumes(snapshots []metrics.MetricsSnapshot) []float64 {
	values := make([]float64, 0, len(snapshots))
	for _, snapshot := range snapshots {
		values = append(values, snapshot.TotalVolume)
	}
	return values
}

// Statistical calculation methods

// calculateMean calculates the arithmetic mean
func (ia *IntelligentAnalyzer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStandardDeviation calculates standard deviation
func (ia *IntelligentAnalyzer) calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	sumSquaredDiffs := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquaredDiffs += diff * diff
	}

	variance := sumSquaredDiffs / float64(len(values)-1)
	return math.Sqrt(variance)
}

// calculatePercentile calculates the specified percentile
func (ia *IntelligentAnalyzer) calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := percentile * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// calculateGrowthRate calculates growth rate using linear regression
func (ia *IntelligentAnalyzer) calculateGrowthRate(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	n := float64(len(values))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Linear regression slope
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Convert to growth rate
	if len(values) > 0 && values[0] != 0 {
		return slope / values[0] // Normalize by initial value
	}

	return 0
}

// determineTrendDirection determines the direction of a trend
func (ia *IntelligentAnalyzer) determineTrendDirection(values []float64) TrendDirection {
	if len(values) < 3 {
		return TrendStable
	}

	growthRate := ia.calculateGrowthRate(values)
	stdDev := ia.calculateStandardDeviation(values, ia.calculateMean(values))
	mean := ia.calculateMean(values)

	// Coefficient of variation for volatility assessment
	cv := stdDev / mean

	if cv > 0.5 {
		return TrendVolatile
	}

	if growthRate > 0.05 {
		return TrendIncreasing
	} else if growthRate < -0.05 {
		return TrendDecreasing
	}

	return TrendStable
}

// calculateTrendStrength calculates the strength of a trend
func (ia *IntelligentAnalyzer) calculateTrendStrength(values []float64) float64 {
	if len(values) < 3 {
		return 0
	}

	growthRate := math.Abs(ia.calculateGrowthRate(values))
	return math.Min(1.0, growthRate*10) // Normalize to 0-1 range
}

// detectSeasonality detects seasonal patterns in the data
func (ia *IntelligentAnalyzer) detectSeasonality(snapshots []metrics.MetricsSnapshot) []SeasonalPattern {
	// Simplified seasonality detection
	// In a real implementation, this would use FFT or other signal processing techniques
	var patterns []SeasonalPattern

	if len(snapshots) > 24 { // Need sufficient data for hourly patterns
		patterns = append(patterns, SeasonalPattern{
			Pattern:     "Daily",
			Strength:    0.3, // Simplified calculation
			Period:      24 * time.Hour,
			Description: "Daily usage pattern detected",
		})
	}

	return patterns
}

// Adaptive threshold management

// updateAdaptiveThresholds updates thresholds based on new data
func (ia *IntelligentAnalyzer) updateAdaptiveThresholds(snapshot metrics.MetricsSnapshot) {
	// Exponential moving average for adaptive thresholds
	alpha := 0.1 // Learning rate

	if current, exists := ia.adaptiveThresholds["latency"]; exists {
		latency := float64(snapshot.AvgLatency.Milliseconds())
		ia.adaptiveThresholds["latency"] = alpha*latency + (1-alpha)*current
	} else {
		ia.adaptiveThresholds["latency"] = float64(snapshot.AvgLatency.Milliseconds())
	}

	if current, exists := ia.adaptiveThresholds["orders_throughput"]; exists {
		ia.adaptiveThresholds["orders_throughput"] = alpha*snapshot.OrdersPerSec + (1-alpha)*current
	} else {
		ia.adaptiveThresholds["orders_throughput"] = snapshot.OrdersPerSec
	}
}

// getAdaptiveThreshold gets adaptive threshold or fallback to default
func (ia *IntelligentAnalyzer) getAdaptiveThreshold(key string, defaultValue float64) float64 {
	if value, exists := ia.adaptiveThresholds[key]; exists {
		return value
	}
	return defaultValue
}

// Impact calculation methods (continued in next part due to length...)
func (ia *IntelligentAnalyzer) calculateLatencyImpact(latency, baseline float64) BusinessImpact {
	impactRatio := (latency - baseline) / baseline

	return BusinessImpact{
		Revenue:        -impactRatio * 10000,  // Revenue loss estimate
		Cost:           impactRatio * 5000,    // Additional cost
		UserExperience: math.Max(0, 1-impactRatio*2), // UX degradation
		Reliability:    math.Max(0, 1-impactRatio),   // Reliability impact
		Scalability:    math.Max(0, 1-impactRatio*1.5), // Scalability impact
		OverallScore:   math.Max(0, 1-impactRatio*1.2), // Overall impact
	}
}

// Additional impact calculation methods...
func (ia *IntelligentAnalyzer) calculateThroughputImpact(current, expected float64) BusinessImpact {
	ratio := current / expected
	impact := 1 - ratio

	return BusinessImpact{
		Revenue:        -impact * 15000,
		Cost:           impact * 8000,
		UserExperience: ratio,
		Reliability:    ratio,
		Scalability:    ratio * 0.8,
		OverallScore:   ratio,
	}
}

func (ia *IntelligentAnalyzer) calculateVarianceImpact(cv float64) BusinessImpact {
	impact := math.Min(1.0, cv)

	return BusinessImpact{
		Revenue:        -impact * 5000,
		Cost:           impact * 3000,
		UserExperience: 1 - impact*0.8,
		Reliability:    1 - impact,
		Scalability:    1 - impact*0.6,
		OverallScore:   1 - impact*0.7,
	}
}

func (ia *IntelligentAnalyzer) calculateMemoryImpact(usage float64) BusinessImpact {
	normalizedUsage := usage / 2000 // Normalize against threshold
	impact := math.Min(1.0, normalizedUsage)

	return BusinessImpact{
		Revenue:        -impact * 12000,
		Cost:           impact * 7000,
		UserExperience: 1 - impact*0.9,
		Reliability:    1 - impact*0.8,
		Scalability:    1 - impact,
		OverallScore:   1 - impact*0.85,
	}
}

func (ia *IntelligentAnalyzer) calculateTradeEfficiencyImpact(tradeRate, orderRate float64) BusinessImpact {
	efficiency := tradeRate / (orderRate * 0.3) // Expected 30% trade rate
	impact := 1 - efficiency

	return BusinessImpact{
		Revenue:        -impact * 20000,
		Cost:           impact * 10000,
		UserExperience: efficiency,
		Reliability:    efficiency,
		Scalability:    efficiency * 0.9,
		OverallScore:   efficiency,
	}
}

func (ia *IntelligentAnalyzer) calculateGrowthImpact(growthRate float64) BusinessImpact {
	// Growth impact is positive for planning but creates pressure
	return BusinessImpact{
		Revenue:        growthRate * 50000,  // Growth opportunity
		Cost:           growthRate * 25000,  // Infrastructure cost
		UserExperience: 1 - growthRate*0.3,  // Potential UX pressure
		Reliability:    1 - growthRate*0.4,  // Reliability pressure
		Scalability:    1 - growthRate*0.8,  // Scalability pressure
		OverallScore:   1 - growthRate*0.5,  // Overall pressure
	}
}

// Confidence calculation
func (ia *IntelligentAnalyzer) calculateConfidence(dataPoints int, variabilityFactor float64) float64 {
	// Base confidence on data points available
	dataConfidence := math.Min(1.0, float64(dataPoints)/float64(ia.config.MinDataPoints*2))

	// Adjust for variability
	variabilityPenalty := math.Min(0.5, variabilityFactor)

	return math.Max(0.1, dataConfidence*(1-variabilityPenalty))
}

// Prediction methods (simplified implementations)

func (ia *IntelligentAnalyzer) predictLoad(snapshots []metrics.MetricsSnapshot, horizon time.Duration) LoadPrediction {
	ordersPerSec := ia.extractOrdersPerSecond(snapshots)
	tradesPerSec := ia.extractTradesPerSecond(snapshots)

	orderGrowth := ia.calculateGrowthRate(ordersPerSec)
	tradeGrowth := ia.calculateGrowthRate(tradesPerSec)

	currentOrders := ordersPerSec[len(ordersPerSec)-1]
	currentTrades := tradesPerSec[len(tradesPerSec)-1]

	// Project forward based on growth rate
	projectionFactor := 1 + orderGrowth*(horizon.Hours()/24) // Daily growth projection

	return LoadPrediction{
		OrdersPerSecond: currentOrders * projectionFactor,
		TradesPerSecond: currentTrades * math.Max(1, 1+tradeGrowth*(horizon.Hours()/24)),
		PeakMultiplier:  1.5, // Assume 50% peak load
		GrowthRate:      orderGrowth,
	}
}

func (ia *IntelligentAnalyzer) calculateRecommendedCapacity(load LoadPrediction, trends TrendAnalysis) CapacityRequirement {
	// Safety margin for capacity planning
	safetyMargin := 1.3

	peakLoad := load.OrdersPerSecond * load.PeakMultiplier * safetyMargin

	return CapacityRequirement{
		ComputeUnits:     int(math.Ceil(peakLoad / 100)), // 100 orders per compute unit
		MemoryGB:         peakLoad * 0.1,                  // 100MB per orders/sec
		StorageGB:        peakLoad * 24 * 0.001,          // Daily storage needs
		NetworkBandwidth: peakLoad * 2,                   // 2 Mbps per orders/sec
		DatabaseIOPS:     int(peakLoad * 10),             // 10 IOPS per orders/sec
	}
}

func (ia *IntelligentAnalyzer) calculatePredictionConfidence(snapshots []metrics.MetricsSnapshot, prediction LoadPrediction) ConfidenceInterval {
	ordersPerSec := ia.extractOrdersPerSecond(snapshots)
	stdDev := ia.calculateStandardDeviation(ordersPerSec, ia.calculateMean(ordersPerSec))

	// 95% confidence interval
	margin := 1.96 * stdDev / math.Sqrt(float64(len(ordersPerSec)))

	return ConfidenceInterval{
		Lower:      prediction.OrdersPerSecond - margin,
		Upper:      prediction.OrdersPerSecond + margin,
		Confidence: 0.95,
	}
}

func (ia *IntelligentAnalyzer) identifyCapacityRiskFactors(trends TrendAnalysis) []string {
	var risks []string

	if trends.TrendStrength > 0.7 {
		risks = append(risks, "High trend strength indicates potential rapid changes")
	}

	if trends.LatencyTrend == TrendIncreasing {
		risks = append(risks, "Increasing latency trend may indicate capacity constraints")
	}

	if trends.ThroughputTrend == TrendVolatile {
		risks = append(risks, "Volatile throughput makes capacity prediction uncertain")
	}

	if len(risks) == 0 {
		risks = append(risks, "No significant risk factors identified")
	}

	return risks
}

// Severity calculation methods

func (ia *IntelligentAnalyzer) calculateLatencySeverity(p95, threshold, stdDev float64) float64 {
	// Severity based on how far above threshold and variability
	excessRatio := (p95 - threshold) / threshold
	variabilityFactor := stdDev / threshold

	severity := excessRatio + variabilityFactor*0.3
	return math.Min(1.0, severity)
}

// Placeholder methods for recommendation generation (to be implemented in recommendations.go)

func (ia *IntelligentAnalyzer) generateBottleneckRecommendations(bottleneck Bottleneck) []Recommendation {
	// This will be implemented in recommendations.go
	return []Recommendation{}
}

func (ia *IntelligentAnalyzer) generateTrendBasedRecommendations(trends TrendAnalysis) []Recommendation {
	// This will be implemented in recommendations.go
	return []Recommendation{}
}

func (ia *IntelligentAnalyzer) generateCapacityRecommendations(prediction CapacityPrediction) []Recommendation {
	// This will be implemented in recommendations.go
	return []Recommendation{}
}

func (ia *IntelligentAnalyzer) prioritizeRecommendations(recommendations []Recommendation) []Recommendation {
	// This will be implemented in recommendations.go
	return recommendations
}