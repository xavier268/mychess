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

	/*
		sq := SqParse("c2")
		fmt.Println("Is C2 attacked by WHITE :", p.IsSquareAttacked(bt, sq, WHITE))
		fmt.Println("Is C2 attacked by BLACK :", p.IsSquareAttacked(bt, sq, BLACK))
		fmt.Println("C2 Bishop BLACK attack set ", p.GetBishopMovesFromSquare(bt, BLACK, sq).String())
		fmt.Println("C2 Bishop WHITE attack set ", p.GetBishopMovesFromSquare(bt, WHITE, sq).String())

		sq = SqParse("f4")
		fmt.Println("Is F4 attacked by WHITE :", p.IsSquareAttacked(bt, sq, WHITE))
		fmt.Println("WHITE Pawns move from F4", p.GetPawnMovesFromSquare(bt, WHITE, sq).String())
		fmt.Println("BLACK Pawns move from F4", p.GetPawnMovesFromSquare(bt, BLACK, sq).String())
	*/

	for _, s := range []string{"c7", "f4", "h5"} {
		sq := SqParse(s)
		fmt.Println(p)
		fmt.Println("Square", sq.String(), "is attacked by WHITE :", p.IsSquareAttacked(bt, sq, WHITE))
		fmt.Println("Square", sq.String(), "is attacked by BLACK :", p.IsSquareAttacked(bt, sq, BLACK))
	}

}
