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

// ServiceInfo represents information about a registered service
type ServiceInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Host        string            `json:"host"`
	Port        string            `json:"port"`
	Protocol    string            `json:"protocol"` // http, https, grpc
	Environment string            `json:"environment"`
	Status      ServiceStatus     `json:"status"`
	Metadata    map[string]string `json:"metadata"`
	RegisteredAt time.Time        `json:"registered_at"`
	LastHeartbeat time.Time       `json:"last_heartbeat"`
	TTL         time.Duration     `json:"ttl"`
}

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	ServiceStatusHealthy   ServiceStatus = "healthy"
	ServiceStatusUnhealthy ServiceStatus = "unhealthy"
	ServiceStatusStarting  ServiceStatus = "starting"
	ServiceStatusStopping  ServiceStatus = "stopping"
)

// ServiceRegistry provides service registration and discovery
type ServiceRegistry struct {
	redis      *redis.Client
	logger     *slog.Logger
	services   map[string]*ServiceInfo
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// RegistryConfig holds configuration for the service registry
type RegistryConfig struct {
	RedisClient *redis.Client
	Logger      *slog.Logger
	KeyPrefix   string
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry(config RegistryConfig) *ServiceRegistry {
	ctx, cancel := context.WithCancel(context.Background())

	registry := &ServiceRegistry{
		redis:    config.RedisClient,
		logger:   config.Logger,
		services: make(map[string]*ServiceInfo),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start background tasks
	registry.wg.Add(1)
	go registry.backgroundCleanup()

	return registry
}

// Register registers a service in the registry
func (sr *ServiceRegistry) Register(ctx context.Context, service *ServiceInfo) error {
	if service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	// Set default values
	if service.TTL == 0 {
		service.TTL = 30 * time.Second
	}
	if service.Protocol == "" {
		service.Protocol = "http"
	}
	if service.Status == "" {
		service.Status = ServiceStatusHealthy
	}

	service.RegisteredAt = time.Now()
	service.LastHeartbeat = time.Now()

	// Store in Redis
	key := sr.getServiceKey(service.Name)
	data, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	err = sr.redis.Set(ctx, key, data, service.TTL*2).Err() // TTL * 2 for safety margin
	if err != nil {
		return fmt.Errorf("failed to register service in Redis: %w", err)
	}

	// Store locally
	sr.mutex.Lock()
	sr.services[service.Name] = service
	sr.mutex.Unlock()

	sr.logger.Info("Service registered",
		"name", service.Name,
		"version", service.Version,
		"host", service.Host,
		"port", service.Port,
		"ttl", service.TTL,
	)

	return nil
}

// Unregister removes a service from the registry
func (sr *ServiceRegistry) Unregister(ctx context.Context, serviceName string) error {
	key := sr.getServiceKey(serviceName)

	err := sr.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to unregister service from Redis: %w", err)
	}

	sr.mutex.Lock()
	delete(sr.services, serviceName)
	sr.mutex.Unlock()

	sr.logger.Info("Service unregistered", "name", serviceName)
	return nil
}

// Discover finds all registered services
func (sr *ServiceRegistry) Discover(ctx context.Context) (map[string]*ServiceInfo, error) {
	pattern := sr.getServiceKey("*")
	keys, err := sr.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get service keys from Redis: %w", err)
	}

	services := make(map[string]*ServiceInfo)

	for _, key := range keys {
		data, err := sr.redis.Get(ctx, key).Result()
		if err != nil {
			sr.logger.Warn("Failed to get service data", "key", key, "error", err)
			continue
		}

		var service ServiceInfo
		if err := json.Unmarshal([]byte(data), &service); err != nil {
			sr.logger.Warn("Failed to unmarshal service data", "key", key, "error", err)
			continue
		}

		services[service.Name] = &service
	}

	return services, nil
}

// FindService finds a specific service by name
func (sr *ServiceRegistry) FindService(ctx context.Context, serviceName string) (*ServiceInfo, error) {
	key := sr.getServiceKey(serviceName)
	data, err := sr.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("service %s not found", serviceName)
		}
		return nil, fmt.Errorf("failed to get service from Redis: %w", err)
	}

	var service ServiceInfo
	if err := json.Unmarshal([]byte(data), &service); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service data: %w", err)
	}

	return &service, nil
}

// FindServicesByEnvironment finds services in a specific environment
func (sr *ServiceRegistry) FindServicesByEnvironment(ctx context.Context, environment string) (map[string]*ServiceInfo, error) {
	allServices, err := sr.Discover(ctx)
	if err != nil {
		return nil, err
	}

	filteredServices := make(map[string]*ServiceInfo)
	for name, service := range allServices {
		if service.Environment == environment {
			filteredServices[name] = service
		}
	}

	return filteredServices, nil
}

// Heartbeat updates the last heartbeat timestamp for a service
func (sr *ServiceRegistry) Heartbeat(ctx context.Context, serviceName string) error {
	service, err := sr.FindService(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("service not found for heartbeat: %w", err)
	}

	service.LastHeartbeat = time.Now()

	// Update in Redis
	key := sr.getServiceKey(serviceName)
	data, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	err = sr.redis.Set(ctx, key, data, service.TTL*2).Err()
	if err != nil {
		return fmt.Errorf("failed to update heartbeat in Redis: %w", err)
	}

	sr.logger.Debug("Service heartbeat updated", "name", serviceName)
	return nil
}

// UpdateServiceStatus updates the status of a service
func (sr *ServiceRegistry) UpdateServiceStatus(ctx context.Context, serviceName string, status ServiceStatus) error {
	service, err := sr.FindService(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("service not found for status update: %w", err)
	}

	service.Status = status
	service.LastHeartbeat = time.Now()

	// Update in Redis
	key := sr.getServiceKey(serviceName)
	data, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	err = sr.redis.Set(ctx, key, data, service.TTL*2).Err()
	if err != nil {
		return fmt.Errorf("failed to update service status in Redis: %w", err)
	}

	sr.logger.Info("Service status updated",
		"name", serviceName,
		"status", status,
	)

	return nil
}

// GetServiceURL returns the full URL for a service
func (sr *ServiceRegistry) GetServiceURL(ctx context.Context, serviceName string) (string, error) {
	service, err := sr.FindService(ctx, serviceName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s:%s", service.Protocol, service.Host, service.Port), nil
}

// Close shuts down the service registry
func (sr *ServiceRegistry) Close() error {
	sr.cancel()
	sr.wg.Wait()
	return nil
}

// backgroundCleanup removes expired services
func (sr *ServiceRegistry) backgroundCleanup() {
	defer sr.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sr.ctx.Done():
			return
		case <-ticker.C:
			sr.cleanupExpiredServices()
		}
	}
}

func (sr *ServiceRegistry) cleanupExpiredServices() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	services, err := sr.Discover(ctx)
	if err != nil {
		sr.logger.Error("Failed to discover services for cleanup", "error", err)
		return
	}

	now := time.Now()
	expired := make([]string, 0)

	for name, service := range services {
		if now.Sub(service.LastHeartbeat) > service.TTL*3 { // Give extra time before cleanup
			expired = append(expired, name)
		}
	}

	for _, serviceName := range expired {
		if err := sr.Unregister(ctx, serviceName); err != nil {
			sr.logger.Error("Failed to unregister expired service",
				"name", serviceName,
				"error", err,
			)
		} else {
			sr.logger.Info("Cleaned up expired service", "name", serviceName)
		}
	}
}

func (sr *ServiceRegistry) getServiceKey(serviceName string) string {
	return fmt.Sprintf("services:%s", serviceName)
}

// HealthChecker for the service registry
type RegistryHealthChecker struct {
	registry *ServiceRegistry
}

// NewRegistryHealthChecker creates a new health checker for the registry
func NewRegistryHealthChecker(registry *ServiceRegistry) *RegistryHealthChecker {
	return &RegistryHealthChecker{registry: registry}
}

// Check performs a health check on the service registry
func (h *RegistryHealthChecker) Check(ctx context.Context) error {
	// Test Redis connectivity
	return h.registry.redis.Ping(ctx).Err()
}

// Name returns the name of the health checker
func (h *RegistryHealthChecker) Name() string {
	return "service-registry"
}