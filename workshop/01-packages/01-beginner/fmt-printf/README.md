[進行ガイド・シナリオ一覧](../../../README.md) | [01-packages の調べ方](../../README.md)

# fmt.Printf の書式指定子を調べよう

先輩のコードで、`fmt.Printf` の書式指定子を見かけました。どんなものか調べてみましょう。

```go
value := "gopher"
fmt.Printf("%#[1]v %[1]T\n", value)
```

（Go Playground で動かす: https://go.dev/play/p/RTNSvn_p2Ai ）

何をやっているコードか調べてみましょう。

この例では、同じ `value` を「Go のリテラルらしい表示」と「型名」の2通りで確認しています。まず `"gopher"` と `string` がどの位置に出るかを予想してから実行すると、書式の役割を追いやすくなります。

## 設問 1: `%v` と `%T`、そして `#` は、それぞれ何を意味する？

`value` にいろいろな値（構造体、マップ、ポインタなど）を入れて実行し、出力がどう変わるか確かめてみましょう。

<details>
<summary>ヒント</summary>

- Overview 冒頭の「Printing」セクションに、`%v` のような書式指定子の一覧表がある
- 少し下の「Other flags」に `#` フラグの説明がある

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `fmt` パッケージを開く。
2. https://pkg.go.dev/fmt の Overview 冒頭の「Printing」セクションを読む。
3. `%v` や `%T` のような書式指定子を、fmt のドキュメントでは **verb** と呼ぶ。verb の一覧表から `%v` と `%T` を探す。
4. 少し下の「Other flags」で `#` フラグの説明を探す。

**答え**

- `%v`: 値をデフォルトのフォーマットで出力する。
- `%T`: 値の型を Go の構文で出力する。
- `#`: alternate format（代替フォーマット）を指定するフラグ。`%v` と組み合わせた `%#v` は、値を Go の構文表現（Go-syntax representation）で出力する。

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

- ドキュメントの Overview で `[` を検索してみよう
- 例えば `fmt.Printf("%[1]s / %[1]s\n", "gopher")` は、1つの引数を何回使うでしょうか。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `fmt` パッケージを開く。
2. https://pkg.go.dev/fmt の Overview で `[` を検索する。
3. 「Explicit argument indexes」セクションにたどり着く。

**答え**

`[n]` は explicit argument index（明示的な引数インデックス）で、「次の verb が使う引数を n 番目に切り替える」指定です。
通常、verb は引数を先頭から順番に消費しますが、`%#[1]v %[1]T` はどちらの verb にも「1 番目の引数を使え」と指定しているため、1 つの `value` が 2 回出力されます。
同じ値を複数のフォーマットで出したいときに、引数を重複して渡さずに済みます。

</details>

---

<details>
<summary>こぼれ話: フラグと引数インデックスの順序</summary>

`#` フラグと `[1]` の順序を入れ替えた `%[1]#v` は動きません。

```go
fmt.Printf("%[1]#v %[1]T\n", value)
// 出力: %!#(string=gopher)v string
```

引数インデックスの直後には verb が来る必要があるため、`#` が不正な verb として扱われ、`%!#(...)` というエラー表記になります。
フラグは引数インデックスより前に書く、と覚えておきましょう。

実際に動かない様子は Go Playground で確認できます: https://go.dev/play/p/fYQqEjznC-T

</details>

---

## 調査の入り口

- https://pkg.go.dev/fmt
