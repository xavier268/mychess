package position

import (
	"crypto/rand"
	"encoding/binary"
)

// Zobrist table parameters
type ZobristTable struct {

	// Parameters for each bit in each bitboard in a position.
	// [board][bit]
	ZobristBitboards [6][64]uint64 // 3KB

	// Parameters for each king position
	// [color][square]
	ZobristKing [2][64]uint64 // 1KB

	// parameters for each castling possibility
	// [color][castlingbits]
	ZobristCastling [2][4]uint64 // 64B

	// to hash the turn
	ZobristTurn uint64 // 8B
}

// DefaultZT is the package-level Zobrist table, initialised once at startup.
// DoMove uses it to maintain Position.Hash incrementally.
var DefaultZT = func() ZobristTable {
	var zt ZobristTable
	zt.Init()
	return zt
}()

// Should be done once, and saved to file ...
func (z *ZobristTable) Init() {

	// Generate enough "good" random numbers
	// L'utilisation de crypto/rand de meilleure qualité que math/rand/v2,
	// mais pas seedable et un peu plus lent.
	// Pas critique ici, car seulement appélé une fois à l'init.
	const size = 6*64*8 + 2*64*8 + 2*4*8 + 8 // ZobristTurn is uint64 (8 bytes)
	rands := make([]byte, size)
	byteIndex := 0
	rand.Read(rands)

	// Fill the table
	for i := range 6 {
		for j := range 64 {
			z.ZobristBitboards[i][j] = binary.LittleEndian.Uint64(rands[byteIndex : byteIndex+8])
			byteIndex += 8
		}
	}
	for i := range 2 {
		for j := range 64 {
			z.ZobristKing[i][j] = binary.LittleEndian.Uint64(rands[byteIndex : byteIndex+8])
			byteIndex += 8
		}
	}
	for i := range 2 {
		for j := range 4 {
			z.ZobristCastling[i][j] = binary.LittleEndian.Uint64(rands[byteIndex : byteIndex+8])
			byteIndex += 8
		}
	}

	z.ZobristTurn = binary.LittleEndian.Uint64(rands[byteIndex : byteIndex+8])
	byteIndex += 8

	if byteIndex != size {
		panic("internal error - Zobrist table not initialized correctly")
	}

}

// Initial hash for a position.
// Subsequent hashes should be caclcuted by difference, based on what changed only !
func (zt *ZobristTable) HashPosition(p Position) uint64 {
	var hash uint64 = 0

	// hash ColOcc
	for c := range 2 {
		for sq := range p.colOcc[c].AllSetSquares {
			hash ^= zt.ZobristBitboards[c][sq]
		}
	}
	// hash pawnOcc
	for sq := range p.pawnOcc.AllSetSquares {
		hash ^= zt.ZobristBitboards[2][sq]
	}
	// hash rookOcc
	for sq := range p.rookOcc.AllSetSquares {
		hash ^= zt.ZobristBitboards[3][sq]
	}
	// hash bishopOcc
	for sq := range p.bishopOcc.AllSetSquares {
		hash ^= zt.ZobristBitboards[4][sq]
	}
	// hash knightOcc
	for sq := range p.knightOcc.AllSetSquares {
		hash ^= zt.ZobristBitboards[5][sq]
	}
	// hash kingOcc
	for c := range 2 {
		hash ^= zt.ZobristKing[c][p.status.GetKingPosition(uint8(c))]
	}
	// hash castling – GetCastleBits returns {0,0x40,0x80,0xC0}; >>6 maps to 0–3
	for c := range 2 {
		hash ^= zt.ZobristCastling[c][p.status.GetCastleBits(uint8(c))>>6]
	}
	// hash turn
	if p.status.GetTurn() == 1 {
		hash ^= zt.ZobristTurn
	}
	return hash
}

// init seeds the Hash field of StartPosition after DefaultZT is ready.
// All init() functions run after package-level vars are initialised, so
// DefaultZT (a package-level var with an initialiser) is guaranteed to be
// populated by the time this runs.
func init() {
	StartPosition.Hash = DefaultZT.HashPosition(StartPosition)
}
