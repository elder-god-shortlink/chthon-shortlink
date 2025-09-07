package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/database"
	"github.com/chthon/shortlink/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AnalyticsService struct {
	db     *database.DB
	config *config.Config
}

type ClickEvent struct {
	ID             string                 `bson:"_id,omitempty" json:"id"`
	ShortCode      string                 `bson:"short_code" json:"short_code"`
	OriginalURL    string                 `bson:"original_url" json:"original_url"`
	UserAgent      string                 `bson:"user_agent" json:"user_agent"`
	IPAddress      string                 `bson:"ip_address" json:"ip_address"`
	Referer        string                 `bson:"referer" json:"referer"`
	Country        string                 `bson:"country" json:"country"`
	City           string                 `bson:"city" json:"city"`
	Device         string                 `bson:"device" json:"device"`
	Browser        string                 `bson:"browser" json:"browser"`
	OS             string                 `bson:"os" json:"os"`
	Timestamp      time.Time              `bson:"timestamp" json:"timestamp"`
	AdditionalData map[string]interface{} `bson:"additional_data,omitempty" json:"additional_data,omitempty"`
}

type AnalyticsStats struct {
	ShortCode    string     `json:"short_code"`
	TotalClicks  int64      `json:"total_clicks"`
	UniqueClicks int64      `json:"unique_clicks"`
	LastClicked  *time.Time `json:"last_clicked,omitempty"`
}

func NewAnalyticsService(db *database.DB, config *config.Config) *AnalyticsService {
	return &AnalyticsService{
		db:     db,
		config: config,
	}
}

func (s *AnalyticsService) ProcessClickEvent(event *ClickEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.MongoDB.Collection("click_events")

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	_, err := collection.InsertOne(ctx, event)
	if err != nil {
		logger.Error("Failed to insert click event", "error", err)
		return err
	}

	logger.Info("Click event processed", "short_code", event.ShortCode)
	return nil
}

func (s *AnalyticsService) GetAnalytics(shortCode string) (*AnalyticsStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.MongoDB.Collection("click_events")

	filter := bson.M{"short_code": shortCode}

	// Count total clicks
	totalClicks, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		logger.Error("Failed to count total clicks", "error", err)
		return nil, err
	}

	// Count unique clicks (by IP address)
	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id": "$ip_address",
		}},
		{"$count": "unique_count"},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		logger.Error("Failed to count unique clicks", "error", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var result struct {
		UniqueCount int64 `bson:"unique_count"`
	}
	var uniqueClicks int64 = 0
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err == nil {
			uniqueClicks = result.UniqueCount
		}
	}

	// Get last clicked time
	var lastEvent ClickEvent
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
	err = collection.FindOne(ctx, filter, opts).Decode(&lastEvent)

	stats := &AnalyticsStats{
		ShortCode:    shortCode,
		TotalClicks:  totalClicks,
		UniqueClicks: uniqueClicks,
	}

	if err == nil {
		stats.LastClicked = &lastEvent.Timestamp
	} else if err != mongo.ErrNoDocuments {
		logger.Error("Failed to get last clicked time", "error", err)
	}

	return stats, nil
}

func (s *AnalyticsService) ParseClickEventFromKafka(message []byte) (*ClickEvent, error) {
	var event ClickEvent
	if err := json.Unmarshal(message, &event); err != nil {
		logger.Error("Failed to parse click event", "error", err)
		return nil, err
	}

	return &event, nil
}
