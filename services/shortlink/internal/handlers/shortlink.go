package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/chthon/shortlink/services/shortlink/internal/service"
	"github.com/gin-gonic/gin"
)

// Handler contains handlers for shortlink operations
type Handler struct {
	service service.ShortlinkService
	config  *config.Config
}

// NewHandler creates a new handler instance
func NewHandler(service service.ShortlinkService, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		config:  cfg,
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(c *gin.Context) {
	response := models.HealthCheckResponse{
		Status:    "ok",
		Services:  map[string]bool{"shortlink-service": true},
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}
	c.JSON(http.StatusOK, response)
}

// CreateShortlink handles shortlink creation
func (h *Handler) CreateShortlink(c *gin.Context) {
	var req models.CreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(uint); ok {
			userID = &id
		}
	}

	// Create shortlink
	response, err := h.service.CreateShortlink(&req, userID)
	if err != nil {
		logger.Error("Failed to create shortlink", "error", err, "user_id", userID)

		status := http.StatusInternalServerError
		if err.Error() == "invalid URL format" ||
			err.Error() == "custom code can only contain alphanumeric characters" ||
			err.Error() == "custom code already exists" {
			status = http.StatusBadRequest
		}

		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to create shortlink",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	logger.Info("Shortlink created successfully", "code", response.Code, "user_id", userID)
	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Shortlink created successfully",
		Data:    response,
	})
}

// GetShortlink handles getting a specific shortlink
func (h *Handler) GetShortlink(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid number",
			Code:    400,
		})
		return
	}

	// Get user ID from context
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		if userId, ok := uid.(uint); ok {
			userID = &userId
		}
	}

	shortlink, err := h.service.GetShortlink(uint(id), userID)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "access denied: not owner of this shortlink" {
			status = http.StatusForbidden
		}

		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to get shortlink",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Shortlink retrieved successfully",
		Data:    shortlink,
	})
}

// UpdateShortlink handles shortlink updates
func (h *Handler) UpdateShortlink(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid number",
			Code:    400,
		})
		return
	}

	var req models.UpdateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	// Get user ID from context
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		if userId, ok := uid.(uint); ok {
			userID = &userId
		}
	}

	shortlink, err := h.service.UpdateShortlink(uint(id), &req, userID)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "access denied: not owner of this shortlink" {
			status = http.StatusForbidden
		}

		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to update shortlink",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	logger.Info("Shortlink updated successfully", "id", id, "user_id", userID)
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Shortlink updated successfully",
		Data:    shortlink,
	})
}

// DeleteShortlink handles shortlink deletion
func (h *Handler) DeleteShortlink(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid ID",
			Message: "ID must be a valid number",
			Code:    400,
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
			Code:    401,
		})
		return
	}

	userIDUint := userID.(uint)
	err = h.service.DeleteShortlink(uint(id), &userIDUint)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "access denied: not owner of this shortlink" {
			status = http.StatusForbidden
		}

		c.JSON(status, models.ErrorResponse{
			Error:   "Failed to delete shortlink",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	logger.Info("Shortlink deleted successfully", "id", id, "user_id", userID)
	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Shortlink deleted successfully",
		Data:    nil,
	})
}

// ListUserShortlinks handles listing user's shortlinks
func (h *Handler) ListUserShortlinks(c *gin.Context) {
	// Get pagination parameters
	var pagination models.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid pagination parameters",
			Message: err.Error(),
			Code:    400,
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User ID not found in context",
			Code:    401,
		})
		return
	}

	response, err := h.service.ListUserShortlinks(userID.(uint), pagination.Page, pagination.PageSize)
	if err != nil {
		logger.Error("Failed to list user shortlinks", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to list shortlinks",
			Message: err.Error(),
			Code:    500,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Shortlinks retrieved successfully",
		Data:    response,
	})
}

// GetShortlinkByCode handles getting shortlink by code (for redirect service)
func (h *Handler) GetShortlinkByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid code",
			Message: "Code parameter is required",
			Code:    400,
		})
		return
	}

	shortlink, err := h.service.GetShortlinkByCode(code)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "shortlink has expired" {
			status = http.StatusGone
		}

		c.JSON(status, models.ErrorResponse{
			Error:   "Shortlink not found",
			Message: err.Error(),
			Code:    status,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Shortlink found",
		Data:    shortlink,
	})
}

// AdminStats handles admin statistics
func (h *Handler) AdminStats(c *gin.Context) {
	// Check if user is admin (should be validated by middleware)
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
		Message: "Statistics retrieved successfully",
		Data:    stats,
	})
}
