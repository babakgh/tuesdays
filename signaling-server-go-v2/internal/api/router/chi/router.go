package chi

import (
	"net/http"
	"strings"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/api/router"
)

// ChiRouter implements the Router interface using a basic http.ServeMux as a placeholder
type ChiRouter struct {
	router     *http.ServeMux
	middleware []func(http.Handler) http.Handler
}

// NewChiRouter creates a new router
func NewChiRouter() router.Router {
	return &ChiRouter{
		router:     http.NewServeMux(),
		middleware: []func(http.Handler) http.Handler{},
	}
}

// Handle registers a handler for a specific method and path
func (r *ChiRouter) Handle(method, path string, handler http.Handler) {
	// Create a method checking wrapper
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Only serve if method matches
		if req.Method == strings.ToUpper(method) {
			handler.ServeHTTP(w, req)
			return
		}
		// Method not allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	// Register with the mux
	r.router.Handle(path, r.wrapMiddleware(wrapped))
}

// HandleFunc registers a handler function for a specific method and path
func (r *ChiRouter) HandleFunc(method, path string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	r.Handle(method, path, http.HandlerFunc(handlerFunc))
}

// Use adds middleware to the router
func (r *ChiRouter) Use(middleware ...func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, middleware...)
}

// wrapMiddleware wraps a handler with all middleware
func (r *ChiRouter) wrapMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in reverse order so the first middleware is executed first
	for i := len(r.middleware) - 1; i >= 0; i-- {
		handler = r.middleware[i](handler)
	}
	return handler
}

// ServeHTTP implements the http.Handler interface
func (r *ChiRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}
