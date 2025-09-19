package reporting

import (
	"testing"
	"time"
)

func TestStandardROICalculator_CalculateROI(t *testing.T) {
	calculator := NewStandardROICalculator()

	tests := []struct {
		name                string
		investment          InvestmentData
		expectedROI         float64
		expectedPayback     float64
		expectedThreeYearROI float64
		tolerance           float64
	}{
		{
			name: "profitable_infrastructure_upgrade",
			investment: InvestmentData{
				InitialCost: 100000.0,
				ProjectName: "Trading System Infrastructure Upgrade",
				Category:    "Technology",
				Benefits: []InvestmentBenefit{
					{
						Type:        "Cost Savings",
						Amount:      40000.0,
						Frequency:   "Annual",
						Description: "Reduced operational costs through automation",
						Confidence:  0.9,
					},
					{
						Type:        "Revenue Increase",
						Amount:      25000.0,
						Frequency:   "Annual",
						Description: "Increased trading capacity and efficiency",
						Confidence:  0.8,
					},
				},
				Risks: []InvestmentRisk{
					{
						Type:        "Implementation",
						Probability: 0.2,
						Impact:      15000.0,
						Description: "Potential delays and cost overruns",
					},
					{
						Type:        "Market",
						Probability: 0.1,
						Impact:      10000.0,
						Description: "Market volatility affecting trading volumes",
					},
				},
				Timeline:     24,
				DiscountRate: 0.08,
				Currency:     "USD",
				Department:   "Technology",
				Stakeholder:  "CTO",
				StartDate:    time.Now(),
			},
			expectedROI:         65.0,
			expectedPayback:     18.5,
			expectedThreeYearROI: 195.0,
			tolerance:           5.0,
		},
		{
			name: "marginal_training_investment",
			investment: InvestmentData{
				InitialCost: 50000.0,
				ProjectName: "Staff Trading Platform Training",
				Category:    "Human Resources",
				Benefits: []InvestmentBenefit{
					{
						Type:        "Productivity Increase",
						Amount:      15000.0,
						Frequency:   "Annual",
						Description: "Improved staff efficiency and reduced errors",
						Confidence:  0.7,
					},
				},
				Risks: []InvestmentRisk{
					{
						Type:        "Adoption",
						Probability: 0.3,
						Impact:      5000.0,
						Description: "Staff resistance to new processes",
					},
				},
				Timeline:     12,
				DiscountRate: 0.08,
				Currency:     "USD",
				Department:   "Human Resources",
				Stakeholder:  "CHRO",
				StartDate:    time.Now(),
			},
			expectedROI:         30.0,
			expectedPayback:     40.0,
			expectedThreeYearROI: 90.0,
			tolerance:           10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calculator.CalculateROI(tt.investment)
			if err != nil {
				t.Fatalf("CalculateROI() error = %v", err)
			}

			if result == nil {
				t.Fatal("Expected ROI calculation result, got nil")
			}

			if abs(result.AnnualROI-tt.expectedROI) > tt.tolerance {
				t.Errorf("AnnualROI = %v, want %v ± %v",
					result.AnnualROI, tt.expectedROI, tt.tolerance)
			}

			if abs(result.PaybackPeriodMonths-tt.expectedPayback) > tt.tolerance {
				t.Errorf("PaybackPeriodMonths = %v, want %v ± %v",
					result.PaybackPeriodMonths, tt.expectedPayback, tt.tolerance)
			}

			if abs(result.ThreeYearROI-tt.expectedThreeYearROI) > tt.tolerance {
				t.Errorf("ThreeYearROI = %v, want %v ± %v",
					result.ThreeYearROI, tt.expectedThreeYearROI, tt.tolerance)
			}

			if result.NetPresentValue <= 0 && result.AnnualROI > 0 {
				t.Error("Expected positive NPV for profitable investment")
			}

			if len(result.CashFlowProjections) != int(tt.investment.Timeline) {
				t.Errorf("Expected %d cash flow projections, got %d",
					tt.investment.Timeline, len(result.CashFlowProjections))
			}

			if result.InitialInvestment != tt.investment.InitialCost {
				t.Errorf("InitialInvestment = %v, want %v",
					result.InitialInvestment, tt.investment.InitialCost)
			}
		})
	}
}

func TestStandardROICalculator_CalculateNPV(t *testing.T) {
	calculator := NewStandardROICalculator()

	investment := InvestmentData{
		InitialCost: 100000.0,
		Benefits: []InvestmentBenefit{
			{
				Type:      "Cost Savings",
				Amount:    50000.0,
				Frequency: "Annual",
			},
		},
		Timeline:     3,
		DiscountRate: 0.10,
	}

	npv, err := calculator.CalculateNPV(investment)
	if err != nil {
		t.Fatalf("CalculateNPV() error = %v", err)
	}

	expectedNPV := 24342.0 // Approximate NPV for this scenario
	tolerance := 1000.0

	if abs(npv-expectedNPV) > tolerance {
		t.Errorf("NPV = %v, want %v ± %v", npv, expectedNPV, tolerance)
	}
}

func TestStandardROICalculator_CalculateIRR(t *testing.T) {
	calculator := NewStandardROICalculator()

	investment := InvestmentData{
		InitialCost: 100000.0,
		Benefits: []InvestmentBenefit{
			{
				Type:      "Revenue Increase",
				Amount:    40000.0,
				Frequency: "Annual",
			},
		},
		Timeline: 5,
	}

	irr, err := calculator.CalculateIRR(investment)
	if err != nil {
		t.Fatalf("CalculateIRR() error = %v", err)
	}

	expectedIRR := 28.65 // Approximate IRR for this cash flow
	tolerance := 2.0

	if abs(irr-expectedIRR) > tolerance {
		t.Errorf("IRR = %v%%, want %v%% ± %v%%", irr, expectedIRR, tolerance)
	}

	if irr < 0 {
		t.Error("IRR should be positive for profitable investment")
	}
}

func TestStandardROICalculator_GenerateProjections(t *testing.T) {
	calculator := NewStandardROICalculator()

	investment := InvestmentData{
		InitialCost: 75000.0,
		ProjectName: "Market Data Enhancement",
		Benefits: []InvestmentBenefit{
			{
				Type:      "Revenue Increase",
				Amount:    30000.0,
				Frequency: "Annual",
			},
			{
				Type:      "Cost Savings",
				Amount:    15000.0,
				Frequency: "Annual",
			},
		},
		Risks: []InvestmentRisk{
			{
				Type:        "Technology",
				Probability: 0.15,
				Impact:      10000.0,
			},
		},
		Timeline: 24,
	}

	projections, err := calculator.GenerateProjections(investment)
	if err != nil {
		t.Fatalf("GenerateProjections() error = %v", err)
	}

	if len(projections.CashFlows) != 24 {
		t.Errorf("Expected 24 monthly cash flows, got %d", len(projections.CashFlows))
	}

	if projections.TotalInvestment != investment.InitialCost {
		t.Errorf("TotalInvestment = %v, want %v",
			projections.TotalInvestment, investment.InitialCost)
	}

	if projections.ProjectedROI <= 0 {
		t.Error("Expected positive projected ROI")
	}

	if len(projections.RiskAdjustments) == 0 {
		t.Error("Expected risk adjustments based on investment risks")
	}

	if len(projections.Scenarios) == 0 {
		t.Error("Expected scenario projections")
	}

	// Verify cash flows are reasonable
	for i, cf := range projections.CashFlows {
		if i == 0 && cf.Amount >= 0 {
			t.Error("First month should show negative cash flow (initial investment)")
		}
		if i > 0 && cf.Amount <= 0 {
			t.Errorf("Month %d should show positive cash flow, got %v", i+1, cf.Amount)
		}
	}
}

func TestStandardROICalculator_EdgeCases(t *testing.T) {
	calculator := NewStandardROICalculator()

	t.Run("zero_initial_cost", func(t *testing.T) {
		investment := InvestmentData{
			InitialCost: 0.0,
			Benefits: []InvestmentBenefit{
				{Type: "Revenue", Amount: 10000.0, Frequency: "Annual"},
			},
			Timeline: 12,
		}

		_, err := calculator.CalculateROI(investment)
		if err == nil {
			t.Error("Expected error for zero initial cost, got nil")
		}
	})

	t.Run("no_benefits", func(t *testing.T) {
		investment := InvestmentData{
			InitialCost: 50000.0,
			Benefits:    []InvestmentBenefit{},
			Timeline:    12,
		}

		result, err := calculator.CalculateROI(investment)
		if err != nil {
			t.Fatalf("CalculateROI() with no benefits error = %v", err)
		}

		if result.AnnualROI != -100.0 {
			t.Errorf("Expected -100%% ROI for no benefits, got %v%%", result.AnnualROI)
		}
	})

	t.Run("very_high_discount_rate", func(t *testing.T) {
		investment := InvestmentData{
			InitialCost: 100000.0,
			Benefits: []InvestmentBenefit{
				{Type: "Revenue", Amount: 30000.0, Frequency: "Annual"},
			},
			Timeline:     36,
			DiscountRate: 0.50, // 50% discount rate
		}

		npv, err := calculator.CalculateNPV(investment)
		if err != nil {
			t.Fatalf("CalculateNPV() with high discount rate error = %v", err)
		}

		if npv > 0 {
			t.Error("Expected negative NPV with very high discount rate")
		}
	})

	t.Run("negative_timeline", func(t *testing.T) {
		investment := InvestmentData{
			InitialCost: 50000.0,
			Benefits: []InvestmentBenefit{
				{Type: "Revenue", Amount: 20000.0, Frequency: "Annual"},
			},
			Timeline: -12,
		}

		_, err := calculator.CalculateROI(investment)
		if err == nil {
			t.Error("Expected error for negative timeline, got nil")
		}
	})
}

func TestStandardROICalculator_SensitivityAnalysis(t *testing.T) {
	calculator := NewStandardROICalculator()

	baseInvestment := InvestmentData{
		InitialCost: 100000.0,
		Benefits: []InvestmentBenefit{
			{
				Type:      "Revenue Increase",
				Amount:    50000.0,
				Frequency: "Annual",
			},
		},
		Timeline:     24,
		DiscountRate: 0.08,
	}

	baseResult, err := calculator.CalculateROI(baseInvestment)
	if err != nil {
		t.Fatalf("Base CalculateROI() error = %v", err)
	}

	// Test sensitivity to benefit changes
	t.Run("benefit_sensitivity", func(t *testing.T) {
		modifiedInvestment := baseInvestment
		modifiedInvestment.Benefits[0].Amount = 60000.0 // 20% increase

		result, err := calculator.CalculateROI(modifiedInvestment)
		if err != nil {
			t.Fatalf("Modified CalculateROI() error = %v", err)
		}

		if result.AnnualROI <= baseResult.AnnualROI {
			t.Error("Expected higher ROI with increased benefits")
		}
	})

	// Test sensitivity to cost changes
	t.Run("cost_sensitivity", func(t *testing.T) {
		modifiedInvestment := baseInvestment
		modifiedInvestment.InitialCost = 120000.0 // 20% increase

		result, err := calculator.CalculateROI(modifiedInvestment)
		if err != nil {
			t.Fatalf("Modified CalculateROI() error = %v", err)
		}

		if result.AnnualROI >= baseResult.AnnualROI {
			t.Error("Expected lower ROI with increased initial cost")
		}
	})

	// Test sensitivity to discount rate changes
	t.Run("discount_rate_sensitivity", func(t *testing.T) {
		modifiedInvestment := baseInvestment
		modifiedInvestment.DiscountRate = 0.12 // Higher discount rate

		npvBase, _ := calculator.CalculateNPV(baseInvestment)
		npvModified, err := calculator.CalculateNPV(modifiedInvestment)
		if err != nil {
			t.Fatalf("Modified CalculateNPV() error = %v", err)
		}

		if npvModified >= npvBase {
			t.Error("Expected lower NPV with higher discount rate")
		}
	})
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}