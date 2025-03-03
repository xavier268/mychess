package position

// ========================================================
// This file contains the tables to compute the attack sets
// ========================================================

// Generate rook mask for square.
// Mask does not include square, nor extreme values on the border of the chess board.
func GenerateRookMaskSq(sq Square) Bitboard {
	r, f := sq.RF()
	return (Rank(r) ^ File(f)).
		// cannot use Interior(), because rook could contain a full border file or rank.
		Unset(Sq(r, 0)).
		Unset(Sq(r, 7)).
		Unset(Sq(0, f)).
		Unset(Sq(7, f))
}

// Generate bishop mask for square.
// Mask does not include square, nor extreme values on the border of the chess board.
func GenerateBishopMaskSq(sq Square) Bitboard {
	//r, f := sq.RF()
	return (Diagonal(sq) ^ AntiDiagonal(sq)) & Interior()

}

// Generate attack set for knight positions
func GenerateKnightAttacksSq(sq Square) Bitboard {
	r, f := sq.RF()
	b := Bitboard(0)
	// Generate KNIGHT positions
	if (r+2 < 8) && (f+1) < 8 {
		b = b.Set(Sq(r+2, f+1))
	}
	if (r+2 < 8) && (f-1) >= 0 {
		b = b.Set(Sq(r+2, f-1))
	}
	if (r-2 >= 0) && (f+1) < 8 {
		b = b.Set(Sq(r-2, f+1))
	}
	if (r-2 >= 0) && (f-1) >= 0 {
		b = b.Set(Sq(r-2, f-1))
	}
	if (r+1 < 8) && (f+2) < 8 {
		b = b.Set(Sq(r+1, f+2))
	}
	if (r+1 < 8) && (f-2) >= 0 {
		b = b.Set(Sq(r+1, f-2))
	}
	if (r-1 >= 0) && (f+2) < 8 {
		b = b.Set(Sq(r-1, f+2))
	}
	if (r-1 >= 0) && (f-2) >= 0 {
		b = b.Set(Sq(r-1, f-2))
	}
	// return
	return b
}

// Castling is NOT covered here
func GenerateKingAttacksSq(sq Square) Bitboard {
	r, f := sq.RF()
	b := Bitboard(0)
	// Generate KING positions
	if (r+1 < 8) && (f+1) < 8 {
		b = b.Set(Sq(r+1, f+1))
	}
	if (r+1 < 8) && (f-1) >= 0 {
		b = b.Set(Sq(r+1, f-1))
	}
	if (r-1 >= 0) && (f+1) < 8 {
		b = b.Set(Sq(r-1, f+1))
	}
	if (r-1 >= 0) && (f-1) >= 0 {
		b = b.Set(Sq(r-1, f-1))
	}
	if r+1 < 8 {
		b = b.Set(Sq(r+1, f))
	}
	if r-1 >= 0 {
		b = b.Set(Sq(r-1, f))
	}
	if (f + 1) < 8 {
		b = b.Set(Sq(r, f+1))
	}
	if (f - 1) >= 0 {
		b = b.Set(Sq(r, f-1))
	}
	// return
	return b
}

// low level generation.
// occ is the occupancy of the board (both colors), already masked with RookMask.
func generateRookAttackSetSqOcc(sq Square, occ Bitboard) Bitboard {
	r, f := sq.RF()
	as := Bitboard(0) // default attack set
	var i int

	// north
	for i = r + 1; i < 8; i++ {
		as = as.Set(Sq(i, f))
		if occ.IsSet(Sq(i, f)) {
			break // break after adding the 1srt occupancy
		}
	}
	// south
	for i = r - 1; i >= 0; i-- {
		as = as.Set(Sq(i, f))
		if occ.IsSet(Sq(i, f)) {
			break // break after adding the 1srt occupancy
		}
	}
	// east
	for i = f + 1; i < 8; i++ {
		as = as.Set(Sq(r, i))
		if occ.IsSet(Sq(r, i)) {
			break // break after adding the 1srt occupancy
		}
	}
	// west
	for i = f - 1; i >= 0; i-- {
		as = as.Set(Sq(r, i))
		if occ.IsSet(Sq(r, i)) {
			break // break after adding the 1srt occupancy
		}
	}
	return as
}

// Generate the magic maps for rook attacks for the given square
func GenerateRookAttacksMagicMapSq(sq Square) (res map[uint64]uint64) {
	res = make(map[uint64]uint64, 1<<6) // start small
	mask := GenerateRookMaskSq(sq)      // mask for the square occupancy
	// generate all possible occupancy within the above mask
	for occ := range mask.AllBitCombinations {
		res[uint64(occ)] = uint64(generateRookAttackSetSqOcc(sq, occ))
	}

	return res
}

func generateBishopAttackSetSqOcc(sq Square, occ Bitboard) Bitboard {
	r, f := sq.RF()
	as := Bitboard(0) // default attack set
	var i, j int

	// north east
	for i, j = r+1, f+1; i < 8 && j < 8; i, j = i+1, j+1 {
		as = as.Set(Sq(i, j))
		if occ.IsSet(Sq(i, j)) {
			break // break after adding the 1srt occupancy
		}
	}
	// north west
	for i, j = r+1, f-1; i < 8 && j >= 0; i, j = i+1, j-1 {
		as = as.Set(Sq(i, j))
		if occ.IsSet(Sq(i, j)) {
			break // break after adding the 1srt occupancy
		}
	}
	// south east
	for i, j = r-1, f+1; i >= 0 && j < 8; i, j = i-1, j+1 {
		as = as.Set(Sq(i, j))
		if occ.IsSet(Sq(i, j)) {
			break // break after adding the 1srt occupancy
		}
	}
	// south west
	for i, j = r-1, f-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		as = as.Set(Sq(i, j))
		if occ.IsSet(Sq(i, j)) {
			break // break after adding the 1srt occupancy
		}
	}
	return as
}

func GenerateBishopAttacksSq(sq Square) (res map[uint64]uint64) {
	res = make(map[uint64]uint64, 1<<4) // start small
	mask := GenerateBishopMaskSq(sq)    // mask for the square occupancy
	// generate all possible occupancy within the above mask
	for occ := range mask.AllBitCombinations {
		res[uint64(occ)] = uint64(generateBishopAttackSetSqOcc(sq, occ))
	}
	return res
}

func GenerateBishopAttacksMagicMapSq(sq Square) (res map[uint64]uint64) {
	res = make(map[uint64]uint64, 1<<8) // start small
	mask := GenerateBishopMaskSq(sq)    // mask for the square occupancy
	// generate all possible occupancy within the above mask
	for occ := range mask.AllBitCombinations {
		res[uint64(occ)] = uint64(generateBishopAttackSetSqOcc(sq, occ))
	}
	return res
}

func GenerateWhitePawnMoveMaskSq(sq Square) Bitboard {
	r, f := sq.RF()
	b := Bitboard(0)

	if r < 7 {
		// Moves
		b = b.Set(Sq(r+1, f))
		if r == 1 {
			b = b.Set(Sq(r+2, f))
		}
	}
	return b
}

func GenerateBlackPawnMoveMaskSq(sq Square) Bitboard {
	r, f := sq.RF()
	b := Bitboard(0)

	if r > 0 {
		// Moves
		b = b.Set(Sq(r-1, f))
		if r == 6 {
			b = b.Set(Sq(r-2, f))
		}
	}
	return b
}

func GenerateWhitePawnCaptureMaskSq(sq Square) Bitboard {
	r, f := sq.RF()
	b := Bitboard(0)
	// Captures
	if r < 7 {
		if f > 0 {
			b = b.Set(Sq(r+1, f-1))
		}
		if f < 7 {
			b = b.Set(Sq(r+1, f+1))
		}
	}
	return b
}

func GenerateBlackPawnCaptureMaskSq(sq Square) Bitboard {
	r, f := sq.RF()
	b := Bitboard(0)
	// Captures
	if r > 0 {
		if f > 0 {
			b = b.Set(Sq(r-1, f-1))
		}
		if f < 7 {
			b = b.Set(Sq(r-1, f+1))
		}
	}
	return b
}
