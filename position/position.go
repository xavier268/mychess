package position

const (
	EMPTY = iota
	PAWN
	BISHOP
	KNIGHT
	ROOK
	QUEEN
	KING
)

const (
	WHITE = 1
	BLACK = 0
)

type Position struct {
	// all pieces together, without en-passant position, but with kings
	whiteOcc  Bitboard
	// all pieces together
	blackOcc  Bitboard
	// all colors together 
	// en passant is indicated as rank 0 for the white pawn (there can never be a black pawn here because of promotion) - or in the status data ?
	pawnOcc   Bitboard 
	// all colors together
	rookOcc   Bitboard 
	// all colors together
	bishopOcc Bitboard 

	// queen appears as both a rook and a bishop
	// king positions derived from status data below
	
	// Detail per bit number :
	// 0-5 : counter of moves without capture and without pawn move (0-63)
	// 6 : who should play ?
	// 7 : reserved
	//
	// 16-21 black king position (0-63)
	// 22 : black can big castle
	// 23 : black can small castle
	//
	// 24-29 white king position (0-63)
	// 30 : white can big castle
	// 31 : white can small castle
	// 
	// 40 : en passant active
	// 41-43 : en passant file number (0-7)
	//
	// 48 - 55 : reserved
	//
	// 56 - 63 : reserved
	status uint64 
}
	     

}
	
}
}
