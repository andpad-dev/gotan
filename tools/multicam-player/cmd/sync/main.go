// Command sync は、3 本の音声からずれ量を測る.
//
// wide を基準に、player の音声が何秒遅れて始まっているかを求める.
// 冒頭の手拍子の候補と、冒頭・末尾 5 分の相互相関の両方を出し、ドリフトの有無を確かめる.
// 入力は 16kHz モノラルの WAV.
//
// usage: sync <基準 wav> <wav>...
//
// 例:
//
//	go run ./cmd/sync work/wide.16k.wav work/player1.16k.wav work/player2.16k.wav
package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"slices"

	"github.com/andpad-dev/gotan/tools/multicam-player/internal/dsp"
	"github.com/andpad-dev/gotan/tools/multicam-player/internal/num"
	"github.com/andpad-dev/gotan/tools/multicam-player/internal/wav"
)

const sr = 16000

func load(path string) []float64 {
	rate, x, err := wav.Read(path)
	if err != nil {
		log.Fatalf("%s: %v", path, err)
	}
	if rate != sr {
		log.Fatalf("%s: サンプリング周波数が %d Hz (%d Hz にする)", path, rate, sr)
	}
	return x
}

// candidate は手拍子の候補. JSON では [秒, 立ち上がりの鋭さ, 音量] の配列にする.
type candidate struct {
	frame        int
	score, level float64
}

func (c candidate) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float64{float64(c.frame) / 100, c.score, c.level})
}

// clapCandidates は、冒頭 searchSec 秒から立ち上がりの鋭い大きなピークを n 個、時刻順に返す.
//
// 10ms ごとの RMS を音量とし、直前 0.5 秒 (自身を含む 50 フレーム) の平均との比を鋭さとする.
// 鋭い順に選び、選んだ候補と 1 秒以内のものは同じ音とみなして飛ばす.
func clapCandidates(x []float64, searchSec, n int) []candidate {
	seg := dsp.Slice(x, 0, searchSec*sr)
	const hop = sr / 100
	// 各フレームの中心 i = f*hop の前後 hop サンプル (i-80 から i+79) の平均二乗. 範囲外は 0 とみなす.
	nf := (len(seg) + hop - 1) / hop
	env := make([]float64, nf)
	for f := range env {
		var s float64
		for _, v := range dsp.Slice(seg, f*hop-hop/2, f*hop+hop/2) {
			s += v * v
		}
		env[f] = math.Sqrt(s / hop)
	}
	score := make([]float64, nf)
	var sum float64
	for f := range env {
		sum += env[f]
		if f >= 50 {
			sum -= env[f-50]
		}
		score[f] = env[f] / (sum/50 + 1e-4)
	}

	order := make([]int, nf)
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int { return cmp.Compare(score[b], score[a]) })
	var out []candidate
	for _, f := range order {
		if !slices.ContainsFunc(out, func(c candidate) bool { return abs(f-c.frame) <= 100 }) {
			out = append(out, candidate{f, score[f], env[f]})
		}
		if len(out) == n {
			break
		}
	}
	slices.SortFunc(out, func(a, b candidate) int { return cmp.Compare(a.frame, b.frame) })
	return out
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// xcorrOffset は ref の refStart 秒からの length 秒を、x の xStart 秒の前後 maxLag 秒でずらして比べ、
// x が ref より何秒遅れているかと、正規化した相関の強さ (0 から 1) を返す.
func xcorrOffset(ref, x []float64, refStart, xStart, length, maxLag float64) (float64, float64) {
	a := dsp.Slice(ref, int(refStart*sr), int((refStart+length)*sr))
	b := dsp.Slice(x, int((xStart-maxLag)*sr), int((xStart+length+maxLag)*sr))
	c := dsp.Correlate(b, a)
	k := dsp.ArgMaxAbs(c)
	if k < 0 {
		log.Fatal("比べる区間が音声の長さを超えている")
	}
	peak := math.Abs(c[k]) / (dsp.Norm(a)*dsp.Norm(b[k:k+len(a)]) + 1e-9)
	lag := (xStart - maxLag) + float64(k)/sr - refStart
	return lag, peak
}

// field は出力 JSON のキーと値. map ではキーの並びが保てないので、並び順どおりに書き出すために使う.
type field struct {
	key   string
	value any
}

// object は、キーを並べた順に書き出す JSON オブジェクト.
type object []field

func (o object) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, f := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		k, err := json.Marshal(f.key)
		if err != nil {
			return nil, err
		}
		v, err := json.Marshal(f.value)
		if err != nil {
			return nil, err
		}
		b.Write(k)
		b.WriteByte(':')
		b.Write(v)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func main() {
	log.SetFlags(0)
	if len(os.Args) < 3 {
		log.Fatal("usage: sync <基準 wav> <wav>...")
	}
	refPath, paths := os.Args[1], os.Args[2:]
	ref := load(refPath)
	claps := object{{refPath, clapCandidates(ref, 600, 5)}}
	var pairs object
	for _, p := range paths {
		x := load(p)
		claps = append(claps, field{p, clapCandidates(x, 600, 5)})
		// 粗い推定: 冒頭側の 5 分を ±60 秒の範囲で比べる.
		head, hpk := xcorrOffset(ref, x, 120, 120, 300, 60)
		// 末尾側: 同じずれを仮定して 5 分を ±5 秒で比べ、ドリフトを測る.
		tailRef := float64(len(ref))/sr - 700
		tail, tpk := xcorrOffset(ref, x, tailRef, tailRef+head, 300, 5)
		pairs = append(pairs, field{p, object{
			{"head_offset_sec", num.Round(head, 4)},
			{"head_peak", num.Round(hpk, 3)},
			{"tail_offset_sec", num.Round(tail, 4)},
			{"tail_peak", num.Round(tpk, 3)},
			{"drift_sec", num.Round(tail-head, 4)},
			{"duration_sec", float64(len(x)) / sr},
		}})
	}
	result := object{
		{"reference", refPath},
		{"clap", claps},
		{"pairs", pairs},
		{"reference_duration_sec", float64(len(ref)) / sr},
	}
	raw, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}
	var out bytes.Buffer
	if err := json.Indent(&out, raw, "", "  "); err != nil {
		log.Fatal(err)
	}
	fmt.Println(out.String())
}
