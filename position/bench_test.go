package position

// Benchmarks for move generation.
//
// Run with:
//   go test ./position/... -bench=. -benchmem -benchtime=3s
//
// Positions used:
//   benchPos     – "middlegame": both sides active, no special flags
//   benchPosCastle – same but with castling rights available
//   benchPosEP    – same but with an en passant phantom set

import "testing"

// ── shared positions (built once at package init) ────────────────────────────

var (
	// Complex middlegame – lots of sliding pieces, both colors, no special flags.
	// White to move.
	benchPos = func() Position {
		p := new(Position).
			AddKing(WHITE, "g1").AddKing(BLACK, "g8").
			AddQueen(WHITE, "d1").AddQueen(BLACK, "d8").
			AddRook(WHITE, "a1", "f1").AddRook(BLACK, "a8", "f8").
			AddBishop(WHITE, "c1", "f4").AddBishop(BLACK, "c8", "e6").
			AddKnight(WHITE, "c3", "f3").AddKnight(BLACK, "c6", "f6").
			AddPawn(WHITE, "a2", "b2", "c2", "d4", "e5", "f2", "g2", "h2").
			AddPawn(BLACK, "a7", "b7", "c7", "d5", "e4", "f7", "g7", "h7")
		return *p
	}()

	// Same position but WHITE has both castling rights (and clear path on king-side).
	benchPosCastle = func() Position {
		p := new(Position).
			AddKing(WHITE, "e1").AddKing(BLACK, "g8").
			AddQueen(WHITE, "d1").AddQueen(BLACK, "d8").
			AddRook(WHITE, "a1", "h1").AddRook(BLACK, "a8", "f8").
			AddBishop(WHITE, "c1", "f4").AddBishop(BLACK, "c8", "e6").
			AddKnight(WHITE, "c3", "f3").AddKnight(BLACK, "c6", "f6").
			AddPawn(WHITE, "a2", "b2", "c2", "d4", "e5", "g2", "h2").
			AddPawn(BLACK, "a7", "b7", "c7", "d5", "e4", "f7", "g7", "h7").
			SetCastle(WHITE, CanCastleKingSide|CanCastleQueenSide)
		return *p
	}()

	// Same complex position but with an en passant phantom.
	// BLACK just played d7-d5; WHITE pawn on e5 can capture en passant to d6.
	// Black queen moved to e8 (instead of d8) so that d8 is free for the phantom.
	benchPosEP = func() Position {
		p := new(Position).
			AddKing(WHITE, "g1").AddKing(BLACK, "g8").
			AddQueen(WHITE, "d1").AddQueen(BLACK, "e8"). // e8 free, d8 used for phantom
			AddRook(WHITE, "a1", "f1").AddRook(BLACK, "a8", "f8").
			AddBishop(WHITE, "c1", "f4").AddBishop(BLACK, "c8", "e6").
			AddKnight(WHITE, "c3", "f3").AddKnight(BLACK, "c6", "f6").
			AddPawn(WHITE, "a2", "b2", "c2", "d4", "e5", "f2", "g2", "h2").
			AddPawn(BLACK, "a7", "b7", "c7", "d5", "e4", "f7", "g7", "h7").
			SetEnPassant(BLACK, "d5") // phantom at d8 (empty); e5 pawn captures to d6
		return *p
	}()
)

// ── 1. BigTable construction ──────────────────────────────────────────────────

func BenchmarkNewBigTable(b *testing.B) {
	b.ReportAllocs()
	var sink *BigTable
	for b.Loop() {
		sink = newBigTable()
	}
	_ = sink
}

// ── 2. Full move list – baseline (no special flags) ───────────────────────────

func BenchmarkGetMoveList_Baseline(b *testing.B) {
	b.ReportAllocs()
	var sink []Move
	for b.Loop() {
		sink = benchPos.GetMoveList()
	}
	_ = sink
}

// ── 3. Full move list – with castling rights ─────────────────────────────────

func BenchmarkGetMoveList_Castling(b *testing.B) {
	b.ReportAllocs()
	var sink []Move
	for b.Loop() {
		sink = benchPosCastle.GetMoveList()
	}
	_ = sink
}

// ── 4. Full move list – with en passant phantom ───────────────────────────────

func BenchmarkGetMoveList_EnPassant(b *testing.B) {
	b.ReportAllocs()
	var sink []Move
	for b.Loop() {
		sink = benchPosEP.GetMoveList()
	}
	_ = sink
}

// ── 5. Per-piece move generation (GetMovesBB) ────────────────────────────────

func BenchmarkGetMovesBB_Queen(b *testing.B) {
	b.ReportAllocs()
	var sink Bitboard
	for b.Loop() {
		sink = benchPos.GetMovesBB(SqParse("d1"))
	}
	_ = sink
}

func BenchmarkGetMovesBB_Rook(b *testing.B) {
	b.ReportAllocs()
	var sink Bitboard
	for b.Loop() {
		sink = benchPos.GetMovesBB(SqParse("f1"))
	}
	_ = sink
}

func BenchmarkGetMovesBB_Bishop(b *testing.B) {
	b.ReportAllocs()
	var sink Bitboard
	for b.Loop() {
		sink = benchPos.GetMovesBB(SqParse("f4"))
	}
	_ = sink
}

func BenchmarkGetMovesBB_Knight(b *testing.B) {
	b.ReportAllocs()
	var sink Bitboard
	for b.Loop() {
		sink = benchPos.GetMovesBB(SqParse("f3"))
	}
	_ = sink
}

func BenchmarkGetMovesBB_Pawn_NoEP(b *testing.B) {
	b.ReportAllocs()
	var sink Bitboard
	for b.Loop() {
		sink = benchPos.GetMovesBB(SqParse("e5"))
	}
	_ = sink
}

func BenchmarkGetMovesBB_Pawn_WithEP(b *testing.B) {
	b.ReportAllocs()
	var sink Bitboard
	for b.Loop() {
		sink = benchPosEP.GetMovesBB(SqParse("e5"))
	}
	_ = sink
}

// ── 6. IsSquareAttacked (used by castling path check) ────────────────────────

func BenchmarkIsSquareAttacked(b *testing.B) {
	b.ReportAllocs()
	var sink bool
	for b.Loop() {
		sink = benchPos.IsSquareAttacked(SqParse("e4"), WHITE)
	}
	_ = sink
}
