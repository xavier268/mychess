package mychess

import (
	"fmt"
	"strings"
)

// Defines a position in time.
type Position struct {
	Board                   [][]int8 // 8 x 8 board with pieces. Ligne, colonne. 0-based index. Black pieces are counted as negative.
	Turn                    int8     // white or black  to play next ?
	EnPassant               Square   // en passant square, or zero-value if no en passant option.
	CanWhiteCastleKingSide  bool
	CanWhiteCastleQueenSide bool
	CanBlackCastleKingSide  bool
	CanBlackCastleQueenSide bool
}

const ( // Black pieces are negative values of white pieces
	EMPTY int8 = iota
	WHITE
	PAWN
	// GHOSTPAWN //  en passant
	KNIGHT
	BISHOP
	ROOK
	QUEEN
	KING
)
const BLACK = -WHITE

// A move. When king moves two squares, it is castling.
// When a pawn moves two squares, it can be captured en passant
type Move struct {
	Piece int8 // this is the arrival value, including potential promotion
	From  Square
	To    Square
}

type Square struct {
	Row int
	Col int
}

// Create empty position, allocating required memory.
func NewPosition() *Position {
	p := new(Position)
	p.Board = make([][]int8, 8)
	// ALlocate board
	for i := 0; i < 8; i++ {
		p.Board[i] = make([]int8, 8)
	}
	p.Turn = WHITE
	p.CanWhiteCastleKingSide = true
	p.CanWhiteCastleQueenSide = true
	p.CanBlackCastleKingSide = true
	p.CanBlackCastleQueenSide = true
	return p
}

// Set the piece, given as KING or -BISHOP, to the locations, specified as A1 or b6 or C4.
// Castle capability unchanged.
func (p *Position) SetPiece(piece int8, where ...string) {
	// Set the piece
	for _, w := range where {
		sq := SquareFromString(w)
		p.Board[sq.Row][sq.Col] = piece
	}
}

func (p *Position) SetNoCastle() {
	p.CanWhiteCastleKingSide = false
	p.CanWhiteCastleQueenSide = false
	p.CanBlackCastleKingSide = false
	p.CanBlackCastleQueenSide = false
}

func SquareFromString(s string) Square {
	if len(s) != 2 {
		panic("Square must be exctly 2 characters")
	}
	s = strings.ToLower(s)
	col := int(s[0] - 'a')
	row := int(s[1] - '1')
	if col < 0 || col > 7 || row < 0 || row > 7 {
		panic("Square out of bounds")
	}
	return Square{row, col}
}

// Reset position to game start position. No allocation is made.
func (p *Position) Reset() *Position {

	// Set pawns
	for j := 0; j < 8; j++ {
		p.Board[1][j] = PAWN
		p.Board[6][j] = -PAWN
	}
	// Set other pieces
	p.Board[0][0] = ROOK
	p.Board[0][1] = KNIGHT
	p.Board[0][2] = BISHOP
	p.Board[0][3] = QUEEN
	p.Board[0][4] = KING
	p.Board[0][5] = BISHOP
	p.Board[0][6] = KNIGHT
	p.Board[0][7] = ROOK
	p.Board[7][0] = -ROOK
	p.Board[7][1] = -KNIGHT
	p.Board[7][2] = -BISHOP
	p.Board[7][3] = -QUEEN
	p.Board[7][4] = -KING
	p.Board[7][5] = -BISHOP
	p.Board[7][6] = -KNIGHT
	p.Board[7][7] = -ROOK

	// Set other values
	p.Turn = WHITE
	p.CanBlackCastleKingSide = true
	p.CanBlackCastleQueenSide = true
	p.CanWhiteCastleKingSide = true
	p.CanWhiteCastleQueenSide = true

	// Set empty squares
	for i := 2; i < 6; i++ {
		for j := 0; j < 8; j++ {
			p.Board[i][j] = EMPTY
		}
	}
	return p
}

var DISPLAY = map[int8]string{
	// White is hollow, black is full
	-WHITE: "♟", WHITE: "♙",
	-PAWN: "♟", PAWN: "♙",
	-KNIGHT: "♞", KNIGHT: "♘",
	-BISHOP: "♝", BISHOP: "♗",
	-ROOK: "♜", ROOK: "♖",
	-QUEEN: "♛", QUEEN: "♕",
	-KING: "♚", KING: "♔",
	EMPTY: " ",
}

func DisplayPiece(piece int8) string {
	return DISPLAY[piece]
}

// Display position as a string.
// Will use special chars to have a nice looking display on a terminal.
func (p *Position) String() string {

	// Display board
	const ll = "   a  b  c  d  e  f  g  h\n"
	const (
		reset   = "\033[0m"
		bgWhite = "\033[47m"
		bgGray  = "\033[100m"
	)

	var buf strings.Builder
	fmt.Fprintln(&buf, reset)
	fmt.Fprintf(&buf, "%s", ll)
	for i := 7; i >= 0; i-- {
		fmt.Fprintf(&buf, "%s %1d", reset, i+1)
		for j := 0; j < 8; j++ {
			if (i+j)%2 != 0 {
				fmt.Fprint(&buf, bgWhite)
			} else {
				fmt.Fprint(&buf, bgGray)
			}
			fmt.Fprintf(&buf, " %s ", DISPLAY[p.Board[i][j]])
		}
		fmt.Fprintf(&buf, "%s %d\n", reset, i+1)
	}
	fmt.Fprintf(&buf, "%s%s", reset, ll)
	return buf.String()
}

// Display move as a string.
func (m *Move) String() string {
	var color string
	if m.Piece >= 0 {
		color = "White"
	} else {
		color = "Black"
	}
	return fmt.Sprintf("%s %s  %s-%s", color, DISPLAY[m.Piece], m.From.String(), m.To.String())
}

// Display square as a string.
func (s *Square) String() string {
	return fmt.Sprintf("%c%c", 'a'+s.Col, '1'+s.Row)
}

// Clone a position
func (p *Position) Clone() *Position {
	clone := NewPosition()
	clone.CopyFrom(p)
	return clone
}

// Copy p2 into p. p2 is unchanged.
func (p *Position) CopyFrom(p2 *Position) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			p.Board[i][j] = p2.Board[i][j]
		}
	}
	p.Turn = p2.Turn
	p.CanBlackCastleKingSide = p2.CanBlackCastleKingSide
	p.CanBlackCastleQueenSide = p2.CanBlackCastleQueenSide
	p.CanWhiteCastleKingSide = p2.CanWhiteCastleKingSide
	p.CanWhiteCastleQueenSide = p2.CanBlackCastleQueenSide
}

// Execute move on current position. No cloning, no allocation.
func (pos *Position) ExecuteMove(m Move) {

	pos.Board[m.To.Row][m.To.Col] = m.Piece
	pos.Board[m.From.Row][m.From.Col] = EMPTY

	// Use existing en passant value to capture the pawn if we targeted the en passant square
	if pos.EnPassant != (Square{}) && m.To == pos.EnPassant {
		if pos.Turn == WHITE {
			pos.Board[4][m.To.Col] = EMPTY // capture the BLACK pawn
		} else {
			pos.Board[3][m.To.Col] = EMPTY // capture the white pawn
		}
	}

	// Now, set a new en passant value, if required
	pos.EnPassant = Square{}
	if m.Piece == PAWN && m.From.Row == 1 && m.To.Row == 3 {
		// set en passant
		pos.EnPassant = Square{2, m.To.Col}
	}
	if m.Piece == -PAWN && m.From.Row == 6 && m.To.Row == 4 {
		// set en passant
		pos.EnPassant = Square{5, m.To.Col}
	}

	// handle castling white and inhibiting when king moves
	if m.Piece == KING {
		if pos.CanWhiteCastleKingSide && m.From == (Square{0, 4}) && m.To == (Square{0, 6}) {
			// white king side
			pos.Board[0][5] = ROOK
			pos.Board[0][7] = EMPTY
		}
		if pos.CanWhiteCastleQueenSide && m.From == (Square{0, 4}) && m.To == (Square{0, 2}) {
			// white queen side
			pos.Board[0][3] = ROOK
			pos.Board[0][0] = EMPTY
		}

		// Inhibit castling when white king moves
		pos.CanWhiteCastleKingSide = false
		pos.CanWhiteCastleQueenSide = false
	}

	// handle castling black and inhibiting when king moves
	if m.Piece == -KING {
		if pos.CanBlackCastleKingSide && m.From == (Square{7, 4}) && m.To == (Square{7, 6}) {
			// black king side
			pos.Board[7][5] = -ROOK
			pos.Board[7][7] = EMPTY
		}
		if pos.CanBlackCastleQueenSide && m.From == (Square{7, 4}) && m.To == (Square{7, 2}) {
			// black queen side
			pos.Board[7][3] = -ROOK
			pos.Board[7][0] = EMPTY
		}

		// Inhibit castling when black king moves
		pos.CanBlackCastleKingSide = false
		pos.CanBlackCastleQueenSide = false
	}

	// Inhibit castling when rook moves
	if m.Piece == ROOK {
		if m.From == (Square{0, 0}) {
			pos.CanWhiteCastleQueenSide = false
		}
		if m.From == (Square{0, 7}) {
			pos.CanWhiteCastleKingSide = false
		}
	}
	if m.Piece == -ROOK {
		if m.From == (Square{7, 0}) {
			pos.CanBlackCastleQueenSide = false
		}
		if m.From == (Square{7, 7}) {
			pos.CanBlackCastleKingSide = false
		}
	}

	// change turn
	pos.Turn = -pos.Turn

}

// Convert color code or piece into string (White / Black)
func StringColor(color int8) string {
	if color > 0 {
		return "White"
	} else {
		return "Black"
	}
}
