package position

func (p Position) GetKnightMovesFromSquareBB(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	return bt.KnightAttacks[sq] & ^p.colOcc[turn]
}

// Not including castling ...
func (p Position) GetKingMovesFromSquareBB(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	return bt.KingAttacks[sq] & ^p.colOcc[turn]
}

func (p Position) GetRookMovesFromSquareBB(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]
	rankAttacks := bt.RookAttackSetRank[sq][occ&bt.RookMaskRank[sq]]
	fileAttacks := bt.RookAttackSetFile[sq][occ&bt.RookMaskFile[sq]]
	return (rankAttacks | fileAttacks) & ^p.colOcc[turn]
}

func (p Position) GetBishopMovesFromSquareBB(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]
	neAttacks := bt.BishopAttackSetNE[sq][occ&bt.BishopMaskNE[sq]]
	nwAttacks := bt.BishopAttackSetNW[sq][occ&bt.BishopMaskNW[sq]]
	return (neAttacks | nwAttacks) & ^p.colOcc[turn]
}

// GetPawnMovesFromSquareBB returns all pawn moves including en passant.
// The PawnAttackSet map handles forward-move blocking and regular captures.
// En passant is detected via phantom pawns (in pawnOcc but not in colOcc) at rank 0 (white EP) or rank 7 (black EP).
func (p Position) GetPawnMovesFromSquareBB(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]
	res = bt.PawnAttackSet[turn][sq][occ&bt.PawnMask[turn][sq]] & ^p.colOcc[turn]

	// En passant: phantom pawns are in pawnOcc but not in colOcc
	phantoms := p.pawnOcc & ^occ
	r, f := sq.RF()
	if turn == WHITE && r == 4 {
		// Black's en passant signal is a phantom at rank 7 (adjacent file)
		if f > 0 && phantoms.IsSet(Sq(7, f-1)) {
			res = res.Set(Sq(5, f-1))
		}
		if f < 7 && phantoms.IsSet(Sq(7, f+1)) {
			res = res.Set(Sq(5, f+1))
		}
	} else if turn == BLACK && r == 3 {
		// White's en passant signal is a phantom at rank 0 (adjacent file)
		if f > 0 && phantoms.IsSet(Sq(0, f-1)) {
			res = res.Set(Sq(2, f-1))
		}
		if f < 7 && phantoms.IsSet(Sq(0, f+1)) {
			res = res.Set(Sq(2, f+1))
		}
	}
	return res
}

func (p Position) GetQueenMovesFromSquareBB(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	return (p.GetBishopMovesFromSquareBB(bt, turn, sq) | p.GetRookMovesFromSquareBB(bt, turn, sq))
}

// All moves from the specified position in a single bitboard
func (p Position) GetMovesBB(bt *BigTable, sq Square) (res Bitboard) {
	turn := p.status.GetTurn()

	switch {
	case p.pawnOcc&(1<<sq) != 0:
		return p.GetPawnMovesFromSquareBB(bt, turn, sq)
	case p.knightOcc&(1<<sq) != 0:
		return p.GetKnightMovesFromSquareBB(bt, turn, sq)
	case p.bishopOcc&p.rookOcc&(1<<sq) != 0:
		return p.GetQueenMovesFromSquareBB(bt, turn, sq)
	case p.bishopOcc&(1<<sq) != 0:
		return p.GetBishopMovesFromSquareBB(bt, turn, sq)
	case p.rookOcc&(1<<sq) != 0:
		return p.GetRookMovesFromSquareBB(bt, turn, sq)
	case sq == p.status.GetKingPosition(turn):
		return p.GetKingMovesFromSquareBB(bt, turn, sq)
	}
	return Bitboard(0)
}

// Compute if the specified square is currently under attack from specified color (by)
func (p Position) IsSquareAttacked(bt *BigTable, sq Square, by uint8) bool {
	return (p.GetKnightMovesFromSquareBB(bt, 1^by, sq)&p.colOcc[by]&p.knightOcc != 0) ||
		(p.GetBishopMovesFromSquareBB(bt, 1^by, sq)&p.colOcc[by]&p.bishopOcc != 0) ||
		(p.GetRookMovesFromSquareBB(bt, 1^by, sq)&p.colOcc[by]&p.rookOcc != 0) ||
		(p.GetKingMovesFromSquareBB(bt, 1^by, sq)&(1<<p.status.GetKingPosition(by)) != 0) ||
		(p.GetPawnMovesFromSquareBB(bt, 1^by, sq)&p.colOcc[by]&p.pawnOcc != 0)
}

// TODO - handle castling !
