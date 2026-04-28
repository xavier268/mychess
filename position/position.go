// Low-level library for board state representation and legal moves generation.
package position

import (
	"fmt"
	"strings"
)

// ============== Position definition ==============================

// Constant to indicate turn (who is expected to play) and index the masks and status.
const (
	WHITE = 0
	BLACK = 1
)

type Position struct {
	// // Occupancies for each side, all actual pieces (king INCLUDED), but en passant marker not included
	colOcc [2]Bitboard
	// all colors together
	// en passant is indicated as a "phantom" pawn between the initial position and the actual move, 2 squares further. The Phantom Pawn has NO occupancy flag, that is the way to recognize it is an "en passant" artefact.
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

	// Zobrist hash of the position; 0 when uninitialized.
	// Maintained incrementally by DoMove; restored atomically by UndoMove.
	Hash uint64
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
	st.KingStatus[side] = (st.KingStatus[side] & 0b00_111111) | castleBits
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

// IsCheck retourne true si le joueur courant est en échec.
func (p Position) IsCheck() bool {
	t := p.status.GetTurn()
	return p.IsSquareAttacked(p.status.GetKingPosition(t), 1^t)
}

// Turn retourne le camp qui doit jouer : WHITE (0) ou BLACK (1).
func (p Position) Turn() uint8 {
	return p.status.GetTurn()
}

// CastleBits retourne le masque des droits de roque pour le camp donné.
// Testez avec position.CanCastleKingSide et position.CanCastleQueenSide.
func (p Position) CastleBits(side uint8) uint8 {
	return p.status.GetCastleBits(side)
}

// KingPosition retourne la case du roi du camp `side` (WHITE ou BLACK).
func (p Position) KingPosition(side uint8) Square {
	return p.status.GetKingPosition(side)
}

// Extract the en passant flag of the opposite player.
// Phantom pawns (in pawnOcc but not in colOcc) at rank 7 are BLACK's EP signals (black double-pushed).
// Phantom pawns at rank 0 are WHITE's EP signals (white double-pushed).
// Returns the opponent's phantom pawn(s) relevant to the current turn.
func (p Position) EnPassantFlag() Bitboard {
	phantoms := p.pawnOcc & ^(p.colOcc[WHITE] | p.colOcc[BLACK])
	if p.status.GetTurn() == WHITE {
		return phantoms & Rank(7) // black's signals
	}
	return phantoms & Rank(0) // white's signals
}

// Zero all enpassant flags (phantom pawns not in any colOcc)
func (p *Position) ResetEnPassantFlag() {
	p.pawnOcc &= p.colOcc[WHITE] | p.colOcc[BLACK]
}

// EnPassantCaptureMask returns the squares to clear in pawnOcc when an en passant capture lands on `capture`.
// Returns 0 if `capture` is not a valid en passant landing square.
// The returned mask includes: the phantom pawn + the actual captured pawn.
func (p Position) EnPassantCaptureMask(capture Square) (remove Bitboard) {
	epFlag := p.EnPassantFlag()
	if epFlag == 0 {
		return 0
	}
	turn := p.status.GetTurn()
	if turn == WHITE && capture.Rank() == 5 {
		// Black's phantom is at rank 7, same file as the landing square
		phantom := Sq(7, capture.File())
		if epFlag.IsSet(phantom) {
			// Remove phantom + the actual black pawn at rank 4
			return phantom.Bitboard() | Sq(4, capture.File()).Bitboard()
		}
	} else if turn == BLACK && capture.Rank() == 2 {
		// White's phantom is at rank 0, same file as the landing square
		phantom := Sq(0, capture.File())
		if epFlag.IsSet(phantom) {
			// Remove phantom + the actual white pawn at rank 3
			return phantom.Bitboard() | Sq(3, capture.File()).Bitboard()
		}
	}
	return 0
}
