package consumer

import (
	"context"
	"time"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/chthon/shortlink/pkg/logger"
	"github.com/chthon/shortlink/services/analytics/internal/service"
	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader   *kafka.Reader
	service  *service.AnalyticsService
	config   *config.Config
	stopChan chan struct{}
}

func NewKafkaConsumer(analyticsService *service.AnalyticsService, config *config.Config) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.Kafka.Brokers,
		Topic:       config.Kafka.ClicksTopic,
		GroupID:     "analytics-service",
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		StartOffset: kafka.LastOffset,
	})

	return &KafkaConsumer{
		reader:   reader,
		service:  analyticsService,
		config:   config,
		stopChan: make(chan struct{}),
	}
}

func (c *KafkaConsumer) Start() error {
	logger.Info("Starting Kafka consumer")

	for {
		select {
		case <-c.stopChan:
			logger.Info("Kafka consumer stopped")
			return nil
		default:
			// Set a timeout for reading messages
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			message, err := c.reader.ReadMessage(ctx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded {
					// Timeout is normal, continue
					continue
				}
				logger.Error("Failed to read message from Kafka", "error", err)
				continue
			}

			// Process the message
			if err := c.processMessage(message); err != nil {
				logger.Error("Failed to process message", "error", err)
			}
		}
	}
}

func (c *KafkaConsumer) Stop() {
	logger.Info("Stopping Kafka consumer")
	close(c.stopChan)
	if c.reader != nil {
		c.reader.Close()
	}
}

func (c *KafkaConsumer) processMessage(message kafka.Message) error {
	logger.Debug("Processing Kafka message", "partition", message.Partition, "offset", message.Offset)

	// Parse click event from message
	clickEvent, err := c.service.ParseClickEventFromKafka(message.Value)
	if err != nil {
		return err
	}

	// Process the click event
	if err := c.service.ProcessClickEvent(clickEvent); err != nil {
		return err
	}

	logger.Debug("Successfully processed click event", "short_code", clickEvent.ShortCode)
	return nil
}

// Health check for the consumer
func (c *KafkaConsumer) IsHealthy() bool {
	// Simple health check - could be enhanced with more sophisticated checks
	return c.reader != nil
}
