# 初級: API の担当者一覧を `[]` で返そう

案件詳細 API のレスポンスをフロントエンドと確認していたところ、担当者がいない案件だけ `{"assignees":null}` が返っていました。画面側は常に配列として処理したいので、空のときも `{"assignees":[]}` で返す契約です。

レビューで次の `json.Marshal` を見かけました。どんなものか調べてみましょう。

まずは、空に見える二つの値で出力がどう変わるかを [Go Playground で実行](https://go.dev/play/p/M_lGwwe4dMX) して観測します。

```go
package main

import (
	"encoding/json"
	"fmt"
)

type assigneeResponse struct {
	Assignees []string `json:"assignees"`
}

func main() {
	nilJSON, _ := json.Marshal(assigneeResponse{})
	emptyJSON, _ := json.Marshal(assigneeResponse{Assignees: []string{}})
	fmt.Println("nil:", string(nilJSON))
	fmt.Println("empty:", string(emptyJSON))
}
```

Go 1.26.4 での実行結果です。

```text
nil: {"assignees":null}
empty: {"assignees":[]}
```

---

## 設問 1: 同じ長さ 0 なのに、なぜ JSON が違うのか？

`assigneeResponse{}` と `assigneeResponse{Assignees: []string{}}` は、どちらも `len` が 0 です。それでも `null` と `[]` に分かれる理由を、スライスの値と `json.Marshal` の規則から説明してください。

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

ゼロ値の `[]string` は `nil` スライスです。一方、`[]string{}` は長さ 0 でも初期化済みの空スライスです。`encoding/json` は `nil` スライスを JSON の `null` として、初期化済みスライスを JSON 配列としてエンコードします。そのため、前者が `{"assignees":null}`、後者が `{"assignees":[]}` になります。

</details>

---

## 設問 2: API 契約どおりに空配列を返すには？

この API では、担当者がいないことも `assignees` フィールドを残して `[]` で表す契約です。`omitempty` を付ける案ではなく、レスポンスをどう組み立てればよいでしょうか。また、なぜ `omitempty` はこの契約に合わないのでしょうか。

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

レスポンスを作るときに `Assignees: []string{}` のように空スライスを明示して初期化します。そうすれば JSON は `[]` になり、フィールドも残ります。`omitempty` は長さ 0 のスライスをフィールドごと省略するため、この API で必要な「担当者はいないが、一覧という項目はある」という表現には使えません。

`null`、空配列、フィールドなしのどれを契約にするかは API 設計上の選択です。この画面との契約では配列を常に反復できる `[]` を選んでいるため、初期化済みの空スライスを返します。

</details>

---

## 調査の入り口

1. [04-deep-dive の調べ方](../../README.md)を開き、Go 言語仕様を起点にします。
2. [Go 言語仕様: Slice types](https://go.dev/ref/spec#Slice_types) で `nil` と空スライスを調べます。
3. 仕様ページから標準ライブラリのドキュメントへ進み、`encoding/json` を確認します。
