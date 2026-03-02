package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/approval-service/resources"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, resources.ErrorResponse{
					Success: false,
					Error:   "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
