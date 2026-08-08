package Data_Map_Internal


type CompareFn func(a, b interface{}) int

const maxDegree = 16
const minItems = (maxDegree - 1) / 2

type Item struct {
	Key   interface{}
	Value interface{}
}

type Node struct {
	items    []Item
	children []*Node
}

func (n *Node) clone() *Node {
	newNode := &Node{
		items:    make([]Item, len(n.items), maxDegree),
		children: make([]*Node, len(n.children), maxDegree+1),
	}
	copy(newNode.items, n.items)
	copy(newNode.children, n.children)
	return newNode
}

func (n *Node) find(key interface{}, cmp CompareFn) (int, bool) {
	for i, item := range n.items {
		c := cmp(key, item.Key)
		if c == 0 {
			return i, true
		} else if c < 0 {
			return i, false
		}
	}
	return len(n.items), false
}

type BTree struct {
	root    *Node
	size    int
	compare CompareFn
}
type anyValuer interface {
	AnyVal() any
}

func asTree(m interface{}) *BTree {
	if val, ok := m.(anyValuer); ok {
		return val.AnyVal().(*BTree)
	}
	return m.(*BTree)
}


func NewBTree(cmp CompareFn) *BTree {
	return &BTree{compare: cmp}
}

func (t *BTree) Size() int { return t.size }

func (t *BTree) Lookup(key interface{}) (interface{}, bool) {
	if t.root == nil {
		return nil, false
	}
	return t.root.lookup(key, t.compare)
}

func (n *Node) lookup(key interface{}, cmp CompareFn) (interface{}, bool) {
	i, found := n.find(key, cmp)
	if found {
		return n.items[i].Value, true
	}
	if len(n.children) == 0 {
		return nil, false
	}
	return n.children[i].lookup(key, cmp)
}

func (t *BTree) Insert(key, value interface{}) *BTree {
	if t.root == nil {
		return &BTree{
			root: &Node{items: []Item{{key, value}}},
			size: 1,
			compare: t.compare,
		}
	}

	newRoot, splitItem, rightNode, replaced := t.root.insert(key, value, t.compare)
	newSize := t.size
	if !replaced {
		newSize++
	}

	if rightNode != nil {
		// Root split
		return &BTree{
			root: &Node{
				items:    []Item{splitItem},
				children: []*Node{newRoot, rightNode},
			},
			size:    newSize,
			compare: t.compare,
		}
	}

	return &BTree{
		root:    newRoot,
		size:    newSize,
		compare: t.compare,
	}
}

func (n *Node) insert(key, value interface{}, cmp CompareFn) (*Node, Item, *Node, bool) {
	i, found := n.find(key, cmp)
	newNode := n.clone()

	if found {
		newNode.items[i].Value = value
		return newNode, Item{}, nil, true
	}

	var replaced bool
	var splitItem Item
	var rightChild *Node

	if len(n.children) == 0 {
		newNode.items = append(newNode.items, Item{})
		copy(newNode.items[i+1:], newNode.items[i:])
		newNode.items[i] = Item{key, value}
	} else {
		var childNew *Node
		childNew, splitItem, rightChild, replaced = n.children[i].insert(key, value, cmp)
		newNode.children[i] = childNew
		if rightChild != nil {
			newNode.items = append(newNode.items, Item{})
			copy(newNode.items[i+1:], newNode.items[i:])
			newNode.items[i] = splitItem

			newNode.children = append(newNode.children, nil)
			copy(newNode.children[i+2:], newNode.children[i+1:])
			newNode.children[i+1] = rightChild
		}
	}

	if len(newNode.items) >= maxDegree {
		mid := maxDegree / 2
		upItem := newNode.items[mid]

		rightNode := &Node{
			items: make([]Item, len(newNode.items[mid+1:]), maxDegree),
		}
		copy(rightNode.items, newNode.items[mid+1:])
		if len(newNode.children) > 0 {
			rightNode.children = make([]*Node, len(newNode.children[mid+1:]), maxDegree+1)
			copy(rightNode.children, newNode.children[mid+1:])
			newNode.children = newNode.children[:mid+1]
		}
		newNode.items = newNode.items[:mid]

		return newNode, upItem, rightNode, replaced
	}

	return newNode, Item{}, nil, replaced
}

func (t *BTree) Delete(key interface{}) *BTree {
	if t.root == nil {
		return t
	}

	newRoot, deleted := t.root.delete(key, t.compare)
	if !deleted {
		return t // unchanged
	}

	// If root has no items and has one child, the child becomes the new root
	if len(newRoot.items) == 0 && len(newRoot.children) == 1 {
		newRoot = newRoot.children[0]
	} else if len(newRoot.items) == 0 && len(newRoot.children) == 0 {
		newRoot = nil
	}

	return &BTree{
		root:    newRoot,
		size:    t.size - 1,
		compare: t.compare,
	}
}

func (n *Node) delete(key interface{}, cmp CompareFn) (*Node, bool) {
	i, found := n.find(key, cmp)
	newNode := n.clone()

	if len(n.children) == 0 {
		if found {
			// Leaf node, just remove the item
			newNode.items = append(newNode.items[:i], newNode.items[i+1:]...)
			return newNode, true
		}
		return nil, false
	}

	if found {
		// Internal node, replace with predecessor
		// Predecessor is the largest item in children[i]
		predChild, predItem := n.children[i].deleteMax()
		newNode.items[i] = predItem
		newNode.children[i] = predChild
		
		newNode.rebalanceChild(i)
		return newNode, true
	}

	// Not found, go down to children[i]
	childNew, deleted := n.children[i].delete(key, cmp)
	if !deleted {
		return nil, false
	}
	newNode.children[i] = childNew
	newNode.rebalanceChild(i)
	return newNode, true
}

func (n *Node) deleteMax() (*Node, Item) {
	newNode := n.clone()
	if len(n.children) == 0 {
		item := newNode.items[len(newNode.items)-1]
		newNode.items = newNode.items[:len(newNode.items)-1]
		return newNode, item
	}
	
	lastIdx := len(n.children) - 1
	childNew, item := n.children[lastIdx].deleteMax()
	newNode.children[lastIdx] = childNew
	newNode.rebalanceChild(lastIdx)
	return newNode, item
}

// rebalanceChild ensures children[i] has at least minItems items.
// If it doesn't, it borrows from a sibling or merges.
func (n *Node) rebalanceChild(i int) {
	child := n.children[i]
	if len(child.items) >= minItems {
		return
	}

	// Try borrowing from left sibling
	if i > 0 && len(n.children[i-1].items) > minItems {
		leftSib := n.children[i-1].clone()
		newChild := child.clone()
		
		// Move n.items[i-1] down to child
		newChild.items = append([]Item{n.items[i-1]}, newChild.items...)
		if len(leftSib.children) > 0 {
			newChild.children = append([]*Node{leftSib.children[len(leftSib.children)-1]}, newChild.children...)
			leftSib.children = leftSib.children[:len(leftSib.children)-1]
		}
		
		// Move leftSib's last item up to n
		n.items[i-1] = leftSib.items[len(leftSib.items)-1]
		leftSib.items = leftSib.items[:len(leftSib.items)-1]
		
		n.children[i-1] = leftSib
		n.children[i] = newChild
		return
	}

	// Try borrowing from right sibling
	if i < len(n.children)-1 && len(n.children[i+1].items) > minItems {
		rightSib := n.children[i+1].clone()
		newChild := child.clone()
		
		// Move n.items[i] down to child
		newChild.items = append(newChild.items, n.items[i])
		if len(rightSib.children) > 0 {
			newChild.children = append(newChild.children, rightSib.children[0])
			rightSib.children = rightSib.children[1:]
		}
		
		// Move rightSib's first item up to n
		n.items[i] = rightSib.items[0]
		rightSib.items = rightSib.items[1:]
		
		n.children[i+1] = rightSib
		n.children[i] = newChild
		return
	}

	// Must merge. We merge child and a sibling.
	if i > 0 {
		// Merge children[i-1] and children[i]
		leftSib := n.children[i-1].clone()
		// Bring n.items[i-1] down
		leftSib.items = append(leftSib.items, n.items[i-1])
		leftSib.items = append(leftSib.items, child.items...)
		if len(child.children) > 0 {
			leftSib.children = append(leftSib.children, child.children...)
		}
		
		// Remove n.items[i-1] and n.children[i]
		n.items = append(n.items[:i-1], n.items[i:]...)
		n.children = append(n.children[:i], n.children[i+1:]...)
		n.children[i-1] = leftSib
	} else {
		// Merge children[i] and children[i+1]
		newChild := child.clone()
		rightSib := n.children[i+1]
		
		newChild.items = append(newChild.items, n.items[i])
		newChild.items = append(newChild.items, rightSib.items...)
		if len(rightSib.children) > 0 {
			newChild.children = append(newChild.children, rightSib.children...)
		}
		
		n.items = append(n.items[:i], n.items[i+1:]...)
		n.children = append(n.children[:i+1], n.children[i+2:]...)
		n.children[i] = newChild
	}
}

type FoldFn func(acc, key, value interface{}) interface{}

func (t *BTree) Foldl(f FoldFn, acc interface{}) interface{} {
	if t.root == nil {
		return acc
	}
	return t.root.foldl(f, acc)
}

func (n *Node) foldl(f FoldFn, acc interface{}) interface{} {
	if len(n.children) == 0 {
		for _, item := range n.items {
			acc = f(acc, item.Key, item.Value)
		}
		return acc
	}

	for i, item := range n.items {
		acc = n.children[i].foldl(f, acc)
		acc = f(acc, item.Key, item.Value)
	}
	acc = n.children[len(n.items)].foldl(f, acc)
	return acc
}

func (t *BTree) Foldr(f FoldFn, acc interface{}) interface{} {
	if t.root == nil {
		return acc
	}
	return t.root.foldr(f, acc)
}

func (n *Node) foldr(f FoldFn, acc interface{}) interface{} {
	if len(n.children) == 0 {
		for i := len(n.items) - 1; i >= 0; i-- {
			item := n.items[i]
			acc = f(acc, item.Key, item.Value)
		}
		return acc
	}

	acc = n.children[len(n.items)].foldr(f, acc)
	for i := len(n.items) - 1; i >= 0; i-- {
		item := n.items[i]
		acc = f(acc, item.Key, item.Value)
		acc = n.children[i].foldr(f, acc)
	}
	return acc
}

func (t *BTree) Keys() []interface{} {
	keys := make([]interface{}, 0, t.size)
	res := t.Foldl(func(acc, key, value interface{}) interface{} {
		return append(acc.([]interface{}), key)
	}, keys)
	return res.([]interface{})
}

func (t *BTree) Values() []interface{} {
	values := make([]interface{}, 0, t.size)
	res := t.Foldl(func(acc, key, value interface{}) interface{} {
		return append(acc.([]interface{}), value)
	}, values)
	return res.([]interface{})
}

type CombineFn func(v1, v2 interface{}) interface{}

func (t *BTree) UnionWith(other *BTree, f CombineFn) *BTree {
	if t.size == 0 {
		return other
	}
	if other.size == 0 {
		return t
	}

	res := t
	finalRes := other.Foldl(func(acc, key, value interface{}) interface{} {
		tree := asTree(acc)
		existing, ok := tree.Lookup(key)
		if ok {
			newVal := f(existing, value)
			return tree.Insert(key, newVal)
		}
		return tree.Insert(key, value)
	}, res)

	return asTree(finalRes)
}

func (t *BTree) IntersectionWith(other *BTree, f CombineFn) *BTree {
	res := NewBTree(t.compare)
	if t.size == 0 || other.size == 0 {
		return res
	}

	smaller, larger := t, other
	if t.size > other.size {
		smaller, larger = other, t
	}

	finalRes := smaller.Foldl(func(acc, key, value interface{}) interface{} {
		tree := asTree(acc)
		if otherVal, ok := larger.Lookup(key); ok {
			var newVal interface{}
			if smaller == t {
				newVal = f(value, otherVal)
			} else {
				newVal = f(otherVal, value)
			}
			return tree.Insert(key, newVal)
		}
		return tree
	}, res)

	return asTree(finalRes)
}

func (t *BTree) Difference(other *BTree) *BTree {
	res := t
	if t.size == 0 || other.size == 0 {
		return res
	}

	finalRes := other.Foldl(func(acc, key, _ interface{}) interface{} {
		tree := asTree(acc)
		return tree.Delete(key)
	}, res)

	return asTree(finalRes)
}


// FFI Wrappers

var Empty = func() interface{} {
	// A dummy cmp function because the BTree struct requires it,
	// but it will be replaced on the first insert since insert creates a new tree.
	// Actually, an empty tree does not know its cmp until the first insert!
	// In our FFI, we pass cmp on EVERY operation.
	// So NewBTree(nil) is fine.
	return NewBTree(nil)
}()

func IsEmpty(m interface{}) bool {
	return asTree(m).Size() == 0
}

func Singleton(k interface{}) func(interface{}) interface{} {
	return func(v interface{}) interface{} {
		return &BTree{
			root: &Node{items: []Item{{k, v}}},
			size: 1,
			compare: nil, // compare will be injected on next insert
		}
	}
}

func InsertImpl(compare func(interface{}) func(interface{}) interface{}, fromOrdering func(interface{}) int, k interface{}, v interface{}, m interface{}) interface{} {
	tree := asTree(m)
	cmp := func(a, b interface{}) int {
		return fromOrdering(compare(a)(b))
	}
	if tree.Size() == 0 {
		return NewBTree(cmp).Insert(k, v)
	}
	// We must ensure the tree uses the current cmp
	tree.compare = cmp
	return tree.Insert(k, v)
}

func InsertWithImpl(compare func(interface{}) func(interface{}) interface{}, fromOrdering func(interface{}) int, f func(interface{}) func(interface{}) interface{}, k interface{}, v interface{}, m interface{}) interface{} {
	tree := asTree(m)
	cmp := func(a, b interface{}) int {
		return fromOrdering(compare(a)(b))
	}
	if tree.Size() == 0 {
		return NewBTree(cmp).Insert(k, v)
	}
	tree.compare = cmp
	
	existing, ok := tree.Lookup(k)
	if ok {
		newVal := f(existing)(v)
		return tree.Insert(k, newVal)
	}
	return tree.Insert(k, v)
}

func LookupImpl(just func(interface{}) interface{}, nothing interface{}, compare func(interface{}) func(interface{}) interface{}, fromOrdering func(interface{}) int, k interface{}, m interface{}) interface{} {
	tree := asTree(m)
	if tree.Size() == 0 {
		return nothing
	}
	cmp := func(a, b interface{}) int {
		return fromOrdering(compare(a)(b))
	}
	tree.compare = cmp
	val, ok := tree.Lookup(k)
	if ok {
		return just(val)
	}
	return nothing
}

func DeleteImpl(compare func(interface{}) func(interface{}) interface{}, fromOrdering func(interface{}) int, k interface{}, m interface{}) interface{} {
	tree := asTree(m)
	if tree.Size() == 0 {
		return m
	}
	cmp := func(a, b interface{}) int {
		return fromOrdering(compare(a)(b))
	}
	tree.compare = cmp
	return tree.Delete(k)
}

func KeysImpl(m interface{}) []interface{} {
	return asTree(m).Keys()
}

func ValuesImpl(m interface{}) []interface{} {
	return asTree(m).Values()
}

func UnionWithImpl(compare func(interface{}) func(interface{}) interface{}, fromOrdering func(interface{}) int, f func(interface{}) func(interface{}) interface{}, m1 interface{}, m2 interface{}) interface{} {
	t1 := asTree(m1)
	t2 := asTree(m2)
	cmp := func(a, b interface{}) int {
		return fromOrdering(compare(a)(b))
	}
	t1.compare = cmp
	t2.compare = cmp
	return t1.UnionWith(t2, func(v1, v2 interface{}) interface{} {
		return f(v1)(v2)
	})
}

func IntersectionWithImpl(compare func(interface{}) func(interface{}) interface{}, fromOrdering func(interface{}) int, f func(interface{}) func(interface{}) interface{}, m1 interface{}, m2 interface{}) interface{} {
	t1 := asTree(m1)
	t2 := asTree(m2)
	cmp := func(a, b interface{}) int {
		return fromOrdering(compare(a)(b))
	}
	t1.compare = cmp
	t2.compare = cmp
	return t1.IntersectionWith(t2, func(v1, v2 interface{}) interface{} {
		return f(v1)(v2)
	})
}

func DifferenceImpl(compare func(interface{}) func(interface{}) interface{}, fromOrdering func(interface{}) int, m1 interface{}, m2 interface{}) interface{} {
	t1 := asTree(m1)
	t2 := asTree(m2)
	cmp := func(a, b interface{}) int {
		return fromOrdering(compare(a)(b))
	}
	t1.compare = cmp
	t2.compare = cmp
	return t1.Difference(t2)
}

func SizeImpl(m interface{}) int {
	return asTree(m).Size()
}

func (n *Node) mapValues(f func(interface{}) interface{}) *Node {
	if n == nil {
		return nil
	}
	newNode := &Node{
		items:    make([]Item, len(n.items), maxDegree),
		children: make([]*Node, len(n.children), maxDegree+1),
	}
	for i, item := range n.items {
		newNode.items[i] = Item{Key: item.Key, Value: f(item.Value)}
	}
	for i, child := range n.children {
		newNode.children[i] = child.mapValues(f)
	}
	return newNode
}

func (t *BTree) MapValues(f func(interface{}) interface{}) *BTree {
	return &BTree{
		root:    t.root.mapValues(f),
		size:    t.size,
		compare: t.compare,
	}
}

func MapImpl(f func(interface{}) interface{}, m interface{}) interface{} {
	return asTree(m).MapValues(f)
}

func FoldlImpl(f func(interface{}) func(interface{}) func(interface{}) interface{}, z interface{}, m interface{}) interface{} {
	tree := asTree(m)
	return tree.Foldl(func(acc, key, value interface{}) interface{} {
		return f(acc)(key)(value)
	}, z)
}

func FoldrImpl(f func(interface{}) func(interface{}) func(interface{}) interface{}, z interface{}, m interface{}) interface{} {
	tree := asTree(m)
	return tree.Foldr(func(acc, key, value interface{}) interface{} {
		return f(acc)(key)(value)
	}, z)
}


func FilterKeysImpl(p func(interface{}) interface{}, m interface{}) interface{} {
	tree := asTree(m)
	res := NewBTree(tree.compare)
	finalRes := tree.Foldl(func(acc, key, value interface{}) interface{} {
		accTree := asTree(acc)
		pv := p(key)
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
