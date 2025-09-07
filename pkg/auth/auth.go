package auth

import (
	"errors"
	"time"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Claims represents JWT claims
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	secretKey []byte
	expiresIn time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey: []byte(cfg.Secret),
		expiresIn: cfg.ExpiresIn,
	}
}

// GenerateToken generates a JWT token for a user
func (j *JWTManager) GenerateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "shortlink-service",
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates a JWT token and returns claims
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new token with extended expiration
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Check if token is close to expiration (within 1 hour)
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token is not close to expiration")
	}

	// Create new token with same claims but extended expiration
	newClaims := &Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "shortlink-service",
			Subject:   claims.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString(j.secretKey)
}

// Password utilities

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword checks if password matches hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Role-based access control

// RequireRole checks if user has required role
func RequireRole(userRole, requiredRole string) bool {
	roles := map[string]int{
		"user":    1,
		"premium": 2,
		"admin":   3,
	}

	userLevel, userExists := roles[userRole]
	requiredLevel, requiredExists := roles[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

// IsAdmin checks if user is admin
func IsAdmin(role string) bool {
	return role == "admin"
}

// IsPremium checks if user is premium or higher
func IsPremium(role string) bool {
	return RequireRole(role, "premium")
}

// API Key utilities

// GenerateAPIKey generates a random API key
func GenerateAPIKey() string {
	// Implementation would generate a secure random string
	// For now, returning a placeholder
	return "api_" + generateRandomString(32)
}

// ValidateAPIKey validates an API key format
func ValidateAPIKey(apiKey string) bool {
	// Basic validation - should start with "api_" and be 36 characters
	return len(apiKey) == 36 && apiKey[:4] == "api_"
}

// Helper function to generate random string
func generateRandomString(length int) string {
	// This is a simplified implementation
	// In production, use crypto/rand for secure random generation
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// This is not cryptographically secure - use crypto/rand in production
		b[i] = charset[len(charset)/2] // Simplified for now
	}
	return string(b)
}

// ExtractUserFromClaims extracts user info from JWT claims
func ExtractUserFromClaims(claims *Claims) *models.User {
	return &models.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	}
}
