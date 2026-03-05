package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/agent"
	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/client"
	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/session"
	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/validator"
)

type HTTPServer struct {
	chatAgent      *agent.ChatAgent
	sessionManager *session.RedisSessionManager
	port           string
}

func NewHTTPServer(chatAgent *agent.ChatAgent, sessionManager *session.RedisSessionManager, port string) *HTTPServer {
	return &HTTPServer{
		chatAgent:      chatAgent,
		sessionManager: sessionManager,
		port:           port,
	}
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.healthHandler)

	// Chat endpoint
	mux.HandleFunc("/api/v1/chat", s.chatHandler)

	// CORS middleware
	handler := s.enableCORS(mux)

	addr := "0.0.0.0:" + s.port
	log.Printf("🚀 Chat Agent Service starting on 0.0.0.0:%s", s.port)
	log.Printf("📡 Endpoints:")
	log.Printf("   GET  /health           - Health check")
	log.Printf("   POST /api/v1/chat      - Chat with AI agent (Markdown response)")
	log.Println()

	return http.ListenAndServe(addr, handler)
}

func (s *HTTPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "chat-agent-service",
		"version": "1.0.0",
	})
}

func (s *HTTPServer) chatHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req agent.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendMarkdownError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = generateSessionID()
		log.Printf("Generated new session ID: %s", sessionID)
	}
	// GUARDRAIL STEP 1: Validate input
	validationResult, err := validator.ValidateInput(req.Message)

	// Log validation result for security monitoring
	validator.LogValidationResult(r.RemoteAddr, req.Message, validationResult, err)

	if err != nil {
		s.sendMarkdownError(w, http.StatusBadRequest, fmt.Sprintf("Input validation failed: %v", err))
		return
	}

	// GUARDRAIL STEP 2: Optional PII detection and masking
	if hasPII, piiTypes := validator.DetectPII(validationResult.Sanitized); hasPII {
		log.Printf("[PII_DETECTED] IP=%s Types=%v", r.RemoteAddr, piiTypes)
		// Optionally mask PII before sending to LLM
		// validationResult.Sanitized = validator.MaskPII(validationResult.Sanitized)
	}

	// Use sanitized input
	req.Message = validationResult.Sanitized

	log.Printf("Processing chat message (tokens: ~%d, risk: %d): %s",
		validator.EstimateTokens(req.Message), validationResult.RiskScore, req.Message)

	history, err := s.sessionManager.GetHistory(sessionID)
	if err != nil {
		log.Printf("[ERROR] Failed to get session history: %v", err)
		history = []client.ChatMessage{}
	}
	resp, err := s.chatAgent.ProcessMessage(req.Message, history)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		s.sendMarkdownError(w, http.StatusInternalServerError, fmt.Sprintf("Error: %v", err))
		return
	}
	if err := s.sessionManager.AddMessage(sessionID, client.ChatMessage{Role: "user", Content: req.Message}); err != nil {
		log.Printf("[ERROR] Failed to add user message to session: %v", err)
	}
	if err := s.sessionManager.AddMessage(sessionID, client.ChatMessage{Role: "assistant", Content: resp.Response}); err != nil {
		log.Printf("[ERROR] Failed to add assistant message to session: %v", err)
	}

	// Return as Markdown format
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("X-Session-ID", sessionID) // Return session ID to client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp.Response))
}

func (s *HTTPServer) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (s *HTTPServer) sendMarkdownError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(status)
	errorMsg := fmt.Sprintf("# Error\n\n%s", message)
	w.Write([]byte(errorMsg))
}

func (s *HTTPServer) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}
