package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/gateway/api-gateway/config"
)

// ProxyHandler handles proxying requests to backend services
type ProxyHandler struct {
	config *config.Config
	client *http.Client
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(cfg *config.Config) *ProxyHandler {
	return &ProxyHandler{
		config: cfg,
		client: &http.Client{},
	}
}

// ProxyRequest is a generic proxy function that forwards requests to backend services
func (h *ProxyHandler) ProxyRequest(targetURL string, stripPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Build the target URL
		path := c.Request.URL.Path
		if stripPrefix != "" {
			path = strings.TrimPrefix(path, stripPrefix)
		}

		// Construct full URL with query parameters
		fullURL := targetURL + path
		if c.Request.URL.RawQuery != "" {
			fullURL += "?" + c.Request.URL.RawQuery
		}

		log.Printf("PROXY: %s %s -> %s", c.Request.Method, c.Request.URL.Path, fullURL)

		// Read request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create new request
		req, err := http.NewRequest(c.Request.Method, fullURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			log.Printf("ERROR: Error creating request: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create proxy request",
				"message": err.Error(),
			})
			return
		}

		// Copy headers from original request
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Send request to backend service
		resp, err := h.client.Do(req)
		if err != nil {
			log.Printf("ERROR: Error forwarding request to %s: %v", fullURL, err)
			c.JSON(http.StatusBadGateway, gin.H{
				"error":   "Service unavailable",
				"message": fmt.Sprintf("Failed to connect to backend service: %v", err),
			})
			return
		}
		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ERROR: Error reading response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to read response",
				"message": err.Error(),
			})
			return
		}

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}

		// Send response back to client
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}

// JiraTriggerService - handles Jira issue creation
func (h *ProxyHandler) JiraTriggerService() gin.HandlerFunc {
	return h.ProxyRequest(h.config.JiraTriggerServiceURL, "/jira")
}

// ChatAgentService - handles AI chat interactions
func (h *ProxyHandler) ChatAgentService() gin.HandlerFunc {
	return h.ProxyRequest(h.config.ChatAgentServiceURL, "/chat")
}

// ApprovalService - handles Slack approval workflows
func (h *ProxyHandler) ApprovalService() gin.HandlerFunc {
	return h.ProxyRequest(h.config.ApprovalServiceURL, "/approval")
}

// OnboardingService - handles service catalog and onboarding
func (h *ProxyHandler) OnboardingService() gin.HandlerFunc {
	return h.ProxyRequest(h.config.OnboardingServiceURL, "/service")
}

// ScoreCardService - handles service scorecard evaluations
func (h *ProxyHandler) ScoreCardService() gin.HandlerFunc {
	return h.ProxyRequest(h.config.ScoreCardServiceURL, "/scorecard")
}

// SonarShellService - handles SonarCloud automation
func (h *ProxyHandler) SonarShellService() gin.HandlerFunc {
	return h.ProxyRequest(h.config.SonarShellServiceURL, "/sonar")
}

func (h *ProxyHandler) PagerDutyService() gin.HandlerFunc {
	return h.ProxyRequest(h.config.PagerDutyServiceURL, "/pd")
}

// HealthCheck returns the gateway health status
func (h *ProxyHandler) HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api-gateway",
			"version": "1.0.0",
			"message": "Gateway is running smoothly",
		})
	}
}
