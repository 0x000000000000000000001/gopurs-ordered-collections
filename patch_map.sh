#!/bin/bash
cat << 'INNER_EOF' >> src/Data/Map/Internal.go

func SubmapImpl(compare func(interface{}) func(interface{}) interface{}, fromOrdering func(interface{}) int, m1 interface{}, m2 interface{}) bool {
	// Not needed by FFI since it is implemented in PureScript
	return true
}

func FilterKeysImpl(p func(interface{}) interface{}, m interface{}) interface{} {
	tree := asTree(m)
	res := NewBTree(tree.compare)
	finalRes := tree.Foldl(func(acc, key, value interface{}) interface{} {
		accTree := asTree(acc)
		pv := p(key)
		// Check if pv is true
		var isTrue bool
		if val, ok := pv.(gopurs_runtime.Value); ok {
			isTrue = val.IntVal != 0
		} else {
			panic("Expected Value from p(key)")
		}
		if isTrue {
			return accTree.Insert(key, value)
		}
		return accTree
	}, res)
	return asTree(finalRes)
}
INNER_EOF
