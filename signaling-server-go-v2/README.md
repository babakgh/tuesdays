# WebRTC Signaling Server

A production-ready WebRTC signaling server implementation in Go with comprehensive observability, graceful shutdown, and clean architecture.

## Features

- **Clean Architecture**: Clear separation of concerns with domain-driven design
- **Configuration Management**: Support for environment variables and YAML config files
- **Graceful Shutdown**: Proper handling of OS signals and connection termination
- **Full Observability**:
  - Structured logging with go-kit/log
  - Prometheus metrics for monitoring
  - OpenTelemetry distributed tracing
- **Health Checks**: Readiness and liveness probes for Kubernetes integration
- **WebSocket Support**: Efficient WebSocket connection management
- **Error Handling**: Robust error handling and recovery patterns

## Project Structure

```
├── cmd/
│   └── server/           # Application entry point
├── config/               # Configuration handling
├── internal/
│   ├── api/              # HTTP API components
│   │   ├── handlers/     # HTTP request handlers
│   │   ├── middleware/   # HTTP middleware
│   │   ├── router/       # Router interface and implementations
│   │   ├── server.go     # HTTP server implementation
│   │   └── websocket/    # WebSocket handling
│   └── observability/    # Observability components
│       ├── logging/      # Logging infrastructure
│       ├── metrics/      # Metrics collection
│       └── tracing/      # Distributed tracing
├── docker/               # Docker-related files
├── go.mod                # Go module file
├── go.sum                # Go dependencies checksum
├── Makefile              # Build and development commands
└── README.md             # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose (for local development)

### Setup

1. Clone the repository

```bash
git clone https://github.com/babakgh/tuesdays/signaling-server-go-v2.git
cd signaling-server-go-v2
```

2. Install dependencies

```bash
go mod tidy
```

3. Build and run

```bash
make run
```

Or using Docker:

```bash
make docker-run
```

### Configuration

The server can be configured using environment variables or a configuration file. Key configuration options:

- `SERVER_PORT`: HTTP server port (default: 8080)
- `LOGGING_LEVEL`: Logging level (default: info)
- `METRICS_ENABLED`: Enable Prometheus metrics (default: true)
- `TRACING_ENABLED`: Enable OpenTelemetry tracing (default: true)

See `config/default.yaml` for more configuration options.

## API Endpoints

- `/health/live`: Liveness probe endpoint
- `/health/ready`: Readiness probe endpoint
- `/metrics`: Prometheus metrics endpoint
- `/ws`: WebSocket connection endpoint

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Build binary
make build

# Run with Docker Compose (with Prometheus and Jaeger)
make up
```

## Monitoring

The server exports metrics in Prometheus format at the `/metrics` endpoint. In the Docker Compose setup, Prometheus is configured to scrape these metrics.

Jaeger UI is available at `http://localhost:16686` for viewing traces when running with Docker Compose.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
