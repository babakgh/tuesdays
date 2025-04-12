# Signaling Server Go v2

A production-ready WebRTC signaling server written in Go, implementing proper startup and graceful shutdown patterns with complete observability through traces, metrics, and logs.

## Features

- Clean architecture with clear separation of concerns
- Configuration through environment variables and YAML files
- Graceful shutdown handling
- Health check endpoints (/health/live, /health/ready)
- Structured logging
- Prometheus metrics
- OpenTelemetry tracing

## Getting Started

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for local development)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/tuesdays/signaling-server-go-v2.git
cd signaling-server-go-v2
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the server:
```bash
go run cmd/server/main.go
```

### Configuration

The server can be configured through:
- Environment variables
- YAML configuration file (default: `config/default.yaml`)

Set the `CONFIG_FILE` environment variable to use a custom configuration file:
```bash
export CONFIG_FILE=/path/to/config.yaml
```

## Development

### Running Tests

```bash
go test ./...
```

### Local Development with Docker

```bash
docker-compose up
```

## API Endpoints

- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

## License

This project is licensed under the MIT License - see the LICENSE file for details. 