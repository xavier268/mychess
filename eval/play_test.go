package eval

import (
	"fmt"
	"mychess/position"
	"testing"
)

func TestGrowth(t *testing.T) {

	root := NewNode(position.NewPosition().Reset())

	root.Grow(3)
	v, d := root.Eval()
	fmt.Println("Eval ", v, d)
	root.Grow(3)
	v, d = root.Eval()
	fmt.Println("Eval ", v, d)

	fmt.Println("Finished !")
}
