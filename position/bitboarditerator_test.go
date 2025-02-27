package position

import "testing"

func TestBitboardIterator(t *testing.T) {
	sq := Sq(5, 2)
	b := GenerateKingAttacksSq(sq)

	for bb := range b.BitCombinations {
		bb.Display()
	}
}
