// Command check-sync は、変換後のファイル同士のずれを、指定した区間で測る. 0 に近ければ同期できている.
//
// 基準の区間を、比べる音声の ±0.5 秒の範囲でずらして相互相関を取り、最も重なる位置をずれとする.
//
// usage: check-sync <開始秒> <終了秒> <基準 wav> <wav>...
//
// 例:
//
//	go run ./cmd/check-sync 60 120 work/final/wide.wav work/final/player1.wav work/final/player2.wav
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/andpad-dev/gotan/tools/multicam-player/internal/dsp"
	"github.com/andpad-dev/gotan/tools/multicam-player/internal/wav"
)

// maxLag は探すずれの幅 (秒).
const maxLag = 0.5

func main() {
	log.SetFlags(0)
	if len(os.Args) < 5 {
		log.Fatal("usage: check-sync <開始秒> <終了秒> <基準 wav> <wav>...")
	}
	a0, err0 := strconv.ParseFloat(os.Args[1], 64)
	a1, err1 := strconv.ParseFloat(os.Args[2], 64)
	if err0 != nil || err1 != nil {
		log.Fatal("開始秒と終了秒は数で渡す")
	}
	refPath, paths := os.Args[3], os.Args[4:]
	sr, ref, err := wav.Read(refPath)
	if err != nil {
		log.Fatalf("%s: %v", refPath, err)
	}
	fs := float64(sr)
	a := dsp.Slice(ref, int(a0*fs), int(a1*fs))
	for _, p := range paths {
		_, x, err := wav.Read(p)
		if err != nil {
			log.Fatalf("%s: %v", p, err)
		}
		b := dsp.Slice(x, int((a0-maxLag)*fs), int((a1+maxLag)*fs))
		k := dsp.ArgMaxAbs(dsp.Correlate(b, a))
		if k < 0 {
			log.Fatalf("%s: 区間が音声の長さを超えている", p)
		}
		fmt.Printf("%.0f-%.0fs\t%s\t%+.1f ms\n", a0, a1, p, (float64(k)/fs-maxLag)*1000)
	}
}
