package service

import "github.com/ruziba3vich/chess_app/internal/genprotos"

type GameService struct {
	genprotos.UnimplementedGameServiceServer
}
