package position

import (
	"math/rand"
	"testing"
)

// TestFuzz_DoUndoRoundTrip joue des parties aléatoires depuis la position de
// départ et vérifie après chaque DoMove / UndoMove :
//   - la cohérence interne de la position (Validate),
//   - l'égalité structurelle avant DoMove vs après UndoMove,
//   - la cohérence du Hash incrémental.
//
// La graine est fixe pour reproductibilité ; remonter NbGames pour explorer plus.
func TestFuzz_DoUndoRoundTrip(t *testing.T) {
	const (
		NbGames = 1000
		MaxPly  = 200
		Seed    = 42
	)
	rng := rand.New(rand.NewSource(Seed))

	for game := range NbGames {
		p := StartPosition
		p.Hash = DefaultZT.HashPosition(p)

		// Trace des coups joués pour pouvoir rejouer la partie en cas d'échec.
		history := make([]Move, 0, MaxPly)

		for ply := range MaxPly {
			// 1. La position courante doit être cohérente.
			if msg := p.Validate(); msg != "" {
				t.Fatalf("game %d ply %d: position courante invalide: %s\nhistorique=%v\n%s",
					game, ply, msg, history, p.DebugString())
			}

			moves := p.GetMoveList()
			if len(moves) == 0 {
				break // mat ou pat
			}

			// 2. Tester DoMove + UndoMove sur TOUS les coups pseudo-légaux.
			for _, m := range moves {
				snapshot := p

				after, enriched := p.DoMove(m)
				if msg := after.Validate(); msg != "" {
					t.Fatalf("game %d ply %d coup %s: invalide APRÈS DoMove: %s\nhistorique=%v\nposition AVANT=\n%s\nposition APRÈS=\n%s",
						game, ply, m, msg, history, snapshot.DebugString(), after.DebugString())
				}

				// Hash incrémental doit correspondre au hash recalculé.
				if want := DefaultZT.HashPosition(after); after.Hash != want {
					t.Fatalf("game %d ply %d coup %s: hash incrémental %016x != recalculé %016x\nhistorique=%v",
						game, ply, m, after.Hash, want, history)
				}

				restored := after.UndoMove(enriched)
				if msg := restored.Validate(); msg != "" {
					t.Fatalf("game %d ply %d coup %s: invalide APRÈS UndoMove: %s\nhistorique=%v\nposition AVANT=\n%s\nposition APRÈS DoMove=\n%s\nposition APRÈS UndoMove=\n%s",
						game, ply, m, msg, history, snapshot.DebugString(), after.DebugString(), restored.DebugString())
				}

				if restored != snapshot {
					t.Fatalf("game %d ply %d coup %s: UndoMove n'a pas restauré la position\nhistorique=%v\nAVANT=\n%s\nAPRÈS UndoMove=\n%s",
						game, ply, m, history, snapshot.DebugString(), restored.DebugString())
				}
			}

			// 3. Jouer un coup légal au hasard (filtre les coups laissant son
			// propre roi en prise). Si on jouait n'importe quel coup pseudo-légal,
			// on pourrait capturer le roi adverse au coup suivant et passer dans
			// un état où status.KingPosition pointe sur une case vide.
			legal := make([]Move, 0, len(moves))
			mover := p.Turn()
			for _, m := range moves {
				after, _ := p.DoMove(m)
				if !after.IsSquareAttacked(after.KingPosition(mover), 1^mover) {
					legal = append(legal, m)
				}
			}
			if len(legal) == 0 {
				break // mat ou pat (aucun pseudo-légal n'était légal)
			}
			m := legal[rng.Intn(len(legal))]
			var enriched Move
			p, enriched = p.DoMove(m)
			history = append(history, enriched)
		}
	}
}
