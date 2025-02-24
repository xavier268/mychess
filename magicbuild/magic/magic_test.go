package magic

import (
	"fmt"
	"testing"
)

var testMap = map[uint64]uint64{
	2:  9,
	3:  9,
	4:  9,
	5:  17,
	6:  17,
	7:  17,
	8:  10,
	9:  10,
	10: 10,
}

func TestMagic(t *testing.T) {
	magic, nbbits, values := DoMagic(testMap)
	fmt.Println("Values : ", values)
	for k, v := range testMap {
		if vv := ApplyMagic(magic, nbbits, values, k); vv != v {
			t.Errorf("Expected %d, got %d", v, vv)
		}
	}
}
