#!/bin/bash

# Test script for Score Card Service API v2
# Advanced scorecards with Gold/Silver/Bronze levels and Traffic Light scoring

BASE_URL="http://localhost:8085"

echo "🧪 Testing Score Card Service API v2"
echo "===================================="
echo ""

# Test 1: Get all scorecard definitions
echo "📋 Test 1: Get All Scorecard Definitions"
echo "GET $BASE_URL/api/v2/scorecards/definitions"
curl -s "$BASE_URL/api/v2/scorecards/definitions" | jq '.'
echo ""
echo ""

# Test 2: Get specific scorecard definition (Code Quality)
echo "📋 Test 2: Get Code Quality Scorecard Definition"
echo "GET $BASE_URL/api/v2/scorecards/definitions/CodeQuality"
curl -s "$BASE_URL/api/v2/scorecards/definitions/CodeQuality" | jq '.'
echo ""
echo ""

# Test 3: Get Production Readiness scorecard (Traffic Light pattern)
echo "🚦 Test 3: Get Production Readiness Scorecard (Traffic Light)"
echo "GET $BASE_URL/api/v2/scorecards/definitions/Production_Readiness"
curl -s "$BASE_URL/api/v2/scorecards/definitions/Production_Readiness" | jq '.'
echo ""
echo ""

# Test 4: Evaluate a service with GOLD level metrics
echo "🥇 Test 4: Evaluate Service with GOLD Level Metrics"
echo "POST $BASE_URL/api/v2/scorecards/evaluate/CodeQuality"
curl -s -X POST "$BASE_URL/api/v2/scorecards/evaluate/CodeQuality" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "gold-service",
    "service_data": {
      "coverage": 92,
      "vulnerabilities": 0,
      "code_smells": 8,
      "duplicated_lines_density": 2.5,
      "has_readme": 1
    }
  }' | jq '.'
echo ""
echo ""

# Test 5: Evaluate a service with SILVER level metrics
echo "🥈 Test 5: Evaluate Service with SILVER Level Metrics"
echo "POST $BASE_URL/api/v2/scorecards/evaluate/CodeQuality"
curl -s -X POST "$BASE_URL/api/v2/scorecards/evaluate/CodeQuality" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "silver-service",
    "service_data": {
      "coverage": 82,
      "vulnerabilities": 3,
      "code_smells": 45,
      "duplicated_lines_density": 4.5,
      "has_readme": 1
    }
  }' | jq '.'
echo ""
echo ""

# Test 6: Evaluate a service with BRONZE level metrics
echo "🥉 Test 6: Evaluate Service with BRONZE Level Metrics"
echo "POST $BASE_URL/api/v2/scorecards/evaluate/CodeQuality"
curl -s -X POST "$BASE_URL/api/v2/scorecards/evaluate/CodeQuality" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "bronze-service",
    "service_data": {
      "coverage": 65,
      "vulnerabilities": 8,
      "code_smells": 100,
      "duplicated_lines_density": 4.8,
      "has_readme": 1
    }
  }' | jq '.'
echo ""
echo ""

# Test 7: Evaluate Production Readiness (Traffic Light - GREEN)
echo "🟢 Test 7: Evaluate Production Readiness - GREEN Level"
echo "POST $BASE_URL/api/v2/scorecards/evaluate/Production_Readiness"
curl -s -X POST "$BASE_URL/api/v2/scorecards/evaluate/Production_Readiness" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "production-ready-service",
    "service_data": {
      "has_readme": 1,
      "days_since_last_commit": 3,
      "contributors": 6,
      "coverage": 85,
      "quality_gate_passed": 1
    }
  }' | jq '.'
echo ""
echo ""

# Test 8: Evaluate Production Readiness (Traffic Light - YELLOW)
echo "🟡 Test 8: Evaluate Production Readiness - YELLOW Level"
echo "POST $BASE_URL/api/v2/scorecards/evaluate/Production_Readiness"
curl -s -X POST "$BASE_URL/api/v2/scorecards/evaluate/Production_Readiness" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "needs-improvement-service",
    "service_data": {
      "has_readme": 1,
      "days_since_last_commit": 25,
      "contributors": 2,
      "coverage": 60,
      "quality_gate_passed": 0
    }
  }' | jq '.'
echo ""
echo ""

# Test 9: Evaluate ALL scorecards for a comprehensive service
echo "🎯 Test 9: Evaluate ALL Scorecards for Comprehensive Service"
echo "POST $BASE_URL/api/v2/scorecards/evaluate"
curl -s -X POST "$BASE_URL/api/v2/scorecards/evaluate" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "comprehensive-service",
    "service_data": {
      "coverage": 85.5,
      "vulnerabilities": 2,
      "code_smells": 15,
      "duplicated_lines_density": 2.5,
      "has_readme": 1,
      "bugs": 8,
      "open_bugs": 3,
      "mttr": 18,
      "deployment_frequency": 5,
      "merged_prs": 25,
      "prs_with_conflicts": 1,
      "open_prs": 4,
      "contributors": 4,
      "days_since_last_commit": 2,
      "quality_gate_passed": 1,
      "security_hotspots": 1
    }
  }' | jq '.'
echo ""
echo ""

echo "✅ All tests completed!"

