# 自作の型にジェネリックメソッドで汎用ヘルパを生やそう

自作の型にジェネリックメソッドを生やしたいです。どういうふうにやればいいか調べよう。

社内ユーティリティに、独自コンテナ型 `Slice[T]` があります。
値を別の型に写す変換ヘルパを、コンテナに **メソッドとして** 生やしたいです。

```go
type Slice[T any] struct {
	items []T
}

// これを書きたい
func (s Slice[T]) Map[F any](f func(T) F) Slice[F] { ... }
```

いま安定版として使っている Go 1.26.5 でこのコードをコンパイルしようとすると、次のように怒られます。

```
syntax error: method must have no type parameters
```

Go 1.27 のリリースノートには「ジェネリックメソッド」が記載されていますが、現時点ではリリースノート自体が Draft です。ここでは安定版 Go 1.26 の仕様と、`gotip` / Playground の `?v=gotip` で試せる将来仕様を混同しないように調べます。

- 仕様書のどこがどう変わったのか
- なぜ Go 1.26 までは書けなかったのか
- インタフェースメソッドではどうなのか

を一次情報で押さえてから、実際に手を動かして書けるようになりましょう。

---

## 設問 1: 仕様書のどこが変わったのか調べよう

Go 1.27 の Draft リリースノートと proposal から、「method に何が書けるようになる予定か」を特定しましょう。現行仕様書の EBNF と、提案されている EBNF の違いまで見てみます。

<details>
<summary>ヒント</summary>

- カテゴリの逆引き手順「[最新バージョンの機能を調査したい](../../README.md)」の通り、まず [Go 1.27 リリースノート](https://go.dev/doc/go1.27) を開く
- Changes to the language セクションに「generic methods」の段落があるはず
- リリースノートの HTML には `go.dev/issue/<番号>` の形で関連 issue が埋め込まれている（開発者ツールでソースを見ると拾える）
- 仕様書側は `Method declarations` セクションを開いて、EBNF が旧版とどう変わっているか比べる

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート の Changes to the language 節](https://go.dev/doc/go1.27#language) を開き、`generic methods` の段落を読む。
2. その段落に埋め込まれた `go.dev/issue/77273` を開いて [proposal #77273 "spec: generic methods for Go"](https://go.dev/issue/77273) の Proposal 節に飛ぶ。旧新の EBNF が並記されている。
3. [現行仕様書の Method declarations](https://go.dev/ref/spec#Method_declarations) セクションを読み、Draft リリースノート・proposal に書かれた将来仕様と区別する。

**答え**

Draft のリリースノートには次のように書かれています。

> Go 1.27 now supports generic methods: a method declaration may declare its own type parameters. This widely anticipated change allows adding generic functions within the namespace of a particular data type where before one had to declare such functions with a scope of the entire package. Note that methods of interfaces may not declare type parameters nor can interface methods be implemented by generic methods.

現行の安定版仕様書 [Method declarations](https://go.dev/ref/spec#Method_declarations) は、まだ次の Go 1.26 の EBNF です。

- **現行 (Go 1.26)**
  ```
  MethodDecl = "func" Receiver MethodName Signature [ FunctionBody ] .
  ```
- proposal #77273 が示す **将来仕様 (Go 1.27 の Draft)**
  ```
  MethodDecl = "func" Receiver MethodName [ TypeParameters ] Signature [ FunctionBody ] .
  ```

proposal では `MethodName` と `Signature` の間に `[ TypeParameters ]` を挿入する案になっています。関数宣言 (`FunctionDecl`) の型パラメータ位置と揃った形です。ただし、これは現行の安定版仕様書に反映された仕様ではありません。

Go 1.26.5 でこの宣言を実行すると、実際には次の構文エラーになります。ジェネリックメソッドを試す場合は、`gotip` または Playground の `?v=gotip` を使います。

```
syntax error: method must have no type parameters
```

</details>

---

## 設問 2: なぜインタフェースメソッドでは依然として書けないか、そして generic method はインタフェースを実装するか

Draft のリリースノートには含みのある一文があります。

> methods of interfaces may not declare type parameters **nor can interface methods be implemented by generic methods**.

つまり Draft で示される Go 1.27 の案でも「インタフェースメソッドに型パラメータは書けない」うえに、「ジェネリックメソッドでインタフェースを実装することもできない」。なぜこういう線引きになったのでしょうか。次のコードを Playground で試して、コンパイラの言い分も一緒に確認しましょう。

```go
package main

import "io"

type Reader struct{}

func (*Reader) Read[E any](p []E) (int, error) { return 0, nil }

func main() {
	var _ io.Reader = (*Reader)(nil)
}
```

（Go Playground で動かす: https://go.dev/play/p/Lqn-PC8jg9E?v=gotip 、Dev/gotip を選んで Run）

<details>
<summary>ヒント</summary>

- proposal #77273 の Background 節に「なぜ歴史的に禁止していたのか」が短く書かれている
- 同じ proposal の Examples 節に、まさに `Reader.Read[E any]` と `io.Reader` の対比が載っている
- Background から辿れる [Type Parameters Proposal の No parameterized methods 節](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#No-parameterized-methods) には、より詳しい議論がある

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [proposal #77273](https://go.dev/issue/77273) の Background 節を読み、歴史的な禁止理由と "A change of view" の考え方を押さえる。
2. 同 proposal の Examples 節で、`Reader.Read[E any]` と `io.Reader` の関係を確認する。
3. 手元 (`gotip`) か Playground で実際にコンパイルして、compiler のエラーメッセージを見る。
4. 深追いしたい人向け: [Type Parameters Proposal の No parameterized methods 節](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#No-parameterized-methods) で当時の議論を読む。

**答え**

**1. なぜインタフェースメソッドには型パラメータを許さないのか**

proposal Background の該当段落がそのまま答えです。

> Go doesn't support such generic interface methods because we don't know how to implement (calls of) them, or at least we don't know how to implement them efficiently. Specifically, because Go doesn't require a concrete type to declare the interfaces it implements, and instead this is a dynamic property, it cannot be known at compile time which of the infinite possible instantiations of concrete methods will be needed at run time.

要点は「Go の interface 充足は動的な性質」であることです。ある型がどの interface を満たすかをコンパイル時に宣言しないので、コンパイラは「実行時にどんな型引数で generic interface method が呼ばれるか」を事前に列挙できません。効率的な実装が難しいので、Go 1.27 でも interface method には型パラメータを持たせない、という判断は据え置きになりました。

**2. なぜそれでも concrete method には許すことにしたのか（"A change of view"）**

proposal の見方の転換はこうです。

> concrete methods are not just a means for implementing interfaces. A method is a function associated with a type, and accessed through the namespace of that type. Therefore methods are useful for organizing code even if they don't ever implement an interface.

メソッドを「インタフェース実装手段」だけと見ると、"concrete に許すなら interface にも許さねば" と縛られます。しかしメソッドは「型の名前空間に属する関数」でもあるので、名前空間としての価値だけを目的に generic concrete method を導入するのは筋が通る、という整理です。

**3. `Reader.Read[E any]` は `io.Reader` を実装するか**

proposal Examples 節にそのままの答えがあります。

> ```
> type Reader struct{ … }
> func (*Reader) Read[E any]([]E) (int, error) { … }
> ```
> does not implement `io.Reader`, even though it might if there were some way to instantiate the method as `(*Reader).Read[byte]` (which there is not, and we are not proposing it).

実際にコンパイルすると、gotip の compiler が同じことを言います。

```
./main.go:10:20: cannot use (*Reader)(nil) (value of type *Reader) as io.Reader value in variable declaration: *Reader does not implement io.Reader (wrong type for method Read)
                have Read[E any]([]E) (int, error)
                want Read([]byte) (int, error)
```

インタフェースの `Read` は `Read([]byte) (int, error)` という「型パラメータを持たないシグネチャ」を要求しますが、こちらの `Read` は `Read[E any]([]E) (int, error)` なので、シグネチャそのものが違うと判定されます。既存 interface 充足のルールは何一つ変わっていない、と読むのが正確です。

</details>

---

## 設問 3: 実際に書いて動かそう — `Slice[T]` に `Map[F any]` を生やす

一次情報で挙動を押さえたら、冒頭の `Slice[T].Map[F any]` を完成させて、`gotip` か Playground（`?v=gotip`）で動かしてみましょう。呼び出し側で型引数を明示しなくても済むか（型推論が効くか）も確かめてください。

<details>
<summary>ヒント</summary>

- 手元で試すなら `go install golang.org/dl/gotip@latest && gotip download` してから `gotip run .`
- Playground は URL に `?v=gotip` を付けるか、Playground 画面の Go version セレクタから "Dev branch" を選ぶ
- 安定版 Go 1.26 のままでは実行できない。`gotip` または Playground の `?v=gotip` を使い、モジュールを作る場合はそのツールチェーンが受け付ける言語バージョンを設定する
- 呼び出し側は `s.Map(func(n int) string { ... })` のように書ける（型引数 `F` は関数リテラルから推論される）

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. 設問 1 の EBNF を思い出しつつ、メソッド名の直後に `[F any]` を置く。
2. 手元で `gotip run .` するか、Playground の Dev/gotip で走らせて出力を確認する。
3. 型推論の効き方は [仕様書の Type inference](https://go.dev/ref/spec#Type_inference) と同じ。関数引数から `F` が推論される。

**答え**

```go
package main

import (
	"fmt"
	"strings"
)

type Slice[T any] struct {
	items []T
}

func (s Slice[T]) Map[F any](f func(T) F) Slice[F] {
	out := Slice[F]{items: make([]F, len(s.items))}
	for i, v := range s.items {
		out.items[i] = f(v)
	}
	return out
}

func (s Slice[T]) Items() []T { return s.items }

func main() {
	nums := Slice[int]{items: []int{1, 2, 3}}
	labels := nums.Map(func(n int) string { return fmt.Sprintf("v%d", n) })
	lengths := labels.Map(func(s string) int { return len(s) })

	fmt.Println(strings.Join(labels.Items(), ","))
	fmt.Println(lengths.Items())
}
```

（Go Playground で動かす: https://go.dev/play/p/ZlBpqqdQJvd?v=gotip ）

出力はこうなります。

```
v1,v2,v3
[2 2 2]
```

- レシーバの型パラメータ `T` はメソッド呼び出しの時点ですでに束縛済み（`nums` は `Slice[int]` なので `T = int`）。
- メソッド固有の型パラメータ `F` は、渡した関数リテラル `func(int) string` の戻り値の型から推論される。明示したければ `nums.Map[string](...)` とも書ける。
- レシーバの型パラメータリスト（`Slice[T]`）とメソッドの型パラメータリスト（`[F any]`）が **2 段構え** で並ぶのが Go 1.27 の generic method の姿です。

</details>

---

<details>
<summary>こぼれ話: 標準ライブラリの <code>Rand.N</code></summary>

同じ proposal #77273 の副産物として、[math/rand/v2 の Rand 型](https://pkg.go.dev/math/rand/v2#Rand) に generic method [Rand.N](https://pkg.go.dev/math/rand/v2#Rand.N) が追加されました。これまでは package レベルの [rand.N](https://pkg.go.dev/math/rand/v2#N) しかなく、既存の `*Rand` から呼びたいときは自前でラップする必要がありました。

```go
package main

import (
	"fmt"
	"math/rand/v2"
)

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	x := r.N(int32(100))
	y := r.N(uint(1000))
	fmt.Printf("int32: %d (type %T)\n", x, x)
	fmt.Printf("uint : %d (type %T)\n", y, y)
}
```

（Go Playground で動かす: https://go.dev/play/p/HQD1EdC7pL6?v=gotip ）

出力:

```
int32: 76 (type int32)
uint : 616 (type uint)
```

Go 1.27 の Draft リリースノートの [math/rand/v2 節](https://go.dev/doc/go1.27#minor_library_changes) に「`Rand` now supports a generic method `N`, matching the behavior of the top-level `N` function.」と明記されています。「なぜ 1.27 でようやくこれが入る予定なのか」を知りたいときは、そのまま proposal #77273 のスレッドを追うのが早いです。

</details>

---

## 調査の入り口

- リリースノート: [Go 1.27 Release Notes #language](https://go.dev/doc/go1.27#language)
- 言語仕様: [Method declarations](https://go.dev/ref/spec#Method_declarations) / [Type parameter declarations](https://go.dev/ref/spec#Type_parameter_declarations)
- Proposal: [#77273 spec: generic methods for Go](https://go.dev/issue/77273)
- 先行議論: [#49085 proposal: spec: allow type parameters in methods](https://go.dev/issue/49085) / [Type Parameters Proposal の No parameterized methods 節](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#No-parameterized-methods)
- 手元で試す: `go install golang.org/dl/gotip@latest && gotip download` / Playground の `?v=gotip`
