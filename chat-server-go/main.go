package main

import (
	"fmt"
	"log"
	"net/http"

	"chat-server-go/transport"
)

func main() {
	// Create WebSocket handler
	wsHandler := transport.NewWebSocketHandler()

	// Set up routes
	http.HandleFunc("/ws", wsHandler.HandleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	// Start server
	port := ":8080"
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
