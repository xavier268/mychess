package game

import (
	"context"
	"fmt"
	"testing"

	"github.com/xavier268/mychess/position"
)

// BenchmarkAnalysis mesure le coût d'une analyse complète (approfondissement itératif
// de la profondeur 1 jusqu'à depth) depuis la position de départ.
//
// Un nouveau Game est créé à chaque itération : la table Z part vide (cold start).
// Cela donne une mesure reproductible du coût réel, sans biais dû au cache chaud.
//
// Lancer : go test ./game/ -bench=BenchmarkAnalysis -benchtime=3s
func BenchmarkAnalysis(b *testing.B) {
	for _, depth := range []uint16{1, 2, 3, 4, 5, 6} {
		b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
			for b.Loop() {
				g := NewGame()
				g.Analysis(context.Background(), depth)
			}
		})
	}
}

// BenchmarkAlphaBeta mesure un seul appel AlphaBeta à profondeur fixe,
// sans itérative deepening ni bénéfice du cache des profondeurs inférieures.
//
// Comparer BenchmarkAlphaBeta/depth=N avec BenchmarkAnalysis/depth=N révèle
// le gain apporté par l'iterative deepening sur le remplissage de la table Z.
//
// Lancer : go test ./game/ -bench=BenchmarkAlphaBeta -benchtime=3s
func BenchmarkAlphaBeta(b *testing.B) {
	for _, depth := range []uint16{1, 2, 3, 4, 5, 6} {
		b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
			for b.Loop() {
				g := NewGame()
				g.AlphaBeta(context.Background(), position.LOST, position.WON, depth)
			}
		})
	}
}
