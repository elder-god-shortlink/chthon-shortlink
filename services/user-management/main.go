package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/database"
	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/chthon/shortlink/services/user-management/internal/handlers"
	"github.com/chthon/shortlink/services/user-management/internal/repository"
	"github.com/chthon/shortlink/services/user-management/internal/routes"
	"github.com/chthon/shortlink/services/user-management/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadServiceConfig("user-management", 8085)
	if err := cfg.Validate(); err != nil {
		log.Fatal("Configuration validation failed:", err)
	}

	// Initialize logger
	logger.InitDefaultLogger(cfg, "user-management-service")
	logger.Info("Starting User Management Service")

	// Initialize database connections
	db, err := database.NewDatabaseConnections(cfg)
	if err != nil {
		logger.Error("Failed to connect to databases", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Auto migrate user models
	if err := db.PostgreSQL.AutoMigrate(&models.User{}); err != nil {
		logger.Error("Failed to migrate database", "error", err)
		os.Exit(1)
	}

	// Initialize repository
	userRepo := repository.NewUserRepository(db.PostgreSQL)

	// Initialize service
	userService := service.NewUserService(userRepo, cfg)

	// Initialize handlers
	h := handlers.NewUserHandler(userService)

	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup routes
	routes.SetupRoutes(router, h)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("User Management Service stopped")
}
