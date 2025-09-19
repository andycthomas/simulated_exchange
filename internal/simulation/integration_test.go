package simulation

import (
	"context"
	"sync"
	"testing"
	"time"

	"simulated_exchange/internal/api/dto"
)

// IntegrationTradingEngine implements a simplified trading engine for integration testing
type IntegrationTradingEngine struct {
	orderBook    map[string]*OrderBook
	orderHistory []dto.PlaceOrderRequest
	tradeHistory []Trade
	mu           sync.RWMutex
}

// OrderBook represents a simple order book for testing
type OrderBook struct {
	Symbol      string
	BuyOrders   []Order
	SellOrders  []Order
	LastPrice   float64
	TotalVolume float64
}

// Order represents an order in the book
type Order struct {
	ID       string
	Side     string
	Quantity float64
	Price    float64
	UserID   string
	Time     time.Time
}

// Trade represents an executed trade
type Trade struct {
	ID          string
	Symbol      string
	BuyOrderID  string
	SellOrderID string
	Quantity    float64
	Price       float64
	Time        time.Time
}

// NewIntegrationTradingEngine creates a new integration trading engine
func NewIntegrationTradingEngine() *IntegrationTradingEngine {
	return &IntegrationTradingEngine{
		orderBook:    make(map[string]*OrderBook),
		orderHistory: make([]dto.PlaceOrderRequest, 0),
		tradeHistory: make([]Trade, 0),
	}
}

// PlaceOrder places an order in the integration engine
func (ite *IntegrationTradingEngine) PlaceOrder(order dto.PlaceOrderRequest) (interface{}, error) {
	ite.mu.Lock()
	defer ite.mu.Unlock()

	// Add to order history
	ite.orderHistory = append(ite.orderHistory, order)

	// Initialize order book if needed
	if _, exists := ite.orderBook[order.Symbol]; !exists {
		ite.orderBook[order.Symbol] = &OrderBook{
			Symbol:      order.Symbol,
			BuyOrders:   make([]Order, 0),
			SellOrders:  make([]Order, 0),
			LastPrice:   50000.0, // Default price
			TotalVolume: 0,
		}
	}

	book := ite.orderBook[order.Symbol]

	// Create order
	newOrder := Order{
		ID:       generateOrderID(),
		Side:     order.Side,
		Quantity: order.Quantity,
		UserID:   "simulated_user", // Default user ID for simulation
		Time:     time.Now(),
	}

	// Set price based on order type
	if order.Type == "market" {
		newOrder.Price = book.LastPrice
	} else {
		newOrder.Price = order.Price
	}

	// Attempt to match order
	executed := ite.matchOrder(book, newOrder)

	// Add remaining quantity to book if not fully executed
	if newOrder.Quantity > 0 && order.Type == "limit" {
		if newOrder.Side == "buy" {
			book.BuyOrders = append(book.BuyOrders, newOrder)
		} else {
			book.SellOrders = append(book.SellOrders, newOrder)
		}
	}

	response := map[string]interface{}{
		"order_id":        newOrder.ID,
		"status":          "FILLED",
		"executed_quantity": order.Quantity - newOrder.Quantity,
		"remaining_quantity": newOrder.Quantity,
	}

	if !executed && order.Type == "limit" {
		response["status"] = "OPEN"
	}

	return response, nil
}

// GetOrderBook returns the current order book for a symbol
func (ite *IntegrationTradingEngine) GetOrderBook(symbol string) *OrderBook {
	ite.mu.RLock()
	defer ite.mu.RUnlock()

	if book, exists := ite.orderBook[symbol]; exists {
		return book
	}
	return nil
}

// GetOrderHistory returns all placed orders
func (ite *IntegrationTradingEngine) GetOrderHistory() []dto.PlaceOrderRequest {
	ite.mu.RLock()
	defer ite.mu.RUnlock()

	history := make([]dto.PlaceOrderRequest, len(ite.orderHistory))
	copy(history, ite.orderHistory)
	return history
}

// GetTradeHistory returns all executed trades
func (ite *IntegrationTradingEngine) GetTradeHistory() []Trade {
	ite.mu.RLock()
	defer ite.mu.RUnlock()

	history := make([]Trade, len(ite.tradeHistory))
	copy(history, ite.tradeHistory)
	return history
}

// GetMarketData returns current market data
func (ite *IntegrationTradingEngine) GetMarketData(symbol string) map[string]interface{} {
	ite.mu.RLock()
	defer ite.mu.RUnlock()

	book := ite.orderBook[symbol]
	if book == nil {
		return nil
	}

	return map[string]interface{}{
		"symbol":        symbol,
		"last_price":    book.LastPrice,
		"total_volume":  book.TotalVolume,
		"buy_orders":    len(book.BuyOrders),
		"sell_orders":   len(book.SellOrders),
	}
}

// Private helper methods

func (ite *IntegrationTradingEngine) matchOrder(book *OrderBook, order Order) bool {
	executed := false

	if order.Side == "buy" {
		// Match against sell orders
		for i := len(book.SellOrders) - 1; i >= 0 && order.Quantity > 0; i-- {
			sellOrder := &book.SellOrders[i]

			// Check if prices match (buy price >= sell price)
			if order.Price >= sellOrder.Price {
				// Execute trade
				tradeQuantity := min(order.Quantity, sellOrder.Quantity)

				trade := Trade{
					ID:          generateTradeID(),
					Symbol:      book.Symbol,
					BuyOrderID:  order.ID,
					SellOrderID: sellOrder.ID,
					Quantity:    tradeQuantity,
					Price:       sellOrder.Price, // Take maker price
					Time:        time.Now(),
				}

				ite.tradeHistory = append(ite.tradeHistory, trade)

				// Update quantities
				order.Quantity -= tradeQuantity
				sellOrder.Quantity -= tradeQuantity
				book.TotalVolume += tradeQuantity
				book.LastPrice = sellOrder.Price

				executed = true

				// Remove sell order if fully filled
				if sellOrder.Quantity == 0 {
					book.SellOrders = append(book.SellOrders[:i], book.SellOrders[i+1:]...)
				}
			}
		}
	} else { // SELL
		// Match against buy orders
		for i := len(book.BuyOrders) - 1; i >= 0 && order.Quantity > 0; i-- {
			buyOrder := &book.BuyOrders[i]

			// Check if prices match (sell price <= buy price)
			if order.Price <= buyOrder.Price {
				// Execute trade
				tradeQuantity := min(order.Quantity, buyOrder.Quantity)

				trade := Trade{
					ID:          generateTradeID(),
					Symbol:      book.Symbol,
					BuyOrderID:  buyOrder.ID,
					SellOrderID: order.ID,
					Quantity:    tradeQuantity,
					Price:       buyOrder.Price, // Take maker price
					Time:        time.Now(),
				}

				ite.tradeHistory = append(ite.tradeHistory, trade)

				// Update quantities
				order.Quantity -= tradeQuantity
				buyOrder.Quantity -= tradeQuantity
				book.TotalVolume += tradeQuantity
				book.LastPrice = buyOrder.Price

				executed = true

				// Remove buy order if fully filled
				if buyOrder.Quantity == 0 {
					book.BuyOrders = append(book.BuyOrders[:i], book.BuyOrders[i+1:]...)
				}
			}
		}
	}

	return executed
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func generateOrderID() string {
	return "order_" + time.Now().Format("20060102150405") + "_" + string(rune(time.Now().Nanosecond()%1000))
}

func generateTradeID() string {
	return "trade_" + time.Now().Format("20060102150405") + "_" + string(rune(time.Now().Nanosecond()%1000))
}

// Integration Tests

func TestIntegration_FullSimulationWithTradingEngine(t *testing.T) {
	// Create components
	priceGen := NewRealisticPriceGenerator(DefaultPriceGeneratorConfig())
	orderGen := NewRealisticOrderGenerator(DefaultOrderGeneratorConfig())
	eventGen := NewPatternEventGenerator(DefaultEventGeneratorConfig())
	engine := NewIntegrationTradingEngine()

	// Create simulator
	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	// Configure simulation
	config := DefaultSimulationConfig()
	config.SimulationDuration = 5 * time.Second
	config.TickInterval = 100 * time.Millisecond
	config.OrderGenerationRate = 2.0 // 2 orders per second

	// Start simulation
	ctx := context.Background()
	err := simulator.StartSimulation(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start simulation: %v", err)
	}

	// Let simulation run
	time.Sleep(3 * time.Second)

	// Stop simulation
	err = simulator.StopSimulation()
	if err != nil {
		t.Fatalf("Failed to stop simulation: %v", err)
	}

	// Verify results
	orderHistory := engine.GetOrderHistory()
	if len(orderHistory) == 0 {
		t.Error("Expected orders to be placed in trading engine")
	}

	tradeHistory := engine.GetTradeHistory()
	t.Logf("Generated %d orders and %d trades", len(orderHistory), len(tradeHistory))

	// Check simulation statistics
	stats := simulator.GetStatistics()
	if stats.OrdersGenerated == 0 {
		t.Error("Expected orders to be generated")
	}

	if stats.PriceUpdates == 0 {
		t.Error("Expected price updates to occur")
	}

	t.Logf("Simulation stats: %d orders generated, %d price updates, %d events triggered",
		stats.OrdersGenerated, stats.PriceUpdates, stats.EventsTriggered)
}

func TestIntegration_OrderMatching(t *testing.T) {
	engine := NewIntegrationTradingEngine()

	// Place a buy order
	buyOrder := dto.PlaceOrderRequest{
		Symbol:   "BTCUSD",
		Side:     "buy",
		Quantity: 1.0,
		Type:     "limit",
		Price:    50000.0,
	}

	resp1, err := engine.PlaceOrder(buyOrder)
	if err != nil {
		t.Fatalf("Failed to place buy order: %v", err)
	}

	// Place a matching sell order
	sellOrder := dto.PlaceOrderRequest{
		Symbol:   "BTCUSD",
		Side:     "sell",
		Quantity: 0.5,
		Type:     "limit",
		Price:    49999.0, // Below buy price, should match
	}

	resp2, err := engine.PlaceOrder(sellOrder)
	if err != nil {
		t.Fatalf("Failed to place sell order: %v", err)
	}

	// Check that trade occurred
	trades := engine.GetTradeHistory()
	if len(trades) == 0 {
		t.Error("Expected a trade to be executed")
	}

	trade := trades[0]
	if trade.Quantity != 0.5 {
		t.Errorf("Expected trade quantity 0.5, got %f", trade.Quantity)
	}

	if trade.Price != buyOrder.Price {
		t.Errorf("Expected trade price %f, got %f", buyOrder.Price, trade.Price)
	}

	// Check order book
	book := engine.GetOrderBook("BTCUSD")
	if len(book.BuyOrders) != 1 {
		t.Errorf("Expected 1 remaining buy order, got %d", len(book.BuyOrders))
	}

	if book.BuyOrders[0].Quantity != 0.5 {
		t.Errorf("Expected remaining buy quantity 0.5, got %f", book.BuyOrders[0].Quantity)
	}

	t.Logf("Order matching successful: %+v", resp1)
	t.Logf("Trade executed: %+v", resp2)
}

func TestIntegration_MarketConditionsImpact(t *testing.T) {
	priceGen := NewRealisticPriceGenerator(DefaultPriceGeneratorConfig())
	orderGen := NewRealisticOrderGenerator(DefaultOrderGeneratorConfig())
	eventGen := NewPatternEventGenerator(DefaultEventGeneratorConfig())
	engine := NewIntegrationTradingEngine()

	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	// Test different market conditions
	conditions := []MarketCondition{
		MarketSteady,
		MarketVolatile,
		MarketBullish,
		MarketBearish,
	}

	for _, condition := range conditions {
		t.Logf("Testing market condition: %s", condition)

		config := DefaultSimulationConfig()
		config.MarketCondition = condition
		config.SimulationDuration = 2 * time.Second
		config.OrderGenerationRate = 5.0

		ctx := context.Background()
		err := simulator.StartSimulation(ctx, config)
		if err != nil {
			t.Fatalf("Failed to start simulation for condition %s: %v", condition, err)
		}

		time.Sleep(1 * time.Second)

		// Check that orders are being generated
		ordersBefore := len(engine.GetOrderHistory())

		time.Sleep(500 * time.Millisecond)

		ordersAfter := len(engine.GetOrderHistory())
		if ordersAfter <= ordersBefore {
			t.Errorf("Expected more orders for condition %s", condition)
		}

		err = simulator.StopSimulation()
		if err != nil {
			t.Fatalf("Failed to stop simulation for condition %s: %v", condition, err)
		}

		t.Logf("Condition %s generated %d orders", condition, ordersAfter-ordersBefore)
	}
}

func TestIntegration_VolatilityInjection(t *testing.T) {
	priceGen := NewRealisticPriceGenerator(DefaultPriceGeneratorConfig())
	orderGen := NewRealisticOrderGenerator(DefaultOrderGeneratorConfig())
	eventGen := NewPatternEventGenerator(DefaultEventGeneratorConfig())
	engine := NewIntegrationTradingEngine()

	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	config := DefaultSimulationConfig()
	config.SimulationDuration = 8 * time.Second
	config.OrderGenerationRate = 3.0

	ctx := context.Background()
	err := simulator.StartSimulation(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start simulation: %v", err)
	}

	// Record initial state
	time.Sleep(1 * time.Second)
	ordersBefore := len(engine.GetOrderHistory())

	// Inject volatility
	err = simulator.InjectVolatility(VolatilitySpike, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to inject volatility: %v", err)
	}

	// Wait for volatility to affect order generation
	time.Sleep(3 * time.Second)
	ordersAfter := len(engine.GetOrderHistory())

	// Stop simulation
	err = simulator.StopSimulation()
	if err != nil {
		t.Fatalf("Failed to stop simulation: %v", err)
	}

	// Check that volatility increased order generation
	ordersDuringVolatility := ordersAfter - ordersBefore
	if ordersDuringVolatility == 0 {
		t.Error("Expected increased order activity during volatility")
	}

	t.Logf("Orders generated during volatility period: %d", ordersDuringVolatility)
}

func TestIntegration_UserBehaviorPatterns(t *testing.T) {
	orderGen := NewRealisticOrderGenerator(DefaultOrderGeneratorConfig())

	// Test different behavior patterns
	patterns := []UserBehaviorPattern{
		BehaviorConservative,
		BehaviorAggressive,
		BehaviorFOMO,
		BehaviorPanic,
		BehaviorMomentum,
		BehaviorMeanRevert,
		BehaviorArbitrage,
	}

	for _, pattern := range patterns {
		t.Logf("Testing behavior pattern: %s", pattern)

		orders := orderGen.SimulateUserBehavior(pattern, 2.0)

		if len(orders) == 0 {
			t.Errorf("Expected orders for pattern %s", pattern)
			continue
		}

		// Analyze order characteristics
		buyCount, sellCount := 0, 0
		totalQuantity := 0.0
		marketOrderCount := 0

		for _, order := range orders {
			if order.Side == "buy" {
				buyCount++
			} else {
				sellCount++
			}

			totalQuantity += order.Quantity

			if order.Type == "market" {
				marketOrderCount++
			}
		}

		avgQuantity := totalQuantity / float64(len(orders))
		marketOrderRatio := float64(marketOrderCount) / float64(len(orders))

		t.Logf("Pattern %s: %d orders, %.2f avg quantity, %.2f market order ratio, %d buy/%d sell",
			pattern, len(orders), avgQuantity, marketOrderRatio, buyCount, sellCount)

		// Basic validation
		if avgQuantity <= 0 {
			t.Errorf("Average quantity should be positive for pattern %s", pattern)
		}
	}
}

func TestIntegration_MarketSentimentEffects(t *testing.T) {
	orderGen := NewRealisticOrderGenerator(DefaultOrderGeneratorConfig())

	sentiments := []MarketSentiment{
		SentimentOptimistic,
		SentimentPessimistic,
		SentimentNeutral,
		SentimentGreedy,
		SentimentFearful,
	}

	for _, sentiment := range sentiments {
		t.Logf("Testing market sentiment: %s", sentiment)

		orderGen.UpdateMarketSentiment(sentiment)

		orders := orderGen.GenerateRealisticOrders("BTCUSD", 50000.0, MarketSteady)

		if len(orders) == 0 {
			t.Errorf("Expected orders for sentiment %s", sentiment)
			continue
		}

		// Analyze buy/sell distribution
		buyCount, sellCount := 0, 0
		for _, order := range orders {
			if order.Side == "buy" {
				buyCount++
			} else {
				sellCount++
			}
		}

		buyRatio := float64(buyCount) / float64(len(orders))

		t.Logf("Sentiment %s: %.2f buy ratio (%d buy, %d sell)",
			sentiment, buyRatio, buyCount, sellCount)

		// Validate sentiment effects (basic checks)
		switch sentiment {
		case SentimentOptimistic, SentimentGreedy:
			if buyRatio < 0.3 {
				t.Logf("Warning: Expected higher buy ratio for %s, got %.2f", sentiment, buyRatio)
			}
		case SentimentPessimistic, SentimentFearful:
			if buyRatio > 0.7 {
				t.Logf("Warning: Expected lower buy ratio for %s, got %.2f", sentiment, buyRatio)
			}
		}
	}
}

func TestIntegration_EventGeneration(t *testing.T) {
	eventGen := NewPatternEventGenerator(DefaultEventGeneratorConfig())

	// Test event generation
	for i := 0; i < 10; i++ {
		event := eventGen.GenerateMarketEvent()

		if event.ID == "" {
			t.Error("Event should have an ID")
		}

		if event.Type == "" {
			t.Error("Event should have a type")
		}

		if event.Severity == "" {
			t.Error("Event should have a severity")
		}

		if event.PriceImpact <= 0 {
			t.Error("Event should have positive price impact")
		}

		if event.Duration <= 0 {
			t.Error("Event should have positive duration")
		}

		if len(event.AffectedSymbols) == 0 {
			t.Error("Event should affect at least one symbol")
		}

		// Inject event and check it's active
		err := eventGen.InjectEvent(event)
		if err != nil {
			t.Errorf("Failed to inject event: %v", err)
		}

		activeEvents := eventGen.GetActiveEvents()
		found := false
		for _, activeEvent := range activeEvents {
			if activeEvent.ID == event.ID {
				found = true
				break
			}
		}

		if !found {
			t.Error("Injected event should be in active events")
		}

		t.Logf("Generated event: %s (%s) - %.2f%% impact for %v",
			event.Description, event.Severity, event.PriceImpact, event.Duration)
	}
}

func TestIntegration_PriceMovementRealism(t *testing.T) {
	priceGen := NewRealisticPriceGenerator(DefaultPriceGeneratorConfig())
	symbol := "BTCUSD"
	basePrice := 50000.0

	priceGen.SetBasePrice(symbol, basePrice)

	// Generate price series
	prices := make([]float64, 100)
	prices[0] = basePrice

	for i := 1; i < 100; i++ {
		timeElapsed := time.Duration(i) * time.Second
		prices[i] = priceGen.GeneratePrice(symbol, prices[i-1], timeElapsed)
	}

	// Analyze price movement characteristics
	returns := make([]float64, 99)
	for i := 1; i < 100; i++ {
		returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
	}

	// Calculate basic statistics
	meanReturn := 0.0
	for _, ret := range returns {
		meanReturn += ret
	}
	meanReturn /= float64(len(returns))

	variance := 0.0
	for _, ret := range returns {
		variance += (ret - meanReturn) * (ret - meanReturn)
	}
	variance /= float64(len(returns))

	volatility := variance // Simplified volatility measure

	t.Logf("Price series statistics:")
	t.Logf("  Starting price: %.2f", prices[0])
	t.Logf("  Ending price: %.2f", prices[99])
	t.Logf("  Mean return: %.6f", meanReturn)
	t.Logf("  Volatility: %.6f", volatility)

	// Basic realism checks
	if prices[99] <= 0 {
		t.Error("Final price should be positive")
	}

	// Check that not all prices are the same (some movement occurred)
	allSame := true
	for i := 1; i < 100; i++ {
		if prices[i] != prices[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Prices should show some variation")
	}

	// Check that volatility is reasonable
	if volatility > 0.1 {
		t.Logf("Warning: Very high volatility detected: %.6f", volatility)
	}
}

func TestIntegration_SimulationPerformance(t *testing.T) {
	priceGen := NewRealisticPriceGenerator(DefaultPriceGeneratorConfig())
	orderGen := NewRealisticOrderGenerator(DefaultOrderGeneratorConfig())
	eventGen := NewPatternEventGenerator(DefaultEventGeneratorConfig())
	engine := NewIntegrationTradingEngine()

	simulator := NewRealisticSimulator(priceGen, orderGen, eventGen, engine)

	config := DefaultSimulationConfig()
	config.SimulationDuration = 3 * time.Second
	config.OrderGenerationRate = 10.0 // High order rate for performance test

	start := time.Now()

	ctx := context.Background()
	err := simulator.StartSimulation(ctx, config)
	if err != nil {
		t.Fatalf("Failed to start simulation: %v", err)
	}

	// Let simulation run
	time.Sleep(2 * time.Second)

	err = simulator.StopSimulation()
	if err != nil {
		t.Fatalf("Failed to stop simulation: %v", err)
	}

	elapsed := time.Since(start)

	// Check performance metrics
	stats := simulator.GetStatistics()
	orderHistory := engine.GetOrderHistory()

	ordersPerSecond := float64(len(orderHistory)) / elapsed.Seconds()
	priceUpdatesPerSecond := float64(stats.PriceUpdates) / elapsed.Seconds()

	t.Logf("Performance metrics:")
	t.Logf("  Simulation time: %v", elapsed)
	t.Logf("  Orders generated: %d (%.2f/sec)", len(orderHistory), ordersPerSecond)
	t.Logf("  Price updates: %d (%.2f/sec)", stats.PriceUpdates, priceUpdatesPerSecond)
	t.Logf("  Events triggered: %d", stats.EventsTriggered)
	t.Logf("  Error count: %d", stats.ErrorCount)

	// Performance validation
	if ordersPerSecond < 1.0 {
		t.Error("Order generation rate too low")
	}

	if priceUpdatesPerSecond < 0.5 {
		t.Error("Price update rate too low")
	}

	if stats.ErrorCount > 0 {
		t.Errorf("Simulation had %d errors", stats.ErrorCount)
	}
}

// Benchmark tests

func BenchmarkSimulation_OrderGeneration(b *testing.B) {
	orderGen := NewRealisticOrderGenerator(DefaultOrderGeneratorConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		orderGen.GenerateRealisticOrders("BTCUSD", 50000.0, MarketSteady)
	}
}

func BenchmarkSimulation_PriceGeneration(b *testing.B) {
	priceGen := NewRealisticPriceGenerator(DefaultPriceGeneratorConfig())
	priceGen.SetBasePrice("BTCUSD", 50000.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		priceGen.GeneratePrice("BTCUSD", 50000.0, time.Duration(i)*time.Second)
	}
}

func BenchmarkSimulation_EventGeneration(b *testing.B) {
	eventGen := NewPatternEventGenerator(DefaultEventGeneratorConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eventGen.GenerateMarketEvent()
	}
}

func BenchmarkSimulation_TradingEngine(b *testing.B) {
	engine := NewIntegrationTradingEngine()

	order := dto.PlaceOrderRequest{
		Symbol:   "BTCUSD",
		Side:     "buy",
		Quantity: 1.0,
		Type:     "market",
		Price:    50000.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.PlaceOrder(order)
	}
}