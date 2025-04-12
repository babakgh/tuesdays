package router

import (
	"net/http"
)

// Router interface for abstracting HTTP routing implementations
type Router interface {
	Handle(method, path string, handler http.Handler)
	HandleFunc(method, path string, handlerFunc func(http.ResponseWriter, *http.Request))
	Use(middleware ...func(http.Handler) http.Handler)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}
