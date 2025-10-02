package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"simulated_exchange/pkg/config"
	"simulated_exchange/services/order-flow-simulator/internal/handlers"
)

// Server represents the HTTP server for order flow simulator
type Server struct {
	config        *config.Config
	healthHandler *handlers.HealthHandler
	flowHandler   *handlers.FlowHandler
	logger        *slog.Logger
	router        *gin.Engine
	httpServer    *http.Server
}

// NewServer creates a new HTTP server
func NewServer(
	config *config.Config,
	healthHandler *handlers.HealthHandler,
	flowHandler *handlers.FlowHandler,
	logger *slog.Logger,
) *Server {
	// Set Gin mode based on environment
	if config.Service.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &Server{
		config:        config,
		healthHandler: healthHandler,
		flowHandler:   flowHandler,
		logger:        logger,
	}

	server.setupRouter()
	return server
}

// setupRouter configures the Gin router with middleware and routes
func (s *Server) setupRouter() {
	s.router = gin.New()

	// Add middleware
	s.router.Use(loggingMiddleware(s.logger))
	s.router.Use(recoveryMiddleware(s.logger))
	s.router.Use(corsMiddleware())

	// Health check routes
	s.router.GET("/health", s.healthHandler.GetHealth)
	s.router.GET("/ready", s.healthHandler.GetReadiness)
	s.router.GET("/live", s.healthHandler.GetLiveness)

	// API routes
	api := s.router.Group("/api")
	{
		// Flow simulation endpoints
		api.GET("/status", s.flowHandler.GetStatus)
		api.POST("/order-rate", s.flowHandler.SetOrderRate)
		api.POST("/volatility", s.flowHandler.SetVolatility)

		// User simulation endpoints
		api.GET("/users", s.flowHandler.GetUserSessions)

		// Market state endpoints
		api.GET("/market/:symbol", s.flowHandler.GetMarketState)

		// Metrics and statistics endpoints
		api.GET("/metrics", s.flowHandler.GetSimulationMetrics)
		api.GET("/symbols/:symbol/stats", s.flowHandler.GetSymbolStats)
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

// Middleware functions

func loggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration,
			"client_ip", c.ClientIP(),
		)
	}
}

func recoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "INTERNAL_SERVER_ERROR",
						"message": "Internal server error",
					},
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}