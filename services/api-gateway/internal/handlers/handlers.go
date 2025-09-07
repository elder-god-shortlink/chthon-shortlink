package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chthon/shortlink/pkg/auth"
	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/gin-gonic/gin"
)

// Handler contains all the handlers for the API Gateway
type Handler struct {
	config     *config.Config
	jwtManager *auth.JWTManager
	httpClient *http.Client
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		config:     cfg,
		jwtManager: auth.NewJWTManager(&cfg.JWT),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetJWTManager returns the JWT manager instance
func (h *Handler) GetJWTManager() *auth.JWTManager {
	return h.jwtManager
}

// Health check handler
func (h *Handler) HealthCheck(c *gin.Context) {
	response := models.HealthCheckResponse{
		Status:    "ok",
		Services:  make(map[string]bool),
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	// Check service health (simplified)
	response.Services["api-gateway"] = true
	response.Services["shortlink-service"] = h.checkServiceHealth("http://shortlink-service:8082/health")
	response.Services["redirect-service"] = h.checkServiceHealth("http://redirect-service:8083/health")
	response.Services["analytics-service"] = h.checkServiceHealth("http://analytics-service:8084/health")
	response.Services["user-management-service"] = h.checkServiceHealth("http://user-management-service:8085/health")

	c.JSON(http.StatusOK, response)
}

// Proxy handlers for different services

// ShortlinkProxy proxies requests to shortlink service
func (h *Handler) ShortlinkProxy(c *gin.Context) {
	h.proxyRequest(c, "http://shortlink-service:8082")
}

// RedirectProxy proxies requests to redirect service
func (h *Handler) RedirectProxy(c *gin.Context) {
	h.proxyRequest(c, "http://redirect-service:8083")
}

// AnalyticsProxy proxies requests to analytics service
func (h *Handler) AnalyticsProxy(c *gin.Context) {
	h.proxyRequest(c, "http://analytics-service:8084")
}

// UserManagementProxy proxies requests to user management service
func (h *Handler) UserManagementProxy(c *gin.Context) {
	h.proxyRequest(c, "http://user-management-service:8085")
}

// Authentication handlers

// Login handles user authentication
func (h *Handler) Login(c *gin.Context) {
	// Forward to user management service
	h.proxyRequest(c, "http://user-management-service:8085")
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	// Forward to user management service
	h.proxyRequest(c, "http://user-management-service:8085")
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	// Forward to user management service
	h.proxyRequest(c, "http://user-management-service:8085")
}

// Admin handlers

// AdminStats provides admin dashboard statistics
func (h *Handler) AdminStats(c *gin.Context) {
	// Get user from context (added by auth middleware)
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "User not found in context",
			Code:    401,
		})
		return
	}

	userModel := user.(*models.User)
	if !auth.IsAdmin(userModel.Role) {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "Forbidden",
			Message: "Admin access required",
			Code:    403,
		})
		return
	}

	// Collect stats from all services
	stats := map[string]interface{}{
		"total_users":    h.getServiceData("http://user-management-service:8085/admin/stats"),
		"total_links":    h.getServiceData("http://shortlink-service:8082/admin/stats"),
		"total_clicks":   h.getServiceData("http://analytics-service:8084/admin/stats"),
		"service_health": h.getAllServiceHealth(),
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Admin statistics retrieved successfully",
		Data:    stats,
	})
}

// Helper methods

// proxyRequest forwards request to target service
func (h *Handler) proxyRequest(c *gin.Context, targetURL string) {
	// Build target URL
	targetURL = strings.TrimSuffix(targetURL, "/") + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Read request body
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Create new request
	req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		logger.Error("Failed to create proxy request", "error", err, "target", targetURL)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create proxy request",
			Code:    500,
		})
		return
	}

	// Copy headers
	for name, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Add correlation ID if exists
	if correlationID := c.GetHeader("X-Correlation-ID"); correlationID != "" {
		req.Header.Set("X-Correlation-ID", correlationID)
	}

	// Make request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		logger.Error("Proxy request failed", "error", err, "target", targetURL)
		c.JSON(http.StatusBadGateway, models.ErrorResponse{
			Error:   "Bad Gateway",
			Message: "Service unavailable",
			Code:    502,
		})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			c.Header(name, value)
		}
	}

	// Copy response
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

// checkServiceHealth checks if a service is healthy
func (h *Handler) checkServiceHealth(url string) bool {
	resp, err := h.httpClient.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// getServiceData gets data from a service endpoint
func (h *Handler) getServiceData(url string) interface{} {
	resp, err := h.httpClient.Get(url)
	if err != nil {
		logger.Error("Failed to get service data", "error", err, "url", url)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.Error("Failed to decode service response", "error", err, "url", url)
		return nil
	}

	return data
}

// getAllServiceHealth gets health status of all services
func (h *Handler) getAllServiceHealth() map[string]bool {
	return map[string]bool{
		"shortlink-service":       h.checkServiceHealth("http://shortlink-service:8082/health"),
		"redirect-service":        h.checkServiceHealth("http://redirect-service:8083/health"),
		"analytics-service":       h.checkServiceHealth("http://analytics-service:8084/health"),
		"user-management-service": h.checkServiceHealth("http://user-management-service:8085/health"),
	}
}
