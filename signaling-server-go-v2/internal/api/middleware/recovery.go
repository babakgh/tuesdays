package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
)

// Recovery middleware recovers from panics
func Recovery(logger logging.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get request ID from header
					requestID := r.Header.Get(RequestIDHeader)

					// Log the panic with stacktrace
					logger.Error("Panic recovered",
						"request_id", requestID,
						"error", fmt.Sprintf("%v", err),
						"stack", string(debug.Stack()),
					)

					// Return 500 to the client
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"Internal Server Error"}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
