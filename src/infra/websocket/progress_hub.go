package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// ProgressUpdate represents a progress update message
type ProgressUpdate struct {
	TaskID     string                 `json:"task_id"`
	TaskType   string                 `json:"task_type"`
	Progress   float64                `json:"progress"`
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	Timestamp  time.Time              `json:"timestamp"`
	UserID     string                 `json:"user_id"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ProgressHub manages WebSocket connections for progress tracking
type ProgressHub struct {
	clients   map[string]map[string]*websocket.Conn // userID -> connectionID -> connection
	broadcast chan ProgressUpdate
	register  chan *ClientConnection
	unregister chan *ClientConnection
	mu        sync.RWMutex
}

// ClientConnection represents a client WebSocket connection
type ClientConnection struct {
	UserID       string
	ConnectionID string
	Conn         *websocket.Conn
	Send         chan ProgressUpdate
}

// NewProgressHub creates a new progress tracking hub
func NewProgressHub() *ProgressHub {
	return &ProgressHub{
		clients:    make(map[string]map[string]*websocket.Conn),
		broadcast:  make(chan ProgressUpdate, 256),
		register:   make(chan *ClientConnection, 16),
		unregister: make(chan *ClientConnection, 16),
	}
}

// Run starts the progress hub
func (h *ProgressHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case update := <-h.broadcast:
			h.broadcastUpdate(update)
		}
	}
}

// registerClient adds a new client connection
func (h *ProgressHub) registerClient(client *ClientConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[string]*websocket.Conn)
	}
	
	h.clients[client.UserID][client.ConnectionID] = client.Conn
	log.Printf("Client registered: userID=%s, connectionID=%s", client.UserID, client.ConnectionID)

	// Send initial connection confirmation
	initialUpdate := ProgressUpdate{
		TaskID:    "connection",
		TaskType:  "connection",
		Progress:  0,
		Status:    "connected",
		Message:   "WebSocket connection established",
		Timestamp: time.Now(),
		UserID:    client.UserID,
	}
	
	h.sendToConnection(client.Conn, initialUpdate)
}

// unregisterClient removes a client connection
func (h *ProgressHub) unregisterClient(client *ClientConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if userConnections, exists := h.clients[client.UserID]; exists {
		if conn, exists := userConnections[client.ConnectionID]; exists {
			conn.Close()
			delete(userConnections, client.ConnectionID)
			
			// Remove user entry if no more connections
			if len(userConnections) == 0 {
				delete(h.clients, client.UserID)
			}
			
			log.Printf("Client unregistered: userID=%s, connectionID=%s", client.UserID, client.ConnectionID)
		}
	}
}

// broadcastUpdate sends an update to all connections for a specific user
func (h *ProgressHub) broadcastUpdate(update ProgressUpdate) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if userConnections, exists := h.clients[update.UserID]; exists {
		for connectionID, conn := range userConnections {
			if !h.sendToConnection(conn, update) {
				// Connection is dead, remove it
				delete(userConnections, connectionID)
			}
		}
		
		// Clean up user entry if no more connections
		if len(userConnections) == 0 {
			delete(h.clients, update.UserID)
		}
	}
}

// sendToConnection sends an update to a specific connection
func (h *ProgressHub) sendToConnection(conn *websocket.Conn, update ProgressUpdate) bool {
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.WriteJSON(update); err != nil {
		log.Printf("Error writing to WebSocket: %v", err)
		return false
	}
	return true
}

// BroadcastProgress sends a progress update to all connections for a user
func (h *ProgressHub) BroadcastProgress(userID, taskID, taskType string, progress float64, status, message string, metadata map[string]interface{}) {
	update := ProgressUpdate{
		TaskID:    taskID,
		TaskType:  taskType,
		Progress:  progress,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		UserID:    userID,
		Metadata:  metadata,
	}

	select {
	case h.broadcast <- update:
	default:
		log.Printf("Progress broadcast buffer full, dropping update for user %s", userID)
	}
}

// GetConnectedUsers returns a list of currently connected users
func (h *ProgressHub) GetConnectedUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		users = append(users, userID)
	}
	return users
}

// GetConnectionCount returns the total number of active connections
func (h *ProgressHub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, userConnections := range h.clients {
		count += len(userConnections)
	}
	return count
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *ProgressHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Extract user ID from query parameters or JWT token
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "Missing user_id"))
		conn.Close()
		return
	}

	// Generate unique connection ID
	connectionID := generateConnectionID()

	client := &ClientConnection{
		UserID:       userID,
		ConnectionID: connectionID,
		Conn:         conn,
		Send:         make(chan ProgressUpdate, 256),
	}

	// Register the client
	h.register <- client

	// Set up ping/pong handlers for connection health
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start goroutines for reading and writing
	go h.writePump(client)
	go h.readPump(client)
}

// readPump handles incoming messages from the client
func (h *ProgressHub) readPump(client *ClientConnection) {
	defer func() {
		h.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512) // Limit message size
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		// Read message from client (ping/pong or control messages)
		messageType, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if messageType == websocket.TextMessage {
			// Handle client control messages if needed
			var controlMessage map[string]interface{}
			if err := json.Unmarshal(message, &controlMessage); err == nil {
				h.handleControlMessage(client, controlMessage)
			}
		}

		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}

// writePump handles outgoing messages to the client
func (h *ProgressHub) writePump(client *ClientConnection) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case update, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(update); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleControlMessage processes control messages from clients
func (h *ProgressHub) handleControlMessage(client *ClientConnection, message map[string]interface{}) {
	messageType, exists := message["type"]
	if !exists {
		return
	}

	switch messageType {
	case "subscribe":
		// Handle task subscription if needed
		if taskID, ok := message["task_id"].(string); ok {
			log.Printf("Client %s subscribed to task %s", client.UserID, taskID)
		}
	case "unsubscribe":
		// Handle task unsubscription if needed
		if taskID, ok := message["task_id"].(string); ok {
			log.Printf("Client %s unsubscribed from task %s", client.UserID, taskID)
		}
	}
}

// generateConnectionID generates a unique connection identifier
func generateConnectionID() string {
	return time.Now().Format("20060102150405") + "-" + generateRandomString(8)
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// UploadProgressTracker tracks upload progress and sends updates
type UploadProgressTracker struct {
	hub       *ProgressHub
	userID    string
	taskID    string
	totalSize int64
	uploaded  int64
}

// NewUploadProgressTracker creates a new upload progress tracker
func NewUploadProgressTracker(hub *ProgressHub, userID, taskID string, totalSize int64) *UploadProgressTracker {
	return &UploadProgressTracker{
		hub:       hub,
		userID:    userID,
		taskID:    taskID,
		totalSize: totalSize,
		uploaded:  0,
	}
}

// UpdateProgress updates the upload progress
func (upt *UploadProgressTracker) UpdateProgress(uploadedBytes int64) {
	upt.uploaded = uploadedBytes
	progress := float64(upt.uploaded) / float64(upt.totalSize) * 100
	
	metadata := map[string]interface{}{
		"uploaded_bytes": upt.uploaded,
		"total_bytes":    upt.totalSize,
		"upload_speed":   calculateUploadSpeed(upt.uploaded, time.Now()),
	}

	var status string
	var message string
	
	switch {
	case progress < 25:
		status = "uploading"
		message = fmt.Sprintf("Uploading... %.1f%%", progress)
	case progress < 50:
		status = "uploading"
		message = fmt.Sprintf("Upload in progress... %.1f%%", progress)
	case progress < 100:
		status = "uploading"
		message = fmt.Sprintf("Almost done... %.1f%%", progress)
	default:
		status = "completed"
		message = "Upload completed successfully"
	}

	upt.hub.BroadcastProgress(upt.userID, upt.taskID, "file_upload", progress, status, message, metadata)
}

// calculateUploadSpeed calculates upload speed (simplified implementation)
func calculateUploadSpeed(uploadedBytes int64, startTime time.Time) float64 {
	duration := time.Since(startTime).Seconds()
	if duration > 0 {
		return float64(uploadedBytes) / duration // bytes per second
	}
	return 0
}