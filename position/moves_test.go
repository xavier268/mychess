package position

import (
	"fmt"
	"testing"
)

func TestGetMoves1(t *testing.T) {
	p := StartPosition

	fmt.Println(p.String())
	moves := p.GetMoveList(bt)
	for i, m := range moves {
		fmt.Println(i, "\t", m.String())
	}
	if len(moves) != 20 {
		t.Error("wrong number of moves, expected 20, got", len(moves))
	}

	p = *rpt1
	fmt.Println(p.String())
	moves = p.GetMoveList(bt)
	for i, m := range moves {
		fmt.Println(i, "\t", m.String())
	}
	if len(moves) != 24 {
		t.Error("wrong number of moves, expected 24, got", len(moves))
	}

}
