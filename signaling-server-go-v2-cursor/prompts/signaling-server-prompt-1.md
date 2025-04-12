# Prompt 1: Production-Ready Go Signaling Server Skeleton

## Problem

Create the foundation for a WebRTC signaling server in Go with production-quality infrastructure. The server should implement proper startup and graceful shutdown patterns, along with complete observability through traces, metrics, and logs.

This server will form the backbone of our WebRTC signaling implementation, so it needs to be robust, observable, and maintainable.

## Supporting Information

The signaling server will eventually handle WebSocket connections and facilitate the WebRTC signaling process between clients. For now, focus on creating a solid foundation with the following characteristics:

- **Architecture**: Follow clean architecture principles with clear separation of concerns
- **Configuration**: Support environment variables and config files
- **Graceful Shutdown**: Handle OS signals and ensure clean termination
- **Observability**:
  - Structured logging with contextual information
  - Prometheus metrics for monitoring
  - Distributed tracing with OpenTelemetry
- **Health Checks**: Implement readiness and liveness endpoints
- **Error Handling**: Robust error handling patterns

Reference the provided WebRTC Signaling Server System Design document for the overall architecture. The server will eventually need to support WebSockets, Redis Pub/Sub, and room management, but those specific implementations will be added in future iterations.

Suggested key interfaces for the adapter patterns:

```go
// Logger interface for abstracting logging implementations
type Logger interface {
    Debug(msg string, keyvals ...interface{})
    Info(msg string, keyvals ...interface{})
    Warn(msg string, keyvals ...interface{})
    Error(msg string, keyvals ...interface{})
    With(keyvals ...interface{}) Logger
}

// Router interface for abstracting HTTP routing implementations
type Router interface {
    Handle(method, path string, handler http.Handler)
    HandleFunc(method, path string, handlerFunc func(http.ResponseWriter, *http.Request))
    Use(middleware ...func(http.Handler) http.Handler)
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// WebSocket interfaces for abstracting WebSocket implementations
type WebSocketHandler interface {
    HandleConnection(w http.ResponseWriter, r *http.Request)
    BroadcastMessage(message []byte) error
    SendMessage(clientID string, message []byte) error
    CloseConnection(clientID string) error
}

type WebSocketConnection interface {
    ReadMessage() (messageType int, p []byte, err error)
    WriteMessage(messageType int, data []byte) error
    Close() error
    SetReadDeadline(t time.Time) error
    SetWriteDeadline(t time.Time) error
}

// Tracer interfaces for abstracting tracing implementations
type Tracer interface {
    StartSpan(name string, opts ...SpanOption) Span
    Inject(ctx context.Context, carrier interface{}) error
    Extract(carrier interface{}) (context.Context, error)
}

type Span interface {
    End()
    SetAttribute(key string, value interface{})
    AddEvent(name string, attributes map[string]interface{})
    RecordError(err error)
    Context() context.Context
}

type SpanOption func(*SpanOptions)
type SpanOptions struct {
    Attributes map[string]interface{}
    Parent     context.Context
}
```

Suggested file structure for the project:

```
signaling-server-go-v2/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── config/
│   ├── config.go                # Configuration structures and loading
│   └── default.yaml             # Default configuration
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP request handlers
│   │   │   └── health.go        # Health check endpoints
│   │   ├── middleware/          # HTTP middleware
│   │   │   ├── logging.go       # Request logging
│   │   │   ├── metrics.go       # Prometheus metrics
│   │   │   ├── recovery.go      # Panic recovery
│   │   │   └── tracing.go       # Request tracing
│   │   ├── router.go            # HTTP router setup
│   │   └── server.go            # HTTP server implementation
│   ├── domain/                  # Domain models and interfaces
│   ├── observability/
│   │   ├── logging/             # Logging configuration
│   │   ├── metrics/             # Metrics configuration
│   │   └── tracing/             # Tracing configuration
│   └── service/                 # Business logic implementations
├── pkg/                         # Public packages that could be imported
│   └── telemetry/               # Telemetry helpers
├── docker/                      # Docker-related files
│   ├── Dockerfile               # Main application Dockerfile
│   └── prometheus/              # Prometheus configuration files
├── docker-compose.yml           # Docker Compose for local development
├── .gitignore                   # Git ignore file
├── go.mod                       # Go module file
├── go.sum                       # Go module checksum
├── Makefile                     # Build and development commands
└── README.md                    # Project documentation
```

This structure follows clean architecture principles and separation of concerns, making the codebase maintainable and extendable.

## Steps to Complete

1. **Project Structure Setup**:
   - Create a Go module with appropriate directory structure
   - Set up dependency management with Go modules
   - Implement a configuration loader supporting environment variables and config files

2. **Server Core Implementation**:
   - Implement a router interface with adapter pattern:
     - Create a common Router interface
     - Implement an adapter for Gorilla mux as the default router
     - This allows for swapping router implementations without changing application code
   - Implement a WebSocket interface with adapter pattern:
     - Create a common WebSocket interface
     - Implement an adapter for Gorilla WebSocket as the default implementation
     - This allows for swapping WebSocket implementations without changing application code
   - Create router setup with basic health check endpoints
   - Implement graceful startup and shutdown patterns
   - Add signal handling (SIGTERM, SIGINT)

3. **Observability Infrastructure**:
   - Implement a logger interface with adapter pattern:
     - Create a common Logger interface
     - Implement an adapter for go-kit/log as the default structured logging library
     - This allows for swapping logging implementations without changing application code
   - Set up Prometheus metrics collection
   - Implement a tracer interface with adapter pattern:
     - Create a common Tracer and Span interfaces
     - Implement an adapter for OpenTelemetry as the default tracing implementation
     - This allows for swapping tracing implementations without changing application code
   - Create middleware for request logging, tracing, and metrics

4. **Health and Monitoring Endpoints**:
   - Implement health check endpoints:
     - `/health/live` for liveness probes (is the server running?)
     - `/health/ready` for readiness probes (is the server ready to accept traffic?)
   - Add a debug endpoint with runtime statistics
   - Create metric endpoints for Prometheus scraping

5. **Error Handling and Recovery**:
   - Implement panic recovery middleware
   - Create consistent error response patterns
   - Add request ID generation and propagation

6. **Testing Infrastructure**:
   - Write unit tests for core server infrastructure:
     - Configuration loading and validation
     - Graceful startup and shutdown sequences
     - Signal handling (SIGTERM, SIGINT)
     - Health check endpoints functionality
   - Create tests for all adapter implementations:
     - Logger adapter tests
     - Router adapter tests
     - WebSocket adapter tests
     - Tracer adapter tests
   - Implement integration tests for observability:
     - Verify metrics collection works properly
     - Ensure trace propagation functions correctly
     - Validate structured logging output
   - Add benchmark tests for performance-critical paths

7. **Development Environment Setup**:
   - Create a Docker Compose file for local development
   - Set up Prometheus for metrics collection locally
   - Configure environment-specific settings for development

Remember that this implementation should focus on creating a production-ready foundation that can be extended with additional functionality in future iterations.
