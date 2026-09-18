package Data_Map_Internal

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func unionEntries(tree *BTree) []Item {
	items := []Item{}
	tree.Foldl(func(_, key, value interface{}) interface{} {
		items = append(items, Item{key, value})
		return nil
	}, nil)
	return items
}

func unionFixture(compare CompareFn, start, size, multiplier int) *BTree {
	tree := newBTree(compare)
	for key := start; key < start+size; key++ {
		tree = tree.Insert(key, key*multiplier)
	}
	return tree
}

func TestUnionWithSizesOrderingAndPersistence(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		compare := func(a, b interface{}) int {
			if reverse {
				return b.(int) - a.(int)
			}
			return a.(int) - b.(int)
		}
		for _, sizes := range [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}, {1, 32}, {32, 1}, {32, 32}, {64, 256}, {128, 256}, {255, 256}} {
			for _, offset := range []int{0, 1, 50} {
				left := unionFixture(compare, 0, sizes[0], 10)
				right := unionFixture(compare, offset, sizes[1], 100)
				leftBefore, rightBefore := unionEntries(left), unionEntries(right)
				for _, combine := range []CombineFn{
					func(a, b interface{}) interface{} { return fmt.Sprintf("(%v|%v)", a, b) },
					func(a, _ interface{}) interface{} { return a },
				} {
					expected := left.UnionWith(right, combine)
					actual := asTree(UnionWithImpl(
						func(a, b interface{}) interface{} { return compare(a, b) },
						func(v interface{}) int { return v.(int) },
						func(a interface{}) func(interface{}) interface{} {
							return func(b interface{}) interface{} { return combine(a, b) }
						}, left, right))
					if !reflect.DeepEqual(unionEntries(expected), unionEntries(actual)) {
						t.Fatalf("union differs: reverse=%t sizes=%v offset=%d", reverse, sizes, offset)
					}
					_ = actual.Insert(-1, "new value")
					if !reflect.DeepEqual(leftBefore, unionEntries(left)) || !reflect.DeepEqual(rightBefore, unionEntries(right)) {
						t.Fatal("union or subsequent insertion mutated an input")
					}
				}
			}
		}
	}
}

func TestUnionRetainsLeftRepresentativeKey(t *testing.T) {
	type key struct {
		rank int
		name string
	}
	compare := func(a, b interface{}) int { return a.(key).rank - b.(key).rank }
	left := newBTree(compare).Insert(key{2, "left"}, "left value")
	right := newBTree(compare)
	for rank := 0; rank < 32; rank++ {
		right = right.Insert(key{rank, "right"}, "right value")
	}
	combine := func(a, b interface{}) interface{} { return a.(string) + "/" + b.(string) }
	actual := unionWithSameOrdering(left, right, combine)
	if !reflect.DeepEqual(unionEntries(left.UnionWith(right, combine)), unionEntries(actual)) {
		t.Fatal("Ord-equivalent collision lost the left representative key or combination order")
	}
	if got := unionEntries(actual)[2]; got.Key != (key{2, "left"}) || got.Value != "left value/right value" {
		t.Fatalf("wrong collision: %#v", got)
	}
}

func TestRawUnionWithDifferentOrderings(t *testing.T) {
	ascending := func(a, b interface{}) int { return a.(int) - b.(int) }
	descending := func(a, b interface{}) int { return b.(int) - a.(int) }
	left := newBTree(ascending).Insert(2, -1)
	right := unionFixture(descending, 0, 32, 10)
	actual := left.UnionWith(right, func(a, _ interface{}) interface{} { return a })
	if actual.Size() != 32 {
		t.Fatalf("raw union size: %d", actual.Size())
	}
	for rank, item := range unionEntries(actual) {
		want := rank * 10
		if rank == 2 {
			want = -1
		}
		if item.Key != rank || item.Value != want {
			t.Fatalf("raw union changed ordering or left bias: %#v", item)
		}
		if value, ok := actual.Lookup(rank); !ok || value != want {
			t.Fatalf("raw union lookup %d: %v, %t", rank, value, ok)
		}
	}
}

var unionSink *BTree

func TestUnionOverlapAllocationBudget(t *testing.T) {
	compare := func(a, b interface{}) int { return a.(int) - b.(int) }
	combine := func(a, _ interface{}) interface{} { return a }
	right := unionFixture(compare, 0, 256, 10)
	for _, size := range []int{64, 128, 255} {
		left := unionFixture(compare, 0, size, 100)
		baseline := testing.AllocsPerRun(10, func() { unionSink = left.UnionWith(right, combine) })
		optimized := testing.AllocsPerRun(10, func() { unionSink = unionWithSameOrdering(left, right, combine) })
		if optimized > baseline {
			t.Errorf("overlap %d/256 allocates more: %g > %g", size, optimized, baseline)
		}
	}
}

// node test/native-map.mjs --bench-union
func BenchmarkMapUnion(b *testing.B) {
	compare := func(a, b interface{}) int { return a.(int) - b.(int) }
	combine := func(a, _ interface{}) interface{} { return a }
	for _, size := range []int{128, 256, 512} {
		for _, overlap := range []bool{false, true} {
			for _, baseline := range []bool{true, false} {
				b.Run(fmt.Sprintf("baseline=%t/overlap=%t/n%d", baseline, overlap, size), func(b *testing.B) {
					key := -1
					if overlap {
						key = size / 2
					}
					left := newBTree(compare).Insert(key, -1)
					right := unionFixture(compare, 0, size, 10)
					b.ReportAllocs()
					b.ResetTimer()
					for iteration := 0; iteration < b.N; iteration++ {
						if baseline {
							unionSink = left.UnionWith(right, combine)
						} else {
							unionSink = unionWithSameOrdering(left, right, combine)
						}
					}
				})
			}
		}
	}
}

func BenchmarkMapUnionHighOverlap(b *testing.B) {
	compare := func(a, b interface{}) int { return a.(int) - b.(int) }
	combine := func(a, _ interface{}) interface{} { return a }
	for _, size := range []int{256, 512} {
		for _, leftSize := range []int{size / 4, size / 2, size - 1} {
			for _, baseline := range []bool{true, false} {
				b.Run(fmt.Sprintf("baseline=%t/left%d-right%d", baseline, leftSize, size), func(b *testing.B) {
					left := unionFixture(compare, 0, leftSize, 10)
					right := unionFixture(compare, 0, size, 100)
					b.ReportAllocs()
					b.ResetTimer()
					for iteration := 0; iteration < b.N; iteration++ {
						if baseline {
							unionSink = left.UnionWith(right, combine)
						} else {
							unionSink = unionWithSameOrdering(left, right, combine)
						}
					}
				})
			}
		}
	}
}

func TestToArrayOrderAndPersistence(t *testing.T) {
	compare := func(a, b interface{}) int { return a.(int) - b.(int) }
	tuple := func(key interface{}) func(interface{}) interface{} {
		return func(value interface{}) interface{} { return [2]int{key.(int), value.(int)} }
	}
	original := newBTree(compare)
	for key := 999; key >= 0; key-- {
		original = original.Insert(key, key*10)
	}
	check := func(tree *BTree, expectedSize int) []interface{} {
		t.Helper()
		items := ToArrayImpl(tuple, tree)
		if len(items) != expectedSize {
			t.Fatalf("length %d != %d", len(items), expectedSize)
		}
		for i, item := range items {
			if item != ([2]int{i, i * 10}) {
				t.Fatalf("entry %d: %v", i, item)
			}
		}
		return items
	}
	before := check(original, 1000)
	updated := original.Insert(1000, 10000)
	check(updated, 1001)
	if after := check(original, 1000); !reflect.DeepEqual(before, after) {
		t.Fatal("input tree changed")
	}
	if len(ToArrayImpl(tuple, Empty)) != 0 {
		t.Fatal("empty map")
	}
	before[0] = [2]int{-1, -1}
	check(original, 1000)
}

func TestConcurrentPersistentMap(t *testing.T) {
	compare := func(a, b interface{}) interface{} {
		if a.(int) < b.(int) {
			return -1
		}
		if a.(int) > b.(int) {
			return 1
		}
		return 0
	}
	// Distinct but equivalent comparator closures must not replace one another
	// on a tree shared by independent computations.
	comparators := []func(interface{}, interface{}) interface{}{
		compare,
		func(a, b interface{}) interface{} { return compare(a, b) },
	}
	ordering := func(value interface{}) int { return value.(int) }
	just := func(value interface{}) interface{} { return value }
	first := func(a interface{}) func(interface{}) interface{} {
		return func(_ interface{}) interface{} { return a }
	}
	combine := func(a interface{}) func(interface{}) interface{} {
		return func(b interface{}) interface{} { return a.(int)*1000 + b.(int) }
	}
	var original interface{} = Empty
	for key := 0; key < 64; key++ {
		original = InsertImpl(compare, ordering, key, key*10, original)
	}
	var workers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		workers.Add(1)
		go func(worker int) {
			defer workers.Done()
			cmp := comparators[worker%len(comparators)]
			for iteration := 0; iteration < 100; iteration++ {
				key := iteration % 64
				if got := LookupImpl(just, nil, cmp, ordering, key, original); got != key*10 {
					t.Errorf("lookup %d: %v", key, got)
				}
				inserted := InsertImpl(cmp, ordering, 64+worker, worker, original)
				if SizeImpl(inserted) != 65 {
					t.Error("insert size")
				}
				if SizeImpl(DeleteImpl(cmp, ordering, key, original)) != 63 {
					t.Error("delete size")
				}
				if SizeImpl(UnionWithImpl(cmp, ordering, first, original, inserted)) != 65 {
					t.Error("union size")
				}
				// A colliding singleton exercises the size-aware Delete+Insert path.
				combined := UnionWithImpl(cmp, ordering, combine, Singleton(key)(worker+1), original)
				if SizeImpl(combined) != 64 {
					t.Error("singleton union size")
				}
				if got := LookupImpl(just, nil, cmp, ordering, key, combined); got != (worker+1)*1000+key*10 {
					t.Errorf("singleton union %d: %v", key, got)
				}
				if SizeImpl(IntersectionWithImpl(cmp, ordering, first, original, inserted)) != 64 {
					t.Error("intersection size")
				}
				if SizeImpl(DifferenceImpl(cmp, ordering, inserted, original)) != 1 {
					t.Error("difference size")
				}
			}
		}(worker)
	}
	workers.Wait()
	if SizeImpl(original) != 64 {
		t.Fatal("shared input was changed")
	}
	for key := 0; key < 64; key++ {
		if got := LookupImpl(just, nil, compare, ordering, key, original); got != key*10 {
			t.Fatalf("shared input key %d changed to %v", key, got)
		}
	}
}
