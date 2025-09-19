package ai

import (
	"fmt"
	"sort"
	"time"
)

// RecommendationEngine generates intelligent recommendations based on performance analysis
type RecommendationEngine struct {
	config                RecommendationConfig
	templateRecommendations map[BottleneckType][]RecommendationTemplate
	impactCalculator      BusinessImpactCalculator
}

// RecommendationConfig holds configuration for recommendation generation
type RecommendationConfig struct {
	MaxRecommendations     int           `json:"max_recommendations"`
	MinConfidenceThreshold float64       `json:"min_confidence_threshold"`
	PriorityWeights        PriorityWeights `json:"priority_weights"`
	DefaultTimeToEffect    time.Duration `json:"default_time_to_effect"`
}

// PriorityWeights defines weights for different impact factors
type PriorityWeights struct {
	Revenue        float64 `json:"revenue"`
	Cost           float64 `json:"cost"`
	UserExperience float64 `json:"user_experience"`
	Reliability    float64 `json:"reliability"`
	Scalability    float64 `json:"scalability"`
}

// RecommendationTemplate defines templates for generating recommendations
type RecommendationTemplate struct {
	Type             RecommendationType `json:"type"`
	TitleTemplate    string            `json:"title_template"`
	DescriptionTemplate string         `json:"description_template"`
	Category         string            `json:"category"`
	Complexity       Complexity        `json:"complexity"`
	TimeToEffect     time.Duration     `json:"time_to_effect"`
	Prerequisites    []string          `json:"prerequisites"`
	ApplicabilityConditions []string   `json:"applicability_conditions"`
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(config RecommendationConfig, impactCalculator BusinessImpactCalculator) *RecommendationEngine {
	engine := &RecommendationEngine{
		config:                  config,
		templateRecommendations: make(map[BottleneckType][]RecommendationTemplate),
		impactCalculator:       impactCalculator,
	}

	engine.initializeTemplates()
	return engine
}

// NewDefaultRecommendationEngine creates engine with default configuration
func NewDefaultRecommendationEngine(impactCalculator BusinessImpactCalculator) *RecommendationEngine {
	return NewRecommendationEngine(DefaultRecommendationConfig(), impactCalculator)
}

// generateBottleneckRecommendations generates recommendations for specific bottlenecks
func (re *RecommendationEngine) generateBottleneckRecommendations(bottleneck Bottleneck) []Recommendation {
	var recommendations []Recommendation

	templates, exists := re.templateRecommendations[bottleneck.Type]
	if !exists {
		return recommendations
	}

	for _, template := range templates {
		if re.isTemplateApplicable(template, bottleneck) {
			rec := re.createRecommendationFromTemplate(template, bottleneck)
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations
}

// generateTrendBasedRecommendations creates proactive recommendations based on trends
func (re *RecommendationEngine) generateTrendBasedRecommendations(trends TrendAnalysis) []Recommendation {
	var recommendations []Recommendation

	// Proactive scaling recommendations based on trends
	if trends.ThroughputTrend == TrendIncreasing && trends.TrendStrength > 0.6 {
		rec := Recommendation{
			Type:        RecommendationTypeScaling,
			Title:       "Proactive Scaling Recommendation",
			Description: fmt.Sprintf("System throughput is trending upward (strength: %.2f). Consider scaling infrastructure proactively.", trends.TrendStrength),
			Impact: BusinessImpact{
				Revenue:        25000,  // Prevent revenue loss
				Cost:           -15000, // Infrastructure cost
				UserExperience: 0.9,
				Reliability:    0.95,
				Scalability:    0.98,
				OverallScore:   0.92,
			},
			Priority:      PriorityHigh,
			Category:      "Infrastructure",
			Complexity:    ComplexityMedium,
			TimeToEffect:  2 * time.Hour,
			Prerequisites: []string{"Infrastructure team approval", "Capacity planning review"},
			Metrics:       []string{"throughput_trend", "capacity_utilization"},
			Confidence:    0.8,
			CreatedAt:     time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	// Latency optimization recommendations
	if trends.LatencyTrend == TrendIncreasing && trends.TrendStrength > 0.5 {
		rec := Recommendation{
			Type:        RecommendationTypeOptimization,
			Title:       "Latency Optimization Required",
			Description: "Increasing latency trend detected. Implement performance optimizations before user impact occurs.",
			Impact: BusinessImpact{
				Revenue:        15000,
				Cost:           -8000,
				UserExperience: 0.85,
				Reliability:    0.9,
				Scalability:    0.8,
				OverallScore:   0.85,
			},
			Priority:     PriorityHigh,
			Category:     "Performance",
			Complexity:   ComplexityMedium,
			TimeToEffect: 4 * time.Hour,
			Prerequisites: []string{"Performance team analysis", "Code review"},
			Metrics:      []string{"latency_trend", "response_time"},
			Confidence:   0.75,
			CreatedAt:    time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	// Stability improvements for volatile trends
	if trends.LatencyTrend == TrendVolatile || trends.ThroughputTrend == TrendVolatile {
		rec := Recommendation{
			Type:        RecommendationTypeArchitecture,
			Title:       "System Stability Enhancement",
			Description: "Volatile performance patterns detected. Consider implementing circuit breakers and load balancing improvements.",
			Impact: BusinessImpact{
				Revenue:        10000,
				Cost:           -12000,
				UserExperience: 0.88,
				Reliability:    0.95,
				Scalability:    0.9,
				OverallScore:   0.88,
			},
			Priority:     PriorityMedium,
			Category:     "Architecture",
			Complexity:   ComplexityHigh,
			TimeToEffect: 8 * time.Hour,
			Prerequisites: []string{"Architecture review", "Testing plan"},
			Metrics:      []string{"volatility_index", "stability_score"},
			Confidence:   0.7,
			CreatedAt:    time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	// Monitoring enhancements for complex trends
	if len(trends.Seasonality) > 0 {
		rec := Recommendation{
			Type:        RecommendationTypeMonitoring,
			Title:       "Enhanced Monitoring for Seasonal Patterns",
			Description: "Seasonal patterns detected. Implement predictive monitoring and auto-scaling based on patterns.",
			Impact: BusinessImpact{
				Revenue:        8000,
				Cost:           -5000,
				UserExperience: 0.8,
				Reliability:    0.9,
				Scalability:    0.95,
				OverallScore:   0.83,
			},
			Priority:     PriorityMedium,
			Category:     "Monitoring",
			Complexity:   ComplexityLow,
			TimeToEffect: 1 * time.Hour,
			Prerequisites: []string{"Monitoring tool configuration"},
			Metrics:      []string{"seasonal_patterns", "prediction_accuracy"},
			Confidence:   0.85,
			CreatedAt:    time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// generateCapacityRecommendations creates recommendations based on capacity predictions
func (re *RecommendationEngine) generateCapacityRecommendations(prediction CapacityPrediction) []Recommendation {
	var recommendations []Recommendation

	// Calculate current vs predicted capacity gap
	currentEstimate := CapacityRequirement{
		ComputeUnits:     5,    // Current estimate
		MemoryGB:         32,   // Current estimate
		StorageGB:        1000, // Current estimate
		NetworkBandwidth: 100,  // Current estimate
		DatabaseIOPS:     500,  // Current estimate
	}

	// Generate scaling recommendations if capacity gap exists
	if prediction.RecommendedCapacity.ComputeUnits > currentEstimate.ComputeUnits {
		gap := prediction.RecommendedCapacity.ComputeUnits - currentEstimate.ComputeUnits
		rec := Recommendation{
			Type:        RecommendationTypeCapacity,
			Title:       "Compute Capacity Scaling Required",
			Description: fmt.Sprintf("Predicted capacity needs require %d additional compute units within %v", gap, prediction.TimeHorizon),
			Impact: BusinessImpact{
				Revenue:        float64(gap) * 5000,  // Revenue protection
				Cost:           -float64(gap) * 2000, // Infrastructure cost
				UserExperience: 0.9,
				Reliability:    0.95,
				Scalability:    0.98,
				OverallScore:   0.91,
			},
			Priority:     re.calculatePriorityFromTimeHorizon(prediction.TimeHorizon),
			Category:     "Capacity Planning",
			Complexity:   ComplexityMedium,
			TimeToEffect: prediction.TimeHorizon / 4, // Plan for 1/4 of horizon
			Prerequisites: []string{"Budget approval", "Infrastructure provisioning"},
			Metrics:      []string{"compute_utilization", "capacity_forecast"},
			Confidence:   prediction.ConfidenceInterval.Confidence,
			CreatedAt:    time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	// Memory scaling recommendations
	if prediction.RecommendedCapacity.MemoryGB > currentEstimate.MemoryGB*1.2 {
		memoryGap := prediction.RecommendedCapacity.MemoryGB - currentEstimate.MemoryGB
		rec := Recommendation{
			Type:        RecommendationTypeCapacity,
			Title:       "Memory Capacity Enhancement",
			Description: fmt.Sprintf("Additional %.0f GB memory required for predicted load growth", memoryGap),
			Impact: BusinessImpact{
				Revenue:        memoryGap * 500,
				Cost:           -memoryGap * 200,
				UserExperience: 0.85,
				Reliability:    0.9,
				Scalability:    0.95,
				OverallScore:   0.87,
			},
			Priority:     PriorityMedium,
			Category:     "Infrastructure",
			Complexity:   ComplexityLow,
			TimeToEffect: 2 * time.Hour,
			Prerequisites: []string{"Memory upgrade planning"},
			Metrics:      []string{"memory_utilization", "memory_forecast"},
			Confidence:   prediction.ConfidenceInterval.Confidence * 0.9,
			CreatedAt:    time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	// Risk mitigation recommendations
	if len(prediction.RiskFactors) > 0 {
		rec := Recommendation{
			Type:        RecommendationTypeMonitoring,
			Title:       "Risk Mitigation Strategy",
			Description: fmt.Sprintf("Implement monitoring and alerting for %d identified risk factors", len(prediction.RiskFactors)),
			Impact: BusinessImpact{
				Revenue:        10000,
				Cost:           -3000,
				UserExperience: 0.8,
				Reliability:    0.95,
				Scalability:    0.85,
				OverallScore:   0.85,
			},
			Priority:     PriorityMedium,
			Category:     "Risk Management",
			Complexity:   ComplexityLow,
			TimeToEffect: 30 * time.Minute,
			Prerequisites: []string{"Risk assessment review"},
			Metrics:      []string{"risk_indicators", "alert_effectiveness"},
			Confidence:   0.8,
			CreatedAt:    time.Now(),
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations
}

// prioritizeRecommendations sorts and prioritizes recommendations
func (re *RecommendationEngine) prioritizeRecommendations(recommendations []Recommendation) []Recommendation {
	// Calculate priority scores for all recommendations
	for i := range recommendations {
		recommendations[i] = re.calculatePriorityScore(recommendations[i])
	}
	// Sort by priority score (highest first)
	sort.Slice(recommendations, func(i, j int) bool {
		return re.getNumericPriority(recommendations[i].Priority) > re.getNumericPriority(recommendations[j].Priority)
	})

	// Remove duplicates and low-confidence recommendations
	filteredRecommendations := re.filterRecommendations(recommendations)

	// Limit to maximum recommendations
	if len(filteredRecommendations) > re.config.MaxRecommendations {
		filteredRecommendations = filteredRecommendations[:re.config.MaxRecommendations]
	}

	return filteredRecommendations
}

// Helper methods for recommendation engine

// initializeTemplates sets up recommendation templates for different bottleneck types
func (re *RecommendationEngine) initializeTemplates() {
	// Latency bottleneck templates
	re.templateRecommendations[BottleneckTypeLatency] = []RecommendationTemplate{
		{
			Type:                    RecommendationTypeOptimization,
			TitleTemplate:           "Optimize %s Performance",
			DescriptionTemplate:     "High latency detected in %s (%.2fms). Consider query optimization, caching, or algorithm improvements.",
			Category:                "Performance",
			Complexity:              ComplexityMedium,
			TimeToEffect:            2 * time.Hour,
			Prerequisites:           []string{"Performance profiling", "Code review"},
			ApplicabilityConditions: []string{"severity > 0.6"},
		},
		{
			Type:                    RecommendationTypeScaling,
			TitleTemplate:           "Scale %s Infrastructure",
			DescriptionTemplate:     "Critical latency issues in %s require immediate infrastructure scaling.",
			Category:                "Infrastructure",
			Complexity:              ComplexityHigh,
			TimeToEffect:            30 * time.Minute,
			Prerequisites:           []string{"Infrastructure approval", "Load balancer configuration"},
			ApplicabilityConditions: []string{"severity > 0.8"},
		},
	}

	// Throughput bottleneck templates
	re.templateRecommendations[BottleneckTypeThroughput] = []RecommendationTemplate{
		{
			Type:                    RecommendationTypeScaling,
			TitleTemplate:           "Increase %s Capacity",
			DescriptionTemplate:     "Low throughput in %s (%.2f ops/sec). Scale horizontally or optimize processing pipeline.",
			Category:                "Capacity",
			Complexity:              ComplexityMedium,
			TimeToEffect:            1 * time.Hour,
			Prerequisites:           []string{"Capacity planning", "Auto-scaling configuration"},
			ApplicabilityConditions: []string{"severity > 0.5"},
		},
		{
			Type:                    RecommendationTypeOptimization,
			TitleTemplate:           "Optimize %s Processing",
			DescriptionTemplate:     "Implement parallel processing and batch optimizations for improved throughput.",
			Category:                "Architecture",
			Complexity:              ComplexityHigh,
			TimeToEffect:            4 * time.Hour,
			Prerequisites:           []string{"Architecture review", "Testing framework"},
			ApplicabilityConditions: []string{"severity > 0.7"},
		},
	}

	// Memory bottleneck templates
	re.templateRecommendations[BottleneckTypeMemory] = []RecommendationTemplate{
		{
			Type:                    RecommendationTypeCapacity,
			TitleTemplate:           "Memory Capacity Upgrade",
			DescriptionTemplate:     "High memory pressure detected (%.2f MB). Increase memory allocation or implement memory optimization.",
			Category:                "Infrastructure",
			Complexity:              ComplexityLow,
			TimeToEffect:            1 * time.Hour,
			Prerequisites:           []string{"Memory monitoring", "Resource allocation"},
			ApplicabilityConditions: []string{"severity > 0.4"},
		},
		{
			Type:                    RecommendationTypeOptimization,
			TitleTemplate:           "Memory Usage Optimization",
			DescriptionTemplate:     "Implement memory pooling, garbage collection tuning, and data structure optimization.",
			Category:                "Performance",
			Complexity:              ComplexityMedium,
			TimeToEffect:            3 * time.Hour,
			Prerequisites:           []string{"Memory profiling", "Performance testing"},
			ApplicabilityConditions: []string{"severity > 0.6"},
		},
	}

	// Add templates for other bottleneck types...
	re.templateRecommendations[BottleneckTypeCPU] = []RecommendationTemplate{
		{
			Type:                    RecommendationTypeScaling,
			TitleTemplate:           "CPU Scaling Required",
			DescriptionTemplate:     "High CPU utilization detected. Scale compute resources or optimize CPU-intensive operations.",
			Category:                "Infrastructure",
			Complexity:              ComplexityMedium,
			TimeToEffect:            45 * time.Minute,
			Prerequisites:           []string{"CPU monitoring", "Load balancing"},
			ApplicabilityConditions: []string{"severity > 0.6"},
		},
	}

	re.templateRecommendations[BottleneckTypeIO] = []RecommendationTemplate{
		{
			Type:                    RecommendationTypeOptimization,
			TitleTemplate:           "I/O Performance Enhancement",
			DescriptionTemplate:     "I/O bottleneck detected. Consider SSD upgrade, connection pooling, or async I/O implementation.",
			Category:                "Infrastructure",
			Complexity:              ComplexityMedium,
			TimeToEffect:            2 * time.Hour,
			Prerequisites:           []string{"I/O profiling", "Storage analysis"},
			ApplicabilityConditions: []string{"severity > 0.5"},
		},
	}
}

// isTemplateApplicable checks if a template applies to the given bottleneck
func (re *RecommendationEngine) isTemplateApplicable(template RecommendationTemplate, bottleneck Bottleneck) bool {
	// Check severity threshold
	if bottleneck.Severity < 0.3 {
		return false
	}

	// Check confidence threshold
	if bottleneck.Confidence < re.config.MinConfidenceThreshold {
		return false
	}

	// Check specific applicability conditions
	for _, condition := range template.ApplicabilityConditions {
		if !re.evaluateCondition(condition, bottleneck) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates applicability conditions
func (re *RecommendationEngine) evaluateCondition(condition string, bottleneck Bottleneck) bool {
	// Simplified condition evaluation
	// In a real implementation, this would be more sophisticated
	switch condition {
	case "severity > 0.6":
		return bottleneck.Severity > 0.6
	case "severity > 0.8":
		return bottleneck.Severity > 0.8
	case "severity > 0.5":
		return bottleneck.Severity > 0.5
	case "severity > 0.7":
		return bottleneck.Severity > 0.7
	case "severity > 0.4":
		return bottleneck.Severity > 0.4
	default:
		return true
	}
}

// createRecommendationFromTemplate creates a recommendation from a template
func (re *RecommendationEngine) createRecommendationFromTemplate(template RecommendationTemplate, bottleneck Bottleneck) Recommendation {
	// Format template strings with bottleneck data
	title := fmt.Sprintf(template.TitleTemplate, bottleneck.Component)
	description := fmt.Sprintf(template.DescriptionTemplate, bottleneck.Component, bottleneck.Severity*100)

	// Calculate priority based on severity and impact
	priority := re.calculatePriorityFromSeverity(bottleneck.Severity)

	return Recommendation{
		Type:          template.Type,
		Title:         title,
		Description:   description,
		Impact:        bottleneck.Impact,
		Priority:      priority,
		Category:      template.Category,
		Complexity:    template.Complexity,
		TimeToEffect:  template.TimeToEffect,
		Prerequisites: template.Prerequisites,
		Metrics:       bottleneck.AffectedMetrics,
		Confidence:    bottleneck.Confidence,
		CreatedAt:     time.Now(),
	}
}

// calculatePriorityFromSeverity determines priority based on severity
func (re *RecommendationEngine) calculatePriorityFromSeverity(severity float64) Priority {
	if severity >= 0.9 {
		return PriorityCritical
	} else if severity >= 0.7 {
		return PriorityHigh
	} else if severity >= 0.4 {
		return PriorityMedium
	}
	return PriorityLow
}

// calculatePriorityFromTimeHorizon determines priority based on time sensitivity
func (re *RecommendationEngine) calculatePriorityFromTimeHorizon(horizon time.Duration) Priority {
	if horizon <= 1*time.Hour {
		return PriorityCritical
	} else if horizon <= 6*time.Hour {
		return PriorityHigh
	} else if horizon <= 24*time.Hour {
		return PriorityMedium
	}
	return PriorityLow
}

// calculatePriorityScore calculates a numeric priority score for sorting
func (re *RecommendationEngine) calculatePriorityScore(rec Recommendation) Recommendation {
	weights := re.config.PriorityWeights

	// Calculate weighted impact score
	impactScore := rec.Impact.Revenue*weights.Revenue +
		rec.Impact.UserExperience*weights.UserExperience +
		rec.Impact.Reliability*weights.Reliability +
		rec.Impact.Scalability*weights.Scalability -
		rec.Impact.Cost*weights.Cost

	// Adjust score based on confidence and complexity
	confidenceMultiplier := rec.Confidence
	complexityPenalty := re.getComplexityPenalty(rec.Complexity)

	finalScore := impactScore * confidenceMultiplier * complexityPenalty

	// Store score in Impact.OverallScore for sorting
	rec.Impact.OverallScore = finalScore

	return rec
}

// getNumericPriority converts priority enum to numeric value for sorting
func (re *RecommendationEngine) getNumericPriority(priority Priority) float64 {
	switch priority {
	case PriorityCritical:
		return 4.0
	case PriorityHigh:
		return 3.0
	case PriorityMedium:
		return 2.0
	case PriorityLow:
		return 1.0
	default:
		return 0.0
	}
}

// getComplexityPenalty returns penalty factor for complexity
func (re *RecommendationEngine) getComplexityPenalty(complexity Complexity) float64 {
	switch complexity {
	case ComplexityLow:
		return 1.0
	case ComplexityMedium:
		return 0.8
	case ComplexityHigh:
		return 0.6
	default:
		return 0.5
	}
}

// filterRecommendations removes duplicates and low-quality recommendations
func (re *RecommendationEngine) filterRecommendations(recommendations []Recommendation) []Recommendation {
	var filtered []Recommendation
	seen := make(map[string]bool)

	for _, rec := range recommendations {
		// Filter by confidence threshold
		if rec.Confidence < re.config.MinConfidenceThreshold {
			continue
		}

		// Remove duplicates based on title
		if seen[rec.Title] {
			continue
		}
		seen[rec.Title] = true

		filtered = append(filtered, rec)
	}

	return filtered
}

// DefaultRecommendationConfig returns default configuration
func DefaultRecommendationConfig() RecommendationConfig {
	return RecommendationConfig{
		MaxRecommendations:     10,
		MinConfidenceThreshold: 0.6,
		PriorityWeights: PriorityWeights{
			Revenue:        0.3,
			Cost:           0.2,
			UserExperience: 0.25,
			Reliability:    0.15,
			Scalability:    0.1,
		},
		DefaultTimeToEffect: 2 * time.Hour,
	}
}

