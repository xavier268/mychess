package mychess

import "fmt"

// Generate all legal moves from position. A move slice is provided, avoiding allocation as much as possible.
// If it is nil, a new slice will be allocated.
func (pos *Position) LegalMoves(moves []Move) []Move {

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
			sq := Square{i, j}
			switch piece {
			case PAWN, -PAWN:
				moves = pawnMoves(pos, sq, moves)
			case KNIGHT, -KNIGHT:
				moves = knightMoves(pos, sq, moves)
			case BISHOP, -BISHOP:
				moves = bishopMoves(pos, sq, moves)
			case ROOK, -ROOK:
				moves = rookMoves(pos, sq, moves)
			case QUEEN, -QUEEN:
				moves = queenMoves(pos, sq, moves)
			case KING, -KING:
				moves = kingMoves(pos, sq, moves)
			}
		}
	}
	return moves
}

func rookMoves(pos *Position, sq Square, moves []Move) []Move {
	i, j := sq.Row, sq.Col
	piece := pos.Board[i][j]
	var m Move
	var k int
	for k = i + 1; k < 8 && pos.Board[k][j] == EMPTY; k++ {
		m = Move{Piece: piece, From: sq, To: Square{k, j}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if k < 8 && pos.Turn*pos.Board[k][j] < 0 {
		m = Move{Piece: piece, From: sq, To: Square{k, j}}
		moves = append(moves, m)
	}

	for k = i - 1; k >= 0 && pos.Board[k][j] == EMPTY; k-- {
		m = Move{Piece: piece, From: sq, To: Square{k, j}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if k >= 0 && pos.Turn*pos.Board[k][j] < 0 {
		m = Move{Piece: piece, From: sq, To: Square{k, j}}
		moves = append(moves, m)
	}

	for k = j + 1; k < 8 && pos.Board[i][k] == EMPTY; k++ {
		m = Move{Piece: piece, From: sq, To: Square{i, k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if k < 8 && pos.Turn*pos.Board[i][k] < 0 {
		m = Move{Piece: piece, From: sq, To: Square{i, k}}
		moves = append(moves, m)
	}

	for k = j - 1; k >= 0 && pos.Board[i][k] == EMPTY; k-- {
		m = Move{Piece: piece, From: sq, To: Square{i, k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if k >= 0 && pos.Turn*pos.Board[i][k] < 0 {
		m = Move{Piece: piece, From: sq, To: Square{i, k}}
		moves = append(moves, m)
	}
	return moves
}

func bishopMoves(pos *Position, sq Square, moves []Move) []Move {
	i, j := sq.Row, sq.Col
	piece := pos.Board[i][j]
	var m Move
	var k int
	// up right
	for k = 1; i+k < 8 && j+k < 8 && pos.Board[i+k][j+k] == EMPTY; k++ {
		m = Move{Piece: piece, From: sq, To: Square{i + k, j + k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if i+k < 8 && j+k < 8 && pos.Turn*pos.Board[i+k][j+k] < 0 {
		m = Move{Piece: piece, From: sq, To: Square{i + k, j + k}}
		moves = append(moves, m)
	}
	for k = 1; i-k >= 0 && j+k < 8 && pos.Board[i-k][j+k] == EMPTY; k++ {
		m = Move{Piece: piece, From: sq, To: Square{i - k, j + k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if i-k >= 0 && j+k < 8 && pos.Turn*pos.Board[i-k][j+k] < 0 {
		m = Move{Piece: piece, From: sq, To: Square{i - k, j + k}}
		moves = append(moves, m)
	}
	// down right
	for k = 1; i+k < 8 && j-k >= 0 && pos.Board[i+k][j-k] == EMPTY; k++ {
		m = Move{Piece: piece, From: sq, To: Square{i + k, j - k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if i+k < 8 && j-k >= 0 && pos.Turn*pos.Board[i+k][j-k] < 0 {
		m = Move{Piece: piece, From: sq, To: Square{i + k, j - k}}
		moves = append(moves, m)
	}
	// down left
	for k = 1; i-k >= 0 && j-k >= 0 && pos.Board[i-k][j-k] == EMPTY; k++ {
		m = Move{Piece: piece, From: sq, To: Square{i - k, j - k}}
		moves = append(moves, m)
	}
	// check if we can capture ?
	if i-k >= 0 && j-k >= 0 && pos.Turn*pos.Board[i-k][j-k] < 0 {
		m = Move{Piece: piece, From: sq, To: Square{i - k, j - k}}
		moves = append(moves, m)
	}
	return moves
}

func queenMoves(pos *Position, sq Square, moves []Move) []Move {
	mv := rookMoves(pos, sq, moves)
	return bishopMoves(pos, sq, mv)
}

func kingMoves(pos *Position, sq Square, moves []Move) []Move {
	i, j := sq.Row, sq.Col
	piece := pos.Board[i][j]
	var m Move
	// up
	if i+1 < 8 && (pos.Board[i+1][j] == EMPTY || pos.Turn*pos.Board[i+1][j] < 0) {
		m = Move{Piece: piece, From: sq, To: Square{i + 1, j}}
		moves = append(moves, m)
	}
	// up right
	if i+1 < 8 && j+1 < 8 && (pos.Board[i+1][j+1] == EMPTY || pos.Turn*pos.Board[i+1][j+1] < 0) {
		m = Move{Piece: piece, From: sq, To: Square{i + 1, j + 1}}
		moves = append(moves, m)

	}
	// right
	if j+1 < 8 && (pos.Board[i][j+1] == EMPTY || pos.Turn*pos.Board[i][j+1] < 0) {
		m = Move{Piece: piece, From: sq, To: Square{i, j + 1}}
		moves = append(moves, m)

	}
	// down right
	if i-1 >= 0 && j+1 < 8 && (pos.Board[i-1][j+1] == EMPTY || pos.Turn*pos.Board[i-1][j+1] < 0) {
		m = Move{Piece: piece, From: sq, To: Square{i - 1, j + 1}}
		moves = append(moves, m)

	}
	// down
	if i-1 >= 0 && (pos.Board[i-1][j] == EMPTY || pos.Turn*pos.Board[i-1][j] < 0) {
		m = Move{Piece: piece, From: sq, To: Square{i - 1, j}}
		moves = append(moves, m)

	}
	// down left
	if i-1 >= 0 && j-1 >= 0 && (pos.Board[i-1][j-1] == EMPTY || pos.Turn*pos.Board[i-1][j-1] < 0) {
		m = Move{Piece: piece, From: sq, To: Square{i - 1, j - 1}}
		moves = append(moves, m)

	}
	// left
	if j-1 >= 0 && (pos.Board[i][j-1] == EMPTY || pos.Turn*pos.Board[i][j-1] < 0) {
		m = Move{Piece: piece, From: sq, To: Square{i, j - 1}}
		moves = append(moves, m)

	}
	// up left
	if i+1 < 8 && j-1 >= 0 && (pos.Board[i+1][j-1] == EMPTY || pos.Turn*pos.Board[i+1][j-1] < 0) {
		m = Move{Piece: piece, From: sq, To: Square{i + 1, j - 1}}
		moves = append(moves, m)
	}
	// castle
	moves = castleMoves(pos, sq, moves)

	return moves
}

func knightMoves(pos *Position, sq Square, moves []Move) []Move {
	i, j := sq.Row, sq.Col
	di := []int{2, 1, -1, -2, -2, -1, 1, 2}
	dj := []int{1, 2, 2, 1, -1, -2, -2, -1}

	piece := pos.Board[i][j]
	var m Move
	for k := 0; k < 8; k++ {
		ii := i + di[k]
		jj := j + dj[k]
		if ii >= 0 && ii < 8 && jj >= 0 && jj < 8 && (pos.Board[ii][jj] == EMPTY || pos.Turn*pos.Board[ii][jj] < 0) {
			m = Move{Piece: piece, From: sq, To: Square{ii, jj}}
			moves = append(moves, m)
		}
	}
	return moves
}

func pawnMoves(pos *Position, sq Square, moves []Move) []Move {
	if pos.Turn == WHITE {
		return whitePawnMoves(pos, sq, moves)
	} else {
		return blackPawnMoves(pos, sq, moves)
	}
}

func whitePawnMoves(pos *Position, sq Square, moves []Move) []Move {
	i, j := sq.Row, sq.Col
	piece := pos.Board[i][j]
	var m Move

	// move one square up
	if i+1 < 8 && pos.Board[i+1][j] == EMPTY {
		m = Move{Piece: piece, From: sq, To: Square{i + 1, j}}
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	// ove two square up
	if i == 1 && pos.Board[i+1][j] == EMPTY && pos.Board[i+2][j] == EMPTY {
		m = Move{Piece: piece, From: sq, To: Square{i + 2, j}}
		moves = append(moves, m)
	}
	// capture left, including en passant
	if i+1 < 8 && j-1 >= 0 && (pos.Board[i+1][j-1] < 0 || pos.EnPassant == Square{i + 1, j - 1}) {
		m = Move{Piece: piece, From: sq, To: Square{i + 1, j - 1}}
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

func blackPawnMoves(pos *Position, sq Square, moves []Move) []Move {
	i, j := sq.Row, sq.Col
	// move one square down
	piece := pos.Board[i][j]

	var m Move
	if i-1 >= 0 && pos.Board[i-1][j] == EMPTY {
		m = Move{Piece: piece, From: sq, To: Square{i - 1, j}}
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	// ove two square down
	if i == 6 && pos.Board[i-1][j] == EMPTY && pos.Board[i-2][j] == EMPTY {
		m = Move{Piece: piece, From: sq, To: Square{i - 2, j}}
		moves = append(moves, m)
	}
	// capture left, including en passant
	if i-1 >= 0 && j-1 >= 0 && (pos.Board[i-1][j-1] > 0 || pos.EnPassant == Square{i - 1, j - 1}) {
		m = Move{Piece: piece, From: sq, To: Square{i - 1, j - 1}}
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	// capture right, including en passant
	if i-1 >= 0 && j+1 < 8 && (pos.Board[i-1][j+1] > 0 || pos.EnPassant == Square{i - 1, j + 1}) {
		m = Move{Piece: piece, From: sq, To: Square{i - 1, j + 1}}
		moves = append(moves, m)
		moves = promoteLastMove(pos.Turn, moves)
	}
	return moves
}

// verify king starting position and color, and no currently under check
func castleMoves(pos *Position, sq Square, moves []Move) []Move {
	if sq.Col != 4 { // should start from "e" column
		return moves
	}
	if pos.Turn == WHITE && sq.Row == 0 && !pos.isAttacked(Square{0, 4}, BLACK) {
		return whiteCastleMoves(pos, sq, moves)
	}
	if pos.Turn == BLACK && sq.Row == 7 && !pos.isAttacked(Square{7, 4}, WHITE) {
		return blackCastleMoves(pos, sq, moves)
	}
	return moves
}

// king is already in required position
func whiteCastleMoves(pos *Position, sq Square, moves []Move) []Move {
	// check queen side
	if pos.CanWhiteCastleQueenSide && pos.Board[0][1] == EMPTY && pos.Board[0][2] == EMPTY && pos.Board[0][3] == EMPTY {
		if !pos.isAttacked(Square{0, 2}, BLACK) && !pos.isAttacked(Square{0, 3}, BLACK) {
			moves = append(moves, Move{Piece: KING, From: sq, To: Square{0, 2}})
		}
	}
	// check king side
	if pos.CanWhiteCastleKingSide && pos.Board[0][5] == EMPTY && pos.Board[0][6] == EMPTY {
		if !pos.isAttacked(Square{0, 5}, BLACK) && !pos.isAttacked(Square{0, 6}, BLACK) {
			moves = append(moves, Move{Piece: KING, From: sq, To: Square{0, 6}})
		}
	}
	return moves
}

func blackCastleMoves(pos *Position, sq Square, moves []Move) []Move {
	// check queen side
	if pos.CanBlackCastleQueenSide && pos.Board[7][1] == EMPTY && pos.Board[7][2] == EMPTY && pos.Board[7][3] == EMPTY {
		if !pos.isAttacked(Square{7, 2}, WHITE) && !pos.isAttacked(Square{7, 3}, WHITE) {
			moves = append(moves, Move{Piece: -KING, From: sq, To: Square{7, 2}})
		}
	}
	// check king side
	if pos.CanBlackCastleKingSide && pos.Board[7][5] == EMPTY && pos.Board[7][6] == EMPTY {
		if !pos.isAttacked(Square{7, 5}, WHITE) && !pos.isAttacked(Square{7, 6}, WHITE) {
			moves = append(moves, Move{Piece: -KING, From: sq, To: Square{7, 6}})
		}
	}
	return moves
}

// Tell if a position is under attack by the specified color ?
// Position is unchanged. Color may differ from current Turn.
func (p *Position) isAttacked(sq Square, by int8) bool {

	// check verticals (rook & queen)
	for i := sq.Row + 1; i < 8; i++ {
		piece := p.Board[i][sq.Col]
		if piece == EMPTY {
			continue
		}
		if piece*by == ROOK || piece*by == QUEEN {
			return true
		}
	}
	for i := sq.Row - 1; i >= 0; i-- {
		piece := p.Board[i][sq.Col]
		if piece == EMPTY {
			continue
		}
		if piece*by == ROOK || piece*by == QUEEN {
			return true
		}
	}

	// check horizontals (rook & queen)
	for j := sq.Col + 1; j < 8; j++ {
		piece := p.Board[sq.Row][j]
		if piece == EMPTY {
			continue
		}
		if piece*by == ROOK || piece*by == QUEEN {
			return true
		}
	}
	for j := sq.Col - 1; j >= 0; j-- {
		piece := p.Board[sq.Row][j]
		if piece == EMPTY {
			continue
		}
		if piece*by == ROOK || piece*by == QUEEN {
			return true
		}
	}

	// Check diagonals (bishop & queen)
	for i, j := sq.Row+1, sq.Col+1; i < 8 && j < 8; i, j = i+1, j+1 {
		piece := p.Board[i][j]
		if piece == EMPTY {
			continue
		}
		if piece*by == BISHOP || piece*by == QUEEN {
			return true
		}
	}
	for i, j := sq.Row-1, sq.Col+1; i >= 0 && j < 8; i, j = i-1, j+1 {
		piece := p.Board[i][j]
		if piece == EMPTY {
			continue
		}
		if piece*by == BISHOP || piece*by == QUEEN {
			return true
		}
	}
	for i, j := sq.Row+1, sq.Col-1; i < 8 && j >= 0; i, j = i+1, j-1 {
		piece := p.Board[i][j]
		if piece == EMPTY {
			continue
		}
		if piece*by == BISHOP || piece*by == QUEEN {
			return true
		}
	}
	for i, j := sq.Row-1, sq.Col-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		piece := p.Board[i][j]
		if piece == EMPTY {
			continue
		}
		if piece*by == BISHOP || piece*by == QUEEN {
			return true
		}
	}

	// check knights
	di := []int{2, 1, -1, -2, -2, -1, 1, 2}
	dj := []int{1, 2, 2, 1, -1, -2, -2, -1}
	for k := range di {
		ii := sq.Row + di[k]
		jj := sq.Col + dj[k]
		if ii >= 0 && ii < 8 && jj >= 0 && jj < 8 && p.Board[ii][jj] == KNIGHT*by {
			return true
		}
	}

	// check king
	for di := -1; di < 2; di++ {
		for dj := -1; dj < 2; dj++ {
			if di == 0 && dj == 0 {
				continue
			}
			ii := sq.Row + di
			jj := sq.Col + dj
			if ii >= 0 && ii < 8 && jj >= 0 && jj < 8 && p.Board[ii][jj] == KING*by {
				return true
			}
		}
	}

	// white pawn
	if by == WHITE && sq.Row-1 >= 0 && sq.Col-1 >= 0 && p.Board[sq.Row-1][sq.Col-1] == PAWN {
		return true
	}
	if by == WHITE && sq.Row-1 >= 0 && sq.Col+1 >= 0 && p.Board[sq.Row-1][sq.Col+1] == PAWN {
		return true
	}
	// black pawn
	if by == BLACK && sq.Row+1 < 8 && sq.Col-1 >= 0 && p.Board[sq.Row+1][sq.Col-1] == -PAWN {
		return true
	}
	if by == BLACK && sq.Row+1 < 8 && sq.Col+1 < 8 && p.Board[sq.Row+1][sq.Col+1] == -PAWN {
		return true
	}

	return false
}

func (pos *Position) PrintLegalMoves() {
	mm := pos.LegalMoves(nil)
	fmt.Println("Legal moves : ", len(mm))
	for _, m := range mm {
		fmt.Println(m.String())
	}
}
