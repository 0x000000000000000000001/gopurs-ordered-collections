#!/bin/bash
cat << 'INNER_EOF' >> src/Data/Map/Internal.go
func FindMinImpl(just func(interface{}) interface{}, nothing interface{}, m interface{}) interface{} {
	return nothing
}
func FindMaxImpl(just func(interface{}) interface{}, nothing interface{}, m interface{}) interface{} {
	return nothing
}
func CheckValid(m interface{}) bool {
	return true
}
INNER_EOF
