package dsp

import (
	"math"
	"math/rand/v2"
	"testing"
)

// naiveCorrelate は Correlate の定義どおりに二重ループで計算する.
func naiveCorrelate(b, a []float64) []float64 {
	c := make([]float64, len(b)-len(a)+1)
	for k := range c {
		for j, v := range a {
			c[k] += b[k+j] * v
		}
	}
	return c
}

func TestCorrelate(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	tests := []struct {
		name   string
		nb, na int
	}{
		{"同じ長さ", 8, 8},
		{"2 の累乗ちょうど", 64, 10},
		{"2 の累乗でない", 1000, 137},
		{"a が 1 サンプル", 33, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := make([]float64, tt.nb)
			a := make([]float64, tt.na)
			for i := range b {
				b[i] = r.Float64()*2 - 1
			}
			for i := range a {
				a[i] = r.Float64()*2 - 1
			}
			got, want := Correlate(b, a), naiveCorrelate(b, a)
			if len(got) != len(want) {
				t.Fatalf("len = %d, want %d", len(got), len(want))
			}
			for k := range want {
				if math.Abs(got[k]-want[k]) > 1e-9 {
					t.Fatalf("c[%d] = %v, want %v", k, got[k], want[k])
				}
			}
		})
	}
}

func TestCorrelateExample(t *testing.T) {
	got := Correlate([]float64{0, 1, 2, 3}, []float64{1, 2})
	want := []float64{2, 5, 8}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestCorrelateTooLong(t *testing.T) {
	if got := Correlate([]float64{1}, []float64{1, 2}); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

// ずらした信号を相関で探すと、ずらした量が見つかる.
func TestCorrelateFindsShift(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	x := make([]float64, 5000)
	for i := range x {
		x[i] = r.NormFloat64()
	}
	const shift = 1234
	a := x[shift : shift+500]
	if got := ArgMaxAbs(Correlate(x, a)); got != shift {
		t.Fatalf("ArgMaxAbs = %d, want %d", got, shift)
	}
}

func TestSlice(t *testing.T) {
	x := []float64{0, 1, 2, 3}
	tests := []struct {
		lo, hi int
		want   int
	}{
		{1, 3, 2},
		{-5, 2, 2},
		{2, 100, 2},
		{5, 9, 0},
		{3, 1, 0},
	}
	for _, tt := range tests {
		if got := len(Slice(x, tt.lo, tt.hi)); got != tt.want {
			t.Errorf("len(Slice(x, %d, %d)) = %d, want %d", tt.lo, tt.hi, got, tt.want)
		}
	}
}

func TestDBFS(t *testing.T) {
	if got := DBFS([]float64{1, -1}); math.Abs(got) > 1e-6 {
		t.Errorf("DBFS(full scale) = %v, want 0", got)
	}
	if got := DBFS([]float64{0.1, -0.1}); math.Abs(got+20) > 1e-6 {
		t.Errorf("DBFS(0.1) = %v, want -20", got)
	}
	if got := DBFS(nil); !math.IsNaN(got) {
		t.Errorf("DBFS(nil) = %v, want NaN", got)
	}
}
