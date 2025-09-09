package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// Server Configuration
	Server ServerConfig `json:"server"`

	// Database Configuration
	Database DatabaseConfig `json:"database"`

	// Redis Configuration
	Redis RedisConfig `json:"redis"`

	// MongoDB Configuration
	MongoDB MongoDBConfig `json:"mongodb"`

	// Kafka Configuration
	Kafka KafkaConfig `json:"kafka"`

	// JWT Configuration
	JWT JWTConfig `json:"jwt"`

	// Rate Limiting
	RateLimit RateLimitConfig `json:"rate_limit"`

	// Logging
	Logging LoggingConfig `json:"logging"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type MongoDBConfig struct {
	URI      string `json:"uri"`
	Database string `json:"database"`
}

type KafkaConfig struct {
	Brokers     []string `json:"brokers"`
	ClicksTopic string   `json:"clicks_topic"`
}

type JWTConfig struct {
	Secret    string        `json:"secret"`
	ExpiresIn time.Duration `json:"expires_in"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	Burst             int `json:"burst"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}
	return &Config{
		Server: ServerConfig{
			Host: getEnv("API_GATEWAY_HOST", "0.0.0.0"),
			Port: getEnvAsInt("API_GATEWAY_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvAsInt("POSTGRES_PORT", 5432),
			Database: getEnv("POSTGRES_DB", "shortlink"),
			Username: getEnv("POSTGRES_USER", "shortlink_user"),
			Password: getEnv("POSTGRES_PASSWORD", "shortlink_password"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "shortlink_analytics"),
		},
		Kafka: KafkaConfig{
			Brokers:     []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			ClicksTopic: getEnv("KAFKA_TOPIC_CLICKS", "shortlink_clicks"),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			ExpiresIn: getEnvAsDuration("JWT_EXPIRES_IN", 24*time.Hour),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
			Burst:             getEnvAsInt("RATE_LIMIT_BURST", 200),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// LoadServiceConfig loads configuration for a specific service
func LoadServiceConfig(serviceName string, defaultPort int) *Config {
	config := LoadConfig()

	// Override port based on service
	switch serviceName {
	case "shortlink":
		config.Server.Port = getEnvAsInt("SHORTLINK_SERVICE_PORT", defaultPort)
	case "redirect":
		config.Server.Port = getEnvAsInt("REDIRECT_SERVICE_PORT", defaultPort)
	case "analytics":
		config.Server.Port = getEnvAsInt("ANALYTICS_SERVICE_PORT", defaultPort)
	case "user-management":
		config.Server.Port = getEnvAsInt("USER_MANAGEMENT_SERVICE_PORT", defaultPort)
	}

	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.Secret == "" || c.JWT.Secret == "your-super-secret-jwt-key" {
		log.Println("WARNING: Using default JWT secret. Please change this in production!")
	}

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		log.Printf("WARNING: Invalid server port: %d", c.Server.Port)
	}

	return nil
}
