[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [02-features の調べ方](../../README.md)

# 自作の型にジェネリックメソッドで汎用ヘルパを生やそう

![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8)

手元に独自コンテナ型 `Slice[T]` があります。
値を別の型へ写す変換ヘルパを **メソッドとして** 生やしたいです。
どういうふうにやればいいか調べよう。

```go
type Slice[T any] struct {
	items []T
}

// これを書きたい
func (s Slice[T]) Map[F any](f func(T) F) Slice[F] { ... }
```

Go 1.26.4 でこのコードをコンパイルすると、次のように拒否されます。

```
syntax error: method must have no type parameters
```

Go 1.27 では「ジェネリックメソッド」が正式に使えるようになりました。
一次情報で次の三点を押さえてから、実際に手を動かして書きましょう。

- 仕様書のどこがどう変わったのか
- なぜ Go 1.26 までは書けなかったのか
- インタフェースメソッドではどうなのか

<details>
<summary>調査の入り口</summary>

1. [02-features の調べ方](../../README.md) を開き、言語仕様・公式ブログ・プロポーザルの逆引き手順を確かめます。
2. [Go 1.27 リリースノートの Changes to the language 節](https://go.dev/doc/go1.27#language) を開きます。Go 1.27 で言語に入った変更の一覧です。
3. [Go 言語仕様: Method declarations](https://go.dev/ref/spec#Method_declarations) と [Go 言語仕様: Type parameter declarations](https://go.dev/ref/spec#Type_parameter_declarations) を開きます。メソッド宣言と型パラメータの構文を定めた節です。
4. [proposal #77273 "spec: generic methods for Go"](https://go.dev/issue/77273) を開きます。ジェネリックメソッドを導入した提案の本文と議論です。
5. [#77549 x/tools: plan for generic methods](https://github.com/golang/go/issues/77549) を開きます。言語機能に対するツール側の追随を扱う Issue です。
6. [Go Code Owners](https://dev.golang.org/owners) を開きます。Go のパッケージごとの担当者一覧です。
7. 先行議論として、[#49085 proposal: spec: allow type parameters in methods](https://go.dev/issue/49085) を開きます。
8. 同じく先行議論の [Type Parameters Proposal の No parameterized methods 節](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#No-parameterized-methods) を開きます。ジェネリクス導入時の設計文書です。
9. 実行環境として、手元の Go のバージョンが 1.27 以降であることを確かめます。
10. 動かすときは、Go Playground で実行するか、手元の Go 1.27 以降で実行します。手元での手順は設問 3 のヒントにあります。

</details>

---

## 設問 1: 仕様書のどこが変わったのか調べよう

Go 1.27 で「method に何が書けるようになったか」を特定しましょう。
情報源はリリースノートと [The Go Programming Language Specification](https://go.dev/ref/spec) です。
仕様書では、該当セクションの **EBNF の変化** まで見ます。

<details>
<summary>ヒント</summary>

- カテゴリの逆引き手順「[最新バージョンの機能を調査したい](../../README.md)」の通り、まず [Go 1.27 リリースノート](https://go.dev/doc/go1.27) を開く
- Changes to the language セクションに「generic methods」の段落があるはず
- リリースノートの HTML には `go.dev/issue/<番号>` の形で関連 issue が埋め込まれている（開発者ツールでソースを見ると拾える）
- 仕様書側は `Method declarations` セクションを開いて、EBNF が旧版とどう変わっているか比べる
- 仕様書の冒頭が `Language version go1.27` であることも確認し、リリースノート・仕様・実行環境の版を揃える

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート の Changes to the language 節](https://go.dev/doc/go1.27#language) を開き、`generic methods` の段落を読む。
2. その段落に埋め込まれた `go.dev/issue/77273` を開いて [proposal #77273 "spec: generic methods for Go"](https://go.dev/issue/77273) の Proposal 節に飛ぶ。旧新の EBNF が並記されている。
3. [仕様書の Method declarations](https://go.dev/ref/spec#Method_declarations) セクションで、実際に採用された EBNF を確認する。
4. 手元の `go version` と Playground の表示が Go 1.27 以降であることを確認する。

**答え**

リリースノートには次のように書かれています。

> Go 1.27 now supports generic methods: a method declaration may declare its own type parameters. This widely anticipated change allows adding generic functions within the namespace of a particular data type where before one had to declare such functions with a scope of the entire package. Note that methods of interfaces may not declare type parameters nor can interface methods be implemented by generic methods.

仕様書 [Method declarations](https://go.dev/ref/spec#Method_declarations) の EBNF がこう変わりました（proposal #77273 の Proposal 節で対比されているとおり）。

- **旧 (Go 1.26 まで)**
  ```
  MethodDecl = "func" Receiver MethodName Signature [ FunctionBody ] .
  ```
- **新 (Go 1.27 以降)**
  ```
  MethodDecl = "func" Receiver MethodName [ TypeParameters ] Signature [ FunctionBody ] .
  ```

`MethodName` と `Signature` の間に `[ TypeParameters ]` が挿入されただけです。関数宣言 (`FunctionDecl`) の型パラメータ位置と揃った形になっています。

Go 1.26.4 のコンパイラはジェネリックメソッドの構文を受け付けないため、次のように断られます。

```
syntax error: method must have no type parameters
```

</details>

---

## 設問 2: インタフェースメソッドとの線引きを調べよう

リリースノートの generic methods の段落は、二つの制約を明記しています。

- インタフェースメソッドは型パラメータを宣言できない
- **ジェネリックメソッドでインタフェースを実装することもできない**

なぜこの線引きなのでしょうか。
次のコードを Playground で試して、コンパイラのエラーメッセージも確認しましょう。

```go
package main

import "io"

type Reader struct{}

func (*Reader) Read[E any](p []E) (int, error) { return 0, nil }

func main() {
	var _ io.Reader = (*Reader)(nil)
}
```

（[Go Playground で動かす](https://go.dev/play/p/Lqn-PC8jg9E)）

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
3. [仕様書の Interface types](https://go.dev/ref/spec#Interface_types) で、interface method は型パラメータを宣言できないという現行規則を確認する。
4. Go 1.27 以降の手元環境か Playground で実際にコンパイルして、compiler のエラーメッセージを見る。
5. 深追いしたい人向け: [Type Parameters Proposal の No parameterized methods 節](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#No-parameterized-methods) で当時の議論を読む。

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

Go 1.27.0 で実際にコンパイルすると、compiler が同じことを言います。

```
./main.go:10:20: cannot use (*Reader)(nil) (value of type *Reader) as io.Reader value in variable declaration: *Reader does not implement io.Reader (wrong type for method Read)
                have Read[E any]([]E) (int, error)
                want Read([]byte) (int, error)
```

インタフェースの `Read` は `Read([]byte) (int, error)` という「型パラメータを持たないシグネチャ」を要求しますが、こちらの `Read` は `Read[E any]([]E) (int, error)` なので、シグネチャそのものが違うと判定されます。既存 interface 充足のルールは何一つ変わっていない、と読むのが正確です。

</details>

---

## 設問 3: `Slice[T]` に `Map[F any]` を生やして動かそう

冒頭の `Slice[T].Map[F any]` を完成させましょう。
Go 1.27 以降の手元環境か Playground で動かします。
呼び出し側で型引数を明示しなくても済むか（型推論が効くか）も確かめてください。

<details>
<summary>ヒント</summary>

- 手元で試す場合は、空の作業ディレクトリに完成したコードを `main.go` として保存する。`go mod init example.com/generic-methods` の後、`go.mod` の `go` 行を `1.27` にし、`go run .` を実行する
- Playground の実行結果に表示される Go バージョンも確認する
- `go.mod` は `go 1.27` にする。Go 1.27 toolchain でも `go 1.26` のままだと、`generic method requires go1.27 or later` という言語バージョンのエラーになる
- 呼び出し側は `s.Map(func(n int) string { ... })` のように書ける（型引数 `F` は関数リテラルから推論される）

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. 設問 1 の EBNF を思い出しつつ、メソッド名の直後に `[F any]` を置く。
2. Playground で走らせるか、空の作業ディレクトリに `main.go` と `go 1.27` の `go.mod` を用意して `go run .` し、出力を確認する。
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

（[Go Playground で動かす](https://go.dev/play/p/ZlBpqqdQJvd)）

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

（[Go Playground で動かす](https://go.dev/play/p/HQD1EdC7pL6)）

出力:

```
int32: 76 (type int32)
uint : 616 (type uint)
```

Go 1.27 リリースノートの [math/rand/v2 節](https://go.dev/doc/go1.27#minor_library_changes) に「`Rand` now supports a generic method `N`, matching the behavior of the top-level `N` function.」と明記されています。「なぜ 1.27 でようやくこれが入ったのか」を知りたいときは、そのまま proposal #77273 のスレッドを追うのが早いです。

</details>

---

<details>
<summary>こぼれ話: 最近の対応を「人」から追う</summary>

言語機能が入った後は、コンパイラだけでなく gopls、解析ライブラリ、デバッガなども新しい構文へ追随します。proposal #77273 の Libraries and tools 節は、過去の経験からツールが追随するまで 1、2 リリースかかる可能性に触れています。では、Go 1.27 の Generic Methods では実際に誰が何を追っていたのでしょうか。

まず [Go 1.27 リリースノート](https://go.dev/doc/go1.27#language) と [proposal #77273](https://go.dev/issue/77273) で対象機能と版を固定します。次に GitHub の `golang/go` Issues で `"generic methods"` を検索し、`Tools` ラベルと `Go1.27` milestone で絞ると、Alan Donovan が起票した [#77549 x/tools: plan for generic methods](https://github.com/golang/go/issues/77549) に到達できます。この Issue は、`x/tools` 全体を監査し、gopls、vulncheck、外部ツールなどの追随作業を集めるための追跡 Issue です。

ここで [Alan Donovan のプロフィール](https://github.com/adonovan) を開くと、同じ人物が最近扱っている Issue やリポジトリを別方向から探せます。ただし、プロフィールの活動量は実装の完了や現在の担当を証明しません。機能名と人物名で見つけた候補から、対象版の Issue、関連 CL、レビュー、バージョン付きソースへ戻って結論を確認します。[Go Code Owners](https://dev.golang.org/owners) もパッケージごとの担当候補と Gerrit 履歴を探す入口ですが、同じく直近の変更履歴で再確認してください。

</details>
