package main

import (
	"fmt"
	"mychess/eval"
	"mychess/position"
)

func main() {
	PLAYER := position.WHITE
	fmt.Println("Choisissez votre camp : (1 : WHITE, -1 : BLACK)")
	fmt.Scan(&PLAYER)
	if PLAYER != 1 {
		PLAYER = -1
	}
	fmt.Println("Vous jouez", position.StringColor(PLAYER))
	fmt.Println("Préparation ...")

	root := eval.NewNode(position.NewPosition().Reset())
	root.Expand()
	root.Expand()
	root.ExpandBestN(10)

	fmt.Println(root.P.String())

	for {
		root.Expand()
		var mi int
		if root.P.Turn == PLAYER { // human
			fmt.Println("Choisissez votre mouvement :")
			for i, m := range root.Moves {
				fmt.Println(i, m.String())
			}

			for fmt.Scan(&mi); mi < 0 || mi >= len(root.Moves); fmt.Scan(&mi) {
				fmt.Println("Choix invalide. Réssayez ...")
			}

		} else { // ordi
			root.ExpandBestN(6)
			mi, _, _ = root.SelectBestMove()
		}

		fmt.Println("Playing : ", root.Moves[mi].String())

		n2 := root.Play(mi)
		if n2 == nil {
			fmt.Println("Game finished !")
			break
		}

		root = n2
		// runtime.GC()

		fmt.Println(root.P.String())
		if root.P.Turn == PLAYER {
			v, depth := root.Eval()
			fmt.Printf("Value of position : %f/%d\n", v, depth)
		}
	}
}
