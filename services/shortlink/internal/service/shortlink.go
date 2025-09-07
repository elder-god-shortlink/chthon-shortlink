package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/chthon/shortlink/pkg/utils"
	"github.com/chthon/shortlink/services/shortlink/internal/repository"
)

// ShortlinkService interface defines business logic operations
type ShortlinkService interface {
	CreateShortlink(req *models.CreateShortLinkRequest, userID *uint) (*models.CreateShortLinkResponse, error)
	GetShortlink(id uint, userID *uint) (*models.ShortLink, error)
	GetShortlinkByCode(code string) (*models.ShortLink, error)
	UpdateShortlink(id uint, req *models.UpdateShortLinkRequest, userID *uint) (*models.ShortLink, error)
	DeleteShortlink(id uint, userID *uint) error
	ListUserShortlinks(userID uint, page, pageSize int) (*models.PaginatedResponse, error)
	GetStats() map[string]interface{}
	ValidateOwnership(linkID uint, userID uint) error
}

// shortlinkService implements ShortlinkService
type shortlinkService struct {
	repo      *repository.Repository
	config    *config.Config
	generator *utils.ShortCodeGenerator
}

// NewShortlinkService creates a new shortlink service
func NewShortlinkService(repo *repository.Repository, cfg *config.Config) ShortlinkService {
	return &shortlinkService{
		repo:      repo,
		config:    cfg,
		generator: utils.NewShortCodeGenerator(),
	}
}

// CreateShortlink creates a new shortlink
func (s *shortlinkService) CreateShortlink(req *models.CreateShortLinkRequest, userID *uint) (*models.CreateShortLinkResponse, error) {
	// Validate URL
	if !utils.IsValidURL(req.URL) {
		return nil, errors.New("invalid URL format")
	}

	// Normalize URL
	normalizedURL := utils.NormalizeURL(req.URL)

	// Generate or validate custom code
	var code string
	var err error

	if req.CustomCode != "" {
		// Validate custom code
		if !utils.IsAlphaNumeric(req.CustomCode) {
			return nil, errors.New("custom code can only contain alphanumeric characters")
		}

		// Check if custom code already exists
		if s.repo.ShortLink.ExistsCode(req.CustomCode) {
			return nil, errors.New("custom code already exists")
		}

		code = req.CustomCode
	} else {
		// Generate unique code
		code, err = s.generateUniqueCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate unique code: %w", err)
		}
	}

	// Create shortlink model
	shortlink := &models.ShortLink{
		Code:        code,
		OriginalURL: normalizedURL,
		UserID:      userID,
		Title:       utils.SanitizeString(req.Title),
		Description: utils.SanitizeString(req.Description),
		IsActive:    true,
		ClickCount:  0,
		ExpiresAt:   req.ExpiresAt,
	}

	// Save to database
	if err := s.repo.ShortLink.Create(shortlink); err != nil {
		logger.Error("Failed to create shortlink", "error", err, "code", code)
		return nil, fmt.Errorf("failed to create shortlink: %w", err)
	}

	// Build response
	baseDomain := s.config.Server.Host
	if s.config.Server.Port != 80 && s.config.Server.Port != 443 {
		baseDomain = fmt.Sprintf("%s:%d", baseDomain, s.config.Server.Port)
	}

	response := &models.CreateShortLinkResponse{
		ID:          shortlink.ID,
		Code:        shortlink.Code,
		ShortURL:    fmt.Sprintf("http://%s/%s", baseDomain, shortlink.Code),
		OriginalURL: shortlink.OriginalURL,
		Title:       shortlink.Title,
		Description: shortlink.Description,
		CreatedAt:   shortlink.CreatedAt,
		ExpiresAt:   shortlink.ExpiresAt,
	}

	logger.Info("Shortlink created", "id", shortlink.ID, "code", code, "user_id", userID)
	return response, nil
}

// GetShortlink retrieves a shortlink by ID
func (s *shortlinkService) GetShortlink(id uint, userID *uint) (*models.ShortLink, error) {
	shortlink, err := s.repo.ShortLink.GetByID(id)
	if err != nil {
		return nil, err
	}

	if shortlink == nil {
		return nil, errors.New("shortlink not found")
	}

	// Check ownership for non-admin users
	if userID != nil && shortlink.UserID != nil && *shortlink.UserID != *userID {
		return nil, errors.New("access denied: not owner of this shortlink")
	}

	return shortlink, nil
}

// GetShortlinkByCode retrieves a shortlink by code
func (s *shortlinkService) GetShortlinkByCode(code string) (*models.ShortLink, error) {
	shortlink, err := s.repo.ShortLink.GetByCode(code)
	if err != nil {
		return nil, err
	}

	if shortlink == nil {
		return nil, errors.New("shortlink not found")
	}

	// Check if expired
	if shortlink.ExpiresAt != nil && time.Now().After(*shortlink.ExpiresAt) {
		return nil, errors.New("shortlink has expired")
	}

	return shortlink, nil
}

// UpdateShortlink updates a shortlink
func (s *shortlinkService) UpdateShortlink(id uint, req *models.UpdateShortLinkRequest, userID *uint) (*models.ShortLink, error) {
	// Get existing shortlink
	shortlink, err := s.GetShortlink(id, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Title != nil {
		shortlink.Title = utils.SanitizeString(*req.Title)
	}

	if req.Description != nil {
		shortlink.Description = utils.SanitizeString(*req.Description)
	}

	if req.IsActive != nil {
		shortlink.IsActive = *req.IsActive
	}

	if req.ExpiresAt != nil {
		shortlink.ExpiresAt = req.ExpiresAt
	}

	// Save changes
	if err := s.repo.ShortLink.Update(shortlink); err != nil {
		logger.Error("Failed to update shortlink", "error", err, "id", id)
		return nil, fmt.Errorf("failed to update shortlink: %w", err)
	}

	logger.Info("Shortlink updated", "id", id, "user_id", userID)
	return shortlink, nil
}

// DeleteShortlink deletes a shortlink
func (s *shortlinkService) DeleteShortlink(id uint, userID *uint) error {
	// Verify ownership
	if err := s.ValidateOwnership(id, *userID); err != nil {
		return err
	}

	// Soft delete
	if err := s.repo.ShortLink.Delete(id); err != nil {
		logger.Error("Failed to delete shortlink", "error", err, "id", id)
		return fmt.Errorf("failed to delete shortlink: %w", err)
	}

	logger.Info("Shortlink deleted", "id", id, "user_id", userID)
	return nil
}

// ListUserShortlinks lists shortlinks for a user with pagination
func (s *shortlinkService) ListUserShortlinks(userID uint, page, pageSize int) (*models.PaginatedResponse, error) {
	offset := utils.CalculateOffset(page, pageSize)

	shortlinks, total, err := s.repo.ShortLink.GetByUserID(userID, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list user shortlinks: %w", err)
	}

	totalPages := utils.CalculateTotalPages(total, pageSize)

	return &models.PaginatedResponse{
		Data:       shortlinks,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// GetStats returns statistics for admin dashboard
func (s *shortlinkService) GetStats() map[string]interface{} {
	return s.repo.ShortLink.GetStats()
}

// ValidateOwnership checks if user owns the shortlink
func (s *shortlinkService) ValidateOwnership(linkID uint, userID uint) error {
	shortlink, err := s.repo.ShortLink.GetByID(linkID)
	if err != nil {
		return err
	}

	if shortlink == nil {
		return errors.New("shortlink not found")
	}

	if shortlink.UserID == nil || *shortlink.UserID != userID {
		return errors.New("access denied: not owner of this shortlink")
	}

	return nil
}

// generateUniqueCode generates a unique short code
func (s *shortlinkService) generateUniqueCode() (string, error) {
	maxAttempts := 10
	codeLength := 6

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Try different generation methods
		var code string
		var err error

		switch attempt % 3 {
		case 0:
			// Random base62 code
			code, err = s.generator.GenerateRandomCode(codeLength)
		case 1:
			// Hash-based code
			code = s.generator.GenerateHashCode(time.Now().String(), codeLength)
		case 2:
			// Base64 code
			code = s.generator.GenerateBase64Code(time.Now().String(), codeLength)
		}

		if err != nil {
			logger.Error("Failed to generate code", "error", err, "attempt", attempt)
			continue
		}

		// Ensure code is alphanumeric and not reserved
		code = strings.ToLower(code)
		if s.isReservedCode(code) {
			continue
		}

		// Check if code already exists
		if !s.repo.ShortLink.ExistsCode(code) {
			return code, nil
		}

		// Increase code length for next attempts
		if attempt >= 5 {
			codeLength = 8
		}
	}

	return "", errors.New("failed to generate unique code after maximum attempts")
}

// isReservedCode checks if a code is reserved
func (s *shortlinkService) isReservedCode(code string) bool {
	reserved := []string{
		"api", "www", "admin", "app", "help", "docs", "blog",
		"health", "status", "login", "logout", "register",
		"dashboard", "analytics", "settings", "profile",
	}

	for _, r := range reserved {
		if code == r {
			return true
		}
	}

	return false
}
