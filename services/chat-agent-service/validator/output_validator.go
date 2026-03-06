package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// Patterns that indicate internal errors that should not be exposed
var internalErrorPatterns = []string{
	"internal server error",
	"database",
	"connection refused",
	"connection failed",
	"stack trace",
	"panic",
	"fatal",
	"localhost",
	"127.0.0.1",
	"10.",      // Private IP ranges
	"192.168.", // Private IP ranges
	"172.16.",  // Private IP ranges
	"sql error",
	"query failed",
	"authentication failed",
	"unauthorized",
}

// System prompt leak patterns
var systemPromptLeakPatterns = []string{
	"you are a helpful assistant",
	"system prompt",
	"internal instruction",
	"your instructions are",
	"i was instructed to",
	"my system prompt",
	"i am programmed to",
	"my instructions say",
	"according to my instructions",
}

// Sensitive data patterns
var sensitiveDataPatterns = []string{
	"api_key",
	"api key",
	"secret_key",
	"secret key",
	"password",
	"bearer ",
	"authorization:",
	"private_key",
	"access_token",
}

// Regex patterns for SHA hashes
// Regex patterns for SHA hashes (case-insensitive)
var (
	// Match full SHA (40 hex chars) - case insensitive
	gitSHAPattern = regexp.MustCompile(`(?i)"sha"\s*:\s*"[a-f0-9]{40}"`)
)

// SanitizeToolResponse removes sensitive data from tool responses
func SanitizeToolResponse(result string) string {
	// Remove SHA hashes (case-insensitive)
	sanitized := gitSHAPattern.ReplaceAllString(result, `"sha":"[REDACTED]"`)

	return sanitized
}

// ValidateToolResponse checks if tool response is safe to send to LLM
func ValidateToolResponse(result string) error {
	if result == "" {
		return fmt.Errorf("tool returned empty result")
	}

	lowerResult := strings.ToLower(result)

	// Check for internal error exposure
	for _, pattern := range internalErrorPatterns {
		if strings.Contains(lowerResult, pattern) {
			return fmt.Errorf("tool response contains internal error information")
		}
	}

	// Check response length (prevent massive responses that waste tokens)
	if len(result) > 50000 {
		return fmt.Errorf("tool response too large: %d characters", len(result))
	}

	return nil
}

// SanitizeToolError converts internal errors to user-friendly messages
func SanitizeToolError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()
	lowerErr := strings.ToLower(errStr)

	// Check if it's an internal error that should be hidden
	for _, pattern := range internalErrorPatterns {
		if strings.Contains(lowerErr, pattern) {
			return "Unable to retrieve data at this time. Please try again later."
		}
	}

	// Check for API errors that expose too much
	if strings.Contains(lowerErr, "api error") {
		return "Unable to retrieve data from the backend service."
	}

	// Safe to return a generic error message
	return "An error occurred while processing your request."
}

// ValidateFinalResponse checks LLM's final response before returning to user
func ValidateFinalResponse(response string) error {
	if response == "" {
		return fmt.Errorf("empty response from LLM")
	}

	lowerResponse := strings.ToLower(response)

	// Check for system prompt leaks
	for _, pattern := range systemPromptLeakPatterns {
		if strings.Contains(lowerResponse, pattern) {
			return fmt.Errorf("response contains system prompt leak")
		}
	}

	// Check for sensitive data exposure
	for _, pattern := range sensitiveDataPatterns {
		if strings.Contains(lowerResponse, pattern) {
			return fmt.Errorf("response may contain sensitive data")
		}
	}

	// Check response length limits
	if len(response) > 10000 {
		return fmt.Errorf("response too long: %d characters", len(response))
	}

	// Allow short responses (e.g., "4", "Yes", "No")
	if len(response) < 1 {
		return fmt.Errorf("response too short: %d characters", len(response))
	}

	return nil
}

// FormatResponse cleans up and improves readability of the final response
func FormatResponse(response string) string {
	// Remove excessive blank lines (more than 2 consecutive newlines)
	re := regexp.MustCompile(`\n{3,}`)
	response = re.ReplaceAllString(response, "\n\n")

	// Trim leading and trailing whitespace
	response = strings.TrimSpace(response)

	return response
}
