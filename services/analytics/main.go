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
	"github.com/chthon/shortlink/services/analytics/internal/consumer"
	"github.com/chthon/shortlink/services/analytics/internal/handlers"
	"github.com/chthon/shortlink/services/analytics/internal/routes"
	"github.com/chthon/shortlink/services/analytics/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadServiceConfig("analytics", 8084)
	if err := cfg.Validate(); err != nil {
		log.Fatal("Configuration validation failed:", err)
	}

	// Initialize logger
	logger.InitDefaultLogger(cfg, "analytics-service")
	logger.Info("Starting Analytics Service")

	// Initialize database connections
	db, err := database.NewDatabaseConnections(cfg)
	if err != nil {
		logger.Error("Failed to connect to databases", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize service
	analyticsService := service.NewAnalyticsService(db, cfg)

	// Initialize Kafka consumer
	kafkaConsumer := consumer.NewKafkaConsumer(analyticsService, cfg)

	// Start Kafka consumer in background
	go func() {
		logger.Info("Starting Kafka consumer")
		if err := kafkaConsumer.Start(); err != nil {
			logger.Error("Failed to start Kafka consumer", "error", err)
		}
	}()

	// Initialize handlers
	h := handlers.NewAnalyticsHandler(analyticsService)

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
		logger.Info("Analytics service starting", "address", server.Addr)
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

	// Stop Kafka consumer
	kafkaConsumer.Stop()

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
