package storage

import (
	"context"
	"fmt"

	"github.com/notnil/chess"
	"github.com/ruziba3vich/chess_app/internal/genprotos"
	"github.com/ruziba3vich/chess_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Storage) CreateGameStorage(ctx context.Context, player1, player2 string, duration int8) (string, error) {
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
		return "", err
	}

	// Get inserted game ID
	return result.InsertedID.(primitive.ObjectID).Hex(), nil

}

func (s *Storage) MakeMove(ctx context.Context, req *genprotos.MakeMoveRequest) (*genprotos.MakeMoveResponse, error) {
	// Retrieve game from Redis
	game, err := s.redisService.GetGame(req.GameId)
	if err != nil {
		return nil, fmt.Errorf("game not found: %s", err.Error())
	}

	// Validate and apply move
	moveStr := req.Move.MoveFrom + req.Move.MoveTo
	if err := game.MoveStr(moveStr); err != nil {
		return &genprotos.MakeMoveResponse{
			Success: false,
			Message: "Invalid move",
			IsCheck: false,
		}, nil
	}

	// Check if the move results in a check
	isCheck := detectCheck(game)

	// After successful move
	resp := &genprotos.MakeMoveResponse{
		Success:     true,
		Message:     "successful move",
		IsCheck:     isCheck,
		IsCheckmate: detectCheckmate(game), // Fixed: was detectCheck
	}

	// If game ends (checkmate), update MongoDB
	if resp.IsCheckmate {
		objID, _ := primitive.ObjectIDFromHex(req.GameId)
		moves := game.Moves()
		protoMoves := make([]genprotos.Move, len(moves))
		for i, move := range moves {
			protoMoves[i] = genprotos.Move{
				MoveFrom: move.S1().String(),
				MoveTo:   move.S2().String(),
			}
		}

		update := bson.M{
			"$set": bson.M{
				"moves": protoMoves,
			},
		}
		_, err := s.database.GamesCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			s.logger.Println("Failed to update game moves in MongoDB:", err)
		}
	}

	return resp, nil
}

// detectCheck checks if the current position is in check
func detectCheck(game *chess.Game) bool {
	// Get the board and current position
	position := game.Position()
	board := position.Board()
	turn := position.Turn() // The player who just moved

	// Find the opponent's king position
	var kingSquare chess.Square
	for sq, piece := range board.SquareMap() {
		if piece.Type() == chess.King && piece.Color() != turn {
			kingSquare = sq
			break
		}
	}

	// Get all legal moves in the current position
	moves := game.ValidMoves()

	// Check if any move attacks the king's square
	for _, move := range moves {
		if move.S2() == kingSquare {
			return true
		}
	}

	return false
}

func (s *Storage) GetGameStats(ctx context.Context, gameID string) (*genprotos.GetGameStatsResponse, error) {
	// Initialize response
	response := &genprotos.GetGameStatsResponse{
		Moves: make([]*genprotos.Move, 0),
	}

	// Retrieve game from Redis
	game, err := s.redisService.GetGame(gameID)
	if err != nil {
		// If not found in Redis, try to get from MongoDB
		objID, err := primitive.ObjectIDFromHex(gameID)
		if err != nil {
			return nil, fmt.Errorf("invalid game ID: %s", err.Error())
		}

		var gameModel models.GameModel
		err = s.database.GamesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&gameModel)
		if err != nil {
			return nil, fmt.Errorf("game not found: %s", err.Error())
		}
		for i := range gameModel.Moves {
			response.Moves = append(response.Moves, &gameModel.Moves[i])
		}
		return response, nil
	}

	// If game is found in Redis, get moves from the chess game
	moves := game.Moves()
	response.Moves = make([]*genprotos.Move, len(moves))

	for i, move := range moves {
		response.Moves[i] = &genprotos.Move{
			MoveFrom: move.S1().String(),
			MoveTo:   move.S2().String(),
		}
	}

	return response, nil
}

func detectCheckmate(game *chess.Game) bool {
	// If the current player has no valid moves and is in check → Checkmate!
	return len(game.ValidMoves()) == 0 && detectCheck(game)
}
