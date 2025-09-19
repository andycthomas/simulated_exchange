package reporting

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestStandardReportGenerator_GenerateExecutiveReport(t *testing.T) {
	generator := NewStandardReportGenerator(
		NewStandardBusinessAnalyzer(),
		NewStandardROICalculator(),
		NewStandardTemplateEngine(),
	)
	ctx := context.Background()

	params := ReportParameters{
		Period:           "Q3_2024",
		CompanyName:     "TradeCorp Exchange",
		Department:      "Technology",
		RequestedBy:     "CEO",
		BusinessData: BusinessData{
			FinancialMetrics: FinancialMetrics{
				Revenue:      1500000.0,
				Expenses:     1000000.0,
				Profit:       500000.0,
				CashFlow:     450000.0,
				DebtToEquity: 0.3,
				CurrentRatio: 2.0,
				QuickRatio:   1.5,
				GrossMargin:  33.3,
				NetMargin:    25.0,
				ROI:          20.0,
			},
			OperationalMetrics: OperationalMetrics{
				SystemUptime:      99.5,
				TransactionVolume: 200000,
				ProcessingSpeed:   60.0,
				ErrorRate:        0.03,
				CustomerRetention: 95.0,
				EmployeeTurnover:  5.0,
				ComplianceScore:   98.0,
				SecurityScore:     92.0,
			},
			MarketData: MarketData{
				MarketVolatility:          12.5,
				CompetitorCount:           6,
				MarketGrowthRate:          15.0,
				CustomerAcquisitionCost:   200.0,
				CustomerLifetimeValue:     3000.0,
				MarketShare:               8.0,
			},
			Period:     "Q3_2024",
			DataSource: "integrated_systems",
			Timestamp:  time.Now(),
		},
		PerformanceData: PerformanceData{
			Revenue:              1500000.0,
			Costs:               1000000.0,
			Transactions:         200000,
			ActiveUsers:          12000,
			SystemUptime:         99.5,
			ResponseTime:         60.0,
			ErrorRate:           0.03,
			CustomerSatisfaction: 4.5,
			MarketShare:         8.0,
			GrowthRate:          15.0,
			Period:              "Q3_2024",
			ComparisonPeriod:    "Q2_2024",
			BenchmarkData: map[string]float64{
				"industry_revenue_growth":    10.0,
				"industry_customer_satisfaction": 4.2,
				"industry_uptime":           99.0,
			},
		},
		InvestmentData: InvestmentData{
			InitialCost: 250000.0,
			ProjectName: "Trading Platform Enhancement",
			Category:    "Technology Upgrade",
			Benefits: []InvestmentBenefit{
				{
					Type:        "Revenue Increase",
					Amount:      100000.0,
					Frequency:   "Annual",
					Description: "Increased trading capacity and efficiency",
					Confidence:  0.85,
				},
				{
					Type:        "Cost Savings",
					Amount:      50000.0,
					Frequency:   "Annual",
					Description: "Reduced operational overhead",
					Confidence:  0.9,
				},
			},
			Risks: []InvestmentRisk{
				{
					Type:        "Implementation",
					Probability: 0.2,
					Impact:      30000.0,
					Description: "Integration complexity and potential delays",
				},
			},
			Timeline:     24,
			DiscountRate: 0.08,
			Currency:     "USD",
			Department:   "Technology",
			Stakeholder:  "CTO",
			StartDate:    time.Now(),
		},
		CostOptimization: CostOptimizationData{
			CurrentCosts: map[string]float64{
				"infrastructure": 200000.0,
				"personnel":      500000.0,
				"licensing":      100000.0,
			},
			ProposedChanges: []OptimizationProposal{
				{
					Category:           "infrastructure",
					Description:        "Cloud migration for better scalability",
					EstimatedSavings:   60000.0,
					ImplementationCost: 40000.0,
					TimeToImplement:    180,
					RiskLevel:         "Medium",
					ImpactAreas:       []string{"scalability", "maintenance"},
				},
			},
			TargetSavings: 100000.0,
			TimeFrame:     365,
		},
		ReportScope:   []string{"performance", "financial", "risk", "roi"},
		OutputFormat:  JSON,
		IncludeCharts: true,
		IncludeMetrics: true,
		CustomFields: map[string]interface{}{
			"quarterly_focus": "Platform Modernization",
			"board_meeting":   true,
		},
	}

	report, err := generator.GenerateExecutiveReport(ctx, params)
	if err != nil {
		t.Fatalf("GenerateExecutiveReport() error = %v", err)
	}

	if report == nil {
		t.Fatal("Expected executive report, got nil")
	}

	// Test report metadata
	if report.Metadata.ReportID == "" {
		t.Error("Expected report ID, got empty string")
	}

	if report.Metadata.GeneratedBy != "StandardReportGenerator" {
		t.Errorf("GeneratedBy = %v, want StandardReportGenerator", report.Metadata.GeneratedBy)
	}

	if report.Metadata.Period != params.Period {
		t.Errorf("Period = %v, want %v", report.Metadata.Period, params.Period)
	}

	// Test executive summary
	if report.ExecutiveSummary.Overview == "" {
		t.Error("Expected executive summary overview, got empty string")
	}

	if len(report.ExecutiveSummary.KeyFindings) == 0 {
		t.Error("Expected key findings in executive summary")
	}

	if len(report.ExecutiveSummary.Recommendations) == 0 {
		t.Error("Expected recommendations in executive summary")
	}

	// Test performance section
	if report.PerformanceSection == nil {
		t.Fatal("Expected performance section, got nil")
	}

	if report.PerformanceSection.OverallScore <= 0 {
		t.Error("Expected positive overall performance score")
	}

	// Test cost optimization section
	if report.CostOptimizationSection == nil {
		t.Fatal("Expected cost optimization section, got nil")
	}

	if report.CostOptimizationSection.TotalEstimatedSavings <= 0 {
		t.Error("Expected positive estimated savings")
	}

	// Test risk mitigation section
	if report.RiskMitigationSection == nil {
		t.Fatal("Expected risk mitigation section, got nil")
	}

	if report.RiskMitigationSection.OverallRiskLevel == "" {
		t.Error("Expected overall risk level assessment")
	}

	// Test ROI analysis section
	if report.ROIAnalysisSection == nil {
		t.Fatal("Expected ROI analysis section, got nil")
	}

	if report.ROIAnalysisSection.AnnualROI <= 0 {
		t.Error("Expected positive annual ROI")
	}
}

func TestStandardReportGenerator_CreateROIAnalysis(t *testing.T) {
	generator := NewStandardReportGenerator(
		NewStandardBusinessAnalyzer(),
		NewStandardROICalculator(),
		NewStandardTemplateEngine(),
	)
	ctx := context.Background()

	investmentData := InvestmentData{
		InitialCost: 500000.0,
		ProjectName: "Advanced Trading Analytics Platform",
		Category:    "Technology Innovation",
		Benefits: []InvestmentBenefit{
			{
				Type:        "Revenue Increase",
				Amount:      200000.0,
				Frequency:   "Annual",
				Description: "Enhanced trading algorithms and market analysis",
				Confidence:  0.8,
			},
			{
				Type:        "Cost Savings",
				Amount:      80000.0,
				Frequency:   "Annual",
				Description: "Automated risk management and reporting",
				Confidence:  0.9,
			},
		},
		Risks: []InvestmentRisk{
			{
				Type:        "Technical",
				Probability: 0.15,
				Impact:      75000.0,
				Description: "Algorithm complexity and integration challenges",
			},
			{
				Type:        "Market",
				Probability: 0.1,
				Impact:      50000.0,
				Description: "Changing market conditions affecting effectiveness",
			},
		},
		Timeline:     36,
		DiscountRate: 0.10,
		Currency:     "USD",
		Department:   "Quantitative Research",
		Stakeholder:  "Head of Quant",
		StartDate:    time.Now(),
	}

	analysis, err := generator.CreateROIAnalysis(ctx, investmentData)
	if err != nil {
		t.Fatalf("CreateROIAnalysis() error = %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected ROI analysis, got nil")
	}

	if analysis.InitialInvestment != investmentData.InitialCost {
		t.Errorf("InitialInvestment = %v, want %v",
			analysis.InitialInvestment, investmentData.InitialCost)
	}

	if analysis.AnnualBenefits <= 0 {
		t.Error("Expected positive annual benefits")
	}

	expectedAnnualBenefits := 280000.0 // 200000 + 80000
	if abs(analysis.AnnualBenefits-expectedAnnualBenefits) > 1000.0 {
		t.Errorf("AnnualBenefits = %v, want approximately %v",
			analysis.AnnualBenefits, expectedAnnualBenefits)
	}

	if analysis.PaybackPeriodMonths <= 0 {
		t.Error("Expected positive payback period")
	}

	if analysis.AnnualROI <= 0 {
		t.Error("Expected positive annual ROI")
	}

	if analysis.NetPresentValue <= 0 {
		t.Error("Expected positive NPV for profitable investment")
	}

	if len(analysis.CashFlowProjections) != int(investmentData.Timeline) {
		t.Errorf("Expected %d cash flow projections, got %d",
			investmentData.Timeline, len(analysis.CashFlowProjections))
	}
}

func TestStandardReportGenerator_BuildRiskAssessment(t *testing.T) {
	generator := NewStandardReportGenerator(
		NewStandardBusinessAnalyzer(),
		NewStandardROICalculator(),
		NewStandardTemplateEngine(),
	)
	ctx := context.Background()

	businessData := BusinessData{
		FinancialMetrics: FinancialMetrics{
			Revenue:      2000000.0,
			Expenses:     1500000.0,
			Profit:       500000.0,
			CashFlow:     400000.0,
			DebtToEquity: 0.4,
			CurrentRatio: 1.8,
			QuickRatio:   1.3,
			GrossMargin:  25.0,
			NetMargin:    20.0,
			ROI:          15.0,
		},
		OperationalMetrics: OperationalMetrics{
			SystemUptime:      98.5,
			TransactionVolume: 150000,
			ProcessingSpeed:   80.0,
			ErrorRate:        0.08,
			CustomerRetention: 88.0,
			EmployeeTurnover:  12.0,
			ComplianceScore:   85.0,
			SecurityScore:     78.0,
		},
		MarketData: MarketData{
			MarketVolatility:          20.0,
			CompetitorCount:           10,
			MarketGrowthRate:          8.0,
			CustomerAcquisitionCost:   400.0,
			CustomerLifetimeValue:     2000.0,
			MarketShare:               5.0,
		},
		Period:     "Q3_2024",
		DataSource: "enterprise_systems",
		Timestamp:  time.Now(),
	}

	assessment, err := generator.BuildRiskAssessment(ctx, businessData)
	if err != nil {
		t.Fatalf("BuildRiskAssessment() error = %v", err)
	}

	if assessment == nil {
		t.Fatal("Expected risk assessment, got nil")
	}

	if assessment.OverallRiskLevel == "" {
		t.Error("Expected overall risk level, got empty string")
	}

	if assessment.RiskScore < 0 || assessment.RiskScore > 100 {
		t.Errorf("RiskScore = %v, want between 0 and 100", assessment.RiskScore)
	}

	if len(assessment.RiskFactors) == 0 {
		t.Error("Expected risk factors, got none")
	}

	if len(assessment.MitigationStrategies) == 0 {
		t.Error("Expected mitigation strategies, got none")
	}

	// Verify risk factors cover major categories
	categories := make(map[string]bool)
	for _, factor := range assessment.RiskFactors {
		categories[factor.Category] = true
	}

	expectedCategories := []string{"Financial", "Operational", "Market"}
	for _, expected := range expectedCategories {
		if !categories[expected] {
			t.Errorf("Expected risk factor in category %v", expected)
		}
	}
}

func TestStandardReportGenerator_OutputFormats(t *testing.T) {
	generator := NewStandardReportGenerator(
		NewStandardBusinessAnalyzer(),
		NewStandardROICalculator(),
		NewStandardTemplateEngine(),
	)

	// Test supported formats
	formats := generator.GetSupportedFormats()
	expectedFormats := []OutputFormat{JSON, HTML, Text, CSV}

	for _, expected := range expectedFormats {
		found := false
		for _, format := range formats {
			if format == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format %v not found in supported formats", expected)
		}
	}

	// Test format setting
	generator.SetOutputFormat(HTML)
	// No direct way to test this without exposing internal state,
	// but we can test that it doesn't panic

	generator.SetOutputFormat(JSON)
	generator.SetOutputFormat(Text)
	generator.SetOutputFormat(CSV)
}

func TestStandardReportGenerator_FormatConversion(t *testing.T) {
	generator := NewStandardReportGenerator(
		NewStandardBusinessAnalyzer(),
		NewStandardROICalculator(),
		NewStandardTemplateEngine(),
	)

	// Create a sample report
	report := &ExecutiveReport{
		Metadata: ReportMetadata{
			ReportID:    "TEST-001",
			Title:       "Test Executive Report",
			Period:      "Q3_2024",
			GeneratedAt: time.Now(),
			GeneratedBy: "Test Suite",
			Version:     "1.0",
		},
		ExecutiveSummary: ExecutiveSummary{
			Overview: "Test overview for format conversion testing",
			KeyFindings: []KeyFinding{
				{
					Finding:    "Test finding 1",
					Impact:     "High",
					Severity:   "Medium",
					Category:   "Performance",
					Evidence:   "Test evidence",
					Confidence: 0.85,
				},
			},
			Recommendations: []Recommendation{
				{
					Priority:     "High",
					Category:     "Technology",
					Description:  "Test recommendation",
					Impact:       "Improve system performance by 25%",
					Timeline:     "90 days",
					Resources:    []string{"Development Team", "DevOps"},
					Success:      "Performance metrics improvement",
					Risk:         "Low implementation risk",
					Cost:         75000.0,
					Confidence:   0.9,
				},
			},
		},
	}

	t.Run("json_format", func(t *testing.T) {
		generator.SetOutputFormat(JSON)
		data, err := generator.convertToFormat(report, JSON)
		if err != nil {
			t.Fatalf("convertToFormat(JSON) error = %v", err)
		}

		if len(data) == 0 {
			t.Error("Expected JSON data, got empty")
		}

		// Basic JSON validation
		if !strings.Contains(string(data), "\"ReportID\":\"TEST-001\"") {
			t.Error("JSON output missing expected report ID")
		}
	})

	t.Run("html_format", func(t *testing.T) {
		generator.SetOutputFormat(HTML)
		data, err := generator.convertToFormat(report, HTML)
		if err != nil {
			t.Fatalf("convertToFormat(HTML) error = %v", err)
		}

		if len(data) == 0 {
			t.Error("Expected HTML data, got empty")
		}

		htmlContent := string(data)
		if !strings.Contains(htmlContent, "<html") {
			t.Error("HTML output missing html tag")
		}

		if !strings.Contains(htmlContent, "TEST-001") {
			t.Error("HTML output missing report ID")
		}
	})

	t.Run("text_format", func(t *testing.T) {
		generator.SetOutputFormat(Text)
		data, err := generator.convertToFormat(report, Text)
		if err != nil {
			t.Fatalf("convertToFormat(Text) error = %v", err)
		}

		if len(data) == 0 {
			t.Error("Expected text data, got empty")
		}

		textContent := string(data)
		if !strings.Contains(textContent, "EXECUTIVE REPORT") {
			t.Error("Text output missing header")
		}

		if !strings.Contains(textContent, "TEST-001") {
			t.Error("Text output missing report ID")
		}
	})

	t.Run("csv_format", func(t *testing.T) {
		generator.SetOutputFormat(CSV)
		data, err := generator.convertToFormat(report, CSV)
		if err != nil {
			t.Fatalf("convertToFormat(CSV) error = %v", err)
		}

		if len(data) == 0 {
			t.Error("Expected CSV data, got empty")
		}

		csvContent := string(data)
		if !strings.Contains(csvContent, "Field,Value") {
			t.Error("CSV output missing header row")
		}

		if !strings.Contains(csvContent, "TEST-001") {
			t.Error("CSV output missing report ID")
		}
	})
}