package handler
import (
	"net/http"
	"github.com/gin-gonic/gin"
)
func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	}
}
func Ready() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}
}
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "# Metrics endpoint\n")
	}
}
func NotImplemented() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "This endpoint is not yet implemented",
		})
	}
}