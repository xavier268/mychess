package position

import (
	"fmt"
	"testing"
)

func TestGenerateRookAttacksTable(t *testing.T) {
	for r := range 4 {
		for f := range 4 {
			sq := Sq(r, f)
			m := GenerateRookAttacksMagicMapSq(sq)
			fmt.Printf("Rook table for %s stats : len %d keys\n", sq.String(), len(m))
		}
	}
}

func TestGenerateBishopAttacksTable(t *testing.T) {
	for r := range 4 {
		for f := range 4 {
			sq := Sq(r, f)
			m := GenerateBishopAttacksMagicMapSq(sq)
			fmt.Printf("Bishop table for %s stats : len %d keys\n", sq.String(), len(m))
		}
	}
}
