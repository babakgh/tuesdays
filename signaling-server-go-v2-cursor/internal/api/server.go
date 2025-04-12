package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tuesdays/signaling-server-go-v2/internal/config"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	router     *mux.Router
	config     *config.Config
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) *Server {
	router := mux.NewRouter()

	// Setup routes
	router.HandleFunc("/health/live", handleLive).Methods(http.MethodGet)
	router.HandleFunc("/health/ready", handleReady).Methods(http.MethodGet)

	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{
		httpServer: server,
		router:     router,
		config:     cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// handleLive handles the liveness probe endpoint
func handleLive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReady handles the readiness probe endpoint
func handleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
