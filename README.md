# mychess

Moteur d'échecs écrit en Go, avec une interface TUI basée sur BubbleTea.

📈 [Historique des performances par version](history.md)

---

## Structure du projet

```
position/   — représentation de l'échiquier, génération des coups, évaluation statique
game/       — recherche alpha-bêta, table de transposition, boucle d'analyse
cache/      — génération et lecture des fichiers de cache pré-calculés
gencache/   — commande standalone pour pré-calculer un cache sur disque
client/     — interface textuelle (TUI)
```

---

## Package `position` — Représentation de l'échiquier

### 1. Indexation des cases — `Square`

`Square` est un `uint8` dans l'intervalle `[0, 63]`. Les cases sont numérotées **rang en premier**, à partir de `a1 = 0` :

```
rang 7  56 57 58 59 60 61 62 63     a8 b8 c8 d8 e8 f8 g8 h8
rang 6  48 49 50 51 52 53 54 55     a7 b7 ...
  ...
rang 1   8  9 10 11 12 13 14 15     a2 b2 ...
rang 0   0  1  2  3  4  5  6  7     a1 b1 c1 d1 e1 f1 g1 h1
         ^                  ^
       col 0             col 7
      (col a)           (col h)
```

Formule : `Square = rang * 8 + colonne` (tous deux indexés à partir de 0).  
Constructeur : `Sq(rang, col int) Square`.  
Parsing : `SqParse("d4") → Square(27)`.

---

### 2. Bitboard

Un `Bitboard` est un `uint64` où le bit `n` correspond à `Square(n)` (bit 0 = a1, bit 63 = h8). Un bit à 1 signifie « quelque chose est ici ». Les bitboards sont la structure de données centrale : unions, intersections et exclusions sont des instructions CPU uniques.

Principales opérations (retournent toutes un nouveau `Bitboard`) :

| Méthode | Description |
|---|---|
| `.Set(sq)` | mettre un bit à 1 |
| `.Unset(sq)` | mettre un bit à 0 |
| `.IsSet(sq)` | tester un bit |
| `.Get(sq)` | retourne 0 ou 1 |
| `.BitCount()` | popcount |
| `.AllSetSquares(yield)` | itérateur range-over-func |
| `.AllBitCombinations(yield)` | énumère tous les sous-ensembles |

Constructeurs utilitaires : `Rank(r)`, `File(f)`, `Diagonal(sq)`, `AntiDiagonal(sq)`, `Interior()`, `Full()`, `Border()`.

---

### 3. Structure `Position`

Au lieu d'un tableau de 64 valeurs de pièces, `Position` utilise **six bitboards indépendants**, chacun suivant une propriété sur les 64 cases simultanément :

```go
type Position struct {
    colOcc    [2]Bitboard  // cases occupées par chaque couleur
    pawnOcc   Bitboard     // cases avec un pion (toutes couleurs + fantôme EP)
    rookOcc   Bitboard     // cases avec une tour ou une dame (toutes couleurs)
    bishopOcc Bitboard     // cases avec un fou ou une dame (toutes couleurs)
    knightOcc Bitboard     // cases avec un cavalier (toutes couleurs)
    status    Status
    Hash      uint64       // hash Zobrist maintenu incrémentalement
}
```

#### 3.1 Occupation par couleur — `colOcc[2]`

`colOcc[WHITE]` a un bit pour chaque case occupée par une pièce blanche (roi inclus). `colOcc[BLACK]` de même pour les noirs. L'union `colOcc[WHITE] | colOcc[BLACK]` donne toutes les cases occupées.

#### 3.2 Occupation par type de pièce

Chacun de `pawnOcc`, `rookOcc`, `bishopOcc`, `knightOcc` est **indépendant de la couleur** : le bit est mis à 1 quelle que soit la couleur propriétaire. Pour déterminer la couleur d'une pièce en `sq`, on intersecte avec `colOcc` :

```
tours blanches = rookOcc & colOcc[WHITE]   (hors dames)
```

#### 3.3 Dames — représentation implicite

Il n'existe **pas de `queenOcc` séparé**. Une dame est stockée à la fois comme tour et comme fou : quand une dame est placée en `sq`, les deux bitboards `rookOcc` et `bishopOcc` ont leur bit à 1 en `sq`.

```
isQueen(sq) = rookOcc.IsSet(sq) && bishopOcc.IsSet(sq)
```

La génération des coups d'une dame est alors naturelle : mouvements de dame = mouvements de tour ∪ mouvements de fou, sans code spécifique.

#### 3.4 Rois — stockés dans `Status`

Les rois ne sont **pas** suivis dans les bitboards de type. Leurs cases sont encodées dans la structure `Status` (6 bits chacun dans `KingStatus[couleur]`). Leur présence dans `colOcc` est maintenue normalement.

---

### 4. Structure `Status`

```go
type Status struct {
    KingStatus [2]uint8   // par couleur : bits de roque (2) + case du roi (6)
    TurnStatus uint8      // bit 0 : trait (WHITE=0, BLACK=1)
}
```

Disposition de `KingStatus[couleur]` :

```
bit 7  bit 6  bits 5-0
  │      │       │
  │      │       └── case du roi (0-63)
  │      └────────── peut roquer côté dame ?
  └───────────────── peut roquer côté roi ?
```

Constantes : `CanCastleKingSide = 0b10000000`, `CanCastleQueenSide = 0b01000000`.

---

### 5. Pions fantômes (en passant)

Lors d'une avance de deux cases, l'opportunité de prise en passant est encodée par un **pion fantôme** : un bit dans `pawnOcc` **sans bit correspondant** dans `colOcc`.

Le fantôme est placé à la **case intermédiaire** — celle que le pion vient de traverser — qui est garantie libre puisque la poussée double serait bloquée si elle ne l'était pas :

| Côté ayant avancé de deux | Rang du fantôme | Exemple |
|---|---|---|
| BLANC (rang 1 → rang 3) | rang 2 | e2-e4 → fantôme en e3 |
| NOIR (rang 6 → rang 4) | rang 5 | e7-e5 → fantôme en e6 |

Le fantôme coïncide exactement avec la case d'atterrissage de la prise en passant adverse, ce qui simplifie la détection : il suffit de vérifier que la case diagonale cible est un fantôme.

Les rangs 2 et 5 peuvent contenir de vrais pions, mais ceux-ci sont dans `colOcc` alors que le fantôme ne l'est pas. La distinction reste sans ambiguïté :

**Détection** (dans `GetPawnMovesFromSquareBB`) :

```
fantômes = pawnOcc & ^(colOcc[WHITE] | colOcc[BLACK])
```

**Nettoyage** : `pawnOcc &= colOcc[WHITE] | colOcc[BLACK]` — supprime tous les bits fantômes en une opération.

Aucune garde de placement n'est nécessaire : la case intermédiaire est toujours libre lors d'une poussée double légale.

---

### 6. Identification d'une pièce — `PieceAt`

```
couleur = colOcc[WHITE].Get(sq) - colOcc[BLACK].Get(sq)   → +1, -1 ou 0
```

Si `couleur == 0` → case vide. Sinon, on consulte les bitboards de type dans l'ordre : pion → cavalier → dame (rook∧bishop) → fou → tour → roi (depuis `Status`).

---

### 7. Tables d'attaque pré-calculées — `BigTable`

`BigTable` est une structure immuable construite une seule fois au démarrage par `NewBigTable()`. Elle contient tous les ensembles d'attaque de toute pièce depuis toute case, indexés par les bits d'occupation pertinents. Une fois construite, **toutes les recherches de coups sont des lectures de map sans allocation**.

#### 7.1 Pièces glissantes — conception par direction

Les attaques glissantes sont découpées par **direction** en quatre tables indépendantes :

| Pièce | Direction | Champ masque | Champ map |
|---|---|---|---|
| Tour | rang (E/O) | `RookMaskRank[sq]` | `RookAttackSetRank[sq]` |
| Tour | colonne (N/S) | `RookMaskFile[sq]` | `RookAttackSetFile[sq]` |
| Fou | diagonale NE/SO | `BishopMaskNE[sq]` | `BishopAttackSetNE[sq]` |
| Fou | diagonale NO/SE | `BishopMaskNW[sq]` | `BishopAttackSetNW[sq]` |

Chaque masque directionnel est plus petit (max 6 bits pour un masque de rang contre 12 bits pour un masque complet de tour), ce qui réduit le nombre d'entrées par map et améliore l'utilisation du cache. Les attaques totales sont reconstituées par OR des deux résultats directionnels.

Construction du masque (exemple tour rang) :

```
RookMaskRank[sq] = Rank(r).Unset(sq).Unset(Sq(r,0)).Unset(Sq(r,7))
```

La case elle-même et les deux cases de bordure sont exclues : les cases de bordure sont toujours accessibles quelle que soit l'occupation, les inclure dans la clé gaspillerait des entrées.

Lookup (tour en `sq`) :

```go
occ         := colOcc[WHITE] | colOcc[BLACK]
rankAttacks := RookAttackSetRank[sq][occ & RookMaskRank[sq]]
fileAttacks := RookAttackSetFile[sq][occ & RookMaskFile[sq]]
attacks     := (rankAttacks | fileAttacks) & ^colOcc[trait]
```

#### 7.2 Pièces non-glissantes

`KingAttacks[sq]` et `KnightAttacks[sq]` sont de simples tableaux `[64]Bitboard` — pas de clé d'occupation nécessaire.

#### 7.3 Tables des pions

```go
PawnMask[couleur][sq]      Bitboard
PawnAttackSet[couleur][sq] map[Bitboard]Bitboard   // occupation → coups
```

Le masque combine les **cases de poussée** (1 ou 2 cases devant) et les **cases de capture diagonale**. La map est indexée par `occTotale & PawnMask` :

- Une case de poussée apparaît dans la valeur si et seulement si elle est **inoccupée** dans la clé. Pour un pion sur son rang de départ, la poussée double est omise si la case intermédiaire est occupée (blocage correct, pas de saut).
- Une case de capture apparaît si et seulement si elle est **occupée** par n'importe quelle pièce. L'appelant filtre ensuite avec `& ^colOcc[trait]` pour exclure ses propres pièces.

La prise en passant n'est **pas** dans la map ; elle est calculée séparément dans `GetPawnMovesFromSquareBB` via la détection du pion fantôme (§5).

---

### 8. Génération des coups

#### 8.1 `GetMovesBB(sq) → Bitboard`

Identifie le type de pièce en `sq` en consultant les bitboards de type dans l'ordre, puis dispatche vers le handler approprié. Retourne un `Bitboard` de toutes les cases accessibles (pseudo-légal — ne filtre pas les auto-échecs).

#### 8.2 `GetMoveList() → []Move`

Itère sur toutes les cases de `colOcc[trait]`, appelle `GetMovesBB` pour chacune, et décompose le bitboard résultant en structs `Move`. Les promotions sont développées en quatre coups (D/T/F/C) en ligne. Le roque est ajouté via `GetCastlingMoveList` qui vérifie :

1. Les bits de droits de roque dans `Status`.
2. Le roi n'est pas actuellement en échec.
3. Toutes les cases intermédiaires sont inoccupées.
4. Aucune case intermédiaire n'est attaquée par l'adversaire.

Pour chaque coup candidat, `DoMove`/`UndoMove` est ensuite appelé afin de :

- **Filtrer les coups illégaux** : les candidats qui laissent le roi du joueur courant en échec sont écartés.
- **Scorer les coups donnant échec** : si le coup met le roi adverse en échec, un bonus `checkBonus` (actuellement 5) est ajouté au `Score`, plaçant les échecs silencieux au-dessus des captures de tour mais en dessous des captures de dame.

Les coups sont retournés **entièrement légaux**, triés par `Score` décroissant.

Barème des scores de tri :

| Situation | Score |
|---|---|
| Capture de dame | 7 |
| Coup donnant échec (silencieux) | 5 |
| Capture de tour | 4 |
| Capture de fou ou cavalier | 3 |
| Capture de pion | 2 |
| Roque | 1 |
| Coup silencieux | 0 |

---

### 9. `DoMove` / `UndoMove`

#### 9.1 Structure `Move`

```go
type Move struct {
    From, To  Square
    Promotion Piece   // EMPTY | CASTLEMOVE | KNIGHT | BISHOP | ROOK | QUEEN
    Score     uint8   // score brut pour le tri (non lié au score de position)

    // Champs de défaire (remplis par DoMove) :
    Captured      Piece   // signé : +blanc, -noir, EMPTY = aucun
    CaptureSquare Square  // == To, sauf en passant (case du pion capturé)
    PrevStatus    Status  // snapshot complet de Status avant le coup
    PrevEPFile    int8    // colonne du fantôme EP avant le coup ; -1 = aucun
    PrevHash      uint64  // hash Zobrist avant le coup
}
```

`GetMoveList` ne renseigne que `From`, `To`, `Promotion` et `Score`. `DoMove` renseigne tous les champs de défaire dans le `Move` retourné.

#### 9.2 `DoMove(m Move) (Position, Move)`

Retourne la nouvelle position **et** le coup enrichi des champs de défaire. Le `Move` retourné doit être transmis intact à `UndoMove`.

Cinq chemins de code selon `m.Promotion` :

| `m.Promotion` | Chemin |
|---|---|
| `CASTLEMOVE` | Déplace roi + tour ; révoque les droits de roque pour cette couleur |
| `KNIGHT/BISHOP/ROOK/QUEEN` | Supprime le pion source ; place la pièce promue à destination ; gère la capture éventuelle |
| `EMPTY` — pion change de colonne vers case vide | Prise en passant : supprime le pion capturé à la case adjacente |
| `EMPTY` — capture normale | Supprime la pièce capturée à destination |
| `EMPTY` — coup silencieux | Déplace la pièce ; gère les effets de bord roi/tour |

Après tout chemin, si une tour a été capturée, le droit de roque correspondant de l'adversaire est révoqué.

**Cycle de vie du fantôme EP dans DoMove** :

1. L'ancien fantôme est extrait du hash (XOR) puis nettoyé de `pawnOcc`.
2. Si le coup est une poussée double, le nouveau fantôme est placé à la case intermédiaire (`Sq(2, col)` pour blanc, `Sq(5, col)` pour noir) dans `pawnOcc` et intégré dans le hash (XOR). Aucune garde d'occupation n'est nécessaire : cette case est toujours libre.

#### 9.3 `UndoMove(m Move) Position`

Restaure la position exactement. Points clés :

- `pp.status = m.PrevStatus` restaure le trait, les cases des rois et tous les droits de roque en une seule affectation.
- `pp.Hash = m.PrevHash` restaure le hash Zobrist atomiquement — sans recalcul.
- Le fantôme EP est restauré en nettoyant tous les fantômes courants, puis en réinsérant celui enregistré dans `m.PrevEPFile` (le cas échéant).

---

### 10. Hachage Zobrist

#### 10.1 Table — `ZobristTable`

```go
ZobristBitboards [6][64]uint64   // index : 0=colOcc[B], 1=colOcc[N],
                                 //         2=pawnOcc, 3=rookOcc,
                                 //         4=bishopOcc, 5=knightOcc
ZobristKing      [2][64]uint64   // case du roi stockée dans Status
ZobristCastling  [2][4]uint64    // index = GetCastleBits()>>6 → 0–3
ZobristTurn      uint64          // XORé quand c'est au tour des noirs
```

`DefaultZT` est un singleton de package initialisé avec `crypto/rand` au démarrage. `StartPosition.Hash` en est dérivé via `init()`.

#### 10.2 Mise à jour incrémentale dans `DoMove`

Les composantes dépendantes du `Status` (bits de roque, cases des rois) utilisent un **schéma en accolade** :

```
XOR out : ZobristCastling[B][old], ZobristCastling[N][old],
          ZobristKing[B][old],     ZobristKing[N][old]
   ... effectuer tous les changements de bitboards et Status ...
XOR in  : ZobristCastling[B][new], ZobristCastling[N][new],
          ZobristKing[B][new],     ZobristKing[N][new]
```

Les révocations de droits de roque (déplacement de sa propre tour, capture d'une tour adverse) ne nécessitent aucun code hash spécifique : `revokeRookCastle` modifie `pp.status` et l'accolade fermante capte automatiquement la nouvelle valeur.

Tout le reste est XORé **en ligne** au fur et à mesure des changements de bitboards :

| Événement | Clés XORées |
|---|---|
| Changement de trait | `ZobristTurn` (inconditionnel) |
| Fantôme EP supprimé | `ZobristBitboards[pawnOcc][fantômeSq]` (accolade ouvrante) |
| Pièce couleur c déplacée from→to | `ZobristBitboards[c][from] ^ ZobristBitboards[c][to]` |
| Type de pièce change en sq | `ZobristBitboards[typeIdx][sq]` (une fois par sq) |
| Nouveau fantôme EP créé | `ZobristBitboards[pawnOcc][fantômeSq]` |

#### 10.3 Recalcul complet — `HashPosition`

Utilisé pour établir le hash de toute position non atteinte via `DoMove` (ex. positions chargées depuis FEN). Complexité O(popcount de tous les bitboards).

---

### 11. Évaluation statique — `Value()`

L'évaluation statique (profondeur 0) est exprimée du point de vue du joueur au trait (convention negamax) :

| Composante | Valeur |
|---|---|
| Mat (roi adverse prenable) | +30 000 |
| Pat | 0 |
| Reine | 90 |
| Tour | 50 |
| Fou / Cavalier | 30 |
| Pion | 10 |
| Contrôle du centre (d4/d5/e4/e5) | +1 par case contrôlée |
| Droits de roque disponibles | +1 par droit |
| Roi en échec | −10 |

Un score positif signifie un avantage pour le joueur qui doit jouer.

---

### 12. Empreinte mémoire

| Structure | Taille statique | Tas (runtime) |
|---|---|---|
| `Position` | 64 octets | 0 (type valeur) |
| `Status` | 3 octets | — |
| `Move` | 48 octets | 0 (type valeur) |
| `BigTable` (shell struct) | 7 168 octets | ~304 Ko total |
| `ZobristTable` | 4 168 octets | 0 |

Le coût tas de `BigTable` (~304 Ko) couvre les 64 × 4 maps directionnelles pour tours/fous plus les 2 × 64 maps de pions. Temps de construction : ~365 µs (coût unique au démarrage).

---

## Package `game` — Analyse

### 1. Negamax avec élagage alpha-bêta

L'algorithme central est un **negamax avec élagage alpha-bêta** (`AlphaBeta`). La convention negamax signifie que le score est toujours exprimé du point de vue du joueur au trait ; lors de l'appel récursif, les bornes et le score sont niés (`-beta, -alpha`).

```
AlphaBeta(α, β, profondeur) :
  1. Consultation de la table de transposition
     → Pruning possible si l'entrée est suffisamment profonde
  2. Si profondeur == 0 → évaluation statique
  3. Génération des coups légaux (filtrés et triés : captures par valeur de pièce,
     échecs au roi adverse +5, coups silencieux 0)
  4. Déplacement du meilleur coup TT en tête de liste
  5. Pour chaque coup :
       DoMove → appel récursif avec (-β, -α) → UndoMove
       Mise à jour de α et du meilleur coup
       Élagage si α ≥ β
  6. Stockage en table de transposition (EXACT / LOWER / UPPER)
```

### 2. Table de transposition — `ZMap`

Tableau de taille fixe (`ZSize` entrées, configurable à la compilation) indexé par `hash % ZSize`.

Chaque entrée contient :

| Champ | Description |
|---|---|
| `Best` | meilleur coup trouvé |
| `Score` + `ScoreType` | valeur et type de borne (EXACT / LOWER / UPPER) |
| `Depth` | profondeur d'analyse |
| `ConfirmH` | 16 bits supérieurs du hash (détection de collision) |
| `Age` | numéro du coup au moment du stockage (politique de remplacement) |

**Politique de remplacement :** même hash → on garde l'entrée la plus profonde ; collision de hash → on remplace si l'entrée stockée est plus ancienne.

**Tailles disponibles** (via build tags) :

| Tag | Entrées | Mémoire approx. |
|---|---|---|
| `low` | 1 M | ~90 Mo |
| *(défaut)* | 5 M | ~400 Mo |
| `high` | 50 M | ~4 Go |
| `ultra` | 500 M | ~25 Go |

### 3. Approfondissement itératif

`Analysis` appelle `AlphaBeta` en boucle avec une profondeur croissante (1, 2, 3, …). À chaque profondeur complète, le meilleur coup est sauvegardé dans `LastRootEntry`. Si l'analyse est interrompue (timeout ou coup joué), le résultat de la dernière profondeur *complète* est utilisé, garantissant un comportement *anytime*.

### 4. Annulation et concurrence

L'analyse tourne dans une goroutine séparée et reçoit un `context.Context`. Chaque nœud de la recherche vérifie `ctx.Err()` (lecture atomique ~5 ns, sans surcoût mesurable). En cas d'annulation, le nœud retourne immédiatement sans écrire en table de transposition, laissant la table dans l'état cohérent issu de la dernière profondeur entièrement terminée.

---

---

## Package `cache` / commande `gencache` — Cache pré-calculé

Le package `cache` et la commande `gencache` permettent de pré-calculer une table de transposition depuis la position initiale et de la sauvegarder sur disque, afin que le moteur démarre avec une table déjà chaude.

### Nommage des fichiers

```
mychess_cache_<N>M_<pct>.bin
```

- `<N>` : taille de la ZMap en millions d'entrées (dépend du build tag)
- `<pct>` : pourcentage de remplissage cible sur 2 chiffres (ex. `60` → table remplie à 60 %)

Exemples : `mychess_cache_5M_60.bin`, `mychess_cache_50M_10.bin`.

### Génération

```
gencache [-dir <répertoire>] [-fill <pct>]
```

| Flag | Défaut | Description |
|---|---|---|
| `-dir` | `.` | répertoire de sortie |
| `-fill` | `75` | pourcentage de remplissage cible (1–99) |

### Chargement au démarrage

`NewGame()` scanne les répertoires `./`, `./bin/` et `../` à la recherche du fichier de cache dont le pourcentage de remplissage est le plus élevé et dont le `ZSize` correspond au build courant. En cas de succès, la `ZobristTable` et la `ZMap` sont restaurées ; l'analyse repart immédiatement à la profondeur déjà explorée. Si aucun fichier compatible n'est trouvé, une table vide est utilisée.

### Format binaire

```
[8]byte   magic       "MYCHCACH"
uint32    version     1
uint64    ZSize       nombre d'entrées (vérifié contre le build)
uint32    fillPct     pourcentage de remplissage (informatif)
[…]       ZobristTable
[…]       ZMap.data
```

---

## Package `client` — Interface TUI

Interface textuelle interactive construite avec [BubbleTea v2](https://charm.land/bubbletea). Permet de jouer contre le moteur, de visualiser l'analyse en cours (profondeur, meilleur coup, score) et de naviguer dans l'historique de la partie.

Au démarrage, le client affiche `Memory model : <N>M, loading ...` pendant le chargement du cache. Une fois le TUI lancé, le terminal bascule sur le buffer alternatif (alt screen) : le message disparaît. À la sortie (`x`), le buffer principal est restauré.
