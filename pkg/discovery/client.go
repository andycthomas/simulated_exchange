package discovery

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// DiscoveryClient provides client-side service discovery functionality
type DiscoveryClient struct {
	registry     *ServiceRegistry
	cache        map[string]*ServiceInfo
	cacheMutex   sync.RWMutex
	cacheTimeout time.Duration
	logger       *slog.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// ClientConfig holds configuration for the discovery client
type ClientConfig struct {
	RedisClient  *redis.Client
	Logger       *slog.Logger
	CacheTimeout time.Duration
}

// NewDiscoveryClient creates a new service discovery client
func NewDiscoveryClient(config ClientConfig) *DiscoveryClient {
	if config.CacheTimeout == 0 {
		config.CacheTimeout = 30 * time.Second
	}

	registryConfig := RegistryConfig{
		RedisClient: config.RedisClient,
		Logger:      config.Logger,
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &DiscoveryClient{
		registry:     NewServiceRegistry(registryConfig),
		cache:        make(map[string]*ServiceInfo),
		cacheTimeout: config.CacheTimeout,
		logger:       config.Logger,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start cache refresh goroutine
	client.wg.Add(1)
	go client.refreshCacheLoop()

	return client
}

// GetService retrieves service information, using cache if available
func (dc *DiscoveryClient) GetService(serviceName string) (*ServiceInfo, error) {
	// Try cache first
	dc.cacheMutex.RLock()
	cached, exists := dc.cache[serviceName]
	dc.cacheMutex.RUnlock()

	if exists && time.Since(cached.LastHeartbeat) < dc.cacheTimeout {
		return cached, nil
	}

	// Cache miss or stale, fetch from registry
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	service, err := dc.registry.FindService(ctx, serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to find service %s: %w", serviceName, err)
	}

	// Update cache
	dc.cacheMutex.Lock()
	dc.cache[serviceName] = service
	dc.cacheMutex.Unlock()

	return service, nil
}

// GetServiceURL returns the full URL for a service
func (dc *DiscoveryClient) GetServiceURL(serviceName string) (string, error) {
	service, err := dc.GetService(serviceName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s:%s", service.Protocol, service.Host, service.Port), nil
}

// GetHealthyServices returns all healthy services in the environment
func (dc *DiscoveryClient) GetHealthyServices(environment string) (map[string]*ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	allServices, err := dc.registry.FindServicesByEnvironment(ctx, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to find services in environment %s: %w", environment, err)
	}

	healthyServices := make(map[string]*ServiceInfo)
	for name, service := range allServices {
		if service.Status == ServiceStatusHealthy {
			healthyServices[name] = service
		}
	}

	return healthyServices, nil
}

// WaitForService waits for a service to become available
func (dc *DiscoveryClient) WaitForService(serviceName string, timeout time.Duration) (*ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for service %s", serviceName)
		case <-ticker.C:
			service, err := dc.GetService(serviceName)
			if err == nil && service.Status == ServiceStatusHealthy {
				return service, nil
			}
		}
	}
}

// RefreshCache manually refreshes the service cache
func (dc *DiscoveryClient) RefreshCache() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	services, err := dc.registry.Discover(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh service cache: %w", err)
	}

	dc.cacheMutex.Lock()
	dc.cache = services
	dc.cacheMutex.Unlock()

	dc.logger.Debug("Service cache refreshed", "services_count", len(services))
	return nil
}

// Close shuts down the discovery client
func (dc *DiscoveryClient) Close() error {
	dc.cancel()
	dc.wg.Wait()
	return dc.registry.Close()
}

// refreshCacheLoop periodically refreshes the service cache
func (dc *DiscoveryClient) refreshCacheLoop() {
	defer dc.wg.Done()

	ticker := time.NewTicker(dc.cacheTimeout / 2) // Refresh at half the cache timeout
	defer ticker.Stop()

	for {
		select {
		case <-dc.ctx.Done():
			return
		case <-ticker.C:
			if err := dc.RefreshCache(); err != nil {
				dc.logger.Error("Failed to refresh service cache", "error", err)
			}
		}
	}
}

// ServiceWatcher provides real-time updates when services change
type ServiceWatcher struct {
	client   *DiscoveryClient
	watchers map[string][]ServiceChangeHandler
	mutex    sync.RWMutex
	logger   *slog.Logger
}

// ServiceChangeHandler is called when a service changes
type ServiceChangeHandler func(serviceName string, service *ServiceInfo, changeType ServiceChangeType)

// ServiceChangeType represents the type of service change
type ServiceChangeType string

const (
	ServiceAdded   ServiceChangeType = "added"
	ServiceUpdated ServiceChangeType = "updated"
	ServiceRemoved ServiceChangeType = "removed"
)

// NewServiceWatcher creates a new service watcher
func NewServiceWatcher(client *DiscoveryClient, logger *slog.Logger) *ServiceWatcher {
	return &ServiceWatcher{
		client:   client,
		watchers: make(map[string][]ServiceChangeHandler),
		logger:   logger,
	}
}

// Watch registers a handler to be called when a specific service changes
func (sw *ServiceWatcher) Watch(serviceName string, handler ServiceChangeHandler) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	sw.watchers[serviceName] = append(sw.watchers[serviceName], handler)
	sw.logger.Debug("Service watcher registered", "service", serviceName)
}

// UnWatch removes all watchers for a service
func (sw *ServiceWatcher) UnWatch(serviceName string) {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	delete(sw.watchers, serviceName)
	sw.logger.Debug("Service watchers removed", "service", serviceName)
}

// LoadBalancer provides simple load balancing for service instances
type LoadBalancer struct {
	client    *DiscoveryClient
	roundRobin map[string]int
	mutex     sync.RWMutex
	logger    *slog.Logger
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(client *DiscoveryClient, logger *slog.Logger) *LoadBalancer {
	return &LoadBalancer{
		client:     client,
		roundRobin: make(map[string]int),
		logger:     logger,
	}
}

// GetServiceInstance returns the next available service instance using round-robin
func (lb *LoadBalancer) GetServiceInstance(serviceName string) (*ServiceInfo, error) {
	// For now, we just return the single service instance
	// In a real microservices environment, you might have multiple instances
	return lb.client.GetService(serviceName)
}

// GetServiceInstanceURL returns the URL for the next available service instance
func (lb *LoadBalancer) GetServiceInstanceURL(serviceName string) (string, error) {
	service, err := lb.GetServiceInstance(serviceName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s:%s", service.Protocol, service.Host, service.Port), nil
}