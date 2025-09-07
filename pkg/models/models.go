package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:'user'"` // user, premium, admin
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	ShortLinks []ShortLink `json:"short_links,omitempty" gorm:"foreignKey:UserID"`
}

// ShortLink represents a shortened link
type ShortLink struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Code        string     `json:"code" gorm:"uniqueIndex;not null;size:8"`
	OriginalURL string     `json:"original_url" gorm:"not null;type:text"`
	UserID      *uint      `json:"user_id" gorm:"index"` // nullable for anonymous links
	Title       string     `json:"title" gorm:"size:255"`
	Description string     `json:"description" gorm:"type:text"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	ClickCount  int64      `json:"click_count" gorm:"default:0"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`

	// Relationships
	User   *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Clicks []ClickEvent `json:"clicks,omitempty" gorm:"foreignKey:ShortLinkCode;references:Code"`
}

// ClickEvent represents a click event for analytics
type ClickEvent struct {
	ID            string    `json:"id" bson:"_id,omitempty"`
	ShortLinkCode string    `json:"short_link_code" bson:"short_link_code"`
	IPAddress     string    `json:"ip_address" bson:"ip_address"`
	UserAgent     string    `json:"user_agent" bson:"user_agent"`
	Referer       string    `json:"referer" bson:"referer"`
	Country       string    `json:"country" bson:"country"`
	City          string    `json:"city" bson:"city"`
	Device        string    `json:"device" bson:"device"`
	Browser       string    `json:"browser" bson:"browser"`
	OS            string    `json:"os" bson:"os"`
	Timestamp     time.Time `json:"timestamp" bson:"timestamp"`
}

// CreateShortLinkRequest represents a request to create a short link
type CreateShortLinkRequest struct {
	URL         string     `json:"url" binding:"required,url"`
	CustomCode  string     `json:"custom_code" binding:"omitempty,min=3,max=12,alphanum"`
	Title       string     `json:"title" binding:"omitempty,max=255"`
	Description string     `json:"description" binding:"omitempty,max=1000"`
	ExpiresAt   *time.Time `json:"expires_at" binding:"omitempty"`
}

// CreateShortLinkResponse represents a response when creating a short link
type CreateShortLinkResponse struct {
	ID          uint       `json:"id"`
	Code        string     `json:"code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	Title       string     `json:"title,omitempty"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// UpdateShortLinkRequest represents a request to update a short link
type UpdateShortLinkRequest struct {
	Title       *string    `json:"title" binding:"omitempty,max=255"`
	Description *string    `json:"description" binding:"omitempty,max=1000"`
	IsActive    *bool      `json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
	User      User   `json:"user"`
}

// AnalyticsResponse represents analytics data
type AnalyticsResponse struct {
	ShortLinkCode string          `json:"short_link_code"`
	TotalClicks   int64           `json:"total_clicks"`
	UniqueClicks  int64           `json:"unique_clicks"`
	ClicksByDate  []ClicksByDate  `json:"clicks_by_date"`
	TopCountries  []CountryStats  `json:"top_countries"`
	TopReferrers  []ReferrerStats `json:"top_referrers"`
	DeviceStats   DeviceStats     `json:"device_stats"`
	BrowserStats  []BrowserStats  `json:"browser_stats"`
}

// ClicksByDate represents clicks grouped by date
type ClicksByDate struct {
	Date   string `json:"date"`
	Clicks int64  `json:"clicks"`
}

// CountryStats represents clicks by country
type CountryStats struct {
	Country string `json:"country"`
	Clicks  int64  `json:"clicks"`
}

// ReferrerStats represents clicks by referrer
type ReferrerStats struct {
	Referrer string `json:"referrer"`
	Clicks   int64  `json:"clicks"`
}

// DeviceStats represents device statistics
type DeviceStats struct {
	Desktop int64 `json:"desktop"`
	Mobile  int64 `json:"mobile"`
	Tablet  int64 `json:"tablet"`
	Other   int64 `json:"other"`
}

// BrowserStats represents browser statistics
type BrowserStats struct {
	Browser string `json:"browser"`
	Clicks  int64  `json:"clicks"`
}

// HealthCheckResponse represents health check response
type HealthCheckResponse struct {
	Status    string          `json:"status"`
	Services  map[string]bool `json:"services"`
	Timestamp time.Time       `json:"timestamp"`
	Version   string          `json:"version"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int64       `json:"total"`
	TotalPages int         `json:"total_pages"`
}
