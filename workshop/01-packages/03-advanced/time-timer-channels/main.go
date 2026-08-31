package main

import (
	"fmt"
	"time"
)

func main() {
	timer := time.NewTimer(time.Hour)
	fmt.Println("channel capacity:", cap(timer.C))
	fmt.Println("stop before firing:", timer.Stop())

	timer.Reset(time.Millisecond)
	<-timer.C
	fmt.Println("reset timer fired")
}
