package chi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChiRouterHandleFunc(t *testing.T) {
	// Create a new router
	router := NewChiRouter()

	// Register a handler
	router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rec, req)

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", rec.Body.String())
	}
}

func TestChiRouterMethodNotAllowed(t *testing.T) {
	// Create a new router
	router := NewChiRouter()

	// Register a handler for GET method
	router.HandleFunc("GET", "/method-test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET response"))
	})

	// Create a test request with POST method
	req := httptest.NewRequest("POST", "/method-test", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rec, req)

	// Check the response - should be method not allowed
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestChiRouterMiddleware(t *testing.T) {
	// Create a new router
	router := NewChiRouter()

	// Add middleware
	middlewareCalled := false
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	})

	// Register a handler
	router.HandleFunc("GET", "/middleware-test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("middleware test"))
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/middleware-test", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rec, req)

	// Check that middleware was called
	if !middlewareCalled {
		t.Error("Expected middleware to be called")
	}

	// Check the response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestChiRouterMultipleMiddleware(t *testing.T) {
	// Create a new router
	router := NewChiRouter()

	// Add middleware
	middleware1Called := false
	middleware2Called := false

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware1Called = true
			next.ServeHTTP(w, r)
		})
	})

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware2Called = true
			next.ServeHTTP(w, r)
		})
	})

	// Register a handler
	router.HandleFunc("GET", "/multi-middleware", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/multi-middleware", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rec, req)

	// Check that both middleware were called
	if !middleware1Called {
		t.Error("Expected middleware1 to be called")
	}

	if !middleware2Called {
		t.Error("Expected middleware2 to be called")
	}
}