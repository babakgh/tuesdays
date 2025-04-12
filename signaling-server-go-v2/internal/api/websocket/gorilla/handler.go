package gorilla

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/babakgh/tuesdays/signaling-server-go-v2/config"
	ws "github.com/babakgh/tuesdays/signaling-server-go-v2/internal/api/websocket"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/logging"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/metrics"
	"github.com/babakgh/tuesdays/signaling-server-go-v2/internal/observability/tracing"
)

// Handler implements WebSocketHandler with a basic implementation
type Handler struct {
	wsConfig   ws.WebSocketConfig
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	logger     logging.Logger
	metrics    *metrics.Metrics
	tracer     tracing.Tracer
	mux        sync.Mutex
	nextID     int
}

// Client represents a connected WebSocket client
type Client struct {
	id      string
	handler *Handler
	send    chan []byte
	logger  logging.Logger
	metrics *metrics.Metrics
	tracer  tracing.Tracer
}

// NewHandler creates a new websocket handler
func NewHandler(cfg config.WebSocketConfig, logger logging.Logger, m *metrics.Metrics, tracer tracing.Tracer) ws.WebSocketHandler {
	wsConfig := ws.NewWebSocketConfig(cfg)
	h := &Handler{
		wsConfig:   wsConfig,
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		logger:     logger.With("component", "websocket"),
		metrics:    m,
		tracer:     tracer,
		nextID:     1,
	}

	// Start the client manager
	go h.run()

	return h
}

// run processes client registration and broadcasts
func (h *Handler) run() {
	for {
		select {
		case client := <-h.register:
			h.mux.Lock()
			h.clients[client.id] = client
			h.mux.Unlock()
			h.logger.Info("Client registered", "client_id", client.id)
			if h.metrics != nil {
				h.metrics.WebSocketConnect()
			}

		case client := <-h.unregister:
			h.mux.Lock()
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.send)
				h.logger.Info("Client unregistered", "client_id", client.id)
				if h.metrics != nil {
					h.metrics.WebSocketDisconnect()
				}
			}
			h.mux.Unlock()

		case message := <-h.broadcast:
			h.mux.Lock()
			for id, client := range h.clients {
				select {
				case client.send <- message:
					// Message sent to client
				default:
					// Failed to send - client buffer full
					close(client.send)
					delete(h.clients, id)
					if h.metrics != nil {
						h.metrics.WebSocketDisconnect()
						h.metrics.WebSocketError("send_buffer_full")
					}
				}
			}
			h.mux.Unlock()
		}
	}
}

// HandleConnection handles a new WebSocket connection
func (h *Handler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would upgrade the connection to WebSocket
	// For now, just create a simulated client and acknowledge the connection
	h.mux.Lock()
	clientID := h.generateClientID()
	h.mux.Unlock()

	client := &Client{
		id:      clientID,
		handler: h,
		send:    make(chan []byte, 256),
		logger:  h.logger.With("client_id", clientID),
		metrics: h.metrics,
		tracer:  h.tracer,
	}

	// Register the client
	h.register <- client

	// Since we can't actually establish a WebSocket connection in this context,
	// we'll send a success response and log it
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"connected","message":"WebSocket connection simulated","client_id":"` + clientID + `"}`))
}

// BroadcastMessage sends a message to all connected clients
func (h *Handler) BroadcastMessage(message []byte) error {
	h.broadcast <- message
	return nil
}

// SendMessage sends a message to a specific client
func (h *Handler) SendMessage(clientID string, message []byte) error {
	h.mux.Lock()
	defer h.mux.Unlock()

	client, ok := h.clients[clientID]
	if !ok {
		h.logger.Error("Client not found", "client_id", clientID)
		return nil
	}

	select {
	case client.send <- message:
		return nil
	default:
		// Client send channel is full - disconnect client
		close(client.send)
		delete(h.clients, clientID)
		if h.metrics != nil {
			h.metrics.WebSocketDisconnect()
			h.metrics.WebSocketError("send_buffer_full")
		}
		return nil
	}
}

// CloseConnection closes a client's connection
func (h *Handler) CloseConnection(clientID string) error {
	h.mux.Lock()
	defer h.mux.Unlock()

	client, ok := h.clients[clientID]
	if !ok {
		return nil
	}

	// Close the send channel to signal disconnect
	close(client.send)
	delete(h.clients, clientID)
	if h.metrics != nil {
		h.metrics.WebSocketDisconnect()
	}

	return nil
}

// generateClientID generates a simple client ID
func (h *Handler) generateClientID() string {
	id := h.nextID
	h.nextID++
	return "client-" + time.Now().Format("20060102150405") + "-" + fmt.Sprint(id)
}
