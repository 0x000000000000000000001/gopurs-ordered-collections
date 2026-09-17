package Data_Map_Internal

import (
	"reflect"
	"sync"
	"testing"
)

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
	compare := func(a interface{}) func(interface{}) interface{} {
		return func(b interface{}) interface{} {
			if a.(int) < b.(int) {
				return -1
			}
			if a.(int) > b.(int) {
				return 1
			}
			return 0
		}
	}
	// Distinct but equivalent comparator closures must not replace one another
	// on a tree shared by independent computations.
	comparators := []func(interface{}) func(interface{}) interface{}{
		compare,
		func(a interface{}) func(interface{}) interface{} {
			return func(b interface{}) interface{} { return compare(a)(b) }
		},
	}
	ordering := func(value interface{}) int { return value.(int) }
	just := func(value interface{}) interface{} { return value }
	first := func(a interface{}) func(interface{}) interface{} {
		return func(_ interface{}) interface{} { return a }
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
