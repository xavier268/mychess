package position

import (
	"fmt"
	"testing"
)

func TestDisplayPosition(t *testing.T) {

	fmt.Println(StartPosition.String())
	StartPosition.Dump()

}

func TestRandomPosition(t *testing.T) {

	bt := NewBigTable()

	p := new(Position).
		AddKing(WHITE, "c2").AddKing(BLACK, "c7").
		AddBishop(WHITE, "a2", "a3").
		AddQueen(BLACK, "e4").
		AddRook(BLACK, "h8").
		AddPawn(BLACK, "a7", "b6").AddPawn(WHITE, "d2", "e3")
	p.UpdateKingThreats(bt)
	fmt.Println(p)
	fmt.Println(p.status)
}
