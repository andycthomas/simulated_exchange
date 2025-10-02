package domain

import (
	"time"
	"simulated_exchange/pkg/shared"
)

// CanGenerateOrder checks if we can generate another order based on rate limits
func (og *OrderGenerator) CanGenerateOrder() bool {
	now := time.Now()

	// Reset counter if new minute
	if now.Sub(og.minuteStart) >= time.Minute {
		og.orderCount = 0
		og.minuteStart = now
	}

	// Check if we've exceeded the per-minute limit
	if og.orderCount >= og.config.MaxOrdersPerMinute {
		og.logger.Debug("Order generation rate limited",
			"orders_this_minute", og.orderCount,
			"max_per_minute", og.config.MaxOrdersPerMinute)
		return false
	}

	return true
}

// AddToBuffer adds an order to the batch buffer
func (og *OrderGenerator) AddToBuffer(order *shared.Order) {
	og.orderBuffer = append(og.orderBuffer, order)
	og.orderCount++
}

// ShouldFlushBuffer determines if the buffer should be flushed
func (og *OrderGenerator) ShouldFlushBuffer() bool {
	now := time.Now()

	// Flush if buffer is full
	if len(og.orderBuffer) >= og.config.BatchSize {
		return true
	}

	// Flush if enough time has passed and buffer has orders
	if len(og.orderBuffer) > 0 && now.Sub(og.lastBatchTime) >= og.config.BatchInterval {
		return true
	}

	return false
}

// FlushBuffer returns all buffered orders and clears the buffer
func (og *OrderGenerator) FlushBuffer() []*shared.Order {
	if len(og.orderBuffer) == 0 {
		return nil
	}

	orders := make([]*shared.Order, len(og.orderBuffer))
	copy(orders, og.orderBuffer)

	// Clear buffer
	og.orderBuffer = og.orderBuffer[:0]
	og.lastBatchTime = time.Now()

	og.logger.Info("Flushing order batch", "batch_size", len(orders))
	return orders
}

// GetOptimalConfig returns a performance-optimized configuration
func GetOptimalConfig() OrderGeneratorConfig {
	return OrderGeneratorConfig{
		BaseOrderRate:      0.01, // 0.01 orders per second (EXTREMELY slow)
		VolatilityBoost:    1.1,  // Only 10% boost during volatility
		MaxOrdersPerMinute: 2,    // Maximum 2 orders per minute
		BatchSize:          1,    // Send 1 order at a time
		BatchInterval:      60 * time.Second, // Flush every minute
		RandomSeed:         time.Now().UnixNano(),
		UserTypeMix: map[string]float64{
			"retail":       0.9,  // 90% retail traders (slower)
			"algorithmic":  0.08, // 8% algo traders
			"institutional": 0.02, // 2% institutional
		},
	}
}