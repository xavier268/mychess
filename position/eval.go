package position

import "math/bits"

// Score d'une position du point de vue du joueur qui doit jouer (turn player).
// Plus le score est élevé, meilleure est la position pour ce joueur.
type Score int16

var (
	WON  Score = 30_000  // Le joueur courant a gagné (roi adverse prenable).
	LOST Score = -30_000 // Le joueur courant est mat.
	DRAW Score = 0       // Partie nulle (pat, répétition, etc.).
)

// MaterialValue calcule l'avantage matériel du joueur courant.
//
// Barème (en points, 1 pion = 10) :
//   - Pion   : 10
//   - Fou    : 30
//   - Cavalier : 30
//   - Tour   : 50
//   - Dame   : 90
//
// Représentation interne : une dame est stockée à la fois dans rookOcc ET bishopOcc.
// Son score est donc : 50 (tour) + 30 (fou) + 10 (bonus dame) = 90.
func (p Position) MaterialValue() Score {

	t := p.status.GetTurn()

	s := 10*bits.OnesCount(uint(p.pawnOcc&p.colOcc[t])) - 10*bits.OnesCount(uint(p.pawnOcc&p.colOcc[1^t]))
	s += 30*bits.OnesCount(uint(p.bishopOcc&p.colOcc[t])) - 30*bits.OnesCount(uint(p.bishopOcc&p.colOcc[1^t]))
	s += 30*bits.OnesCount(uint(p.knightOcc&p.colOcc[t])) - 30*bits.OnesCount(uint(p.knightOcc&p.colOcc[1^t]))
	// Les tours pures (rookOcc sans bishopOcc) valent 50.
	// Les dames (rookOcc ∩ bishopOcc) sont comptées ici à 50, puis à 30 (fou), puis +10 ci-dessous → total 90.
	s += 50*bits.OnesCount(uint(p.rookOcc&p.colOcc[t])) - 50*bits.OnesCount(uint(p.rookOcc&p.colOcc[1^t]))
	// Bonus dame : 10 points supplémentaires pour les pièces dans rookOcc ∩ bishopOcc (les dames).
	s += 10*bits.OnesCount(uint(p.rookOcc&p.bishopOcc&p.colOcc[t])) - 10*bits.OnesCount(uint(p.rookOcc&p.bishopOcc&p.colOcc[1^t]))
	return Score(s)
}

// IsCheck retourne true si le joueur courant est en échec.
func (p Position) IsCheck() bool {
	t := p.status.GetTurn()
	return p.IsSquareAttacked(p.status.GetKingPosition(t), 1^t)
}

// Turn retourne le camp qui doit jouer : WHITE (0) ou BLACK (1).
func (p Position) Turn() uint8 {
	return p.status.GetTurn()
}

// KingPosition retourne la case du roi du camp `side` (WHITE ou BLACK).
func (p Position) KingPosition(side uint8) Square {
	return p.status.GetKingPosition(side)
}

// Value évalue la position de manière statique (appelée à profondeur 0 dans l'alpha/beta).
// Retourne le score du point de vue du joueur courant : positif = bon pour lui.
//
// Composantes du score :
//   - WON  si le joueur courant peut capturer le roi adverse (position illégale de l'adversaire)
//   - Malus de 10 si le joueur courant est en échec
//   - +1 par droit de roque disponible (mobilité future)
//   - +1 par case du centre occupée (d4, d5, e4, e5)
//   - Avantage matériel (voir MaterialValue)
func (p Position) Value() Score {

	t := p.status.GetTurn()
	s := Score(0)

	// Si l'adversaire a laissé son roi en prise, le joueur courant gagne.
	// (Cela signifie que l'adversaire a joué un coup illégal.)
	if p.IsSquareAttacked(p.status.GetKingPosition(1^t), t) {
		return WON
	}

	// Malus si le joueur courant est sous échec : la position est dangereuse.
	if p.IsSquareAttacked(p.status.GetKingPosition(t), 1^t) {
		s += Score(-10)
	}

	// Bonus pour chaque droit de roque encore disponible (avantage positionnel futur).
	s += Score(bits.OnesCount(uint(p.status.KingStatus[t]&CanCastle)) - bits.OnesCount(uint(p.status.KingStatus[1^t]&CanCastle)))

	// Bonus pour chaque case centrale occupée (d4, d5, e4, e5).
	s += Score(bits.OnesCount(uint(p.colOcc[t]&Center())) - bits.OnesCount(uint(p.colOcc[1^t]&Center())))

	// Avantage matériel.
	s += p.MaterialValue()

	return s
}
