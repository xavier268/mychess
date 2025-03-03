package position

import (
	"fmt"
	"mychess/magic"
)

// Bigtable contains all the fixed structures required to compute attack sets and moves of a given position.
// There should only be one such straucture, and it is immutable.
// It should be cached in file as much as possible, since its construction is cpu/memory intensive.
type BigTable struct {
	// Precomputed simple attacks
	KingAttacks   [64]Bitboard
	KnightAttacks [64]Bitboard

	// Sliding pieces require a mask
	// Queen is a combination of bishop & rook
	RookMask        [64]Bitboard
	BishopMask      [64]Bitboard
	PawnCaptureMask [2][64]Bitboard
	PawnMoveMask    [2][64]Bitboard

	// Complex attack sets are stored in a (single) MagicMap
	// Tables are :
	// 0-  rookattacksets
	// 1-  bishop attack sets
	// 2-  XXXXXX whitepawn attack sets  ==> INUTILE ?? // TODO
	// Since we have less than 4 tables, we do not need to use reduced squares and mirror the results.
	*magic.MagicMap
}

// Create and initialize a new BigTable
func NewBigTable() *BigTable {
	b := new(BigTable)
	b.MagicMap = new(magic.MagicMap)

	// Initialize the simple attacks
	for sq := Square(0); sq < 64; sq++ {
		b.KingAttacks[sq] = GenerateKingAttacksSq(sq)
		b.KnightAttacks[sq] = GenerateKnightAttacksSq(sq)
		b.RookMask[sq] = GenerateRookMaskSq(sq)
		b.BishopMask[sq] = GenerateBishopMaskSq(sq)
		b.PawnCaptureMask[WHITE][sq] = GenerateWhitePawnCaptureMaskSq(sq)
		b.PawnCaptureMask[BLACK][sq] = GenerateBlackPawnCaptureMaskSq(sq)
		b.PawnMoveMask[WHITE][sq] = GenerateWhitePawnMoveMaskSq(sq)
		b.PawnMoveMask[BLACK][sq] = GenerateBlackPawnMoveMaskSq(sq)
	}

	// Prepare table entries into the MagicMap
	// 0-  rookattacksets
	// 1-  bishop attack sets
	tes := make([]magic.TableEntry, 0, 256)
	var te magic.TableEntry
	for sq := Square(0); sq < 64; sq++ {
		// rook - table 0
		te = magic.TableEntry{Sqt: uint8(SquareTable(sq, 0)), Values: GenerateRookAttacksMagicMapSq(sq)}
		tes = append(tes, te)
		// bishop - table 1
		te = magic.TableEntry{Sqt: uint8(SquareTable(sq, 1)), Values: GenerateBishopAttacksMagicMapSq(sq)}
		tes = append(tes, te)
	}
	st := magic.InitMagicMap(b.MagicMap, tes...)
	fmt.Println(st.String())

	return b
}
