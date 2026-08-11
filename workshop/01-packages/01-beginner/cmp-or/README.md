# cmp.Or で通知先チャンネルを決めよう

社内通知サービスでは、個人設定・チーム設定・全社既定の順に送信先チャンネルを選びます。設定されていない項目は空文字です。個人設定が空なのに、チーム設定の `#backend-alerts` が選ばれるコードを見て、`cmp.Or` を見かけました。どんなものか調べてみましょう。

次のコードを [Go Playground で動かす](https://go.dev/play/p/3edBoI6v8PU) と、設定値を優先順に選ぶ結果と、すべて未設定の場合の結果を確認できます。

```go
package main

import (
	"cmp"
	"fmt"
)

func main() {
	personalChannel := ""
	teamChannel := "#backend-alerts"
	companyDefault := "#general"
	fmt.Println(cmp.Or(personalChannel, teamChannel, companyDefault))
	fmt.Printf("all unavailable: %q\n", cmp.Or("", ""))
}
```

実行結果:

```text
#backend-alerts
all unavailable: ""
```

---

## 設問 1: なぜチーム設定が選ばれる？

`personalChannel` は空なのに、出力はなぜ `#backend-alerts` になるのでしょうか。`cmp.Or` は引数をどの順番で見て、どの値を返すか調べてください。

たとえば `cmp.Or("", "#backend-alerts", "#general")` と書いたとき、空文字を読み飛ばしてどの値が選ばれるかを予想してから調べてみましょう。

<details>
<summary>ヒント</summary>

- [カテゴリの調べ方](../../README.md) から標準パッケージを開く。
- `cmp` パッケージの関数一覧で `Or` を探し、「zero value」という言葉に注目する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `cmp` パッケージを開く。
2. [cmp.Or](https://pkg.go.dev/cmp#Or) の説明を読む。
3. [cmp.go の実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmp/cmp.go;l=67) を開き、引数を先頭から比較する処理を確認する。

**答え**

`cmp.Or` は引数を左から順に見て、ゼロ値ではない最初の値を返します。`string` のゼロ値は空文字なので、空の `personalChannel` は飛ばされ、次の `teamChannel` である `#backend-alerts` が返ります。

優先順位は引数の並びで表します。このコードでは個人設定、チーム設定、全社既定の順です。

</details>

---

## 設問 2: すべて未設定なら、何を確認すべき？

`cmp.Or("", "")` が返す `""` は、設定が見つからなかったことを示します。通知を送らずに設定エラーとして扱うか、さらに既定値を足すかは業務ルール次第です。`cmp.Or` が返す情報だけで、その判断に必要なことは分かるでしょうか。

<details>
<summary>ヒント</summary>

- `Or` の「すべての値がゼロ値」の場合を確認する。
- 関数の戻り値の型と個数から、どの情報を返しているかを考える。
- `cmp.Or("", "")` と `cmp.Or("", "#general")` の結果を並べ、「候補が見つからなかった」ことを戻り値だけで区別できるか考える。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.22 Release Notes](https://go.dev/doc/go1.22) の `cmp` の項目を読む。
2. [cmp.Or](https://pkg.go.dev/cmp#Or) で、すべてゼロ値の場合の戻り値を確認する。
3. [実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmp/cmp.go;l=69) で、最後に返す値を確認する。

**答え**

すべての引数がゼロ値なら、`cmp.Or` はその型のゼロ値を返します。`string` では空文字です。戻り値は選ばれた文字列だけで、どの候補から選ばれたかや「未設定だった」という別の状態は返しません。

そのため、空文字を「未設定」として扱える今回の設定では、空文字のチェックを加えて業務ルールを適用します。空文字そのものを有効な通知先として区別したい設計なら、値だけでなく設定の有無も表せるデータ構造を用意する必要があります。

例えば「候補がない」と「設定に明示的な空文字が入っている」を区別したいなら、`value string` だけでなく `found bool` も一緒に返す関数にします。`cmp.Or` は値の優先順位を決める関数で、設定の有無までは記録しません。

</details>

---

<details>
<summary>こぼれ話</summary>

`cmp.Or` の型引数は `comparable` です。文字列・数値・ポインタなどのゼロ値を使った優先順位には向きますが、比較できない型を直接渡す用途には使えません。

</details>

---

## 調査の入り口

- [Go Documentation](https://go.dev/doc/)
- [01-packages の調べ方](../../README.md)
- [package cmp](https://pkg.go.dev/cmp)
