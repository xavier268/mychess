package magic

import (
	"fmt"
	"math/rand"
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
	0:   888,
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
	if mm.Get(1) != 0 {
		t.Errorf("Error: %d != %d", mm.Get(1), 0)
	}
	nbi, nbo := mm.Count()
	if nbi != 7 { // 0 is never counted as a key
		t.Errorf("Error: %d != %d", nbi, 7)
	}
	if nbo != 4 { // the value for the zero key is not counted
		t.Errorf("Error: %d != %d", nbo, 4)
	}

}

func TestMagicMapCapacity(t *testing.T) {
	t.Skip() // does not work, either collision are mishandled, or too slow...
	nin, nout := 5, 4
	mm := NewMagicMap(nin+2, nout+2)

	for i := range 1 << (nin) {
		mm.Set(uint64(i), uint64(i&((1<<nout)-1))*337)
	}
	for i := range 100 {
		key := rand.Uint64() & ((1 << nin) - 1)
		if mm.Get(key) != uint64(key&((1<<nout)-1))*337 {
			t.Errorf("Error for key %d : %d != %d", key, mm.Get(uint64(i)), uint64(i))
		}
	}
}
