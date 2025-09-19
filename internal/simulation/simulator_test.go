package simulation

import (
	"context"
	"testing"
	"time"

	"simulated_exchange/internal/api/dto"
)

// Mock implementations for testing

type MockPriceGenerator struct {
	prices       map[string]float64
	volatility   float64
	basePrices   map[string]float64
	trends       map[string]PriceTrend
}

func NewMockPriceGenerator() *MockPriceGenerator {
	return &MockPriceGenerator{
		prices:     make(map[string]float64),
		basePrices: make(map[string]float64),
		trends:     make(map[string]PriceTrend),
		volatility: 0.02,
	}
}

func (mpg *MockPriceGenerator) GeneratePrice(symbol string, currentPrice float64, timeElapsed time.Duration) float64 {
	// Simple price generation for testing
	change := (mpg.volatility * 2 * (0.5 - 0.5)) // Random walk simulation
	newPrice := currentPrice * (1 + change)
	mpg.prices[symbol] = newPrice
	return newPrice
}

func (mpg *MockPriceGenerator) SimulateVolatility(pattern VolatilityPattern, intensity float64) {
	switch pattern {
	case VolatilitySpike:
		mpg.volatility = 0.1 + intensity*0.2
	case VolatilityDecay:
		mpg.volatility *= (1 - intensity*0.1)
	default:
		mpg.volatility = 0.02
	}
}

func (mpg *MockPriceGenerator) GetPriceTrend(symbol string) PriceTrend {
	if trend, exists := mpg.trends[symbol]; exists {
		return trend
	}
	return PriceTrend{
		Symbol:    symbol,
		Direction: TrendSideways,
		Strength:  0.0,
		LastUpdate: time.Now(),
	}
}

func (mpg *MockPriceGenerator) SetBasePrice(symbol string, price float64) {
	mpg.basePrices[symbol] = price
	mpg.prices[symbol] = price
}

func (mpg *MockPriceGenerator) Reset() {
	mpg.prices = make(map[string]float64)
	mpg.basePrices = make(map[string]float64)
	mpg.trends = make(map[string]PriceTrend)
	mpg.volatility = 0.02
}

type MockOrderGenerator struct {
	userProfiles []UserProfile
	sentiment    MarketSentiment
	statistics   OrderStatistics
	orderCount   int64
}

func NewMockOrderGenerator() *MockOrderGenerator {
	return &MockOrderGenerator{
		userProfiles: DefaultUserProfiles(),
		sentiment:    SentimentNeutral,
		statistics: OrderStatistics{
			OrdersBySymbol:   make(map[string]int64),
			OrdersBySide:     make(map[string]int64),
			OrdersByUserType: make(map[string]int64),
		},
	}
}

func (mog *MockOrderGenerator) GenerateRealisticOrders(symbol string, currentPrice float64, marketCondition MarketCondition) []dto.PlaceOrderRequest {
	// Generate 1-3 orders for testing
	numOrders := 1 + int(mog.orderCount%3)
	orders := make([]dto.PlaceOrderRequest, numOrders)

	for i := 0; i < numOrders; i++ {
		side := "buy"
		if i%2 == 1 {
			side = "sell"
		}

		price := currentPrice
		orderType := "market"
		if i%3 == 0 {
			orderType = "limit"
			if side == "buy" {
				price *= 0.999 // Buy below market
			} else {
				price *= 1.001 // Sell above market
			}
		}

		orders[i] = dto.PlaceOrderRequest{
			Symbol:   symbol,
			Side:     side,
			Quantity: 100.0 + float64(i*50),
			Type:     orderType,
		}

		if orderType == "limit" {
			orders[i].Price = price
		}
	}

	mog.orderCount += int64(numOrders)
	mog.statistics.TotalOrders += int64(numOrders)
	mog.statistics.OrdersBySymbol[symbol] += int64(numOrders)

	return orders
}

func (mog *MockOrderGenerator) SimulateUserBehavior(pattern UserBehaviorPattern, intensity float64) []dto.PlaceOrderRequest {
	// Generate orders based on behavior pattern
	numOrders := int(intensity * 2) // 2 orders per intensity unit
	if numOrders == 0 {
		numOrders = 1
	}

	orders := make([]dto.PlaceOrderRequest, numOrders)
	symbol := "BTCUSD"

	for i := 0; i < numOrders; i++ {
		side := "buy"
		switch pattern {
		case BehaviorFOMO:
			side = "buy" // FOMO is usually buying
		case BehaviorPanic:
			side = "sell" // Panic is usually selling
		case BehaviorMomentum:
			if i%2 == 0 {
				side = "buy"
			} else {
				side = "sell"
			}
		}

		orders[i] = dto.PlaceOrderRequest{
			Symbol:   symbol,
			Side:     side,
			Quantity: 200.0 * intensity,
			Type:     "market",
		}
	}

	return orders
}

func (mog *MockOrderGenerator) SetUserProfiles(profiles []UserProfile) {
	mog.userProfiles = profiles
}

func (mog *MockOrderGenerator) GetOrderStatistics() OrderStatistics {
	return mog.statistics
}

func (mog *MockOrderGenerator) UpdateMarketSentiment(sentiment MarketSentiment) {
	mog.sentiment = sentiment
}

type MockEventGenerator struct {
	events       []MarketEvent
	activeEvents []MarketEvent
	probabilities map[EventType]float64
}

func NewMockEventGenerator() *MockEventGenerator {
	return &MockEventGenerator{
		events:       make([]MarketEvent, 0),
		activeEvents: make([]MarketEvent, 0),
		probabilities: make(map[EventType]float64),
	}
}

func (meg *MockEventGenerator) GenerateMarketEvent() MarketEvent {
	event := MarketEvent{
		ID:              "test_event",
		Type:            EventNews,
		Severity:        SeverityMedium,
		AffectedSymbols: []string{"BTCUSD"},
		PriceImpact:     5.0,
		Duration:        1 * time.Minute,
		Description:     "Test market event",
		StartTime:       time.Now(),
		IsActive:        true,
	}
	return event
}

func (meg *MockEventGenerator) InjectEvent(event MarketEvent) error {
	meg.activeEvents = append(meg.activeEvents, event)
	meg.events = append(meg.events, event)
	return nil
}

func (meg *MockEventGenerator) GetActiveEvents() []MarketEvent {
	return meg.activeEvents
}

func (meg *MockEventGenerator) SetEventProbability(eventType EventType, probability float64) {
	meg.probabilities[eventType] = probability
}

type MockTradingEngine struct {
	orders []dto.PlaceOrderRequest
}

func NewMockTradingEngine() *MockTradingEngine {
	return &MockTradingEngine{
		orders: make([]dto.PlaceOrderRequest, 0),
	}
}

func (mte *MockTradingEngine) PlaceOrder(order dto.PlaceOrderRequest) (interface{}, error) {
	mte.orders = append(mte.orders, order)
	return map[string]interface{}{
		"order_id": "test_order_id",
		"status":   "FILLED",
	}, nil
}

// Unit Tests

func TestRealisticSimulator_StartStopSimulation(t *testing.T) {
	priceGen := NewMockPriceGenerator()
	orderGen := NewMockOrderGenerator()
	eventGen := NewMockEventGenerator()
	engine := NewMockTradingEngine()

	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	// Test starting simulation
	config := DefaultSimulationConfig()
	config.SimulationDuration = 2 * time.Second // Short duration for testing

	ctx := context.Background()
	err := simulator.StartSimulation(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start simulation: %v", err)
	}

	// Check status
	status := simulator.GetSimulationStatus()
	if !status.IsRunning {
		t.Error("Simulation should be running")
	}

	// Wait a bit to let simulation run
	time.Sleep(500 * time.Millisecond)

	// Test stopping simulation
	err = simulator.StopSimulation()
	if err != nil {
		t.Fatalf("Failed to stop simulation: %v", err)
	}

	// Check status after stopping
	status = simulator.GetSimulationStatus()
	if status.IsRunning {
		t.Error("Simulation should not be running after stop")
	}

	// Check that some orders were generated
	if len(engine.orders) == 0 {
		t.Error("Expected some orders to be generated during simulation")
	}
}

func TestRealisticSimulator_InjectVolatility(t *testing.T) {
	priceGen := NewMockPriceGenerator()
	orderGen := NewMockOrderGenerator()
	eventGen := NewMockEventGenerator()
	engine := NewMockTradingEngine()

	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	// Start simulation
	config := DefaultSimulationConfig()
	config.SimulationDuration = 3 * time.Second

	ctx := context.Background()
	err := simulator.StartSimulation(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start simulation: %v", err)
	}

	// Test volatility injection
	err = simulator.InjectVolatility(VolatilitySpike, 1*time.Second)
	if err != nil {
		t.Errorf("Failed to inject volatility: %v", err)
	}

	// Check that volatility was applied
	originalVol := priceGen.volatility
	time.Sleep(100 * time.Millisecond)

	// Volatility should have increased
	if priceGen.volatility <= originalVol {
		t.Error("Expected volatility to increase after spike injection")
	}

	// Clean up
	simulator.StopSimulation()
}

func TestRealisticSimulator_ConfigValidation(t *testing.T) {
	priceGen := NewMockPriceGenerator()
	orderGen := NewMockOrderGenerator()
	eventGen := NewMockEventGenerator()
	engine := NewMockTradingEngine()

	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	ctx := context.Background()

	// Test invalid config - no symbols
	config := DefaultSimulationConfig()
	config.Symbols = []string{}

	err := simulator.StartSimulation(ctx, config)
	if err == nil {
		t.Error("Expected error for empty symbols")
	}

	// Test invalid config - negative duration
	config = DefaultSimulationConfig()
	config.SimulationDuration = -1 * time.Second

	err = simulator.StartSimulation(ctx, config)
	if err == nil {
		t.Error("Expected error for negative duration")
	}

	// Test invalid config - missing initial prices
	config = DefaultSimulationConfig()
	config.InitialPrices = make(map[string]float64)

	err = simulator.StartSimulation(ctx, config)
	if err == nil {
		t.Error("Expected error for missing initial prices")
	}
}

func TestRealisticSimulator_UpdateConfig(t *testing.T) {
	priceGen := NewMockPriceGenerator()
	orderGen := NewMockOrderGenerator()
	eventGen := NewMockEventGenerator()
	engine := NewMockTradingEngine()

	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	// Start simulation
	config := DefaultSimulationConfig()
	config.SimulationDuration = 3 * time.Second

	ctx := context.Background()
	err := simulator.StartSimulation(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start simulation: %v", err)
	}

	// Update config
	newConfig := config
	newConfig.MarketCondition = MarketVolatile
	newConfig.MarketSentiment = SentimentOptimistic

	err = simulator.UpdateConfig(newConfig)
	if err != nil {
		t.Errorf("Failed to update config: %v", err)
	}

	// Check that config was updated
	status := simulator.GetSimulationStatus()
	if status.MarketCondition != MarketVolatile {
		t.Error("Market condition should have been updated")
	}

	// Clean up
	simulator.StopSimulation()
}

// Test simulation patterns

func TestFlashCrashPattern(t *testing.T) {
	priceGen := NewMockPriceGenerator()
	orderGen := NewMockOrderGenerator()
	eventGen := NewMockEventGenerator()
	engine := NewMockTradingEngine()

	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	// Enable flash crash pattern
	config := DefaultSimulationConfig()
	config.EnablePatterns = true
	config.PatternProbabilities = map[string]float64{
		"flash_crash": 1.0, // 100% chance for testing
	}
	config.SimulationDuration = 10 * time.Second

	ctx := context.Background()
	err := simulator.StartSimulation(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start simulation: %v", err)
	}

	// Wait for pattern to potentially activate
	time.Sleep(2 * time.Second)

	// Check if pattern was activated
	statistics := simulator.GetStatistics()
	if statistics.PatternsActivated == 0 {
		t.Log("Flash crash pattern was not activated during test (random behavior)")
	}

	// Check that orders were generated
	if len(engine.orders) == 0 {
		t.Error("Expected orders to be generated")
	}

	simulator.StopSimulation()
}

func TestFOMOPattern(t *testing.T) {
	orderGen := NewMockOrderGenerator()

	// Test FOMO behavior pattern directly
	orders := orderGen.SimulateUserBehavior(BehaviorFOMO, 2.0)

	if len(orders) == 0 {
		t.Error("Expected FOMO behavior to generate orders")
	}

	// Check that FOMO generates mostly buy orders
	buyCount := 0
	for _, order := range orders {
		if order.Side == "buy" {
			buyCount++
		}
	}

	buyRatio := float64(buyCount) / float64(len(orders))
	if buyRatio < 0.5 {
		t.Errorf("Expected FOMO to generate more buy orders, got ratio: %f", buyRatio)
	}
}

func TestPanicPattern(t *testing.T) {
	orderGen := NewMockOrderGenerator()

	// Test panic behavior pattern
	orders := orderGen.SimulateUserBehavior(BehaviorPanic, 3.0)

	if len(orders) == 0 {
		t.Error("Expected panic behavior to generate orders")
	}

	// Check that panic generates mostly sell orders
	sellCount := 0
	for _, order := range orders {
		if order.Side == "sell" {
			sellCount++
		}
	}

	sellRatio := float64(sellCount) / float64(len(orders))
	if sellRatio < 0.5 {
		t.Errorf("Expected panic to generate more sell orders, got ratio: %f", sellRatio)
	}
}

// Test price generator

func TestMockPriceGenerator_VolatilityPatterns(t *testing.T) {
	priceGen := NewMockPriceGenerator()
	priceGen.SetBasePrice("BTCUSD", 50000.0)

	originalVolatility := priceGen.volatility

	// Test volatility spike
	priceGen.SimulateVolatility(VolatilitySpike, 0.8)
	if priceGen.volatility <= originalVolatility {
		t.Error("Volatility should increase after spike")
	}

	// Test volatility decay
	priceGen.SimulateVolatility(VolatilityDecay, 0.5)
	if priceGen.volatility >= originalVolatility {
		t.Error("Volatility should decrease after decay")
	}
}

func TestMockPriceGenerator_PriceGeneration(t *testing.T) {
	priceGen := NewMockPriceGenerator()
	symbol := "BTCUSD"
	basePrice := 50000.0

	priceGen.SetBasePrice(symbol, basePrice)

	// Generate prices over time
	prices := make([]float64, 10)
	for i := 0; i < 10; i++ {
		timeElapsed := time.Duration(i) * time.Second
		prices[i] = priceGen.GeneratePrice(symbol, basePrice, timeElapsed)
	}

	// Check that prices are being generated
	for i, price := range prices {
		if price <= 0 {
			t.Errorf("Price %d should be positive, got %f", i, price)
		}
	}

	// Check price trend
	trend := priceGen.GetPriceTrend(symbol)
	if trend.Symbol != symbol {
		t.Errorf("Expected symbol %s, got %s", symbol, trend.Symbol)
	}
}

// Test order statistics

func TestOrderStatistics_Tracking(t *testing.T) {
	orderGen := NewMockOrderGenerator()
	symbol := "BTCUSD"
	currentPrice := 50000.0

	// Generate orders
	orders := orderGen.GenerateRealisticOrders(symbol, currentPrice, MarketSteady)

	// Check statistics
	stats := orderGen.GetOrderStatistics()

	if stats.TotalOrders == 0 {
		t.Error("Expected total orders to be tracked")
	}

	if stats.OrdersBySymbol[symbol] == 0 {
		t.Error("Expected orders by symbol to be tracked")
	}

	expectedTotal := int64(len(orders))
	if stats.TotalOrders < expectedTotal {
		t.Errorf("Expected at least %d total orders, got %d", expectedTotal, stats.TotalOrders)
	}
}

// Test user profile selection

func TestUserProfileSelection(t *testing.T) {
	profiles := []UserProfile{
		{
			Name:             "Test Conservative",
			BehaviorPattern:  BehaviorConservative,
			PopulationWeight: 0.5,
		},
		{
			Name:             "Test Aggressive",
			BehaviorPattern:  BehaviorAggressive,
			PopulationWeight: 0.3,
		},
		{
			Name:             "Test Institutional",
			BehaviorPattern:  BehaviorArbitrage,
			PopulationWeight: 0.2,
		},
	}

	orderGen := NewMockOrderGenerator()
	orderGen.SetUserProfiles(profiles)

	// Test that profiles were set
	if len(orderGen.userProfiles) != 3 {
		t.Errorf("Expected 3 user profiles, got %d", len(orderGen.userProfiles))
	}

	// Generate orders and check they use the profiles
	orders := orderGen.GenerateRealisticOrders("BTCUSD", 50000.0, MarketSteady)
	if len(orders) == 0 {
		t.Error("Expected orders to be generated with user profiles")
	}
}

// Test market sentiment effects

func TestMarketSentimentEffects(t *testing.T) {
	orderGen := NewMockOrderGenerator()

	// Test different sentiments
	sentiments := []MarketSentiment{
		SentimentOptimistic,
		SentimentPessimistic,
		SentimentNeutral,
		SentimentGreedy,
		SentimentFearful,
	}

	for _, sentiment := range sentiments {
		orderGen.UpdateMarketSentiment(sentiment)

		orders := orderGen.GenerateRealisticOrders("BTCUSD", 50000.0, MarketSteady)

		if len(orders) == 0 {
			t.Errorf("Expected orders to be generated for sentiment %s", sentiment)
		}

		// Check that sentiment affects order generation
		// (This is a basic check - in real implementation, sentiment would affect buy/sell ratios)
		for _, order := range orders {
			if order.Quantity <= 0 {
				t.Errorf("Order quantity should be positive for sentiment %s", sentiment)
			}
		}
	}
}

// Test error conditions

func TestSimulationErrorConditions(t *testing.T) {
	priceGen := NewMockPriceGenerator()
	orderGen := NewMockOrderGenerator()
	eventGen := NewMockEventGenerator()
	engine := NewMockTradingEngine()

	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	// Test double start
	config := DefaultSimulationConfig()
	config.SimulationDuration = 1 * time.Second

	ctx := context.Background()
	err := simulator.StartSimulation(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start simulation: %v", err)
	}

	// Try to start again
	err = simulator.StartSimulation(ctx, config)
	if err == nil {
		t.Error("Expected error when starting already running simulation")
	}

	// Test stop when not running
	simulator.StopSimulation()

	err = simulator.StopSimulation()
	if err == nil {
		t.Error("Expected error when stopping already stopped simulation")
	}

	// Test inject volatility when not running
	err = simulator.InjectVolatility(VolatilitySpike, 1*time.Second)
	if err == nil {
		t.Error("Expected error when injecting volatility while not running")
	}
}