// implemnts evaluation strategy of positions
package eval

import (
	"mychess/position"
)

type (
	Position = position.Position
	Move     = position.Move
)

const (
	WHITE  = position.WHITE
	BLACK  = position.BLACK
	PAWN   = position.PAWN
	KNIGHT = position.KNIGHT
	BISHOP = position.BISHOP
	ROOK   = position.ROOK
	QUEEN  = position.QUEEN
	KING   = position.KING
	EMPTY  = position.EMPTY
)

// Node creates a tree of evaluated positions.
// Node is garanteed to have the full Moves legal from this position.
// Children are matching moves, but may be nil if not yet evaluated.
// Closing Stop channel stops evaluation for this position and all its children.
type Node struct {
	P     *Position // p.Turn defines whose turn it is to play - immutable
	Moves []Move    // all legal moves from this position - immutable
	value float64   // value of this position, from the point of view of the player who is going to play (p.Turn) - immutable

	children []*Node // children of this position. Matching the moves. Nil if not yet evaluated.
}

// Create a new Node for the provided position.
// Value is initially set to a simple piece count and legal moves are computed.
func NewNode(p *Position) *Node {
	n := &Node{P: p}
	n.Moves = p.LegalMoves(nil)
	n.children = make([]*Node, len(n.Moves))
	// Set initial node value without looking at children ...
	if len(n.Moves) != 0 { // normal game continue ...
		n.value = basicEval(p)
		return n
	} else { // verify stalemate or draw ?
		if p.IsCheck(p.Turn) {
			// stalemate !
			n.value = WORSTVALUE
		} else {
			// draw
			n.value = 0
		}
	}
	return n
}

// Compute value of given Node, using a min/max strategy, based only upon current available subtrees.
func (n *Node) Eval() (v float64, depth int) {

	// evaluate value recursively
	v, depth = n.value, 0 // this will be either draw or stalemate if no move, basicValue if a legal move is available.
	for _, c := range n.children {
		if c == nil {
			continue // keep value
		}
		vc, dc := c.Eval()
		v = max(v, -vc)
		depth = max(dc+1, depth)
	}
	return v, depth
}

// return index of -1 if no legal move available (value will reflect stalemate or draw).
// if no children were analysed, will not suggest a "best move" and return -1.
// It is garanteed that n.children[best] is a non nil node if indx >= 0.
func (n *Node) SelectBestMove() (indx int, moveValue float64, depth int) {

	indx = -1
	moveValue = WORSTVALUE
	depth = 0

	// Envision all legal moves
	for i := range n.Moves {
		c := n.children[i]
		if c == nil {
			continue
		}
		v, d := c.Eval()
		if -v > moveValue {
			moveValue = -v
			indx = i
			depth = d
		}
	}
	return indx, moveValue, depth // best Move and its value/depth
}

// Add exactly 1 layer to the tree
func (n *Node) Expand() {
	// expand children
	for i, m := range n.Moves {
		if n.children[i] == nil {
			// create children if it does not exists
			p2 := n.P.Clone()
			p2.ExecuteMove(m) // turn has changed ...
			n.children[i] = NewNode(p2)
			// no recursion here
		} else {
			// expand the children recursively
			n.children[i].Expand()
		}
	}
}

// Explore the best branch, and expand its leave.
func (n *Node) ExpandBest() {
	b := n.findBestLeave()
	b.Expand()
}

func (n *Node) ExpandBestN(count int) {
	n.ExpandBest()
	if count > 1 {
		n.ExpandBestN(count - 1)
	}
}

// could be this n, if no further information
func (n *Node) findBestLeave() *Node {
	indx, _, _ := n.SelectBestMove()
	if indx == -1 {
		return n
	}
	return n.children[indx].findBestLeave()
}
