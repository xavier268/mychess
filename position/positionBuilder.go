package position

import "fmt"

// Utilities to create/modify specific positions
// Positions are specified as strings : a1 or C4 ...

// Side is WHITE or BLACK only

func (p *Position) AddPawn(side uint8, pawns ...string) *Position {
	side = side & 1
	for _, s := range pawns {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.pawnOcc |= 1 << sq
	}
	return p
}

func (p *Position) AddKnight(side uint8, knights ...string) *Position {
	side = side & 1
	for _, s := range knights {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.knightOcc |= 1 << sq
	}
	return p
}

func (p *Position) AddBishop(side uint8, bishops ...string) *Position {
	side = side & 1
	for _, s := range bishops {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.bishopOcc |= 1 << sq
	}
	return p
}

func (p *Position) AddRook(side uint8, rooks ...string) *Position {
	side = side & 1
	for _, s := range rooks {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.rookOcc |= 1 << sq
	}
	return p
}

func (p *Position) AddQueen(side uint8, queens ...string) *Position {
	side = side & 1
	for _, s := range queens {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.rookOcc |= 1 << sq
		p.bishopOcc |= 1 << sq
	}
	return p
}

// ... and sets castling rights to NONE !
func (p *Position) AddKing(side uint8, kingsq string) *Position {
	side = side & 1
	sq := uint8(SqParse(kingsq))
	p.colOcc[side] |= 1 << sq
	p.status.KingStatus[side] = sq
	return p
}

// SetEnPassant records an en passant opportunity for the given side.
// Only the file of where is used. The phantom is placed at rank 2 (white's
// double push) or rank 5 (black's double push) — the capture landing square.
func (p *Position) SetEnPassant(side uint8, where string) *Position {
	side = side & 1
	file := SqParse(where).File()
	p.pawnOcc |= 1 << (Sq(2+3*int(side), file))
	return p
}

// Add specify castling rights
func (p *Position) SetCastle(side uint8, castleBits uint8) *Position {
	side = side & 1
	p.status.KingStatus[side] |= (castleBits & CanCastle)
	return p
}

// SetTurn sets whose turn it is (WHITE or BLACK).
func (p *Position) SetTurn(turn uint8) {
	p.status.SetTurn(turn)
}

// ParseFEN parses the piece-placement section of a FEN string and returns a
// Position with the pieces placed. Hash, turn, castling rights, and en passant
// are all zero/cleared; the caller must set them and call
// DefaultZT.HashPosition to obtain a consistent hash.
// Returns an error if the FEN contains an unrecognised character.
func ParseFEN(fen string) (Position, error) {
	var p Position
	rank := 7
	file := 0
	for _, ch := range fen {
		switch {
		case ch == '/':
			rank--
			file = 0
			if rank < 0 {
				return Position{}, fmt.Errorf("ParseFEN: too many ranks in %q", fen)
			}
		case ch >= '1' && ch <= '8':
			file += int(ch - '0')
		default:
			if file > 7 || rank < 0 {
				return Position{}, fmt.Errorf("ParseFEN: square out of range at %c", ch)
			}
			sq := Sq(rank, file)
			switch ch {
			case 'P':
				p.AddPawn(WHITE, sq.String())
			case 'N':
				p.AddKnight(WHITE, sq.String())
			case 'B':
				p.AddBishop(WHITE, sq.String())
			case 'R':
				p.AddRook(WHITE, sq.String())
			case 'Q':
				p.AddQueen(WHITE, sq.String())
			case 'K':
				p.AddKing(WHITE, sq.String())
			case 'p':
				p.AddPawn(BLACK, sq.String())
			case 'n':
				p.AddKnight(BLACK, sq.String())
			case 'b':
				p.AddBishop(BLACK, sq.String())
			case 'r':
				p.AddRook(BLACK, sq.String())
			case 'q':
				p.AddQueen(BLACK, sq.String())
			case 'k':
				p.AddKing(BLACK, sq.String())
			default:
				return Position{}, fmt.Errorf("ParseFEN: unknown character %q", ch)
			}
			file++
		}
	}
	return p, nil
}
