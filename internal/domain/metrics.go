package domain

import (
	"errors"
)

type PerformanceMetrics struct {
	OrdersPerSecond     float64 `json:"orders_per_second"`
	AverageLatencyMs    float64 `json:"average_latency_ms"`
	SystemLoadPercent   float64 `json:"system_load_percent"`
	OptimizationScore   float64 `json:"optimization_score"`
}

func NewPerformanceMetrics(ordersPerSecond, averageLatencyMs, systemLoadPercent, optimizationScore float64) (*PerformanceMetrics, error) {
	if ordersPerSecond < 0 {
		return nil, errors.New("orders per second cannot be negative")
	}
	if averageLatencyMs < 0 {
		return nil, errors.New("average latency cannot be negative")
	}
	if systemLoadPercent < 0 || systemLoadPercent > 100 {
		return nil, errors.New("system load percent must be between 0 and 100")
	}
	if optimizationScore < 0 || optimizationScore > 100 {
		return nil, errors.New("optimization score must be between 0 and 100")
	}

	return &PerformanceMetrics{
		OrdersPerSecond:   ordersPerSecond,
		AverageLatencyMs:  averageLatencyMs,
		SystemLoadPercent: systemLoadPercent,
		OptimizationScore: optimizationScore,
	}, nil
}

func (pm *PerformanceMetrics) IsValid() error {
	if pm.OrdersPerSecond < 0 {
		return errors.New("orders per second cannot be negative")
	}
	if pm.AverageLatencyMs < 0 {
		return errors.New("average latency cannot be negative")
	}
	if pm.SystemLoadPercent < 0 || pm.SystemLoadPercent > 100 {
		return errors.New("system load percent must be between 0 and 100")
	}
	if pm.OptimizationScore < 0 || pm.OptimizationScore > 100 {
		return errors.New("optimization score must be between 0 and 100")
	}
	return nil
}

func (pm *PerformanceMetrics) IsHealthy() bool {
	return pm.SystemLoadPercent < 80 && pm.AverageLatencyMs < 100 && pm.OptimizationScore > 70
}

func (pm *PerformanceMetrics) CalculateEfficiency() float64 {
	if pm.AverageLatencyMs == 0 {
		return 0
	}
	return (pm.OrdersPerSecond * pm.OptimizationScore) / (pm.AverageLatencyMs * pm.SystemLoadPercent / 100)
}