package handler

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	clients   map[string]map[*websocket.Conn]bool // runID -> connections
	broadcast chan WebSocketMessage
	mu        sync.RWMutex
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	h := &WebSocketHandler{
		clients:   make(map[string]map[*websocket.Conn]bool),
		broadcast: make(chan WebSocketMessage, 256),
	}

	// Start broadcast goroutine
	go h.handleBroadcast()

	return h
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	RunID     string      `json:"run_id"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// HandleWebSocket handles WebSocket connections for a specific run
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	runID := c.Param("runId")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	h.mu.Lock()
	if h.clients[runID] == nil {
		h.clients[runID] = make(map[*websocket.Conn]bool)
	}
	h.clients[runID][conn] = true
	h.mu.Unlock()

	log.Printf("WebSocket client connected for run: %s", runID)

	// Send initial status
	h.sendInitialStatus(conn, runID)

	// Unregister client on disconnect
	defer func() {
		h.mu.Lock()
		delete(h.clients[runID], conn)
		if len(h.clients[runID]) == 0 {
			delete(h.clients, runID)
		}
		h.mu.Unlock()
		log.Printf("WebSocket client disconnected for run: %s", runID)
	}()

	// Read messages from client (for subscription management)
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle client messages (e.g., subscribe, unsubscribe)
		log.Printf("Received message from client: %v", msg)
	}
}

// sendInitialStatus sends initial status to newly connected client
func (h *WebSocketHandler) sendInitialStatus(conn *websocket.Conn, runID string) {
	msg := WebSocketMessage{
		Type:  "connected",
		RunID: runID,
		Data: map[string]interface{}{
			"message": "Connected to workflow updates",
		},
		Timestamp: time.Now(),
	}

	conn.WriteJSON(msg)
}

// handleBroadcast handles broadcasting messages to clients
func (h *WebSocketHandler) handleBroadcast() {
	for msg := range h.broadcast {
		h.mu.RLock()
		clients := h.clients[msg.RunID]
		h.mu.RUnlock()

		for conn := range clients {
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				conn.Close()
				h.mu.Lock()
				delete(h.clients[msg.RunID], conn)
				h.mu.Unlock()
			}
		}
	}
}

// BroadcastWorkflowStatus broadcasts workflow status update
func (h *WebSocketHandler) BroadcastWorkflowStatus(runID, phase, status string, progress int) {
	msg := WebSocketMessage{
		Type:  "workflow_status",
		RunID: runID,
		Data: map[string]interface{}{
			"phase":    phase,
			"status":   status,
			"progress": progress,
		},
		Timestamp: time.Now(),
	}

	h.broadcast <- msg
}

// BroadcastAgentActivity broadcasts agent activity update
func (h *WebSocketHandler) BroadcastAgentActivity(runID, agent, status, message string, details map[string]interface{}) {
	msg := WebSocketMessage{
		Type:  "agent_activity",
		RunID: runID,
		Data: map[string]interface{}{
			"agent":   agent,
			"status":  status,
			"message": message,
			"details": details,
		},
		Timestamp: time.Now(),
	}

	h.broadcast <- msg
}

// BroadcastLog broadcasts log entry
func (h *WebSocketHandler) BroadcastLog(runID, level, message, agent string, details map[string]interface{}) {
	msg := WebSocketMessage{
		Type:  "log",
		RunID: runID,
		Data: map[string]interface{}{
			"level":   level,
			"message": message,
			"agent":   agent,
			"details": details,
		},
		Timestamp: time.Now(),
	}

	h.broadcast <- msg
}

// BroadcastPhaseComplete broadcasts phase completion
func (h *WebSocketHandler) BroadcastPhaseComplete(runID, phase string, durationMS int64, metadata map[string]interface{}) {
	msg := WebSocketMessage{
		Type:  "phase_complete",
		RunID: runID,
		Data: map[string]interface{}{
			"phase":       phase,
			"duration_ms": durationMS,
			"metadata":    metadata,
		},
		Timestamp: time.Now(),
	}

	h.broadcast <- msg
}

// SimulateWorkflow simulates workflow updates for testing
func (h *WebSocketHandler) SimulateWorkflow(runID string) {
	go func() {
		// Phase 1: Spec Analysis
		h.BroadcastWorkflowStatus(runID, "phase_1", "in_progress", 10)
		h.BroadcastLog(runID, "info", "Starting Phase 1: Spec Analysis", "spec_analyzer", nil)
		time.Sleep(2 * time.Second)

		h.BroadcastAgentActivity(runID, "spec_analyzer", "active", "Parsing OpenAPI specification", map[string]interface{}{
			"endpoints_found": 3,
		})
		time.Sleep(2 * time.Second)

		h.BroadcastWorkflowStatus(runID, "phase_1", "completed", 33)
		h.BroadcastPhaseComplete(runID, "phase_1", 4000, map[string]interface{}{
			"endpoints_analyzed": 3,
			"schemas_extracted":  5,
		})

		// Phase 2: Test Generation
		time.Sleep(1 * time.Second)
		h.BroadcastWorkflowStatus(runID, "phase_2", "in_progress", 40)
		h.BroadcastLog(runID, "info", "Starting Phase 2: Test Generation", "payload_generator", nil)

		agents := []string{"smart_data_generator", "mutation_generator", "security_generator", "performance_generator"}
		for i, agent := range agents {
			time.Sleep(2 * time.Second)
			h.BroadcastAgentActivity(runID, agent, "active", "Generating test payloads", map[string]interface{}{
				"payloads_generated": 10 + i*5,
			})
			h.BroadcastWorkflowStatus(runID, "phase_2", "in_progress", 40+i*5)
		}

		time.Sleep(2 * time.Second)
		h.BroadcastWorkflowStatus(runID, "phase_2", "completed", 66)
		h.BroadcastPhaseComplete(runID, "phase_2", 12000, map[string]interface{}{
			"tests_generated": 45,
		})

		// Phase 3: Test Execution
		time.Sleep(1 * time.Second)
		h.BroadcastWorkflowStatus(runID, "phase_3", "in_progress", 70)
		h.BroadcastLog(runID, "info", "Starting Phase 3: Test Execution", "test_executor", nil)

		for i := 0; i < 5; i++ {
			time.Sleep(2 * time.Second)
			h.BroadcastAgentActivity(runID, "test_executor", "active", "Executing tests", map[string]interface{}{
				"tests_executed": (i + 1) * 9,
				"tests_passed":   (i + 1) * 8,
			})
			h.BroadcastWorkflowStatus(runID, "phase_3", "in_progress", 70+i*5)
		}

		time.Sleep(2 * time.Second)
		h.BroadcastLog(runID, "warn", "Schema drift detected: field 'created_at' added", "schema_drift_detector", map[string]interface{}{
			"endpoint":   "GET /users",
			"drift_type": "field_added",
		})

		time.Sleep(1 * time.Second)
		h.BroadcastWorkflowStatus(runID, "phase_3", "completed", 100)
		h.BroadcastPhaseComplete(runID, "phase_3", 18000, map[string]interface{}{
			"tests_executed": 45,
			"tests_passed":   40,
			"tests_failed":   3,
			"tests_skipped":  2,
		})

		h.BroadcastLog(runID, "info", "Workflow completed successfully", "system", nil)
	}()
}
