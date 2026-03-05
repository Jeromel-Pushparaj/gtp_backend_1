package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/handler"
)

// SetupRoutes configures all routes for the service
func SetupRoutes(router *gin.Engine, serviceHandler *handler.ServiceHandler) {
	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Organization-scoped service routes
		org := api.Group("/org/:org_id")
		{
			// Service routes under organization
			service := org.Group("/service")
			{
				// Fetch/Create service from sonar-shell-test API
				service.POST("", serviceHandler.FetchServices)

				// Get all services for an organization
				service.GET("", serviceHandler.GetAllServices)

				// Get specific service by ID within organization
				service.GET("/:id", serviceHandler.GetService)
			}
		}

		// Legacy routes (for backward compatibility - optional)
		service := api.Group("/service")
		{
			// Fetch/Create service from sonar-shell-test API
			service.POST("", serviceHandler.FetchServicesLegacy)

			// Get all services from cache
			service.GET("", serviceHandler.GetAllServicesLegacy)

			// Get specific service from cache by ID
			service.GET("/:id", serviceHandler.GetServiceLegacy)
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "service-catalog",
		})
	})

	// Serve OpenAPI YAML file
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.File("./openapi.yaml")
	})

	// Serve Swagger UI
	router.GET("/swagger", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Service Catalog API - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "/openapi.yaml",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>
		`)
	})
}
