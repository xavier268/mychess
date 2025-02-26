package position

import (
	"math/rand/v2"
	"testing"
)

func TestDisplay(t *testing.T) {
	b := Bitboard(0b110000_01100000)
	b.Display()
	b = Bitboard(1)
	b.Display()
}

func TestVMirror(t *testing.T) {

	b := Bitboard(rand.Uint64())
	b.Display()
	c := b.VMirror()
	c.Display()

	for sq := Square(0); sq < 64; sq++ {
		if b.Get(sq) != c.Get(sq.VMirror()) {
			t.Errorf("VMirror failed for square %d", sq)
		}
	}
}

func TestHMirror(t *testing.T) {

	b := Bitboard(rand.Uint64())
	b.Display()
	c := b.HMirror()
	c.Display()

	for sq := Square(0); sq < 64; sq++ {
		if b.Get(sq) != c.Get(sq.HMirror()) {
			t.Errorf("HMirror failed for square %d", sq)
		}
	}

}
