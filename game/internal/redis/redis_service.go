package redis_service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ruziba3vich/chess_app/pkg/config"
)

type RedisService struct {
	client *redis.Client
	config *config.Config
}

func NewRedisService(client *redis.Client, config *config.Config) *RedisService {
	return &RedisService{
		client: client,
		config: config,
	}
}

func (r *RedisService) AddPlayerToQueue(playerID string, playerRank int32, duration int8) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("%s_%dmin", r.config.GameConfig.MatchMakingQueueName, duration)
	score := float64(playerRank)
	err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: playerID,
	}).Err()

	return err
}

func (r *RedisService) FindMatch(playerID string, playerRank int32, duration int32) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("%s_%dmin", r.config.GameConfig.MatchMakingQueueName, duration) // Match within same duration
	rankRange := int32(r.config.GameConfig.RankRange)
	minRank := float64(playerRank - rankRange)
	maxRank := float64(playerRank + rankRange)

	// Find closest match within the same duration queue
	matches, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:   fmt.Sprintf("%f", minRank),
		Max:   fmt.Sprintf("%f", maxRank),
		Count: 1,
	}).Result()
	if err != nil || len(matches) == 0 {
		return "", fmt.Errorf("no match found")
	}

	// Remove matched player from queue
	match := matches[0]

	if err := r.RemovePlayer(match, key, r.config.GameConfig.SearchDuration); err != nil {
		return "", err
	}

	return match, nil
}

func (r *RedisService) RemovePlayer(playerID string, key string, duration int8) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Minute)
	defer cancel()

	return r.client.ZRem(ctx, key, playerID).Err()
}
