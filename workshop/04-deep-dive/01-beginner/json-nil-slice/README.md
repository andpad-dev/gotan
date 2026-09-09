[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# 初級: encoding/json の nil スライスと空スライスを調べよう

**実行環境**: ブラウザだけ。Go のインストールは不要です。

API が `{"values":null}` を返していました。利用側は常に配列として処理するため、空のときも `{"values":[]}` で返す契約です。

レビューで次の `json.Marshal` を見かけました。どんなものか調べてみましょう。

まずは、空に見える二つの値で出力がどう変わるかを [Go Playground で実行](https://go.dev/play/p/jz0RLXHZtWU) して観測します。

```go
package main

import (
	"encoding/json"
	"fmt"
)

type response struct {
	Values []string `json:"values"`
}

func main() {
	nilJSON, _ := json.Marshal(response{})
	emptyJSON, _ := json.Marshal(response{Values: []string{}})
	fmt.Println("nil:", string(nilJSON))
	fmt.Println("empty:", string(emptyJSON))
}
```

Go 1.26.4 での実行結果です。

```text
nil: {"values":null}
empty: {"values":[]}
```

<details>
<summary>調査の入り口</summary>

まず [04-deep-dive の調べ方](../../README.md)を開き、Go 言語仕様を起点にします。

そのうえで、次のどれかから入ります。

- [Go 言語仕様: Slice types](https://go.dev/ref/spec#Slice_types) — `nil` と空スライス
- 仕様ページから標準ライブラリのドキュメントへ進み、`encoding/json` を確認

</details>

---

## 設問 1: 同じ長さ 0 なのに、なぜ JSON が違うのか？

観測した二つの値は、どちらも `len` が 0 です。それでも `null` と `[]` に分かれる理由を、スライスの値と `json.Marshal` の規則から説明してください。

<details>
<summary>ヒント</summary>

スライス型の仕様で、宣言しただけの値と空の複合リテラルで作った値がどう区別されるかを調べます。次に、標準ライブラリの JSON エンコード規則でスライスを検索しましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 言語仕様の Slice types](https://go.dev/ref/spec#Slice_types) を開き、`nil` のスライスと初期化済みの空スライスが区別されることを確認する。
2. 仕様ページの標準ライブラリへの入口から [encoding/json の `Marshal`](https://pkg.go.dev/encoding/json#Marshal) を開き、スライスの JSON エンコード規則を読む。
3. ドキュメントから実装へ進み、固定版の [`sliceEncoder.encode`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/encoding/json/encode.go;l=843) が `IsNil` のとき `null` を書き出すことを確認する。

**答え**

ゼロ値の `[]string` は `nil` スライスです。一方、`[]string{}` は長さ 0 でも初期化済みの空スライスです。`encoding/json` は `nil` スライスを JSON の `null` として、初期化済みスライスを JSON 配列としてエンコードします。そのため、前者が `{"values":null}`、後者が `{"values":[]}` になります。

</details>

---

## 設問 2: API 契約どおりに空配列を返すには？

この API では、要素がなくても `values` フィールドを残して `[]` で表す契約です。`omitempty` を付ける案ではなく、レスポンスをどう組み立てればよいでしょうか。また、なぜ `omitempty` はこの契約に合わないのでしょうか。

<details>
<summary>ヒント</summary>

`Marshal` のフィールドタグの説明で、長さ 0 のスライスがどう扱われるかを確認します。`null`、`[]`、フィールドが存在しない場合の三つを区別して考えてください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 言語仕様の Slice types](https://go.dev/ref/spec#Slice_types) で、空スライスを初期化する方法を確認する。
2. [encoding/json の `Marshal`](https://pkg.go.dev/encoding/json#Marshal) で `omitempty` の説明を読み、長さ 0 のスライスも空値と判定されることを確認する。
3. 設問 1 の [固定版実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/encoding/json/encode.go;l=843) と実行結果を照合する。

**答え**

レスポンスを作るときに `Values: []string{}` のように空スライスを明示して初期化します。そうすれば JSON は `[]` になり、フィールドも残ります。`omitempty` は長さ 0 のスライスをフィールドごと省略するため、この契約には使えません。

`null`、空配列、フィールドなしのどれを契約にするかは API 設計上の選択です。ここでは常に反復できる `[]` を選んでいるため、初期化済みの空スライスを返します。

</details>
