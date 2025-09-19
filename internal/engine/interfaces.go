package engine

import (
	"time"
	"simulated_exchange/internal/types"
)

type OrderProcessor interface {
	PlaceOrder(order types.Order) error
	CancelOrder(id string) error
	GetOrderBook(symbol string) (types.OrderBook, error)
}

// MetricsRecorder interface for recording performance metrics
type MetricsRecorder interface {
	RecordOrderEvent(orderID, symbol string, side types.OrderSide, orderType types.OrderType, quantity, price float64, latency time.Duration)
	RecordTradeEvent(tradeID, symbol string, quantity, price float64, latency time.Duration, buyOrderID, sellOrderID string)
}

type TradeExecutor interface {
	ExecuteTrade(buyOrder types.Order, sellOrder types.Order, quantity float64, price float64) (types.Trade, error)
}

type OrderMatcher interface {
	FindMatches(newOrder types.Order, existingOrders []types.Order) []types.Match
}

type OrderRepository interface {
	Save(order types.Order) error
	GetByID(id string) (types.Order, error)
	GetBySymbol(symbol string) ([]types.Order, error)
	Delete(id string) error
	GetAll() ([]types.Order, error)
}

type TradeRepository interface {
	Save(trade types.Trade) error
	GetByID(id string) (types.Trade, error)
	GetBySymbol(symbol string) ([]types.Trade, error)
	GetAll() ([]types.Trade, error)
}