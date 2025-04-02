 # WebSocket Chat Server

A WebSocket-powered chat room server built with Go and the Gorilla WebSocket package.

## Features

- Real-time WebSocket communication
- Support for multiple chat rooms
- Member presence tracking
- Broadcast messages to all connected members
- List connected members
- Health check endpoint

## Getting Started

### Prerequisites

- Go 1.16 or later

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd chat-server-go
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run main.go
```

The server will start on port 8080.

## API Endpoints

### WebSocket Endpoint
- `ws://localhost:8080/ws`
  - Establishes a WebSocket connection
  - Supports the following commands:
    - `broadcast`: Send a message to all connected members
    - `list`: Get list of connected members
    - `me`: Get current member information

### Health Check
- `GET http://localhost:8080/health`
  - Returns "ok" if the server is running

## WebSocket Protocol

### Commands (Client → Server)

1. Broadcast Message
```json
{
  "command": "broadcast",
  "message": "Hello, world!"
}
```

2. List Members
```json
{
  "command": "list"
}
```

3. Get Member Info
```json
{
  "command": "me"
}
```

### Events (Server → Client)

1. Broadcast Event
```json
{
  "event": "broadcast",
  "member": "member123",
  "message": "Hello, world!"
}
```

2. List Event
```json
{
  "event": "list",
  "members": ["member123", "member456"]
}
```

3. Member Info Event
```json
{
  "event": "me",
  "member": "member123",
  "id": "abc-123-xyz"
}
```

## Testing

The project includes integration tests written in Python using the `websockets` and `pytest` libraries. To run the tests:

1. Install Python dependencies:
```bash
pip install websockets pytest
```

2. Run the tests:
```bash
python tests/integration_test.py
```

## Architecture

The project follows a clean architecture with the following layers:

- `transport/`: WebSocket handlers and HTTP endpoints
- `wire/`: Command/event JSON parsing and validation
- `commands/`: Command implementations and business logic
- `persistence/`: Member store implementations
- `domain/`: Core interfaces and domain models
