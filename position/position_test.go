package position

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestDisplayPosition(t *testing.T) {

	fmt.Println(StartPosition.String())
	StartPosition.PrintWithOverlay(1<<4 | 1<<60)
	StartPosition.Dump()

}

func TestSizes(t *testing.T) {
	fmt.Printf("Size of Position : %d bytes\n", uint64(unsafe.Sizeof(Position{})))
	fmt.Printf("Size of Status : %d bytes\n", uint64(unsafe.Sizeof(Status{})))
	fmt.Printf("Size of ZobristTable : %d bytes\n", uint64(unsafe.Sizeof(ZobristTable{})))

	fmt.Printf("Size of BigTable : %d bytes\n", uint64(unsafe.Sizeof(BigTable{})))

}

func TestRandomPosition(t *testing.T) {

	bt := NewBigTable()

	p := new(Position).
		AddKing(WHITE, "c2").AddKing(BLACK, "c7").
		AddBishop(WHITE, "a2", "a3").
		AddQueen(BLACK, "e4").
		AddRook(BLACK, "h8").
		AddPawn(BLACK, "a7", "b6").AddPawn(WHITE, "d2", "e3", "h5")
	fmt.Println(p)
	fmt.Println(p.status)

	// Verify attacks ...

	type wb struct {
		w bool // attacked by white ?
		b bool // attacked by black ?
	}
	testSqu := map[string]wb{
		// no attacks
		"c7": {false, false},
		"g7": {false, false},
		"h3": {false, false},
		"a6": {false, false},

		// both attacks
		"f4": {true, true},
		"d4": {true, true},
		"e3": {true, true},
		"d6": {true, true},

		// black only attacks
		"h5": {false, true},
		"a5": {false, true},
		"b7": {false, true},

		// white only attacks
		"d1": {true, false},
	}

	for s, res := range testSqu {
		sq := SqParse(s)
		fmt.Println(p)
		fmt.Println("Square", sq.String(), "is attacked by WHITE :", p.IsSquareAttacked(bt, sq, WHITE))
		fmt.Println("Square", sq.String(), "is attacked by BLACK :", p.IsSquareAttacked(bt, sq, BLACK))
		if p.IsSquareAttacked(bt, sq, WHITE) != res.w {
			t.Error("Square", sq.String(), "is attacked by WHITE :", p.IsSquareAttacked(bt, sq, WHITE))
		}
		if p.IsSquareAttacked(bt, sq, BLACK) != res.b {
			t.Error("Square", sq.String(), "is attacked by BLACK :", p.IsSquareAttacked(bt, sq, BLACK))
		}
	}

}
