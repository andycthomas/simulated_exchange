package engine

import (
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"simulated_exchange/internal/repository"
	"simulated_exchange/internal/types"
)

func TestTradingEngine_Integration_FullOrderFlow(t *testing.T) {
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	engine := NewTradingEngine(orderRepo, tradeRepo, matcher, executor)

	sellOrder := types.Order{
		ID:       "sell1",
		Symbol:   "AAPL",
		Side:     types.Sell,
		Type:     types.Limit,
		Quantity: 100,
		Price:    150.0,
	}

	err := engine.PlaceOrder(sellOrder)
	require.NoError(t, err)

	buyOrder := types.Order{
		ID:       "buy1",
		Symbol:   "AAPL",
		Side:     types.Buy,
		Type:     types.Limit,
		Quantity: 80,
		Price:    150.0,
	}

	err = engine.PlaceOrder(buyOrder)
	require.NoError(t, err)

	orderBook, err := engine.GetOrderBook("AAPL")
	require.NoError(t, err)

	assert.Equal(t, "AAPL", orderBook.Symbol)
	assert.Len(t, orderBook.Bids, 0)
	assert.Len(t, orderBook.Asks, 1)
	assert.Equal(t, 20.0, orderBook.Asks[0].Quantity)

	trades, err := tradeRepo.GetBySymbol("AAPL")
	require.NoError(t, err)
	assert.Len(t, trades, 1)
	assert.Equal(t, 80.0, trades[0].Quantity)
	assert.Equal(t, 150.0, trades[0].Price)
}

func TestTradingEngine_Integration_MarketOrder(t *testing.T) {
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	engine := NewTradingEngine(orderRepo, tradeRepo, matcher, executor)

	sellOrder1 := types.Order{
		ID:       "sell1",
		Symbol:   "AAPL",
		Side:     types.Sell,
		Type:     types.Limit,
		Quantity: 50,
		Price:    150.0,
	}

	sellOrder2 := types.Order{
		ID:       "sell2",
		Symbol:   "AAPL",
		Side:     types.Sell,
		Type:     types.Limit,
		Quantity: 30,
		Price:    149.0,
	}

	err := engine.PlaceOrder(sellOrder1)
	require.NoError(t, err)

	err = engine.PlaceOrder(sellOrder2)
	require.NoError(t, err)

	marketBuyOrder := types.Order{
		ID:       "buy1",
		Symbol:   "AAPL",
		Side:     types.Buy,
		Type:     types.Market,
		Quantity: 60,
	}

	err = engine.PlaceOrder(marketBuyOrder)
	require.NoError(t, err)

	orderBook, err := engine.GetOrderBook("AAPL")
	require.NoError(t, err)

	assert.Len(t, orderBook.Bids, 0)
	assert.Len(t, orderBook.Asks, 1)
	assert.Equal(t, 20.0, orderBook.Asks[0].Quantity)
	assert.Equal(t, 150.0, orderBook.Asks[0].Price)

	trades, err := tradeRepo.GetBySymbol("AAPL")
	require.NoError(t, err)
	assert.Len(t, trades, 2)
}

func TestTradingEngine_Integration_PriceTimePriority(t *testing.T) {
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	engine := NewTradingEngine(orderRepo, tradeRepo, matcher, executor)

	time1 := time.Now()
	time2 := time1.Add(time.Second)

	sellOrder1 := types.Order{
		ID:        "sell1",
		Symbol:    "AAPL",
		Side:      types.Sell,
		Type:      types.Limit,
		Quantity:  50,
		Price:     150.0,
		Timestamp: time1,
	}

	sellOrder2 := types.Order{
		ID:        "sell2",
		Symbol:    "AAPL",
		Side:      types.Sell,
		Type:      types.Limit,
		Quantity:  30,
		Price:     150.0,
		Timestamp: time2,
	}

	sellOrder3 := types.Order{
		ID:        "sell3",
		Symbol:    "AAPL",
		Side:      types.Sell,
		Type:      types.Limit,
		Quantity:  40,
		Price:     149.0,
		Timestamp: time2,
	}

	err := engine.PlaceOrder(sellOrder1)
	require.NoError(t, err)

	err = engine.PlaceOrder(sellOrder2)
	require.NoError(t, err)

	err = engine.PlaceOrder(sellOrder3)
	require.NoError(t, err)

	buyOrder := types.Order{
		ID:       "buy1",
		Symbol:   "AAPL",
		Side:     types.Buy,
		Type:     types.Limit,
		Quantity: 100,
		Price:    151.0,
	}

	err = engine.PlaceOrder(buyOrder)
	require.NoError(t, err)

	trades, err := tradeRepo.GetBySymbol("AAPL")
	require.NoError(t, err)
	assert.Len(t, trades, 3)

	// Sort trades by timestamp to ensure deterministic order
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].Timestamp.Before(trades[j].Timestamp)
	})

	assert.Equal(t, 149.0, trades[0].Price)
	assert.Equal(t, 40.0, trades[0].Quantity)

	assert.Equal(t, 150.0, trades[1].Price)
	assert.Equal(t, 50.0, trades[1].Quantity)

	assert.Equal(t, 150.0, trades[2].Price)
	assert.Equal(t, 10.0, trades[2].Quantity)

	orderBook, err := engine.GetOrderBook("AAPL")
	require.NoError(t, err)
	assert.Len(t, orderBook.Bids, 0)
	assert.Len(t, orderBook.Asks, 1)
	assert.Equal(t, 20.0, orderBook.Asks[0].Quantity)
}

func TestTradingEngine_Integration_CancelOrder(t *testing.T) {
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	engine := NewTradingEngine(orderRepo, tradeRepo, matcher, executor)

	order := types.Order{
		ID:       "order1",
		Symbol:   "AAPL",
		Side:     types.Buy,
		Type:     types.Limit,
		Quantity: 100,
		Price:    150.0,
	}

	err := engine.PlaceOrder(order)
	require.NoError(t, err)

	orderBook, err := engine.GetOrderBook("AAPL")
	require.NoError(t, err)
	assert.Len(t, orderBook.Bids, 1)

	err = engine.CancelOrder("order1")
	require.NoError(t, err)

	orderBook, err = engine.GetOrderBook("AAPL")
	require.NoError(t, err)
	assert.Len(t, orderBook.Bids, 0)
}

func TestTradingEngine_Integration_ThreadSafety(t *testing.T) {
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	engine := NewTradingEngine(orderRepo, tradeRepo, matcher, executor)

	const numGoroutines = 10
	const ordersPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < ordersPerGoroutine; j++ {
				order := types.Order{
					Symbol:   "AAPL",
					Side:     types.Buy,
					Type:     types.Limit,
					Quantity: 10,
					Price:    float64(150 + routineID),
				}
				engine.PlaceOrder(order)
			}
		}(i)

		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < ordersPerGoroutine; j++ {
				order := types.Order{
					Symbol:   "AAPL",
					Side:     types.Sell,
					Type:     types.Limit,
					Quantity: 10,
					Price:    float64(150 + routineID),
				}
				engine.PlaceOrder(order)
			}
		}(i)
	}

	wg.Wait()

	orderBook, err := engine.GetOrderBook("AAPL")
	require.NoError(t, err)

	trades, err := tradeRepo.GetAll()
	require.NoError(t, err)

	totalTradeQuantity := 0.0
	for _, trade := range trades {
		totalTradeQuantity += trade.Quantity
	}

	totalOrderQuantity := 0.0
	for _, bid := range orderBook.Bids {
		totalOrderQuantity += bid.Quantity
	}
	for _, ask := range orderBook.Asks {
		totalOrderQuantity += ask.Quantity
	}

	expectedTotalQuantity := float64(numGoroutines * ordersPerGoroutine * 2 * 10)
	actualTotalQuantity := totalTradeQuantity*2 + totalOrderQuantity

	assert.Equal(t, expectedTotalQuantity, actualTotalQuantity)
}

func TestTradingEngine_Integration_PartialFills(t *testing.T) {
	orderRepo := repository.NewMemoryOrderRepository()
	tradeRepo := repository.NewMemoryTradeRepository()
	matcher := NewPriceTimeOrderMatcher()
	executor := NewSimpleTradeExecutor(tradeRepo)

	engine := NewTradingEngine(orderRepo, tradeRepo, matcher, executor)

	bigSellOrder := types.Order{
		ID:       "sell1",
		Symbol:   "AAPL",
		Side:     types.Sell,
		Type:     types.Limit,
		Quantity: 1000,
		Price:    150.0,
	}

	err := engine.PlaceOrder(bigSellOrder)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		buyOrder := types.Order{
			Symbol:   "AAPL",
			Side:     types.Buy,
			Type:     types.Limit,
			Quantity: 100,
			Price:    150.0,
		}

		err = engine.PlaceOrder(buyOrder)
		require.NoError(t, err)
	}

	orderBook, err := engine.GetOrderBook("AAPL")
	require.NoError(t, err)

	assert.Len(t, orderBook.Bids, 0)
	assert.Len(t, orderBook.Asks, 1)
	assert.Equal(t, 500.0, orderBook.Asks[0].Quantity)

	trades, err := tradeRepo.GetBySymbol("AAPL")
	require.NoError(t, err)
	assert.Len(t, trades, 5)

	totalTradeQuantity := 0.0
	for _, trade := range trades {
		totalTradeQuantity += trade.Quantity
		assert.Equal(t, 150.0, trade.Price)
		assert.Equal(t, 100.0, trade.Quantity)
	}
	assert.Equal(t, 500.0, totalTradeQuantity)
}