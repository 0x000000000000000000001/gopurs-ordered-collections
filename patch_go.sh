#!/bin/bash
sed -i.bak 's/return f(acc)(key)(value)/fmt.Printf("FoldlImpl value: %#v\\n", value); return f(acc)(key)(value)/' src/Data/Map/Internal.go
sed -i.bak 's/import (/import (\n\t"fmt"/' src/Data/Map/Internal.go
