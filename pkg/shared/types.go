package shared

import (
	"time"
)

// Core Domain Types (Shared across all services)

type OrderSide string
type OrderType string
type OrderStatus string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

const (
	OrderTypeMarket   OrderType = "MARKET"
	OrderTypeLimit    OrderType = "LIMIT"
	OrderTypeStopLoss OrderType = "STOP_LOSS"
)

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPartial   OrderStatus = "PARTIAL"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

// Order represents a trading order
type Order struct {
	ID        string      `json:"id" db:"id"`
	UserID    string      `json:"user_id" db:"user_id"`
	Symbol    string      `json:"symbol" db:"symbol"`
	Side      OrderSide   `json:"side" db:"side"`
	Type      OrderType   `json:"type" db:"type"`
	Price     float64     `json:"price" db:"price"`
	Quantity  float64     `json:"quantity" db:"quantity"`
	Status    OrderStatus `json:"status" db:"status"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// Trade represents an executed trade
type Trade struct {
	ID          string    `json:"id" db:"id"`
	BuyOrderID  string    `json:"buy_order_id" db:"buy_order_id"`
	SellOrderID string    `json:"sell_order_id" db:"sell_order_id"`
	Symbol      string    `json:"symbol" db:"symbol"`
	Price       float64   `json:"price" db:"price"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// OrderBook represents the order book for a symbol
type OrderBook struct {
	Symbol    string    `json:"symbol"`
	Bids      []Order   `json:"bids"`
	Asks      []Order   `json:"asks"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PriceUpdate represents a price update for a symbol
type PriceUpdate struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// MarketData represents market information
type MarketData struct {
	Symbol           string    `json:"symbol"`
	CurrentPrice     float64   `json:"current_price"`
	PreviousPrice    float64   `json:"previous_price"`
	DailyHigh        float64   `json:"daily_high"`
	DailyLow         float64   `json:"daily_low"`
	DailyVolume      float64   `json:"daily_volume"`
	PriceChange      float64   `json:"price_change"`
	PriceChangePerc  float64   `json:"price_change_percent"`
	Timestamp        time.Time `json:"timestamp"`
}

// User represents a trading user
type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	IsActive     bool      `json:"is_active" db:"is_active"`
}

// Match represents a matching pair of orders
type Match struct {
	BuyOrder  Order   `json:"buy_order"`
	SellOrder Order   `json:"sell_order"`
	Quantity  float64 `json:"quantity"`
	Price     float64 `json:"price"`
}

// Metrics represents system performance metrics
type Metrics struct {
	Timestamp      time.Time            `json:"timestamp"`
	OrderCount     int64                `json:"order_count"`
	TradeCount     int64                `json:"trade_count"`
	TotalVolume    float64              `json:"total_volume"`
	OrdersPerSec   float64              `json:"orders_per_sec"`
	TradesPerSec   float64              `json:"trades_per_sec"`
	AvgLatency     time.Duration        `json:"avg_latency"`
	SymbolMetrics  map[string]Metrics   `json:"symbol_metrics,omitempty"`
}

// Event types for inter-service communication
type EventType string

const (
	EventTypeOrderPlaced    EventType = "order.placed"
	EventTypeOrderCancelled EventType = "order.cancelled"
	EventTypeTradeExecuted  EventType = "trade.executed"
	EventTypePriceUpdate    EventType = "price.updated"
	EventTypeMarketData     EventType = "market.data"
)

// Event represents an event in the system
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// ServiceInfo represents service health and information
type ServiceInfo struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    time.Duration     `json:"uptime"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}