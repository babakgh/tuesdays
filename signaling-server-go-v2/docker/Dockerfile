# Build stage
FROM golang:1.21-alpine AS build

WORKDIR /app

# Install dependencies required for building
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/server ./cmd/server

# Final stage
FROM alpine:3.17

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the build stage
COPY --from=build /bin/server /app/server

# Copy configuration files
COPY config/default.yaml /app/config/default.yaml

# Set environment variables
ENV SERVER_CONFIG_PATH=/app/config/default.yaml

# Expose ports
EXPOSE 8080

# Run the application
CMD ["/app/server"]