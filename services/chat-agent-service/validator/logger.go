package validator

import "log"

// LogValidationResult logs the validation result for monitoring
func LogValidationResult(ip string, input string, result *ValidationResult, err error) {
	if err != nil && result != nil {
		log.Printf("[GUARDRAIL_BLOCK] IP=%s Length=%d RiskScore=%d Reasons=%v Error=%v",
			ip, len(input), result.RiskScore, result.Reasons, err)
	} else if result != nil && result.RiskScore > 0 {
		log.Printf("[GUARDRAIL_WARN] IP=%s Length=%d RiskScore=%d Reasons=%v",
			ip, len(input), result.RiskScore, result.Reasons)
	} else if result != nil {
		log.Printf("[GUARDRAIL_PASS] IP=%s Length=%d Tokens=%d",
			ip, len(input), EstimateTokens(input))
	} else {
		log.Printf("[GUARDRAIL_ERROR] IP=%s Error=%v", ip, err)
	}
}