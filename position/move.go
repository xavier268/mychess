package position

import (
	"fmt"
	"slices"
)

type Move struct {
	From, To Square
	// Quelle piece si promotion ? EMPTY si pas de promotion.
	Promotion Piece
	// Used to rank moves to select the one to explore first
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
	// order by decreasing score
	slices.SortFunc(moves, func(a, b Move) int { return int(b.Score) - int(a.Score) })
	return moves
}
