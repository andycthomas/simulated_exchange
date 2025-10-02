package domain

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"
	"simulated_exchange/pkg/shared"
)

// OrderMatcher implements shared.OrderMatcher interface
type OrderMatcher struct {
	logger *slog.Logger
}

// NewOrderMatcher creates a new order matcher
func NewOrderMatcher(logger *slog.Logger) *OrderMatcher {
	return &OrderMatcher{
		logger: logger,
	}
}

// FindMatches finds matching orders for a new order using price-time priority
func (m *OrderMatcher) FindMatches(ctx context.Context, newOrder *shared.Order, existingOrders []*shared.Order) ([]*shared.Match, error) {
	var matches []*shared.Match
	var candidates []*shared.Order

	// Filter matching candidates (opposite side, same symbol)
	for _, existingOrder := range existingOrders {
		if existingOrder.Symbol != newOrder.Symbol {
			continue
		}
		if existingOrder.Side == newOrder.Side {
			continue
		}
		if existingOrder.Status != shared.OrderStatusPending && existingOrder.Status != shared.OrderStatusPartial {
			continue
		}
		if m.canMatch(newOrder, existingOrder) {
			candidates = append(candidates, existingOrder)
		}
	}

	if len(candidates) == 0 {
		return matches, nil
	}

	// Sort candidates by price-time priority
	m.sortByPriceTimePriority(candidates, newOrder.Side)

	// Match orders
	remainingQuantity := newOrder.Quantity
	for _, candidate := range candidates {
		if remainingQuantity <= 0 {
			break
		}

		matchQuantity := min(remainingQuantity, candidate.Quantity)
		matchPrice := m.determineMatchPrice(newOrder, candidate)

		var buyOrder, sellOrder shared.Order
		if newOrder.Side == shared.OrderSideBuy {
			buyOrder = *newOrder
			sellOrder = *candidate
		} else {
			buyOrder = *candidate
			sellOrder = *newOrder
		}

		match := &shared.Match{
			BuyOrder:  buyOrder,
			SellOrder: sellOrder,
			Quantity:  matchQuantity,
			Price:     matchPrice,
		}

		matches = append(matches, match)
		remainingQuantity -= matchQuantity

		m.logger.Debug("Found order match",
			"buy_order_id", buyOrder.ID,
			"sell_order_id", sellOrder.ID,
			"symbol", newOrder.Symbol,
			"quantity", matchQuantity,
			"price", matchPrice,
		)
	}

	return matches, nil
}

// ExecuteTrade executes a trade from a match
func (m *OrderMatcher) ExecuteTrade(ctx context.Context, match *shared.Match) (*shared.Trade, error) {
	trade := &shared.Trade{
		ID:          uuid.New().String(),
		BuyOrderID:  match.BuyOrder.ID,
		SellOrderID: match.SellOrder.ID,
		Symbol:      match.BuyOrder.Symbol,
		Price:       match.Price,
		Quantity:    match.Quantity,
		CreatedAt:   time.Now(),
	}

	m.logger.Info("Executing trade",
		"trade_id", trade.ID,
		"buy_order_id", trade.BuyOrderID,
		"sell_order_id", trade.SellOrderID,
		"symbol", trade.Symbol,
		"price", trade.Price,
		"quantity", trade.Quantity,
	)

	return trade, nil
}

// canMatch determines if two orders can be matched
func (m *OrderMatcher) canMatch(newOrder *shared.Order, existingOrder *shared.Order) bool {
	// Market orders can always match
	if newOrder.Type == shared.OrderTypeMarket || existingOrder.Type == shared.OrderTypeMarket {
		return true
	}

	// For limit orders, check price compatibility
	if newOrder.Side == shared.OrderSideBuy {
		// Buy order price must be >= sell order price
		return newOrder.Price >= existingOrder.Price
	} else {
		// Sell order price must be <= buy order price
		return newOrder.Price <= existingOrder.Price
	}
}

// determineMatchPrice determines the execution price for a match
func (m *OrderMatcher) determineMatchPrice(newOrder *shared.Order, existingOrder *shared.Order) float64 {
	// If existing order is market order, use new order price
	if existingOrder.Type == shared.OrderTypeMarket {
		if newOrder.Type == shared.OrderTypeMarket {
			// Both market orders - use a default price (this shouldn't happen in practice)
			return 0
		}
		return newOrder.Price
	}

	// If new order is market order, use existing order price
	if newOrder.Type == shared.OrderTypeMarket {
		return existingOrder.Price
	}

	// Both are limit orders - use existing order price (price-time priority)
	return existingOrder.Price
}

// sortByPriceTimePriority sorts orders by price-time priority
func (m *OrderMatcher) sortByPriceTimePriority(orders []*shared.Order, newOrderSide shared.OrderSide) {
	sort.Slice(orders, func(i, j int) bool {
		orderI, orderJ := orders[i], orders[j]

		// For buy orders matching against sell orders: lowest sell price first
		// For sell orders matching against buy orders: highest buy price first
		if newOrderSide == shared.OrderSideBuy {
			// Matching against sell orders - prefer lowest price
			if orderI.Price != orderJ.Price {
				return orderI.Price < orderJ.Price
			}
		} else {
			// Matching against buy orders - prefer highest price
			if orderI.Price != orderJ.Price {
				return orderI.Price > orderJ.Price
			}
		}

		// If prices are equal, use time priority (earlier orders first)
		return orderI.CreatedAt.Before(orderJ.CreatedAt)
	})
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}