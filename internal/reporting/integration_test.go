package reporting

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestReportingSystemIntegration(t *testing.T) {
	ctx := context.Background()

	// Initialize all components
	businessAnalyzer := NewStandardBusinessAnalyzer()
	roiCalculator := NewStandardROICalculator()
	templateEngine := NewStandardTemplateEngine()
	reportGenerator := NewStandardReportGenerator(businessAnalyzer, roiCalculator, templateEngine)

	// Comprehensive test data representing a real trading platform scenario
	testParams := ReportParameters{
		Period:      "Q3_2024",
		CompanyName: "TradeCorp Financial Exchange",
		Department:  "Executive Leadership",
		RequestedBy: "Chief Executive Officer",
		BusinessData: BusinessData{
			FinancialMetrics: FinancialMetrics{
				Revenue:      2500000.0,
				Expenses:     1800000.0,
				Profit:       700000.0,
				CashFlow:     650000.0,
				DebtToEquity: 0.35,
				CurrentRatio: 2.2,
				QuickRatio:   1.9,
				GrossMargin:  28.0,
				NetMargin:    28.0,
				ROI:          22.5,
			},
			OperationalMetrics: OperationalMetrics{
				SystemUptime:      99.7,
				TransactionVolume: 350000,
				ProcessingSpeed:   42.0,
				ErrorRate:        0.025,
				CustomerRetention: 94.5,
				EmployeeTurnover:  6.8,
				ComplianceScore:   96.5,
				SecurityScore:     91.0,
			},
			MarketData: MarketData{
				MarketVolatility:          14.2,
				CompetitorCount:           7,
				MarketGrowthRate:          18.5,
				CustomerAcquisitionCost:   185.0,
				CustomerLifetimeValue:     3200.0,
				MarketShare:               11.2,
			},
			Period:     "Q3_2024",
			DataSource: "enterprise_data_warehouse",
			Timestamp:  time.Now(),
		},
		PerformanceData: PerformanceData{
			Revenue:              2500000.0,
			Costs:               1800000.0,
			Transactions:         350000,
			ActiveUsers:          18500,
			SystemUptime:         99.7,
			ResponseTime:         42.0,
			ErrorRate:           0.025,
			CustomerSatisfaction: 4.6,
			MarketShare:         11.2,
			GrowthRate:          18.5,
			Period:              "Q3_2024",
			ComparisonPeriod:    "Q2_2024",
			BenchmarkData: map[string]float64{
				"industry_revenue_growth":        12.0,
				"industry_customer_satisfaction": 4.3,
				"industry_uptime":               99.2,
				"industry_response_time":        65.0,
				"industry_error_rate":           0.045,
			},
		},
		InvestmentData: InvestmentData{
			InitialCost: 750000.0,
			ProjectName: "Next-Generation Trading Infrastructure",
			Category:    "Strategic Technology Investment",
			Benefits: []InvestmentBenefit{
				{
					Type:        "Revenue Increase",
					Amount:      380000.0,
					Frequency:   "Annual",
					Description: "Enhanced trading capacity and algorithm efficiency",
					Confidence:  0.82,
				},
				{
					Type:        "Cost Savings",
					Amount:      140000.0,
					Frequency:   "Annual",
					Description: "Reduced operational costs through automation",
					Confidence:  0.88,
				},
				{
					Type:        "Risk Reduction",
					Amount:      65000.0,
					Frequency:   "Annual",
					Description: "Improved compliance and reduced regulatory risk",
					Confidence:  0.75,
				},
			},
			Risks: []InvestmentRisk{
				{
					Type:        "Implementation",
					Probability: 0.25,
					Impact:      120000.0,
					Description: "Complex integration with legacy systems",
				},
				{
					Type:        "Market",
					Probability: 0.15,
					Impact:      85000.0,
					Description: "Market volatility affecting expected returns",
				},
				{
					Type:        "Technology",
					Probability: 0.1,
					Impact:      50000.0,
					Description: "Emerging technology adoption risks",
				},
			},
			Timeline:     30,
			DiscountRate: 0.09,
			Currency:     "USD",
			Department:   "Technology & Operations",
			Stakeholder:  "Chief Technology Officer",
			StartDate:    time.Now().AddDate(0, 1, 0),
		},
		CostOptimization: CostOptimizationData{
			CurrentCosts: map[string]float64{
				"infrastructure":     450000.0,
				"personnel":          800000.0,
				"licensing":          150000.0,
				"compliance":         120000.0,
				"data_feeds":         180000.0,
			},
			ProposedChanges: []OptimizationProposal{
				{
					Category:           "infrastructure",
					Description:        "Migrate to hybrid cloud architecture",
					EstimatedSavings:   95000.0,
					ImplementationCost: 65000.0,
					TimeToImplement:    150,
					RiskLevel:         "Medium",
					ImpactAreas:       []string{"scalability", "reliability", "cost_efficiency"},
				},
				{
					Category:           "data_feeds",
					Description:        "Consolidate and optimize market data providers",
					EstimatedSavings:   42000.0,
					ImplementationCost: 18000.0,
					TimeToImplement:    90,
					RiskLevel:         "Low",
					ImpactAreas:       []string{"data_quality", "cost_reduction"},
				},
				{
					Category:           "personnel",
					Description:        "Implement advanced automation for routine operations",
					EstimatedSavings:   125000.0,
					ImplementationCost: 85000.0,
					TimeToImplement:    180,
					RiskLevel:         "Medium",
					ImpactAreas:       []string{"efficiency", "accuracy", "scalability"},
				},
			},
			TargetSavings: 200000.0,
			TimeFrame:     365,
		},
		ReportScope:    []string{"performance", "financial", "risk", "roi", "optimization"},
		OutputFormat:   JSON,
		IncludeCharts:  true,
		IncludeMetrics: true,
		CustomFields: map[string]interface{}{
			"strategic_initiative": "Digital Transformation 2024",
			"board_presentation":   true,
			"quarterly_review":     true,
			"stakeholder_count":    12,
		},
	}

	// Test full report generation
	t.Run("full_executive_report_generation", func(t *testing.T) {
		start := time.Now()
		report, err := reportGenerator.GenerateExecutiveReport(ctx, testParams)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("GenerateExecutiveReport() error = %v", err)
		}

		if report == nil {
			t.Fatal("Expected complete executive report, got nil")
		}

		// Performance test - should complete within reasonable time
		if duration > 5*time.Second {
			t.Errorf("Report generation took %v, expected < 5s", duration)
		}

		// Validate report completeness
		validateReportCompleteness(t, report, testParams)

		// Validate data consistency
		validateDataConsistency(t, report, testParams)

		// Test report serialization to JSON
		jsonData, err := json.Marshal(report)
		if err != nil {
			t.Fatalf("JSON serialization error = %v", err)
		}

		if len(jsonData) < 1000 {
			t.Error("JSON report seems incomplete, too small")
		}

		t.Logf("Report generation completed in %v", duration)
		t.Logf("Generated report size: %d bytes", len(jsonData))
	})

	// Test individual component integration
	t.Run("business_analyzer_integration", func(t *testing.T) {
		analysis, err := businessAnalyzer.AnalyzePerformance(ctx, testParams.PerformanceData)
		if err != nil {
			t.Fatalf("AnalyzePerformance() error = %v", err)
		}

		costAnalysis, err := businessAnalyzer.CalculateCostSavings(ctx, testParams.CostOptimization)
		if err != nil {
			t.Fatalf("CalculateCostSavings() error = %v", err)
		}

		riskAssessment, err := businessAnalyzer.AssessRisk(ctx, testParams.BusinessData)
		if err != nil {
			t.Fatalf("AssessRisk() error = %v", err)
		}

		// Validate that analyses are comprehensive
		if analysis.OverallScore <= 0 || analysis.OverallScore > 100 {
			t.Errorf("Invalid performance score: %v", analysis.OverallScore)
		}

		if costAnalysis.TotalEstimatedSavings <= 0 {
			t.Error("Expected positive cost savings estimation")
		}

		if riskAssessment.RiskScore < 0 || riskAssessment.RiskScore > 100 {
			t.Errorf("Invalid risk score: %v", riskAssessment.RiskScore)
		}
	})

	// Test ROI calculator integration
	t.Run("roi_calculator_integration", func(t *testing.T) {
		roiAnalysis, err := roiCalculator.CalculateROI(testParams.InvestmentData)
		if err != nil {
			t.Fatalf("CalculateROI() error = %v", err)
		}

		npv, err := roiCalculator.CalculateNPV(testParams.InvestmentData)
		if err != nil {
			t.Fatalf("CalculateNPV() error = %v", err)
		}

		irr, err := roiCalculator.CalculateIRR(testParams.InvestmentData)
		if err != nil {
			t.Fatalf("CalculateIRR() error = %v", err)
		}

		projections, err := roiCalculator.GenerateProjections(testParams.InvestmentData)
		if err != nil {
			t.Fatalf("GenerateProjections() error = %v", err)
		}

		// Validate ROI calculations are realistic
		if roiAnalysis.AnnualROI <= 0 {
			t.Error("Expected positive ROI for profitable investment")
		}

		if npv <= 0 {
			t.Error("Expected positive NPV for profitable investment")
		}

		if irr <= 0 {
			t.Error("Expected positive IRR for profitable investment")
		}

		if len(projections.CashFlows) != int(testParams.InvestmentData.Timeline) {
			t.Errorf("Expected %d projections, got %d",
				testParams.InvestmentData.Timeline, len(projections.CashFlows))
		}
	})

	// Test template engine integration
	t.Run("template_engine_integration", func(t *testing.T) {
		report, err := reportGenerator.GenerateExecutiveReport(ctx, testParams)
		if err != nil {
			t.Fatalf("GenerateExecutiveReport() for template test error = %v", err)
		}

		// Test HTML template rendering
		htmlData, err := templateEngine.RenderTemplate("executive_report", report)
		if err != nil {
			t.Fatalf("RenderTemplate() error = %v", err)
		}

		htmlContent := string(htmlData)
		if !strings.Contains(htmlContent, "<html") {
			t.Error("HTML template missing html tag")
		}

		if !strings.Contains(htmlContent, testParams.CompanyName) {
			t.Error("HTML template missing company name")
		}

		if !strings.Contains(htmlContent, testParams.Period) {
			t.Error("HTML template missing report period")
		}

		// Test template validation
		if err := templateEngine.ValidateTemplate("executive_report", htmlContent); err != nil {
			t.Errorf("Template validation failed: %v", err)
		}
	})

	// Test multi-format output generation
	t.Run("multi_format_output_integration", func(t *testing.T) {
		formats := []OutputFormat{JSON, HTML, Text, CSV}

		for _, format := range formats {
			t.Run(string(format), func(t *testing.T) {
				testParams.OutputFormat = format
				reportGenerator.SetOutputFormat(format)

				report, err := reportGenerator.GenerateExecutiveReport(ctx, testParams)
				if err != nil {
					t.Fatalf("GenerateExecutiveReport() for %v format error = %v", format, err)
				}

				if report == nil {
					t.Fatalf("Expected report in %v format, got nil", format)
				}

				// Format-specific validations would go here
				// For now, just ensure report was generated
				if report.Metadata.ReportID == "" {
					t.Errorf("Report ID missing in %v format", format)
				}
			})
		}
	})

	// Test performance metrics under load
	t.Run("performance_load_test", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping load test in short mode")
		}

		const numReports = 10
		reports := make([]*ExecutiveReport, numReports)
		errors := make([]error, numReports)

		start := time.Now()

		// Generate multiple reports concurrently
		done := make(chan int, numReports)
		for i := 0; i < numReports; i++ {
			go func(index int) {
				reports[index], errors[index] = reportGenerator.GenerateExecutiveReport(ctx, testParams)
				done <- index
			}(i)
		}

		// Wait for all reports to complete
		for i := 0; i < numReports; i++ {
			<-done
		}

		duration := time.Since(start)
		avgDuration := duration / numReports

		// Check for errors
		errorCount := 0
		for i, err := range errors {
			if err != nil {
				errorCount++
				t.Errorf("Report %d generation error: %v", i, err)
			}
		}

		if errorCount > 0 {
			t.Errorf("Failed to generate %d out of %d reports", errorCount, numReports)
		}

		// Performance expectations
		if avgDuration > 2*time.Second {
			t.Errorf("Average report generation time %v too slow, expected < 2s", avgDuration)
		}

		t.Logf("Generated %d reports in %v (avg: %v)", numReports, duration, avgDuration)
	})
}

func validateReportCompleteness(t *testing.T, report *ExecutiveReport, params ReportParameters) {
	t.Helper()

	// Validate metadata
	if report.Metadata.ReportID == "" {
		t.Error("Report metadata missing ID")
	}

	if report.Metadata.Period != params.Period {
		t.Errorf("Report period mismatch: got %v, want %v",
			report.Metadata.Period, params.Period)
	}

	// Validate executive summary
	if report.ExecutiveSummary.Overview == "" {
		t.Error("Executive summary missing overview")
	}

	if len(report.ExecutiveSummary.KeyFindings) == 0 {
		t.Error("Executive summary missing key findings")
	}

	if len(report.ExecutiveSummary.Recommendations) == 0 {
		t.Error("Executive summary missing recommendations")
	}

	// Validate sections based on report scope
	for _, scope := range params.ReportScope {
		switch scope {
		case "performance":
			if report.PerformanceSection == nil {
				t.Error("Performance section missing despite being in scope")
			}
		case "roi":
			if report.ROIAnalysisSection == nil {
				t.Error("ROI analysis section missing despite being in scope")
			}
		case "risk":
			if report.RiskMitigationSection == nil {
				t.Error("Risk mitigation section missing despite being in scope")
			}
		case "optimization":
			if report.CostOptimizationSection == nil {
				t.Error("Cost optimization section missing despite being in scope")
			}
		}
	}
}

func validateDataConsistency(t *testing.T, report *ExecutiveReport, params ReportParameters) {
	t.Helper()

	// Validate financial data consistency
	if report.PerformanceSection != nil {
		if report.PerformanceSection.Revenue != params.PerformanceData.Revenue {
			t.Errorf("Revenue inconsistency: report %v, params %v",
				report.PerformanceSection.Revenue, params.PerformanceData.Revenue)
		}
	}

	// Validate ROI data consistency
	if report.ROIAnalysisSection != nil {
		if report.ROIAnalysisSection.InitialInvestment != params.InvestmentData.InitialCost {
			t.Errorf("Initial investment inconsistency: report %v, params %v",
				report.ROIAnalysisSection.InitialInvestment, params.InvestmentData.InitialCost)
		}
	}

	// Validate cost optimization consistency
	if report.CostOptimizationSection != nil {
		totalCurrentCosts := 0.0
		for _, cost := range params.CostOptimization.CurrentCosts {
			totalCurrentCosts += cost
		}

		if abs(report.CostOptimizationSection.CurrentTotalCosts-totalCurrentCosts) > 0.01 {
			t.Errorf("Total costs inconsistency: report %v, calculated %v",
				report.CostOptimizationSection.CurrentTotalCosts, totalCurrentCosts)
		}
	}
}