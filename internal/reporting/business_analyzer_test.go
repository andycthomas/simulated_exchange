package reporting

import (
	"context"
	"testing"
	"time"
)

func TestStandardBusinessAnalyzer_AnalyzePerformance(t *testing.T) {
	analyzer := NewStandardBusinessAnalyzer()
	ctx := context.Background()

	tests := []struct {
		name     string
		data     PerformanceData
		expected float64 // expected overall score range
		minScore float64
		maxScore float64
	}{
		{
			name: "high_performance_trading_system",
			data: PerformanceData{
				Revenue:           1500000.0,
				Costs:             800000.0,
				Transactions:      250000,
				ActiveUsers:       15000,
				SystemUptime:      99.8,
				ResponseTime:      45.0,
				ErrorRate:         0.02,
				CustomerSatisfaction: 4.7,
				MarketShare:       12.5,
				GrowthRate:        15.2,
				Period:            "Q3_2024",
				ComparisonPeriod:  "Q2_2024",
				BenchmarkData: map[string]float64{
					"industry_revenue_growth":    8.5,
					"industry_customer_satisfaction": 4.2,
					"industry_uptime":           99.2,
				},
			},
			minScore: 85.0,
			maxScore: 95.0,
		},
		{
			name: "underperforming_system",
			data: PerformanceData{
				Revenue:           500000.0,
				Costs:             600000.0,
				Transactions:      50000,
				ActiveUsers:       3000,
				SystemUptime:      97.5,
				ResponseTime:      150.0,
				ErrorRate:         0.15,
				CustomerSatisfaction: 3.2,
				MarketShare:       2.1,
				GrowthRate:        -2.5,
				Period:            "Q3_2024",
				ComparisonPeriod:  "Q2_2024",
				BenchmarkData: map[string]float64{
					"industry_revenue_growth":    8.5,
					"industry_customer_satisfaction": 4.2,
					"industry_uptime":           99.2,
				},
			},
			minScore: 25.0,
			maxScore: 45.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := analyzer.AnalyzePerformance(ctx, tt.data)
			if err != nil {
				t.Fatalf("AnalyzePerformance() error = %v", err)
			}

			if analysis == nil {
				t.Fatal("Expected analysis result, got nil")
			}

			if analysis.OverallScore < tt.minScore || analysis.OverallScore > tt.maxScore {
				t.Errorf("OverallScore = %v, want between %v and %v",
					analysis.OverallScore, tt.minScore, tt.maxScore)
			}

			if len(analysis.KeyMetrics) == 0 {
				t.Error("Expected key metrics, got none")
			}

			if len(analysis.Trends) == 0 {
				t.Error("Expected trends analysis, got none")
			}

			if analysis.Grade == "" {
				t.Error("Expected performance grade, got empty string")
			}

			if len(analysis.Recommendations) == 0 {
				t.Error("Expected recommendations, got none")
			}
		})
	}
}

func TestStandardBusinessAnalyzer_CalculateCostSavings(t *testing.T) {
	analyzer := NewStandardBusinessAnalyzer()
	ctx := context.Background()

	optimizationData := CostOptimizationData{
		CurrentCosts: map[string]float64{
			"infrastructure": 150000.0,
			"personnel":      300000.0,
			"licensing":      50000.0,
			"maintenance":    25000.0,
		},
		ProposedChanges: []OptimizationProposal{
			{
				Category:        "infrastructure",
				Description:     "Migrate to cloud-native architecture",
				EstimatedSavings: 45000.0,
				ImplementationCost: 20000.0,
				TimeToImplement: 90,
				RiskLevel:       "Medium",
				ImpactAreas:     []string{"scalability", "maintenance", "reliability"},
			},
			{
				Category:        "personnel",
				Description:     "Automate trading operations monitoring",
				EstimatedSavings: 80000.0,
				ImplementationCost: 35000.0,
				TimeToImplement: 120,
				RiskLevel:       "Low",
				ImpactAreas:     []string{"efficiency", "accuracy", "availability"},
			},
		},
		TargetSavings: 100000.0,
		TimeFrame:     365,
	}

	analysis, err := analyzer.CalculateCostSavings(ctx, optimizationData)
	if err != nil {
		t.Fatalf("CalculateCostSavings() error = %v", err)
	}

	expectedTotalSavings := 125000.0 // 45000 + 80000
	if analysis.TotalEstimatedSavings != expectedTotalSavings {
		t.Errorf("TotalEstimatedSavings = %v, want %v",
			analysis.TotalEstimatedSavings, expectedTotalSavings)
	}

	expectedImplementationCost := 55000.0 // 20000 + 35000
	if analysis.TotalImplementationCost != expectedImplementationCost {
		t.Errorf("TotalImplementationCost = %v, want %v",
			analysis.TotalImplementationCost, expectedImplementationCost)
	}

	expectedNetSavings := 70000.0 // 125000 - 55000
	if analysis.NetSavings != expectedNetSavings {
		t.Errorf("NetSavings = %v, want %v", analysis.NetSavings, expectedNetSavings)
	}

	if len(analysis.Opportunities) != 2 {
		t.Errorf("Expected 2 opportunities, got %d", len(analysis.Opportunities))
	}

	if analysis.TargetAchievementRate <= 0 {
		t.Error("Expected positive target achievement rate")
	}
}

func TestStandardBusinessAnalyzer_AssessRisk(t *testing.T) {
	analyzer := NewStandardBusinessAnalyzer()
	ctx := context.Background()

	businessData := BusinessData{
		FinancialMetrics: FinancialMetrics{
			Revenue:         1200000.0,
			Expenses:        800000.0,
			Profit:          400000.0,
			CashFlow:        350000.0,
			DebtToEquity:    0.25,
			CurrentRatio:    2.1,
			QuickRatio:      1.8,
			GrossMargin:     33.3,
			NetMargin:       25.0,
			ROI:             18.5,
		},
		OperationalMetrics: OperationalMetrics{
			SystemUptime:        99.5,
			TransactionVolume:   180000,
			ProcessingSpeed:     75.0,
			ErrorRate:          0.05,
			CustomerRetention:   92.5,
			EmployeeTurnover:    8.2,
			ComplianceScore:     95.0,
			SecurityScore:       88.0,
		},
		MarketData: MarketData{
			MarketVolatility:    15.2,
			CompetitorCount:     8,
			MarketGrowthRate:    12.5,
			CustomerAcquisitionCost: 250.0,
			CustomerLifetimeValue:   2500.0,
			MarketShare:         8.5,
		},
		Period:     "Q3_2024",
		DataSource: "internal_systems",
		Timestamp:  time.Now(),
	}

	assessment, err := analyzer.AssessRisk(ctx, businessData)
	if err != nil {
		t.Fatalf("AssessRisk() error = %v", err)
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

	categoryFound := false
	for _, factor := range assessment.RiskFactors {
		if factor.Category == "Financial" || factor.Category == "Operational" || factor.Category == "Market" {
			categoryFound = true
			break
		}
	}
	if !categoryFound {
		t.Error("Expected risk factors with standard categories")
	}
}

func TestStandardBusinessAnalyzer_EdgeCases(t *testing.T) {
	analyzer := NewStandardBusinessAnalyzer()
	ctx := context.Background()

	t.Run("zero_revenue_performance", func(t *testing.T) {
		data := PerformanceData{
			Revenue:     0.0,
			Costs:       10000.0,
			Transactions: 0,
			ActiveUsers:  0,
			SystemUptime: 99.9,
			ResponseTime: 50.0,
			ErrorRate:   0.01,
			CustomerSatisfaction: 5.0,
			MarketShare: 0.0,
			GrowthRate:  0.0,
			Period:      "Q1_2024",
		}

		analysis, err := analyzer.AnalyzePerformance(ctx, data)
		if err != nil {
			t.Fatalf("AnalyzePerformance() with zero revenue error = %v", err)
		}

		if analysis.OverallScore > 30.0 {
			t.Errorf("Expected low score for zero revenue, got %v", analysis.OverallScore)
		}
	})

	t.Run("negative_costs_savings", func(t *testing.T) {
		optimizationData := CostOptimizationData{
			CurrentCosts: map[string]float64{
				"infrastructure": 50000.0,
			},
			ProposedChanges: []OptimizationProposal{
				{
					Category:           "infrastructure",
					Description:        "Expensive upgrade with minimal benefits",
					EstimatedSavings:   10000.0,
					ImplementationCost: 80000.0,
					TimeToImplement:    60,
					RiskLevel:         "High",
				},
			},
			TargetSavings: 50000.0,
			TimeFrame:     365,
		}

		analysis, err := analyzer.CalculateCostSavings(ctx, optimizationData)
		if err != nil {
			t.Fatalf("CalculateCostSavings() with high implementation cost error = %v", err)
		}

		if analysis.NetSavings >= 0 {
			t.Errorf("Expected negative net savings, got %v", analysis.NetSavings)
		}
	})

	t.Run("perfect_metrics_risk_assessment", func(t *testing.T) {
		businessData := BusinessData{
			FinancialMetrics: FinancialMetrics{
				Revenue:      2000000.0,
				Expenses:     1000000.0,
				Profit:       1000000.0,
				CashFlow:     950000.0,
				DebtToEquity: 0.1,
				CurrentRatio: 3.0,
				QuickRatio:   2.5,
				GrossMargin:  50.0,
				NetMargin:    50.0,
				ROI:          30.0,
			},
			OperationalMetrics: OperationalMetrics{
				SystemUptime:      99.99,
				TransactionVolume: 500000,
				ProcessingSpeed:   25.0,
				ErrorRate:        0.001,
				CustomerRetention: 98.0,
				EmployeeTurnover:  2.0,
				ComplianceScore:   100.0,
				SecurityScore:     100.0,
			},
			MarketData: MarketData{
				MarketVolatility:          5.0,
				CompetitorCount:           3,
				MarketGrowthRate:          25.0,
				CustomerAcquisitionCost:   100.0,
				CustomerLifetimeValue:     5000.0,
				MarketShare:               20.0,
			},
			Period:     "Q3_2024",
			DataSource: "internal_systems",
			Timestamp:  time.Now(),
		}

		assessment, err := analyzer.AssessRisk(ctx, businessData)
		if err != nil {
			t.Fatalf("AssessRisk() with perfect metrics error = %v", err)
		}

		if assessment.RiskScore > 30.0 {
			t.Errorf("Expected low risk score for perfect metrics, got %v", assessment.RiskScore)
		}

		if assessment.OverallRiskLevel != "Low" {
			t.Errorf("Expected Low risk level for perfect metrics, got %v", assessment.OverallRiskLevel)
		}
	})
}

func TestStandardBusinessAnalyzer_PerformanceGrading(t *testing.T) {
	analyzer := NewStandardBusinessAnalyzer()

	tests := []struct {
		score    float64
		expected string
	}{
		{95.0, "A+"},
		{87.0, "A"},
		{82.0, "B+"},
		{75.0, "B"},
		{68.0, "C+"},
		{62.0, "C"},
		{55.0, "D+"},
		{48.0, "D"},
		{25.0, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			grade := analyzer.determineGrade(tt.score)
			if grade != tt.expected {
				t.Errorf("determineGrade(%v) = %v, want %v", tt.score, grade, tt.expected)
			}
		})
	}
}