package position

import (
	"fmt"
	"strings"
)

// =======================
// Piece based approach (printing & testing only)
// =======================

const (
	WHITE = 1
	BLACK = -1
)

type Piece int16

const (
	EMPTY  Piece = 0
	PAWN   Piece = 1
	KNIGHT Piece = 1 << 2
	BISHOP Piece = 1 << 3
	ROOK   Piece = 1 << 4
	QUEEN  Piece = ROOK | BISHOP
	KING   Piece = 1 << 5
)

var PieceRepresentation = map[Piece]rune{
	PAWN:    'P',
	BISHOP:  'B',
	KNIGHT:  'N',
	ROOK:    'R',
	QUEEN:   'Q',
	KING:    'K',
	EMPTY:   '.',
	-PAWN:   'p',
	-BISHOP: 'b',
	-KNIGHT: 'n',
	-ROOK:   'r',
	-QUEEN:  'q',
	-KING:   'k',
}

// Print position.
// Rank 8 is printed first.
func (p Position) String() string {
	buf := new(strings.Builder)

	buf.WriteString("-- ")
	for i := 'a'; i <= 'h'; i++ {
		buf.WriteRune(i)
		buf.WriteRune(' ')
	}
	buf.WriteString("--")

	for r := 7; r >= 0; r-- {
		fmt.Fprintf(buf, "\n%d  ", r+1)
		for f := 0; f < 8; f++ {
			buf.WriteRune(PieceRepresentation[p.PieceAt(Sq(r, f))])
			buf.WriteRune(' ')
		}
		fmt.Fprintf(buf, " %d", r+1)
	}
	buf.WriteString("\n-- ")
	for i := 'a'; i <= 'h'; i++ {
		buf.WriteRune(i)
		buf.WriteRune(' ')
	}
	buf.WriteString("--")
	return buf.String()
}

// Piece at specified square.
func (p Position) PieceAt(sq Square) Piece {

	//handle kings differently
	if sq == p.GetBlackKingSquare() {
		return -KING
	}
	if sq == p.GetWhiteKingSquare() {
		return KING
	}

	// Normal pieces
	color := Piece(p.whiteOcc.Get(sq) - p.blackOcc.Get(sq))
	piece := Piece(p.pawnOcc.Get(sq) | p.knightOcc.Get(sq)<<2 | p.bishopOcc.Get(sq)<<3 | p.rookOcc.Get(sq)<<4)
	return color * piece
}

func (p Position) Dump() {
	fmt.Println("White occ : ", p.whiteOcc.String())
	fmt.Println("Black occ : ", p.blackOcc.String())
	fmt.Println("Pawn occ : ", p.pawnOcc.String())
	fmt.Println("Knight occ : ", p.knightOcc.String())
	fmt.Println("Bishop occ : ", p.bishopOcc.String())
	fmt.Println("Rook occ : ", p.rookOcc.String())
	fmt.Println("White king sq : ", Bitboard(1<<p.GetWhiteKingSquare()).String())
	fmt.Println("Black king sq : ", Bitboard(1<<p.GetBlackKingSquare()).String())
	fmt.Printf("Status : %64b\n", p.status)
}

func (p Position) GetWhiteKingSquare() Square {
	return Square(p.status.GetWhiteKingSquare())
}

func (p Position) GetBlackKingSquare() Square {
	return Square(p.status.GetBlackKingSquare())
}
