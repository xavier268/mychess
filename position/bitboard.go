package position

import (
	"fmt"
	"strings"
)

// Square from 0 to 63
type Square int

// A 64-bit map
type BitBoard uint64

func (b BitBoard) IsSet(pos Square) bool {
	return b&(1<<pos) != 0
}

func (b *BitBoard) Set(pos Square) {
	*b |= 1 << pos
}

func (b *BitBoard) Unset(pos Square) {
	*b &= ^(1 << pos)
}

func (b BitBoard) String() string {
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

func (b BitBoard) Display() {
	fmt.Printf("Bitboard : %016X\n%s\n", uint64(b), b.String())
}
