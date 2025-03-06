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
	key := occ & bt.RookMask[sq]
	attacks := Bitboard(bt.Get(uint8(SquareTable(sq, 0)), uint64(key)))
	return attacks & ^p.colOcc[turn]
}

func (p Position) GetBishopMovesFromSquareBB(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]
	key := occ & bt.BishopMask[sq]
	attacks := Bitboard(bt.Get(uint8(SquareTable(sq, 1)), uint64(key)))
	return attacks & ^p.colOcc[turn]
}

func (p Position) GetPawnMovesFromSquareBB(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]
	return (bt.PawnCaptureMask[turn][sq] & p.colOcc[1^turn]) | // capture ONLY if opponent
		(bt.PawnMoveMask[turn][sq] & ^occ) // Move ONLY if empty
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

// TODO - handle en passant & castling !
