[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [01-packages の調べ方](../../README.md)

# strings.Cut で連携設定のキーと値を分けよう

![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8)

外部 SaaS との連携設定では、管理画面から `region=asia-east1` のような `key=value` 形式の項目が届きます。実装担当は、区切り記号が欠けた入力を「値が空」と誤認しないよう、まず設定を安全に分けたいと考えました。コードに `strings.Cut` を見かけました。どんなものか調べてみましょう。

次の観測ログを [Go Playground で動かす](https://go.dev/play/p/qzzfmNtViQU) と、区切り記号がある設定とない設定で結果が変わります。

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	for _, setting := range []string{"region=asia-east1", "region"} {
		key, value, found := strings.Cut(setting, "=")
		fmt.Printf("%q -> key=%q value=%q found=%t\n", setting, key, value, found)
	}
}
```

実行結果:

```text
"region=asia-east1" -> key="region" value="asia-east1" found=true
"region" -> key="region" value="" found=false
```

---

## 設問 1: 3 つの戻り値は何を伝える？

`strings.Cut` の 3 つの戻り値はそれぞれ何でしょうか。`"region"` の場合に `value == ""` だけでは「値が空なのか、そもそも `=` がなかったのか」を判定できない理由も説明してください。

<details>
<summary>ヒント</summary>

- まず [カテゴリの調べ方](../../README.md) から標準パッケージのドキュメントを開く。
- `Cut` を検索し、戻り値の名前と「区切り記号がない場合」の一文を読む。
- `"region="` と `"region"` を比べると、どちらも `value` が空でも `found` が異なることに注目する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `strings` パッケージを開く。
2. [strings.Cut](https://pkg.go.dev/strings#Cut) の説明を読む。
3. 実装の [strings.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/strings/strings.go;l=1273) で、`Cut` が内部実装へ委譲していることも確認する。

**答え**

戻り値は順に、最初の区切り記号より前の文字列 `before`、後ろの文字列 `after`、区切り記号を見つけたかを示す `found` です。

`"region"` を `"="` で切ると `before` は元の `"region"`、`after` は `""`、`found` は `false` になります。`after == ""` は `"region="` のように値が本当に空の場合にも起こるため、設定形式を検証するには `found` を確認します。

</details>

---

## 設問 2: どこで切れる？

連携先から `callback=https://example.com/a=b` が届きました。`strings.Cut(setting, "=")` はどこで切れるでしょうか。最初の `=` だけを意味のある境界とする今回の形式で、なぜこの性質が役立つか考えてください。

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

結果は `before == "callback"`、`after == "https://example.com/a=b"`、`found == true` です。`Cut` は最初の一致だけで切ります。

キーと残りの値を一度に分けたい形式なら、値に `=` が含まれていても失われないため便利です。値もさらに構造化されているなら、`after` に対してもう一度 `Cut` する、と段階的に読み取れます。

</details>

---

<details>
<summary>こぼれ話</summary>

接頭辞や接尾辞だけを確かめたい札には、同じ `strings` パッケージの `CutPrefix` と `CutSuffix` もあります。どちらも、見つからないときに元の文字列と `false` を返す設計です。

</details>

---

## 調査の入り口

- [Go Documentation](https://go.dev/doc/)
- [01-packages の調べ方](../../README.md)
- [package strings](https://pkg.go.dev/strings)
