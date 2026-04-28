package position

import (
	"fmt"
	"testing"
)

func TestPSTVisual(t *testing.T) {

	fmt.Println("Displaying visually the weights used for PST")

	fmt.Println("PST for Knights :")

	for rank := 7; rank >= 0; rank-- {
		fmt.Printf("%2d : ", rank+1)
		for file := 0; file < 8; file++ {
			fmt.Printf("%3d ", pstKnight[Sq(rank, file)])
		}
		fmt.Println()
	}

}
