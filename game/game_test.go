package game

// Tests du package game — focalisés sur Analysis (synchrone).
//
// Note sur la génération de coups : GetMoveList() génère des coups PSEUDO-LÉGAUX.
// La détection de mat ne passe donc pas par len(moves)==0, mais par Value() qui
// retourne WON quand le roi adverse est prenable au coup suivant.
// La détection pat/mat réelle (len==0 + IsCheck) ne se déclenche que si toutes
// les pseudo-légales laissent effectivement le roi en prise.

import (
	"context"
	"testing"
	"time"

	"mychess/position"
)

// ── Helpers ──────────────────────────────────────────────────────────────────

// mateInOneGame construit une partie dans une position où les Blancs font mat en 1 : Qa7→g7#
//
//	Blancs : Roi h6, Dame a7  |  Noirs : Roi h8  |  Trait aux Blancs.
//
// Vérification géométrique de Qg7# :
//   - Le Roi noir sur h8 est en échec (diagonale g7→h8).
//   - g8 : couvert par Qg7 (colonne g).
//   - h7 : couvert par Qg7 (rangée 7) et Roi blanc h6.
//   - g7 : défendu par Roi blanc h6 (case adjacente) → le Roi noir ne peut pas capturer.
//
// Profondeur 2 requise : à d=1, AlphaBeta joue Qa7→g7 mais ne vérifie pas les réponses
// noires. À d=2, tous les pseudo-coups noirs conduisent à une position où Value() retourne
// WON (Roi noir prenable), ce qui identifie la séquence comme gagnante.
func mateInOneGame(t *testing.T) *Game {
	t.Helper()
	var pos position.Position
	pos.AddKing(position.WHITE, "h6")
	pos.AddQueen(position.WHITE, "a7")
	pos.AddKing(position.BLACK, "h8")
	pos.Hash = position.DefaultZT.HashPosition(pos)

	g := NewGame(context.Background())
	g.Position = pos
	return g
}

// firstLegalMove retourne le premier coup pseudo-légal de la position courante.
func firstLegalMove(g *Game) position.Move {
	moves := g.Position.GetMoveList()
	if len(moves) == 0 {
		panic("firstLegalMove: aucun coup dans la position courante")
	}
	return moves[0]
}

// ── NewGame ──────────────────────────────────────────────────────────────────

func TestNewGame_initialState(t *testing.T) {
	g := NewGame(context.Background())
	if g.Position != position.StartPosition {
		t.Error("NewGame: la position initiale devrait être StartPosition")
	}
	if len(g.History) != 0 {
		t.Errorf("NewGame: History devrait être vide, contient %d entrées", len(g.History))
	}
	if len(g.Z) != 0 {
		t.Errorf("NewGame: Z devrait être vide, contient %d entrées", len(g.Z))
	}
}

// ── Play ─────────────────────────────────────────────────────────────────────

func TestPlay_updatesPositionAndHistory(t *testing.T) {
	g := NewGame(context.Background())
	before := g.Position

	g.Play(firstLegalMove(g))

	if g.Position == before {
		t.Error("Play: la position n'a pas changé après le coup")
	}
	if len(g.History) != 1 {
		t.Errorf("Play: History devrait contenir 1 coup, contient %d", len(g.History))
	}
}

// ── RetractPlay ──────────────────────────────────────────────────────────────

func TestRetractPlay_restoresPosition(t *testing.T) {
	g := NewGame(context.Background())
	before := g.Position

	g.Play(firstLegalMove(g))
	if err := g.RetractPlay(); err != nil {
		t.Fatalf("RetractPlay: erreur inattendue : %v", err)
	}

	if g.Position != before {
		t.Error("RetractPlay: la position devrait être restaurée à l'état initial")
	}
	if len(g.History) != 0 {
		t.Errorf("RetractPlay: History devrait être vide après annulation, contient %d coups", len(g.History))
	}
}

func TestRetractPlay_errorOnEmptyHistory(t *testing.T) {
	g := NewGame(context.Background())
	if err := g.RetractPlay(); err == nil {
		t.Error("RetractPlay: devrait retourner une erreur si l'historique est vide")
	}
}

func TestRetractPlay_doublePlayDoubleRetract(t *testing.T) {
	g := NewGame(context.Background())
	before := g.Position

	g.Play(firstLegalMove(g))
	g.Play(firstLegalMove(g))

	if err := g.RetractPlay(); err != nil {
		t.Fatalf("premier RetractPlay: %v", err)
	}
	if err := g.RetractPlay(); err != nil {
		t.Fatalf("deuxième RetractPlay: %v", err)
	}

	if g.Position != before {
		t.Error("après deux Play + deux RetractPlay, la position devrait être restaurée")
	}
	if len(g.History) != 0 {
		t.Errorf("History devrait être vide, contient %d coups", len(g.History))
	}
}

// ── AutoPlay ─────────────────────────────────────────────────────────────────

func TestAutoPlay_errorWithoutAnalysis(t *testing.T) {
	g := NewGame(context.Background())
	if err := g.AutoPlay(); err == nil {
		t.Error("AutoPlay: devrait retourner une erreur quand Z est vide")
	}
}

func TestAutoPlay_playsAfterAnalysis(t *testing.T) {
	g := NewGame(context.Background())
	g.Analysis(context.Background(), 1)

	before := g.Position
	if err := g.AutoPlay(); err != nil {
		t.Fatalf("AutoPlay: erreur inattendue après analyse : %v", err)
	}
	if g.Position == before {
		t.Error("AutoPlay: la position devrait avoir changé")
	}
	if len(g.History) != 1 {
		t.Errorf("AutoPlay: History devrait contenir 1 coup, contient %d", len(g.History))
	}
}

// ── Analysis ─────────────────────────────────────────────────────────────────

func TestAnalysis_populatesZ(t *testing.T) {
	g := NewGame(context.Background())
	g.Analysis(context.Background(), 2)

	if len(g.Z) == 0 {
		t.Error("Analysis: Z devrait être rempli après une analyse à profondeur 2")
	}
}

func TestAnalysis_rootEntryPresent(t *testing.T) {
	g := NewGame(context.Background())
	g.Analysis(context.Background(), 2)

	entry, found := g.Z[g.Position.Hash]
	if !found {
		t.Fatal("Analysis: Z ne contient pas d'entrée pour la position racine")
	}
	if entry.Depth < 1 {
		t.Errorf("Analysis: l'entrée racine a Depth=%d, attendu ≥ 1", entry.Depth)
	}
	if entry.Best == (position.Move{}) {
		t.Error("Analysis: l'entrée racine devrait avoir un meilleur coup")
	}
}

func TestAnalysis_returnedDepth(t *testing.T) {
	g := NewGame(context.Background())
	depth := g.Analysis(context.Background(), 2)
	if depth != 2 {
		t.Errorf("Analysis: devrait retourner 2 pour une recherche complète, obtenu %d", depth)
	}
}

func TestAnalysis_deeperSearchOverwritesShallower(t *testing.T) {
	g := NewGame(context.Background())

	g.Analysis(context.Background(), 1)
	depthAfter1 := g.Z[g.Position.Hash].Depth

	g.Analysis(context.Background(), 2)
	depthAfter2 := g.Z[g.Position.Hash].Depth

	if depthAfter2 <= depthAfter1 {
		t.Errorf("une analyse plus profonde devrait mettre à jour l'entrée (avant=%d, après=%d)",
			depthAfter1, depthAfter2)
	}
}

func TestAnalysis_cancelledBeforeStart(t *testing.T) {
	g := NewGame(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // annulé avant même de démarrer

	depth := g.Analysis(ctx, 10)
	if depth != 0 {
		t.Errorf("Analysis: devrait retourner 0 si le contexte est déjà annulé, obtenu %d", depth)
	}
	if len(g.Z) != 0 {
		t.Errorf("Analysis: Z devrait rester vide si annulé avant le démarrage, contient %d entrées", len(g.Z))
	}
}

// TestAnalysis_zCoherenceAfterCancel vérifie l'invariant principal de l'implémentation :
// aucune entrée de Z n'a une profondeur supérieure à la dernière profondeur complète.
//
// Garantie : AlphaBeta ne stocke rien dans Z si g.Ctx est annulé pendant la recherche.
// Seules les profondeurs entièrement terminées sont donc visibles dans la table.
//
// Le time.Sleep est nécessaire ici : Analysis est synchrone mais lancé dans une goroutine
// pour permettre l'annulation à mi-chemin. On attend suffisamment pour que la profondeur 1
// se termine (~1 ms en pratique), puis on annule avant la profondeur 2.
func TestAnalysis_zCoherenceAfterCancel(t *testing.T) {
	g := NewGame(context.Background())
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan uint16, 1)
	go func() {
		ch <- g.Analysis(ctx, 10)
	}()

	// 20 ms laissent la profondeur 1 se terminer (< 1 ms en pratique sur la position de départ).
	time.Sleep(20 * time.Millisecond)
	cancel()
	completedDepth := <-ch

	// Invariant : toute entrée de Z doit avoir Depth ≤ completedDepth.
	for hash, entry := range g.Z {
		if entry.Depth > completedDepth {
			t.Errorf("incohérence Z : hash=%x Depth=%d > profondeur complète=%d",
				hash, entry.Depth, completedDepth)
		}
	}
	if completedDepth == 0 && len(g.Z) > 0 {
		t.Errorf("Z devrait être vide si completedDepth=0, contient %d entrées", len(g.Z))
	}
}

// TestAnalysis_findsWinningScore vérifie que l'analyse identifie la position
// comme gagnante pour les Blancs sur le problème mat-en-1.
//
// À profondeur 2 : White joue Qa7→g7 (d=2→1), puis tous les pseudo-coups noirs
// conduisent à d=0 où Value() retourne WON (Roi noir prenable par la Dame).
// Le score remonté à la racine doit donc être WON.
func TestAnalysis_findsWinningScore(t *testing.T) {
	g := mateInOneGame(t)
	completedDepth := g.Analysis(context.Background(), 2)
	if completedDepth < 2 {
		t.Fatalf("Analysis n'a pas complété la profondeur 2 (complétée : %d)", completedDepth)
	}

	entry, found := g.Z[g.Position.Hash]
	if !found {
		t.Fatal("Z ne contient pas d'entrée pour la position racine")
	}
	if entry.Score != position.WON {
		t.Errorf("score attendu = WON (%d), obtenu %d", position.WON, entry.Score)
	}
}

// TestAnalysis_bestMoveLeadsToCheck vérifie que le meilleur coup trouvé
// met effectivement le Roi adverse en échec (premier pas du mat).
func TestAnalysis_bestMoveLeadsToCheck(t *testing.T) {
	g := mateInOneGame(t)
	g.Analysis(context.Background(), 2)

	entry, found := g.Z[g.Position.Hash]
	if !found || entry.Best == (position.Move{}) {
		t.Fatal("Z ne contient pas de meilleur coup pour la position racine")
	}

	// Joue le meilleur coup et vérifie que le Roi adverse est en échec.
	newPos, _ := g.Position.DoMove(entry.Best)
	if !newPos.IsCheck() {
		t.Errorf("le meilleur coup (%v) devrait mettre le Roi adverse en échec", entry.Best)
	}
}

// ── PruneZ ───────────────────────────────────────────────────────────────────

func TestPruneZ_reducesToTargetSize(t *testing.T) {
	g := NewGame(context.Background())
	g.Analysis(context.Background(), 2)

	initial := len(g.Z)
	if initial < 2 {
		t.Skip("Z trop petite pour tester la purge")
	}
	target := initial / 2
	g.PruneZ(target)

	if len(g.Z) > target {
		t.Errorf("PruneZ: attendu len(Z) ≤ %d, obtenu %d", target, len(g.Z))
	}
}

func TestPruneZ_noOpWhenSmallEnough(t *testing.T) {
	g := NewGame(context.Background())
	g.Analysis(context.Background(), 2)

	before := len(g.Z)
	g.PruneZ(before + 1000)

	if len(g.Z) != before {
		t.Errorf("PruneZ: ne devrait rien supprimer si déjà sous la cible (avant=%d après=%d)",
			before, len(g.Z))
	}
}

// TestPruneZ_removesOldestEntries injecte des entrées avec des Ages connus
// et vérifie que PruneZ supprime d'abord les plus anciennes (Age le plus bas).
func TestPruneZ_removesOldestEntries(t *testing.T) {
	g := NewGame(context.Background())

	// 6 entrées avec Ages 0..5 (hashes distincts pour éviter les collisions).
	for i := uint64(0); i < 6; i++ {
		g.Z[i] = ZEntry{Age: int64(i), Depth: 1}
	}

	// Purge à 3 : Ages 0, 1, 2 doivent disparaître ; Ages 3, 4, 5 doivent rester.
	g.PruneZ(3)

	if len(g.Z) > 3 {
		t.Errorf("PruneZ: attendu len(Z) ≤ 3, obtenu %d", len(g.Z))
	}
	for hash, entry := range g.Z {
		if entry.Age < 3 {
			t.Errorf("PruneZ: entrée hash=%d Age=%d aurait dû être supprimée", hash, entry.Age)
		}
	}
}
