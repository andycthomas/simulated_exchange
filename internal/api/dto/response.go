package dto

import "time"

// APIResponse provides a consistent structure for all API responses
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents error information in API responses
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// OrderResponse represents order information in API responses
type OrderResponse struct {
	ID       string  `json:"id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`
	Type     string  `json:"type"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Status   string  `json:"status"`
}

// OrderBookResponse represents order book information
type OrderBookResponse struct {
	Symbol string              `json:"symbol"`
	Bids   []OrderBookEntryDTO `json:"bids"`
	Asks   []OrderBookEntryDTO `json:"asks"`
}

// OrderBookEntryDTO represents a single order book entry
type OrderBookEntryDTO struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// MetricsResponse represents real-time metrics
type MetricsResponse struct {
	OrderCount     int64                      `json:"order_count"`
	TradeCount     int64                      `json:"trade_count"`
	TotalVolume    float64                    `json:"total_volume"`
	AvgLatency     string                     `json:"avg_latency"`
	OrdersPerSec   float64                    `json:"orders_per_sec"`
	TradesPerSec   float64                    `json:"trades_per_sec"`
	SymbolMetrics  map[string]SymbolMetricsDTO `json:"symbol_metrics"`
	Analysis       *PerformanceAnalysisDTO    `json:"analysis,omitempty"`
}

// SymbolMetricsDTO represents symbol-specific metrics
type SymbolMetricsDTO struct {
	OrderCount int64   `json:"order_count"`
	TradeCount int64   `json:"trade_count"`
	Volume     float64 `json:"volume"`
	AvgPrice   float64 `json:"avg_price"`
}

// PerformanceAnalysisDTO represents performance analysis data
type PerformanceAnalysisDTO struct {
	Timestamp       time.Time        `json:"timestamp"`
	TrendDirection  string           `json:"trend_direction"`
	Bottlenecks     []BottleneckDTO  `json:"bottlenecks"`
	Recommendations []string         `json:"recommendations"`
}

// BottleneckDTO represents a performance bottleneck
type BottleneckDTO struct {
	Type        string  `json:"type"`
	Severity    float64 `json:"severity"`
	Description string  `json:"description"`
}

// HealthResponse represents system health status
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Version   string            `json:"version"`
}

// PlaceOrderResponse represents the response after placing an order
type PlaceOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// CancelOrderResponse represents the response after canceling an order
type CancelOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}