package middleware

import (
	"net/http"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/metrics"
)

// Metrics middleware records metrics for HTTP requests
func Metrics(metrics *metrics.Metrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response wrapper to capture metrics
			rw := &metricsResponseWriter{&responseWriter{w, http.StatusOK, 0}}

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Record metrics
			if metrics != nil {
				metrics.RecordHTTPRequest(
					r.Method,
					r.URL.Path,
					rw.Status(),
					time.Since(start),
					rw.Size(),
				)
			}
		})
	}
}

// metricsResponseWriter is a wrapper for responseWriter that provides status and size methods
type metricsResponseWriter struct {
	*responseWriter
}

// Status returns the HTTP status code
func (rw *metricsResponseWriter) Status() int {
	return rw.status
}

// Size returns the response size
func (rw *metricsResponseWriter) Size() int {
	return rw.size
}
