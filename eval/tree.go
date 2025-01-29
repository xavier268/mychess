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
// Value is initially set to a simple piece count.
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
// Tree expansion is frozen during evaluation.
func (n *Node) Eval() (v float64, depth int) {

	// TODO - need better protection against nil dereferencing !
	// evaluate value recursively
	if len(n.Moves) == 0 {
		return n.value, 1 // TODO - ici, se poser la question du draw vs stalemate pour fixer la valeur ?
	}
	v, depth = WORSTVALUE, -1
	for _, c := range n.children[1:] {
		// TODO - Mettre un marqueur si on a trouvé au moins un children non nul ... sinon, utiliser value !
		if c == nil {
			continue
		}
		vc, dc := c.Eval()
		v = max(v, -vc)
		depth = max(dc+1, depth)
	}
	return v, depth
}

// return zero-value if no legal move available (stalemate or draw).
// The whole tree will be locked during evaluation.
func (n *Node) SelectBestMove() (move Move, moveValue float64, depth int) {
	// TODO - A réecrire pour les cas ou les children ne sont pas tous développés ?
	move = Move{}
	moveValue = WORSTVALUE
	depth = 0

	// Envision all legal moves
	for i, c := range n.children {
		if c == nil {
			continue
		}
		v, d := c.Eval()
		if -v > moveValue {
			moveValue = -v
			move = c.Moves[i]
			depth = d
		}
	}
	return move, moveValue, depth // best Move and its value/depth
}

// Grow a certain number of layers
func (n *Node) Grow(depth int) {

	// TODO - A réecrire pour "rajouter" une layer de l'epaisseur donnée au bout de l'arbre et ne pas seulement contraidre la propfondeur de l'arbre ?
	// TODO - ou une fonction Expand qui fait ça 1 fois, et que l'on itérer ?
	if depth <= 0 {
		return
	}

	// expand children
	for i, m := range n.Moves {
		if n.children[i] == nil {
			// create children if it does not exists
			nc := NewNode(n.P.Clone())
			nc.P.ExecuteMove(m)
			n.children[i] = nc
		}
		// grow the children
		n.children[i].Grow(depth - 1)
	}
}
