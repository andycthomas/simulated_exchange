package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketSubscriber implements the DemoSubscriber interface for WebSocket connections
type WebSocketSubscriber struct {
	id         string
	conn       *websocket.Conn
	updateChan chan DemoUpdate
	stopChan   chan struct{}
	logger     *slog.Logger
	active     bool
	mu         sync.RWMutex

	// Subscription filters
	updateTypes []UpdateType
	filters     map[string]interface{}

	// Connection management
	lastPing    time.Time
	lastPong    time.Time
	writeTimeout time.Duration
	readTimeout  time.Duration
}

// WebSocketHub manages WebSocket connections and broadcasts
type WebSocketHub struct {
	config      WebSocketConfig
	logger      *slog.Logger
	upgrader    websocket.Upgrader

	// Connection management
	subscribers map[string]*WebSocketSubscriber
	mu          sync.RWMutex

	// Broadcast channels
	broadcast   chan DemoUpdate
	register    chan *WebSocketSubscriber
	unregister  chan *WebSocketSubscriber
	shutdown    chan struct{}

	// Statistics
	connectionCount int64
	messageCount    int64
	errorCount      int64
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub(config WebSocketConfig, logger *slog.Logger) *WebSocketHub {
	return &WebSocketHub{
		config: config,
		logger: logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for demo purposes
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		subscribers: make(map[string]*WebSocketSubscriber),
		broadcast:   make(chan DemoUpdate, 1000),
		register:    make(chan *WebSocketSubscriber),
		unregister:  make(chan *WebSocketSubscriber),
		shutdown:    make(chan struct{}),
	}
}

// Start begins the WebSocket hub operations
func (hub *WebSocketHub) Start(ctx context.Context) {
	hub.logger.Info("Starting WebSocket hub")

	// Start the main hub loop
	go hub.run(ctx)

	// Start ping routine
	go hub.pingClients(ctx)
}

// Stop gracefully shuts down the WebSocket hub
func (hub *WebSocketHub) Stop(ctx context.Context) error {
	hub.logger.Info("Stopping WebSocket hub")

	// Close all connections
	hub.mu.RLock()
	for _, subscriber := range hub.subscribers {
		subscriber.Close()
	}
	hub.mu.RUnlock()

	// Signal shutdown
	close(hub.shutdown)

	return nil
}

// HandleWebSocket handles new WebSocket connections
func (hub *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := hub.upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.logger.Error("Failed to upgrade connection", "error", err)
		hub.errorCount++
		return
	}

	// Check connection limit
	hub.mu.RLock()
	currentConnections := len(hub.subscribers)
	hub.mu.RUnlock()

	if currentConnections >= hub.config.MaxConnections {
		hub.logger.Warn("Connection limit reached", "current", currentConnections, "max", hub.config.MaxConnections)
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Connection limit reached"))
		conn.Close()
		return
	}

	// Create subscriber
	subscriber := NewWebSocketSubscriber(conn, hub.config, hub.logger)

	// Register subscriber
	hub.register <- subscriber

	// Start subscriber goroutines
	go subscriber.readPump(hub.unregister)
	go subscriber.writePump()

	hub.connectionCount++
	hub.logger.Info("New WebSocket connection", "id", subscriber.id, "total_connections", currentConnections+1)
}

// BroadcastUpdate sends an update to all subscribers
func (hub *WebSocketHub) BroadcastUpdate(update DemoUpdate) {
	select {
	case hub.broadcast <- update:
		hub.messageCount++
	default:
		hub.logger.Warn("Broadcast channel full, dropping update", "type", update.Type)
		hub.errorCount++
	}
}

// GetStats returns hub statistics
func (hub *WebSocketHub) GetStats() map[string]interface{} {
	hub.mu.RLock()
	defer hub.mu.RUnlock()

	return map[string]interface{}{
		"active_connections": len(hub.subscribers),
		"total_connections":  hub.connectionCount,
		"messages_sent":      hub.messageCount,
		"errors":            hub.errorCount,
		"max_connections":   hub.config.MaxConnections,
	}
}

// Private methods

func (hub *WebSocketHub) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-hub.shutdown:
			return
		case subscriber := <-hub.register:
			hub.mu.Lock()
			hub.subscribers[subscriber.id] = subscriber
			hub.mu.Unlock()
			hub.logger.Debug("Registered WebSocket subscriber", "id", subscriber.id)

		case subscriber := <-hub.unregister:
			hub.mu.Lock()
			if _, exists := hub.subscribers[subscriber.id]; exists {
				delete(hub.subscribers, subscriber.id)
				subscriber.Close()
			}
			hub.mu.Unlock()
			hub.logger.Debug("Unregistered WebSocket subscriber", "id", subscriber.id)

		case update := <-hub.broadcast:
			hub.mu.RLock()
			for _, subscriber := range hub.subscribers {
				if subscriber.shouldReceiveUpdate(update) {
					select {
					case subscriber.updateChan <- update:
					default:
						// Subscriber channel full, close connection
						hub.logger.Warn("Subscriber channel full, closing connection", "id", subscriber.id)
						subscriber.Close()
					}
				}
			}
			hub.mu.RUnlock()
		}
	}
}

func (hub *WebSocketHub) pingClients(ctx context.Context) {
	ticker := time.NewTicker(hub.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hub.shutdown:
			return
		case <-ticker.C:
			hub.mu.RLock()
			for _, subscriber := range hub.subscribers {
				if err := subscriber.ping(); err != nil {
					hub.logger.Debug("Failed to ping subscriber", "id", subscriber.id, "error", err)
				}
			}
			hub.mu.RUnlock()
		}
	}
}

// WebSocketSubscriber Implementation

// NewWebSocketSubscriber creates a new WebSocket subscriber
func NewWebSocketSubscriber(conn *websocket.Conn, config WebSocketConfig, logger *slog.Logger) *WebSocketSubscriber {
	id := fmt.Sprintf("ws_%d", time.Now().UnixNano())

	subscriber := &WebSocketSubscriber{
		id:           id,
		conn:         conn,
		updateChan:   make(chan DemoUpdate, 100),
		stopChan:     make(chan struct{}),
		logger:       logger,
		active:       true,
		updateTypes:  []UpdateType{}, // Subscribe to all by default
		filters:      make(map[string]interface{}),
		lastPing:     time.Now(),
		lastPong:     time.Now(),
		writeTimeout: config.WriteTimeout,
		readTimeout:  config.ReadTimeout,
	}

	// Set connection limits
	conn.SetReadLimit(config.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(config.ReadTimeout))
	conn.SetPongHandler(func(string) error {
		subscriber.lastPong = time.Now()
		conn.SetReadDeadline(time.Now().Add(config.ReadTimeout))
		return nil
	})

	return subscriber
}

// GetID returns the subscriber ID
func (ws *WebSocketSubscriber) GetID() string {
	return ws.id
}

// SendUpdate sends an update to the subscriber
func (ws *WebSocketSubscriber) SendUpdate(update DemoUpdate) error {
	if !ws.IsActive() {
		return fmt.Errorf("subscriber inactive")
	}

	select {
	case ws.updateChan <- update:
		return nil
	default:
		return fmt.Errorf("update channel full")
	}
}

// IsActive returns whether the subscriber is active
func (ws *WebSocketSubscriber) IsActive() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.active
}

// Close closes the subscriber connection
func (ws *WebSocketSubscriber) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.active {
		return nil
	}

	ws.active = false
	close(ws.stopChan)

	// Send close message
	ws.conn.SetWriteDeadline(time.Now().Add(ws.writeTimeout))
	ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	return ws.conn.Close()
}

// SetSubscription sets the update types and filters for the subscriber
func (ws *WebSocketSubscriber) SetSubscription(updateTypes []UpdateType, filters map[string]interface{}) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.updateTypes = updateTypes
	ws.filters = filters

	ws.logger.Debug("Updated subscription", "id", ws.id, "types", updateTypes)
}

// Private methods

func (ws *WebSocketSubscriber) readPump(unregister chan<- *WebSocketSubscriber) {
	defer func() {
		unregister <- ws
	}()

	for {
		select {
		case <-ws.stopChan:
			return
		default:
			_, message, err := ws.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					ws.logger.Error("WebSocket read error", "id", ws.id, "error", err)
				}
				return
			}

			// Process incoming message
			if err := ws.handleMessage(message); err != nil {
				ws.logger.Warn("Failed to handle message", "id", ws.id, "error", err)
			}
		}
	}
}

func (ws *WebSocketSubscriber) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ws.stopChan:
			return
		case update := <-ws.updateChan:
			ws.conn.SetWriteDeadline(time.Now().Add(ws.writeTimeout))

			// Create WebSocket message
			message := WebSocketMessage{
				Type:    string(update.Type),
				Payload: update,
			}

			if err := ws.conn.WriteJSON(message); err != nil {
				ws.logger.Error("Failed to write message", "id", ws.id, "error", err)
				return
			}

		case <-ticker.C:
			// Send ping
			if err := ws.ping(); err != nil {
				return
			}
		}
	}
}

func (ws *WebSocketSubscriber) handleMessage(message []byte) error {
	var wsMessage WebSocketMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	switch wsMessage.Type {
	case "subscribe":
		return ws.handleSubscription(wsMessage.Payload)
	case "unsubscribe":
		return ws.handleUnsubscription(wsMessage.Payload)
	case "ping":
		return ws.handlePing()
	default:
		ws.logger.Debug("Unknown message type", "type", wsMessage.Type, "id", ws.id)
	}

	return nil
}

func (ws *WebSocketSubscriber) handleSubscription(payload interface{}) error {
	data, ok := payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription payload")
	}

	// Parse update types
	var updateTypes []UpdateType
	if types, exists := data["update_types"]; exists {
		if typesList, ok := types.([]interface{}); ok {
			for _, t := range typesList {
				if typeStr, ok := t.(string); ok {
					updateTypes = append(updateTypes, UpdateType(typeStr))
				}
			}
		}
	}

	// Parse filters
	filters := make(map[string]interface{})
	if filterData, exists := data["filters"]; exists {
		if filterMap, ok := filterData.(map[string]interface{}); ok {
			filters = filterMap
		}
	}

	ws.SetSubscription(updateTypes, filters)

	// Send acknowledgment
	ack := WebSocketMessage{
		Type: "subscription_ack",
		Payload: map[string]interface{}{
			"update_types": updateTypes,
			"filters":      filters,
		},
	}

	ws.conn.SetWriteDeadline(time.Now().Add(ws.writeTimeout))
	return ws.conn.WriteJSON(ack)
}

func (ws *WebSocketSubscriber) handleUnsubscription(payload interface{}) error {
	// Clear subscription
	ws.SetSubscription([]UpdateType{}, make(map[string]interface{}))

	// Send acknowledgment
	ack := WebSocketMessage{
		Type: "unsubscription_ack",
		Payload: map[string]interface{}{
			"status": "unsubscribed",
		},
	}

	ws.conn.SetWriteDeadline(time.Now().Add(ws.writeTimeout))
	return ws.conn.WriteJSON(ack)
}

func (ws *WebSocketSubscriber) handlePing() error {
	pong := WebSocketMessage{
		Type: "pong",
		Payload: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}

	ws.conn.SetWriteDeadline(time.Now().Add(ws.writeTimeout))
	return ws.conn.WriteJSON(pong)
}

func (ws *WebSocketSubscriber) ping() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.active {
		return fmt.Errorf("subscriber inactive")
	}

	ws.lastPing = time.Now()
	ws.conn.SetWriteDeadline(time.Now().Add(ws.writeTimeout))
	return ws.conn.WriteMessage(websocket.PingMessage, nil)
}

func (ws *WebSocketSubscriber) shouldReceiveUpdate(update DemoUpdate) bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	// If no specific types are subscribed, send all updates
	if len(ws.updateTypes) == 0 {
		return true
	}

	// Check if update type is subscribed
	for _, subscribedType := range ws.updateTypes {
		if update.Type == subscribedType {
			return ws.matchesFilters(update)
		}
	}

	return false
}

func (ws *WebSocketSubscriber) matchesFilters(update DemoUpdate) bool {
	if len(ws.filters) == 0 {
		return true
	}

	// Apply filters based on update source, type, etc.
	if sourceFilter, exists := ws.filters["source"]; exists {
		if sourceStr, ok := sourceFilter.(string); ok && sourceStr != update.Source {
			return false
		}
	}

	// Add more filter logic as needed

	return true
}

// DemoWebSocketSubscriber adapts WebSocketSubscriber to DemoSubscriber interface
type DemoWebSocketSubscriber struct {
	*WebSocketSubscriber
}

// NewDemoWebSocketSubscriber creates a new demo WebSocket subscriber
func NewDemoWebSocketSubscriber(conn *websocket.Conn, config WebSocketConfig, logger *slog.Logger) DemoSubscriber {
	ws := NewWebSocketSubscriber(conn, config, logger)
	return &DemoWebSocketSubscriber{WebSocketSubscriber: ws}
}

// Integration with DemoController

// WebSocketDemoIntegration provides WebSocket integration for the demo controller
type WebSocketDemoIntegration struct {
	hub        *WebSocketHub
	controller DemoController
	logger     *slog.Logger
}

// NewWebSocketDemoIntegration creates a new WebSocket demo integration
func NewWebSocketDemoIntegration(
	config WebSocketConfig,
	controller DemoController,
	logger *slog.Logger,
) *WebSocketDemoIntegration {
	return &WebSocketDemoIntegration{
		hub:        NewWebSocketHub(config, logger),
		controller: controller,
		logger:     logger,
	}
}

// Start begins the WebSocket integration
func (wdi *WebSocketDemoIntegration) Start(ctx context.Context) error {
	wdi.hub.Start(ctx)

	// Subscribe hub to controller updates
	// This would typically be done through a callback mechanism
	// For now, we'll implement it as a polling mechanism
	go wdi.forwardUpdates(ctx)

	return nil
}

// Stop gracefully shuts down the WebSocket integration
func (wdi *WebSocketDemoIntegration) Stop(ctx context.Context) error {
	return wdi.hub.Stop(ctx)
}

// GetHandler returns the HTTP handler for WebSocket connections
func (wdi *WebSocketDemoIntegration) GetHandler() http.HandlerFunc {
	return wdi.hub.HandleWebSocket
}

// GetStats returns integration statistics
func (wdi *WebSocketDemoIntegration) GetStats() map[string]interface{} {
	return wdi.hub.GetStats()
}

func (wdi *WebSocketDemoIntegration) forwardUpdates(ctx context.Context) {
	// This is a simplified implementation
	// In a real system, you'd implement a proper callback or event system
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get current status and broadcast if changed
			if loadStatus, err := wdi.controller.GetLoadTestStatus(ctx); err == nil {
				update := DemoUpdate{
					Type:      UpdateLoadTestStatus,
					Timestamp: time.Now(),
					Data:      loadStatus,
					Source:    "websocket_integration",
				}
				wdi.hub.BroadcastUpdate(update)
			}

			if chaosStatus, err := wdi.controller.GetChaosTestStatus(ctx); err == nil {
				update := DemoUpdate{
					Type:      UpdateChaosTestStatus,
					Timestamp: time.Now(),
					Data:      chaosStatus,
					Source:    "websocket_integration",
				}
				wdi.hub.BroadcastUpdate(update)
			}

			if systemStatus, err := wdi.controller.GetSystemStatus(ctx); err == nil {
				update := DemoUpdate{
					Type:      UpdateSystemStatus,
					Timestamp: time.Now(),
					Data:      systemStatus,
					Source:    "websocket_integration",
				}
				wdi.hub.BroadcastUpdate(update)
			}
		}
	}
}