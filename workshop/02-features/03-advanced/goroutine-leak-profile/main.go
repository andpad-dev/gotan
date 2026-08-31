package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

type result struct {
	value int
	err   error
}

func processWorkItem(id int) (int, error) {
	if id == 2 {
		return 0, errors.New("boom")
	}
	time.Sleep(100 * time.Millisecond)
	return id * 10, nil
}

func processWorkItems(ids []int) ([]int, error) {
	ch := make(chan result)
	for _, id := range ids {
		go func(id int) {
			v, err := processWorkItem(id)
			ch <- result{v, err}
		}(id)
	}
	var out []int
	for range ids {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		out = append(out, r.value)
	}
	return out, nil
}

func main() {
	fmt.Println("go version:", runtime.Version())
	_, err := processWorkItems([]int{0, 1, 2, 3, 4})
	fmt.Fprintln(os.Stderr, "processWorkItems err:", err)

	time.Sleep(300 * time.Millisecond)

	fmt.Println("=== goroutine profile ===")
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	fmt.Println("=== goroutineleak profile ===")
	pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
}
