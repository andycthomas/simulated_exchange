package engine

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"simulated_exchange/internal/types"
)

type SimpleTradeExecutor struct {
	tradeRepo TradeRepository
}

func NewSimpleTradeExecutor(tradeRepo TradeRepository) *SimpleTradeExecutor {
	return &SimpleTradeExecutor{
		tradeRepo: tradeRepo,
	}
}

func (te *SimpleTradeExecutor) ExecuteTrade(buyOrder types.Order, sellOrder types.Order, quantity float64, price float64) (types.Trade, error) {
	if buyOrder.Symbol != sellOrder.Symbol {
		return types.Trade{}, fmt.Errorf("symbol mismatch: buy order symbol %s, sell order symbol %s", buyOrder.Symbol, sellOrder.Symbol)
	}

	if buyOrder.Side != types.Buy {
		return types.Trade{}, fmt.Errorf("invalid buy order side: %s", buyOrder.Side)
	}

	if sellOrder.Side != types.Sell {
		return types.Trade{}, fmt.Errorf("invalid sell order side: %s", sellOrder.Side)
	}

	if quantity <= 0 {
		return types.Trade{}, fmt.Errorf("invalid quantity: %f", quantity)
	}

	if price < 0 {
		return types.Trade{}, fmt.Errorf("invalid price: %f", price)
	}

	if quantity > buyOrder.Quantity {
		return types.Trade{}, fmt.Errorf("trade quantity %f exceeds buy order quantity %f", quantity, buyOrder.Quantity)
	}

	if quantity > sellOrder.Quantity {
		return types.Trade{}, fmt.Errorf("trade quantity %f exceeds sell order quantity %f", quantity, sellOrder.Quantity)
	}

	trade := types.Trade{
		ID:          uuid.New().String(),
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		Symbol:      buyOrder.Symbol,
		Quantity:    quantity,
		Price:       price,
		Timestamp:   time.Now(),
	}

	err := te.tradeRepo.Save(trade)
	if err != nil {
		return types.Trade{}, fmt.Errorf("failed to save trade: %w", err)
	}

	return trade, nil
}