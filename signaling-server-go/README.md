# Signaling Server Go

A production-ready WebRTC signaling server implementation in Go.

## Features

- Clean architecture with clear separation of concerns
- Configuration management with environment variables and YAML support
- Graceful shutdown handling
- Comprehensive observability:
  - Structured logging with Zap
  - Prometheus metrics
  - OpenTelemetry tracing
- Health check endpoints
- Robust error handling

## Getting Started

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile commands)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/tuesdays/signaling-server-go.git
cd signaling-server-go
```

2. Install dependencies:
```bash
go mod download
```

### Configuration

The server can be configured using:
- Environment variables
- YAML configuration file (`config/default.yaml`)

Key configuration options:
- Server host and port
- Logging level and format
- Metrics collection settings
- Tracing configuration
- Health check endpoints

### Running the Server

```bash
go run cmd/server/main.go
```

Or using the Makefile:
```bash
make run
```

### Health Checks

The server exposes the following health check endpoints:
- `/health/live` - Liveness probe
- `/health/ready` - Readiness probe

### Metrics

Prometheus metrics are available at `/metrics` when enabled in the configuration.

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 