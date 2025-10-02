package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"simulated_exchange/pkg/shared"
)

// PostgresUserRepository implements shared.UserRepository using PostgreSQL
type PostgresUserRepository struct {
	db *sqlx.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sqlx.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Create inserts a new user into the database
func (r *PostgresUserRepository) Create(ctx context.Context, user *shared.User) error {
	query := `
		INSERT INTO trading.users (id, username, email, password_hash, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.CreatedAt, user.UpdatedAt, user.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*shared.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, is_active
		FROM trading.users
		WHERE id = $1`

	var user shared.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, shared.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by their username
func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*shared.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, is_active
		FROM trading.users
		WHERE username = $1`

	var user shared.User
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, shared.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email
func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*shared.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, is_active
		FROM trading.users
		WHERE email = $1`

	var user shared.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, shared.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update updates an existing user
func (r *PostgresUserRepository) Update(ctx context.Context, user *shared.User) error {
	query := `
		UPDATE trading.users
		SET username = $2, email = $3, password_hash = $4, updated_at = $5, is_active = $6
		WHERE id = $1`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.UpdatedAt, user.IsActive)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return shared.ErrUserNotFound
	}

	return nil
}

// Delete removes a user from the database (soft delete by setting is_active to false)
func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE trading.users SET is_active = false, updated_at = $2 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return shared.ErrUserNotFound
	}

	return nil
}

// GetActiveUsers retrieves all active users
func (r *PostgresUserRepository) GetActiveUsers(ctx context.Context) ([]*shared.User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at, updated_at, is_active
		FROM trading.users
		WHERE is_active = true
		ORDER BY created_at DESC`

	var users []shared.User
	err := r.db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	// Convert to slice of pointers
	result := make([]*shared.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

// CheckUsernameExists checks if a username already exists
func (r *PostgresUserRepository) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM trading.users WHERE username = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, username)
	if err != nil {
		return false, fmt.Errorf("failed to check username exists: %w", err)
	}

	return exists, nil
}

// CheckEmailExists checks if an email already exists
func (r *PostgresUserRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM trading.users WHERE email = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email exists: %w", err)
	}

	return exists, nil
}