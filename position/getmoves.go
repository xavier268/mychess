package position

// It is always assumed  white has to move

func (p Position) GetKnightMovesFromSquare(bt *BigTable, sq Square) (res Bitboard) {
	return bt.KnightAttacks[sq] & ^p.whiteOcc
}

func (p Position) GetKingMovesFromSquare(bt *BigTable, sq Square) (res Bitboard) {
	return bt.KingAttacks[sq] & ^p.whiteOcc
}

func (p Position) GetRookMovesFromSquare(bt *BigTable, sq Square) (res Bitboard) {
	occ := p.whiteOcc | p.blackOcc
	key := occ & bt.RookMask[sq]
	attacks := Bitboard(bt.Get(uint8(SquareTable(sq, 0)), uint64(key)))
	return attacks & ^p.whiteOcc
}

func (p Position) GetBishopMovesFromSquare(bt *BigTable, sq Square) (res Bitboard) {
	occ := p.whiteOcc | p.blackOcc
	key := occ & bt.BishopMask[sq]
	attacks := Bitboard(bt.Get(uint8(SquareTable(sq, 1)), uint64(key)))
	return attacks & ^p.whiteOcc
}

func (p Position) GetWhitePawnMovesFromSquare(bt *BigTable, sq Square) (res Bitboard) {
	occ := p.whiteOcc | p.blackOcc
	return (bt.WhitePawnCaptureMask[sq] & p.blackOcc) | // capture ONLY if opponent
		(bt.WhitePawnMoveMask[sq] & ^occ) // Move ONLY if empty
}

// Queen moves not needed - handled automatically ...

// All moves from the specified position in a single bitboard
func (p Position) GetMoves(bt *BigTable, sq Square) (res Bitboard) {

	return p.GetKnightMovesFromSquare(bt, sq) |
		p.GetBishopMovesFromSquare(bt, sq) |
		p.GetKingMovesFromSquare(bt, sq) |
		p.GetRookMovesFromSquare(bt, sq) |
		p.GetWhitePawnMovesFromSquare(bt, sq)
}

// Compute if the specified square (one of the kings) is currently under attack.
func (p Position) IsWhiteKingAttacked(bt *BigTable) bool {
	sq := p.GetWhiteKingSquare()

	return (p.GetKnightMovesFromSquare(bt, sq)&p.blackOcc&p.knightOcc != 0) || // black knights attacking ?
		(p.GetBishopMovesFromSquare(bt, sq)&p.blackOcc&p.bishopOcc != 0) ||
		(p.GetRookMovesFromSquare(bt, sq)&p.blackOcc&p.rookOcc != 0) ||
		(p.GetKingMovesFromSquare(bt, sq)&(1<<p.GetBlackKingSquare()) != 0) ||
		(p.GetWhitePawnMovesFromSquare(bt, sq)&p.blackOcc&p.pawnOcc != 0)

	panic("todo")

}
