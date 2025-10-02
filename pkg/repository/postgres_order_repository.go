package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"simulated_exchange/pkg/shared"
)

// PostgresOrderRepository implements shared.OrderRepository using PostgreSQL
type PostgresOrderRepository struct {
	db *sqlx.DB
}

// NewPostgresOrderRepository creates a new PostgreSQL order repository
func NewPostgresOrderRepository(db *sqlx.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Create inserts a new order into the database
func (r *PostgresOrderRepository) Create(ctx context.Context, order *shared.Order) error {
	query := `
		INSERT INTO trading.orders (id, user_id, symbol, side, type, price, quantity, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		order.ID, order.UserID, order.Symbol, order.Side, order.Type,
		order.Price, order.Quantity, order.Status, order.CreatedAt, order.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetByID retrieves an order by its ID
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*shared.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, price, quantity, status, created_at, updated_at
		FROM trading.orders
		WHERE id = $1`

	var order shared.Order
	err := r.db.GetContext(ctx, &order, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, shared.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order by ID: %w", err)
	}

	return &order, nil
}

// GetByUserID retrieves all orders for a specific user
func (r *PostgresOrderRepository) GetByUserID(ctx context.Context, userID string) ([]*shared.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, price, quantity, status, created_at, updated_at
		FROM trading.orders
		WHERE user_id = $1
		ORDER BY created_at DESC`

	var orders []shared.Order
	err := r.db.SelectContext(ctx, &orders, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user ID: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.Order, len(orders))
	for i := range orders {
		result[i] = &orders[i]
	}

	return result, nil
}

// GetBySymbol retrieves all orders for a specific symbol
func (r *PostgresOrderRepository) GetBySymbol(ctx context.Context, symbol string) ([]*shared.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, price, quantity, status, created_at, updated_at
		FROM trading.orders
		WHERE symbol = $1 AND status IN ('PENDING', 'PARTIAL')
		ORDER BY created_at ASC`

	var orders []shared.Order
	err := r.db.SelectContext(ctx, &orders, query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by symbol: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.Order, len(orders))
	for i := range orders {
		result[i] = &orders[i]
	}

	return result, nil
}

// GetByStatus retrieves all orders with a specific status
func (r *PostgresOrderRepository) GetByStatus(ctx context.Context, status shared.OrderStatus) ([]*shared.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, price, quantity, status, created_at, updated_at
		FROM trading.orders
		WHERE status = $1
		ORDER BY created_at DESC`

	var orders []shared.Order
	err := r.db.SelectContext(ctx, &orders, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.Order, len(orders))
	for i := range orders {
		result[i] = &orders[i]
	}

	return result, nil
}

// Update updates an existing order
func (r *PostgresOrderRepository) Update(ctx context.Context, order *shared.Order) error {
	query := `
		UPDATE trading.orders
		SET user_id = $2, symbol = $3, side = $4, type = $5, price = $6,
		    quantity = $7, status = $8, updated_at = $9
		WHERE id = $1`

	order.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		order.ID, order.UserID, order.Symbol, order.Side, order.Type,
		order.Price, order.Quantity, order.Status, order.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return shared.ErrOrderNotFound
	}

	return nil
}

// Delete removes an order from the database
func (r *PostgresOrderRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM trading.orders WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return shared.ErrOrderNotFound
	}

	return nil
}

// GetActiveOrders retrieves all active orders (PENDING or PARTIAL status)
func (r *PostgresOrderRepository) GetActiveOrders(ctx context.Context) ([]*shared.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, price, quantity, status, created_at, updated_at
		FROM trading.orders
		WHERE status IN ('PENDING', 'PARTIAL')
		ORDER BY created_at ASC`

	var orders []shared.Order
	err := r.db.SelectContext(ctx, &orders, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.Order, len(orders))
	for i := range orders {
		result[i] = &orders[i]
	}

	return result, nil
}

// GetOrdersInTimeRange retrieves orders within a specific time range
func (r *PostgresOrderRepository) GetOrdersInTimeRange(ctx context.Context, start, end time.Time) ([]*shared.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, price, quantity, status, created_at, updated_at
		FROM trading.orders
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC`

	var orders []shared.Order
	err := r.db.SelectContext(ctx, &orders, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders in time range: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.Order, len(orders))
	for i := range orders {
		result[i] = &orders[i]
	}

	return result, nil
}