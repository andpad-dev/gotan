[進行ガイド・シナリオ一覧](../../../README.md) | [01-packages の調べ方](../../README.md)

# slog.Handler インタフェースを実装しよう

ログの出力を YAML 形式で行うという意思決定がされました。
そのためには slog のログハンドラーを自作する必要があります。
一次情報のみを辿って、実装や単体テストに必要な情報を集めましょう。

## 設問 1: slog.Handler インタフェースについて調べよう

slog.Handler インタフェースにはメソッドがいくつあり、それぞれ何をするメソッドでしょうか？
どのメソッドがハンドラーの中心になるか考えてみましょう。
また、標準でこのインタフェースを実装している型も探してみましょう。

例えば `logger.Info("login", "user", "alice")` を呼ぶと、ログを受け取って出力する役割と、出力前に共通属性を付ける役割は同じでしょうか。ログがハンドラーへ届く流れを想像してから、4つのメソッドを分類してみましょう。

<details>
<summary>ヒント</summary>

- まずメソッド名と説明の最初の 1 文で雑に予想を立てる（英語が難しければ翻訳する。詳しくはまだ調べなくてよい）
- [Handler のインタフェース定義](https://pkg.go.dev/log/slog#Handler) と各メソッドのコメントを、呼び出し側の `Logger` がどの順で使うかという観点で読む。
- [Example](https://pkg.go.dev/log/slog#pkg-examples) の LevelHandler はラッパーなので簡単に読める。今回の用途はこの方法（ラッパー）でよいか考えてみよう。
- 標準の実装を列挙するときは、`TextHandler` / `JSONHandler` / `MultiHandler` が型なのか、`DiscardHandler` が値なのかも区別する。
- `Enabled` は「このログを通すか」、`Handle` は「通したログをどう出力するか」、`WithAttrs` / `WithGroup` は「後続のログに共通情報をどう持たせるか」と仮置きして、コメントで確かめる。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://pkg.go.dev/log/slog を開き、`Handler` を検索してジャンプする。
2. インタフェース定義とメソッドごとのコメントを読む。
3. [Example](https://pkg.go.dev/log/slog#pkg-examples) と Index から、Handler を実装している標準の型を探す。

**答え**

メソッドは 4 つあります。

- `Enabled(context.Context, Level) bool`: そのレベルのログを処理するかどうかを判定する。
- `Handle(context.Context, Record) error`: ログレコードを受け取って出力する。ハンドラーの中心となるメソッド。
- `WithAttrs([]Attr) Handler`: 属性を追加した新しいハンドラーを返す。
- `WithGroup(string) Handler`: グループ名を付けた新しいハンドラーを返す。

標準には `TextHandler` / `JSONHandler` / `MultiHandler` という `Handler` 実装型があります。また、すべてのログを捨てる `DiscardHandler` は `Handler` 型のパッケージ変数です。

Example の [Wrapping](https://pkg.go.dev/log/slog#example-package-Wrapping) にある LevelHandler は、既存ハンドラーを包んでレベル判定だけ差し替えるラッパーです。
出力の一部だけ変えたいならこの方法で十分ですが、今回は出力形式そのものを YAML にしたいので、`Handle` を自分で書く必要があり、ラッパーでは足りません。

</details>

---

## 設問 2: slog.Handler を実装する際の注意点を調べよう

ハンドラーの自作には注意点がいくつかあり、ドキュメントに書かれています。
探して読み、どんな注意点かグループで議論しましょう。

- slog.Handler を埋め込んで、必要なメソッドだけ実装するのはなぜダメなのでしょうか？
- slog.Value の Resolve メソッドを呼ばないと、どんなログがうまく出力されなくなるのでしょうか？

次のコードで、2つの注意点を同じ入力から観測してください。[Go Playground で実行する](https://go.dev/play/p/wXj_QdVwc19) と、埋め込みだけのハンドラーで panic が起きることと、`Resolve` の有無でパスワードの表示が変わることを確認できます。

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
)

type badHandler struct {
	slog.Handler
}

func (badHandler) Handle(context.Context, slog.Record) error {
	return nil
}

type password string

func (password) LogValue() slog.Value {
	return slog.StringValue("REDACTED")
}

func main() {
	func() {
		defer func() {
			fmt.Println("embedded handler panicked:", recover() != nil)
		}()
		slog.New(badHandler{}).Info("login")
	}()

	attr := slog.Any("password", password("secret"))
	fmt.Printf("without Resolve: %v\n", attr.Value.Any())
	fmt.Printf("with Resolve: %v\n", attr.Value.Resolve().Any())
}
```

実行結果:

```text
embedded handler panicked: true
without Resolve: secret
with Resolve: REDACTED
```

<details>
<summary>ヒント</summary>

- slog のドキュメントの Overview に「[Writing a handler](https://pkg.go.dev/log/slog#hdr-Writing_a_handler)」という節がある
- Example も参考になる: [DiscardHandler](https://pkg.go.dev/log/slog#example-package-DiscardHandler) / [Wrapping](https://pkg.go.dev/log/slog#example-package-Wrapping)
- さらに詳しいガイドが https://golang.org/s/slog-handler-guide にある
  - ガイドの IndentHandler の `mu` フィールドが、なぜ `sync.Mutex` ではなく `*sync.Mutex`（ポインタ）なのかにも注目する。
- ガイド冒頭の「埋め込み」を読んだら、Logger が `Enabled`、`WithAttrs`、`WithGroup`、`Handle` をどのように組み合わせて呼ぶかを、メソッドの委譲とコピーの観点から整理する。結論を先に決めず、4 メソッドそれぞれの契約と照合する。
- `LogValuer` の例として、パスワード型が `LogValue` で `slog.String("password", "REDACTED")` を返すケースを考える。`Resolve` を呼ばずに値をそのまま文字列化すると何が失われるか、実際の出力を想像してから確認する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/wXj_QdVwc19) を実行し、panic の有無と `Resolve` 前後の出力を観測する。
2. slog のドキュメントの Overview にある「[Writing a handler](https://pkg.go.dev/log/slog#hdr-Writing_a_handler)」節を読む。
3. そこからリンクされている詳細ガイド https://golang.org/s/slog-handler-guide （[golang/example の slog-handler-guide](https://github.com/golang/example/blob/master/slog-handler-guide/README.md)）を読む。

**答え**

主な注意点は 3 つです。

**1. 4 つのメソッドをすべて実装する。**
`slog.Handler` を埋め込んで必要なメソッドだけ実装したくなりますが、Logger と Handler は密結合なのでそれでは動きません。
ガイドの冒頭に明記されています。

> it is tempting to embed slog.Handler in your custom handler and implement only the methods that you need. Loggers and handlers are too tightly coupled for that to work. You should implement all four handler methods.

**2. `WithAttrs` / `WithGroup` はハンドラーをコピーして返す。**
ガイドの IndentHandler の `mu` フィールドが `sync.Mutex` ではなく `*sync.Mutex` なのはこのためです。
コピーされたハンドラー同士は同じ出力先（`io.Writer`）を共有するので、Mutex を値で持つとコピーごとに別のロックになり、排他が効かなくなります。
ポインタで持つことで、すべてのコピーが同じ Mutex を共有します。

**3. 属性の値は `Resolve` してから使う。**
`Handle` の中で属性を処理するときは、まず `slog.Value.Resolve` を呼びます。
`slog.LogValuer` を実装した値は、`Resolve` によって初めて `LogValue` メソッドが呼ばれ、本来ログに出したい値に変換されます。
呼ばないと、LogValuer の値（たとえばパスワードを `REDACTED` に置き換えるような型）が意図した形で出力されません。

</details>

---

## 設問 3: slog.Handler のテストの書き方を調べよう

自作ハンドラーが slog.Handler の決まりごとを守れているか、単体テストで確認する方法を調べましょう。

<details>
<summary>ヒント</summary>

- https://golang.org/s/slog-handler-guide にテストについて書かれた部分がある
- 標準パッケージに slog.Handler のテスト専用パッケージがある

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. slog-handler-guide の「[Testing](https://github.com/golang/example/blob/master/slog-handler-guide/README.md#testing)」節を読む。
2. そこで使われている [testing/slogtest](https://pkg.go.dev/testing/slogtest) パッケージのドキュメントを読む。

**答え**

標準パッケージの `testing/slogtest` を使います。
`slogtest.TestHandler` に「自作ハンドラー」と「出力結果を map のスライスに変換して返す関数」を渡すと、Handler が守るべき決まりごと（属性の解決、グループの扱い、空の属性の無視など）を一括で検証してくれます。
自分でテストケースを列挙しなくても、slog.Handler の仕様準拠を網羅的に確認できるのがポイントです。

</details>

---

## 調査の入り口

- https://pkg.go.dev/log/slog
