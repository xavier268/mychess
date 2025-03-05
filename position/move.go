package position

// A optimiser pour compacité ?
type Move struct {
	From, To, // 2 * 6 bits = 12 bits
	Promotion, // 4 choices, 2 bits.
	Score  // to rank moves for alpha/beta ? 2 bits enough, if added to promotion ?
	uint16
}