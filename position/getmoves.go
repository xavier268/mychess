package position

// Answers depends on who is expected to move (see status.Turn)

func (p Position) GetKnightMovesFromSquare(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	return bt.KnightAttacks[sq] & ^p.colOcc[turn]
}

func (p Position) GetKingMovesFromSquare(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	return bt.KingAttacks[sq] & ^p.colOcc[turn]
}

func (p Position) GetRookMovesFromSquare(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]
	key := occ & bt.RookMask[sq]
	attacks := Bitboard(bt.Get(uint8(SquareTable(sq, 0)), uint64(key)))
	return attacks & ^p.colOcc[turn]
}

func (p Position) GetBishopMovesFromSquare(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]
	key := occ & bt.BishopMask[sq]
	attacks := Bitboard(bt.Get(uint8(SquareTable(sq, 1)), uint64(key)))
	return attacks & ^p.colOcc[turn]
}

func (p Position) GetPawnMovesFromSquare(bt *BigTable, turn uint8, sq Square) (res Bitboard) {
	occ := p.colOcc[WHITE] | p.colOcc[BLACK]
	return (bt.PawnCaptureMask[turn][sq] & p.colOcc[1^turn]) | // capture ONLY if opponent
		(bt.PawnMoveMask[turn][sq] & ^occ) // Move ONLY if empty
}

// Queen moves not needed - handled automatically ...

// All moves from the specified position in a single bitboard
func (p Position) GetMoves(bt *BigTable, sq Square) (res Bitboard) {
	turn := p.status.Turn()
	return p.GetKnightMovesFromSquare(bt, turn, sq) |
		p.GetBishopMovesFromSquare(bt, turn, sq) |
		p.GetKingMovesFromSquare(bt, turn, sq) |
		p.GetRookMovesFromSquare(bt, turn, sq) |
		p.GetPawnMovesFromSquare(bt, turn, sq)
}

// Compute if the specified square is currently under attack from specified color (by)
func (p Position) IsSquareAttacked(bt *BigTable, sq Square, by uint8) bool {
	return (p.GetKnightMovesFromSquare(bt, 1^by, sq)&p.colOcc[by]&p.knightOcc != 0) ||
		(p.GetBishopMovesFromSquare(bt, 1^by, sq)&p.colOcc[by]&p.bishopOcc != 0) ||
		(p.GetRookMovesFromSquare(bt, 1^by, sq)&p.colOcc[by]&p.rookOcc != 0) ||
		(p.GetKingMovesFromSquare(bt, 1^by, sq)&(1<<p.status.KingPosition[by]) != 0) ||
		(p.GetPawnMovesFromSquare(bt, 1^by, sq)&p.colOcc[by]&p.pawnOcc != 0)
}

// TODO - handle en passant & castling !
