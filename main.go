package main

import (
	"fmt"
	"mychess/eval"
	"mychess/position"
	"runtime"
)

const VERSION = "0.1.1"

var HINT = true    // should we display hint ?
var VERBOSE = true // should we display statistics ?

func main() {
	fmt.Println("Mychess Version ", VERSION)
	PLAYER := position.WHITE
	fmt.Println("Choisissez votre camp : (1 : WHITE, -1 : BLACK)")
	fmt.Scan(&PLAYER)
	fmt.Println("HeapValue", eval.HeapValue()/1000000, "Mo")
	if PLAYER != 1 {
		PLAYER = -1
	}
	fmt.Println("Vous jouez", position.StringColor(PLAYER))
	fmt.Println("Préparation ...")

	root := eval.NewNode(position.NewPosition().Reset())
	root.Expand()
	root.Expand()
	fmt.Println("HeapValue", eval.HeapValue()/1000000, "Mo")
	fmt.Println(root.ExpandBFSLimit(eval.NewDefaultLimit()))
	fmt.Println("HeapValue", eval.HeapValue()/1000000, "Mo")

	fmt.Println(root.P.String())

	for {
		root.Expand0() // required to compensate depth loss from node selection
		var mi, md int
		var mv float64
		if root.P.Turn == PLAYER { // human
			if len(root.Moves) == 0 {
				fmt.Println("Game over !")
				if root.P.IsCheck(PLAYER) {
					fmt.Println("Checkmate - You lost !")
				} else {
					fmt.Println("Pat !")
				}
				break
			}
			if HINT {
				mi, mv, md = root.SelectBestMove()
			}
			fmt.Println("Choisissez votre mouvement :")
			for i, m := range root.Moves {
				fmt.Printf("%2d  %s ", i, m.String())
				if HINT && i == mi {
					fmt.Printf("<=== (Recommanded : value=%f, depth=%d)\n", mv, md)
				} else {
					fmt.Println()
				}
			}
			for fmt.Scan(&mi); mi < 0 || mi >= len(root.Moves); fmt.Scan(&mi) {
				fmt.Println("Choix invalide. Réssayez ...")
			}

		} else { // ordi
			runtime.GC()
			fmt.Println(root.ExpandBFSLimit(eval.NewDefaultLimit()))
			fmt.Println("HeapValue", eval.HeapValue()/1000000, "Mo")
			runtime.GC()
			fmt.Println(root.ExpandBestLimit(eval.NewDefaultLimit()))
			fmt.Println("HeapValue", eval.HeapValue()/1000000, "Mo")
			mi, _, _ = root.SelectBestMove()
		}

		fmt.Println("Playing : ", root.Moves[mi].String())

		n2 := root.Play(mi)
		if n2 == nil {
			fmt.Println("Game finished !")
			if root.P.IsCheck(-PLAYER) {
				fmt.Println("Checkmate - you won !")
			} else {
				fmt.Println("Pat !")
			}
			break
		}

		root = n2
		// runtime.GC()

		fmt.Println(root.P.String())
		if root.P.Turn == PLAYER {
			v, depth := root.Eval()
			fmt.Printf("Value of position : %f (depth :%d, evaluations : %d)\n", v, depth, root.Count())
		}
	}
}
