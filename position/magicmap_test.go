package position

import (
	"fmt"
	"testing"
)

var testmap = map[uint64]uint64{
	11:  11111,
	12:  333,
	13:  999,
	987: 0,
	45:  999,
	47:  333,
	49:  333,
}

func TestMagicMapDump(t *testing.T) {

	mm := NewMagicMap_12_6()
	mm.Dump()

	for k, v := range testmap {
		mm.Set(k, v)
		mm.Dump()
	}
	fmt.Println("Verification ...")
	fmt.Println("Keys :", mm.AllKeys())
	fmt.Println("Values :", mm.AllValues())
	for k, v := range testmap {
		if mm.Get(k) != v {
			t.Errorf("Error: %d != %d", mm.Get(k), v)
		}
	}
	if mm.Get(0) != 0 {
		t.Errorf("Error: %d != %d", mm.Get(0), 0)
	}
	if mm.Get(1) != 0 {
		t.Errorf("Error: %d != %d", mm.Get(1), 0)
	}
	nbi, nbo := mm.Count()
	if nbi != 7 {
		t.Errorf("Error: %d != %d", nbi, 7)
	}
	if nbo != 4 {
		t.Errorf("Error: %d != %d", nbo, 4)
	}

}
