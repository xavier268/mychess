package position

type Position struct {
	// all pieces together, EXCLUDING en-passant position, but INCLUDING kings
	whiteOcc Bitboard
	// all pieces together
	blackOcc Bitboard
	// all colors together
	// en passant is indicated as rank 0 for the white pawn (there can never be a black pawn here because of promotion) - or in the status data ?
	pawnOcc Bitboard
	// all colors together
	rookOcc Bitboard
	// all colors together
	bishopOcc Bitboard
	// all colors together
	knightOcc Bitboard

	// queen appears as both a rook and a bishop
	// king positions derived from status data below

	// Detail per bit number :
	//
	// byte 0
	// 0-6: white king position (0 / 63)
	// 7  : white can castle queen side
	// 8  : white can castle king side
	//
	// byte 1
	// 0-6: black king position (0-63)
	// 7  : black can castle queen side
	// 8 :  black can castle king side
	//
	// byte 2
	// 0-6 : counter of ply without capture and without pawn move (0 / 127)
	// 7   : if set, colors of the physical board have been reversed, to ensure WHITE is always expected to play in this position.
	//
	// byte 3
	// reserved - whose king is under attack ( mine or yours ? ), mat or draw, or game-over flags ?
	//
	// byte 6 & 7 ( 0 / 65 535)
	// uint16 representing total number of ply so far
	status uint64
}

const (
	StartWhiteOcc  Bitboard = 0xFFFF
	StartBlackOcc  Bitboard = 0xFFFF << 48
	StartPawnOcc   Bitboard = (0xFF << 8) | (0xFF << 48)
	StartRookOcc   Bitboard = 0x81 | (0x81 << 56)
	StartKnightOcc Bitboard = 0x42 | (0x42 << 56)
	StartBishopOcc Bitboard = (1 << 2) | (1 << 5) | (1 << (2 + 56)) | (1 << (5 + 56))
	StartQueenOcc  Bitboard = 1<<3 | (1 << (3 + 56))
	StartKingOcc   Bitboard = 1<<4 | (1 << (4 + 56))

	CanCastleKingSide  = 0b10000000
	CanCastleQueenSide = 0b01000000
	CanCastle          = CanCastleQueenSide | CanCastleKingSide
)

var StartPosition = Position{
	whiteOcc:  StartWhiteOcc,
	blackOcc:  StartBlackOcc,
	pawnOcc:   StartPawnOcc,
	rookOcc:   StartRookOcc | StartQueenOcc,
	bishopOcc: StartBishopOcc | StartQueenOcc,
	knightOcc: StartKnightOcc,
	status: CanCastle | 4 | // white king position
		(CanCastle|60)<<8, // black king position

}

func (p Position) GetWhiteKingSquare() Square {
	return Square(p.status & 0b00111111)
}

func (p Position) GetBlackKingSquare() Square {
	return Square((p.status >> 8) & 0b00111111)
}
