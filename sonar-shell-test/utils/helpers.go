package utils

import "fmt"

// GetRatingLabel converts a numeric rating to a letter grade with emoji
func GetRatingLabel(rating string) string {
	switch rating {
	case "1.0":
		return "A ✅"
	case "2.0":
		return "B ⚠️"
	case "3.0":
		return "C ⚠️"
	case "4.0":
		return "D ❌"
	case "5.0":
		return "E ❌"
	default:
		return rating
	}
}

// GetIssueIcon returns an emoji icon for an issue type
func GetIssueIcon(issueType string) string {
	switch issueType {
	case "BUG":
		return "🐛"
	case "VULNERABILITY":
		return "🔒"
	case "CODE_SMELL":
		return "💩"
	default:
		return "⚠️"
	}
}

// GenerateSonarYML generates the SonarCloud workflow YAML content
func GenerateSonarYML(sonarOrgKey, sonarProjectKey, environmentName string) string {
	return fmt.Sprintf(`name: SonarCloud Scan

on:
  push:
    branches:
      - main
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  sonarcloud:
    name: SonarCloud Analysis
    runs-on: ubuntu-latest
    environment: %s

    steps:
      - name: Checkout Repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: SonarCloud Scan
        uses: SonarSource/sonarqube-scan-action@master
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
        with:
          args: >
            -Dsonar.projectKey=%s
            -Dsonar.organization=%s
`, environmentName, sonarProjectKey, sonarOrgKey)
}

