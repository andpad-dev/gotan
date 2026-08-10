# 中級: リクエスト後の監査ログを期限付きで残そう

受注ステータスを更新する HTTP API では、応答を返したあとにも監査ログを保存します。クライアントが接続を閉じると `r.Context()` が取り消され、監査ログも途中で止まってしまいました。一方で、ログにはリクエスト ID を残し、失敗した書き込みが無期限に続かないようにする必要があります。

「HTTP リクエスト終了後も監査ログの書き込みを一定時間だけ続ける」ことをやりたいです。どういうふうにやればいいか調べよう。

[Go Playground で実行](https://go.dev/play/p/02PmNtj9kUO) し、親を取り消した後のエラーとリクエスト ID を観測してください。

```go
package main

import (
	"context"
	"fmt"
)

type requestIDKey struct{}

func main() {
	parent, cancel := context.WithCancel(
		context.WithValue(context.Background(), requestIDKey{}, "req-42"),
	)
	detached := context.WithoutCancel(parent)
	cancel()

	fmt.Println("parent:", parent.Err())
	fmt.Println("detached:", detached.Err())
	fmt.Println("request ID:", detached.Value(requestIDKey{}))
}
```

Go 1.26.4 での実行結果です。

```text
parent: context canceled
detached: <nil>
request ID: req-42
```

---

## 設問 1: 親の取消を切り離しても、何を引き継げるのか？

実行例では親が `context canceled` でも、派生させた値は取り消されず、リクエスト ID を取得できます。この派生方法は、親から何を引き継ぎ、`Deadline`、`Done`、`Err` をどう変えるのでしょうか。

<details>
<summary>ヒント</summary>

この API が追加された Go のリリースノートから `context` を探します。次にパッケージのドキュメントと実装を順に読んで、値の探索と取消通知を分けて確認してください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.21 リリースノートの context](https://go.dev/doc/go1.21#context) を開き、親の取消を伝播しない派生コンテキストが追加されたことを確認する。
2. リリースノートから [context の `WithoutCancel`](https://pkg.go.dev/context#WithoutCancel) を開き、`Deadline`、`Done`、`Err` の規則を読む。
3. 固定版の [context 実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/context/context.go;l=592) で、各メソッドと `Value` の実装を確認する。

**答え**

`context.WithoutCancel(parent)` は、親の値をたどれる派生コンテキストを返しますが、親が取り消されても取り消されません。返されたコンテキストは期限を持たず、`Done` は `nil`、`Err` は `nil` です。したがって実行例では親の `Err` だけが `context canceled` になり、`detached` からはリクエスト ID を引き続き取得できます。

</details>

---

## 設問 2: 監査ログの処理を無期限にしないには？

親の取消を切り離すだけでは、停止しない書き込みを待ち続けるおそれがあります。監査ログの書き込みには、どのように独自の期限を付けるべきでしょうか。また、`Done` が `nil` であることは、`select` で待つ処理にどんな影響を与えますか。

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
4. 固定版の [context 実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/context/context.go;l=592) と照合する。

**答え**

監査ログ用には、親を `WithoutCancel` で切り離したあと、その結果から `WithTimeout` で短い独自の期限を作ります。そして処理が終わったら、返された `CancelFunc` を必ず呼びます。`WithoutCancel` 単体の `Done` は `nil` なので、そこだけを待つ `select` のケースは選ばれず、取消を待つ仕組みにはなりません。独自のタイムアウトを重ねることで、クライアント切断には左右されず、外部のログ保存先が止まっても一定時間で処理を打ち切れます。

</details>

---

## 設問 3: この分離ができたことをどう検証するか？

ハンドラーのテストでは、クライアント切断を親コンテキストの取消として再現できます。監査ログ処理について、どの二つの観測を確認すれば「リクエスト ID は残すが、クライアントの取消には従わず、独自の期限には従う」という要件を確かめられるでしょうか。

<details>
<summary>ヒント</summary>

実行例の三行を、テストで観測する状態へ言い換えます。そのうえで、設問 2 の独自期限を短い値にし、ログ保存先を制御できるテストダブルで確認する順序を考えてください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.21 リリースノートの context](https://go.dev/doc/go1.21#context) から、親の取消を伝播しないという目的を確認する。
2. [context の `WithoutCancel`](https://pkg.go.dev/context#WithoutCancel) と [context の `WithTimeout`](https://pkg.go.dev/context#WithTimeout) を順に読み、値の伝播と独自の期限を照合する。
3. [実行例](https://go.dev/play/p/02PmNtj9kUO) を基準に、親の取消後も値を読める状態を再現する。

**答え**

一つ目は、親を取り消した後も監査ログの処理がリクエスト ID を取得でき、ただちに `context canceled` にならないことです。これは冒頭の `detached: <nil>` と `request ID: req-42` を再現する観測です。二つ目は、ログ保存先が応答しない場合でも、監査ログ用に作った独自の期限で処理が終わることです。

この二つを分けて検証すれば、単にバックグラウンド化しただけではなく、冒頭の「クライアント切断で記録が止まる」問題を解消しながら、設問 2 で見つけた `Done` が `nil` という性質による無期限待ちも防げていると説明できます。

</details>

---

## 調査の入り口

1. [04-deep-dive の調べ方](../../README.md)を開き、Go の公式ドキュメントを起点にします。
2. [Go 1.21 リリースノート: context](https://go.dev/doc/go1.21#context) で追加された API の目的を調べます。
3. リリースノートから `context` パッケージのドキュメントへ進みます。
