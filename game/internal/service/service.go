package service

import (
	"context"

	"github.com/ruziba3vich/chess_app/internal/game_service"
	"github.com/ruziba3vich/chess_app/internal/genprotos"
	"github.com/ruziba3vich/chess_app/internal/storage"
)

type GameService struct {
	genprotos.UnimplementedGameServiceServer
	storage     *storage.Storage
	gameService *game_service.MatchmakingService
}

func NewGameService(storage *storage.Storage) *GameService {
	return &GameService{
		storage: storage,
	}
}

func (g *GameService) CreateGame(ctx context.Context, req *genprotos.CreateGameRequest) error {
	return g.gameService.AddPlayer(ctx, req.PlayerId, float64(req.PlayerRank), req.Duration, make(chan string))
}
func (g *GameService) GetGameStats(ctx context.Context, req *genprotos.GetGameStatsRequest) (*genprotos.GetGameStatsResponse, error) {
	return g.storage.GetGameStats(ctx, req.GameId)
}

func (g *GameService) MakeMove(ctx context.Context, req *genprotos.MakeMoveRequest) (*genprotos.MakeMoveResponse, error) {
	return g.storage.MakeMove(ctx, req)
}
