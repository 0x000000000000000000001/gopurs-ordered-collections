package main

import (
	"fmt"
	"unsafe"
)

type Value struct {
	Type      uint8
	IntVal    int64
	UnsafePtr unsafe.Pointer
}

func Func(f func(Value) Value) Value {
	return Value{Type: 1, UnsafePtr: *(*unsafe.Pointer)(unsafe.Pointer(&f))}
}

func Apply(f Value, arg Value) Value {
	return (*(*func(Value) Value)(unsafe.Pointer(&f.UnsafePtr)))(arg)
}

var _Gopurs_ToNumber = Func(func(arg0 Value) Value {
	fmt.Printf("arg: Type=%d, IntVal=%d, UnsafePtr=%v\n", arg0.Type, arg0.IntVal, arg0.UnsafePtr)
	return arg0
})

func main() {
	arg := Value{Type: 6, IntVal: 6114, UnsafePtr: nil}
	Apply(_Gopurs_ToNumber, arg)
}
