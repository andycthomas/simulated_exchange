package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"simulated_exchange/pkg/discovery"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusUnknown   HealthStatus = "unknown"
)

// ServiceHealth represents the health status of a single service
type ServiceHealth struct {
	Name         string                 `json:"name"`
	Status       HealthStatus           `json:"status"`
	LastChecked  time.Time              `json:"last_checked"`
	ResponseTime time.Duration          `json:"response_time"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	URL          string                 `json:"url"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status         HealthStatus             `json:"status"`
	Timestamp      time.Time                `json:"timestamp"`
	Services       map[string]ServiceHealth `json:"services"`
	Summary        HealthSummary            `json:"summary"`
	OverallUptime  time.Duration            `json:"overall_uptime"`
}

// HealthSummary provides a summary of system health
type HealthSummary struct {
	TotalServices    int `json:"total_services"`
	HealthyServices  int `json:"healthy_services"`
	DegradedServices int `json:"degraded_services"`
	UnhealthyServices int `json:"unhealthy_services"`
	UnknownServices  int `json:"unknown_services"`
}

// HealthAggregator collects health information from all services
type HealthAggregator struct {
	discoveryClient *discovery.DiscoveryClient
	httpClient      *http.Client
	logger          *slog.Logger
	cache           *SystemHealth
	cacheMutex      sync.RWMutex
	checkInterval   time.Duration
	timeout         time.Duration
	startTime       time.Time
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// HealthAggregatorConfig holds configuration for the health aggregator
type HealthAggregatorConfig struct {
	DiscoveryClient *discovery.DiscoveryClient
	Logger          *slog.Logger
	CheckInterval   time.Duration
	Timeout         time.Duration
}

// NewHealthAggregator creates a new health aggregator
func NewHealthAggregator(config HealthAggregatorConfig) *HealthAggregator {
	if config.CheckInterval == 0 {
		config.CheckInterval = 30 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	aggregator := &HealthAggregator{
		discoveryClient: config.DiscoveryClient,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger:        config.Logger,
		checkInterval: config.CheckInterval,
		timeout:       config.Timeout,
		startTime:     time.Now(),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Initialize cache
	aggregator.cache = &SystemHealth{
		Status:    StatusUnknown,
		Timestamp: time.Now(),
		Services:  make(map[string]ServiceHealth),
		Summary:   HealthSummary{},
	}

	return aggregator
}

// Start begins health monitoring
func (ha *HealthAggregator) Start() error {
	ha.logger.Info("Starting health aggregator")

	// Start health checking loop
	ha.wg.Add(1)
	go ha.healthCheckLoop()

	return nil
}

// Stop stops health monitoring
func (ha *HealthAggregator) Stop() error {
	ha.logger.Info("Stopping health aggregator")
	ha.cancel()
	ha.wg.Wait()
	return nil
}

// GetSystemHealth returns the current system health status
func (ha *HealthAggregator) GetSystemHealth() SystemHealth {
	ha.cacheMutex.RLock()
	defer ha.cacheMutex.RUnlock()

	// Create a copy to avoid race conditions
	health := *ha.cache
	health.Services = make(map[string]ServiceHealth)
	for name, service := range ha.cache.Services {
		health.Services[name] = service
	}
	health.OverallUptime = time.Since(ha.startTime)

	return health
}

// GetServiceHealth returns health status for a specific service
func (ha *HealthAggregator) GetServiceHealth(serviceName string) (ServiceHealth, bool) {
	ha.cacheMutex.RLock()
	defer ha.cacheMutex.RUnlock()

	service, exists := ha.cache.Services[serviceName]
	return service, exists
}

// healthCheckLoop periodically checks all services
func (ha *HealthAggregator) healthCheckLoop() {
	defer ha.wg.Done()

	ticker := time.NewTicker(ha.checkInterval)
	defer ticker.Stop()

	// Perform initial check
	ha.performHealthChecks()

	for {
		select {
		case <-ha.ctx.Done():
			return
		case <-ticker.C:
			ha.performHealthChecks()
		}
	}
}

// performHealthChecks checks all registered services
func (ha *HealthAggregator) performHealthChecks() {
	// Get all healthy services from discovery
	services, err := ha.discoveryClient.GetHealthyServices("production")
	if err != nil {
		ha.logger.Error("Failed to get services from discovery", "error", err)
		return
	}

	// Check each service in parallel
	var wg sync.WaitGroup
	serviceHealths := make(map[string]ServiceHealth)
	mutex := sync.Mutex{}

	for serviceName, serviceInfo := range services {
		wg.Add(1)
		go func(name string, info *discovery.ServiceInfo) {
			defer wg.Done()

			health := ha.checkServiceHealth(name, info)

			mutex.Lock()
			serviceHealths[name] = health
			mutex.Unlock()
		}(serviceName, serviceInfo)
	}

	wg.Wait()

	// Update cache with results
	ha.updateCache(serviceHealths)
}

// checkServiceHealth checks a single service
func (ha *HealthAggregator) checkServiceHealth(serviceName string, serviceInfo *discovery.ServiceInfo) ServiceHealth {
	startTime := time.Now()
	healthURL := fmt.Sprintf("%s://%s:%s/health", serviceInfo.Protocol, serviceInfo.Host, serviceInfo.Port)

	health := ServiceHealth{
		Name:        serviceName,
		Status:      StatusUnknown,
		LastChecked: startTime,
		URL:         healthURL,
	}

	// Create request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ha.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		health.Status = StatusUnhealthy
		health.Error = fmt.Sprintf("Failed to create request: %v", err)
		health.ResponseTime = time.Since(startTime)
		return health
	}

	// Set headers
	req.Header.Set("User-Agent", "health-aggregator/1.0")
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := ha.httpClient.Do(req)
	if err != nil {
		health.Status = StatusUnhealthy
		health.Error = fmt.Sprintf("Request failed: %v", err)
		health.ResponseTime = time.Since(startTime)
		return health
	}
	defer resp.Body.Close()

	health.ResponseTime = time.Since(startTime)

	// Check status code
	if resp.StatusCode == http.StatusOK {
		health.Status = StatusHealthy
	} else if resp.StatusCode == http.StatusServiceUnavailable {
		health.Status = StatusDegraded
	} else {
		health.Status = StatusUnhealthy
		health.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	// Try to parse response body for additional details
	var healthResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err == nil {
		health.Details = healthResponse
	}

	return health
}

// updateCache updates the cached system health
func (ha *HealthAggregator) updateCache(serviceHealths map[string]ServiceHealth) {
	ha.cacheMutex.Lock()
	defer ha.cacheMutex.Unlock()

	ha.cache.Services = serviceHealths
	ha.cache.Timestamp = time.Now()

	// Calculate summary
	summary := HealthSummary{
		TotalServices: len(serviceHealths),
	}

	for _, health := range serviceHealths {
		switch health.Status {
		case StatusHealthy:
			summary.HealthyServices++
		case StatusDegraded:
			summary.DegradedServices++
		case StatusUnhealthy:
			summary.UnhealthyServices++
		case StatusUnknown:
			summary.UnknownServices++
		}
	}

	ha.cache.Summary = summary

	// Determine overall status
	if summary.UnhealthyServices > 0 {
		ha.cache.Status = StatusUnhealthy
	} else if summary.DegradedServices > 0 || summary.UnknownServices > 0 {
		ha.cache.Status = StatusDegraded
	} else if summary.HealthyServices > 0 {
		ha.cache.Status = StatusHealthy
	} else {
		ha.cache.Status = StatusUnknown
	}

	ha.logger.Debug("Health check completed",
		"total_services", summary.TotalServices,
		"healthy", summary.HealthyServices,
		"degraded", summary.DegradedServices,
		"unhealthy", summary.UnhealthyServices,
		"overall_status", ha.cache.Status,
	)
}

// HealthHandler provides HTTP endpoints for health information
type HealthHandler struct {
	aggregator *HealthAggregator
	logger     *slog.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(aggregator *HealthAggregator, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		aggregator: aggregator,
		logger:     logger,
	}
}

// GetSystemHealth handles GET /health/system
func (hh *HealthHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	health := hh.aggregator.GetSystemHealth()

	w.Header().Set("Content-Type", "application/json")

	// Set status code based on health
	switch health.Status {
	case StatusHealthy:
		w.WriteHeader(http.StatusOK)
	case StatusDegraded:
		w.WriteHeader(http.StatusOK) // Still operational
	case StatusUnhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		hh.logger.Error("Failed to encode health response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetServiceHealth handles GET /health/service/{name}
func (hh *HealthHandler) GetServiceHealth(w http.ResponseWriter, r *http.Request) {
	// Extract service name from URL (this would depend on your router)
	serviceName := r.URL.Path[len("/health/service/"):]

	health, exists := hh.aggregator.GetServiceHealth(serviceName)
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch health.Status {
	case StatusHealthy:
		w.WriteHeader(http.StatusOK)
	case StatusDegraded:
		w.WriteHeader(http.StatusOK)
	case StatusUnhealthy:
		w.WriteHeader(http.StatusServiceUnavailable)
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		hh.logger.Error("Failed to encode service health response", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}