Build a WebSocket-powered chat room server using Go and the Gorilla WebSocket package.

## Chat Scenarios

The following scenarios describe how the chat system behaves and what should be logged at key moments:

**Scenario 1: Connecting to the Chat Room**

- **Given** a client wants to join the chat,
- **When** they establish a WebSocket connection with the server,
- **Then** the server assigns a unique member ID and name,
- **And** sends a `broadcast` message from the system saying `"member123 has joined!"`,
- **And** logs: `🔌 Member member123 connected`.

**Scenario 2: Sending a Broadcast Message**

- **Given** a client is connected to the server,
- **When** they send a `broadcast` message,
- **Then** the server relays the message to all members, including the sender,
- **And** logs: `📤 Broadcast from member123: Hello world!`.

**Scenario 3: Disconnecting from the Chat Room**

- **Given** a client is connected to the server,
- **When** they disconnect from the server,
- **Then** the server removes the member from its presence list,
- **And** logs: `🔌 Member member123 disconnected`.

## Functional Overview

1. **Command and Event Structure**

The system distinguishes between commands (client → server) and events (server → client).

To promote type safety and extensibility, events can also follow a common interface similar to commands.

- **Commands** are messages initiated by a member and must include a `command` field.
- **Events** are messages pushed by the server and must include an `event` field.

Each command is implemented as a concrete type that satisfies the following interface:

```go
// Command defines the interface all client commands must implement
// to execute logic against the server context.
type Command interface {
  Execute() error
}
```

Each event can also implement a structured interface:

```go
// Event defines the interface all server-pushed messages should implement
type Event interface {
  Name() string           // e.g., "broadcast", "system"
  Payload() interface{}   // the data to be encoded into the JSON response
}
```

### Command & Event Reference

- **Broadcast** — sends a message to all members in the same room.

  **Command**

  ```json
  {
    "command": "broadcast",
    "message": "Hello, world!"
  }
  ```

  **Event**

  ```json
  {
    "event": "broadcast",
    "member": "member123",
    "message": "Hello, world!"
  }
  ```

  **System Message (if no members)**

  ```json
  {
    "event": "system",
    "message": "No other members are currently connected."
  }
  ```

- **Me** — returns the connected member’s ID and name.

  **Command**

  ```json
  {
    "command": "me"
  }
  ```

  **Event**

  ```json
  {
    "event": "me",
    "member": "member123",
    "id": "abc-123-xyz"
  }
  ```

- **List** — retrieves the list of members currently connected.

  **Command**

  ```json
  {
    "command": "list"
  }
  ```

  **Event**

  ```json
  {
    "event": "list",
    "members": ["member123", "member456"]
  }
  ```

2. **Integration Testing**

Integration tests are written in **Python** using the `websockets` and `pytest` libraries. This setup is portable, easy to extend, and future-proof across different backend implementations.

- Each test spawns one or more asynchronous WebSocket clients.
- Commands are sent as JSON payloads following the defined protocol.
- Tests await server responses and assert behavior.
- Timeouts are enforced for all expected messages to prevent hangs.
- The test framework is reusable across environments and can run as part of CI/CD pipelines.

* No frontend is required.

#### Test 1: Two members connect to the room

- Simulate two WebSocket members connecting to the chat room.
- After both connections, assert that the server's internal member storage contains **exactly two members**.
- Member 1 should send a `list` command and receive the correct JSON response from the server
- Log: `✅ Test 1 passed: 2 members connected and present in storage`

#### Test 2: Broadcast message is delivered to all members

- Simulate two members connecting.
- Member 1 sends a `broadcast` command with content: "Hello from member1!".
- Assert that both members receive the message.
- Log: `✅ Test 2 passed: Broadcast message received by all members`

#### Test 3: Disconnect removes member from list

- Simulate two members connecting.
- Member 2 disconnects.
- Member 1 sends a `list` command.
- Assert that the server returns only Member 1 in the list.
- Log: `✅ Test 3 passed: Disconnected member removed from server list`

#### Test 4: New member receives welcome commands

- Simulate one member connecting.
- Assert that the member receives a `me` command.
- Assert that the member receives a `broadcast` message like: "member123 joined!".
- Log: `✅ Test 4 passed: Member received me + joined broadcast`

## Endpoints

The server exposes the following endpoints:

- `GET /health`

  - Simple health check endpoint. Returns 200 OK with a plain message like `"ok"`.

- `GET /ws`

  - Upgrades the HTTP connection to a WebSocket.
  - Clients must connect here to join the chat room.
  - Once connected, clients can send and receive JSON-formatted messages according to the supported commands (`me`, `broadcast`, `list`).

## Architecture

### Design Patterns

The architecture of this chat room server is designed using key software design patterns to support extensibility, testability, and concurrency:

- **Observer Pattern**: The `Broadcaster` implements this pattern to distribute messages to all connected members (observers) when a new message is broadcast.

- **Command Pattern**: Incoming WebSocket messages are treated as structured command objects (e.g., `broadcast`, `list`, `me`) that implement a common `Execute()` interface. These commands are parsed and dispatched through a central dispatcher. This pattern supports type-safe, decoupled, and testable command logic.

- **Strategy Pattern**: If different broadcast behaviors or delivery mechanisms are introduced (e.g., private messages, filtered topics), they can be implemented as interchangeable strategies.

- **Dependency Injection + Interface Segregation**: Key interfaces like `MemberStore` and `Broadcaster` decouple the logic from implementation, allowing easy testing and future substitution (e.g., Redis-backed store).

### Component Responsibilities

- **WebSocketHandler**: Upgrades HTTP to WebSocket and manages the connection lifecycle.
- **Command (interface)**: Defines a contract for dispatching command objects parsed from incoming messages. Each concrete command (e.g., `ListCommand`, `BroadcastCommand`) implements the `Command` interface with an `Execute(ctx)` method, and is routed through the dispatcher for execution.
- **MemberStore (interface)**: Manages connected members and supports add/remove/list operations.
- **Broadcaster (interface)**: Delivers messages to all connected members.
- **MemberSession**: Represents a connected member, storing WebSocket connection, name, and ID.

### Layered Structure

- **Transport Layer**: Handles WebSocket upgrades and connection management.
- **Wire Layer**: Parses JSON commands and validates structure.
- **Commands Layer**: Contains logic for each command (`broadcast`, `list`, etc.).
- **Persistence Layer**: Currently an in-memory store, abstracted for easy replacement.

### Folder Structure Alignment

This layered structure can be reflected in the folder layout of the codebase for improved clarity and maintainability:

- `transport/` — WebSocket handlers and HTTP endpoints
- `wire/` — Command/event JSON parsing and validation logic
- `commands/` — Command implementations and business logic
- `persistence/` — Member store implementations (e.g., in-memory, Redis)
- `domain/` — Core interfaces like `Command`, `Broadcaster`, `MemberStore`

Aligning the folder structure with these layers helps enforce separation of concerns and simplifies navigation as the system grows.

