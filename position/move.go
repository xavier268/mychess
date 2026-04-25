package position

import (
	"fmt"
	"slices"
)

// CASTLEMOVE is the Promotion field sentinel for castling moves.
const CASTLEMOVE Piece = KING + ROOK

// Move represents one half-move (ply).
//
// The Promotion field encodes the move class:
//   - EMPTY      : normal move (push, capture, en passant)
//   - CASTLEMOVE : castling; king-side vs queen-side inferred from From/To squares
//   - KNIGHT / BISHOP / ROOK / QUEEN : pawn promotion to that piece type
//
// The undo fields are populated by DoMove and must be passed intact to UndoMove.
type Move struct {
	From, To  Square
	Promotion Piece // EMPTY / CASTLEMOVE / promotion piece type (always positive)
	Score     uint8 // Il ne s'agit pas du score de la position, mais d'un score grossier pour prioriser l'examen des killer moves.

	// ── undo fields (populated by DoMove) ──────────────────────────────────
	Captured      Piece  // captured piece with colour: +white, -black, EMPTY=none
	CaptureSquare Square // square of the captured piece (== To, except en passant)
	PrevStatus    Status // full status snapshot before this move
	PrevEPFile    int8   // file of the en passant phantom before this move; -1 = none
	PrevHash      uint64 // Zobrist hash before this move; restored by UndoMove
}

func (m Move) String() string {
	if m.Promotion == EMPTY {
		return fmt.Sprintf("%s - %s (%d)", m.From.String(), m.To.String(), m.Score)
	}
	return fmt.Sprintf("%s - %s (%d) => %s", m.From.String(), m.To.String(), m.Score, m.Promotion.String())
}

// ── piece-type helpers ────────────────────────────────────────────────────────

// clearPieceAt clears all piece-type bitboard bits at sq (does not touch colOcc).
func (pp *Position) clearPieceAt(sq Square) {
	bb := ^sq.Bitboard()
	pp.pawnOcc &= bb
	pp.rookOcc &= bb
	pp.bishopOcc &= bb
	pp.knightOcc &= bb
}

// setPieceAt sets the type-bitboard bit(s) for an absolute piece type at sq.
// KING has no type-bitboard entry (it lives in Status).
func (pp *Position) setPieceAt(sq Square, pieceType Piece) {
	switch pieceType {
	case PAWN:
		pp.pawnOcc |= sq.Bitboard()
	case KNIGHT:
		pp.knightOcc |= sq.Bitboard()
	case BISHOP:
		pp.bishopOcc |= sq.Bitboard()
	case ROOK:
		pp.rookOcc |= sq.Bitboard()
	case QUEEN:
		pp.rookOcc |= sq.Bitboard()
		pp.bishopOcc |= sq.Bitboard()
	}
}

// pabs returns the absolute value of a Piece (strips the colour sign).
func pabs(p Piece) Piece {
	if p < 0 {
		return -p
	}
	return p
}

// revokeRookCastle clears the castle right for the given colour if sq is that
// colour's initial rook corner.
func (pp *Position) revokeRookCastle(color uint8, sq Square) {
	switch color {
	case WHITE:
		if sq == 0 {
			pp.status.KingStatus[WHITE] &= ^uint8(CanCastleQueenSide)
		}
		if sq == 7 {
			pp.status.KingStatus[WHITE] &= ^uint8(CanCastleKingSide)
		}
	case BLACK:
		if sq == 56 {
			pp.status.KingStatus[BLACK] &= ^uint8(CanCastleQueenSide)
		}
		if sq == 63 {
			pp.status.KingStatus[BLACK] &= ^uint8(CanCastleKingSide)
		}
	}
}

// ── Zobrist hash helpers ──────────────────────────────────────────────────────

// ZobristBitboards index constants – must match HashPosition.
const (
	zbPawnOcc   = 2
	zbRookOcc   = 3
	zbBishopOcc = 4
	zbKnightOcc = 5
)

// hashXORPieceType XORs the type-bitboard Zobrist key for piece type pt at sq.
// KING is excluded: its ZobristKing contribution is handled by the status
// bracket in DoMove (XOR out old, XOR in new around all bitboard changes).
func (pp *Position) hashXORPieceType(sq Square, pt Piece) {
	switch pt {
	case PAWN:
		pp.Hash ^= DefaultZT.ZobristBitboards[zbPawnOcc][sq]
	case KNIGHT:
		pp.Hash ^= DefaultZT.ZobristBitboards[zbKnightOcc][sq]
	case BISHOP:
		pp.Hash ^= DefaultZT.ZobristBitboards[zbBishopOcc][sq]
	case ROOK:
		pp.Hash ^= DefaultZT.ZobristBitboards[zbRookOcc][sq]
	case QUEEN:
		pp.Hash ^= DefaultZT.ZobristBitboards[zbRookOcc][sq]
		pp.Hash ^= DefaultZT.ZobristBitboards[zbBishopOcc][sq]
	}
}

// ── move-list generation ──────────────────────────────────────────────────────

// checkBonus is added to a move's score when it gives check to the opponent.
// Value chosen to rank quiet checks above rook captures (4) but below queen captures (7).
const checkBonus uint8 = 5

// GetMoveList returns all legal moves for the side to move, sorted by descending score.
// Illegal moves that leave the own king in check are filtered out here.
// Moves that give check receive a score bonus of checkBonus.
// Promotion is expanded into four moves (Q/R/B/N).
// Castling moves are included.
// Undo fields are NOT populated here; DoMove fills them.
func (p Position) GetMoveList() []Move {
	turn := p.status.GetTurn()
	promotionRank := int((1 - turn) * 7) // WHITE → 7, BLACK → 0

	candidates := make([]Move, 0, 32)
	for fromSq := range p.colOcc[turn].AllSetSquares {
		bb := p.GetMovesBB(fromSq)
		for toSq := range bb.AllSetSquares {
			// Score captures by piece value
			sc := uint8(p.rookOcc.Get(toSq)*4 +
				p.bishopOcc.Get(toSq)*3 +
				p.knightOcc.Get(toSq)*3 +
				p.pawnOcc.Get(toSq)*2)

			// Pawn reaching the last rank → expand into four promotion moves
			if p.pawnOcc.IsSet(fromSq) && toSq.Rank() == promotionRank {
				for _, piece := range []Piece{QUEEN, ROOK, BISHOP, KNIGHT} {
					candidates = append(candidates, Move{
						From:      fromSq,
						To:        toSq,
						Promotion: piece,
						Score:     sc + 2*uint8(piece),
					})
				}
			} else {
				candidates = append(candidates, Move{
					From:      fromSq,
					To:        toSq,
					Promotion: EMPTY,
					Score:     sc,
				})
			}
		}
	}
	candidates = append(candidates, p.GetCastlingMoveList()...)

	moves := make([]Move, 0, len(candidates))
	for _, m := range candidates {
		pp, _ := p.DoMove(m)
		if pp.IsSquareAttacked(pp.KingPosition(turn), 1^turn) {
			continue // own king left in check: illegal
		}
		if pp.IsCheck() {
			m.Score += checkBonus // move gives check to opponent
		}
		moves = append(moves, m)
	}

	slices.SortFunc(moves, func(a, b Move) int { return int(b.Score) - int(a.Score) })
	return moves
}

// GetCastlingMoveList returns castling moves for the side to move, if legal.
func (p Position) GetCastlingMoveList() []Move {
	turn := p.status.GetTurn()
	cb := p.status.GetCastleBits(turn)
	if cb == 0 || p.IsSquareAttacked(p.status.GetKingPosition(turn), 1-turn) {
		return nil
	}

	moves := make([]Move, 0, 2)
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]

	if turn == WHITE {
		if cb&CanCastleKingSide != 0 &&
			occ&((1<<5)|(1<<6)) == 0 &&
			!p.IsSquareAttacked(5, BLACK) && !p.IsSquareAttacked(6, BLACK) {
			moves = append(moves, Move{From: 4, To: 6, Promotion: CASTLEMOVE, Score: 1})
		}
		if cb&CanCastleQueenSide != 0 &&
			occ&((1<<1)|(1<<2)|(1<<3)) == 0 &&
			!p.IsSquareAttacked(3, BLACK) && !p.IsSquareAttacked(2, BLACK) {
			moves = append(moves, Move{From: 4, To: 2, Promotion: CASTLEMOVE, Score: 1})
		}
	} else {
		if cb&CanCastleKingSide != 0 &&
			occ&((1<<61)|(1<<62)) == 0 &&
			!p.IsSquareAttacked(61, WHITE) && !p.IsSquareAttacked(62, WHITE) {
			moves = append(moves, Move{From: 60, To: 62, Promotion: CASTLEMOVE, Score: 1})
		}
		if cb&CanCastleQueenSide != 0 &&
			occ&((1<<57)|(1<<58)|(1<<59)) == 0 &&
			!p.IsSquareAttacked(59, WHITE) && !p.IsSquareAttacked(58, WHITE) {
			moves = append(moves, Move{From: 60, To: 58, Promotion: CASTLEMOVE, Score: 1})
		}
	}

	return moves
}

// ── DoMove ────────────────────────────────────────────────────────────────────

// DoMove applies m to p and returns the resulting position together with the
// move enriched with all undo information (Captured, CaptureSquare, PrevStatus,
// PrevEPFile, PrevHash).  Pass the returned Move to UndoMove to restore p exactly.
//
// Hash is maintained incrementally using DefaultZT:
//   - Status-dependent contributions (castle bits, king squares) are bracketed:
//     XOR out before any change, XOR in after all changes.
//   - colOcc and type-bitboard changes are XORed inline alongside each bitboard op.
//   - The turn flip is unconditional (ZobristTurn).
//   - The EP phantom is XORed out at the bracket and XORed in again if a new
//     phantom is created by a double pawn push.
func (p Position) DoMove(m Move) (Position, Move) {
	turn := p.status.GetTurn()
	opponent := 1 ^ turn
	pp := p

	// ── 1. Save undo information ──────────────────────────────────────────────
	m.PrevStatus = p.status
	m.PrevHash = p.Hash

	phantoms := p.pawnOcc & ^(p.colOcc[WHITE] | p.colOcc[BLACK])
	m.PrevEPFile = -1
	for sq := range phantoms.AllSetSquares {
		m.PrevEPFile = int8(sq.File())
		break // at most one phantom at a time
	}

	// ── 2. Hash bracket – XOR out status-dependent contributions ─────────────
	// Castle bits: GetCastleBits returns {0,0x40,0x80,0xC0}; >>6 maps to 0-3.
	pp.Hash ^= DefaultZT.ZobristCastling[WHITE][p.status.GetCastleBits(WHITE)>>6]
	pp.Hash ^= DefaultZT.ZobristCastling[BLACK][p.status.GetCastleBits(BLACK)>>6]
	// King squares (in Status, separate from colOcc):
	pp.Hash ^= DefaultZT.ZobristKing[WHITE][p.status.GetKingPosition(WHITE)]
	pp.Hash ^= DefaultZT.ZobristKing[BLACK][p.status.GetKingPosition(BLACK)]
	// EP phantom in pawnOcc (rank 7 for WHITE's turn target, rank 0 for BLACK's):
	if m.PrevEPFile >= 0 {
		phantomRank := int(1-turn) * 7 // WHITE to move → 7; BLACK to move → 0
		pp.Hash ^= DefaultZT.ZobristBitboards[zbPawnOcc][Sq(phantomRank, int(m.PrevEPFile))]
	}

	// ── 3. Clear the outgoing en passant phantom ──────────────────────────────
	pp.pawnOcc &= p.colOcc[WHITE] | p.colOcc[BLACK]

	// ── 4. Switch turn (unconditional XOR) ────────────────────────────────────
	pp.status.SwitchTurn()
	pp.Hash ^= DefaultZT.ZobristTurn

	switch m.Promotion {

	// ── Castling ──────────────────────────────────────────────────────────────
	case CASTLEMOVE:
		m.Captured = EMPTY
		m.CaptureSquare = m.To

		// Rook squares derived from king movement direction
		var rookFrom, rookTo Square
		if m.From < m.To { // king-side
			rookFrom = m.To + 1
			rookTo = m.To - 1
		} else { // queen-side
			rookFrom = m.To - 2
			rookTo = m.To + 1
		}

		// Move king (ZobristKing handled by the bracket; only colOcc changes here)
		pp.colOcc[turn] = (pp.colOcc[turn] & ^m.From.Bitboard()) | m.To.Bitboard()
		pp.Hash ^= DefaultZT.ZobristBitboards[turn][m.From]
		pp.Hash ^= DefaultZT.ZobristBitboards[turn][m.To]
		pp.status.SetKingPosition(turn, m.To)
		pp.status.SetCastleBits(turn, 0)

		// Move rook (colOcc + rookOcc)
		pp.colOcc[turn] = (pp.colOcc[turn] & ^rookFrom.Bitboard()) | rookTo.Bitboard()
		pp.Hash ^= DefaultZT.ZobristBitboards[turn][rookFrom]
		pp.Hash ^= DefaultZT.ZobristBitboards[turn][rookTo]
		pp.rookOcc = (pp.rookOcc & ^rookFrom.Bitboard()) | rookTo.Bitboard()
		pp.Hash ^= DefaultZT.ZobristBitboards[zbRookOcc][rookFrom]
		pp.Hash ^= DefaultZT.ZobristBitboards[zbRookOcc][rookTo]

	// ── Pawn promotion ────────────────────────────────────────────────────────
	case KNIGHT, BISHOP, ROOK, QUEEN:
		m.Captured = p.PieceAt(m.To) // EMPTY or opponent piece
		m.CaptureSquare = m.To

		// Remove any piece at destination
		if m.Captured != EMPTY {
			pp.colOcc[opponent] &= ^m.To.Bitboard()
			pp.Hash ^= DefaultZT.ZobristBitboards[opponent][m.To]
			pp.clearPieceAt(m.To)
			pp.hashXORPieceType(m.To, pabs(m.Captured))
		}

		// Remove pawn from source
		pp.pawnOcc &= ^m.From.Bitboard()
		pp.colOcc[turn] = (pp.colOcc[turn] & ^m.From.Bitboard()) | m.To.Bitboard()
		pp.Hash ^= DefaultZT.ZobristBitboards[turn][m.From]
		pp.Hash ^= DefaultZT.ZobristBitboards[zbPawnOcc][m.From]
		// Place promoted piece at destination
		pp.Hash ^= DefaultZT.ZobristBitboards[turn][m.To]
		pp.setPieceAt(m.To, m.Promotion)
		pp.hashXORPieceType(m.To, m.Promotion)

	// ── Normal move (push, capture, en passant) ────────────────────────────────
	default: // EMPTY
		movedType := pabs(p.PieceAt(m.From))

		// Detect en passant: pawn changes file to an empty square
		isEP := movedType == PAWN &&
			m.From.File() != m.To.File() &&
			p.PieceAt(m.To) == EMPTY

		if isEP {
			m.CaptureSquare = Sq(m.From.Rank(), m.To.File())
			if turn == WHITE {
				m.Captured = -PAWN
			} else {
				m.Captured = PAWN
			}
			// Remove captured pawn (at the rank of the moving pawn, adjacent file)
			pp.pawnOcc &= ^m.CaptureSquare.Bitboard()
			pp.colOcc[opponent] &= ^m.CaptureSquare.Bitboard()
			pp.Hash ^= DefaultZT.ZobristBitboards[opponent][m.CaptureSquare]
			pp.Hash ^= DefaultZT.ZobristBitboards[zbPawnOcc][m.CaptureSquare]
		} else {
			m.Captured = p.PieceAt(m.To)
			m.CaptureSquare = m.To
			if m.Captured != EMPTY {
				pp.colOcc[opponent] &= ^m.To.Bitboard()
				pp.Hash ^= DefaultZT.ZobristBitboards[opponent][m.To]
				pp.clearPieceAt(m.To)
				pp.hashXORPieceType(m.To, pabs(m.Captured))
			}
		}

		// Move piece (colOcc always; type bitboard for non-kings)
		pp.colOcc[turn] = (pp.colOcc[turn] & ^m.From.Bitboard()) | m.To.Bitboard()
		pp.Hash ^= DefaultZT.ZobristBitboards[turn][m.From]
		pp.Hash ^= DefaultZT.ZobristBitboards[turn][m.To]
		pp.clearPieceAt(m.From)
		pp.setPieceAt(m.To, movedType)
		if movedType != KING {
			// King's ZobristKing contribution is handled by the status bracket.
			pp.hashXORPieceType(m.From, movedType)
			pp.hashXORPieceType(m.To, movedType)
		}

		switch movedType {
		case KING:
			pp.status.SetKingPosition(turn, m.To)
			pp.status.SetCastleBits(turn, 0)
		case ROOK:
			pp.revokeRookCastle(turn, m.From)
		case PAWN:
			// Double push: set new en passant phantom.
			// Only place the phantom if the target square is unoccupied; if a
			// piece already sits there the phantom would be invisible (masked out
			// by colOcc in phantom detection) and would corrupt pawnOcc on undo.
			allOcc := pp.colOcc[WHITE] | pp.colOcc[BLACK]
			if turn == WHITE && m.From.Rank() == 1 && m.To.Rank() == 3 {
				phantomSq := Sq(0, m.From.File())
				if !allOcc.IsSet(phantomSq) {
					pp.pawnOcc |= phantomSq.Bitboard()
					pp.Hash ^= DefaultZT.ZobristBitboards[zbPawnOcc][phantomSq]
				}
			} else if turn == BLACK && m.From.Rank() == 6 && m.To.Rank() == 4 {
				phantomSq := Sq(7, m.From.File())
				if !allOcc.IsSet(phantomSq) {
					pp.pawnOcc |= phantomSq.Bitboard()
					pp.Hash ^= DefaultZT.ZobristBitboards[zbPawnOcc][phantomSq]
				}
			}
		}
	}

	// Revoke opponent's castling right if their rook was captured.
	// (The castle-bit hash change is handled by the closing bracket below.)
	if m.Captured != EMPTY && pabs(m.Captured) == ROOK {
		pp.revokeRookCastle(opponent, m.CaptureSquare)
	}

	// ── 5. Hash bracket – XOR in new status-dependent contributions ───────────
	pp.Hash ^= DefaultZT.ZobristCastling[WHITE][pp.status.GetCastleBits(WHITE)>>6]
	pp.Hash ^= DefaultZT.ZobristCastling[BLACK][pp.status.GetCastleBits(BLACK)>>6]
	pp.Hash ^= DefaultZT.ZobristKing[WHITE][pp.status.GetKingPosition(WHITE)]
	pp.Hash ^= DefaultZT.ZobristKing[BLACK][pp.status.GetKingPosition(BLACK)]

	return pp, m
}

// ── UndoMove ──────────────────────────────────────────────────────────────────

// UndoMove restores the position before DoMove was called.
// m must be the enriched Move returned by DoMove (undo fields must be intact).
// Hash is restored atomically from m.PrevHash — no re-computation needed.
func (p Position) UndoMove(m Move) Position {
	pp := p
	turn := m.PrevStatus.GetTurn()
	opponent := 1 ^ turn

	// Restore all status bits (turn, king positions, castle rights) and hash.
	pp.status = m.PrevStatus
	pp.Hash = m.PrevHash

	switch m.Promotion {

	// ── Undo castling ─────────────────────────────────────────────────────────
	case CASTLEMOVE:
		var rookFrom, rookTo Square
		if m.From < m.To { // king-side
			rookFrom = m.To + 1
			rookTo = m.To - 1
		} else { // queen-side
			rookFrom = m.To - 2
			rookTo = m.To + 1
		}
		// Restore king
		pp.colOcc[turn] = (pp.colOcc[turn] & ^m.To.Bitboard()) | m.From.Bitboard()
		// Restore rook
		pp.colOcc[turn] = (pp.colOcc[turn] & ^rookTo.Bitboard()) | rookFrom.Bitboard()
		pp.rookOcc = (pp.rookOcc & ^rookTo.Bitboard()) | rookFrom.Bitboard()

	// ── Undo promotion ────────────────────────────────────────────────────────
	case KNIGHT, BISHOP, ROOK, QUEEN:
		// Remove promoted piece at destination
		pp.colOcc[turn] &= ^m.To.Bitboard()
		pp.clearPieceAt(m.To)
		// Restore pawn at source
		pp.colOcc[turn] |= m.From.Bitboard()
		pp.pawnOcc |= m.From.Bitboard()
		// Restore captured piece (promotion-capture)
		if m.Captured != EMPTY {
			pp.colOcc[opponent] |= m.CaptureSquare.Bitboard()
			pp.setPieceAt(m.CaptureSquare, pabs(m.Captured))
		}

	// ── Undo normal move (including en passant) ────────────────────────────────
	default: // EMPTY
		// Determine moved piece from the post-move position (p, unchanged)
		movedType := pabs(p.PieceAt(m.To))

		// Move piece back from destination to source
		pp.colOcc[turn] = (pp.colOcc[turn] & ^m.To.Bitboard()) | m.From.Bitboard()
		pp.clearPieceAt(m.To)
		pp.setPieceAt(m.From, movedType)

		// Restore captured piece (at CaptureSquare, which ≠ To for en passant)
		if m.Captured != EMPTY {
			pp.colOcc[opponent] |= m.CaptureSquare.Bitboard()
			pp.setPieceAt(m.CaptureSquare, pabs(m.Captured))
		}
	}

	// ── Restore en passant phantom ────────────────────────────────────────────
	// Clear any phantom created by the move being undone (e.g. a double push)
	pp.pawnOcc &= pp.colOcc[WHITE] | pp.colOcc[BLACK]
	// Restore the phantom that existed before the move
	if m.PrevEPFile >= 0 {
		// Phantom rank: 7 when it's WHITE's turn (black created it), 0 when BLACK's turn
		phantomRank := 7 * (1 - int(turn))
		pp.pawnOcc |= Sq(phantomRank, int(m.PrevEPFile)).Bitboard()
	}

	return pp
}
