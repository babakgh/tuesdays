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

Suggested file structure for the project:

```
signaling-server-go/
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
   - Implement a clean HTTP server in Go
   - Create router setup with basic health check endpoints
   - Implement graceful startup and shutdown patterns
   - Add signal handling (SIGTERM, SIGINT)

3. **Observability Infrastructure**:
   - Integrate a structured logging library (e.g., zap, zerolog)
   - Set up Prometheus metrics collection
   - Implement OpenTelemetry tracing
   - Create middleware for request logging, tracing, and metrics

4. **Basic Health and Debug Endpoints**:
   - Implement `/health/live` and `/health/ready` endpoints
   - Add a debug endpoint with runtime statistics
   - Create metric endpoints for Prometheus scraping

5. **Error Handling and Recovery**:
   - Implement panic recovery middleware
   - Create consistent error response patterns
   - Add request ID generation and propagation

6. **Basic Testing**:
   - Write unit tests for core functionality
   - Create integration tests for startup/shutdown
   - Add benchmark tests for critical paths

Remember that this implementation should focus on creating a production-ready foundation that can be extended with additional functionality in future iterations.
