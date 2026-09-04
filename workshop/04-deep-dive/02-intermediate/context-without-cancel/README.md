[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# 中級: context.WithoutCancel に独自の期限を付けよう

![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8)

親コンテキストの値は引き継ぎつつ、親のキャンセルは切り離したいコードがあります。ただし、派生した処理が無期限に続かないよう、独自の期限も必要です。どう組み合わせるか調べましょう。

[Go Playground で実行](https://go.dev/play/p/r-de7RWWAPE) してください。観測するのは、親を取り消した後のエラー、値、期限と `Done`、独自タイムアウトです。

```go
package main

import (
	"context"
	"fmt"
	"time"
)

type markerKey struct{}

func main() {
	parent, cancel := context.WithCancel(
		context.WithValue(context.Background(), markerKey{}, "kept"),
	)
	detached := context.WithoutCancel(parent)
	cancel()

	_, hasDeadline := detached.Deadline()
	fmt.Println("parent:", parent.Err())
	fmt.Println("detached:", detached.Err())
	fmt.Println("value:", detached.Value(markerKey{}))
	fmt.Println("detached has deadline:", hasDeadline)
	fmt.Println("detached Done is nil:", detached.Done() == nil)

	limited, stop := context.WithTimeout(detached, 10*time.Millisecond)
	defer stop()
	<-limited.Done()
	fmt.Println("limited:", limited.Err())
}
```

Go 1.27.0 での実行結果です。

```text
parent: context canceled
detached: <nil>
value: kept
detached has deadline: false
detached Done is nil: true
limited: context deadline exceeded
```

---

## 設問 1: 親のキャンセルを切り離しても、何を引き継げるのか？

実行例では親が `context canceled` でも、派生させた値は取り消されず、リクエスト ID を取得できます。この派生方法は、親から何を引き継ぎ、`Deadline`、`Done`、`Err` をどう変えるのでしょうか。

<details>
<summary>ヒント</summary>

この API が追加された Go のリリースノートから `context` を探します。次にパッケージのドキュメントと実装を順に読んで、値の探索とキャンセル通知を分けて確認してください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.21 リリースノートの context](https://go.dev/doc/go1.21#context) を開き、親のキャンセルを伝播しない派生コンテキストが追加されたことを確認する。
2. リリースノートから [context の `WithoutCancel`](https://pkg.go.dev/context#WithoutCancel) を開き、`Deadline`、`Done`、`Err` の規則を読む。
3. [Go 1.27.0 の `WithoutCancel` と各メソッド](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/context/context.go;l=582-615) を開き、宣言、`Deadline`、`Done`、`Err`、`Value` の順に実装を確認する。

**答え**

`context.WithoutCancel(parent)` は、親の値をたどれる派生コンテキストを返しますが、親が取り消されても取り消されません。返されたコンテキストは期限を持たず、`Done` は `nil`、`Err` は `nil` です。したがって実行例では親の `Err` だけが `context canceled` になり、`detached` からはリクエスト ID を引き続き取得できます。

</details>

---

## 設問 2: 派生した処理を無期限にしないには？

親のキャンセルを切り離すだけでは、処理を待ち続けるおそれがあります。どのように独自の期限を付けるべきでしょうか。また、`Done` が `nil` であることは、`select` で待つ処理にどんな影響を与えますか。

<details>
<summary>ヒント</summary>

設問 1 で確認した `Done` の性質を、チャネルの `nil` 値を含む `select` の規則と照合します。次に `context` パッケージで時間制限を作る関数と、その戻り値を解放する必要性を調べましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 言語仕様の Select statements](https://go.dev/ref/spec#Select_statements) で、`nil` チャネルの通信は選択されないことを確認する。
2. [context の `WithoutCancel`](https://pkg.go.dev/context#WithoutCancel) で `Done` が `nil` であることを再確認する。
3. 同じパッケージの [context の `WithTimeout`](https://pkg.go.dev/context#WithTimeout) を開き、期限と `CancelFunc` の規則を確認する。
4. [Go 1.27.0 の `WithoutCancel`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/context/context.go;l=582-615) と [`WithTimeout`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/context/context.go;l=696-709) の実装を開き、切り離した親から新しい期限付きコンテキストを作る流れを照合する。

**答え**

親を `WithoutCancel` で切り離したあと、その結果から `WithTimeout` で短い独自の期限を作ります。そして処理が終わったら、返された `CancelFunc` を必ず呼びます。`WithoutCancel` 単体の `Done` は `nil` なので、そこだけを待つ `select` のケースは選ばれず、キャンセルを待つ仕組みにはなりません。独自のタイムアウトを重ねることで、一定時間で処理を打ち切れます。

</details>

---

## 設問 3: この分離ができたことをどう検証するか？

確かめたい要件は「値は残すが、親のキャンセルには従わず、独自の期限には従う」です。どの二つの観測を確認すればよいでしょうか。

<details>
<summary>ヒント</summary>

実行例の三行を、テストで観測する状態へ言い換えます。そのうえで、設問 2 の独自期限を短い値にして確認する順序を考えてください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.21 リリースノートの context](https://go.dev/doc/go1.21#context) から、親のキャンセルを伝播しないという目的を確認する。
2. [context の `WithoutCancel`](https://pkg.go.dev/context#WithoutCancel) と [context の `WithTimeout`](https://pkg.go.dev/context#WithTimeout) を順に読み、値の伝播と独自の期限を照合する。
3. [実行例](https://go.dev/play/p/r-de7RWWAPE) を基準に、親のキャンセル後も値を読めることと、独自の期限では `context deadline exceeded` になることを再現する。

**答え**

一つ目は、親を取り消した後も派生したコンテキストから値を取得でき、ただちに `context canceled` にならないことです。これは `detached: <nil>` と `value: kept` を再現する観測です。二つ目は、独自の期限で処理が終わり、`context deadline exceeded` になることです。

この二つを分けて検証すれば、親のキャンセルを切り離しながら、設問 2 で見つけた `Done` が `nil` という性質による無期限待ちも防げていると説明できます。

</details>

---

## 調査の入り口

1. [04-deep-dive の調べ方](../../README.md)を開き、Go の公式ドキュメントを起点にします。
2. [Go 1.21 リリースノート: context](https://go.dev/doc/go1.21#context) で追加された API の目的を調べます。
3. リリースノートから `context` パッケージのドキュメントへ進みます。
