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

	// mm contains the initial map : input to output

	// vals contains deduped output values
	vals := make(map[uint64]bool, len(mm))
	for _, v := range mm {
		vals[v] = true
	}

	// m2 maps output values to indexes
	m2 := make(map[uint64]uint64, len(mm))

	NbFrom := len(mm)
	NbTo := len(vals)
	NbBits = int(math.Ceil(math.Log2(float64(NbTo))))
	fmt.Printf("Trying to compress a map from %d input values to %d output values (%d bit index for values)\n", NbFrom, NbTo, NbBits)

	for {

		magic = rand.Uint64()
		fmt.Println("Debug : trying ", magic)

		// clear m2 map
		for k := range m2 {
			delete(m2, k)
		}

		// test magic ?
		for inv, outv := range mm {
			fmt.Println("   ", inv, "-->", outv)
			idx := (magic * inv) >> (64 - NbBits)
			if idx2, ok := m2[outv]; !ok { // this value had no index yet
				m2[outv] = idx
				fmt.Println("  ", inv, "=>", idx, "-->", outv)
			} else { // this value already had an index, idx2 - is it the same as the computed idx ?
				if idx2 != idx { // magic number is invalid !
					fmt.Println("Failed  : value", outv, "idx", idx, "idx2", idx2)
					magic = 0
					break // abort loop
				}
			}
		}
		//  if magic is invalid ? main loop !
		if magic == 0 {
			continue
		}
		// here, magic should be valid
		fmt.Printf("Found valid magic : %d\n", magic)
		values = make([]uint64, 1<<(NbBits))
		for outv := range vals { // loop over deduplicated out values
			values[m2[outv]] = outv
			return magic, NbBits, values
		}
	}
}

// Apply computed magic nbr to input to generate output
func ApplyMagic(magic uint64, NbBits int, values []uint64, input uint64) (output uint64) {
	return values[(magic*input)>>(64-NbBits)]
}
