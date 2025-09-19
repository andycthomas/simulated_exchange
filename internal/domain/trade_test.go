package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrade_ValidTrade(t *testing.T) {
	trade, err := NewTrade("trade-1", "buy-order-1", "sell-order-1", "BTCUSD", 50000.0, 1.5)

	require.NoError(t, err)
	assert.Equal(t, "trade-1", trade.ID)
	assert.Equal(t, "buy-order-1", trade.BuyOrderID)
	assert.Equal(t, "sell-order-1", trade.SellOrderID)
	assert.Equal(t, "BTCUSD", trade.Symbol)
	assert.Equal(t, 50000.0, trade.Price)
	assert.Equal(t, 1.5, trade.Quantity)
	assert.WithinDuration(t, time.Now(), trade.Timestamp, time.Second)
}

func TestNewTrade_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		buyOrderID  string
		sellOrderID string
		symbol      string
		price       float64
		quantity    float64
		expectedErr string
	}{
		{
			name:        "empty ID",
			id:          "",
			buyOrderID:  "buy-order-1",
			sellOrderID: "sell-order-1",
			symbol:      "BTCUSD",
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "trade ID cannot be empty",
		},
		{
			name:        "empty buy order ID",
			id:          "trade-1",
			buyOrderID:  "",
			sellOrderID: "sell-order-1",
			symbol:      "BTCUSD",
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "buy order ID cannot be empty",
		},
		{
			name:        "empty sell order ID",
			id:          "trade-1",
			buyOrderID:  "buy-order-1",
			sellOrderID: "",
			symbol:      "BTCUSD",
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "sell order ID cannot be empty",
		},
		{
			name:        "empty symbol",
			id:          "trade-1",
			buyOrderID:  "buy-order-1",
			sellOrderID: "sell-order-1",
			symbol:      "",
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "symbol cannot be empty",
		},
		{
			name:        "zero price",
			id:          "trade-1",
			buyOrderID:  "buy-order-1",
			sellOrderID: "sell-order-1",
			symbol:      "BTCUSD",
			price:       0,
			quantity:    1.0,
			expectedErr: "price must be positive",
		},
		{
			name:        "negative price",
			id:          "trade-1",
			buyOrderID:  "buy-order-1",
			sellOrderID: "sell-order-1",
			symbol:      "BTCUSD",
			price:       -100.0,
			quantity:    1.0,
			expectedErr: "price must be positive",
		},
		{
			name:        "zero quantity",
			id:          "trade-1",
			buyOrderID:  "buy-order-1",
			sellOrderID: "sell-order-1",
			symbol:      "BTCUSD",
			price:       50000.0,
			quantity:    0,
			expectedErr: "quantity must be positive",
		},
		{
			name:        "negative quantity",
			id:          "trade-1",
			buyOrderID:  "buy-order-1",
			sellOrderID: "sell-order-1",
			symbol:      "BTCUSD",
			price:       50000.0,
			quantity:    -1.0,
			expectedErr: "quantity must be positive",
		},
		{
			name:        "same buy and sell order IDs",
			id:          "trade-1",
			buyOrderID:  "order-1",
			sellOrderID: "order-1",
			symbol:      "BTCUSD",
			price:       50000.0,
			quantity:    1.0,
			expectedErr: "buy order ID and sell order ID cannot be the same",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trade, err := NewTrade(tt.id, tt.buyOrderID, tt.sellOrderID, tt.symbol, tt.price, tt.quantity)

			assert.Error(t, err)
			assert.Nil(t, trade)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestTrade_IsValid(t *testing.T) {
	validTrade := &Trade{
		ID:          "trade-1",
		BuyOrderID:  "buy-order-1",
		SellOrderID: "sell-order-1",
		Symbol:      "BTCUSD",
		Price:       50000.0,
		Quantity:    1.0,
		Timestamp:   time.Now(),
	}

	assert.NoError(t, validTrade.IsValid())

	invalidTrade := &Trade{
		ID:          "",
		BuyOrderID:  "buy-order-1",
		SellOrderID: "sell-order-1",
		Symbol:      "BTCUSD",
		Price:       50000.0,
		Quantity:    1.0,
		Timestamp:   time.Now(),
	}

	assert.Error(t, invalidTrade.IsValid())
	assert.Contains(t, invalidTrade.IsValid().Error(), "trade ID cannot be empty")
}

func TestTrade_Value(t *testing.T) {
	trade := &Trade{
		ID:          "trade-1",
		BuyOrderID:  "buy-order-1",
		SellOrderID: "sell-order-1",
		Symbol:      "BTCUSD",
		Price:       50000.0,
		Quantity:    1.5,
		Timestamp:   time.Now(),
	}

	expectedValue := 50000.0 * 1.5
	assert.Equal(t, expectedValue, trade.Value())
}

func TestTrade_ValueWithZeroPrice(t *testing.T) {
	trade := &Trade{
		ID:          "trade-1",
		BuyOrderID:  "buy-order-1",
		SellOrderID: "sell-order-1",
		Symbol:      "BTCUSD",
		Price:       0,
		Quantity:    1.5,
		Timestamp:   time.Now(),
	}

	assert.Equal(t, 0.0, trade.Value())
}

func TestTrade_ValueWithZeroQuantity(t *testing.T) {
	trade := &Trade{
		ID:          "trade-1",
		BuyOrderID:  "buy-order-1",
		SellOrderID: "sell-order-1",
		Symbol:      "BTCUSD",
		Price:       50000.0,
		Quantity:    0,
		Timestamp:   time.Now(),
	}

	assert.Equal(t, 0.0, trade.Value())
}