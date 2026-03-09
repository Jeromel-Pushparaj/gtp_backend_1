package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/config"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/constants"
	controllerv1 "github.com/jeromelp/gtp_backend_1/services/approval-service/controller/v1"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/db"
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

	if cfg.LogLevel == constants.LogLevelRelease {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	database, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("Database initialization error: %v\n", err)
	}
	defer database.Close()
	log.Printf("Database initialized")

	approvalRepo := db.NewApprovalRepository(database)

	kafkaService, err := servicev1.NewKafkaService(cfg.KafkaBrokers, cfg.KafkaGroupID)
	if err != nil {
		log.Fatalf("Kafka service initialization error: %v\n", err)
	}
	defer kafkaService.Close()
	log.Printf("Kafka service initialized")

	slackService := servicev1.NewSlackService(cfg.SlackBotToken)
	log.Printf("Slack service initialized")

	approvalService := servicev1.NewApprovalService(approvalRepo, slackService, kafkaService)
	log.Printf("Approval service initialized")

	consumerService := servicev1.NewConsumerService(kafkaService, approvalService)
	if err := consumerService.Start(); err != nil {
		log.Fatalf("Consumer service start error: %v\n", err)
	}
	log.Printf("Kafka consumer started")

	businessConsumer, err := servicev1.NewBusinessConsumerService(cfg.KafkaBrokers, "business-consumer-group", kafkaService)
	if err != nil {
		log.Fatalf("Business consumer service initialization error: %v\n", err)
	}
	defer businessConsumer.Close()
	if err := businessConsumer.Start(); err != nil {
		log.Fatalf("Business consumer service start error: %v\n", err)
	}
	log.Printf("Business Kafka consumer started")

	socketHandler := servicev1.NewSocketHandler(cfg.SlackBotToken, cfg.SlackAppToken, approvalService)
	go socketHandler.Start()
	log.Printf("Slack Socket Mode handler started")

	slackController := controllerv1.NewSlackController(slackService)
	approvalController := controllerv1.NewApprovalController(approvalRepo, slackService, kafkaService)
	log.Printf("Controllers initialized")

	router := routes.SetupRoutes(slackController, approvalController)
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
	log.Printf("   GET  /api/v1/approval/all                 - Get all approvals")
	log.Printf("   GET  /api/v1/approval/pending             - Get pending approvals")
	log.Printf("   POST /api/v1/approval/by-id               - Get approval by ID")

	log.Printf("   POST /api/v1/slack/dm-channel/get         - Get DM channel id")
	log.Printf("   POST /api/v1/approval/domain-change       - Create domain change approval request")

	log.Printf("   POST /api/v1/approval/request             - Create approval request (publishes to Kafka)")
	log.Printf("Slack Socket Mode: ENABLED")
	log.Printf("Kafka Consumer 1: Listening on %s", constants.KafkaTopicApprovalRequested)
	log.Printf("Kafka Consumer 2: Listening on %s", constants.KafkaTopicApprovalCompleted)
	log.Printf("Kafka Producer: Publishing to %s, %s, %s", constants.KafkaTopicApprovalCompleted, constants.KafkaTopicActionExecuted, constants.KafkaTopicActionRejected)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
