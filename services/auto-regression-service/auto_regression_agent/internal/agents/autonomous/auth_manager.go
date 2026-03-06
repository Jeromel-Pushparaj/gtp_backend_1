package autonomous

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// AuthManager manages authentication for API testing
type AuthManager struct {
	securitySchemes map[string]SecurityScheme
	credentials     map[string]string // scheme name -> credential
	tokens          map[string]TokenInfo
	mu              sync.RWMutex
}

// SecurityScheme represents an authentication scheme from OpenAPI spec
type SecurityScheme struct {
	Name         string `json:"name"`
	Type         string `json:"type"`          // apiKey, http, oauth2, openIdConnect
	Scheme       string `json:"scheme"`        // bearer, basic
	In           string `json:"in"`            // header, query, cookie
	HeaderName   string `json:"header_name"`   // For apiKey type
	BearerFormat string `json:"bearer_format"` // JWT, etc.
}

// TokenInfo stores token with expiry
type TokenInfo struct {
	Token     string
	ExpiresAt time.Time
	Type      string // Bearer, Basic, etc.
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return &AuthManager{
		securitySchemes: make(map[string]SecurityScheme),
		credentials:     make(map[string]string),
		tokens:          make(map[string]TokenInfo),
	}
}

// LoadSecuritySchemes loads security schemes from OpenAPI spec
func (am *AuthManager) LoadSecuritySchemes(spec map[string]interface{}) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Navigate to components.securitySchemes
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		log.Printf("Info: No components found in spec")
		return nil
	}

	securitySchemes, ok := components["securitySchemes"].(map[string]interface{})
	if !ok {
		log.Printf("Info: No securitySchemes found in spec")
		return nil
	}

	// Parse each security scheme
	for name, schemeData := range securitySchemes {
		schemeMap, ok := schemeData.(map[string]interface{})
		if !ok {
			continue
		}

		scheme := SecurityScheme{
			Name: name,
		}

		if schemeType, ok := schemeMap["type"].(string); ok {
			scheme.Type = schemeType
		}

		if schemeScheme, ok := schemeMap["scheme"].(string); ok {
			scheme.Scheme = schemeScheme
		}

		if in, ok := schemeMap["in"].(string); ok {
			scheme.In = in
		}

		if headerName, ok := schemeMap["name"].(string); ok {
			scheme.HeaderName = headerName
		}

		if bearerFormat, ok := schemeMap["bearerFormat"].(string); ok {
			scheme.BearerFormat = bearerFormat
		}

		am.securitySchemes[name] = scheme
		log.Printf("🔐 Loaded security scheme: %s (type: %s, scheme: %s)", name, scheme.Type, scheme.Scheme)
	}

	return nil
}

// LoadCredentialsFromEnv loads credentials from environment variables
func (am *AuthManager) LoadCredentialsFromEnv() {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Common environment variable patterns
	envPatterns := map[string]string{
		"API_KEY":         "API_KEY",
		"BEARER_TOKEN":    "BEARER_TOKEN",
		"AUTH_TOKEN":      "AUTH_TOKEN",
		"ACCESS_TOKEN":    "ACCESS_TOKEN",
		"JWT_TOKEN":       "JWT_TOKEN",
		"BASIC_AUTH_USER": "BASIC_AUTH_USER",
		"BASIC_AUTH_PASS": "BASIC_AUTH_PASS",
		"OAUTH_TOKEN":     "OAUTH_TOKEN",
		"X_API_KEY":       "X_API_KEY",
	}

	for schemeName, envKey := range envPatterns {
		if value := os.Getenv(envKey); value != "" {
			am.credentials[schemeName] = value
			log.Printf("🔑 Loaded credential from env: %s", envKey)
		}
	}

	// Also check for scheme-specific env vars
	for schemeName := range am.securitySchemes {
		envKey := strings.ToUpper(strings.ReplaceAll(schemeName, "-", "_"))
		if value := os.Getenv(envKey); value != "" {
			am.credentials[schemeName] = value
			log.Printf("🔑 Loaded credential for scheme '%s' from env: %s", schemeName, envKey)
		}
	}
}

// SetCredential manually sets a credential for a scheme
func (am *AuthManager) SetCredential(schemeName, credential string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.credentials[schemeName] = credential
	log.Printf("🔑 Credential set for scheme: %s", schemeName)
}

// GetAuthHeaders generates authentication headers for a request
func (am *AuthManager) GetAuthHeaders(ctx context.Context, securityRequirements []string) (map[string]string, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	headers := make(map[string]string)

	// If no security requirements specified, try to use any available auth
	if len(securityRequirements) == 0 {
		// Use first available credential
		for schemeName, credential := range am.credentials {
			if scheme, ok := am.securitySchemes[schemeName]; ok {
				am.applyAuthToHeaders(headers, scheme, credential)
				return headers, nil
			}
		}
		// No auth available
		return headers, nil
	}

	// Apply auth for each required security scheme
	for _, schemeName := range securityRequirements {
		scheme, ok := am.securitySchemes[schemeName]
		if !ok {
			log.Printf("Warning: Security scheme '%s' not found in spec", schemeName)
			continue
		}

		credential, ok := am.credentials[schemeName]
		if !ok {
			// Try common fallbacks
			credential = am.findFallbackCredential(scheme)
			if credential == "" {
				log.Printf("Warning: No credential found for security scheme '%s'", schemeName)
				continue
			}
		}

		am.applyAuthToHeaders(headers, scheme, credential)
	}

	return headers, nil
}

// applyAuthToHeaders applies authentication to headers based on scheme type
func (am *AuthManager) applyAuthToHeaders(headers map[string]string, scheme SecurityScheme, credential string) {
	switch scheme.Type {
	case "http":
		// HTTP authentication (Bearer, Basic)
		switch strings.ToLower(scheme.Scheme) {
		case "bearer":
			headers["Authorization"] = fmt.Sprintf("Bearer %s", credential)
			log.Printf("🔐 Applied Bearer token authentication")
		case "basic":
			// Expect credential in format "username:password"
			encoded := base64.StdEncoding.EncodeToString([]byte(credential))
			headers["Authorization"] = fmt.Sprintf("Basic %s", encoded)
			log.Printf("🔐 Applied Basic authentication")
		default:
			headers["Authorization"] = credential
			log.Printf("🔐 Applied custom HTTP authentication: %s", scheme.Scheme)
		}

	case "apiKey":
		// API Key authentication
		headerName := scheme.HeaderName
		if headerName == "" {
			headerName = "X-API-Key" // Default
		}

		switch strings.ToLower(scheme.In) {
		case "header":
			headers[headerName] = credential
			log.Printf("🔐 Applied API Key in header: %s", headerName)
		case "query":
			// Query params handled separately in HTTP client
			log.Printf("🔐 API Key in query parameter: %s (not yet implemented)", headerName)
		case "cookie":
			headers["Cookie"] = fmt.Sprintf("%s=%s", headerName, credential)
			log.Printf("🔐 Applied API Key in cookie: %s", headerName)
		}

	case "oauth2":
		// OAuth2 - treat as bearer token
		headers["Authorization"] = fmt.Sprintf("Bearer %s", credential)
		log.Printf("🔐 Applied OAuth2 token")

	case "openIdConnect":
		// OpenID Connect - treat as bearer token
		headers["Authorization"] = fmt.Sprintf("Bearer %s", credential)
		log.Printf("🔐 Applied OpenID Connect token")

	default:
		log.Printf("Warning: Unknown security scheme type: %s", scheme.Type)
	}
}

// findFallbackCredential tries to find a credential using common patterns
func (am *AuthManager) findFallbackCredential(scheme SecurityScheme) string {
	// Try common credential keys based on scheme type
	switch scheme.Type {
	case "http":
		if strings.ToLower(scheme.Scheme) == "bearer" {
			// Try bearer token variants
			for _, key := range []string{"BEARER_TOKEN", "AUTH_TOKEN", "ACCESS_TOKEN", "JWT_TOKEN"} {
				if cred, ok := am.credentials[key]; ok {
					return cred
				}
			}
		} else if strings.ToLower(scheme.Scheme) == "basic" {
			// Try basic auth variants
			if cred, ok := am.credentials["BASIC_AUTH_USER"]; ok {
				if pass, ok := am.credentials["BASIC_AUTH_PASS"]; ok {
					return fmt.Sprintf("%s:%s", cred, pass)
				}
			}
		}

	case "apiKey":
		// Try API key variants
		for _, key := range []string{"API_KEY", "X_API_KEY"} {
			if cred, ok := am.credentials[key]; ok {
				return cred
			}
		}

	case "oauth2", "openIdConnect":
		// Try OAuth token variants
		for _, key := range []string{"OAUTH_TOKEN", "ACCESS_TOKEN", "BEARER_TOKEN"} {
			if cred, ok := am.credentials[key]; ok {
				return cred
			}
		}
	}

	return ""
}

// GetSecuritySchemes returns all loaded security schemes
func (am *AuthManager) GetSecuritySchemes() map[string]SecurityScheme {
	am.mu.RLock()
	defer am.mu.RUnlock()

	schemes := make(map[string]SecurityScheme)
	for k, v := range am.securitySchemes {
		schemes[k] = v
	}
	return schemes
}
