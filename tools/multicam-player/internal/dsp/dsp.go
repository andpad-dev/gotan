// Package dsp は、音声の時刻合わせに使う信号処理 (FFT による相互相関と音量) をまとめる.
package dsp

import (
	"math"
	"math/bits"
	"math/cmplx"
)

// fft は x を長さ len(x) (2 の累乗) の離散フーリエ変換で置き換える.
// inverse が true なら逆変換し、1/len(x) の正規化も行う.
func fft(x []complex128, inverse bool) {
	n := len(x)
	if n <= 1 {
		return
	}
	shift := 64 - bits.TrailingZeros(uint(n))
	for i := range n {
		if j := int(bits.Reverse64(uint64(i)) >> shift); i < j {
			x[i], x[j] = x[j], x[i]
		}
	}
	sign := -1.0
	if inverse {
		sign = 1
	}
	// 回転因子は最大の段の分だけ作り、小さい段では間引いて使う.
	w := make([]complex128, n/2)
	for k := range w {
		s, c := math.Sincos(sign * 2 * math.Pi * float64(k) / float64(n))
		w[k] = complex(c, s)
	}
	for size := 2; size <= n; size <<= 1 {
		half, stride := size/2, n/size
		for start := 0; start < n; start += size {
			for k := range half {
				t := w[k*stride] * x[start+k+half]
				x[start+k+half] = x[start+k] - t
				x[start+k] += t
			}
		}
	}
	if inverse {
		inv := complex(1/float64(n), 0)
		for i := range x {
			x[i] *= inv
		}
	}
}

// Correlate は、a を b の上でずらしながら掛け合わせた和を、a が b に収まる位置ごとに返す.
//
// 戻り値の c[k] は sum_j b[k+j]*a[j] で、長さは len(b)-len(a)+1.
// scipy.signal.fftconvolve(b, a[::-1], mode="valid") と同じ値になる.
// len(a) > len(b) のときは nil を返す.
//
// 例:
//
//	c := dsp.Correlate([]float64{0, 1, 2, 3}, []float64{1, 2})
//	fmt.Println(c) // [2 5 8]
func Correlate(b, a []float64) []float64 {
	if len(a) == 0 || len(a) > len(b) {
		return nil
	}
	// k+j は常に len(b) 未満なので、FFT の長さが len(b) 以上あれば循環の折り返しは起きない.
	n := 1
	for n < len(b) {
		n <<= 1
	}
	fb := make([]complex128, n)
	fa := make([]complex128, n)
	for i, v := range b {
		fb[i] = complex(v, 0)
	}
	for i, v := range a {
		fa[i] = complex(v, 0)
	}
	fft(fb, false)
	fft(fa, false)
	for i := range fb {
		fb[i] *= cmplx.Conj(fa[i])
	}
	fft(fb, true)
	c := make([]float64, len(b)-len(a)+1)
	for k := range c {
		c[k] = real(fb[k])
	}
	return c
}

// ArgMaxAbs は |x[i]| が最大になる最初の i を返す. x が空なら -1.
func ArgMaxAbs(x []float64) int {
	best, at := -1.0, -1
	for i, v := range x {
		if a := math.Abs(v); a > best {
			best, at = a, i
		}
	}
	return at
}

// Norm は x のユークリッドノルムを返す.
func Norm(x []float64) float64 {
	var s float64
	for _, v := range x {
		s += v * v
	}
	return math.Sqrt(s)
}

// DBFS は x の RMS を dBFS で返す. 無音で -Inf にならないよう 1e-9 を足す.
// x が空なら NaN を返す (numpy の空配列の平均と同じ).
func DBFS(x []float64) float64 {
	if len(x) == 0 {
		return math.NaN()
	}
	var s float64
	for _, v := range x {
		s += v * v
	}
	return 20 * math.Log10(math.Sqrt(s/float64(len(x)))+1e-9)
}

// Slice は x[lo:hi] を、範囲外を切り詰めて返す. numpy のスライスと同じく、はみ出しはエラーにしない.
func Slice(x []float64, lo, hi int) []float64 {
	lo = max(0, min(lo, len(x)))
	hi = max(lo, min(hi, len(x)))
	return x[lo:hi]
}
