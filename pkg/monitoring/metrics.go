package monitoring

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector provides centralized metrics collection
type MetricsCollector struct {
	registry *prometheus.Registry
	logger   *slog.Logger

	// Common metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// Trading specific metrics
	ordersTotal          *prometheus.CounterVec
	orderProcessingTime  *prometheus.HistogramVec
	tradesTotal          *prometheus.CounterVec
	activeOrders         *prometheus.GaugeVec
	orderBookDepth       *prometheus.GaugeVec

	// Market simulation metrics
	priceUpdatesTotal    *prometheus.CounterVec
	currentPrices        *prometheus.GaugeVec
	marketVolatility     *prometheus.GaugeVec
	simulationErrors     *prometheus.CounterVec

	// System metrics
	serviceUptime        *prometheus.GaugeVec
	serviceHealth        *prometheus.GaugeVec
	redisConnections     *prometheus.GaugeVec
	postgresConnections  *prometheus.GaugeVec

	// Event bus metrics
	eventsPublished      *prometheus.CounterVec
	eventsProcessed      *prometheus.CounterVec
	eventProcessingTime  *prometheus.HistogramVec

	startTime time.Time
	mutex     sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *slog.Logger) *MetricsCollector {
	registry := prometheus.NewRegistry()

	mc := &MetricsCollector{
		registry:  registry,
		logger:    logger,
		startTime: time.Now(),
	}

	mc.initializeMetrics()
	mc.registerMetrics()

	return mc
}

// initializeMetrics creates all metric definitions
func (mc *MetricsCollector) initializeMetrics() {
	// HTTP metrics
	mc.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status_code"},
	)

	mc.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "endpoint"},
	)

	mc.httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"service", "method", "endpoint"},
	)

	// Trading metrics
	mc.ordersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_total",
			Help: "Total number of orders processed",
		},
		[]string{"service", "type", "side", "status"},
	)

	mc.orderProcessingTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_processing_duration_seconds",
			Help:    "Order processing duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"service", "type", "side"},
	)

	mc.tradesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trades_total",
			Help: "Total number of trades executed",
		},
		[]string{"service", "symbol"},
	)

	mc.activeOrders = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_orders",
			Help: "Current number of active orders",
		},
		[]string{"service", "symbol", "side"},
	)

	mc.orderBookDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "order_book_depth",
			Help: "Current order book depth",
		},
		[]string{"service", "symbol", "side"},
	)

	// Market simulation metrics
	mc.priceUpdatesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "price_updates_total",
			Help: "Total number of price updates generated",
		},
		[]string{"service", "symbol"},
	)

	mc.currentPrices = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "current_prices",
			Help: "Current asset prices",
		},
		[]string{"service", "symbol"},
	)

	mc.marketVolatility = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "market_volatility",
			Help: "Current market volatility indicators",
		},
		[]string{"service", "symbol"},
	)

	mc.simulationErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "simulation_errors_total",
			Help: "Total number of simulation errors",
		},
		[]string{"service", "type"},
	)

	// System metrics
	mc.serviceUptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_uptime_seconds",
			Help: "Service uptime in seconds",
		},
		[]string{"service"},
	)

	mc.serviceHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_health",
			Help: "Service health status (1=healthy, 0=unhealthy)",
		},
		[]string{"service"},
	)

	mc.redisConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_connections",
			Help: "Current Redis connections",
		},
		[]string{"service", "type"},
	)

	mc.postgresConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "postgres_connections",
			Help: "Current PostgreSQL connections",
		},
		[]string{"service", "type"},
	)

	// Event bus metrics
	mc.eventsPublished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_published_total",
			Help: "Total number of events published",
		},
		[]string{"service", "event_type"},
	)

	mc.eventsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_processed_total",
			Help: "Total number of events processed",
		},
		[]string{"service", "event_type", "status"},
	)

	mc.eventProcessingTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "event_processing_duration_seconds",
			Help:    "Event processing duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"service", "event_type"},
	)
}

// registerMetrics registers all metrics with the registry
func (mc *MetricsCollector) registerMetrics() {
	mc.registry.MustRegister(
		mc.httpRequestsTotal,
		mc.httpRequestDuration,
		mc.httpRequestsInFlight,
		mc.ordersTotal,
		mc.orderProcessingTime,
		mc.tradesTotal,
		mc.activeOrders,
		mc.orderBookDepth,
		mc.priceUpdatesTotal,
		mc.currentPrices,
		mc.marketVolatility,
		mc.simulationErrors,
		mc.serviceUptime,
		mc.serviceHealth,
		mc.redisConnections,
		mc.postgresConnections,
		mc.eventsPublished,
		mc.eventsProcessed,
		mc.eventProcessingTime,
	)
}

// GetRegistry returns the Prometheus registry
func (mc *MetricsCollector) GetRegistry() *prometheus.Registry {
	return mc.registry
}

// GetHandler returns the Prometheus HTTP handler
func (mc *MetricsCollector) GetHandler() http.Handler {
	return promhttp.HandlerFor(mc.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// HTTP Metrics methods
func (mc *MetricsCollector) RecordHTTPRequest(service, method, endpoint string, statusCode int, duration time.Duration) {
	mc.httpRequestsTotal.WithLabelValues(service, method, endpoint, strconv.Itoa(statusCode)).Inc()
	mc.httpRequestDuration.WithLabelValues(service, method, endpoint).Observe(duration.Seconds())
}

func (mc *MetricsCollector) HTTPRequestStart(service, method, endpoint string) {
	mc.httpRequestsInFlight.WithLabelValues(service, method, endpoint).Inc()
}

func (mc *MetricsCollector) HTTPRequestEnd(service, method, endpoint string) {
	mc.httpRequestsInFlight.WithLabelValues(service, method, endpoint).Dec()
}

// Trading Metrics methods
func (mc *MetricsCollector) RecordOrder(service, orderType, side, status string, processingTime time.Duration) {
	mc.ordersTotal.WithLabelValues(service, orderType, side, status).Inc()
	mc.orderProcessingTime.WithLabelValues(service, orderType, side).Observe(processingTime.Seconds())
}

func (mc *MetricsCollector) RecordTrade(service, symbol string) {
	mc.tradesTotal.WithLabelValues(service, symbol).Inc()
}

func (mc *MetricsCollector) SetActiveOrders(service, symbol, side string, count float64) {
	mc.activeOrders.WithLabelValues(service, symbol, side).Set(count)
}

func (mc *MetricsCollector) SetOrderBookDepth(service, symbol, side string, depth float64) {
	mc.orderBookDepth.WithLabelValues(service, symbol, side).Set(depth)
}

// Market Simulation Metrics methods
func (mc *MetricsCollector) RecordPriceUpdate(service, symbol string) {
	mc.priceUpdatesTotal.WithLabelValues(service, symbol).Inc()
}

func (mc *MetricsCollector) SetCurrentPrice(service, symbol string, price float64) {
	mc.currentPrices.WithLabelValues(service, symbol).Set(price)
}

func (mc *MetricsCollector) SetMarketVolatility(service, symbol string, volatility float64) {
	mc.marketVolatility.WithLabelValues(service, symbol).Set(volatility)
}

func (mc *MetricsCollector) RecordSimulationError(service, errorType string) {
	mc.simulationErrors.WithLabelValues(service, errorType).Inc()
}

// System Metrics methods
func (mc *MetricsCollector) UpdateServiceUptime(service string) {
	mc.serviceUptime.WithLabelValues(service).Set(time.Since(mc.startTime).Seconds())
}

func (mc *MetricsCollector) SetServiceHealth(service string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	mc.serviceHealth.WithLabelValues(service).Set(value)
}

func (mc *MetricsCollector) SetRedisConnections(service, connectionType string, count float64) {
	mc.redisConnections.WithLabelValues(service, connectionType).Set(count)
}

func (mc *MetricsCollector) SetPostgresConnections(service, connectionType string, count float64) {
	mc.postgresConnections.WithLabelValues(service, connectionType).Set(count)
}

// Event Bus Metrics methods
func (mc *MetricsCollector) RecordEventPublished(service, eventType string) {
	mc.eventsPublished.WithLabelValues(service, eventType).Inc()
}

func (mc *MetricsCollector) RecordEventProcessed(service, eventType, status string, processingTime time.Duration) {
	mc.eventsProcessed.WithLabelValues(service, eventType, status).Inc()
	mc.eventProcessingTime.WithLabelValues(service, eventType).Observe(processingTime.Seconds())
}

// MetricsMiddleware creates HTTP middleware for automatic metrics collection
func (mc *MetricsCollector) MetricsMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Track request start
			mc.HTTPRequestStart(serviceName, r.Method, r.URL.Path)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start)
			mc.RecordHTTPRequest(serviceName, r.Method, r.URL.Path, wrapped.statusCode, duration)
			mc.HTTPRequestEnd(serviceName, r.Method, r.URL.Path)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// PeriodicMetricsUpdater updates certain metrics periodically
type PeriodicMetricsUpdater struct {
	collector *MetricsCollector
	logger    *slog.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewPeriodicMetricsUpdater creates a new periodic metrics updater
func NewPeriodicMetricsUpdater(collector *MetricsCollector, logger *slog.Logger) *PeriodicMetricsUpdater {
	ctx, cancel := context.WithCancel(context.Background())

	return &PeriodicMetricsUpdater{
		collector: collector,
		logger:    logger,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins periodic metrics updates
func (pmu *PeriodicMetricsUpdater) Start(serviceName string) {
	pmu.wg.Add(1)
	go pmu.updateLoop(serviceName)
}

// Stop stops periodic metrics updates
func (pmu *PeriodicMetricsUpdater) Stop() {
	pmu.cancel()
	pmu.wg.Wait()
}

// updateLoop periodically updates metrics
func (pmu *PeriodicMetricsUpdater) updateLoop(serviceName string) {
	defer pmu.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-pmu.ctx.Done():
			return
		case <-ticker.C:
			pmu.collector.UpdateServiceUptime(serviceName)
		}
	}
}