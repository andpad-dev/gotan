package wav

import (
	"encoding/binary"
	"testing"
)

// build は 16bit PCM の WAV を組み立てる. extra は fmt と data の間に挟むチャンク.
func build(sr, channels int, samples []int16, extra []byte) []byte {
	le := binary.LittleEndian
	var b []byte
	b = append(b, "RIFF\x00\x00\x00\x00WAVE"...)
	b = append(b, "fmt "...)
	b = le.AppendUint32(b, 16)
	b = le.AppendUint16(b, 1)
	b = le.AppendUint16(b, uint16(channels))
	b = le.AppendUint32(b, uint32(sr))
	b = le.AppendUint32(b, uint32(sr*channels*2))
	b = le.AppendUint16(b, uint16(channels*2))
	b = le.AppendUint16(b, 16)
	b = append(b, extra...)
	b = append(b, "data"...)
	b = le.AppendUint32(b, uint32(len(samples)*2))
	for _, s := range samples {
		b = le.AppendUint16(b, uint16(s))
	}
	return b
}

func TestParse(t *testing.T) {
	// ffmpeg は fmt と data の間に LIST チャンクを書く. 奇数長のチャンクは 1 バイト埋める.
	list := append([]byte("LIST\x03\x00\x00\x00abc"), 0)
	tests := []struct {
		name     string
		in       []byte
		wantSR   int
		wantVals []float64
	}{
		{"モノラル", build(16000, 1, []int16{0, 16384, -32768}, nil), 16000, []float64{0, 0.5, -1}},
		{"LIST チャンク付き", build(16000, 1, []int16{16384}, list), 16000, []float64{0.5}},
		{"ステレオは先頭チャンネル", build(48000, 2, []int16{16384, -1, -16384, -1}, nil), 48000, []float64{0.5, -0.5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr, x, err := Parse(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if sr != tt.wantSR {
				t.Errorf("sr = %d, want %d", sr, tt.wantSR)
			}
			if len(x) != len(tt.wantVals) {
				t.Fatalf("x = %v, want %v", x, tt.wantVals)
			}
			for i := range x {
				if x[i] != tt.wantVals[i] {
					t.Fatalf("x = %v, want %v", x, tt.wantVals)
				}
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	float32WAV := build(16000, 1, []int16{0}, nil)
	binary.LittleEndian.PutUint16(float32WAV[34:36], 32)
	tests := []struct {
		name string
		in   []byte
	}{
		{"空", nil},
		{"RIFF でない", []byte("RIFX\x00\x00\x00\x00WAVE")},
		{"16bit でない", float32WAV},
		{"data がない", build(16000, 1, nil, nil)[:36]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := Parse(tt.in); err == nil {
				t.Fatal("エラーにならなかった")
			}
		})
	}
}
