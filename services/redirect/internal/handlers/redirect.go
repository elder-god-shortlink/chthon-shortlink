package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/chthon/shortlink/pkg/utils"
	"github.com/chthon/shortlink/services/redirect/internal/service"
	"github.com/gin-gonic/gin"
)

// Handler contains handlers for redirect operations
type Handler struct {
	service service.RedirectService
	config  *config.Config
}

// NewHandler creates a new handler instance
func NewHandler(service service.RedirectService, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		config:  cfg,
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	response := models.HealthCheckResponse{
		Status:    "ok",
		Services:  map[string]bool{"redirect-service": true},
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}
	c.JSON(http.StatusOK, response)
}

// RedirectHandler handles shortlink redirects
func (h *Handler) RedirectHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: "Short code is required",
			Code:    400,
		})
		return
	}

	// Get original URL
	originalURL, err := h.service.GetOriginalURL(code)
	if err != nil {
		status := http.StatusNotFound
		message := "Shortlink not found"

		if err.Error() == "shortlink has expired" {
			status = http.StatusGone
			message = "Shortlink has expired"
		}

		// Log the failed attempt
		logger.Info("Redirect failed",
			"code", code,
			"error", err.Error(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent())

		c.JSON(status, models.ErrorResponse{
			Error:   "Redirect failed",
			Message: message,
			Code:    status,
		})
		return
	}

	// Log the click event
	ip := utils.ExtractIPFromRequest(map[string]string{
		"X-Real-IP":        c.GetHeader("X-Real-IP"),
		"X-Forwarded-For":  c.GetHeader("X-Forwarded-For"),
		"CF-Connecting-IP": c.GetHeader("CF-Connecting-IP"),
		"Remote-Addr":      c.ClientIP(),
	})

	userAgent := c.Request.UserAgent()
	referer := c.Request.Referer()

	// Log click asynchronously
	go func() {
		if err := h.service.LogClick(code, ip, userAgent, referer); err != nil {
			logger.Error("Failed to log click", "error", err, "code", code)
		}
	}()

	// Log successful redirect
	logger.Info("Redirect successful",
		"code", code,
		"original_url", originalURL,
		"ip", ip,
		"user_agent", userAgent,
		"referer", referer)

	// Perform redirect
	c.Redirect(http.StatusFound, originalURL)
}

// PreviewHandler provides preview information without redirecting
func (h *Handler) PreviewHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: "Short code is required",
			Code:    400,
		})
		return
	}

	// Get original URL (this also caches it)
	originalURL, err := h.service.GetOriginalURL(code)
	if err != nil {
		status := http.StatusNotFound
		message := "Shortlink not found"

		if err.Error() == "shortlink has expired" {
			status = http.StatusGone
			message = "Shortlink has expired"
		}

		c.JSON(status, models.ErrorResponse{
			Error:   "Preview failed",
			Message: message,
			Code:    status,
		})
		return
	}

	// Extract domain for safety
	domain := utils.ExtractDomain(originalURL)

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Preview information retrieved",
		Data: map[string]interface{}{
			"code":         code,
			"original_url": originalURL,
			"domain":       domain,
			"preview_url":  originalURL,
		},
	})
}

// CheckHandler checks if a shortlink exists without redirecting
func (h *Handler) CheckHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: "Short code is required",
			Code:    400,
		})
		return
	}

	// Check if shortlink exists
	_, err := h.service.GetOriginalURL(code)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "shortlink has expired" {
			status = http.StatusGone
		}

		c.JSON(status, models.ErrorResponse{
			Error:   "Check failed",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Shortlink exists",
		Data: map[string]interface{}{
			"code":   code,
			"exists": true,
		},
	})
}

// StatsHandler provides redirect statistics (admin only)
func (h *Handler) StatsHandler(c *gin.Context) {
	// Check if user is admin (should be validated by middleware in API Gateway)
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "Forbidden",
			Message: "Admin access required",
			Code:    403,
		})
		return
	}

	stats := h.service.GetStats()
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Redirect statistics retrieved",
		Data:    stats,
	})
}

// QRCodeHandler generates QR code for shortlink
func (h *Handler) QRCodeHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: "Short code is required",
			Code:    400,
		})
		return
	}

	// Check if shortlink exists
	_, err := h.service.GetOriginalURL(code)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "shortlink has expired" {
			status = http.StatusGone
		}

		c.JSON(status, models.ErrorResponse{
			Error:   "QR code generation failed",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	// Build shortlink URL
	baseDomain := h.config.Server.Host
	if h.config.Server.Port != 80 && h.config.Server.Port != 443 {
		baseDomain = fmt.Sprintf("%s:%d", baseDomain, h.config.Server.Port)
	}
	shortURL := fmt.Sprintf("http://%s/%s", baseDomain, code)

	// In a real implementation, you would generate QR code image here
	// For now, return QR code data URL placeholder
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "QR code generated",
		Data: map[string]interface{}{
			"code":     code,
			"url":      shortURL,
			"qr_code":  "data:image/png;base64,placeholder", // Would be actual QR code
			"download": fmt.Sprintf("/qr/%s/download", code),
		},
	})
}
