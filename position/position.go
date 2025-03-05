package position

// Constant to indicate turn (who is expected to play ) and index the masks and status.
const (
	WHITE = 0
	BLACK = 1
)

type Position struct {
	// // Occupancies for each side, all actual pieces (king INCLUDED), but en passant marker not included
	colOcc [2]Bitboard
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
	status Status
}

const (
	StartWhiteOcc  Bitboard = 0xFFFF
	StartBlackOcc  Bitboard = 0xFFFF << 48
	StartPawnOcc   Bitboard = (0xFF << 8) | (0xFF << 48)
	StartRookOcc   Bitboard = 0x81 | (0x81 << 56)
	StartKnightOcc Bitboard = 0x42 | (0x42 << 56)
	StartBishopOcc Bitboard = (1 << 2) | (1 << 5) | (1 << (2 + 56)) | (1 << (5 + 56))
	StartQueenOcc  Bitboard = 1<<3 | (1 << (3 + 56))
)

var StartPosition = Position{
	colOcc: [2]Bitboard{
		StartWhiteOcc,
		StartBlackOcc,
	},
	pawnOcc:   StartPawnOcc,
	rookOcc:   StartRookOcc | StartQueenOcc,
	bishopOcc: StartBishopOcc | StartQueenOcc,
	knightOcc: StartKnightOcc,
	status:    StartStatus,
}
