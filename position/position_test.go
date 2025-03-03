package position

import (
	"fmt"
	"testing"
)

func TestDisplayPosition(t *testing.T) {

	fmt.Println(StartPosition.String())
	StartPosition.Dump()

}
