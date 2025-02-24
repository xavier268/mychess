// Compute magic nbr to optimized a map[uint64]uint64
package magic

import (
	"fmt"
	"math"
	"math/rand/v2"
)

// brute force the magic nbr corresponding the provided map
func DoMagic(mm map[uint64]uint64) uint64 {

	keys := make([]uint64, 0, len(mm))
	for k := range mm {
		keys = append(keys, k)
	}
	m2 := make(map[uint64]int, len(mm)) // map the target values to the index

	NbFrom := len(mm)
	NbTo := len(keys)
	NbBits := uint64(math.Ceil(math.Log2(float64(NbTo))))
	fmt.Printf("Trying to compress a map from %d values to %d values (%d bit index for values)\n", NbFrom, NbTo, NbBits)

	for i := rand.Uint64(); true; {
		fmt.Printf(".")
		magic := i
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
				if m2[v] != idx {
					break
				}
			}
		

	}

}
