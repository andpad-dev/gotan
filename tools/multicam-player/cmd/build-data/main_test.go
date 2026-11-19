package main

import "testing"

func TestToOutput(t *testing.T) {
	segs := []segment{{11.5, 284}, {297, 3235.5}}
	tests := []struct {
		name string
		t    float64
		want float64
	}{
		{"最初の区間より前は 0", 5, 0},
		{"最初の区間の中", 100, 88.5},
		{"切った区間の中は次の区間の頭", 290, 272.5},
		{"2 つ目の区間の中", 300, 275.5},
		{"最後の区間より後は末尾", 4000, 272.5 + 2938.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toOutput(segs, tt.t); got != tt.want {
				t.Errorf("toOutput(%v) = %v, want %v", tt.t, got, tt.want)
			}
		})
	}
}
