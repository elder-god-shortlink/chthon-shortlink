package routes

import (
	"github.com/chthon/shortlink/services/user-management/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, userHandler *handlers.UserHandler) {
	// Health check
	router.GET("/health", userHandler.Health)

	// Authentication routes (public)
	auth := router.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
	}

	// User management routes
	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			// Get all users (with pagination)
			users.GET("", userHandler.GetAllUsers)

			// Get user by ID
			users.GET("/:id", userHandler.GetUser)

			// Update user
			users.PUT("/:id", userHandler.UpdateUser)

			// Delete user
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}
}
