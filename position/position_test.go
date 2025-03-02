package position

import (
	"fmt"
	"testing"
)

func TestDisplayPosition(t *testing.T) {
	rev := StartPosition
	fmt.Println(rev.String())
	//StartPosition.Dump()
	fmt.Println("Black K", rev.GetBlackKingSquare())
	fmt.Println("White K", rev.GetWhiteKingSquare())

	fmt.Println("Reversing start position")
	rev = rev.VMirror()
	fmt.Println("Black K", rev.GetBlackKingSquare())
	fmt.Println("White K", rev.GetWhiteKingSquare())
	fmt.Println(rev.String())
}
