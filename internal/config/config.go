package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database DatabaseConfig `json:"database"`

	// Logging configuration
	Logging LoggingConfig `json:"logging"`

	// Metrics configuration
	Metrics MetricsConfig `json:"metrics"`

	// Simulation configuration
	Simulation SimulationConfig `json:"simulation"`

	// Health check configuration
	Health HealthConfig `json:"health"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	Environment  string        `json:"environment"`
	EnableCORS   bool          `json:"enable_cors"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	ConnectionString string        `json:"connection_string"`
	MaxConnections   int           `json:"max_connections"`
	MaxIdleTime      time.Duration `json:"max_idle_time"`
	MaxLifetime      time.Duration `json:"max_lifetime"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"` // json or text
	OutputFile string `json:"output_file"`
	MaxSize    int    `json:"max_size"`    // megabytes
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`     // days
}

// MetricsConfig contains metrics collection settings
type MetricsConfig struct {
	Enabled        bool          `json:"enabled"`
	CollectionTime time.Duration `json:"collection_time"`
	RetentionTime  time.Duration `json:"retention_time"`
	ExportInterval time.Duration `json:"export_interval"`
}

// SimulationConfig contains market simulation settings
type SimulationConfig struct {
	Enabled           bool          `json:"enabled"`
	Symbols           []string      `json:"symbols"`
	OrdersPerSecond   float64       `json:"orders_per_second"`
	VolatilityFactor  float64       `json:"volatility_factor"`
	WorkerCount       int           `json:"worker_count"`
	PatternInterval   time.Duration `json:"pattern_interval"`
	EnableVolatility  bool          `json:"enable_volatility"`
}

// HealthConfig contains health check settings
type HealthConfig struct {
	Enabled       bool          `json:"enabled"`
	CheckInterval time.Duration `json:"check_interval"`
	Timeout       time.Duration `json:"timeout"`
	Endpoint      string        `json:"endpoint"`
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvOrDefault("SERVER_PORT", "8080"),
			Host:         getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationOrDefault("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationOrDefault("SERVER_IDLE_TIMEOUT", 120*time.Second),
			Environment:  getEnvOrDefault("ENVIRONMENT", "development"),
			EnableCORS:   getBoolOrDefault("ENABLE_CORS", true),
		},
		Database: DatabaseConfig{
			ConnectionString: getEnvOrDefault("DATABASE_URL", ""),
			MaxConnections:   getIntOrDefault("DB_MAX_CONNECTIONS", 25),
			MaxIdleTime:      getDurationOrDefault("DB_MAX_IDLE_TIME", 30*time.Minute),
			MaxLifetime:      getDurationOrDefault("DB_MAX_LIFETIME", 1*time.Hour),
		},
		Logging: LoggingConfig{
			Level:      getEnvOrDefault("LOG_LEVEL", "info"),
			Format:     getEnvOrDefault("LOG_FORMAT", "json"),
			OutputFile: getEnvOrDefault("LOG_FILE", ""),
			MaxSize:    getIntOrDefault("LOG_MAX_SIZE", 100),
			MaxBackups: getIntOrDefault("LOG_MAX_BACKUPS", 3),
			MaxAge:     getIntOrDefault("LOG_MAX_AGE", 28),
		},
		Metrics: MetricsConfig{
			Enabled:        getBoolOrDefault("METRICS_ENABLED", true),
			CollectionTime: getDurationOrDefault("METRICS_COLLECTION_TIME", 60*time.Second),
			RetentionTime:  getDurationOrDefault("METRICS_RETENTION_TIME", 24*time.Hour),
			ExportInterval: getDurationOrDefault("METRICS_EXPORT_INTERVAL", 5*time.Minute),
		},
		Simulation: SimulationConfig{
			Enabled:          getBoolOrDefault("SIMULATION_ENABLED", true),
			Symbols:          getStringSliceOrDefault("SIMULATION_SYMBOLS", []string{"BTCUSD", "ETHUSD", "ADAUSD"}),
			OrdersPerSecond:  getFloatOrDefault("SIMULATION_ORDERS_PER_SEC", 10.0),
			VolatilityFactor: getFloatOrDefault("SIMULATION_VOLATILITY", 0.05),
			WorkerCount:      getIntOrDefault("SIMULATION_WORKERS", 4),
			PatternInterval:  getDurationOrDefault("SIMULATION_PATTERN_INTERVAL", 5*time.Minute),
			EnableVolatility: getBoolOrDefault("SIMULATION_ENABLE_VOLATILITY", true),
		},
		Health: HealthConfig{
			Enabled:       getBoolOrDefault("HEALTH_ENABLED", true),
			CheckInterval: getDurationOrDefault("HEALTH_CHECK_INTERVAL", 30*time.Second),
			Timeout:       getDurationOrDefault("HEALTH_TIMEOUT", 10*time.Second),
			Endpoint:      getEnvOrDefault("HEALTH_ENDPOINT", "/health"),
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

	if c.Server.Environment != "development" && c.Server.Environment != "staging" && c.Server.Environment != "production" {
		return fmt.Errorf("invalid environment: %s", c.Server.Environment)
	}

	if c.Logging.Level != "debug" && c.Logging.Level != "info" && c.Logging.Level != "warn" && c.Logging.Level != "error" {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	if c.Logging.Format != "json" && c.Logging.Format != "text" {
		return fmt.Errorf("invalid log format: %s", c.Logging.Format)
	}

	if c.Simulation.Enabled && len(c.Simulation.Symbols) == 0 {
		return fmt.Errorf("simulation symbols are required when simulation is enabled")
	}

	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
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