package position

import "math/bits"

// Calcule la phase du jeu : 24 = début de partie, 0 = fin de partie
// Compte le nombre de pièces (hors pions et roi) sur l'échiquier, quelle que soient leurs couleurs.
// Comptage des points de phase :
// 1 pour fou
// 1 pour cavalier
// 2 pour tour
// 4 pour reine (bonus de 1 par rapport à fou + tour).
func (p Position) Phase() int {
	return bits.OnesCount(uint(p.knightOcc)) +
		bits.OnesCount(uint(p.bishopOcc)) +
		2*bits.OnesCount(uint(p.rookOcc)) +
		bits.OnesCount(uint(p.rookOcc&p.bishopOcc))
}
