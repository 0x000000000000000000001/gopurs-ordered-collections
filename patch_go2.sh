#!/bin/bash
git checkout src/Data/Map/Internal.go
sed -i.bak 's/return f(acc)(key)(value)/println("FoldlImpl value type:"); return f(acc)(key)(value)/' src/Data/Map/Internal.go
