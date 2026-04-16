package game

import (
	"context"
	"mychess/position"
)

// A Game capture the context of an on-going game.
type Game struct {
	// The current position of the game (includes the turn)
	Position position.Position
	// past moves until now
	History []position.Move
	// Context for game
	Ctx context.Context
	// ZobristHash -> ZEntry
	Z map[uint64]ZEntry
}

type ZEntry struct {
	// Upper avalaible score
	Upper position.Score
	// Lower available score
	Lower position.Score
	// Current best move identified
	Best position.Move
	// Depth of analysis at this stage
	Depth int
}

func NewGame(ctx context.Context) *Game {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Game{
		Position: position.StartPosition,
		History:  make([]position.Move, 0, 100),
		Ctx:      ctx,
		Z:        make(map[uint64]ZEntry, 1000),
	}
}
