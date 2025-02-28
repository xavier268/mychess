package magic2

import (
	"fmt"
	"testing"
	"unsafe"
)

func TestPadding(t *testing.T) {
	type st struct {
		A uint64 // 8 bytes
		B uint8  // 1 byte
	}
	var arr [10]st     // => padded to 16 bytes x 10 !!!
	var arr2 [10]uint8 // => packed bytes array x 10 !!!

	fmt.Printf("[10]struct{uint64,uint8} size in memory : %d bytes = 10 x %d bytes\n", unsafe.Sizeof(arr), unsafe.Sizeof(st{}))
	fmt.Printf("[10]uint8 size in memory : %d bytes = 10 x %d bytes\n", unsafe.Sizeof(arr2), unsafe.Sizeof(uint8(0)))
}

func TestNextPowerOfTwo(t *testing.T) {
	data := []uint64{
		0, 0,
		1, 1,
		2, 2,
		3, 4,
		4, 4,
		5, 8,
		6, 8,
		7, 8,
		8, 8,
		9, 16,
		10, 16,
		11, 16,
		12, 16,
		13, 16,
		14, 16,
		15, 16,
		16, 16,
		17, 32,
		18, 32,
		19, 32,
		20, 32,
		21, 32,
		22, 32,
		23, 32,
		24, 32,
		25, 32,
		26, 32,
		27, 32,
		28, 32,
		29, 32,
		30, 32,
		31, 32,
		32, 32,
		50, 64,
		64, 64,
		100, 128,
		128, 128,
		200, 256,
		256, 256,
		300, 512,
		512, 512,
		1000, 1024,
		1024, 1024,
		10000, 16384,
		16384, 16384,
		16385, 32768,
		32768, 32768,
		60000, 65536,
		65536, 65536,
		100_000, 131_072,
		131_072, 131_072,
		200_000, 262_144,
		16777216, 16777216,
		536870912, 536870912,
		1<<50 - 100, 1 << 50,
		1 << 50, 1 << 50,
		1 << 60, 1 << 60,
		1 << 62, 1 << 62,
		1<<63 - 1, 1 << 63,
		1<<63 - 100, 1 << 63,
		1 << 63, 1 << 63,
	}

	for i := 0; i < len(data); i += 2 {
		if NextPowerOfTwo(data[i]) != data[i+1] {
			t.Errorf("nextPowerOfTwo(%d) = %d, want %d", data[i], NextPowerOfTwo(data[i]), data[i+1])
		}
	}

}

func TestCreateEmptyMagic(t *testing.T) {
	_, st := CreateMagic2()
	fmt.Print(st.String())
}
