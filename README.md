# mychess
experimental chess engine in go

## Attack tables dimensionning 

Number of occupancies vs. Number of attack sets.

| | bishop | | | |  | rook  | | | | |
|---|---|---|---|---|---|---|---|---|---|---|
| | A | B | C | D | || A | B | C | D |
| 8 | 64/7 | 32/6 | 32/10 | 32/12 | | 8 | 4096/49 | 2048/42 | 2048/70 | 2048/84 |
| 7 | 32/6 | 32/6 | 32/10 | 32/12 | | 7 | 2048/42 | 1024/36 | 1024/60 | 1024/72 |
| 6 | 32/10 | 32/10 | 128/40 | 128/48 | | 6 | 2048/70 | 1024/60 | 1024/100 | 1024/120 |
| 5 | 32/12 | 32/12 | 128/48 | 512/108 | | 5 | 2048/84 | 1024/72 | 1024/120 | 1024/144 |
| | | | | | | | | | |
| | A | B | C | D | || A | B | C | D |
| 8 | 9.14 | 5.33 | 3.20 | 2.67 | | 8 | 83.59 | 48.76 | 29.26 | 24.38 |
| 7 | 5.33 | 5.33 | 3.20 | 2.67 | | 7 | 48.76 | 28.44 | 17.07 | 14.22 |
| 6 | 3.20 | 3.20 | 3.20 | 2.67 | | 6 | 29.26 | 17.07 | 10.24 | 8.53 |
| 5 | 2.67 | 2.67 | 2.67 | 4.74 | | 5 | 24.38 | 14.22 | 8.53 | 7.11 |

## Changements à faire 

Ne pas faire de miroir de position mais avoir une variable qui dit à qui c'est le tour (0 ou 1) dans la position, 
Indexer les tableau des occupancies white/black pour passer facilement de l'un à l'autre sans transposer le reste.
Dupliquer les tables pawnMask pour les noirs et les blancs, sous forme de tableaux à double index
Réecrire pawnAttacks specifique pour les noirs 
Reecrire les fonctions GetMoves en tenant compte du joueur qui doit jouer.

Prevoir un hash de position pour detecter les répétitions (6 derniers ply indentiques) (inclure position, castle, ... mais PAS les compteurs de coup !)

## Poursuite de la conception

Implementer les roques

Implementer la prise en passant.

Iterateur sur une poistion qui sort les positions filles.
Meta iterateur qui filtre les positions illegales.

Gestion des arrêts de parties (mats, draw, repetition/echec perpetuel, ...)

Puis passer aux Nodes, avec la notion de valeur de la position ...