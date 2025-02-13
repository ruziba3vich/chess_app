package game_service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ruziba3vich/chess_app/internal/storage"
	"github.com/ruziba3vich/chess_app/pkg/config"
)

type MatchmakingService struct {
	redisClient    *redis.Client
	playerChannels map[string]chan string
	mutex          sync.Mutex
	wg             *sync.WaitGroup
	config         *config.Config
	storage        *storage.Storage
	logger         *log.Logger
	luaScript      string
}

func NewMatchmakingService(
	redisClient *redis.Client,
	playerChannels map[string]chan string,
	config *config.Config,
	storage *storage.Storage,
	wg *sync.WaitGroup,
	logger *log.Logger,
	luaScript string,
) *MatchmakingService {
	return &MatchmakingService{
		redisClient:    redisClient,
		playerChannels: playerChannels,
		config:         config,
		storage:        storage,
		wg:             wg,
		logger:         logger,
		luaScript:      luaScript,
	}
}

func (m *MatchmakingService) AddPlayer(ctx context.Context, playerID string, score float64, duration int8, playerChannel chan string) {
	m.mutex.Lock()
	m.playerChannels[playerID] = playerChannel
	m.mutex.Unlock()

	// Redis key based on duration (e.g., "score_queue_10min")
	queueKey := fmt.Sprintf("%s_%dmin", m.config.GameConfig.ScoreQueue, duration)

	err := m.redisClient.ZAdd(ctx, queueKey, redis.Z{
		Score:  score,
		Member: playerID,
	}).Err()
	if err != nil {
		m.logger.Println("Error adding player to queue:", err)
		return
	}

}

func (m *MatchmakingService) MatchPlayers(ctx context.Context, minDiff, maxDiff int, duration int8) {
	var wg sync.WaitGroup

	for range m.config.GameConfig.WorkerPoolSize {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.matchWorker(ctx, minDiff, maxDiff, duration)
		}()
	}

	wg.Wait()
}

func (m *MatchmakingService) matchWorker(ctx context.Context, minDiff, maxDiff int, duration int8) error {
	// Use the correct Redis key based on duration
	queueKey := fmt.Sprintf("%s_%dmin", m.config.GameConfig.ScoreQueue, duration)

	backoff := 500 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("could not find an opponent, please retry")
		default:
			players, err := m.redisClient.Eval(ctx, m.luaScript, []string{queueKey},
				fmt.Sprintf("%d", minDiff), fmt.Sprintf("%d", maxDiff)).Result()
			if err != nil || players == nil {
				time.Sleep(backoff)
				if backoff < 2*time.Second {
					backoff *= 2
				}
				continue
			}

			res, ok := players.([]interface{})
			if !ok || len(res) < 2 {
				continue
			}

			player1, _ := res[0].(string)
			player2, _ := res[1].(string)

			if err := m.handleMatch(ctx, player1, player2, duration); err != nil {
				return err
			}
			backoff = 500 * time.Millisecond
		}
	}
}

func (m *MatchmakingService) handleMatch(ctx context.Context, player1, player2 string, duration int8) error {
	gameResp, err := m.storage.CreateGameStorage(ctx, player1, player2, duration)
	if err != nil {
		m.logger.Println("Error creating game:", err)
		return err
	}

	m.mutex.Lock()
	if ch1, ok := m.playerChannels[player1]; ok {
		ch1 <- gameResp.GameId
	}
	if ch2, ok := m.playerChannels[player2]; ok {
		ch2 <- gameResp.GameId
	}
	m.mutex.Unlock()

	return m.redisClient.Publish(ctx, m.config.GameConfig.RedisChannel,
		fmt.Sprintf("%s:%s:%s", player1, player2, gameResp.GameId)).Err()
}
