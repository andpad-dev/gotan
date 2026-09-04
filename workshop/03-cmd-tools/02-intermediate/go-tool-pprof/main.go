package main

import "fmt"

var result uint64

func transform(seed uint64) uint64 {
	for range 4_000_000 {
		seed = seed*2862933555777941757 + 3037000493
	}
	return seed
}

func main() {
	fmt.Println(transform(1))
}
