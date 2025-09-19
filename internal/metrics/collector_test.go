package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"simulated_exchange/internal/types"
)

func TestNewRealTimeMetrics(t *testing.T) {
	t.Run("default window size", func(t *testing.T) {
		collector := NewRealTimeMetrics(0)
		assert.NotNil(t, collector)
		assert.Equal(t, 60*time.Second, collector.windowSize)
	})

	t.Run("custom window size", func(t *testing.T) {
		windowSize := 30 * time.Second
		collector := NewRealTimeMetrics(windowSize)
		assert.NotNil(t, collector)
		assert.Equal(t, windowSize, collector.windowSize)
	})
}

func TestRealTimeMetrics_RecordOrder(t *testing.T) {
	collector := NewRealTimeMetrics(60 * time.Second)

	t.Run("record single order", func(t *testing.T) {
		event := OrderEvent{
			OrderID:   "order1",
			Symbol:    "AAPL",
			Side:      types.Buy,
			Type:      types.Limit,
			Quantity:  100,
			Price:     150.0,
			Timestamp: time.Now(),
			Latency:   10 * time.Millisecond,
		}

		collector.RecordOrder(event)

		metrics := collector.GetCurrentMetrics()
		assert.Equal(t, int64(1), metrics.OrderCount)
		assert.Equal(t, int64(0), metrics.TradeCount)
		assert.Equal(t, 10*time.Millisecond, metrics.AvgLatency)
	})

	t.Run("record multiple orders", func(t *testing.T) {
		collector.Reset()

		events := []OrderEvent{
			{
				OrderID:   "order1",
				Symbol:    "AAPL",
				Side:      types.Buy,
				Type:      types.Limit,
				Quantity:  100,
				Price:     150.0,
				Timestamp: time.Now(),
				Latency:   10 * time.Millisecond,
			},
			{
				OrderID:   "order2",
				Symbol:    "AAPL",
				Side:      types.Sell,
				Type:      types.Limit,
				Quantity:  50,
				Price:     149.0,
				Timestamp: time.Now(),
				Latency:   20 * time.Millisecond,
			},
		}

		for _, event := range events {
			collector.RecordOrder(event)
		}

		metrics := collector.GetCurrentMetrics()
		assert.Equal(t, int64(2), metrics.OrderCount)
		assert.Equal(t, 15*time.Millisecond, metrics.AvgLatency) // (10+20)/2
		assert.Equal(t, 20*time.Millisecond, metrics.MaxLatency)
		assert.Equal(t, 10*time.Millisecond, metrics.MinLatency)
	})

	t.Run("automatic timestamp assignment", func(t *testing.T) {
		collector.Reset()
		startTime := time.Now()

		event := OrderEvent{
			OrderID:  "order1",
			Symbol:   "AAPL",
			Side:     types.Buy,
			Type:     types.Limit,
			Quantity: 100,
			Price:    150.0,
			// Timestamp intentionally left zero
			Latency: 10 * time.Millisecond,
		}

		collector.RecordOrder(event)

		// Verify timestamp was set automatically
		assert.True(t, event.Timestamp.After(startTime) || event.Timestamp.Equal(startTime))
	})
}

func TestRealTimeMetrics_RecordTrade(t *testing.T) {
	collector := NewRealTimeMetrics(60 * time.Second)

	t.Run("record single trade", func(t *testing.T) {
		event := TradeEvent{
			TradeID:     "trade1",
			Symbol:      "AAPL",
			Quantity:    50,
			Price:       149.5,
			Timestamp:   time.Now(),
			Latency:     5 * time.Millisecond,
			BuyOrderID:  "buy1",
			SellOrderID: "sell1",
		}

		collector.RecordTrade(event)

		metrics := collector.GetCurrentMetrics()
		assert.Equal(t, int64(0), metrics.OrderCount)
		assert.Equal(t, int64(1), metrics.TradeCount)
		assert.Equal(t, 50.0, metrics.TotalVolume)
		assert.Equal(t, 5*time.Millisecond, metrics.AvgLatency)
	})

	t.Run("record multiple trades", func(t *testing.T) {
		collector.Reset()

		events := []TradeEvent{
			{
				TradeID:     "trade1",
				Symbol:      "AAPL",
				Quantity:    50,
				Price:       149.5,
				Timestamp:   time.Now(),
				Latency:     5 * time.Millisecond,
				BuyOrderID:  "buy1",
				SellOrderID: "sell1",
			},
			{
				TradeID:     "trade2",
				Symbol:      "AAPL",
				Quantity:    30,
				Price:       150.0,
				Timestamp:   time.Now(),
				Latency:     15 * time.Millisecond,
				BuyOrderID:  "buy2",
				SellOrderID: "sell2",
			},
		}

		for _, event := range events {
			collector.RecordTrade(event)
		}

		metrics := collector.GetCurrentMetrics()
		assert.Equal(t, int64(2), metrics.TradeCount)
		assert.Equal(t, 80.0, metrics.TotalVolume) // 50 + 30
		assert.Equal(t, 10*time.Millisecond, metrics.AvgLatency) // (5+15)/2
	})
}

func TestRealTimeMetrics_CalculateMetrics_TimeWindows(t *testing.T) {
	collector := NewRealTimeMetrics(60 * time.Second)

	t.Run("time window filtering", func(t *testing.T) {
		now := time.Now()

		// Record events at different times
		oldEvent := OrderEvent{
			OrderID:   "old_order",
			Symbol:    "AAPL",
			Side:      types.Buy,
			Type:      types.Limit,
			Quantity:  100,
			Price:     150.0,
			Timestamp: now.Add(-90 * time.Second), // Outside 60s window
			Latency:   10 * time.Millisecond,
		}

		recentEvent := OrderEvent{
			OrderID:   "recent_order",
			Symbol:    "AAPL",
			Side:      types.Sell,
			Type:      types.Limit,
			Quantity:  50,
			Price:     149.0,
			Timestamp: now.Add(-30 * time.Second), // Within 60s window
			Latency:   20 * time.Millisecond,
		}

		collector.RecordOrder(oldEvent)
		collector.RecordOrder(recentEvent)

		// Calculate metrics for 60-second window
		metrics := collector.CalculateMetrics(60 * time.Second)

		// Should only include the recent event
		assert.Equal(t, int64(1), metrics.OrderCount)
		assert.Equal(t, 20*time.Millisecond, metrics.AvgLatency)
	})

	t.Run("custom window duration", func(t *testing.T) {
		collector.Reset()
		now := time.Now()

		events := []OrderEvent{
			{
				OrderID:   "order1",
				Symbol:    "AAPL",
				Timestamp: now.Add(-45 * time.Second),
				Latency:   10 * time.Millisecond,
			},
			{
				OrderID:   "order2",
				Symbol:    "AAPL",
				Timestamp: now.Add(-15 * time.Second),
				Latency:   20 * time.Millisecond,
			},
		}

		for _, event := range events {
			collector.RecordOrder(event)
		}

		// 30-second window should only include order2
		metrics30s := collector.CalculateMetrics(30 * time.Second)
		assert.Equal(t, int64(1), metrics30s.OrderCount)

		// 60-second window should include both
		metrics60s := collector.CalculateMetrics(60 * time.Second)
		assert.Equal(t, int64(2), metrics60s.OrderCount)
	})
}

func TestRealTimeMetrics_SymbolMetrics(t *testing.T) {
	collector := NewRealTimeMetrics(60 * time.Second)

	t.Run("per-symbol metrics calculation", func(t *testing.T) {
		// Record events for different symbols
		events := []interface{}{
			OrderEvent{
				OrderID:   "order1",
				Symbol:    "AAPL",
				Side:      types.Buy,
				Timestamp: time.Now(),
				Latency:   10 * time.Millisecond,
			},
			TradeEvent{
				TradeID:   "trade1",
				Symbol:    "AAPL",
				Quantity:  100,
				Price:     150.0,
				Timestamp: time.Now(),
				Latency:   5 * time.Millisecond,
			},
			OrderEvent{
				OrderID:   "order2",
				Symbol:    "GOOGL",
				Side:      types.Sell,
				Timestamp: time.Now(),
				Latency:   15 * time.Millisecond,
			},
			TradeEvent{
				TradeID:   "trade2",
				Symbol:    "GOOGL",
				Quantity:  50,
				Price:     2800.0,
				Timestamp: time.Now(),
				Latency:   8 * time.Millisecond,
			},
		}

		for _, event := range events {
			switch e := event.(type) {
			case OrderEvent:
				collector.RecordOrder(e)
			case TradeEvent:
				collector.RecordTrade(e)
			}
		}

		metrics := collector.GetCurrentMetrics()

		// Verify overall metrics
		assert.Equal(t, int64(2), metrics.OrderCount)
		assert.Equal(t, int64(2), metrics.TradeCount)

		// Verify AAPL metrics
		aaplMetrics, exists := metrics.SymbolMetrics["AAPL"]
		require.True(t, exists)
		assert.Equal(t, "AAPL", aaplMetrics.Symbol)
		assert.Equal(t, int64(1), aaplMetrics.OrderCount)
		assert.Equal(t, int64(1), aaplMetrics.TradeCount)
		assert.Equal(t, 100.0, aaplMetrics.Volume)
		assert.Equal(t, 150.0, aaplMetrics.AvgPrice)
		assert.Equal(t, 150.0, aaplMetrics.HighPrice)
		assert.Equal(t, 150.0, aaplMetrics.LowPrice)

		// Verify GOOGL metrics
		googlMetrics, exists := metrics.SymbolMetrics["GOOGL"]
		require.True(t, exists)
		assert.Equal(t, "GOOGL", googlMetrics.Symbol)
		assert.Equal(t, int64(1), googlMetrics.OrderCount)
		assert.Equal(t, int64(1), googlMetrics.TradeCount)
		assert.Equal(t, 50.0, googlMetrics.Volume)
		assert.Equal(t, 2800.0, googlMetrics.AvgPrice)
	})

	t.Run("price statistics calculation", func(t *testing.T) {
		collector.Reset()

		trades := []TradeEvent{
			{
				TradeID:   "trade1",
				Symbol:    "AAPL",
				Price:     148.0,
				Quantity:  100,
				Timestamp: time.Now(),
			},
			{
				TradeID:   "trade2",
				Symbol:    "AAPL",
				Price:     152.0,
				Quantity:  50,
				Timestamp: time.Now(),
			},
			{
				TradeID:   "trade3",
				Symbol:    "AAPL",
				Price:     150.0,
				Quantity:  75,
				Timestamp: time.Now(),
			},
		}

		for _, trade := range trades {
			collector.RecordTrade(trade)
		}

		metrics := collector.GetCurrentMetrics()
		aaplMetrics := metrics.SymbolMetrics["AAPL"]

		assert.Equal(t, 148.0, aaplMetrics.LowPrice)
		assert.Equal(t, 152.0, aaplMetrics.HighPrice)
		assert.Equal(t, 150.0, aaplMetrics.LastPrice)
		assert.Equal(t, 150.0, aaplMetrics.AvgPrice) // (148+152+150)/3
	})
}

func TestRealTimeMetrics_RateCalculations(t *testing.T) {
	collector := NewRealTimeMetrics(60 * time.Second)

	t.Run("calculate orders and trades per second", func(t *testing.T) {
		now := time.Now()
		windowStart := now.Add(-10 * time.Second)

		// Record 5 orders and 3 trades over 10 seconds
		for i := 0; i < 5; i++ {
			collector.RecordOrder(OrderEvent{
				OrderID:   "order" + string(rune('1'+i)),
				Symbol:    "AAPL",
				Timestamp: windowStart.Add(time.Duration(i) * 2 * time.Second),
				Latency:   time.Millisecond,
			})
		}

		for i := 0; i < 3; i++ {
			collector.RecordTrade(TradeEvent{
				TradeID:   "trade" + string(rune('1'+i)),
				Symbol:    "AAPL",
				Quantity:  100,
				Price:     150.0,
				Timestamp: windowStart.Add(time.Duration(i) * 3 * time.Second),
				Latency:   time.Millisecond,
			})
		}

		metrics := collector.CalculateMetrics(10 * time.Second)

		// 5 orders over 10 seconds = 0.5 orders/sec
		assert.InDelta(t, 0.5, metrics.OrdersPerSec, 0.1)

		// 3 trades over 10 seconds = 0.3 trades/sec
		assert.InDelta(t, 0.3, metrics.TradesPerSec, 0.1)

		// 300 volume over 10 seconds = 30 volume/sec
		assert.InDelta(t, 30.0, metrics.VolumePerSec, 1.0)
	})
}

func TestRealTimeMetrics_Reset(t *testing.T) {
	collector := NewRealTimeMetrics(60 * time.Second)

	// Add some data
	collector.RecordOrder(OrderEvent{
		OrderID:   "order1",
		Symbol:    "AAPL",
		Timestamp: time.Now(),
		Latency:   10 * time.Millisecond,
	})

	collector.RecordTrade(TradeEvent{
		TradeID:   "trade1",
		Symbol:    "AAPL",
		Quantity:  100,
		Price:     150.0,
		Timestamp: time.Now(),
		Latency:   5 * time.Millisecond,
	})

	// Verify data exists
	metrics := collector.GetCurrentMetrics()
	assert.Equal(t, int64(1), metrics.OrderCount)
	assert.Equal(t, int64(1), metrics.TradeCount)

	// Reset and verify data is cleared
	collector.Reset()
	metricsAfterReset := collector.GetCurrentMetrics()
	assert.Equal(t, int64(0), metricsAfterReset.OrderCount)
	assert.Equal(t, int64(0), metricsAfterReset.TradeCount)
	assert.Equal(t, 0.0, metricsAfterReset.TotalVolume)
	assert.Len(t, metricsAfterReset.SymbolMetrics, 0)
}

func TestRealTimeMetrics_ThreadSafety(t *testing.T) {
	collector := NewRealTimeMetrics(60 * time.Second)

	// Test concurrent access
	const numGoroutines = 10
	const eventsPerGoroutine = 100

	done := make(chan bool, numGoroutines*2)

	// Start goroutines recording orders
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < eventsPerGoroutine; j++ {
				collector.RecordOrder(OrderEvent{
					OrderID:   "order" + string(rune('0'+routineID)) + string(rune('0'+j)),
					Symbol:    "AAPL",
					Timestamp: time.Now(),
					Latency:   time.Millisecond,
				})
			}
			done <- true
		}(i)
	}

	// Start goroutines recording trades
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < eventsPerGoroutine; j++ {
				collector.RecordTrade(TradeEvent{
					TradeID:   "trade" + string(rune('0'+routineID)) + string(rune('0'+j)),
					Symbol:    "AAPL",
					Quantity:  10,
					Price:     150.0,
					Timestamp: time.Now(),
					Latency:   time.Millisecond,
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines*2; i++ {
		<-done
	}

	// Verify final metrics
	metrics := collector.GetCurrentMetrics()
	assert.Equal(t, int64(numGoroutines*eventsPerGoroutine), metrics.OrderCount)
	assert.Equal(t, int64(numGoroutines*eventsPerGoroutine), metrics.TradeCount)
	assert.Equal(t, float64(numGoroutines*eventsPerGoroutine*10), metrics.TotalVolume)
}

func TestRealTimeMetrics_EventCleanup(t *testing.T) {
	// Use a short window for testing cleanup
	collector := NewRealTimeMetrics(5 * time.Second)

	now := time.Now()

	// Record old events that should be cleaned up
	oldEvent := OrderEvent{
		OrderID:   "old_order",
		Symbol:    "AAPL",
		Timestamp: now.Add(-10 * time.Second), // Outside window
		Latency:   10 * time.Millisecond,
	}

	recentEvent := OrderEvent{
		OrderID:   "recent_order",
		Symbol:    "AAPL",
		Timestamp: now.Add(-2 * time.Second), // Within window
		Latency:   20 * time.Millisecond,
	}

	collector.RecordOrder(oldEvent)
	collector.RecordOrder(recentEvent)

	// Trigger cleanup by recording another event
	collector.RecordOrder(OrderEvent{
		OrderID:   "trigger_cleanup",
		Symbol:    "AAPL",
		Timestamp: now,
		Latency:   15 * time.Millisecond,
	})

	// Verify that only recent events are counted
	metrics := collector.GetCurrentMetrics()

	// Should only count recent events (within 5-second window)
	assert.LessOrEqual(t, metrics.OrderCount, int64(2))
	assert.Greater(t, metrics.OrderCount, int64(0))
}