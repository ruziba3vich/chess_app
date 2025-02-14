package storage

import (
	"context"
	"fmt"
	"log"

	redisservice "github.com/ruziba3vich/chess_app/internal/redis_service"
	"github.com/ruziba3vich/chess_app/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	DB struct {
		Client          *mongo.Client
		GamesCollection *mongo.Collection
	}
	Storage struct {
		database     *DB
		logger       *log.Logger
		redisService *redisservice.RedisStorage
	}
)

func NewStorage(database *DB, logger *log.Logger, redisService *redisservice.RedisStorage) *Storage {
	return &Storage{
		database:     database,
		logger:       logger,
		redisService: redisService,
	}
}

// ConnectDB establishes a connection to MongoDB
func ConnectDB(cfg *config.Config, ctx context.Context) (*DB, error) {
	clientOptions := options.Client().ApplyURI(cfg.DbConfig.MongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %s", err.Error())
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %s", err.Error())
	}

	return &DB{
		Client:          client,
		GamesCollection: client.Database(cfg.DbConfig.MongoDB).Collection(cfg.DbConfig.Collection),
	}, nil
}

// DisconnectDB to disconnect the db
func (db *DB) DisconnectDB(ctx context.Context) error {
	if err := db.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %s", err.Error())
	}
	return nil
}
