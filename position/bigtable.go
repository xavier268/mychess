package position

// Bigtable contains all the fixed structures required to compute attack sets and moves of a given position.
// There should only be one such straucture, and it is immutable.
// It should be cached in file as much as possible, since its construction is cpu/memory intensive.
type BigTable struct {
	// Precomputed simple attacks
	// Square -> AttackSet
	KingAttacks   [64]Bitboard
	KnightAttacks [64]Bitboard

	// Sliding pieces require a mask
	// Queen is a combination of the 4 bishop & rook masks
	RookMaskRank, RookMaskFile [64]Bitboard
	BishopMaskNE, BishopMaskNW [64]Bitboard
	PawnMask                   [2][64]Bitboard // (color, square) -> Mask of all useful info (2 cases devant, les deux cases de prises possibles)

	// Attack sets for sliding pieces
	RookAttackSetRank, RookAttackSetFile [64]map[Bitboard]Bitboard // (square, maskedOccupancy) -> AttackSet
	BishopAttackSetNE, BishopAttackSetNW [64]map[Bitboard]Bitboard // (square, maskedOccupancy) -> AttackSet

	// Attack sets for pawns
	PawnAttackSet [2][64]map[Bitboard]Bitboard // (color, square) -> (occupancy -> AttackSet)

}

// Create and initialize a new BigTable
func NewBigTable() *BigTable {
	b := new(BigTable)

	// Initialize masks
	for sq := Square(0); sq < 64; sq++ {
		r, f := sq.RF()

		b.KingAttacks[sq] = GenerateKingAttacksSq(sq)
		b.KnightAttacks[sq] = GenerateKnightAttacksSq(sq)

		b.RookMaskRank[sq] = Rank(r).Unset(sq).Unset(Sq(r, 0)).Unset(Sq(r, 7))
		b.RookMaskFile[sq] = File(f).Unset(sq).Unset(Sq(0, f)).Unset(Sq(7, f))
		b.BishopMaskNE[sq] = Diagonal(sq).Unset(sq) & Interior()
		b.BishopMaskNW[sq] = AntiDiagonal(sq).Unset(sq) & Interior()
		b.PawnMask[WHITE][sq] = GenerateWhitePawnMoveMaskSq(sq) | GenerateWhitePawnCaptureMaskSq(sq)
		b.PawnMask[BLACK][sq] = GenerateBlackPawnMoveMaskSq(sq) | GenerateBlackPawnCaptureMaskSq(sq)
	}

	// Build attack set maps
	for sq := Square(0); sq < 64; sq++ {
		mask := b.RookMaskRank[sq]
		b.RookAttackSetRank[sq] = make(map[Bitboard]Bitboard, 1<<mask.BitCount())
		for occ := range mask.AllBitCombinations {
			b.RookAttackSetRank[sq][occ] = generateRookRankAttackSetSqOcc(sq, occ)
		}

		mask = b.RookMaskFile[sq]
		b.RookAttackSetFile[sq] = make(map[Bitboard]Bitboard, 1<<mask.BitCount())
		for occ := range mask.AllBitCombinations {
			b.RookAttackSetFile[sq][occ] = generateRookFileAttackSetSqOcc(sq, occ)
		}

		mask = b.BishopMaskNE[sq]
		b.BishopAttackSetNE[sq] = make(map[Bitboard]Bitboard, 1<<mask.BitCount())
		for occ := range mask.AllBitCombinations {
			b.BishopAttackSetNE[sq][occ] = generateBishopNEAttackSetSqOcc(sq, occ)
		}

		mask = b.BishopMaskNW[sq]
		b.BishopAttackSetNW[sq] = make(map[Bitboard]Bitboard, 1<<mask.BitCount())
		for occ := range mask.AllBitCombinations {
			b.BishopAttackSetNW[sq][occ] = generateBishopNWAttackSetSqOcc(sq, occ)
		}

		b.PawnAttackSet[WHITE][sq] = generatePawnAttackMapSq(WHITE, sq)
		b.PawnAttackSet[BLACK][sq] = generatePawnAttackMapSq(BLACK, sq)
	}

	return b
}
