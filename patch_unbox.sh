#!/bin/bash
sed -i.bak 's/v := \*(\*any)(ptr)/if uintptr(ptr) < 4096 { panic(fmt.Sprintf("Invalid ptr in Unbox[int]: %v", ptr)) }\n\t\t\t\tv := \*(\*any)(ptr)/g' output/gopurs_runtime/runtime.go
