# 初級: defer の引数評価と実行順を調べよう

CSV インポートの処理で、開始・完了の状態を監査ログへ残しています。処理は完了しているのに、後片付けのログだけ `started` と出てしまいました。ログを確実に出すために `defer` を使っています。

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

Go 1.26.4 での実行結果です。

```text
closure: completed
direct: started
```

---

## 設問 1: `defer` した関数呼び出しの引数はいつ評価されるのか？

`status` を `completed` へ代入するのは、最初の `defer fmt.Println("direct:", status)` より後です。それでも `direct` が `started` になる理由を、`defer` 文の評価時点と実行時点を分けて説明してください。

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

無名関数側は `completed` を表示し、しかも `direct` より先に表示されます。無名関数が `status` を読む時点と、複数の `defer` の実行順を調べ、冒頭の監査ログが古い状態を記録した理由を説明してください。

<details>
<summary>ヒント</summary>

設問 1 で確認した「関数値と引数」の規則を、無名関数の本体が変数を参照する場合と比べます。続けて、同じ仕様節で複数の `defer` の順番を確認しましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 言語仕様の Defer statements](https://go.dev/ref/spec#Defer_statements) で、複数の遅延呼び出しが逆順に実行されることを確認する。
2. [Go Blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) の遅延実行の例を、`status` を無名関数の本体で読むコードと比べる。
3. [実行例](https://go.dev/play/p/XEEZ1ax4se2) の出力順と値を照合する。

**答え**

無名関数に `status` を引数として渡してはいないので、`closure` の本体は遅延呼び出しが実行される時点で変数を参照します。その時点では代入済みで `completed` です。また、遅延呼び出しは積み重ねた逆順に実行されるため、後から登録した無名関数が先に表示されます。

つまり、冒頭の `direct` ログが古いのは「後片付けで出力したから」ではなく、引数が開始時に評価・保存されていたからです。完了時の状態を残す必要がある監査ログでは、値をいつ固定したいのかを決めて、無名関数か引数かを選びます。

</details>

---

## 調査の入り口

1. [04-deep-dive の調べ方](../../README.md)を開き、Go 言語仕様を起点にします。
2. [Go 言語仕様: Defer statements](https://go.dev/ref/spec#Defer_statements) で評価時点と実行順を調べます。
3. [Go Blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) で観測結果を照合します。
