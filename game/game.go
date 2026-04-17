package game

import (
	"context"
	"errors"
	"mychess/position"
	"sort"
	"sync"
	"sync/atomic"
)

// Game capture le contexte d'une partie en cours.
type Game struct {
	// Position courante (inclut le tour à jouer).
	Position position.Position
	// Historique des coups joués depuis le début de la partie.
	History []position.Move
	// Table de transposition : Zobrist hash → ZEntry.
	Z map[uint64]ZEntry

	// mu est tenu par la goroutine d'analyse pendant toute sa durée.
	// Play() et AnalysisAsync() se synchronisent via ce verrou.
	mu sync.Mutex

	// cancelAnalysis annule le contexte de l'analyse en cours.
	// Nil si aucune analyse n'a jamais été lancée.
	// Appeler cancel plusieurs fois est un no-op (garanti par le package context).
	cancelAnalysis context.CancelFunc

	// Champs atomiques lisibles depuis n'importe quelle goroutine (UI, etc.)
	// sans synchronisation supplémentaire.
	analysisRunning atomic.Bool  // true pendant qu'une goroutine d'analyse tourne
	zSize           atomic.Int64 // approximation de len(Z), mise à jour à chaque écriture
	lastRootEntry   atomic.Value // stores rootSnapshot (hash + ZEntry de la position racine)
}

// rootSnapshot associe un Zobrist hash à la ZEntry de la position racine.
// Stocké dans lastRootEntry pour que AutoPlay puisse vérifier que l'entrée
// correspond bien à la position courante même si Z a été élaguée.
type rootSnapshot struct {
	Hash  uint64
	Entry ZEntry
}

type ZEntry struct {
	// Best move found until now (or null move)
	Best position.Move
	// Score (upper, lower, exact)
	Score position.Score
	// Score type : UPPER, LOWER, EXACT
	ScoreType ScoreType
	// Depth of analysis at this stage
	Depth uint16
	// When was this entry last updated ?
	Age int64
}

type ScoreType uint8

const (
	UPPER ScoreType = iota
	LOWER
	EXACT
)

func NewGame() *Game {
	return &Game{
		Position: position.StartPosition,
		History:  make([]position.Move, 0, 100),
		Z:        make(map[uint64]ZEntry, 1000),
	}
}

// IsAnalysisRunning retourne true si une goroutine d'analyse tourne actuellement.
// Sûr à appeler depuis n'importe quelle goroutine (lecture atomique).
func (g *Game) IsAnalysisRunning() bool { return g.analysisRunning.Load() }

// ZTableSize retourne le nombre approximatif d'entrées dans la table Z.
// La valeur est mise à jour à chaque écriture dans AlphaBeta (lecture atomique).
func (g *Game) ZTableSize() int { return int(g.zSize.Load()) }

// LastRootEntry retourne la dernière ZEntry enregistrée pour la position racine,
// et true si elle existe. Sûr à appeler depuis n'importe quelle goroutine.
func (g *Game) LastRootEntry() (ZEntry, bool) {
	v := g.lastRootEntry.Load()
	if v == nil {
		return ZEntry{}, false
	}
	return v.(rootSnapshot).Entry, true
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

	// Crée un sous-contexte annulable, indépendant de parentCtx.
	// Play() utilisera ce cancel pour interrompre proprement l'analyse.
	ctx, cancel := context.WithCancel(parentCtx)
	g.cancelAnalysis = cancel

	// Acquiert le verrou AVANT de lancer la goroutine.
	// Dès le retour de AnalysisAsync, mu est tenu : tout appel à Play()
	// bloquera sur mu.Lock() jusqu'à la fin effective de l'analyse.
	g.analysisRunning.Store(true)
	g.mu.Lock()
	go func() {
		defer g.mu.Unlock()
		defer g.analysisRunning.Store(false)
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
	// Si Z est déjà pleine avant de démarrer (entrées d'autres positions), on élague
	// d'emblée pour garantir que le premier AlphaBeta puisse écrire l'entrée racine.
	if len(g.Z) >= maxZEntries {
		g.pruneZLocked(maxZEntries / 2)
	}
	for d := uint16(1); d <= maxDepth; d++ {
		// Fenêtre initiale maximale [LOST, WON] : recherche complète sans aspiration.
		g.AlphaBeta(ctx, position.LOST, position.WON, d)

		if ctx.Err() != nil {
			// Le niveau d a été interrompu : la table Z n'a pas été modifiée pour ce niveau.
			// On retourne la dernière profondeur entièrement terminée.
			return depth
		}

		// Niveau d terminé avec succès : on le mémorise comme dernier niveau complet.
		depth = d
		// Publie l'entrée racine pour que l'UI puisse lire score/meilleur coup
		// sans accès concurrent à la map Z, et qu'AutoPlay puisse y accéder
		// même si Z est élaguée ultérieurement.
		if entry, ok := g.Z[g.Position.Hash]; ok {
			g.lastRootEntry.Store(rootSnapshot{Hash: g.Position.Hash, Entry: entry})
		}
		// Élagage automatique : évite la croissance non bornée de Z.
		// pruneZLocked peut être appelé ici car l'analyse tient déjà mu.
		if len(g.Z) >= maxZEntries {
			g.pruneZLocked(maxZEntries / 2)
		}
	}
	return // maxDepth atteinte sans interruption
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

// RetractPlay annule le dernier coup joué et restaure la position précédente.
//
// Si une analyse asynchrone est en cours, elle est d'abord stoppée proprement
// (cancel + attente de la goroutine) — même comportement que Play et AutoPlay.
//
// Retourne une erreur si l'historique est vide (aucun coup à annuler).
func (g *Game) RetractPlay() error {
	if g.cancelAnalysis != nil {
		g.cancelAnalysis()
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if len(g.History) == 0 {
		return errors.New("aucun coup à annuler : l'historique est vide")
	}

	// Dépile le dernier coup. Il contient toutes les informations nécessaires
	// à UndoMove (PrevStatus, PrevHash, pièce capturée, etc.).
	lastMove := g.History[len(g.History)-1]
	g.History = g.History[:len(g.History)-1]

	g.Position = g.Position.UndoMove(lastMove)
	return nil
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

	entry, found := g.Z[g.Position.Hash]
	if !found || !moveIsValid(entry.Best) {
		if snap := g.lastRootEntry.Load(); snap != nil {
			rs := snap.(rootSnapshot)
			if rs.Hash == g.Position.Hash && moveIsValid(rs.Entry.Best) {
				entry, found = rs.Entry, true
			}
		}
	}
	if !found || !moveIsValid(entry.Best) {
		return errors.New("aucun coup disponible : lancez une analyse avant d'appeler AutoPlay")
	}

	var enrichedMove position.Move
	g.Position, enrichedMove = g.Position.DoMove(entry.Best)
	g.History = append(g.History, enrichedMove)
	return nil
}

// maxZEntries est la taille maximale de la table Z avant élagage automatique.
// Chaque ZEntry fait ~32 octets : 2 000 000 entrées ≈ 64 Mo.
const maxZEntries = 2_000_000

// PruneZ supprime les entrées les plus anciennes de la table Z jusqu'à ce que
// sa taille soit inférieure à size. Si une analyse asynchrone est en cours,
// elle est d'abord stoppée proprement.
//
// L'ancienneté d'une entrée est déterminée par son champ Age, qui vaut
// len(g.History) au moment où l'entrée a été écrite : les petites valeurs
// correspondent aux positions analysées tôt dans la partie.
//
// Complexité : O(n log n) en temps, O(n) en mémoire supplémentaire (n = len(Z)).
func (g *Game) PruneZ(size int) {
	if g.cancelAnalysis != nil {
		g.cancelAnalysis()
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pruneZLocked(size)
}

// pruneZLocked est identique à PruneZ mais suppose que mu est déjà tenu par
// l'appelant (typiquement l'analyse elle-même). Ne tente pas d'acquérir mu.
func (g *Game) pruneZLocked(size int) {
	if len(g.Z) <= size {
		return
	}

	// Extrait les clés avec leur âge dans une slice pour pouvoir les trier.
	// On ne stocke que (hash, age) pour limiter la mémoire utilisée.
	type entry struct {
		hash uint64
		age  int64
	}
	entries := make([]entry, 0, len(g.Z))
	for hash, e := range g.Z {
		entries = append(entries, entry{hash, e.Age})
	}

	// Trie par âge croissant : les plus anciens (petit Age) en premier.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].age < entries[j].age
	})

	// Supprime depuis le début (les plus anciens) jusqu'à atteindre size.
	for _, e := range entries {
		if len(g.Z) <= size {
			break
		}
		delete(g.Z, e.hash)
	}
	g.zSize.Store(int64(len(g.Z)))
}
