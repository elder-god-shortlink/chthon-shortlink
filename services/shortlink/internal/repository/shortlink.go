package repository

import (
	"errors"
	"fmt"

	"github.com/chthon/shortlink/pkg/database"
	"github.com/chthon/shortlink/pkg/models"
	"gorm.io/gorm"
)

// ShortlinkRepository interface defines shortlink data operations
type ShortlinkRepository interface {
	Create(shortlink *models.ShortLink) error
	GetByCode(code string) (*models.ShortLink, error)
	GetByID(id uint) (*models.ShortLink, error)
	GetByUserID(userID uint, offset, limit int) ([]*models.ShortLink, int64, error)
	Update(shortlink *models.ShortLink) error
	Delete(id uint) error
	ExistsCode(code string) bool
	IncrementClickCount(code string) error
	GetStats() map[string]interface{}
}

// Repository contains all repositories
type Repository struct {
	ShortLink ShortlinkRepository
	db        *database.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *database.DB) *Repository {
	return &Repository{
		ShortLink: &shortlinkRepository{db: db.PostgreSQL},
		db:        db,
	}
}

// shortlinkRepository implements ShortlinkRepository
type shortlinkRepository struct {
	db *gorm.DB
}

// Create creates a new shortlink
func (r *shortlinkRepository) Create(shortlink *models.ShortLink) error {
	if err := r.db.Create(shortlink).Error; err != nil {
		return fmt.Errorf("failed to create shortlink: %w", err)
	}
	return nil
}

// GetByCode retrieves a shortlink by its code
func (r *shortlinkRepository) GetByCode(code string) (*models.ShortLink, error) {
	var shortlink models.ShortLink
	if err := r.db.Where("code = ? AND is_active = ?", code, true).First(&shortlink).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil for not found
		}
		return nil, fmt.Errorf("failed to get shortlink by code: %w", err)
	}
	return &shortlink, nil
}

// GetByID retrieves a shortlink by its ID
func (r *shortlinkRepository) GetByID(id uint) (*models.ShortLink, error) {
	var shortlink models.ShortLink
	if err := r.db.First(&shortlink, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get shortlink by ID: %w", err)
	}
	return &shortlink, nil
}

// GetByUserID retrieves shortlinks for a specific user with pagination
func (r *shortlinkRepository) GetByUserID(userID uint, offset, limit int) ([]*models.ShortLink, int64, error) {
	var shortlinks []*models.ShortLink
	var total int64

	// Count total records
	if err := r.db.Model(&models.ShortLink{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user shortlinks: %w", err)
	}

	// Get paginated records
	if err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&shortlinks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user shortlinks: %w", err)
	}

	return shortlinks, total, nil
}

// Update updates a shortlink
func (r *shortlinkRepository) Update(shortlink *models.ShortLink) error {
	if err := r.db.Save(shortlink).Error; err != nil {
		return fmt.Errorf("failed to update shortlink: %w", err)
	}
	return nil
}

// Delete deletes a shortlink (soft delete by setting is_active = false)
func (r *shortlinkRepository) Delete(id uint) error {
	if err := r.db.Model(&models.ShortLink{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to delete shortlink: %w", err)
	}
	return nil
}

// ExistsCode checks if a code already exists
func (r *shortlinkRepository) ExistsCode(code string) bool {
	var count int64
	r.db.Model(&models.ShortLink{}).Where("code = ?", code).Count(&count)
	return count > 0
}

// IncrementClickCount increments the click count for a shortlink
func (r *shortlinkRepository) IncrementClickCount(code string) error {
	if err := r.db.Model(&models.ShortLink{}).
		Where("code = ?", code).
		UpdateColumn("click_count", gorm.Expr("click_count + ?", 1)).Error; err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}
	return nil
}

// GetStats returns statistics for admin dashboard
func (r *shortlinkRepository) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Total links
	var totalLinks int64
	r.db.Model(&models.ShortLink{}).Count(&totalLinks)
	stats["total_links"] = totalLinks

	// Active links
	var activeLinks int64
	r.db.Model(&models.ShortLink{}).Where("is_active = ?", true).Count(&activeLinks)
	stats["active_links"] = activeLinks

	// Total clicks
	var totalClicks int64
	r.db.Model(&models.ShortLink{}).Select("COALESCE(SUM(click_count), 0)").Scan(&totalClicks)
	stats["total_clicks"] = totalClicks

	// Links created today
	var linksToday int64
	r.db.Model(&models.ShortLink{}).Where("DATE(created_at) = CURRENT_DATE").Count(&linksToday)
	stats["links_today"] = linksToday

	// Top 10 most clicked links
	var topLinks []map[string]interface{}
	r.db.Model(&models.ShortLink{}).
		Select("code, original_url, click_count").
		Where("is_active = ?", true).
		Order("click_count DESC").
		Limit(10).
		Scan(&topLinks)
	stats["top_links"] = topLinks

	return stats
}

// AutoMigrate runs auto migration for the database schema
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.ShortLink{},
	)
}
