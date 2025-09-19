package domain

import (
	"errors"
	"time"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPartial   OrderStatus = "PARTIAL"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Symbol    string      `json:"symbol"`
	Side      OrderSide   `json:"side"`
	Type      OrderType   `json:"type"`
	Price     float64     `json:"price"`
	Quantity  float64     `json:"quantity"`
	Status    OrderStatus `json:"status"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewOrder(id, userID, symbol string, side OrderSide, orderType OrderType, price, quantity float64) (*Order, error) {
	if id == "" {
		return nil, errors.New("order ID cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}
	if side != OrderSideBuy && side != OrderSideSell {
		return nil, errors.New("invalid order side")
	}
	if orderType != OrderTypeMarket && orderType != OrderTypeLimit {
		return nil, errors.New("invalid order type")
	}
	if price < 0 {
		return nil, errors.New("price cannot be negative")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	if orderType == OrderTypeLimit && price == 0 {
		return nil, errors.New("limit orders must have a price greater than zero")
	}

	return &Order{
		ID:        id,
		UserID:    userID,
		Symbol:    symbol,
		Side:      side,
		Type:      orderType,
		Price:     price,
		Quantity:  quantity,
		Status:    OrderStatusPending,
		Timestamp: time.Now(),
	}, nil
}

func (o *Order) IsValid() error {
	if o.ID == "" {
		return errors.New("order ID cannot be empty")
	}
	if o.UserID == "" {
		return errors.New("user ID cannot be empty")
	}
	if o.Symbol == "" {
		return errors.New("symbol cannot be empty")
	}
	if o.Side != OrderSideBuy && o.Side != OrderSideSell {
		return errors.New("invalid order side")
	}
	if o.Type != OrderTypeMarket && o.Type != OrderTypeLimit {
		return errors.New("invalid order type")
	}
	if o.Price < 0 {
		return errors.New("price cannot be negative")
	}
	if o.Quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if o.Type == OrderTypeLimit && o.Price == 0 {
		return errors.New("limit orders must have a price greater than zero")
	}
	return nil
}

func (o *Order) UpdateStatus(status OrderStatus) error {
	validStatuses := map[OrderStatus]bool{
		OrderStatusPending:   true,
		OrderStatusPartial:   true,
		OrderStatusFilled:    true,
		OrderStatusCancelled: true,
		OrderStatusRejected:  true,
	}

	if !validStatuses[status] {
		return errors.New("invalid order status")
	}

	o.Status = status
	return nil
}