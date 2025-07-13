package handlers

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/usman-007/checkbox-backend/internal/monitoring"
	"github.com/usman-007/checkbox-backend/internal/services"
)

type WebSocketHandler struct {
	checkboxService *services.CheckboxService
	redisClient     *redis.Client
	clients  map[*websocket.Conn]bool
	mutex    sync.Mutex 
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new instance of WebSocketHandler
func NewWebSocketHandler(checkboxService *services.CheckboxService, redisClient *redis.Client) *WebSocketHandler {
	if checkboxService == nil {
		log.Fatal("CheckboxService is nil in NewWebSocketHandler")
	}
	if redisClient == nil {
		log.Fatal("RedisClient is nil in NewWebSocketHandler")
	}

	return &WebSocketHandler{
		checkboxService: checkboxService,
		redisClient:     redisClient,
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow all origins for development - Consider restricting in production
			CheckOrigin: func(r *http.Request) bool {
				// log.Printf("WebSocket CheckOrigin: Host=%s, Origin=%s", r.Host, r.Header.Get("Origin")) // Debug logging
				return true // Be careful with this in production
			},
		},
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket and handles the connection lifecycle
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	defer conn.Close()
	log.Printf("WebSocket connection established: %s", conn.RemoteAddr())

	// Register new client connection
	h.mutex.Lock()
	if h.clients == nil {
		// This should not happen if NewWebSocketHandler is used correctly, but defensive check
		log.Println("CRITICAL: h.clients map is nil during client registration!")
		h.clients = make(map[*websocket.Conn]bool)
	}
	h.clients[conn] = true
	// Update WebSocket metrics
	monitoring.WebSocketConnections.Inc()
	h.mutex.Unlock()

	// --- Unregister client when the connection closes ---
	defer func() {
		h.mutex.Lock()
		if _, ok := h.clients[conn]; ok {
			delete(h.clients, conn)
			// Update WebSocket metrics
			monitoring.WebSocketConnections.Dec()
			log.Printf("Client unregistered: %s. Remaining clients: %d", conn.RemoteAddr(), len(h.clients))
		} else {
			log.Printf("Attempted to unregister client %s but it was already removed.", conn.RemoteAddr())
		}
		h.mutex.Unlock()
	}()
	// --- End Unregister ---

	// --- Send initial state to the newly connected client ---
		checkboxes, err := h.checkboxService.GetAllCheckboxes()
		if err != nil {
			log.Printf("Failed to get initial checkbox state for client %s: %v", conn.RemoteAddr(), err)
		} else {
			// Use WriteJSON for the initial state as it's likely a complex object/slice
			h.mutex.Lock() // Lock needed if GetAllCheckboxes modifies shared state (unlikely) or if WriteJSON isn't thread-safe for the conn
			err := conn.WriteJSON(checkboxes)
			h.mutex.Unlock()
			if err != nil {
				log.Printf("Failed to send initial state to client %s: %v", conn.RemoteAddr(), err)
				return 
			}
		}

	// --- End Initial State ---

	// --- Keep-alive and Disconnect Detection Loop ---
	// Read messages from the client. This loop primarily serves to detect
	// when the client disconnects. We don't expect specific messages here
	// unless the client is designed to send them.
	for {
		// ReadMessage blocks until a message is received or an error occurs (like disconnect)
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			// Check if the error indicates a normal closure or an unexpected error
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				log.Printf("Error reading message from client %s: %v", conn.RemoteAddr(), err)
			} else {
				log.Printf("Client %s disconnected.", conn.RemoteAddr())
			}
			break
		}
		// Record incoming message metric
		monitoring.WebSocketMessagesTotal.WithLabelValues("received").Inc()
		// Optional: Handle messages received from the client if needed
		log.Printf("Received message from client %s (type %d): %s", conn.RemoteAddr(), messageType, message)
	}

}

// StartRedisSubscription starts listening for Redis Pub/Sub messages and broadcasts updates
func (h *WebSocketHandler) StartRedisSubscription() {

	ctx := context.Background() 
	pubsub := h.redisClient.Subscribe(ctx, "checkbox_updates")
	_, err := pubsub.Receive(ctx)
	if err != nil {
		log.Printf("FATAL: Failed to subscribe to Redis channel 'checkbox_updates': %v", err)
		// Depending on application structure, might want to panic or signal failure
		return // Exit if subscription fails
	}

	// Get the channel for receiving messages. This is the idiomatic way.
	ch := pubsub.Channel()

	// Start a goroutine to process incoming messages from the channel
	go func() {
		// Ensure resources are cleaned up when this goroutine exits.
		// Closing the pubsub here will also cause the range loop over 'ch' below to terminate.
		defer pubsub.Close()
		defer log.Println("Exiting Redis message listener goroutine.")

		// This loop reads from the channel until it's closed.
		for msg := range ch {
			// Log the raw payload received
			log.Printf("Received message payload from Redis 'checkbox_updates': %s", msg.Payload)

			h.mutex.Lock() 

			if h.clients == nil {
				log.Println("CRITICAL: h.clients map is nil in broadcast goroutine!")
				h.mutex.Unlock() // Unlock before skipping
				continue       // Skip this message if map is nil
			}

			clientCount := len(h.clients)
			broadcastCount := 0

			// Iterate over the clients map safely
			for client, active := range h.clients {
				// If you use the boolean value, you can check if a client is marked inactive
				if !active {
					continue
				}

				err := client.WriteMessage(websocket.TextMessage, []byte(msg.Payload))

				if err != nil {
					// Log the error and remove the problematic client connection
					log.Printf("Error sending message to client %p (%s): %v. Closing and removing.", client, client.RemoteAddr(), err)
					// Close the connection (best effort, might already be closed)
					client.Close()
					// Remove the client from the map
					delete(h.clients, client)
				} else {
					// Record outgoing message metric
					monitoring.WebSocketMessagesTotal.WithLabelValues("sent").Inc()
					// log.Printf("Successfully sent message to client %p (%s)", client, client.RemoteAddr()) // Verbose logging
					broadcastCount++
				}
			} // End of client loop

			if clientCount > 0 {
				log.Printf("Finished broadcasting message to %d out of %d initial clients.", broadcastCount, clientCount)
			}

			h.mutex.Unlock() 
		}
	}() 
}