package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"simulated_exchange/pkg/shared"
)

// PostgresDB implements the shared.DBConnection interface
type PostgresDB struct {
	db *sqlx.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(connectionString string) (*PostgresDB, error) {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Configure connection pool for high-performance trading
	db.SetMaxOpenConns(200)        // Maximum concurrent connections
	db.SetMaxIdleConns(50)         // Keep more idle connections ready
	db.SetConnMaxLifetime(30 * time.Minute)  // Refresh connections regularly
	db.SetConnMaxIdleTime(5 * time.Minute)   // Close idle connections after 5 minutes

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// Ping checks if the database connection is alive
func (p *PostgresDB) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Begin starts a new transaction
func (p *PostgresDB) Begin(ctx context.Context) (shared.Transaction, error) {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &PostgresTransaction{tx: tx}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// Stats returns database statistics
func (p *PostgresDB) Stats() interface{} {
	return p.db.Stats()
}

// GetDB returns the underlying sqlx.DB instance
func (p *PostgresDB) GetDB() *sqlx.DB {
	return p.db
}

// PostgresTransaction implements the shared.Transaction interface
type PostgresTransaction struct {
	tx *sqlx.Tx
}

// Commit commits the transaction
func (t *PostgresTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *PostgresTransaction) Rollback() error {
	return t.tx.Rollback()
}

// Exec executes a query without returning rows
func (t *PostgresTransaction) Exec(ctx context.Context, query string, args ...interface{}) error {
	_, err := t.tx.ExecContext(ctx, query, args...)
	return err
}

// Query executes a query that returns rows
func (t *PostgresTransaction) Query(ctx context.Context, query string, args ...interface{}) (shared.Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &PostgresRows{rows: rows}, nil
}

// QueryRow executes a query that returns a single row
func (t *PostgresTransaction) QueryRow(ctx context.Context, query string, args ...interface{}) shared.Row {
	row := t.tx.QueryRowContext(ctx, query, args...)
	return &PostgresRow{row: row}
}

// PostgresRows implements the shared.Rows interface
type PostgresRows struct {
	rows *sql.Rows
}

// Next returns true if there is another row
func (r *PostgresRows) Next() bool {
	return r.rows.Next()
}

// Scan copies the columns in the current row into the values pointed at by dest
func (r *PostgresRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

// Close closes the rows iterator
func (r *PostgresRows) Close() error {
	return r.rows.Close()
}

// Err returns the error, if any, that was encountered during iteration
func (r *PostgresRows) Err() error {
	return r.rows.Err()
}

// PostgresRow implements the shared.Row interface
type PostgresRow struct {
	row *sql.Row
}

// Scan copies the columns in the current row into the values pointed at by dest
func (r *PostgresRow) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// HealthChecker implements the shared.HealthChecker interface for PostgreSQL
type PostgresHealthChecker struct {
	db *PostgresDB
}

// NewPostgresHealthChecker creates a new health checker for PostgreSQL
func NewPostgresHealthChecker(db *PostgresDB) *PostgresHealthChecker {
	return &PostgresHealthChecker{db: db}
}

// Check performs a health check on the database
func (h *PostgresHealthChecker) Check(ctx context.Context) error {
	return h.db.Ping(ctx)
}

// Name returns the name of the health checker
func (h *PostgresHealthChecker) Name() string {
	return "postgres"
}