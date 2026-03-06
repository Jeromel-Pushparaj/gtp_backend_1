package handler

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SpecHandler handles spec-related requests
type SpecHandler struct {
	orchestrator *orchestration.Orchestrator
}

// NewSpecHandler creates a new spec handler
func NewSpecHandler(orch *orchestration.Orchestrator) *SpecHandler {
	return &SpecHandler{
		orchestrator: orch,
	}
}

// UploadSpecRequest represents the upload spec request
type UploadSpecRequest struct {
	Name    string `json:"name" binding:"required"`
	Service string `json:"service" binding:"required"`
	TeamID  string `json:"team_id" binding:"required"`
	RunMode string `json:"run_mode"` // smoke, full, nightly
}

// UploadSpec handles spec upload and processing
func (h *SpecHandler) UploadSpec(c *gin.Context) {
	var content []byte
	var format string
	var name, service, teamID string
	var runMode workflow.RunMode

	// Check Content-Type to determine how to parse the request
	contentType := c.GetHeader("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart form upload
		if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to parse form data",
			})
			return
		}

		// Get file from form
		file, header, err := c.Request.FormFile("spec")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Spec file is required (field name: 'spec')",
			})
			return
		}
		defer file.Close()

		// Read file content
		content, err = io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read spec file",
			})
			return
		}

		// Determine format from file extension
		ext := strings.ToLower(filepath.Ext(header.Filename))
		format = "yaml"
		if ext == ".json" {
			format = "json"
		}

		// Get form values
		name = c.PostForm("name")
		service = c.PostForm("service")
		teamID = c.PostForm("team_id")
		runModeStr := c.PostForm("run_mode")

		// Set defaults
		if name == "" {
			name = header.Filename
		}
		if service == "" {
			service = "default"
		}
		if teamID == "" {
			teamID = "default-team"
		}
		if runModeStr != "" {
			runMode = workflow.RunMode(runModeStr)
		} else {
			runMode = workflow.RunModeFull
		}
	} else {
		// Handle raw body upload (YAML/JSON)
		var err error
		content, err = io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read request body",
			})
			return
		}

		// Determine format from Content-Type
		if strings.Contains(contentType, "json") {
			format = "json"
		} else {
			format = "yaml"
		}

		// Set defaults for raw upload
		name = "uploaded-spec"
		service = "default"
		teamID = "default-team"
		runMode = workflow.RunModeFull
	}

	// Validate run mode
	if !isValidRunMode(runMode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid run mode. Must be: smoke, full, or nightly",
		})
		return
	}

	// Generate spec ID
	specID := uuid.New().String()

	// Process spec
	result, err := h.orchestrator.ProcessSpec(
		c.Request.Context(),
		specID,
		teamID,
		content,
		format,
		runMode,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process spec",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"spec_id":             specID,
		"name":                name,
		"service":             service,
		"team_id":             teamID,
		"workflow_id":         result.WorkflowID,
		"endpoints_extracted": result.EndpointsExtracted,
		"jobs_created":        result.JobsCreated,
		"jobs_enqueued":       result.JobsEnqueued,
		"run_mode":            runMode,
		"metadata":            result.Metadata,
	})
}

// ValidateSpec validates a spec without creating jobs
func (h *SpecHandler) ValidateSpec(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to parse form data",
		})
		return
	}

	// Get file from form
	file, header, err := c.Request.FormFile("spec")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Spec file is required",
		})
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read spec file",
		})
		return
	}

	// Determine format
	ext := strings.ToLower(filepath.Ext(header.Filename))
	format := "yaml"
	if ext == ".json" {
		format = "json"
	}

	// Validate spec
	result, err := h.orchestrator.ValidateSpec(c.Request.Context(), content, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Validation failed",
		})
		return
	}

	if !result.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":  false,
			"errors": result.Errors,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSpecStatus gets the status of a spec processing
func (h *SpecHandler) GetSpecStatus(c *gin.Context) {
	specID := c.Param("spec_id")

	// TODO: Implement status retrieval from storage
	// For now, return placeholder
	c.JSON(http.StatusOK, gin.H{
		"spec_id": specID,
		"status":  "processing",
		"message": "Status retrieval not yet implemented",
	})
}

// isValidRunMode validates run mode
func isValidRunMode(mode workflow.RunMode) bool {
	return mode == workflow.RunModeSmoke ||
		mode == workflow.RunModeFull ||
		mode == workflow.RunModeNightly
}
