package eval

import (
	"fmt"
	"mychess/position"
	"testing"
)

func TestCount(t *testing.T) {

	var n *Node

	fmt.Printf("HeapSpace used %.2f%%\n", 100*HeapPercentage())

	n = NewNode(position.NewPosition().Reset())
	fmt.Println("\nReset count : ", n.Count())
	fmt.Printf("HeapSpace used %.2f%%\n", 100*HeapPercentage())
	n.Expand0()
	fmt.Println("Expand0", n.Count())
	fmt.Printf("HeapSpace used %.2f%%\n", 100*HeapPercentage())
	n.Expand0()
	fmt.Println("Expand0", n.Count())
	fmt.Printf("HeapSpace used %.2f%%\n", 100*HeapPercentage())

	n = NewNode(position.NewPosition().Reset())
	fmt.Println("\nReset count : ", n.Count())
	fmt.Printf("HeapSpace used %.2f%%\n", 100*HeapPercentage())
	n.Expand()
	fmt.Println("Expand", n.Count())
	fmt.Printf("HeapSpace used %.2f%%\n", 100*HeapPercentage())
	n.Expand0()
	fmt.Println("Expand0", n.Count())
	fmt.Printf("HeapSpace used %.2f%%\n", 100*HeapPercentage())
	n.Expand()
	fmt.Println("Expand", n.Count())
	fmt.Printf("HeapSpace used %.2f%%\n", 100*HeapPercentage())
	n.ExpandBestN(3)
	fmt.Println("ExpandBest x 3", n.Count())
	fmt.Printf("HeapSpace used %.2f%%\n", 100*HeapPercentage())

}

func TestLimit1(t *testing.T) {

	n := NewNode(position.NewPosition().Reset())
	err := n.ExpandBFSLimit(NewDefaultLimit())

	fmt.Printf("%s\nExpandBFS with limit count = %d\n", err, n.Count())
}

func TestLimit2(t *testing.T) {

	n := NewNode(position.NewPosition().Reset())
	n.Expand0()

	err := n.ExpandBestLimit(NewDefaultLimit())

	fmt.Printf("%s\nExpandBest with limit count = %d\n", err, n.Count())
}
