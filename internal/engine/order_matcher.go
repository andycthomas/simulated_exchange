package engine

import (
	"sort"

	"simulated_exchange/internal/types"
)

type PriceTimeOrderMatcher struct{}

func NewPriceTimeOrderMatcher() *PriceTimeOrderMatcher {
	return &PriceTimeOrderMatcher{}
}

func (m *PriceTimeOrderMatcher) FindMatches(newOrder types.Order, existingOrders []types.Order) []types.Match {
	var matches []types.Match
	var candidates []types.Order

	for _, existingOrder := range existingOrders {
		if existingOrder.Symbol != newOrder.Symbol {
			continue
		}
		if existingOrder.Side == newOrder.Side {
			continue
		}
		if m.canMatch(newOrder, existingOrder) {
			candidates = append(candidates, existingOrder)
		}
	}

	m.sortByPriceTimePriority(candidates, newOrder.Side)

	remainingQuantity := newOrder.Quantity
	for _, candidate := range candidates {
		if remainingQuantity <= 0 {
			break
		}

		matchQuantity := min(remainingQuantity, candidate.Quantity)
		matchPrice := m.determineMatchPrice(newOrder, candidate)

		var buyOrder, sellOrder types.Order
		if newOrder.Side == types.Buy {
			buyOrder = newOrder
			sellOrder = candidate
		} else {
			buyOrder = candidate
			sellOrder = newOrder
		}

		matches = append(matches, types.Match{
			BuyOrder:  buyOrder,
			SellOrder: sellOrder,
			Quantity:  matchQuantity,
			Price:     matchPrice,
		})

		remainingQuantity -= matchQuantity
	}

	return matches
}

func (m *PriceTimeOrderMatcher) canMatch(newOrder types.Order, existingOrder types.Order) bool {
	if newOrder.Type == types.Market || existingOrder.Type == types.Market {
		return true
	}

	if newOrder.Side == types.Buy {
		return newOrder.Price >= existingOrder.Price
	} else {
		return newOrder.Price <= existingOrder.Price
	}
}

func (m *PriceTimeOrderMatcher) determineMatchPrice(newOrder types.Order, existingOrder types.Order) float64 {
	if existingOrder.Type == types.Market {
		if newOrder.Type == types.Market {
			return 0
		}
		return newOrder.Price
	}

	if newOrder.Type == types.Market {
		return existingOrder.Price
	}

	return existingOrder.Price
}

func (m *PriceTimeOrderMatcher) sortByPriceTimePriority(orders []types.Order, newOrderSide types.OrderSide) {
	sort.Slice(orders, func(i, j int) bool {
		orderI, orderJ := orders[i], orders[j]

		if newOrderSide == types.Buy {
			if orderI.Price != orderJ.Price {
				return orderI.Price < orderJ.Price
			}
		} else {
			if orderI.Price != orderJ.Price {
				return orderI.Price > orderJ.Price
			}
		}

		return orderI.Timestamp.Before(orderJ.Timestamp)
	})
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}