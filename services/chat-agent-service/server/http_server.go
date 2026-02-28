package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/agent"
)

type HTTPServer struct {
	chatAgent *agent.ChatAgent
	port      string
}

func NewHTTPServer(chatAgent *agent.ChatAgent, port string) *HTTPServer {
	return &HTTPServer{
		chatAgent: chatAgent,
		port:      port,
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

	addr := ":" + s.port
	log.Printf("🚀 Chat Agent Service starting on port %s", s.port)
	log.Printf("📡 Endpoints:")
	log.Printf("   GET  /health           - Health check")
	log.Printf("   POST /api/v1/chat      - Chat with AI agent")
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
		s.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Message == "" {
		s.sendError(w, http.StatusBadRequest, "Message is required")
		return
	}

	log.Printf("Processing chat message: %s", req.Message)

	resp, err := s.chatAgent.ProcessMessage(req.Message)
	if err != nil {
		log.Printf("Error processing message: %v", err)
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Error: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (s *HTTPServer) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
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

