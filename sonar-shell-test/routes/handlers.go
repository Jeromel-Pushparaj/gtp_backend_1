package routes

import (
	"fmt"
	"log"
	"os"
	"sonar-automation/controllers"
	"sonar-automation/models"
	"sonar-automation/services"
)

// ListSecretsHandler lists all secrets across all repositories
func ListSecretsHandler() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          GitHub Repository Secrets Viewer               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load configuration
	config, err := models.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v\n", err)
	}

	fmt.Printf("Organization: %s\n", config.Organization)
	fmt.Println()

	// Create GitHub service
	gs := services.NewGitHubService(config.GitHubPAT)

	// List all repositories
	fmt.Printf("🔍 Fetching repositories from %s...\n", config.Organization)
	repos, err := gs.ListRepositories(config.Organization)
	if err != nil {
		log.Fatalf("❌ Failed to list repositories: %v\n", err)
	}

	fmt.Printf("✅ Found %d repositories\n\n", len(repos))

	totalSecrets := 0

	// List secrets for each repository
	for _, repo := range repos {
		repoName := repo.GetName()

		if repo.GetArchived() {
			fmt.Printf("📦 %s (archived - skipped)\n\n", repoName)
			continue
		}

		fmt.Printf("📦 %s\n", repoName)

		// List repository-level secrets
		secrets, err := gs.ListSecrets(config.Organization, repoName)
		if err != nil {
			fmt.Printf("   ❌ Error listing repository secrets: %v\n", err)
		} else if len(secrets) > 0 {
			fmt.Printf("   📋 Repository Secrets (%d):\n", len(secrets))
			for _, secret := range secrets {
				fmt.Printf("      ✅ %-20s Created: %s  |  Updated: %s\n",
					secret.Name,
					secret.CreatedAt.Format("2006-01-02 15:04:05"),
					secret.UpdatedAt.Format("2006-01-02 15:04:05"))
				totalSecrets++
			}
		}

		// List environment secrets
		envName := config.EnvironmentName
		envSecrets, err := gs.ListEnvSecrets(config.Organization, repoName, envName)
		if err != nil {
			fmt.Printf("   ⚠️  No environment '%s' or error: %v\n", envName, err)
		} else if len(envSecrets) > 0 {
			fmt.Printf("   🌍 Environment '%s' Secrets (%d):\n", envName, len(envSecrets))
			for _, secret := range envSecrets {
				fmt.Printf("      ✅ %-20s Created: %s  |  Updated: %s\n",
					secret.Name,
					secret.CreatedAt.Format("2006-01-02 15:04:05"),
					secret.UpdatedAt.Format("2006-01-02 15:04:05"))
				totalSecrets++
			}
		}

		if len(secrets) == 0 && (envSecrets == nil || len(envSecrets) == 0) {
			fmt.Printf("   ⚠️  No secrets found\n")
		}

		fmt.Println()
	}

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                      SUMMARY                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("Total repositories scanned: %d\n", len(repos))
	fmt.Printf("Total secrets found: %d\n", totalSecrets)
	fmt.Println()
}

// AddEnvSecretsHandler adds environment secrets to all repositories
func AddEnvSecretsHandler() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║      Add Environment Secrets to All Repositories        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load configuration
	config, err := models.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v\n", err)
	}

	fmt.Printf("Organization: %s\n", config.Organization)
	fmt.Printf("Environment: %s\n", config.EnvironmentName)
	fmt.Println()

	// Create services
	gs := services.NewGitHubService(config.GitHubPAT)
	rc := controllers.NewRepositoryController(gs, nil, config)

	// List all repositories
	fmt.Printf("🔍 Fetching repositories from %s...\n", config.Organization)
	repos, err := gs.ListRepositories(config.Organization)
	if err != nil {
		log.Fatalf("❌ Failed to list repositories: %v\n", err)
	}

	fmt.Printf("✅ Found %d repositories\n\n", len(repos))

	successCount := 0
	skipCount := 0
	errorCount := 0

	// Add environment secrets to each repository
	for _, repo := range repos {
		repoName := repo.GetName()

		if repo.GetArchived() {
			fmt.Printf("📦 %s (archived - skipped)\n\n", repoName)
			skipCount++
			continue
		}

		fmt.Printf("📦 %s\n", repoName)

		// Add environment secrets
		err := rc.SetupEnvironmentSecrets(repoName)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n\n", err)
			errorCount++
			continue
		}

		// Verify secrets
		err = rc.VerifyEnvironmentSecrets(repoName)
		if err != nil {
			fmt.Printf("  ⚠️  Warning: %v\n", err)
		}

		fmt.Printf("  ✅ Environment secrets added successfully\n\n")
		successCount++
	}

	// Summary
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                      SUMMARY                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("Total repositories: %d\n", len(repos))
	fmt.Printf("✅ Successfully processed: %d\n", successCount)
	fmt.Printf("⏭️  Skipped (archived): %d\n", skipCount)
	fmt.Printf("❌ Errors: %d\n", errorCount)
	fmt.Println()

	if errorCount > 0 {
		os.Exit(1)
	}
}

// UpdateWorkflowsHandler updates existing workflows to use environment
func UpdateWorkflowsHandler() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║       Update Workflows to Use Environment Secrets       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load configuration
	config, err := models.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v\n", err)
	}

	fmt.Printf("Organization: %s\n", config.Organization)
	fmt.Printf("Environment: %s\n", config.EnvironmentName)
	fmt.Println()

	// Create services
	gs := services.NewGitHubService(config.GitHubPAT)
	rc := controllers.NewRepositoryController(gs, nil, config)

	// List all repositories
	fmt.Printf("🔍 Fetching repositories from %s...\n", config.Organization)
	repos, err := gs.ListRepositories(config.Organization)
	if err != nil {
		log.Fatalf("❌ Failed to list repositories: %v\n", err)
	}

	fmt.Printf("✅ Found %d repositories\n\n", len(repos))

	successCount := 0
	skipCount := 0
	errorCount := 0

	// Update workflows in each repository
	for _, repo := range repos {
		repoName := repo.GetName()

		if repo.GetArchived() {
			fmt.Printf("📦 %s (archived - skipped)\n\n", repoName)
			skipCount++
			continue
		}

		fmt.Printf("📦 %s\n", repoName)

		// Get default branch
		defaultBranch, err := gs.GetDefaultBranch(config.Organization, repoName)
		if err != nil {
			fmt.Printf("  ❌ Failed to get default branch: %v\n\n", err)
			errorCount++
			continue
		}

		// Check if sonar.yml exists
		sonarPath := ".github/workflows/sonar.yml"
		exists, err := gs.CheckFileExists(config.Organization, repoName, sonarPath, defaultBranch)
		if err != nil {
			fmt.Printf("  ❌ Failed to check sonar.yml: %v\n\n", err)
			errorCount++
			continue
		}

		if !exists {
			fmt.Printf("  ⚠️  sonar.yml not found, skipping\n\n")
			skipCount++
			continue
		}

		// Update the workflow file
		err = rc.UpdateWorkflowToUseEnvironment(repoName, defaultBranch)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n\n", err)
			errorCount++
			continue
		}

		fmt.Printf("  ✅ Workflow updated successfully\n\n")
		successCount++
	}

	// Summary
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                      SUMMARY                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("Total repositories: %d\n", len(repos))
	fmt.Printf("✅ Successfully updated: %d\n", successCount)
	fmt.Printf("⏭️  Skipped (archived/no workflow): %d\n", skipCount)
	fmt.Printf("❌ Errors: %d\n", errorCount)
	fmt.Println()

	if errorCount > 0 {
		os.Exit(1)
	}
}

// FullSetupHandler performs complete setup: SonarCloud project creation, GitHub setup, and result fetching
func FullSetupHandler() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Full SonarCloud & GitHub Setup                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load configuration
	config, err := models.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v\n", err)
	}

	fmt.Printf("Organization: %s\n", config.Organization)
	fmt.Printf("SonarCloud Org: %s\n", config.SonarOrgKey)
	fmt.Printf("Environment: %s\n", config.EnvironmentName)
	fmt.Println()

	// Create services
	gs := services.NewGitHubService(config.GitHubPAT)
	sc := services.NewSonarCloudService(config.SonarToken, config.SonarOrgKey)
	rc := controllers.NewRepositoryController(gs, sc, config)

	// List all repositories
	fmt.Printf("🔍 Fetching repositories from %s...\n", config.Organization)
	repos, err := gs.ListRepositories(config.Organization)
	if err != nil {
		log.Fatalf("❌ Failed to list repositories: %v\n", err)
	}

	fmt.Printf("✅ Found %d repositories\n\n", len(repos))

	successCount := 0
	skipCount := 0
	errorCount := 0

	// Process each repository
	for _, repo := range repos {
		repoName := repo.GetName()

		if repo.GetArchived() {
			fmt.Printf("📦 %s (archived - skipped)\n\n", repoName)
			skipCount++
			continue
		}

		fmt.Printf("📦 Processing repository: %s\n", repoName)

		// Process repository with full setup
		err := rc.ProcessRepositoryFullSetup(repoName)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n\n", err)
			errorCount++
			continue
		}

		fmt.Printf("  ✅ Repository setup complete!\n\n")
		successCount++
	}

	// Summary
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                      SUMMARY                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("Total repositories: %d\n", len(repos))
	fmt.Printf("✅ Successfully processed: %d\n", successCount)
	fmt.Printf("⏭️  Skipped (archived): %d\n", skipCount)
	fmt.Printf("❌ Errors: %d\n", errorCount)
	fmt.Println()

	if errorCount > 0 {
		os.Exit(1)
	}
}

// FetchResultsHandler fetches analysis results from SonarCloud for all projects
func FetchResultsHandler() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║        Fetch SonarCloud Analysis Results                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load configuration
	config, err := models.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v\n", err)
	}

	fmt.Printf("Organization: %s\n", config.Organization)
	fmt.Printf("SonarCloud Org: %s\n", config.SonarOrgKey)
	fmt.Println()

	// Create services
	gs := services.NewGitHubService(config.GitHubPAT)
	sc := services.NewSonarCloudService(config.SonarToken, config.SonarOrgKey)
	rc := controllers.NewRepositoryController(gs, sc, config)

	// List all repositories
	fmt.Printf("🔍 Fetching repositories from %s...\n", config.Organization)
	repos, err := gs.ListRepositories(config.Organization)
	if err != nil {
		log.Fatalf("❌ Failed to list repositories: %v\n", err)
	}

	fmt.Printf("✅ Found %d repositories\n\n", len(repos))

	// Fetch results for each repository
	for _, repo := range repos {
		repoName := repo.GetName()

		if repo.GetArchived() {
			continue
		}

		projectKey := fmt.Sprintf("%s_%s", config.SonarOrgKey, repoName)

		fmt.Printf("📊 %s\n", repoName)
		fmt.Printf("   Project Key: %s\n", projectKey)

		// Fetch and display results
		err := rc.FetchAndDisplayResults(projectKey)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n\n", err)
			continue
		}

		fmt.Println()
	}
}

// DefaultHandler handles the default case when no flags are provided
func DefaultHandler() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║   SonarCloud Automation - Organization-wide Setup       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Load configuration
	config, err := models.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v\n", err)
	}

	fmt.Printf("Organization: %s\n", config.Organization)
	fmt.Printf("SonarCloud Org: %s\n", config.SonarOrgKey)
	fmt.Printf("Default Branch: %s\n", config.DefaultBranch)
	fmt.Printf("Environment: %s\n", config.EnvironmentName)
	fmt.Println()

	// Create services
	gs := services.NewGitHubService(config.GitHubPAT)
	rc := controllers.NewRepositoryController(gs, nil, config)

	// List all repositories
	fmt.Printf("🔍 Fetching repositories from %s...\n", config.Organization)
	repos, err := gs.ListRepositories(config.Organization)
	if err != nil {
		log.Fatalf("❌ Failed to list repositories: %v\n", err)
	}

	fmt.Printf("✅ Found %d repositories\n", len(repos))

	if len(repos) == 0 {
		fmt.Println("⚠️  No repositories found in organization")
		os.Exit(0)
	}

	// Process each repository
	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, repo := range repos {
		repoName := repo.GetName()

		// Skip archived repositories
		if repo.GetArchived() {
			fmt.Printf("\n📦 %s (archived, skipping)\n", repoName)
			skipCount++
			continue
		}

		err := rc.ProcessRepository(repoName)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			errorCount++
		} else {
			successCount++
		}
	}

	// Summary
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                      SUMMARY                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("Total repositories: %d\n", len(repos))
	fmt.Printf("✅ Successfully processed: %d\n", successCount)
	fmt.Printf("⏭️  Skipped (archived/existing): %d\n", skipCount)
	fmt.Printf("❌ Errors: %d\n", errorCount)
	fmt.Println()

	if errorCount > 0 {
		os.Exit(1)
	}
}

