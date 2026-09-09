[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [01-packages の調べ方](../../README.md)

# fmt.Sprintf で文字列を組み立てよう

**実行環境**: ブラウザだけ。Go のインストールは要りません。

同僚のコードで、値を埋め込んで文字列を作っている次のコードを見かけました。

```go
package main

import "fmt"

func main() {
	name := "gopher"
	points := 42
	msg := fmt.Sprintf("ユーザー %s のポイントは %d です", name, points)
	fmt.Println(msg)
	fmt.Printf("%T\n", msg)
}
```

（Go Playground で動かす: https://go.dev/play/p/uDdRXVMggVb ）

実行結果:

```
ユーザー gopher のポイントは 42 です
string
```

`fmt.Sprintf` がどんなものか調べてみましょう。

<details>
<summary>調査の入り口</summary>

まず [01-packages の調べ方](../../README.md) を開き、標準パッケージのドキュメント（pkg.go.dev）の開き方を確かめます。

そのうえで、次のどれかから入ります。

- [Go Documentation](https://go.dev/doc/) — 標準ライブラリの `fmt` パッケージを開く入口
- [package fmt](https://pkg.go.dev/fmt) — `Sprintf` の説明と、書式指定子の説明（Overview）はこのページにある

</details>

---

## 設問 1: `Sprintf` は何を返す？ `%s` や `%d` は何を意味する？

`fmt.Println` は画面に出力しますが、この `fmt.Sprintf` は何を返しているのでしょうか。
また、`%s` や `%d` の部分に `name` と `points` がどう対応しているのか確かめてみましょう。

<details>
<summary>ヒント</summary>

- https://pkg.go.dev/fmt で `f` キーを押し、`Sprintf` を検索して関数のシグネチャ（戻り値の型）を見る
- verb の意味は Overview 冒頭の「Printing」セクションの一覧表にある

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `fmt` パッケージを開く。
2. https://pkg.go.dev/fmt を開き、`f` キーで検索ダイアログを出して `Sprintf` を打ち込み、関数のところへジャンプする。
3. シグネチャ `func Sprintf(format string, a ...any) string` を見て、戻り値が `string` であることを確認する。
4. Overview 冒頭の「Printing」セクションに戻り、`%s` と `%d` の意味を verb の一覧表で確認する。

**答え**

- `fmt.Sprintf` は、`Printf` と同じ書式指定でフォーマットした結果を **画面に出力せず `string` として返す** 関数です（`S` は String の S）。作った文字列を変数に入れて後で使いたいときに使います。
- `%s`: 引数を文字列として埋め込む verb。
- `%d`: 引数を 10 進整数として埋め込む verb。
- `format` に書いた `%s` `%d` が、後ろに渡した引数 `name` `points` に左から順に対応します。

```go
name := "gopher"
points := 42
msg := fmt.Sprintf("ユーザー %s のポイントは %d です", name, points)
fmt.Println(msg)
// 出力: ユーザー gopher のポイントは 42 です
fmt.Printf("%T\n", msg)
// 出力: string
```

`fmt.Printf(...)` は「フォーマットして即出力」、`fmt.Sprintf(...)` は「フォーマットして文字列を返す」——このペアで覚えておくと便利です。

</details>

---

## 設問 2: 数値の桁揃えやゼロ埋めをしたい

数値の幅を揃えたり、`007` のようにゼロ埋めしたり、小数点以下の桁数を固定したりしたくなります。
`%d` や `%f` に幅や精度をどう指定すればよいか調べてみましょう。

次のコードを実行して、指定した幅・精度と実際の出力を対応付けてから、書式の指定方法を調べましょう。

```go
package main

import "fmt"

func main() {
	fmt.Println(fmt.Sprintf("[%5d]", 42))
	fmt.Println(fmt.Sprintf("[%-5d]", 42))
	fmt.Println(fmt.Sprintf("[%05d]", 42))
	fmt.Println(fmt.Sprintf("[%.2f]", 3.14159))
	fmt.Println(fmt.Sprintf("[%8.2f]", 3.14159))
}
```

（[Go Playground で動かす](https://go.dev/play/p/SMFWIRANv3B)）

実行結果:

```
[   42]
[42   ]
[00042]
[3.14]
[    3.14]
```

<details>
<summary>ヒント</summary>

- Overview の「Printing」の中ほどにある「Width and precision」セクションを読む
- `%` と verb の **あいだ** に数字を書く

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `fmt` パッケージを開く。
2. https://pkg.go.dev/fmt の Overview で「Width and precision」セクションを探す。
3. 幅は `%` の直後、精度は `.` に続けて書くこと、`0` フラグでゼロ埋めになること、`-` フラグで左寄せになることを確認する。

**答え**

- `%5d`: 幅 5 で右寄せ（足りない分は空白）。
- `%-5d`: 幅 5 で左寄せ。
- `%05d`: 幅 5 でゼロ埋め。
- `%.2f`: 小数点以下 2 桁。
- `%8.2f`: 幅 8・小数点以下 2 桁。
- `%%`: verb ではなく、文字としての `%` を出力する。

```go
fmt.Println(fmt.Sprintf("[%5d]", 42))      // [   42]
fmt.Println(fmt.Sprintf("[%-5d]", 42))     // [42   ]
fmt.Println(fmt.Sprintf("[%05d]", 42))     // [00042]
fmt.Println(fmt.Sprintf("[%.2f]", 3.14159))   // [3.14]
fmt.Println(fmt.Sprintf("[%8.2f]", 3.14159))  // [    3.14]
fmt.Println(fmt.Sprintf("%d%%", 50))       // 50%
```

（Go Playground で動かす: https://go.dev/play/p/nUCAOoJdeG6 ）

幅・精度の指定は、設問 1 と同じく `fmt.Sprintf` で組み立てた文字列にもそのまま効きます。

</details>

---

<details>
<summary>こぼれ話: 名前の頭文字は関数の役割を表している</summary>

`fmt` の出力系関数は、頭文字が役割を表しています。

- 何もなし（`Print`, `Printf`, `Println`）: 標準出力へ書き出す。
- `S`（`Sprint`, `Sprintf`, `Sprintln`）: 文字列として返す。
- `F`（`Fprint`, `Fprintf`, `Fprintln`）: 指定した `io.Writer` へ書き出す。

さらに末尾の `f` は「format 文字列を取る」、`ln` は「末尾に改行を付け、引数の間に空白を入れる」という意味です。
この規則を知っておくと、`Sprintf` を初めて見ても「文字列を返す、フォーマット版」だと名前から推測できます。

</details>
