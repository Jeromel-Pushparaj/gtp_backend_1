package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// ValidationResult contains the result of input validation
type ValidationResult struct {
	Valid     bool
	RiskScore int
	Reasons   []string
	Sanitized string
}

// ValidateInput performs comprehensive input validation with risk scoring
func ValidateInput(input string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:     true,
		RiskScore: 0,
		Reasons:   []string{},
		Sanitized: input,
	}

	// Step 1: Sanitize input
	result.Sanitized = SanitizeInput(input)

	// Step 2: Basic checks (hard failures)
	if strings.TrimSpace(result.Sanitized) == "" {
		return nil, fmt.Errorf("input cannot be empty")
	}

	if len(result.Sanitized) > MaxInputLength {
		return nil, fmt.Errorf("input too long (maximum %d characters)", MaxInputLength)
	}

	if len(result.Sanitized) < MinInputLength {
		return nil, fmt.Errorf("input too short (minimum %d characters)", MinInputLength)
	}

	// Step 3: Token estimation (hard failure)
	tokens := EstimateTokens(result.Sanitized)
	if tokens > MaxTokens {
		return nil, fmt.Errorf("input exceeds token limit (%d tokens, max %d)", tokens, MaxTokens)
	}

	// Step 4: Risk scoring (soft checks)
	normalizedInput := normalizeUnicode(result.Sanitized)
	lowerInput := strings.ToLower(normalizedInput)

	// Check prompt injection patterns
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerInput, pattern) {
			// Critical patterns get instant block weight
			if isCriticalPattern(pattern) {
				result.RiskScore += 3 // Changed from 2 to 3
				result.Reasons = append(result.Reasons, fmt.Sprintf("critical pattern: %s", pattern))
			} else {
				result.RiskScore++
				result.Reasons = append(result.Reasons, fmt.Sprintf("suspicious pattern: %s", pattern))
			}
		}
	}

	// Check SQL injection patterns
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(lowerInput) {
			result.RiskScore += 3 // Higher weight for SQL injection
			result.Reasons = append(result.Reasons, "SQL injection pattern detected")
			break
		}
	}

	// Check system override patterns
	for _, pattern := range systemOverridePatterns {
		if pattern.MatchString(lowerInput) {
			result.RiskScore++
			result.Reasons = append(result.Reasons, "system override attempt detected")
			break // Only count once
		}
	}

	// Check special character ratio
	if hasExcessiveSpecialChars(result.Sanitized) {
		result.RiskScore++
		result.Reasons = append(result.Reasons, "excessive special characters")
	}

	// Check for repeated patterns (spam detection)
	if hasRepeatedPatterns(result.Sanitized) {
		result.RiskScore++
		result.Reasons = append(result.Reasons, "repeated patterns detected")
	}

	// Final risk assessment
	if result.RiskScore >= RiskThreshold {
		result.Valid = false
		return result, fmt.Errorf("input blocked: risk score %d/%d - %s",
			result.RiskScore, RiskThreshold, strings.Join(result.Reasons, ", "))
	}

	return result, nil
}

// SanitizeInput cleans and normalizes input
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters (except newline, tab, carriage return)
	var sanitized strings.Builder
	for _, r := range input {
		if r == '\n' || r == '\r' || r == '\t' || !unicode.IsControl(r) {
			sanitized.WriteRune(r)
		}
	}

	input = sanitized.String()

	// Normalize whitespace
	input = strings.TrimSpace(input)

	// Collapse multiple spaces
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")

	// Unicode normalization
	input = normalizeUnicode(input)

	return input
}

// normalizeUnicode prevents homograph attacks
func normalizeUnicode(input string) string {
	// NFKC normalization: Compatibility decomposition followed by canonical composition
	normalized := norm.NFKC.String(input)

	// Filter to allowed Unicode ranges
	var result strings.Builder
	for _, r := range normalized {
		if isAllowedUnicode(r) {
			result.WriteRune(r)
		} else {
			// Replace suspicious Unicode with space
			result.WriteRune(' ')
		}
	}

	return result.String()
}

// isAllowedUnicode checks if a rune is in allowed Unicode ranges
func isAllowedUnicode(r rune) bool {
	return (r >= 0x0020 && r <= 0x007E) || // Basic Latin + ASCII punctuation
		(r >= 0x00A0 && r <= 0x00FF) || // Latin-1 Supplement
		unicode.IsSpace(r) ||
		unicode.IsPunct(r)
}

// EstimateTokens provides rough token estimation
func EstimateTokens(text string) int {
	words := strings.Fields(text)
	tokenCount := 0

	for _, word := range words {
		// Average: 1 word ≈ 1.3 tokens
		tokenCount += len(word)/4 + 1
	}

	return tokenCount
}

// hasExcessiveSpecialChars detects too many special characters
func hasExcessiveSpecialChars(input string) bool {
	specialCharCount := 0
	for _, char := range input {
		if !isAlphanumericOrCommon(char) {
			specialCharCount++
		}
	}

	if len(input) == 0 {
		return false
	}

	ratio := float64(specialCharCount) / float64(len(input))
	return ratio > 0.3
}

// hasRepeatedPatterns detects spam/garbage
// hasRepeatedPatterns detects spam/garbage
func hasRepeatedPatterns(input string) bool {
	// Check for same character repeated many times (manual approach)
	if len(input) < 11 {
		return false
	}

	// Count consecutive identical characters
	maxConsecutive := 1
	currentConsecutive := 1

	runes := []rune(input) // Handle Unicode properly
	for i := 1; i < len(runes); i++ {
		if runes[i] == runes[i-1] {
			currentConsecutive++
			if currentConsecutive > maxConsecutive {
				maxConsecutive = currentConsecutive
			}
		} else {
			currentConsecutive = 1
		}
	}

	// If any character repeats more than 10 times consecutively
	if maxConsecutive > 10 {
		return true
	}

	// Check for same word repeated many times
	words := strings.Fields(input)
	if len(words) < 5 {
		return false // Too short to determine
	}

	wordCount := make(map[string]int)
	for _, word := range words {
		wordCount[strings.ToLower(word)]++
		if wordCount[strings.ToLower(word)] > 5 {
			return true
		}
	}

	return false
}

// isAlphanumericOrCommon checks if character is common/safe
func isAlphanumericOrCommon(r rune) bool {
	return unicode.IsLetter(r) ||
		unicode.IsNumber(r) ||
		unicode.IsSpace(r) ||
		strings.ContainsRune(".,?!-_'\"()[]{}:;", r)
}

// DetectPII detects personally identifiable information
func DetectPII(input string) (hasPII bool, piiTypes []string) {
	for piiType, pattern := range piiPatterns {
		if pattern.MatchString(input) {
			hasPII = true
			piiTypes = append(piiTypes, piiType)
		}
	}
	return
}

// MaskPII masks personally identifiable information
func MaskPII(input string) string {
	masked := input

	// Mask email
	masked = piiPatterns["email"].ReplaceAllString(masked, "[EMAIL_REDACTED]")

	// Mask credit card
	masked = piiPatterns["credit_card"].ReplaceAllString(masked, "[CARD_REDACTED]")

	// Mask SSN
	masked = piiPatterns["ssn"].ReplaceAllString(masked, "[SSN_REDACTED]")

	// Mask phone
	masked = piiPatterns["phone"].ReplaceAllString(masked, "[PHONE_REDACTED]")

	return masked
}

// isCriticalPattern identifies high-risk patterns
func isCriticalPattern(pattern string) bool {
	criticalPatterns := []string{
		"ignore previous instructions",
		"ignore all previous",
		"forget previous",
		"you are now",
		"act as",
	}

	for _, critical := range criticalPatterns {
		if pattern == critical {
			return true
		}
	}
	return false
}
