package game

import (
	"context"
	"mychess/position"
)

// A Game capture the context of an on-going game.
type Game struct {
	// The current position of the game (includes the turn)
	Position position.Position
	// past moves played until now
	History []position.Move
	// Context for game
	Ctx context.Context
	// ZobristHash -> ZEntry
	Z map[uint64]ZEntry
}

type ZEntry struct {
	// Best move found until now (or null move)
	Best position.Move
	// Score (upper, lower, exact)
	Score position.Score
	// Score type : UPPER, LOWER, EXACT
	ScoreType ScoreType
	// Depth of analysis at this stage
	Depth int16
	// When was this entry last updated ?
	Age uint8
}

type ScoreType uint8

const (
	UPPER ScoreType = iota
	LOWER
	EXACT
)

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
