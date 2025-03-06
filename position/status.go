package position

import (
	"fmt"
	"strings"
)

// Special type to handle status word.
type Status struct {
	// Bit 0-5 : king square position
	// Bit 6-7 : castle bits
	KingStatus [2]uint8
	// Bit 0 : 	Who should play WHITE/BLACK
	// Bit 1-4 : en passant file
	TurnStatus uint8
}

const (
	// Castle bits
	CanCastleKingSide  = 0b10000000
	CanCastleQueenSide = 0b01000000
	CanCastle          = CanCastleQueenSide | CanCastleKingSide
)

// Status at start of game
var StartStatus = Status{
	KingStatus: [2]uint8{
		CanCastle | 4,
		CanCastle | 60,
	},
	TurnStatus: WHITE,
}

// Who should move from here ?
func (st Status) GetTurn() uint8 {
	return st.TurnStatus & 1
}

func (st *Status) SetTurn(turn uint8) {
	st.TurnStatus = (st.TurnStatus & 0b1110) | (turn & 1)
}

func (st Status) CanCastle(side uint8, castleBits uint8) bool {
	return st.KingStatus[side]&castleBits != 0
}

func (st Status) GetCastleBits(side uint8) uint8 {
	return uint8(st.KingStatus[side] & CanCastle)
}

func (st Status) GetKingPosition(side uint8) Square {
	return Square(st.KingStatus[side] & 0b11_1111)
}

func (st Status) String() string {
	buf := new(strings.Builder)
	fmt.Fprintln(buf, "\n--- Status ---")
	fmt.Fprintf(buf, "Turn: %d ( %d = WHITE, %d = BLACK )\n", st.GetTurn(), WHITE, BLACK)

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

	return buf.String()
}

func Bool2uint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}
