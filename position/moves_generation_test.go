package position

// Systematic move generation tests.
//
// Conventions:
//   - All positions are set up with WHITE to move unless noted otherwise.
//   - assertMovesFromSq checks the exact set of destinations for one piece.
//   - assertMoveListCount checks the total number of legal pseudo-legal moves.
//
// Categories covered:
//   1. Pawn – single push, double push, blocking, capture, border files
//   2. Pawn – en passant (white captures, black captures, file boundaries)
//   3. Knight – center, corner, edge
//   4. Bishop – open board, blocked, capture
//   5. Rook   – open board, blocked, capture
//   6. Queen  – open board, combined rays
//   7. King   – center, corner, edge
//   8. Castling – both sides, both colors, blocked, through-check
//   9. Starting position count

import (
	"fmt"
	"runtime"
	"testing"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// assertMovesFromSq verifies that GetMovesBB for `from` returns exactly the
// squares listed in `expected` (and no others).
func assertMovesFromSq(t *testing.T, p Position, from string, expected ...string) {
	t.Helper()
	fromSq := SqParse(from)
	got := p.GetMovesBB(fromSq)

	want := Bitboard(0)
	for _, s := range expected {
		want = want.Set(SqParse(s))
	}

	extra := got & ^want
	missing := want & ^got

	if extra != 0 || missing != 0 {
		t.Errorf("moves from %s:\n  extra   : %v\n  missing : %v",
			from, squareList(extra), squareList(missing))
	}
}

func squareList(b Bitboard) []string {
	var out []string
	for sq := range b.AllSetSquares {
		out = append(out, sq.String())
	}
	return out
}

// ── 1. PAWN MOVES ─────────────────────────────────────────────────────────────

func TestPawnSinglePush(t *testing.T) {
	// White pawn on rank 3 (no double-push), nothing ahead
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d4")
	assertMovesFromSq(t, *p, "d4", "d5")
}

func TestPawnDoublePush(t *testing.T) {
	// White pawn on rank 1 – both d3 and d4 available
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d2")
	assertMovesFromSq(t, *p, "d2", "d3", "d4")
}

func TestPawnDoublePushBlockedIntermediate(t *testing.T) {
	// d3 occupied – double push must be disallowed
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d2").AddPawn(BLACK, "d3")
	assertMovesFromSq(t, *p, "d2" /* nothing */)
}

func TestPawnDoublePushBlockedTarget(t *testing.T) {
	// d4 occupied but d3 clear – only single push
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d2").AddPawn(BLACK, "d4")
	assertMovesFromSq(t, *p, "d2", "d3")
}

func TestPawnCaptures(t *testing.T) {
	// White pawn on d4, black pawns on c5 and e5
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d4").AddPawn(BLACK, "c5", "e5")
	assertMovesFromSq(t, *p, "d4", "d5", "c5", "e5")
}

func TestPawnCapturesBorderFileA(t *testing.T) {
	// Pawn on a4: can only capture towards b5 (no left capture)
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "a4").AddPawn(BLACK, "b5")
	assertMovesFromSq(t, *p, "a4", "a5", "b5")
}

func TestPawnCapturesBorderFileH(t *testing.T) {
	// Pawn on h4: can only capture towards g5
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "h4").AddPawn(BLACK, "g5")
	assertMovesFromSq(t, *p, "h4", "h5", "g5")
}

func TestPawnNoCaptureOwnPiece(t *testing.T) {
	// Own piece on capture diagonal must not be a move
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d4", "e5")
	assertMovesFromSq(t, *p, "d4", "d5") // e5 is own piece
}

func TestBlackPawnMoves(t *testing.T) {
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(BLACK, "d7")
	p.status.SetTurn(BLACK)
	assertMovesFromSq(t, *p, "d7", "d6", "d5")
}

func TestBlackPawnCaptures(t *testing.T) {
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(BLACK, "d5").AddPawn(WHITE, "c4", "e4")
	p.status.SetTurn(BLACK)
	assertMovesFromSq(t, *p, "d5", "d4", "c4", "e4")
}

// ── 2. EN PASSANT ────────────────────────────────────────────────────────────

func TestEnPassantWhiteCapturesLeft(t *testing.T) {
	// White pawn d5, black just played c7-c5 → phantom at c8(rank7,file2)
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d5").AddPawn(BLACK, "c5").
		SetEnPassant(BLACK, "c5")
	assertMovesFromSq(t, *p, "d5", "d6", "c6") // c6 = en passant landing
}

func TestEnPassantWhiteCapturesRight(t *testing.T) {
	// White pawn d5, black just played e7-e5.
	// Black king must NOT be on e8: SetEnPassant(BLACK,"e5") places a phantom there.
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "a8").
		AddPawn(WHITE, "d5").AddPawn(BLACK, "e5").
		SetEnPassant(BLACK, "e5")
	assertMovesFromSq(t, *p, "d5", "d6", "e6")
}

func TestEnPassantWhiteFileA(t *testing.T) {
	// White pawn a5, black just played b7-b5 → can capture en passant right
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "a5").AddPawn(BLACK, "b5").
		SetEnPassant(BLACK, "b5")
	assertMovesFromSq(t, *p, "a5", "a6", "b6")
}

func TestEnPassantWhiteFileH(t *testing.T) {
	// White pawn h5, black just played g7-g5
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "h5").AddPawn(BLACK, "g5").
		SetEnPassant(BLACK, "g5")
	assertMovesFromSq(t, *p, "h5", "h6", "g6")
}

func TestEnPassantBlackCaptures(t *testing.T) {
	// Black pawn e4, white just played d2-d4
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(BLACK, "e4").AddPawn(WHITE, "d4").
		SetEnPassant(WHITE, "d4")
	p.status.SetTurn(BLACK)
	assertMovesFromSq(t, *p, "e4", "e3", "d3") // d3 = en passant landing
}

func TestEnPassantNotAvailableWhenNoPhantom(t *testing.T) {
	// No en passant signal set
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d5").AddPawn(BLACK, "e5")
	assertMovesFromSq(t, *p, "d5", "d6") // no en passant
}

func TestEnPassantWrongRank(t *testing.T) {
	// Pawn not at rank 4 cannot en passant even if phantom exists
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddPawn(WHITE, "d3").AddPawn(BLACK, "e5").
		SetEnPassant(BLACK, "e5")
	assertMovesFromSq(t, *p, "d3", "d4") // d3 is rank 2, not rank 4
}

// ── 3. KNIGHT ────────────────────────────────────────────────────────────────

func TestKnightCenter(t *testing.T) {
	p := new(Position).AddKing(WHITE, "a1").AddKing(BLACK, "h8").
		AddKnight(WHITE, "d4")
	assertMovesFromSq(t, *p, "d4", "b3", "b5", "c2", "c6", "e2", "e6", "f3", "f5")
}

func TestKnightCornerA1(t *testing.T) {
	p := new(Position).AddKing(WHITE, "h1").AddKing(BLACK, "h8").
		AddKnight(WHITE, "a1")
	assertMovesFromSq(t, *p, "a1", "b3", "c2")
}

func TestKnightCornerH8(t *testing.T) {
	p := new(Position).AddKing(WHITE, "a1").AddKing(BLACK, "a8").
		AddKnight(WHITE, "h8")
	assertMovesFromSq(t, *p, "h8", "g6", "f7")
}

func TestKnightEdgeFileA(t *testing.T) {
	p := new(Position).AddKing(WHITE, "h1").AddKing(BLACK, "h8").
		AddKnight(WHITE, "a4")
	assertMovesFromSq(t, *p, "a4", "b2", "b6", "c3", "c5")
}

func TestKnightBlockedByOwnPiece(t *testing.T) {
	// Own pieces on b3 and c2 – knight cannot land there
	p := new(Position).AddKing(WHITE, "h1").AddKing(BLACK, "h8").
		AddKnight(WHITE, "a1").AddPawn(WHITE, "b3", "c2")
	assertMovesFromSq(t, *p, "a1" /* no destinations */)
}

// ── 4. BISHOP ────────────────────────────────────────────────────────────────

func TestBishopOpenBoard(t *testing.T) {
	// Kings on h1/a8 so neither lies on d4's diagonals (NE: a1-h8, NW: a7-g1).
	p := new(Position).AddKing(WHITE, "h1").AddKing(BLACK, "a8").
		AddBishop(WHITE, "d4")
	// NE: e5,f6,g7,h8  SW: c3,b2,a1  NW: c5,b6,a7  SE: e3,f2,g1
	assertMovesFromSq(t, *p, "d4",
		"e5", "f6", "g7", "h8",
		"c3", "b2", "a1",
		"c5", "b6", "a7",
		"e3", "f2", "g1")
}

func TestBishopBlockedByOwnPiece(t *testing.T) {
	// Own pawn on e5 blocks the NE ray
	p := new(Position).AddKing(WHITE, "a1").AddKing(BLACK, "h8").
		AddBishop(WHITE, "d4").AddPawn(WHITE, "e5")
	assertMovesFromSq(t, *p, "d4",
		"c3", "b2",
		"c5", "b6", "a7",
		"e3", "f2", "g1")
}

func TestBishopCaptureOpponent(t *testing.T) {
	// Black pawn on e5 – bishop can capture it but not go further
	p := new(Position).AddKing(WHITE, "a1").AddKing(BLACK, "h8").
		AddBishop(WHITE, "d4").AddPawn(BLACK, "e5")
	assertMovesFromSq(t, *p, "d4",
		"e5", // capture
		"c3", "b2",
		"c5", "b6", "a7",
		"e3", "f2", "g1")
}

// ── 5. ROOK ───────────────────────────────────────────────────────────────────

func TestRookOpenBoard(t *testing.T) {
	p := new(Position).AddKing(WHITE, "a1").AddKing(BLACK, "h8").
		AddRook(WHITE, "d4")
	// Rank: a4,b4,c4,e4,f4,g4,h4
	// File: d1,d2,d3,d5,d6,d7,d8
	assertMovesFromSq(t, *p, "d4",
		"a4", "b4", "c4", "e4", "f4", "g4", "h4",
		"d1", "d2", "d3", "d5", "d6", "d7", "d8")
}

func TestRookBlockedByOwnPiece(t *testing.T) {
	// Own pawn on d6 blocks the north ray
	p := new(Position).AddKing(WHITE, "a1").AddKing(BLACK, "h8").
		AddRook(WHITE, "d4").AddPawn(WHITE, "d6")
	assertMovesFromSq(t, *p, "d4",
		"a4", "b4", "c4", "e4", "f4", "g4", "h4",
		"d1", "d2", "d3", "d5") // stops before d6
}

func TestRookCaptureOpponent(t *testing.T) {
	p := new(Position).AddKing(WHITE, "a1").AddKing(BLACK, "h8").
		AddRook(WHITE, "d4").AddPawn(BLACK, "d6")
	assertMovesFromSq(t, *p, "d4",
		"a4", "b4", "c4", "e4", "f4", "g4", "h4",
		"d1", "d2", "d3", "d5", "d6") // can capture d6
}

func TestRookBorderRank1(t *testing.T) {
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "a1")
	// Rank: b1,c1,d1 (e1=king)   File: a2..a8
	assertMovesFromSq(t, *p, "a1",
		"b1", "c1", "d1",
		"a2", "a3", "a4", "a5", "a6", "a7", "a8")
}

// ── 6. QUEEN ─────────────────────────────────────────────────────────────────

func TestQueenOpenBoard(t *testing.T) {
	// Kings on h1/a8 (not on any of d4's rays).
	p := new(Position).AddKing(WHITE, "h1").AddKing(BLACK, "a8").
		AddQueen(WHITE, "d4")
	assertMovesFromSq(t, *p, "d4",
		// rook rays
		"a4", "b4", "c4", "e4", "f4", "g4", "h4",
		"d1", "d2", "d3", "d5", "d6", "d7", "d8",
		// bishop rays
		"e5", "f6", "g7", "h8",
		"c3", "b2", "a1",
		"c5", "b6", "a7",
		"e3", "f2", "g1")
}

// ── 7. KING ───────────────────────────────────────────────────────────────────

func TestKingCenter(t *testing.T) {
	p := new(Position).AddKing(WHITE, "d4").AddKing(BLACK, "h8")
	assertMovesFromSq(t, *p, "d4",
		"c3", "d3", "e3", "c4", "e4", "c5", "d5", "e5")
}

func TestKingCornerA1(t *testing.T) {
	p := new(Position).AddKing(WHITE, "a1").AddKing(BLACK, "h8")
	assertMovesFromSq(t, *p, "a1", "a2", "b1", "b2")
}

func TestKingCornerH8(t *testing.T) {
	p := new(Position).AddKing(WHITE, "a1").AddKing(BLACK, "h8")
	p.status.SetTurn(BLACK)
	assertMovesFromSq(t, *p, "h8", "g7", "g8", "h7")
}

func TestKingEdgeFileA(t *testing.T) {
	p := new(Position).AddKing(WHITE, "a4").AddKing(BLACK, "h8")
	assertMovesFromSq(t, *p, "a4", "a3", "a5", "b3", "b4", "b5")
}

func TestKingBlockedByOwnPiece(t *testing.T) {
	p := new(Position).AddKing(WHITE, "d4").AddKing(BLACK, "h8").
		AddPawn(WHITE, "c3", "d3", "e3")
	assertMovesFromSq(t, *p, "d4",
		"c4", "e4", "c5", "d5", "e5") // c3,d3,e3 blocked
}

// ── 8. CASTLING ───────────────────────────────────────────────────────────────

func TestCastleWhiteKingside(t *testing.T) {
	// Clear f1,g1; castling rights set
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "h1").SetCastle(WHITE, CanCastleKingSide)
	moves := p.GetMoveList()
	found := false
	for _, m := range moves {
		if m.From == SqParse("e1") && m.To == SqParse("g1") && m.Promotion == CASTLEMOVE {
			found = true
		}
	}
	if !found {
		t.Error("expected white kingside castling move e1-g1")
	}
}

func TestCastleWhiteQueenside(t *testing.T) {
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "a1").SetCastle(WHITE, CanCastleQueenSide)
	moves := p.GetMoveList()
	found := false
	for _, m := range moves {
		if m.From == SqParse("e1") && m.To == SqParse("c1") && m.Promotion == CASTLEMOVE {
			found = true
		}
	}
	if !found {
		t.Error("expected white queenside castling move e1-c1")
	}
}

func TestCastleBlackKingside(t *testing.T) {
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(BLACK, "h8").SetCastle(BLACK, CanCastleKingSide)
	p.status.SetTurn(BLACK)
	moves := p.GetMoveList()
	found := false
	for _, m := range moves {
		if m.From == SqParse("e8") && m.To == SqParse("g8") && m.Promotion == CASTLEMOVE {
			found = true
		}
	}
	if !found {
		t.Error("expected black kingside castling move e8-g8")
	}
}

func TestCastleBlackQueenside(t *testing.T) {
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(BLACK, "a8").SetCastle(BLACK, CanCastleQueenSide)
	p.status.SetTurn(BLACK)
	moves := p.GetMoveList()
	found := false
	for _, m := range moves {
		if m.From == SqParse("e8") && m.To == SqParse("c8") && m.Promotion == CASTLEMOVE {
			found = true
		}
	}
	if !found {
		t.Error("expected black queenside castling move e8-c8")
	}
}

func TestCastleBlockedByOwnPiece(t *testing.T) {
	// f1 occupied – kingside castling must be blocked
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "h1").AddBishop(WHITE, "f1").
		SetCastle(WHITE, CanCastleKingSide)
	moves := p.GetMoveList()
	for _, m := range moves {
		if m.From == SqParse("e1") && m.To == SqParse("g1") && m.Promotion == CASTLEMOVE {
			t.Error("castling must not be possible with f1 occupied")
		}
	}
}

func TestCastleBlockedByOpponent(t *testing.T) {
	// g1 occupied by black piece
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "h1").AddPawn(BLACK, "g1").
		SetCastle(WHITE, CanCastleKingSide)
	moves := p.GetMoveList()
	for _, m := range moves {
		if m.From == SqParse("e1") && m.To == SqParse("g1") && m.Promotion == CASTLEMOVE {
			t.Error("castling must not be possible with g1 occupied")
		}
	}
}

func TestCastleKingInCheck(t *testing.T) {
	// King in check – castling must be disallowed
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "h1").AddRook(BLACK, "e6"). // attacks e1
		SetCastle(WHITE, CanCastleKingSide)
	moves := p.GetMoveList()
	for _, m := range moves {
		if m.Promotion == CASTLEMOVE {
			t.Error("castling must not be possible when king is in check")
		}
	}
}

func TestCastleThroughAttackedSquare(t *testing.T) {
	// f1 attacked by black rook – kingside castling disallowed
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "h1").AddRook(BLACK, "f6"). // attacks f1
		SetCastle(WHITE, CanCastleKingSide)
	moves := p.GetMoveList()
	for _, m := range moves {
		if m.From == SqParse("e1") && m.To == SqParse("g1") && m.Promotion == CASTLEMOVE {
			t.Error("castling must not be possible when passing through attacked square f1")
		}
	}
}

func TestNoCastleRightsSet(t *testing.T) {
	// Castle bits not set – no castling even with clear path
	p := new(Position).AddKing(WHITE, "e1").AddKing(BLACK, "e8").
		AddRook(WHITE, "h1") // no SetCastle
	moves := p.GetMoveList()
	for _, m := range moves {
		if m.Promotion == CASTLEMOVE {
			t.Error("castling must not be possible without castle rights")
		}
	}
}

// ── 9. STARTING POSITION ─────────────────────────────────────────────────────

func TestStartingPositionMoveCount(t *testing.T) {
	moves := StartPosition.GetMoveList()
	if len(moves) != 20 {
		t.Errorf("expected 20 moves from start, got %d", len(moves))
	}
}

// ── 10. BIGTABLE MEMORY ───────────────────────────────────────────────────────

func TestBigTableMemory(t *testing.T) {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	newBt := newBigTable()
	runtime.ReadMemStats(&after)
	_ = newBt // keep alive until after measurement
	allocated := after.TotalAlloc - before.TotalAlloc
	fmt.Printf("BigTable allocated: %d KB (~%d MB)\n",
		allocated/1024, allocated/(1024*1024))
}
