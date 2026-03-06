package autonomous

import "strings"

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// stripMarkdownCodeBlocks removes markdown code block markers from LLM responses
// Handles formats like:
//
//	```json
//	{...}
//	```
//
// or just:
//
//	```
//	{...}
//	```
//
// Also handles cases where there's explanatory text before the code block:
//
//	Some explanation text...
//	```json
//	{...}
//	```
func stripMarkdownCodeBlocks(response string) string {
	// Trim whitespace
	response = strings.TrimSpace(response)

	// Find the first occurrence of ``` (code block start)
	startIdx := strings.Index(response, "```")
	if startIdx == -1 {
		// No code block found, return as is
		return response
	}

	// Find the closing ``` after the opening one
	endIdx := strings.Index(response[startIdx+3:], "```")
	if endIdx == -1 {
		// No closing code block found, return as is
		return response
	}

	// Extract the content between the code blocks
	// startIdx+3 skips the opening ```
	// We need to also skip the language identifier (e.g., "json\n")
	codeBlockContent := response[startIdx+3 : startIdx+3+endIdx]

	// Remove the language identifier if present (e.g., "json\n")
	lines := strings.Split(codeBlockContent, "\n")
	if len(lines) > 1 {
		// Skip the first line if it looks like a language identifier
		firstLine := strings.TrimSpace(lines[0])
		if firstLine == "json" || firstLine == "JSON" || firstLine == "" {
			lines = lines[1:]
		}
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}
