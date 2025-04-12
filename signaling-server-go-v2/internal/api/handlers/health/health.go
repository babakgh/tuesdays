package health

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
)

// Status represents the status of a health check
type Status string

const (
	// StatusUp indicates the service is up and running
	StatusUp Status = "UP"

	// StatusDown indicates the service is down
	StatusDown Status = "DOWN"
)

// HealthResponse represents the response from a health check endpoint
type HealthResponse struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Checks    map[string]CheckStatus `json:"checks,omitempty"`
}

// CheckStatus represents the status of a specific health check
type CheckStatus struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

// Handler is the health check handler
type Handler struct {
	logger      logging.Logger
	checks      map[string]func() (Status, string)
	readyChecks map[string]func() (Status, string)
}

// NewHandler creates a new health check handler
func NewHandler(logger logging.Logger) *Handler {
	return &Handler{
		logger:      logger.With("component", "health"),
		checks:      make(map[string]func() (Status, string)),
		readyChecks: make(map[string]func() (Status, string)),
	}
}

// AddLivenessCheck adds a check to the liveness endpoint
func (h *Handler) AddLivenessCheck(name string, check func() (Status, string)) {
	h.checks[name] = check
}

// AddReadinessCheck adds a check to the readiness endpoint
func (h *Handler) AddReadinessCheck(name string, check func() (Status, string)) {
	h.readyChecks[name] = check
}

// LiveHandler handles liveness check requests
func (h *Handler) LiveHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling liveness check")

	resp := HealthResponse{
		Status:    StatusUp,
		Timestamp: time.Now().UTC(),
		Checks:    make(map[string]CheckStatus),
	}

	for name, check := range h.checks {
		status, message := check()
		resp.Checks[name] = CheckStatus{
			Status:  status,
			Message: message,
		}

		if status == StatusDown {
			resp.Status = StatusDown
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if resp.Status == StatusDown {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode health response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ReadyHandler handles readiness check requests
func (h *Handler) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Handling readiness check")

	resp := HealthResponse{
		Status:    StatusUp,
		Timestamp: time.Now().UTC(),
		Checks:    make(map[string]CheckStatus),
	}

	// First run all liveness checks
	for name, check := range h.checks {
		status, message := check()
		resp.Checks[name] = CheckStatus{
			Status:  status,
			Message: message,
		}

		if status == StatusDown {
			resp.Status = StatusDown
		}
	}

	// Then run all readiness-specific checks
	for name, check := range h.readyChecks {
		status, message := check()
		resp.Checks[name] = CheckStatus{
			Status:  status,
			Message: message,
		}

		if status == StatusDown {
			resp.Status = StatusDown
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if resp.Status == StatusDown {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to encode health response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
