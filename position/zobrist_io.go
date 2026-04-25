package position

import (
	"encoding/binary"
	"io"
)

// Encode serializes the ZobristTable to w in little-endian binary format.
func (zt *ZobristTable) Encode(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, zt)
}

// Decode deserializes a ZobristTable from r.
func (zt *ZobristTable) Decode(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, zt)
}

// RestoreDefaultZT replaces DefaultZT with zt and recomputes StartPosition.Hash.
// Must be called before any search when loading a pre-computed ZMap from cache.
func RestoreDefaultZT(zt ZobristTable) {
	DefaultZT = zt
	StartPosition.Hash = DefaultZT.HashPosition(StartPosition)
}
