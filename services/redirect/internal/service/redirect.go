package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/database"
	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/pkg/models"
	"github.com/chthon/shortlink/pkg/utils"
)

// RedirectService interface defines redirect operations
type RedirectService interface {
	GetOriginalURL(code string) (string, error)
	LogClick(code, ip, userAgent, referer string) error
	GetStats() map[string]interface{}
}

// redirectService implements RedirectService
type redirectService struct {
	db          *gorm.DB
	redis       *redis.Client
	kafkaWriter *kafka.Writer
	config      *config.Config
}

// NewRedirectService creates a new redirect service
func NewRedirectService(db *database.DB, cfg *config.Config) RedirectService {
	// Initialize Kafka writer
	kafkaWriter := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Kafka.Brokers...),
		Topic:        cfg.Kafka.ClicksTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        true, // Async for better performance
	}

	return &redirectService{
		db:          db.PostgreSQL,
		redis:       db.Redis,
		kafkaWriter: kafkaWriter,
		config:      cfg,
	}
}

// GetOriginalURL retrieves the original URL for a short code
func (s *redirectService) GetOriginalURL(code string) (string, error) {
	ctx := context.Background()

	// Try Redis cache first
	cacheKey := fmt.Sprintf("shortlink:%s", code)
	cachedURL, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		logger.Debug("Cache hit for shortlink", "code", code)
		return cachedURL, nil
	}

	// Cache miss or error, query database
	logger.Debug("Cache miss for shortlink", "code", code, "redis_error", err)

	var shortlink models.ShortLink
	if err := s.db.Where("code = ? AND is_active = ?", code, true).First(&shortlink).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("shortlink not found")
		}
		logger.Error("Database error while fetching shortlink", "error", err, "code", code)
		return "", fmt.Errorf("database error: %w", err)
	}

	// Check if expired
	if shortlink.ExpiresAt != nil && time.Now().After(*shortlink.ExpiresAt) {
		return "", fmt.Errorf("shortlink has expired")
	}

	// Cache the result for future requests (TTL: 1 hour)
	if err := s.redis.Set(ctx, cacheKey, shortlink.OriginalURL, time.Hour).Err(); err != nil {
		logger.Warn("Failed to cache shortlink", "error", err, "code", code)
	}

	// Increment click count in database (async)
	go func() {
		if err := s.incrementClickCount(code); err != nil {
			logger.Error("Failed to increment click count", "error", err, "code", code)
		}
	}()

	logger.Info("Shortlink found", "code", code, "url", shortlink.OriginalURL)
	return shortlink.OriginalURL, nil
}

// LogClick logs a click event to Kafka for analytics
func (s *redirectService) LogClick(code, ip, userAgent, referer string) error {
	// Extract device, browser, OS from user agent
	device, browser, os := utils.ParseUserAgent(userAgent)

	// Create click event
	clickEvent := models.ClickEvent{
		ShortLinkCode: code,
		IPAddress:     ip,
		UserAgent:     userAgent,
		Referer:       referer,
		Device:        device,
		Browser:       browser,
		OS:            os,
		Timestamp:     time.Now().UTC(),
	}

	// Serialize to JSON
	clickData, err := json.Marshal(clickEvent)
	if err != nil {
		logger.Error("Failed to serialize click event", "error", err, "code", code)
		return fmt.Errorf("failed to serialize click event: %w", err)
	}

	// Send to Kafka
	message := kafka.Message{
		Key:   []byte(code),
		Value: clickData,
		Time:  time.Now(),
	}

	// Write to Kafka (async)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.kafkaWriter.WriteMessages(ctx, message); err != nil {
			logger.Error("Failed to send click event to Kafka", "error", err, "code", code)
		} else {
			logger.Debug("Click event sent to Kafka", "code", code, "ip", ip)
		}
	}()

	return nil
}

// incrementClickCount increments the click count in database
func (s *redirectService) incrementClickCount(code string) error {
	if err := s.db.Model(&models.ShortLink{}).
		Where("code = ?", code).
		UpdateColumn("click_count", gorm.Expr("click_count + ?", 1)).Error; err != nil {
		return fmt.Errorf("failed to increment click count: %w", err)
	}
	return nil
}

// GetStats returns statistics for admin dashboard
func (s *redirectService) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	ctx := context.Background()

	// Redis stats
	redisInfo, err := s.redis.Info(ctx, "stats").Result()
	if err != nil {
		logger.Error("Failed to get Redis stats", "error", err)
		stats["redis_error"] = err.Error()
	} else {
		// Parse Redis info
		lines := strings.Split(redisInfo, "\r\n")
		redisStats := make(map[string]string)
		for _, line := range lines {
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					redisStats[parts[0]] = parts[1]
				}
			}
		}
		stats["redis_stats"] = redisStats
	}

	// Database stats
	var totalRedirects int64
	s.db.Model(&models.ShortLink{}).Select("COALESCE(SUM(click_count), 0)").Scan(&totalRedirects)
	stats["total_redirects"] = totalRedirects

	// Recent redirects (last 24 hours)
	var recentRedirects int64
	s.db.Model(&models.ShortLink{}).
		Where("updated_at > ?", time.Now().Add(-24*time.Hour)).
		Select("COALESCE(SUM(click_count), 0)").
		Scan(&recentRedirects)
	stats["redirects_24h"] = recentRedirects

	// Top redirected links
	var topLinks []map[string]interface{}
	s.db.Model(&models.ShortLink{}).
		Select("code, original_url, click_count").
		Where("is_active = ? AND click_count > 0", true).
		Order("click_count DESC").
		Limit(10).
		Scan(&topLinks)
	stats["top_links"] = topLinks

	return stats
}

// Close closes the service connections
func (s *redirectService) Close() error {
	if s.kafkaWriter != nil {
		if err := s.kafkaWriter.Close(); err != nil {
			logger.Error("Failed to close Kafka writer", "error", err)
			return err
		}
	}
	return nil
}
