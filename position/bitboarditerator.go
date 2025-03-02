package position

import "math/bits"

// AllBitCombinations iterates through all possible bit combinations of this mask
// using the yield function pattern with Go's newer iteration approach.
func (mask Bitboard) AllBitCombinations(yield func(Bitboard) bool) {
	// First yield the empty combination (0)
	if !yield(0) {
		return
	}

	// Extract set bits from the mask
	var bits []Bitboard
	tempMask := mask
	for tempMask > 0 {
		lowestBit := tempMask & -tempMask
		bits = append(bits, lowestBit)
		tempMask &= ^lowestBit
	}

	// Calculate the total number of combinations (2^bitCount)
	bitCount := len(bits)
	total := 1 << bitCount

	// Generate and yield all non-empty combinations
	for i := 1; i < total; i++ {
		var combination Bitboard
		for j, bit := range bits {
			if (i & (1 << j)) != 0 {
				combination |= bit
			}
		}

		if !yield(combination) {
			return
		}
	}
}

// Range iterator that returns the Squares from this Bitboard that are set.
// Ordered in natural order.
func (b Bitboard) AllSetSquares(yield func(Square) bool) {

	// Early exit if no bits are set
	bb := uint64(b)
	if bb == 0 {
		return
	}

	// Iterate through all set bits
	for bb != 0 {
		// Find position of the least significant bit
		// This is equivalent to the Square index
		sq := Square(bits.TrailingZeros64(bb))

		// Call the yield function and check if we should continue
		if !yield(Square(sq)) {
			return
		}

		// Clear the least significant bit and continue
		bb &= bb - 1
	}
}
