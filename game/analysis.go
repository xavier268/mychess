package game

import "mychess/position"

func (g *Game) AlphaBeta(alpha, beta position.Score, depth int16) position.Score {

	// alpha : le meilleur score que moi, le joueur qui doit jouer, a réussi à garantir jusque là.
	// beta : le pire score que MON ADVERSAIRE, qui vient de jouer, peut m'imposer.

	hash := g.Position.Hash
	oldAlpha := alpha

	// 1: Verification de la table Z
	entry, found := g.Z[hash]
	if found && entry.Depth >= depth {
		switch entry.ScoreType {
		case EXACT:
			return entry.Score
		case LOWER:
			alpha = max(alpha, entry.Score)
		case UPPER:
			beta = min(beta, entry.Score)
		}

		if alpha >= beta { // Coupure, inutile de regarder cette branche que l'adversaire ne choisira jamais !
			return entry.Score
		}
	}

	// 2. CONDITION D'ARRÊT
	if depth == 0 {
		return g.Position.Value() // Valeur intrinsèque de la position !
	}

	// 3. RECHERCHE DES COUPS
	moves := g.Position.GetMoveList()
	// mettre entry.Best en premier, s'il y en a un ?
	if (found && entry.Best != position.Move{}) {
		for i, move := range moves {
			if move == entry.Best {
				moves[0], moves[i] = moves[i], moves[0]
				break
			}
		}
	}

	bestScore := position.LOST - 1
	var bestMove position.Move

	for _, move := range moves {
		g.Position, move = g.Position.DoMove(move)
		// Note le signe '-' et l'inversion alpha/beta (NegaMax)
		score := -g.AlphaBeta(-beta, -alpha, depth-1)
		g.Position.UndoMove(move)

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		alpha = max(alpha, score)
		if alpha >= beta {
			break // Coupure Beta
		}
	}

	// 4. SAUVEGARDE DANS LA TABLE (STORE)
	newEntry := ZEntry{
		Score: position.Score(bestScore),
		Best:  bestMove,
		Depth: depth,
		Age:   uint8(len(g.History)),
	}

	if bestScore <= oldAlpha {
		newEntry.ScoreType = UPPER
	} else if bestScore >= beta {
		newEntry.ScoreType = LOWER
	} else {
		newEntry.ScoreType = EXACT
	}

	g.Z[hash] = newEntry

	return bestScore
}
