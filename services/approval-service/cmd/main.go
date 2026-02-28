package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/config"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	controllerv1 "github.com/jeromelp/gtp_backend_1/services/approval-service/controller/v1"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/routes"
	servicev1 "github.com/jeromelp/gtp_backend_1/services/approval-service/service/v1"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v\n", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Service Name: %s", cfg.ServiceName)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Log Level: %s", cfg.LogLevel)
	log.Printf("Slack Bot Name: %s", cfg.SlackBotName)
	log.Printf("Slack Bot ID: %s", cfg.SlackBotID)

	if cfg.LogLevel == constants.LogLevelRelease {
		gin.SetMode(gin.ReleaseMode)
		log.Printf("Gin mode set to: RELEASE")
	} else {
		gin.SetMode(gin.DebugMode)
		log.Printf("Gin mode set to: DEBUG")
	}

	slackService := servicev1.NewSlackService(cfg.SlackBotToken)
	log.Printf("Slack service initialized")

	slackController := controllerv1.NewSlackController(slackService)
	log.Printf("Controllers initialized")

	router := routes.SetupRoutes(slackController)
	log.Printf("Routes configured")

	addr := fmt.Sprintf("%s:%s", cfg.ServiceHost, cfg.ServicePort)

	log.Printf("Starting %s on %s", cfg.ServiceName, addr)

	log.Printf("API Endpoints:")
	log.Printf("   GET  /health                              - Health check")
	log.Printf("   POST /api/v1/slack/channel/create         - Create Slack channel")
	log.Printf("   POST /api/v1/slack/member/add             - Add member to channel")
	log.Printf("   GET  /api/v1/slack/users/all              - Get all users")
	log.Printf("   POST /api/v1/slack/user/by-name           - Get user by name")
	log.Printf("   POST /api/v1/slack/user/by-id             - Get user by ID")
	log.Printf("   GET  /api/v1/slack/channels/all           - Get all channels")
	log.Printf("   POST /api/v1/slack/channel/by-name        - Get channel by name")
	log.Printf("   POST /api/v1/slack/channel/by-id          - Get channel by ID")
	log.Printf("   POST /api/v1/slack/message/send           - Send message to channel")

	log.Println()

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
