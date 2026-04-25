package position

// Regression tests for the EP phantom placement fix.
//
// Old design (buggy): phantom at rank 0 (white) / rank 7 (black) — those
// squares can be occupied by pieces (e.g. the king on e1 when white plays
// e2-e4), silently preventing the phantom from being placed and making en
// passant unavailable even though it should be legal.
//
// New design (correct): phantom at the intermediate square — rank 2 for
// white's double push, rank 5 for black's double push. That square is always
// empty because the pawn just traversed it, so no placement guard is needed
// and EP is always available after a double push regardless of what sits on
// the back rank.
//
// Each test:
//  1. Calls DoMove for the double push with a piece occupying the old phantom
//     square (rank 0 / rank 7).
//  2. Verifies the resulting position: pawn at destination, blocking piece
//     intact, phantom present at the intermediate square, hash consistent.
//  3. Verifies that the EP capture IS in the opponent's move list.
//  4. Calls UndoMove and verifies the original position is restored exactly.

import "testing"

// phantoms returns bits set in pawnOcc but not in colOcc (EP phantoms).
func phantoms(p Position) Bitboard {
	return p.pawnOcc & ^(p.colOcc[WHITE] | p.colOcc[BLACK])
}

// assertPhantomAt fails if no EP phantom is present at sq.
func assertPhantomAt(t *testing.T, p Position, sq string, label string) {
	t.Helper()
	s := SqParse(sq)
	if !phantoms(p).IsSet(s) {
		t.Errorf("[%s] expected EP phantom at %s, pawnOcc=%016x colOcc=%016x",
			label, sq, uint64(p.pawnOcc), uint64(p.colOcc[WHITE]|p.colOcc[BLACK]))
	}
}

// assertHasMove fails if the move from→to is absent from p's legal move list.
func assertHasMove(t *testing.T, p Position, from, to string, label string) {
	t.Helper()
	f, tk := SqParse(from), SqParse(to)
	for _, m := range p.GetMoveList() {
		if m.From == f && m.To == tk {
			return
		}
	}
	t.Errorf("[%s] move %s→%s not found in move list", label, from, to)
}

// assertDoMoveUndoMove performs the DoMove/UndoMove round-trip, checking hash
// consistency, and returns the intermediate position for further inspection.
func assertDoMoveUndoMove(t *testing.T, orig Position, m Move, label string) (after Position) {
	t.Helper()
	orig.Hash = DefaultZT.HashPosition(orig)

	after, done := orig.DoMove(m)

	if want := DefaultZT.HashPosition(after); after.Hash != want {
		t.Errorf("[%s] DoMove: incremental hash %016x != full hash %016x", label, after.Hash, want)
	}

	restored := after.UndoMove(done)
	assertPositionsEqual(t, orig, restored, label)
	if restored.Hash != orig.Hash {
		t.Errorf("[%s] UndoMove: hash %016x != original %016x", label, restored.Hash, orig.Hash)
	}
	return after
}

// ── White double push ─────────────────────────────────────────────────────────

// TestEP_White_OwnKingOnOldPhantomSquare verifies that white's e2-e4 creates
// a phantom at e3 (rank 2) even though the white king occupies e1 (the old
// rank-0 phantom square). Black's adjacent pawn on d4 must be able to capture
// en passant to e3.
//
//	8  . . . . . . . k
//	4  . . . p . . . .   (black pawn d4 — would capture to e3)
//	2  . . . . P . . .   (white pawn e2 — to push)
//	1  . . . . K . . .   (white king e1 — was blocking phantom in old design)
func TestEP_White_OwnKingOnOldPhantomSquare(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "h8").
		AddPawn(WHITE, "e2").AddPawn(BLACK, "d4")

	after := assertDoMoveUndoMove(t, orig, mkMove("e2", "e4", EMPTY),
		"white e2-e4, king on e1")

	if !after.pawnOcc.IsSet(SqParse("e4")) || !after.colOcc[WHITE].IsSet(SqParse("e4")) {
		t.Error("white pawn not at e4 after double push")
	}
	if after.KingPosition(WHITE) != SqParse("e1") {
		t.Error("white king not on e1 after push")
	}
	assertPhantomAt(t, after, "e3", "white e2-e4 king on e1")
	assertHasMove(t, after, "d4", "e3", "EP d4xe3 available despite king on e1")
}

// TestEP_White_EnemyRookOnOldPhantomSquare verifies that white's d2-d4 creates
// a phantom at d3 (rank 2) even though a black rook occupies d1 (old rank-0
// square). Black's adjacent pawn on c4 must be able to capture en passant to d3.
//
//	8  . . . . . . . k
//	4  . . p . . . . .   (black pawn c4 — would capture to d3)
//	2  . . . P . . . .   (white pawn d2 — to push)
//	1  . . . r . K . .   (black rook d1 — was blocking phantom; white king f1)
func TestEP_White_EnemyRookOnOldPhantomSquare(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "f1").AddKing(BLACK, "h8").
		AddPawn(WHITE, "d2").AddPawn(BLACK, "c4").
		AddRook(BLACK, "d1")

	after := assertDoMoveUndoMove(t, orig, mkMove("d2", "d4", EMPTY),
		"white d2-d4, black rook on d1")

	if !after.pawnOcc.IsSet(SqParse("d4")) || !after.colOcc[WHITE].IsSet(SqParse("d4")) {
		t.Error("white pawn not at d4 after double push")
	}
	if !after.rookOcc.IsSet(SqParse("d1")) || !after.colOcc[BLACK].IsSet(SqParse("d1")) {
		t.Error("black rook not on d1 after push")
	}
	assertPhantomAt(t, after, "d3", "white d2-d4 black rook on d1")
	assertHasMove(t, after, "c4", "d3", "EP c4xd3 available despite rook on d1")
}

// ── Black double push ─────────────────────────────────────────────────────────

// TestEP_Black_OwnKingOnOldPhantomSquare verifies that black's e7-e5 creates
// a phantom at e6 (rank 5) even though the black king occupies e8 (the old
// rank-7 phantom square). White's adjacent pawn on d5 must be able to capture
// en passant to e6.
//
//	8  . . . . k . . .   (black king e8 — was blocking phantom in old design)
//	7  . . . . p . . .   (black pawn e7 — to push)
//	5  . . . P . . . .   (white pawn d5 — would capture to e6)
//	1  . . . . . . . K
func TestEP_Black_OwnKingOnOldPhantomSquare(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "h1").AddKing(BLACK, "e8").
		AddPawn(BLACK, "e7").AddPawn(WHITE, "d5")
	orig.status.SetTurn(BLACK)

	after := assertDoMoveUndoMove(t, orig, mkMove("e7", "e5", EMPTY),
		"black e7-e5, king on e8")

	if !after.pawnOcc.IsSet(SqParse("e5")) || !after.colOcc[BLACK].IsSet(SqParse("e5")) {
		t.Error("black pawn not at e5 after double push")
	}
	if after.KingPosition(BLACK) != SqParse("e8") {
		t.Error("black king not on e8 after push")
	}
	assertPhantomAt(t, after, "e6", "black e7-e5 king on e8")
	assertHasMove(t, after, "d5", "e6", "EP d5xe6 available despite king on e8")
}

// TestEP_Black_EnemyRookOnOldPhantomSquare verifies that black's d7-d5 creates
// a phantom at d6 (rank 5) even though a white rook occupies d8 (old rank-7
// square). White's adjacent pawn on c5 must be able to capture en passant to d6.
//
//	8  . . . R . . . k   (white rook d8 — was blocking phantom; black king h8)
//	7  . . . p . . . .   (black pawn d7 — to push)
//	5  . . P . . . . .   (white pawn c5 — would capture to d6)
//	1  . . . . . . . K
func TestEP_Black_EnemyRookOnOldPhantomSquare(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "h1").AddKing(BLACK, "h8").
		AddPawn(BLACK, "d7").AddPawn(WHITE, "c5").
		AddRook(WHITE, "d8")
	orig.status.SetTurn(BLACK)

	after := assertDoMoveUndoMove(t, orig, mkMove("d7", "d5", EMPTY),
		"black d7-d5, white rook on d8")

	if !after.pawnOcc.IsSet(SqParse("d5")) || !after.colOcc[BLACK].IsSet(SqParse("d5")) {
		t.Error("black pawn not at d5 after double push")
	}
	if !after.rookOcc.IsSet(SqParse("d8")) || !after.colOcc[WHITE].IsSet(SqParse("d8")) {
		t.Error("white rook not on d8 after push")
	}
	assertPhantomAt(t, after, "d6", "black d7-d5 white rook on d8")
	assertHasMove(t, after, "c5", "d6", "EP c5xd6 available despite rook on d8")
}
