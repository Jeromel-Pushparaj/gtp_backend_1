package main

import (
	"log"

	"github.com/joho/godotenv"
	"sonar-automation/models"
	"sonar-automation/services"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found or error loading it, using system environment variables")
	}

	// Load configuration
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

	log.Println("╔══════════════════════════════════════════════════════════╗")
	log.Println("║          Database Seeding - Organization Setup          ║")
	log.Println("╚══════════════════════════════════════════════════════════╝")
	log.Println()

	// Check if organization already exists
	existingOrg, err := db.GetOrganizationByName(config.Organization)
	if err == nil {
		log.Printf("⚠️  Organization '%s' already exists in database\n", config.Organization)
		log.Printf("   ID: %d\n", existingOrg.ID)
		log.Printf("   SonarCloud Org Key: %s\n", existingOrg.SonarOrgKey)
		log.Printf("   Jira Domain: %s\n", existingOrg.JiraDomain)
		log.Printf("   Jira Email: %s\n", existingOrg.JiraEmail)
		log.Println()
		log.Println("✅ Database already seeded. Use this organization for metrics collection.")
		return
	}

	// Create organization with credentials
	org := &models.Organization{
		Name:        config.Organization,
		GitHubPAT:   config.GitHubPAT,
		SonarToken:  config.SonarToken,
		SonarOrgKey: config.SonarOrgKey,
		JiraToken:   config.JiraToken,
		JiraDomain:  config.JiraDomain,
		JiraEmail:   config.JiraEmail,
	}

	if err := db.CreateOrganization(org); err != nil {
		log.Fatalf("❌ Failed to create organization: %v\n", err)
	}

	log.Println("✅ Organization created successfully!")
	log.Println()
	log.Printf("   Organization: %s\n", org.Name)
	log.Printf("   Database ID: %d\n", org.ID)
	log.Printf("   SonarCloud Org Key: %s\n", org.SonarOrgKey)
	log.Printf("   Jira Domain: %s\n", org.JiraDomain)
	log.Printf("   Jira Email: %s\n", org.JiraEmail)
	log.Printf("   GitHub PAT: %s...\n", config.GitHubPAT[:20])
	log.Printf("   SonarCloud Token: %s...\n", config.SonarToken[:20])
	log.Printf("   Jira Token: %s...\n", config.JiraToken[:20])
	log.Println()
	log.Println("✅ Database seeded successfully!")
	log.Println()
	log.Println("Next steps:")
	log.Println("  1. Start the API server: ./bin/server -server")
	log.Println("  2. Collect metrics for a repository:")
	log.Println("     POST http://localhost:8080/api/v1/metrics/github/collect?repo=<repo-name>")
	log.Println("  3. View stored metrics:")
	log.Println("     GET http://localhost:8080/api/v1/metrics/github/stored?repo=<repo-name>")
	log.Println()
}

