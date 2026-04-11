# `position` package — internal representation

This document describes how chess positions, moves, and attack tables are
represented and computed internally.

---

## 1. Square indexing

A `Square` is a `uint8` in the range `[0, 63]`.

Squares are numbered **rank-first**, starting from `a1 = 0`:

```
rank 7  56 57 58 59 60 61 62 63     a8 b8 c8 d8 e8 f8 g8 h8
rank 6  48 49 50 51 52 53 54 55     a7 b7 ...
  ...
rank 1   8  9 10 11 12 13 14 15     a2 b2 ...
rank 0   0  1  2  3  4  5  6  7     a1 b1 c1 d1 e1 f1 g1 h1
         ^                  ^
       file 0             file 7
       (a file)           (h file)
```

Formula: `Square = rank * 8 + file`, both 0-based.

Constructor: `Sq(rank, file int) Square`.  
Parsing:     `SqParse("d4") → Square(27)`.

---

## 2. Bitboard

A `Bitboard` is a `uint64` where bit `n` corresponds to `Square(n)`.

```
Bit 63 = h8    ...    Bit 0 = a1
```

A set bit means "something is here". Bitboards are the core data structure
for representing sets of squares efficiently: unions, intersections, and
exclusions are single CPU instructions.

Key operations (all return a new `Bitboard`):

| Method | Description |
|--------|-------------|
| `.Set(sq)` | set a bit |
| `.Unset(sq)` | clear a bit |
| `.IsSet(sq)` | test a bit |
| `.Get(sq)` | returns 0 or 1 |
| `.BitCount()` | popcount |
| `.AllSetSquares(yield)` | range-over-func iterator |
| `.AllBitCombinations(yield)` | enumerate all subsets |

Constructors: `Rank(r)`, `File(f)`, `Diagonal(sq)`, `AntiDiagonal(sq)`,
`Interior()`, `Full()`, `Border()`.

---

## 3. Position — the bitboard array approach

Instead of one array of 64 piece values, `Position` uses **six independent
bitboards**, each tracking one property across all 64 squares simultaneously.

```go
type Position struct {
    colOcc    [2]Bitboard  // which squares each color occupies
    pawnOcc   Bitboard     // squares with a pawn (either color, + EP phantom)
    rookOcc   Bitboard     // squares with a rook or queen (either color)
    bishopOcc Bitboard     // squares with a bishop or queen (either color)
    knightOcc Bitboard     // squares with a knight (either color)
    status    Status
    Hash      uint64       // Zobrist hash; 0 = uninitialized
}
```

### 3.1 Color occupancy — `colOcc[2]`

`colOcc[WHITE]` has a bit set for every square occupied by a white piece
(including the king).  
`colOcc[BLACK]` is the same for black.

The union `colOcc[WHITE] | colOcc[BLACK]` gives all occupied squares.

### 3.2 Piece-type occupancy

Each of `pawnOcc`, `rookOcc`, `bishopOcc`, `knightOcc` is **color-agnostic**:
a bit is set regardless of which player owns the piece.

To determine the color of a piece at a given square, intersect with `colOcc`:

```
white rooks = rookOcc & colOcc[WHITE]   (excluding queens)
```

### 3.3 Queens — implicit representation

There is **no separate `queenOcc`**.  A queen is stored as *both* a rook and
a bishop: when a queen is placed on `sq`, both `rookOcc` and `bishopOcc` have
their bit set at `sq`.

Identifying a queen at `sq`:

```
isQueen(sq) = rookOcc.IsSet(sq) && bishopOcc.IsSet(sq)
```

This makes move generation natural: queen moves = rook moves ∪ bishop moves.

### 3.4 Kings — stored in `Status`

Kings are **not** tracked in any of the piece-type bitboards.  Their squares
are embedded in the `Status` struct (6 bits each inside `KingStatus[color]`).
King presence in `colOcc` is still set normally.

---

## 4. Status word

```go
type Status struct {
    KingStatus [2]uint8   // per-color: castling bits (2) + king square (6)
    TurnStatus uint8      // bit 0: whose turn (WHITE=0, BLACK=1)
}
```

### 4.1 `KingStatus[color]` layout

```
bit 7  bit 6  bits 5-0
  │      │       │
  │      │       └── king square (0-63)
  │      └────────── can castle queen-side?
  └───────────────── can castle king-side?
```

Constants: `CanCastleKingSide = 0b10000000`, `CanCastleQueenSide = 0b01000000`.

---

## 5. En passant — phantom pawn

When a pawn makes a double push, an en passant opportunity must be recorded
for the opponent's next move.  Rather than a separate field, this is encoded
as a **phantom pawn**: a bit set in `pawnOcc` with **no corresponding bit**
in either `colOcc`.

The phantom square encodes the **file** of the double-pushed pawn; the rank
encodes **which side** created the opportunity:

| Side that double-pushed | Phantom rank | Phantom square formula |
|---|---|---|
| WHITE | rank 0 | `Sq(0, file)` |
| BLACK | rank 7 | `Sq(7, file)` |

Ranks 0 and 7 can never hold real pawns (they would have promoted), so a
pawn bit there without a corresponding `colOcc` bit is unambiguously an
en passant marker.

**Detection** (in `GetPawnMovesFromSquareBB`):

```
phantoms = pawnOcc & ^(colOcc[WHITE] | colOcc[BLACK])
```

For a white pawn at rank 4 (its 5th rank), a phantom at `Sq(7, file±1)` means
black just double-pushed on the adjacent file. The en passant landing square
is `Sq(5, file±1)` (one rank above the black pawn's current position).

**Clearing** (`ResetEnPassantFlag`): `pawnOcc &= colOcc[WHITE] | colOcc[BLACK]`
— removes all phantom bits in one operation.

**Placement guard**: the phantom is only written when its target square is
unoccupied.  If a piece already sits on rank 0/7 at the same file (e.g. the
king on e1 while the e-pawn double-pushes), the phantom would be masked out
by `colOcc` and would corrupt `pawnOcc` on undo; it is silently skipped and
the en passant opportunity for that specific case is not recorded.

---

## 6. Piece identification at a square — `PieceAt`

```
color = colOcc[WHITE].Get(sq) - colOcc[BLACK].Get(sq)   → +1, -1, or 0
```

If `color == 0` → empty.  Otherwise, look up the type bitboards in order:
pawn → knight → queen (rook∧bishop) → bishop → rook → king (from `Status`).

---

## 7. BigTable — precomputed attack tables

`BigTable` is an immutable structure built once at startup by `NewBigTable()`.
It contains every attack set that any piece can have from any square, indexed
by the relevant occupancy bits.  Once built, **all move lookups are
allocation-free map reads**.

### 7.1 Sliding pieces — per-direction design

Instead of a single combined mask per piece, sliding attacks are split by
**direction** into four independent tables:

| Piece | Direction | Mask field | Map field |
|---|---|---|---|
| Rook | rank (E/W) | `RookMaskRank[sq]` | `RookAttackSetRank[sq]` |
| Rook | file (N/S) | `RookMaskFile[sq]` | `RookAttackSetFile[sq]` |
| Bishop | NE/SW diagonal | `BishopMaskNE[sq]` | `BishopAttackSetNE[sq]` |
| Bishop | NW/SE diagonal | `BishopMaskNW[sq]` | `BishopAttackSetNW[sq]` |

**Why per-direction?**  Each direction's mask is smaller (max 6 bits for a
rank mask vs. 12 bits for a full rook mask).  Smaller masks → fewer map
entries → better cache utilisation.  Total attacks are recovered by OR-ing the
two direction results.

**Mask construction** (rook rank example):

```
RookMaskRank[sq] = Rank(r).Unset(sq).Unset(Sq(r,0)).Unset(Sq(r,7))
```

The square itself and the two border files are excluded: border squares are
always reachable regardless of occupancy, so including them in the key wastes
map entries without adding information.

**Lookup** (rook at `sq`):

```go
occ          := colOcc[WHITE] | colOcc[BLACK]
rankAttacks  := RookAttackSetRank[sq][occ & RookMaskRank[sq]]
fileAttacks  := RookAttackSetFile[sq][occ & RookMaskFile[sq]]
attacks      := (rankAttacks | fileAttacks) & ^colOcc[turn]
```

### 7.2 Non-sliding pieces

`KingAttacks[sq]` and `KnightAttacks[sq]` are plain `[64]Bitboard` arrays —
no occupancy key needed.

### 7.3 Pawn attack maps

```go
PawnMask[color][sq]      Bitboard                       // relevant squares
PawnAttackSet[color][sq] map[Bitboard]Bitboard          // occ → moves
```

The mask combines **forward move squares** (1 or 2 ahead) and **diagonal
capture squares**.  The map is keyed by `totalOcc & PawnMask`:

- A forward square appears in the value if and only if it is **unoccupied** in
  the key.  For a pawn on its starting rank, the double-push square is omitted
  if the intermediate square is occupied (correct blocking, no jumping).
- A capture square appears in the value if and only if it is **occupied** by
  any piece.  The caller then filters with `& ^colOcc[turn]` to exclude own
  pieces, leaving only opponent captures.

En passant is **not** in the map; it is computed separately in
`GetPawnMovesFromSquareBB` using the phantom-pawn detection described in §5.

---

## 8. Move generation

### 8.1 `GetMovesBB(bt, sq) → Bitboard`

Identifies the piece type at `sq` by checking the type bitboards in order,
then dispatches to the appropriate piece handler.  Returns a `Bitboard` of all
reachable squares (pseudo-legal — does not filter self-checks).

### 8.2 `GetMoveList(bt) → []Move`

Iterates over all squares in `colOcc[turn]`, calls `GetMovesBB` for each,
and unpacks the resulting bitboard into `Move` structs.  Promotions are
expanded into four moves (Q/R/B/N) inline.  Castling is appended via
`GetCastlingMoveList` which checks:

1. Castle rights bits in `Status`.
2. The king is not currently in check.
3. All intermediate squares are unoccupied.
4. No intermediate square is attacked by the opponent.

Moves are returned sorted by a simple capture-value score (captures first).

**Pseudo-legal only**: moves that leave the own king in check are not filtered.
The caller is responsible for verifying legality, typically with:

```go
func (p Position) IsLegal(bt *BigTable, m Move) bool {
    after, _ := p.DoMove(m)
    turn := p.status.GetTurn()
    return !after.IsSquareAttacked(bt, after.status.GetKingPosition(turn), 1-turn)
}
```

---

## 9. DoMove / UndoMove

### 9.1 The `Move` struct

```go
type Move struct {
    From, To  Square
    Promotion Piece   // EMPTY | CASTLEMOVE | KNIGHT | BISHOP | ROOK | QUEEN

    // Populated by DoMove; required by UndoMove:
    Captured      Piece   // signed: +white, -black, EMPTY = none
    CaptureSquare Square  // == To, except en passant (where it's the pawn's square)
    PrevStatus    Status  // full Status snapshot before the move
    PrevEPFile    int8    // file of the EP phantom before the move; -1 = none
    PrevHash      uint64  // Zobrist hash before the move
}
```

`GetMoveList` fills only `From`, `To`, `Promotion`, and `Score`.
`DoMove` fills all undo fields in the returned `Move`.

### 9.2 `DoMove(m Move) (Position, Move)`

Returns the new position **and** the move enriched with undo fields.  The
returned `Move` must be passed intact to `UndoMove`.

Handles five code paths:

| `m.Promotion` | Path |
|---|---|
| `CASTLEMOVE` | Move king + rook; clear castle rights for that color |
| `KNIGHT/BISHOP/ROOK/QUEEN` | Remove pawn at source; place promoted piece at dest; handle capture |
| `EMPTY` — pawn changes file to empty square | En passant: remove captured pawn from adjacent square |
| `EMPTY` — normal capture | Remove captured piece at dest |
| `EMPTY` — quiet move | Move piece; handle king/rook side-effects |

After every path, if a rook was captured, the opponent's corresponding castle
right is revoked.

**EP phantom lifecycle inside DoMove**:

1. Old phantom is XORed out of the hash, then cleared from `pawnOcc`.
2. If the move is a double pawn push and the phantom target square is free, the
   new phantom is set in `pawnOcc` and XORed into the hash.

### 9.3 `UndoMove(m Move) Position`

Restores the position exactly.  Key points:

- `pp.status = m.PrevStatus` restores turn, king squares, and all castle bits
  in a single assignment.
- `pp.Hash = m.PrevHash` restores the Zobrist hash atomically — no
  re-computation.
- The EP phantom is restored by clearing all current phantoms, then
  re-inserting the one recorded in `m.PrevEPFile` (if any).

---

## 10. Zobrist hashing

### 10.1 Table layout (`ZobristTable`)

```go
ZobristBitboards [6][64]uint64   // index: 0=colOcc[W], 1=colOcc[B],
                                 //        2=pawnOcc, 3=rookOcc,
                                 //        4=bishopOcc, 5=knightOcc
ZobristKing      [2][64]uint64   // king square stored in Status
ZobristCastling  [2][4]uint64    // castle-bits index = GetCastleBits()>>6 → 0–3
ZobristTurn      uint64          // XORed in when it is BLACK's turn
```

`DefaultZT` is a package-level singleton initialised with `crypto/rand` at
startup.  `StartPosition.Hash` is seeded from it via `init()`.

### 10.2 Incremental update strategy in `DoMove`

Status-dependent components (castle bits, king squares) use a **bracket**
pattern:

```
XOR out: ZobristCastling[W][old], ZobristCastling[B][old],
         ZobristKing[W][old],     ZobristKing[B][old]
   ... make all bitboard and Status changes ...
XOR in:  ZobristCastling[W][new], ZobristCastling[B][new],
         ZobristKing[W][new],     ZobristKing[B][new]
```

This means castle-right revocations (own rook move, opponent rook capture)
require no special-case hash code — `revokeRookCastle` modifies `pp.status`
and the closing bracket picks up the new value automatically.

Everything else is XORed **inline** as each bitboard changes:

| Event | Keys XORed |
|---|---|
| Turn flip | `ZobristTurn` (unconditional — turn always changes) |
| EP phantom removed | `ZobristBitboards[pawnOcc][phantomSq]` (opening bracket) |
| Piece moves color c, from→to | `ZobristBitboards[c][from] ^ ZobristBitboards[c][to]` |
| Piece type changes at sq | `ZobristBitboards[typeIdx][sq]` (once per sq) |
| New EP phantom created | `ZobristBitboards[pawnOcc][phantomSq]` |

### 10.3 Full recomputation — `HashPosition(p)`

Used to establish the hash for any position not reached via `DoMove` (e.g.
positions loaded from FEN).  O(popcount of all bitboards).

---

## 11. Memory layout summary

| Structure | Static size | Runtime (heap) |
|---|---|---|
| `Position` | 64 bytes | 0 (value type) |
| `Status` | 3 bytes | — |
| `Move` | 48 bytes | 0 (value type) |
| `BigTable` (struct shell) | 7 168 bytes | ~304 KB total |
| `ZobristTable` | 4 168 bytes | 0 |

`BigTable` heap cost: ~304 KB for all 64 × 4 direction maps for rooks/bishops
plus 2 × 64 pawn maps.  Build time: ~365 µs (one-time cost at startup).
