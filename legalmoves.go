package mychess

// Generate all legal moves from position. A move slice is provided, avoiding allocation as much as possible.
// If it is nil, a new slice will be allocated.
func (pos *Position) LegalMoves(moves []Move) []Move {
	// if pos.Draw || pos.StaleMate {
	// 	return nil
	// }
	if moves == nil {
		moves = make([]Move, 0, 40)
	} else {
		moves = moves[:0]
	}

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			piece := pos.Board[i][j]
			if pos.Turn*piece <= 0 {
				continue // skip if piece is empty or not your own !
			}
			switch piece {
			case PAWN, -PAWN:
				moves = pawnMoves(pos, i, j, moves)
			case KNIGHT, -KNIGHT:
				moves = knightMoves(pos, i, j, moves)
			case BISHOP, -BISHOP:
				moves = bishopMoves(pos, i, j, moves)
			case ROOK, -ROOK:
				moves = rookMoves(pos, i, j, moves)
			case QUEEN, -QUEEN:
				moves = queenMoves(pos, i, j, moves)
			case KING, -KING:
				moves = kingMoves(pos, i, j, moves)
			}
		}
	}
	return moves
}

func rookMoves(pos *Position, i, j int, moves []Move) []Move {
	piece := pos.Board[i][j]
	from := Square{i, j}
	var m Move
	var k int
	for k = i + 1; k < 8 && pos.Board[k][j] == EMPTY; k++ {
		m = Move{Piece: piece, From: from, To: Square{k, j}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if k < 8 && pos.Turn*pos.Board[k][j] < 0 {
		m = Move{Piece: piece, From: from, To: Square{k, j}}
		moves = append(moves, m)
	}

	for k = i - 1; k >= 0 && pos.Board[k][j] == EMPTY; k-- {
		m = Move{Piece: piece, From: from, To: Square{k, j}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if k >= 0 && pos.Turn*pos.Board[k][j] < 0 {
		m = Move{Piece: piece, From: from, To: Square{k, j}}
		moves = append(moves, m)
	}

	for k = j + 1; k < 8 && pos.Board[i][k] == EMPTY; k++ {
		m = Move{Piece: piece, From: from, To: Square{i, k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if k < 8 && pos.Turn*pos.Board[i][k] < 0 {
		m = Move{Piece: piece, From: from, To: Square{i, k}}
		moves = append(moves, m)
	}

	for k = j - 1; k >= 0 && pos.Board[i][k] == EMPTY; k-- {
		m = Move{Piece: piece, From: from, To: Square{i, k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if k >= 0 && pos.Turn*pos.Board[i][k] < 0 {
		m = Move{Piece: piece, From: from, To: Square{i, k}}
		moves = append(moves, m)
	}
	return moves
}

func bishopMoves(pos *Position, i, j int, moves []Move) []Move {
	piece := pos.Board[i][j]
	from := Square{i, j}
	var m Move
	var k int
	// up right
	for k = 1; i+k < 8 && j+k < 8 && pos.Board[i+k][j+k] == EMPTY; k++ {
		m = Move{Piece: piece, From: from, To: Square{i + k, j + k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if i+k < 8 && j+k < 8 && pos.Turn*pos.Board[i+k][j+k] < 0 {
		m = Move{Piece: piece, From: from, To: Square{i + k, j + k}}
		moves = append(moves, m)
	}
	for k = 1; i-k >= 0 && j+k < 8 && pos.Board[i-k][j+k] == EMPTY; k++ {
		m = Move{Piece: piece, From: from, To: Square{i - k, j + k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if i-k >= 0 && j+k < 8 && pos.Turn*pos.Board[i-k][j+k] < 0 {
		m = Move{Piece: piece, From: from, To: Square{i - k, j + k}}
		moves = append(moves, m)
	}
	// down right
	for k = 1; i+k < 8 && j-k >= 0 && pos.Board[i+k][j-k] == EMPTY; k++ {
		m = Move{Piece: piece, From: from, To: Square{i + k, j - k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if i+k < 8 && j-k >= 0 && pos.Turn*pos.Board[i+k][j-k] < 0 {
		m = Move{Piece: piece, From: from, To: Square{i + k, j - k}}
		moves = append(moves, m)
	}
	// down left
	for k = 1; i-k >= 0 && j-k >= 0 && pos.Board[i-k][j-k] == EMPTY; k++ {
		m = Move{Piece: piece, From: from, To: Square{i - k, j - k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if i-k >= 0 && j-k >= 0 && pos.Turn*pos.Board[i-k][j-k] < 0 {
		m = Move{Piece: piece, From: from, To: Square{i - k, j - k}}
		moves = append(moves, m)
	}
	return moves
}

func queenMoves(pos *Position, i, j int, moves []Move) []Move {
	mv := rookMoves(pos, i, j, moves)
	return bishopMoves(pos, i, j, mv)
}

func kingMoves(pos *Position, i, j int, moves []Move) []Move {

	piece := pos.Board[i][j]
	from := Square{i, j}
	var m Move
	// up
	if i+1 < 8 && (pos.Board[i+1][j] == EMPTY || pos.Turn*pos.Board[i+1][j] < 0) {
		m = Move{Piece: piece, From: from, To: Square{i + 1, j}}
		moves = append(moves, m)
	}
	// up right
	if i+1 < 8 && j+1 < 8 && (pos.Board[i+1][j+1] == EMPTY || pos.Turn*pos.Board[i+1][j+1] < 0) {
		m = Move{Piece: piece, From: from, To: Square{i + 1, j + 1}}
		moves = append(moves, m)

	}
	// right
	if j+1 < 8 && (pos.Board[i][j+1] == EMPTY || pos.Turn*pos.Board[i][j+1] < 0) {
		m = Move{Piece: piece, From: from, To: Square{i, j + 1}}
		moves = append(moves, m)

	}
	// down right
	if i-1 >= 0 && j+1 < 8 && (pos.Board[i-1][j+1] == EMPTY || pos.Turn*pos.Board[i-1][j+1] < 0) {
		m = Move{Piece: piece, From: from, To: Square{i - 1, j + 1}}
		moves = append(moves, m)

	}
	// down
	if i-1 >= 0 && (pos.Board[i-1][j] == EMPTY || pos.Turn*pos.Board[i-1][j] < 0) {
		m = Move{Piece: piece, From: from, To: Square{i - 1, j}}
		moves = append(moves, m)

	}
	// down left
	if i-1 >= 0 && j-1 >= 0 && (pos.Board[i-1][j-1] == EMPTY || pos.Turn*pos.Board[i-1][j-1] < 0) {
		m = Move{Piece: piece, From: from, To: Square{i - 1, j - 1}}
		moves = append(moves, m)

	}
	// left
	if j-1 >= 0 && (pos.Board[i][j-1] == EMPTY || pos.Turn*pos.Board[i][j-1] < 0) {
		m = Move{Piece: piece, From: from, To: Square{i, j - 1}}
		moves = append(moves, m)

	}
	// up left
	if i+1 < 8 && j-1 >= 0 && (pos.Board[i+1][j-1] == EMPTY || pos.Turn*pos.Board[i+1][j-1] < 0) {
		m = Move{Piece: piece, From: from, To: Square{i + 1, j - 1}}
		moves = append(moves, m)
	}
	return moves
}

func knightMoves(pos *Position, i, j int, moves []Move) []Move {
	di := []int{2, 1, -1, -2, -2, -1, 1, 2}
	dj := []int{1, 2, 2, 1, -1, -2, -2, -1}

	piece := pos.Board[i][j]
	from := Square{i, j}
	var m Move
	for k := 0; k < 8; k++ {
		ii := i + di[k]
		jj := j + dj[k]
		if ii >= 0 && ii < 8 && jj >= 0 && jj < 8 && (pos.Board[ii][jj] == EMPTY || pos.Turn*pos.Board[ii][jj] < 0) {
			m = Move{Piece: piece, From: from, To: Square{ii, jj}}
			moves = append(moves, m)
		}
	}
	return moves
}

func pawnMoves(pos *Position, i, j int, moves []Move) []Move {
	if pos.Turn == WHITE {
		return whitePawnMoves(pos, i, j, moves)
	} else {
		return blackPawnMoves(pos, i, j, moves)
	}
}

func whitePawnMoves(pos *Position, i, j int, moves []Move) []Move {
	piece := pos.Board[i][j]
	from := Square{i, j}
	var m Move

	// move one square up
	if i+1 < 8 && pos.Board[i+1][j] == EMPTY {
		m = Move{Piece: piece, From: from, To: Square{i + 1, j}}
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	// ove two square up
	if i == 1 && pos.Board[i+1][j] == EMPTY && pos.Board[i+2][j] == EMPTY {
		m = Move{Piece: piece, From: from, To: Square{i + 2, j}}
		moves = append(moves, m)
	}
	// capture left, including en passant
	if i+1 < 8 && j-1 >= 0 && (pos.Board[i+1][j-1] < 0 || pos.EnPassant == Square{i + 1, j - 1}) {
		m = Move{Piece: piece, From: from, To: Square{i + 1, j - 1}}
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	// capture right, including en passant
	if i+1 < 8 && j+1 < 8 && (pos.Board[i+1][j+1] < 0 || pos.EnPassant == Square{i + 1, j + 1}) {
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	return moves
}

// add all possible promotions, without changing to and from position of the last move.
func promoteLastMove(turn int8, moves []Move) []Move {
	if len(moves) == 0 {
		return moves
	}
	last := moves[len(moves)-1]
	if turn == WHITE && last.To.Row == 7 {
		last.Piece = QUEEN
		moves = append(moves, last)
		last.Piece = KNIGHT
		moves = append(moves, last)
		last.Piece = ROOK
		moves = append(moves, last)
		last.Piece = BISHOP
		moves = append(moves, last)
	}

	if turn == BLACK && last.To.Row == 0 {
		last.Piece = -QUEEN
		moves = append(moves, last)
		last.Piece = -KNIGHT
		moves = append(moves, last)
		last.Piece = -ROOK
		moves = append(moves, last)
		last.Piece = -BISHOP
		moves = append(moves, last)
	}
	return moves
}

func blackPawnMoves(pos *Position, i, j int, moves []Move) []Move {
	// move one square down
	piece := pos.Board[i][j]
	from := Square{i, j}
	var m Move
	if i-1 >= 0 && pos.Board[i-1][j] == EMPTY {
		m = Move{Piece: piece, From: from, To: Square{i - 1, j}}
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	// ove two square down
	if i == 6 && pos.Board[i-1][j] == EMPTY && pos.Board[i-2][j] == EMPTY {
		m = Move{Piece: piece, From: from, To: Square{i - 2, j}}
		moves = append(moves, m)
	}
	// capture left, including en passant
	if i-1 >= 0 && j-1 >= 0 && (pos.Board[i-1][j-1] > 0 || pos.EnPassant == Square{i - 1, j - 1}) {
		m = Move{Piece: piece, From: from, To: Square{i - 1, j - 1}}
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	// capture right, including en passant
	if i-1 >= 0 && j+1 < 8 && (pos.Board[i-1][j+1] > 0 || pos.EnPassant == Square{i - 1, j + 1}) {
		m = Move{Piece: piece, From: from, To: Square{i - 1, j + 1}}
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	return moves
}

// Check if provided square is exposed by "who" player. Who is white or black.
// CHECK LOGIC AND TESTING NEEDED !!
/*
func (pos *Position) isExposedBy(sq Square, who int8) bool {
	 	// look up
	   	for i := sq.Row + 1; i < 8; i++ {
	   		piece := pos.Board[i][sq.Col]
	   		if piece == EMPTY {
	   			continue
	   		}
	   		if piece*who == ROOK || piece*who == QUEEN {
	   			return true
	   		}
	   		break
	   	}
	   	// look down
	   	for i := sq.Row - 1; i >= 0; i-- {
	   		piece := pos.Board[i][sq.Col]
	   		if piece == EMPTY {
	   			continue
	   		}
	   		if piece*who == ROOK || piece*who == QUEEN {
	   			return true
	   		}
	   		break
	   	}
	   	// look right
	   	for j := sq.Col + 1; j < 8; j++ {
	   		piece := pos.Board[sq.Row][j]
	   		if piece == EMPTY {
	   			continue
	   		}
	   		if piece*who == ROOK || piece*who == QUEEN {
	   			return true
	   		}
	   		break
	   	}
	   	// look left
	   	for j := sq.Col - 1; j >= 0; j-- {
	   		piece := pos.Board[sq.Row][j]
	   		if piece == EMPTY {
	   			continue
	   		}
	   		if piece*who == ROOK || piece*who == QUEEN {
	   			return true
	   		}
	   		break
	   	}
	   	// look up right
	   	for i, j := sq.Row+1, sq.Col+1; i < 8 && j < 8; i, j = i+1, j+1 {
	   		piece := pos.Board[i][j]
	   		if piece == EMPTY {
	   			continue
	   		}
	   		if piece*who == BISHOP || piece*who == QUEEN {
	   			return true
	   		}
	   		break
	   	}
	   	// look down right
	   	for i, j := sq.Row-1, sq.Col+1; i >= 0 && j < 8; i, j = i-1, j+1 {
	   		piece := pos.Board[i][j]
	   		if piece == EMPTY {
	   			continue
	   		}
	   		if piece*who == BISHOP || piece*who == QUEEN {
	   			return true
	   		}
	   		break
	   	}
	   	// look down left
	   	for i, j := sq.Row-1, sq.Col-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
	   		piece := pos.Board[i][j]
	   		if piece == EMPTY {
	   			continue
	   		}
	   		if piece*who == BISHOP || piece*who == QUEEN {
	   			return true
	   		}
	   		break
	   	}
	   	// look up left
	   	for i, j := sq.Row+1, sq.Col-1; i < 8 && j >= 0; i, j = i+1, j-1 {
	   		piece := pos.Board[i][j]
	   		if piece == EMPTY {
	   			continue
	   		}
	   		if piece*who == BISHOP || piece*who == QUEEN {
	   			return true
	   		}
	   		break
	   	}
	   	// look up knight
	   	if sq.Row+2 < 8 && sq.Col+1 < 8 && pos.Board[sq.Row+2][sq.Col+1] == who*KNIGHT {
	   		return true
	   	}
	   	if sq.Row+1 < 8 && sq.Col+2 < 8 && pos.Board[sq.Row+1][sq.Col+2] == who*KNIGHT {
	   		return true
	   	}
	   	if sq.Row-1 >= 0 && sq.Col+2 < 8 && pos.Board[sq.Row-1][sq.Col+2] == who*KNIGHT {
	   		return true
	   	}
	   	if sq.Row-2 >= 0 && sq.Col+1 < 8 && pos.Board[sq.Row-2][sq.Col+1] == who*KNIGHT {
	   		return true
	   	}
	   	if sq.Row-2 >= 0 && sq.Col-1 >= 0 && pos.Board[sq.Row-2][sq.Col-1] == who*KNIGHT {
	   		return true
	   	}
	   	if sq.Row-1 >= 0 && sq.Col-2 >= 0 && pos.Board[sq.Row-1][sq.Col-2] == who*KNIGHT {
	   		return true
	   	}
	   	if sq.Row+1 < 8 && sq.Col-2 >= 0 && pos.Board[sq.Row+1][sq.Col-2] == who*KNIGHT {
	   		return true
	   	}
	   	if sq.Row+2 < 8 && sq.Col-1 >= 0 && pos.Board[sq.Row+2][sq.Col-1] == who*KNIGHT {
	   		return true
	   	}

	   	// Look for king
	   	if sq.Row+1 < 8 && pos.Board[sq.Row+1][sq.Col] == who*KING {
	   		return true
	   	}
	   	if sq.Row-1 >= 0 && pos.Board[sq.Row-1][sq.Col] == who*KING {
	   		return true
	   	}
	   	if sq.Col+1 < 8 && pos.Board[sq.Row][sq.Col+1] == who*KING {
	   		return true
	   	}
	   	if sq.Col-1 >= 0 && pos.Board[sq.Row][sq.Col-1] == who*KING {
	   		return true
	   	}
	   	if sq.Row+1 < 8 && sq.Col+1 < 8 && pos.Board[sq.Row+1][sq.Col+1] == who*KING {
	   		return true
	   	}
	   	if sq.Row-1 >= 0 && sq.Col+1 < 8 && pos.Board[sq.Row-1][sq.Col+1] == who*KING {
	   		return true
	   	}
	   	if sq.Row-1 >= 0 && sq.Col-1 >= 0 && pos.Board[sq.Row-1][sq.Col-1] == who*KING {
	   		return true
	   	}
	   	if sq.Row+1 < 8 && sq.Col-1 >= 0 && pos.Board[sq.Row+1][sq.Col-1] == who*KING {
	   		return true
	   	}

	   	// Look for pawn
	   	if who == WHITE && sq.Row-1 >= 0 && sq.Col-1 >= 0 && pos.Board[sq.Row-1][sq.Col-1] == who*PAWN {
	   		return true
	   	}
	   	if who == WHITE && sq.Row-1 >= 0 && sq.Col+1 < 8 && pos.Board[sq.Row-1][sq.Col+1] == who*PAWN {
	   		return true
	   	}
	   	if who == BLACK && sq.Row+1 < 8 && sq.Col-1 >= 0 && pos.Board[sq.Row+1][sq.Col-1] == who*PAWN {
	   		return true
	   	}
	   	if who == BLACK && sq.Row+1 < 8 && sq.Col+1 < 8 && pos.Board[sq.Row+1][sq.Col+1] == who*PAWN {
	   		return true
	   	}

	return false
}
*/
