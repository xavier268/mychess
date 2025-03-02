package position

import "fmt"

// ====================================
// Square object
// ====================================

// Square from 0 to 63
type Square uint8

// Rank/File coordinates of a given square
// 0-based
func (s Square) RF() (rank int, file int) {
	return int(s >> 3), int(s & 7)
}

// Create a square from the rank/file coordinates
func Sq(rank, file int) Square {
	return Square(rank<<3 | file)
}

// create a square from the string expression, eg : "d2"
func SqParse(s string) Square {
	if len(s) != 2 {
		panic("invalid square string")
	}
	file := int(s[0] - 'a')
	rank := int(s[1] - '1')
	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		panic("invalid square")
	}
	return Sq(rank, file)
}

func (s Square) String() string {
	rank, file := s.RF()
	return fmt.Sprintf("%c%d", 'a'+file, rank+1)
}

func (s Square) IsValid() bool {
	return s < 64
}

func (s Square) Rank() int {
	return int(s >> 3)
}

func (s Square) File() int {
	return int(s & 7)
}

func (s Square) VMirror() Square {
	return s ^ 56
}

func (s Square) HMirror() Square {
	return s ^ 7
}

// ========================
// Quadrant manipulation
// =========================

type Quadrant uint8

const (
	QuadMask          = 0b100100
	QA1      Quadrant = 0 & QuadMask
	QH1      Quadrant = 7 & QuadMask
	QA8      Quadrant = 56 & QuadMask
	QH8      Quadrant = 63 & QuadMask
)

// Which quadrant is this square in ?
func (s Square) Quadrant() uint8 {
	return uint8(s) & QuadMask
}

// ================
// RedSquare : Reduced Square, applying symetries to comme in A1 Quandrant (QA1)
// ================

// type RedSquare uint8

// const (
// 	RedSquareMask = 0b011_011
// )

// func (s Square) Reduce() RedSquare {
// 	return RedSquare(s & RedSquareMask)
// }

// // Full square given reduced square and quadrant
// func (s RedSquare) Square(q Quadrant) Square {
// 	return Square(uint8(s) | uint8(q))
// }

// ===============
// SQT : combining a square (possibly reduced) and a table id
// ===============

type SQT uint8

// Constructor. Valid tables ids 0-3
func SquareTable(sq Square, tableId uint8) SQT {
	return SQT(uint8(sq) | (tableId << 6))
}

// // Constructor. Valid table ids are 0-15
// func RedSquareTable(rs RedSquare, tableId uint8) SQT {
// 	return SQT(uint8(rs) | (tableId << 5) | ((tableId & 0b_0000_1000) >> 1))
// }

func (s SQT) Square() Square {
	return Square(s & 0b_0011_1111)
}

// func (s SQT) RedSquare() RedSquare {
// 	return RedSquare(uint8(s) & 0b_0001_1011)
// }

// Table when using full square
// Result is 0 - 3
func (s SQT) Table() uint8 {
	return (uint8(s) >> 6)
}

// Table when using reduced square
// Result is 0-15
// func (s SQT) RedTable() uint8 {
// 	return uint8(s>>5) | (uint8(s&0b_0000_0100) << 1)
// }
