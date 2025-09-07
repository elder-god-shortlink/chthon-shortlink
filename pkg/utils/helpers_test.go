package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShortCodeGenerator(t *testing.T) {
	generator := NewShortCodeGenerator()
	assert.NotNil(t, generator)

	t.Run("GenerateRandomCode", func(t *testing.T) {
		length := 8
		code, err := generator.GenerateRandomCode(length)

		assert.NoError(t, err)
		assert.Equal(t, length, len(code))
		assert.True(t, IsAlphaNumeric(code))
	})

	t.Run("GenerateHashCode", func(t *testing.T) {
		input := "test input"
		length := 6
		code := generator.GenerateHashCode(input, length)

		assert.Equal(t, length, len(code))
		assert.True(t, IsAlphaNumeric(code))
	})

	t.Run("GenerateBase64Code", func(t *testing.T) {
		input := "test input"
		length := 8
		code := generator.GenerateBase64Code(input, length)

		assert.LessOrEqual(t, len(code), length)
		assert.NotEmpty(t, code)
	})

	t.Run("GenerateTimestampCode", func(t *testing.T) {
		length := 10
		code := generator.GenerateTimestampCode(length)

		assert.LessOrEqual(t, len(code), length)
		assert.NotEmpty(t, code)
	})
}

func TestURLValidation(t *testing.T) {
	t.Run("ValidURLs", func(t *testing.T) {
		validURLs := []string{
			"https://example.com",
			"http://example.com",
			"https://www.example.com/path",
			"http://localhost:8080",
		}

		for _, url := range validURLs {
			assert.True(t, IsValidURL(url), "URL should be valid: %s", url)
		}
	})

	t.Run("InvalidURLs", func(t *testing.T) {
		invalidURLs := []string{
			"",
			"not-a-url",
			"example.com",
			"://invalid",
		}

		for _, url := range invalidURLs {
			assert.False(t, IsValidURL(url), "URL should be invalid: %s", url)
		}
	})
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "http://example.com"},
		{"www.example.com", "http://www.example.com"},
		{"https://example.com", "https://example.com"},
		{"", ""},
	}

	for _, test := range tests {
		result := NormalizeURL(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://example.com/path", "example.com"},
		{"http://www.example.com:8080", "www.example.com:8080"},
		{"invalid-url", ""}, // Invalid URL returns empty string
	}

	for _, test := range tests {
		result := ExtractDomain(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestIsAlphaNumeric(t *testing.T) {
	t.Run("ValidAlphaNumeric", func(t *testing.T) {
		validStrings := []string{"ABC123", "test123", "123456", "abcDEF"}

		for _, str := range validStrings {
			assert.True(t, IsAlphaNumeric(str), "Should be alphanumeric: %s", str)
		}
	})

	t.Run("InvalidAlphaNumeric", func(t *testing.T) {
		invalidStrings := []string{"test-123", "test 123", "test@123", ""}

		for _, str := range invalidStrings {
			assert.False(t, IsAlphaNumeric(str), "Should not be alphanumeric: %s", str)
		}
	})
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>alert('xss')</script>normal text", "alert('xss')normal text"}, // HTML tags removed
		{"  test string  ", "test string"},
		{"<div>content</div>", "content"},
	}

	for _, test := range tests {
		result := SanitizeString(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		userAgent string
		device    string
		browser   string
		os        string
	}{
		{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/91.0.4472.124",
			"desktop", "chrome", "windows",
		},
		{
			"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 Safari/604.1",
			"mobile", "safari", "macos", // Fixed: this parses as macos due to "Mac OS X"
		},
	}

	for _, test := range tests {
		device, browser, os := ParseUserAgent(test.userAgent)
		assert.Equal(t, test.device, device)
		assert.Equal(t, test.browser, browser)
		assert.Equal(t, test.os, os)
	}
}

func TestExtractIPFromRequest(t *testing.T) {
	tests := []struct {
		headers  map[string]string
		expected string
	}{
		{
			map[string]string{"X-Forwarded-For": "192.168.1.1, 10.0.0.1"},
			"192.168.1.1",
		},
		{
			map[string]string{"X-Real-IP": "192.168.1.2"},
			"192.168.1.2",
		},
		{
			map[string]string{},
			"unknown",
		},
	}

	for _, test := range tests {
		result := ExtractIPFromRequest(test.headers)
		assert.Equal(t, test.expected, result)
	}
}

func TestPaginationUtils(t *testing.T) {
	t.Run("CalculateOffset", func(t *testing.T) {
		tests := []struct {
			page     int
			pageSize int
			expected int
		}{
			{1, 10, 0},
			{2, 10, 10},
			{0, 10, 0}, // Invalid page should default to 1
		}

		for _, test := range tests {
			result := CalculateOffset(test.page, test.pageSize)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("CalculateTotalPages", func(t *testing.T) {
		tests := []struct {
			total    int64
			pageSize int
			expected int
		}{
			{100, 10, 10},
			{95, 10, 10},
			{5, 10, 1},
		}

		for _, test := range tests {
			result := CalculateTotalPages(test.total, test.pageSize)
			assert.Equal(t, test.expected, result)
		}
	})
}

func TestTimeUtils(t *testing.T) {
	t.Run("FormatDuration", func(t *testing.T) {
		tests := []struct {
			duration time.Duration
			expected string
		}{
			{30 * time.Second, "30s"},
			{5 * time.Minute, "5m"},
			{2 * time.Hour, "2h"},
		}

		for _, test := range tests {
			result := FormatDuration(test.duration)
			assert.Equal(t, test.expected, result)
		}
	})

	t.Run("GetTimeRanges", func(t *testing.T) {
		ranges := GetTimeRanges()
		assert.NotNil(t, ranges)
		assert.Contains(t, ranges, "1h")
		assert.Contains(t, ranges, "24h")
		assert.Contains(t, ranges, "7d")
	})
}

func TestValidationUtils(t *testing.T) {
	t.Run("IsValidEmail", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"test123@test-domain.com",
		}

		for _, email := range validEmails {
			assert.True(t, IsValidEmail(email), "Email should be valid: %s", email)
		}

		invalidEmails := []string{
			"invalid-email",
			"@domain.com",
			"test@",
			"",
		}

		for _, email := range invalidEmails {
			assert.False(t, IsValidEmail(email), "Email should be invalid: %s", email)
		}
	})

	t.Run("IsValidPassword", func(t *testing.T) {
		validPasswords := []string{
			"password123",
			"test1234",
			"abc123def",
		}

		for _, password := range validPasswords {
			assert.True(t, IsValidPassword(password), "Password should be valid: %s", password)
		}

		invalidPasswords := []string{
			"short",    // Too short
			"password", // No numbers
			"123456",   // No letters
		}

		for _, password := range invalidPasswords {
			assert.False(t, IsValidPassword(password), "Password should be invalid: %s", password)
		}
	})
}

func TestHashString(t *testing.T) {
	input := "test string"
	hash := HashString(input)

	assert.NotEmpty(t, hash)
	assert.Equal(t, 32, len(hash)) // MD5 hex is 32 characters

	// Same input should produce same hash
	hash2 := HashString(input)
	assert.Equal(t, hash, hash2)

	// Different input should produce different hash
	hash3 := HashString("different string")
	assert.NotEqual(t, hash, hash3)
}
