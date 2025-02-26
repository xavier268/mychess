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
	// byte 4 & 5 ( 0 / 65 535)
	// uint16 representing total number of ply so far
	//
	// byte 6 & 7 (-32 000 / +32 000)
	// int16 representing material value of the board ?
	status uint64
}
