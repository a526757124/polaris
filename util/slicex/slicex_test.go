package slicex

import (
	"testing"
	"fmt"
)

func TestFindIndex(t *testing.T) {
	vals := []string{"1", "2", "3"}
	find := "1"
	fmt.Println(FindIndex(vals, find))
	fmt.Println(vals[0:0])
}
