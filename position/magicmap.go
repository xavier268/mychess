package position

import (
	"fmt"
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
	Count() (in, out int)  // actual nb of distinct values
	Capa() (in, out int)   // capacity, in max number of distinct values
	Size() (bytes int)     // memory size in bytes
	Set(key, value uint64) // set
	AllKeys() []uint64     // get all keys - testing only
	AllValues() []uint64   // get all values - testing only
	Dump()                 // dump key -> values - testing only
}

const magic uint64 = 11400714819323198485

// 16 bits hash function optimized for speed
func hash16(key uint64) uint16 {
	return uint16((key * magic) >> 24) // middle bits have the most entropy
}

const (
	max12 = 1 << 12
	max11 = 1 << 11
	max6  = 1 << 6
)

// ====================================================
// 12 bits -> 6 bits
// ====================================================

func NewMagicMap_12_6() MagicMap {
	mm := new(magicMap_12_6)
	return mm
}

type magicMap_12_6 struct { // capa is 4096 in/64 out
	keys   [max12]uint64 // keyindex ->  input keys
	index  [max12]uint8  // keyindex -> valueindex ( NOTE : assumes index < 256 for up to 256 distict output values per magicmap).
	values [max6]uint64  // valueindex -> output values
	nbin   uint16        // nb of keys registered
	nbout  uint8         // nb of keys and values registered
}

// Get implements MagicMap.
func (mm magicMap_12_6) Get(key uint64) (value uint64) {
	const maxin = max12
	if key == 0 {
		return 0
	}
	// look for key, linear search if collision, return 0 if not found
	keyindex := (maxin - 1) & hash16(key)
	for {
		if mm.keys[keyindex] == 0 {
			return 0 // not found
		}
		if mm.keys[keyindex] == key {
			return mm.values[mm.index[keyindex]]
		}
		keyindex = (keyindex + 1) & (maxin - 1)
	}
}

func (mm magicMap_12_6) Size() (bytes int) {
	return int(unsafe.Sizeof((mm)))
}

func (mm magicMap_12_6) Count() (in, out int) {
	return int(mm.nbin), int(mm.nbout)
}

func (mm magicMap_12_6) Capa() (in, out int) {
	return max12, max6
}

func (mm *magicMap_12_6) Set(key, value uint64) {
	const maxin = max12
	const maxout = max6

	// check remaining capacity
	if mm.nbout == maxout {
		panic("trying to set a new value beyond value capacity")
	}
	if mm.nbin == maxin {
		panic("trying to set a new key beyond key capacity")
	}
	if key == 0 {
		panic("key cannot be 0")
	}
	// look for value if it exists
	var valueindex uint8
	found := false
	for valueindex = 0; valueindex < mm.nbout; valueindex++ {
		if mm.values[valueindex] == value {
			found = true
			break
		}
	}
	// value was not previously known, lets register it.
	if !found {
		valueindex = uint8(mm.nbout)
		mm.values[mm.nbout] = value
		mm.nbout++
	}
	fmt.Printf("DEBUG : registerd value %d with valueindex %d\n", value, valueindex)
	// OK, now we have the index for the value in valueIndex.

	// look for key, linear search until empty slot found
	keyindex := (maxin - 1) & hash16(key)
	for i := 0; i < (maxin); i = (i + 1) & (maxin - 1) { // never more than 2^12 tries
		if mm.keys[keyindex] == 0 {
			mm.keys[keyindex] = key
			mm.index[keyindex] = valueindex
			mm.nbin++
			return
		}
	}
	panic("internal logic error")
}

func (mm magicMap_12_6) AllKeys() []uint64 {
	keys := make([]uint64, 0, mm.nbin)
	for _, k := range mm.keys {
		if k != 0 {
			keys = append(keys, k)
		}
	}
	return keys
}

func (mm magicMap_12_6) AllValues() (values []uint64) {
	return append(values, mm.values[:mm.nbout]...)
}

func (mm magicMap_12_6) Dump() {
	maxin, maxout := mm.Capa()
	nbin, nbout := mm.Count()
	size := mm.Size()
	fmt.Printf("MagicMap : %d/%d keys, %d/%d values (memory used : %d bytes)\n", nbin, maxin, nbout, maxout, size)
	for i, k := range mm.AllKeys() {
		v := mm.Get(k)
		fmt.Printf("%05d :   %016X (%20d) -> %016X (%20d)\n", i, k, k, v, v)
	}
}
