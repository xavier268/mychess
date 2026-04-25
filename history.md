# Historique des performances — mychess

**Plateforme de mesure :** Intel Core i7-10700 @ 2.90GHz · Windows/amd64 · 16 CPU logiques  
**Sources :** fichiers `bench-<VERSION>-<DATE>.txt` générés par `go test -bench=. -benchmem`

---

## Repères de lecture

Les deux benchmarks principaux du package `game` mesurent la force de calcul de l'IA :

- **`BenchmarkAnalysis/depth=N`** — recherche complète par approfondissement itératif (iterative deepening) jusqu'à la profondeur N depuis la position initiale, table de transposition froide (nouveau `Game` à chaque itération).
- **`BenchmarkAlphaBeta/depth=N`** — appel alpha-bêta unique à profondeur fixe, sans bénéfice du cache des profondeurs inférieures.

Dans le package `position` :

- **`BenchmarkGetMoveList_Baseline`** — coût de génération d'une liste de coups depuis une position de milieu de partie représentative.
- **`BenchmarkGetMovesBB_*`** — coût de génération du bitboard de déplacements pour chaque type de pièce.

La colonne **`B/op`** reflète la mémoire allouée par itération de benchmark ; elle est dominée par la taille statique de la table de transposition (`ZMap`) embarquée dans la struct `Game`.

---

## v0.3.1 — Référence initiale

**Contenu :** moteur jouable avec alpha-bêta et table de transposition Zobrist (1 million d'entrées, ~90 Mo).  
La génération de coups est *pseudo-légale* : `GetMoveList` retourne tous les coups apparemment valides sans vérifier si le roi reste en échec ; ce filtrage est fait *a posteriori* dans la boucle alpha-bêta (`DoMove` → `IsSquareAttacked` → `UndoMove`).

| Benchmark | Valeur |
|---|---|
| Analysis depth=6 | 375 ms/op · 144 MB/op · 233 K allocs |
| AlphaBeta depth=6 | 380 ms/op · 146 MB/op · 237 K allocs |
| GetMoveList Baseline | **1 413 ns/op** |

---

## v0.3.2 — Détection des fins de partie

**Changements :** ajout de la détection mat/pat dans le client (ckck for mat/pat, gestion des fins de partie).  
Aucune modification du moteur de recherche ni de la génération de coups.

**Impact sur les performances :** négligeable — les chiffres restent dans la marge de bruit statistique (±5 %).

| Benchmark | v0.3.1 | v0.3.2 | Δ |
|---|---|---|---|
| Analysis depth=6 | 375 ms | 391 ms | +4 % (bruit) |
| GetMoveList Baseline | 1 413 ns | 1 826 ns | +29 % (bruit) |

---

## v0.3.3 — Table de transposition ×10

**Changements :** la `ZMap` passe de **1 million à 10 millions d'entrées** (`ZSize = 10_000_000`). Objectif : moins de collisions dans la table Zobrist, meilleure réutilisation des positions calculées. Des fins de partie incorrectes sont également corrigées.

**Impact sur les performances :**

La mémoire allouée par itération **bondit de 40 MB à 200 MB** (+5×) : la `ZMap` est un tableau statique embarqué dans la struct `Game`, elle est entièrement allouée à froid à chaque itération de benchmark.

En revanche, le **temps d'exécution reste quasi identique** malgré cette surcharge mémoire : à profondeur 6, l'analyse passe de 375 ms à 385 ms (+3 %). Le gain en qualité de cache Zobrist compense l'overhead d'initialisation, et le benchmarking à froid (cold start) ne reflète pas l'avantage d'une table plus grande sur une vraie partie.

| Benchmark | v0.3.2 | v0.3.3 | Δ |
|---|---|---|---|
| Analysis depth=6 | 391 ms · 144 MB | 385 ms · **303 MB** | –1 % · **+110 % mémoire** |
| GetMoveList Baseline | 1 826 ns | 1 541 ns | –16 % |

---

## v0.3.5 — Modèles mémoire configurables

**Changements :** introduction de quatre tailles de `ZMap` sélectionnables par build tag :

| Tag | ZSize | Mémoire indicative |
|---|---|---|
| `low` | 1 M | ~90 Mo |
| *(défaut)* | 5 M | ~400 Mo |
| `high` | 50 M | ~4 Go |
| `ultra` | 500 M | ~25 Go |

Les benchmarks ci-dessous sont compilés avec le tag par défaut (5 M entrées). La taille `ZEntry` a également été ajustée, maintenant la consommation mémoire observée à ~200 MB/op malgré la réduction de `ZSize`.

**Impact sur les performances :** légère régression sur la vitesse de recherche et sur la génération de coups, probablement liée au remaniement des structs de taille.

| Benchmark | v0.3.3 | v0.3.5 | Δ |
|---|---|---|---|
| Analysis depth=6 | 385 ms | **433 ms** | +13 % |
| AlphaBeta depth=6 | 393 ms | **446 ms** | +14 % |
| GetMoveList Baseline | 1 541 ns | **2 355 ns** | +53 % |

---

## v0.3.6 — Renommage de module, corrections de bugs

**Changements :** renommage du module de `mychess` en `github.com/xavier268/mychess`. Correction d'un bug dans la capture de références de bits (*build ref capture*). Suppression de code mort, amélioration de l'affichage des statistiques.

**Impact sur les performances :** **récupération partielle** des régressions de v0.3.5, sans doute grâce à la suppression du code mort et à la correction du bug de capture.

| Benchmark | v0.3.5 | v0.3.6 | Δ |
|---|---|---|---|
| Analysis depth=6 | 433 ms | **405 ms** | –6 % |
| AlphaBeta depth=6 | 446 ms | **412 ms** | –8 % |
| GetMoveList Baseline | 2 355 ns | **1 785 ns** | –24 % |

---

## v0.4.0 — Filtrage légal systématique dans `GetMoveList`

**Changements :** refactoring architectural majeur. Jusqu'ici, `GetMoveList` générait des coups *pseudo-légaux*, et la boucle alpha-bêta filtrait les illégaux après `DoMove` + `IsSquareAttacked` + `UndoMove`. Désormais, **`GetMoveList` retourne uniquement des coups légaux**, en exécutant ce filtrage en interne pour chaque candidat. Un **bonus de score est également attribué aux coups qui donnent échec** (`checkBonus = 5`), améliorant l'ordre d'exploration.

En contrepartie, la boucle alpha-bêta est simplifiée (suppression de ~30 lignes de code de filtrage + debug).

**Impact sur les performances :** c'est la modification la plus coûteuse de l'historique en termes de vitesse brute.

| Benchmark | v0.3.6 | v0.4.0 | Δ |
|---|---|---|---|
| **GetMoveList Baseline** | 1 785 ns | **6 730 ns** | **+277 %** |
| Analysis depth=6 | 405 ms · 302 MB | **677 ms** · 282 MB | **+67 %** · –7 % mémoire |
| AlphaBeta depth=6 | 412 ms | **724 ms** | **+76 %** |
| Analysis depth=5 | 29 ms | **46 ms** | +58 % |

`GetMoveList` est maintenant presque 4× plus lent : pour chaque coup candidat, il exécute `DoMove`, teste si le roi est attaqué, puis `UndoMove`. La recherche globale ralentit de 67–76 % à grande profondeur.

En revanche, la **mémoire diminue de 7 % à profondeur 6** (302 MB → 282 MB) et le nombre d'allocations passe de 231 K à 217 K (–6 %) : le meilleur ordonnancement des coups (bonus d'échec) permet à l'élagage alpha-bêta de couper plus tôt, explorant moins de nœuds totaux.

---

## v0.4.1 — Correction du bug en passant

**Changements :** correction d'un bug dans la gestion de la prise *en passant* (EP) dans `position/move.go` et `position/movesBB.go`. Le moteur explorait jusqu'ici un sous-ensemble incorrect de positions EP, ce qui réduisait artificiellement l'espace de recherche.

**Impact sur les performances :** nouvelle dégradation de la vitesse, attendue et saine — le moteur explore maintenant les positions EP qu'il omettait auparavant.

| Benchmark | v0.4.0 | v0.4.1 | Δ |
|---|---|---|---|
| GetMoveList Baseline | 6 730 ns | **7 630 ns** | +13 % |
| Analysis depth=6 | 677 ms · 217 K allocs | **822 ms** · **244 K allocs** | +21 % · +12 % |
| AlphaBeta depth=6 | 724 ms | **774 ms** | +7 % |
| Analysis depth=4 | 12.8 ms · 1 599 allocs | 14.3 ms · **1 727 allocs** | +12 % · +8 % |

L'augmentation des allocations à toutes les profondeurs confirme que le correctif ouvre de véritables branches de jeu supplémentaires dans l'arbre de recherche.

---

## Synthèse

```
GetMoveList Baseline (ns/op) :

v0.3.1  ████ 1 413
v0.3.2  █████ 1 826
v0.3.3  ████ 1 541
v0.3.5  ███████ 2 355
v0.3.6  █████ 1 785
v0.4.0  ████████████████████ 6 730   ← filtrage légal dans GetMoveList
v0.4.1  ██████████████████████ 7 630  ← correction EP

Analysis depth=6 (ms/op) :

v0.3.1  ████ 375
v0.3.2  ████ 391
v0.3.3  ████ 385
v0.3.5  █████ 433
v0.3.6  ████ 405
v0.4.0  ████████ 677   ← coût du filtrage légal systématique
v0.4.1  █████████ 822  ← coût du correctif EP (plus de positions explorées)
```

Les choix de conception reflètent un arbitrage explicite : la v0.4.0 sacrifie ~70 % de vitesse brute pour garantir la légalité des coups dès leur génération et améliorer l'ordre d'exploration ; la v0.4.1 accepte un surcoût supplémentaire pour corriger une règle du jeu. La piste d'optimisation naturelle est de rendre le filtrage légal plus économique (tables d'épinglage, détection de coups *absolument légaux* sans Do/Undo complet).
