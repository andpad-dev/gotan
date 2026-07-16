# fmt パッケージのドキュメントを読もう

先輩のコードで、次の 1 行を見かけました。

```go
fmt.Printf("%#[1]v %[1]T\n", value)
```

（Go Playground で動かす: https://go.dev/play/p/fbq6AWx_dDE ）

何をやっているコードか調べてみましょう。

## 設問 1: `%v` と `%T`、そして `#` は、それぞれ何を意味する？

`value` にいろいろな値（構造体、マップ、ポインタなど）を入れて実行し、出力がどう変わるか確かめてみましょう。

<details>
<summary>ヒント</summary>

- Overview 冒頭の「Printing」セクションに、`%v` のような書式指定子の一覧表がある
- 少し下の「Other flags」に `#` フラグの説明がある

</details>

<details>
<summary>答え</summary>

### 調査ルート

1. https://pkg.go.dev/fmt を開き、Overview 冒頭の「Printing」セクションを読む。
2. `%v` や `%T` のような書式指定子を、fmt のドキュメントでは **verb** と呼ぶ。verb の一覧表から `%v` と `%T` を探す。
3. 少し下の「Other flags」で `#` フラグの説明を探す。

### 答え

- `%v`: 値をデフォルトのフォーマットで出力する。
- `%T`: 値の型を Go の構文で出力する。
- `#`: alternate format（代替フォーマット）を指定するフラグ。`%v` と組み合わせた `%#v` は、値を Go の構文表現（Go-syntax representation）で出力する。

つまり `%#v` と `%T` を並べたこのコードは、「値の中身と型を、どちらも Go の構文で確認する」デバッグの定番イディオムです。

```go
value := struct{ Name string }{"gopher"}
fmt.Printf("%#[1]v %[1]T\n", value)
// 出力: struct { Name string }{Name:"gopher"} struct { Name string }
```

</details>

## 設問 2: `[1]` は何をしている？

引数の `value` は 1 つしか渡していないのに、なぜ 2 回出力されるのでしょうか？

<details>
<summary>ヒント</summary>

- ドキュメントの Overview で `[` を検索してみよう

</details>

<details>
<summary>答え</summary>

### 調査ルート

1. https://pkg.go.dev/fmt の Overview で `[` を検索する。
2. 「Explicit argument indexes」セクションにたどり着く。

### 答え

`[n]` は explicit argument index（明示的な引数インデックス）で、「次の verb が使う引数を n 番目に切り替える」指定です。
通常、verb は引数を先頭から順番に消費しますが、`%#[1]v %[1]T` はどちらの verb にも「1 番目の引数を使え」と指定しているため、1 つの `value` が 2 回出力されます。
同じ値を複数のフォーマットで出したいときに、引数を重複して渡さずに済みます。

</details>

<details>
<summary>こぼれ話: フラグと引数インデックスの順序</summary>

`#` フラグと `[1]` の順序を入れ替えた `%[1]#v` は動きません。

```go
fmt.Printf("%[1]#v %[1]T\n", value)
// 出力: {%!#(string=gopher)}v struct { Name string }
```

引数インデックスの直後には verb が来る必要があるため、`#` が不正な verb として扱われ、`%!#(...)` というエラー表記になります。
フラグは引数インデックスより前に書く、と覚えておきましょう。

実際に動かない様子は Go Playground で確認できます: https://go.dev/play/p/aplHlg0U0Xc

</details>

## 調査の入り口

- https://pkg.go.dev/fmt
