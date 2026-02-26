package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"sonar-automation/api"
	"sonar-automation/models"
	"sonar-automation/routes"
	"sonar-automation/services"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found or error loading it, using system environment variables")
	}

	// Command-line flags
	serverMode := flag.Bool("server", false, "Run as API server")
	port := flag.String("port", "8080", "API server port (only used with -server)")
	listSecretsFlag := flag.Bool("list-secrets", false, "List all secrets in repositories")
	addEnvSecretsFlag := flag.Bool("add-env-secrets", false, "Add environment secrets to existing repositories")
	updateWorkflowsFlag := flag.Bool("update-workflows", false, "Update existing workflows to use environment")
	fullSetupFlag := flag.Bool("full-setup", false, "Full setup: Create SonarCloud projects, add workflows, and fetch results")
	fetchResultsFlag := flag.Bool("fetch-results", false, "Fetch analysis results from SonarCloud")
	flag.Parse()

	// If server mode is enabled, start API server
	if *serverMode {
		config, err := models.LoadConfig()
		if err != nil {
			log.Fatalf("❌ Configuration error: %v\n", err)
		}

		// Initialize database
		db, err := services.NewDatabaseService(config.DatabasePath)
		if err != nil {
			log.Fatalf("❌ Database initialization error: %v\n", err)
		}
		defer db.Close()

		log.Printf("✅ Database initialized at: %s\n", config.DatabasePath)

		apiKey := os.Getenv("API_KEY") // Optional API key for authentication
		server := api.NewServer(config, *port, apiKey, db)

		log.Fatal(server.Start())
		return
	}

	// CLI mode - Route to appropriate handler based on flags
	if *listSecretsFlag {
		routes.ListSecretsHandler()
		return
	}

	if *addEnvSecretsFlag {
		routes.AddEnvSecretsHandler()
		return
	}

	if *updateWorkflowsFlag {
		routes.UpdateWorkflowsHandler()
		return
	}

	if *fullSetupFlag {
		routes.FullSetupHandler()
		return
	}

	if *fetchResultsFlag {
		routes.FetchResultsHandler()
		return
	}

	// Default handler
	routes.DefaultHandler()
}
