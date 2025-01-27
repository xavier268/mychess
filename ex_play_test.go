package mychess

func ExamplePosition_PrintLegalMoves_t1() {
	p := NewPosition().Reset()
	p.PrintLegalMoves()

	// output:
	// Legal moves :  20
	// White ♘  b1-c3
	// White ♘  b1-a3
	// White ♘  g1-h3
	// White ♘  g1-f3
	// White ♙  a2-a3
	// White ♙  a2-a4
	// White ♙  b2-b3
	// White ♙  b2-b4
	// White ♙  c2-c3
	// White ♙  c2-c4
	// White ♙  d2-d3
	// White ♙  d2-d4
	// White ♙  e2-e3
	// White ♙  e2-e4
	// White ♙  f2-f3
	// White ♙  f2-f4
	// White ♙  g2-g3
	// White ♙  g2-g4
	// White ♙  h2-h3
	// White ♙  h2-h4

}

func ExamplePosition_PrintLegalMoves_t2() {
	p := NewPosition()

	p.SetPiece(KING, "e1")
	p.SetPiece(-KING, "e8")
	p.PrintLegalMoves()

	// output:
	// Legal moves :  5
	// White ♔  e1-e2
	// White ♔  e1-f2
	// White ♔  e1-f1
	// White ♔  e1-d1
	// White ♔  e1-d2
}
