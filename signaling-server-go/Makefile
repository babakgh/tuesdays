.PHONY: build run test lint clean

# Build variables
BINARY_NAME=signaling-server
GO=go

# Build the application
build:
	$(GO) build -o $(BINARY_NAME) cmd/server/main.go

# Run the application
run:
	$(GO) run cmd/server/main.go

# Run tests
test:
	$(GO) test -v ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	$(GO) clean 