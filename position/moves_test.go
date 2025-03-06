package position

import (
	"fmt"
	"testing"
)

func TestGetMoves(t *testing.T) {
	p := StartPosition
	fmt.Println(p.String())
	bt := NewBigTable()
	moves := p.GetMoveList(bt)
	for i, m := range moves {
		fmt.Println(i, "\t", m.String())
	}
}
