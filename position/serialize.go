package position

import "math/bits"

// Split the bit board per bit.
func (b Bitboard) Serialize() []Bitboard {
	result := make([]Bitboard, 0, bits.OnesCount64(uint64(b))) // Pre-allocate exact capacity

	// Continue as long as there are bits set
	for b != 0 {
		// Extract the lowest set bit
		lowestBit := b & -b // Or: num & (^num + 1)

		// Add it to our result
		result = append(result, lowestBit)

		// Clear that bit from the original number
		b &= b - 1 // This clears the lowest set bit
	}

	return result
}
