package position

// BitCombinations iterates through all possible bit combinations of this mask
// using the yield function pattern with Go's newer iteration approach.
func (mask Bitboard) BitCombinations(yield func(Bitboard) bool) {
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
