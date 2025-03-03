package position

import (
	"fmt"
	"strings"
)

// Special type to handle status word.
type Status struct {
	Plies               uint16
	PliesWithoutCapture uint8
	KingPosition        [2]Square
	CastleBits          [2]uint8
	// Bit 0 : who should play ?
	// Bit 1 : white king under threat
	// Bit 2 : black king under threat
	// Bit 3 : game over (no more legal moves or draw)
	// Bit 4 : draw
	Game uint8
}

// Status at start of game
var StartStatus = Status{
	Plies:               0,
	PliesWithoutCapture: 0,
	KingPosition:        [2]Square{4, 60},
	CastleBits:          [2]uint8{CanCastle, CanCastle},
	Game:                0,
}

// Who should move from here ?
func (st Status) Turn() uint8 {
	return st.Game & 1
}

func (st Status) KingUnderThreat(color uint8) bool {
	return st.Game&(1<<(color+1)) != 0
}

func (st Status) GameOver() bool {
	return st.Game&0b1000 != 0
}

func (st Status) Draw() bool {
	return st.Game&0b0100 != 0
}

func (st Status) CanCastle(side uint8, castleBits uint8) bool {
	return st.CastleBits[side]&castleBits != 0

}

func (st Status) SetKingThreatBit(side uint8, value uint8) Status {
	st.Game = (st.Game & ^(1 << (side + 1))) | ((value & 1) << (side + 1))
	return st
}

const (
	// Castle bits
	CanCastleKingSide  = 0b10000000
	CanCastleQueenSide = 0b01000000
	CanCastle          = CanCastleQueenSide | CanCastleKingSide
)

func (st Status) String() string {
	buf := new(strings.Builder)
	fmt.Fprintln(buf, "\n--- Status ---")
	fmt.Fprintf(buf, "Turn: %d ( %d = WHITE, %d = BLACK )\n", st.Turn(), WHITE, BLACK)
	fmt.Fprintf(buf, "Plies: %d\n", st.Plies)
	fmt.Fprintf(buf, "Plies without capture: %d\n", st.PliesWithoutCapture)
	if st.KingUnderThreat(WHITE) {
		fmt.Fprintf(buf, "WHITE king under threat\n")
	}
	if st.KingUnderThreat(BLACK) {
		fmt.Fprintf(buf, "BLACK king under threat\n")
	}
	if st.CanCastle(WHITE, CanCastleKingSide) {
		fmt.Fprintf(buf, "WHITE can castle king side\n")
	}
	if st.CanCastle(WHITE, CanCastleQueenSide) {
		fmt.Fprintf(buf, "WHITE can castle queen side\n")
	}
	if st.CanCastle(BLACK, CanCastleKingSide) {
		fmt.Fprintf(buf, "BLACK can castle king side\n")
	}
	if st.CanCastle(BLACK, CanCastleQueenSide) {
		fmt.Fprintf(buf, "BLACK can castle queen side\n")
	}
	if st.Draw() {
		fmt.Fprintf(buf, "Game is a Draw\n")
	}
	if st.GameOver() {
		fmt.Fprintf(buf, "Game is over !\n")
	}
	return buf.String()
}

func Bool2uint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
