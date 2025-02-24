// Compute magic nbr to optimized a map[uint64]uint64
package magic

import (
	"fmt"
	"math"
	"math/rand/v2"
)

// brute force the magic nbr corresponding the provided map
// returns the magic number and the slice, from the index (generated with magic number) to the value
func DoMagic(inputOutputMap map[uint64]uint64) (magic uint64, NbBits int, values []uint64) {

	// dedupOutVals contains deduped output values
	dedupOutVals := make(map[uint64]bool, len(inputOutputMap))
	for _, v := range inputOutputMap {
		dedupOutVals[v] = true
	}

	const MAXUINT = 0xFFFFFFFFFFFFFFFF // sentinel value for out values

	NbIn := len(inputOutputMap)
	NbOut := len(dedupOutVals)
	NbBits = int(math.Ceil(math.Log2(float64(NbOut))))
	fmt.Printf("Trying to compress a map from %d input values to %d output values (%d bit index for values)\n", NbIn, NbOut, NbBits)

	// slice of index to values -
	values = make([]uint64, 1<<(NbBits))
	for {

		magic = rand.Uint64()
		fmt.Println("Debug : trying ", magic)

		// clear values
		for k := range values {
			values[k] = MAXUINT // sentinel value. This value is normally never found in real life.
		}

		// test magic ?
		for inv, outv := range inputOutputMap {
			//fmt.Println("Testing :", inv, "-->", outv)
			idx := (magic * inv) >> (64 - NbBits)
			if values[idx] == MAXUINT { // index not yet set to this value (excpt values[0] which is always set to 0)
				values[idx] = outv
				//fmt.Printf("Setting index\n%d ->[%d]->%d\n", inv, idx, outv)
			} else { // this value already had an index, idx2 - is it the same as the computed idx ?
				if values[idx] != outv {
					//fmt.Printf("conflicting values for index :[%d]->%d  and %d\nabort\n", idx, outv, values[idx])
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
		fmt.Printf("Found valid magic : %d\noutputToIndexMap%v\n", magic, values)
		return magic, NbBits, values

	}
}

// Apply computed magic nbr to input to generate output
func ApplyMagic(magic uint64, NbBits int, values []uint64, input uint64) (output uint64) {
	return values[(magic*input)>>(64-NbBits)]
}
