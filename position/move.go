package position

import (
	"fmt"
	"slices"
)

// Flag to signal a castle move.
const CASTLEMOVE Piece = KING + ROOK

type Move struct {
	From, To Square
	// Used to mark special moves.
	// Always positive, even for BLACK
	// Flag a castle move with CASTLEMOVE
	// Flag a promotion with KNIGHT, QUEEN, ROOK or BISHOP
	Promotion Piece
	// Used to rank moves to select which one to explore first
	Score uint8
}

func (m Move) String() string {
	if m.Promotion == EMPTY {
		return fmt.Sprintf("%s - %s (%d)", m.From.String(), m.To.String(), m.Score)
	} else {
		return fmt.Sprintf("%s - %s (%d) => %s", m.From.String(), m.To.String(), m.Score, m.Promotion.String())
	}
}

// Illegal moves (self checks) are not yet removed.
// Promotion is managed here.
func (p Position) GetMoveList(bt *BigTable) []Move {
	var moves []Move = make([]Move, 0, 32)

	turn := p.status.GetTurn() // 0 : white, 1 black
	for fromSq := range p.colOcc[turn].AllSetSquares {
		bb := p.GetMovesBB(bt, fromSq)
		for toSq := range bb.AllSetSquares {

			// better score for bigger capture
			sc := uint8(p.rookOcc.Get(toSq)*4 + p.bishopOcc.Get(toSq)*3 + p.knightOcc.Get(toSq)*3 + p.pawnOcc.Get(toSq)*2) // sc max 4

			// handle promotion moves
			if p.pawnOcc.Get(toSq) != 0 && toSq.Rank() == int((1-turn)*7) {
				for _, piece := range []Piece{QUEEN, ROOK, BISHOP, KNIGHT} {
					// improve score for promotion
					moves = append(moves, Move{fromSq, toSq, piece, sc + 2*uint8(piece)}) // sc max 14
				}
			} else { // no promotion
				moves = append(moves, Move{fromSq, toSq, EMPTY, sc})
			}
		}
	}
	// append the castling moves, if any ?
	moves = append(moves, p.GetCastlingMoveList(bt)...)

	// order by decreasing score
	slices.SortFunc(moves, func(a, b Move) int { return int(b.Score) - int(a.Score) })
	return moves
}

// Verifie que le roi n'est pas en echec, que les cases intermédiaires sont vides,
// que le roi ne va pas passer sur une case en echec, que le roque reste possible.
func (p Position) GetCastlingMoveList(bt *BigTable) []Move {
	turn := p.status.GetTurn()
	cb := p.status.GetCastleBits(turn)
	if cb == 0 || p.IsSquareAttacked(bt, p.status.GetKingPosition(turn), 1-turn) {
		return nil
	}
	// Both king and rook are supposed to be at the correct position, since the castle bits are set.
	// We do not verify this...
	var moves []Move = make([]Move, 0, 4)
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]

	if turn == WHITE { // WHITE
		if (cb&CanCastleKingSide != 0) &&
			((occ&(1<<5) | (1 << 6)) == 0) &&
			!p.IsSquareAttacked(bt, 5, BLACK) && !p.IsSquareAttacked(bt, 6, BLACK) {
			moves = append(moves, Move{4, 6, CASTLEMOVE, 1}) // sc 1
		}
		if (cb&CanCastleQueenSide != 0) &&
			((occ&(1<<1) | (1 << 2) | (1 << 3)) == 0) &&
			!p.IsSquareAttacked(bt, 3, BLACK) && !p.IsSquareAttacked(bt, 2, BLACK) {
			moves = append(moves, Move{4, 2, CASTLEMOVE, 1}) // sc 1
		}

	} else { // BLACK
		if (cb&CanCastleKingSide != 0) &&
			((occ&(1<<61) | (1 << 62)) == 0) &&
			!p.IsSquareAttacked(bt, 61, WHITE) && !p.IsSquareAttacked(bt, 62, WHITE) {
			moves = append(moves, Move{60, 62, CASTLEMOVE, 1}) // sc 1
		}
		if (cb&CanCastleQueenSide != 0) &&
			((occ&(1<<57) | (1 << 58) | (1 << 59)) == 0) &&
			!p.IsSquareAttacked(bt, 59, WHITE) && !p.IsSquareAttacked(bt, 58, WHITE) {
			moves = append(moves, Move{60, 58, CASTLEMOVE, 1}) // sc 1
		}
	}

	return moves
}

// Generate error if move is illegal.
// Status is updated.
// TODO - manage pawn advancing 2 squres, manage castling, reconsider how to store en passant ?
func (p Position) DoMove(m Move) (pp Position, err error) {
	pp = p
	turn := p.status.GetTurn() // 0 : white, 1 black
	newturn := 1 - turn
	pp.status.TurnStatus = newturn

	switch m.Promotion {
	case EMPTY: // Normal move
		// Erase start position
		pp.colOcc[turn] &= ^(1 << m.From)
		pp.bishopOcc &= ^(1 << m.From)
		pp.knightOcc &= ^(1 << m.From)
		pp.rookOcc &= ^(1 << m.From)
		pp.pawnOcc &= ^(1 << m.From)
	}

	panic("todo")

}
