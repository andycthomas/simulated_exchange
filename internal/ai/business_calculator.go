package ai

import (
	"fmt"
	"math"
	"time"

	"simulated_exchange/internal/metrics"
)
// ROICalculator implements BusinessImpactCalculator with sophisticated financial analysis
type ROICalculator struct {
	config           ROIConfig
	marketParameters MarketParameters
	costModel        CostModel
}

// ROIConfig holds configuration for ROI calculations
type ROIConfig struct {
	DiscountRate          float64 `json:"discount_rate"`           // Annual discount rate for NPV
	RiskFreeRate          float64 `json:"risk_free_rate"`          // Risk-free rate for calculations
	MarketRiskPremium     float64 `json:"market_risk_premium"`     // Market risk premium
	ProjectRiskMultiplier float64 `json:"project_risk_multiplier"` // Project-specific risk multiplier
	TaxRate               float64 `json:"tax_rate"`                // Corporate tax rate
	InflationRate         float64 `json:"inflation_rate"`          // Annual inflation rate
}

// MarketParameters defines market-specific parameters for impact calculation
type MarketParameters struct {
	RevenuePerOrder       float64 `json:"revenue_per_order"`        // Average revenue per order
	RevenuePerTrade       float64 `json:"revenue_per_trade"`        // Average revenue per trade
	CostPerLatencyMs      float64 `json:"cost_per_latency_ms"`      // Cost impact per ms of latency
	RevenueAtRiskPercent  float64 `json:"revenue_at_risk_percent"`  // Percentage of revenue at risk
	CustomerLTVMultiplier float64 `json:"customer_ltv_multiplier"`  // Customer lifetime value multiplier
	CompetitiveImpact     float64 `json:"competitive_impact"`       // Impact on competitive position
}

// CostModel defines cost structure for different optimization activities
type CostModel struct {
	HourlyDeveloperCost    float64            `json:"hourly_developer_cost"`
	HourlyInfraCost        float64            `json:"hourly_infra_cost"`
	ComputeUnitCost        float64            `json:"compute_unit_cost"`         // Annual cost per compute unit
	MemoryGBCost           float64            `json:"memory_gb_cost"`            // Annual cost per GB memory
	StorageGBCost          float64            `json:"storage_gb_cost"`           // Annual cost per GB storage
	NetworkBandwidthCost   float64            `json:"network_bandwidth_cost"`    // Annual cost per Mbps
	DatabaseIOPSCost       float64            `json:"database_iops_cost"`        // Annual cost per IOPS
	MaintenanceMultiplier  float64            `json:"maintenance_multiplier"`    // Maintenance cost as multiplier
	ComplexityCostFactors  map[Complexity]float64 `json:"complexity_cost_factors"` // Cost factors by complexity
}

// NewROICalculator creates a new ROI calculator
func NewROICalculator(config ROIConfig, marketParams MarketParameters, costModel CostModel) *ROICalculator {
	return &ROICalculator{
		config:           config,
		marketParameters: marketParams,
		costModel:        costModel,
	}
}

// NewDefaultROICalculator creates calculator with default parameters
func NewDefaultROICalculator() *ROICalculator {
	return NewROICalculator(
		DefaultROIConfig(),
		DefaultMarketParameters(),
		DefaultCostModel(),
	)
}

// CalculateROI computes return on investment for performance improvements
func (calc *ROICalculator) CalculateROI(recommendation Recommendation, currentMetrics metrics.MetricsSnapshot) ROIAnalysis {
	// Calculate implementation cost
	implementationCost := calc.calculateImplementationCost(recommendation)

	// Calculate annual savings
	annualSavings := calc.calculateAnnualSavings(recommendation, currentMetrics)

	// Calculate payback period
	paybackPeriod := calc.calculatePaybackPeriod(implementationCost, annualSavings)

	// Calculate ROI percentage
	roiPercentage := calc.calculateROIPercentage(implementationCost, annualSavings)

	// Calculate Net Present Value (NPV)
	npv := calc.calculateNPV(implementationCost, annualSavings, 5) // 5-year horizon

	// Calculate Internal Rate of Return (IRR)
	irr := calc.calculateIRR(implementationCost, annualSavings, 5)

	// Apply risk adjustment
	riskAdjustment := calc.calculateRiskAdjustment(recommendation)

	// Generate assumptions
	assumptions := calc.generateAssumptions(recommendation, currentMetrics)

	// Perform sensitivity analysis
	sensitivityAnalysis := calc.performSensitivityAnalysis(implementationCost, annualSavings)

	return ROIAnalysis{
		InitialInvestment:   implementationCost,
		AnnualSavings:       annualSavings,
		PaybackPeriod:       paybackPeriod,
		ROIPercentage:       roiPercentage,
		NPV:                 npv,
		IRR:                 irr,
		RiskAdjustment:      riskAdjustment,
		Assumptions:         assumptions,
		SensitivityAnalysis: sensitivityAnalysis,
	}
}

// EstimateCostSavings projects cost savings from optimization recommendations
func (calc *ROICalculator) EstimateCostSavings(recommendations []Recommendation, timeHorizon time.Duration) CostSavingsEstimate {
	totalSavings := 0.0
	totalImplementationCost := 0.0
	savingsByCategory := make(map[string]float64)
	var breakdown []CostSavingsBreakdown

	// Calculate aggregate savings across all recommendations
	for _, rec := range recommendations {
		// Calculate individual recommendation savings
		annualSaving := calc.calculateRecommendationSavings(rec)
		implementationCost := calc.calculateImplementationCost(rec)

		// Project savings over time horizon
		projectedSavings := annualSaving * (timeHorizon.Hours() / 8760) // Convert to annual fraction

		totalSavings += projectedSavings
		totalImplementationCost += implementationCost

		// Categorize savings
		category := rec.Category
		savingsByCategory[category] += projectedSavings

		// Add to breakdown
		breakdown = append(breakdown, CostSavingsBreakdown{
			Category:    category,
			Description: rec.Title,
			Amount:      projectedSavings,
			Frequency:   "Annual",
			Confidence:  rec.Confidence,
		})
	}

	// Calculate net savings
	netSavings := totalSavings - totalImplementationCost

	// Calculate overall confidence level
	confidenceLevel := calc.calculateOverallConfidence(recommendations)

	return CostSavingsEstimate{
		TimeHorizon:        timeHorizon,
		TotalSavings:       totalSavings,
		SavingsByCategory:  savingsByCategory,
		ImplementationCost: totalImplementationCost,
		NetSavings:         netSavings,
		ConfidenceLevel:    confidenceLevel,
		Breakdown:          breakdown,
	}
}

// Private calculation methods

// calculateImplementationCost calculates the cost to implement a recommendation
func (calc *ROICalculator) calculateImplementationCost(rec Recommendation) float64 {
	baseCost := 0.0

	// Calculate labor cost based on complexity and time to effect
	complexityFactor := calc.costModel.ComplexityCostFactors[rec.Complexity]
	laborHours := rec.TimeToEffect.Hours() * complexityFactor
	laborCost := laborHours * calc.costModel.HourlyDeveloperCost

	baseCost += laborCost

	// Add infrastructure costs based on recommendation type
	switch rec.Type {
	case RecommendationTypeScaling:
		baseCost += calc.calculateScalingCost(rec)
	case RecommendationTypeCapacity:
		baseCost += calc.calculateCapacityCost(rec)
	case RecommendationTypeArchitecture:
		baseCost += calc.calculateArchitectureCost(rec)
	case RecommendationTypeOptimization:
		baseCost += calc.calculateOptimizationCost(rec)
	case RecommendationTypeMonitoring:
		baseCost += calc.calculateMonitoringCost(rec)
	case RecommendationTypeMaintenance:
		baseCost += calc.calculateMaintenanceCost(rec)
	}

	// Apply maintenance multiplier for ongoing costs
	if rec.Type == RecommendationTypeCapacity || rec.Type == RecommendationTypeScaling {
		baseCost *= calc.costModel.MaintenanceMultiplier
	}

	return baseCost
}

// calculateAnnualSavings calculates expected annual savings from a recommendation
func (calc *ROICalculator) calculateAnnualSavings(rec Recommendation, currentMetrics metrics.MetricsSnapshot) float64 {
	savings := 0.0

	// Revenue protection/enhancement
	if rec.Impact.Revenue > 0 {
		savings += rec.Impact.Revenue
	}

	// Cost reduction
	if rec.Impact.Cost < 0 {
		savings += math.Abs(rec.Impact.Cost)
	}

	// Calculate operational efficiency savings
	efficiencySavings := calc.calculateEfficiencySavings(rec, currentMetrics)
	savings += efficiencySavings

	// Calculate customer experience value
	cxValue := calc.calculateCustomerExperienceValue(rec, currentMetrics)
	savings += cxValue

	// Calculate competitive advantage value
	competitiveValue := calc.calculateCompetitiveValue(rec)
	savings += competitiveValue

	// Apply confidence adjustment
	savings *= rec.Confidence

	return savings
}

// calculatePaybackPeriod calculates how long it takes to recover the investment
func (calc *ROICalculator) calculatePaybackPeriod(investment, annualSavings float64) time.Duration {
	if annualSavings <= 0 {
		return time.Duration(0) // Infinite payback
	}

	yearsToPayback := investment / annualSavings
	return time.Duration(yearsToPayback * float64(time.Hour*24*365))
}

// calculateROIPercentage calculates the ROI as a percentage
func (calc *ROICalculator) calculateROIPercentage(investment, annualSavings float64) float64 {
	if investment <= 0 {
		return 0
	}

	return ((annualSavings - investment) / investment) * 100
}

// calculateNPV calculates Net Present Value over specified years
func (calc *ROICalculator) calculateNPV(investment, annualSavings float64, years int) float64 {
	npv := -investment // Initial investment is negative cash flow

	for year := 1; year <= years; year++ {
		// Discount future cash flows
		discountedSavings := annualSavings / math.Pow(1+calc.config.DiscountRate, float64(year))
		npv += discountedSavings
	}

	return npv
}

// calculateIRR calculates Internal Rate of Return using Newton-Raphson method
func (calc *ROICalculator) calculateIRR(investment, annualSavings float64, years int) float64 {
	// Simplified IRR calculation
	// For equal annual cash flows: IRR ≈ (Annual Cash Flow / Initial Investment) - 1
	if investment <= 0 {
		return 0
	}

	// Initial guess
	irr := 0.1 // 10%

	// Newton-Raphson iterations
	for i := 0; i < 100; i++ {
		npv := -investment
		npvDerivative := 0.0

		for year := 1; year <= years; year++ {
			factor := math.Pow(1+irr, float64(year))
			npv += annualSavings / factor
			npvDerivative -= float64(year) * annualSavings / math.Pow(1+irr, float64(year+1))
		}

		if math.Abs(npv) < 0.01 {
			break
		}

		irr = irr - npv/npvDerivative

		// Prevent negative or unrealistic IRR
		if irr < -0.99 || irr > 10 {
			irr = 0.1
			break
		}
	}

	return irr * 100 // Return as percentage
}

// calculateRiskAdjustment applies risk adjustment based on recommendation characteristics
func (calc *ROICalculator) calculateRiskAdjustment(rec Recommendation) float64 {
	baseRisk := calc.config.ProjectRiskMultiplier

	// Adjust risk based on complexity
	complexityRisk := map[Complexity]float64{
		ComplexityLow:    0.1,
		ComplexityMedium: 0.2,
		ComplexityHigh:   0.4,
	}

	// Adjust risk based on recommendation type
	typeRisk := map[RecommendationType]float64{
		RecommendationTypeOptimization: 0.1,
		RecommendationTypeMonitoring:   0.05,
		RecommendationTypeScaling:      0.15,
		RecommendationTypeCapacity:     0.2,
		RecommendationTypeArchitecture: 0.3,
		RecommendationTypeMaintenance:  0.1,
	}

	totalRisk := baseRisk + complexityRisk[rec.Complexity] + typeRisk[rec.Type]

	// Adjust for confidence level
	confidenceAdjustment := (1 - rec.Confidence) * 0.2

	return totalRisk + confidenceAdjustment
}

// Support calculation methods

// calculateScalingCost calculates cost for scaling recommendations
func (calc *ROICalculator) calculateScalingCost(rec Recommendation) float64 {
	// Estimate scaling cost based on current infrastructure
	// This is a simplified calculation
	baseCost := 10000.0 // Base infrastructure cost

	// Adjust based on recommendation severity/impact
	impactMultiplier := rec.Impact.OverallScore
	return baseCost * impactMultiplier
}

// calculateCapacityCost calculates cost for capacity recommendations
func (calc *ROICalculator) calculateCapacityCost(rec Recommendation) float64 {
	// Estimate capacity costs
	return 15000.0 * rec.Impact.OverallScore
}

// calculateArchitectureCost calculates cost for architecture recommendations
func (calc *ROICalculator) calculateArchitectureCost(rec Recommendation) float64 {
	// Architecture changes are typically more expensive
	return 25000.0 * rec.Impact.OverallScore
}

// calculateOptimizationCost calculates cost for optimization recommendations
func (calc *ROICalculator) calculateOptimizationCost(rec Recommendation) float64 {
	// Optimization is primarily labor cost
	return 5000.0 * rec.Impact.OverallScore
}

// calculateMonitoringCost calculates cost for monitoring recommendations
func (calc *ROICalculator) calculateMonitoringCost(rec Recommendation) float64 {
	// Monitoring setup and tools
	return 3000.0 * rec.Impact.OverallScore
}

// calculateMaintenanceCost calculates cost for maintenance recommendations
func (calc *ROICalculator) calculateMaintenanceCost(rec Recommendation) float64 {
	// Maintenance activities
	return 2000.0 * rec.Impact.OverallScore
}

// calculateRecommendationSavings calculates savings for individual recommendation
func (calc *ROICalculator) calculateRecommendationSavings(rec Recommendation) float64 {
	return math.Abs(rec.Impact.Revenue) + math.Abs(rec.Impact.Cost)
}

// calculateEfficiencySavings calculates operational efficiency savings
func (calc *ROICalculator) calculateEfficiencySavings(rec Recommendation, currentMetrics metrics.MetricsSnapshot) float64 {
	// Calculate efficiency improvements based on current metrics
	savings := 0.0

	// Latency improvements
	if rec.Impact.UserExperience > 0.8 {
		latencyValue := currentMetrics.OrdersPerSec * calc.marketParameters.CostPerLatencyMs * 24 * 365
		savings += latencyValue * 0.1 // 10% improvement estimate
	}

	// Throughput improvements
	if rec.Impact.Scalability > 0.8 {
		throughputValue := currentMetrics.TradesPerSec * calc.marketParameters.RevenuePerTrade * 24 * 365
		savings += throughputValue * 0.05 // 5% improvement estimate
	}

	return savings
}

// calculateCustomerExperienceValue calculates value from improved customer experience
func (calc *ROICalculator) calculateCustomerExperienceValue(rec Recommendation, currentMetrics metrics.MetricsSnapshot) float64 {
	if rec.Impact.UserExperience < 0.7 {
		return 0
	}

	// Estimate customer retention value
	dailyOrders := currentMetrics.OrdersPerSec * 24 * 3600
	annualOrders := dailyOrders * 365
	customerValue := annualOrders * calc.marketParameters.RevenuePerOrder * calc.marketParameters.CustomerLTVMultiplier

	// UX improvement factor
	uxImprovement := (rec.Impact.UserExperience - 0.7) / 0.3 // Normalize to 0-1

	return customerValue * uxImprovement * 0.02 // 2% max improvement
}

// calculateCompetitiveValue calculates competitive advantage value
func (calc *ROICalculator) calculateCompetitiveValue(rec Recommendation) float64 {
	if rec.Impact.Scalability < 0.8 {
		return 0
	}

	// Competitive advantage in trading is significant
	competitiveValue := 50000.0 // Base competitive value
	return competitiveValue * rec.Impact.Scalability * calc.marketParameters.CompetitiveImpact
}

// calculateOverallConfidence calculates confidence across multiple recommendations
func (calc *ROICalculator) calculateOverallConfidence(recommendations []Recommendation) float64 {
	if len(recommendations) == 0 {
		return 0
	}

	totalConfidence := 0.0
	for _, rec := range recommendations {
		totalConfidence += rec.Confidence
	}

	avgConfidence := totalConfidence / float64(len(recommendations))

	// Adjust for number of recommendations (more recommendations = higher confidence)
	diversificationBonus := math.Min(0.1, float64(len(recommendations))*0.02)

	return math.Min(1.0, avgConfidence+diversificationBonus)
}

// generateAssumptions creates list of assumptions for the ROI analysis
func (calc *ROICalculator) generateAssumptions(rec Recommendation, currentMetrics metrics.MetricsSnapshot) []string {
	assumptions := []string{
		fmt.Sprintf("Discount rate: %.1f%%", calc.config.DiscountRate*100),
		fmt.Sprintf("Implementation completed within estimated timeframe: %v", rec.TimeToEffect),
		"Current performance trends continue without intervention",
		"No major competitive or regulatory changes",
		fmt.Sprintf("Confidence level: %.1f%%", rec.Confidence*100),
	}

	// Add specific assumptions based on recommendation type
	switch rec.Type {
	case RecommendationTypeScaling:
		assumptions = append(assumptions, "Infrastructure scaling provides linear performance improvement")
	case RecommendationTypeOptimization:
		assumptions = append(assumptions, "Code optimizations maintain current functionality")
	case RecommendationTypeCapacity:
		assumptions = append(assumptions, "Capacity expansion matches predicted growth patterns")
	}

	return assumptions
}

// performSensitivityAnalysis performs sensitivity analysis on key variables
func (calc *ROICalculator) performSensitivityAnalysis(investment, annualSavings float64) map[string]float64 {
	baseROI := ((annualSavings - investment) / investment) * 100

	sensitivity := make(map[string]float64)

	// Test +/- 20% changes in key variables
	scenarios := map[string]float64{
		"cost_+20%":        (annualSavings - investment*1.2) / (investment*1.2) * 100,
		"cost_-20%":        (annualSavings - investment*0.8) / (investment*0.8) * 100,
		"savings_+20%":     (annualSavings*1.2 - investment) / investment * 100,
		"savings_-20%":     (annualSavings*0.8 - investment) / investment * 100,
		"discount_rate_+2%": calc.calculateROIWithDiscountRate(investment, annualSavings, calc.config.DiscountRate+0.02),
		"discount_rate_-2%": calc.calculateROIWithDiscountRate(investment, annualSavings, calc.config.DiscountRate-0.02),
	}

	for scenario, roi := range scenarios {
		sensitivity[scenario] = roi - baseROI // Change from base case
	}

	return sensitivity
}

// calculateROIWithDiscountRate calculates ROI with different discount rate
func (calc *ROICalculator) calculateROIWithDiscountRate(investment, annualSavings, discountRate float64) float64 {
	npv := -investment
	for year := 1; year <= 5; year++ {
		npv += annualSavings / math.Pow(1+discountRate, float64(year))
	}
	return (npv / investment) * 100
}

// Default configuration methods

// DefaultROIConfig returns default ROI calculation configuration
func DefaultROIConfig() ROIConfig {
	return ROIConfig{
		DiscountRate:          0.08,  // 8% annual discount rate
		RiskFreeRate:          0.02,  // 2% risk-free rate
		MarketRiskPremium:     0.06,  // 6% market risk premium
		ProjectRiskMultiplier: 0.15,  // 15% project risk
		TaxRate:               0.25,  // 25% corporate tax rate
		InflationRate:         0.025, // 2.5% annual inflation
	}
}

// DefaultMarketParameters returns default market parameters
func DefaultMarketParameters() MarketParameters {
	return MarketParameters{
		RevenuePerOrder:       2.50,   // $2.50 per order
		RevenuePerTrade:       5.00,   // $5.00 per trade
		CostPerLatencyMs:      0.01,   // $0.01 per ms of latency
		RevenueAtRiskPercent:  0.05,   // 5% revenue at risk
		CustomerLTVMultiplier: 5.0,    // 5x annual revenue as LTV
		CompetitiveImpact:     0.1,    // 10% competitive impact factor
	}
}

// DefaultCostModel returns default cost model
func DefaultCostModel() CostModel {
	return CostModel{
		HourlyDeveloperCost:   150.0, // $150/hour for developers
		HourlyInfraCost:       50.0,  // $50/hour for infrastructure
		ComputeUnitCost:       1200.0, // $1200/year per compute unit
		MemoryGBCost:          50.0,   // $50/year per GB memory
		StorageGBCost:         2.0,    // $2/year per GB storage
		NetworkBandwidthCost:  100.0,  // $100/year per Mbps
		DatabaseIOPSCost:      5.0,    // $5/year per IOPS
		MaintenanceMultiplier: 1.2,    // 20% additional for maintenance
		ComplexityCostFactors: map[Complexity]float64{
			ComplexityLow:    1.0,
			ComplexityMedium: 2.0,
			ComplexityHigh:   4.0,
		},
	}
}