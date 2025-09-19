package types

import (
	"time"
)

type OrderType string
type OrderSide string

const (
	Market OrderType = "MARKET"
	Limit  OrderType = "LIMIT"
)

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

type Order struct {
	ID        string
	Symbol    string
	Side      OrderSide
	Type      OrderType
	Quantity  float64
	Price     float64
	Timestamp time.Time
}

type Trade struct {
	ID           string
	BuyOrderID   string
	SellOrderID  string
	Symbol       string
	Quantity     float64
	Price        float64
	Timestamp    time.Time
}

type OrderBook struct {
	Symbol string
	Bids   []Order
	Asks   []Order
}

type Match struct {
	BuyOrder  Order
	SellOrder Order
	Quantity  float64
	Price     float64
}