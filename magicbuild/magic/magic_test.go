package magic

import (
	"fmt"
	"testing"
)

var testMap = map[uint64]uint64{
	// 0: 170,
	1:  90,
	2:  170,
	3:  90,
	4:  90,
	5:  170,
	6:  170,
	7:  170,
	8:  100,
	9:  100,
	10: 100,
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
