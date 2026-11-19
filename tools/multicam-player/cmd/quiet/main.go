// Command quiet は、区間内で 0.5 秒ごとの音量 (dBFS) を出し、カットに向いた静かな切れ目を探す.
//
// usage: quiet <wav> <開始秒> <終了秒>
//
// 例:
//
//	go run ./cmd/quiet work/wide.16k.wav 280 300
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andpad-dev/gotan/tools/multicam-player/internal/dsp"
	"github.com/andpad-dev/gotan/tools/multicam-player/internal/num"
	"github.com/andpad-dev/gotan/tools/multicam-player/internal/wav"
)

// step は 1 行にまとめる秒数.
const step = 0.5

func main() {
	log.SetFlags(0)
	if len(os.Args) != 4 {
		log.Fatal("usage: quiet <wav> <開始秒> <終了秒>")
	}
	r, err := num.ParseFloats(os.Args[2:])
	if err != nil {
		log.Fatal(err)
	}
	sr, x, err := wav.Read(os.Args[1])
	if err != nil {
		log.Fatalf("%s: %v", os.Args[1], err)
	}
	fs := float64(sr)
	for t := r[0]; t < r[1]; t += step {
		fmt.Printf("%7.1f %6.1f\n", t, dsp.DBFS(dsp.Slice(x, int(t*fs), int((t+step)*fs))))
	}
}
