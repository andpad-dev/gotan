[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# slog.Value はなぜ == で比較できないのか

**実行環境**: ブラウザだけ。Go のインストールは不要です。

次のコードはコンパイルできません。

```go
package main

import (
	"fmt"
	"log/slog"
)

func main() {
	v1 := slog.StringValue("gopher")
	v2 := slog.StringValue("gopher")
	fmt.Println(v1 == v2) // コンパイルエラー.
	// fmt.Println(v1.Equal(v2)) // こちらは true.
}
```

（[Go Playground で確かめる](https://go.dev/play/p/MPeWX_kJChR)）

コンパイルエラー:

```
invalid operation: v1 == v2 (struct containing [0]func() cannot be compared)
```

なんでこうなってるの？背景を調べて、その仕組みを言語仕様から説明しなさい。

<details>
<summary>調査の入り口</summary>

まず [04-deep-dive の調べ方](../../README.md) を開き、仕様・実装・設計背景をたどる順番を確かめます。

そのうえで、次のどれかから入ります。

- [Go Documentation](https://go.dev/doc/) — 標準ライブラリの `log/slog` パッケージを開く入口
- [The Go Programming Language Specification](https://go.dev/ref/spec) — Go 言語仕様。`==` の規則もここで定義されている
- [go1.27.0 の src/log/slog/value.go](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/value.go;l=21) — `slog.Value` の型定義
- [`==` を禁止した CL 479516](https://go-review.googlesource.com/c/go/+/479516) と [Issue #56345](https://github.com/golang/go/issues/56345) — `==` を禁止した変更と、その背景の issue

</details>

---

## 設問 1: 比較を禁止している仕掛けを特定しよう

slog.Value の型定義を読み、`==` を禁止している仕掛けを見つけましょう。

- https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/value.go;l=21

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、標準ライブラリの `log/slog` パッケージを開く。
2. [go1.27.0 の src/log/slog/value.go](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/value.go;l=21) を開く。
3. `type Value struct` の定義を読む。

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

`[0]func()` 自体のサイズと、フィールドを先頭・末尾に置いた構造体のサイズを区別してください。
`slog.Value` ではこの仕掛けがメモリ消費を増やさない理由を、仕様とコンパイラ実装の両方から説明しましょう。

次のコードを実行し、ゼロ長配列と `slog.Value` のサイズを確認してください。

```go
package main

import (
	"fmt"
	"log/slog"
	"unsafe"
)

type withArrayFirst struct {
	_   [0]func()
	num uint64
	any any
}

type withoutArray struct {
	num uint64
	any any
}

type withArrayLast struct {
	num uint64
	any any
	_ [0]func()
}

func main() {
	fmt.Println(unsafe.Sizeof([0]func(){}))
	fmt.Println(unsafe.Sizeof(withArrayFirst{}))
	fmt.Println(unsafe.Sizeof(withoutArray{}))
	fmt.Println(unsafe.Sizeof(withArrayLast{}))
	fmt.Println(unsafe.Sizeof(slog.Value{}))
}
```

（[Go Playground で動かす](https://go.dev/play/p/w9nhI8d9nnl)）

実行結果:

```text
0
24
24
32
24
```

<details>
<summary>ヒント</summary>

- まず仕様の「Size and alignment guarantees」で、ゼロ長配列そのものについて保証される範囲を確認します。
- 末尾配置だけ 32 バイトになる理由は、Go 1.27.0 のコンパイラソースで `zero-sized field` を検索してください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/w9nhI8d9nnl) を実行し、先頭配置と末尾配置のサイズ差を観測する。
2. 仕様の「[Size and alignment guarantees](https://go.dev/ref/spec#Size_and_alignment_guarantees)」を読む。
3. [Go 1.27.0 の `cmd/compile/internal/types/size.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/types/size.go;l=516) で、末尾のゼロサイズフィールドへ追加するパディングと [Issue #9401](https://github.com/golang/go/issues/9401) への参照を確認する。

**答え**

仕様が長さ 0 の配列のサイズを 0 と保証しています。

> A struct or array type has size zero if it contains no fields (or elements, respectively) that have a size greater than zero.

仕様が保証するのは、`[0]func()` という型自体のサイズが 0 であることです。構造体へ追加しても総サイズが増えないことまでは保証していません。

Go の gc コンパイラは、ゼロサイズフィールドが非ゼロサイズ構造体の末尾にある場合、そのフィールドのアドレスが次のヒープオブジェクトを指さないようパディングを加えます。実行すると、先頭配置はフィールドなしと同じ 24 バイトですが、末尾配置は 32 バイトです。`slog.Value` は `_ [0]func()` を先頭に置くことで、比較不可能という型の性質を追加しつつ、Go 1.27.0 の実装では総サイズを 24 バイトに保っています。

</details>

---

## 設問 4: そもそもなぜ == を禁止したいのか

仮に `==` が使えたとして、slog.Value の比較は期待どおりに動くのでしょうか。
型定義の各フィールドのコメントを読んで、禁止したい理由を考えましょう。
また、slog.Value 同士を比較したいときはどうすればよいかも調べましょう。

次のコードで、文字列の比較と、比較不可能な値を含む `AnyValue` の比較をそれぞれ試してください。

```go
package main

import (
	"fmt"
	"log/slog"
	"strings"
	"unsafe"
)

func main() {
	literal := "X"
	computed := strings.ToUpper("x")
	fmt.Println("strings equal:", literal == computed)
	fmt.Println("data pointers equal:", unsafe.StringData(literal) == unsafe.StringData(computed))
	fmt.Println("Value.Equal string:", slog.StringValue(literal).Equal(slog.StringValue(computed)))

	sliceValue := slog.AnyValue([]int{1})
	func() {
		defer func() {
			fmt.Println("Value.Equal slice panicked:", recover() != nil)
		}()
		fmt.Println(sliceValue.Equal(sliceValue))
	}()
}
```

（[Go Playground で動かす](https://go.dev/play/p/d6m0c3sgC6F)）

実行結果:

```text
strings equal: true
data pointers equal: false
Value.Equal string: true
Value.Equal slice panicked: true
```

<details>
<summary>ヒント</summary>

- 文字列の Value が「何と何の組み合わせ」で保持されているかに注目
- `any` フィールドには何でも入る。比較できない値が入った interface 同士を `==` するとどうなるか、仕様の「Comparison operators」の interface の項を読もう
- `value.go` の履歴から、`disallow ==` を追加した変更と関連 issue をたどると設計理由を確認できます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/d6m0c3sgC6F) を実行し、内容が同じ文字列のデータポインタが異なることと、`Value.Equal` の通常の結果・panic を観測する。
2. [go1.27.0 の `Value` 定義](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/value.go;l=21) の `num` / `any` フィールドのコメントを読む。
3. 仕様の「[Comparison operators](https://go.dev/ref/spec#Comparison_operators)」の interface の項を読む。
4. `value.go` の履歴から [`==` を禁止した CL 479516](https://go-review.googlesource.com/c/go/+/479516) と [Issue #56345](https://github.com/golang/go/issues/56345) をたどり、変更理由と文字列の例を読む。
5. [Value.Equal](https://pkg.go.dev/log/slog#Value.Equal) の説明と [Go 1.27.0 の実装](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/value.go;l=420) を確認する。

**答え**

仮に `==` が許されても、slog.Value の比較は期待どおりに動かないためです。

**文字列の比較が壊れる。**
フィールドのコメントを読むと、文字列の Value は「長さを `num` に、先頭ポインタを `any`（`stringptr` 型）に」分解して保持しています。
この表現のままフィールド単位で比較すると文字列の内容ではなくポインタが比較されます。実験のリテラル `"X"` と実行時に作った `strings.ToUpper("x")` は内容が同じでもデータポインタが異なるため、内容ではなく内部表現を比べる方式では false になります。これは CL 479516 が挙げる反例と同じ条件です。

**実行時 panic の危険がある。**
`any` フィールドには任意の値が入ります。
仕様にはこうあります。

> A comparison of two interface values with identical dynamic types causes a run-time panic if that type is not comparable.

比較不可能な値（スライスなど）を持つ Value 同士を `==` すると、コンパイルは通っても実行時に panic します。

このように「コンパイルは通るが結果が信頼できない・panic し得る」比較を、型レベルで禁止するのが `_ [0]func()` の役割です。
内容を比較するには [`Value.Equal`](https://pkg.go.dev/log/slog#Value.Equal) を使えますが、`KindAny` や `KindLogValuer` で比較不可能な値を保持している場合は `Value.Equal` 自体も panic し得ます。安全に扱うには、比較可能な値だけを `AnyValue` に渡すなど、値の型に応じた設計が必要です。

</details>

---

<details>
<summary>こぼれ話: フィールドが先頭に置かれている理由</summary>

`_ [0]func()` を構造体の末尾に置くと、パディングが入って構造体が大きくなります（実行した結果: 先頭なら 24 バイト、末尾だと 32 バイト）。
末尾のゼロサイズフィールドへのポインタが構造体の外を指してしまうのを防ぐためです。背景は [Issue #9401](https://github.com/golang/go/issues/9401) と [Go 1.27.0 のコンパイラ実装](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/types/size.go;l=516) から追えます。

</details>
