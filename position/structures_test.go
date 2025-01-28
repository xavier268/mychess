package position

import (
	"fmt"
	"testing"
)

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
