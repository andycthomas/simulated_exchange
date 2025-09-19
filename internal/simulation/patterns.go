package simulation

import (
	"fmt"
	"math/rand"
	"time"
)

// PatternManager manages predefined simulation patterns
type PatternManager struct {
	patterns        map[string]*SimulationPattern
	eventGenerator  EventGenerator
	rng            *rand.Rand
}

// PatternEventGenerator implements EventGenerator for pattern-driven events
type PatternEventGenerator struct {
	eventHistory    []MarketEvent
	eventProbs      map[EventType]float64
	activeEvents    []MarketEvent
	rng            *rand.Rand
	config         EventGeneratorConfig
}

// EventGeneratorConfig configures event generation
type EventGeneratorConfig struct {
	BaseEventRate     float64            `json:"base_event_rate"`
	EventProbabilities map[EventType]float64 `json:"event_probabilities"`
	EventDurations    map[EventType]time.Duration `json:"event_durations"`
	MarketImpactRanges map[EventSeverity]ImpactRange `json:"market_impact_ranges"`
	RandomSeed        int64              `json:"random_seed"`
}

// ImpactRange defines the range of market impact for different severities
type ImpactRange struct {
	MinImpact float64 `json:"min_impact"`
	MaxImpact float64 `json:"max_impact"`
}

// NewPatternManager creates a new pattern manager
func NewPatternManager(eventGen EventGenerator) *PatternManager {
	pm := &PatternManager{
		patterns:       make(map[string]*SimulationPattern),
		eventGenerator: eventGen,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	pm.initializeStandardPatterns()
	return pm
}

// NewPatternEventGenerator creates a new event generator
func NewPatternEventGenerator(config EventGeneratorConfig) *PatternEventGenerator {
	if config.RandomSeed == 0 {
		config.RandomSeed = time.Now().UnixNano()
	}

	peg := &PatternEventGenerator{
		eventHistory:   make([]MarketEvent, 0),
		eventProbs:     make(map[EventType]float64),
		activeEvents:   make([]MarketEvent, 0),
		rng:           rand.New(rand.NewSource(config.RandomSeed)),
		config:        config,
	}

	peg.initializeDefaultProbabilities()
	return peg
}

// GetPattern returns a simulation pattern by name
func (pm *PatternManager) GetPattern(name string) *SimulationPattern {
	return pm.patterns[name]
}

// ListPatterns returns all available pattern names
func (pm *PatternManager) ListPatterns() []string {
	names := make([]string, 0, len(pm.patterns))
	for name := range pm.patterns {
		names = append(names, name)
	}
	return names
}

// CreateCustomPattern creates a custom simulation pattern
func (pm *PatternManager) CreateCustomPattern(name, description string, phases []PatternPhase, duration time.Duration) *SimulationPattern {
	pattern := &SimulationPattern{
		Name:        name,
		Description: description,
		Duration:    duration,
		Phases:      phases,
		Triggers:    make([]PatternTrigger, 0),
		Parameters:  make(map[string]interface{}),
	}

	pm.patterns[name] = pattern
	return pattern
}

// initializeStandardPatterns creates predefined simulation patterns
func (pm *PatternManager) initializeStandardPatterns() {
	// Flash Crash Pattern
	pm.patterns["flash_crash"] = &SimulationPattern{
		Name:        "Flash Crash",
		Description: "Sudden dramatic price decline followed by partial recovery",
		Duration:    5 * time.Minute,
		Phases: []PatternPhase{
			{
				Name:            "Pre-crash",
				Duration:        30 * time.Second,
				MarketCondition: MarketSteady,
				VolatilityLevel: 0.1,
				OrderIntensity:  1.0,
				PriceDirection:  TrendSideways,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorConservative: 0.7,
					BehaviorAggressive:   0.2,
					BehaviorArbitrage:    0.1,
				},
			},
			{
				Name:            "Crash",
				Duration:        90 * time.Second,
				MarketCondition: MarketCrash,
				VolatilityLevel: 0.8,
				OrderIntensity:  5.0,
				PriceDirection:  TrendDown,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorPanic:      0.6,
					BehaviorAggressive: 0.3,
					BehaviorArbitrage:  0.1,
				},
			},
			{
				Name:            "Recovery",
				Duration:        3 * time.Minute,
				MarketCondition: MarketRecovery,
				VolatilityLevel: 0.4,
				OrderIntensity:  2.0,
				PriceDirection:  TrendUp,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorMeanRevert: 0.4,
					BehaviorFOMO:       0.3,
					BehaviorConservative: 0.3,
				},
			},
		},
		Triggers: []PatternTrigger{
			{
				Type:      TriggerRandom,
				Condition: "probability",
				Parameters: map[string]interface{}{
					"probability": 0.001, // 0.1% chance per check
				},
				NextPhase: "Crash",
			},
		},
		Parameters: map[string]interface{}{
			"crash_magnitude": 0.15,     // 15% price drop
			"recovery_ratio":  0.6,      // Recover 60% of the drop
			"trigger_volume":  1000000,  // Large volume trigger
		},
	}

	// FOMO Spike Pattern
	pm.patterns["fomo_spike"] = &SimulationPattern{
		Name:        "FOMO Spike",
		Description: "Fear of missing out driven price spike with correction",
		Duration:    10 * time.Minute,
		Phases: []PatternPhase{
			{
				Name:            "Initial Run",
				Duration:        2 * time.Minute,
				MarketCondition: MarketBullish,
				VolatilityLevel: 0.3,
				OrderIntensity:  1.5,
				PriceDirection:  TrendUp,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorMomentum:   0.5,
					BehaviorAggressive: 0.3,
					BehaviorFOMO:       0.2,
				},
			},
			{
				Name:            "FOMO Phase",
				Duration:        3 * time.Minute,
				MarketCondition: MarketVolatile,
				VolatilityLevel: 0.6,
				OrderIntensity:  4.0,
				PriceDirection:  TrendUp,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorFOMO:       0.7,
					BehaviorAggressive: 0.2,
					BehaviorMomentum:   0.1,
				},
			},
			{
				Name:            "Correction",
				Duration:        5 * time.Minute,
				MarketCondition: MarketBearish,
				VolatilityLevel: 0.4,
				OrderIntensity:  2.0,
				PriceDirection:  TrendDown,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorMeanRevert: 0.4,
					BehaviorConservative: 0.4,
					BehaviorPanic:      0.2,
				},
			},
		},
		Parameters: map[string]interface{}{
			"spike_magnitude":     0.25,  // 25% price increase
			"correction_ratio":    0.4,   // 40% correction
			"fomo_trigger_rate":   0.1,   // 10% price increase triggers FOMO
		},
	}

	// Whale Dump Pattern
	pm.patterns["whale_dump"] = &SimulationPattern{
		Name:        "Whale Dump",
		Description: "Large holder selling pressure causing sustained decline",
		Duration:    15 * time.Minute,
		Phases: []PatternPhase{
			{
				Name:            "Accumulation Signs",
				Duration:        2 * time.Minute,
				MarketCondition: MarketSteady,
				VolatilityLevel: 0.15,
				OrderIntensity:  1.2,
				PriceDirection:  TrendSideways,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorConservative: 0.6,
					BehaviorAggressive:   0.3,
					BehaviorArbitrage:    0.1,
				},
			},
			{
				Name:            "Initial Dump",
				Duration:        3 * time.Minute,
				MarketCondition: MarketBearish,
				VolatilityLevel: 0.5,
				OrderIntensity:  3.0,
				PriceDirection:  TrendDown,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorPanic:      0.4,
					BehaviorMeanRevert: 0.3,
					BehaviorConservative: 0.3,
				},
			},
			{
				Name:            "Sustained Selling",
				Duration:        10 * time.Minute,
				MarketCondition: MarketBearish,
				VolatilityLevel: 0.3,
				OrderIntensity:  1.8,
				PriceDirection:  TrendDown,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorMeanRevert: 0.5,
					BehaviorConservative: 0.3,
					BehaviorPanic:      0.2,
				},
			},
		},
		Parameters: map[string]interface{}{
			"dump_magnitude":    0.20,    // 20% price decline
			"whale_order_size":  100000,  // Large order sizes
			"selling_duration":  600,     // 10 minutes of selling
		},
	}

	// Morning Pump Pattern
	pm.patterns["morning_pump"] = &SimulationPattern{
		Name:        "Morning Pump",
		Description: "Market opening pump followed by profit taking",
		Duration:    30 * time.Minute,
		Phases: []PatternPhase{
			{
				Name:            "Pre-market",
				Duration:        5 * time.Minute,
				MarketCondition: MarketSteady,
				VolatilityLevel: 0.08,
				OrderIntensity:  0.5,
				PriceDirection:  TrendSideways,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorConservative: 0.8,
					BehaviorAggressive:   0.2,
				},
			},
			{
				Name:            "Opening Pump",
				Duration:        10 * time.Minute,
				MarketCondition: MarketBullish,
				VolatilityLevel: 0.4,
				OrderIntensity:  2.5,
				PriceDirection:  TrendUp,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorMomentum:   0.4,
					BehaviorFOMO:       0.3,
					BehaviorAggressive: 0.3,
				},
			},
			{
				Name:            "Profit Taking",
				Duration:        15 * time.Minute,
				MarketCondition: MarketSideways,
				VolatilityLevel: 0.2,
				OrderIntensity:  1.2,
				PriceDirection:  TrendSideways,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorConservative: 0.5,
					BehaviorMeanRevert: 0.3,
					BehaviorArbitrage:  0.2,
				},
			},
		},
		Parameters: map[string]interface{}{
			"pump_magnitude":   0.08,   // 8% price increase
			"volume_spike":     3.0,    // 3x normal volume
			"opening_time":     "09:00",
		},
	}

	// Consolidation Pattern
	pm.patterns["consolidation"] = &SimulationPattern{
		Name:        "Consolidation",
		Description: "Sideways price action with decreasing volatility",
		Duration:    60 * time.Minute,
		Phases: []PatternPhase{
			{
				Name:            "Range Formation",
				Duration:        20 * time.Minute,
				MarketCondition: MarketSideways,
				VolatilityLevel: 0.2,
				OrderIntensity:  1.0,
				PriceDirection:  TrendSideways,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorConservative: 0.5,
					BehaviorMeanRevert: 0.3,
					BehaviorArbitrage:  0.2,
				},
			},
			{
				Name:            "Tight Range",
				Duration:        30 * time.Minute,
				MarketCondition: MarketSteady,
				VolatilityLevel: 0.1,
				OrderIntensity:  0.8,
				PriceDirection:  TrendSideways,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorConservative: 0.6,
					BehaviorArbitrage:  0.3,
					BehaviorMeanRevert: 0.1,
				},
			},
			{
				Name:            "Breakout Setup",
				Duration:        10 * time.Minute,
				MarketCondition: MarketVolatile,
				VolatilityLevel: 0.3,
				OrderIntensity:  1.5,
				PriceDirection:  TrendSideways,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorAggressive: 0.4,
					BehaviorMomentum:   0.3,
					BehaviorConservative: 0.3,
				},
			},
		},
		Parameters: map[string]interface{}{
			"range_percentage": 0.05,   // 5% trading range
			"breakout_trigger": 0.03,   // 3% move triggers breakout
			"volume_decline":   0.7,    // Volume declines to 70%
		},
	}

	// News Spike Pattern
	pm.patterns["news_spike"] = &SimulationPattern{
		Name:        "News Spike",
		Description: "Sharp price movement due to news event",
		Duration:    8 * time.Minute,
		Phases: []PatternPhase{
			{
				Name:            "Pre-news",
				Duration:        1 * time.Minute,
				MarketCondition: MarketSteady,
				VolatilityLevel: 0.1,
				OrderIntensity:  1.0,
				PriceDirection:  TrendSideways,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorConservative: 0.7,
					BehaviorAggressive:   0.3,
				},
			},
			{
				Name:            "News Reaction",
				Duration:        2 * time.Minute,
				MarketCondition: MarketVolatile,
				VolatilityLevel: 0.9,
				OrderIntensity:  6.0,
				PriceDirection:  TrendUp, // Positive news
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorFOMO:       0.5,
					BehaviorAggressive: 0.3,
					BehaviorMomentum:   0.2,
				},
			},
			{
				Name:            "Digestion",
				Duration:        5 * time.Minute,
				MarketCondition: MarketSideways,
				VolatilityLevel: 0.3,
				OrderIntensity:  1.5,
				PriceDirection:  TrendSideways,
				UserBehavior: map[UserBehaviorPattern]float64{
					BehaviorConservative: 0.4,
					BehaviorMeanRevert: 0.3,
					BehaviorArbitrage:  0.3,
				},
			},
		},
		Parameters: map[string]interface{}{
			"news_impact":      0.12,   // 12% price impact
			"reaction_speed":   5.0,    // 5 seconds reaction time
			"volume_multiplier": 5.0,   // 5x volume spike
		},
	}
}

// Event Generator Implementation

// GenerateMarketEvent creates random market events
func (peg *PatternEventGenerator) GenerateMarketEvent() MarketEvent {
	eventType := peg.selectEventType()
	severity := peg.selectEventSeverity()

	event := MarketEvent{
		ID:              fmt.Sprintf("event_%d_%s", time.Now().Unix(), eventType),
		Type:            eventType,
		Severity:        severity,
		AffectedSymbols: peg.selectAffectedSymbols(),
		PriceImpact:     peg.calculatePriceImpact(severity),
		Duration:        peg.getEventDuration(eventType),
		Description:     peg.generateEventDescription(eventType, severity),
		StartTime:       time.Now(),
		IsActive:        true,
	}

	return event
}

// InjectEvent manually introduces a specific event
func (peg *PatternEventGenerator) InjectEvent(event MarketEvent) error {
	event.StartTime = time.Now()
	event.IsActive = true

	peg.activeEvents = append(peg.activeEvents, event)
	peg.eventHistory = append(peg.eventHistory, event)

	// Schedule event deactivation
	go func() {
		time.Sleep(event.Duration)
		peg.deactivateEvent(event.ID)
	}()

	return nil
}

// GetActiveEvents returns currently active market events
func (peg *PatternEventGenerator) GetActiveEvents() []MarketEvent {
	// Clean up expired events
	now := time.Now()
	activeEvents := make([]MarketEvent, 0)

	for _, event := range peg.activeEvents {
		if now.Sub(event.StartTime) < event.Duration {
			activeEvents = append(activeEvents, event)
		}
	}

	peg.activeEvents = activeEvents
	return activeEvents
}

// SetEventProbability configures event generation probabilities
func (peg *PatternEventGenerator) SetEventProbability(eventType EventType, probability float64) {
	peg.eventProbs[eventType] = probability
}

// Private helper methods for event generation

func (peg *PatternEventGenerator) initializeDefaultProbabilities() {
	peg.eventProbs = map[EventType]float64{
		EventEarnings:     0.05,  // 5% chance
		EventNews:         0.15,  // 15% chance
		EventRegulatory:   0.02,  // 2% chance
		EventEconomic:     0.08,  // 8% chance
		EventGeopolitical: 0.03,  // 3% chance
		EventTechnical:    0.01,  // 1% chance
		EventCorporate:    0.04,  // 4% chance
	}
}

func (peg *PatternEventGenerator) selectEventType() EventType {
	totalProb := 0.0
	for _, prob := range peg.eventProbs {
		totalProb += prob
	}

	r := peg.rng.Float64() * totalProb
	cumProb := 0.0

	for eventType, prob := range peg.eventProbs {
		cumProb += prob
		if r <= cumProb {
			return eventType
		}
	}

	return EventNews // Default fallback
}

func (peg *PatternEventGenerator) selectEventSeverity() EventSeverity {
	r := peg.rng.Float64()

	switch {
	case r < 0.5:
		return SeverityLow
	case r < 0.8:
		return SeverityMedium
	case r < 0.95:
		return SeverityHigh
	default:
		return SeverityCrisis
	}
}

func (peg *PatternEventGenerator) selectAffectedSymbols() []string {
	symbols := []string{"BTCUSD", "ETHUSD", "ADAUSD"}

	// Randomly select 1-3 symbols
	numSymbols := peg.rng.Intn(3) + 1
	selected := make([]string, 0, numSymbols)

	for i := 0; i < numSymbols && i < len(symbols); i++ {
		idx := peg.rng.Intn(len(symbols))
		selected = append(selected, symbols[idx])

		// Remove selected symbol to avoid duplicates
		symbols = append(symbols[:idx], symbols[idx+1:]...)
	}

	return selected
}

func (peg *PatternEventGenerator) calculatePriceImpact(severity EventSeverity) float64 {
	var min, max float64

	switch severity {
	case SeverityLow:
		min, max = 0.5, 2.0    // 0.5% to 2%
	case SeverityMedium:
		min, max = 2.0, 5.0    // 2% to 5%
	case SeverityHigh:
		min, max = 5.0, 12.0   // 5% to 12%
	case SeverityCrisis:
		min, max = 12.0, 25.0  // 12% to 25%
	default:
		min, max = 1.0, 3.0
	}

	return min + peg.rng.Float64()*(max-min)
}

func (peg *PatternEventGenerator) getEventDuration(eventType EventType) time.Duration {
	baseDuration := map[EventType]time.Duration{
		EventEarnings:     30 * time.Minute,
		EventNews:         15 * time.Minute,
		EventRegulatory:   2 * time.Hour,
		EventEconomic:     45 * time.Minute,
		EventGeopolitical: 4 * time.Hour,
		EventTechnical:    10 * time.Minute,
		EventCorporate:    1 * time.Hour,
	}

	base := baseDuration[eventType]
	if base == 0 {
		base = 20 * time.Minute
	}

	// Add randomness: 50% to 150% of base duration
	multiplier := 0.5 + peg.rng.Float64()
	return time.Duration(float64(base) * multiplier)
}

func (peg *PatternEventGenerator) generateEventDescription(eventType EventType, severity EventSeverity) string {
	templates := map[EventType][]string{
		EventEarnings: {
			"Quarterly earnings report released",
			"Company beats earnings expectations",
			"Disappointing earnings results",
		},
		EventNews: {
			"Breaking news affects market sentiment",
			"Major announcement drives trading activity",
			"Market reacts to latest developments",
		},
		EventRegulatory: {
			"New regulations announced",
			"Regulatory clarity provided",
			"Government policy change",
		},
		EventEconomic: {
			"Economic indicators released",
			"Central bank announcement",
			"Inflation data published",
		},
		EventGeopolitical: {
			"Geopolitical tensions rise",
			"International trade agreement",
			"Political uncertainty affects markets",
		},
		EventTechnical: {
			"Technical analysis breakout",
			"Key support level tested",
			"Resistance level breakthrough",
		},
		EventCorporate: {
			"Corporate partnership announced",
			"Merger and acquisition activity",
			"Product launch announcement",
		},
	}

	eventTemplates := templates[eventType]
	if len(eventTemplates) == 0 {
		return "Market event occurred"
	}

	template := eventTemplates[peg.rng.Intn(len(eventTemplates))]

	// Add severity qualifier
	severityPrefix := map[EventSeverity]string{
		SeverityLow:    "Minor: ",
		SeverityMedium: "Moderate: ",
		SeverityHigh:   "Major: ",
		SeverityCrisis: "Critical: ",
	}

	return severityPrefix[severity] + template
}

func (peg *PatternEventGenerator) deactivateEvent(eventID string) {
	for i, event := range peg.activeEvents {
		if event.ID == eventID {
			peg.activeEvents[i].IsActive = false
			// Remove from active events
			peg.activeEvents = append(peg.activeEvents[:i], peg.activeEvents[i+1:]...)
			break
		}
	}
}

// Pattern Execution Helpers

// ExecutePattern applies a simulation pattern to the market
func (pm *PatternManager) ExecutePattern(patternName string, priceGen PriceGenerator, orderGen OrderGenerator) error {
	pattern := pm.patterns[patternName]
	if pattern == nil {
		return fmt.Errorf("pattern '%s' not found", patternName)
	}

	// Execute each phase of the pattern
	for _, phase := range pattern.Phases {
		if err := pm.executePhase(phase, priceGen, orderGen); err != nil {
			return fmt.Errorf("failed to execute phase '%s': %w", phase.Name, err)
		}
	}

	return nil
}

func (pm *PatternManager) executePhase(phase PatternPhase, priceGen PriceGenerator, orderGen OrderGenerator) error {
	// Apply volatility changes
	volatilityPattern := pm.mapMarketConditionToVolatility(phase.MarketCondition)
	priceGen.SimulateVolatility(volatilityPattern, phase.VolatilityLevel)

	// Apply user behavior changes
	for behavior, weight := range phase.UserBehavior {
		orderGen.SimulateUserBehavior(behavior, weight*phase.OrderIntensity)
	}

	// Wait for phase duration
	time.Sleep(phase.Duration)

	return nil
}

func (pm *PatternManager) mapMarketConditionToVolatility(condition MarketCondition) VolatilityPattern {
	switch condition {
	case MarketVolatile:
		return VolatilitySpike
	case MarketCrash:
		return VolatilitySpike
	case MarketSteady:
		return VolatilityDecay
	case MarketSideways:
		return VolatilityDecay
	default:
		return VolatilityOscillate
	}
}

// DefaultEventGeneratorConfig returns default event generator configuration
func DefaultEventGeneratorConfig() EventGeneratorConfig {
	return EventGeneratorConfig{
		BaseEventRate: 0.1, // 10% chance per check interval
		EventProbabilities: map[EventType]float64{
			EventEarnings:     0.05,
			EventNews:         0.15,
			EventRegulatory:   0.02,
			EventEconomic:     0.08,
			EventGeopolitical: 0.03,
			EventTechnical:    0.01,
			EventCorporate:    0.04,
		},
		EventDurations: map[EventType]time.Duration{
			EventEarnings:     30 * time.Minute,
			EventNews:         15 * time.Minute,
			EventRegulatory:   2 * time.Hour,
			EventEconomic:     45 * time.Minute,
			EventGeopolitical: 4 * time.Hour,
			EventTechnical:    10 * time.Minute,
			EventCorporate:    1 * time.Hour,
		},
		MarketImpactRanges: map[EventSeverity]ImpactRange{
			SeverityLow:    {MinImpact: 0.5, MaxImpact: 2.0},
			SeverityMedium: {MinImpact: 2.0, MaxImpact: 5.0},
			SeverityHigh:   {MinImpact: 5.0, MaxImpact: 12.0},
			SeverityCrisis: {MinImpact: 12.0, MaxImpact: 25.0},
		},
		RandomSeed: 0,
	}
}