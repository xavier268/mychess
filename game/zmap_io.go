package game

import (
	"encoding/binary"
	"io"
	"unsafe"
)

// FillPercent returns the percentage of used entries in the ZMap (0–100).
func (z *ZMap) FillPercent() float64 {
	return 100.0 * float64(z.cellCount) / float64(ZSize)
}

// Encode writes the raw data array followed by cellCount to w.
// Uses unsafe for zero-copy bulk transfer of the large entry array.
func (z *ZMap) Encode(w io.Writer) error {
	size := int(unsafe.Sizeof(z.data))
	b := unsafe.Slice((*byte)(unsafe.Pointer(&z.data[0])), size)
	if _, err := w.Write(b); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, int64(z.cellCount))
}

// Decode reads the raw data array and cellCount from r.
func (z *ZMap) Decode(r io.Reader) error {
	size := int(unsafe.Sizeof(z.data))
	b := unsafe.Slice((*byte)(unsafe.Pointer(&z.data[0])), size)
	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}
	var cellCount int64
	if err := binary.Read(r, binary.LittleEndian, &cellCount); err != nil {
		return err
	}
	z.cellCount = int(cellCount)
	return nil
}
