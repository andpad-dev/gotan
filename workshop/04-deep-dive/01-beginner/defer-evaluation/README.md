[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# 初級: defer の引数評価と実行順を調べよう

**実行環境**: ブラウザだけ。Go のインストールは不要です。

状態を `started` から `completed` へ変えるコードで `defer` を使っています。二つの遅延呼び出しが異なる値を出す理由を調べましょう。

次のコードの `defer` を見かけました。どんなものか調べてみましょう。

[Go Playground で実行](https://go.dev/play/p/XEEZ1ax4se2) すると、二つのログが異なる値で、しかも逆順に出ます。

```go
package main

import "fmt"

func main() {
	status := "started"
	defer fmt.Println("direct:", status)
	defer func() {
		fmt.Println("closure:", status)
	}()

	status = "completed"
}
```

Go 1.27.0 での実行結果です。

```text
closure: completed
direct: started
```

<details>
<summary>調査の入り口</summary>

まず [04-deep-dive の調べ方](../../README.md)を開き、Go 言語仕様を起点にします。

そのうえで、次のどれかから入ります。

- [Go 言語仕様: Defer statements](https://go.dev/ref/spec#Defer_statements) — `defer` 文の規則を定義する節
- [Go 言語仕様: Function literals](https://go.dev/ref/spec#Function_literals) — 関数リテラルの規則を定義する節
- [Go Blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) — 観測結果の照合先

</details>

---

## 設問 1: `defer` した関数呼び出しの引数はいつ評価されるのか？

コードでは、最初に `defer fmt.Println("direct:", status)` を書いています。`status` を `completed` へ代入するのは、その後です。それでも `direct` が `started` になる理由を、`defer` 文の評価時点と実行時点を分けて説明してください。

<details>
<summary>ヒント</summary>

Go 言語仕様で `defer` 文を探します。関数呼び出しそのものと、関数値・引数の評価について、別々に書かれている箇所に注目してください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 言語仕様の Defer statements](https://go.dev/ref/spec#Defer_statements) を開き、`defer` 文で評価されるものと、関数呼び出しが実行される時点を確認する。
2. [Go Blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) で、引数の評価時点を小さな例と照合する。

**答え**

`defer` 文に書いた関数値と引数は、その `defer` 文を実行した時点で評価されます。`fmt.Println("direct:", status)` の `status` は、この時点では `started` です。関数呼び出し自体は `main` から戻るときまで遅延しますが、渡す引数の値はすでに保存されているため、`direct: started` になります。

</details>

---

## 設問 2: なぜ `closure` は完了後の値で、先に表示されるのか？

無名関数側は `completed` を表示し、しかも `direct` より先に表示されます。無名関数が `status` を読む時点と、複数の `defer` の実行順を調べてください。

<details>
<summary>ヒント</summary>

Go 言語仕様では、無名関数を `Function literals` と呼びます。この節で外側の変数をどう参照するかを調べ、設問 1 の「関数値と引数」の評価規則と分けて考えます。続けて `Defer statements` で、複数の `defer` の順番を確認しましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 言語仕様の Function literals](https://go.dev/ref/spec#Function_literals) を開き、関数リテラルと外側の関数が参照する変数を共有する規則を確認する。
2. [Go 言語仕様の Defer statements](https://go.dev/ref/spec#Defer_statements) で、複数の遅延呼び出しが逆順に実行されることを確認する。
3. [Go Blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) の例と、[実行例](https://go.dev/play/p/XEEZ1ax4se2) の出力順・値を照合する。

**答え**

関数リテラルは外側の関数と `status` 変数を共有します。無名関数に `status` を引数として渡して固定してはいないため、`closure` の本体は遅延呼び出しが実行される時点の変数を参照します。その時点では代入済みで `completed` です。また、遅延呼び出しは積み重ねた逆順に実行されるため、後から登録した無名関数が先に表示されます。

`direct` が古い値を出すのは、引数が `defer` 文の実行時に評価・保存されていたからです。値をいつ固定したいのかを決めて、無名関数か引数かを選びます。

</details>
