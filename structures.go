package mychess

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
	GHOSTPAWN //  en passant
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
	Piece int8
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
	return p
}

// Reset position. No allocation is made.
func (p *Position) Reset() {

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
}
