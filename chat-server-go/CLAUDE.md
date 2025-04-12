# Go WebSocket Chat Server Development Guide

## Build/Test Commands
- Build project: `go build`
- Run server: `go run main.go`
- Run all tests: `go test ./...`
- Run specific package tests: `go test ./transport/...`
- Run single test: `go test -run TestName ./package/...` (e.g., `go test -run TestMemoryStore_ConcurrentAccess ./persistence`)
- Integration tests: `cd tests && pytest test_chat.py`
- Test with race detection: `go test -race ./...`
- Test with verbose output: `go test -v ./...`

## Code Style Guidelines
- **Imports**: Group stdlib, 3rd party, then local packages, alphabetically within groups
- **Error Handling**: Check errors immediately and return with descriptive messages
- **Types**: Define interfaces in domain package; implement interfaces in other packages
- **Naming**: Use camelCase; exported functions are PascalCase
- **Documentation**: Comment all exported functions with proper Go doc format
- **Testing**: Name tests `TestFunctionName_Scenario`; use testify/assert for assertions
- **Concurrency**: Use mutex for protecting shared state; avoid goroutine leaks
- **Architecture**: Follow the layered architecture (domain → transport → commands → persistence)