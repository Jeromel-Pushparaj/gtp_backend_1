package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/gateway/api-gateway/config"
	"github.com/jeromelp/gtp_backend_1/gateway/api-gateway/handlers"
	"github.com/jeromelp/gtp_backend_1/gateway/api-gateway/middleware"
)

func main() {
	// Print banner
	printBanner()

	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Setup middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS(cfg))

	// Setup rate limiting
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitDuration)
	router.Use(rateLimiter.RateLimit())

	// Create proxy handler
	proxyHandler := handlers.NewProxyHandler(cfg)

	// Health check endpoint
	router.GET("/health", proxyHandler.HealthCheck())

	// Jira Trigger Service Routes - Port 8086
	// Handles Jira issue creation and management
	jiraGroup := router.Group("/jira")
	{
		jiraGroup.Any("/*path", proxyHandler.JiraTriggerService())
	}

	// Chat Agent Service Routes - Port 8082
	// Handles AI-powered chat interactions
	chatGroup := router.Group("/chat")
	{
		chatGroup.Any("/*path", proxyHandler.ChatAgentService())
	}

	// Approval Service Routes - Port 8083
	// Handles Slack approval workflows and notifications
	approvalGroup := router.Group("/approval")
	{
		approvalGroup.Any("/*path", proxyHandler.ApprovalService())
	}

	// Onboarding Service Routes - Port 8084
	// Handles service catalog and onboarding
	onboardingGroup := router.Group("/service")
	{
		onboardingGroup.Any("/*path", proxyHandler.OnboardingService())
	}

	// ScoreCard Service Routes - Port 8085
	// Handles service scorecard evaluations
	scorecardGroup := router.Group("/scorecard")
	{
		scorecardGroup.Any("/*path", proxyHandler.ScoreCardService())
	}

	// SonarShell Service Routes - Port 8080
	// Handles SonarCloud automation and metrics
	sonarGroup := router.Group("/sonar")
	{
		sonarGroup.Any("/*path", proxyHandler.SonarShellService())
	}

	pagerDutyGroup := router.Group("/pd")
	{
		pagerDutyGroup.Any("/*path", proxyHandler.PagerDutyService())
	}

	// Print route information
	printRoutes(cfg)

	// Configure HTTP server with timeouts
	address := fmt.Sprintf("%s:%s", cfg.GatewayHost, cfg.GatewayPort)
	server := &http.Server{
		Addr:           address,
		Handler:        router,
		ReadTimeout:    cfg.ServerReadTimeout,
		WriteTimeout:   cfg.ServerWriteTimeout,
		IdleTimeout:    cfg.ServerIdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	log.Printf("INFO: API Gateway starting on %s (Environment: %s)", address, cfg.Environment)
	log.Printf("INFO: Rate Limit: %d requests per %v", cfg.RateLimitRequests, cfg.RateLimitDuration)
	log.Printf("INFO: Server Timeouts - Read: %v, Write: %v, Idle: %v",
		cfg.ServerReadTimeout, cfg.ServerWriteTimeout, cfg.ServerIdleTimeout)
	log.Println("INFO: All systems operational")

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ERROR: Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("INFO: Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("ERROR: Server forced to shutdown: %v", err)
	}

	log.Println("INFO: Server exited gracefully")
}

// printBanner prints the application banner
func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║          GTP Backend API Gateway                          ║
║                                                           ║
║     Connecting all your microservices                     ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

// printRoutes prints all available routes and their backend services
func printRoutes(cfg *config.Config) {
	log.Println("\nAvailable Routes:")
	log.Println("================================================================")
	log.Printf("  GET    /health                    -> Gateway Health Check")
	log.Println("================================================================")
	log.Printf("  ANY    /jira/*                    -> %s", cfg.JiraTriggerServiceURL)
	log.Printf("  ANY    /chat/*                    -> %s", cfg.ChatAgentServiceURL)
	log.Printf("  ANY    /approval/*                -> %s", cfg.ApprovalServiceURL)
	log.Printf("  ANY    /onboarding/*              -> %s", cfg.OnboardingServiceURL)
	log.Printf("  ANY    /scorecard/*               -> %s", cfg.ScoreCardServiceURL)
	log.Printf("  ANY    /sonar/*                   -> %s", cfg.SonarShellServiceURL)
	log.Printf("  ANY    /pd/*                      -> %s", cfg.PagerDutyServiceURL)
	log.Println("================================================================")
	log.Println()
}
