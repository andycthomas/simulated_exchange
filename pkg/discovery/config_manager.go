package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConfigManager provides centralized configuration management
type ConfigManager struct {
	redis       *redis.Client
	logger      *slog.Logger
	cache       map[string]interface{}
	cacheMutex  sync.RWMutex
	subscribers map[string][]ConfigChangeHandler
	subsMutex   sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// ConfigChangeHandler is called when configuration changes
type ConfigChangeHandler func(key string, oldValue, newValue interface{})

// ConfigManagerConfig holds configuration for the config manager
type ConfigManagerConfig struct {
	RedisClient *redis.Client
	Logger      *slog.Logger
	KeyPrefix   string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(config ConfigManagerConfig) *ConfigManager {
	ctx, cancel := context.WithCancel(context.Background())

	cm := &ConfigManager{
		redis:       config.RedisClient,
		logger:      config.Logger,
		cache:       make(map[string]interface{}),
		subscribers: make(map[string][]ConfigChangeHandler),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start background tasks
	cm.wg.Add(1)
	go cm.watchConfigChanges()

	return cm
}

// Set stores a configuration value
func (cm *ConfigManager) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal config value: %w", err)
	}

	configKey := cm.getConfigKey(key)
	err = cm.redis.Set(ctx, configKey, data, 0).Err() // No expiration for config
	if err != nil {
		return fmt.Errorf("failed to set config in Redis: %w", err)
	}

	// Update cache
	cm.cacheMutex.Lock()
	oldValue := cm.cache[key]
	cm.cache[key] = value
	cm.cacheMutex.Unlock()

	// Notify subscribers
	cm.notifySubscribers(key, oldValue, value)

	cm.logger.Debug("Configuration set", "key", key)
	return nil
}

// Get retrieves a configuration value
func (cm *ConfigManager) Get(ctx context.Context, key string) (interface{}, error) {
	// Try cache first
	cm.cacheMutex.RLock()
	cached, exists := cm.cache[key]
	cm.cacheMutex.RUnlock()

	if exists {
		return cached, nil
	}

	// Cache miss, fetch from Redis
	configKey := cm.getConfigKey(key)
	data, err := cm.redis.Get(ctx, configKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("configuration key %s not found", key)
		}
		return nil, fmt.Errorf("failed to get config from Redis: %w", err)
	}

	var value interface{}
	if err := json.Unmarshal([]byte(data), &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config value: %w", err)
	}

	// Update cache
	cm.cacheMutex.Lock()
	cm.cache[key] = value
	cm.cacheMutex.Unlock()

	return value, nil
}

// GetString retrieves a string configuration value
func (cm *ConfigManager) GetString(ctx context.Context, key string) (string, error) {
	value, err := cm.Get(ctx, key)
	if err != nil {
		return "", err
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("configuration value for key %s is not a string", key)
	}

	return str, nil
}

// GetInt retrieves an integer configuration value
func (cm *ConfigManager) GetInt(ctx context.Context, key string) (int, error) {
	value, err := cm.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	// JSON unmarshaling creates float64 for numbers
	if f, ok := value.(float64); ok {
		return int(f), nil
	}

	if i, ok := value.(int); ok {
		return i, nil
	}

	return 0, fmt.Errorf("configuration value for key %s is not a number", key)
}

// GetFloat retrieves a float configuration value
func (cm *ConfigManager) GetFloat(ctx context.Context, key string) (float64, error) {
	value, err := cm.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	if f, ok := value.(float64); ok {
		return f, nil
	}

	if i, ok := value.(int); ok {
		return float64(i), nil
	}

	return 0, fmt.Errorf("configuration value for key %s is not a number", key)
}

// GetBool retrieves a boolean configuration value
func (cm *ConfigManager) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := cm.Get(ctx, key)
	if err != nil {
		return false, err
	}

	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("configuration value for key %s is not a boolean", key)
	}

	return b, nil
}

// Delete removes a configuration value
func (cm *ConfigManager) Delete(ctx context.Context, key string) error {
	configKey := cm.getConfigKey(key)
	err := cm.redis.Del(ctx, configKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete config from Redis: %w", err)
	}

	// Remove from cache
	cm.cacheMutex.Lock()
	oldValue := cm.cache[key]
	delete(cm.cache, key)
	cm.cacheMutex.Unlock()

	// Notify subscribers
	cm.notifySubscribers(key, oldValue, nil)

	cm.logger.Debug("Configuration deleted", "key", key)
	return nil
}

// GetAll retrieves all configuration values with a prefix
func (cm *ConfigManager) GetAll(ctx context.Context, prefix string) (map[string]interface{}, error) {
	pattern := cm.getConfigKey(prefix + "*")
	keys, err := cm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get config keys from Redis: %w", err)
	}

	configs := make(map[string]interface{})

	for _, fullKey := range keys {
		// Extract the original key by removing the prefix
		key := cm.extractOriginalKey(fullKey)

		data, err := cm.redis.Get(ctx, fullKey).Result()
		if err != nil {
			cm.logger.Warn("Failed to get config data", "key", key, "error", err)
			continue
		}

		var value interface{}
		if err := json.Unmarshal([]byte(data), &value); err != nil {
			cm.logger.Warn("Failed to unmarshal config data", "key", key, "error", err)
			continue
		}

		configs[key] = value
	}

	return configs, nil
}

// Subscribe registers a handler to be called when a configuration changes
func (cm *ConfigManager) Subscribe(key string, handler ConfigChangeHandler) {
	cm.subsMutex.Lock()
	defer cm.subsMutex.Unlock()

	cm.subscribers[key] = append(cm.subscribers[key], handler)
	cm.logger.Debug("Configuration subscriber registered", "key", key)
}

// Unsubscribe removes all subscribers for a configuration key
func (cm *ConfigManager) Unsubscribe(key string) {
	cm.subsMutex.Lock()
	defer cm.subsMutex.Unlock()

	delete(cm.subscribers, key)
	cm.logger.Debug("Configuration subscribers removed", "key", key)
}

// LoadInitialConfig loads configuration from environment or defaults
func (cm *ConfigManager) LoadInitialConfig(ctx context.Context, defaults map[string]interface{}) error {
	for key, defaultValue := range defaults {
		// Check if config already exists
		_, err := cm.Get(ctx, key)
		if err != nil {
			// Config doesn't exist, set default
			if err := cm.Set(ctx, key, defaultValue); err != nil {
				return fmt.Errorf("failed to set default config %s: %w", key, err)
			}
			cm.logger.Info("Set default configuration", "key", key, "value", defaultValue)
		}
	}

	return nil
}

// Close shuts down the configuration manager
func (cm *ConfigManager) Close() error {
	cm.cancel()
	cm.wg.Wait()
	return nil
}

// watchConfigChanges watches for configuration changes in Redis
func (cm *ConfigManager) watchConfigChanges() {
	defer cm.wg.Done()

	pattern := cm.getConfigKey("*")
	pubsub := cm.redis.PSubscribe(cm.ctx, "__keyspace@0__:"+pattern)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			// Extract the configuration key from the Redis keyspace notification
			configKey := cm.extractConfigKeyFromNotification(msg.Channel)
			if configKey == "" {
				continue
			}

			// Refresh the specific configuration
			cm.refreshConfig(configKey)
		case <-time.After(30 * time.Second):
			// Periodic full refresh as fallback
			cm.refreshAllConfigs()
		}
	}
}

func (cm *ConfigManager) refreshConfig(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get current cached value
	cm.cacheMutex.RLock()
	oldValue := cm.cache[key]
	cm.cacheMutex.RUnlock()

	// Fetch new value
	newValue, err := cm.Get(ctx, key)
	if err != nil {
		cm.logger.Warn("Failed to refresh config", "key", key, "error", err)
		return
	}

	// Notify subscribers if value changed
	if !configValuesEqual(oldValue, newValue) {
		cm.notifySubscribers(key, oldValue, newValue)
	}
}

func (cm *ConfigManager) refreshAllConfigs() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	configs, err := cm.GetAll(ctx, "")
	if err != nil {
		cm.logger.Error("Failed to refresh all configs", "error", err)
		return
	}

	cm.cacheMutex.Lock()
	cm.cache = configs
	cm.cacheMutex.Unlock()

	cm.logger.Debug("All configurations refreshed", "count", len(configs))
}

func (cm *ConfigManager) notifySubscribers(key string, oldValue, newValue interface{}) {
	cm.subsMutex.RLock()
	handlers := cm.subscribers[key]
	cm.subsMutex.RUnlock()

	for _, handler := range handlers {
		go handler(key, oldValue, newValue)
	}
}

func (cm *ConfigManager) getConfigKey(key string) string {
	return fmt.Sprintf("config:%s", key)
}

func (cm *ConfigManager) extractOriginalKey(fullKey string) string {
	prefix := "config:"
	if len(fullKey) > len(prefix) {
		return fullKey[len(prefix):]
	}
	return fullKey
}

func (cm *ConfigManager) extractConfigKeyFromNotification(channel string) string {
	// Redis keyspace notifications format: __keyspace@0__:config:key
	prefix := "__keyspace@0__:config:"
	if len(channel) > len(prefix) {
		return channel[len(prefix):]
	}
	return ""
}

func configValuesEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}