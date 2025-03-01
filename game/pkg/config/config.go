package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// DbConfig holds the database configuration
type (
	DbConfig struct {
		MongoURI   string
		MongoDB    string
		Collection string
	}

	// Config holds the application configuration
	Config struct {
		DbConfig     *DbConfig
		GameConfig   *GameConfig
		Port         string
		Protocol     string
		RedisURI     string
		KafkaBrokers string // Kafka brokers (comma-separated)
		KafkaTopic   string // Kafka topic for move events
	}

	// GameConfig keeps the game configuration elements
	GameConfig struct {
		ScoreQueue     string // queue name in redis for users to be grouped in
		RankRange      int8   // players will be choosen in this range of score
		SearchDuration int8   // game is gonna be in search for opponent for this many minutes
		RedisChannel   string
		WorkerPoolSize int8
	}
)

// LoadConfig reads configuration from environment variables or .env file
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables if set.")
	}

	searchDurationStr := getEnv("SEARCH_DURATION", "1")
	searchDurationInt, _ := strconv.Atoi(searchDurationStr)
	workerPoolSizeStr, _ := strconv.Atoi(getEnv("WORKER_POOL_SIZE", "5"))
	workerPoolSizeInt8 := int8(workerPoolSizeStr)

	return &Config{
		DbConfig: &DbConfig{
			MongoURI:   getEnv("MONGO_URI", "mongodb://localhost:27017"),
			MongoDB:    getEnv("MONGO_DB", "test"),
			Collection: getEnv("MONGO_COLLECTION", "users"),
		},
		GameConfig: &GameConfig{
			ScoreQueue:     getEnv("MATCH_MAKING_QUEUE_NAME", "match_making_queue_name"),
			SearchDuration: int8(searchDurationInt),
			RedisChannel:   getEnv("REDIS_CHANNEL", "redis_channel"),
			WorkerPoolSize: workerPoolSizeInt8,
		},
		Port:         getEnv("PORT", "8080"),
		Protocol:     getEnv("PROTOCOL", "tcp"),
		RedisURI:     getEnv("REDIS_URI", "redis:6379"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "chess-moves"),
	}, nil
}

// Helper function to get environment variables with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// Getters for private fields
func (c *Config) GetKafkaBrokers() string {
	return c.KafkaBrokers
}

func (c *Config) GetKafkaTopic() string {
	return c.KafkaTopic
}
