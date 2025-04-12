.PHONY: build run test lint docker docker-build docker-run up down clean test-coverage test-verbose

BIN_NAME=server
SRC_DIR=./cmd/server
BIN_DIR=./bin

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BIN_NAME) $(SRC_DIR)

run: build
	$(BIN_DIR)/$(BIN_NAME)

# Run all tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	@echo "Coverage report:"
	go tool cover -func=coverage.out

# Run tests verbosely
test-verbose:
	go test -v -race ./...

# Run linter
lint:
	golangci-lint run ./...

# Docker commands
docker-build:
	docker build -t signaling-server:latest -f docker/Dockerfile .

docker-run: docker-build
	docker run -p 8080:8080 signaling-server:latest

# Docker Compose commands
up:
	docker-compose up -d

down:
	docker-compose down

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)
	rm -f coverage.out
	docker-compose down -v