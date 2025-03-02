package position

// Special type to handle status word.
// Status is made of 3 groups of bytes :
//
// byte 0-1 : white details
//		bit 0-5 : King position
//		bit 6 : can castle king side
//		bit 7 : can castle king side
//		bit 8 : white king under check
//		bit 9-15 : ruf
//
// byte 2-3 : black details
//		same format
//
// byte 4-5 : uint16 count global ply counter // max 65K !
// byte 6 : uint8 count ply without capture or pawn move  // max 256
// byte 7 :
//	   	bit 1 : reversed colors vs actual physical color
//	   	bit 2 : game over
//	   	bit 3 : draw position
//		bit 4-7 : ruf
type Status uint64

const (
	CanCastleKingSide  = 0b10000000
	CanCastleQueenSide = 0b01000000
	CanCastle          = CanCastleQueenSide | CanCastleKingSide
)

func (st Status) GetWhiteKingSquare() Square {
	return Square(st & 0x3F)
}
func (st Status) SetWhiteKingSquare(sq Square) Status {
	return st & ^Status(0x3F) | Status(sq)
}

func (st Status) GetBlackKingSquare() Square {
	return Square((st >> 16) & 0x3F)
}

func (st Status) SetBlackKingSquare(sq Square) Status {
	return (st & ^Status(0x3F<<16)) | Status(sq)<<16
}

func (st Status) GetWhiteCastleBits() uint8 {
	return uint8(st & 0xC0)
}

func (st Status) GetBlackCastleBits() uint8 {
	return uint8((st >> 16) & 0xC0)
}

func (st Status) ReverseSwitchBit() Status {
	return st ^ Status(1<<56)
}

func (st Status) VMirror() Status {
	b0 := Status(uint8(st.GetBlackKingSquare().VMirror()) | (st.GetBlackCastleBits()))
	b1 := (st >> 8) & 0xFF
	b2 := Status(uint8(st.GetWhiteKingSquare().VMirror()) | (st.GetWhiteCastleBits()))
	b3 := (st >> 24) & 0xFF
	rest := (st.ReverseSwitchBit() &^ 0x_FFFF_FFFF)
	return Status(b0 | b1<<8 | b2<<16 | b3<<24 | rest)
}
