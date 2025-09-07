package handlers

import (
	"net/http"

	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/services/analytics/internal/service"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service *service.AnalyticsService
}

func NewAnalyticsHandler(service *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
	}
}

// GetAnalytics godoc
// @Summary Get analytics for a short code
// @Description Get click analytics for a specific short code
// @Tags analytics
// @Accept json
// @Produce json
// @Param short_code path string true "Short code"
// @Success 200 {object} service.AnalyticsStats
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /analytics/{short_code} [get]
func (h *AnalyticsHandler) GetAnalytics(c *gin.Context) {
	shortCode := c.Param("short_code")
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Short code is required"})
		return
	}

	stats, err := h.service.GetAnalytics(shortCode)
	if err != nil {
		logger.Error("Failed to get analytics", "short_code", shortCode, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Health godoc
// @Summary Health check
// @Description Check if the analytics service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *AnalyticsHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "analytics",
	})
}

// Metrics godoc
// @Summary Get service metrics
// @Description Get basic metrics about the analytics service
// @Tags metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metrics [get]
func (h *AnalyticsHandler) Metrics(c *gin.Context) {
	// Basic metrics - could be enhanced with Prometheus metrics
	c.JSON(http.StatusOK, gin.H{
		"service": "analytics",
		"uptime":  "running", // Could add actual uptime calculation
		"version": "1.0.0",
	})
}
