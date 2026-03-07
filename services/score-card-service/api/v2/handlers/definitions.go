package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jeromelp/gtp_backen_1/service/score-card-service/internal/scorecards"
)

// GetAllScorecardDefinitions handles GET /api/v2/scorecards/definitions
func (h *ScorecardV2Handler) GetAllScorecardDefinitions(c *gin.Context) {
	definitions := scorecards.GetAllScorecardDefinitions()

	c.JSON(http.StatusOK, gin.H{
		"count":      len(definitions),
		"scorecards": definitions,
	})
}

// GetScorecardDefinition handles GET /api/v2/scorecards/definitions/:name
func (h *ScorecardV2Handler) GetScorecardDefinition(c *gin.Context) {
	name := c.Param("name")

	scorecard, found := getScorecardByName(name)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Scorecard definition not found",
		})
		return
	}

	c.JSON(http.StatusOK, scorecard)
}

