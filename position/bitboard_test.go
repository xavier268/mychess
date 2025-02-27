package position

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

func TestDisplay(t *testing.T) {
	b := Bitboard(0b110000_01100000)
	b.Display()
	b = Bitboard(0x8040201008040201) // diag
	b.Display()
	b = Rank(2)
	b.Display()
	b = File(2)
	b.Display()
	b = Interior()
	b.Display()
	b = Border()
	b.Display()
}

func TestRookMasks(t *testing.T) {
	for sq := Square(0); sq < 64; sq += 13 {
		fmt.Println("Square", sq.String())
		GenerateRookMaskSq(sq).Display()
	}
}

func TestBishopMasks(t *testing.T) {
	for sq := Square(0); sq < 64; sq += 13 {
		fmt.Println("Square", sq.String())
		GenerateBishopMaskSq(sq).Display()
	}
}

func TestKnightAttacks(t *testing.T) {
	for sq := Square(0); sq < 64; sq += 13 {
		fmt.Println("Square", sq.String())
		GenerateKnightAttacksSq(sq).Display()
	}
}

func TestKingAttacks(t *testing.T) {
	for sq := Square(0); sq < 64; sq += 13 {
		fmt.Println("Square", sq.String())
		GenerateKingAttacksSq(sq).Display()
	}
}

func TestDiagonal(t *testing.T) {
	for sq := Square(0); sq < 64; sq += 5 {
		fmt.Println("Square", sq.String())
		Diagonal(sq).Display()
		AntiDiagonal(sq).Display()
	}
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
