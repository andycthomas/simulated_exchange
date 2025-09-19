package engine

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"simulated_exchange/internal/types"
)

type TradingEngine struct {
	orderRepo     OrderRepository
	tradeRepo     TradeRepository
	matcher       OrderMatcher
	executor      TradeExecutor
	mutex         sync.RWMutex
}

func NewTradingEngine(
	orderRepo OrderRepository,
	tradeRepo TradeRepository,
	matcher OrderMatcher,
	executor TradeExecutor,
) *TradingEngine {
	return &TradingEngine{
		orderRepo: orderRepo,
		tradeRepo: tradeRepo,
		matcher:   matcher,
		executor:  executor,
		mutex:     sync.RWMutex{},
	}
}

func (te *TradingEngine) PlaceOrder(order types.Order) error {
	if err := te.validateOrder(order); err != nil {
		return fmt.Errorf("invalid order: %w", err)
	}

	if order.ID == "" {
		order.ID = uuid.New().String()
	}

	if order.Timestamp.IsZero() {
		order.Timestamp = time.Now()
	}

	te.mutex.Lock()
	defer te.mutex.Unlock()

	if err := te.orderRepo.Save(order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	existingOrders, err := te.orderRepo.GetBySymbol(order.Symbol)
	if err != nil {
		return fmt.Errorf("failed to get existing orders: %w", err)
	}

	var ordersToMatch []types.Order
	for _, existingOrder := range existingOrders {
		if existingOrder.ID != order.ID {
			ordersToMatch = append(ordersToMatch, existingOrder)
		}
	}

	matches := te.matcher.FindMatches(order, ordersToMatch)

	remainingQuantity := order.Quantity
	for _, match := range matches {
		if remainingQuantity <= 0 {
			break
		}

		trade, err := te.executor.ExecuteTrade(match.BuyOrder, match.SellOrder, match.Quantity, match.Price)
		if err != nil {
			return fmt.Errorf("failed to execute trade: %w", err)
		}

		updatedOrder, err := te.updateOrderQuantities(match, trade.Quantity, order.ID)
		if err != nil {
			return fmt.Errorf("failed to update order quantities: %w", err)
		}

		if updatedOrder != nil {
			order = *updatedOrder
		}

		remainingQuantity -= trade.Quantity
	}

	if remainingQuantity > 0 {
		order.Quantity = remainingQuantity
		if err := te.orderRepo.Save(order); err != nil {
			return fmt.Errorf("failed to save remaining order: %w", err)
		}
	} else {
		if err := te.orderRepo.Delete(order.ID); err != nil {
			return fmt.Errorf("failed to delete filled order: %w", err)
		}
	}

	return nil
}

func (te *TradingEngine) CancelOrder(id string) error {
	if id == "" {
		return fmt.Errorf("order ID cannot be empty")
	}

	te.mutex.Lock()
	defer te.mutex.Unlock()

	_, err := te.orderRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	if err := te.orderRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return nil
}

func (te *TradingEngine) GetOrderBook(symbol string) (types.OrderBook, error) {
	if symbol == "" {
		return types.OrderBook{}, fmt.Errorf("symbol cannot be empty")
	}

	te.mutex.RLock()
	defer te.mutex.RUnlock()

	orders, err := te.orderRepo.GetBySymbol(symbol)
	if err != nil {
		return types.OrderBook{}, fmt.Errorf("failed to get orders for symbol: %w", err)
	}

	var bids, asks []types.Order

	for _, order := range orders {
		if order.Side == types.Buy {
			bids = append(bids, order)
		} else {
			asks = append(asks, order)
		}
	}

	sort.Slice(bids, func(i, j int) bool {
		if bids[i].Price != bids[j].Price {
			return bids[i].Price > bids[j].Price
		}
		return bids[i].Timestamp.Before(bids[j].Timestamp)
	})

	sort.Slice(asks, func(i, j int) bool {
		if asks[i].Price != asks[j].Price {
			return asks[i].Price < asks[j].Price
		}
		return asks[i].Timestamp.Before(asks[j].Timestamp)
	})

	return types.OrderBook{
		Symbol: symbol,
		Bids:   bids,
		Asks:   asks,
	}, nil
}

func (te *TradingEngine) validateOrder(order types.Order) error {
	if order.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	if order.Side != types.Buy && order.Side != types.Sell {
		return fmt.Errorf("invalid order side: %s", order.Side)
	}

	if order.Type != types.Market && order.Type != types.Limit {
		return fmt.Errorf("invalid order type: %s", order.Type)
	}

	if order.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive: %f", order.Quantity)
	}

	if order.Type == types.Limit && order.Price <= 0 {
		return fmt.Errorf("limit order price must be positive: %f", order.Price)
	}

	return nil
}

func (te *TradingEngine) updateOrderQuantities(match types.Match, tradeQuantity float64, newOrderID string) (updatedNewOrder *types.Order, err error) {
	buyOrder, err := te.orderRepo.GetByID(match.BuyOrder.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get buy order: %w", err)
	}

	sellOrder, err := te.orderRepo.GetByID(match.SellOrder.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sell order: %w", err)
	}

	buyOrder.Quantity -= tradeQuantity
	sellOrder.Quantity -= tradeQuantity

	var newOrderUpdated *types.Order

	if buyOrder.ID == newOrderID {
		newOrderUpdated = &buyOrder
	} else {
		if buyOrder.Quantity <= 0 {
			if err := te.orderRepo.Delete(buyOrder.ID); err != nil {
				return nil, fmt.Errorf("failed to delete filled buy order: %w", err)
			}
		} else {
			if err := te.orderRepo.Save(buyOrder); err != nil {
				return nil, fmt.Errorf("failed to update buy order: %w", err)
			}
		}
	}

	if sellOrder.ID == newOrderID {
		newOrderUpdated = &sellOrder
	} else {
		if sellOrder.Quantity <= 0 {
			if err := te.orderRepo.Delete(sellOrder.ID); err != nil {
				return nil, fmt.Errorf("failed to delete filled sell order: %w", err)
			}
		} else {
			if err := te.orderRepo.Save(sellOrder); err != nil {
				return nil, fmt.Errorf("failed to update sell order: %w", err)
			}
		}
	}

	return newOrderUpdated, nil
}