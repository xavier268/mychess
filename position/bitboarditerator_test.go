package position

import (
	"fmt"
	"testing"
)

func TestBitboardIterator(t *testing.T) {
	sq := Sq(5, 2)
	b := GenerateKingAttacksSq(sq)

	for bb := range b.AllBitCombinations {
		bb.Display()
		fmt.Printf("Set squares :")
		for sq := range bb.AllSetSquares {
			fmt.Printf(" %d,", sq)
		}
		fmt.Println()
	}
}
