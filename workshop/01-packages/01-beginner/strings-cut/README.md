[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [01-packages の調べ方](../../README.md)

# strings.Cut で最初の区切り位置を分けよう

![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8)

レビュー中に、文字列を最初の `=` で分けるコードを見かけました。
区切り記号がない場合と、区切り後が空の場合を区別できるか調べましょう。

次の観測ログを [Go Playground で動かす](https://go.dev/play/p/ieMFoFpSNtz) と、区切り記号の有無で結果が変わります。

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	for _, input := range []string{"left=right", "left"} {
		before, after, found := strings.Cut(input, "=")
		fmt.Printf("%q -> before=%q after=%q found=%t\n", input, before, after, found)
	}
}
```

実行結果:

```text
"left=right" -> before="left" after="right" found=true
"left" -> before="left" after="" found=false
```

<details>
<summary>調査の入り口</summary>

1. [01-packages の調べ方](../../README.md) を開き、標準パッケージのドキュメントの開き方を確かめます。
2. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `strings` パッケージを開きます。
3. [package strings](https://pkg.go.dev/strings) を開きます。`Cut` 関数の説明はこのページにあります。

</details>

---

## 設問 1: 3 つの戻り値は何を伝える？

`strings.Cut` の 3 つの戻り値はそれぞれ何でしょうか。`"left"` の場合、`after == ""` だけでは「区切り後が空なのか、`=` がなかったのか」を判定できません。その理由も説明してください。

<details>
<summary>ヒント</summary>

- まず [カテゴリの調べ方](../../README.md) から標準パッケージのドキュメントを開く。
- `Cut` を検索し、戻り値の名前と「区切り記号がない場合」の一文を読む。
- `"left="` と `"left"` を比べると、どちらも `after` が空でも `found` が異なることに注目する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `strings` パッケージを開く。
2. [strings.Cut](https://pkg.go.dev/strings#Cut) の説明を読む。
3. 実装の [strings.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/strings/strings.go;l=1273) で、`Cut` が内部実装へ委譲していることも確認する。

**答え**

戻り値は順に、最初の区切り記号より前の文字列 `before`、後ろの文字列 `after`、区切り記号を見つけたかを示す `found` です。

`"left"` を `"="` で切ると `before` は元の `"left"`、`after` は `""`、`found` は `false` になります。`after == ""` は `"left="` のように区切り後が空の場合にも起こるため、区切りの有無は `found` で確認します。

</details>

---

## 設問 2: どこで切れる？

`strings.Cut("left=middle=right", "=")` はどこで切れるでしょうか。最初の `=` だけで分ける性質を確認してください。

<details>
<summary>ヒント</summary>

- `Cut` の説明にある `first instance` に注目する。
- `before` と `after` を、もう一度同じ関数に渡せるか考える。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `strings` パッケージを開く。
2. [strings.Cut](https://pkg.go.dev/strings#Cut) の「first instance」の説明を読む。
3. [実装の該当箇所](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/strings/strings.go;l=1273) を開き、`Cut` の責務が 1 回の分割であることを確認する。

**答え**

結果は `before == "left"`、`after == "middle=right"`、`found == true` です。`Cut` は最初の一致だけで切ります。

最初の部分と残りを一度に分けたい場合、残りに `=` が含まれていても失われません。さらに分けるなら、`after` に対してもう一度 `Cut` できます。

</details>

---

<details>
<summary>こぼれ話</summary>

接頭辞や接尾辞だけを確かめたい札には、同じ `strings` パッケージの `CutPrefix` と `CutSuffix` もあります。どちらも、見つからないときに元の文字列と `false` を返す設計です。

</details>
