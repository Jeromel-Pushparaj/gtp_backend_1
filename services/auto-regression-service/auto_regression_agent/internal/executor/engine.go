package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/execution"
)

// Engine executes test manifests deterministically
type Engine struct {
	httpClient  *HTTPClient
	validator   *Validator
	varResolver *VariableResolver
}

// Config represents executor configuration
type Config struct {
	Timeout        time.Duration
	MaxConcurrency int
}

// NewEngine creates a new test execution engine
func NewEngine(config Config) *Engine {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 10
	}

	return &Engine{
		httpClient:  NewHTTPClient(config.Timeout),
		validator:   NewValidator(),
		varResolver: NewVariableResolver(),
	}
}

// Execute executes a test manifest
func (e *Engine) Execute(ctx context.Context, manifest *execution.TestManifest) (*execution.ExecutionReport, error) {
	startTime := time.Now()

	report := &execution.ExecutionReport{
		ManifestID:   manifest.ID,
		ManifestName: manifest.Name,
		StartedAt:    startTime,
		Results:      make([]execution.TestResult, 0),
		Summary: execution.ExecutionSummary{
			Total: len(manifest.Tests),
		},
	}

	// Initialize variable resolver with manifest variables
	e.varResolver.SetVariables(manifest.Variables)

	// Execute setup steps
	if err := e.executeSetup(ctx, manifest.Setup, manifest.BaseURL); err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}

	// Execute tests
	if manifest.Config.Parallel {
		e.executeParallel(ctx, manifest, report)
	} else {
		e.executeSequential(ctx, manifest, report)
	}

	// Execute teardown steps
	if err := e.executeTeardown(ctx, manifest.Teardown, manifest.BaseURL); err != nil {
		// Log but don't fail the entire execution
		fmt.Printf("Warning: teardown failed: %v\n", err)
	}

	// Finalize report
	report.CompletedAt = time.Now()
	report.Duration = report.CompletedAt.Sub(report.StartedAt).Milliseconds()

	// Calculate summary
	for _, result := range report.Results {
		switch result.Status {
		case execution.TestStatusPassed:
			report.Summary.Passed++
		case execution.TestStatusFailed:
			report.Summary.Failed++
		case execution.TestStatusSkipped:
			report.Summary.Skipped++
		case execution.TestStatusError:
			report.Summary.Errors++
		}
	}

	return report, nil
}

// executeSequential executes tests sequentially
func (e *Engine) executeSequential(ctx context.Context, manifest *execution.TestManifest, report *execution.ExecutionReport) {
	for _, test := range manifest.Tests {
		result := e.executeTest(ctx, test, manifest.BaseURL)
		report.Results = append(report.Results, result)

		// Stop on failure if configured
		if manifest.Config.StopOnFailure && result.Status == execution.TestStatusFailed {
			// Mark remaining tests as skipped
			for i := len(report.Results); i < len(manifest.Tests); i++ {
				skippedResult := execution.TestResult{
					TestID:   manifest.Tests[i].ID,
					TestName: manifest.Tests[i].Name,
					Status:   execution.TestStatusSkipped,
				}
				report.Results = append(report.Results, skippedResult)
			}
			break
		}
	}
}

// executeParallel executes tests in parallel
func (e *Engine) executeParallel(ctx context.Context, manifest *execution.TestManifest, report *execution.ExecutionReport) {
	maxConcurrency := manifest.Config.MaxConcurrency
	if maxConcurrency == 0 {
		maxConcurrency = 10
	}

	// Create semaphore for concurrency control
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := make([]execution.TestResult, len(manifest.Tests))

	for i, test := range manifest.Tests {
		wg.Add(1)
		go func(index int, tc execution.TestCase) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			result := e.executeTest(ctx, tc, manifest.BaseURL)

			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, test)
	}

	wg.Wait()
	report.Results = results
}

// executeTest executes a single test case
func (e *Engine) executeTest(ctx context.Context, test execution.TestCase, baseURL string) execution.TestResult {
	startTime := time.Now()

	result := execution.TestResult{
		TestID:     test.ID,
		TestName:   test.Name,
		StartedAt:  startTime,
		Assertions: make([]execution.AssertionResult, 0),
	}

	// Resolve variables in test
	resolvedTest, err := e.varResolver.ResolveTest(test)
	if err != nil {
		result.Status = execution.TestStatusError
		result.Error = fmt.Sprintf("variable resolution failed: %v", err)
		result.CompletedAt = time.Now()
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	// Execute HTTP request
	response, err := e.httpClient.Execute(ctx, HTTPRequest{
		BaseURL:     baseURL,
		Endpoint:    resolvedTest.Endpoint,
		Method:      resolvedTest.Method,
		Headers:     resolvedTest.Headers,
		PathParams:  resolvedTest.PathParams,
		QueryParams: resolvedTest.QueryParams,
		Payload:     resolvedTest.Payload,
		Timeout:     time.Duration(resolvedTest.Timeout) * time.Second,
	})

	if err != nil {
		result.Status = execution.TestStatusError
		result.Error = fmt.Sprintf("HTTP request failed: %v", err)
		result.CompletedAt = time.Now()
		result.Duration = time.Since(startTime).Milliseconds()
		return result
	}

	// Store request and response details
	result.Request = e.buildRequestDetails(resolvedTest, baseURL)
	result.Response = e.buildResponseDetails(response)

	// Validate assertions
	assertionResults := e.validator.Validate(resolvedTest.Assertions, response)
	result.Assertions = assertionResults

	// Determine overall test status
	allPassed := true
	for _, ar := range assertionResults {
		if !ar.Passed {
			allPassed = false
			break
		}
	}

	if allPassed {
		result.Status = execution.TestStatusPassed
	} else {
		result.Status = execution.TestStatusFailed
	}

	result.CompletedAt = time.Now()
	result.Duration = time.Since(startTime).Milliseconds()

	return result
}

// executeSetup executes setup steps
func (e *Engine) executeSetup(ctx context.Context, steps []execution.SetupStep, baseURL string) error {
	for _, step := range steps {
		response, err := e.httpClient.Execute(ctx, HTTPRequest{
			BaseURL:  baseURL,
			Endpoint: step.Endpoint,
			Method:   step.Method,
			Payload:  step.Payload,
		})

		if err != nil {
			return fmt.Errorf("setup step %s failed: %w", step.ID, err)
		}

		// Extract variables from response
		if step.Extract != nil {
			for varName, jsonPath := range step.Extract {
				value, err := extractJSONPath(response.Body, jsonPath)
				if err != nil {
					return fmt.Errorf("failed to extract %s from setup step %s: %w", varName, step.ID, err)
				}
				e.varResolver.SetVariable(varName, value)
			}
		}
	}
	return nil
}

// executeTeardown executes teardown steps
func (e *Engine) executeTeardown(ctx context.Context, steps []execution.TeardownStep, baseURL string) error {
	for _, step := range steps {
		_, err := e.httpClient.Execute(ctx, HTTPRequest{
			BaseURL:  baseURL,
			Endpoint: step.Endpoint,
			Method:   step.Method,
		})

		if err != nil {
			return fmt.Errorf("teardown step %s failed: %w", step.ID, err)
		}
	}
	return nil
}

// buildRequestDetails builds request details for the report
func (e *Engine) buildRequestDetails(test execution.TestCase, baseURL string) execution.RequestDetails {
	url := baseURL + test.Endpoint

	// Replace path parameters in URL
	for key, value := range test.PathParams {
		url = replacePathParam(url, key, value)
	}

	// Generate curl command for easy debugging and manual testing
	httpReq := HTTPRequest{
		BaseURL:     baseURL,
		Endpoint:    test.Endpoint,
		Method:      test.Method,
		Headers:     test.Headers,
		PathParams:  test.PathParams,
		QueryParams: test.QueryParams,
		Payload:     test.Payload,
	}
	curlCommand := e.httpClient.GenerateCurl(httpReq)

	return execution.RequestDetails{
		Method:      test.Method,
		URL:         url,
		Curl:        curlCommand,
		Headers:     test.Headers,
		QueryParams: test.QueryParams,
		Body:        test.Payload,
	}
}

// buildResponseDetails builds response details for the report
func (e *Engine) buildResponseDetails(response *HTTPResponse) execution.ResponseDetails {
	return execution.ResponseDetails{
		StatusCode:   response.StatusCode,
		Headers:      response.Headers,
		Body:         response.Body,
		RawBody:      string(response.RawBody),
		ResponseTime: response.ResponseTime,
	}
}

// LoadManifest loads a test manifest from a JSON file
func LoadManifest(path string) (*execution.TestManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest execution.TestManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	return &manifest, nil
}

// SaveReport saves an execution report to a JSON file
func SaveReport(report *execution.ExecutionReport, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}
