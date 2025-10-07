package domain

import (
	"context"
	"log/slog"
	"math"
	"math/rand"
	"time"

	"simulated_exchange/pkg/shared"
)

// OrderGeneratorConfig holds configuration for the order generator
type OrderGeneratorConfig struct {
	BaseOrderRate   float64            `json:"base_order_rate"`   // Base orders per second (reduced for performance)
	VolatilityBoost float64            `json:"volatility_boost"`  // Multiplier during volatile periods
	UserTypeMix     map[string]float64 `json:"user_type_mix"`     // Percentage of each user type
	RandomSeed      int64              `json:"random_seed"`       // Random seed for reproducibility
	MaxOrdersPerMinute int             `json:"max_orders_per_minute"` // Maximum orders per minute (rate limit)
	BatchSize       int                `json:"batch_size"`        // Number of orders to batch together
	BatchInterval   time.Duration      `json:"batch_interval"`    // Interval between batches
}

// OrderGenerator generates realistic trading orders based on market conditions
type OrderGenerator struct {
	config         OrderGeneratorConfig
	logger         *slog.Logger
	random         *rand.Rand
	currentRates   map[string]float64 // Current order generation rates per symbol
	volatilityMode bool               // Whether we're in high volatility mode
	symbols        []string           // Available trading symbols
	orderBuffer    []*shared.Order    // Buffer for batch processing
	lastBatchTime  time.Time          // Last time batch was sent
	orderCount     int                // Orders generated in current minute
	minuteStart    time.Time          // Start of current minute for rate limiting
}

// NewOrderGenerator creates a new order generator
func NewOrderGenerator(config OrderGeneratorConfig, logger *slog.Logger) *OrderGenerator {
	return &OrderGenerator{
		config:       config,
		logger:       logger,
		random:       rand.New(rand.NewSource(config.RandomSeed)),
		currentRates: make(map[string]float64),
		symbols:      []string{"BTC", "ETH", "ADA", "DOT", "SOL", "MATIC"},
	}
}

// GenerateOrder creates a new order based on current market conditions
func (og *OrderGenerator) GenerateOrder(ctx context.Context, userType string, symbol string, currentPrice float64) (*shared.Order, error) {
	order := &shared.Order{
		ID:        og.generateOrderID(),
		UserID:    og.generateUserID(),
		Symbol:    symbol,
		Type:      og.determineOrderType(userType),
		Side:      og.determineOrderSide(userType, symbol),
		Quantity:  og.generateQuantity(userType, currentPrice),
		Status:    shared.OrderStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set price based on order type
	switch order.Type {
	case shared.OrderTypeMarket:
		order.Price = 0 // Market orders don't have a price
	case shared.OrderTypeLimit:
		order.Price = og.generateLimitPrice(currentPrice, order.Side, userType)
	case shared.OrderTypeStopLoss:
		order.Price = og.generateStopPrice(currentPrice, order.Side, userType)
	}

	og.logger.Debug("Generated order",
		"order_id", order.ID,
		"user_type", userType,
		"symbol", symbol,
		"type", order.Type,
		"side", order.Side,
		"quantity", order.Quantity,
		"price", order.Price,
	)

	return order, nil
}

// GetOrderRate returns the current order generation rate for a symbol
func (og *OrderGenerator) GetOrderRate(symbol string) float64 {
	rate, exists := og.currentRates[symbol]
	if !exists {
		rate = og.config.BaseOrderRate
		og.currentRates[symbol] = rate
	}

	if og.volatilityMode {
		rate *= og.config.VolatilityBoost
	}

	return rate
}

// SetVolatilityMode enables or disables high volatility order generation
func (og *OrderGenerator) SetVolatilityMode(enabled bool) {
	og.volatilityMode = enabled
	og.logger.Info("Volatility mode changed", "enabled", enabled)
}

// AdjustRateForSymbol adjusts order generation rate for a specific symbol
func (og *OrderGenerator) AdjustRateForSymbol(symbol string, multiplier float64) {
	baseRate := og.config.BaseOrderRate
	og.currentRates[symbol] = baseRate * multiplier

	og.logger.Debug("Adjusted order rate",
		"symbol", symbol,
		"multiplier", multiplier,
		"new_rate", og.currentRates[symbol],
	)
}

// GetSupportedSymbols returns the list of supported trading symbols
func (og *OrderGenerator) GetSupportedSymbols() []string {
	return og.symbols
}

// Private helper methods

func (og *OrderGenerator) generateOrderID() string {
	return "ORD-" + og.generateRandomString(12)
}

func (og *OrderGenerator) generateUserID() string {
	// Use ACTUAL database user UUIDs from init-db.sql
	userUUIDs := []string{
		"af2d2361-7772-447b-b2ec-bbf2f6c77be4", // admin@trading.local
		"6169d80e-a6af-45ae-a5cf-2f4c39d0e84e", // trader1@trading.local
		"a8705dd2-44f6-4b77-85c3-28c1816d42e5", // trader2@trading.local
	}
	return userUUIDs[og.random.Intn(len(userUUIDs))]
}

func (og *OrderGenerator) generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[og.random.Intn(len(charset))]
	}
	return string(result)
}

func (og *OrderGenerator) determineOrderType(userType string) shared.OrderType {
	switch userType {
	case "conservative":
		// Conservative users prefer limit orders
		if og.random.Float64() < 0.7 {
			return shared.OrderTypeLimit
		}
		return shared.OrderTypeMarket
	case "aggressive":
		// Aggressive users use more market orders
		if og.random.Float64() < 0.6 {
			return shared.OrderTypeMarket
		}
		return shared.OrderTypeLimit
	case "momentum":
		// Momentum traders use stops and limits
		rand := og.random.Float64()
		if rand < 0.4 {
			return shared.OrderTypeLimit
		} else if rand < 0.7 {
			return shared.OrderTypeMarket
		}
		return shared.OrderTypeStopLoss
	default:
		return shared.OrderTypeLimit
	}
}

func (og *OrderGenerator) determineOrderSide(userType string, symbol string) shared.OrderSide {
	// For simplicity, use 50/50 buy/sell ratio
	// In reality, this would be based on market sentiment, trends, etc.
	if og.random.Float64() < 0.5 {
		return shared.OrderSideBuy
	}
	return shared.OrderSideSell
}

func (og *OrderGenerator) generateQuantity(userType string, currentPrice float64) float64 {
	var baseQuantity float64

	switch userType {
	case "conservative":
		// Smaller, more conservative quantities
		baseQuantity = 0.1 + og.random.Float64()*0.5 // 0.1 to 0.6
	case "aggressive":
		// Larger quantities
		baseQuantity = 0.5 + og.random.Float64()*2.0 // 0.5 to 2.5
	case "momentum":
		// Variable quantities based on momentum
		baseQuantity = 0.2 + og.random.Float64()*1.5 // 0.2 to 1.7
	default:
		baseQuantity = 0.3 + og.random.Float64()*1.0 // 0.3 to 1.3
	}

	// Adjust quantity based on price (higher price = smaller quantity)
	priceAdjustment := math.Max(0.1, 1000.0/currentPrice)
	return baseQuantity * priceAdjustment
}

func (og *OrderGenerator) generateLimitPrice(currentPrice float64, side shared.OrderSide, userType string) float64 {
	var priceOffset float64

	switch userType {
	case "conservative":
		// Conservative users place orders further from market price
		priceOffset = 0.02 + og.random.Float64()*0.03 // 2-5% from market
	case "aggressive":
		// Aggressive users place orders closer to market price
		priceOffset = 0.005 + og.random.Float64()*0.015 // 0.5-2% from market
	default:
		priceOffset = 0.01 + og.random.Float64()*0.02 // 1-3% from market
	}

	if side == shared.OrderSideBuy {
		// Buy orders below market price
		return currentPrice * (1 - priceOffset)
	}
	// Sell orders above market price
	return currentPrice * (1 + priceOffset)
}

func (og *OrderGenerator) generateStopPrice(currentPrice float64, side shared.OrderSide, userType string) float64 {
	// Stop losses are typically 5-15% away from current price
	stopDistance := 0.05 + og.random.Float64()*0.10 // 5-15%

	if side == shared.OrderSideBuy {
		// Buy stop above current price (stop buy)
		return currentPrice * (1 + stopDistance)
	}
	// Sell stop below current price (stop loss)
	return currentPrice * (1 - stopDistance)
}