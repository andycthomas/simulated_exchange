package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"simulated_exchange/pkg/shared"
)

// RedisClient implements shared.CacheRepository using Redis
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	return &RedisClient{client: rdb}
}

// Ping tests the connection to Redis
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Set stores a value in Redis with expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Get retrieves a value from Redis and unmarshals it
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return shared.NewBusinessError("CACHE_MISS", "key not found in cache")
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	return nil
}

// Delete removes a key from Redis
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// Exists checks if a key exists in Redis
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if key %s exists: %w", key, err)
	}

	return count > 0, nil
}

// SetOrderBook stores an order book in Redis
func (r *RedisClient) SetOrderBook(ctx context.Context, symbol string, orderBook *shared.OrderBook) error {
	key := fmt.Sprintf("orderbook:%s", symbol)
	return r.Set(ctx, key, orderBook, 10*time.Minute)
}

// GetOrderBook retrieves an order book from Redis
func (r *RedisClient) GetOrderBook(ctx context.Context, symbol string) (*shared.OrderBook, error) {
	key := fmt.Sprintf("orderbook:%s", symbol)
	var orderBook shared.OrderBook
	err := r.Get(ctx, key, &orderBook)
	if err != nil {
		return nil, err
	}
	return &orderBook, nil
}

// SetMarketData stores market data in Redis
func (r *RedisClient) SetMarketData(ctx context.Context, symbol string, data *shared.MarketData) error {
	key := fmt.Sprintf("marketdata:%s", symbol)
	return r.Set(ctx, key, data, 5*time.Minute)
}

// GetMarketData retrieves market data from Redis
func (r *RedisClient) GetMarketData(ctx context.Context, symbol string) (*shared.MarketData, error) {
	key := fmt.Sprintf("marketdata:%s", symbol)
	var marketData shared.MarketData
	err := r.Get(ctx, key, &marketData)
	if err != nil {
		return nil, err
	}
	return &marketData, nil
}

// SetPrice stores the latest price for a symbol
func (r *RedisClient) SetPrice(ctx context.Context, symbol string, price float64) error {
	key := fmt.Sprintf("price:%s", symbol)
	return r.client.Set(ctx, key, price, 1*time.Hour).Err()
}

// GetPrice retrieves the latest price for a symbol
func (r *RedisClient) GetPrice(ctx context.Context, symbol string) (float64, error) {
	key := fmt.Sprintf("price:%s", symbol)
	price, err := r.client.Get(ctx, key).Float64()
	if err != nil {
		if err == redis.Nil {
			return 0, shared.NewBusinessError("PRICE_NOT_FOUND", "price not found in cache")
		}
		return 0, fmt.Errorf("failed to get price for %s: %w", symbol, err)
	}
	return price, nil
}

// SetUserSession stores user session data
func (r *RedisClient) SetUserSession(ctx context.Context, sessionID string, userID string, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Set(ctx, key, userID, expiration).Err()
}

// GetUserSession retrieves user session data
func (r *RedisClient) GetUserSession(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	userID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", shared.NewBusinessError("SESSION_NOT_FOUND", "session not found")
		}
		return "", fmt.Errorf("failed to get session %s: %w", sessionID, err)
	}
	return userID, nil
}

// DeleteUserSession removes a user session
func (r *RedisClient) DeleteUserSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Del(ctx, key).Err()
}

// IncrementCounter increments a counter and returns the new value
func (r *RedisClient) IncrementCounter(ctx context.Context, key string) (int64, error) {
	value, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment counter %s: %w", key, err)
	}
	return value, nil
}

// SetCounterExpiration sets expiration for a counter
func (r *RedisClient) SetCounterExpiration(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// GetListLength returns the length of a Redis list
func (r *RedisClient) GetListLength(ctx context.Context, key string) (int64, error) {
	length, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get list length for %s: %w", key, err)
	}
	return length, nil
}

// PushToList pushes an item to the end of a Redis list
func (r *RedisClient) PushToList(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = r.client.RPush(ctx, key, data).Err()
	if err != nil {
		return fmt.Errorf("failed to push to list %s: %w", key, err)
	}

	return nil
}

// PopFromList pops an item from the beginning of a Redis list
func (r *RedisClient) PopFromList(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.LPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return shared.NewBusinessError("LIST_EMPTY", "list is empty")
		}
		return fmt.Errorf("failed to pop from list %s: %w", key, err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value from list %s: %w", key, err)
	}

	return nil
}

// TrimList trims a Redis list to the specified range
func (r *RedisClient) TrimList(ctx context.Context, key string, start, stop int64) error {
	err := r.client.LTrim(ctx, key, start, stop).Err()
	if err != nil {
		return fmt.Errorf("failed to trim list %s: %w", key, err)
	}
	return nil
}

// RedisHealthChecker implements the shared.HealthChecker interface for Redis
type RedisHealthChecker struct {
	client *RedisClient
}

// NewRedisHealthChecker creates a new health checker for Redis
func NewRedisHealthChecker(client *RedisClient) *RedisHealthChecker {
	return &RedisHealthChecker{client: client}
}

// Check performs a health check on Redis
func (h *RedisHealthChecker) Check(ctx context.Context) error {
	return h.client.Ping(ctx)
}

// Name returns the name of the health checker
func (h *RedisHealthChecker) Name() string {
	return "redis"
}