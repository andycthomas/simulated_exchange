package metrics

import (
	"time"

	"simulated_exchange/internal/types"
)

// OrderEvent represents an order-related event for metrics collection
type OrderEvent struct {
	OrderID   string
	Symbol    string
	Side      types.OrderSide
	Type      types.OrderType
	Quantity  float64
	Price     float64
	Timestamp time.Time
	Latency   time.Duration // Time from order creation to processing
}

// TradeEvent represents a trade-related event for metrics collection
type TradeEvent struct {
	TradeID     string
	Symbol      string
	Quantity    float64
	Price       float64
	Timestamp   time.Time
	Latency     time.Duration // Time from order matching to trade execution
	BuyOrderID  string
	SellOrderID string
}

// MetricsSnapshot represents metrics calculated for a specific time window
type MetricsSnapshot struct {
	WindowStart    time.Time
	WindowEnd      time.Time
	OrderCount     int64
	TradeCount     int64
	TotalVolume    float64
	AvgLatency     time.Duration
	MaxLatency     time.Duration
	MinLatency     time.Duration
	OrdersPerSec   float64
	TradesPerSec   float64
	VolumePerSec   float64
	SymbolMetrics  map[string]SymbolMetrics
}

// SymbolMetrics represents metrics for a specific trading symbol
type SymbolMetrics struct {
	Symbol       string
	OrderCount   int64
	TradeCount   int64
	Volume       float64
	AvgPrice     float64
	HighPrice    float64
	LowPrice     float64
	LastPrice    float64
	AvgLatency   time.Duration
}

// PerformanceAnalysis represents the result of performance analysis
type PerformanceAnalysis struct {
	Timestamp           time.Time
	LatencyTrend        TrendDirection
	ThroughputTrend     TrendDirection
	PredictedThroughput float64
	Bottlenecks         []Bottleneck
	Recommendations     []string
}

// TrendDirection represents the direction of a performance trend
type TrendDirection string

const (
	TrendUp    TrendDirection = "UP"
	TrendDown  TrendDirection = "DOWN"
	TrendFlat  TrendDirection = "FLAT"
)

// Bottleneck represents a detected performance bottleneck
type Bottleneck struct {
	Type        string
	Severity    float64 // 0.0 to 1.0
	Description string
	Component   string
}

// MetricsCollector interface for collecting trading metrics
type MetricsCollector interface {
	RecordOrder(event OrderEvent)
	RecordTrade(event TradeEvent)
	CalculateMetrics(windowDuration time.Duration) MetricsSnapshot
	GetCurrentMetrics() MetricsSnapshot
	Reset()
}

// PerformanceAnalyzer interface for analyzing performance metrics
type PerformanceAnalyzer interface {
	AnalyzeLatency(snapshots []MetricsSnapshot) LatencyAnalysis
	PredictThroughput(snapshots []MetricsSnapshot) ThroughputPrediction
	DetectBottlenecks(snapshot MetricsSnapshot) []Bottleneck
	GenerateRecommendations(analysis PerformanceAnalysis) []string
}

// LatencyAnalysis represents latency analysis results
type LatencyAnalysis struct {
	Trend            TrendDirection
	CurrentLatency   time.Duration
	PredictedLatency time.Duration
	PercentileP50    time.Duration
	PercentileP95    time.Duration
	PercentileP99    time.Duration
}

// ThroughputPrediction represents throughput prediction results
type ThroughputPrediction struct {
	Trend                TrendDirection
	CurrentThroughput    float64
	PredictedThroughput  float64
	MaxObservedThroughput float64
	ConfidenceLevel      float64
}

// MetricsService interface for orchestrating metrics collection and analysis
type MetricsService interface {
	Start() error
	Stop() error
	GetRealTimeMetrics() MetricsSnapshot
	GetPerformanceAnalysis() PerformanceAnalysis
	IsHealthy() bool
}