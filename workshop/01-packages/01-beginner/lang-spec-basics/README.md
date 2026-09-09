[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [01-packages の調べ方](../../README.md)

# Go の for 文と switch 文を仕様書で調べよう

**実行環境**: ブラウザだけ。Go のインストールは不要です。

レビュー中に `for` 文と `switch` 文を見かけました。実行結果から Go 言語仕様の該当箇所を探してみましょう。

<details>
<summary>調査の入り口</summary>

まず [01-packages の調べ方](../../README.md) を開き、言語仕様の節の探し方を確かめます。

そのうえで、次のどれかから入ります。

- [Go Documentation](https://go.dev/doc/) — Language Specification から Go 言語仕様を開く入口
- [Go 言語仕様](https://go.dev/ref/spec) — `for` 文と `switch` 文の規則は、この 1 ページにまとまった仕様書の中にある

</details>

---

## 設問 1: for 文

他言語で見かける `for value in values` のような反復を、Go ではどう書くのでしょうか。次のコードでスライス `values` の要素を順に表示する構文を観察してください。そのうえで、`range` が返す二つの値と `_` の役割まで調べてください。

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

- まず `for value in values` と `for _, value := range values` をそれぞれ試し、エラーになる方と動く方を比べます。
- 仕様書の目次で「For statements」を開いたら、親の構文定義だけで止まらず、下位の「For statements with range clause」へ進みます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/bJ7XfrbuNHo) を実行し、`values` の要素が順に表示されることを観測する。
2. [Go 言語仕様](https://go.dev/ref/spec) を開き、目次から「For statements」を選んで `ForStmt` の 3 形式を確認する。
3. 直下の [For statements with range clause](https://go.dev/ref/spec#For_range) へ進み、`RangeClause` の構文と、スライスを反復したときの第 1・第 2 の値を確認する。

**答え**

`for` 文には、条件だけを書く形式、初期化・条件・後処理を書く形式、`range` 節を使う形式があります。`for value in values` という構文はなく、コレクションを反復するときは `range` を使います。

```
ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
```

スライスに対する `range` の第 1 の値はインデックス、第 2 の値は要素です。冒頭の `for _, value := range values` は、不要なインデックスを空白識別子 `_` で捨て、要素だけを `value` で受け取っています。

</details>

---

## 設問 2: switch 文

Go の switch 文は、C 言語の switch 文と違って、break が不要なことは知っています。
次のコードでは `x == 1` のとき `one` だけが表示されます。C 言語の switch 文のように、1 つの case にマッチしたら次の case も実行させたいです。どうすればよいでしょうか？

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

- `x == 1` で `one` だけが出る状態を確認してから、仕様書の「Switch statements」直下にある「Expression switches」へ進みます。
- 次の case の式をもう一度判定するのか、無条件に処理を移すのかにも注目します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/FhO9_Qu5-ca) を実行し、`x == 1` では `one` だけが表示されることを観測する。
2. [Go 言語仕様](https://go.dev/ref/spec) を開き、目次から「Switch statements」を選ぶ。
3. 下位の [Expression switches](https://go.dev/ref/spec#Expression_switches) で、暗黙の break と `fallthrough` の規則を確認する。

**答え**

`case 1` の最後に `fallthrough` を置くと、次の case の式を判定せず、その最初の文へ処理を移せます。

```go
package main

import "fmt"

func main() {
	x := 1
	switch x {
	case 1:
		fmt.Println("one")
		fallthrough
	case 2:
		fmt.Println("two")
	}
}
```

[Go Playground で動かす](https://go.dev/play/p/fOZxxOIZNDA) と、次のように両方の case の処理が実行されます。

```text
one
two
```


</details>

---

<details>
<summary>こぼれ話: 仕様書内で迷わないために</summary>

[Go 言語仕様](https://go.dev/ref/spec) は 1 ページの HTML に目次と各節がまとまっています。親の節には構文定義だけがあり、具体的な規則が下位節に分かれていることがあります。今回の `range` なら `#For_range`、式 switch なら `#Expression_switches` の節リンクまで共有すると、班の全員が同じ根拠を開けます。

</details>
