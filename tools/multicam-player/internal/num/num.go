// Package num は、コマンドの出力で使う数値の丸めと書式をまとめる.
package num

import (
	"strconv"
	"strings"
)

// Round は x を小数 digits 桁に丸める.
// 2 進数の値を 10 進の文字列にしてから丸めるので、0.25 のような境界も表示どおりに丸まる.
//
// 例:
//
//	num.Round(78.46, 1) // 78.5
func Round(x float64, digits int) float64 {
	v, _ := strconv.ParseFloat(strconv.FormatFloat(x, 'f', digits, 64), 64)
	return v
}

// Float は x を、整数値でも小数点を付けた最短の表記にする (0 は "0.0", 78.5 は "78.5").
// data.js の秒数を、小数として書く既存の形式に合わせるために使う.
func Float(x float64) string {
	s := strconv.FormatFloat(x, 'f', -1, 64)
	if !strings.ContainsAny(s, ".eEn") {
		s += ".0"
	}
	return s
}

// JSONFloat は、JSON に小数点付きで書き出す float64.
type JSONFloat float64

// MarshalJSON は Float と同じ表記で書き出す.
func (f JSONFloat) MarshalJSON() ([]byte, error) {
	return []byte(Float(float64(f))), nil
}

// ParseFloats は文字列を順に float64 に直す. 1 つでも数でなければエラーを返す.
func ParseFloats(ss []string) ([]float64, error) {
	out := make([]float64, len(ss))
	for i, s := range ss {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}
