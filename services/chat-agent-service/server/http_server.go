package server

import (
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jeromelp/gtp_backend_1/services/chat-agent-service/agent"
    "github.com/jeromelp/gtp_backend_1/services/chat-agent-service/client"
    "github.com/jeromelp/gtp_backend_1/services/chat-agent-service/session"
    "github.com/jeromelp/gtp_backend_1/services/chat-agent-service/validator"
)

type HTTPServer struct {
    chatAgent      *agent.ChatAgent
    sessionManager *session.RedisSessionManager
    port           string
    router         *gin.Engine
}

type HTTPChatResponse struct {
    Status    string `json:"status"`      // "success" or "error"
    Message   string `json:"message"`     // Status message
    SessionID string `json:"session_id"`  // For conversation memory
    Response  string `json:"response"`    // Markdown content from Groq
}

func NewHTTPServer(chatAgent *agent.ChatAgent, sessionManager *session.RedisSessionManager, port string) *HTTPServer {
    router := gin.New()
    
    // Middleware
    router.Use(gin.Recovery())
    router.Use(gin.Logger())
    router.Use(corsMiddleware())
    
    server := &HTTPServer{
        chatAgent:      chatAgent,
        sessionManager: sessionManager,
        port:           port,
        router:         router,
    }
    
    server.setupRoutes()
    return server
}

func (s *HTTPServer) setupRoutes() {
    s.router.GET("/health", s.healthHandler)
    
    v1 := s.router.Group("/api/v1")
    {
        v1.POST("/chat", s.chatHandler)
    }
}

func (s *HTTPServer) Start() error {
    addr := "0.0.0.0:" + s.port
    log.Printf("Chat Agent Service starting on 0.0.0.0:%s", s.port)
    log.Printf("Endpoints:")
    log.Printf("   GET  /health           - Health check")
    log.Printf("   POST /api/v1/chat      - Chat with AI agent (JSON with Markdown response)")
    log.Println()
    
    return s.router.Run(addr)
}

func (s *HTTPServer) healthHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "healthy",
        "service": "chat-agent-service",
        "version": "1.0.0",
    })
}

func (s *HTTPServer) chatHandler(c *gin.Context) {
    var req agent.ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        s.sendError(c, http.StatusBadRequest, "Invalid request body", "")
        return
    }
    
    sessionID := req.SessionID
    if sessionID == "" {
        sessionID = generateSessionID()
        log.Printf("Generated new session ID: %s", sessionID)
    }
    
    // GUARDRAIL STEP 1: Validate input
    validationResult, err := validator.ValidateInput(req.Message)
    validator.LogValidationResult(c.ClientIP(), req.Message, validationResult, err)
    
    if err != nil {
        s.sendError(c, http.StatusBadRequest, fmt.Sprintf("Input validation failed: %v", err), sessionID)
        return
    }
    
    // GUARDRAIL STEP 2: Optional PII detection
    if hasPII, piiTypes := validator.DetectPII(validationResult.Sanitized); hasPII {
        log.Printf("[PII_DETECTED] IP=%s Types=%v", c.ClientIP(), piiTypes)
    }
    
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
        s.sendError(c, http.StatusInternalServerError, fmt.Sprintf("Error: %v", err), sessionID)
        return
    }
    
    if err := s.sessionManager.AddMessage(sessionID, client.ChatMessage{Role: "user", Content: req.Message}); err != nil {
        log.Printf("[ERROR] Failed to add user message to session: %v", err)
    }
    if err := s.sessionManager.AddMessage(sessionID, client.ChatMessage{Role: "assistant", Content: resp.Response}); err != nil {
        log.Printf("[ERROR] Failed to add assistant message to session: %v", err)
    }
    
    c.JSON(http.StatusOK, HTTPChatResponse{
        Status:    "success",
        Message:   "Chat response generated successfully",
        SessionID: sessionID,
        Response:  resp.Response,
    })
}

func (s *HTTPServer) sendError(c *gin.Context, status int, message string, sessionID string) {
    c.JSON(status, HTTPChatResponse{
        Status:    "error",
        Message:   message,
        SessionID: sessionID,
        Response:  "",
    })
}

func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusOK)
            return
        }
        
        c.Next()
    }
}

func generateSessionID() string {
    return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}