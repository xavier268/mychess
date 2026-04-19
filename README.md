# mychess

Moteur d'échecs écrit en Go, avec une interface TUI basée sur BubbleTea.

---

## Structure du projet

```
position/   — représentation de l'échiquier, génération des coups, évaluation statique
game/       — recherche alpha-bêta, table de transposition, boucle d'analyse
client/     — interface textuelle (TUI)
```

---

## Package `position` — Représentation de l'échiquier

### Bitboards

L'échiquier est représenté par un ensemble de **bitboards** : des entiers 64 bits où chaque bit correspond à une case (bit 0 = a1, bit 63 = h8, encodage `rang*8 + colonne`).

```
Position
├── colOcc [2]Bitboard   — cases occupées par chaque couleur (WHITE=0, BLACK=1)
├── pawnOcc  Bitboard    — cases occupées par des pions (toutes couleurs)
├── rookOcc  Bitboard    — cases occupées par des tours et dames
├── bishopOcc Bitboard   — cases occupées par des fous et dames
├── knightOcc Bitboard   — cases occupées par des cavaliers
├── status   Status      — informations compactes (voir ci-dessous)
└── Hash     uint64      — hash Zobrist maintenu incrémentalement
```

Les dames apparaissent à la fois dans `rookOcc` et `bishopOcc` : les mouvements d'une dame sont ainsi calculés comme l'union des mouvements de tour et de fou, sans code spécifique.

### Structure `Status`

Les informations d'état sont compressées dans deux octets :

```
KingStatus[couleur]  — bits 0-5 : case du roi ; bits 6-7 : droits de roque
TurnStatus           — bit 0 : trait (0=blanc, 1=noir)
```

### Pions fantômes (en passant)

La prise en passant est représentée par un **pion fantôme** : lors d'une avance de deux cases, un pion fantôme est placé dans `pawnOcc` à la case de destination adverse (rang 0 pour les blancs, rang 7 pour les noirs), mais *pas* dans `colOcc`. Ce fantôme est invisible comme pièce réelle mais sert de cible à la prise. Il est nettoyé automatiquement lors du coup suivant, ce qui évite toute corruption d'état lors du défaire-refaire.

### Tables d'attaque pré-calculées (`BigTable`)

Une table globale immuable `BT *BigTable` est initialisée au démarrage :

- Attaques fixes : rois, cavaliers, pions (tableaux indexés par case)
- Attaques glissantes : tours et fous, représentées par des maps `occupancy → Bitboard` par case et par axe (rang, colonne, diagonale NE, diagonale NW)

La légalité d'un coup est vérifiée *après* application par lookup inverse : on recalcule les attaquants de la case du roi avec les tables.

### Hachage Zobrist

Le hash est maintenu de façon incrémentale dans `DoMove`/`UndoMove` par XOR :

- Une valeur aléatoire par (type de pièce, case) pour les pièces
- Des valeurs séparées pour la position des rois, les droits de roque et le trait
- La contribution de l'état de roque et des pions fantômes est extraite avant modification, puis réinsérée après

### Coups (`Move`)

```go
type Move struct {
    From, To      Square
    Promotion     Piece    // EMPTY, CASTLEMOVE ou pièce de promotion
    Score         uint8    // Score d'ordre de tri pour l'alpha-bêta
    // Champs de défaire (remplis par DoMove) :
    Captured      Piece
    CaptureSquare Square
    PrevStatus    Status
    PrevEPFile    int8
    PrevHash      uint64
}
```

`DoMove` prend un snapshot complet de l'état dans le coup lui-même ; `UndoMove` restaure atomiquement depuis ce snapshot, sans reconstruction incrémentale.

### Évaluation statique (`Value()`)

L'évaluation statique (profondeur 0) comprend :

| Composante | Valeur |
|---|---|
| Mat | ±30 000 |
| Pat | 0 |
| Reine | 90 |
| Tour | 50 |
| Fou / Cavalier | 30 |
| Pion | 10 |
| Contrôle du centre (d4/d5/e4/e5) | +1 par case contrôlée |
| Droits de roque disponibles | +1 par droit |
| Roi en échec | −10 |

Tous les scores sont du point de vue du joueur au trait (convention negamax).

---

## Package `game` — Analyse

### Negamax avec élagage alpha-bêta

L'algorithme central est un **negamax avec élagage alpha-bêta** (`AlphaBeta`). La convention negamax signifie que le score est toujours exprimé du point de vue du joueur au trait ; au moment de rappeler récursivement, les bornes et le score sont niés (`-beta, -alpha`).

```
AlphaBeta(α, β, profondeur) :
  1. Consultation de la table de transposition
     → Pruning possible si l'entrée est suffisamment profonde
  2. Si profondeur == 0 → évaluation statique
  3. Génération des pseudo-coups légaux (triés par score de capture)
  4. Déplacement du meilleur coup TT en tête de liste
  5. Pour chaque coup :
       DoMove → appel récursif avec (-β, -α) → UndoMove
       Filtrage de légalité : rejeter si le roi est en échec
       Mise à jour de α et du meilleur coup
       Élagage si α ≥ β
  6. Stockage en table de transposition (EXACT / LOWER / UPPER)
```

### Table de transposition (`ZMap`)

Tableau de taille fixe (`ZSize` entrées, configurable à la compilation) indexé par `hash % ZSize`.

Chaque entrée contient :
- `Best` : meilleur coup trouvé
- `Score` + `ScoreType` (EXACT / LOWER / UPPER)
- `Depth` : profondeur d'analyse
- `ConfirmH` : 16 bits supérieurs du hash (détection de collision)
- `Age` : numéro du coup au moment du stockage (politique de remplacement)

**Politique de remplacement :** même hash → on garde l'entrée la plus profonde ; collision de hash → on remplace si l'entrée stockée est plus ancienne.

**Tailles disponibles** (via build tags) :

| Tag | Entrées | Mémoire approx. |
|---|---|---|
| `size_low` | 100 K | 8 Mo |
| *(défaut)* | 5 M | ~400 Mo |
| `size_high` | 20 M | ~1,6 Go |

### Approfondissement itératif

`Analysis` appelle `AlphaBeta` en boucle avec une profondeur croissante (1, 2, 3, …). À chaque profondeur complète, le meilleur coup est sauvegardé dans `LastRootEntry`. Si l'analyse est interrompue (timeout ou coup joué), le résultat de la dernière profondeur *complète* est utilisé, garantissant un comportement *anytime*.

### Annulation et concurrence

L'analyse tourne dans une goroutine séparée et reçoit un `context.Context`. Chaque nœud de la recherche vérifie la validité du contexte ; en cas d'annulation, le nœud retourne sans écrire en table de transposition (pour ne pas polluer la table avec des scores incomplets).
