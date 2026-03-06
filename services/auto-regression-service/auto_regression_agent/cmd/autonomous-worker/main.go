package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/agents/autonomous"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/config"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/events"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/vectordb"
)

func main() {
	// Parse flags
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	agentType := flag.String("agent", "all", "Agent type to run: all, discovery, designer, payload, executor, analyzer, feedback")
	flag.Parse()

	if *agentType == "all" {
		log.Println("🤖 Starting Autonomous AI Agent System (ALL AGENTS)...")
	} else {
		log.Printf("🤖 Starting %s Agent...\n", *agentType)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		MaxRetries:   cfg.Redis.MaxRetries,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	// Initialize LLM client
	if cfg.LLM.APIKey == "" {
		log.Fatalf("LLM API key not configured. Set LLM_API_KEY or OPENAI_API_KEY environment variable")
	}

	llmClient := llm.NewClient(llm.Config{
		APIKey:   cfg.LLM.APIKey,
		BaseURL:  cfg.LLM.BaseURL,
		Model:    cfg.LLM.Model,
		Provider: cfg.LLM.Provider,
		Timeout:  time.Duration(cfg.LLM.Timeout) * time.Second,
	})
	log.Printf("✅ LLM client initialized: provider=%s, model=%s", cfg.LLM.Provider, cfg.LLM.Model)

	// Initialize event system
	eventBus := events.NewBus(redisClient)
	messageBus := events.NewMessageBus(redisClient)
	consensusEngine := events.NewConsensusEngine(redisClient, eventBus)
	log.Println("✅ Event system initialized")

	// Initialize vector store for FeedbackAgent (optional)
	var vectorStore *vectordb.Store
	if cfg.VectorDB.Enabled && cfg.VectorDB.Host != "" {
		log.Printf("🔗 Connecting to vector database at %s:%d...", cfg.VectorDB.Host, cfg.VectorDB.Port)
		embClient := vectordb.NewEmbeddingClient(vectordb.EmbeddingConfig{
			APIKey:  cfg.LLM.APIKey,
			BaseURL: cfg.LLM.BaseURL,
		})
		vectorStore, err = vectordb.NewStore(vectordb.StoreConfig{
			Host:     cfg.VectorDB.Host,
			Port:     cfg.VectorDB.Port,
			User:     cfg.VectorDB.User,
			Password: cfg.VectorDB.Password,
			Database: cfg.VectorDB.Database,
			SSLMode:  cfg.VectorDB.SSLMode,
		}, embClient)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to connect to vector store: %v", err)
			log.Printf("⚠️  FeedbackAgent will run without learning capabilities")
			vectorStore = nil
		} else {
			log.Println("✅ Vector database connected")
		}
	} else {
		log.Println("ℹ️  Vector database not configured, FeedbackAgent learning disabled")
	}

	// Create and start agents based on agent type
	var agentsToStop []interface{ Stop() }

	switch *agentType {
	case "all":
		log.Println("🤖 Creating all autonomous agents...")

		discoveryAgent := autonomous.NewDiscoveryAgent(llmClient, eventBus, messageBus, consensusEngine)
		designerAgent := autonomous.NewDesignerAgent(llmClient, eventBus, messageBus, consensusEngine)
		payloadAgent := autonomous.NewPayloadAgent(llmClient, eventBus, messageBus, consensusEngine)
		executorAgent := autonomous.NewExecutorAgent(llmClient, eventBus, messageBus, consensusEngine)
		analyzerAgent := autonomous.NewAnalyzerAgent(llmClient, eventBus, messageBus, consensusEngine)
		feedbackAgent := autonomous.NewFeedbackAgent(llmClient, eventBus, messageBus, consensusEngine, vectorStore)

		log.Println("✅ All agents created (including FeedbackAgent)")
		log.Println("🚀 Starting all agents...")

		if err := discoveryAgent.Start(ctx); err != nil {
			log.Fatalf("Failed to start discovery agent: %v", err)
		}
		if err := designerAgent.Start(ctx); err != nil {
			log.Fatalf("Failed to start designer agent: %v", err)
		}
		if err := payloadAgent.Start(ctx); err != nil {
			log.Fatalf("Failed to start payload agent: %v", err)
		}
		if err := executorAgent.Start(ctx); err != nil {
			log.Fatalf("Failed to start executor agent: %v", err)
		}
		if err := analyzerAgent.Start(ctx); err != nil {
			log.Fatalf("Failed to start analyzer agent: %v", err)
		}
		if err := feedbackAgent.Start(ctx); err != nil {
			log.Fatalf("Failed to start feedback agent: %v", err)
		}

		agentsToStop = []interface{ Stop() }{discoveryAgent, designerAgent, payloadAgent, executorAgent, analyzerAgent, feedbackAgent}

		log.Println("✅ All agents started and listening")
		log.Println("")
		log.Println("═══════════════════════════════════════════════════════════")
		log.Println("🤖 AUTONOMOUS AI AGENT SYSTEM READY")
		log.Println("═══════════════════════════════════════════════════════════")
		log.Println("")
		log.Println("Active Agents:")
		log.Println("  🔍 Discovery Agent   - Analyzes OpenAPI specs with AI")
		log.Println("  🎨 Designer Agent    - Designs test strategies with AI")
		log.Println("  🎲 Payload Agent     - Generates test data with AI")
		log.Println("  ⚡ Executor Agent    - Executes tests")
		log.Println("  📊 Analyzer Agent    - Analyzes results and provides feedback")
		log.Println("  🧠 Feedback Agent    - Learns patterns and stores in vector DB")
		log.Println("")
		log.Println("Agents are now autonomous and will collaborate automatically!")
		log.Println("Upload a Swagger/OpenAPI spec to see them in action.")
		log.Println("")
		log.Println("Press Ctrl+C to shutdown...")
		log.Println("═══════════════════════════════════════════════════════════")

	case "discovery":
		log.Println("🔍 Creating Discovery Agent...")
		agent := autonomous.NewDiscoveryAgent(llmClient, eventBus, messageBus, consensusEngine)
		if err := agent.Start(ctx); err != nil {
			log.Fatalf("Failed to start discovery agent: %v", err)
		}
		agentsToStop = []interface{ Stop() }{agent}
		log.Println("✅ Discovery Agent started and listening for spec_uploaded events")
		log.Println("🔍 Will analyze OpenAPI specs using GPT-4")
		log.Println("Press Ctrl+C to shutdown...")

	case "designer":
		log.Println("🎨 Creating Designer Agent...")
		agent := autonomous.NewDesignerAgent(llmClient, eventBus, messageBus, consensusEngine)
		if err := agent.Start(ctx); err != nil {
			log.Fatalf("Failed to start designer agent: %v", err)
		}
		agentsToStop = []interface{ Stop() }{agent}
		log.Println("✅ Designer Agent started and listening for spec_analyzed events")
		log.Println("🎨 Will create test strategies using GPT-4 and request consensus")
		log.Println("Press Ctrl+C to shutdown...")

	case "payload":
		log.Println("🎲 Creating Payload Agent...")
		agent := autonomous.NewPayloadAgent(llmClient, eventBus, messageBus, consensusEngine)
		if err := agent.Start(ctx); err != nil {
			log.Fatalf("Failed to start payload agent: %v", err)
		}
		agentsToStop = []interface{ Stop() }{agent}
		log.Println("✅ Payload Agent started and listening for strategy_approved events")
		log.Println("🎲 Will generate test data using GPT-3.5")
		log.Println("Press Ctrl+C to shutdown...")

	case "executor":
		log.Println("⚡ Creating Executor Agent...")
		agent := autonomous.NewExecutorAgent(llmClient, eventBus, messageBus, consensusEngine)
		if err := agent.Start(ctx); err != nil {
			log.Fatalf("Failed to start executor agent: %v", err)
		}
		agentsToStop = []interface{ Stop() }{agent}
		log.Println("✅ Executor Agent started and listening for payloads_ready events")
		log.Println("⚡ Will execute HTTP tests")
		log.Println("Press Ctrl+C to shutdown...")

	case "analyzer":
		log.Println("📊 Creating Analyzer Agent...")
		agent := autonomous.NewAnalyzerAgent(llmClient, eventBus, messageBus, consensusEngine)
		if err := agent.Start(ctx); err != nil {
			log.Fatalf("Failed to start analyzer agent: %v", err)
		}
		agentsToStop = []interface{ Stop() }{agent}
		log.Println("✅ Analyzer Agent started and listening for tests_complete events")
		log.Println("📊 Will analyze results using GPT-4 and provide feedback")
		log.Println("Press Ctrl+C to shutdown...")

	case "feedback":
		log.Println("🧠 Creating Feedback Agent...")
		agent := autonomous.NewFeedbackAgent(llmClient, eventBus, messageBus, consensusEngine, vectorStore)
		if err := agent.Start(ctx); err != nil {
			log.Fatalf("Failed to start feedback agent: %v", err)
		}
		agentsToStop = []interface{ Stop() }{agent}
		log.Println("✅ Feedback Agent started and listening for test results")
		log.Println("🧠 Will learn patterns and store in vector database")
		log.Println("Press Ctrl+C to shutdown...")

	default:
		log.Fatalf("Unknown agent type: %s. Valid options: all, discovery, designer, payload, executor, analyzer, feedback", *agentType)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("")
	log.Println("🛑 Shutting down autonomous agent system...")

	// Stop all agents
	for _, agent := range agentsToStop {
		agent.Stop()
	}

	// Close event bus
	eventBus.Close()

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("✅ Shutdown complete")
}
