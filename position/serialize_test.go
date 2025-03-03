package position

import (
	"fmt"
	"math/bits"
	"math/rand/v2"
	"testing"
)

func TestSerialize(t *testing.T) {

	for i := Bitboard(0); i < 15; i++ {
		verifySerialize(t, i)
		verifySerialize(t, 1<<i)
		verifySerialize(t, Bitboard(rand.Uint64()))
	}
}

func TestSerializeAll(t *testing.T) {
	for range 10 {
		b := Bitboard(rand.Uint64())
		// Use iterator
		bb1 := make([]Bitboard, 0, 50)
		for sb := range b.AllSerialized {
			bb1 = append(bb1, sb)
		}
		// compare with slice method
		bb2 := b.Serialize()
		if len(bb1) != len(bb2) {
			t.Errorf("Wrong length got %d", len(bb1))
		}
		for i := range bb1 {
			if bb1[i] != bb2[i] {
				t.Errorf("Wrong bitboard got %d", bb1[i])
			}
		}
	}
}

func verifySerialize(t *testing.T, b Bitboard) {

	bbs := b.Serialize()
	fmt.Printf("\n %d ", uint64(b))
	if len(bbs) != bits.OnesCount64((uint64(b))) {
		t.Errorf("Wrong slice length got %d", len(bbs))
	}
	s := Bitboard(0)
	for _, k := range bbs {
		if bits.OnesCount64((uint64)(k)) != 1 {
			t.Errorf("Wrong bitboard got %d", k)
		}
		s += k
	}
	if s != b {
		t.Errorf("Wrong bitboard got %d", s)
	}
	fmt.Printf(" OK !")
}
