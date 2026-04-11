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
	Score     uint8

	// ── undo fields (populated by DoMove) ──────────────────────────────────
	Captured      Piece  // captured piece with colour: +white, -black, EMPTY=none
	CaptureSquare Square // square of the captured piece (== To, except en passant)
	PrevStatus    Status // full status snapshot before this move
	PrevEPFile    int8   // file of the en passant phantom before this move; -1 = none
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

// ── move-list generation ──────────────────────────────────────────────────────

// GetMoveList returns all pseudo-legal moves for the side to move.
// Illegal moves that leave the own king in check are NOT filtered.
// Promotion is expanded into four moves (Q/R/B/N).
// Undo fields are NOT populated here; DoMove fills them.
func (p Position) GetMoveList(bt *BigTable) []Move {
	moves := make([]Move, 0, 32)
	turn := p.status.GetTurn()
	promotionRank := int((1 - turn) * 7) // WHITE → 7, BLACK → 0

	for fromSq := range p.colOcc[turn].AllSetSquares {
		bb := p.GetMovesBB(bt, fromSq)
		for toSq := range bb.AllSetSquares {
			// Score captures by piece value
			sc := uint8(p.rookOcc.Get(toSq)*4 +
				p.bishopOcc.Get(toSq)*3 +
				p.knightOcc.Get(toSq)*3 +
				p.pawnOcc.Get(toSq)*2)

			// Pawn reaching the last rank → expand into four promotion moves
			if p.pawnOcc.IsSet(fromSq) && toSq.Rank() == promotionRank {
				for _, piece := range []Piece{QUEEN, ROOK, BISHOP, KNIGHT} {
					moves = append(moves, Move{
						From:      fromSq,
						To:        toSq,
						Promotion: piece,
						Score:     sc + 2*uint8(piece),
					})
				}
			} else {
				moves = append(moves, Move{
					From:      fromSq,
					To:        toSq,
					Promotion: EMPTY,
					Score:     sc,
				})
			}
		}
	}

	moves = append(moves, p.GetCastlingMoveList(bt)...)
	slices.SortFunc(moves, func(a, b Move) int { return int(b.Score) - int(a.Score) })
	return moves
}

// GetCastlingMoveList returns castling moves for the side to move, if legal.
func (p Position) GetCastlingMoveList(bt *BigTable) []Move {
	turn := p.status.GetTurn()
	cb := p.status.GetCastleBits(turn)
	if cb == 0 || p.IsSquareAttacked(bt, p.status.GetKingPosition(turn), 1-turn) {
		return nil
	}

	moves := make([]Move, 0, 2)
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]

	if turn == WHITE {
		if cb&CanCastleKingSide != 0 &&
			occ&((1<<5)|(1<<6)) == 0 &&
			!p.IsSquareAttacked(bt, 5, BLACK) && !p.IsSquareAttacked(bt, 6, BLACK) {
			moves = append(moves, Move{From: 4, To: 6, Promotion: CASTLEMOVE, Score: 1})
		}
		if cb&CanCastleQueenSide != 0 &&
			occ&((1<<1)|(1<<2)|(1<<3)) == 0 &&
			!p.IsSquareAttacked(bt, 3, BLACK) && !p.IsSquareAttacked(bt, 2, BLACK) {
			moves = append(moves, Move{From: 4, To: 2, Promotion: CASTLEMOVE, Score: 1})
		}
	} else {
		if cb&CanCastleKingSide != 0 &&
			occ&((1<<61)|(1<<62)) == 0 &&
			!p.IsSquareAttacked(bt, 61, WHITE) && !p.IsSquareAttacked(bt, 62, WHITE) {
			moves = append(moves, Move{From: 60, To: 62, Promotion: CASTLEMOVE, Score: 1})
		}
		if cb&CanCastleQueenSide != 0 &&
			occ&((1<<57)|(1<<58)|(1<<59)) == 0 &&
			!p.IsSquareAttacked(bt, 59, WHITE) && !p.IsSquareAttacked(bt, 58, WHITE) {
			moves = append(moves, Move{From: 60, To: 58, Promotion: CASTLEMOVE, Score: 1})
		}
	}

	return moves
}

// ── DoMove ────────────────────────────────────────────────────────────────────

// DoMove applies m to p and returns the resulting position together with the
// move enriched with all undo information (Captured, CaptureSquare, PrevStatus,
// PrevEPFile).  Pass the returned Move to UndoMove to restore p exactly.
func (p Position) DoMove(m Move) (Position, Move) {
	turn := p.status.GetTurn()
	opponent := 1 ^ turn
	pp := p

	// ── 1. Save undo information ──────────────────────────────────────────────
	m.PrevStatus = p.status

	phantoms := p.pawnOcc & ^(p.colOcc[WHITE] | p.colOcc[BLACK])
	m.PrevEPFile = -1
	for sq := range phantoms.AllSetSquares {
		m.PrevEPFile = int8(sq.File())
		break // at most one phantom at a time
	}

	// ── 2. Clear the outgoing en passant phantom ──────────────────────────────
	pp.pawnOcc &= p.colOcc[WHITE] | p.colOcc[BLACK]

	// ── 3. Switch turn ────────────────────────────────────────────────────────
	pp.status.SwitchTurn()

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

		// Move king
		pp.colOcc[turn] = (pp.colOcc[turn] & ^m.From.Bitboard()) | m.To.Bitboard()
		pp.status.SetKingPosition(turn, m.To)
		pp.status.SetCastleBits(turn, 0)

		// Move rook
		pp.colOcc[turn] = (pp.colOcc[turn] & ^rookFrom.Bitboard()) | rookTo.Bitboard()
		pp.rookOcc = (pp.rookOcc & ^rookFrom.Bitboard()) | rookTo.Bitboard()

	// ── Pawn promotion ────────────────────────────────────────────────────────
	case KNIGHT, BISHOP, ROOK, QUEEN:
		m.Captured = p.PieceAt(m.To) // EMPTY or opponent piece
		m.CaptureSquare = m.To

		// Remove any piece at destination
		if m.Captured != EMPTY {
			pp.colOcc[opponent] &= ^m.To.Bitboard()
			pp.clearPieceAt(m.To)
		}

		// Remove pawn from source; place promoted piece at destination
		pp.pawnOcc &= ^m.From.Bitboard()
		pp.colOcc[turn] = (pp.colOcc[turn] & ^m.From.Bitboard()) | m.To.Bitboard()
		pp.setPieceAt(m.To, m.Promotion)

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
			pp.pawnOcc &= ^m.CaptureSquare.Bitboard()
			pp.colOcc[opponent] &= ^m.CaptureSquare.Bitboard()
		} else {
			m.Captured = p.PieceAt(m.To)
			m.CaptureSquare = m.To
			if m.Captured != EMPTY {
				pp.colOcc[opponent] &= ^m.To.Bitboard()
				pp.clearPieceAt(m.To)
			}
		}

		// Move piece
		pp.colOcc[turn] = (pp.colOcc[turn] & ^m.From.Bitboard()) | m.To.Bitboard()
		pp.clearPieceAt(m.From)
		pp.setPieceAt(m.To, movedType)

		switch movedType {
		case KING:
			pp.status.SetKingPosition(turn, m.To)
			pp.status.SetCastleBits(turn, 0)
		case ROOK:
			pp.revokeRookCastle(turn, m.From)
		case PAWN:
			// Double push: set new en passant phantom.
			// Only place the phantom if the target square is unoccupied; if a
			// piece already sits there the phantom would be invisible (it would
			// be masked out by colOcc in phantom detection) and would also
			// permanently corrupt pawnOcc after subsequent UndoMove.
			allOcc := pp.colOcc[WHITE] | pp.colOcc[BLACK]
			if turn == WHITE && m.From.Rank() == 1 && m.To.Rank() == 3 {
				phantomSq := Sq(0, m.From.File())
				if !allOcc.IsSet(phantomSq) {
					pp.pawnOcc |= phantomSq.Bitboard()
				}
			} else if turn == BLACK && m.From.Rank() == 6 && m.To.Rank() == 4 {
				phantomSq := Sq(7, m.From.File())
				if !allOcc.IsSet(phantomSq) {
					pp.pawnOcc |= phantomSq.Bitboard()
				}
			}
		}
	}

	// Revoke opponent's castling right if their rook was captured
	if m.Captured != EMPTY && pabs(m.Captured) == ROOK {
		pp.revokeRookCastle(opponent, m.CaptureSquare)
	}

	return pp, m
}

// ── UndoMove ──────────────────────────────────────────────────────────────────

// UndoMove restores the position before DoMove was called.
// m must be the enriched Move returned by DoMove (undo fields must be intact).
func (p Position) UndoMove(m Move) Position {
	pp := p
	turn := m.PrevStatus.GetTurn()
	opponent := 1 ^ turn

	// Restore all status bits (turn, king positions, castle rights)
	pp.status = m.PrevStatus

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
