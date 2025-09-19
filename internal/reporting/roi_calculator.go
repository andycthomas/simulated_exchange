package reporting

import (
	"fmt"
	"math"
	"time"
)

// StandardROICalculator implements ROICalculator interface
type StandardROICalculator struct {
	defaultDiscountRate float64
	inflationRate       float64
	riskAdjustment      float64
}

// NewStandardROICalculator creates a new ROI calculator
func NewStandardROICalculator() *StandardROICalculator {
	return &StandardROICalculator{
		defaultDiscountRate: 0.10, // 10% default discount rate
		inflationRate:       0.03, // 3% inflation rate
		riskAdjustment:      0.02, // 2% risk adjustment
	}
}

// CalculateROI computes return on investment metrics
func (src *StandardROICalculator) CalculateROI(investment InvestmentData) (*ROICalculation, error) {
	if investment.InitialCost <= 0 {
		return nil, fmt.Errorf("initial cost must be greater than zero")
	}

	if len(investment.ExpectedBenefits) == 0 {
		return nil, fmt.Errorf("expected benefits cannot be empty")
	}

	// Calculate annual benefits
	annualBenefits := src.calculateAnnualBenefits(investment)

	// Calculate cumulative benefits
	cumulativeBenefits := src.calculateCumulativeBenefits(annualBenefits)

	// Calculate payback period
	paybackPeriod := src.calculatePaybackPeriodMonths(investment.InitialCost, annualBenefits)

	// Calculate ROI for different periods
	threeYearROI := src.calculatePeriodROI(investment.InitialCost, annualBenefits, 3)
	fiveYearROI := src.calculatePeriodROI(investment.InitialCost, annualBenefits, 5)
	totalROI := src.calculateTotalROI(investment.InitialCost, cumulativeBenefits)

	// Calculate annual ROI
	annualROI := src.calculateAnnualROI(investment.InitialCost, annualBenefits)

	// Determine breakeven point
	breakevenPoint := src.calculateBreakevenPoint(investment, annualBenefits)

	// Additional investment metrics
	investmentMetrics := src.calculateAdditionalMetrics(investment, annualBenefits)

	return &ROICalculation{
		InitialInvestment: investment.InitialCost,
		AnnualBenefits:    annualBenefits,
		PaybackPeriod:     paybackPeriod,
		ThreeYearROI:      threeYearROI,
		FiveYearROI:       fiveYearROI,
		BreakevenPoint:    breakevenPoint,
		TotalROI:          totalROI,
		AnnualROI:         annualROI,
		InvestmentMetrics: investmentMetrics,
	}, nil
}

// CalculatePaybackPeriod determines investment payback timeframe
func (src *StandardROICalculator) CalculatePaybackPeriod(investment InvestmentData) (*PaybackAnalysis, error) {
	if investment.InitialCost <= 0 {
		return nil, fmt.Errorf("initial cost must be greater than zero")
	}

	annualBenefits := src.calculateAnnualBenefits(investment)
	monthlyBenefits := src.calculateMonthlyBenefits(annualBenefits)

	// Simple payback period
	simplePayback := src.calculateSimplePayback(investment.InitialCost, monthlyBenefits)

	// Discounted payback period
	discountedPayback := src.calculateDiscountedPayback(investment.InitialCost, monthlyBenefits, src.defaultDiscountRate)

	// Risk-adjusted payback period
	riskAdjustedPayback := src.calculateRiskAdjustedPayback(investment, monthlyBenefits)

	// Cash flow analysis
	cashFlowAnalysis := src.analyzeCashFlow(investment, monthlyBenefits)

	// Payback sensitivity analysis
	sensitivityAnalysis := src.calculatePaybackSensitivity(investment, monthlyBenefits)

	return &PaybackAnalysis{
		SimplePaybackMonths:      simplePayback,
		DiscountedPaybackMonths:  discountedPayback,
		RiskAdjustedPaybackMonths: riskAdjustedPayback,
		BreakevenAnalysis:        cashFlowAnalysis,
		SensitivityAnalysis:      sensitivityAnalysis,
		PaybackConfidence:        src.calculatePaybackConfidence(investment),
		MonthlyBreakdown:         src.createMonthlyBreakdown(investment, monthlyBenefits),
	}, nil
}

// CalculateNPV computes net present value
func (src *StandardROICalculator) CalculateNPV(investment InvestmentData, discountRate float64) (*NPVAnalysis, error) {
	if discountRate < 0 || discountRate > 1 {
		return nil, fmt.Errorf("discount rate must be between 0 and 1")
	}

	if investment.InitialCost <= 0 {
		return nil, fmt.Errorf("initial cost must be greater than zero")
	}

	annualBenefits := src.calculateAnnualBenefits(investment)
	annualCosts := src.calculateAnnualCosts(investment)

	// Calculate NPV
	npv := src.calculateNetPresentValue(investment.InitialCost, annualBenefits, annualCosts, discountRate)

	// Calculate profitability index
	profitabilityIndex := src.calculateProfitabilityIndex(investment.InitialCost, annualBenefits, discountRate)

	// Present value breakdown
	pvBreakdown := src.calculatePVBreakdown(annualBenefits, annualCosts, discountRate)

	// Sensitivity analysis
	sensitivityAnalysis := src.calculateNPVSensitivity(investment, annualBenefits, annualCosts, discountRate)

	// Risk analysis
	riskAnalysis := src.analyzeNPVRisk(investment, npv)

	return &NPVAnalysis{
		NetPresentValue:     npv,
		ProfitabilityIndex:  profitabilityIndex,
		DiscountRate:        discountRate,
		PresentValueBreakdown: pvBreakdown,
		SensitivityAnalysis: sensitivityAnalysis,
		RiskAnalysis:        riskAnalysis,
		Recommendation:      src.generateNPVRecommendation(npv, profitabilityIndex),
	}, nil
}

// ProjectCashFlow forecasts cash flow over specified period
func (src *StandardROICalculator) ProjectCashFlow(investment InvestmentData, years int) (*CashFlowProjection, error) {
	if years <= 0 {
		return nil, fmt.Errorf("projection period must be greater than zero")
	}

	// Calculate annual cash flows
	annualCashFlows := src.calculateAnnualCashFlows(investment, years)

	// Calculate monthly cash flows for first year
	monthlyCashFlows := src.calculateMonthlyCashFlows(investment)

	// Calculate cumulative cash flows
	cumulativeCashFlows := src.calculateCumulativeCashFlows(annualCashFlows)

	// Free cash flow analysis
	freeCashFlow := src.calculateFreeCashFlow(investment, annualCashFlows)

	// Cash flow ratios
	cashFlowRatios := src.calculateCashFlowRatios(annualCashFlows, investment)

	// Seasonal analysis
	seasonalAnalysis := src.analyzeSeasonality(investment, monthlyCashFlows)

	// Risk factors
	riskFactors := src.identifyCashFlowRisks(investment)

	return &CashFlowProjection{
		ProjectionPeriodYears: years,
		AnnualCashFlows:      annualCashFlows,
		MonthlyCashFlows:     monthlyCashFlows,
		CumulativeCashFlows:  cumulativeCashFlows,
		FreeCashFlow:         freeCashFlow,
		CashFlowRatios:       cashFlowRatios,
		SeasonalAnalysis:     seasonalAnalysis,
		RiskFactors:          riskFactors,
		ProjectionConfidence: src.calculateProjectionConfidence(investment),
	}, nil
}

// CalculateIRR computes internal rate of return
func (src *StandardROICalculator) CalculateIRR(investment InvestmentData) (*IRRAnalysis, error) {
	if investment.InitialCost <= 0 {
		return nil, fmt.Errorf("initial cost must be greater than zero")
	}

	annualBenefits := src.calculateAnnualBenefits(investment)
	annualCosts := src.calculateAnnualCosts(investment)

	// Calculate net cash flows
	netCashFlows := src.calculateNetCashFlows(investment.InitialCost, annualBenefits, annualCosts)

	// Calculate IRR using Newton-Raphson method
	irr := src.calculateInternalRateOfReturn(netCashFlows)

	// Calculate Modified IRR (MIRR)
	mirr := src.calculateModifiedIRR(netCashFlows, src.defaultDiscountRate)

	// Risk-adjusted IRR
	riskAdjustedIRR := src.calculateRiskAdjustedIRR(irr, investment)

	// IRR sensitivity analysis
	sensitivityAnalysis := src.calculateIRRSensitivity(investment, netCashFlows)

	// Benchmark comparison
	benchmarkComparison := src.compareWithBenchmarks(irr, investment.InvestmentType)

	return &IRRAnalysis{
		InternalRateOfReturn: irr,
		ModifiedIRR:         mirr,
		RiskAdjustedIRR:     riskAdjustedIRR,
		BenchmarkComparison: benchmarkComparison,
		SensitivityAnalysis: sensitivityAnalysis,
		ConfidenceLevel:     src.calculateIRRConfidence(investment),
		Recommendation:      src.generateIRRRecommendation(irr, riskAdjustedIRR),
	}, nil
}

// Helper methods for calculations

func (src *StandardROICalculator) calculateAnnualBenefits(investment InvestmentData) []float64 {
	maxYears := 5 // Default to 5-year analysis
	annualBenefits := make([]float64, maxYears)

	// Aggregate benefits by year
	for _, benefit := range investment.ExpectedBenefits {
		year := benefit.Year
		if year > 0 && year <= maxYears {
			annualBenefits[year-1] += benefit.Amount
		}
	}

	// Apply growth rates and adjustments
	for i := 1; i < maxYears; i++ {
		if annualBenefits[i] == 0 && annualBenefits[i-1] > 0 {
			// If no specific benefit for this year, apply growth
			growthRate := src.estimateGrowthRate(investment)
			annualBenefits[i] = annualBenefits[i-1] * (1 + growthRate)
		}
	}

	return annualBenefits
}

func (src *StandardROICalculator) calculateAnnualCosts(investment InvestmentData) []float64 {
	maxYears := 5
	annualCosts := make([]float64, maxYears)

	// Aggregate ongoing costs by year
	for _, cost := range investment.OngoingCosts {
		year := cost.Year
		if year > 0 && year <= maxYears {
			annualCosts[year-1] += cost.Amount
		}
	}

	// Apply inflation to costs
	for i := 1; i < maxYears; i++ {
		if annualCosts[i] == 0 && annualCosts[i-1] > 0 {
			annualCosts[i] = annualCosts[i-1] * (1 + src.inflationRate)
		}
	}

	return annualCosts
}

func (src *StandardROICalculator) calculateCumulativeBenefits(annualBenefits []float64) []float64 {
	cumulative := make([]float64, len(annualBenefits))
	sum := 0.0
	for i, benefit := range annualBenefits {
		sum += benefit
		cumulative[i] = sum
	}
	return cumulative
}

func (src *StandardROICalculator) calculatePaybackPeriodMonths(initialCost float64, annualBenefits []float64) float64 {
	monthlyBenefits := src.calculateMonthlyBenefits(annualBenefits)

	cumulativeBenefit := 0.0
	for month, benefit := range monthlyBenefits {
		cumulativeBenefit += benefit
		if cumulativeBenefit >= initialCost {
			// Linear interpolation for more precise calculation
			prevCumulative := cumulativeBenefit - benefit
			ratio := (initialCost - prevCumulative) / benefit
			return float64(month) + ratio
		}
	}

	// If payback not achieved in calculated period
	return -1 // Indicates payback period exceeds projection
}

func (src *StandardROICalculator) calculateMonthlyBenefits(annualBenefits []float64) []float64 {
	totalMonths := len(annualBenefits) * 12
	monthlyBenefits := make([]float64, totalMonths)

	for year, annualBenefit := range annualBenefits {
		monthlyBenefit := annualBenefit / 12.0
		for month := 0; month < 12; month++ {
			monthlyBenefits[year*12+month] = monthlyBenefit
		}
	}

	return monthlyBenefits
}

func (src *StandardROICalculator) calculatePeriodROI(initialCost float64, annualBenefits []float64, years int) float64 {
	if years > len(annualBenefits) {
		years = len(annualBenefits)
	}

	totalBenefits := 0.0
	for i := 0; i < years; i++ {
		totalBenefits += annualBenefits[i]
	}

	if initialCost == 0 {
		return 0
	}

	return ((totalBenefits - initialCost) / initialCost) * 100
}

func (src *StandardROICalculator) calculateTotalROI(initialCost float64, cumulativeBenefits []float64) float64 {
	if len(cumulativeBenefits) == 0 || initialCost == 0 {
		return 0
	}

	totalBenefits := cumulativeBenefits[len(cumulativeBenefits)-1]
	return ((totalBenefits - initialCost) / initialCost) * 100
}

func (src *StandardROICalculator) calculateAnnualROI(initialCost float64, annualBenefits []float64) float64 {
	if len(annualBenefits) == 0 || initialCost == 0 {
		return 0
	}

	// Calculate average annual benefit
	totalBenefits := 0.0
	for _, benefit := range annualBenefits {
		totalBenefits += benefit
	}
	avgAnnualBenefit := totalBenefits / float64(len(annualBenefits))

	// Annual ROI = (Average Annual Benefit / Initial Investment) * 100
	return (avgAnnualBenefit / initialCost) * 100
}

func (src *StandardROICalculator) calculateBreakevenPoint(investment InvestmentData, annualBenefits []float64) time.Time {
	startDate := time.Now()
	if len(investment.ImplementationPlan.Phases) > 0 {
		startDate = investment.ImplementationPlan.Phases[0].StartDate
	}

	paybackMonths := src.calculatePaybackPeriodMonths(investment.InitialCost, annualBenefits)
	if paybackMonths < 0 {
		// Return far future date if breakeven not achieved
		return startDate.AddDate(10, 0, 0)
	}

	months := int(paybackMonths)
	days := int((paybackMonths - float64(months)) * 30)

	return startDate.AddDate(0, months, days)
}

func (src *StandardROICalculator) calculateNetPresentValue(initialCost float64, annualBenefits, annualCosts []float64, discountRate float64) float64 {
	npv := -initialCost // Initial investment is negative cash flow

	for year, benefit := range annualBenefits {
		cost := 0.0
		if year < len(annualCosts) {
			cost = annualCosts[year]
		}

		netCashFlow := benefit - cost
		discountFactor := math.Pow(1+discountRate, float64(year+1))
		npv += netCashFlow / discountFactor
	}

	return npv
}

func (src *StandardROICalculator) calculateProfitabilityIndex(initialCost float64, annualBenefits []float64, discountRate float64) float64 {
	presentValueBenefits := 0.0

	for year, benefit := range annualBenefits {
		discountFactor := math.Pow(1+discountRate, float64(year+1))
		presentValueBenefits += benefit / discountFactor
	}

	if initialCost == 0 {
		return 0
	}

	return presentValueBenefits / initialCost
}

func (src *StandardROICalculator) calculateInternalRateOfReturn(netCashFlows []float64) float64 {
	// Newton-Raphson method for IRR calculation
	rate := 0.1 // Initial guess: 10%
	tolerance := 0.0001
	maxIterations := 100

	for i := 0; i < maxIterations; i++ {
		npv := 0.0
		derivative := 0.0

		for period, cashFlow := range netCashFlows {
			factor := math.Pow(1+rate, float64(period))
			npv += cashFlow / factor
			if period > 0 {
				derivative -= float64(period) * cashFlow / (factor * (1 + rate))
			}
		}

		if math.Abs(npv) < tolerance {
			break
		}

		if derivative == 0 {
			break
		}

		rate = rate - npv/derivative

		// Prevent negative rates
		if rate < -0.99 {
			rate = -0.99
		}
	}

	return rate * 100 // Return as percentage
}

func (src *StandardROICalculator) estimateGrowthRate(investment InvestmentData) float64 {
	// Default growth rate based on investment type
	switch investment.InvestmentType {
	case InvestmentTechnology:
		return 0.15 // 15% growth for technology investments
	case InvestmentMarketing:
		return 0.10 // 10% growth for marketing investments
	case InvestmentOperational:
		return 0.08 // 8% growth for operational improvements
	default:
		return 0.05 // 5% default growth
	}
}

func (src *StandardROICalculator) calculateAdditionalMetrics(investment InvestmentData, annualBenefits []float64) map[string]interface{} {
	metrics := make(map[string]interface{})

	// Risk-adjusted return
	riskScore := src.calculateInvestmentRisk(investment)
	riskAdjustment := 1.0 - (riskScore * 0.1) // Reduce returns based on risk

	totalBenefits := 0.0
	for _, benefit := range annualBenefits {
		totalBenefits += benefit
	}

	metrics["risk_score"] = riskScore
	metrics["risk_adjusted_roi"] = src.calculateTotalROI(investment.InitialCost, []float64{totalBenefits * riskAdjustment})
	metrics["benefit_cost_ratio"] = totalBenefits / investment.InitialCost
	metrics["investment_efficiency"] = src.calculateInvestmentEfficiency(investment, annualBenefits)
	metrics["payback_ratio"] = investment.InitialCost / (totalBenefits / float64(len(annualBenefits)))

	return metrics
}

func (src *StandardROICalculator) calculateInvestmentRisk(investment InvestmentData) float64 {
	// Simple risk scoring based on various factors
	riskScore := 0.0

	// Risk based on investment type
	switch investment.InvestmentType {
	case InvestmentTechnology:
		riskScore += 0.3
	case InvestmentMarketing:
		riskScore += 0.4
	case InvestmentOperational:
		riskScore += 0.2
	default:
		riskScore += 0.25
	}

	// Risk based on investment size
	if investment.InitialCost > 1000000 {
		riskScore += 0.2
	} else if investment.InitialCost > 100000 {
		riskScore += 0.1
	}

	// Risk based on implementation complexity
	if len(investment.ImplementationPlan.Phases) > 5 {
		riskScore += 0.1
	}

	// Risk based on number of risk factors
	riskScore += float64(len(investment.RiskFactors)) * 0.05

	return math.Min(1.0, riskScore)
}

func (src *StandardROICalculator) calculateInvestmentEfficiency(investment InvestmentData, annualBenefits []float64) float64 {
	if len(annualBenefits) == 0 || investment.InitialCost == 0 {
		return 0
	}

	totalBenefits := 0.0
	for _, benefit := range annualBenefits {
		totalBenefits += benefit
	}

	// Efficiency = Total Benefits / (Initial Cost + Complexity Factor)
	complexityFactor := float64(len(investment.ImplementationPlan.Phases)) * 1000
	return totalBenefits / (investment.InitialCost + complexityFactor)
}

// Additional supporting data structures

type ExpectedBenefit struct {
	Year        int     `json:"year"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

type OngoingCost struct {
	Year        int     `json:"year"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	IsRecurring bool    `json:"is_recurring"`
}

type ImplementationPlan struct {
	Phases      []ImplementationPhase `json:"phases"`
	TotalDuration time.Duration       `json:"total_duration"`
	Dependencies []string             `json:"dependencies"`
	Resources    []string             `json:"resources"`
}

type ImplementationPhase struct {
	Name        string    `json:"name"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Cost        float64   `json:"cost"`
	Description string    `json:"description"`
	Deliverables []string `json:"deliverables"`
}

type PaybackAnalysis struct {
	SimplePaybackMonths       float64                `json:"simple_payback_months"`
	DiscountedPaybackMonths   float64                `json:"discounted_payback_months"`
	RiskAdjustedPaybackMonths float64                `json:"risk_adjusted_payback_months"`
	BreakevenAnalysis         map[string]interface{} `json:"breakeven_analysis"`
	SensitivityAnalysis       map[string]interface{} `json:"sensitivity_analysis"`
	PaybackConfidence         float64                `json:"payback_confidence"`
	MonthlyBreakdown          []MonthlyBreakdown     `json:"monthly_breakdown"`
}

type NPVAnalysis struct {
	NetPresentValue       float64                `json:"net_present_value"`
	ProfitabilityIndex    float64                `json:"profitability_index"`
	DiscountRate          float64                `json:"discount_rate"`
	PresentValueBreakdown map[string]interface{} `json:"present_value_breakdown"`
	SensitivityAnalysis   map[string]interface{} `json:"sensitivity_analysis"`
	RiskAnalysis          map[string]interface{} `json:"risk_analysis"`
	Recommendation        string                 `json:"recommendation"`
}

type IRRAnalysis struct {
	InternalRateOfReturn float64                `json:"internal_rate_of_return"`
	ModifiedIRR          float64                `json:"modified_irr"`
	RiskAdjustedIRR      float64                `json:"risk_adjusted_irr"`
	BenchmarkComparison  map[string]interface{} `json:"benchmark_comparison"`
	SensitivityAnalysis  map[string]interface{} `json:"sensitivity_analysis"`
	ConfidenceLevel      float64                `json:"confidence_level"`
	Recommendation       string                 `json:"recommendation"`
}

type CashFlowProjection struct {
	ProjectionPeriodYears int                    `json:"projection_period_years"`
	AnnualCashFlows      []float64              `json:"annual_cash_flows"`
	MonthlyCashFlows     []float64              `json:"monthly_cash_flows"`
	CumulativeCashFlows  []float64              `json:"cumulative_cash_flows"`
	FreeCashFlow         []float64              `json:"free_cash_flow"`
	CashFlowRatios       map[string]interface{} `json:"cash_flow_ratios"`
	SeasonalAnalysis     map[string]interface{} `json:"seasonal_analysis"`
	RiskFactors          []string               `json:"risk_factors"`
	ProjectionConfidence float64                `json:"projection_confidence"`
}

type MonthlyBreakdown struct {
	Month              int     `json:"month"`
	MonthlyCashFlow    float64 `json:"monthly_cash_flow"`
	CumulativeCashFlow float64 `json:"cumulative_cash_flow"`
	ROIToDate          float64 `json:"roi_to_date"`
	BreakevenStatus    string  `json:"breakeven_status"`
}

// Missing helper methods for ROI calculator

func (src *StandardROICalculator) calculateSimplePayback(initialCost float64, monthlyBenefits []float64) float64 {
	cumulativeBenefit := 0.0
	for month, benefit := range monthlyBenefits {
		cumulativeBenefit += benefit
		if cumulativeBenefit >= initialCost {
			// Linear interpolation for more precise calculation
			prevCumulative := cumulativeBenefit - benefit
			ratio := (initialCost - prevCumulative) / benefit
			return float64(month) + ratio
		}
	}
	return -1 // Payback not achieved in projection period
}

func (src *StandardROICalculator) calculateDiscountedPayback(initialCost float64, monthlyBenefits []float64, discountRate float64) float64 {
	cumulativePV := 0.0
	monthlyRate := discountRate / 12.0

	for month, benefit := range monthlyBenefits {
		discountFactor := math.Pow(1+monthlyRate, float64(month+1))
		presentValue := benefit / discountFactor
		cumulativePV += presentValue

		if cumulativePV >= initialCost {
			prevCumulative := cumulativePV - presentValue
			ratio := (initialCost - prevCumulative) / presentValue
			return float64(month) + ratio
		}
	}
	return -1
}

func (src *StandardROICalculator) calculateRiskAdjustedPayback(investment InvestmentData, monthlyBenefits []float64) float64 {
	riskScore := src.calculateInvestmentRisk(investment)
	riskAdjustment := 1.0 - (riskScore * 0.2) // Reduce benefits by risk factor

	adjustedBenefits := make([]float64, len(monthlyBenefits))
	for i, benefit := range monthlyBenefits {
		adjustedBenefits[i] = benefit * riskAdjustment
	}

	return src.calculateSimplePayback(investment.InitialCost, adjustedBenefits)
}

func (src *StandardROICalculator) analyzeCashFlow(investment InvestmentData, monthlyBenefits []float64) map[string]interface{} {
	analysis := make(map[string]interface{})

	totalBenefits := 0.0
	minMonthlyBenefit := math.Inf(1)
	maxMonthlyBenefit := math.Inf(-1)

	for _, benefit := range monthlyBenefits {
		totalBenefits += benefit
		if benefit < minMonthlyBenefit {
			minMonthlyBenefit = benefit
		}
		if benefit > maxMonthlyBenefit {
			maxMonthlyBenefit = benefit
		}
	}

	avgMonthlyBenefit := totalBenefits / float64(len(monthlyBenefits))

	analysis["total_benefits"] = totalBenefits
	analysis["average_monthly_benefit"] = avgMonthlyBenefit
	analysis["min_monthly_benefit"] = minMonthlyBenefit
	analysis["max_monthly_benefit"] = maxMonthlyBenefit
	analysis["benefit_volatility"] = (maxMonthlyBenefit - minMonthlyBenefit) / avgMonthlyBenefit
	analysis["cash_flow_stability"] = avgMonthlyBenefit / maxMonthlyBenefit

	return analysis
}

func (src *StandardROICalculator) calculatePaybackSensitivity(investment InvestmentData, monthlyBenefits []float64) map[string]interface{} {
	sensitivity := make(map[string]interface{})

	basePayback := src.calculateSimplePayback(investment.InitialCost, monthlyBenefits)

	// Test sensitivity to benefit changes
	scenarios := []float64{0.8, 0.9, 1.1, 1.2} // -20%, -10%, +10%, +20%
	scenarioResults := make(map[string]float64)

	for _, multiplier := range scenarios {
		adjustedBenefits := make([]float64, len(monthlyBenefits))
		for i, benefit := range monthlyBenefits {
			adjustedBenefits[i] = benefit * multiplier
		}
		payback := src.calculateSimplePayback(investment.InitialCost, adjustedBenefits)
		scenarioResults[fmt.Sprintf("%.1fx", multiplier)] = payback
	}

	sensitivity["base_payback"] = basePayback
	sensitivity["scenarios"] = scenarioResults
	sensitivity["sensitivity_index"] = (scenarioResults["1.2x"] - scenarioResults["0.8x"]) / basePayback

	return sensitivity
}

func (src *StandardROICalculator) calculatePaybackConfidence(investment InvestmentData) float64 {
	confidence := 0.8 // Base confidence

	// Adjust confidence based on risk factors
	riskScore := src.calculateInvestmentRisk(investment)
	confidence -= riskScore * 0.3

	// Adjust confidence based on investment type
	switch investment.InvestmentType {
	case InvestmentTechnology:
		confidence -= 0.1 // Technology has higher uncertainty
	case InvestmentOperational:
		confidence += 0.1 // Operational improvements are more predictable
	}

	// Ensure confidence stays within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

func (src *StandardROICalculator) createMonthlyBreakdown(investment InvestmentData, monthlyBenefits []float64) []MonthlyBreakdown {
	breakdown := make([]MonthlyBreakdown, len(monthlyBenefits))
	cumulativeCashFlow := 0.0

	for i, benefit := range monthlyBenefits {
		cumulativeCashFlow += benefit

		roiToDate := 0.0
		if investment.InitialCost > 0 {
			roiToDate = ((cumulativeCashFlow - investment.InitialCost) / investment.InitialCost) * 100
		}

		breakevenStatus := "Below Breakeven"
		if cumulativeCashFlow >= investment.InitialCost {
			breakevenStatus = "Above Breakeven"
		}

		breakdown[i] = MonthlyBreakdown{
			Month:              i + 1,
			MonthlyCashFlow:    benefit,
			CumulativeCashFlow: cumulativeCashFlow,
			ROIToDate:          roiToDate,
			BreakevenStatus:    breakevenStatus,
		}
	}

	return breakdown
}

func (src *StandardROICalculator) calculatePVBreakdown(annualBenefits, annualCosts []float64, discountRate float64) map[string]interface{} {
	breakdown := make(map[string]interface{})

	totalPVBenefits := 0.0
	totalPVCosts := 0.0
	yearlyPV := make([]map[string]float64, len(annualBenefits))

	for year, benefit := range annualBenefits {
		cost := 0.0
		if year < len(annualCosts) {
			cost = annualCosts[year]
		}

		discountFactor := math.Pow(1+discountRate, float64(year+1))
		pvBenefit := benefit / discountFactor
		pvCost := cost / discountFactor

		totalPVBenefits += pvBenefit
		totalPVCosts += pvCost

		yearlyPV[year] = map[string]float64{
			"year":        float64(year + 1),
			"pv_benefit":  pvBenefit,
			"pv_cost":     pvCost,
			"net_pv":      pvBenefit - pvCost,
		}
	}

	breakdown["total_pv_benefits"] = totalPVBenefits
	breakdown["total_pv_costs"] = totalPVCosts
	breakdown["net_pv"] = totalPVBenefits - totalPVCosts
	breakdown["yearly_breakdown"] = yearlyPV

	return breakdown
}

func (src *StandardROICalculator) calculateNPVSensitivity(investment InvestmentData, annualBenefits, annualCosts []float64, discountRate float64) map[string]interface{} {
	sensitivity := make(map[string]interface{})

	baseNPV := src.calculateNetPresentValue(investment.InitialCost, annualBenefits, annualCosts, discountRate)

	// Test different discount rates
	rates := []float64{discountRate - 0.02, discountRate - 0.01, discountRate + 0.01, discountRate + 0.02}
	rateResults := make(map[string]float64)

	for _, rate := range rates {
		if rate > 0 {
			npv := src.calculateNetPresentValue(investment.InitialCost, annualBenefits, annualCosts, rate)
			rateResults[fmt.Sprintf("%.1f%%", rate*100)] = npv
		}
	}

	sensitivity["base_npv"] = baseNPV
	sensitivity["rate_sensitivity"] = rateResults

	return sensitivity
}

func (src *StandardROICalculator) analyzeNPVRisk(investment InvestmentData, npv float64) map[string]interface{} {
	risk := make(map[string]interface{})

	riskScore := src.calculateInvestmentRisk(investment)

	risk["risk_score"] = riskScore
	risk["npv_at_risk"] = npv * riskScore
	risk["risk_adjusted_npv"] = npv * (1 - riskScore*0.3)

	if npv > 0 {
		risk["recommendation"] = "Positive NPV indicates good investment"
	} else {
		risk["recommendation"] = "Negative NPV suggests reconsidering investment"
	}

	return risk
}

func (src *StandardROICalculator) generateNPVRecommendation(npv, profitabilityIndex float64) string {
	if npv > 0 && profitabilityIndex > 1.0 {
		return "Strong recommendation: Positive NPV and profitable"
	} else if npv > 0 {
		return "Moderate recommendation: Positive NPV"
	} else {
		return "Not recommended: Negative NPV"
	}
}

func (src *StandardROICalculator) calculateAnnualCashFlows(investment InvestmentData, years int) []float64 {
	cashFlows := make([]float64, years)
	annualBenefits := src.calculateAnnualBenefits(investment)
	annualCosts := src.calculateAnnualCosts(investment)

	for i := 0; i < years; i++ {
		benefit := 0.0
		cost := 0.0

		if i < len(annualBenefits) {
			benefit = annualBenefits[i]
		}
		if i < len(annualCosts) {
			cost = annualCosts[i]
		}

		cashFlows[i] = benefit - cost
	}

	return cashFlows
}

func (src *StandardROICalculator) calculateMonthlyCashFlows(investment InvestmentData) []float64 {
	annualBenefits := src.calculateAnnualBenefits(investment)
	return src.calculateMonthlyBenefits(annualBenefits)
}

func (src *StandardROICalculator) calculateCumulativeCashFlows(annualCashFlows []float64) []float64 {
	cumulative := make([]float64, len(annualCashFlows))
	sum := 0.0

	for i, cashFlow := range annualCashFlows {
		sum += cashFlow
		cumulative[i] = sum
	}

	return cumulative
}

func (src *StandardROICalculator) calculateFreeCashFlow(investment InvestmentData, annualCashFlows []float64) []float64 {
	// Simplified free cash flow calculation
	freeCashFlow := make([]float64, len(annualCashFlows))

	for i, cashFlow := range annualCashFlows {
		// Assume 10% of cash flow goes to capital expenditures
		capex := cashFlow * 0.1
		freeCashFlow[i] = cashFlow - capex
	}

	return freeCashFlow
}

func (src *StandardROICalculator) calculateCashFlowRatios(annualCashFlows []float64, investment InvestmentData) map[string]interface{} {
	ratios := make(map[string]interface{})

	if len(annualCashFlows) > 0 {
		totalCashFlow := 0.0
		for _, cf := range annualCashFlows {
			totalCashFlow += cf
		}

		ratios["cash_flow_to_investment"] = totalCashFlow / investment.InitialCost
		ratios["average_annual_cash_flow"] = totalCashFlow / float64(len(annualCashFlows))
		ratios["cash_flow_stability"] = src.calculateStability(annualCashFlows)
	}

	return ratios
}

func (src *StandardROICalculator) analyzeSeasonality(investment InvestmentData, monthlyCashFlows []float64) map[string]interface{} {
	seasonality := make(map[string]interface{})

	if len(monthlyCashFlows) >= 12 {
		quarterly := make([]float64, 4)
		for i, cf := range monthlyCashFlows[:12] {
			quarter := i / 3
			quarterly[quarter] += cf
		}

		maxQuarter := 0.0
		minQuarter := math.Inf(1)
		for _, q := range quarterly {
			if q > maxQuarter {
				maxQuarter = q
			}
			if q < minQuarter {
				minQuarter = q
			}
		}

		seasonality["has_seasonality"] = (maxQuarter - minQuarter) / maxQuarter > 0.2
		seasonality["seasonal_variation"] = (maxQuarter - minQuarter) / maxQuarter
		seasonality["quarterly_breakdown"] = quarterly
	}

	return seasonality
}

func (src *StandardROICalculator) identifyCashFlowRisks(investment InvestmentData) []string {
	risks := []string{}

	// Common cash flow risks
	risks = append(risks, "Market volatility affecting revenue")
	risks = append(risks, "Implementation delays")
	risks = append(risks, "Cost overruns")

	// Investment-specific risks
	switch investment.InvestmentType {
	case InvestmentTechnology:
		risks = append(risks, "Technology obsolescence", "Integration challenges")
	case InvestmentMarketing:
		risks = append(risks, "Campaign effectiveness uncertainty", "Market response variability")
	case InvestmentOperational:
		risks = append(risks, "Process adoption resistance", "Operational disruption")
	}

	return risks
}

func (src *StandardROICalculator) calculateProjectionConfidence(investment InvestmentData) float64 {
	confidence := 0.8

	// Adjust based on investment characteristics
	riskScore := src.calculateInvestmentRisk(investment)
	confidence -= riskScore * 0.3

	// Adjust based on data quality
	if len(investment.ExpectedBenefits) < 3 {
		confidence -= 0.1
	}

	return math.Max(0.1, math.Min(1.0, confidence))
}

func (src *StandardROICalculator) calculateNetCashFlows(initialCost float64, annualBenefits, annualCosts []float64) []float64 {
	maxLen := len(annualBenefits)
	if len(annualCosts) > maxLen {
		maxLen = len(annualCosts)
	}

	netCashFlows := make([]float64, maxLen+1)
	netCashFlows[0] = -initialCost

	for i := 0; i < maxLen; i++ {
		benefit := 0.0
		cost := 0.0

		if i < len(annualBenefits) {
			benefit = annualBenefits[i]
		}
		if i < len(annualCosts) {
			cost = annualCosts[i]
		}

		netCashFlows[i+1] = benefit - cost
	}

	return netCashFlows
}

func (src *StandardROICalculator) calculateModifiedIRR(netCashFlows []float64, reinvestmentRate float64) float64 {
	if len(netCashFlows) < 2 {
		return 0
	}

	// Separate positive and negative cash flows
	pv := math.Abs(netCashFlows[0]) // Initial investment (negative)
	fv := 0.0

	// Calculate future value of positive cash flows
	for i := 1; i < len(netCashFlows); i++ {
		if netCashFlows[i] > 0 {
			years := float64(len(netCashFlows) - 1 - i)
			fv += netCashFlows[i] * math.Pow(1+reinvestmentRate, years)
		}
	}

	if pv == 0 {
		return 0
	}

	// Calculate MIRR
	n := float64(len(netCashFlows) - 1)
	mirr := math.Pow(fv/pv, 1/n) - 1

	return mirr * 100 // Return as percentage
}

func (src *StandardROICalculator) calculateRiskAdjustedIRR(irr float64, investment InvestmentData) float64 {
	riskScore := src.calculateInvestmentRisk(investment)
	riskPremium := riskScore * 5.0 // 5% risk premium per risk point

	return irr - riskPremium
}

func (src *StandardROICalculator) calculateIRRSensitivity(investment InvestmentData, netCashFlows []float64) map[string]interface{} {
	sensitivity := make(map[string]interface{})

	baseIRR := src.calculateInternalRateOfReturn(netCashFlows)

	// Test sensitivity to cash flow changes
	scenarios := []float64{0.8, 0.9, 1.1, 1.2}
	scenarioResults := make(map[string]float64)

	for _, multiplier := range scenarios {
		adjustedCashFlows := make([]float64, len(netCashFlows))
		adjustedCashFlows[0] = netCashFlows[0] // Keep initial investment unchanged

		for i := 1; i < len(netCashFlows); i++ {
			adjustedCashFlows[i] = netCashFlows[i] * multiplier
		}

		irr := src.calculateInternalRateOfReturn(adjustedCashFlows)
		scenarioResults[fmt.Sprintf("%.1fx", multiplier)] = irr
	}

	sensitivity["base_irr"] = baseIRR
	sensitivity["scenarios"] = scenarioResults

	return sensitivity
}

func (src *StandardROICalculator) compareWithBenchmarks(irr float64, investmentType InvestmentType) map[string]interface{} {
	comparison := make(map[string]interface{})

	// Industry benchmark IRRs
	benchmarks := map[InvestmentType]float64{
		InvestmentTechnology:     15.0,
		InvestmentMarketing:      12.0,
		InvestmentOperational:    10.0,
		InvestmentInfrastructure: 8.0,
		InvestmentPersonnel:      7.0,
		InvestmentStrategic:      20.0,
	}

	benchmark := benchmarks[investmentType]
	if benchmark == 0 {
		benchmark = 10.0 // Default benchmark
	}

	comparison["benchmark_irr"] = benchmark
	comparison["performance_vs_benchmark"] = irr - benchmark

	if irr > benchmark {
		comparison["performance_rating"] = "Above Benchmark"
	} else if irr > benchmark*0.9 {
		comparison["performance_rating"] = "Near Benchmark"
	} else {
		comparison["performance_rating"] = "Below Benchmark"
	}

	return comparison
}

func (src *StandardROICalculator) calculateIRRConfidence(investment InvestmentData) float64 {
	return src.calculatePaybackConfidence(investment) // Reuse similar logic
}

func (src *StandardROICalculator) generateIRRRecommendation(irr, riskAdjustedIRR float64) string {
	if riskAdjustedIRR > 15.0 {
		return "Highly recommended: Excellent returns even after risk adjustment"
	} else if riskAdjustedIRR > 10.0 {
		return "Recommended: Good returns considering risk"
	} else if riskAdjustedIRR > 5.0 {
		return "Moderate recommendation: Acceptable returns but monitor risks"
	} else {
		return "Not recommended: Returns too low relative to risk"
	}
}

func (src *StandardROICalculator) calculateStability(values []float64) float64 {
	if len(values) < 2 {
		return 1.0
	}

	mean := 0.0
	for _, v := range values {
		mean += v
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values))

	stdDev := math.Sqrt(variance)

	if mean == 0 {
		return 0.0
	}

	coefficientOfVariation := stdDev / math.Abs(mean)
	stability := 1.0 / (1.0 + coefficientOfVariation)

	return stability
}