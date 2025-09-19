package repository

import (
	"fmt"
	"sync"

	"simulated_exchange/internal/types"
)

type MemoryTradeRepository struct {
	trades map[string]types.Trade
	mutex  sync.RWMutex
}

func NewMemoryTradeRepository() *MemoryTradeRepository {
	return &MemoryTradeRepository{
		trades: make(map[string]types.Trade),
		mutex:  sync.RWMutex{},
	}
}

func (r *MemoryTradeRepository) Save(trade types.Trade) error {
	if trade.ID == "" {
		return fmt.Errorf("trade ID cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.trades[trade.ID] = trade
	return nil
}

func (r *MemoryTradeRepository) GetByID(id string) (types.Trade, error) {
	if id == "" {
		return types.Trade{}, fmt.Errorf("trade ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	trade, exists := r.trades[id]
	if !exists {
		return types.Trade{}, fmt.Errorf("trade with ID %s not found", id)
	}

	return trade, nil
}

func (r *MemoryTradeRepository) GetBySymbol(symbol string) ([]types.Trade, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var trades []types.Trade
	for _, trade := range r.trades {
		if trade.Symbol == symbol {
			trades = append(trades, trade)
		}
	}

	return trades, nil
}

func (r *MemoryTradeRepository) GetAll() ([]types.Trade, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	trades := make([]types.Trade, 0, len(r.trades))
	for _, trade := range r.trades {
		trades = append(trades, trade)
	}

	return trades, nil
}