// A home-made chess program constructed for educationnal and testing purposes.
// However, playing level is already quite interesting ...
package mychess

const VERSION = "0.4.4"
const COPYRIGHT = "(c) 2025-2026 by Xavier Gandillot (aka xavier268)"

// binary format constants for caching files identification
var CacheMagic = [8]byte{'M', 'Y', 'C', 'H', 'C', 'A', 'C', 'H'}

// Dynamic vars
var (
	BUILDDATE string
	BUILDHASH string
)
