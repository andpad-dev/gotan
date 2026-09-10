[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# 中級: 3添字スライス式で通知対象を分離しよう

![実行環境: 指定なし](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-%E6%8C%87%E5%AE%9A%E3%81%AA%E3%81%97-9E9E9E)

案件の担当者一覧から先頭 2 人へ通知し、当番担当者を 1 人だけ追加するバッチがあります。通知用の一覧を作ったつもりが、もとの案件データの 3 人目まで当番担当者に置き換わってしまいました。画面表示や後続の同期処理に影響させてはいけません。

「共有している担当者一覧に影響させず、通知対象だけへ当番を追加する」ことをやりたいです。どういうふうにやればいいか調べよう。

[Go Playground で実行](https://go.dev/play/p/-4j2pYJbiJT) し、単純なスライス式と、三つの添字を使うスライス式を比べてください。

```go
package main

import "fmt"

func main() {
	shared := []string{"aya", "bo", "chi"}
	view := shared[:2]
	view = append(view, "on-call")
	fmt.Printf("shared=%q\n", shared)
	fmt.Printf("view=%q\n", view)

	isolatedSource := []string{"aya", "bo", "chi"}
	isolated := isolatedSource[:2:2]
	isolated = append(isolated, "on-call")
	fmt.Printf("isolatedSource=%q\n", isolatedSource)
	fmt.Printf("isolated=%q\n", isolated)
}
```

Go 1.26.4 での実行結果です。

```text
shared=["aya" "bo" "on-call"]
view=["aya" "bo" "on-call"]
isolatedSource=["aya" "bo" "chi"]
isolated=["aya" "bo" "on-call"]
```

---

## 設問 1: もとの一覧の 3 人目はなぜ置き換わったのか？

`view := shared[:2]` の直後、`view` の長さは 2 です。`append` した `"on-call"` が、なぜ `shared` の 3 人目 `"chi"` を置き換えたのでしょうか。`len` と `cap`、および `append` の規則で説明してください。

<details>
<summary>ヒント</summary>

スライス式の仕様で、添字を二つだけ指定したときの容量を確認します。次に `append` の仕様で、容量が足りる場合と足りない場合の違いを調べます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 言語仕様の Slice expressions](https://go.dev/ref/spec#Slice_expressions) を開き、二つの添字を使うスライス式の長さと容量を確認する。
2. [Go 言語仕様の Appending and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices) で、追加後の長さが容量を超えない場合の `append` の規則を読む。
3. [実行例](https://go.dev/play/p/-4j2pYJbiJT) の `shared` と `view` の出力を照合する。

**答え**

`shared` の長さと容量はいずれも 3 で、`shared[:2]` の長さは 2、容量は 3 です。したがって `append(view, "on-call")` の結果の長さ 3 は容量に収まります。仕様では、この場合の `append` は同じ基底配列を再利用できます。追加値は基底配列の添字 2 に書かれるため、同じ配列を見ている `shared` の 3 人目も `"on-call"` になります。

</details>

---

## 設問 2: 通知用の追加だけを分離するには？

`isolatedSource[:2:2]` では、追加後も `isolatedSource` が `"chi"` のままです。三つ目の添字は何を制限し、なぜ `append` の結果だけが当番担当者を含む別の一覧になるのでしょうか。

<details>
<summary>ヒント</summary>

スライス式の仕様で、三つの添字を使う形式の最後の添字が何を表すかを調べます。その容量と、追加後に必要な長さを比べてください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 言語仕様の Slice expressions](https://go.dev/ref/spec#Slice_expressions) で、三つの添字を使う完全スライス式の容量が最後の添字で決まることを確認する。
2. [Go 言語仕様の Appending and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices) で、容量が足りない `append` が新しい基底配列を割り当てることを確認する。
3. 実装も追いたければ、固定版の [runtime の `growslice`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/runtime/slice.go;l=178) を開き、容量を増やす処理の入口を確認する。

**答え**

完全スライス式 `isolatedSource[:2:2]` は、長さを 2、容量も 2 にします。そこへ 1 人追加すると必要な長さは 3 で容量を超えるため、`append` は新しい基底配列を割り当てます。その新しい配列だけに `"on-call"` が追加されるので、通知用の `isolated` は増え、もとの `isolatedSource` は変わりません。

</details>

---

## 設問 3: 業務コードではどこまで分離を明示するか？

通知用一覧がもとの案件データを変更しないことを、テストでどう確かめますか。また、完全スライス式と、最初から新しいスライスへ要素をコピーする方法は、それぞれどんな意図をコードに表せるでしょうか。

<details>
<summary>ヒント</summary>

まず、実行例にある二組の出力をテストの期待値として言い換えます。その後、仕様の `copy` と `append` の説明を読み、どの時点で別の基底配列が必要になるかを比較してください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 言語仕様の Appending and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices) で、`append` と `copy` の結果を確認する。
2. [Go 言語仕様の Slice expressions](https://go.dev/ref/spec#Slice_expressions) に戻り、完全スライス式で容量を限定する意味を照合する。
3. [実行例](https://go.dev/play/p/-4j2pYJbiJT) のように、元の一覧が `["aya" "bo" "chi"]` のまま、通知用だけが当番を含むことを期待値にする。

**答え**

テストでは、通知用の一覧へ追加した後も、元の案件の担当者が `aya, bo, chi` のままであることを確認します。同時に、通知用だけが `aya, bo, on-call` になることを確認すると、今回の事故を再現できます。

完全スライス式は「次の `append` でこの範囲を越えて書き換えない」と容量で表す方法です。一方、最初から `append([]string(nil), source[:2]...)` のようにコピーすれば、コピーした時点で別の基底配列を持つことを明示できます。どちらも要素そのものは浅くコピーします。どちらを選ぶかは、容量制限を活用する意図と、最初から独立した一覧を作る意図のどちらを読み手へ強く伝えたいかで決めます。

</details>

---

## 調査の入り口

1. [04-deep-dive の調べ方](../../README.md)を開き、Go 言語仕様を起点にします。
2. [Go 言語仕様: Slice expressions](https://go.dev/ref/spec#Slice_expressions) で `len` と `cap` を調べます。
3. [Go 言語仕様: Appending and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices) で追加時の規則を確かめます。
