package domain

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"simulated_exchange/pkg/shared"
)

// TradingService implements shared.TradingService interface
type TradingService struct {
	orderRepo    shared.OrderRepository
	tradeRepo    shared.TradeRepository
	cache        shared.CacheRepository
	eventBus     shared.EventBus
	orderMatcher shared.OrderMatcher
	logger       *slog.Logger
}

// NewTradingService creates a new trading service
func NewTradingService(
	orderRepo shared.OrderRepository,
	tradeRepo shared.TradeRepository,
	cache shared.CacheRepository,
	eventBus shared.EventBus,
	orderMatcher shared.OrderMatcher,
	logger *slog.Logger,
) *TradingService {
	return &TradingService{
		orderRepo:    orderRepo,
		tradeRepo:    tradeRepo,
		cache:        cache,
		eventBus:     eventBus,
		orderMatcher: orderMatcher,
		logger:       logger,
	}
}

// PlaceOrder places a new order in the system
func (s *TradingService) PlaceOrder(ctx context.Context, order *shared.Order) (*shared.Order, error) {
	// Validate order
	if err := s.validateOrder(order); err != nil {
		return nil, err
	}

	// Generate order ID if not provided
	if order.ID == "" {
		order.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now
	order.Status = shared.OrderStatusPending

	s.logger.Info("Placing order",
		"order_id", order.ID,
		"user_id", order.UserID,
		"symbol", order.Symbol,
		"side", order.Side,
		"type", order.Type,
		"quantity", order.Quantity,
		"price", order.Price,
	)

	// Save order to database
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, shared.NewServiceErrorWithCause("trading", "place_order", "failed to save order", err)
	}

	// Try to match the order
	if err := s.processOrderMatching(ctx, order); err != nil {
		s.logger.Warn("Order matching failed", "order_id", order.ID, "error", err)
		// Don't return error here - order is placed but not matched
	}

	// Publish order placed event
	if err := s.eventBus.Publish(ctx, &shared.Event{
		Type:   shared.EventTypeOrderPlaced,
		Source: "trading-api",
		Data: map[string]interface{}{
			"order_id": order.ID,
			"user_id":  order.UserID,
			"symbol":   order.Symbol,
			"side":     order.Side,
			"type":     order.Type,
			"price":    order.Price,
			"quantity": order.Quantity,
		},
	}); err != nil {
		s.logger.Warn("Failed to publish order placed event", "error", err)
	}

	// Update order book cache
	if err := s.updateOrderBookCache(ctx, order.Symbol); err != nil {
		s.logger.Warn("Failed to update order book cache", "error", err)
	}

	return order, nil
}

// CancelOrder cancels an existing order
func (s *TradingService) CancelOrder(ctx context.Context, orderID string) error {
	// Get order from database
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Check if order can be cancelled
	if order.Status == shared.OrderStatusFilled {
		return shared.NewBusinessError(shared.ErrCodeOrderAlreadyFilled, "order is already filled")
	}
	if order.Status == shared.OrderStatusCancelled {
		return shared.NewBusinessError(shared.ErrCodeOrderAlreadyCancelled, "order is already cancelled")
	}

	s.logger.Info("Cancelling order", "order_id", orderID, "user_id", order.UserID)

	// Update order status
	order.Status = shared.OrderStatusCancelled
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return shared.NewServiceErrorWithCause("trading", "cancel_order", "failed to update order", err)
	}

	// Publish order cancelled event
	if err := s.eventBus.Publish(ctx, &shared.Event{
		Type:   shared.EventTypeOrderCancelled,
		Source: "trading-api",
		Data: map[string]interface{}{
			"order_id": orderID,
			"user_id":  order.UserID,
		},
	}); err != nil {
		s.logger.Warn("Failed to publish order cancelled event", "error", err)
	}

	// Update order book cache
	if err := s.updateOrderBookCache(ctx, order.Symbol); err != nil {
		s.logger.Warn("Failed to update order book cache", "error", err)
	}

	return nil
}

// GetOrder retrieves an order by ID
func (s *TradingService) GetOrder(ctx context.Context, orderID string) (*shared.Order, error) {
	return s.orderRepo.GetByID(ctx, orderID)
}

// GetRecentOrders retrieves recent orders with optional limit
func (s *TradingService) GetRecentOrders(ctx context.Context, limit int) ([]*shared.Order, error) {
	// Use the OrderRepository's GetOrdersInTimeRange or create a simple method
	// For now, we'll get active orders as a proxy for recent orders
	orders, err := s.orderRepo.GetActiveOrders(ctx)
	if err != nil {
		return nil, err
	}

	// If we have more orders than the limit, return the most recent ones
	if len(orders) > limit {
		return orders[:limit], nil
	}

	return orders, nil
}

// GetOrderBook retrieves the current order book for a symbol
func (s *TradingService) GetOrderBook(ctx context.Context, symbol string) (*shared.OrderBook, error) {
	// Try to get from cache first
	orderBook, err := s.cache.GetOrderBook(ctx, symbol)
	if err == nil {
		return orderBook, nil
	}

	// If not in cache, build from database
	orders, err := s.orderRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}

	// Separate bids and asks
	var bids, asks []shared.Order
	for _, order := range orders {
		if order.Status == shared.OrderStatusPending || order.Status == shared.OrderStatusPartial {
			if order.Side == shared.OrderSideBuy {
				bids = append(bids, *order)
			} else {
				asks = append(asks, *order)
			}
		}
	}

	orderBook = &shared.OrderBook{
		Symbol:    symbol,
		Bids:      bids,
		Asks:      asks,
		UpdatedAt: time.Now(),
	}

	// Cache the order book
	if err := s.cache.SetOrderBook(ctx, symbol, orderBook); err != nil {
		s.logger.Warn("Failed to cache order book", "error", err)
	}

	return orderBook, nil
}

// GetUserOrders retrieves all orders for a specific user
func (s *TradingService) GetUserOrders(ctx context.Context, userID string) ([]*shared.Order, error) {
	return s.orderRepo.GetByUserID(ctx, userID)
}

// GetTrades retrieves recent trades for a symbol
func (s *TradingService) GetTrades(ctx context.Context, symbol string, limit int) ([]*shared.Trade, error) {
	return s.tradeRepo.GetRecentTrades(ctx, limit)
}

// processOrderMatching attempts to match an order with existing orders
func (s *TradingService) processOrderMatching(ctx context.Context, newOrder *shared.Order) error {
	// Get existing orders for the same symbol
	existingOrders, err := s.orderRepo.GetBySymbol(ctx, newOrder.Symbol)
	if err != nil {
		return err
	}

	// Filter for matching orders (opposite side)
	var matchableOrders []*shared.Order
	for _, order := range existingOrders {
		if order.Side != newOrder.Side && (order.Status == shared.OrderStatusPending || order.Status == shared.OrderStatusPartial) {
			matchableOrders = append(matchableOrders, order)
		}
	}

	if len(matchableOrders) == 0 {
		return nil // No matches possible
	}

	// Find matches
	matches, err := s.orderMatcher.FindMatches(ctx, newOrder, matchableOrders)
	if err != nil {
		return err
	}

	// Execute trades for each match
	for _, match := range matches {
		trade, err := s.orderMatcher.ExecuteTrade(ctx, match)
		if err != nil {
			s.logger.Error("Failed to execute trade", "error", err)
			continue
		}

		// Save trade to database
		if err := s.tradeRepo.Create(ctx, trade); err != nil {
			s.logger.Error("Failed to save trade", "trade_id", trade.ID, "error", err)
			continue
		}

		// Update order quantities and statuses
		if err := s.updateOrdersAfterTrade(ctx, match, trade.Quantity); err != nil {
			s.logger.Error("Failed to update orders after trade", "trade_id", trade.ID, "error", err)
		}

		// Publish trade executed event
		if err := s.eventBus.Publish(ctx, &shared.Event{
			Type:   shared.EventTypeTradeExecuted,
			Source: "trading-api",
			Data: map[string]interface{}{
				"trade_id":      trade.ID,
				"buy_order_id":  trade.BuyOrderID,
				"sell_order_id": trade.SellOrderID,
				"symbol":        trade.Symbol,
				"price":         trade.Price,
				"quantity":      trade.Quantity,
			},
		}); err != nil {
			s.logger.Warn("Failed to publish trade executed event", "error", err)
		}

		s.logger.Info("Trade executed",
			"trade_id", trade.ID,
			"symbol", trade.Symbol,
			"price", trade.Price,
			"quantity", trade.Quantity,
		)
	}

	return nil
}

// updateOrdersAfterTrade updates order quantities and statuses after a trade
func (s *TradingService) updateOrdersAfterTrade(ctx context.Context, match *shared.Match, tradeQuantity float64) error {
	// Update buy order
	buyOrder := match.BuyOrder
	buyOrder.Quantity -= tradeQuantity
	if buyOrder.Quantity <= 0 {
		buyOrder.Status = shared.OrderStatusFilled
	} else {
		buyOrder.Status = shared.OrderStatusPartial
	}
	buyOrder.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, &buyOrder); err != nil {
		return fmt.Errorf("failed to update buy order: %w", err)
	}

	// Update sell order
	sellOrder := match.SellOrder
	sellOrder.Quantity -= tradeQuantity
	if sellOrder.Quantity <= 0 {
		sellOrder.Status = shared.OrderStatusFilled
	} else {
		sellOrder.Status = shared.OrderStatusPartial
	}
	sellOrder.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, &sellOrder); err != nil {
		return fmt.Errorf("failed to update sell order: %w", err)
	}

	return nil
}

// updateOrderBookCache updates the cached order book for a symbol
func (s *TradingService) updateOrderBookCache(ctx context.Context, symbol string) error {
	orderBook, err := s.GetOrderBook(ctx, symbol)
	if err != nil {
		return err
	}

	return s.cache.SetOrderBook(ctx, symbol, orderBook)
}

// validateOrder validates an order before processing
func (s *TradingService) validateOrder(order *shared.Order) error {
	if order.UserID == "" {
		return shared.NewValidationError("user_id", "user ID is required")
	}

	if order.Symbol == "" {
		return shared.NewValidationError("symbol", "symbol is required")
	}

	if order.Side != shared.OrderSideBuy && order.Side != shared.OrderSideSell {
		return shared.NewValidationError("side", "side must be BUY or SELL")
	}

	if order.Type != shared.OrderTypeMarket && order.Type != shared.OrderTypeLimit {
		return shared.NewValidationError("type", "type must be MARKET or LIMIT")
	}

	if order.Quantity <= 0 {
		return shared.NewValidationError("quantity", "quantity must be positive")
	}

	if order.Type == shared.OrderTypeLimit && order.Price <= 0 {
		return shared.NewValidationError("price", "price must be positive for limit orders")
	}

	return nil
}