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

	key := fmt.Sprintf("matchmaking_queue:%d", duration)

	// Find players within the rank range
	results, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", playerRank-200),
		Max: fmt.Sprintf("%d", playerRank+200),
	}).Result()

	if err != nil || len(results) == 0 {
		return "", fmt.Errorf("no suitable opponent found")
	}

	for _, result := range results {
		opponentID := result
		if opponentID != playerID {
			if err := r.RemovePlayer(opponentID, key, int8(duration)); err != nil {
				return "", err
			}
			return opponentID, nil
		}
	}

	return "", fmt.Errorf("no suitable opponent found")
}

func (r *RedisService) RemovePlayer(playerID string, key string, duration int8) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Minute)
	defer cancel()

	return r.client.ZRem(ctx, key, playerID).Err()
}
