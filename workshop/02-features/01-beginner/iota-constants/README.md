[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [02-features の調べ方](../../README.md)

# iota を使った定数宣言を読もう

![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8)

先輩のコードで、次のような定数宣言を見かけました。

```go
type Color int

const (
    Red Color = iota
    Green
    Blue
)
```

（Go Playground で動かす: https://go.dev/play/p/mBjgkLshasU ）

`fmt.Println(Red, Green, Blue)` すると `0 1 2` と出ます。
`iota` とはどんなもので、なぜ 2 行目以降を省略しても値が変わっていくのでしょうか。調べてみましょう。

設問1では、まず `iota` が何者か（関数・キーワード・定数として使える識別子のどれか）に絞ります。`Green` と `Blue` の式を省略できる理由は設問2で扱います。

## 設問 1: `iota` とは何？

最初の `Red` にだけ書かれている `iota` の正体を、言語仕様書で調べてみましょう。

`const` ブロック内の各行が、`iota` の値を決める「位置」になると考えて、1行目と3行目で値がどう変わるかを予想してみてください。

<details>
<summary>ヒント</summary>

- Go 言語仕様書は 1 枚 HTML なので、`f` キー（ブラウザのページ内検索）で `iota` を探すのが早いです。
- `Constant declarations` セクションの中に、`Iota` という見出しがあります。
- 仕様書の [Iota](https://go.dev/ref/spec#Iota) と [Predeclared identifiers](https://go.dev/ref/spec#Predeclared_identifiers) を順に読み、`iota` がキーワード一覧にあるかも確認します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://go.dev/ref/spec を開き、ページ内検索で `Iota` を探す。
2. `Constant declarations` の中の `Iota` セクションを読む。

**答え**

`iota` は **定数宣言 (`ConstDecl`) の中でだけ使える、あらかじめ宣言された定数（predeclared identifier）** です。関数でも、予約語であるキーワードでもありません。
仕様書 [Iota](https://go.dev/ref/spec#Iota) には次のように書かれています。

> Within a constant declaration, the predeclared identifier `iota` represents successive untyped integer constants. Its value is the index of the respective ConstSpec in that constant declaration, starting at zero.

要点は 2 つです。

- 値は **untyped integer**（型なし整数定数）。
- 値は、その `const ( ... )` ブロック内の **ConstSpec（ここでは各定数宣言の行）のインデックス**（0 始まり）。1 行目が 0、2 行目が 1、3 行目が 2 ……となる。

つまり `Red Color = iota` の時点で `Red = 0` が確定し、`Color` 型として定義されます。

</details>

---

## 設問 2: なぜ `Green` と `Blue` は式を省略できる？

`Green` と `Blue` には `= iota` すら書かれていないのに、なぜ `1`・`2` という値が入るのでしょうか。
`iota` の性質だけでは説明しきれません。もう 1 つ、定数宣言のルールがあります。

`Red = iota / Green / Blue` の省略を、`Red = iota / Green = iota / Blue = iota` と書き直した場合と比べてみましょう。省略が値を直接増やしているのか、式を繰り返しているだけなのかを切り分けます。

<details>
<summary>ヒント</summary>

- 仕様書の `Constant declarations` セクション本文に、「式リストが省略されたときの挙動」の説明があります。
- 「同じ式リストを繰り返す」ようなイメージのルールです。
- `A = 10 / B / C` と `Red = iota / Green / Blue` を並べ、前の式を繰り返した結果がそれぞれどうなるか予想します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://go.dev/ref/spec#Constant_declarations を開く。
2. 「Within a parenthesized `const` declaration list ...」で始まる段落を読む。

**答え**

省略は「値を1つ増やす」記法ではありません。前の行の式リストを繰り返し、その繰り返された式の中で `iota` が現在の位置の値になります。

仕様書 [Constant declarations](https://go.dev/ref/spec#Constant_declarations) には次のルールがあります。

> Within a parenthesized `const` declaration list the expression list may be omitted from any but the first ConstSpec. Such an empty list is equivalent to the textual substitution of the first preceding non-empty expression list and its type if any.

これを **implicit repetition（暗黙の繰り返し）** と呼びます。

- 括弧付きの `const ( ... )` の中では、2 行目以降の `ConstSpec` は式リスト（と型）を丸ごと省略できる。
- 省略されたときは、**その直前にある「省略されていない式リストと型」がテキストとしてそのままコピーされる**。

したがって、

```go
const (
    Red Color = iota
    Green
    Blue
)
```

は、コンパイラから見ると

```go
const (
    Red   Color = iota
    Green Color = iota
    Blue  Color = iota
)
```

と等価です。
そして `iota` は「その ConstSpec のインデックス」なので、`Red` では 0、`Green` では 1、`Blue` では 2 に評価されます。
2 つのルール（implicit repetition と iota のインデックス性）が組み合わさって、Go 定番の enum イディオムができています。

</details>

---

<details>
<summary>こぼれ話: <code>1 &lt;&lt; iota</code> でビットフラグ</summary>

`iota` は式の中で使えるので、`1 << iota` と書くとビットフラグが作れます。

```go
type Perm uint

const (
    Read Perm = 1 << iota
    Write
    Execute
)
```

（Go Playground で動かす: https://go.dev/play/p/QrVksK9QllL ）

`fmt.Println(Read, Write, Execute, Read|Write)` の出力は `1 2 4 3` になります。

implicit repetition で 2 行目以降にコピーされるのは **式リストそのもの**（`1 << iota`）であって、値ではありません。
`iota` は各行で再評価されるので、それぞれ `1<<0=1`、`1<<1=2`、`1<<2=4` になります。

[Effective Go の Constants](https://go.dev/doc/effective_go#constants) でも、この `1 << iota` を使った `ByteSize`（KB, MB, GB, ...）の例が紹介されています。

</details>

---

## 調査の入り口

- https://go.dev/ref/spec
- https://go.dev/doc/effective_go
