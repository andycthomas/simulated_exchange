package reporting

import (
	"time"
)

// Missing data structures referenced in interfaces but not yet defined

// Visualization data structures

// GraphData represents graph visualization data
type GraphData struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	XAxis       AxisData               `json:"x_axis"`
	YAxis       AxisData               `json:"y_axis"`
	Series      []DataSeries           `json:"series"`
	Options     map[string]interface{} `json:"options"`
	Width       int                    `json:"width"`
	Height      int                    `json:"height"`
}

// TableData represents tabular data for display
type TableData struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	Headers     []string                 `json:"headers"`
	Rows        [][]interface{}          `json:"rows"`
	Summary     map[string]interface{}   `json:"summary"`
	Formatting  map[string]string        `json:"formatting"`
	Sortable    bool                     `json:"sortable"`
	Filterable  bool                     `json:"filterable"`
}

// DashboardData represents dashboard layout and widgets
type DashboardData struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Layout      string                 `json:"layout"`
	Widgets     []DashboardWidget      `json:"widgets"`
	Filters     []DashboardFilter      `json:"filters"`
	RefreshRate int                    `json:"refresh_rate"`
	Options     map[string]interface{} `json:"options"`
}

// DashboardWidget represents individual dashboard widgets
type DashboardWidget struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Data     interface{}            `json:"data"`
	Position WidgetPosition         `json:"position"`
	Size     WidgetSize             `json:"size"`
	Options  map[string]interface{} `json:"options"`
}

// DashboardFilter represents dashboard filtering options
type DashboardFilter struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Options []string               `json:"options"`
	Default interface{}            `json:"default"`
}

// WidgetPosition represents widget position on dashboard
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WidgetSize represents widget dimensions
type WidgetSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// InteractiveElement represents interactive visualization elements
type InteractiveElement struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Data        interface{}            `json:"data"`
	Actions     []InteractiveAction    `json:"actions"`
	Events      []string               `json:"events"`
	Options     map[string]interface{} `json:"options"`
}

// InteractiveAction represents interactive actions
type InteractiveAction struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
	Callback    string      `json:"callback"`
}

// AxisData represents chart axis configuration
type AxisData struct {
	Label       string      `json:"label"`
	Type        string      `json:"type"`
	Min         float64     `json:"min"`
	Max         float64     `json:"max"`
	Step        float64     `json:"step"`
	Format      string      `json:"format"`
	Categories  []string    `json:"categories"`
	ShowGrid    bool        `json:"show_grid"`
	ShowLabels  bool        `json:"show_labels"`
}

// DataSeries represents chart data series
type DataSeries struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Data        []DataPoint            `json:"data"`
	Color       string                 `json:"color"`
	Visible     bool                   `json:"visible"`
	Options     map[string]interface{} `json:"options"`
}

// DataPoint represents individual data points
type DataPoint struct {
	X     interface{} `json:"x"`
	Y     interface{} `json:"y"`
	Label string      `json:"label"`
	Color string      `json:"color"`
	Size  float64     `json:"size"`
}

// KeyFinding represents a significant finding in the analysis
type KeyFinding struct {
	Finding     string  `json:"finding"`
	Impact      string  `json:"impact"`
	Severity    string  `json:"severity"`
	Category    string  `json:"category"`
	Evidence    string  `json:"evidence"`
	Confidence  float64 `json:"confidence"`
}

// MetricSummary provides a summary of critical metrics
type MetricSummary struct {
	Name        string  `json:"name"`
	Current     float64 `json:"current"`
	Previous    float64 `json:"previous"`
	Target      float64 `json:"target"`
	Trend       string  `json:"trend"`
	Unit        string  `json:"unit"`
	Description string  `json:"description"`
}

// ScoreCard represents overall performance scoring
type ScoreCard struct {
	Score       float64 `json:"score"`
	Grade       Grade   `json:"grade"`
	MaxScore    float64 `json:"max_score"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
}

// Recommendation represents actionable recommendations
type Recommendation struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    string    `json:"priority"`
	Category    string    `json:"category"`
	Impact      string    `json:"impact"`
	Effort      string    `json:"effort"`
	Timeline    string    `json:"timeline"`
	Cost        float64   `json:"cost"`
	Benefits    []string  `json:"benefits"`
	Risks       []string  `json:"risks"`
	Owner       string    `json:"owner"`
	DueDate     time.Time `json:"due_date"`
}

// ComparisonMetrics holds comparison data
type ComparisonMetrics struct {
	CurrentPeriod   map[string]float64 `json:"current_period"`
	PreviousPeriod  map[string]float64 `json:"previous_period"`
	YearOverYear    map[string]float64 `json:"year_over_year"`
	IndustryAverage map[string]float64 `json:"industry_average"`
	BestInClass     map[string]float64 `json:"best_in_class"`
	Variance        map[string]float64 `json:"variance"`
}

// PerformanceFactor represents factors affecting performance
type PerformanceFactor struct {
	Name        string  `json:"name"`
	Impact      float64 `json:"impact"`
	Direction   string  `json:"direction"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
}

// AreaSummary provides summary of specific business areas
type AreaSummary struct {
	Area        string                 `json:"area"`
	Score       float64                `json:"score"`
	Grade       Grade                  `json:"grade"`
	Trend       TrendDirection         `json:"trend"`
	KeyMetrics  map[string]float64     `json:"key_metrics"`
	Strengths   []string               `json:"strengths"`
	Weaknesses  []string               `json:"weaknesses"`
	Actions     []string               `json:"actions"`
	Details     map[string]interface{} `json:"details"`
}

// ImprovementArea identifies areas needing improvement
type ImprovementArea struct {
	Area        string   `json:"area"`
	Priority    string   `json:"priority"`
	Gap         float64  `json:"gap"`
	Potential   float64  `json:"potential"`
	Actions     []string `json:"actions"`
	Timeline    string   `json:"timeline"`
	Investment  float64  `json:"investment"`
	ExpectedROI float64  `json:"expected_roi"`
}

// CostSavingsAnalysis contains cost savings analysis results
type CostSavingsAnalysis struct {
	CurrentCosts          float64                `json:"current_costs"`
	PotentialSavings      float64                `json:"potential_savings"`
	SavingsOpportunities  []SavingsOpportunity   `json:"savings_opportunities"`
	CostBreakdown         CostBreakdown          `json:"cost_breakdown"`
	EfficiencyScore       float64                `json:"efficiency_score"`
	IndustryComparison    map[string]interface{} `json:"industry_comparison"`
	RiskAssessment        map[string]interface{} `json:"risk_assessment"`
	ImplementationPlan    []ImplementationStep   `json:"implementation_plan"`
	Timeline              string                 `json:"timeline"`
}

// CostBreakdown provides detailed cost analysis
type CostBreakdown struct {
	TotalCosts    float64            `json:"total_costs"`
	Categories    map[string]float64 `json:"categories"`
	FixedCosts    float64            `json:"fixed_costs"`
	VariableCosts float64            `json:"variable_costs"`
	DirectCosts   float64            `json:"direct_costs"`
	IndirectCosts float64            `json:"indirect_costs"`
	Trends        []CostTrend        `json:"trends"`
}

// SavingsOpportunity represents a cost savings opportunity
type SavingsOpportunity struct {
	ID                   string  `json:"id"`
	Title                string  `json:"title"`
	Description          string  `json:"description"`
	Category             string  `json:"category"`
	PotentialSavings     float64 `json:"potential_savings"`
	ImplementationCost   float64 `json:"implementation_cost"`
	ImplementationEffort string  `json:"implementation_effort"`
	Timeline             string  `json:"timeline"`
	Risk                 string  `json:"risk"`
	Priority             string  `json:"priority"`
	NetBenefit           float64 `json:"net_benefit"`
	PaybackPeriod        float64 `json:"payback_period"`
}

// CostTrend represents cost trends over time
type CostTrend struct {
	Period    string  `json:"period"`
	Amount    float64 `json:"amount"`
	Change    float64 `json:"change"`
	Category  string  `json:"category"`
	Direction string  `json:"direction"`
}

// EfficiencyMetrics contains efficiency measurements
type EfficiencyMetrics struct {
	OverallEfficiency   float64 `json:"overall_efficiency"`
	ProductivityIndex   float64 `json:"productivity_index"`
	CostPerUnit         float64 `json:"cost_per_unit"`
	ResourceUtilization float64 `json:"resource_utilization"`
	QualityIndex        float64 `json:"quality_index"`
	WasteReduction      float64 `json:"waste_reduction"`
}

// RiskAnalysis contains comprehensive risk analysis
type RiskAnalysis struct {
	OverallRiskScore      float64             `json:"overall_risk_score"`
	RiskLevel            string              `json:"risk_level"`
	RiskCategories       map[string]float64  `json:"risk_categories"`
	RiskMatrix           RiskMatrix          `json:"risk_matrix"`
	TopRisks             []RiskFactor        `json:"top_risks"`
	RiskTrends           []RiskTrend         `json:"risk_trends"`
	MitigationPriorities []MitigationPriority `json:"mitigation_priorities"`
	RecommendedActions   []string            `json:"recommended_actions"`
}

// RiskMatrix represents risk probability vs impact matrix
type RiskMatrix struct {
	Matrix      [][]RiskMatrixCell `json:"matrix"`
	Dimensions  MatrixDimensions   `json:"dimensions"`
	Legend      map[string]string  `json:"legend"`
	RiskCounts  map[string]int     `json:"risk_counts"`
}

// RiskMatrixCell represents a cell in the risk matrix
type RiskMatrixCell struct {
	Probability string       `json:"probability"`
	Impact      string       `json:"impact"`
	RiskLevel   string       `json:"risk_level"`
	RiskCount   int          `json:"risk_count"`
	Risks       []RiskFactor `json:"risks"`
}

// MatrixDimensions defines the risk matrix dimensions
type MatrixDimensions struct {
	ProbabilityLevels []string `json:"probability_levels"`
	ImpactLevels      []string `json:"impact_levels"`
	RiskLevels        []string `json:"risk_levels"`
}

// MitigationPlan represents a risk mitigation strategy
type MitigationPlan struct {
	RiskID          string    `json:"risk_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Strategy        string    `json:"strategy"`
	Actions         []string  `json:"actions"`
	Owner           string    `json:"owner"`
	Timeline        string    `json:"timeline"`
	EstimatedCost   float64   `json:"estimated_cost"`
	ExpectedImpact  string    `json:"expected_impact"`
	Priority        string    `json:"priority"`
	Status          string    `json:"status"`
	StartDate       time.Time `json:"start_date"`
	TargetDate      time.Time `json:"target_date"`
}

// MitigationPriority represents prioritized mitigation strategies
type MitigationPriority struct {
	Priority    int          `json:"priority"`
	RiskFactor  RiskFactor   `json:"risk_factor"`
	Actions     []string     `json:"actions"`
	Timeline    string       `json:"timeline"`
	Investment  float64      `json:"investment"`
	Impact      string       `json:"impact"`
}

// ComplianceStatus represents compliance assessment
type ComplianceStatus struct {
	OverallStatus     string                    `json:"overall_status"`
	ComplianceScore   float64                   `json:"compliance_score"`
	Areas             map[string]ComplianceArea `json:"areas"`
	Violations        []ComplianceViolation     `json:"violations"`
	Recommendations   []string                  `json:"recommendations"`
	NextAuditDate     time.Time                 `json:"next_audit_date"`
	LastAuditDate     time.Time                 `json:"last_audit_date"`
}

// ComplianceArea represents a specific compliance area
type ComplianceArea struct {
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Score       float64 `json:"score"`
	LastReview  time.Time `json:"last_review"`
	NextReview  time.Time `json:"next_review"`
	Issues      []string `json:"issues"`
	Actions     []string `json:"actions"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	ID          string    `json:"id"`
	Area        string    `json:"area"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	DetectedDate time.Time `json:"detected_date"`
	DueDate     time.Time `json:"due_date"`
	Owner       string    `json:"owner"`
}

// RiskTrend represents risk trends over time
type RiskTrend struct {
	Period     string  `json:"period"`
	RiskScore  float64 `json:"risk_score"`
	Category   string  `json:"category"`
	Change     float64 `json:"change"`
	Events     []string `json:"events"`
}

// RiskFactor represents an individual risk factor
type RiskFactor struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Probability float64   `json:"probability"`
	Impact      string    `json:"impact"`
	RiskScore   float64   `json:"risk_score"`
	Priority    string    `json:"priority"`
	Owner       string    `json:"owner"`
	Status      string    `json:"status"`
	LastReview  time.Time `json:"last_review"`
	NextReview  time.Time `json:"next_review"`
}

// TrendAnalysis provides trend analysis across metrics
type TrendAnalysis struct {
	OverallTrend    TrendDirection       `json:"overall_trend"`
	PeriodAnalysis  []PeriodTrend        `json:"period_analysis"`
	MetricTrends    map[string]TrendData `json:"metric_trends"`
	Seasonality     SeasonalAnalysis     `json:"seasonality"`
	Forecasts       []Forecast           `json:"forecasts"`
	Confidence      float64              `json:"confidence"`
}

// PeriodTrend represents trend for a specific period
type PeriodTrend struct {
	Period    string         `json:"period"`
	Trend     TrendDirection `json:"trend"`
	Change    float64        `json:"change"`
	Metrics   map[string]float64 `json:"metrics"`
	Highlights []string      `json:"highlights"`
}

// TrendData contains detailed trend information
type TrendData struct {
	Direction   TrendDirection `json:"direction"`
	Strength    float64        `json:"strength"`
	Consistency float64        `json:"consistency"`
	Acceleration float64       `json:"acceleration"`
	Volatility  float64        `json:"volatility"`
	R2          float64        `json:"r2"`
}

// SeasonalAnalysis provides seasonal pattern analysis
type SeasonalAnalysis struct {
	HasSeasonality   bool                   `json:"has_seasonality"`
	SeasonalStrength float64                `json:"seasonal_strength"`
	Patterns         map[string]interface{} `json:"patterns"`
	PeakPeriods      []string               `json:"peak_periods"`
	LowPeriods       []string               `json:"low_periods"`
}

// Forecast represents future projections
type Forecast struct {
	Period     string  `json:"period"`
	Value      float64 `json:"value"`
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Confidence float64 `json:"confidence"`
	Method     string  `json:"method"`
}

// BenchmarkAnalysis provides benchmark comparison
type BenchmarkAnalysis struct {
	CompanyMetrics      map[string]float64      `json:"company_metrics"`
	IndustryBenchmarks  map[string]float64      `json:"industry_benchmarks"`
	TopPerformers       map[string]float64      `json:"top_performers"`
	Comparisons         map[string]BenchmarkComparison `json:"comparisons"`
	PerformanceGaps     []PerformanceGap        `json:"performance_gaps"`
	CompetitivePosition string                  `json:"competitive_position"`
}

// BenchmarkComparison represents individual metric comparison
type BenchmarkComparison struct {
	MetricName    string  `json:"metric_name"`
	CompanyValue  float64 `json:"company_value"`
	IndustryValue float64 `json:"industry_value"`
	Difference    float64 `json:"difference"`
	PercentDiff   float64 `json:"percent_diff"`
	Ranking       int     `json:"ranking"`
	Quartile      string  `json:"quartile"`
}

// PerformanceGap identifies gaps vs benchmarks
type PerformanceGap struct {
	MetricName     string  `json:"metric_name"`
	GapSize        float64 `json:"gap_size"`
	GapPercent     float64 `json:"gap_percent"`
	Priority       string  `json:"priority"`
	Recommendation string  `json:"recommendation"`
	Actions        []string `json:"actions"`
	Timeline       string  `json:"timeline"`
	Investment     float64 `json:"investment"`
}

// RegionalMetrics provides performance by region
type RegionalMetrics struct {
	Region      string             `json:"region"`
	Performance map[string]float64 `json:"performance"`
	Ranking     int                `json:"ranking"`
	Trend       TrendDirection     `json:"trend"`
	Highlights  []string           `json:"highlights"`
	Challenges  []string           `json:"challenges"`
}

// BenchmarkData contains benchmark information
type BenchmarkData struct {
	IndustryAverages map[string]float64 `json:"industry_averages"`
	TopPerformers    map[string]float64 `json:"top_performers"`
	BottomQuartile   map[string]float64 `json:"bottom_quartile"`
	MedianValues     map[string]float64 `json:"median_values"`
	DataSource       string             `json:"data_source"`
	LastUpdated      time.Time          `json:"last_updated"`
}

// AppendixData contains supplementary information
type AppendixData struct {
	DataSources       []DataSource       `json:"data_sources"`
	Methodology       string             `json:"methodology"`
	Assumptions       []string           `json:"assumptions"`
	Limitations       []string           `json:"limitations"`
	DetailedMetrics   map[string]interface{} `json:"detailed_metrics"`
	RawData           map[string]interface{} `json:"raw_data"`
	Calculations      map[string]interface{} `json:"calculations"`
	Glossary          map[string]string  `json:"glossary"`
}

// DataSource represents a data source
type DataSource struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	LastUpdated time.Time `json:"last_updated"`
	Reliability string    `json:"reliability"`
	URL         string    `json:"url"`
}

// HistoricalRecord represents historical data point
type HistoricalRecord struct {
	Date        time.Time              `json:"date"`
	Period      string                 `json:"period"`
	Metrics     map[string]float64     `json:"metrics"`
	Events      []string               `json:"events"`
	Context     string                 `json:"context"`
	Source      string                 `json:"source"`
	Quality     string                 `json:"quality"`
	Adjustments map[string]interface{} `json:"adjustments"`
}

// ImplementationStep represents a step in implementation plan
type ImplementationStep struct {
	Step        int       `json:"step"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Owner       string    `json:"owner"`
	Resources   []string  `json:"resources"`
	Dependencies []string `json:"dependencies"`
	Deliverables []string `json:"deliverables"`
	Cost        float64   `json:"cost"`
	Status      string    `json:"status"`
}

// CompetitorProfile represents competitor analysis
type CompetitorProfile struct {
	Name        string             `json:"name"`
	MarketShare float64            `json:"market_share"`
	Strengths   []string           `json:"strengths"`
	Weaknesses  []string           `json:"weaknesses"`
	Strategy    string             `json:"strategy"`
	Performance map[string]float64 `json:"performance"`
	ThreatLevel string             `json:"threat_level"`
}

// Stakeholder represents project stakeholder
type Stakeholder struct {
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Department  string   `json:"department"`
	Influence   string   `json:"influence"`
	Interest    string   `json:"interest"`
	Involvement string   `json:"involvement"`
	Contact     string   `json:"contact"`
}

// SuccessMetric represents success measurement criteria
type SuccessMetric struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Target      float64   `json:"target"`
	Unit        string    `json:"unit"`
	Baseline    float64   `json:"baseline"`
	Current     float64   `json:"current"`
	Timeline    string    `json:"timeline"`
	Owner       string    `json:"owner"`
	Status      string    `json:"status"`
}

// InvestmentOption represents alternative investment options
type InvestmentOption struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Cost        float64        `json:"cost"`
	Benefits    []string       `json:"benefits"`
	Risks       []string       `json:"risks"`
	Timeline    string         `json:"timeline"`
	ROI         float64        `json:"roi"`
	Ranking     int            `json:"ranking"`
	Pros        []string       `json:"pros"`
	Cons        []string       `json:"cons"`
}

// SensitivityAnalysis represents sensitivity analysis results
type SensitivityAnalysis struct {
	Variables   []SensitivityVariable `json:"variables"`
	Scenarios   []Scenario           `json:"scenarios"`
	BaseCase    map[string]float64   `json:"base_case"`
	BestCase    map[string]float64   `json:"best_case"`
	WorstCase   map[string]float64   `json:"worst_case"`
	Confidence  float64              `json:"confidence"`
}

// SensitivityVariable represents a variable in sensitivity analysis
type SensitivityVariable struct {
	Name        string  `json:"name"`
	BaseValue   float64 `json:"base_value"`
	MinValue    float64 `json:"min_value"`
	MaxValue    float64 `json:"max_value"`
	Impact      float64 `json:"impact"`
	Correlation float64 `json:"correlation"`
}

// Scenario represents a scenario in analysis
type Scenario struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Probability float64            `json:"probability"`
	Variables   map[string]float64 `json:"variables"`
	Outcomes    map[string]float64 `json:"outcomes"`
	Impact      string             `json:"impact"`
}

// RiskAssessment represents comprehensive risk assessment
type RiskAssessment struct {
	OverallRiskScore     float64            `json:"overall_risk_score"`
	RiskLevel           string             `json:"risk_level"`
	RiskCategories      map[string]float64 `json:"risk_categories"`
	RiskMatrix          RiskMatrix         `json:"risk_matrix"`
	TopRisks            []RiskFactor       `json:"top_risks"`
	MitigationStrategies []MitigationPlan  `json:"mitigation_strategies"`
	ComplianceStatus    ComplianceStatus   `json:"compliance_status"`
	RiskTrends          []RiskTrend        `json:"risk_trends"`
	ActionPlan          []string           `json:"action_plan"`
	MonitoringPlan      []string           `json:"monitoring_plan"`
}