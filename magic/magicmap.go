package magic

import (
	"fmt"
	"math/bits"
	"unsafe"
)

// MagicMap is a map optimized (memory/cpu) for retrieval, using 64-bits words both as keys and values.
// Optimization uses the fact that the number of distinct input values and output values are smaller that 2^64, with much less output than input.
// Set panic if attempting to store beyond capacity.
// Set panic if attempting to set the 0 key.
// You are not supposed to set the same key twice.
// Not optimized for storing. Only Get() is optimized.
// CAUTION : key should NEVER be 0. (0 is used as a marker for free slots).
// 0 value returned for keys not found.
type MagicMap interface {
	Get(key uint64) (value uint64) // get - this is the main, optimized, method
	// The following methods are not optimized, do not use at runtime
	Count() (in, out int)           // actual nb of distinct values, (0 key is always defined, not counted)
	Capa() (in, out int)            // capacity, in max number of distinct values
	Size() (bytes int)              // memory size in bytes
	Set(key, value uint64)          // set
	AllKeys() []uint64              // get all keys - testing only, excluding 0 key
	AllValues() []uint64            // get all values - testing only, excluding the value of the 0 key
	Dump()                          // dump key -> values - testing only
	SetMagic(magic uint64) MagicMap //magic value - testing only
	Stats() MagicStats              // statistics on collisions - testing only
}

const defaultmagic uint64 = 11400714819323198485

type MagicStats struct {
	Coll      int // nbr of key collisions
	Maxsearch int // max nbr of key searches
	Sumsearch int // sum of key searches of the entire keyset
}

// Construct in a deterministic way a magicmap from a go map.
func GoMap2MagicMap(m map[uint64]uint64) (mm MagicMap) {

	fmt.Printf("Processing a go map of %d bytes\n", int(unsafe.Sizeof((mm))))

	// Compte nbr of keys and nbr of DISTINCTS values
	nbkeys := len(m)
	dedup := make(map[uint64]bool, len(m))
	// dedup values
	for _, v := range m {
		dedup[v] = true
	}
	nbvalues := len(dedup)
	dedup = nil

	// compute how many bits are needed to store the keys and the distincts values
	in, out := bits.Len(uint(nbkeys)), bits.Len(uint(nbvalues))

	// Create the adequate type of magic map
	mm = NewMagicMap(in, out)

	// Fill the map
	for k, v := range m {
		mm.Set((k), (v))
	}

	fmt.Printf("Created and filled a magicmap ( mem size %d bytes)\n", mm.Size())

	return mm
}
