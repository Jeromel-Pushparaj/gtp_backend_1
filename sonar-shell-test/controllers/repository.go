package controllers

import (
	"fmt"
	"time"

	"sonar-automation/models"
	"sonar-automation/services"
	"sonar-automation/utils"
)

// RepositoryController handles repository-related operations
type RepositoryController struct {
	githubService     *services.GitHubService
	sonarCloudService *services.SonarCloudService
	config            *models.Config
}

// NewRepositoryController creates a new repository controller
func NewRepositoryController(gs *services.GitHubService, sc *services.SonarCloudService, cfg *models.Config) *RepositoryController {
	return &RepositoryController{
		githubService:     gs,
		sonarCloudService: sc,
		config:            cfg,
	}
}

// ProcessRepository processes a single repository
func (rc *RepositoryController) ProcessRepository(repoName string) error {
	fmt.Printf("\n📦 Processing repository: %s\n", repoName)

	// Get default branch
	defaultBranch, err := rc.githubService.GetDefaultBranch(rc.config.Organization, repoName)
	if err != nil {
		return fmt.Errorf("  ❌ Failed to get default branch: %w", err)
	}
	fmt.Printf("  ℹ️  Default branch: %s\n", defaultBranch)

	// Check if sonar.yml exists
	sonarPath := ".github/workflows/sonar.yml"
	exists, err := rc.githubService.CheckFileExists(rc.config.Organization, repoName, sonarPath, defaultBranch)
	if err != nil {
		return fmt.Errorf("  ❌ Failed to check sonar.yml: %w", err)
	}

	if exists {
		fmt.Printf("  ✅ sonar.yml already exists, skipping\n")
		return nil
	}

	fmt.Printf("  ⚠️  sonar.yml not found, proceeding with setup...\n")

	// STEP 1 & 2: Add secrets to environment FIRST (before creating the workflow file)
	fmt.Printf("  🔐 Setting up environment and secrets...\n")
	err = rc.SetupEnvironmentSecrets(repoName)
	if err != nil {
		return fmt.Errorf("  ❌ Failed to setup environment secrets: %w", err)
	}

	// Verify secrets were added
	fmt.Printf("  🔍 Verifying environment secrets...\n")
	err = rc.VerifyEnvironmentSecrets(repoName)
	if err != nil {
		fmt.Printf("  ⚠️  Warning: %v\n", err)
	}

	// STEP 3: Generate sonar.yml content
	sonarProjectKey := fmt.Sprintf("%s_%s", rc.config.SonarOrgKey, repoName)
	sonarContent := utils.GenerateSonarYML(rc.config.SonarOrgKey, sonarProjectKey, rc.config.EnvironmentName)

	// Create sonar.yml (push to repo)
	fmt.Printf("  📤 Pushing sonar.yml to repository...\n")
	err = rc.githubService.CreateFile(
		rc.config.Organization,
		repoName,
		sonarPath,
		"ci: Add SonarCloud workflow [automated]",
		sonarContent,
		defaultBranch,
	)
	if err != nil {
		return fmt.Errorf("  ❌ Failed to create sonar.yml: %w", err)
	}
	fmt.Printf("  ✅ sonar.yml pushed successfully\n")

	fmt.Printf("  ✅ Repository setup complete!\n")
	return nil
}

// SetupEnvironmentSecrets creates an environment and adds secrets to it
func (rc *RepositoryController) SetupEnvironmentSecrets(repoName string) error {
	envName := rc.config.EnvironmentName

	// Check if environment exists, create if not
	fmt.Printf("    [1/3] Creating environment '%s'...\n", envName)
	err := rc.githubService.CreateEnvironment(rc.config.Organization, repoName, envName)
	if err != nil {
		// Check if it's a "already exists" error - if so, that's fine
		fmt.Printf("    ℹ️  Environment '%s' may already exist: %v\n", envName, err)
		// Continue anyway - we'll try to add secrets
	} else {
		fmt.Printf("    ✅ Environment '%s' created successfully\n", envName)
	}

	// Get environment public key for encryption
	fmt.Printf("    Getting public key for environment '%s'...\n", envName)
	publicKey, err := rc.githubService.GetEnvironmentPublicKey(rc.config.Organization, repoName, envName)
	if err != nil {
		return fmt.Errorf("failed to get environment public key for '%s': %w", envName, err)
	}
	fmt.Printf("    ✅ Public key retrieved\n")

	// STEP 2: Encrypt and add GH_PAT (GitHub doesn't allow secrets starting with GITHUB_)
	fmt.Printf("    [2/3] Adding GH_PAT to environment...\n")
	encryptedPAT, err := rc.githubService.EncryptSecret(publicKey, rc.config.GitHubPAT)
	if err != nil {
		return fmt.Errorf("failed to encrypt GH_PAT: %w", err)
	}

	err = rc.githubService.CreateOrUpdateEnvSecret(
		rc.config.Organization,
		repoName,
		envName,
		"GH_PAT",
		encryptedPAT,
		publicKey.GetKeyID(),
	)
	if err != nil {
		return err
	}
	fmt.Printf("    ✅ GH_PAT added to environment\n")

	// STEP 3: Encrypt and add SONAR_TOKEN
	fmt.Printf("    [3/3] Adding SONAR_TOKEN to environment...\n")
	encryptedSonar, err := rc.githubService.EncryptSecret(publicKey, rc.config.SonarToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt SONAR_TOKEN: %w", err)
	}

	err = rc.githubService.CreateOrUpdateEnvSecret(
		rc.config.Organization,
		repoName,
		envName,
		"SONAR_TOKEN",
		encryptedSonar,
		publicKey.GetKeyID(),
	)
	if err != nil {
		return err
	}
	fmt.Printf("    ✅ SONAR_TOKEN added to environment\n")

	return nil
}

// VerifyEnvironmentSecrets verifies that secrets were successfully added to the environment
func (rc *RepositoryController) VerifyEnvironmentSecrets(repoName string) error {
	envName := rc.config.EnvironmentName

	secrets, err := rc.githubService.ListEnvSecrets(rc.config.Organization, repoName, envName)
	if err != nil {
		return fmt.Errorf("failed to list environment secrets: %w", err)
	}

	// Check for required secrets
	requiredSecrets := map[string]bool{
		"GH_PAT":      false,
		"SONAR_TOKEN": false,
	}

	for _, secret := range secrets {
		if _, exists := requiredSecrets[secret.Name]; exists {
			requiredSecrets[secret.Name] = true
			fmt.Printf("    ✅ %s verified in '%s' environment (created: %s)\n",
				secret.Name, envName, secret.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	// Check if all required secrets are present
	missing := []string{}
	for name, found := range requiredSecrets {
		if !found {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing secrets in environment '%s': %v", envName, missing)
	}

	return nil
}

// UpdateWorkflowToUseEnvironment updates an existing sonar.yml to use environment
func (rc *RepositoryController) UpdateWorkflowToUseEnvironment(repoName, branch string) error {
	sonarPath := ".github/workflows/sonar.yml"

	fmt.Printf("  📥 Getting current workflow file...\n")

	// Get current file content and SHA
	_, sha, err := rc.githubService.GetFileContent(rc.config.Organization, repoName, sonarPath, branch)
	if err != nil {
		return fmt.Errorf("failed to get current workflow: %w", err)
	}

	fmt.Printf("  ✅ Current file SHA: %s\n", sha[:7])

	// Generate new sonar.yml content with environment
	sonarProjectKey := fmt.Sprintf("%s_%s", rc.config.SonarOrgKey, repoName)
	sonarContent := utils.GenerateSonarYML(rc.config.SonarOrgKey, sonarProjectKey, rc.config.EnvironmentName)

	fmt.Printf("  📝 Updating workflow to use '%s' environment...\n", rc.config.EnvironmentName)

	// Update the file
	err = rc.githubService.UpdateFile(
		rc.config.Organization,
		repoName,
		sonarPath,
		"ci: Update SonarCloud workflow to use environment [automated]",
		sonarContent,
		branch,
		sha,
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow: %w", err)
	}

	fmt.Printf("  ✅ Workflow file updated\n")

	return nil
}

// ProcessRepositoryFullSetup performs complete setup for a repository
func (rc *RepositoryController) ProcessRepositoryFullSetup(repoName string) error {
	// Get default branch
	defaultBranch, err := rc.githubService.GetDefaultBranch(rc.config.Organization, repoName)
	if err != nil {
		return fmt.Errorf("failed to get default branch: %w", err)
	}
	fmt.Printf("  ℹ️  Default branch: %s\n", defaultBranch)

	projectKey := fmt.Sprintf("%s_%s", rc.config.SonarOrgKey, repoName)

	// ═══════════════════════════════════════════════════════════════
	// STEP 1: Create SonarCloud Project
	// ═══════════════════════════════════════════════════════════════
	fmt.Printf("  [1/4] Creating SonarCloud project...\n")
	exists, err := rc.sonarCloudService.ProjectExists(projectKey)
	if err != nil {
		return fmt.Errorf("failed to check project existence: %w", err)
	}

	if exists {
		fmt.Printf("        ✅ SonarCloud project already exists: %s\n", projectKey)
	} else {
		fmt.Printf("        📝 Creating project: %s\n", projectKey)
		err = rc.sonarCloudService.CreateProject(projectKey, repoName)
		if err != nil {
			return fmt.Errorf("failed to create SonarCloud project: %w", err)
		}
		fmt.Printf("        ✅ SonarCloud project created successfully\n")

		// Set main branch
		err = rc.sonarCloudService.SetMainBranch(projectKey, defaultBranch)
		if err != nil {
			fmt.Printf("        ⚠️  Warning: Could not set main branch: %v\n", err)
		} else {
			fmt.Printf("        ✅ Main branch set to: %s\n", defaultBranch)
		}
	}

	// ═══════════════════════════════════════════════════════════════
	// STEP 2: Check and Create Environment & Secrets
	// ═══════════════════════════════════════════════════════════════
	fmt.Printf("  [2/4] Checking environment and secrets...\n")

	// Check if environment exists
	envName := rc.config.EnvironmentName
	envSecrets, err := rc.githubService.ListEnvSecrets(rc.config.Organization, repoName, envName)
	envExists := err == nil && envSecrets != nil

	if envExists && len(envSecrets) >= 2 {
		fmt.Printf("        ✅ Environment '%s' exists with %d secrets\n", envName, len(envSecrets))
	} else {
		fmt.Printf("        📝 Setting up environment '%s' and secrets...\n", envName)
		err = rc.SetupEnvironmentSecrets(repoName)
		if err != nil {
			return fmt.Errorf("failed to setup environment secrets: %w", err)
		}
		fmt.Printf("        ✅ Environment and secrets configured\n")
	}

	// ═══════════════════════════════════════════════════════════════
	// STEP 3: Check and Push sonar.yml
	// ═══════════════════════════════════════════════════════════════
	fmt.Printf("  [3/4] Checking and pushing sonar.yml...\n")
	sonarPath := ".github/workflows/sonar.yml"
	workflowExists, err := rc.githubService.CheckFileExists(rc.config.Organization, repoName, sonarPath, defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to check sonar.yml: %w", err)
	}

	if workflowExists {
		fmt.Printf("        ✅ sonar.yml already exists\n")
	} else {
		fmt.Printf("        📝 Creating sonar.yml workflow...\n")
		sonarProjectKey := fmt.Sprintf("%s_%s", rc.config.SonarOrgKey, repoName)
		sonarContent := utils.GenerateSonarYML(rc.config.SonarOrgKey, sonarProjectKey, rc.config.EnvironmentName)

		err = rc.githubService.CreateFile(
			rc.config.Organization,
			repoName,
			sonarPath,
			"ci: Add SonarCloud workflow [automated]",
			sonarContent,
			defaultBranch,
		)
		if err != nil {
			return fmt.Errorf("failed to create sonar.yml: %w", err)
		}
		fmt.Printf("        ✅ sonar.yml created and pushed to repository\n")
	}

	// ═══════════════════════════════════════════════════════════════
	// STEP 4: Fetch Analysis Results from SonarCloud API
	// ═══════════════════════════════════════════════════════════════
	fmt.Printf("  [4/4] Fetching analysis results from SonarCloud...\n")

	// Wait a moment for SonarCloud to be ready
	time.Sleep(2 * time.Second)

	err = rc.FetchAndDisplayResults(projectKey)
	if err != nil {
		fmt.Printf("        ⚠️  Could not fetch results: %v\n", err)
		fmt.Printf("        ℹ️  Note: Results will be available after the first workflow run\n")
		fmt.Printf("        ℹ️  Push a commit to trigger the analysis\n")
	}

	return nil
}

// FetchAndDisplayResults fetches and displays SonarCloud analysis results
func (rc *RepositoryController) FetchAndDisplayResults(projectKey string) error {
	// Get quality gate status
	qgStatus, err := rc.sonarCloudService.GetQualityGateStatus(projectKey)
	if err != nil {
		return fmt.Errorf("failed to get quality gate status: %w", err)
	}

	// Get measures
	measures, err := rc.sonarCloudService.GetProjectMeasures(projectKey)
	if err != nil {
		return fmt.Errorf("failed to get measures: %w", err)
	}

	// Get issues summary
	issues, err := rc.sonarCloudService.GetIssues(projectKey, 10)
	if err != nil {
		return fmt.Errorf("failed to get issues: %w", err)
	}

	// Display results
	fmt.Printf("   ╔════════════════════════════════════════════════╗\n")
	fmt.Printf("   ║           SonarCloud Analysis Results         ║\n")
	fmt.Printf("   ╚════════════════════════════════════════════════╝\n")

	// Quality Gate
	qgIcon := "✅"
	if qgStatus.Status == "ERROR" {
		qgIcon = "❌"
	} else if qgStatus.Status == "WARN" {
		qgIcon = "⚠️"
	}
	fmt.Printf("   %s Quality Gate: %s\n", qgIcon, qgStatus.Status)
	fmt.Println()

	// Measures
	fmt.Printf("   📊 Metrics:\n")
	measureMap := make(map[string]string)
	for _, m := range measures {
		measureMap[m.Metric] = m.Value
	}

	displayMeasure := func(label, metric, unit string) {
		if val, ok := measureMap[metric]; ok {
			fmt.Printf("      %-25s %s%s\n", label+":", val, unit)
		}
	}

	displayMeasure("Lines of Code", "ncloc", "")
	displayMeasure("Bugs", "bugs", "")
	displayMeasure("Vulnerabilities", "vulnerabilities", "")
	displayMeasure("Code Smells", "code_smells", "")
	displayMeasure("Coverage", "coverage", "%")
	displayMeasure("Duplications", "duplicated_lines_density", "%")

	// Ratings
	fmt.Println()
	fmt.Printf("   ⭐ Ratings:\n")
	displayRating := func(label, metric string) {
		if val, ok := measureMap[metric]; ok {
			rating := utils.GetRatingLabel(val)
			fmt.Printf("      %-25s %s\n", label+":", rating)
		}
	}

	displayRating("Maintainability", "sqale_rating")
	displayRating("Reliability", "reliability_rating")
	displayRating("Security", "security_rating")

	// Issues summary
	fmt.Println()
	fmt.Printf("   🐛 Issues: %d total\n", issues.Total)
	if len(issues.Issues) > 0 {
		fmt.Printf("      Recent issues:\n")
		for i, issue := range issues.Issues {
			if i >= 5 {
				break
			}
			typeIcon := utils.GetIssueIcon(issue.Type)
			fmt.Printf("      %s [%s] %s (Line %d)\n", typeIcon, issue.Severity, issue.Message, issue.Line)
		}
		if issues.Total > 5 {
			fmt.Printf("      ... and %d more\n", issues.Total-5)
		}
	}

	return nil
}

