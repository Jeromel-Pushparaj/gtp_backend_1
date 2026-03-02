package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

type Server struct {
	mcpServer *mcp_golang.Server
	baseURL   string
	apiKey    string
}

func NewServer(baseURL, apiKey string) *Server {
	return &Server{
		mcpServer: mcp_golang.NewServer(stdio.NewStdioServerTransport()),
		baseURL:   baseURL,
		apiKey:    apiKey,
	}
}

func (s *Server) RegisterTools() error {
	if err := s.registerHealthCheck(); err != nil {
		return err
	}
	if err := s.registerSonarCloudTools(); err != nil {
		return err
	}
	if err := s.registerGitHubTools(); err != nil {
		return err
	}
	if err := s.registerJiraTools(); err != nil {
		return err
	}
	if err := s.registerMetricsTools(); err != nil {
		return err
	}
	if err := s.registerOrgTools(); err != nil {
		return err
	}
	if err := s.registerRepoTools(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Serve() error {
	return s.mcpServer.Serve()
}

func (s *Server) makeRequest(method, endpoint string, queryParams map[string]string, body interface{}) (string, error) {
	fullURL := s.baseURL + endpoint

	if len(queryParams) > 0 {
		params := url.Values{}
		for k, v := range queryParams {
			params.Add(k, v)
		}
		fullURL += "?" + params.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return "", err
	}

	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(responseBody))
	}

	return string(responseBody), nil
}

type HealthCheckArgs struct{}

func (s *Server) registerHealthCheck() error {
	return s.mcpServer.RegisterTool("health_check", "Check the health status of the service", func(args HealthCheckArgs) (*mcp_golang.ToolResponse, error) {
		result, err := s.makeRequest("GET", "/health", nil, nil)
		if err != nil {
			return nil, err
		}
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
	})
}

