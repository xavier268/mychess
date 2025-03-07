package position

import (
	"fmt"
	"strings"
)

// ============== Position definition ==============================

// Constant to indicate turn (who is expected to play ) and index the masks and status.
const (
	WHITE = 0
	BLACK = 1
)

type Position struct {
	// // Occupancies for each side, all actual pieces (king INCLUDED), but en passant marker not included
	colOcc [2]Bitboard
	// all colors together
	// en passant is indicated as rank 0 for the white pawn (there can never be a black pawn here because of promotion), WHITHOUT any occupancy flag.
	pawnOcc Bitboard
	// all colors together
	rookOcc Bitboard
	// all colors together
	bishopOcc Bitboard
	// all colors together
	knightOcc Bitboard

	// queen appears as both a rook and a bishop
	// king positions derived from status data below
	status Status
}

// Special type to handle status word.
type Status struct {
	// Bit 0-5 : king square position
	// Bit 6-7 : castle bits
	KingStatus [2]uint8
	// Bit 0 : 	Who should play WHITE/BLACK
	// Bit 1-7 : RUF
	TurnStatus uint8
}

const (
	// Castle bits
	CanCastleKingSide  = 0b10000000
	CanCastleQueenSide = 0b01000000
	CanCastle          = CanCastleQueenSide | CanCastleKingSide
)

const (
	StartWhiteOcc  Bitboard = 0xFFFF
	StartBlackOcc  Bitboard = 0xFFFF << 48
	StartPawnOcc   Bitboard = (0xFF << 8) | (0xFF << 48)
	StartRookOcc   Bitboard = 0x81 | (0x81 << 56)
	StartKnightOcc Bitboard = 0x42 | (0x42 << 56)
	StartBishopOcc Bitboard = (1 << 2) | (1 << 5) | (1 << (2 + 56)) | (1 << (5 + 56))
	StartQueenOcc  Bitboard = 1<<3 | (1 << (3 + 56))
)

var StartPosition = Position{
	colOcc: [2]Bitboard{
		StartWhiteOcc,
		StartBlackOcc,
	},
	pawnOcc:   StartPawnOcc,
	rookOcc:   StartRookOcc | StartQueenOcc,
	bishopOcc: StartBishopOcc | StartQueenOcc,
	knightOcc: StartKnightOcc,
	status:    StartStatus,
}

// Status at start of game
var StartStatus = Status{
	KingStatus: [2]uint8{
		CanCastle | 4,
		CanCastle | 60,
	},
	TurnStatus: WHITE,
}

// ============== Getter/Setter ===========================

// Who should move from here ?
func (st Status) GetTurn() uint8 {
	return st.TurnStatus & 1
}

func (st *Status) SetTurn(turn uint8) {
	st.TurnStatus = (st.TurnStatus & 0b1110) | (turn & 1)
}

func (st *Status) SwitchTurn() {
	st.TurnStatus ^= 1
}

func (st Status) CanCastle(side uint8, castleBits uint8) bool {
	return st.KingStatus[side]&castleBits != 0
}

func (st Status) GetCastleBits(side uint8) uint8 {
	return uint8(st.KingStatus[side] & CanCastle)
}

func (st *Status) SetCastleBits(side uint8, castleBits uint8) {
	st.KingStatus[side] = (st.KingStatus[side] & 0b11_000000) | castleBits
}

func (st Status) GetKingPosition(side uint8) Square {
	return Square(st.KingStatus[side] & 0b11_1111)
}

func (st *Status) SetKingPosition(side uint8, sq Square) {
	st.KingStatus[side] = (st.KingStatus[side] & 0b11_000000) | uint8(sq)
}

func (st Status) String() string {
	buf := new(strings.Builder)
	fmt.Fprintln(buf, "\n--- Status ---")
	fmt.Fprintf(buf, "Turn: %d ( %d = WHITE, %d = BLACK )\n", st.GetTurn(), WHITE, BLACK)

	if st.CanCastle(WHITE, CanCastleKingSide) {
		fmt.Fprintf(buf, "WHITE can castle king side\n")
	}
	if st.CanCastle(WHITE, CanCastleQueenSide) {
		fmt.Fprintf(buf, "WHITE can castle queen side\n")
	}
	if st.CanCastle(BLACK, CanCastleKingSide) {
		fmt.Fprintf(buf, "BLACK can castle king side\n")
	}
	if st.CanCastle(BLACK, CanCastleQueenSide) {
		fmt.Fprintf(buf, "BLACK can castle queen side\n")
	}

	return buf.String()
}

// Extract the en passant flag of the opposite player.
// It is a pawn position that has no corresponding occupancy and one the last or first rank).
// 0 means no en passant flag in place.
func (p Position) EnPassantFlag() Bitboard {
	occ := p.colOcc[1-p.status.GetTurn()]
	return p.pawnOcc & ^occ
}

// Zero all enpassant flags
func (p *Position) ResetEnPassantFlag() {
	p.pawnOcc &= p.colOcc[WHITE] | p.colOcc[BLACK]
}

// Capture is the end position of any move.
// p.turn is playing, so enpassant will be triggered based on (1-p.turn) enpassant flag.
// If capture is not an active enpassant target, return 0
// If capture is an active enpassant square, remove is a map of pawn and occupancy(1 - p.turn) we will need to zero.
func (p Position) EnPassantCaptureMask(capture Square) (remove Bitboard) {

	if remove = p.EnPassantFlag(); remove != 0 {
		if turn := p.status.GetTurn(); turn == WHITE {
			if capture.Rank() == 5 {
				return remove | capture.South().Bitboard() // removes both the enpassant flag and the adverse pawn
			}

		} else { // turn == BLACK
			if capture.Rank() == 2 {
				return remove | capture.North().Bitboard() // removes both the enpassant flag and the adverse pawn
			}
		}
	}
	return 0
}
