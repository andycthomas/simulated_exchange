package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gin-gonic/gin"
	"simulated_exchange/pkg/config"
	"simulated_exchange/pkg/monitoring"
	"simulated_exchange/services/trading-api/internal/handlers"
	"simulated_exchange/services/trading-api/internal/middleware"
)

// Server represents the HTTP server
type Server struct {
	config            *config.Config
	orderHandler      *handlers.OrderHandler
	healthHandler     *handlers.HealthHandler
	metricsHandler    *handlers.MetricsHandler
	metricsCollector  *monitoring.MetricsCollector
	logger            *slog.Logger
	router            *gin.Engine
	httpServer        *http.Server
	startTime         time.Time
}

// NewServer creates a new HTTP server
func NewServer(
	config *config.Config,
	orderHandler *handlers.OrderHandler,
	healthHandler *handlers.HealthHandler,
	metricsHandler *handlers.MetricsHandler,
	metricsCollector *monitoring.MetricsCollector,
	logger *slog.Logger,
) *Server {
	// Set Gin mode based on environment
	if config.Service.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &Server{
		config:           config,
		orderHandler:     orderHandler,
		healthHandler:    healthHandler,
		metricsHandler:   metricsHandler,
		metricsCollector: metricsCollector,
		logger:           logger,
		startTime:        time.Now(),
	}

	server.setupRouter()
	return server
}

// setupRouter configures the Gin router with middleware and routes
func (s *Server) setupRouter() {
	s.router = gin.New()

	// Add middleware
	s.router.Use(middleware.LoggingMiddleware(s.logger))
	s.router.Use(middleware.RecoveryMiddleware(s.logger))
	s.router.Use(middleware.CORSMiddleware())
	s.router.Use(middleware.RequestIDMiddleware())

	// Health check routes (no authentication required)
	s.router.GET("/health", s.healthHandler.GetHealth)
	s.router.GET("/ready", s.healthHandler.GetReadiness)
	s.router.GET("/live", s.healthHandler.GetLiveness)

	// Prometheus metrics endpoint (for Prometheus scraping)
	s.router.GET("/metrics", gin.WrapH(s.metricsCollector.GetHandler()))

	// pprof debug endpoints (for profiling and flamegraphs)
	debug := s.router.Group("/debug/pprof")
	{
		debug.GET("/", gin.WrapF(pprof.Index))
		debug.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		debug.GET("/profile", gin.WrapF(pprof.Profile))
		debug.POST("/symbol", gin.WrapF(pprof.Symbol))
		debug.GET("/symbol", gin.WrapF(pprof.Symbol))
		debug.GET("/trace", gin.WrapF(pprof.Trace))
		debug.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		debug.GET("/block", gin.WrapH(pprof.Handler("block")))
		debug.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		debug.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		debug.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		debug.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}

	// API routes
	api := s.router.Group("/api")
	{
		// Add API middleware
		api.Use(middleware.RateLimitMiddleware())
		api.Use(middleware.ValidationMiddleware())

		// Health endpoints under /api
		api.GET("/health", s.healthHandler.GetHealth)

		// Metrics endpoint
		api.GET("/metrics", s.metricsHandler.GetMetrics)

		// Order endpoints
		orders := api.Group("/orders")
		{
			orders.POST("", s.orderHandler.PlaceOrder)
			orders.GET("", s.orderHandler.GetOrders) // Add GET all orders
			orders.GET("/:id", s.orderHandler.GetOrder)
			orders.DELETE("/:id", s.orderHandler.CancelOrder)
		}

		// Order book endpoints
		api.GET("/orderbook/:symbol", s.orderHandler.GetOrderBook)

		// User order endpoints
		api.GET("/users/:user_id/orders", s.orderHandler.GetUserOrders)
	}

	// Service info endpoint
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     s.config.Service.Name,
			"version":     s.config.Service.Version,
			"environment": s.config.Service.Environment,
			"status":      "running",
		})
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := s.config.Server.Host + ":" + s.config.Server.Port

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	s.logger.Info("Starting HTTP server",
		"addr", addr,
		"environment", s.config.Service.Environment,
	)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}

	s.logger.Info("Stopping HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown HTTP server gracefully", "error", err)
		return err
	}

	s.logger.Info("HTTP server stopped successfully")
	return nil
}

// GetRouter returns the Gin router for testing
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}