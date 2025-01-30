package eval

import (
	"math"
	"math/rand/v2"
)

type Piece = int8

var (
	BESTVALUE  = math.Inf(1)
	WORSTVALUE = math.Inf(-1)
)

// Basic evaluation of a position, by setting a value for each piece.
// Counted from the point of view of the player who has to play now.
// Positive = better.
// A small random value is always added for non predictibility.
func basicEval(p *Position) float64 {
	var v float64
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			v += pieceValue(p.Board[i][j])
		}
	}
	if p.CanWhiteCastleKingSide {
		v += 0.5
	}
	if p.CanWhiteCastleQueenSide {
		v += 0.5
	}
	if p.CanBlackCastleKingSide {
		v -= 0.5
	}
	if p.CanBlackCastleQueenSide {
		v -= 0.5
	}
	v += 0.0001 * (rand.Float64() - 0.5)
	return v * float64(p.Turn)
}

func pieceValue(piece Piece) float64 {
	switch piece {
	case EMPTY:
		return 0
	case PAWN:
		return 1
	case -PAWN:
		return -1
	case KNIGHT:
		return 3
	case -KNIGHT:
		return -3
	case BISHOP:
		return 3
	case -BISHOP:
		return -3
	case ROOK:
		return 5
	case -ROOK:
		return -5
	case QUEEN:
		return 9
	case -QUEEN:
		return -9
	case KING:
		return 100
	case -KING:
		return -100
	default:
		panic("invalid piece")
	}
}
