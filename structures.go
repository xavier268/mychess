package mychess

import (
	"fmt"
	"strings"
)

// Defines a position in time.
type Position struct {
	Board          [][]int8 // 8 x 8 board with pieces. Ligne, colonne. 0-based index. Black pieces are counted as negative.
	WhiteKing      Square   // Where is the white king
	BlackKing      Square   // Where is the black king
	WhiteKingMoved bool     // Did the king moved already (cannot castle any more)
	BlackKingMoved bool     // Did the king moved already (cannot castle any more)
	History        []Move   // History of moves from start of game
	Value          float64  // positive means favorable for white, negative for black
	Turn           int8     // white or black  to play next ?
	Draw           bool     // true if position is a draw
	StaleMate      bool     // true if position is a draw - the player which should play now lost.
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
	p.History = make([]Move, 0, 50)
	return p
}

// Set the piece, given as KING or -BISHOP, to the locations, specified as A1 or b6 or C4.
// Useful for debugging.
func (p *Position) SetPiece(piece int8, where ...string) {
	// Set the piece
	for _, w := range where {
		sq, err := SquareFromString(w)
		if err != nil {
			fmt.Println("Ignoring coordinates : ", w, "because", err)
			continue
		}
		p.Board[sq.Row][sq.Col] = piece
		if piece == KING {
			p.WhiteKing = sq
		}
		if piece == -KING {
			p.BlackKing = sq
		}
	}
}

func SquareFromString(s string) (Square, error) {
	if len(s) != 2 {
		return Square{}, fmt.Errorf("Square must be exctly 2 characters")
	}
	s = strings.ToLower(s)
	col := int(s[0] - 'a')
	row := int(s[1] - '1')
	if col < 0 || col > 7 || row < 0 || row > 7 {
		return Square{}, fmt.Errorf("Square out of bounds")
	}
	return Square{row, col}, nil
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
	p.History = p.History[:0]
	p.BlackKing = Square{7, 4}
	p.WhiteKing = Square{0, 4}
	p.BlackKingMoved = false
	p.WhiteKingMoved = false
	p.Turn = WHITE
	p.Draw = false
	p.StaleMate = false
	p.Value = 0.0
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
		// fgBlack = "\033[30m"
		// fgWhite = "\033[97m"
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
	p.WhiteKing = p2.WhiteKing
	p.BlackKing = p2.BlackKing
	p.WhiteKingMoved = p2.WhiteKingMoved
	p.BlackKingMoved = p2.BlackKingMoved
	p.History = append(p.History[:0], p2.History...)
	p.Value = p2.Value
	p.Turn = p2.Turn
	p.Draw = p2.Draw
	p.StaleMate = p2.StaleMate
}

// Execute move on current position. No cloning, no allocation.
func (pos *Position) ExecuteMove(m Move) {

	// TODO castling !

	// TODO EN passant !

	pos.Board[m.To.Row][m.To.Col] = m.Piece
	pos.Board[m.From.Row][m.From.Col] = EMPTY
	if m.Piece == KING {
		pos.WhiteKing = m.To
		pos.WhiteKingMoved = true
	}
	if m.Piece == -KING {
		pos.BlackKing = m.To
		pos.BlackKingMoved = true
	}
	// Update history
	pos.History = append(pos.History, m)
	// change turn
	pos.Turn = -pos.Turn

}

func StringColor(color int8) string {
	if color > 0 {
		return "White"
	} else {
		return "Black"
	}
}
