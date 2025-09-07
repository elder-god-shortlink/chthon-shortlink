package routes

import (
	"github.com/chthon/shortlink/services/shortlink/internal/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the Shortlink service
func SetupRoutes(r *gin.Engine, h *handlers.Handler) {
	// Health check
	r.GET("/health", h.HealthCheck)

	// Service info
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "shortlink-service",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Shortlink operations
		links := v1.Group("/links")
		{
			links.POST("/", h.CreateShortlink)      // Create shortlink
			links.GET("/", h.ListUserShortlinks)    // List user's shortlinks
			links.GET("/:id", h.GetShortlink)       // Get specific shortlink
			links.PUT("/:id", h.UpdateShortlink)    // Update shortlink
			links.DELETE("/:id", h.DeleteShortlink) // Delete shortlink
		}

		// Code lookup (for redirect service)
		v1.GET("/code/:code", h.GetShortlinkByCode)

		// Admin routes
		admin := v1.Group("/admin")
		{
			admin.GET("/stats", h.AdminStats)
		}
	}

	// Internal routes (for service-to-service communication)
	internal := r.Group("/internal")
	{
		internal.GET("/code/:code", h.GetShortlinkByCode) // Used by redirect service
	}
}
