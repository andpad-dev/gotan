package main

import (
	"fmt"
	"sort"
)

func main() {
	xs := []int{3, 1, 2}
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
	for i := 0; i < len(xs); i++ {
		fmt.Println(i, xs[i])
	}
}
