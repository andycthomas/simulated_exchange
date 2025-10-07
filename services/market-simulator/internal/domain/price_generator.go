package domain

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sync"
	"time"

	"simulated_exchange/pkg/shared"
)

// PriceGeneratorConfig configures price generation behavior
type PriceGeneratorConfig struct {
	BaseVolatility    float64 `json:"base_volatility"`
	VolatilityDecay   float64 `json:"volatility_decay"`
	SpreadPercentage  float64 `json:"spread_percentage"`
	PriceStepSize     float64 `json:"price_step_size"`
	TrendPersistence  float64 `json:"trend_persistence"`
	MeanReversion     float64 `json:"mean_reversion"`
	HistorySize       int     `json:"history_size"`
	RandomSeed        int64   `json:"random_seed"`
}

// PriceState tracks price information for a symbol
type PriceState struct {
	Symbol          string                  `json:"symbol"`
	CurrentPrice    float64                 `json:"current_price"`
	PreviousPrice   float64                 `json:"previous_price"`
	BasePrice       float64                 `json:"base_price"`
	DailyOpen       float64                 `json:"daily_open"`
	DailyHigh       float64                 `json:"daily_high"`
	DailyLow        float64                 `json:"daily_low"`
	Volume          float64                 `json:"volume"`
	LastUpdate      time.Time               `json:"last_update"`
	PriceHistory    []shared.PriceUpdate    `json:"price_history"`
	VolatilityIndex float64                 `json:"volatility_index"`
	TrendDirection  string                  `json:"trend_direction"`
	TrendStrength   float64                 `json:"trend_strength"`
}

// PriceGenerator implements shared.PriceService interface for price generation
type PriceGenerator struct {
	config PriceGeneratorConfig
	logger *slog.Logger

	// Price state for each symbol
	prices            map[string]*PriceState
	basePrices        map[string]float64
	currentVolatility map[string]float64
	mu                sync.RWMutex

	// Random number generator
	rng *rand.Rand
}

// NewPriceGenerator creates a new price generator
func NewPriceGenerator(config PriceGeneratorConfig, logger *slog.Logger) *PriceGenerator {
	if config.RandomSeed == 0 {
		config.RandomSeed = time.Now().UnixNano()
	}

	return &PriceGenerator{
		config:            config,
		logger:            logger,
		prices:            make(map[string]*PriceState),
		basePrices:        make(map[string]float64),
		currentVolatility: make(map[string]float64),
		rng:               rand.New(rand.NewSource(config.RandomSeed)),
	}
}

// SetBasePrice establishes baseline price for symbol
func (pg *PriceGenerator) SetBasePrice(symbol string, price float64) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	if price <= 0 {
		return fmt.Errorf("price must be positive")
	}

	pg.basePrices[symbol] = price
	pg.currentVolatility[symbol] = pg.config.BaseVolatility
	pg.initializePriceState(symbol, price)

	pg.logger.Info("Base price set", "symbol", symbol, "price", price)
	return nil
}

// GeneratePrice creates next price based on current market conditions
func (pg *PriceGenerator) GeneratePrice(symbol string) (*shared.PriceUpdate, error) {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	priceState, exists := pg.prices[symbol]
	if !exists {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}

	// Calculate time-based factors
	timeFactor := pg.calculateTimeFactor(time.Since(priceState.LastUpdate))

	// Get current volatility for symbol
	volatility := pg.getCurrentVolatility(symbol)

	// Calculate price change components
	trendComponent := pg.calculateTrendComponent(symbol, timeFactor)
	randomComponent := pg.calculateRandomComponent(volatility, timeFactor)
	meanReversionComponent := pg.calculateMeanReversionComponent(symbol)

	// Combine components
	totalChange := trendComponent + randomComponent + meanReversionComponent

	// Apply change to current price
	newPrice := priceState.CurrentPrice * (1 + totalChange)

	// Ensure price is positive and apply step size
	newPrice = math.Max(newPrice, pg.config.PriceStepSize)
	newPrice = pg.roundToStepSize(newPrice)

	// Generate volume
	volume := pg.generateVolume(symbol, newPrice, volatility)

	// Update price state
	pg.updatePriceState(symbol, newPrice, volume)

	// Create price update
	priceUpdate := &shared.PriceUpdate{
		Symbol:    symbol,
		Price:     newPrice,
		Volume:    volume,
		Timestamp: time.Now(),
	}

	pg.logger.Debug("Generated price",
		"symbol", symbol,
		"price", newPrice,
		"change", totalChange,
		"volume", volume,
	)

	return priceUpdate, nil
}

// GetCurrentPrice returns the current price for a symbol
func (pg *PriceGenerator) GetCurrentPrice(symbol string) (float64, error) {
	pg.mu.RLock()
	defer pg.mu.RUnlock()

	priceState, exists := pg.prices[symbol]
	if !exists {
		return 0, fmt.Errorf("symbol %s not found", symbol)
	}

	return priceState.CurrentPrice, nil
}

// GetPriceHistory returns price history for a symbol
func (pg *PriceGenerator) GetPriceHistory(symbol string, limit int) ([]shared.PriceUpdate, error) {
	pg.mu.RLock()
	defer pg.mu.RUnlock()

	priceState, exists := pg.prices[symbol]
	if !exists {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}

	history := priceState.PriceHistory
	if limit > 0 && len(history) > limit {
		// Return the most recent entries
		start := len(history) - limit
		history = history[start:]
	}

	return history, nil
}

// SimulateVolatility applies volatility pattern to price generation
func (pg *PriceGenerator) SimulateVolatility(symbol string, pattern string, intensity float64) error {
	pg.mu.Lock()
	defer pg.mu.Unlock()

	_, exists := pg.prices[symbol]
	if !exists {
		return fmt.Errorf("symbol %s not found", symbol)
	}

	switch pattern {
	case "spike":
		pg.currentVolatility[symbol] = pg.config.BaseVolatility * (1 + intensity*3)
	case "decay":
		pg.currentVolatility[symbol] *= (1 - intensity*0.1)
	case "oscillate":
		oscFactor := math.Sin(float64(time.Now().Unix())/60) * intensity
		pg.currentVolatility[symbol] = pg.config.BaseVolatility * (1 + oscFactor)
	case "random":
		randomFactor := pg.rng.Float64() * intensity
		pg.currentVolatility[symbol] = pg.config.BaseVolatility * (1 + randomFactor*2)
	default:
		return fmt.Errorf("unknown volatility pattern: %s", pattern)
	}

	// Ensure volatility stays within reasonable bounds
	pg.currentVolatility[symbol] = math.Max(pg.currentVolatility[symbol], pg.config.BaseVolatility*0.1)
	pg.currentVolatility[symbol] = math.Min(pg.currentVolatility[symbol], pg.config.BaseVolatility*10)

	pg.logger.Info("Volatility applied",
		"symbol", symbol,
		"pattern", pattern,
		"intensity", intensity,
		"new_volatility", pg.currentVolatility[symbol],
	)

	return nil
}

// Private helper methods

func (pg *PriceGenerator) initializePriceState(symbol string, price float64) {
	now := time.Now()

	pg.prices[symbol] = &PriceState{
		Symbol:          symbol,
		CurrentPrice:    price,
		PreviousPrice:   price,
		BasePrice:       price,
		DailyOpen:       price,
		DailyHigh:       price,
		DailyLow:        price,
		Volume:          0,
		LastUpdate:      now,
		PriceHistory:    make([]shared.PriceUpdate, 0, pg.config.HistorySize),
		VolatilityIndex: pg.config.BaseVolatility,
		TrendDirection:  "sideways",
		TrendStrength:   0.0,
	}

	if pg.currentVolatility[symbol] == 0 {
		pg.currentVolatility[symbol] = pg.config.BaseVolatility
	}
}

func (pg *PriceGenerator) getCurrentVolatility(symbol string) float64 {
	if vol, exists := pg.currentVolatility[symbol]; exists {
		return vol
	}
	return pg.config.BaseVolatility
}

func (pg *PriceGenerator) calculateTimeFactor(timeSince time.Duration) float64 {
	// Convert to hours and apply square root scaling
	hours := timeSince.Hours()
	if hours <= 0 {
		return 0.1 // Minimum time factor
	}
	return math.Sqrt(hours)
}

func (pg *PriceGenerator) calculateTrendComponent(symbol string, timeFactor float64) float64 {
	priceState := pg.prices[symbol]

	// Simple trend calculation based on recent price movement
	if len(priceState.PriceHistory) < 5 {
		return 0.0
	}

	// Calculate moving average trend
	recentPrices := priceState.PriceHistory[len(priceState.PriceHistory)-5:]
	trend := 0.0
	for i := 1; i < len(recentPrices); i++ {
		change := (recentPrices[i].Price - recentPrices[i-1].Price) / recentPrices[i-1].Price
		trend += change
	}
	trend /= float64(len(recentPrices) - 1)

	// Apply persistence and time factor
	return trend * pg.config.TrendPersistence * timeFactor * 0.1
}

func (pg *PriceGenerator) calculateRandomComponent(volatility, timeFactor float64) float64 {
	// Generate normally distributed random walk
	randomValue := pg.rng.NormFloat64()

	// Scale by volatility and time
	return randomValue * volatility * timeFactor * 0.01
}

func (pg *PriceGenerator) calculateMeanReversionComponent(symbol string) float64 {
	basePrice, exists := pg.basePrices[symbol]
	if !exists {
		return 0.0
	}

	priceState := pg.prices[symbol]
	if priceState == nil {
		return 0.0
	}

	// Calculate deviation from base price
	deviation := (priceState.CurrentPrice - basePrice) / basePrice

	// Apply mean reversion force
	reversionForce := -deviation * pg.config.MeanReversion

	return reversionForce * 0.001 // Scale down
}

func (pg *PriceGenerator) roundToStepSize(price float64) float64 {
	if pg.config.PriceStepSize <= 0 {
		return price
	}
	return math.Round(price/pg.config.PriceStepSize) * pg.config.PriceStepSize
}

func (pg *PriceGenerator) generateVolume(symbol string, price float64, volatility float64) float64 {
	baseVolume := 1000.0

	// Volume increases with volatility
	volumeMultiplier := 1.0 + volatility*2

	// Add realistic random variation (log-normal distribution for trading volume)
	// Most trades are small, occasional large spikes
	randomFactor := math.Exp(pg.rng.NormFloat64() * 0.8) // Log-normal: mean ~1, wide variance

	// Occasional volume spikes (5% chance of 3-10x volume)
	if pg.rng.Float64() < 0.05 {
		randomFactor *= 3.0 + pg.rng.Float64()*7.0
	}

	// Occasional quiet periods (10% chance of very low volume)
	if pg.rng.Float64() < 0.10 {
		randomFactor *= 0.1 + pg.rng.Float64()*0.2
	}

	// Time-of-day pattern (higher volume during "market hours")
	hour := time.Now().Hour()
	timeMultiplier := 1.0
	if hour >= 9 && hour <= 16 { // Simulated market hours
		timeMultiplier = 1.5 + pg.rng.Float64()*0.5
	} else {
		timeMultiplier = 0.3 + pg.rng.Float64()*0.4
	}

	return baseVolume * volumeMultiplier * randomFactor * timeMultiplier
}

func (pg *PriceGenerator) updatePriceState(symbol string, newPrice, volume float64) {
	priceState := pg.prices[symbol]
	now := time.Now()

	// Update prices
	priceState.PreviousPrice = priceState.CurrentPrice
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
	priceUpdate := shared.PriceUpdate{
		Symbol:    symbol,
		Price:     newPrice,
		Volume:    volume,
		Timestamp: now,
	}

	priceState.PriceHistory = append(priceState.PriceHistory, priceUpdate)

	// Maintain history size limit
	if len(priceState.PriceHistory) > pg.config.HistorySize {
		priceState.PriceHistory = priceState.PriceHistory[1:]
	}

	// Reset daily values if new day
	if pg.isNewTradingDay(priceState.LastUpdate, now) {
		priceState.DailyOpen = newPrice
		priceState.DailyHigh = newPrice
		priceState.DailyLow = newPrice
	}

	// Decay volatility
	pg.decayVolatility(symbol)
}

func (pg *PriceGenerator) decayVolatility(symbol string) {
	if vol, exists := pg.currentVolatility[symbol]; exists {
		// Exponential decay towards base volatility
		decayFactor := 1.0 - pg.config.VolatilityDecay
		newVol := vol*decayFactor + pg.config.BaseVolatility*(1-decayFactor)
		pg.currentVolatility[symbol] = newVol
	}
}

func (pg *PriceGenerator) isNewTradingDay(lastUpdate, currentTime time.Time) bool {
	// Simple check: different day
	return lastUpdate.Day() != currentTime.Day()
}