package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"simulated_exchange/internal/api/handlers"
	"simulated_exchange/internal/api/middleware"
)

// Server represents the HTTP server with dependency injection
type Server struct {
	router   *gin.Engine
	handlers *Handlers
	config   *Config
}

// Config holds server configuration
type Config struct {
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Handlers holds all HTTP handlers
type Handlers struct {
	OrderHandler   handlers.OrderHandler
	MetricsHandler handlers.MetricsHandler
	DemoHandler    handlers.DemoHandler
}

// Dependencies interface for dependency injection into server
type Dependencies interface {
	GetOrderService() handlers.OrderService
	GetMetricsService() handlers.MetricsService
	GetDemoController() handlers.DemoController
}

// NewServer creates a new server with dependency injection
func NewServer(deps Dependencies, config *Config) *Server {
	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Create handlers with dependency injection
	orderHandler := handlers.NewOrderHandler(deps.GetOrderService())
	metricsHandler := handlers.NewMetricsHandler(deps.GetMetricsService())
	demoHandler := handlers.NewDemoHandler(deps.GetDemoController())

	handlers := &Handlers{
		OrderHandler:   orderHandler,
		MetricsHandler: metricsHandler,
		DemoHandler:    demoHandler,
	}

	server := &Server{
		router:   router,
		handlers: handlers,
		config:   config,
	}

	// Setup middleware and routes
	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures all middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware (must be first)
	s.router.Use(middleware.ErrorHandlerMiddleware())

	// Logging middleware
	s.router.Use(middleware.LoggingMiddleware())

	// Security headers
	s.router.Use(middleware.SecurityHeadersMiddleware())

	// CORS middleware
	s.router.Use(middleware.CORSMiddleware())

	// Content type validation
	s.router.Use(middleware.ContentTypeMiddleware())

	// Rate limiting
	s.router.Use(middleware.RateLimitMiddleware())

	// Validation middleware
	s.router.Use(middleware.ValidationMiddleware())
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Dashboard route
	s.router.GET("/dashboard", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		data, err := os.ReadFile("web/templates/dashboard.html")
		if err != nil {
			c.String(http.StatusNotFound, "Dashboard not found")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Static file serving
	s.router.Static("/static", "./web/static")

	// Redirect root to dashboard for better UX
	s.router.GET("/", func(c *gin.Context) {
		// If Accept header indicates browser request, redirect to dashboard
		accept := c.GetHeader("Accept")
		if strings.Contains(accept, "text/html") {
			c.Redirect(http.StatusMovedPermanently, "/dashboard")
			return
		}
		// Otherwise return JSON API info
		c.JSON(http.StatusOK, gin.H{
			"service": "simulated-exchange-api",
			"version": "1.0.0",
			"status":  "running",
			"web_dashboard": "/dashboard",
		})
	})

	// API version group
	api := s.router.Group("/api")

	// Order endpoints
	orders := api.Group("/orders")
	{
		orders.POST("", s.handlers.OrderHandler.PlaceOrder)
		orders.GET("/:id", s.handlers.OrderHandler.GetOrder)
		orders.DELETE("/:id", s.handlers.OrderHandler.CancelOrder)
	}

	// Metrics endpoints
	api.GET("/metrics", s.handlers.MetricsHandler.GetMetrics)
	api.GET("/health", s.handlers.MetricsHandler.GetHealth)

	// Demo endpoints
	demo := s.router.Group("/demo")
	{
		// Load testing
		demo.POST("/load-test", s.handlers.DemoHandler.StartLoadTest)
		demo.GET("/load-test/status", s.handlers.DemoHandler.GetLoadTestStatus)
		demo.GET("/load-test/results", s.handlers.DemoHandler.GetLoadTestResults)
		demo.DELETE("/load-test", s.handlers.DemoHandler.StopLoadTest)

		// Chaos testing
		demo.POST("/chaos-test", s.handlers.DemoHandler.StartChaosTest)
		demo.GET("/chaos-test/status", s.handlers.DemoHandler.GetChaosTestStatus)
		demo.DELETE("/chaos-test", s.handlers.DemoHandler.StopChaosTest)

		// System management
		demo.POST("/reset", s.handlers.DemoHandler.ResetSystem)
		demo.GET("/status", s.handlers.DemoHandler.GetSystemStatus)
	}

	// Root health check
	s.router.GET("/health", s.handlers.MetricsHandler.GetHealth)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := ":" + s.config.Port
	log.Printf("Starting server on port %s", s.config.Port)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	return server.ListenAndServe()
}

// StartWithContext starts the HTTP server with graceful shutdown support
func (s *Server) StartWithContext(ctx context.Context) error {
	addr := ":" + s.config.Port
	log.Printf("Starting server on port %s", s.config.Port)

	server := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Println("Shutting down server...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	}
}

// GetRouter returns the gin router for testing
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:         "8080",
		Environment:  "development",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
}

// DependencyContainer implements the Dependencies interface
type DependencyContainer struct {
	orderService   handlers.OrderService
	metricsService handlers.MetricsService
	demoController handlers.DemoController
}

// NewDependencyContainer creates a new dependency container
func NewDependencyContainer(orderService handlers.OrderService, metricsService handlers.MetricsService, demoController handlers.DemoController) Dependencies {
	return &DependencyContainer{
		orderService:   orderService,
		metricsService: metricsService,
		demoController: demoController,
	}
}

// GetOrderService returns the order service
func (dc *DependencyContainer) GetOrderService() handlers.OrderService {
	return dc.orderService
}

// GetMetricsService returns the metrics service
func (dc *DependencyContainer) GetMetricsService() handlers.MetricsService {
	return dc.metricsService
}

// GetDemoController returns the demo controller
func (dc *DependencyContainer) GetDemoController() handlers.DemoController {
	return dc.demoController
}