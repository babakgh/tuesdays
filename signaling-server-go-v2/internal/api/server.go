package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/api/handlers/health"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/api/middleware"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/api/router"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/api/websocket"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/metrics"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/tracing"
)

// Server represents the HTTP server for the signaling service
type Server struct {
	cfg           *config.Config
	router        router.Router
	httpServer    *http.Server
	logger        logging.Logger
	metrics       *metrics.Metrics
	tracer        tracing.Tracer
	wsHandler     websocket.WebSocketHandler
	healthHandler *health.Handler
}

// NewServer creates a new server with the given configuration
func NewServer(
	cfg *config.Config,
	router router.Router,
	logger logging.Logger,
	metrics *metrics.Metrics,
	tracer tracing.Tracer,
	wsHandler websocket.WebSocketHandler,
) *Server {
	s := &Server{
		cfg:       cfg,
		router:    router,
		logger:    logger.With("component", "server"),
		metrics:   metrics,
		tracer:    tracer,
		wsHandler: wsHandler,
	}

	// Create and configure the HTTP server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      s.router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Create health handler
	s.healthHandler = health.NewHandler(logger)

	// Register routes and middleware
	s.registerMiddleware()
	s.registerRoutes()

	return s
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting server", "address", s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("Failed to start server", "error", err)
		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server")

	// Create a new context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// Shutdown the HTTP server
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Failed to shutdown server gracefully", "error", err)
		return err
	}

	return nil
}

// registerMiddleware registers middleware for the server
func (s *Server) registerMiddleware() {
	// Add core middleware
	s.router.Use(middleware.Recovery(s.logger))
	s.router.Use(middleware.Logging(s.logger))

	// Add metrics middleware if enabled
	if s.cfg.Metrics.Enabled {
		s.router.Use(middleware.Metrics(s.metrics))
	}

	// Add tracing middleware if enabled
	if s.cfg.Tracing.Enabled {
		s.router.Use(middleware.Tracing(s.tracer))
	}
}

// registerRoutes registers routes for the server
func (s *Server) registerRoutes() {
	// Register health check endpoints
	s.router.HandleFunc("GET", s.cfg.Monitoring.LivenessPath, s.healthHandler.LiveHandler)
	s.router.HandleFunc("GET", s.cfg.Monitoring.ReadinessPath, s.healthHandler.ReadyHandler)

	// Register WebSocket endpoint
	s.router.HandleFunc("GET", s.cfg.WebSocket.Path, s.wsHandler.HandleConnection)

	// Register metrics endpoint if enabled
	if s.cfg.Metrics.Enabled {
		s.router.Handle("GET", s.cfg.Metrics.Path, metrics.MetricsHandler())
	}
}
