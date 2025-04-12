package middleware

import (
	"net/http"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/tracing"
)

// Tracing middleware adds tracing to HTTP requests
func Tracing(tracer tracing.Tracer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract trace context from the request headers if present
			propagatedCtx, _ := tracer.Extract(r.Header)
			if propagatedCtx != nil {
				ctx = propagatedCtx
			}

			// Start a new span for this request
			span := tracer.StartSpan("http.request",
				tracing.WithParent(ctx),
				tracing.WithAttributes(map[string]interface{}{
					"http.method": r.Method,
					"http.url":    r.URL.String(),
					"http.host":   r.Host,
				}),
			)
			defer span.End()

			// Get the request ID and add it as an attribute
			requestID := r.Header.Get(RequestIDHeader)
			if requestID != "" {
				span.SetAttribute("request_id", requestID)
			}

			// Create a response wrapper to capture the status code
			rw := &tracingResponseWriter{&responseWriter{w, http.StatusOK, 0}}

			// Inject the span context into the response headers for propagation
			_ = tracer.Inject(span.Context(), w.Header())

			// Call the next handler with the span context
			next.ServeHTTP(rw, r.WithContext(span.Context()))

			// Record the status code as an attribute
			span.SetAttribute("http.status_code", rw.Status())
		})
	}
}

// tracingResponseWriter is a wrapper for responseWriter that provides status method
type tracingResponseWriter struct {
	*responseWriter
}

// Status returns the HTTP status code
func (rw *tracingResponseWriter) Status() int {
	return rw.status
}
