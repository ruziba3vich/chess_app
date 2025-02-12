package storage

import (
	"context"

	"github.com/ruziba3vich/chess_app/internal/genprotos"
	"github.com/ruziba3vich/chess_app/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Storage) CreateGameStorage(ctx context.Context, player1, player2 string, duration int8) (*genprotos.CreateGameResponse, error) {
	// Create game model with both player IDs and duration
	game := models.GameModel{
		Players:  []string{player1, player2},
		Moves:    []genprotos.Move{}, // Empty moves at the start
		Duration: duration,           // Store duration
	}

	// Insert into MongoDB
	result, err := s.database.GamesCollection.InsertOne(ctx, game)
	if err != nil {
		s.logger.Println("Error inserting game:", err)
		return nil, err
	}

	// Get inserted game ID
	gameID := result.InsertedID.(primitive.ObjectID).Hex()

	return &genprotos.CreateGameResponse{GameId: gameID}, nil
}
