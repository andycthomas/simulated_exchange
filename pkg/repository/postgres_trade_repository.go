package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"simulated_exchange/pkg/shared"
)

// PostgresTradeRepository implements shared.TradeRepository using PostgreSQL
type PostgresTradeRepository struct {
	db *sqlx.DB
}

// NewPostgresTradeRepository creates a new PostgreSQL trade repository
func NewPostgresTradeRepository(db *sqlx.DB) *PostgresTradeRepository {
	return &PostgresTradeRepository{db: db}
}

// Create inserts a new trade into the database
func (r *PostgresTradeRepository) Create(ctx context.Context, trade *shared.Trade) error {
	query := `
		INSERT INTO trading.trades (id, buy_order_id, sell_order_id, symbol, price, quantity, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		trade.ID, trade.BuyOrderID, trade.SellOrderID, trade.Symbol,
		trade.Price, trade.Quantity, trade.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create trade: %w", err)
	}

	return nil
}

// GetByID retrieves a trade by its ID
func (r *PostgresTradeRepository) GetByID(ctx context.Context, id string) (*shared.Trade, error) {
	query := `
		SELECT id, buy_order_id, sell_order_id, symbol, price, quantity, created_at
		FROM trading.trades
		WHERE id = $1`

	var trade shared.Trade
	err := r.db.GetContext(ctx, &trade, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, shared.ErrTradeNotFound
		}
		return nil, fmt.Errorf("failed to get trade by ID: %w", err)
	}

	return &trade, nil
}

// GetByOrderID retrieves all trades for a specific order
func (r *PostgresTradeRepository) GetByOrderID(ctx context.Context, orderID string) ([]*shared.Trade, error) {
	query := `
		SELECT id, buy_order_id, sell_order_id, symbol, price, quantity, created_at
		FROM trading.trades
		WHERE buy_order_id = $1 OR sell_order_id = $1
		ORDER BY created_at DESC`

	var trades []shared.Trade
	err := r.db.SelectContext(ctx, &trades, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades by order ID: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.Trade, len(trades))
	for i := range trades {
		result[i] = &trades[i]
	}

	return result, nil
}

// GetBySymbol retrieves all trades for a specific symbol
func (r *PostgresTradeRepository) GetBySymbol(ctx context.Context, symbol string) ([]*shared.Trade, error) {
	query := `
		SELECT id, buy_order_id, sell_order_id, symbol, price, quantity, created_at
		FROM trading.trades
		WHERE symbol = $1
		ORDER BY created_at DESC`

	var trades []shared.Trade
	err := r.db.SelectContext(ctx, &trades, query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades by symbol: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.Trade, len(trades))
	for i := range trades {
		result[i] = &trades[i]
	}

	return result, nil
}

// GetTradesInTimeRange retrieves trades within a specific time range
func (r *PostgresTradeRepository) GetTradesInTimeRange(ctx context.Context, start, end time.Time) ([]*shared.Trade, error) {
	query := `
		SELECT id, buy_order_id, sell_order_id, symbol, price, quantity, created_at
		FROM trading.trades
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC`

	var trades []shared.Trade
	err := r.db.SelectContext(ctx, &trades, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades in time range: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.Trade, len(trades))
	for i := range trades {
		result[i] = &trades[i]
	}

	return result, nil
}

// GetRecentTrades retrieves the most recent trades with a limit
func (r *PostgresTradeRepository) GetRecentTrades(ctx context.Context, limit int) ([]*shared.Trade, error) {
	query := `
		SELECT id, buy_order_id, sell_order_id, symbol, price, quantity, created_at
		FROM trading.trades
		ORDER BY created_at DESC
		LIMIT $1`

	var trades []shared.Trade
	err := r.db.SelectContext(ctx, &trades, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent trades: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.Trade, len(trades))
	for i := range trades {
		result[i] = &trades[i]
	}

	return result, nil
}

// GetVolumeBySymbol returns the total trading volume for a symbol within a time range
func (r *PostgresTradeRepository) GetVolumeBySymbol(ctx context.Context, symbol string, start, end time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(quantity * price), 0) as volume
		FROM trading.trades
		WHERE symbol = $1 AND created_at >= $2 AND created_at <= $3`

	var volume float64
	err := r.db.GetContext(ctx, &volume, query, symbol, start, end)
	if err != nil {
		return 0, fmt.Errorf("failed to get volume by symbol: %w", err)
	}

	return volume, nil
}

// GetTradeCount returns the total number of trades within a time range
func (r *PostgresTradeRepository) GetTradeCount(ctx context.Context, start, end time.Time) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM trading.trades
		WHERE created_at >= $1 AND created_at <= $2`

	var count int64
	err := r.db.GetContext(ctx, &count, query, start, end)
	if err != nil {
		return 0, fmt.Errorf("failed to get trade count: %w", err)
	}

	return count, nil
}

// GetTradeCountBySymbol returns the number of trades for a specific symbol within a time range
func (r *PostgresTradeRepository) GetTradeCountBySymbol(ctx context.Context, symbol string, start, end time.Time) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM trading.trades
		WHERE symbol = $1 AND created_at >= $2 AND created_at <= $3`

	var count int64
	err := r.db.GetContext(ctx, &count, query, symbol, start, end)
	if err != nil {
		return 0, fmt.Errorf("failed to get trade count by symbol: %w", err)
	}

	return count, nil
}

// GetLatestPriceBySymbol returns the latest trade price for a symbol
func (r *PostgresTradeRepository) GetLatestPriceBySymbol(ctx context.Context, symbol string) (float64, error) {
	query := `
		SELECT price
		FROM trading.trades
		WHERE symbol = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var price float64
	err := r.db.GetContext(ctx, &price, query, symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, shared.ErrTradeNotFound
		}
		return 0, fmt.Errorf("failed to get latest price by symbol: %w", err)
	}

	return price, nil
}