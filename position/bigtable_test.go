package position

import (
	"fmt"
	"testing"
)

func TestNewBigTableInit(t *testing.T) {
	fmt.Println("Trying to create a BigTable")
	_ = newBigTable()
	fmt.Println("BigTable created")
}
