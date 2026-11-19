// Package wav は、ffmpeg が書き出す 16bit PCM の WAV を読む.
//
// 対応するのは非圧縮の 16bit PCM だけで、複数チャンネルなら先頭チャンネルだけを返す.
// 音声の解析には `ffmpeg -i <入力> -vn -ac 1 -ar 16000 <出力>.wav` で作った WAV を渡す想定.
package wav

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// Read は path の WAV を読み、サンプリング周波数と、-1 以上 1 未満に正規化したサンプルを返す.
//
// 例:
//
//	sr, x, err := wav.Read("work/wide.16k.wav")
//	if err != nil {
//		return err
//	}
//	fmt.Println(sr, len(x)) // 16000 51876320
func Read(path string) (int, []float64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}
	return Parse(b)
}

// Parse は WAV のバイト列を読み、サンプリング周波数と正規化したサンプルを返す.
func Parse(b []byte) (int, []float64, error) {
	if len(b) < 12 || string(b[0:4]) != "RIFF" || string(b[8:12]) != "WAVE" {
		return 0, nil, errors.New("wav: RIFF/WAVE ヘッダがない")
	}
	var (
		sr, channels int
		haveFmt      bool
	)
	le := binary.LittleEndian
	// fmt と data 以外のチャンク (ffmpeg が書く LIST など) は読み飛ばす.
	for p := 12; p+8 <= len(b); {
		id := string(b[p : p+4])
		size := int(le.Uint32(b[p+4 : p+8]))
		body := b[p+8:]
		if size > len(body) {
			size = len(body)
		}
		body = body[:size]
		switch id {
		case "fmt ":
			if size < 16 {
				return 0, nil, errors.New("wav: fmt チャンクが短い")
			}
			format := le.Uint16(body[0:2])
			channels = int(le.Uint16(body[2:4]))
			sr = int(le.Uint32(body[4:8]))
			bits := le.Uint16(body[14:16])
			// 0xFFFE (WAVE_FORMAT_EXTENSIBLE) は、ffmpeg が 3 チャンネル以上で書く形式.
			if (format != 1 && format != 0xFFFE) || bits != 16 || channels < 1 {
				return 0, nil, fmt.Errorf("wav: 16bit PCM ではない (format=%d, bits=%d)", format, bits)
			}
			haveFmt = true
		case "data":
			if !haveFmt {
				return 0, nil, errors.New("wav: data チャンクが fmt チャンクより前にある")
			}
			frame := 2 * channels
			n := size / frame
			x := make([]float64, n)
			for i := range n {
				x[i] = float64(int16(le.Uint16(body[i*frame:]))) / 32768
			}
			return sr, x, nil
		}
		// チャンクは 2 バイト境界にそろえて並ぶ.
		p += 8 + size + size%2
	}
	return 0, nil, errors.New("wav: data チャンクがない")
}
