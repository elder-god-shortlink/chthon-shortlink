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
	"github.com/chthon/shortlink/services/shortlink/internal/handlers"
	"github.com/chthon/shortlink/services/shortlink/internal/repository"
	"github.com/chthon/shortlink/services/shortlink/internal/routes"
	"github.com/chthon/shortlink/services/shortlink/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadServiceConfig("shortlink", 8082)
	if err := cfg.Validate(); err != nil {
		log.Fatal("Configuration validation failed:", err)
	}

	// Initialize logger
	logger.InitDefaultLogger(cfg, "shortlink-service")
	logger.Info("Starting Shortlink Service")

	// Initialize database connections
	db, err := database.NewDatabaseConnections(cfg)
	if err != nil {
		logger.Error("Failed to connect to databases", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Auto-migrate database schema
	if err := repository.AutoMigrate(db.PostgreSQL); err != nil {
		logger.Error("Failed to migrate database", "error", err)
		os.Exit(1)
	}

	// Initialize repository
	repo := repository.NewRepository(db)

	// Initialize service
	shortlinkService := service.NewShortlinkService(repo, cfg)

	// Initialize handlers
	h := handlers.NewHandler(shortlinkService, cfg)

	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	r := gin.New()
	r.Use(gin.Recovery())

	// Setup routes
	routes.SetupRoutes(r, h)

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Shortlink service starting", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
