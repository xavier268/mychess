package position

import (
	"fmt"
	"strings"
)

// ====================================
// Square object
// ====================================

// Square from 0 to 63
type Square int

// Rank/File coordinates of a given square
// 0-based
func (s Square) RF() (rank int, file int) {
	return int(s) / 8, int(s) % 8
}

// Create a square from the rank/file coordinates
func Sq(rank, file int) Square {
	return Square(rank*8 + file)
}

// create a square from the string expression, eg : "d2"
func SqParse(s string) Square {
	if len(s) != 2 {
		panic("invalid square string")
	}
	file := int(s[0] - 'a')
	rank := int(s[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		panic("invalid square")
	}
	return Sq(rank, file)
}

func (s Square) String() string {
	rank, file := s.RF()
	return fmt.Sprintf("%c%d", 'a'+file, rank+1)
}

func (s Square) IsValid() bool {
	return s >= 0 && s < 64
}

// ======================================================
// BitBoard object
//=======================================================

// A 64-bit bitmap
type Bitboard uint64

func (b Bitboard) IsSet(pos Square) bool {
	return b&(1<<pos) != 0
}

func (b *Bitboard) Set(pos Square) {
	*b |= 1 << pos
}

func (b *Bitboard) Unset(pos Square) {
	*b &= ^(1 << pos)
}

func (b Bitboard) String() string {
	sb := new(strings.Builder)
	fmt.Fprintf(sb, "\n   ")
	for i := 0; i < 8; i++ {
		fmt.Fprintf(sb, " %c ", 'a'+i)
	}
	for i := Square(0); i < 64; i++ {
		if i%8 == 0 {
			fmt.Fprintf(sb, "\n%d  ", i/8+1)
		}
		if b.IsSet(i) {
			sb.WriteString(" \u25CF ")
		} else {
			sb.WriteString(" . ")
		}
		if i%8 == 7 {
			fmt.Fprintf(sb, "  %d", i/8+1)
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
