// Command clap は、手拍子の立ち上がりの時刻をサンプル単位で求める.
//
// おおよその秒の前後 0.3 秒を調べ、区間の最大振幅の 30% を初めて超えたサンプルを立ち上がりとする.
// おおよその秒は cmd/sync が出す手拍子の候補から選ぶ.
//
// usage: clap <wav> <おおよその秒> [<wav> <秒> ...]
//
// 例:
//
//	go run ./cmd/clap work/wide.16k.wav 10.3 work/player1.16k.wav 1.3
package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/andpad-dev/gotan/tools/multicam-player/internal/dsp"
	"github.com/andpad-dev/gotan/tools/multicam-player/internal/wav"
)

func main() {
	log.SetFlags(0)
	args := os.Args[1:]
	if len(args) == 0 || len(args)%2 != 0 {
		log.Fatal("usage: clap <wav> <おおよその秒> [<wav> <秒> ...]")
	}
	for i := 0; i < len(args); i += 2 {
		path := args[i]
		approx, err := strconv.ParseFloat(args[i+1], 64)
		if err != nil {
			log.Fatal(err)
		}
		sr, x, err := wav.Read(path)
		if err != nil {
			log.Fatalf("%s: %v", path, err)
		}
		s := max(0, int((approx-0.3)*float64(sr)))
		fmt.Printf("%s\t%.4f\n", path, float64(s+onset(dsp.Slice(x, s, s+int(0.6*float64(sr)))))/float64(sr))
	}
}

// onset は、seg の最大振幅の 30% を初めて超えたサンプルの位置を返す. 見つからなければ 0.
func onset(seg []float64) int {
	var peak float64
	for _, v := range seg {
		peak = math.Max(peak, math.Abs(v))
	}
	for i, v := range seg {
		if math.Abs(v) > peak*0.3 {
			return i
		}
	}
	return 0
}
