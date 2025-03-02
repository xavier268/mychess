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

	// Status is made of 3 groups of bytes :
	//
	// byte 0-1 : white details
	// 		bit 0-5 : King position
	// 		bit 6 : can castle king side
	// 		bit 7 : can castle king side
	// 		bit 8 : white king under check
	//		bit 9-15 : ruf
	//
	// byte 2-3 : black details
	// 		same format
	//
	// byte 4-5 : uint16 count global ply counter // max 65K !
	// byte 6 : uint8 count ply without capture or pawn move  // max 256
	// byte 7 :
	//    	bit 1 : reversed colors vs actual physical color
	//    	bit 2 : game over
	//    	bit 3 : draw position
	// 		bit 4-7 : ruf
	//
	status uint64
}

// Switch status to reflect change of side
func VMirrorStatus(status uint64) uint64 {
	status = (status&0xFFFF)<<8*2 | (status>>8*2)&0xFFFF | (status & 0xFFFF_FFFF) // switch bytes 0-1 et 2-3
	status = status ^ (1 << 8 * 7)                                                // reverse bit 0 in byte 7, rest of status left unchanged
	return status
}

const (
	StartWhiteOcc  Bitboard = 0xFFFF
	StartBlackOcc  Bitboard = 0xFFFF << 48
	StartPawnOcc   Bitboard = (0xFF << 8) | (0xFF << 48)
	StartRookOcc   Bitboard = 0x81 | (0x81 << 56)
	StartKnightOcc Bitboard = 0x42 | (0x42 << 56)
	StartBishopOcc Bitboard = (1 << 2) | (1 << 5) | (1 << (2 + 56)) | (1 << (5 + 56))
	StartQueenOcc  Bitboard = 1<<3 | (1 << (3 + 56))
	//StartKingOcc   Bitboard = 1<<4 | (1 << (4 + 56))

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
	status: (CanCastle | 4) | // white king position
		(CanCastle|60)<<16, // black king position
}

func (p Position) GetWhiteKingSquare() Square {
	return Square(p.status & 0x3F)
}

func (p Position) GetBlackKingSquare() Square {
	return Square((p.status >> 16) & 0x3F)
}

// Change side ...
func (p Position) VMirror() Position {

	return Position{
		whiteOcc:  p.blackOcc.VMirror(),
		blackOcc:  p.whiteOcc.VMirror(),
		pawnOcc:   p.pawnOcc.VMirror(),
		rookOcc:   p.rookOcc.VMirror(),
		bishopOcc: p.bishopOcc.VMirror(),
		knightOcc: p.knightOcc.VMirror(),
		status:    VMirrorStatus(p.status),
	}

}
