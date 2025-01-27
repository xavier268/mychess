package mychess

import "fmt"

func ExamplePosition_PrintLegalMoves_t1() {
	p := NewPosition().Reset()
	p.PrintLegalMoves()

	// output:
	// Legal moves :  20
	// White N  b1-c3
	// White N  b1-a3
	// White N  g1-h3
	// White N  g1-f3
	// White P  a2-a3
	// White P  a2-a4
	// White P  b2-b3
	// White P  b2-b4
	// White P  c2-c3
	// White P  c2-c4
	// White P  d2-d3
	// White P  d2-d4
	// White P  e2-e3
	// White P  e2-e4
	// White P  f2-f3
	// White P  f2-f4
	// White P  g2-g3
	// White P  g2-g4
	// White P  h2-h3
	// White P  h2-h4
}

func ExamplePosition_PrintLegalMoves_t2() {
	p := NewPosition()

	p.SetPiece(KING, "e1")
	p.SetPiece(-KING, "e8")
	p.PrintLegalMoves()

	// output:
	// Legal moves :  5
	// White K  e1-e2
	// White K  e1-f2
	// White K  e1-f1
	// White K  e1-d1
	// White K  e1-d2
}

func ExamplePosition_PrintLegalMoves_enpassant() {
	p := NewPosition()

	p.SetPiece(KING, "e1")
	p.SetPiece(-KING, "e8")

	p.SetPiece(PAWN, "e2")        // white pawn
	p.SetPiece(-PAWN, "d4", "e7") // black pawn

	p.PrintLegalMoves()

	p.ExecuteMove(Move{PAWN, SquareFromString("e2"), SquareFromString("e4")})
	fmt.Println()
	p.PrintLegalMoves()

	// output:
	// Legal moves :  6
	// White K  e1-f2
	// White K  e1-f1
	// White K  e1-d1
	// White K  e1-d2
	// White P  e2-e3
	// White P  e2-e4
	//
	// Legal moves :  8
	// Black p  d4-d3
	// Black p  d4-e3
	// Black p  e7-e6
	// Black p  e7-e5
	// Black k  e8-f8
	// Black k  e8-f7
	// Black k  e8-d7
	// Black k  e8-d8

}

func ExamplePosition_PrintLegalMoves_castling() {

	p := NewPosition()

	p.SetPiece(KING, "e1")
	p.SetPiece(-KING, "e8")
	p.SetPiece(ROOK, "a1")
	p.SetPiece(-ROOK, "a8")
	p.SetPiece(ROOK, "h1")
	p.SetPiece(-ROOK, "h8")

	p.PrintLegalMoves()

	p.ExecuteMove(Move{KING, SquareFromString("e1"), SquareFromString("c1")})
	fmt.Println()
	p.PrintLegalMoves()

	// output:
	// Legal moves :  26
	// White R  a1-a2
	// White R  a1-a3
	// White R  a1-a4
	// White R  a1-a5
	// White R  a1-a6
	// White R  a1-a7
	// White R  a1-a8
	// White R  a1-b1
	// White R  a1-c1
	// White R  a1-d1
	// White K  e1-e2
	// White K  e1-f2
	// White K  e1-f1
	// White K  e1-d1
	// White K  e1-d2
	// White K  e1-c1
	// White K  e1-g1
	// White R  h1-h2
	// White R  h1-h3
	// White R  h1-h4
	// White R  h1-h5
	// White R  h1-h6
	// White R  h1-h7
	// White R  h1-h8
	// White R  h1-g1
	// White R  h1-f1
	//
	// Legal moves :  23
	// Black r  a8-a7
	// Black r  a8-a6
	// Black r  a8-a5
	// Black r  a8-a4
	// Black r  a8-a3
	// Black r  a8-a2
	// Black r  a8-a1
	// Black r  a8-b8
	// Black r  a8-c8
	// Black r  a8-d8
	// Black k  e8-f8
	// Black k  e8-f7
	// Black k  e8-e7
	// Black k  e8-g8
	// Black r  h8-h7
	// Black r  h8-h6
	// Black r  h8-h5
	// Black r  h8-h4
	// Black r  h8-h3
	// Black r  h8-h2
	// Black r  h8-h1
	// Black r  h8-g8
	// Black r  h8-f8

}

func ExamplePosition_PrintLegalMoves_mate() {
	p := NewPosition()

	p.SetPiece(KING, "e1")
	p.SetPiece(-KING, "e8")

	p.SetPiece(-ROOK, "a1")
	p.SetPiece(-ROOK, "a2")

	p.SetPiece(PAWN, "e3")
	p.SetPiece(ROOK, "h8")

	p.PrintLegalMoves()

	// output:
	// Legal moves :  0
}
