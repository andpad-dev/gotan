// Command build-data は、チャプター案 (wide の時刻) を再生用ファイルの時刻に直し、data.js を書き出す.
//
// 文字起こしの区切りは話し始めの位置なので、見る人が文脈をつかめるように
// チャプターを lead 秒だけ前倒しする. 一覧の表示、シークバーの印、ジャンプ先は同じ時刻になる.
//
// usage: build-data <chapters.json> <out data.js> <開始秒> <終了秒> [<開始秒> <終了秒> ...]
//
// 例:
//
//	go run ./cmd/build-data work/chapters.draft.json data.js 11.5 284.0 297.0 3235.5
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/andpad-dev/gotan/tools/multicam-player/internal/num"
)

// lead はチャプターを前倒しする秒数.
const lead = 3.0

// segment は、wide の時刻で表した残す区間.
type segment struct{ start, end float64 }

type video struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	File  string `json:"file"`
}

type chapter struct {
	T     num.JSONFloat `json:"t"`
	Title string        `json:"title"`
}

// data は data.js の window.GOTAN に入れる値. フィールドの並びがそのまま出力の並びになる.
type data struct {
	Title string `json:"title"`
	// MediaBase は、動画を URL で置くときの置き場所 (例: "https://example.com/gotan/"). 空なら media/ から読む.
	MediaBase string  `json:"mediaBase"`
	Videos    []video `json:"videos"`
	// Wipe はワイプに映す映像の id.
	Wipe         string         `json:"wipe"`
	DefaultAudio string         `json:"defaultAudio"`
	Offsets      map[string]int `json:"offsets"`
	Chapters     []chapter      `json:"chapters"`
}

const header = `// プレイヤーに渡すデータ. file:// で開いても読めるように、JSON ではなく JS として置く.
// go run ./cmd/build-data で生成する. 再生用ファイルの段階で同期済みなので offsets は 0.
// ?dev=1 で微調整した値は offsets に書き戻す.
`

func main() {
	log.SetFlags(0)
	if len(os.Args) < 5 || (len(os.Args)-3)%2 != 0 {
		log.Fatal("usage: build-data <chapters.json> <out data.js> <開始秒> <終了秒> [<開始秒> <終了秒> ...]")
	}
	src, out := os.Args[1], os.Args[2]
	vals, err := num.ParseFloats(os.Args[3:])
	if err != nil {
		log.Fatal(err)
	}
	var segs []segment
	for i := 0; i < len(vals); i += 2 {
		segs = append(segs, segment{vals[i], vals[i+1]})
	}

	raw, err := os.ReadFile(src)
	if err != nil {
		log.Fatal(err)
	}
	var drafts []struct {
		T     float64 `json:"t"`
		Title string  `json:"title"`
	}
	if err := json.Unmarshal(raw, &drafts); err != nil {
		log.Fatalf("%s: %v", src, err)
	}
	chapters := make([]chapter, len(drafts))
	for i, c := range drafts {
		t := num.Round(math.Max(0, toOutput(segs, c.T)-lead), 1)
		chapters[i] = chapter{T: num.JSONFloat(t), Title: c.Title}
	}

	d := data{
		Title: "Go探無比 プレイ映像: slog.Value はなぜ == で比較できないのか",
		Videos: []video{
			{"player1", "プレイヤー 1 の画面", "player1.mp4"},
			{"player2", "プレイヤー 2 の画面", "player2.mp4"},
			{"wide", "2 人の様子", "wide.mp4"},
		},
		Wipe:         "wide",
		DefaultAudio: "wide",
		Offsets:      map[string]int{"player1": 0, "player2": 0, "wide": 0},
		Chapters:     chapters,
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(d); err != nil {
		log.Fatal(err)
	}
	js := header + "window.GOTAN = " + string(bytes.TrimRight(buf.Bytes(), "\n")) + ";\n"
	if err := os.WriteFile(out, []byte(js), 0o644); err != nil {
		log.Fatal(err)
	}
	for _, c := range chapters {
		t := float64(c.T)
		fmt.Printf("%2d:%04.1f  %s\n", int(t/60), math.Mod(t, 60), c.Title)
	}
}

// toOutput は wide の時刻 t を、残した区間をつないだ時刻に直す.
// 切った区間内の時刻は次の区間の頭に寄せ、最後の区間より後の時刻は末尾に寄せる.
//
// 例: 区間が [11.5, 284] と [297, 3235.5] のとき、toOutput(segs, 300) は 272.5+3 = 275.5.
func toOutput(segs []segment, t float64) float64 {
	var acc float64
	for _, s := range segs {
		if t < s.end {
			return acc + math.Max(0, t-s.start)
		}
		acc += s.end - s.start
	}
	return acc
}
