// Compute magic nbr to optimized a map[uint64]uint64
package magic

import (
	"fmt"
	"math"
	"math/rand/v2"
)

// brute force the magic nbr corresponding the provided map
// returns the magic number and the slice, from the index (generated with magic number) to the value
func DoMagic(mm map[uint64]uint64) (magic uint64, NbBits int, values []uint64) {

	// Dedup output values
	vals := make(map[uint64]bool, len(mm))
	for _, v := range mm {
		vals[v] = true
	}

	m2 := make(map[uint64]uint64, len(mm)) // map the target values to the index pointing to that value

	NbFrom := len(mm)
	NbTo := len(vals)
	NbBits = int(math.Ceil(math.Log2(float64(NbTo))))
	fmt.Printf("Trying to compress a map from %d input values to %d output values (%d bit index for values)\n", NbFrom, NbTo, NbBits)

	for {

		magic := rand.Uint64()
		// clear m2 map
		for k := range m2 {
			delete(m2, k)
		}

		// test magic ?
		for k, v := range mm {
			idx := (magic * k) >> (64 - NbBits)
			if _, ok := m2[v]; !ok { // this value had no index yet
				m2[v] = idx
			} else { // this value already had an index - is it the same ?
				if m2[v] != idx { // magic number is invalid !
					magic = 0
					break // abort loop
				}
			}
		}
		// here, test succeeded if magic != 0
		if magic != 0 {
			fmt.Printf("Found valid magic : %d\n", magic)
			values = make([]uint64, 1<<NbTo)
			for val, idx := range m2 {
				values[idx] = val
			}
			return magic, NbBits, values
		}
	}
}

// Apply computed magic nbr to input to generate output
func ApplyMagic(magic uint64, NbBits int, values []uint64, input uint64) (output uint64) {
	return values[(magic*input)>>(64-NbBits)]
}
