package eval

import (
	"fmt"
	"mychess/position"
	"testing"
)

func TestExpandEval(t *testing.T) {

	root := NewNode(position.NewPosition().Reset())

	v, d := root.Eval()
	fmt.Println("Eval ", v, d)
	root.Expand()
	v, d = root.Eval()
	fmt.Println("Eval ", v, d)
	root.Expand()
	v, d = root.Eval()
	fmt.Println("Eval ", v, d)
	root.Expand()
	v, d = root.Eval()
	fmt.Println("Eval ", v, d)
	root.Expand()
	v, d = root.Eval()
	fmt.Println("Eval ", v, d)
	root.Expand()
	v, d = root.Eval()
	fmt.Println("Eval ", v, d)

	fmt.Println("Finished !")
}

func TestAutoPlay(t *testing.T) {
	root := NewNode(position.NewPosition().Reset())

	// Systematic 2 level exploration
	root.Expand()
	root.Expand()
	root.Expand()
	root.ExpandBestN(10)

	// some partial random exploration

	for i := 0; i < 10; i++ {
		fmt.Println(root.P.String())
		root.Expand()
		root.ExpandBest()
		mi, v, d := root.SelectBestMove()
		if mi < 0 {
			fmt.Printf("%d)  %s    (val: %f/levels : %d)\n", i+1, "no more move", v, d)
		} else {
			m := root.Moves[mi]
			fmt.Printf("%d)  %s    (val: %f/levels : %d)\n", i+1, m.String(), v, d)
		}
		n2 := root.Play(mi)
		if n2 == nil {
			fmt.Println("Game finished !")
			break
		}
		root = n2
	}
}
