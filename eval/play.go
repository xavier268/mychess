package eval

// Play the move, destroy n other child, return the new root node or nil if move is invalid.
// n itself is not automatically destroyed (will be garbage collected later).
func (n *Node) Play(m Move) (n2 *Node) {
	for i, mi := range n.Moves {
		if mi == m {
			return n.children[i]
		}
	}
	return nil
}

// Plays the best possible move.
// return nil if no move available.
func (n *Node) PlayBest() (n2 *Node, value float64, depth int) {
	m, value, depth := n.SelectBestMove()
	n2 = n.Play(m)
	return n2, value, depth
}
