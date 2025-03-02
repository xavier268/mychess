package position

import (
	"fmt"
	"testing"
)

func TestNewBigTableInit(t *testing.T) {
	fmt.Println("Trying to create a BigTable")
	_ = NewBigTable()
	fmt.Println("BigTable created")
}
