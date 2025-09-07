package routes

import (
	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/services/api-gateway/internal/handlers"
	"github.com/chthon/shortlink/services/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes for the API Gateway
func SetupRoutes(r *gin.Engine, h *handlers.Handler, cfg *config.Config) {
	// Health check
	r.GET("/health", h.HealthCheck)
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Chthon ShortLink API Gateway",
			"version": "1.0.0",
			"docs":    "/docs",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", h.Login)
			auth.POST("/register", h.Register)
			auth.POST("/refresh", h.RefreshToken)
		}

		// Public routes (no auth required)
		public := v1.Group("/public")
		{
			// Redirect handling - these need to be fast
			public.GET("/:code", h.RedirectProxy)
			public.HEAD("/:code", h.RedirectProxy)
		}

		// Protected routes (JWT required)
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(h.GetJWTManager()))
		{
			// User routes
			user := protected.Group("/user")
			{
				user.GET("/profile", h.UserManagementProxy)
				user.PUT("/profile", h.UserManagementProxy)
				user.DELETE("/profile", h.UserManagementProxy)
				user.GET("/links", h.ShortlinkProxy)
			}

			// Shortlink routes
			links := protected.Group("/links")
			{
				links.POST("/", h.ShortlinkProxy)             // Create shortlink
				links.GET("/", h.ShortlinkProxy)              // List user's shortlinks
				links.GET("/:id", h.ShortlinkProxy)           // Get specific shortlink
				links.PUT("/:id", h.ShortlinkProxy)           // Update shortlink
				links.DELETE("/:id", h.ShortlinkProxy)        // Delete shortlink
				links.GET("/:id/analytics", h.AnalyticsProxy) // Get link analytics
			}

			// Analytics routes
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/dashboard", h.AnalyticsProxy)
				analytics.GET("/links/:code", h.AnalyticsProxy)
				analytics.GET("/links/:code/export", h.AnalyticsProxy)
			}
		}

		// Premium routes (premium/admin only)
		premium := v1.Group("/premium")
		premium.Use(middleware.JWTAuth(h.GetJWTManager()))
		premium.Use(middleware.RequireRole("premium"))
		{
			premium.GET("/analytics/advanced", h.AnalyticsProxy)
			premium.POST("/links/bulk", h.ShortlinkProxy)
			premium.GET("/links/export", h.ShortlinkProxy)
		}

		// Admin routes (admin only)
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(h.GetJWTManager()))
		admin.Use(middleware.RequireRole("admin"))
		{
			// Dashboard and statistics
			admin.GET("/stats", h.AdminStats)
			admin.GET("/dashboard", h.AdminStats)

			// User management
			users := admin.Group("/users")
			{
				users.GET("/", h.UserManagementProxy)
				users.GET("/:id", h.UserManagementProxy)
				users.PUT("/:id", h.UserManagementProxy)
				users.DELETE("/:id", h.UserManagementProxy)
				users.PUT("/:id/role", h.UserManagementProxy)
				users.PUT("/:id/status", h.UserManagementProxy)
			}

			// Link management
			adminLinks := admin.Group("/links")
			{
				adminLinks.GET("/", h.ShortlinkProxy)
				adminLinks.GET("/:id", h.ShortlinkProxy)
				adminLinks.PUT("/:id", h.ShortlinkProxy)
				adminLinks.DELETE("/:id", h.ShortlinkProxy)
			}

			// System management
			system := admin.Group("/system")
			{
				system.GET("/health", h.HealthCheck)
				system.GET("/metrics", h.AnalyticsProxy)
				system.POST("/maintenance", h.AdminStats)
			}
		}

		// API Key routes (for external integrations)
		api := v1.Group("/api")
		api.Use(middleware.APIKeyAuth())
		{
			api.POST("/links", h.ShortlinkProxy)            // Create link via API
			api.GET("/links/:code", h.ShortlinkProxy)       // Get link info via API
			api.GET("/links/:code/stats", h.AnalyticsProxy) // Get stats via API
		}
	}

	// Webhook routes (for external services)
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/analytics", h.AnalyticsProxy)
		webhooks.POST("/notifications", h.UserManagementProxy)
	}

	// Metrics endpoint for Prometheus
	r.GET("/metrics", middleware.Metrics(), func(c *gin.Context) {
		// In production, this would serve Prometheus metrics
		c.JSON(200, gin.H{
			"message": "Metrics endpoint - integrate with Prometheus",
		})
	})

	// Documentation routes
	docs := r.Group("/docs")
	{
		docs.Static("/", "./docs")
		docs.GET("/swagger", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Swagger documentation",
				"url":     "/docs/swagger.json",
			})
		})
	}
}
