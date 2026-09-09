[シナリオ一覧](../SCENARIOS.md) | [ワークショップ進行ガイド](../README.md)

# チュートリアル: fmt.Printf の書式指定子を調べよう

**実行環境**: ブラウザだけ。Go のインストールは不要です。

同僚のコードで、`fmt.Printf` の書式指定子を見かけました。どんなものか調べてみましょう。

```go
package main

import "fmt"

func main() {
	value := "gopher"
	fmt.Printf("%#[1]v %[1]T\n", value)
}
```

[Go Playground で動かす](https://go.dev/play/p/RTNSvn_p2Ai) と、Go 1.27.0 では次のように出力されます。

```text
"gopher" string
```

何をやっているコードか調べてみましょう。

この例では、同じ `value` を「Go のリテラルらしい表示」と「型名」の2通りで確認しています。上の `"gopher"` と `string` が、それぞれどの書式から出たかを対応付けながら調べてみましょう。

<details>
<summary>調査の入り口</summary>

次のどちらかから入ります。

- [Go Documentation](https://go.dev/doc/) — Go の公式ドキュメントの入口。標準ライブラリはここからたどれる
- [package fmt](https://pkg.go.dev/fmt) — 書式指定子の説明は Overview にある

</details>

---

## 設問 1: `%v` と `%T`、そして `#` は、それぞれ何を意味する？

`value` にいろいろな値（構造体、マップ、ポインタなど）を入れて実行し、出力がどう変わるか確かめてみましょう。

<details>
<summary>ヒント</summary>

- Overview 冒頭の「Printing」にある verb の表で、`%v`、`%#v`、`%T` の 3 行を直接比べる。
- 「Other flags」の `#` は、`%#b` や `%#x` など verb ごとの別の効果を列挙している。そこに `%v` がなければ、verb の表へ戻る。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `fmt` パッケージを開く。
2. [fmt の「Printing」](https://pkg.go.dev/fmt#hdr-Printing) を開く。
3. `%v` や `%T` のような書式指定子を、fmt のドキュメントでは **verb** と呼ぶ。General の表から `%v`、`%#v`、`%T` を探し、3 行を比べる。
4. 同じ節の「Other flags」を読み、`#` の効果は verb ごとに異なり、そこには `%#v` の説明がないことを確認する。
5. [Go 1.27.0 の `fmtFlags`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/fmt/format.go;l=26-36) で、`%#v` が通常の `#` とは別の `sharpV` として扱われる理由を確認する。

**答え**

- `%v`: 値をデフォルトのフォーマットで出力する。
- `%T`: 値の型を Go の構文で出力する。
- `#`: 書式フラグの一つで、効果は組み合わせる verb によって異なる。`%v` との組み合わせは verb の表に `%#v` という別形式として載っており、値を Go の構文表現（Go-syntax representation）で出力する。

つまり `%#v` と `%T` を並べたこのコードは、「値の中身と型を、どちらも Go の構文で確認する」デバッグの定番イディオムです。

```go
value := "gopher"
fmt.Printf("%#[1]v %[1]T\n", value)
// 出力: "gopher" string
```

`%v` なら `gopher` と出るところが、`%#v` では Go のリテラルどおり `"gopher"`（クォート付き）になります。
構造体やマップを入れると、`%#v` と `%v` の差はもっとはっきり見えます。

</details>

---

## 設問 2: `[1]` は何をしている？

引数の `value` は 1 つしか渡していないのに、なぜ 2 回出力されるのでしょうか？

<details>
<summary>ヒント</summary>

- ドキュメントの Overview で `[` をページ内検索（Ctrl+F / Cmd+F）してみよう
- 例えば `fmt.Printf("%[1]s / %[1]s\n", "gopher")` は、1つの引数を何回使うでしょうか。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `fmt` パッケージを開く。
2. [fmt の Overview](https://pkg.go.dev/fmt) で `[` をページ内検索する。
3. [Explicit argument indexes](https://pkg.go.dev/fmt#hdr-Explicit_argument_indexes) にたどり着き、`[n]` がどの引数を選ぶかを読む。

**答え**

`[n]` は explicit argument index（明示的な引数インデックス）で、「次の verb が使う引数を n 番目に切り替える」指定です。
通常、verb は引数を先頭から順番に消費しますが、`%#[1]v %[1]T` はどちらの verb にも「1 番目の引数を使え」と指定しているため、1 つの `value` が 2 回出力されます。
同じ値を複数のフォーマットで出したいときに、引数を重複して渡さずに済みます。

</details>

---

<details>
<summary>こぼれ話: フラグと引数インデックスの順序</summary>

`#` フラグと `[1]` の順序を入れ替えた `%[1]#v` は動きません。

次の完全なコードを [Go Playground で動かす](https://go.dev/play/p/HczVP8-FaCF) と、正しい順序と誤った順序を同じ入力で比較できます。

```go
package main

import "fmt"

func main() {
	value := "gopher"
	fmt.Printf("%#[1]v %[1]T\n", value)
	fmt.Printf("%[1]#v %[1]T\n", value)
}
```

```text
"gopher" string
%!#(string=gopher)v string
```

引数インデックスの直後には verb が来る必要があるため、`#` が不正な verb として扱われ、`%!#(...)` というエラー表記になります。
フラグは引数インデックスより前に書く、と覚えておきましょう。

</details>
