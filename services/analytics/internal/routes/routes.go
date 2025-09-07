package routes

import (
	"github.com/chthon/shortlink/services/analytics/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, analyticsHandler *handlers.AnalyticsHandler) {
	// Health check
	router.GET("/health", analyticsHandler.Health)

	// Metrics
	router.GET("/metrics", analyticsHandler.Metrics)

	// Analytics routes
	api := router.Group("/api/v1")
	{
		analytics := api.Group("/analytics")
		{
			// Get analytics for a specific short code
			analytics.GET("/:short_code", analyticsHandler.GetAnalytics)
		}
	}
}
