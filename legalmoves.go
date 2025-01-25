package mychess

// Generate all legal moves from position. A move slice is provided, avoiding allocation as much as possible.
// If it is nil, a new slice will be allocated.
func (pos *Position) LegalMoves(moves []Move) []Move {
	if pos.Draw || pos.StaleMate {
		return nil
	}
	if moves == nil {
		moves = make([]Move, 0, 40)
	} else {
		moves = moves[:0]
	}

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := pos.Board[i][j]
			if pos.Turn*piece <= 0 {
				continue
			}
			switch piece {
			case PAWN, -PAWN:
				moves = pawnLegalMoves(pos, moves)
			case KNIGHT, -KNIGHT:
				moves = knightLegalMoves(pos, moves)
			case BISHOP, -BISHOP:
				moves = bishopLegalMoves(pos, moves)
			case ROOK, -ROOK:
				moves = rookLegalMoves(pos, moves)
			case QUEEN, -QUEEN:
				moves = queenLegalMoves(pos, moves)
			case KING, -KING:
				moves = kingLegalMoves(pos, moves)
			}
		}
	}
	return moves
}

func rookLegalMoves(pos *Position, moves []Move) []Move {
	panic("todo")
}

func bishopLegalMoves(pos *Position, moves []Move) []Move {
	panic("todo")
}

func queenLegalMoves(pos *Position, moves []Move) []Move {
	panic("todo")
}
func kingLegalMoves(pos *Position, moves []Move) []Move {
	panic("todo")
}
func knightLegalMoves(pos *Position, moves []Move) []Move {
	panic("todo")
}
func pawnLegalMoves(pos *Position, moves []Move) []Move {
	panic("todo")
}
