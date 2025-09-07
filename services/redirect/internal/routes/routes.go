package routes

import (
	"github.com/chthon/shortlink/services/redirect/internal/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the Redirect service
func SetupRoutes(r *gin.Engine, h *handlers.Handler) {
	// Health check
	r.GET("/health", h.HealthCheck)

	// Service info
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "redirect-service",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Main redirect route - this should be the primary route
	r.GET("/:code", h.RedirectHandler)

	// Alternative redirect routes
	r.HEAD("/:code", h.RedirectHandler) // For HEAD requests

	// API routes
	api := r.Group("/api/v1")
	{
		// Preview shortlink without redirecting
		api.GET("/preview/:code", h.PreviewHandler)

		// Check if shortlink exists
		api.GET("/check/:code", h.CheckHandler)

		// QR code generation
		api.GET("/qr/:code", h.QRCodeHandler)

		// Admin statistics
		admin := api.Group("/admin")
		{
			admin.GET("/stats", h.StatsHandler)
		}
	}

	// Internal routes (for service-to-service communication)
	internal := r.Group("/internal")
	{
		internal.GET("/check/:code", h.CheckHandler)
		internal.GET("/stats", h.StatsHandler)
	}
}
