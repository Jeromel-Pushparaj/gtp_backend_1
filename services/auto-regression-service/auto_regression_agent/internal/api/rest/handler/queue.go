package handler

import (
	"net/http"

	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration"
	"github.com/gin-gonic/gin"
)

// QueueHandler handles queue-related requests
type QueueHandler struct {
	orchestrator *orchestration.Orchestrator
}

// NewQueueHandler creates a new queue handler
func NewQueueHandler(orch *orchestration.Orchestrator) *QueueHandler {
	return &QueueHandler{
		orchestrator: orch,
	}
}

// GetStats returns queue statistics
func (h *QueueHandler) GetStats(c *gin.Context) {
	stats, err := h.orchestrator.GetQueueStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get queue stats",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

