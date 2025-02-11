package storage

import (
	"context"

	"github.com/ruziba3vich/chess_app/internal/genprotos"
)

func (s *Storage) CreateGameStorage(ctx *context.Context, req *genprotos.CreateGameRequest) (*genprotos.CreateGameResponse, error) {
	if err := s.redis_service.AddPlayerToQueue(
		req.GetPlayerId(),
		req.GetPlayerRank(),
		int8(req.GetDuration())); err != nil {
		s.logger.Println(err.Error())
		return nil, err
	}
	opponent, err := s.redis_service.FindMatch(
		req.GetPlayerId(),
		req.GetPlayerRank(),
		req.GetDuration(),
	)
	if err != nil {
		s.logger.Println(err.Error())
		return nil, err
	}
	
	return &genprotos.CreateGameResponse{}, nil
}
