// Command loudness は、30 秒ごとの音量 (dBFS) を並べ、収録の頭と尻で人が話しているかを見る.
//
// 複数の WAV を渡すと列に並べ、短い方の長さまで出す.
//
// usage: loudness <wav>...
//
// 例:
//
//	go run ./cmd/loudness work/wide.16k.wav work/player1.16k.wav
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/andpad-dev/gotan/tools/multicam-player/internal/dsp"
	"github.com/andpad-dev/gotan/tools/multicam-player/internal/wav"
)

// win は 1 行にまとめる秒数.
const win = 30

func main() {
	log.SetFlags(0)
	paths := os.Args[1:]
	if len(paths) == 0 {
		log.Fatal("usage: loudness <wav>...")
	}
	cols := make([][]float64, len(paths))
	rows := -1
	for i, p := range paths {
		sr, x, err := wav.Read(p)
		if err != nil {
			log.Fatalf("%s: %v", p, err)
		}
		n := sr * win
		for j := 0; j+n <= len(x); j += n {
			cols[i] = append(cols[i], dsp.DBFS(x[j:j+n]))
		}
		if rows < 0 || len(cols[i]) < rows {
			rows = len(cols[i])
		}
	}

	names := make([]string, len(paths))
	for i, p := range paths {
		name, _, _ := strings.Cut(filepath.Base(p), ".")
		if r := []rune(name); len(r) > 7 {
			name = string(r[:7])
		}
		names[i] = fmt.Sprintf("%7s", name)
	}
	fmt.Println("time   " + strings.Join(names, "  "))
	for r := range rows {
		vals := make([]string, len(cols))
		for i, c := range cols {
			vals[i] = fmt.Sprintf("%7.1f", c[r])
		}
		fmt.Printf("%3d:%02d %s\n", r*win/60, r*win%60, strings.Join(vals, "  "))
	}
}
