package position

import "math/bits"

// Score of a given position, from the point of view of the player who should play next (turn player)
// The larger, the better.
type Score int32

var (
	WON  Score = 100_000_000  // the player who is expected to play has already won.
	LOST Score = -100_000_000 // The player who is about to play has already lost. Check mate.
	DRAW Score = 0
)

// La valeur des pièces de la position.
// Du point de vue de celui qui va jouer (le turn)
// Le pion (100 point) ; le fou (300 points) ;
// le cavalier (300 points) ; la tour (500 points) ;
// la dame (900 points).
func (p Position) MaterialValue() Score {

	t := p.status.GetTurn()
	s := 100*bits.OnesCount(uint(p.pawnOcc&p.colOcc[t])) - 100*bits.OnesCount(uint(p.pawnOcc&p.colOcc[1^t]))
	s += 300*bits.OnesCount(uint(p.bishopOcc&p.colOcc[t])) - 300*bits.OnesCount(uint(p.bishopOcc&p.colOcc[1^t]))
	s += 300*bits.OnesCount(uint(p.knightOcc&p.colOcc[t])) - 300*bits.OnesCount(uint(p.knightOcc&p.colOcc[1^t]))
	s += 300*bits.OnesCount(uint(p.rookOcc&p.colOcc[t])) - 300*bits.OnesCount(uint(p.rookOcc&p.colOcc[1^t]))
	s += 100*bits.OnesCount(uint(p.rookOcc&p.bishopOcc&p.colOcc[t])) - 100*bits.OnesCount(uint(p.rookOcc&p.bishopOcc&p.colOcc[1^t])) // dame
	return Score(s)
}

// MaterialValue + some bonus points for checks, center occupancy, castling capabilities, ...
func (p Position) Value() Score {

	t := p.status.GetTurn()
	s := Score(0)

	// If I am about to play and could capture opponent king, I won !
	if p.IsSquareAttacked(p.status.GetKingPosition(t)) {
		return WON
	}

	// If I am about to play and I am currently under check, add 100 malus points
	if p.IsSquareAttacked(p.status.GetKingPosition(1 ^ t)) {
		s += Score(-100)
	}

	// add 3 points for each castling capabilities
	s += Score(3*bits.OnesCount(uint(p.status.KingStatus[t]&CanCastle)) - 3*bits.OnesCount(uint(p.status.KingStatus[1^t]&CanCastle)))

	// add 10 points for each square occupying the center
	s += Score(10*bits.OnesCount(uint(p.colOcc[t]&Center())) - 10*bits.OnesCount(uint(p.colOcc[1^t]&Center())))

	// add material value
	s += p.MaterialValue()

	return s
}
