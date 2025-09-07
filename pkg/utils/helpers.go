package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ShortCodeGenerator generates short codes using various methods
type ShortCodeGenerator struct {
	base62Chars string
}

// NewShortCodeGenerator creates a new short code generator
func NewShortCodeGenerator() *ShortCodeGenerator {
	return &ShortCodeGenerator{
		base62Chars: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
	}
}

// GenerateRandomCode generates a random base62 code
func (g *ShortCodeGenerator) GenerateRandomCode(length int) (string, error) {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(g.base62Chars))))
		if err != nil {
			return "", err
		}
		result[i] = g.base62Chars[num.Int64()]
	}
	return string(result), nil
}

// GenerateHashCode generates a hash-based code
func (g *ShortCodeGenerator) GenerateHashCode(input string, length int) string {
	hash := md5.Sum([]byte(input + time.Now().String()))
	hashStr := hex.EncodeToString(hash[:])

	// Convert hex to base62-like chars
	result := ""
	for i := 0; i < length && i < len(hashStr); i++ {
		charIndex := int(hashStr[i]) % len(g.base62Chars)
		result += string(g.base62Chars[charIndex])
	}

	return result
}

// GenerateBase64Code generates a base64-based code
func (g *ShortCodeGenerator) GenerateBase64Code(input string, length int) string {
	hash := md5.Sum([]byte(input))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:])

	if len(encoded) > length {
		return encoded[:length]
	}
	return encoded
}

// GenerateTimestampCode generates a timestamp-based code
func (g *ShortCodeGenerator) GenerateTimestampCode(length int) string {
	timestamp := time.Now().UnixNano()
	encoded := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", timestamp)))

	if len(encoded) > length {
		return encoded[:length]
	}
	return encoded
}

// URL utilities

// IsValidURL validates if the URL is valid
func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

// NormalizeURL normalizes URL by adding scheme if missing
func NormalizeURL(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}

	// Add http:// if no scheme provided
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Normalize the URL
	parsedURL.Host = strings.ToLower(parsedURL.Host)

	return parsedURL.String()
}

// ExtractDomain extracts domain from URL
func ExtractDomain(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsedURL.Host
}

// String utilities

// IsAlphaNumeric checks if string contains only alphanumeric characters
func IsAlphaNumeric(s string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", s)
	return match
}

// SanitizeString sanitizes string by removing dangerous characters
func SanitizeString(s string) string {
	// Remove HTML tags and trim whitespace
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// User Agent parsing

// ParseUserAgent parses user agent string to extract device, browser, OS
func ParseUserAgent(userAgent string) (device, browser, os string) {
	ua := strings.ToLower(userAgent)

	// Device detection
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		device = "mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		device = "tablet"
	} else {
		device = "desktop"
	}

	// Browser detection
	if strings.Contains(ua, "chrome") {
		browser = "chrome"
	} else if strings.Contains(ua, "firefox") {
		browser = "firefox"
	} else if strings.Contains(ua, "safari") {
		browser = "safari"
	} else if strings.Contains(ua, "edge") {
		browser = "edge"
	} else if strings.Contains(ua, "opera") {
		browser = "opera"
	} else {
		browser = "other"
	}

	// OS detection
	if strings.Contains(ua, "windows") {
		os = "windows"
	} else if strings.Contains(ua, "macos") || strings.Contains(ua, "mac os") {
		os = "macos"
	} else if strings.Contains(ua, "linux") {
		os = "linux"
	} else if strings.Contains(ua, "android") {
		os = "android"
	} else if strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		os = "ios"
	} else {
		os = "other"
	}

	return device, browser, os
}

// ExtractIPFromRequest extracts IP address from request headers
func ExtractIPFromRequest(headers map[string]string) string {
	// Check common headers for real IP
	headerNames := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Client-IP",
		"CF-Connecting-IP", // Cloudflare
		"True-Client-IP",   // Akamai
	}

	for _, header := range headerNames {
		if ip, exists := headers[header]; exists && ip != "" {
			// X-Forwarded-For can contain multiple IPs, get the first one
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				if len(ips) > 0 {
					return strings.TrimSpace(ips[0])
				}
			}
			return ip
		}
	}

	return "unknown"
}

// Pagination utilities

// CalculateOffset calculates offset for pagination
func CalculateOffset(page, pageSize int) int {
	if page <= 0 {
		page = 1
	}
	return (page - 1) * pageSize
}

// CalculateTotalPages calculates total pages
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

// Time utilities

// FormatDuration formats duration in human readable format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.0fh", d.Hours())
	} else {
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
}

// GetTimeRanges returns common time ranges for analytics
func GetTimeRanges() map[string]time.Time {
	now := time.Now()
	return map[string]time.Time{
		"1h":  now.Add(-1 * time.Hour),
		"24h": now.Add(-24 * time.Hour),
		"7d":  now.Add(-7 * 24 * time.Hour),
		"30d": now.Add(-30 * 24 * time.Hour),
		"90d": now.Add(-90 * 24 * time.Hour),
	}
}

// HashString creates MD5 hash of string
func HashString(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// Validation utilities

// IsValidEmail validates email format
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPassword validates password strength
func IsValidPassword(password string) bool {
	if len(password) < 6 {
		return false
	}

	// Check for at least one number and one letter
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)

	return hasNumber && hasLetter
}
