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

// materialValue calcule l'avantage matériel du joueur courant.
//
// Barème (en points, 1 pion = 100) :
//   - Pion   : 100
//   - Fou    : 300
//   - Cavalier : 300
//   - Tour   : 500
//   - Dame   : 900
//
// Représentation interne : une dame est stockée à la fois dans rookOcc ET bishopOcc.
// Son score est donc : 500 (tour) + 300 (fou) + 100 (bonus dame) = 900.
func (p Position) materialValue() Score {

	t := p.status.GetTurn()

	s := 100*bits.OnesCount(uint(p.pawnOcc&p.colOcc[t])) - 100*bits.OnesCount(uint(p.pawnOcc&p.colOcc[1^t]))
	s += 300*bits.OnesCount(uint(p.bishopOcc&p.colOcc[t])) - 300*bits.OnesCount(uint(p.bishopOcc&p.colOcc[1^t]))
	s += 300*bits.OnesCount(uint(p.knightOcc&p.colOcc[t])) - 300*bits.OnesCount(uint(p.knightOcc&p.colOcc[1^t]))
	// Les tours pures (rookOcc sans bishopOcc) valent 500.
	// Les dames (rookOcc ∩ bishopOcc) sont comptées ici à 500, puis à 300 (fou), puis +100 ci-dessous → total 900.
	s += 500*bits.OnesCount(uint(p.rookOcc&p.colOcc[t])) - 500*bits.OnesCount(uint(p.rookOcc&p.colOcc[1^t]))
	// Bonus dame : 100 points supplémentaires pour les pièces dans rookOcc ∩ bishopOcc (les dames).
	s += 100*bits.OnesCount(uint(p.rookOcc&p.bishopOcc&p.colOcc[t])) - 100*bits.OnesCount(uint(p.rookOcc&p.bishopOcc&p.colOcc[1^t]))
	return Score(s)
}

// Value évalue la position de manière statique (appelée à profondeur 0 dans l'alpha/beta).
// Retourne le score en "centipions" du point de vue du joueur courant : positif = bon pour lui.
//
// Composantes du score :
//   - WON  si le joueur courant peut capturer le roi adverse (position illégale de l'adversaire)
//   - +1 par droit de roque disponible (mobilité future)
//   - Avantage matériel (voir MaterialValue)
//   - PST score (valeurs de cases spécifiques, pondérées par la phase de la partie)
func (p Position) Value() Score {

	t := p.status.GetTurn()
	s := Score(0)

	// Si l'adversaire a laissé son roi en prise, le joueur courant gagne.
	// (Cela signifie que l'adversaire a joué un coup illégal.)
	if p.IsSquareAttacked(p.status.GetKingPosition(1^t), t) {
		return WON
	}

	// // Malus si le joueur courant est sous échec : la position est dangereuse.
	// if p.IsSquareAttacked(p.status.GetKingPosition(t), 1^t) {
	// 	s += Score(-10)
	// }

	// Bonus pour chaque droit de roque encore disponible (avantage positionnel futur).
	s += Score(bits.OnesCount(uint(p.status.KingStatus[t]&CanCastle)) - bits.OnesCount(uint(p.status.KingStatus[1^t]&CanCastle)))

	// Bonus/malus PST
	s += p.pstValue()

	// Avantage matériel.
	s += p.materialValue()

	return s
}
