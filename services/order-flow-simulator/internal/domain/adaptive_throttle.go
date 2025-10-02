package domain

import (
	"time"
	"log/slog"
)

// AdaptiveThrottle monitors system performance and adjusts order generation rate
type AdaptiveThrottle struct {
	logger              *slog.Logger
	baseRate           float64
	currentRate        float64
	errorCount         int
	successCount       int
	lastAdjustment     time.Time
	maxRate            float64
	minRate            float64
	adjustmentInterval time.Duration
}

// NewAdaptiveThrottle creates a new adaptive throttle controller
func NewAdaptiveThrottle(baseRate float64, logger *slog.Logger) *AdaptiveThrottle {
	return &AdaptiveThrottle{
		logger:             logger,
		baseRate:           baseRate,
		currentRate:        baseRate,
		maxRate:            baseRate * 2.0,   // Never exceed 2x base rate
		minRate:            baseRate * 0.1,   // Never go below 10% of base rate
		adjustmentInterval: 60 * time.Second, // Adjust every minute
		lastAdjustment:     time.Now(),
	}
}

// RecordSuccess records a successful order placement
func (at *AdaptiveThrottle) RecordSuccess() {
	at.successCount++
}

// RecordError records a failed order placement
func (at *AdaptiveThrottle) RecordError() {
	at.errorCount++
	at.logger.Warn("Order placement error recorded",
		"error_count", at.errorCount,
		"success_count", at.successCount)
}

// ShouldThrottle determines if order generation should be throttled
func (at *AdaptiveThrottle) ShouldThrottle() bool {
	now := time.Now()

	// Only adjust every minute
	if now.Sub(at.lastAdjustment) < at.adjustmentInterval {
		return false
	}

	totalRequests := at.errorCount + at.successCount
	if totalRequests == 0 {
		return false
	}

	errorRate := float64(at.errorCount) / float64(totalRequests)

	// If error rate > 20%, throttle aggressively
	if errorRate > 0.2 {
		at.currentRate = at.currentRate * 0.5 // Reduce by 50%
		if at.currentRate < at.minRate {
			at.currentRate = at.minRate
		}
		at.logger.Warn("High error rate detected, throttling order generation",
			"error_rate", errorRate,
			"new_rate", at.currentRate)
	} else if errorRate < 0.05 && at.currentRate < at.maxRate {
		// If error rate < 5%, gradually increase
		at.currentRate = at.currentRate * 1.1 // Increase by 10%
		if at.currentRate > at.maxRate {
			at.currentRate = at.maxRate
		}
		at.logger.Info("Low error rate, increasing order generation rate",
			"error_rate", errorRate,
			"new_rate", at.currentRate)
	}

	// Reset counters
	at.errorCount = 0
	at.successCount = 0
	at.lastAdjustment = now

	return false
}

// GetCurrentRate returns the current throttled rate
func (at *AdaptiveThrottle) GetCurrentRate() float64 {
	return at.currentRate
}

// GetThrottleDelay calculates delay needed for current rate
func (at *AdaptiveThrottle) GetThrottleDelay() time.Duration {
	if at.currentRate <= 0 {
		return time.Minute // Fallback to 1 minute delay
	}
	return time.Duration(float64(time.Second) / at.currentRate)
}

// IsHealthy returns true if the system appears healthy for order generation
func (at *AdaptiveThrottle) IsHealthy() bool {
	totalRequests := at.errorCount + at.successCount
	if totalRequests < 5 {
		return true // Not enough data
	}

	errorRate := float64(at.errorCount) / float64(totalRequests)
	return errorRate < 0.3 // Healthy if error rate < 30%
}