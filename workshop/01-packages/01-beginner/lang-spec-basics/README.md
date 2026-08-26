# for 文の構文を仕様書で調べよう

あなたは超久々に Go コーディングをします。
`for` 文を見かけました。どんなものか調べてみましょう。

言語仕様についてド忘れしていることもあるので、Go 言語仕様を読んでみましょう。

## 設問 1: for文

カウンター変数を使ったものは当然あるとして `in` のような構文があったはずですが、Go ではどう書くのでしょうか？ 次のコードで `values` の要素を順に表示する書き方を観察してから、`for value in values` と書けるか調べてみましょう。

```go
package main

import "fmt"

func main() {
	values := []string{"a", "b"}
	for _, value := range values {
		fmt.Println(value)
	}
}
```

（[Go Playground で動かす](https://go.dev/play/p/bJ7XfrbuNHo)）

実行結果:

```text
a
b
```

<details>
<summary>ヒント</summary>

- まず `values := []string{"a", "b"}` を用意し、`for value in values` と `for _, value := range values` をそれぞれ試してみましょう。エラーになる方と動く方を比べてから、仕様書の「For statements」を開きます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/bJ7XfrbuNHo) を実行し、`values` の要素が順に表示されることを観測する。
2. https://go.dev/ref/spec を開き、目次から「For statements」を選ぶ。
3. `for` 文の構文を確認する。

**答え**

- 以下のように、`for` 文は 3 種類の構文を持っています。
- このうち、`in` のような構文はありませんが、 `range` キーワードを使った構文があり、これが `in` のような意味合いを持っています。

```
ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
```

[do it](https://go.dev/play/p/WCZlhaPrgG_p)

</details>

---

## 設問 2: switch文

Go の switch 文は、C 言語の switch 文と違って、break が不要なことは知っています。
次のコードでは `x == 1` のとき `one` だけが表示されます。C 言語の switch 文のように、1 つの case にマッチしたら、次の case も実行されるようにするにはどうすればよいでしょうか？

```go
package main

import "fmt"

func main() {
	x := 1
	switch x {
	case 1:
		fmt.Println("one")
	case 2:
		fmt.Println("two")
	}
}
```

（[Go Playground で動かす](https://go.dev/play/p/FhO9_Qu5-ca)）

実行結果:

```text
one
```

<details>
<summary>ヒント</summary>

- `x := 1` とし、`case 1` で `one`、`case 2` で `two` を表示する短い switch を作ります。`x` が 1 のとき `two` も表示したい場合に、仕様書で必要な構文を探してみましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/FhO9_Qu5-ca) を実行し、`x == 1` では `one` だけが表示されることを観測する。
2. https://go.dev/ref/spec を開き、目次から「Switch statements」を選ぶ。
3. `fallthrough` キーワードの説明を確認する。

**答え**

- `fallthrough` キーワードを使うことで、C 言語の switch 文のように、1 つの case にマッチしたら、次の case も実行されるようにできます。

```go
switch x {
case 1:
    fmt.Println("one")
    fallthrough
case 2:
    fmt.Println("two")
}
```

[just do it](https://go.dev/play/p/HumLQAbZ-eQ)


</details>

---

<details>
<summary>こぼれ話: </summary>

[https://go.dev/ref/spec](https://go.dev/ref/spec)（Go言語仕様書）は、プログラミング言語の仕様書としては異例なほど短く、1ページのHTMLにまとまっているのが最大の特徴です。  
C++やJavaの仕様書が分厚い辞書のようであるのに対し、Goの仕様書は「休日の午後だけで読み切れる」ことを意図して設計されています。

</details>

---

## 調査の入り口

- https://go.dev/ref/spec
