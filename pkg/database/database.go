package database

import (
	"context"
	"fmt"
	"time"

	"github.com/chthon/shortlink/pkg/config"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB holds all database connections
type DB struct {
	PostgreSQL *gorm.DB
	Redis      *redis.Client
	MongoDB    *mongo.Database
}

// NewDatabaseConnections creates new database connections
func NewDatabaseConnections(cfg *config.Config) (*DB, error) {
	db := &DB{}

	// PostgreSQL connection
	if pg, err := NewPostgreSQLConnection(&cfg.Database); err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	} else {
		db.PostgreSQL = pg
	}

	// Redis connection
	if rdb, err := NewRedisConnection(&cfg.Redis); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	} else {
		db.Redis = rdb
	}

	// MongoDB connection
	if mdb, err := NewMongoDBConnection(&cfg.MongoDB); err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	} else {
		db.MongoDB = mdb
	}

	return db, nil
}

// NewPostgreSQLConnection creates a new PostgreSQL connection
func NewPostgreSQLConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.Username, cfg.Password, cfg.Database, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// NewRedisConnection creates a new Redis connection
func NewRedisConnection(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

// NewMongoDBConnection creates a new MongoDB connection
func NewMongoDBConnection(cfg *config.MongoDBConfig) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	// Test connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client.Database(cfg.Database), nil
}

// Close closes all database connections
func (db *DB) Close() error {
	// Close PostgreSQL
	if db.PostgreSQL != nil {
		if sqlDB, err := db.PostgreSQL.DB(); err == nil {
			sqlDB.Close()
		}
	}

	// Close Redis
	if db.Redis != nil {
		db.Redis.Close()
	}

	// Close MongoDB
	if db.MongoDB != nil {
		if client := db.MongoDB.Client(); client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client.Disconnect(ctx)
		}
	}

	return nil
}

// Health check functions
func (db *DB) HealthCheck() map[string]bool {
	health := make(map[string]bool)

	// PostgreSQL health check
	if db.PostgreSQL != nil {
		if sqlDB, err := db.PostgreSQL.DB(); err == nil {
			health["postgresql"] = sqlDB.Ping() == nil
		} else {
			health["postgresql"] = false
		}
	}

	// Redis health check
	if db.Redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := db.Redis.Ping(ctx).Result()
		health["redis"] = err == nil
	}

	// MongoDB health check
	if db.MongoDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := db.MongoDB.Client().Ping(ctx, nil)
		health["mongodb"] = err == nil
	}

	return health
}
