[進行ガイド・シナリオ一覧](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# slog.Value はなぜ == で比較できないのか

次のコードはコンパイルできません。

```go
v1 := slog.StringValue("gopher")
v2 := slog.StringValue("gopher")
fmt.Println(v1 == v2)
// invalid operation: v1 == v2 (struct containing [0]func() cannot be compared)
```

（Go Playground で確かめる: https://go.dev/play/p/MPeWX_kJChR ）

なんでこうなってるの？背景を調べて、その仕組みを言語仕様から説明しなさい。

## 設問 1: 比較を禁止している仕掛けを特定しよう

slog.Value の型定義を読み、`==` を禁止している仕掛けを見つけましょう。

- https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/log/slog/value.go;l=22

<details>
<summary>答え</summary>

**調査ルート**

1. [go1.26.5 の src/log/slog/value.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/log/slog/value.go;l=22) を開く。
2. `type Value struct` の定義を読む。

**答え**

定義の先頭に、この 1 行があります。

```go
type Value struct {
	_ [0]func() // disallow ==
	...
}
```

コメントが答えそのものです。
`_ [0]func()`（長さ 0 の、関数の配列）というフィールドが、`==` を禁止する仕掛けです。
フィールド名がブランク識別子 `_` なので、初期化も参照も不要で、型の性質だけを構造体に与えています。

</details>

---

## 設問 2: なぜその仕掛けで比較できなくなるのか、言語仕様から説明しよう

設問 1 で見つけたフィールドがあると、なぜ `==` がコンパイルエラーになるのでしょうか。
[The Go Programming Language Specification](https://go.dev/ref/spec) を根拠に説明しましょう。

<details>
<summary>ヒント</summary>

- 仕様の「Comparison operators」の節に、型ごとの比較可能性のルールが列挙されている
- 「構造体」「配列」「関数」それぞれの比較可能性を確認しよう

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://go.dev/ref/spec の「[Comparison operators](https://go.dev/ref/spec#Comparison_operators)」を読む。
2. struct / array / function それぞれの比較可能性のルールを拾う。

**答え**

仕様には次の 3 つのルールがあります。

> Struct types are comparable if all their field types are comparable.

> Array types are comparable if their array element types are comparable.

> Slice, map, and function types are not comparable.

これを連鎖させると答えになります。

1. 関数型 `func()` は比較不可能。
2. 要素型が比較不可能なので、配列型 `[0]func()` も比較不可能（長さは無関係）。
3. 比較不可能なフィールドを含むので、構造体 `slog.Value` も比較不可能。

実際に `==` を書くとコンパイルエラーになります: https://go.dev/play/p/MPeWX_kJChR

```
invalid operation: v1 == v2 (struct containing [0]func() cannot be compared)
```

</details>

---

## 設問 3: なぜ「長さ 0 の配列」なのか

この仕掛けのフィールドが slog.Value のメモリ消費を増やさないことを、仕様を根拠に説明しましょう。
`unsafe.Sizeof` を使って実測でも確かめてみましょう。

<details>
<summary>ヒント</summary>

- 仕様の「Size and alignment guarantees」の節を読もう

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. 仕様の「[Size and alignment guarantees](https://go.dev/ref/spec#Size_and_alignment_guarantees)」を読む。
2. `unsafe.Sizeof` で実測する。

**答え**

仕様が長さ 0 の配列のサイズを 0 と保証しています。

> A struct or array type has size zero if it contains no fields (or elements, respectively) that have a size greater than zero.

つまり `[0]func()` は「比較不可能」という型の性質だけを持ち込み、メモリは 1 バイトも消費しません。
実測でも、このフィールドの有無で構造体のサイズは変わりません: https://go.dev/play/p/SVs4LjlKXGo

```go
fmt.Println(unsafe.Sizeof([0]func(){})) // 0
fmt.Println(unsafe.Sizeof(withArray{})) // 24 (フィールドあり)
fmt.Println(unsafe.Sizeof(withoutArray{})) // 24 (フィールドなし)
```

</details>

---

## 設問 4: そもそもなぜ == を禁止したいのか

仮に `==` が使えたとして、slog.Value の比較は期待どおりに動くのでしょうか。
型定義の各フィールドのコメントを読んで、禁止したい理由を考えましょう。
また、slog.Value 同士を比較したいときはどうすればよいかも調べましょう。

<details>
<summary>ヒント</summary>

- 文字列の Value が「何と何の組み合わせ」で保持されているかに注目
- `any` フィールドには何でも入る。比較できない値が入った interface 同士を `==` するとどうなるか、仕様の「Comparison operators」の interface の項を読もう

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. value.go の `num` / `any` フィールドのコメントを読む。
2. 仕様の「Comparison operators」の interface の項を読む。
3. pkg.go.dev/log/slog で Value のメソッド一覧から比較用のメソッドを探す。

**答え**

仮に `==` が許されても、slog.Value の比較は期待どおりに動かないためです。

**文字列の比較が壊れる。**
フィールドのコメントを読むと、文字列の Value は「長さを `num` に、先頭ポインタを `any`（`stringptr` 型）に」分解して保持しています。
この表現のまま `==` すると文字列の内容ではなくポインタが比較されるので、同じ `"gopher"` 同士でも false になり得ます。

**実行時 panic の危険がある。**
`any` フィールドには任意の値が入ります。
仕様にはこうあります。

> A comparison of two interface values with identical dynamic types causes a run-time panic if that type is not comparable.

比較不可能な値（スライスなど）を持つ Value 同士を `==` すると、コンパイルは通っても実行時に panic します。

このように「コンパイルは通るが結果が信頼できない・panic し得る」比較を、型レベルで禁止するのが `_ [0]func()` の役割です。
代わりに、内容を正しく比較する [`Value.Equal`](https://pkg.go.dev/log/slog#Value.Equal) メソッドが用意されています。

</details>

---

<details>
<summary>こぼれ話: フィールドが先頭に置かれている理由</summary>

`_ [0]func()` を構造体の末尾に置くと、サイズ 0 のはずがパディングが入って構造体が大きくなります（実測: 先頭なら 24 バイト、末尾だと 32 バイト）。
末尾のゼロサイズフィールドへのポインタが構造体の外を指してしまうのを防ぐためで、ゼロサイズフィールドは先頭に置くのが定石です。

</details>

---

## 調査の入り口

- https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/log/slog/value.go;l=22
- https://go.dev/ref/spec
