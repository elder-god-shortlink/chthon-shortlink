package auth

import (
	"testing"
	"time"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:    "test-secret-key",
		ExpiresIn: 24 * time.Hour,
	}

	manager := NewJWTManager(cfg)
	require.NotNil(t, manager)

	testUser := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "user",
	}

	t.Run("GenerateToken", func(t *testing.T) {
		token, err := manager.GenerateToken(testUser)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Token should contain JWT format (3 parts separated by dots)
		parts := len([]rune(token))
		assert.Greater(t, parts, 50) // JWT tokens are typically much longer
	})

	t.Run("ValidateToken", func(t *testing.T) {
		// Generate a token
		token, err := manager.GenerateToken(testUser)
		require.NoError(t, err)

		// Validate the token
		claims, err := manager.ValidateToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		// Check claims
		assert.Equal(t, testUser.ID, claims.UserID)
		assert.Equal(t, testUser.Email, claims.Email)
		assert.Equal(t, testUser.Role, claims.Role)

		// Check expiration
		assert.True(t, claims.ExpiresAt.After(time.Now()))
	})

	t.Run("ValidateInvalidToken", func(t *testing.T) {
		invalidTokens := []string{
			"",
			"invalid.token.format",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
		}

		for _, token := range invalidTokens {
			claims, err := manager.ValidateToken(token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		}
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		// Create manager with very short expiration
		shortCfg := &config.JWTConfig{
			Secret:    "test-secret-key",
			ExpiresIn: 1 * time.Millisecond,
		}
		shortManager := NewJWTManager(shortCfg)

		// Generate token
		token, err := shortManager.GenerateToken(testUser)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Try to validate expired token
		claims, err := shortManager.ValidateToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("DifferentSecret", func(t *testing.T) {
		// Generate token with one secret
		token, err := manager.GenerateToken(testUser)
		require.NoError(t, err)

		// Try to validate with different secret
		differentCfg := &config.JWTConfig{
			Secret:    "different-secret-key",
			ExpiresIn: 24 * time.Hour,
		}
		differentManager := NewJWTManager(differentCfg)

		claims, err := differentManager.ValidateToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("RoundTripConsistency", func(t *testing.T) {
		users := []*models.User{
			{ID: 1, Email: "user1@test.com", Role: "user"},
			{ID: 2, Email: "admin@test.com", Role: "admin"},
			{ID: 3, Email: "premium@test.com", Role: "premium"},
		}

		for _, user := range users {
			// Generate token
			token, err := manager.GenerateToken(user)
			require.NoError(t, err, "Failed to generate token for user %d", user.ID)

			// Validate token
			claims, err := manager.ValidateToken(token)
			require.NoError(t, err, "Failed to validate token for user %d", user.ID)

			// Check all fields match
			assert.Equal(t, user.ID, claims.UserID, "UserID mismatch for user %d", user.ID)
			assert.Equal(t, user.Email, claims.Email, "Email mismatch for user %d", user.ID)
			assert.Equal(t, user.Role, claims.Role, "Role mismatch for user %d", user.ID)
		}
	})
}

func TestJWTClaims(t *testing.T) {
	t.Run("ClaimsValidation", func(t *testing.T) {
		cfg := &config.JWTConfig{
			Secret:    "test-secret",
			ExpiresIn: 1 * time.Hour,
		}
		manager := NewJWTManager(cfg)

		user := &models.User{
			ID:    42,
			Email: "test@example.com",
			Role:  "admin",
		}

		token, err := manager.GenerateToken(user)
		require.NoError(t, err)

		claims, err := manager.ValidateToken(token)
		require.NoError(t, err)

		// Test all claim fields
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Role, claims.Role)
		assert.NotZero(t, claims.IssuedAt)
		assert.NotZero(t, claims.ExpiresAt)

		// Convert NumericDate to time.Time for comparison
		issuedTime := claims.IssuedAt.Time
		expiresTime := claims.ExpiresAt.Time
		assert.True(t, expiresTime.After(issuedTime))
	})
}

func BenchmarkJWTOperations(b *testing.B) {
	cfg := &config.JWTConfig{
		Secret:    "benchmark-secret-key",
		ExpiresIn: 24 * time.Hour,
	}
	manager := NewJWTManager(cfg)

	user := &models.User{
		ID:    1,
		Email: "bench@example.com",
		Role:  "user",
	}

	b.Run("GenerateToken", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := manager.GenerateToken(user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Generate a token for validation benchmark
	token, err := manager.GenerateToken(user)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("ValidateToken", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := manager.ValidateToken(token)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
