# Project Dependencies

This project requires the following Go dependencies to be installed:

```bash
go get github.com/go-chi/chi/v5
go get github.com/go-kit/log
go get github.com/google/uuid
go get github.com/gorilla/websocket
go get github.com/prometheus/client_golang
go get github.com/spf13/viper
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
go get go.opentelemetry.io/otel/sdk
go get google.golang.org/grpc
```

## Note on Implementation

The project has been set up with a complete production-ready structure for a WebRTC signaling server. The implementation includes:

1. **Clean Architecture**: Clear separation of concerns with interfaces and adapters
2. **Configuration Management**: Support for environment variables and config files
3. **Observability**:
   - Structured logging with go-kit/log
   - Prometheus metrics
   - OpenTelemetry tracing
4. **Health Checks**: Readiness and liveness endpoints
5. **WebSocket Support**: WebSocket handler for client connections
6. **Graceful Shutdown**: Proper signal handling and shutdown sequence

The current implementation should be considered a foundation that needs to be built upon. To complete the project:

1. Run `go mod tidy` to resolve dependencies
2. Add specific business logic for WebRTC signaling
3. Implement room management and other signaling-specific features
4. Add more unit and integration tests

This implementation provides a robust foundation with production-quality infrastructure that can be extended for WebRTC signaling purposes.