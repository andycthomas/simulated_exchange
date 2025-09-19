package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"simulated_exchange/internal/config"
)

// HealthService manages health checks for all application components
type HealthService struct {
	config config.HealthConfig
	logger *slog.Logger

	checks     map[string]HealthCheck
	checksMux  sync.RWMutex

	lastStatus HealthStatus
	statusMux  sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

// NewHealthService creates a new health service
func NewHealthService(cfg config.HealthConfig, logger *slog.Logger) *HealthService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &HealthService{
		config:     cfg,
		logger:     logger,
		checks:     make(map[string]HealthCheck),
		lastStatus: HealthStatus{
			Status:    "starting",
			Timestamp: time.Now(),
			Checks:    make(map[string]string),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	if cfg.Enabled {
		go service.runPeriodicChecks()
	}

	return service
}

// RegisterCheck registers a new health check
func (h *HealthService) RegisterCheck(name string, check HealthCheck) {
	h.checksMux.Lock()
	defer h.checksMux.Unlock()

	h.checks[name] = check
	h.logger.Info("Health check registered", "name", name)
}

// Check performs all registered health checks
func (h *HealthService) Check(ctx context.Context) HealthStatus {
	start := time.Now()

	h.checksMux.RLock()
	checks := make(map[string]HealthCheck, len(h.checks))
	for name, check := range h.checks {
		checks[name] = check
	}
	h.checksMux.RUnlock()

	status := HealthStatus{
		Status:    "healthy",
		Timestamp: start,
		Checks:    make(map[string]string),
	}

	// Create timeout context for health checks
	checkCtx, cancel := context.WithTimeout(ctx, h.config.Timeout)
	defer cancel()

	// Run all health checks
	for name, check := range checks {
		if err := check(checkCtx); err != nil {
			status.Status = "unhealthy"
			status.Checks[name] = "failed: " + err.Error()
			status.ErrorCount++
			status.LastError = err.Error()
			h.logger.Warn("Health check failed", "name", name, "error", err)
		} else {
			status.Checks[name] = "healthy"
		}
	}

	status.Duration = time.Since(start)

	// Update last status
	h.statusMux.Lock()
	h.lastStatus = status
	h.statusMux.Unlock()

	return status
}

// GetStatus returns the last health check status
func (h *HealthService) GetStatus() HealthStatus {
	h.statusMux.RLock()
	defer h.statusMux.RUnlock()
	return h.lastStatus
}

// runPeriodicChecks runs health checks periodically
func (h *HealthService) runPeriodicChecks() {
	ticker := time.NewTicker(h.config.CheckInterval)
	defer ticker.Stop()

	h.logger.Info("Starting periodic health checks", "interval", h.config.CheckInterval)

	for {
		select {
		case <-h.ctx.Done():
			h.logger.Info("Stopping periodic health checks")
			return
		case <-ticker.C:
			status := h.Check(h.ctx)
			h.logger.Debug("Health check completed",
				"status", status.Status,
				"duration", status.Duration,
				"checks", len(status.Checks),
				"errors", status.ErrorCount,
			)
		}
	}
}

// Shutdown stops the health service
func (h *HealthService) Shutdown() {
	h.logger.Info("Shutting down health service")
	h.cancel()
}