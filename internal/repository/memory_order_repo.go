package repository

import (
	"fmt"
	"sync"

	"simulated_exchange/internal/types"
)

type MemoryOrderRepository struct {
	orders map[string]types.Order
	mutex  sync.RWMutex
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		orders: make(map[string]types.Order),
		mutex:  sync.RWMutex{},
	}
}

func (r *MemoryOrderRepository) Save(order types.Order) error {
	if order.ID == "" {
		return fmt.Errorf("order ID cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.orders[order.ID] = order
	return nil
}

func (r *MemoryOrderRepository) GetByID(id string) (types.Order, error) {
	if id == "" {
		return types.Order{}, fmt.Errorf("order ID cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return types.Order{}, fmt.Errorf("order with ID %s not found", id)
	}

	return order, nil
}

func (r *MemoryOrderRepository) GetBySymbol(symbol string) ([]types.Order, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var orders []types.Order
	for _, order := range r.orders {
		if order.Symbol == symbol {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

func (r *MemoryOrderRepository) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("order ID cannot be empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.orders[id]; !exists {
		return fmt.Errorf("order with ID %s not found", id)
	}

	delete(r.orders, id)
	return nil
}

func (r *MemoryOrderRepository) GetAll() ([]types.Order, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	orders := make([]types.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}

	return orders, nil
}