package position

// ========================================================
// This file contains the tables to compute the attack sets
// ========================================================

// Global tables used to compute attack sets
var (
	// Tables for rook attacks
	RookAttacks [64]MagicMap // derives attacks from occupancy
	RookMasks   [64]BitBoard // max potential
	// Tables for bishop attacks
	BishopAttacks [64]MagicMap // derives attack from occupancy
	BishopMasks   [64]BitBoard // max potential attack
	// Tables for knight attacks
	KnightMasks [64]BitBoard // max attack
	// Tables for king attacks
	KingAttacks [64]BitBoard // max attack
	// Tables for pawn attacks
	WhitePawnAttacks [64]BitBoard // max attack
	BlackPawnAttacks [64]BitBoard // max attack

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
