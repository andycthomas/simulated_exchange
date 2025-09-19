package domain

import (
	"errors"
	"time"
)

type Trade struct {
	ID          string    `json:"id"`
	BuyOrderID  string    `json:"buy_order_id"`
	SellOrderID string    `json:"sell_order_id"`
	Symbol      string    `json:"symbol"`
	Price       float64   `json:"price"`
	Quantity    float64   `json:"quantity"`
	Timestamp   time.Time `json:"timestamp"`
}

func NewTrade(id, buyOrderID, sellOrderID, symbol string, price, quantity float64) (*Trade, error) {
	if id == "" {
		return nil, errors.New("trade ID cannot be empty")
	}
	if buyOrderID == "" {
		return nil, errors.New("buy order ID cannot be empty")
	}
	if sellOrderID == "" {
		return nil, errors.New("sell order ID cannot be empty")
	}
	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}
	if price <= 0 {
		return nil, errors.New("price must be positive")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	if buyOrderID == sellOrderID {
		return nil, errors.New("buy order ID and sell order ID cannot be the same")
	}

	return &Trade{
		ID:          id,
		BuyOrderID:  buyOrderID,
		SellOrderID: sellOrderID,
		Symbol:      symbol,
		Price:       price,
		Quantity:    quantity,
		Timestamp:   time.Now(),
	}, nil
}

func (t *Trade) IsValid() error {
	if t.ID == "" {
		return errors.New("trade ID cannot be empty")
	}
	if t.BuyOrderID == "" {
		return errors.New("buy order ID cannot be empty")
	}
	if t.SellOrderID == "" {
		return errors.New("sell order ID cannot be empty")
	}
	if t.Symbol == "" {
		return errors.New("symbol cannot be empty")
	}
	if t.Price <= 0 {
		return errors.New("price must be positive")
	}
	if t.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if t.BuyOrderID == t.SellOrderID {
		return errors.New("buy order ID and sell order ID cannot be the same")
	}
	return nil
}

func (t *Trade) Value() float64 {
	return t.Price * t.Quantity
}