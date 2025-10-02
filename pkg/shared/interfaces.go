package shared

import (
	"context"
	"time"
)

// Repository Interfaces (Dependency Inversion Principle)

// OrderRepository defines the interface for order persistence
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	GetByUserID(ctx context.Context, userID string) ([]*Order, error)
	GetBySymbol(ctx context.Context, symbol string) ([]*Order, error)
	GetByStatus(ctx context.Context, status OrderStatus) ([]*Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id string) error
	GetActiveOrders(ctx context.Context) ([]*Order, error)
	GetOrdersInTimeRange(ctx context.Context, start, end time.Time) ([]*Order, error)
}

// TradeRepository defines the interface for trade persistence
type TradeRepository interface {
	Create(ctx context.Context, trade *Trade) error
	GetByID(ctx context.Context, id string) (*Trade, error)
	GetByOrderID(ctx context.Context, orderID string) ([]*Trade, error)
	GetBySymbol(ctx context.Context, symbol string) ([]*Trade, error)
	GetTradesInTimeRange(ctx context.Context, start, end time.Time) ([]*Trade, error)
	GetRecentTrades(ctx context.Context, limit int) ([]*Trade, error)
}

// UserRepository defines the interface for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

// Cache Interface (for Redis integration)

// CacheRepository defines the interface for caching operations
type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetOrderBook(ctx context.Context, symbol string, orderBook *OrderBook) error
	GetOrderBook(ctx context.Context, symbol string) (*OrderBook, error)
	SetMarketData(ctx context.Context, symbol string, data *MarketData) error
	GetMarketData(ctx context.Context, symbol string) (*MarketData, error)
}

// Event Bus Interface (for inter-service communication)

// EventBus defines the interface for event publishing and subscribing
type EventBus interface {
	Publish(ctx context.Context, event *Event) error
	Subscribe(ctx context.Context, eventType EventType, handler EventHandler) error
	Unsubscribe(ctx context.Context, eventType EventType) error
	Close() error
}

// EventHandler defines the signature for event handlers
type EventHandler func(ctx context.Context, event *Event) error

// Service Interfaces (following Interface Segregation Principle)

// TradingService defines the interface for trading operations
type TradingService interface {
	PlaceOrder(ctx context.Context, order *Order) (*Order, error)
	CancelOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	GetRecentOrders(ctx context.Context, limit int) ([]*Order, error)
	GetOrderBook(ctx context.Context, symbol string) (*OrderBook, error)
	GetUserOrders(ctx context.Context, userID string) ([]*Order, error)
	GetTrades(ctx context.Context, symbol string, limit int) ([]*Trade, error)
}

// PriceService defines the interface for price and market data operations
type PriceService interface {
	GetCurrentPrice(ctx context.Context, symbol string) (float64, error)
	GetMarketData(ctx context.Context, symbol string) (*MarketData, error)
	GetPriceHistory(ctx context.Context, symbol string, from, to time.Time) ([]PriceUpdate, error)
	UpdatePrice(ctx context.Context, update *PriceUpdate) error
}

// MetricsService defines the interface for metrics operations
type MetricsService interface {
	RecordOrder(ctx context.Context, order *Order) error
	RecordTrade(ctx context.Context, trade *Trade) error
	GetMetrics(ctx context.Context, duration time.Duration) (*Metrics, error)
	GetSystemHealth(ctx context.Context) (*ServiceInfo, error)
}

// OrderMatcher defines the interface for order matching logic
type OrderMatcher interface {
	FindMatches(ctx context.Context, newOrder *Order, existingOrders []*Order) ([]*Match, error)
	ExecuteTrade(ctx context.Context, match *Match) (*Trade, error)
}

// SimulatorService defines the interface for market simulation
type SimulatorService interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool
	GetStatus(ctx context.Context) (*ServiceInfo, error)
	InjectVolatility(ctx context.Context, pattern string, intensity float64) error
}

// NotificationService defines the interface for notifications
type NotificationService interface {
	NotifyOrderUpdate(ctx context.Context, order *Order) error
	NotifyTradeExecution(ctx context.Context, trade *Trade) error
	NotifyPriceUpdate(ctx context.Context, update *PriceUpdate) error
}

// Configuration Interface

// Config defines the interface for service configuration
type Config interface {
	GetString(key string) string
	GetInt(key string) int
	GetFloat64(key string) float64
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	GetStringSlice(key string) []string
}

// Health Check Interface

// HealthChecker defines the interface for health checking
type HealthChecker interface {
	Check(ctx context.Context) error
	Name() string
}

// Database Connection Interface

// DBConnection defines the interface for database operations
type DBConnection interface {
	Ping(ctx context.Context) error
	Begin(ctx context.Context) (Transaction, error)
	Close() error
	Stats() interface{}
}

// Transaction defines the interface for database transactions
type Transaction interface {
	Commit() error
	Rollback() error
	Exec(ctx context.Context, query string, args ...interface{}) error
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) Row
}

// Rows defines the interface for query results
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

// Row defines the interface for single row results
type Row interface {
	Scan(dest ...interface{}) error
}