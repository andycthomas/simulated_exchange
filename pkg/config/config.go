package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config represents the application configuration
type Config struct {
	Service  ServiceConfig  `json:"service"`
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Logging  LoggingConfig  `json:"logging"`
	Metrics  MetricsConfig  `json:"metrics"`
}

// ServiceConfig contains service-specific configuration
type ServiceConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	EnableCORS   bool          `json:"enable_cors"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Host         string        `json:"host"`
	Port         string        `json:"port"`
	Database     string        `json:"database"`
	Username     string        `json:"username"`
	Password     string        `json:"password"`
	SSLMode      string        `json:"ssl_mode"`
	MaxOpenConns int           `json:"max_open_conns"`
	MaxIdleConns int           `json:"max_idle_conns"`
	MaxLifetime  time.Duration `json:"max_lifetime"`
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputFile string `json:"output_file"`
}

// MetricsConfig contains metrics collection settings
type MetricsConfig struct {
	Enabled        bool          `json:"enabled"`
	CollectionTime time.Duration `json:"collection_time"`
	RetentionTime  time.Duration `json:"retention_time"`
	ExportInterval time.Duration `json:"export_interval"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Service: ServiceConfig{
			Name:        getEnvOrDefault("SERVICE_NAME", "trading-api"),
			Version:     getEnvOrDefault("SERVICE_VERSION", "1.0.0"),
			Environment: getEnvOrDefault("ENVIRONMENT", "development"),
		},
		Server: ServerConfig{
			Port:         getEnvOrDefault("SERVER_PORT", "8080"),
			Host:         getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationOrDefault("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationOrDefault("SERVER_IDLE_TIMEOUT", 120*time.Second),
			EnableCORS:   getBoolOrDefault("ENABLE_CORS", true),
		},
		Database: DatabaseConfig{
			Host:         getEnvOrDefault("DB_HOST", "postgres"),
			Port:         getEnvOrDefault("DB_PORT", "5432"),
			Database:     getEnvOrDefault("DB_NAME", "trading_db"),
			Username:     getEnvOrDefault("DB_USER", "trading_user"),
			Password:     getEnvOrDefault("DB_PASSWORD", "trading_pass"),
			SSLMode:      getEnvOrDefault("DB_SSL_MODE", "disable"),
			MaxOpenConns: getIntOrDefault("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getIntOrDefault("DB_MAX_IDLE_CONNS", 25),
			MaxLifetime:  getDurationOrDefault("DB_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnvOrDefault("REDIS_HOST", "redis"),
			Port:     getEnvOrDefault("REDIS_PORT", "6379"),
			Password: getEnvOrDefault("REDIS_PASSWORD", ""),
			Database: getIntOrDefault("REDIS_DB", 0),
		},
		Logging: LoggingConfig{
			Level:      getEnvOrDefault("LOG_LEVEL", "info"),
			Format:     getEnvOrDefault("LOG_FORMAT", "json"),
			OutputFile: getEnvOrDefault("LOG_FILE", ""),
		},
		Metrics: MetricsConfig{
			Enabled:        getBoolOrDefault("METRICS_ENABLED", true),
			CollectionTime: getDurationOrDefault("METRICS_COLLECTION_TIME", 60*time.Second),
			RetentionTime:  getDurationOrDefault("METRICS_RETENTION_TIME", 24*time.Hour),
			ExportInterval: getDurationOrDefault("METRICS_EXPORT_INTERVAL", 5*time.Minute),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.Service.Environment != "development" && c.Service.Environment != "staging" && c.Service.Environment != "production" {
		return fmt.Errorf("invalid environment: %s", c.Service.Environment)
	}

	if c.Logging.Level != "debug" && c.Logging.Level != "info" && c.Logging.Level != "warn" && c.Logging.Level != "error" {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	if c.Logging.Format != "json" && c.Logging.Format != "text" {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	return nil
}

// GetDatabaseConnectionString returns the PostgreSQL connection string
func (c *Config) GetDatabaseConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.Username,
		c.Database.Password, c.Database.Database, c.Database.SSLMode)
}

// GetRedisAddress returns the Redis address
func (c *Config) GetRedisAddress() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Service.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Service.Environment == "production"
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getStringSliceOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}