package redisservice

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/notnil/chess"
)

type RedisStorage struct {
	Pool *redis.Pool
}

func NewRedisStorage(pool *redis.Pool) *RedisStorage {
	return &RedisStorage{Pool: pool}
}

func (r *RedisStorage) SaveGame(gameID string, game *chess.Game) error {
	conn := r.Pool.Get()
	defer conn.Close()

	gameJSON, err := json.Marshal(game)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", "game:"+gameID, gameJSON)
	return err
}

func (r *RedisStorage) GetGame(gameID string) (*chess.Game, error) {
	conn := r.Pool.Get()
	defer conn.Close()

	gameJSON, err := redis.String(conn.Do("GET", "game:"+gameID))
	if err != nil {
		return nil, err
	}

	var game chess.Game
	if err := json.Unmarshal([]byte(gameJSON), &game); err != nil {
		return nil, err
	}

	return &game, nil
}
