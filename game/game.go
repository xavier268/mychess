// High-level library for analysis and scoring.
package game

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/xavier268/mychess/position"
)

// Game capture le contexte d'une partie en cours.
type Game struct {
	// Position courante (inclut le tour à jouer).
	Position position.Position
	// Historique des coups joués depuis le début de la partie.
	History []position.Move
	// Table de transposition : Zobrist hash (uint64) → ZEntry.
	Z *ZMap
	// Logger pour tracer les événements importants (analyse, coups, erreurs).
	Log log.Logger

	// mu est tenu par la goroutine d'analyse pendant toute sa durée.
	// mu protege Z et Position, pas History.
	// Play() et AnalysisAsync() se synchronisent via ce verrou.
	mu sync.Mutex

	// cancelAnalysis annule le contexte de l'analyse en cours.
	// Nil si aucune analyse n'a jamais été lancée.
	// Appeler cancel plusieurs fois est un no-op (garanti par le package context).
	cancelAnalysis context.CancelFunc

	// LastRootEntry est l'entrée de la position racine issue de la dernière
	// profondeur entièrement terminée. Elle est mise à jour dans Analysis après
	// chaque profondeur complète, et sert de source fiable pour AutoPlay même
	// quand Z a écrasé l'entrée racine (table saturée).
	LastRootEntry ZEntry
}

func NewGame() *Game {
	if zt, zmap, path, ok := tryLoadCache(SearchDirs); ok {
		position.RestoreDefaultZT(zt)
		g := &Game{
			Position: position.StartPosition,
			History:  make([]position.Move, 0, 100),
			Z:        zmap,
			Log:      *log.Default(),
		}
		g.Log.Printf("cache loaded: %s (fill %.1f%%)", path, zmap.FillPercent())
		return g
	}
	return &Game{
		Position: position.StartPosition,
		History:  make([]position.Move, 0, 100),
		Z:        NewZMap(),
		Log:      *log.Default(),
	}
}

// AnalysisAsync lance l'analyse en arrière-plan.
//
// Si une analyse est déjà en cours, elle est d'abord annulée et on attend
// qu'elle se termine avant de lancer la nouvelle (via mu.Lock()).
//
// Un sous-contexte est créé depuis parentCtx : son cancel est stocké dans
// g.cancelAnalysis pour que Play() puisse stopper l'analyse proprement.
//
// mu est acquis ici, dans la goroutine appelante, avant le go — ce qui garantit
// qu'aucune fenêtre de race n'existe entre le lancement et la prise du verrou.
func (g *Game) AnalysisAsync(parentCtx context.Context, maxDepth uint16) {

	// Stoppe l'éventuelle analyse précédente.
	if g.cancelAnalysis != nil {
		g.cancelAnalysis()
	}

	// avoid nil contexts
	if parentCtx == nil {
		parentCtx = context.Background()
	}

	// Crée un sous-contexte annulable, indépendant de parentCtx.
	// Play() utilisera ce cancel pour interrompre proprement l'analyse.
	ctx, cancel := context.WithCancel(parentCtx)
	g.cancelAnalysis = cancel

	// Acquiert le verrou AVANT de lancer la goroutine.
	// Dès le retour de AnalysisAsync, mu est tenu : tout appel à Play()
	// bloquera sur mu.Lock() jusqu'à la fin effective de l'analyse.
	g.mu.Lock()
	go func() {
		defer g.mu.Unlock()
		g.Analysis(ctx, maxDepth)
	}()
}

// Analysis remplit la table de transposition Z par approfondissement itératif :
// elle appelle AlphaBeta pour des profondeurs croissantes (1, 2, 3, …, maxDepth)
// et s'arrête dès que le contexte expire ou que maxDepth est atteinte.
//
// Garantie de cohérence : si le contexte expire en cours d'analyse à la profondeur d,
// AlphaBeta détecte l'interruption nœud par nœud et remonte sans rien écrire dans Z.
// Seules les profondeurs entièrement terminées sont reflétées dans la table.
//
// Retourne la dernière profondeur entièrement explorée (0 si aucune n'a pu l'être).
func (g *Game) Analysis(ctx context.Context, maxDepth uint16) (depth uint16) {

	// avoid nil contexts
	if ctx == nil {
		ctx = context.Background()
	}

	rootHash := g.Position.Hash
	for d := uint16(1); d <= maxDepth; d++ {
		// Fenêtre initiale maximale [LOST, WON] : recherche complète sans aspiration.
		g.AlphaBeta(ctx, position.LOST, position.WON, d)

		if ctx.Err() != nil {
			// Le niveau d a été interrompu : la table Z n'a pas été modifiée pour ce niveau.
			// On retourne la dernière profondeur entièrement terminée.
			return depth
		}

		// Niveau d terminé avec succès : épingle l'entrée racine avant qu'elle
		// soit éventuellement écrasée par la suite de la recherche.
		if entry, found := g.Z.Get(rootHash); found {
			g.LastRootEntry = entry
		}
		depth = d
	}
	return depth // maxDepth atteinte sans interruption
}

// Play joue le coup m : met à jour la position et l'ajoute à l'historique.
//
// Si une analyse asynchrone est en cours, Play l'arrête proprement avant de jouer :
//  1. Cancel du contexte d'analyse → AlphaBeta remonte nœud par nœud sans corrompre Z.
//  2. mu.Lock() attend la fin effective de la goroutine (bref, quelques µs).
//  3. Le coup est appliqué.
//
// La table Z n'est pas vidée : ses entrées restent exploitables pour la prochaine analyse.
func (g *Game) Play(m position.Move) {
	// Annule l'analyse en cours le cas échéant.
	// No-op si cancelAnalysis est nil ou si le contexte est déjà annulé.
	if g.cancelAnalysis != nil {
		g.cancelAnalysis()
	}

	// Attend que la goroutine d'analyse relâche le verrou.
	// Si aucune analyse ne tourne, Lock() réussit immédiatement.
	g.mu.Lock()
	defer g.mu.Unlock()

	var enrichedMove position.Move
	g.Position, enrichedMove = g.Position.DoMove(m)
	g.History = append(g.History, enrichedMove)
}

// AutoPlay joue le meilleur coup connu pour la position courante, d'après la table Z.
//
// Même comportement que Play vis-à-vis de l'analyse asynchrone :
//  1. Cancel du contexte si une analyse tourne.
//  2. mu.Lock() attend la fin effective de la goroutine.
//  3. Le meilleur coup est lu dans Z, puis joué.
//
// Retourne une erreur si aucun coup n'est disponible (aucune analyse lancée,
// ou position non encore visitée dans Z).
func (g *Game) AutoPlay() error {
	if g.cancelAnalysis != nil {
		g.cancelAnalysis()
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Le meilleur coup est d'abord cherché dans Z.
	// Si l'entrée a été élaguée (pruneZLocked peut supprimer la racine quand toutes
	// les entrées ont le même Age), on se replie sur le snapshot atomique lastRootEntry
	// qui n'est jamais élaguée et dont le hash permet de vérifier la fraîcheur.
	// moveIsValid retourne true si le Move a un coup From→To significatif.
	// On compare seulement From/To/Promotion : les champs Undo (PrevHash, etc.)
	// sont remplis par DoMove et ne doivent pas entrer dans le test de validité.
	moveIsValid := func(m position.Move) bool {
		return m.From != m.To || m.Promotion != position.EMPTY
	}

	entry, found := g.Z.Get(g.Position.Hash)
	if !found || !moveIsValid(entry.Best) {
		entry = g.LastRootEntry
	}
	if !moveIsValid(entry.Best) {
		return errors.New("aucun coup disponible : lancez une analyse avant d'appeler AutoPlay")
	}

	var enrichedMove position.Move
	g.Position, enrichedMove = g.Position.DoMove(entry.Best)
	g.History = append(g.History, enrichedMove)
	return nil
}
