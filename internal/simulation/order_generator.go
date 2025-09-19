package simulation

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"simulated_exchange/internal/api/dto"
)

// RealisticOrderGenerator implements OrderGenerator interface
type RealisticOrderGenerator struct {
	// User profiles and behavior
	userProfiles     []UserProfile
	marketSentiment  MarketSentiment
	mu               sync.RWMutex

	// Order generation state
	orderCount       int64
	lastOrderTime    time.Time
	orderStatistics  OrderStatistics

	// Behavior modeling
	behaviorModels   map[UserBehaviorPattern]*BehaviorModel
	sentimentFactors map[MarketSentiment]SentimentFactors

	// Market microstructure
	spreadModeling   SpreadModel
	liquidityModel   LiquidityModel

	// Random number generator
	rng *rand.Rand

	// Configuration
	config OrderGeneratorConfig
}

// BehaviorModel defines how different user types behave
type BehaviorModel struct {
	OrderSizeDistribution  DistributionParams `json:"order_size_distribution"`
	OrderFrequencyPattern  FrequencyPattern   `json:"order_frequency_pattern"`
	PriceOffsetBehavior    PriceOffsetBehavior `json:"price_offset_behavior"`
	TimeInForceBehavior    TimeInForceBehavior `json:"time_in_force_behavior"`
	MarketOrderRatio       float64            `json:"market_order_ratio"`
	BuySellBias            float64            `json:"buy_sell_bias"` // -1 (all sell) to 1 (all buy)
	ReactionTimeDistrib    DistributionParams `json:"reaction_time_distribution"`
}

// DistributionParams defines statistical distribution parameters
type DistributionParams struct {
	Type     string  `json:"type"`     // "normal", "lognormal", "exponential", "uniform"
	Mean     float64 `json:"mean"`
	StdDev   float64 `json:"std_dev"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Skewness float64 `json:"skewness"`
}

// FrequencyPattern defines order generation timing
type FrequencyPattern struct {
	BaseFrequency    float64            `json:"base_frequency"`    // orders per minute
	PeakHoursMultiplier float64         `json:"peak_hours_multiplier"`
	WeekendMultiplier   float64         `json:"weekend_multiplier"`
	VolatilityMultiplier float64        `json:"volatility_multiplier"`
	BurstProbability    float64         `json:"burst_probability"`
	BurstIntensity      float64         `json:"burst_intensity"`
}

// PriceOffsetBehavior defines how users price their orders
type PriceOffsetBehavior struct {
	MarketOrderProb     float64 `json:"market_order_probability"`
	LimitOrderOffsetPct float64 `json:"limit_order_offset_percentage"`
	AggressiveOrderProb float64 `json:"aggressive_order_probability"`
	PassiveOrderProb    float64 `json:"passive_order_probability"`
}

// TimeInForceBehavior defines order duration preferences
type TimeInForceBehavior struct {
	IOCProbability    float64       `json:"ioc_probability"`    // Immediate or Cancel
	GTCProbability    float64       `json:"gtc_probability"`    // Good Till Cancel
	GTDProbability    float64       `json:"gtd_probability"`    // Good Till Date
	DefaultDuration   time.Duration `json:"default_duration"`
}

// SentimentFactors define how market sentiment affects order generation
type SentimentFactors struct {
	OrderVolumeMultiplier  float64 `json:"order_volume_multiplier"`
	BuySellBias           float64 `json:"buy_sell_bias"`
	AggressivenessBoost   float64 `json:"aggressiveness_boost"`
	LiquidityConsumption  float64 `json:"liquidity_consumption"`
}

// SpreadModel models bid-ask spread behavior
type SpreadModel struct {
	BaseSpread        float64 `json:"base_spread"`
	VolatilityFactor  float64 `json:"volatility_factor"`
	LiquidityFactor   float64 `json:"liquidity_factor"`
	NewsFactor        float64 `json:"news_factor"`
}

// LiquidityModel models market liquidity
type LiquidityModel struct {
	DepthLevels       []LiquidityLevel `json:"depth_levels"`
	RegenerationRate  float64         `json:"regeneration_rate"`
	VolatilityImpact  float64         `json:"volatility_impact"`
}

// LiquidityLevel represents liquidity at a price level
type LiquidityLevel struct {
	PriceOffset   float64 `json:"price_offset"`   // Percentage from mid-price
	Volume        float64 `json:"volume"`         // Available volume
	RefreshRate   float64 `json:"refresh_rate"`   // How quickly it regenerates
}

// OrderGeneratorConfig configures order generation behavior
type OrderGeneratorConfig struct {
	BaseOrderRate      float64           `json:"base_order_rate"`
	MarketHoursBoost   float64           `json:"market_hours_boost"`
	VolatilityBoost    float64           `json:"volatility_boost"`
	NewsEventBoost     float64           `json:"news_event_boost"`
	UserTypeMix        map[string]float64 `json:"user_type_mix"`
	RandomSeed         int64             `json:"random_seed"`
	RealtimeMode       bool              `json:"realtime_mode"`
}

// NewRealisticOrderGenerator creates a new order generator
func NewRealisticOrderGenerator(config OrderGeneratorConfig) *RealisticOrderGenerator {
	if config.RandomSeed == 0 {
		config.RandomSeed = time.Now().UnixNano()
	}

	rog := &RealisticOrderGenerator{
		userProfiles:     DefaultUserProfiles(),
		marketSentiment:  SentimentNeutral,
		behaviorModels:   make(map[UserBehaviorPattern]*BehaviorModel),
		sentimentFactors: make(map[MarketSentiment]SentimentFactors),
		rng:              rand.New(rand.NewSource(config.RandomSeed)),
		config:           config,
		orderStatistics: OrderStatistics{
			OrdersBySymbol:   make(map[string]int64),
			OrdersBySide:     make(map[string]int64),
			OrdersByUserType: make(map[string]int64),
		},
	}

	// Initialize behavior models
	rog.initializeBehaviorModels()
	rog.initializeSentimentFactors()
	rog.initializeMarketMicrostructure()

	return rog
}

// GenerateRealisticOrders creates orders based on current market conditions
func (rog *RealisticOrderGenerator) GenerateRealisticOrders(symbol string, currentPrice float64, marketCondition MarketCondition) []dto.PlaceOrderRequest {
	rog.mu.Lock()
	defer rog.mu.Unlock()

	var orders []dto.PlaceOrderRequest

	// Calculate order generation rate based on market conditions
	generationRate := rog.calculateOrderGenerationRate(marketCondition, currentPrice)

	// Determine number of orders to generate this tick
	numOrders := rog.sampleOrderCount(generationRate)

	// Generate orders for each user profile
	for i := 0; i < numOrders; i++ {
		userProfile := rog.selectUserProfile()
		order := rog.generateOrderForUser(symbol, currentPrice, userProfile, marketCondition)

		if order != nil {
			orders = append(orders, *order)
		}
	}

	// Update statistics
	rog.updateOrderStatistics(symbol, orders)

	return orders
}

// SimulateUserBehavior generates orders based on user behavior patterns
func (rog *RealisticOrderGenerator) SimulateUserBehavior(pattern UserBehaviorPattern, intensity float64) []dto.PlaceOrderRequest {
	rog.mu.Lock()
	defer rog.mu.Unlock()

	var orders []dto.PlaceOrderRequest

	// Get behavior model for pattern
	behaviorModel, exists := rog.behaviorModels[pattern]
	if !exists {
		return orders
	}

	// Calculate number of orders based on intensity
	baseCount := int(intensity * 10) // Scale intensity to order count
	orderCount := rog.rng.Intn(baseCount) + 1

	// Generate orders according to behavior pattern
	for i := 0; i < orderCount; i++ {
		// Select a symbol (simplified - could be more sophisticated)
		symbols := []string{"BTCUSD", "ETHUSD", "ADAUSD"}
		symbol := symbols[rog.rng.Intn(len(symbols))]
		currentPrice := 50000.0 // Simplified - would get from price generator

		order := rog.generateBehaviorDrivenOrder(symbol, currentPrice, pattern, behaviorModel, intensity)
		if order != nil {
			orders = append(orders, *order)
		}
	}

	return orders
}

// SetUserProfiles configures different types of simulated users
func (rog *RealisticOrderGenerator) SetUserProfiles(profiles []UserProfile) {
	rog.mu.Lock()
	defer rog.mu.Unlock()

	rog.userProfiles = profiles
}

// GetOrderStatistics returns statistics about generated orders
func (rog *RealisticOrderGenerator) GetOrderStatistics() OrderStatistics {
	rog.mu.RLock()
	defer rog.mu.RUnlock()

	// Return a copy
	stats := rog.orderStatistics
	stats.LastGenerationTime = rog.lastOrderTime
	return stats
}

// UpdateMarketSentiment changes overall market sentiment
func (rog *RealisticOrderGenerator) UpdateMarketSentiment(sentiment MarketSentiment) {
	rog.mu.Lock()
	defer rog.mu.Unlock()

	rog.marketSentiment = sentiment
}

// Private helper methods

func (rog *RealisticOrderGenerator) initializeBehaviorModels() {
	// Conservative behavior model
	rog.behaviorModels[BehaviorConservative] = &BehaviorModel{
		OrderSizeDistribution: DistributionParams{
			Type:   "lognormal",
			Mean:   500.0,
			StdDev: 200.0,
			Min:    10.0,
			Max:    2000.0,
		},
		OrderFrequencyPattern: FrequencyPattern{
			BaseFrequency:        0.5, // 0.5 orders per minute
			PeakHoursMultiplier:  1.2,
			WeekendMultiplier:    0.8,
			VolatilityMultiplier: 0.9, // Less active during volatility
			BurstProbability:     0.1,
			BurstIntensity:       2.0,
		},
		PriceOffsetBehavior: PriceOffsetBehavior{
			MarketOrderProb:     0.2,
			LimitOrderOffsetPct: 0.5, // 0.5% away from market
			AggressiveOrderProb: 0.3,
			PassiveOrderProb:    0.7,
		},
		MarketOrderRatio: 0.2,
		BuySellBias:      0.0, // Neutral
		ReactionTimeDistrib: DistributionParams{
			Type:   "exponential",
			Mean:   5.0, // 5 seconds average reaction time
			StdDev: 3.0,
			Min:    1.0,
			Max:    30.0,
		},
	}

	// Aggressive behavior model
	rog.behaviorModels[BehaviorAggressive] = &BehaviorModel{
		OrderSizeDistribution: DistributionParams{
			Type:   "normal",
			Mean:   2000.0,
			StdDev: 1000.0,
			Min:    100.0,
			Max:    20000.0,
		},
		OrderFrequencyPattern: FrequencyPattern{
			BaseFrequency:        3.0, // 3 orders per minute
			PeakHoursMultiplier:  1.5,
			WeekendMultiplier:    1.0,
			VolatilityMultiplier: 1.8, // More active during volatility
			BurstProbability:     0.3,
			BurstIntensity:       5.0,
		},
		PriceOffsetBehavior: PriceOffsetBehavior{
			MarketOrderProb:     0.6,
			LimitOrderOffsetPct: 0.1, // 0.1% away from market
			AggressiveOrderProb: 0.8,
			PassiveOrderProb:    0.2,
		},
		MarketOrderRatio: 0.6,
		BuySellBias:      0.0,
		ReactionTimeDistrib: DistributionParams{
			Type:   "exponential",
			Mean:   0.5, // 0.5 seconds average reaction time
			StdDev: 0.3,
			Min:    0.1,
			Max:    3.0,
		},
	}

	// Momentum behavior model
	rog.behaviorModels[BehaviorMomentum] = &BehaviorModel{
		OrderSizeDistribution: DistributionParams{
			Type:   "normal",
			Mean:   1000.0,
			StdDev: 500.0,
			Min:    50.0,
			Max:    10000.0,
		},
		OrderFrequencyPattern: FrequencyPattern{
			BaseFrequency:        1.5,
			PeakHoursMultiplier:  1.3,
			WeekendMultiplier:    0.9,
			VolatilityMultiplier: 2.0, // Very active during trends
			BurstProbability:     0.4,
			BurstIntensity:       3.0,
		},
		PriceOffsetBehavior: PriceOffsetBehavior{
			MarketOrderProb:     0.5,
			LimitOrderOffsetPct: 0.2,
			AggressiveOrderProb: 0.7,
			PassiveOrderProb:    0.3,
		},
		MarketOrderRatio: 0.5,
		BuySellBias:      0.2, // Slight buy bias (momentum followers)
		ReactionTimeDistrib: DistributionParams{
			Type:   "normal",
			Mean:   2.0,
			StdDev: 1.0,
			Min:    0.5,
			Max:    10.0,
		},
	}

	// FOMO behavior model
	rog.behaviorModels[BehaviorFOMO] = &BehaviorModel{
		OrderSizeDistribution: DistributionParams{
			Type:   "lognormal",
			Mean:   1500.0,
			StdDev: 1200.0,
			Min:    100.0,
			Max:    50000.0,
		},
		OrderFrequencyPattern: FrequencyPattern{
			BaseFrequency:        2.0,
			PeakHoursMultiplier:  2.0,
			WeekendMultiplier:    1.5,
			VolatilityMultiplier: 3.0, // Extremely active during volatility
			BurstProbability:     0.6,
			BurstIntensity:       8.0,
		},
		PriceOffsetBehavior: PriceOffsetBehavior{
			MarketOrderProb:     0.8, // Mostly market orders
			LimitOrderOffsetPct: 0.05,
			AggressiveOrderProb: 0.9,
			PassiveOrderProb:    0.1,
		},
		MarketOrderRatio: 0.8,
		BuySellBias:      0.6, // Strong buy bias during FOMO
		ReactionTimeDistrib: DistributionParams{
			Type:   "exponential",
			Mean:   0.3,
			StdDev: 0.2,
			Min:    0.05,
			Max:    2.0,
		},
	}

	// Panic behavior model
	rog.behaviorModels[BehaviorPanic] = &BehaviorModel{
		OrderSizeDistribution: DistributionParams{
			Type:   "lognormal",
			Mean:   3000.0,
			StdDev: 2000.0,
			Min:    500.0,
			Max:    100000.0,
		},
		OrderFrequencyPattern: FrequencyPattern{
			BaseFrequency:        4.0,
			PeakHoursMultiplier:  1.0,
			WeekendMultiplier:    1.0,
			VolatilityMultiplier: 5.0, // Panic during volatility
			BurstProbability:     0.8,
			BurstIntensity:       10.0,
		},
		PriceOffsetBehavior: PriceOffsetBehavior{
			MarketOrderProb:     0.9, // Almost all market orders
			LimitOrderOffsetPct: 0.02,
			AggressiveOrderProb: 0.95,
			PassiveOrderProb:    0.05,
		},
		MarketOrderRatio: 0.9,
		BuySellBias:      -0.8, // Strong sell bias during panic
		ReactionTimeDistrib: DistributionParams{
			Type:   "exponential",
			Mean:   0.1,
			StdDev: 0.05,
			Min:    0.01,
			Max:    1.0,
		},
	}

	// Mean reversion behavior model
	rog.behaviorModels[BehaviorMeanRevert] = &BehaviorModel{
		OrderSizeDistribution: DistributionParams{
			Type:   "normal",
			Mean:   800.0,
			StdDev: 400.0,
			Min:    50.0,
			Max:    5000.0,
		},
		OrderFrequencyPattern: FrequencyPattern{
			BaseFrequency:        1.0,
			PeakHoursMultiplier:  1.1,
			WeekendMultiplier:    0.9,
			VolatilityMultiplier: 1.5, // More active during extremes
			BurstProbability:     0.2,
			BurstIntensity:       2.5,
		},
		PriceOffsetBehavior: PriceOffsetBehavior{
			MarketOrderProb:     0.3,
			LimitOrderOffsetPct: 1.0, // Far from market
			AggressiveOrderProb: 0.2,
			PassiveOrderProb:    0.8,
		},
		MarketOrderRatio: 0.3,
		BuySellBias:      0.0, // Contrarian - bias depends on market direction
		ReactionTimeDistrib: DistributionParams{
			Type:   "normal",
			Mean:   10.0,
			StdDev: 5.0,
			Min:    2.0,
			Max:    60.0,
		},
	}

	// Arbitrage behavior model
	rog.behaviorModels[BehaviorArbitrage] = &BehaviorModel{
		OrderSizeDistribution: DistributionParams{
			Type:   "normal",
			Mean:   10000.0,
			StdDev: 5000.0,
			Min:    1000.0,
			Max:    100000.0,
		},
		OrderFrequencyPattern: FrequencyPattern{
			BaseFrequency:        0.3, // Less frequent but larger
			PeakHoursMultiplier:  1.1,
			WeekendMultiplier:    1.0,
			VolatilityMultiplier: 0.8, // Less affected by volatility
			BurstProbability:     0.1,
			BurstIntensity:       1.5,
		},
		PriceOffsetBehavior: PriceOffsetBehavior{
			MarketOrderProb:     0.1,
			LimitOrderOffsetPct: 0.05, // Very tight spreads
			AggressiveOrderProb: 0.3,
			PassiveOrderProb:    0.7,
		},
		MarketOrderRatio: 0.1,
		BuySellBias:      0.0, // Perfectly balanced
		ReactionTimeDistrib: DistributionParams{
			Type:   "exponential",
			Mean:   0.1,
			StdDev: 0.05,
			Min:    0.01,
			Max:    1.0,
		},
	}
}

func (rog *RealisticOrderGenerator) initializeSentimentFactors() {
	rog.sentimentFactors[SentimentOptimistic] = SentimentFactors{
		OrderVolumeMultiplier: 1.3,
		BuySellBias:          0.4,
		AggressivenessBoost:  0.2,
		LiquidityConsumption: 1.2,
	}

	rog.sentimentFactors[SentimentNeutral] = SentimentFactors{
		OrderVolumeMultiplier: 1.0,
		BuySellBias:          0.0,
		AggressivenessBoost:  0.0,
		LiquidityConsumption: 1.0,
	}

	rog.sentimentFactors[SentimentPessimistic] = SentimentFactors{
		OrderVolumeMultiplier: 0.8,
		BuySellBias:          -0.3,
		AggressivenessBoost:  -0.1,
		LiquidityConsumption: 0.9,
	}

	rog.sentimentFactors[SentimentFearful] = SentimentFactors{
		OrderVolumeMultiplier: 0.6,
		BuySellBias:          -0.6,
		AggressivenessBoost:  -0.3,
		LiquidityConsumption: 0.7,
	}

	rog.sentimentFactors[SentimentGreedy] = SentimentFactors{
		OrderVolumeMultiplier: 1.8,
		BuySellBias:          0.7,
		AggressivenessBoost:  0.5,
		LiquidityConsumption: 1.6,
	}
}

func (rog *RealisticOrderGenerator) initializeMarketMicrostructure() {
	rog.spreadModeling = SpreadModel{
		BaseSpread:       0.001, // 0.1% base spread
		VolatilityFactor: 2.0,
		LiquidityFactor:  0.5,
		NewsFactor:       1.5,
	}

	rog.liquidityModel = LiquidityModel{
		DepthLevels: []LiquidityLevel{
			{PriceOffset: 0.001, Volume: 10000, RefreshRate: 0.9},
			{PriceOffset: 0.002, Volume: 20000, RefreshRate: 0.8},
			{PriceOffset: 0.005, Volume: 50000, RefreshRate: 0.7},
			{PriceOffset: 0.01, Volume: 100000, RefreshRate: 0.6},
		},
		RegenerationRate: 0.1,
		VolatilityImpact: 0.3,
	}
}

func (rog *RealisticOrderGenerator) calculateOrderGenerationRate(marketCondition MarketCondition, currentPrice float64) float64 {
	baseRate := rog.config.BaseOrderRate

	// Adjust based on market condition
	var conditionMultiplier float64
	switch marketCondition {
	case MarketSteady:
		conditionMultiplier = 1.0
	case MarketVolatile:
		conditionMultiplier = 2.0
	case MarketBullish:
		conditionMultiplier = 1.5
	case MarketBearish:
		conditionMultiplier = 1.3
	case MarketCrash:
		conditionMultiplier = 3.0
	case MarketRecovery:
		conditionMultiplier = 1.8
	default:
		conditionMultiplier = 1.0
	}

	// Apply market sentiment
	sentimentFactors := rog.sentimentFactors[rog.marketSentiment]
	sentimentMultiplier := sentimentFactors.OrderVolumeMultiplier

	// Time-based adjustments (market hours, etc.)
	timeMultiplier := rog.calculateTimeMultiplier()

	return baseRate * conditionMultiplier * sentimentMultiplier * timeMultiplier
}

func (rog *RealisticOrderGenerator) calculateTimeMultiplier() float64 {
	now := time.Now()
	hour := now.Hour()

	// Peak hours: 9-11 AM and 2-4 PM (simplified)
	if (hour >= 9 && hour <= 11) || (hour >= 14 && hour <= 16) {
		return rog.config.MarketHoursBoost
	}

	// Weekend trading (crypto markets)
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return 0.8 // Reduced weekend activity
	}

	return 1.0
}

func (rog *RealisticOrderGenerator) sampleOrderCount(rate float64) int {
	// Use Poisson distribution for order arrival
	lambda := rate
	if lambda <= 0 {
		return 0
	}

	// Simplified Poisson sampling using exponential intervals
	count := 0
	timeSum := 0.0

	for timeSum < 1.0 { // 1 minute interval
		timeSum += -math.Log(rog.rng.Float64()) / lambda
		if timeSum < 1.0 {
			count++
		}
	}

	return count
}

func (rog *RealisticOrderGenerator) selectUserProfile() UserProfile {
	totalWeight := 0.0
	for _, profile := range rog.userProfiles {
		totalWeight += profile.PopulationWeight
	}

	r := rog.rng.Float64() * totalWeight
	cumWeight := 0.0

	for _, profile := range rog.userProfiles {
		cumWeight += profile.PopulationWeight
		if r <= cumWeight {
			return profile
		}
	}

	// Fallback to first profile
	return rog.userProfiles[0]
}

func (rog *RealisticOrderGenerator) generateOrderForUser(symbol string, currentPrice float64, userProfile UserProfile, marketCondition MarketCondition) *dto.PlaceOrderRequest {
	behaviorModel := rog.behaviorModels[userProfile.BehaviorPattern]
	if behaviorModel == nil {
		return nil
	}

	// Determine order side
	side := rog.determineOrderSide(userProfile, behaviorModel, marketCondition)

	// Determine order size
	orderSize := rog.sampleOrderSize(behaviorModel.OrderSizeDistribution, userProfile)

	// Determine order type and price
	orderType, price := rog.determineOrderTypeAndPrice(symbol, currentPrice, behaviorModel, side)

	// Create order request
	order := &dto.PlaceOrderRequest{
		Symbol:   symbol,
		Side:     side,
		Quantity: orderSize,
		Type:     orderType,
		Price:    price, // Always set price (market orders will use current price)
	}

	rog.orderCount++
	rog.lastOrderTime = time.Now()

	return order
}

func (rog *RealisticOrderGenerator) generateBehaviorDrivenOrder(symbol string, currentPrice float64, pattern UserBehaviorPattern, behaviorModel *BehaviorModel, intensity float64) *dto.PlaceOrderRequest {
	// Apply intensity to behavior parameters
	adjustedModel := *behaviorModel // Copy
	adjustedModel.OrderFrequencyPattern.BurstIntensity *= intensity
	adjustedModel.PriceOffsetBehavior.AggressiveOrderProb *= (1 + intensity*0.5)

	// Generate order with adjusted behavior
	side := rog.determineOrderSideBehaviorDriven(pattern, intensity)
	orderSize := rog.sampleOrderSize(adjustedModel.OrderSizeDistribution, UserProfile{})
	orderType, price := rog.determineOrderTypeAndPrice(symbol, currentPrice, &adjustedModel, side)

	order := &dto.PlaceOrderRequest{
		Symbol:   symbol,
		Side:     side,
		Quantity: orderSize,
		Type:     orderType,
		Price:    price,
	}

	return order
}

func (rog *RealisticOrderGenerator) determineOrderSide(userProfile UserProfile, behaviorModel *BehaviorModel, marketCondition MarketCondition) string {
	// Base bias from behavior model
	bias := behaviorModel.BuySellBias

	// Apply market sentiment
	sentimentFactors := rog.sentimentFactors[rog.marketSentiment]
	bias += sentimentFactors.BuySellBias

	// Apply market condition bias
	switch marketCondition {
	case MarketBullish:
		bias += 0.3
	case MarketBearish:
		bias -= 0.3
	case MarketCrash:
		bias -= 0.5 // Panic selling
	case MarketRecovery:
		bias += 0.2 // Buying the dip
	}

	// Add some randomness
	bias += (rog.rng.Float64() - 0.5) * 0.2

	// Convert bias to buy/sell decision
	if bias > 0 {
		if rog.rng.Float64() < (bias+1)/2 {
			return "buy"
		}
		return "sell"
	} else {
		if rog.rng.Float64() < (1+bias)/2 {
			return "buy"
		}
		return "sell"
	}
}

func (rog *RealisticOrderGenerator) determineOrderSideBehaviorDriven(pattern UserBehaviorPattern, intensity float64) string {
	switch pattern {
	case BehaviorFOMO:
		// FOMO is usually buying
		if rog.rng.Float64() < 0.8+intensity*0.1 {
			return "buy"
		}
		return "sell"
	case BehaviorPanic:
		// Panic is usually selling
		if rog.rng.Float64() < 0.8+intensity*0.1 {
			return "sell"
		}
		return "buy"
	case BehaviorMomentum:
		// Momentum follows the trend (simplified as random for now)
		if rog.rng.Float64() < 0.6 {
			return "buy"
		}
		return "sell"
	case BehaviorMeanRevert:
		// Mean reversion goes against the trend (simplified as random for now)
		if rog.rng.Float64() < 0.4 {
			return "buy"
		}
		return "sell"
	default:
		if rog.rng.Float64() < 0.5 {
			return "buy"
		}
		return "sell"
	}
}

func (rog *RealisticOrderGenerator) sampleOrderSize(distribution DistributionParams, userProfile UserProfile) float64 {
	var size float64

	switch distribution.Type {
	case "normal":
		size = rog.rng.NormFloat64()*distribution.StdDev + distribution.Mean
	case "lognormal":
		// Generate lognormal distribution
		normal := rog.rng.NormFloat64()
		size = math.Exp(math.Log(distribution.Mean) + normal*distribution.StdDev)
	case "exponential":
		size = -math.Log(rog.rng.Float64()) * distribution.Mean
	case "uniform":
		size = rog.rng.Float64()*(distribution.Max-distribution.Min) + distribution.Min
	default:
		size = distribution.Mean
	}

	// Clamp to min/max bounds
	if size < distribution.Min {
		size = distribution.Min
	}
	if size > distribution.Max {
		size = distribution.Max
	}

	// Apply user wealth constraints
	if userProfile.Wealth > 0 {
		maxAffordable := userProfile.Wealth * 0.1 // Max 10% of wealth per order
		if size > maxAffordable {
			size = maxAffordable
		}
	}

	return math.Round(size*100) / 100 // Round to 2 decimal places
}

func (rog *RealisticOrderGenerator) determineOrderTypeAndPrice(symbol string, currentPrice float64, behaviorModel *BehaviorModel, side string) (string, float64) {
	// Determine if this should be a market order
	if rog.rng.Float64() < behaviorModel.PriceOffsetBehavior.MarketOrderProb {
		return "market", currentPrice
	}

	// Generate limit order with price offset
	offsetPct := behaviorModel.PriceOffsetBehavior.LimitOrderOffsetPct

	// Add randomness to offset
	offsetPct *= (0.5 + rog.rng.Float64()) // 0.5x to 1.5x the base offset

	var price float64
	if side == "BUY" {
		// Buy orders below market price
		price = currentPrice * (1 - offsetPct/100)
	} else {
		// Sell orders above market price
		price = currentPrice * (1 + offsetPct/100)
	}

	return "limit", math.Round(price*100) / 100
}

func (rog *RealisticOrderGenerator) updateOrderStatistics(symbol string, orders []dto.PlaceOrderRequest) {
	rog.orderStatistics.TotalOrders += int64(len(orders))
	rog.orderStatistics.OrdersBySymbol[symbol] += int64(len(orders))

	totalSize := 0.0
	for _, order := range orders {
		rog.orderStatistics.OrdersBySide[order.Side]++
		totalSize += order.Quantity
	}

	// Update average order size (exponential moving average)
	if rog.orderStatistics.TotalOrders > 0 {
		newAvg := totalSize / float64(len(orders))
		if rog.orderStatistics.AverageOrderSize == 0 {
			rog.orderStatistics.AverageOrderSize = newAvg
		} else {
			alpha := 0.1 // Smoothing factor
			rog.orderStatistics.AverageOrderSize = alpha*newAvg + (1-alpha)*rog.orderStatistics.AverageOrderSize
		}
	}
}

// DefaultOrderGeneratorConfig returns default configuration
func DefaultOrderGeneratorConfig() OrderGeneratorConfig {
	return OrderGeneratorConfig{
		BaseOrderRate:    5.0, // 5 orders per minute base rate
		MarketHoursBoost: 1.5,
		VolatilityBoost:  2.0,
		NewsEventBoost:   3.0,
		UserTypeMix: map[string]float64{
			"conservative": 0.6,
			"aggressive":   0.25,
			"institutional": 0.15,
		},
		RandomSeed:   0,
		RealtimeMode: true,
	}
}