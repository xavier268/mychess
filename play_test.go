package mychess

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestPlayRandomGame(t *testing.T) {
	p := NewPosition().Reset()
	for i := 0; i < 50; i++ {
		fmt.Println("Game :", i, "\n", p.String())
		mm := p.LegalMoves(make([]Move, 0, 40))
		if len(mm) == 0 {
			fmt.Println("No more legal moves !")
			break
		}
		// select random move
		m := mm[rand.Intn(len(mm))]
		fmt.Printf("Playing %s\n", m.String())
		if piece := p.Board[m.To.Row][m.To.Col]; piece != EMPTY {
			fmt.Println("Capture", StringColor(piece), DISPLAY[piece])
		}
		p.ExecuteMove(m)
	}
}

func TestPreparedPosition1(t *testing.T) {
	p := tp1()

	mm := p.LegalMoves(make([]Move, 0, 40))
	if len(mm) == 0 {
		fmt.Println("No more legal moves !")
	}
	fmt.Println("Legal moves :", len(mm))
	// for _, m := range mm {
	// 	fmt.Println(m.String())
	// }
	if len(mm) != 44 {
		t.Errorf("Expected 44 legal moves, got %d", len(mm))
	}

}

// Test position 1
func tp1() *Position {

	p := NewPosition()

	p.SetPiece(KING, "d3")
	p.SetPiece(QUEEN, "d4")
	p.SetPiece(KNIGHT, "d5")
	p.SetPiece(BISHOP, "c5")
	p.SetPiece(ROOK, "e5")
	p.SetPiece(PAWN, "a2", "c3")

	p.SetPiece(-ROOK, "b3")
	p.SetPiece(-KING, "c8")

	return p
}
