package simulation

import (
	"context"
	"time"

	"simulated_exchange/internal/api/dto"
)

// TradingEngine interface for simulation integration
type TradingEngine interface {
	// PlaceOrder places an order in the trading engine
	PlaceOrder(order dto.PlaceOrderRequest) (interface{}, error)
}

// MarketSimulator interface defines the main simulation control
type MarketSimulator interface {
	// StartSimulation begins the market simulation with given parameters
	StartSimulation(ctx context.Context, config SimulationConfig) error

	// StopSimulation gracefully stops the simulation
	StopSimulation() error

	// InjectVolatility introduces artificial volatility into the market
	InjectVolatility(pattern VolatilityPattern, duration time.Duration) error

	// GetSimulationStatus returns current simulation state
	GetSimulationStatus() SimulationStatus

	// UpdateConfig updates simulation parameters during runtime
	UpdateConfig(config SimulationConfig) error
}

// PriceGenerator interface defines price generation behavior
type PriceGenerator interface {
	// GeneratePrice creates next price based on current market conditions
	GeneratePrice(symbol string, currentPrice float64, timeElapsed time.Duration) float64

	// SimulateVolatility applies volatility pattern to price generation
	SimulateVolatility(pattern VolatilityPattern, intensity float64)

	// GetPriceTrend returns current price trend information
	GetPriceTrend(symbol string) PriceTrend

	// SetBasePrice establishes baseline price for symbol
	SetBasePrice(symbol string, price float64)

	// Reset clears all price history and patterns
	Reset()
}

// OrderGenerator interface defines order generation behavior
type OrderGenerator interface {
	// GenerateRealisticOrders creates orders based on current market conditions
	GenerateRealisticOrders(symbol string, currentPrice float64, marketCondition MarketCondition) []dto.PlaceOrderRequest

	// SimulateUserBehavior generates orders based on user behavior patterns
	SimulateUserBehavior(pattern UserBehaviorPattern, intensity float64) []dto.PlaceOrderRequest

	// SetUserProfiles configures different types of simulated users
	SetUserProfiles(profiles []UserProfile)

	// GetOrderStatistics returns statistics about generated orders
	GetOrderStatistics() OrderStatistics

	// UpdateMarketSentiment changes overall market sentiment
	UpdateMarketSentiment(sentiment MarketSentiment)
}

// EventGenerator interface for market events and news simulation
type EventGenerator interface {
	// GenerateMarketEvent creates random market events
	GenerateMarketEvent() MarketEvent

	// InjectEvent manually introduces a specific event
	InjectEvent(event MarketEvent) error

	// GetActiveEvents returns currently active market events
	GetActiveEvents() []MarketEvent

	// SetEventProbability configures event generation probabilities
	SetEventProbability(eventType EventType, probability float64)
}

// Configuration Structures

// SimulationConfig defines simulation parameters
type SimulationConfig struct {
	// Basic parameters
	Symbols               []string      `json:"symbols"`
	SimulationDuration    time.Duration `json:"simulation_duration"`
	TickInterval          time.Duration `json:"tick_interval"`
	OrderGenerationRate   float64       `json:"order_generation_rate"`   // orders per second
	PriceUpdateFrequency  time.Duration `json:"price_update_frequency"`

	// Market conditions
	BaseVolatility        float64            `json:"base_volatility"`        // 0.0 to 1.0
	TradingHours          TradingHours       `json:"trading_hours"`
	InitialPrices         map[string]float64 `json:"initial_prices"`
	MarketCondition       MarketCondition    `json:"market_condition"`

	// User behavior
	UserProfiles          []UserProfile      `json:"user_profiles"`
	MarketSentiment       MarketSentiment    `json:"market_sentiment"`
	NewsEventFrequency    time.Duration      `json:"news_event_frequency"`

	// Advanced settings
	EnablePatterns        bool               `json:"enable_patterns"`
	PatternProbabilities  map[string]float64 `json:"pattern_probabilities"`
	ReactionDelays        ReactionDelays     `json:"reaction_delays"`
}

// SimulationStatus represents current simulation state
type SimulationStatus struct {
	IsRunning           bool               `json:"is_running"`
	StartTime           time.Time          `json:"start_time"`
	RunningDuration     time.Duration      `json:"running_duration"`
	OrdersGenerated     int64              `json:"orders_generated"`
	PriceUpdates        int64              `json:"price_updates"`
	CurrentPrices       map[string]float64 `json:"current_prices"`
	ActivePatterns      []string           `json:"active_patterns"`
	MarketCondition     MarketCondition    `json:"market_condition"`
	LastError           error              `json:"last_error,omitempty"`
}

// Market Condition Enums and Types

// MarketCondition represents overall market state
type MarketCondition string

const (
	MarketSteady      MarketCondition = "STEADY"      // Normal trading conditions
	MarketVolatile    MarketCondition = "VOLATILE"    // High volatility period
	MarketBullish     MarketCondition = "BULLISH"     // Rising market trend
	MarketBearish     MarketCondition = "BEARISH"     // Falling market trend
	MarketSideways    MarketCondition = "SIDEWAYS"    // Range-bound trading
	MarketCrash       MarketCondition = "CRASH"       // Sudden dramatic decline
	MarketRecovery    MarketCondition = "RECOVERY"    // Recovery after crash
)

// VolatilityPattern defines volatility injection patterns
type VolatilityPattern string

const (
	VolatilitySpike     VolatilityPattern = "SPIKE"       // Sudden volatility increase
	VolatilityDecay     VolatilityPattern = "DECAY"       // Gradual volatility decrease
	VolatilityOscillate VolatilityPattern = "OSCILLATE"   // Cyclical volatility changes
	VolatilityRandom    VolatilityPattern = "RANDOM"      // Random volatility bursts
	VolatilityNews      VolatilityPattern = "NEWS"        // News-driven volatility
)

// UserBehaviorPattern defines simulated user behavior types
type UserBehaviorPattern string

const (
	BehaviorConservative UserBehaviorPattern = "CONSERVATIVE" // Risk-averse trading
	BehaviorAggressive   UserBehaviorPattern = "AGGRESSIVE"   // High-frequency trading
	BehaviorMomentum     UserBehaviorPattern = "MOMENTUM"     // Trend-following
	BehaviorMeanRevert   UserBehaviorPattern = "MEAN_REVERT"  // Contrarian trading
	BehaviorFOMO          UserBehaviorPattern = "FOMO"         // Fear of missing out
	BehaviorPanic        UserBehaviorPattern = "PANIC"        // Panic selling
	BehaviorArbitrage    UserBehaviorPattern = "ARBITRAGE"    // Price difference exploitation
)

// MarketSentiment represents overall market mood
type MarketSentiment string

const (
	SentimentOptimistic MarketSentiment = "OPTIMISTIC" // Positive market outlook
	SentimentNeutral    MarketSentiment = "NEUTRAL"    // Balanced sentiment
	SentimentPessimistic MarketSentiment = "PESSIMISTIC" // Negative market outlook
	SentimentFearful    MarketSentiment = "FEARFUL"    // High fear levels
	SentimentGreedy     MarketSentiment = "GREEDY"     // Excessive optimism
)

// Data Structures

// UserProfile defines characteristics of simulated users
type UserProfile struct {
	Name                string              `json:"name"`
	BehaviorPattern     UserBehaviorPattern `json:"behavior_pattern"`
	RiskTolerance       float64             `json:"risk_tolerance"`       // 0.0 to 1.0
	OrderSizeRange      OrderSizeRange      `json:"order_size_range"`
	TradingFrequency    float64             `json:"trading_frequency"`    // orders per minute
	PreferredSymbols    []string            `json:"preferred_symbols"`
	ReactionTime        time.Duration       `json:"reaction_time"`
	Wealth              float64             `json:"wealth"`               // Available capital
	PopulationWeight    float64             `json:"population_weight"`    // Percentage of total users
}

// OrderSizeRange defines order size distribution
type OrderSizeRange struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Mean     float64 `json:"mean"`
	StdDev   float64 `json:"std_dev"`
}

// PriceTrend represents price movement analysis
type PriceTrend struct {
	Symbol              string        `json:"symbol"`
	Direction           TrendDirection `json:"direction"`
	Strength            float64       `json:"strength"`           // 0.0 to 1.0
	Duration            time.Duration `json:"duration"`
	VolatilityLevel     float64       `json:"volatility_level"`   // Current volatility
	SupportLevel        float64       `json:"support_level"`
	ResistanceLevel     float64       `json:"resistance_level"`
	LastUpdate          time.Time     `json:"last_update"`
}

// TrendDirection represents price trend direction
type TrendDirection string

const (
	TrendUp       TrendDirection = "UP"
	TrendDown     TrendDirection = "DOWN"
	TrendSideways TrendDirection = "SIDEWAYS"
)

// OrderStatistics tracks order generation metrics
type OrderStatistics struct {
	TotalOrders         int64              `json:"total_orders"`
	OrdersBySymbol      map[string]int64   `json:"orders_by_symbol"`
	OrdersBySide        map[string]int64   `json:"orders_by_side"`
	OrdersByUserType    map[string]int64   `json:"orders_by_user_type"`
	AverageOrderSize    float64            `json:"average_order_size"`
	AverageSpread       float64            `json:"average_spread"`
	LastGenerationTime  time.Time          `json:"last_generation_time"`
}

// MarketEvent represents external events affecting the market
type MarketEvent struct {
	ID               string        `json:"id"`
	Type             EventType     `json:"type"`
	Severity         EventSeverity `json:"severity"`
	AffectedSymbols  []string      `json:"affected_symbols"`
	PriceImpact      float64       `json:"price_impact"`      // Percentage impact
	Duration         time.Duration `json:"duration"`
	Description      string        `json:"description"`
	StartTime        time.Time     `json:"start_time"`
	IsActive         bool          `json:"is_active"`
}

// EventType defines types of market events
type EventType string

const (
	EventEarnings     EventType = "EARNINGS"     // Earnings announcements
	EventNews         EventType = "NEWS"         // General news events
	EventRegulatory   EventType = "REGULATORY"   // Regulatory changes
	EventEconomic     EventType = "ECONOMIC"     // Economic indicators
	EventGeopolitical EventType = "GEOPOLITICAL" // Political events
	EventTechnical    EventType = "TECHNICAL"    // Technical issues
	EventCorporate    EventType = "CORPORATE"    // Corporate actions
)

// EventSeverity defines impact level of market events
type EventSeverity string

const (
	SeverityLow    EventSeverity = "LOW"    // Minor impact
	SeverityMedium EventSeverity = "MEDIUM" // Moderate impact
	SeverityHigh   EventSeverity = "HIGH"   // Major impact
	SeverityCrisis EventSeverity = "CRISIS" // Market-moving impact
)

// TradingHours defines when market is active
type TradingHours struct {
	OpenTime      time.Time `json:"open_time"`
	CloseTime     time.Time `json:"close_time"`
	TimeZone      string    `json:"timezone"`
	PreMarket     bool      `json:"pre_market"`
	AfterHours    bool      `json:"after_hours"`
	WeekendTrading bool     `json:"weekend_trading"`
}

// ReactionDelays defines how quickly users react to different events
type ReactionDelays struct {
	PriceChange     time.Duration `json:"price_change"`
	NewsEvent       time.Duration `json:"news_event"`
	LargeOrder      time.Duration `json:"large_order"`
	MarketCondition time.Duration `json:"market_condition"`
}

// Pattern Configuration

// SimulationPattern represents predefined simulation scenarios
type SimulationPattern struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Duration      time.Duration          `json:"duration"`
	Phases        []PatternPhase         `json:"phases"`
	Triggers      []PatternTrigger       `json:"triggers"`
	Parameters    map[string]interface{} `json:"parameters"`
}

// PatternPhase represents a phase within a simulation pattern
type PatternPhase struct {
	Name            string          `json:"name"`
	Duration        time.Duration   `json:"duration"`
	MarketCondition MarketCondition `json:"market_condition"`
	VolatilityLevel float64         `json:"volatility_level"`
	OrderIntensity  float64         `json:"order_intensity"`
	PriceDirection  TrendDirection  `json:"price_direction"`
	UserBehavior    map[UserBehaviorPattern]float64 `json:"user_behavior"`
}

// PatternTrigger defines conditions that activate pattern changes
type PatternTrigger struct {
	Type        TriggerType            `json:"type"`
	Condition   string                 `json:"condition"`
	Parameters  map[string]interface{} `json:"parameters"`
	NextPhase   string                 `json:"next_phase"`
}

// TriggerType defines different trigger mechanisms
type TriggerType string

const (
	TriggerTime         TriggerType = "TIME"          // Time-based trigger
	TriggerPrice        TriggerType = "PRICE"         // Price-based trigger
	TriggerVolume       TriggerType = "VOLUME"        // Volume-based trigger
	TriggerVolatility   TriggerType = "VOLATILITY"    // Volatility-based trigger
	TriggerEvent        TriggerType = "EVENT"         // Event-based trigger
	TriggerRandom       TriggerType = "RANDOM"        // Random trigger
)

// Error Types

// SimulationError represents simulation-specific errors
type SimulationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func (e SimulationError) Error() string {
	return e.Message
}

// Common error codes
const (
	ErrorCodeInvalidConfig    = "INVALID_CONFIG"
	ErrorCodeSimulationFailed = "SIMULATION_FAILED"
	ErrorCodePatternNotFound  = "PATTERN_NOT_FOUND"
	ErrorCodeGeneratorError   = "GENERATOR_ERROR"
	ErrorCodeInvalidSymbol    = "INVALID_SYMBOL"
	ErrorCodeRuntimeError     = "RUNTIME_ERROR"
)

// Default Configurations

// DefaultSimulationConfig returns a basic simulation configuration
func DefaultSimulationConfig() SimulationConfig {
	return SimulationConfig{
		Symbols:              []string{"BTCUSD", "ETHUSD", "ADAUSD"},
		SimulationDuration:   1 * time.Hour,
		TickInterval:         100 * time.Millisecond,
		OrderGenerationRate:  5.0, // 5 orders per second
		PriceUpdateFrequency: 1 * time.Second,
		BaseVolatility:       0.1,
		TradingHours: TradingHours{
			OpenTime:       time.Date(0, 1, 1, 9, 0, 0, 0, time.UTC),
			CloseTime:      time.Date(0, 1, 1, 17, 0, 0, 0, time.UTC),
			TimeZone:       "UTC",
			PreMarket:      false,
			AfterHours:     false,
			WeekendTrading: true, // Crypto markets are 24/7
		},
		InitialPrices: map[string]float64{
			"BTCUSD": 50000.0,
			"ETHUSD": 3000.0,
			"ADAUSD": 1.5,
		},
		MarketCondition: MarketSteady,
		UserProfiles:    DefaultUserProfiles(),
		MarketSentiment: SentimentNeutral,
		NewsEventFrequency: 30 * time.Minute,
		EnablePatterns: true,
		PatternProbabilities: map[string]float64{
			"flash_crash": 0.01,
			"fomo_spike":  0.02,
			"whale_dump":  0.015,
		},
		ReactionDelays: ReactionDelays{
			PriceChange:     500 * time.Millisecond,
			NewsEvent:       2 * time.Second,
			LargeOrder:      200 * time.Millisecond,
			MarketCondition: 1 * time.Second,
		},
	}
}

// DefaultUserProfiles returns default user behavior profiles
func DefaultUserProfiles() []UserProfile {
	return []UserProfile{
		{
			Name:            "Conservative Retail",
			BehaviorPattern: BehaviorConservative,
			RiskTolerance:   0.3,
			OrderSizeRange: OrderSizeRange{
				Min:    10.0,
				Max:    1000.0,
				Mean:   200.0,
				StdDev: 100.0,
			},
			TradingFrequency: 0.5, // 0.5 orders per minute
			PreferredSymbols: []string{"BTCUSD", "ETHUSD"},
			ReactionTime:     5 * time.Second,
			Wealth:           10000.0,
			PopulationWeight: 0.6, // 60% of traders
		},
		{
			Name:            "Aggressive Trader",
			BehaviorPattern: BehaviorAggressive,
			RiskTolerance:   0.8,
			OrderSizeRange: OrderSizeRange{
				Min:    100.0,
				Max:    10000.0,
				Mean:   2000.0,
				StdDev: 1500.0,
			},
			TradingFrequency: 3.0, // 3 orders per minute
			PreferredSymbols: []string{"BTCUSD", "ETHUSD", "ADAUSD"},
			ReactionTime:     500 * time.Millisecond,
			Wealth:           100000.0,
			PopulationWeight: 0.25, // 25% of traders
		},
		{
			Name:            "Institutional",
			BehaviorPattern: BehaviorArbitrage,
			RiskTolerance:   0.5,
			OrderSizeRange: OrderSizeRange{
				Min:    1000.0,
				Max:    100000.0,
				Mean:   25000.0,
				StdDev: 15000.0,
			},
			TradingFrequency: 1.0, // 1 order per minute
			PreferredSymbols: []string{"BTCUSD", "ETHUSD"},
			ReactionTime:     100 * time.Millisecond,
			Wealth:           1000000.0,
			PopulationWeight: 0.15, // 15% of traders
		},
	}
}