package game

import "mychess/position"

// AlphaBeta implémente l'algorithme Negamax avec élagage alpha-bêta et table de transposition.
//
// Convention Negamax : le score est TOUJOURS du point de vue du joueur courant.
// Un score positif signifie un avantage pour le joueur qui doit jouer.
//
// Paramètres :
//   - alpha : meilleur score que le joueur courant peut déjà garantir (borne basse).
//   - beta  : meilleur score que l'adversaire peut déjà garantir (borne haute).
//     Si le score retourné >= beta, l'adversaire ignorera cette branche (coupure bêta).
//   - depth : profondeur restante à analyser (0 = évaluation statique).
//
// Interruption : AlphaBeta consulte g.Ctx à chaque nœud. Dès que le contexte est annulé,
// il retourne 0 immédiatement et remonte la pile sans écrire dans Z, laissant la table
// dans l'état cohérent issu de la dernière profondeur entièrement terminée.
func (g *Game) AlphaBeta(alpha, beta position.Score, depth int16) position.Score {

	// ── VÉRIFICATION DU CONTEXTE ───────────────────────────────────────────────
	// ctx.Err() est goroutine-safe et retourne immédiatement après annulation
	// (valeur mise en cache, lecture atomique ~5 ns). On peut l'appeler à chaque nœud
	// sans surcoût mesurable. Pas besoin d'un flag intermédiaire.
	if g.Ctx.Err() != nil {
		return 0 // on remonte sans toucher à Z
	}

	hash := g.Position.Hash
	// oldAlpha capture la fenêtre initiale avant tout ajustement par la table de transposition.
	// Il sera utilisé à la fin pour déterminer le type de borne (UPPER/LOWER/EXACT).
	oldAlpha := alpha

	// ── 1. CONSULTATION DE LA TABLE DE TRANSPOSITION ──────────────────────────
	// Si cette position a déjà été analysée à une profondeur suffisante,
	// on peut réutiliser ou affiner le résultat sans refaire la recherche complète.
	entry, found := g.Z[hash]
	if found && entry.Depth >= depth {
		switch entry.ScoreType {
		case EXACT:
			// Valeur exacte connue : on retourne directement, pas besoin de chercher.
			return entry.Score
		case LOWER:
			// Borne basse : le vrai score est >= entry.Score.
			// On relève alpha sans perdre d'information.
			alpha = max(alpha, entry.Score)
		case UPPER:
			// Borne haute : le vrai score est <= entry.Score.
			// On abaisse beta sans perdre d'information.
			beta = min(beta, entry.Score)
		}
		if alpha >= beta {
			// Fenêtre [alpha, beta] vide : l'adversaire a déjà une meilleure alternative.
			return entry.Score
		}
	}

	// ── 2. CAS DE BASE : profondeur 0 → évaluation statique ───────────────────
	if depth == 0 {
		return g.Position.Value()
	}

	// ── 3. GÉNÉRATION DES COUPS ────────────────────────────────────────────────
	moves := g.Position.GetMoveList()

	// CAS TERMINAL : aucun coup légal disponible.
	// Il faut distinguer l'échec et mat du pat, car les scores sont différents.
	if len(moves) == 0 {
		if g.Position.IsCheck() {
			// Échec et mat : le joueur courant a perdu.
			return position.LOST
		}
		// Pat : aucun coup mais le roi n'est pas en échec → partie nulle.
		return position.DRAW
	}

	// OPTIMISATION — Mise en tête du meilleur coup issu de la table de transposition.
	// Les coupures bêta sont trouvées plus tôt si on évalue d'abord le meilleur coup connu,
	// ce qui réduit considérablement le nombre de nœuds à explorer.
	if found && (entry.Best != position.Move{}) {
		for i, move := range moves {
			if move == entry.Best {
				moves[0], moves[i] = moves[i], moves[0]
				break
			}
		}
	}

	bestScore := position.LOST - 1 // sentinelle : pire que tout score réel
	var bestMove position.Move

	for _, move := range moves {
		g.Position, move = g.Position.DoMove(move)

		// Negamax : on appelle récursivement pour l'adversaire.
		// La fenêtre est inversée et niée : [alpha, beta] → [-beta, -alpha].
		// Le score retourné est du point de vue de l'adversaire, on le nie pour l'avoir du nôtre.
		score := -g.AlphaBeta(-beta, -alpha, depth-1)

		// IMPORTANT : UndoMove retourne la nouvelle position restaurée — il faut capturer le résultat.
		// (Position est un type valeur en Go ; ignorer le retour laisserait g.Position inchangée.)
		g.Position = g.Position.UndoMove(move)

		// Si le contexte a expiré pendant l'appel récursif, on remonte sans stocker.
		// Le résultat partiel (certains coups explorés, d'autres non) ne doit pas polluer Z.
		if g.Ctx.Err() != nil {
			return 0
		}

		if score > bestScore {
			bestScore = score
			bestMove = move
		}
		alpha = max(alpha, score)
		if alpha >= beta {
			// Coupure bêta : l'adversaire dispose déjà d'une meilleure alternative,
			// il ne laissera jamais cette position se produire → on arrête la recherche.
			break
		}
	}

	// ── 4. SAUVEGARDE DANS LA TABLE DE TRANSPOSITION ──────────────────────────
	// Cette section n'est atteinte que si toute la boucle s'est terminée sans interruption.
	//
	// Type de borne selon la relation entre bestScore et la fenêtre initiale [oldAlpha, beta] :
	//   UPPER : tous les coups sont sous oldAlpha  → borne haute (vrai score ≤ bestScore)
	//   LOWER : un coup a causé une coupure bêta   → borne basse (vrai score ≥ bestScore)
	//   EXACT : score compris dans la fenêtre       → valeur exacte
	newEntry := ZEntry{
		Score: bestScore,
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

	// On n'écrase une entrée existante que si la nouvelle est au moins aussi profonde.
	// Remplacer une analyse à profondeur 5 par une à profondeur 2 ferait perdre du travail.
	if existing, ok := g.Z[hash]; !ok || newEntry.Depth >= existing.Depth {
		g.Z[hash] = newEntry
	}

	return bestScore
}
