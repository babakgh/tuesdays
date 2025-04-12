package middleware

import (
	"net/http"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
)

// RequestIDHeader is the header that contains the request ID
const RequestIDHeader = "X-Request-ID"

// Logging middleware logs request information
func Logging(logger logging.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get request ID from context (added by the RequestID middleware)
			requestID := r.Header.Get(RequestIDHeader)
			ctxLogger := logger.With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			ctxLogger.Info("Request started")

			// Create a response wrapper to capture the status code
			rw := &responseWriter{w, http.StatusOK, 0}

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Log the response details
			ctxLogger.Info("Request completed",
				"status", rw.status,
				"size", rw.size,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

// responseWriter wraps the standard http.ResponseWriter to capture the status code and response size
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write captures the response size
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
