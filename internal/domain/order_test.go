package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrder_ValidOrder(t *testing.T) {
	order, err := NewOrder("order-1", "user-1", "BTCUSD", OrderSideBuy, OrderTypeLimit, 50000.0, 1.5)

	require.NoError(t, err)
	assert.Equal(t, "order-1", order.ID)
	assert.Equal(t, "user-1", order.UserID)
	assert.Equal(t, "BTCUSD", order.Symbol)
	assert.Equal(t, OrderSideBuy, order.Side)
	assert.Equal(t, OrderTypeLimit, order.Type)
	assert.Equal(t, 50000.0, order.Price)
	assert.Equal(t, 1.5, order.Quantity)
	assert.Equal(t, OrderStatusPending, order.Status)
	assert.WithinDuration(t, time.Now(), order.Timestamp, time.Second)
}

func TestNewOrder_MarketOrder(t *testing.T) {
	order, err := NewOrder("order-1", "user-1", "BTCUSD", OrderSideSell, OrderTypeMarket, 0, 1.0)

	require.NoError(t, err)
	assert.Equal(t, OrderTypeMarket, order.Type)
	assert.Equal(t, 0.0, order.Price)
}

func TestNewOrder_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		userID      string
		symbol      string
		side        OrderSide
		orderType   OrderType
		price       float64
		quantity    float64
		expectedErr string
	}{
		{
			name:        "empty ID",
			id:          "",
			userID:      "user-1",
			symbol:      "BTCUSD",
			side:        OrderSideBuy,
			orderType:   OrderTypeLimit,
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "order ID cannot be empty",
		},
		{
			name:        "empty user ID",
			id:          "order-1",
			userID:      "",
			symbol:      "BTCUSD",
			side:        OrderSideBuy,
			orderType:   OrderTypeLimit,
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "user ID cannot be empty",
		},
		{
			name:        "empty symbol",
			id:          "order-1",
			userID:      "user-1",
			symbol:      "",
			side:        OrderSideBuy,
			orderType:   OrderTypeLimit,
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "symbol cannot be empty",
		},
		{
			name:        "invalid side",
			id:          "order-1",
			userID:      "user-1",
			symbol:      "BTCUSD",
			side:        "INVALID",
			orderType:   OrderTypeLimit,
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "invalid order side",
		},
		{
			name:        "invalid type",
			id:          "order-1",
			userID:      "user-1",
			symbol:      "BTCUSD",
			side:        OrderSideBuy,
			orderType:   "INVALID",
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "invalid order type",
		},
		{
			name:        "negative price",
			id:          "order-1",
			userID:      "user-1",
			symbol:      "BTCUSD",
			side:        OrderSideBuy,
			orderType:   OrderTypeLimit,
			price:       -100.0,
			quantity:    1.0,
			expectedErr: "price cannot be negative",
		},
		{
			name:        "zero quantity",
			id:          "order-1",
			userID:      "user-1",
			symbol:      "BTCUSD",
			side:        OrderSideBuy,
			orderType:   OrderTypeLimit,
			price:       50000.0,
			quantity:    0,
			expectedErr: "quantity must be positive",
		},
		{
			name:        "negative quantity",
			id:          "order-1",
			userID:      "user-1",
			symbol:      "BTCUSD",
			side:        OrderSideBuy,
			orderType:   OrderTypeLimit,
			price:       50000.0,
			quantity:    -1.0,
			expectedErr: "quantity must be positive",
		},
		{
			name:        "limit order with zero price",
			id:          "order-1",
			userID:      "user-1",
			symbol:      "BTCUSD",
			side:        OrderSideBuy,
			orderType:   OrderTypeLimit,
			price:       0,
			quantity:    1.0,
			expectedErr: "limit orders must have a price greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(tt.id, tt.userID, tt.symbol, tt.side, tt.orderType, tt.price, tt.quantity)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestOrder_IsValid(t *testing.T) {
	validOrder := &Order{
		ID:        "order-1",
		UserID:    "user-1",
		Symbol:    "BTCUSD",
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     50000.0,
		Quantity:  1.0,
		Status:    OrderStatusPending,
		Timestamp: time.Now(),
	}

	assert.NoError(t, validOrder.IsValid())

	invalidOrder := &Order{
		ID:        "",
		UserID:    "user-1",
		Symbol:    "BTCUSD",
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     50000.0,
		Quantity:  1.0,
		Status:    OrderStatusPending,
		Timestamp: time.Now(),
	}

	assert.Error(t, invalidOrder.IsValid())
}

func TestOrder_UpdateStatus(t *testing.T) {
	order := &Order{
		ID:        "order-1",
		UserID:    "user-1",
		Symbol:    "BTCUSD",
		Side:      OrderSideBuy,
		Type:      OrderTypeLimit,
		Price:     50000.0,
		Quantity:  1.0,
		Status:    OrderStatusPending,
		Timestamp: time.Now(),
	}

	err := order.UpdateStatus(OrderStatusFilled)
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusFilled, order.Status)

	err = order.UpdateStatus("INVALID_STATUS")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order status")
}

func TestOrderSideConstants(t *testing.T) {
	assert.Equal(t, OrderSide("BUY"), OrderSideBuy)
	assert.Equal(t, OrderSide("SELL"), OrderSideSell)
}

func TestOrderTypeConstants(t *testing.T) {
	assert.Equal(t, OrderType("MARKET"), OrderTypeMarket)
	assert.Equal(t, OrderType("LIMIT"), OrderTypeLimit)
}

func TestOrderStatusConstants(t *testing.T) {
	assert.Equal(t, OrderStatus("PENDING"), OrderStatusPending)
	assert.Equal(t, OrderStatus("PARTIAL"), OrderStatusPartial)
	assert.Equal(t, OrderStatus("FILLED"), OrderStatusFilled)
	assert.Equal(t, OrderStatus("CANCELLED"), OrderStatusCancelled)
	assert.Equal(t, OrderStatus("REJECTED"), OrderStatusRejected)
}