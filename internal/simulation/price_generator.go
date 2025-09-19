package simulation

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// RealisticPriceGenerator implements PriceGenerator interface
type RealisticPriceGenerator struct {
	// Price state for each symbol
	prices       map[string]*PriceState
	basePrices   map[string]float64
	mu           sync.RWMutex

	// Volatility settings
	baseVolatility    float64
	currentVolatility map[string]float64
	volatilityDecay   float64

	// Market microstructure
	spreadPercentage  float64
	priceStepSize     float64

	// Trend analysis
	trendAnalysis     map[string]*TrendAnalysis
	supportResistance map[string]*SupportResistance

	// Random number generator with seed
	rng *rand.Rand

	// Configuration
	config PriceGeneratorConfig
}

// PriceState tracks price information for a symbol
type PriceState struct {
	Symbol          string        `json:"symbol"`
	CurrentPrice    float64       `json:"current_price"`
	PreviousPrice   float64       `json:"previous_price"`
	DailyOpen       float64       `json:"daily_open"`
	DailyHigh       float64       `json:"daily_high"`
	DailyLow        float64       `json:"daily_low"`
	Volume          float64       `json:"volume"`
	LastUpdate      time.Time     `json:"last_update"`
	PriceHistory    []PricePoint  `json:"price_history"`
	VolatilityIndex float64       `json:"volatility_index"`
}

// PricePoint represents a point in price history
type PricePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
}

// TrendAnalysis tracks price trends
type TrendAnalysis struct {
	Direction       TrendDirection `json:"direction"`
	Strength        float64        `json:"strength"`
	StartTime       time.Time      `json:"start_time"`
	Duration        time.Duration  `json:"duration"`
	PriceChange     float64        `json:"price_change"`
	VolumeProfile   VolumeProfile  `json:"volume_profile"`
	MomentumFactor  float64        `json:"momentum_factor"`
}

// VolumeProfile tracks volume characteristics
type VolumeProfile struct {
	AverageVolume  float64 `json:"average_volume"`
	VolumeSpike    bool    `json:"volume_spike"`
	VolumeRatio    float64 `json:"volume_ratio"` // Current vs average
}

// SupportResistance tracks key price levels
type SupportResistance struct {
	SupportLevels    []float64 `json:"support_levels"`
	ResistanceLevels []float64 `json:"resistance_levels"`
	LastUpdate       time.Time `json:"last_update"`
	Confidence       float64   `json:"confidence"`
}

// PriceGeneratorConfig configures price generation behavior
type PriceGeneratorConfig struct {
	BaseVolatility     float64           `json:"base_volatility"`
	VolatilityDecay    float64           `json:"volatility_decay"`
	SpreadPercentage   float64           `json:"spread_percentage"`
	PriceStepSize      float64           `json:"price_step_size"`
	TrendPersistence   float64           `json:"trend_persistence"`
	MeanReversion      float64           `json:"mean_reversion"`
	HistorySize        int               `json:"history_size"`
	RandomSeed         int64             `json:"random_seed"`
	MarketHours        TradingHours      `json:"market_hours"`
	HolidayAdjustment  float64           `json:"holiday_adjustment"`
}

// NewRealisticPriceGenerator creates a new price generator
func NewRealisticPriceGenerator(config PriceGeneratorConfig) *RealisticPriceGenerator {
	if config.RandomSeed == 0 {
		config.RandomSeed = time.Now().UnixNano()
	}

	return &RealisticPriceGenerator{
		prices:            make(map[string]*PriceState),
		basePrices:        make(map[string]float64),
		currentVolatility: make(map[string]float64),
		trendAnalysis:     make(map[string]*TrendAnalysis),
		supportResistance: make(map[string]*SupportResistance),
		baseVolatility:    config.BaseVolatility,
		volatilityDecay:   config.VolatilityDecay,
		spreadPercentage:  config.SpreadPercentage,
		priceStepSize:     config.PriceStepSize,
		rng:               rand.New(rand.NewSource(config.RandomSeed)),
		config:            config,
	}
}

// GeneratePrice creates next price based on current market conditions
func (rpg *RealisticPriceGenerator) GeneratePrice(symbol string, currentPrice float64, timeElapsed time.Duration) float64 {
	rpg.mu.Lock()
	defer rpg.mu.Unlock()

	// Initialize price state if needed
	if _, exists := rpg.prices[symbol]; !exists {
		rpg.initializePriceState(symbol, currentPrice)
	}

	priceState := rpg.prices[symbol]

	// Update price state
	priceState.PreviousPrice = priceState.CurrentPrice
	priceState.CurrentPrice = currentPrice

	// Calculate time-based factors
	timeFactor := rpg.calculateTimeFactor(timeElapsed)

	// Get current volatility for symbol
	volatility := rpg.getCurrentVolatility(symbol)

	// Calculate price change components
	trendComponent := rpg.calculateTrendComponent(symbol, timeFactor)
	randomComponent := rpg.calculateRandomComponent(symbol, volatility, timeFactor)
	meanReversionComponent := rpg.calculateMeanReversionComponent(symbol)
	supportResistanceComponent := rpg.calculateSupportResistanceComponent(symbol, currentPrice)

	// Combine components
	totalChange := trendComponent + randomComponent + meanReversionComponent + supportResistanceComponent

	// Apply change to current price
	newPrice := currentPrice * (1 + totalChange)

	// Ensure price is positive and apply step size
	newPrice = math.Max(newPrice, rpg.priceStepSize)
	newPrice = rpg.roundToStepSize(newPrice)

	// Update price state
	rpg.updatePriceState(symbol, newPrice, timeElapsed)

	// Update trend analysis
	rpg.updateTrendAnalysis(symbol, newPrice, timeElapsed)

	// Update support/resistance levels
	rpg.updateSupportResistance(symbol, newPrice)

	// Decay volatility
	rpg.decayVolatility(symbol)

	return newPrice
}

// SimulateVolatility applies volatility pattern to price generation
func (rpg *RealisticPriceGenerator) SimulateVolatility(pattern VolatilityPattern, intensity float64) {
	rpg.mu.Lock()
	defer rpg.mu.Unlock()

	// Apply volatility to all symbols
	for symbol := range rpg.prices {
		switch pattern {
		case VolatilitySpike:
			rpg.currentVolatility[symbol] = rpg.baseVolatility * (1 + intensity*3)
		case VolatilityDecay:
			rpg.currentVolatility[symbol] *= (1 - intensity*0.1)
		case VolatilityOscillate:
			// Create oscillating volatility
			oscFactor := math.Sin(float64(time.Now().Unix())/60) * intensity
			rpg.currentVolatility[symbol] = rpg.baseVolatility * (1 + oscFactor)
		case VolatilityRandom:
			// Random volatility bursts
			randomFactor := rpg.rng.Float64() * intensity
			rpg.currentVolatility[symbol] = rpg.baseVolatility * (1 + randomFactor*2)
		case VolatilityNews:
			// News-driven volatility (temporary spike with gradual decay)
			newsSpikeVolatility := rpg.baseVolatility * (1 + intensity*2)
			rpg.currentVolatility[symbol] = math.Max(rpg.currentVolatility[symbol], newsSpikeVolatility)
		}

		// Ensure volatility stays within reasonable bounds
		rpg.currentVolatility[symbol] = math.Max(rpg.currentVolatility[symbol], rpg.baseVolatility*0.1)
		rpg.currentVolatility[symbol] = math.Min(rpg.currentVolatility[symbol], rpg.baseVolatility*10)
	}
}

// GetPriceTrend returns current price trend information
func (rpg *RealisticPriceGenerator) GetPriceTrend(symbol string) PriceTrend {
	rpg.mu.RLock()
	defer rpg.mu.RUnlock()

	trend := PriceTrend{
		Symbol:      symbol,
		Direction:   TrendSideways,
		Strength:    0.0,
		LastUpdate:  time.Now(),
	}

	if analysis, exists := rpg.trendAnalysis[symbol]; exists {
		trend.Direction = analysis.Direction
		trend.Strength = analysis.Strength
		trend.Duration = time.Since(analysis.StartTime)
		trend.VolatilityLevel = rpg.getCurrentVolatility(symbol)

		if sr, exists := rpg.supportResistance[symbol]; exists {
			if len(sr.SupportLevels) > 0 {
				trend.SupportLevel = sr.SupportLevels[len(sr.SupportLevels)-1]
			}
			if len(sr.ResistanceLevels) > 0 {
				trend.ResistanceLevel = sr.ResistanceLevels[len(sr.ResistanceLevels)-1]
			}
		}
	}

	return trend
}

// SetBasePrice establishes baseline price for symbol
func (rpg *RealisticPriceGenerator) SetBasePrice(symbol string, price float64) {
	rpg.mu.Lock()
	defer rpg.mu.Unlock()

	rpg.basePrices[symbol] = price
	rpg.currentVolatility[symbol] = rpg.baseVolatility
	rpg.initializePriceState(symbol, price)
}

// Reset clears all price history and patterns
func (rpg *RealisticPriceGenerator) Reset() {
	rpg.mu.Lock()
	defer rpg.mu.Unlock()

	rpg.prices = make(map[string]*PriceState)
	rpg.basePrices = make(map[string]float64)
	rpg.currentVolatility = make(map[string]float64)
	rpg.trendAnalysis = make(map[string]*TrendAnalysis)
	rpg.supportResistance = make(map[string]*SupportResistance)
}

// Private helper methods

func (rpg *RealisticPriceGenerator) initializePriceState(symbol string, price float64) {
	now := time.Now()

	rpg.prices[symbol] = &PriceState{
		Symbol:          symbol,
		CurrentPrice:    price,
		PreviousPrice:   price,
		DailyOpen:       price,
		DailyHigh:       price,
		DailyLow:        price,
		Volume:          0,
		LastUpdate:      now,
		PriceHistory:    make([]PricePoint, 0, rpg.config.HistorySize),
		VolatilityIndex: rpg.baseVolatility,
	}

	rpg.trendAnalysis[symbol] = &TrendAnalysis{
		Direction:      TrendSideways,
		Strength:       0.0,
		StartTime:      now,
		Duration:       0,
		PriceChange:    0.0,
		MomentumFactor: 0.0,
		VolumeProfile: VolumeProfile{
			AverageVolume: 1000.0,
			VolumeSpike:   false,
			VolumeRatio:   1.0,
		},
	}

	rpg.supportResistance[symbol] = &SupportResistance{
		SupportLevels:    []float64{price * 0.95, price * 0.90},
		ResistanceLevels: []float64{price * 1.05, price * 1.10},
		LastUpdate:       now,
		Confidence:       0.5,
	}

	if rpg.currentVolatility[symbol] == 0 {
		rpg.currentVolatility[symbol] = rpg.baseVolatility
	}
}

func (rpg *RealisticPriceGenerator) getCurrentVolatility(symbol string) float64 {
	if vol, exists := rpg.currentVolatility[symbol]; exists {
		return vol
	}
	return rpg.baseVolatility
}

func (rpg *RealisticPriceGenerator) calculateTimeFactor(timeElapsed time.Duration) float64 {
	// Convert to hours and apply square root scaling
	hours := timeElapsed.Hours()
	if hours <= 0 {
		return 0.1 // Minimum time factor
	}
	return math.Sqrt(hours)
}

func (rpg *RealisticPriceGenerator) calculateTrendComponent(symbol string, timeFactor float64) float64 {
	analysis := rpg.trendAnalysis[symbol]
	if analysis == nil {
		return 0.0
	}

	// Calculate trend strength decay over time
	trendAge := time.Since(analysis.StartTime).Hours()
	ageDecay := math.Exp(-trendAge / 24.0) // Decay over 24 hours

	// Apply momentum and persistence
	trendStrength := analysis.Strength * ageDecay * rpg.config.TrendPersistence

	// Convert direction to multiplier
	var directionMultiplier float64
	switch analysis.Direction {
	case TrendUp:
		directionMultiplier = 1.0
	case TrendDown:
		directionMultiplier = -1.0
	default:
		directionMultiplier = 0.0
	}

	return trendStrength * directionMultiplier * timeFactor * 0.001 // Scale down
}

func (rpg *RealisticPriceGenerator) calculateRandomComponent(symbol string, volatility, timeFactor float64) float64 {
	// Generate normally distributed random walk
	randomValue := rpg.rng.NormFloat64()

	// Scale by volatility and time
	return randomValue * volatility * timeFactor * 0.01
}

func (rpg *RealisticPriceGenerator) calculateMeanReversionComponent(symbol string) float64 {
	basePrice, exists := rpg.basePrices[symbol]
	if !exists {
		return 0.0
	}

	priceState := rpg.prices[symbol]
	if priceState == nil {
		return 0.0
	}

	// Calculate deviation from base price
	deviation := (priceState.CurrentPrice - basePrice) / basePrice

	// Apply mean reversion force
	reversionForce := -deviation * rpg.config.MeanReversion

	return reversionForce * 0.001 // Scale down
}

func (rpg *RealisticPriceGenerator) calculateSupportResistanceComponent(symbol string, currentPrice float64) float64 {
	sr := rpg.supportResistance[symbol]
	if sr == nil {
		return 0.0
	}

	var component float64

	// Check for support levels (price bounces up)
	for _, support := range sr.SupportLevels {
		if currentPrice <= support*1.01 && currentPrice >= support*0.99 {
			// Near support, add upward pressure
			component += 0.001 * sr.Confidence
		}
	}

	// Check for resistance levels (price bounces down)
	for _, resistance := range sr.ResistanceLevels {
		if currentPrice >= resistance*0.99 && currentPrice <= resistance*1.01 {
			// Near resistance, add downward pressure
			component -= 0.001 * sr.Confidence
		}
	}

	return component
}

func (rpg *RealisticPriceGenerator) roundToStepSize(price float64) float64 {
	if rpg.priceStepSize <= 0 {
		return price
	}
	return math.Round(price/rpg.priceStepSize) * rpg.priceStepSize
}

func (rpg *RealisticPriceGenerator) updatePriceState(symbol string, newPrice float64, timeElapsed time.Duration) {
	priceState := rpg.prices[symbol]
	now := time.Now()

	// Update current values
	priceState.CurrentPrice = newPrice
	priceState.LastUpdate = now

	// Update daily high/low
	if newPrice > priceState.DailyHigh {
		priceState.DailyHigh = newPrice
	}
	if newPrice < priceState.DailyLow {
		priceState.DailyLow = newPrice
	}

	// Add to price history
	pricePoint := PricePoint{
		Timestamp: now,
		Price:     newPrice,
		Volume:    rpg.generateVolume(symbol, newPrice),
	}

	priceState.PriceHistory = append(priceState.PriceHistory, pricePoint)

	// Maintain history size limit
	if len(priceState.PriceHistory) > rpg.config.HistorySize {
		priceState.PriceHistory = priceState.PriceHistory[1:]
	}

	// Update volatility index
	priceState.VolatilityIndex = rpg.calculateVolatilityIndex(symbol)

	// Reset daily values if new day
	if rpg.isNewTradingDay(priceState.LastUpdate, now) {
		priceState.DailyOpen = newPrice
		priceState.DailyHigh = newPrice
		priceState.DailyLow = newPrice
	}
}

func (rpg *RealisticPriceGenerator) generateVolume(symbol string, price float64) float64 {
	// Generate realistic volume based on price movement
	priceState := rpg.prices[symbol]

	baseVolume := 1000.0
	if priceState != nil && len(priceState.PriceHistory) > 0 {
		// Use average historical volume as base
		totalVolume := 0.0
		for _, point := range priceState.PriceHistory {
			totalVolume += point.Volume
		}
		baseVolume = totalVolume / float64(len(priceState.PriceHistory))
	}

	// Volume increases with price volatility
	volatility := rpg.getCurrentVolatility(symbol)
	volumeMultiplier := 1.0 + volatility*2

	// Add random variation
	randomFactor := 0.5 + rpg.rng.Float64()

	return baseVolume * volumeMultiplier * randomFactor
}

func (rpg *RealisticPriceGenerator) calculateVolatilityIndex(symbol string) float64 {
	priceState := rpg.prices[symbol]
	if len(priceState.PriceHistory) < 2 {
		return rpg.baseVolatility
	}

	// Calculate standard deviation of recent price changes
	const lookback = 20
	startIdx := len(priceState.PriceHistory) - lookback
	if startIdx < 0 {
		startIdx = 0
	}

	var returns []float64
	for i := startIdx + 1; i < len(priceState.PriceHistory); i++ {
		prevPrice := priceState.PriceHistory[i-1].Price
		currPrice := priceState.PriceHistory[i].Price
		if prevPrice > 0 {
			ret := (currPrice - prevPrice) / prevPrice
			returns = append(returns, ret)
		}
	}

	if len(returns) == 0 {
		return rpg.baseVolatility
	}

	// Calculate standard deviation
	mean := 0.0
	for _, ret := range returns {
		mean += ret
	}
	mean /= float64(len(returns))

	variance := 0.0
	for _, ret := range returns {
		variance += (ret - mean) * (ret - mean)
	}
	variance /= float64(len(returns))

	return math.Sqrt(variance)
}

func (rpg *RealisticPriceGenerator) updateTrendAnalysis(symbol string, newPrice float64, timeElapsed time.Duration) {
	analysis := rpg.trendAnalysis[symbol]
	priceState := rpg.prices[symbol]

	if len(priceState.PriceHistory) < 5 {
		return // Need more data
	}

	// Calculate trend using linear regression on recent prices
	const trendLookback = 10
	startIdx := len(priceState.PriceHistory) - trendLookback
	if startIdx < 0 {
		startIdx = 0
	}

	slope := rpg.calculatePriceSlope(priceState.PriceHistory[startIdx:])

	// Determine trend direction and strength
	var newDirection TrendDirection
	strength := math.Abs(slope) * 1000 // Scale for visibility

	if slope > 0.001 {
		newDirection = TrendUp
	} else if slope < -0.001 {
		newDirection = TrendDown
	} else {
		newDirection = TrendSideways
	}

	// Update trend if direction changed
	if newDirection != analysis.Direction {
		analysis.Direction = newDirection
		analysis.StartTime = time.Now()
		analysis.PriceChange = 0.0
	}

	analysis.Strength = math.Min(strength, 1.0) // Cap at 1.0
	analysis.Duration = time.Since(analysis.StartTime)

	// Update price change since trend start
	if len(priceState.PriceHistory) > 0 {
		startPrice := priceState.PriceHistory[0].Price
		analysis.PriceChange = (newPrice - startPrice) / startPrice
	}

	// Calculate momentum factor
	if len(priceState.PriceHistory) >= 2 {
		recentChange := (newPrice - priceState.PriceHistory[len(priceState.PriceHistory)-2].Price) / priceState.PriceHistory[len(priceState.PriceHistory)-2].Price
		analysis.MomentumFactor = recentChange * 100 // Scale to percentage
	}
}

func (rpg *RealisticPriceGenerator) calculatePriceSlope(history []PricePoint) float64 {
	if len(history) < 2 {
		return 0.0
	}

	n := float64(len(history))
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, point := range history {
		x := float64(i)
		y := point.Price
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope using least squares
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0.0
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	return slope
}

func (rpg *RealisticPriceGenerator) updateSupportResistance(symbol string, newPrice float64) {
	sr := rpg.supportResistance[symbol]
	priceState := rpg.prices[symbol]

	if len(priceState.PriceHistory) < 10 {
		return // Need more data
	}

	now := time.Now()

	// Update every 5 minutes to avoid too frequent recalculations
	if time.Since(sr.LastUpdate) < 5*time.Minute {
		return
	}

	// Find local minima and maxima in recent history
	const lookback = 50
	startIdx := len(priceState.PriceHistory) - lookback
	if startIdx < 0 {
		startIdx = 0
	}

	recentHistory := priceState.PriceHistory[startIdx:]

	// Find pivot points
	supports := rpg.findPivotLows(recentHistory)
	resistances := rpg.findPivotHighs(recentHistory)

	// Update support and resistance levels
	sr.SupportLevels = supports
	sr.ResistanceLevels = resistances
	sr.LastUpdate = now
	sr.Confidence = rpg.calculateSRConfidence(supports, resistances, newPrice)
}

func (rpg *RealisticPriceGenerator) findPivotLows(history []PricePoint) []float64 {
	if len(history) < 5 {
		return []float64{}
	}

	var supports []float64
	for i := 2; i < len(history)-2; i++ {
		price := history[i].Price

		// Check if this is a local minimum
		if price < history[i-1].Price && price < history[i-2].Price &&
		   price < history[i+1].Price && price < history[i+2].Price {
			supports = append(supports, price)
		}
	}

	// Remove duplicates and sort
	supports = rpg.removeDuplicateLevels(supports, 0.01) // 1% tolerance
	return supports
}

func (rpg *RealisticPriceGenerator) findPivotHighs(history []PricePoint) []float64 {
	if len(history) < 5 {
		return []float64{}
	}

	var resistances []float64
	for i := 2; i < len(history)-2; i++ {
		price := history[i].Price

		// Check if this is a local maximum
		if price > history[i-1].Price && price > history[i-2].Price &&
		   price > history[i+1].Price && price > history[i+2].Price {
			resistances = append(resistances, price)
		}
	}

	// Remove duplicates and sort
	resistances = rpg.removeDuplicateLevels(resistances, 0.01) // 1% tolerance
	return resistances
}

func (rpg *RealisticPriceGenerator) removeDuplicateLevels(levels []float64, tolerance float64) []float64 {
	if len(levels) <= 1 {
		return levels
	}

	var unique []float64
	for _, level := range levels {
		isUnique := true
		for _, existing := range unique {
			if math.Abs((level-existing)/existing) < tolerance {
				isUnique = false
				break
			}
		}
		if isUnique {
			unique = append(unique, level)
		}
	}

	return unique
}

func (rpg *RealisticPriceGenerator) calculateSRConfidence(supports, resistances []float64, currentPrice float64) float64 {
	// Calculate confidence based on the number of levels and their proximity to current price
	totalLevels := len(supports) + len(resistances)
	if totalLevels == 0 {
		return 0.0
	}

	// Base confidence on number of levels found
	confidence := math.Min(float64(totalLevels)/10.0, 1.0) // Max confidence at 10 levels

	// Reduce confidence if current price is far from any level
	minDistance := math.Inf(1)
	for _, level := range supports {
		distance := math.Abs((currentPrice - level) / level)
		if distance < minDistance {
			minDistance = distance
		}
	}
	for _, level := range resistances {
		distance := math.Abs((currentPrice - level) / level)
		if distance < minDistance {
			minDistance = distance
		}
	}

	if minDistance < 0.05 { // Within 5%
		confidence *= 1.0
	} else if minDistance < 0.1 { // Within 10%
		confidence *= 0.8
	} else {
		confidence *= 0.5
	}

	return confidence
}

func (rpg *RealisticPriceGenerator) decayVolatility(symbol string) {
	if vol, exists := rpg.currentVolatility[symbol]; exists {
		// Exponential decay towards base volatility
		decayFactor := 1.0 - rpg.volatilityDecay
		newVol := vol*decayFactor + rpg.baseVolatility*(1-decayFactor)
		rpg.currentVolatility[symbol] = newVol
	}
}

func (rpg *RealisticPriceGenerator) isNewTradingDay(lastUpdate, currentTime time.Time) bool {
	// Simple check: different day
	return lastUpdate.Day() != currentTime.Day()
}

// DefaultPriceGeneratorConfig returns default configuration
func DefaultPriceGeneratorConfig() PriceGeneratorConfig {
	return PriceGeneratorConfig{
		BaseVolatility:    0.02,   // 2% base volatility
		VolatilityDecay:   0.01,   // 1% decay per update
		SpreadPercentage:  0.001,  // 0.1% spread
		PriceStepSize:     0.01,   // $0.01 step size
		TrendPersistence:  0.7,    // 70% trend persistence
		MeanReversion:     0.1,    // 10% mean reversion strength
		HistorySize:       1000,   // Keep last 1000 price points
		RandomSeed:        0,      // Will be set to current time
		HolidayAdjustment: 0.5,    // 50% activity on holidays
	}
}