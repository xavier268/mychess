package position

import (
	"fmt"
	"math/bits"
	"strings"
)

// ======================================================
// BitBoard object
//=======================================================

// A 64-bit bitmap
type Bitboard uint64

func (b Bitboard) Set(pos Square) Bitboard {
	return b | (1 << pos)
}
func (b Bitboard) Unset(pos Square) Bitboard {
	return b & ^(1 << pos)
}
func (b Bitboard) IsSet(pos Square) bool {
	return b&(1<<pos) != 0
}

func (b Bitboard) Get(bit Square) int {
	return int((b >> bit) & 1)
}

// affiche un bitboard avec les rank/files
// le rank 0 est en bas.
func (b Bitboard) String() string {
	sb := new(strings.Builder)
	fmt.Fprintf(sb, "\n   ")
	for i := 0; i < 8; i++ {
		fmt.Fprintf(sb, " %c ", 'a'+i)
	}
	for i := Square(0); i < 64; i++ {
		if i%8 == 0 {
			fmt.Fprintf(sb, "\n%d  ", i.VMirror()/8+1)
		}
		if b.IsSet(i.VMirror()) {
			sb.WriteString(" \u25CF ")
		} else {
			sb.WriteString(" . ")
		}
		if i%8 == 7 {
			fmt.Fprintf(sb, "  %d", i.VMirror()/8+1)
		}
	}
	fmt.Fprintf(sb, "\n   ")
	for i := 0; i < 8; i++ {
		fmt.Fprintf(sb, " %c ", 'a'+i)
	}
	return sb.String()

}

func (b Bitboard) Display() {
	fmt.Printf("Bitboard : %016X\n%s\n", uint64(b), b.String())
}

// Count nbr of bits in b
func (b Bitboard) BitCount() int {
	return bits.OnesCount64(uint64(b))
}

// ======================================================
// Uint64 transformations
//=======================================================

// Vertical mirror - use to exchange white-black
func (x Bitboard) VMirror() Bitboard {
	const (
		k1 Bitboard = 0x00FF00FF00FF00FF
		k2 Bitboard = 0x0000FFFF0000FFFF
	)
	x = ((x >> 8) & k1) | ((x & k1) << 8)
	x = ((x >> 16) & k2) | ((x & k2) << 16)
	x = (x >> 32) | (x << 32)
	return x
}

func (x Bitboard) HMirror() Bitboard {
	const (
		k1 Bitboard = 0x5555555555555555
		k2 Bitboard = 0x3333333333333333
		k4 Bitboard = 0x0F0F0F0F0F0F0F0F
	)
	x = ((x >> 1) & k1) | ((x & k1) << 1)
	x = ((x >> 2) & k2) | ((x & k2) << 2)
	x = ((x >> 4) & k4) | ((x & k4) << 4)
	return x
}

//=============================
// Bitboard constructors and constants
//==============================

func Rank(i int) Bitboard {
	return 0xFF << (i * 8)
}

func File(i int) Bitboard {
	return 0x0101010101010101 << i
}

func Full() Bitboard {
	return 0xFFFFFFFFFFFFFFFF
}

func Border() Bitboard {
	return 0xFF818181818181FF
}

func Interior() Bitboard {
	return ^Bitboard(0xFF818181818181FF)
}

// Pre-calculated diagonal and anti-diagonal masks (indexed on r + f)
var (
	antidiagonals = [15]Bitboard{
		0x1,
		0x102,
		0x10204,
		0x1020408,
		0x102040810,
		0x10204081020,
		0x1020408102040,
		0x102040810204080,
		0x204081020408000,
		0x408102040800000,
		0x810204080000000,
		0x1020408000000000,
		0x2040800000000000,
		0x4080000000000000,
		0x8000000000000000,
	}

	diagonals = [15]Bitboard{
		0x80,
		0x8040,
		0x804020,
		0x80402010,
		0x8040201008,
		0x804020100804,
		0x80402010080402,
		0x8040201008040201,
		0x4020100804020100,
		0x2010080402010000,
		0x1008040201000000,
		0x804020100000000,
		0x402010000000000,
		0x201000000000000,
		0x100000000000000,
	}
)

func AntiDiagonal(sq Square) Bitboard {
	r, f := sq.RF()
	return antidiagonals[r+f]
}

func Diagonal(sq Square) Bitboard {
	r, f := sq.RF()
	return diagonals[r+7-f]
}
