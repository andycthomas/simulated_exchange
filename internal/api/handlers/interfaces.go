package handlers

import (
	"github.com/gin-gonic/gin"
)

// OrderHandler interface defines HTTP endpoints for order operations
type OrderHandler interface {
	PlaceOrder(c *gin.Context)
	GetOrder(c *gin.Context)
	CancelOrder(c *gin.Context)
}

// MetricsHandler interface defines HTTP endpoints for metrics and health
type MetricsHandler interface {
	GetMetrics(c *gin.Context)
	GetHealth(c *gin.Context)
}

// HandlerDependencies interface for dependency injection
type HandlerDependencies interface {
	GetOrderService() OrderService
	GetMetricsService() MetricsService
}

// OrderService interface for business logic
type OrderService interface {
	PlaceOrder(orderID, symbol string, side, orderType string, quantity, price float64) error
	GetOrder(orderID string) (Order, error)
	CancelOrder(orderID string) error
	GetOrderBook(symbol string) (OrderBook, error)
}

// MetricsService interface for metrics operations
type MetricsService interface {
	GetRealTimeMetrics() MetricsSnapshot
	GetPerformanceAnalysis() PerformanceAnalysis
	IsHealthy() bool
}

// Domain models for handler dependencies
type Order struct {
	ID       string
	Symbol   string
	Side     string
	Type     string
	Quantity float64
	Price    float64
	Status   string
}

type OrderBook struct {
	Symbol string
	Bids   []OrderBookEntry
	Asks   []OrderBookEntry
}

type OrderBookEntry struct {
	Price    float64
	Quantity float64
}

type MetricsSnapshot struct {
	OrderCount     int64
	TradeCount     int64
	TotalVolume    float64
	AvgLatency     string
	OrdersPerSec   float64
	TradesPerSec   float64
	SymbolMetrics  map[string]SymbolMetrics
}

type SymbolMetrics struct {
	OrderCount int64
	TradeCount int64
	Volume     float64
	AvgPrice   float64
}

type PerformanceAnalysis struct {
	Timestamp       string
	TrendDirection  string
	Bottlenecks     []Bottleneck
	Recommendations []string
}

type Bottleneck struct {
	Type        string
	Severity    float64
	Description string
}