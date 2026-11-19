package num

import "testing"

func TestRound(t *testing.T) {
	tests := []struct {
		x      float64
		digits int
		want   float64
	}{
		{78.46, 1, 78.5},
		{-9.07342, 4, -9.0734},
		// 0.25 は 2 進数でちょうど表せるので、偶数への丸めで 0.2 になる (Python の round と同じ).
		{0.25, 1, 0.2},
		// 2.675 は 2 進数では 2.67499... なので 2.67 になる.
		{2.675, 2, 2.67},
	}
	for _, tt := range tests {
		if got := Round(tt.x, tt.digits); got != tt.want {
			t.Errorf("Round(%v, %d) = %v, want %v", tt.x, tt.digits, got, tt.want)
		}
	}
}

func TestFloat(t *testing.T) {
	tests := []struct {
		x    float64
		want string
	}{
		{0, "0.0"},
		{78.5, "78.5"},
		{3069, "3069.0"},
		{-1.25, "-1.25"},
	}
	for _, tt := range tests {
		if got := Float(tt.x); got != tt.want {
			t.Errorf("Float(%v) = %q, want %q", tt.x, got, tt.want)
		}
	}
}

func TestParseFloats(t *testing.T) {
	got, err := ParseFloats([]string{"11.5", "284"})
	if err != nil || len(got) != 2 || got[0] != 11.5 || got[1] != 284 {
		t.Fatalf("ParseFloats = %v, %v", got, err)
	}
	if _, err := ParseFloats([]string{"x"}); err == nil {
		t.Fatal("数でない文字列でエラーにならなかった")
	}
}
