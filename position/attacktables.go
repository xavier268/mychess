package position

// ========================================================
// This file contains the tables to compute the attack sets
// ========================================================

// Global tables used to compute attack sets
var (

// NB :
// Pawn attacks and moves and Castling do not use precomputed tables.
// Queen is a combination of rook & bishop

// All tables have a square symetry vertical/horizontal, and the square input is reduced to 1/4 of the board (16 instead of 64).

// // Tables for rook attacks
// RookMasks   [16]Bitboard // mask for querying rootAttacks table
// RookAttacks [16]MagicMap // derives attacks from occupancy

// // Tables for bishop attacks
// BishopMasks   [16]Bitboard // mask for querying bishop table
// BishopAttacks [16]MagicMap // derives attack from occupancy

// // Tables for knight attacks
// KnightAttacks [16]Bitboard // max attack

// // Tables for king attacks
// KingAttacks [16]Bitboard // max attack

)

// Load tables from file
func LoadTables(tableFileName string) error {
	panic("todo")
}

// Save tables to file
func SaveTables(tableFileName string) error {
	panic("todo")
}

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
	for occ := range mask.BitCombinations {
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

func GenerateBishopAttacksMagicMapSq(sq Square) (res map[uint64]uint64) {
	res = make(map[uint64]uint64, 1<<4) // start small
	mask := GenerateBishopMaskSq(sq)    // mask for the square occupancy
	// generate all possible occupancy within the above mask
	for occ := range mask.BitCombinations {
		res[uint64(occ)] = uint64(generateBishopAttackSetSqOcc(sq, occ))
	}
	return res
}
