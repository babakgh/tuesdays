package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/api"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/api/router/chi"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/api/websocket/gorilla"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging/kitlog"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/metrics"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/tracing/otel"
)

func main() {
	// Get the configuration path from environment variables
	configPath := config.GetConfigPath()

	// Load the configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := kitlog.NewKitLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Set the default logger instance
	logging.SetDefaultLogger(logger)

	logger.Info("Starting signaling server", "version", "1.0.0")

	// Initialize tracer
	logger.Info("Initializing tracer")
	provider, err := otel.Initialize(cfg.Tracing)
	if err != nil {
		logger.Error("Failed to initialize tracer", "error", err)
		os.Exit(1)
	}

	// If tracing is enabled, ensure we shut it down properly
	if provider != nil {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := otel.Shutdown(ctx, provider); err != nil {
				logger.Error("Failed to shutdown tracer", "error", err)
			}
		}()
	}

	tracer, err := otel.NewOTelTracer(cfg.Tracing)
	if err != nil {
		logger.Error("Failed to create tracer", "error", err)
		os.Exit(1)
	}

	// Initialize metrics
	logger.Info("Initializing metrics")
	m := metrics.NewMetrics(cfg.Metrics)

	// Create router
	router := chi.NewChiRouter()

	// Create WebSocket handler
	wsHandler := gorilla.NewHandler(cfg.WebSocket, logger, m, tracer)

	// Create server
	server := api.NewServer(cfg, router, logger, m, tracer, wsHandler)

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	logger.Info("Starting server")
	go func() {
		if err := server.Start(); err != nil {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for signal
	sig := <-sigCh
	logger.Info("Received signal", "signal", sig.String())

	// Create a context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// Perform graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server gracefully", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped")
}
