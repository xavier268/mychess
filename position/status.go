package position

// Special type to handle status word.
// Status is made of 3 groups of bytes :
//
// byte 0-1 : white details
//		bit 0-5 : King position
//		bit 6 : can castle king side
//		bit 7 : can castle king side
//		bit 8 : white king under check
//		bit 9-15 : ruf
//
// byte 2-3 : black details
//		same format
//
// byte 4-5 : uint16 count global ply counter // max 65K !
// byte 6 : uint8 count ply without capture or pawn move  // max 256
// byte 7 :
//	   	bit 1 : reversed colors vs actual physical color
//	   	bit 2 : game over
//	   	bit 3 : draw position
//		bit 4-7 : ruf

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

func (st Status) IsGameOver() bool {
	return st.Game&0b1000 != 0
}

func (st Status) IsDraw() bool {
	return st.Game&0b0100 != 0
}

const (
	// Castle bits
	CanCastleKingSide  = 0b10000000
	CanCastleQueenSide = 0b01000000
	CanCastle          = CanCastleQueenSide | CanCastleKingSide
)
