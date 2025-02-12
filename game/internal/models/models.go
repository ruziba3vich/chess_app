package models

import (
	"github.com/ruziba3vich/chess_app/internal/genprotos"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	GameModel struct {
		ID       primitive.ObjectID `bson:"_id,omitempty"`
		Players  []string           `bson:"players"`
		Duration int8               `bson:"duration"`
		Moves    []genprotos.Move   `bson:"moves"`
	}
)
