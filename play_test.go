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

func TestPreparaPosition(t *testing.T) {
	p := NewPosition()

	p.SetPiece(KING, "d3")
	p.SetPiece(ROOK, "d4")

	p.SetPiece(-KING, "c8")
	p.SetPiece(-QUEEN, "c7", "d8")

	fmt.Println(p.String())

}
