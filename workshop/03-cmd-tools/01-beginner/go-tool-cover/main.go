package main

import "fmt"

func calculate(n int) int {
	if n <= 0 {
		return 0
	}
	if n >= 10 {
		return n * 8
	}
	return n * 10
}

func main() {
	fmt.Println(calculate(3))
}
