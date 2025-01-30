package main

import (
	"fmt"
	"mychess/eval"
	"mychess/position"
)

func main() {
	PLAYER := position.WHITE
	fmt.Println("Vous jouez", position.StringColor(PLAYER))
	fmt.Println("Préparation ...")

	root := eval.NewNode(position.NewPosition().Reset())
	root.Expand()
	root.Expand()
	root.Expand()
	root.ExpandBestN(10)

	fmt.Println(root.P.String())

	for {
		root.Expand()
		root.ExpandBest()
		var mi int
		if root.P.Turn == PLAYER {
			fmt.Println("Choisissez votre mouvement :")
			for i, m := range root.Moves {
				fmt.Println(i, m.String())
			}
			// human
			fmt.Scan(&mi)
		} else {
			// ordi
			mi, _, _ = root.SelectBestMove()
		}

		fmt.Println("Playing : ", root.Moves[mi].String())

		n2 := root.Play(mi)
		if n2 == nil {
			fmt.Println("Game finished !")
			break
		}

		root = n2

		fmt.Println(root.P.String())
		if root.P.Turn == PLAYER {
			v, depth := root.Eval()
			fmt.Printf("Value of position : %f/%d\n", v, depth)
		}
	}
}
