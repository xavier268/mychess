package position

import (
	"fmt"
	"strings"
)

// =======================
// Piece based approach (printing & testing only)
// =======================

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

// Print position with the provided overlay
func (p Position) PrintWithOverlay(overlay Bitboard) {
	before := "\033[41m" // before overlay
	after := "\033[0m"   // after overlay

	fmt.Printf("Overlay : 0x%x\n", uint64(overlay))
	fmt.Printf("-- ")
	for i := 'a'; i <= 'h'; i++ {
		fmt.Printf("%c ", i)
	}
	fmt.Printf("--")

	for r := 7; r >= 0; r-- {
		fmt.Printf("\n%d  ", r+1)
		for f := 0; f < 8; f++ {
			if overlay.Get(Sq(r, f)) == 1 {
				fmt.Printf("%s", before)
				fmt.Printf("%c", PieceRepresentation[p.PieceAt(Sq(r, f))])
				fmt.Printf("%s", after)
			} else {
				fmt.Printf("%c", PieceRepresentation[p.PieceAt(Sq(r, f))])
			}
			fmt.Printf(" ")
		}
		fmt.Printf(" %d", r+1)
	}
	fmt.Printf("\n-- ")
	for i := 'a'; i <= 'h'; i++ {
		fmt.Printf("%c ", i)
	}
	fmt.Printf("--")
	fmt.Println()
}

// Piece at specified square.
func (p Position) PieceAt(sq Square) Piece {

	//handle kings differently
	if sq == p.status.GetKingPosition(WHITE) {
		return KING
	}
	if sq == p.status.GetKingPosition(BLACK) {
		return -KING
	}

	// Normal pieces
	color := Piece(p.colOcc[WHITE].Get(sq) - p.colOcc[BLACK].Get(sq))
	piece := Piece(p.pawnOcc.Get(sq) | p.knightOcc.Get(sq)<<2 | p.bishopOcc.Get(sq)<<3 | p.rookOcc.Get(sq)<<4)
	return color * piece
}

func (p Position) Dump() {
	fmt.Println("White occ : ", p.colOcc[WHITE].String())
	fmt.Println("Black occ : ", p.colOcc[BLACK].String())
	fmt.Println("Pawn occ : ", p.pawnOcc.String())
	fmt.Println("Knight occ : ", p.knightOcc.String())
	fmt.Println("Bishop occ : ", p.bishopOcc.String())
	fmt.Println("Rook occ : ", p.rookOcc.String())
	fmt.Println("White king sq : ", Bitboard(1<<p.status.GetKingPosition(WHITE)).String())
	fmt.Println("Black king sq : ", Bitboard(1<<p.status.GetKingPosition(BLACK)).String())
	fmt.Printf("Status : %s\n", p.status.String())
}
