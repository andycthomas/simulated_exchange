package domain

import (
	"context"
	"time"
)

type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	GetByUserID(ctx context.Context, userID string) ([]*Order, error)
	GetBySymbol(ctx context.Context, symbol string) ([]*Order, error)
	GetByStatus(ctx context.Context, status OrderStatus) ([]*Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id string) error
	GetActiveOrders(ctx context.Context) ([]*Order, error)
	GetOrdersInTimeRange(ctx context.Context, start, end time.Time) ([]*Order, error)
}

type TradeRepository interface {
	Save(ctx context.Context, trade *Trade) error
	GetByID(ctx context.Context, id string) (*Trade, error)
	GetByOrderID(ctx context.Context, orderID string) ([]*Trade, error)
	GetBySymbol(ctx context.Context, symbol string) ([]*Trade, error)
	GetTradesInTimeRange(ctx context.Context, start, end time.Time) ([]*Trade, error)
	GetRecentTrades(ctx context.Context, limit int) ([]*Trade, error)
	Delete(ctx context.Context, id string) error
}

type MetricsCalculator interface {
	CalculateOrdersPerSecond(ctx context.Context, duration time.Duration) (float64, error)
	CalculateAverageLatency(ctx context.Context, duration time.Duration) (float64, error)
	CalculateSystemLoad(ctx context.Context) (float64, error)
	CalculateOptimizationScore(ctx context.Context, duration time.Duration) (float64, error)
	GetPerformanceMetrics(ctx context.Context, duration time.Duration) (*PerformanceMetrics, error)
}

type OrderMatcher interface {
	MatchOrders(ctx context.Context, newOrder *Order) ([]*Trade, error)
	GetOrderBook(ctx context.Context, symbol string) (*OrderBook, error)
}

type OrderBook struct {
	Symbol   string             `json:"symbol"`
	BuyOrders []*Order          `json:"buy_orders"`
	SellOrders []*Order         `json:"sell_orders"`
	LastUpdated time.Time       `json:"last_updated"`
}