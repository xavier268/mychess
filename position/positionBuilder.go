package position

// Utilities to create/modify specific positions
// Positions are specified as strings : a1 or C4 ...

// Side is WHITE or BLACK only

func (p *Position) AddPawn(side uint8, pawns ...string) *Position {
	side = side & 1
	for _, s := range pawns {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.pawnOcc |= 1 << sq
	}
	return p
}

func (p *Position) AddKnight(side uint8, knights ...string) *Position {
	side = side & 1
	for _, s := range knights {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.knightOcc |= 1 << sq
	}
	return p
}

func (p *Position) AddBishop(side uint8, bishops ...string) *Position {
	side = side & 1
	for _, s := range bishops {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.bishopOcc |= 1 << sq
	}
	return p
}

func (p *Position) AddRook(side uint8, rooks ...string) *Position {
	side = side & 1
	for _, s := range rooks {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.rookOcc |= 1 << sq
	}
	return p
}

func (p *Position) AddQueen(side uint8, queens ...string) *Position {
	side = side & 1
	for _, s := range queens {
		sq := SqParse(s)
		p.colOcc[side] |= 1 << sq
		p.rookOcc |= 1 << sq
		p.bishopOcc |= 1 << sq
	}
	return p
}

// ... and sets castling rights to NONE !
func (p *Position) AddKing(side uint8, kingsq string) *Position {
	side = side & 1
	sq := uint8(SqParse(kingsq))
	p.colOcc[side] |= 1 << sq
	p.status.KingStatus[side] = sq
	return p
}

// SetEnPassant records an en passant opportunity for the given side.
// Only the file of where is used. The phantom is placed at rank 2 (white's
// double push) or rank 5 (black's double push) — the capture landing square.
func (p *Position) SetEnPassant(side uint8, where string) *Position {
	side = side & 1
	file := SqParse(where).File()
	p.pawnOcc |= 1 << (Sq(2+3*int(side), file))
	return p
}

// Add specify castling rights
func (p *Position) SetCastle(side uint8, castleBits uint8) *Position {
	side = side & 1
	p.status.KingStatus[side] |= (castleBits & CanCastle)
	return p
}
