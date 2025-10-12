package utils

import (
	"fmt"
	"testing"
)

func TestStartWith(t *testing.T) {
	fmt.Println(StartsWithAlphanumeric("?abc"))
	fmt.Println(StartsWithAlphanumeric("?123")) // false
	fmt.Println(StartsWithAlphanumeric("a?bc")) // true (因为第一个字符是 'a')
}
