package reporting

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// VisualizationGenerator handles data visualization preparation
type VisualizationGenerator struct {
	chartCounter int
	colorPalette []string
	themes       map[string]ChartTheme
}

// ChartTheme defines visual styling for charts
type ChartTheme struct {
	PrimaryColor   string   `json:"primary_color"`
	SecondaryColor string   `json:"secondary_color"`
	AccentColors   []string `json:"accent_colors"`
	BackgroundColor string  `json:"background_color"`
	TextColor      string   `json:"text_color"`
	GridColor      string   `json:"grid_color"`
	FontFamily     string   `json:"font_family"`
	FontSize       int      `json:"font_size"`
}

// NewVisualizationGenerator creates a new visualization generator
func NewVisualizationGenerator() *VisualizationGenerator {
	return &VisualizationGenerator{
		chartCounter: 0,
		colorPalette: []string{
			"#3b82f6", "#ef4444", "#22c55e", "#f59e0b", "#8b5cf6",
			"#06b6d4", "#f97316", "#84cc16", "#ec4899", "#6366f1",
		},
		themes: getDefaultThemes(),
	}
}

// GenerateVisualizationData creates visualization data for executive report
func (vg *VisualizationGenerator) GenerateVisualizationData(report *ExecutiveReport) VisualizationData {
	viz := VisualizationData{
		Charts:          []ChartData{},
		Graphs:          []GraphData{},
		Tables:          []TableData{},
		Dashboards:      []DashboardData{},
		InteractiveData: []InteractiveElement{},
	}

	// Generate performance charts
	viz.Charts = append(viz.Charts, vg.createPerformanceCharts(report)...)

	// Generate financial charts
	viz.Charts = append(viz.Charts, vg.createFinancialCharts(report)...)

	// Generate risk charts
	viz.Charts = append(viz.Charts, vg.createRiskCharts(report)...)

	// Generate trend graphs
	viz.Graphs = append(viz.Graphs, vg.createTrendGraphs(report)...)

	// Generate data tables
	viz.Tables = append(viz.Tables, vg.createDataTables(report)...)

	// Generate dashboard components
	viz.Dashboards = append(viz.Dashboards, vg.createDashboards(report)...)

	// Generate interactive elements
	viz.InteractiveData = append(viz.InteractiveData, vg.createInteractiveElements(report)...)

	return viz
}

// Chart creation methods

func (vg *VisualizationGenerator) createPerformanceCharts(report *ExecutiveReport) []ChartData {
	charts := []ChartData{}

	// KPI Overview Chart
	kpiChart := ChartData{
		ChartID:     vg.getNextChartID(),
		ChartType:   ChartTypeBar,
		Title:       "Key Performance Indicators",
		Description: "Current performance across key business metrics",
		XAxis: AxisData{
			Label: "Metrics",
			Type:  "category",
			Categories: []string{
				"Customer Satisfaction",
				"Employee Engagement",
				"Operational Efficiency",
				"Market Share",
			},
		},
		YAxis: AxisData{
			Label: "Percentage",
			Type:  "value",
			Min:   0,
			Max:   100,
		},
		DataSeries: []DataSeries{
			{
				Name: "Current Performance",
				Data: []DataPoint{
					{Y: report.PerformanceSection.KPIMetrics.CustomerSatisfaction, Label: "Customer Satisfaction"},
					{Y: report.PerformanceSection.KPIMetrics.EmployeeEngagement, Label: "Employee Engagement"},
					{Y: report.PerformanceSection.KPIMetrics.OperationalEfficiency, Label: "Operational Efficiency"},
					{Y: report.PerformanceSection.KPIMetrics.MarketShare, Label: "Market Share"},
				},
				Color: vg.getNextColor(),
			},
		},
		Colors:  vg.getChartColors(4),
		Options: vg.getDefaultChartOptions(),
	}
	charts = append(charts, kpiChart)

	// Performance Grade Distribution
	gradeChart := vg.createPerformanceGradeChart(report)
	charts = append(charts, gradeChart)

	return charts
}

func (vg *VisualizationGenerator) createFinancialCharts(report *ExecutiveReport) []ChartData {
	charts := []ChartData{}

	// ROI Analysis Chart
	roiChart := ChartData{
		ChartID:     vg.getNextChartID(),
		ChartType:   ChartTypeLine,
		Title:       "ROI Projection Over Time",
		Description: "Return on investment projection for the next 5 years",
		XAxis: AxisData{
			Label: "Year",
			Type:  "category",
			Categories: []string{"Year 1", "Year 2", "Year 3", "Year 4", "Year 5"},
		},
		YAxis: AxisData{
			Label: "ROI Percentage",
			Type:  "value",
			Min:   0,
		},
		DataSeries: []DataSeries{
			{
				Name:  "ROI Projection",
				Data:  vg.createROIProjectionData(report.ROIAnalysis),
				Color: vg.getNextColor(),
			},
		},
		Colors:  vg.getChartColors(1),
		Options: vg.getDefaultChartOptions(),
	}
	charts = append(charts, roiChart)

	// Investment Breakdown Pie Chart
	investmentChart := vg.createInvestmentBreakdownChart(report)
	charts = append(charts, investmentChart)

	return charts
}

func (vg *VisualizationGenerator) createRiskCharts(report *ExecutiveReport) []ChartData {
	charts := []ChartData{}

	// Risk Matrix Heatmap
	riskMatrix := ChartData{
		ChartID:     vg.getNextChartID(),
		ChartType:   ChartTypeHeatmap,
		Title:       "Risk Assessment Matrix",
		Description: "Risk probability vs impact analysis",
		XAxis: AxisData{
			Label:      "Impact",
			Type:       "category",
			Categories: []string{"Low", "Medium", "High", "Critical"},
		},
		YAxis: AxisData{
			Label:      "Probability",
			Type:       "category",
			Categories: []string{"Low", "Medium", "High", "Very High"},
		},
		DataSeries: []DataSeries{
			{
				Name: "Risk Distribution",
				Data: vg.createRiskMatrixData(report.RiskMitigation.RiskMatrix),
				Color: "#ef4444",
			},
		},
		Colors:  []string{"#22c55e", "#f59e0b", "#ef4444", "#dc2626"},
		Options: vg.getRiskMatrixOptions(),
	}
	charts = append(charts, riskMatrix)

	// Top Risks Chart
	topRisksChart := vg.createTopRisksChart(report)
	charts = append(charts, topRisksChart)

	return charts
}

func (vg *VisualizationGenerator) createTrendGraphs(report *ExecutiveReport) []GraphData {
	graphs := []GraphData{}

	// Performance Trend Graph
	performanceTrend := GraphData{
		GraphID:     vg.getNextGraphID(),
		GraphType:   "line",
		Title:       "Performance Trends",
		Description: "Historical performance trends across key metrics",
		XAxis:       "Time Period",
		YAxis:       "Performance Score",
		DataSeries: []GraphSeries{
			{
				Name:   "Overall Performance",
				Points: vg.createPerformanceTrendData(report),
				Color:  "#3b82f6",
			},
		},
		Options: map[string]interface{}{
			"showPoints":    true,
			"smoothCurves":  true,
			"showGridLines": true,
			"interactive":   true,
		},
	}
	graphs = append(graphs, performanceTrend)

	// Cost Savings Graph
	costSavingsGraph := vg.createCostSavingsGraph(report)
	graphs = append(graphs, costSavingsGraph)

	return graphs
}

func (vg *VisualizationGenerator) createDataTables(report *ExecutiveReport) []TableData {
	tables := []TableData{}

	// Executive Summary Table
	summaryTable := TableData{
		TableID:     vg.getNextTableID(),
		Title:       "Executive Summary Metrics",
		Description: "Key metrics and findings summary",
		Headers:     []string{"Metric", "Current", "Target", "Status", "Trend"},
		Rows:        vg.createSummaryTableRows(report),
		Styling: TableStyling{
			HeaderColor:    "#f3f4f6",
			AlternateRows:  true,
			BorderStyle:    "solid",
			ResponsiveMode: true,
		},
		Options: map[string]interface{}{
			"sortable":   true,
			"filterable": false,
			"exportable": true,
		},
	}
	tables = append(tables, summaryTable)

	// Risk Assessment Table
	riskTable := vg.createRiskAssessmentTable(report)
	tables = append(tables, riskTable)

	// Cost Optimization Table
	costTable := vg.createCostOptimizationTable(report)
	tables = append(tables, costTable)

	return tables
}

func (vg *VisualizationGenerator) createDashboards(report *ExecutiveReport) []DashboardData {
	dashboards := []DashboardData{}

	// Executive Dashboard
	execDashboard := DashboardData{
		DashboardID: "executive_dashboard",
		Title:       "Executive Performance Dashboard",
		Description: "High-level view of business performance",
		Widgets: []DashboardWidget{
			{
				WidgetID:   "overall_score",
				Type:       "gauge",
				Title:      "Overall Score",
				Value:      report.ExecutiveSummary.OverallScore.Score,
				Target:     100,
				Unit:       "points",
				Position:   WidgetPosition{Row: 1, Column: 1, Width: 2, Height: 2},
				Styling:    vg.getGaugeWidgetStyling(),
			},
			{
				WidgetID:   "kpi_summary",
				Type:       "metrics",
				Title:      "Key Metrics",
				Data:       vg.createKPIWidgetData(report),
				Position:   WidgetPosition{Row: 1, Column: 3, Width: 4, Height: 2},
				Styling:    vg.getMetricsWidgetStyling(),
			},
			{
				WidgetID:   "trend_chart",
				Type:       "chart",
				Title:      "Performance Trend",
				ChartData:  vg.createTrendWidgetData(report),
				Position:   WidgetPosition{Row: 2, Column: 1, Width: 6, Height: 3},
				Styling:    vg.getChartWidgetStyling(),
			},
		},
		Layout: DashboardLayout{
			Columns:    6,
			RowHeight:  80,
			Responsive: true,
			Theme:      "executive",
		},
		Options: map[string]interface{}{
			"autoRefresh":  false,
			"exportable":   true,
			"interactive":  true,
			"fullscreen":   true,
		},
	}
	dashboards = append(dashboards, execDashboard)

	return dashboards
}

func (vg *VisualizationGenerator) createInteractiveElements(report *ExecutiveReport) []InteractiveElement {
	elements := []InteractiveElement{}

	// Risk Heat Map
	riskHeatMap := InteractiveElement{
		ElementID:   "risk_heatmap",
		Type:        "heatmap",
		Title:       "Interactive Risk Heat Map",
		Description: "Click on risk areas to view details",
		Data:        vg.createRiskHeatMapData(report),
		Interactions: []Interaction{
			{
				Type:        "click",
				Target:      "cell",
				Action:      "showTooltip",
				Parameters:  map[string]interface{}{"position": "mouse"},
			},
			{
				Type:        "hover",
				Target:      "cell",
				Action:      "highlight",
				Parameters:  map[string]interface{}{"intensity": 0.3},
			},
		},
		Styling: map[string]interface{}{
			"colorScheme": "risk",
			"animation":   "fade",
			"responsive":  true,
		},
	}
	elements = append(elements, riskHeatMap)

	// Performance Gauge
	performanceGauge := vg.createPerformanceGauge(report)
	elements = append(elements, performanceGauge)

	return elements
}

// Helper methods for specific chart types

func (vg *VisualizationGenerator) createPerformanceGradeChart(report *ExecutiveReport) ChartData {
	return ChartData{
		ChartID:     vg.getNextChartID(),
		ChartType:   ChartTypeGauge,
		Title:       "Overall Performance Grade",
		Description: "Current performance grade and score",
		DataSeries: []DataSeries{
			{
				Name: "Performance Score",
				Data: []DataPoint{
					{Value: report.ExecutiveSummary.OverallScore.Score, Label: string(report.ExecutiveSummary.OverallScore.Grade)},
				},
				Color: vg.getGradeColor(report.ExecutiveSummary.OverallScore.Grade),
			},
		},
		Options: map[string]interface{}{
			"min":         0,
			"max":         100,
			"thresholds":  []float64{50, 70, 85, 95},
			"colors":      []string{"#dc2626", "#f59e0b", "#22c55e", "#3b82f6"},
			"showValue":   true,
			"showGrade":   true,
		},
	}
}

func (vg *VisualizationGenerator) createInvestmentBreakdownChart(report *ExecutiveReport) ChartData {
	return ChartData{
		ChartID:     vg.getNextChartID(),
		ChartType:   ChartTypePie,
		Title:       "Investment Allocation",
		Description: "Breakdown of investment by category",
		DataSeries: []DataSeries{
			{
				Name: "Investment Distribution",
				Data: vg.createInvestmentData(report.ROIAnalysis),
				Color: "",
			},
		},
		Colors: vg.getChartColors(5),
		Options: map[string]interface{}{
			"showLabels":     true,
			"showLegend":     true,
			"showTooltips":   true,
			"innerRadius":    30,
			"outerRadius":    100,
		},
	}
}

func (vg *VisualizationGenerator) createTopRisksChart(report *ExecutiveReport) ChartData {
	return ChartData{
		ChartID:     vg.getNextChartID(),
		ChartType:   ChartTypeBar,
		Title:       "Top Risk Factors",
		Description: "Highest priority risks by impact score",
		XAxis: AxisData{
			Label: "Risk Factor",
			Type:  "category",
		},
		YAxis: AxisData{
			Label: "Risk Score",
			Type:  "value",
			Min:   0,
			Max:   100,
		},
		DataSeries: []DataSeries{
			{
				Name: "Risk Score",
				Data: vg.createTopRisksData(report.RiskMitigation.RiskAnalysis.TopRisks),
				Color: "#ef4444",
			},
		},
		Colors: []string{"#dc2626", "#ef4444", "#f87171"},
		Options: map[string]interface{}{
			"horizontal": true,
			"showValues": true,
		},
	}
}

// Data creation helper methods

func (vg *VisualizationGenerator) createROIProjectionData(roi ROIAnalysis) []DataPoint {
	data := []DataPoint{}
	cumulativeROI := 0.0

	for i, benefit := range roi.ROICalculation.AnnualBenefits {
		if i < 5 { // Limit to 5 years
			cumulativeROI += (benefit / roi.ROICalculation.InitialInvestment) * 100
			data = append(data, DataPoint{
				Value: cumulativeROI,
				Label: fmt.Sprintf("Year %d", i+1),
			})
		}
	}

	return data
}

func (vg *VisualizationGenerator) createRiskMatrixData(matrix RiskMatrix) []DataPoint {
	data := []DataPoint{}

	for i, row := range matrix.Matrix {
		for j, cell := range row {
			data = append(data, DataPoint{
				X:     float64(j),
				Y:     float64(i),
				Value: float64(cell.RiskCount),
				Label: cell.RiskLevel,
				Color: vg.getRiskLevelColor(cell.RiskLevel),
			})
		}
	}

	return data
}

func (vg *VisualizationGenerator) createPerformanceTrendData(report *ExecutiveReport) []DataPoint {
	points := []DataPoint{}

	// Create sample trend data (in real implementation, this would come from historical data)
	for i := 0; i < 12; i++ {
		points = append(points, DataPoint{
			X:     float64(i),
			Y:     report.ExecutiveSummary.OverallScore.Score + (math.Sin(float64(i)*0.5) * 10),
			Label: fmt.Sprintf("Month %d", i+1),
		})
	}

	return points
}

func (vg *VisualizationGenerator) createInvestmentData(roi ROIAnalysis) []DataPoint {
	return []DataPoint{
		{Value: 40, Label: "Technology", Color: "#3b82f6"},
		{Value: 25, Label: "Operations", Color: "#22c55e"},
		{Value: 20, Label: "Marketing", Color: "#f59e0b"},
		{Value: 10, Label: "Infrastructure", Color: "#8b5cf6"},
		{Value: 5, Label: "Other", Color: "#6b7280"},
	}
}

func (vg *VisualizationGenerator) createTopRisksData(risks []RiskFactor) []DataPoint {
	data := []DataPoint{}

	// Sort risks by score (descending) and take top 5
	sort.Slice(risks, func(i, j int) bool {
		return risks[i].RiskScore > risks[j].RiskScore
	})

	maxRisks := 5
	if len(risks) < maxRisks {
		maxRisks = len(risks)
	}

	for i := 0; i < maxRisks; i++ {
		data = append(data, DataPoint{
			Value: risks[i].RiskScore,
			Label: risks[i].Name,
			Color: vg.getRiskLevelColor(risks[i].Priority),
		})
	}

	return data
}

// Widget and dashboard helper methods

func (vg *VisualizationGenerator) createKPIWidgetData(report *ExecutiveReport) map[string]interface{} {
	return map[string]interface{}{
		"metrics": []map[string]interface{}{
			{
				"name":  "Customer Satisfaction",
				"value": report.PerformanceSection.KPIMetrics.CustomerSatisfaction,
				"unit":  "%",
				"trend": "up",
			},
			{
				"name":  "Employee Engagement",
				"value": report.PerformanceSection.KPIMetrics.EmployeeEngagement,
				"unit":  "%",
				"trend": "stable",
			},
			{
				"name":  "Operational Efficiency",
				"value": report.PerformanceSection.KPIMetrics.OperationalEfficiency,
				"unit":  "%",
				"trend": "up",
			},
			{
				"name":  "Market Share",
				"value": report.PerformanceSection.KPIMetrics.MarketShare,
				"unit":  "%",
				"trend": "down",
			},
		},
	}
}

// Utility methods

func (vg *VisualizationGenerator) getNextChartID() string {
	vg.chartCounter++
	return fmt.Sprintf("chart_%d", vg.chartCounter)
}

func (vg *VisualizationGenerator) getNextGraphID() string {
	return fmt.Sprintf("graph_%d", time.Now().UnixNano())
}

func (vg *VisualizationGenerator) getNextTableID() string {
	return fmt.Sprintf("table_%d", time.Now().UnixNano())
}

func (vg *VisualizationGenerator) getNextColor() string {
	return vg.colorPalette[vg.chartCounter%len(vg.colorPalette)]
}

func (vg *VisualizationGenerator) getChartColors(count int) []string {
	colors := []string{}
	for i := 0; i < count; i++ {
		colors = append(colors, vg.colorPalette[i%len(vg.colorPalette)])
	}
	return colors
}

func (vg *VisualizationGenerator) getGradeColor(grade Grade) string {
	switch grade {
	case GradeExcellent:
		return "#22c55e"
	case GradeGood:
		return "#84cc16"
	case GradeSatisfactory:
		return "#eab308"
	case GradeNeedsImprovement:
		return "#f97316"
	case GradePoor:
		return "#ef4444"
	case GradeCritical:
		return "#dc2626"
	default:
		return "#6b7280"
	}
}

func (vg *VisualizationGenerator) getRiskLevelColor(level string) string {
	switch level {
	case "low":
		return "#22c55e"
	case "medium":
		return "#f59e0b"
	case "high":
		return "#ef4444"
	case "critical":
		return "#dc2626"
	default:
		return "#6b7280"
	}
}

func (vg *VisualizationGenerator) getDefaultChartOptions() map[string]interface{} {
	return map[string]interface{}{
		"responsive":     true,
		"maintainAspectRatio": false,
		"animation":      true,
		"legend":         map[string]bool{"display": true},
		"tooltip":        map[string]bool{"enabled": true},
		"grid":           map[string]bool{"display": true},
	}
}

func (vg *VisualizationGenerator) getRiskMatrixOptions() map[string]interface{} {
	return map[string]interface{}{
		"responsive":     true,
		"showLabels":     true,
		"cellSize":       40,
		"colorScale":     "RdYlGn_r",
		"showTooltips":   true,
		"interactive":    true,
	}
}

func getDefaultThemes() map[string]ChartTheme {
	return map[string]ChartTheme{
		"executive": {
			PrimaryColor:    "#1f2937",
			SecondaryColor:  "#374151",
			AccentColors:    []string{"#3b82f6", "#ef4444", "#22c55e", "#f59e0b"},
			BackgroundColor: "#ffffff",
			TextColor:       "#1f2937",
			GridColor:       "#e5e7eb",
			FontFamily:      "Inter, sans-serif",
			FontSize:        12,
		},
		"financial": {
			PrimaryColor:    "#059669",
			SecondaryColor:  "#047857",
			AccentColors:    []string{"#10b981", "#ef4444", "#3b82f6", "#f59e0b"},
			BackgroundColor: "#ffffff",
			TextColor:       "#1f2937",
			GridColor:       "#d1fae5",
			FontFamily:      "Inter, sans-serif",
			FontSize:        12,
		},
	}
}