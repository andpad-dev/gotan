// Command wb-gains は、wide の色補正 (ホワイトバランス) の倍率を求める.
//
// 指定した時刻のフレームを ffmpeg で取り出し、元画像と grayworld 補正後の RGB 平均の比を取る.
// その中央値を全編共通の倍率とし、scripts/encode.sh の WIDE_COLOR に書く colorchannelmixer の形で出す.
//
// usage: wb-gains <動画> <秒>...
//
// 例:
//
//	go run ./cmd/wb-gains src/wide.mp4 60 600 1800
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/andpad-dev/gotan/tools/multicam-player/internal/num"
)

// 平均を取るときのフレームの大きさ. 色の平均を見るだけなので縮小して軽くする.
const w, h = 320, 180

// meanRGB は path の t 秒のフレームに vf を掛け、RGB それぞれの平均を返す.
func meanRGB(path, t, vf string) ([3]float64, error) {
	var m [3]float64
	cmd := exec.Command("ffmpeg", "-loglevel", "error", "-ss", t, "-i", path,
		"-frames:v", "1", "-vf", fmt.Sprintf("scale=%d:%d,%sformat=rgb24", w, h, vf),
		"-f", "rawvideo", "-")
	cmd.Stderr = os.Stderr
	raw, err := cmd.Output()
	if err != nil {
		return m, fmt.Errorf("ffmpeg (%s 秒): %w", t, err)
	}
	if len(raw) != w*h*3 {
		return m, fmt.Errorf("ffmpeg (%s 秒): フレームの大きさが %d バイト (%d バイトのはず)", t, len(raw), w*h*3)
	}
	for i, v := range raw {
		m[i%3] += float64(v)
	}
	for c := range m {
		m[c] /= w * h
	}
	return m, nil
}

// median は x の中央値を返す. 個数が偶数なら中央の 2 つの平均.
func median(x []float64) float64 {
	s := slices.Sorted(slices.Values(x))
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return (s[n/2-1] + s[n/2]) / 2
}

func main() {
	log.SetFlags(0)
	if len(os.Args) < 3 {
		log.Fatal("usage: wb-gains <動画> <秒>...")
	}
	path, times := os.Args[1], os.Args[2:]
	var gains [3][]float64
	rows := make([]string, len(times))
	for i, t := range times {
		a, err := meanRGB(path, t, "")
		if err != nil {
			log.Fatal(err)
		}
		b, err := meanRGB(path, t, "format=gbrpf32le,grayworld,")
		if err != nil {
			log.Fatal(err)
		}
		vals := make([]string, 3)
		for c := range 3 {
			g := b[c] / a[c]
			gains[c] = append(gains[c], g)
			vals[c] = num.Float(num.Round(g, 3))
		}
		rows[i] = "[" + strings.Join(vals, ", ") + "]"
	}
	fmt.Println("per-frame: [" + strings.Join(rows, ", ") + "]")
	fmt.Printf("colorchannelmixer=rr=%.3f:gg=%.3f:bb=%.3f\n", median(gains[0]), median(gains[1]), median(gains[2]))
}
