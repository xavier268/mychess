package mychess

import (
	"fmt"
	"testing"
)

func TestDisplay(t *testing.T) {
	// Create a new position
	pos := NewPosition().Reset()
	// Display it
	fmt.Println(pos)
}

func TestDisplayMoves(t *testing.T) {
	mm := []Move{
		{PAWN, Square{1, 1}, Square{2, 1}},
		{-PAWN, Square{1, 2}, Square{2, 2}},
		{QUEEN, Square{1, 3}, Square{2, 3}},
		{-QUEEN, Square{1, 4}, Square{2, 4}},
		{KING, Square{1, 5}, Square{2, 5}},
	}
	for _, m := range mm {
		fmt.Println(m.String())
	}
}

func TestLegalMovesAtStart(t *testing.T) {
	// Create a new position
	pos := NewPosition().Reset()
	// Display it
	fmt.Println(pos)
	// Display legal moves
	moves := pos.LegalMoves(make([]Move, 0, 40))
	fmt.Println("Legal moves : ")
	for _, m := range moves {
		fmt.Println(m.String())
	}
}
