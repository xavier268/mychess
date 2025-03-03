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

// Range iterator over serialized boards for each bit set
func (b Bitboard) AllSerialized(yield func(Bitboard) bool) {
	// Continue as long as there are bits set
	for b != 0 {
		// Extract the lowest set bit
		lowestBit := b & -b // Or: num & (^num + 1)

		if !yield(lowestBit) {
			return
		}

		// Clear that bit from the original number
		b &= b - 1 // This clears the lowest set bit
	}

}
