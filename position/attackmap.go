package position

// ========================================================
// This file contains the tables to compute the attack sets
// ========================================================

// Global tables used to compute attack sets
var (

	// Tables for rook attacks
	RookAttacks [64]MagicMap // derives attacks from occupancy
	RookMasks   [64]Bitboard // max potential
	// Tables for bishop attacks
	BishopAttacks [64]MagicMap // derives attack from occupancy
	BishopMasks   [64]Bitboard // max potential attack
	// Tables for knight attacks
	KnightMasks [64]Bitboard // max attack
	// Tables for king attacks
	KingAttacks [64]Bitboard // max attack
	// Tables for pawn attacks
	WhitePawnAttacks [64]Bitboard // max attack
	BlackPawnAttacks [64]Bitboard // max attack

	// NB : Queen is a combination of rook & bishop
)

// Load tables from file
func LoadTables(tableFileName string) error {
	panic("todo")
}

// Save tables to file
func SaveTables(tableFileName string) error {
	panic("todo")
}

func initRooksTables() {
	for sq := Square(0); sq < 64; sq++ {
		r, f := sq.RF()
		// compute the rook mask, does not include self.
		var m Bitboard
		for k := 1; k < 7; k++ {
			if r+k < 8 {
				m.Set(Sq(r+k, f))
			}
			if r-k >= 0 {
				m.Set(Sq(r-k, f))
			}
			if f+k < 8 {
				m.Set(Sq(r, f+k))
			}
			if f-k >= 0 {
				m.Set(Sq(r, f-k))
			}
		}
		RookMasks[sq] = m

		// select the right size of attack magicMap ?

		// create and initialize magic map

		panic("todo")
		// TBC XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

	}
}
