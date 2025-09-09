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

// HealthCheck returns the health status of all services
//
//	@Summary		Check system health
//	@Description	Returns the health status of API Gateway and all microservices
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	object{status=string,services=object,timestamp=string,version=string}	"Health status of all services"
//	@Router			/health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	response := models.HealthCheckResponse{
		Status:    "ok",
		Services:  make(map[string]bool),
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	// Check service health (simplified)
	response.Services["api-gateway"] = true
	response.Services["shortlink-service"] = h.checkServiceHealth("http://localhost:8081/health")
	response.Services["redirect-service"] = h.checkServiceHealth("http://localhost:8082/health")
	response.Services["analytics-service"] = h.checkServiceHealth("http://localhost:8083/health")
	response.Services["user-management-service"] = h.checkServiceHealth("http://localhost:8084/health")

	c.JSON(http.StatusOK, response)
}

// Proxy handlers for different services

// ShortlinkProxy proxies requests to shortlink service
// @Summary Shortlink service operations
// @Description Handle various shortlink operations including creation, retrieval, update, and deletion of short URLs
// @Tags Shortlinks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/user/links [post]
// @Router /api/v1/user/links [get]
// @Router /api/v1/user/links/{id} [get]
// @Router /api/v1/user/links/{id} [put]
// @Router /api/v1/user/links/{id} [delete]
func (h *Handler) ShortlinkProxy(c *gin.Context) {
	h.proxyRequest(c, "http://localhost:8081")
}

// RedirectProxy proxies requests to redirect service
// @Summary URL redirection service
// @Description Handle URL redirection and click tracking for short URLs
// @Tags Redirect
// @Accept json
// @Produce json
// @Router /api/v1/public/{shortCode} [get]
// @Param shortCode path string true "Short URL code for redirection"
// @Success 302 {string} string "Redirect to original URL"
// @Failure 404 {object} map[string]interface{} "Short URL not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
func (h *Handler) RedirectProxy(c *gin.Context) {
	h.proxyRequest(c, "http://localhost:8082")
}

// AnalyticsProxy proxies requests to analytics service
// @Summary Analytics and reporting service
// @Description Handle analytics data collection, processing, and reporting for short URLs
// @Tags Analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/user/analytics [get]
// @Router /api/v1/premium/analytics/detailed [get]
// @Router /api/v1/admin/analytics [get]
// @Success 200 {object} map[string]interface{} "Analytics data and reports"
// @Failure 401 {object} map[string]interface{} "Unauthorized access"
// @Failure 500 {object} map[string]interface{} "Internal server error"
func (h *Handler) AnalyticsProxy(c *gin.Context) {
	h.proxyRequest(c, "http://localhost:8083")
}

// UserManagementProxy proxies requests to user management service
// @Summary User management service
// @Description Handle user account operations including profile management, settings, and user data
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/user/profile [get]
// @Router /api/v1/user/profile [put]
// @Router /api/v1/admin/users [get]
// @Router /api/v1/admin/users/{id} [get]
// @Router /api/v1/admin/users/{id} [put]
// @Router /api/v1/admin/users/{id} [delete]
// @Success 200 {object} map[string]interface{} "User data and operations"
// @Failure 401 {object} map[string]interface{} "Unauthorized access"
// @Failure 403 {object} map[string]interface{} "Forbidden - insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Internal server error"
func (h *Handler) UserManagementProxy(c *gin.Context) {
	h.proxyRequest(c, "http://localhost:8084")
}

// Authentication handlers

// Login handles user authentication
// @Summary User login authentication
// @Description Authenticate user with email and password to receive JWT access token
// @Tags Authentication
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param email formData string true "User email address"
// @Param password formData string true "User password"
// @Success 200 {object} map[string]interface{} "Login successful with access_token and refresh_token"
// @Failure 400 {object} map[string]interface{} "Missing email or password"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	// Forward to user management service
	h.proxyRequest(c, "http://localhost:8084")
}

// Register handles user registration
// @Summary Create new user account
// @Description Register a new user account with email, password and role
// @Tags Authentication
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param email formData string true "User email address"
// @Param password formData string true "User password"
// @Param role formData string false "User role (default: user)"
// @Success 201 {object} map[string]interface{} "Registration successful with user details"
// @Failure 400 {object} map[string]interface{} "Missing required fields"
// @Failure 409 {object} map[string]interface{} "User already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	// Forward to user management service
	h.proxyRequest(c, "http://localhost:8084")
}

// RefreshToken handles JWT token refresh
// @Summary Refresh JWT access token
// @Description Refresh an expired JWT access token using a valid refresh token to maintain authenticated session
// @Tags Authentication
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param refresh_token formData string true "Valid refresh token for generating new access token"
// @Success 200 {object} map[string]interface{} "Successfully refreshed token with new access_token and refresh_token"
// @Failure 400 {object} map[string]interface{} "Missing refresh token in request"
// @Failure 401 {object} map[string]interface{} "Invalid or expired refresh token"
// @Failure 500 {object} map[string]interface{} "Internal server error during token refresh"
// @Router /api/v1/auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	// Forward to user management service
	h.proxyRequest(c, "http://localhost:8084")
}

// Admin handlers

// AdminStats provides admin dashboard statistics
// @Summary Get admin dashboard statistics
// @Description Retrieve comprehensive statistics for admin dashboard including user counts, link metrics, and system health
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Admin statistics including user counts, link metrics, and system status"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing authentication token"
// @Failure 403 {object} map[string]interface{} "Forbidden - user does not have admin privileges"
// @Failure 500 {object} map[string]interface{} "Internal server error while fetching statistics"
// @Router /api/v1/admin/stats [get]
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
		logger.Error("Health check failed", "url", url, "error", err)
		return false
	}
	defer resp.Body.Close()

	success := resp.StatusCode == http.StatusOK
	logger.Info("Health check result", "url", url, "status", resp.StatusCode, "success", success)
	return success
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
