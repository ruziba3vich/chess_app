package game_service_test

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ruziba3vich/chess_app/internal/game_service"
	"github.com/ruziba3vich/chess_app/internal/storage"
	"github.com/ruziba3vich/chess_app/pkg/config"
)

type MockStorage struct {
	mock.Mock
	storage.Storage
}

func (m *MockStorage) CreateGameStorage(ctx context.Context, player1, player2 string, duration int8) (string, error) {
	args := m.Called(ctx, player1, player2, duration)
	return args.Get(0).(string), args.Error(1)
}

func TestMatchmakingService(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mockStorage := new(MockStorage)
	config, _ := config.LoadConfig()
	logger := log.New(os.Stdout, "", log.LstdFlags)
	playerChannels := make(map[string]chan string)
	wg := &sync.WaitGroup{}

	// Read Lua script from file
	luaScript := `
	local key = KEYS[1]
	local minScore = tonumber(ARGV[1])
	local maxScore = tonumber(ARGV[2])
	
	-- Get players within the score range
	local candidates = redis.call('ZRANGEBYSCORE', key, minScore, maxScore, 'LIMIT', 0, 5)
	
	if #candidates < 2 then return nil end
	
	local p1 = tostring(candidates[1])
	local p2 = tostring(candidates[2])
	
	-- Remove players from the queue
	redis.call('ZREM', key, p1, p2)
	
	-- Return matched players
	return {p1, p2}
	`
	logger.Println(luaScript)

	service := game_service.NewMatchmakingService(redisClient, playerChannels, config, storage.NewStorage(nil, logger, nil), wg, logger, luaScript)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go service.MatchPlayers(ctx, 10, 50, 10)
	// Create player channels
	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	// Add players
	service.AddPlayer(context.Background(), "player1", 1500, 10, ch1)
	service.AddPlayer(context.Background(), "player2", 1520, 10, ch2)

	// Run matchmaking in a goroutine

	// Validate match result
	select {
	case gameID := <-ch1:
		assert.Equal(t, "game123", gameID)
	case <-time.After(2 * time.Second):
		t.Fatal("Player 1 did not receive a match")
	}

	select {
	case gameID := <-ch2:
		assert.Equal(t, "game123", gameID)
	case <-time.After(2 * time.Second):
		t.Fatal("Player 2 did not receive a match")
	}

	mockStorage.AssertExpectations(t)
}
