package mychess

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestPlayRandomGame(t *testing.T) {
	t.Skip()
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

func TestPreparedPosition1LegalMoves(t *testing.T) {
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

func TestEnPassant2(t *testing.T) {
	p := tp2()
	if p.EnPassant != (Square{}) {
		t.Errorf("Expected no en passant square, got %s", p.EnPassant.String())
	}
	fmt.Println("Test position #2, before en passant\n", p.String())
	m := Move{Piece: PAWN, From: SquareFromString("b2"), To: SquareFromString("b4")}
	p.ExecuteMove(m)
	fmt.Println("Test position #2, after en passant activated\n", p.String())
	if p.EnPassant != SquareFromString("b3") {
		t.Errorf("Expected en passant capture square, got %s", p.EnPassant.String())
	}

	mm := p.LegalMoves(make([]Move, 0, 40))
	fmt.Println("BLACK legal moves", len(mm))
	// for _, m := range mm {
	// 	fmt.Println(m.String())
	// }
	if len(mm) != 22 {
		t.Errorf("Expected 22 legal moves, got %d", len(mm))
	}
	fmt.Println("Assume BLACK ROOK takes en passant b3")
	m = Move{Piece: -ROOK, From: SquareFromString("e3"), To: SquareFromString("b3")}
	fmt.Println(m.String())
	p.ExecuteMove(m)
	fmt.Println("Test position #2, after black rook took en passant\n", p.String())
	if p.EnPassant != (Square{}) {
		t.Errorf("Expected no en passant square, got %s", p.EnPassant.String())
	}
	if p.Board[3][1] != EMPTY {
		t.Errorf("Expected empty square at b3 after black rook took en passant b4, but got %s", StringColor(p.Board[3][1]))
	}
}

func TestDisplayPositions(t *testing.T) {
	p := NewPosition()
	fmt.Println("Empty position\n", p.String())
	p = NewPosition().Reset()
	fmt.Println("Initial position\n", p.String())
	p = tp1()
	fmt.Println("Test position #1\n", p.String())
	p = tp2()
	fmt.Println("Test position #2\n", p.String())
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

// Test position 2 - en passant captures
func tp2() *Position {

	p := NewPosition()

	p.SetPiece(KING, "e1")
	p.SetPiece(-KING, "e8")

	p.SetPiece(PAWN, "b2")

	p.SetPiece(-PAWN, "a4", "c4")

	p.SetPiece(-ROOK, "e3")

	return p
}
