package eval

// Play the move, based on its index, destroy n other child, return the new root node or nil if move is invalid.
// n itself is not automatically destroyed (will be garbage collected later).
func (n *Node) Play(mi int) (n2 *Node) {

	if mi >= 0 && mi < len(n.Moves) {
		return n.children[mi]
	}
	return nil
}

// Plays the best possible move.
// return nil if no move available.
func (n *Node) PlayBest() (n2 *Node, value float64, depth int) {
	mi, value, depth := n.SelectBestMove()
	n2 = n.Play(mi)
	return n2, value, depth
}
