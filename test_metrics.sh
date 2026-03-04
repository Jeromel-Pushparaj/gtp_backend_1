#!/bin/bash

# Individual Metrics API Test Script
# Tests all individual metric endpoints

REPO="delivery-management-frontend"
PROJECT="DM"
BASE_URL="http://localhost:8080"

echo "🔍 Testing Individual Metrics APIs..."
echo "================================================"
echo ""

# GitHub Metrics
echo "1️⃣ GitHub Metrics:"
echo "-------------------"
curl -s "$BASE_URL/api/v1/github/metrics?repo=$REPO" | jq '.data | {has_readme, open_prs, merged_prs, contributors, branches, commits_last_90_days}'
echo ""

# README Check
echo "2️⃣ README Check:"
echo "-------------------"
curl -s "$BASE_URL/api/v1/github/readme?repo=$REPO" | jq '.data'
echo ""

# Branches
echo "3️⃣ Branches:"
echo "-------------------"
curl -s "$BASE_URL/api/v1/github/branches?repo=$REPO" | jq '.data | {repository, count, branches}'
echo ""

# SonarCloud Metrics
echo "4️⃣ SonarCloud Metrics:"
echo "-------------------"
curl -s "$BASE_URL/api/v1/sonar/metrics?repo=$REPO" | jq '.data | {coverage, bugs, vulnerabilities, code_smells, quality_gate_status}'
echo ""

# Jira Metrics
echo "5️⃣ Jira Metrics:"
echo "-------------------"
curl -s "$BASE_URL/api/v1/jira/metrics?project=$PROJECT" | jq '.data | {open_bugs, mttr, total_issues}'
echo ""

# Jira Open Bugs
echo "6️⃣ Jira Open Bugs:"
echo "-------------------"
curl -s "$BASE_URL/api/v1/jira/bugs/open?project=$PROJECT" | jq '.data | {project_key, open_bugs}'
echo ""

# Organization Members
echo "7️⃣ Organization Members:"
echo "-------------------"
curl -s "$BASE_URL/api/v1/github/org/members" | jq '.data | {count: length}'
echo ""

# Organization Teams
echo "8️⃣ Organization Teams:"
echo "-------------------"
curl -s "$BASE_URL/api/v1/github/org/teams" | jq '.data | {count: length}'
echo ""

echo "================================================"
echo "✅ Test complete!"
echo ""
echo "💡 Tip: To test a different repository, run:"
echo "   REPO=your-repo-name PROJECT=YOUR-PROJECT ./test_metrics.sh"

