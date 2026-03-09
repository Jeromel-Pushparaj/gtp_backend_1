package validator

import "regexp"

// Configuration constants
const (
	MaxInputLength = 10000 // Maximum characters allowed
	MinInputLength = 1     // Minimum characters required
	MaxTokens      = 2000  // Maximum estimated tokens
	RiskThreshold  = 3     // Block if risk score >= this value
)

// Suspicious patterns for prompt injection detection
var suspiciousPatterns = []string{
	"ignore previous instructions",
	"ignore all previous",
	"disregard previous",
	"forget previous",
	"new instructions:",
	"system:",
	"assistant:",
	"<script>",
	"javascript:",
	"eval(",
	"exec(",
	"__import__",
	"you are now",
	"pretend to be",
}

// Critical patterns that should have higher weight (instant block)
var criticalPatterns = []string{
	"ignore previous instructions",
	"ignore all previous",
	"forget previous",
	"you are now",
	"disregard previous",
}

// SQL injection patterns
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(union\s+select|drop\s+table|insert\s+into|delete\s+from)`),
	regexp.MustCompile(`(?i)(or\s+['"]?\d+['"]?\s*=\s*['"]?\d+['"]?)`),
	regexp.MustCompile(`(?i)(and\s+['"]?\d+['"]?\s*=\s*['"]?\d+['"]?)`),
	regexp.MustCompile(`(?i)('[^']*'\s*=\s*'[^']*')`), // Single-quoted equality
	regexp.MustCompile(`(?i)("[^"]*"\s*=\s*"[^"]*")`), // Double-quoted equality
	regexp.MustCompile(`(?i)(exec\s*\(|execute\s*\()`),
	regexp.MustCompile(`(?i)(--|;|\/\*|\*\/)`),
}

// System override patterns - more specific to avoid false positives
var systemOverridePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(you\s+are\s+now\s+(a|an|the)?\s*(admin|root|system|hacker|god))`),
	regexp.MustCompile(`(?i)(act\s+as\s+(a|an|the)?\s*(admin|root|system|hacker|god|jailbreak))`),
	regexp.MustCompile(`(?i)(pretend\s+to\s+be\s+(a|an|the)?\s*(admin|root|system|hacker))`),
	regexp.MustCompile(`(?i)(roleplay\s+as\s+(admin|system|root))`),
	regexp.MustCompile(`(?i)(role\s*:\s*system)`),
}

// PII detection patterns (optional)
var piiPatterns = map[string]*regexp.Regexp{
	"email":       regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
	"credit_card": regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`),
	"ssn":         regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
	"phone":       regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
}
