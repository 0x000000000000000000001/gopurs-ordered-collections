package mapbridge

import (
	r "gopurs/output/runtime"
	"reflect"
	"testing"
)

// Nonstandard tokens ensure the FFI keeps consulting fromOrdering.
var less, equal, greater = r.Int(37), r.Int(19), r.Int(53)
var fromOrdering = r.Func(func(value r.Value) r.Value {
	switch r.Unbox[int64](value) {
	case 37:
		return r.Int(-1)
	case 19:
		return r.Int(0)
	case 53:
		return r.Int(1)
	default:
		panic("unknown ordering token")
	}
})
var just = r.Func(func(value r.Value) r.Value { return value })
var nothing = r.Int(-999)
var combine = r.Func2(func(a, b r.Value) r.Value { return r.Int(10*r.Unbox[int64](a) + r.Unbox[int64](b)) })

func comparator(curried, reverse bool, calls *int) r.Value {
	fn := func(a, b r.Value) r.Value {
		if calls != nil {
			*calls++
		}
		x, y := r.Unbox[int64](a), r.Unbox[int64](b)
		if reverse {
			x, y = y, x
		}
		if x < y {
			return less
		}
		if x > y {
			return greater
		}
		return equal
	}
	if !curried {
		return r.Func2(fn)
	}
	return r.Func(func(a r.Value) r.Value { return r.Func(func(b r.Value) r.Value { return fn(a, b) }) })
}
func insert(cmp r.Value, key, value int64, tree r.Value) r.Value {
	return r.Apply5(_Gopurs_Map_InsertImpl, cmp, fromOrdering, r.Int(key), r.Int(value), tree)
}
func lookup(cmp r.Value, key int64, tree r.Value) int64 {
	return r.Unbox[int64](r.Apply6(_Gopurs_Map_LookupImpl, just, nothing, cmp, fromOrdering, r.Int(key), tree))
}
func keys(tree r.Value) []int64 {
	raw := Map_KeysImpl(tree)
	result := make([]int64, len(raw))
	for i, key := range raw {
		result[i] = r.Unbox[int64](r.Box(key))
	}
	return result
}
func TestGeneratedBridgeCustomOrderingAndPersistence(t *testing.T) {
	for _, curried := range []bool{false, true} {
		for _, reverse := range []bool{false, true} {
			calls := 0
			cmp := comparator(curried, reverse, &calls)
			tree := r.Box(Map_Empty)
			for _, key := range []int64{3, 1, 4, 2} {
				tree = insert(cmp, key, key, tree)
			}
			expected := []int64{1, 2, 3, 4}
			if reverse {
				expected = []int64{4, 3, 2, 1}
			}
			if got := keys(tree); !reflect.DeepEqual(got, expected) {
				t.Fatalf("curried=%t reverse=%t keys=%v", curried, reverse, got)
			}
			for key := int64(1); key <= 4; key++ {
				if got := lookup(cmp, key, tree); got != key {
					t.Fatalf("lookup %d=%d", key, got)
				}
			}
			if lookup(cmp, 9, tree) != -999 {
				t.Fatal("missing lookup")
			}
			updated := r.Apply6(_Gopurs_Map_InsertWithImpl, cmp, fromOrdering, combine, r.Int(2), r.Int(7), tree)
			if lookup(cmp, 2, updated) != 27 || lookup(cmp, 2, tree) != 2 {
				t.Fatal("insertWith or persistence")
			}
			deleted := r.Apply4(_Gopurs_Map_DeleteImpl, cmp, fromOrdering, r.Int(3), tree)
			if lookup(cmp, 3, deleted) != -999 || lookup(cmp, 3, tree) != 3 {
				t.Fatal("delete or persistence")
			}
			other := insert(cmp, 5, 5, insert(cmp, 2, 8, r.Box(Map_Empty)))
			union := r.Apply5(_Gopurs_Map_UnionWithImpl, cmp, fromOrdering, combine, tree, other)
			if lookup(cmp, 2, union) != 28 || lookup(cmp, 5, union) != 5 || Map_SizeImpl(union) != 5 {
				t.Fatal("unionWith")
			}
			intersection := r.Apply5(_Gopurs_Map_IntersectionWithImpl, cmp, fromOrdering, combine, tree, other)
			if !reflect.DeepEqual(keys(intersection), []int64{2}) || lookup(cmp, 2, intersection) != 28 {
				t.Fatal("intersectionWith")
			}
			difference := r.Apply4(_Gopurs_Map_DifferenceImpl, cmp, fromOrdering, tree, other)
			expectedDifference := []int64{1, 3, 4}
			if reverse {
				expectedDifference = []int64{4, 3, 1}
			}
			if !reflect.DeepEqual(keys(difference), expectedDifference) {
				t.Fatal("difference")
			}
			if !reflect.DeepEqual(keys(tree), expected) || lookup(cmp, 2, tree) != 2 || lookup(cmp, 2, other) != 8 {
				t.Fatal("input map changed")
			}
			if calls == 0 {
				t.Fatal("custom comparator was not used")
			}
		}
	}
}

var lookupSink int64

func BenchmarkMapLookup(b *testing.B) {
	for _, shape := range []struct {
		name    string
		curried bool
	}{{"runtime-arity2", false}, {"runtime-curried", true}} {
		b.Run(shape.name, func(b *testing.B) {
			cmp := comparator(shape.curried, false, nil)
			tree := r.Box(Map_Empty)
			for key := int64(0); key < 1024; key++ {
				tree = insert(cmp, key, key, tree)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lookupSink = lookup(cmp, int64(i%1024), tree)
			}
		})
	}
}
