// implemnts evaluation strategy of positions
package eval

import (
	"mychess/position"
	"sync"
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
// Node is garanteed to have the full Moves legal fro this position.
// Children are matchning moves, but may be nil if not yet evaluated.
// Closing Stop channel stops evaluation for this position and all its children.
type Node struct {
	P        *Position     // p.Turn defines whose turn it is to play
	Moves    []Move        // all legal moves from this position
	Children []*Node       // children of this position. Matching the moves. Nil if not yet evaluated.
	mu       sync.Mutex    // protect child slice modifications
	value    float64       // value of this position, from the point of view of the player who is going to play (p.Turn). Never modified once set. Does not consider the children values.
	Stop     chan struct{} // closing this channel stops evalution for this position
}

// Create a new Node for the provided position.
// Value is initially set to a simple piece count.
func NewNode(p *Position) *Node {
	n := &Node{P: p}
	n.Stop = make(chan struct{}) // unbuffered !
	n.Moves = p.LegalMoves(nil)
	n.Children = make([]*Node, len(n.Moves))
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

// Grow the tree, evaluating its ci-th legal move, and adding it to the node.
func (n *Node) ComputeIthChild(i int) *Node {
	pc := n.P.Clone()
	// apply move
	pc.ExecuteMove(n.Moves[i])
	c := NewNode(pc)
	n.mu.Lock()
	n.Children[i] = c
	n.mu.Unlock()
	return n
}

// Compute value of given Node, using a min/max strategy, based only upon current available subtrees.
// Tree expansion is frozen during evaluation.
func (n *Node) Eval() float64 {
	// evaluate value recursively
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.Children) == 0 {
		return n.value
	}
	v := -n.Children[0].Eval()
	for _, c := range n.Children[1:] {
		v = max(v, -c.Eval())
	}
	return v
}

// return zero-value if no legal move available (stalemate or draw).
func (n *Node) SelectBestMove() (Move, float64) {
	m := Move{}
	mv := WORSTVALUE
	// evaluate value recursively
	n.mu.Lock()
	defer n.mu.Unlock()

	for i, c := range n.Children {
		if c == nil {
			continue
		}
		v := -c.Eval()
		if v > mv {
			mv = v
			m = c.Moves[i]
		}
	}
	return m, mv // best Move and its value
}
