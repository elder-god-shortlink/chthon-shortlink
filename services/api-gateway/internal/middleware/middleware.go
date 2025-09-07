package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chthon/shortlink/pkg/auth"
	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/chthon/shortlink/pkg/utils"
	"github.com/gin-gonic/gin"
)

// Logger middleware for request logging
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, Authorization, X-API-Key, X-Correlation-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Correlation-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestID middleware adds correlation ID to requests
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get correlation ID from header or generate new one
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			generator := utils.NewShortCodeGenerator()
			correlationID, _ = generator.GenerateRandomCode(16)
		}

		// Set correlation ID in response header
		c.Header("X-Correlation-ID", correlationID)

		// Store in context for logging
		c.Set("correlation_id", correlationID)

		c.Next()
	}
}

// Rate limiting implementation
type rateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

var limiter *rateLimiter

// RateLimit middleware for rate limiting
func RateLimit(cfg *config.Config) gin.HandlerFunc {
	if limiter == nil {
		limiter = &rateLimiter{
			requests: make(map[string][]time.Time),
			limit:    cfg.RateLimit.RequestsPerMinute,
			window:   time.Minute,
		}

		// Cleanup goroutine
		go limiter.cleanup()
	}

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !limiter.allow(key) {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error:   "Rate limit exceeded",
				Message: "Too many requests. Please try again later.",
				Code:    429,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allow checks if request is allowed under rate limit
func (rl *rateLimiter) allow(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Initialize if first request
	if _, exists := rl.requests[key]; !exists {
		rl.requests[key] = []time.Time{now}
		return true
	}

	// Remove old requests outside window
	requests := rl.requests[key]
	var validRequests []time.Time
	for _, reqTime := range requests {
		if now.Sub(reqTime) < rl.window {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if under limit
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

// cleanup removes old entries from rate limiter
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()

		for key, requests := range rl.requests {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < rl.window {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validRequests
			}
		}

		rl.mutex.Unlock()
	}
}

// JWTAuth middleware for JWT authentication
func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication token required",
				Code:    401,
			})
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid authentication token",
				Code:    401,
			})
			c.Abort()
			return
		}

		// Set user in context
		user := auth.ExtractUserFromClaims(claims)
		c.Set("user", user)
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// OptionalJWTAuth middleware for optional JWT authentication
func OptionalJWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token != "" {
			if claims, err := jwtManager.ValidateToken(token); err == nil {
				user := auth.ExtractUserFromClaims(claims)
				c.Set("user", user)
				c.Set("user_id", claims.UserID)
				c.Set("user_role", claims.Role)
			}
		}
		c.Next()
	}
}

// RequireRole middleware for role-based access control
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
				Code:    401,
			})
			c.Abort()
			return
		}

		if !auth.RequireRole(userRole.(string), requiredRole) {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "Insufficient permissions",
				Code:    403,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// APIKeyAuth middleware for API key authentication
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "API key required",
				Code:    401,
			})
			c.Abort()
			return
		}

		if !auth.ValidateAPIKey(apiKey) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Invalid API key",
				Code:    401,
			})
			c.Abort()
			return
		}

		// In production, you would validate the API key against a database
		// and set user context based on the API key
		c.Set("api_key", apiKey)
		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// Metrics middleware for Prometheus metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()

		// In production, you would record these metrics to Prometheus
		logger.Debug("Request metrics",
			"method", method,
			"path", path,
			"status", status,
			"duration", duration,
		)
	}
}

// Security headers middleware
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}
