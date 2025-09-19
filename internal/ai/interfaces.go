package ai

import (
	"time"

	"simulated_exchange/internal/metrics"
)

// PerformanceAI interface defines AI-powered performance analysis capabilities
type PerformanceAI interface {
	// AnalyzeBottlenecks identifies performance bottlenecks from metrics data
	AnalyzeBottlenecks(snapshots []metrics.MetricsSnapshot) []Bottleneck

	// PredictCapacity forecasts system capacity needs based on trends
	PredictCapacity(snapshots []metrics.MetricsSnapshot, timeHorizon time.Duration) CapacityPrediction

	// GenerateRecommendations creates actionable recommendations for optimization
	GenerateRecommendations(analysis PerformanceAnalysis) []Recommendation
}

// BusinessImpactCalculator interface for calculating business value of optimizations
type BusinessImpactCalculator interface {
	// CalculateROI computes return on investment for performance improvements
	CalculateROI(recommendation Recommendation, currentMetrics metrics.MetricsSnapshot) ROIAnalysis

	// EstimateCostSavings projects cost savings from optimization recommendations
	EstimateCostSavings(recommendations []Recommendation, timeHorizon time.Duration) CostSavingsEstimate
}

// ReportGenerator interface for creating executive reports
type ReportGenerator interface {
	// GenerateExecutiveReport creates comprehensive performance reports
	GenerateExecutiveReport(analysis PerformanceAnalysis, format ReportFormat) ([]byte, error)

	// GenerateSummaryReport creates concise summary reports
	GenerateSummaryReport(analysis PerformanceAnalysis) ExecutiveSummary
}

// Data structures for AI analysis

// Bottleneck represents a performance bottleneck with severity and impact
type Bottleneck struct {
	Type           BottleneckType    `json:"type"`
	Component      string            `json:"component"`
	Severity       float64           `json:"severity"`        // 0.0 to 1.0
	Impact         BusinessImpact    `json:"impact"`
	Description    string            `json:"description"`
	DetectedAt     time.Time         `json:"detected_at"`
	AffectedMetrics []string         `json:"affected_metrics"`
	Confidence     float64           `json:"confidence"`      // 0.0 to 1.0
}

// BottleneckType enum for different types of bottlenecks
type BottleneckType string

const (
	BottleneckTypeLatency    BottleneckType = "LATENCY"
	BottleneckTypeThroughput BottleneckType = "THROUGHPUT"
	BottleneckTypeMemory     BottleneckType = "MEMORY"
	BottleneckTypeCPU        BottleneckType = "CPU"
	BottleneckTypeIO         BottleneckType = "IO"
	BottleneckTypeNetwork    BottleneckType = "NETWORK"
	BottleneckTypeDatabase   BottleneckType = "DATABASE"
	BottleneckTypeCapacity   BottleneckType = "CAPACITY"
)

// Recommendation represents an AI-generated optimization recommendation
type Recommendation struct {
	Type         RecommendationType `json:"type"`
	Title        string             `json:"title"`
	Description  string             `json:"description"`
	Impact       BusinessImpact     `json:"impact"`
	Priority     Priority           `json:"priority"`
	Category     string             `json:"category"`
	Complexity   Complexity         `json:"complexity"`
	TimeToEffect time.Duration      `json:"time_to_effect"`
	Prerequisites []string          `json:"prerequisites"`
	Metrics      []string           `json:"metrics"`
	Confidence   float64            `json:"confidence"`
	CreatedAt    time.Time          `json:"created_at"`
}

// RecommendationType enum for different recommendation types
type RecommendationType string

const (
	RecommendationTypeScaling      RecommendationType = "SCALING"
	RecommendationTypeOptimization RecommendationType = "OPTIMIZATION"
	RecommendationTypeArchitecture RecommendationType = "ARCHITECTURE"
	RecommendationTypeCapacity     RecommendationType = "CAPACITY"
	RecommendationTypeMonitoring   RecommendationType = "MONITORING"
	RecommendationTypeMaintenance  RecommendationType = "MAINTENANCE"
)

// BusinessImpact represents the business impact of a bottleneck or recommendation
type BusinessImpact struct {
	Revenue           float64 `json:"revenue"`             // Potential revenue impact
	Cost              float64 `json:"cost"`                // Cost impact or savings
	UserExperience    float64 `json:"user_experience"`     // UX impact score (0-1)
	Reliability       float64 `json:"reliability"`         // Reliability impact (0-1)
	Scalability       float64 `json:"scalability"`         // Scalability impact (0-1)
	OverallScore      float64 `json:"overall_score"`       // Weighted overall impact
}

// Priority enum for recommendation prioritization
type Priority string

const (
	PriorityCritical Priority = "CRITICAL"
	PriorityHigh     Priority = "HIGH"
	PriorityMedium   Priority = "MEDIUM"
	PriorityLow      Priority = "LOW"
)

// Complexity enum for implementation complexity
type Complexity string

const (
	ComplexityLow    Complexity = "LOW"
	ComplexityMedium Complexity = "MEDIUM"
	ComplexityHigh   Complexity = "HIGH"
)

// CapacityPrediction represents future capacity needs
type CapacityPrediction struct {
	TimeHorizon         time.Duration      `json:"time_horizon"`
	PredictedLoad       LoadPrediction     `json:"predicted_load"`
	RecommendedCapacity CapacityRequirement `json:"recommended_capacity"`
	ConfidenceInterval  ConfidenceInterval `json:"confidence_interval"`
	Assumptions         []string           `json:"assumptions"`
	RiskFactors         []string           `json:"risk_factors"`
	CreatedAt           time.Time          `json:"created_at"`
}

// LoadPrediction represents predicted system load
type LoadPrediction struct {
	OrdersPerSecond float64 `json:"orders_per_second"`
	TradesPerSecond float64 `json:"trades_per_second"`
	PeakMultiplier  float64 `json:"peak_multiplier"`
	GrowthRate      float64 `json:"growth_rate"`
}

// CapacityRequirement represents recommended system capacity
type CapacityRequirement struct {
	ComputeUnits    int     `json:"compute_units"`
	MemoryGB        float64 `json:"memory_gb"`
	StorageGB       float64 `json:"storage_gb"`
	NetworkBandwidth float64 `json:"network_bandwidth_mbps"`
	DatabaseIOPS    int     `json:"database_iops"`
}

// ConfidenceInterval represents statistical confidence bounds
type ConfidenceInterval struct {
	Lower      float64 `json:"lower"`
	Upper      float64 `json:"upper"`
	Confidence float64 `json:"confidence"` // e.g., 0.95 for 95%
}

// ROIAnalysis represents return on investment analysis
type ROIAnalysis struct {
	InitialInvestment  float64       `json:"initial_investment"`
	AnnualSavings      float64       `json:"annual_savings"`
	PaybackPeriod      time.Duration `json:"payback_period"`
	ROIPercentage      float64       `json:"roi_percentage"`
	NPV                float64       `json:"npv"`                // Net Present Value
	IRR                float64       `json:"irr"`                // Internal Rate of Return
	RiskAdjustment     float64       `json:"risk_adjustment"`
	Assumptions        []string      `json:"assumptions"`
	SensitivityAnalysis map[string]float64 `json:"sensitivity_analysis"`
}

// CostSavingsEstimate represents projected cost savings
type CostSavingsEstimate struct {
	TimeHorizon     time.Duration            `json:"time_horizon"`
	TotalSavings    float64                  `json:"total_savings"`
	SavingsByCategory map[string]float64     `json:"savings_by_category"`
	ImplementationCost float64               `json:"implementation_cost"`
	NetSavings      float64                  `json:"net_savings"`
	ConfidenceLevel float64                  `json:"confidence_level"`
	Breakdown       []CostSavingsBreakdown   `json:"breakdown"`
}

// CostSavingsBreakdown represents detailed savings breakdown
type CostSavingsBreakdown struct {
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Frequency   string    `json:"frequency"`
	Confidence  float64   `json:"confidence"`
}

// PerformanceAnalysis represents comprehensive performance analysis
type PerformanceAnalysis struct {
	ID                  string               `json:"id"`
	Timestamp           time.Time            `json:"timestamp"`
	TimeRange           TimeRange            `json:"time_range"`
	Bottlenecks         []Bottleneck         `json:"bottlenecks"`
	Recommendations     []Recommendation     `json:"recommendations"`
	CapacityPrediction  CapacityPrediction   `json:"capacity_prediction"`
	TrendAnalysis       TrendAnalysis        `json:"trend_analysis"`
	PerformanceScore    float64              `json:"performance_score"`
	HealthStatus        HealthStatus         `json:"health_status"`
	BusinessImpact      BusinessImpact       `json:"business_impact"`
	Confidence          float64              `json:"confidence"`
}

// TimeRange represents a time period for analysis
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	LatencyTrend    TrendDirection `json:"latency_trend"`
	ThroughputTrend TrendDirection `json:"throughput_trend"`
	VolumeTrend     TrendDirection `json:"volume_trend"`
	ErrorRateTrend  TrendDirection `json:"error_rate_trend"`
	TrendStrength   float64        `json:"trend_strength"`
	Seasonality     []SeasonalPattern `json:"seasonality"`
}

// TrendDirection enum for trend analysis
type TrendDirection string

const (
	TrendIncreasing TrendDirection = "INCREASING"
	TrendDecreasing TrendDirection = "DECREASING"
	TrendStable     TrendDirection = "STABLE"
	TrendVolatile   TrendDirection = "VOLATILE"
)

// SeasonalPattern represents detected seasonal patterns
type SeasonalPattern struct {
	Pattern     string    `json:"pattern"`
	Strength    float64   `json:"strength"`
	Period      time.Duration `json:"period"`
	Description string    `json:"description"`
}

// HealthStatus represents overall system health
type HealthStatus string

const (
	HealthExcellent HealthStatus = "EXCELLENT"
	HealthGood      HealthStatus = "GOOD"
	HealthFair      HealthStatus = "FAIR"
	HealthPoor      HealthStatus = "POOR"
	HealthCritical  HealthStatus = "CRITICAL"
)

// ExecutiveSummary represents a high-level summary for executives
type ExecutiveSummary struct {
	OverallHealth       HealthStatus      `json:"overall_health"`
	PerformanceScore    float64           `json:"performance_score"`
	CriticalIssues      int               `json:"critical_issues"`
	KeyRecommendations  []string          `json:"key_recommendations"`
	BusinessImpact      BusinessImpact    `json:"business_impact"`
	ROISummary          ROISummaryItem    `json:"roi_summary"`
	NextActions         []string          `json:"next_actions"`
	GeneratedAt         time.Time         `json:"generated_at"`
}

// ROISummaryItem represents summarized ROI information
type ROISummaryItem struct {
	TotalInvestment   float64 `json:"total_investment"`
	ExpectedSavings   float64 `json:"expected_savings"`
	PaybackMonths     int     `json:"payback_months"`
	OverallROI        float64 `json:"overall_roi"`
}

// ReportFormat enum for different report formats
type ReportFormat string

const (
	ReportFormatJSON     ReportFormat = "JSON"
	ReportFormatText     ReportFormat = "TEXT"
	ReportFormatMarkdown ReportFormat = "MARKDOWN"
	ReportFormatPDF      ReportFormat = "PDF"
)

// MLAnalysisConfig represents configuration for machine learning analysis
type MLAnalysisConfig struct {
	MinDataPoints         int           `json:"min_data_points"`
	ConfidenceThreshold   float64       `json:"confidence_threshold"`
	TrendSmoothingFactor  float64       `json:"trend_smoothing_factor"`
	SeasonalityLookback   time.Duration `json:"seasonality_lookback"`
	OutlierDetectionSigma float64       `json:"outlier_detection_sigma"`
	PredictionHorizon     time.Duration `json:"prediction_horizon"`
}

// DefaultMLAnalysisConfig returns default ML analysis configuration
func DefaultMLAnalysisConfig() MLAnalysisConfig {
	return MLAnalysisConfig{
		MinDataPoints:         10,
		ConfidenceThreshold:   0.7,
		TrendSmoothingFactor:  0.3,
		SeasonalityLookback:   24 * time.Hour,
		OutlierDetectionSigma: 2.5,
		PredictionHorizon:     1 * time.Hour,
	}
}

// AnalysisError represents errors in AI analysis
type AnalysisError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func (e AnalysisError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrorCodeInsufficientData = "INSUFFICIENT_DATA"
	ErrorCodeInvalidInput     = "INVALID_INPUT"
	ErrorCodeAnalysisFailed   = "ANALYSIS_FAILED"
	ErrorCodePredictionFailed = "PREDICTION_FAILED"
	ErrorCodeConfigError      = "CONFIG_ERROR"
)