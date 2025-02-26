package position

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
	Count() (in, out int)  // actual nb of distinct values, (0 key is always defined, not counted)
	Capa() (in, out int)   // capacity, in max number of distinct values
	Size() (bytes int)     // memory size in bytes
	Set(key, value uint64) // set
	AllKeys() []uint64     // get all keys - testing only, excluding 0 key
	AllValues() []uint64   // get all values - testing only, excluding the value of the 0 key
	Dump()                 // dump key -> values - testing only
}

const defaultmagic uint64 = 11400714819323198485
