package position

import "math/bits"

// Score of a given position, from the point of view of the player who should play next (turn player)
// The larger, the better.
type Score int16

var (
	WON  Score = 30_000  // the player who is expected to play has already won.
	LOST Score = -30_000 // The player who is about to play has already lost. Check mate.
	DRAW Score = 0
)

// La valeur des pièces de la position.
// Du point de vue de celui qui va jouer (le turn)
// Le pion (10 point) ; le fou (30 points) ;
// le cavalier (30 points) ; la tour (50 points) ;
// la dame (90 points).
func (p Position) MaterialValue() Score {

	t := p.status.GetTurn()
	s := 10*bits.OnesCount(uint(p.pawnOcc&p.colOcc[t])) - 10*bits.OnesCount(uint(p.pawnOcc&p.colOcc[1^t]))
	s += 30*bits.OnesCount(uint(p.bishopOcc&p.colOcc[t])) - 30*bits.OnesCount(uint(p.bishopOcc&p.colOcc[1^t]))
	s += 30*bits.OnesCount(uint(p.knightOcc&p.colOcc[t])) - 30*bits.OnesCount(uint(p.knightOcc&p.colOcc[1^t]))
	s += 30*bits.OnesCount(uint(p.rookOcc&p.colOcc[t])) - 30*bits.OnesCount(uint(p.rookOcc&p.colOcc[1^t]))
	s += 10*bits.OnesCount(uint(p.rookOcc&p.bishopOcc&p.colOcc[t])) - 10*bits.OnesCount(uint(p.rookOcc&p.bishopOcc&p.colOcc[1^t])) // dame
	return Score(s)
}

// MaterialValue + some bonus points for checks, center occupancy, castling capabilities, ...
func (p Position) Value() Score {

	t := p.status.GetTurn()
	s := Score(0)

	// If I am about to play and could capture opponent king, I won !
	if p.IsSquareAttacked(p.status.GetKingPosition(1^t), t) {
		return WON
	}

	// If I am about to play and I am currently under check by opponent, add 10 malus points
	if p.IsSquareAttacked(p.status.GetKingPosition(t), 1^t) {
		s += Score(-10)
	}

	// add 1 points for each castling capabilities
	s += Score(bits.OnesCount(uint(p.status.KingStatus[t]&CanCastle)) - bits.OnesCount(uint(p.status.KingStatus[1^t]&CanCastle)))

	// add 1 points for each square occupying the center
	s += Score(bits.OnesCount(uint(p.colOcc[t]&Center())) - bits.OnesCount(uint(p.colOcc[1^t]&Center())))

	// add material value
	s += p.MaterialValue()

	return s
}
