package position

import "testing"

func TestBitBoardDisplay(t *testing.T) {
	b := Bitboard(0x_1F_AA_01)
	b.Display()
}
