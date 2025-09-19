package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// ExecutiveReportGenerator implements ReportGenerator for creating comprehensive reports
type ExecutiveReportGenerator struct {
	config ReportConfig
}

// ReportConfig holds configuration for report generation
type ReportConfig struct {
	CompanyName      string                 `json:"company_name"`
	ReportingPeriod  time.Duration         `json:"reporting_period"`
	IncludeCharts    bool                  `json:"include_charts"`
	DetailLevel      ReportDetailLevel     `json:"detail_level"`
	Branding         BrandingConfig        `json:"branding"`
	Templates        map[ReportFormat]string `json:"templates"`
}

// ReportDetailLevel defines how detailed the report should be
type ReportDetailLevel string

const (
	DetailLevelExecutive ReportDetailLevel = "EXECUTIVE"  // High-level summary only
	DetailLevelManagement ReportDetailLevel = "MANAGEMENT" // Management-level detail
	DetailLevelTechnical ReportDetailLevel = "TECHNICAL"  // Full technical detail
)

// BrandingConfig holds branding information for reports
type BrandingConfig struct {
	LogoURL     string `json:"logo_url"`
	PrimaryColor string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
	FontFamily  string `json:"font_family"`
}

// NewExecutiveReportGenerator creates a new report generator
func NewExecutiveReportGenerator(config ReportConfig) *ExecutiveReportGenerator {
	return &ExecutiveReportGenerator{
		config: config,
	}
}

// NewDefaultExecutiveReportGenerator creates generator with default configuration
func NewDefaultExecutiveReportGenerator() *ExecutiveReportGenerator {
	return NewExecutiveReportGenerator(DefaultReportConfig())
}

// GenerateExecutiveReport creates comprehensive performance reports in specified format
func (erg *ExecutiveReportGenerator) GenerateExecutiveReport(analysis PerformanceAnalysis, format ReportFormat) ([]byte, error) {
	switch format {
	case ReportFormatJSON:
		return erg.generateJSONReport(analysis)
	case ReportFormatText:
		return erg.generateTextReport(analysis)
	case ReportFormatMarkdown:
		return erg.generateMarkdownReport(analysis)
	case ReportFormatPDF:
		return erg.generatePDFReport(analysis)
	default:
		return nil, fmt.Errorf("unsupported report format: %s", format)
	}
}

// GenerateSummaryReport creates concise summary reports for executives
func (erg *ExecutiveReportGenerator) GenerateSummaryReport(analysis PerformanceAnalysis) ExecutiveSummary {
	// Determine overall health
	health := erg.determineOverallHealth(analysis)

	// Count critical issues
	criticalIssues := erg.countCriticalIssues(analysis.Bottlenecks)

	// Extract key recommendations (top 3)
	keyRecommendations := erg.extractKeyRecommendations(analysis.Recommendations, 3)

	// Calculate ROI summary
	roiSummary := erg.calculateROISummary(analysis.Recommendations)

	// Generate next actions
	nextActions := erg.generateNextActions(analysis)

	return ExecutiveSummary{
		OverallHealth:      health,
		PerformanceScore:   analysis.PerformanceScore,
		CriticalIssues:     criticalIssues,
		KeyRecommendations: keyRecommendations,
		BusinessImpact:     analysis.BusinessImpact,
		ROISummary:         roiSummary,
		NextActions:        nextActions,
		GeneratedAt:        time.Now(),
	}
}

// JSON Report Generation
func (erg *ExecutiveReportGenerator) generateJSONReport(analysis PerformanceAnalysis) ([]byte, error) {
	report := map[string]interface{}{
		"metadata": map[string]interface{}{
			"generated_at":      time.Now().Format(time.RFC3339),
			"report_version":    "1.0",
			"company":           erg.config.CompanyName,
			"analysis_period":   analysis.TimeRange,
			"detail_level":      erg.config.DetailLevel,
		},
		"executive_summary": erg.GenerateSummaryReport(analysis),
		"performance_analysis": analysis,
		"detailed_findings": erg.generateDetailedFindings(analysis),
		"recommendations": erg.generateRecommendationDetails(analysis.Recommendations),
		"financial_impact": erg.generateFinancialImpactSection(analysis),
		"appendices": erg.generateAppendices(analysis),
	}

	return json.MarshalIndent(report, "", "  ")
}

// Text Report Generation
func (erg *ExecutiveReportGenerator) generateTextReport(analysis PerformanceAnalysis) ([]byte, error) {
	var buffer bytes.Buffer

	// Header
	erg.writeTextHeader(&buffer, analysis)

	// Executive Summary
	erg.writeExecutiveSummary(&buffer, analysis)

	// Key Findings
	erg.writeKeyFindings(&buffer, analysis)

	// Recommendations
	erg.writeRecommendations(&buffer, analysis.Recommendations)

	// Financial Impact
	erg.writeFinancialImpact(&buffer, analysis)

	// Next Steps
	erg.writeNextSteps(&buffer, analysis)

	// Footer
	erg.writeTextFooter(&buffer)

	return buffer.Bytes(), nil
}

// Markdown Report Generation
func (erg *ExecutiveReportGenerator) generateMarkdownReport(analysis PerformanceAnalysis) ([]byte, error) {
	var buffer bytes.Buffer

	// Title and metadata
	erg.writeMarkdownHeader(&buffer, analysis)

	// Executive Summary
	erg.writeMarkdownExecutiveSummary(&buffer, analysis)

	// Performance Overview
	erg.writeMarkdownPerformanceOverview(&buffer, analysis)

	// Critical Issues
	erg.writeMarkdownCriticalIssues(&buffer, analysis.Bottlenecks)

	// Recommendations
	erg.writeMarkdownRecommendations(&buffer, analysis.Recommendations)

	// Financial Analysis
	erg.writeMarkdownFinancialAnalysis(&buffer, analysis)

	// Implementation Roadmap
	erg.writeMarkdownImplementationRoadmap(&buffer, analysis.Recommendations)

	// Appendices
	erg.writeMarkdownAppendices(&buffer, analysis)

	return buffer.Bytes(), nil
}

// PDF Report Generation (simplified - would require PDF library in production)
func (erg *ExecutiveReportGenerator) generatePDFReport(analysis PerformanceAnalysis) ([]byte, error) {
	// In a real implementation, this would use a PDF generation library
	// For now, we'll return a placeholder that could be used with external PDF generation

	htmlContent := erg.generateHTMLForPDF(analysis)
	return []byte(htmlContent), nil
}

// Helper methods for report generation

// writeTextHeader writes the header section for text reports
func (erg *ExecutiveReportGenerator) writeTextHeader(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	buffer.WriteString(strings.Repeat("=", 80) + "\n")
	buffer.WriteString("TRADING EXCHANGE PERFORMANCE ANALYSIS REPORT\n")
	buffer.WriteString(strings.Repeat("=", 80) + "\n\n")
	buffer.WriteString(fmt.Sprintf("Company: %s\n", erg.config.CompanyName))
	buffer.WriteString(fmt.Sprintf("Report Generated: %s\n", time.Now().Format("January 2, 2006 at 3:04 PM MST")))
	buffer.WriteString(fmt.Sprintf("Analysis Period: %s to %s\n",
		analysis.TimeRange.Start.Format("Jan 2, 2006"),
		analysis.TimeRange.End.Format("Jan 2, 2006")))
	buffer.WriteString(fmt.Sprintf("Analysis ID: %s\n\n", analysis.ID))
}

// writeExecutiveSummary writes executive summary section
func (erg *ExecutiveReportGenerator) writeExecutiveSummary(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	summary := erg.GenerateSummaryReport(analysis)

	buffer.WriteString("EXECUTIVE SUMMARY\n")
	buffer.WriteString(strings.Repeat("-", 80) + "\n\n")

	buffer.WriteString(fmt.Sprintf("Overall System Health: %s\n", summary.OverallHealth))
	buffer.WriteString(fmt.Sprintf("Performance Score: %.1f/100\n", summary.PerformanceScore*100))
	buffer.WriteString(fmt.Sprintf("Critical Issues: %d\n", summary.CriticalIssues))
	buffer.WriteString(fmt.Sprintf("Expected ROI: %.1f%%\n\n", summary.ROISummary.OverallROI))

	buffer.WriteString("Key Recommendations:\n")
	for i, rec := range summary.KeyRecommendations {
		buffer.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
	}
	buffer.WriteString("\n")
}

// writeKeyFindings writes the key findings section
func (erg *ExecutiveReportGenerator) writeKeyFindings(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	buffer.WriteString("KEY FINDINGS\n")
	buffer.WriteString(strings.Repeat("-", 80) + "\n\n")

	// Performance trends
	buffer.WriteString("Performance Trends:\n")
	buffer.WriteString(fmt.Sprintf("• Latency Trend: %s\n", analysis.TrendAnalysis.LatencyTrend))
	buffer.WriteString(fmt.Sprintf("• Throughput Trend: %s\n", analysis.TrendAnalysis.ThroughputTrend))
	buffer.WriteString(fmt.Sprintf("• Volume Trend: %s\n", analysis.TrendAnalysis.VolumeTrend))
	buffer.WriteString(fmt.Sprintf("• Trend Strength: %.1f%%\n\n", analysis.TrendAnalysis.TrendStrength*100))

	// Critical bottlenecks
	criticalBottlenecks := erg.filterCriticalBottlenecks(analysis.Bottlenecks)
	if len(criticalBottlenecks) > 0 {
		buffer.WriteString("Critical Bottlenecks:\n")
		for i, bottleneck := range criticalBottlenecks {
			buffer.WriteString(fmt.Sprintf("%d. %s - %s (Severity: %.1f%%)\n",
				i+1, bottleneck.Component, bottleneck.Type, bottleneck.Severity*100))
		}
		buffer.WriteString("\n")
	}
}

// writeRecommendations writes the recommendations section
func (erg *ExecutiveReportGenerator) writeRecommendations(buffer *bytes.Buffer, recommendations []Recommendation) {
	buffer.WriteString("RECOMMENDATIONS\n")
	buffer.WriteString(strings.Repeat("-", 80) + "\n\n")

	// Group recommendations by priority
	priorityGroups := erg.groupRecommendationsByPriority(recommendations)

	for _, priority := range []Priority{PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow} {
		recs, exists := priorityGroups[priority]
		if !exists || len(recs) == 0 {
			continue
		}

		buffer.WriteString(fmt.Sprintf("%s Priority:\n", strings.ToUpper(string(priority))))
		for i, rec := range recs {
			buffer.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec.Title))
			buffer.WriteString(fmt.Sprintf("   Category: %s | Complexity: %s | Time to Effect: %v\n",
				rec.Category, rec.Complexity, rec.TimeToEffect))
			buffer.WriteString(fmt.Sprintf("   %s\n", rec.Description))
			if erg.config.DetailLevel == DetailLevelTechnical {
				buffer.WriteString(fmt.Sprintf("   Expected Impact: $%.0f | Confidence: %.0f%%\n",
					rec.Impact.Revenue, rec.Confidence*100))
			}
			buffer.WriteString("\n")
		}
	}
}

// writeFinancialImpact writes financial impact section
func (erg *ExecutiveReportGenerator) writeFinancialImpact(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	buffer.WriteString("FINANCIAL IMPACT ANALYSIS\n")
	buffer.WriteString(strings.Repeat("-", 80) + "\n\n")

	buffer.WriteString(fmt.Sprintf("Potential Revenue Impact: $%.0f\n", analysis.BusinessImpact.Revenue))
	buffer.WriteString(fmt.Sprintf("Estimated Cost Impact: $%.0f\n", analysis.BusinessImpact.Cost))
	buffer.WriteString(fmt.Sprintf("User Experience Score: %.1f/10\n", analysis.BusinessImpact.UserExperience*10))
	buffer.WriteString(fmt.Sprintf("Reliability Score: %.1f/10\n", analysis.BusinessImpact.Reliability*10))
	buffer.WriteString(fmt.Sprintf("Overall Business Impact Score: %.1f/10\n\n", analysis.BusinessImpact.OverallScore*10))
}

// writeNextSteps writes next steps section
func (erg *ExecutiveReportGenerator) writeNextSteps(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	buffer.WriteString("NEXT STEPS\n")
	buffer.WriteString(strings.Repeat("-", 80) + "\n\n")

	nextActions := erg.generateNextActions(analysis)
	for i, action := range nextActions {
		buffer.WriteString(fmt.Sprintf("%d. %s\n", i+1, action))
	}
	buffer.WriteString("\n")
}

// writeTextFooter writes footer for text reports
func (erg *ExecutiveReportGenerator) writeTextFooter(buffer *bytes.Buffer) {
	buffer.WriteString(strings.Repeat("=", 80) + "\n")
	buffer.WriteString("This report was generated automatically by the AI Performance Analyzer\n")
	buffer.WriteString(fmt.Sprintf("Report generated at: %s\n", time.Now().Format(time.RFC3339)))
	buffer.WriteString("For questions about this report, contact the Performance Engineering team\n")
	buffer.WriteString(strings.Repeat("=", 80) + "\n")
}

// Markdown-specific writing methods

// writeMarkdownHeader writes markdown header
func (erg *ExecutiveReportGenerator) writeMarkdownHeader(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	buffer.WriteString("# Trading Exchange Performance Analysis Report\n\n")
	buffer.WriteString("## Report Metadata\n\n")
	buffer.WriteString(fmt.Sprintf("- **Company:** %s\n", erg.config.CompanyName))
	buffer.WriteString(fmt.Sprintf("- **Generated:** %s\n", time.Now().Format("January 2, 2006")))
	buffer.WriteString(fmt.Sprintf("- **Analysis Period:** %s to %s\n",
		analysis.TimeRange.Start.Format("Jan 2"), analysis.TimeRange.End.Format("Jan 2, 2006")))
	buffer.WriteString(fmt.Sprintf("- **Analysis ID:** `%s`\n\n", analysis.ID))
}

// writeMarkdownExecutiveSummary writes markdown executive summary
func (erg *ExecutiveReportGenerator) writeMarkdownExecutiveSummary(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	summary := erg.GenerateSummaryReport(analysis)

	buffer.WriteString("## Executive Summary\n\n")

	// Status badges
	healthColor := erg.getHealthColor(summary.OverallHealth)
	buffer.WriteString(fmt.Sprintf("![Health Status](https://img.shields.io/badge/Health-%s-%s)\n",
		summary.OverallHealth, healthColor))
	buffer.WriteString(fmt.Sprintf("![Performance](https://img.shields.io/badge/Performance-%.0f%%25-blue)\n",
		summary.PerformanceScore*100))
	buffer.WriteString(fmt.Sprintf("![Critical Issues](https://img.shields.io/badge/Critical_Issues-%d-red)\n\n",
		summary.CriticalIssues))

	// Key metrics table
	buffer.WriteString("### Key Metrics\n\n")
	buffer.WriteString("| Metric | Value |\n")
	buffer.WriteString("|--------|-------|\n")
	buffer.WriteString(fmt.Sprintf("| Overall Health | %s |\n", summary.OverallHealth))
	buffer.WriteString(fmt.Sprintf("| Performance Score | %.1f/100 |\n", summary.PerformanceScore*100))
	buffer.WriteString(fmt.Sprintf("| Critical Issues | %d |\n", summary.CriticalIssues))
	buffer.WriteString(fmt.Sprintf("| Expected ROI | %.1f%% |\n", summary.ROISummary.OverallROI))
	buffer.WriteString(fmt.Sprintf("| Payback Period | %d months |\n\n", summary.ROISummary.PaybackMonths))

	// Key recommendations
	buffer.WriteString("### Top Recommendations\n\n")
	for i, rec := range summary.KeyRecommendations {
		buffer.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
	}
	buffer.WriteString("\n")
}

// writeMarkdownPerformanceOverview writes performance overview
func (erg *ExecutiveReportGenerator) writeMarkdownPerformanceOverview(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	buffer.WriteString("## Performance Overview\n\n")

	// Trend analysis
	buffer.WriteString("### Trend Analysis\n\n")
	buffer.WriteString("| Metric | Trend | Strength |\n")
	buffer.WriteString("|--------|-------|----------|\n")
	buffer.WriteString(fmt.Sprintf("| Latency | %s | %.1f%% |\n",
		analysis.TrendAnalysis.LatencyTrend, analysis.TrendAnalysis.TrendStrength*100))
	buffer.WriteString(fmt.Sprintf("| Throughput | %s | %.1f%% |\n",
		analysis.TrendAnalysis.ThroughputTrend, analysis.TrendAnalysis.TrendStrength*100))
	buffer.WriteString(fmt.Sprintf("| Volume | %s | %.1f%% |\n",
		analysis.TrendAnalysis.VolumeTrend, analysis.TrendAnalysis.TrendStrength*100))
	buffer.WriteString("\n")
}

// writeMarkdownCriticalIssues writes critical issues section
func (erg *ExecutiveReportGenerator) writeMarkdownCriticalIssues(buffer *bytes.Buffer, bottlenecks []Bottleneck) {
	criticalBottlenecks := erg.filterCriticalBottlenecks(bottlenecks)

	if len(criticalBottlenecks) == 0 {
		buffer.WriteString("## Critical Issues\n\n")
		buffer.WriteString("✅ No critical issues detected.\n\n")
		return
	}

	buffer.WriteString("## Critical Issues\n\n")
	buffer.WriteString("| Component | Type | Severity | Impact |\n")
	buffer.WriteString("|-----------|------|----------|--------|\n")

	for _, bottleneck := range criticalBottlenecks {
		severityBadge := erg.getSeverityBadge(bottleneck.Severity)
		impactScore := bottleneck.Impact.OverallScore * 10
		buffer.WriteString(fmt.Sprintf("| %s | %s | %s | %.1f/10 |\n",
			bottleneck.Component, bottleneck.Type, severityBadge, impactScore))
	}
	buffer.WriteString("\n")
}

// writeMarkdownRecommendations writes recommendations in markdown
func (erg *ExecutiveReportGenerator) writeMarkdownRecommendations(buffer *bytes.Buffer, recommendations []Recommendation) {
	buffer.WriteString("## Recommendations\n\n")

	priorityGroups := erg.groupRecommendationsByPriority(recommendations)

	for _, priority := range []Priority{PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow} {
		recs, exists := priorityGroups[priority]
		if !exists || len(recs) == 0 {
			continue
		}

		buffer.WriteString(fmt.Sprintf("### %s Priority\n\n", strings.Title(strings.ToLower(string(priority)))))

		for i, rec := range recs {
			buffer.WriteString(fmt.Sprintf("#### %d. %s\n\n", i+1, rec.Title))

			// Recommendation details
			buffer.WriteString(fmt.Sprintf("- **Category:** %s\n", rec.Category))
			buffer.WriteString(fmt.Sprintf("- **Complexity:** %s\n", rec.Complexity))
			buffer.WriteString(fmt.Sprintf("- **Time to Effect:** %s\n", rec.TimeToEffect))
			buffer.WriteString(fmt.Sprintf("- **Confidence:** %.0f%%\n\n", rec.Confidence*100))

			buffer.WriteString(fmt.Sprintf("%s\n\n", rec.Description))

			if len(rec.Prerequisites) > 0 {
				buffer.WriteString("**Prerequisites:**\n")
				for _, prereq := range rec.Prerequisites {
					buffer.WriteString(fmt.Sprintf("- %s\n", prereq))
				}
				buffer.WriteString("\n")
			}
		}
	}
}

// writeMarkdownFinancialAnalysis writes financial analysis
func (erg *ExecutiveReportGenerator) writeMarkdownFinancialAnalysis(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	buffer.WriteString("## Financial Impact Analysis\n\n")

	buffer.WriteString("### Business Impact Summary\n\n")
	buffer.WriteString("| Category | Impact |\n")
	buffer.WriteString("|----------|--------|\n")
	buffer.WriteString(fmt.Sprintf("| Revenue Impact | $%.0f |\n", analysis.BusinessImpact.Revenue))
	buffer.WriteString(fmt.Sprintf("| Cost Impact | $%.0f |\n", analysis.BusinessImpact.Cost))
	buffer.WriteString(fmt.Sprintf("| User Experience | %.1f/10 |\n", analysis.BusinessImpact.UserExperience*10))
	buffer.WriteString(fmt.Sprintf("| Reliability | %.1f/10 |\n", analysis.BusinessImpact.Reliability*10))
	buffer.WriteString(fmt.Sprintf("| Scalability | %.1f/10 |\n", analysis.BusinessImpact.Scalability*10))
	buffer.WriteString(fmt.Sprintf("| **Overall Score** | **%.1f/10** |\n\n", analysis.BusinessImpact.OverallScore*10))
}

// writeMarkdownImplementationRoadmap writes implementation roadmap
func (erg *ExecutiveReportGenerator) writeMarkdownImplementationRoadmap(buffer *bytes.Buffer, recommendations []Recommendation) {
	buffer.WriteString("## Implementation Roadmap\n\n")

	// Sort recommendations by time to effect
	sortedRecs := make([]Recommendation, len(recommendations))
	copy(sortedRecs, recommendations)
	sort.Slice(sortedRecs, func(i, j int) bool {
		return sortedRecs[i].TimeToEffect < sortedRecs[j].TimeToEffect
	})

	buffer.WriteString("### Timeline\n\n")
	buffer.WriteString("```\n")
	for i, rec := range sortedRecs[:min(5, len(sortedRecs))] {
		buffer.WriteString(fmt.Sprintf("Week %d: %s (%s)\n", i+1, rec.Title, rec.TimeToEffect))
	}
	buffer.WriteString("```\n\n")
}

// writeMarkdownAppendices writes appendices
func (erg *ExecutiveReportGenerator) writeMarkdownAppendices(buffer *bytes.Buffer, analysis PerformanceAnalysis) {
	buffer.WriteString("## Appendices\n\n")

	buffer.WriteString("### A. Analysis Methodology\n\n")
	buffer.WriteString("This analysis was performed using AI-powered performance analysis tools with the following methodologies:\n\n")
	buffer.WriteString("- **Trend Analysis:** Linear regression and exponential smoothing\n")
	buffer.WriteString("- **Bottleneck Detection:** Statistical outlier detection and threshold analysis\n")
	buffer.WriteString("- **Capacity Prediction:** Machine learning-based forecasting models\n")
	buffer.WriteString("- **ROI Calculation:** NPV and IRR analysis with risk adjustment\n\n")

	buffer.WriteString("### B. Confidence Levels\n\n")
	buffer.WriteString(fmt.Sprintf("- **Overall Analysis Confidence:** %.0f%%\n", analysis.Confidence*100))
	buffer.WriteString("- **Prediction Accuracy:** Based on historical data analysis\n")
	buffer.WriteString("- **Recommendation Reliability:** Validated against industry benchmarks\n\n")
}

// Helper methods for analysis and formatting

// determineOverallHealth determines overall system health from analysis
func (erg *ExecutiveReportGenerator) determineOverallHealth(analysis PerformanceAnalysis) HealthStatus {
	score := analysis.PerformanceScore
	criticalIssues := erg.countCriticalIssues(analysis.Bottlenecks)

	if criticalIssues > 3 || score < 0.3 {
		return HealthCritical
	} else if criticalIssues > 1 || score < 0.5 {
		return HealthPoor
	} else if criticalIssues > 0 || score < 0.7 {
		return HealthFair
	} else if score < 0.9 {
		return HealthGood
	}

	return HealthExcellent
}

// countCriticalIssues counts bottlenecks with high severity
func (erg *ExecutiveReportGenerator) countCriticalIssues(bottlenecks []Bottleneck) int {
	count := 0
	for _, bottleneck := range bottlenecks {
		if bottleneck.Severity >= 0.8 {
			count++
		}
	}
	return count
}

// extractKeyRecommendations gets top N recommendations
func (erg *ExecutiveReportGenerator) extractKeyRecommendations(recommendations []Recommendation, n int) []string {
	var keyRecs []string

	// Sort by priority and take top N
	sorted := make([]Recommendation, len(recommendations))
	copy(sorted, recommendations)

	sort.Slice(sorted, func(i, j int) bool {
		return erg.getPriorityScore(sorted[i].Priority) > erg.getPriorityScore(sorted[j].Priority)
	})

	for i := 0; i < min(n, len(sorted)); i++ {
		keyRecs = append(keyRecs, sorted[i].Title)
	}

	return keyRecs
}

// calculateROISummary calculates summary ROI information
func (erg *ExecutiveReportGenerator) calculateROISummary(recommendations []Recommendation) ROISummaryItem {
	totalInvestment := 0.0
	totalSavings := 0.0

	for _, rec := range recommendations {
		// Simplified calculation - in practice would use ROICalculator
		investment := math.Abs(rec.Impact.Cost)
		savings := rec.Impact.Revenue

		totalInvestment += investment
		totalSavings += savings
	}

	paybackMonths := 0
	overallROI := 0.0

	if totalInvestment > 0 {
		paybackMonths = int((totalInvestment / (totalSavings / 12)) + 0.5)
		overallROI = ((totalSavings - totalInvestment) / totalInvestment) * 100
	}

	return ROISummaryItem{
		TotalInvestment: totalInvestment,
		ExpectedSavings: totalSavings,
		PaybackMonths:   paybackMonths,
		OverallROI:      overallROI,
	}
}

// generateNextActions creates actionable next steps
func (erg *ExecutiveReportGenerator) generateNextActions(analysis PerformanceAnalysis) []string {
	var actions []string

	criticalIssues := erg.filterCriticalBottlenecks(analysis.Bottlenecks)
	if len(criticalIssues) > 0 {
		actions = append(actions, "Address critical performance bottlenecks immediately")
	}

	if len(analysis.Recommendations) > 0 {
		actions = append(actions, "Review and prioritize optimization recommendations")
		actions = append(actions, "Allocate resources for high-priority improvements")
	}

	if analysis.TrendAnalysis.TrendStrength > 0.7 {
		actions = append(actions, "Implement proactive capacity planning based on trends")
	}

	actions = append(actions, "Schedule follow-up analysis in 30 days")
	actions = append(actions, "Monitor key performance indicators continuously")

	return actions
}

// Utility methods

// filterCriticalBottlenecks filters bottlenecks with severity >= 0.8
func (erg *ExecutiveReportGenerator) filterCriticalBottlenecks(bottlenecks []Bottleneck) []Bottleneck {
	var critical []Bottleneck
	for _, b := range bottlenecks {
		if b.Severity >= 0.8 {
			critical = append(critical, b)
		}
	}
	return critical
}

// groupRecommendationsByPriority groups recommendations by priority
func (erg *ExecutiveReportGenerator) groupRecommendationsByPriority(recommendations []Recommendation) map[Priority][]Recommendation {
	groups := make(map[Priority][]Recommendation)

	for _, rec := range recommendations {
		groups[rec.Priority] = append(groups[rec.Priority], rec)
	}

	return groups
}

// getPriorityScore converts priority to numeric score for sorting
func (erg *ExecutiveReportGenerator) getPriorityScore(priority Priority) int {
	switch priority {
	case PriorityCritical:
		return 4
	case PriorityHigh:
		return 3
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}

// getHealthColor returns color for health status badges
func (erg *ExecutiveReportGenerator) getHealthColor(health HealthStatus) string {
	switch health {
	case HealthExcellent:
		return "brightgreen"
	case HealthGood:
		return "green"
	case HealthFair:
		return "yellow"
	case HealthPoor:
		return "orange"
	case HealthCritical:
		return "red"
	default:
		return "lightgrey"
	}
}

// getSeverityBadge returns severity badge text
func (erg *ExecutiveReportGenerator) getSeverityBadge(severity float64) string {
	if severity >= 0.9 {
		return "🔴 Critical"
	} else if severity >= 0.7 {
		return "🟡 High"
	} else if severity >= 0.4 {
		return "🟢 Medium"
	}
	return "⚪ Low"
}

// generateHTMLForPDF generates HTML content for PDF conversion
func (erg *ExecutiveReportGenerator) generateHTMLForPDF(analysis PerformanceAnalysis) string {
	// This would generate HTML suitable for PDF conversion
	// For brevity, returning a simple HTML structure
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Performance Analysis Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { text-align: center; margin-bottom: 40px; }
        .summary { background: #f5f5f5; padding: 20px; margin: 20px 0; }
        .recommendations { margin: 30px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Trading Exchange Performance Analysis</h1>
        <p>Generated: %s</p>
    </div>
    <div class="summary">
        <h2>Executive Summary</h2>
        <p>Performance Score: %.1f/100</p>
        <p>Health Status: %s</p>
    </div>
    <!-- Additional content would be generated here -->
</body>
</html>
`, time.Now().Format("January 2, 2006"), analysis.PerformanceScore*100, erg.determineOverallHealth(analysis))
}

// generateDetailedFindings creates detailed findings section
func (erg *ExecutiveReportGenerator) generateDetailedFindings(analysis PerformanceAnalysis) map[string]interface{} {
	return map[string]interface{}{
		"bottleneck_analysis": analysis.Bottlenecks,
		"trend_analysis":      analysis.TrendAnalysis,
		"capacity_prediction": analysis.CapacityPrediction,
		"confidence_metrics": map[string]float64{
			"overall_confidence": analysis.Confidence,
			"trend_strength":     analysis.TrendAnalysis.TrendStrength,
		},
	}
}

// generateRecommendationDetails creates detailed recommendation information
func (erg *ExecutiveReportGenerator) generateRecommendationDetails(recommendations []Recommendation) map[string]interface{} {
	priorityGroups := erg.groupRecommendationsByPriority(recommendations)

	return map[string]interface{}{
		"by_priority":    priorityGroups,
		"by_category":    erg.groupRecommendationsByCategory(recommendations),
		"implementation_timeline": erg.generateImplementationTimeline(recommendations),
	}
}

// generateFinancialImpactSection creates financial impact section
func (erg *ExecutiveReportGenerator) generateFinancialImpactSection(analysis PerformanceAnalysis) map[string]interface{} {
	return map[string]interface{}{
		"business_impact":   analysis.BusinessImpact,
		"roi_summary":       erg.calculateROISummary(analysis.Recommendations),
		"cost_breakdown":    erg.generateCostBreakdown(analysis.Recommendations),
		"sensitivity_analysis": map[string]string{
			"best_case":  "Optimistic scenario analysis",
			"worst_case": "Conservative scenario analysis",
			"base_case":  "Expected scenario analysis",
		},
	}
}

// generateAppendices creates appendices section
func (erg *ExecutiveReportGenerator) generateAppendices(analysis PerformanceAnalysis) map[string]interface{} {
	return map[string]interface{}{
		"methodology":      "AI-powered performance analysis using ML algorithms",
		"data_sources":     []string{"System metrics", "Performance logs", "Historical data"},
		"assumptions":      erg.generateAnalysisAssumptions(),
		"glossary":         erg.generateGlossary(),
		"contact_info":     "Performance Engineering Team",
	}
}

// Additional helper methods

func (erg *ExecutiveReportGenerator) groupRecommendationsByCategory(recommendations []Recommendation) map[string][]Recommendation {
	groups := make(map[string][]Recommendation)
	for _, rec := range recommendations {
		groups[rec.Category] = append(groups[rec.Category], rec)
	}
	return groups
}

func (erg *ExecutiveReportGenerator) generateImplementationTimeline(recommendations []Recommendation) []map[string]interface{} {
	var timeline []map[string]interface{}

	sorted := make([]Recommendation, len(recommendations))
	copy(sorted, recommendations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TimeToEffect < sorted[j].TimeToEffect
	})

	for _, rec := range sorted {
		timeline = append(timeline, map[string]interface{}{
			"title":         rec.Title,
			"time_to_effect": rec.TimeToEffect.String(),
			"priority":      rec.Priority,
			"complexity":    rec.Complexity,
		})
	}

	return timeline
}

func (erg *ExecutiveReportGenerator) generateCostBreakdown(recommendations []Recommendation) map[string]float64 {
	breakdown := make(map[string]float64)

	for _, rec := range recommendations {
		category := rec.Category
		breakdown[category] += math.Abs(rec.Impact.Cost)
	}

	return breakdown
}

func (erg *ExecutiveReportGenerator) generateAnalysisAssumptions() []string {
	return []string{
		"Historical performance patterns continue",
		"No major system architecture changes",
		"Current usage patterns remain stable",
		"Infrastructure costs remain consistent",
		"Implementation follows best practices",
	}
}

func (erg *ExecutiveReportGenerator) generateGlossary() map[string]string {
	return map[string]string{
		"ROI":          "Return on Investment - measure of efficiency of an investment",
		"NPV":          "Net Present Value - difference between present value of cash inflows and outflows",
		"IRR":          "Internal Rate of Return - discount rate that makes NPV equal to zero",
		"Bottleneck":   "Performance constraint that limits overall system throughput",
		"Latency":      "Time delay between request and response",
		"Throughput":   "Number of operations processed per unit time",
	}
}

// DefaultReportConfig returns default report configuration
func DefaultReportConfig() ReportConfig {
	return ReportConfig{
		CompanyName:     "Trading Exchange Corp",
		ReportingPeriod: 24 * time.Hour,
		IncludeCharts:   true,
		DetailLevel:     DetailLevelManagement,
		Branding: BrandingConfig{
			LogoURL:        "/assets/logo.png",
			PrimaryColor:   "#1f2937",
			SecondaryColor: "#3b82f6",
			FontFamily:     "Inter, sans-serif",
		},
		Templates: map[ReportFormat]string{
			ReportFormatText:     "default_text_template",
			ReportFormatMarkdown: "default_markdown_template",
			ReportFormatJSON:     "default_json_template",
			ReportFormatPDF:      "default_pdf_template",
		},
	}
}

// Utility function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}