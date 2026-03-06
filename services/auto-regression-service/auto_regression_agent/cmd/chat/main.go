package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/chat"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/config"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/vectordb"
)

var (
	workflowID    = flag.String("workflow", "", "Workflow ID to load context from (required)")
	configPath    = flag.String("config", "configs/config.yaml", "Path to configuration file")
	resumeSession = flag.String("resume", "", "Resume a previous session by ID")
	noColor       = flag.Bool("no-color", false, "Disable colored output")
)

func main() {
	flag.Parse()

	if *workflowID == "" && *resumeSession == "" {
		fmt.Println("Error: --workflow or --resume flag is required")
		fmt.Println("Usage: opentest-chat --workflow <workflow_id>")
		fmt.Println("       opentest-chat --resume <session_id>")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	// Initialize vector store (optional)
	var vectorStore *vectordb.Store
	if cfg.VectorDB.Enabled && cfg.VectorDB.Host != "" {
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
			log.Printf("Warning: Failed to connect to vector store: %v", err)
			log.Printf("Continuing without vector memory...")
			vectorStore = nil
		}
	}

	// Create formatter
	formatter := chat.NewFormatter(!*noColor)

	// Create or resume session
	var session *chat.ChatSession
	historyManager := chat.NewHistoryManager("")

	if *resumeSession != "" {
		// Resume existing session
		sessionData, err := historyManager.LoadSessionData(*resumeSession)
		if err != nil {
			log.Fatalf("Failed to load session: %v", err)
		}
		*workflowID = sessionData.WorkflowID

		session, err = chat.NewChatSession(chat.SessionConfig{
			WorkflowID:  sessionData.WorkflowID,
			LLMClient:   llmClient,
			VectorStore: vectorStore,
		})
		if err != nil {
			log.Fatalf("Failed to create session: %v", err)
		}
		session.History = sessionData.History
		fmt.Println(formatter.FormatInfo(fmt.Sprintf("Resumed session: %s", *resumeSession)))
	} else {
		// Create new session
		session, err = chat.NewChatSession(chat.SessionConfig{
			WorkflowID:  *workflowID,
			LLMClient:   llmClient,
			VectorStore: vectorStore,
		})
		if err != nil {
			log.Fatalf("Failed to create session: %v", err)
		}
	}

	// Load workflow context
	if err := session.LoadContext(); err != nil {
		log.Printf("Warning: Failed to load context: %v", err)
	}

	// Display welcome message
	fmt.Print(formatter.FormatWelcome(*workflowID))
	fmt.Print(formatter.FormatContextLoaded(session.GetContextSummary()))

	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println(formatter.FormatWarning("Received interrupt, saving session..."))
		saveSession(session, historyManager, formatter)
		fmt.Print(formatter.FormatGoodbye())
		os.Exit(0)
	}()

	// Start interactive loop
	reader := bufio.NewReader(os.Stdin)
	runChatLoop(ctx, session, reader, formatter, historyManager)
}

func runChatLoop(ctx context.Context, session *chat.ChatSession, reader *bufio.Reader,
	formatter *chat.Formatter, historyManager *chat.HistoryManager) {

	for {
		fmt.Print(formatter.FormatUserPrompt())

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(formatter.FormatError(err))
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			if handleCommand(ctx, input, session, formatter, historyManager) {
				return // Exit command received
			}
			continue
		}

		// Process message through agent
		response, _, err := session.SendMessage(ctx, input)
		if err != nil {
			fmt.Println(formatter.FormatError(err))
			continue
		}

		fmt.Print(formatter.FormatAssistantResponse(response))
	}
}

func handleCommand(ctx context.Context, cmd string, session *chat.ChatSession,
	formatter *chat.Formatter, historyManager *chat.HistoryManager) bool {

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}

	switch parts[0] {
	case "/exit", "/quit", "/q":
		saveSession(session, historyManager, formatter)
		fmt.Print(formatter.FormatGoodbye())
		return true

	case "/status":
		fmt.Print(formatter.FormatStatus(session))

	case "/reload":
		if err := session.ReloadContext(); err != nil {
			fmt.Println(formatter.FormatError(err))
		} else {
			fmt.Println(formatter.FormatSuccess("Context reloaded"))
			fmt.Print(formatter.FormatContextLoaded(session.GetContextSummary()))
		}

	case "/history":
		fmt.Print(formatter.FormatHistory(session.GetConversationHistory()))

	case "/save":
		saveSession(session, historyManager, formatter)

	case "/clear":
		session.History = nil
		fmt.Println(formatter.FormatSuccess("Conversation history cleared"))

	case "/refine":
		// Refine test suite based on feedback
		feedback := strings.TrimPrefix(cmd, "/refine")
		feedback = strings.TrimSpace(feedback)
		if feedback == "" {
			fmt.Println(formatter.FormatWarning("Usage: /refine <feedback>"))
			fmt.Println(formatter.FormatInfo("Example: /refine Add more negative tests for authentication"))
			break
		}
		fmt.Println(formatter.FormatInfo("🔧 Refining test suite..."))
		response, err := session.RefineTestSuite(ctx, feedback)
		if err != nil {
			fmt.Println(formatter.FormatError(err))
		} else {
			fmt.Print(formatter.FormatAssistantResponse(response))
		}

	case "/run":
		// Run the test suite
		baseURL := ""
		if len(parts) > 1 {
			baseURL = parts[1]
		}
		fmt.Println(formatter.FormatInfo("🚀 Running test suite..."))
		response, err := session.RunTests(ctx, baseURL)
		if err != nil {
			fmt.Println(formatter.FormatError(err))
		} else {
			fmt.Print(formatter.FormatAssistantResponse(response))
		}

	case "/recommend", "/recommendations":
		// Generate recommendations
		fmt.Println(formatter.FormatInfo("💡 Generating recommendations..."))
		response, err := session.GenerateRecommendations(ctx)
		if err != nil {
			fmt.Println(formatter.FormatError(err))
		} else {
			fmt.Print(formatter.FormatAssistantResponse(response))
		}

	case "/help", "/?":
		fmt.Print(formatter.FormatHelp())

	default:
		fmt.Println(formatter.FormatWarning(fmt.Sprintf("Unknown command: %s. Type /help for available commands.", parts[0])))
	}

	return false
}

func saveSession(session *chat.ChatSession, historyManager *chat.HistoryManager, formatter *chat.Formatter) {
	filePath, err := historyManager.SaveSession(session)
	if err != nil {
		fmt.Println(formatter.FormatError(fmt.Errorf("failed to save session: %w", err)))
		return
	}
	fmt.Println(formatter.FormatSuccess(fmt.Sprintf("Session saved to: %s", filePath)))
}
