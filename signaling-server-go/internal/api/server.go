package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tuesdays/signaling-server-go/config"
	"github.com/tuesdays/signaling-server-go/internal/api/middleware"
	"go.uber.org/zap"
)

type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	server *http.Server
	router *gin.Engine
}

func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.TracingMiddleware())

	// Setup routes
	setupRoutes(router, cfg)

	return &Server{
		cfg:    cfg,
		logger: logger,
		router: router,
	}
}

func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func setupRoutes(router *gin.Engine, cfg *config.Config) {
	// Health endpoints
	router.GET(cfg.Health.LivePath, handleLive)
	router.GET(cfg.Health.ReadyPath, handleReady)

	// Metrics endpoint
	if cfg.Metrics.Enabled {
		router.GET(cfg.Metrics.Path, handleMetrics)
	}
}

func handleLive(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleReady(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleMetrics(c *gin.Context) {
	// TODO: Implement Prometheus metrics handler
	c.Status(http.StatusNotImplemented)
}
