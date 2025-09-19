package metrics

import (
	"sort"
	"sync"
	"time"
)

// RealTimeMetrics implements MetricsCollector with time-windowed metrics calculation
type RealTimeMetrics struct {
	mutex         sync.RWMutex
	orderEvents   []OrderEvent
	tradeEvents   []TradeEvent
	windowSize    time.Duration
	symbolMetrics map[string]*symbolData
}

type symbolData struct {
	orderCount int64
	tradeCount int64
	volume     float64
	prices     []float64
	latencies  []time.Duration
}

// NewRealTimeMetrics creates a new RealTimeMetrics collector
func NewRealTimeMetrics(windowSize time.Duration) *RealTimeMetrics {
	if windowSize == 0 {
		windowSize = 60 * time.Second // Default to 60 seconds
	}

	return &RealTimeMetrics{
		orderEvents:   make([]OrderEvent, 0),
		tradeEvents:   make([]TradeEvent, 0),
		windowSize:    windowSize,
		symbolMetrics: make(map[string]*symbolData),
	}
}

// RecordOrder records an order event
func (rtm *RealTimeMetrics) RecordOrder(event OrderEvent) {
	rtm.mutex.Lock()
	defer rtm.mutex.Unlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	rtm.orderEvents = append(rtm.orderEvents, event)
	rtm.updateSymbolMetrics(event.Symbol, true, false, 0, 0, event.Latency)
	rtm.cleanOldEvents()
}

// RecordTrade records a trade event
func (rtm *RealTimeMetrics) RecordTrade(event TradeEvent) {
	rtm.mutex.Lock()
	defer rtm.mutex.Unlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	rtm.tradeEvents = append(rtm.tradeEvents, event)
	rtm.updateSymbolMetrics(event.Symbol, false, true, event.Quantity, event.Price, event.Latency)
	rtm.cleanOldEvents()
}

// CalculateMetrics calculates metrics for the specified time window
func (rtm *RealTimeMetrics) CalculateMetrics(windowDuration time.Duration) MetricsSnapshot {
	rtm.mutex.RLock()
	defer rtm.mutex.RUnlock()

	now := time.Now()
	windowStart := now.Add(-windowDuration)

	return rtm.calculateMetricsForWindow(windowStart, now)
}

// GetCurrentMetrics returns current metrics for the default window size
func (rtm *RealTimeMetrics) GetCurrentMetrics() MetricsSnapshot {
	return rtm.CalculateMetrics(rtm.windowSize)
}

// Reset clears all collected metrics
func (rtm *RealTimeMetrics) Reset() {
	rtm.mutex.Lock()
	defer rtm.mutex.Unlock()

	rtm.orderEvents = rtm.orderEvents[:0]
	rtm.tradeEvents = rtm.tradeEvents[:0]
	rtm.symbolMetrics = make(map[string]*symbolData)
}

// calculateMetricsForWindow calculates metrics for a specific time window
func (rtm *RealTimeMetrics) calculateMetricsForWindow(windowStart, windowEnd time.Time) MetricsSnapshot {
	var orderCount, tradeCount int64
	var totalVolume float64
	var latencies []time.Duration
	symbolMetrics := make(map[string]SymbolMetrics)
	symbolDataMap := make(map[string]*symbolData)

	// Process order events
	for _, event := range rtm.orderEvents {
		if event.Timestamp.After(windowStart) && event.Timestamp.Before(windowEnd) {
			orderCount++
			latencies = append(latencies, event.Latency)

			if _, exists := symbolDataMap[event.Symbol]; !exists {
				symbolDataMap[event.Symbol] = &symbolData{
					prices:    make([]float64, 0),
					latencies: make([]time.Duration, 0),
				}
			}
			symbolDataMap[event.Symbol].orderCount++
			symbolDataMap[event.Symbol].latencies = append(symbolDataMap[event.Symbol].latencies, event.Latency)
		}
	}

	// Process trade events
	for _, event := range rtm.tradeEvents {
		if event.Timestamp.After(windowStart) && event.Timestamp.Before(windowEnd) {
			tradeCount++
			totalVolume += event.Quantity
			latencies = append(latencies, event.Latency)

			if _, exists := symbolDataMap[event.Symbol]; !exists {
				symbolDataMap[event.Symbol] = &symbolData{
					prices:    make([]float64, 0),
					latencies: make([]time.Duration, 0),
				}
			}
			symbolDataMap[event.Symbol].tradeCount++
			symbolDataMap[event.Symbol].volume += event.Quantity
			symbolDataMap[event.Symbol].prices = append(symbolDataMap[event.Symbol].prices, event.Price)
			symbolDataMap[event.Symbol].latencies = append(symbolDataMap[event.Symbol].latencies, event.Latency)
		}
	}

	// Calculate aggregated latency metrics
	avgLatency, maxLatency, minLatency := calculateLatencyMetrics(latencies)

	// Calculate per-symbol metrics
	for symbol, data := range symbolDataMap {
		symbolMetrics[symbol] = rtm.calculateSymbolMetrics(symbol, data)
	}

	// Calculate rates (per second)
	windowDuration := windowEnd.Sub(windowStart).Seconds()
	ordersPerSec := float64(orderCount) / windowDuration
	tradesPerSec := float64(tradeCount) / windowDuration
	volumePerSec := totalVolume / windowDuration

	return MetricsSnapshot{
		WindowStart:   windowStart,
		WindowEnd:     windowEnd,
		OrderCount:    orderCount,
		TradeCount:    tradeCount,
		TotalVolume:   totalVolume,
		AvgLatency:    avgLatency,
		MaxLatency:    maxLatency,
		MinLatency:    minLatency,
		OrdersPerSec:  ordersPerSec,
		TradesPerSec:  tradesPerSec,
		VolumePerSec:  volumePerSec,
		SymbolMetrics: symbolMetrics,
	}
}

// calculateSymbolMetrics calculates metrics for a specific symbol
func (rtm *RealTimeMetrics) calculateSymbolMetrics(symbol string, data *symbolData) SymbolMetrics {
	metrics := SymbolMetrics{
		Symbol:     symbol,
		OrderCount: data.orderCount,
		TradeCount: data.tradeCount,
		Volume:     data.volume,
	}

	if len(data.prices) > 0 {
		sort.Float64s(data.prices)
		metrics.HighPrice = data.prices[len(data.prices)-1]
		metrics.LowPrice = data.prices[0]
		metrics.LastPrice = data.prices[len(data.prices)-1]

		// Calculate average price
		sum := 0.0
		for _, price := range data.prices {
			sum += price
		}
		metrics.AvgPrice = sum / float64(len(data.prices))
	}

	if len(data.latencies) > 0 {
		avg, _, _ := calculateLatencyMetrics(data.latencies)
		metrics.AvgLatency = avg
	}

	return metrics
}

// updateSymbolMetrics updates the symbol-specific metrics
func (rtm *RealTimeMetrics) updateSymbolMetrics(symbol string, isOrder, isTrade bool, quantity, price float64, latency time.Duration) {
	if _, exists := rtm.symbolMetrics[symbol]; !exists {
		rtm.symbolMetrics[symbol] = &symbolData{
			prices:    make([]float64, 0),
			latencies: make([]time.Duration, 0),
		}
	}

	data := rtm.symbolMetrics[symbol]

	if isOrder {
		data.orderCount++
	}

	if isTrade {
		data.tradeCount++
		data.volume += quantity
		data.prices = append(data.prices, price)
	}

	data.latencies = append(data.latencies, latency)
}

// cleanOldEvents removes events outside the current window
func (rtm *RealTimeMetrics) cleanOldEvents() {
	now := time.Now()
	cutoff := now.Add(-rtm.windowSize)

	// Clean order events
	validOrders := rtm.orderEvents[:0]
	for _, event := range rtm.orderEvents {
		if event.Timestamp.After(cutoff) {
			validOrders = append(validOrders, event)
		}
	}
	rtm.orderEvents = validOrders

	// Clean trade events
	validTrades := rtm.tradeEvents[:0]
	for _, event := range rtm.tradeEvents {
		if event.Timestamp.After(cutoff) {
			validTrades = append(validTrades, event)
		}
	}
	rtm.tradeEvents = validTrades
}

// calculateLatencyMetrics calculates average, max, and min latency
func calculateLatencyMetrics(latencies []time.Duration) (avg, max, min time.Duration) {
	if len(latencies) == 0 {
		return 0, 0, 0
	}

	var total time.Duration
	min = latencies[0]
	max = latencies[0]

	for _, latency := range latencies {
		total += latency
		if latency < min {
			min = latency
		}
		if latency > max {
			max = latency
		}
	}

	avg = total / time.Duration(len(latencies))
	return avg, max, min
}