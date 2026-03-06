package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/executor"
)

func main() {
	// Parse command-line flags
	manifestPath := flag.String("manifest", "", "Path to test manifest JSON file (required)")
	outputPath := flag.String("output", "", "Path to output report JSON file (optional, defaults to ./reports/{manifest_id}-report.json)")
	timeout := flag.Int("timeout", 30, "Default timeout in seconds for HTTP requests")
	maxConcurrency := flag.Int("concurrency", 10, "Maximum concurrent test executions (for parallel mode)")
	flag.Parse()

	// Validate required flags
	if *manifestPath == "" {
		fmt.Println("Error: -manifest flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Load test manifest
	log.Printf("Loading test manifest from: %s", *manifestPath)
	manifest, err := executor.LoadManifest(*manifestPath)
	if err != nil {
		log.Fatalf("Failed to load manifest: %v", err)
	}

	log.Printf("Loaded manifest: %s (%s)", manifest.Name, manifest.ID)
	log.Printf("Tests to execute: %d", len(manifest.Tests))

	// Create executor engine
	engine := executor.NewEngine(executor.Config{
		Timeout:        time.Duration(*timeout) * time.Second,
		MaxConcurrency: *maxConcurrency,
	})

	// Execute tests
	log.Println("Starting test execution...")
	ctx := context.Background()
	report, err := engine.Execute(ctx, manifest)
	if err != nil {
		log.Fatalf("Test execution failed: %v", err)
	}

	// Determine output path
	reportPath := *outputPath
	if reportPath == "" {
		// Default to ./reports/{manifest_id}-report.json
		reportsDir := "./reports"
		if err := os.MkdirAll(reportsDir, 0755); err != nil {
			log.Fatalf("Failed to create reports directory: %v", err)
		}
		reportPath = filepath.Join(reportsDir, fmt.Sprintf("%s-report.json", manifest.ID))
	}

	// Save report
	log.Printf("Saving execution report to: %s", reportPath)
	if err := executor.SaveReport(report, reportPath); err != nil {
		log.Fatalf("Failed to save report: %v", err)
	}

	// Print summary
	separator := "============================================================"
	fmt.Println("\n" + separator)
	fmt.Println("TEST EXECUTION SUMMARY")
	fmt.Println(separator)
	fmt.Printf("Manifest:     %s\n", report.ManifestName)
	fmt.Printf("Duration:     %dms\n", report.Duration)
	fmt.Printf("Total Tests:  %d\n", report.Summary.Total)
	fmt.Printf("Passed:       %d\n", report.Summary.Passed)
	fmt.Printf("Failed:       %d\n", report.Summary.Failed)
	fmt.Printf("Skipped:      %d\n", report.Summary.Skipped)
	fmt.Printf("Errors:       %d\n", report.Summary.Errors)
	fmt.Println(separator)

	// Print detailed results
	if report.Summary.Failed > 0 || report.Summary.Errors > 0 {
		fmt.Println("\nFAILED TESTS:")
		for _, result := range report.Results {
			if result.Status == "failed" || result.Status == "error" {
				fmt.Printf("\n  ❌ %s (%s)\n", result.TestName, result.TestID)
				if result.Error != "" {
					fmt.Printf("     Error: %s\n", result.Error)
				}
				for _, assertion := range result.Assertions {
					if !assertion.Passed {
						fmt.Printf("     - %s: %s\n", assertion.Type, assertion.Message)
					}
				}
			}
		}
	}

	// Print passed tests
	if report.Summary.Passed > 0 {
		fmt.Println("\nPASSED TESTS:")
		for _, result := range report.Results {
			if result.Status == "passed" {
				fmt.Printf("  ✅ %s (%dms)\n", result.TestName, result.Duration)
			}
		}
	}

	fmt.Printf("\nReport saved to: %s\n", reportPath)

	// Exit with appropriate code
	if report.Summary.Failed > 0 || report.Summary.Errors > 0 {
		os.Exit(1)
	}
}
