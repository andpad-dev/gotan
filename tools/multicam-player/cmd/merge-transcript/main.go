// Command merge-transcript は、player1 と player2 の文字起こしを wide の時刻軸に揃えて 1 本にする.
//
// player の音声の t 秒は、wide の t + (手拍子の時刻差) 秒にあたる.
// 入力は work/player1.transcript.json と work/player2.transcript.json (cmd/transcribe の出力).
//
// usage: merge-transcript <out.tsv>
//
// 例:
//
//	go run ./cmd/merge-transcript work/transcript.tsv
package main

import (
	"bufio"
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"slices"
	"strings"
)

// source は 1 本分の文字起こしと、wide の時刻に直すために足す秒数.
// shift は「wide の手拍子の時刻 − player の手拍子の時刻」で、cmd/clap の出力から求めた.
type source struct {
	tag, path string
	shift     float64
}

var sources = []source{
	{"P1", "work/player1.transcript.json", 10.3101 - 1.2606},
	{"P2", "work/player2.transcript.json", 10.3101 - 1.1711},
}

type row struct {
	t         float64
	tag, text string
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Fatal("usage: merge-transcript <out.tsv>")
	}
	var rows []row
	for _, s := range sources {
		raw, err := os.ReadFile(s.path)
		if err != nil {
			log.Fatal(err)
		}
		var segs []struct {
			Start float64 `json:"start"`
			Text  string  `json:"text"`
		}
		if err := json.Unmarshal(raw, &segs); err != nil {
			log.Fatalf("%s: %v", s.path, err)
		}
		for _, g := range segs {
			rows = append(rows, row{g.Start + s.shift, s.tag, strings.TrimSpace(g.Text)})
		}
	}
	slices.SortFunc(rows, func(a, b row) int {
		return cmp.Or(cmp.Compare(a.t, b.t), cmp.Compare(a.tag, b.tag), cmp.Compare(a.text, b.text))
	})

	f, err := os.Create(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	w := bufio.NewWriter(f)
	for _, r := range rows {
		// bufio.Writer は書き込みのエラーを覚えておき、Flush で返すので、ここでは見ない.
		_, _ = fmt.Fprintf(w, "%02d:%05.2f\t%s\t%s\n", int(math.Floor(r.t/60)), math.Mod(r.t, 60), r.tag, r.text)
	}
	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(rows))
}
