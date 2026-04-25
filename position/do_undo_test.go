package position

// Round-trip tests for DoMove / UndoMove.
//
// Each test:
//  1. Constructs a specific position.
//  2. Builds a Move (without undo fields).
//  3. Calls DoMove → records enriched move + new position.
//  4. Calls UndoMove on the new position → must equal the original exactly.
//
// Categories:
//   A. Normal moves   – single push, double push, simple capture
//   B. En passant     – white captures, black captures; EP phantom persists through a non-EP move
//   C. Castling       – kingside + queenside, both colors
//   D. Promotion      – push + capture promotions
//   E. Castle-right revocation via rook move or rook capture
//   F. Sequential     – two moves and two undos must recover original

import (
	"fmt"
	"testing"
)

// ── helper ────────────────────────────────────────────────────────────────────

// assertRoundTrip verifies:
//  1. DoMove followed by UndoMove returns the exact original position.
//  2. After DoMove the incremental hash matches the from-scratch hash.
//  3. After UndoMove the hash is restored to the original.
func assertRoundTrip(t *testing.T, orig Position, m Move, label string) {
	t.Helper()

	// Test positions built with the builder have Hash=0.  Establish the
	// correct starting hash so the incremental-vs-full comparison is valid.
	orig.Hash = DefaultZT.HashPosition(orig)

	after, doneMove := orig.DoMove(m)

	// Incremental hash after DoMove must match the full recomputation.
	if want := DefaultZT.HashPosition(after); after.Hash != want {
		t.Errorf("[%s] DoMove: incremental hash %016x != full hash %016x",
			label, after.Hash, want)
	}

	restored := after.UndoMove(doneMove)
	assertPositionsEqual(t, orig, restored, label)

	// Hash must be restored to the original value.
	if restored.Hash != orig.Hash {
		t.Errorf("[%s] UndoMove: hash %016x != original %016x",
			label, restored.Hash, orig.Hash)
	}
}

// assertPositionsEqual compares two positions field by field for a useful diff.
func assertPositionsEqual(t *testing.T, want, got Position, label string) {
	t.Helper()
	if want == got {
		return
	}
	if want.colOcc[WHITE] != got.colOcc[WHITE] {
		t.Errorf("[%s] colOcc[WHITE] want %016x got %016x", label, want.colOcc[WHITE], got.colOcc[WHITE])
	}
	if want.colOcc[BLACK] != got.colOcc[BLACK] {
		t.Errorf("[%s] colOcc[BLACK] want %016x got %016x", label, want.colOcc[BLACK], got.colOcc[BLACK])
	}
	if want.pawnOcc != got.pawnOcc {
		t.Errorf("[%s] pawnOcc want %016x got %016x", label, want.pawnOcc, got.pawnOcc)
	}
	if want.rookOcc != got.rookOcc {
		t.Errorf("[%s] rookOcc want %016x got %016x", label, want.rookOcc, got.rookOcc)
	}
	if want.bishopOcc != got.bishopOcc {
		t.Errorf("[%s] bishopOcc want %016x got %016x", label, want.bishopOcc, got.bishopOcc)
	}
	if want.knightOcc != got.knightOcc {
		t.Errorf("[%s] knightOcc want %016x got %016x", label, want.knightOcc, got.knightOcc)
	}
	if want.status != got.status {
		t.Errorf("[%s] status want %+v got %+v", label, want.status, got.status)
	}
}

// move constructs a bare Move (no undo fields – DoMove fills those).
func mkMove(from, to string, promo Piece) Move {
	return Move{From: SqParse(from), To: SqParse(to), Promotion: promo}
}

// ── A. Normal moves ───────────────────────────────────────────────────────────

func TestRoundTrip_SinglePush(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d4")
	assertRoundTrip(t, orig, mkMove("d4", "d5", EMPTY), "white single push d4-d5")
}

func TestRoundTrip_DoublePush(t *testing.T) {
	// Double push creates an EP phantom at Sq(2, file) = rank-2 d-file = d3.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d2")
	assertRoundTrip(t, orig, mkMove("d2", "d4", EMPTY), "white double push d2-d4")
}

func TestRoundTrip_BlackDoublePush(t *testing.T) {
	// Black double push creates EP phantom at Sq(5, file) = rank-5 e-file = e6.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "a8").
		AddPawn(BLACK, "e7")
	orig.status.SetTurn(BLACK)
	assertRoundTrip(t, orig, mkMove("e7", "e5", EMPTY), "black double push e7-e5")
}

func TestRoundTrip_PawnCapture(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "e4").AddPawn(BLACK, "d5")
	assertRoundTrip(t, orig, mkMove("e4", "d5", EMPTY), "white pawn captures d5")
}

func TestRoundTrip_KnightMove(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddKnight(WHITE, "g1")
	assertRoundTrip(t, orig, mkMove("g1", "f3", EMPTY), "knight g1-f3")
}

// ── B. En passant ─────────────────────────────────────────────────────────────

func TestRoundTrip_EnPassantWhite(t *testing.T) {
	// Black just double-pushed d7-d5; phantom at d6 (Sq(5,3)).
	// White pawn on e5 captures en passant to d6.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "a8").
		AddPawn(WHITE, "e5").AddPawn(BLACK, "d5").
		SetEnPassant(BLACK, "d5")
	assertRoundTrip(t, orig, mkMove("e5", "d6", EMPTY), "white en passant e5xd6")
}

func TestRoundTrip_EnPassantBlack(t *testing.T) {
	// White just double-pushed e2-e4; phantom at e3 (Sq(2,4)).
	// Black pawn on d4 captures en passant to e3.
	orig := *new(Position).
		AddKing(WHITE, "h1").AddKing(BLACK, "e8").
		AddPawn(BLACK, "d4").AddPawn(WHITE, "e4").
		SetEnPassant(WHITE, "e4")
	orig.status.SetTurn(BLACK)
	assertRoundTrip(t, orig, mkMove("d4", "e3", EMPTY), "black en passant d4xe3")
}

func TestRoundTrip_EPPhantomPreservedThroughNonEPMove(t *testing.T) {
	// There is an active EP phantom (black just pushed d7-d5).
	// White makes an unrelated king move (e1-e2), NOT capturing en passant.
	// After DoMove+UndoMove the EP phantom must be fully restored.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "a8").
		AddPawn(WHITE, "e5").AddPawn(BLACK, "d5").
		SetEnPassant(BLACK, "d5")
	assertRoundTrip(t, orig, mkMove("e1", "e2", EMPTY), "EP phantom survives non-EP move")
}

// ── C. Castling ───────────────────────────────────────────────────────────────

func TestRoundTrip_CastleWhiteKingSide(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "h1").
		SetCastle(WHITE, CanCastleKingSide)
	assertRoundTrip(t, orig, Move{From: 4, To: 6, Promotion: CASTLEMOVE, Score: 1}, "white kingside castle")
}

func TestRoundTrip_CastleWhiteQueenSide(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "a1").
		SetCastle(WHITE, CanCastleQueenSide)
	assertRoundTrip(t, orig, Move{From: 4, To: 2, Promotion: CASTLEMOVE, Score: 1}, "white queenside castle")
}

func TestRoundTrip_CastleBlackKingSide(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(BLACK, "h8").
		SetCastle(BLACK, CanCastleKingSide)
	orig.status.SetTurn(BLACK)
	assertRoundTrip(t, orig, Move{From: 60, To: 62, Promotion: CASTLEMOVE, Score: 1}, "black kingside castle")
}

func TestRoundTrip_CastleBlackQueenSide(t *testing.T) {
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(BLACK, "a8").
		SetCastle(BLACK, CanCastleQueenSide)
	orig.status.SetTurn(BLACK)
	assertRoundTrip(t, orig, Move{From: 60, To: 58, Promotion: CASTLEMOVE, Score: 1}, "black queenside castle")
}

// ── D. Promotion ──────────────────────────────────────────────────────────────

func TestRoundTrip_PromotionPush(t *testing.T) {
	// White pawn on e7 pushes to e8 and promotes to queen.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "a8").
		AddPawn(WHITE, "e7")
	assertRoundTrip(t, orig, mkMove("e7", "e8", QUEEN), "promotion push e7-e8=Q")
}

func TestRoundTrip_PromotionCapture(t *testing.T) {
	// White pawn on e7 captures d8 (black bishop) and promotes to knight.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "a8").
		AddPawn(WHITE, "e7").AddBishop(BLACK, "d8")
	assertRoundTrip(t, orig, mkMove("e7", "d8", KNIGHT), "promotion capture e7xd8=N")
}

func TestRoundTrip_BlackPromotionPush(t *testing.T) {
	// Black pawn on d2 pushes to d1 and promotes to rook.
	orig := *new(Position).
		AddKing(WHITE, "h1").AddKing(BLACK, "e8").
		AddPawn(BLACK, "d2")
	orig.status.SetTurn(BLACK)
	assertRoundTrip(t, orig, mkMove("d2", "d1", ROOK), "black promotion push d2-d1=R")
}

// ── E. Castle-right revocation ────────────────────────────────────────────────

func TestRoundTrip_WhiteRookMoveLosesRight(t *testing.T) {
	// White rook moves off h1 → WHITE loses kingside castle right.
	// After undo the right must be back.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "h1").
		SetCastle(WHITE, CanCastle)
	assertRoundTrip(t, orig, mkMove("h1", "h4", EMPTY), "white rook h1-h4 loses castling right")
}

func TestRoundTrip_BlackRookMoveLosesRight(t *testing.T) {
	// Black rook moves off a8 → BLACK loses queenside castle right.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(BLACK, "a8").
		SetCastle(BLACK, CanCastle)
	orig.status.SetTurn(BLACK)
	assertRoundTrip(t, orig, mkMove("a8", "a5", EMPTY), "black rook a8-a5 loses queenside right")
}

func TestRoundTrip_CaptureOpponentRookRevokesRight(t *testing.T) {
	// White rook captures black's h8 rook.
	// BLACK's kingside castle right must be revoked by DoMove and restored by UndoMove.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "h3").AddRook(BLACK, "h8").
		SetCastle(BLACK, CanCastleKingSide)
	assertRoundTrip(t, orig, mkMove("h3", "h8", EMPTY), "white rook captures h8 rook")
}

// ── F. Sequential moves ───────────────────────────────────────────────────────

func TestRoundTrip_TwoMovesSequential(t *testing.T) {
	// Start with StartPosition; make two half-moves, undo both, verify full recovery.
	orig := StartPosition

	// Move 1: e2-e4 (white double push)
	after1, dm1 := orig.DoMove(mkMove("e2", "e4", EMPTY))
	// Move 2: e7-e5 (black double push)
	after2, dm2 := after1.DoMove(mkMove("e7", "e5", EMPTY))

	back1 := after2.UndoMove(dm2)
	back2 := back1.UndoMove(dm1)

	assertPositionsEqual(t, after1, back1, "after undoing move 2")
	assertPositionsEqual(t, orig, back2, "after undoing both moves")
}

func TestRoundTrip_EPCreatedThenUsed(t *testing.T) {
	// 1. White plays d2-d4 (creates EP phantom).
	// 2. Black plays c4xd3 (en passant).
	// 3. Undo both: should recover the original position exactly.
	orig := *new(Position).
		AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d2").AddPawn(BLACK, "c4")

	after1, dm1 := orig.DoMove(mkMove("d2", "d4", EMPTY))
	// Black captures en passant: c4→d3
	after2, dm2 := after1.DoMove(mkMove("c4", "d3", EMPTY))

	back1 := after2.UndoMove(dm2)
	back2 := back1.UndoMove(dm1)

	assertPositionsEqual(t, after1, back1, "after undoing EP capture")
	assertPositionsEqual(t, orig, back2, "after undoing double push + EP capture")
}

// ── G. Explicit hash verification ────────────────────────────────────────────

// assertHashConsistent verifies that p.Hash == DefaultZT.HashPosition(p).
func assertHashConsistent(t *testing.T, p Position, label string) {
	t.Helper()
	want := DefaultZT.HashPosition(p)
	if p.Hash != want {
		t.Errorf("[%s] hash mismatch: got %016x, want %016x", label, p.Hash, want)
	}
}

func TestHash_StartPosition(t *testing.T) {
	assertHashConsistent(t, StartPosition, "StartPosition")
}

func TestHash_IncrementalVsFullAfterManyMoves(t *testing.T) {
	// Play a short forced sequence from the start position, verifying the
	// incremental hash after every half-move.
	sequence := []Move{
		mkMove("e2", "e4", EMPTY), // 1. e4
		mkMove("e7", "e5", EMPTY), // 1. ...e5
		mkMove("g1", "f3", EMPTY), // 2. Nf3
		mkMove("b8", "c6", EMPTY), // 2. ...Nc6
		mkMove("f1", "c4", EMPTY), // 3. Bc4
		mkMove("g8", "f6", EMPTY), // 3. ...Nf6
	}

	p := StartPosition
	for i, m := range sequence {
		p, _ = p.DoMove(m)
		assertHashConsistent(t, p, fmt.Sprintf("move %d", i+1))
	}
}

// ── sanity print (not a test, just a visual aid) ──────────────────────────────

func TestDoMoveVisual(t *testing.T) {
	p := StartPosition
	fmt.Println("=== Start ===")
	fmt.Println(p.String())

	m := mkMove("e2", "e4", EMPTY)
	p2, dm := p.DoMove(m)
	fmt.Printf("=== After %s ===\n", dm.String())
	fmt.Println(p2.String())

	p3 := p2.UndoMove(dm)
	fmt.Println("=== After UndoMove ===")
	fmt.Println(p3.String())

	if p != p3 {
		t.Error("UndoMove did not restore original position")
	}
}
