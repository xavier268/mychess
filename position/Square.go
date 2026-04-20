package position

import (
	"fmt"
	"strings"
)

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
	s = strings.ToLower(s)
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

// no bound checks !
func (sq Square) North() Square {
	return (sq + 8) & 63
}

// no bound checks !
func (sq Square) South() Square {
	return (sq - 8) & 63
}

// no bound checks !
func (sq Square) East() Square {
	return (sq + 1) & 63
}

// no bound checks !
func (sq Square) West() Square {
	return (sq - 1) & 63
}

func (sq Square) Bitboard() Bitboard {
	return 1 << sq
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
