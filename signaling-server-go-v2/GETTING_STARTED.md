# Getting Started

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose (optional, for containerized development)

## Quick Start

1. Clone the repository and navigate to the project directory

```bash
git clone https://github.com/babakgh/tuesdays/signaling-server-go-v2.git
cd signaling-server-go-v2
```

2. Install dependencies

```bash
go mod tidy
```

3. Build and run the server

```bash
make build
make run
```

Or simply:

```bash
go run cmd/server/main.go
```

4. Test WebSocket connectivity

You can use a tool like `wscat` to test WebSocket connections:

```bash
wscat -c ws://localhost:8080/ws
```

Then try sending a message to join a room:

```json
{"type":"join","room":"test-room","sender":"client1"}
```

## Development with Docker

To run the server and its dependencies (Prometheus, Jaeger) using Docker Compose:

```bash
make up
```

This will start the following services:

- **Signaling Server**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Jaeger UI**: http://localhost:16686

To stop the Docker Compose services:

```bash
make down
```

## Environment Variables

The server can be configured using environment variables:

- `SERVER_PORT`: HTTP server port (default: 8080)
- `SERVER_HOST`: Host to listen on (default: 0.0.0.0)
- `LOGGING_LEVEL`: Logging level (default: info)
- `METRICS_ENABLED`: Enable Prometheus metrics (default: true)
- `TRACING_ENABLED`: Enable OpenTelemetry tracing (default: true)

See the configuration file `config/default.yaml` for all available options.

## Running Tests

To run the tests:

```bash
make test
```

or

```bash
go test ./...
```

## Project Structure

- **cmd/server**: Application entry point
- **config**: Configuration handling
- **internal/api**: HTTP API and WebSocket components
- **internal/observability**: Logging, metrics, and tracing
- **docker**: Docker-related files

See the README.md for a more detailed overview of the project structure.
