package position

import (
	"fmt"
	"testing"
)

func TestGenerateRookAttacksTable(t *testing.T) {
	for sq := Square(0); sq < 63; sq += 17 {
		mm := GenerateRookAttacksMagicMapSq(sq)
		st := mm.Stats()
		fmt.Printf("Rook table for %s stats : coll %d, max %d\n", sq.String(), st.Coll, st.Maxsearch)
	}
}
