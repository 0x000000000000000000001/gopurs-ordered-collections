package main

import (
	"fmt"
)

func cmpInt(a, b interface{}) int {
	ai := a.(int)
	bi := b.(int)
	if ai < bi { return -1 }
	if ai > bi { return 1 }
	return 0
}

// I will copy the BTree code here and run it!
