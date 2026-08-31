# slog.Handler インタフェースを実装しよう

ログの出力を YAML 形式で行うという意思決定がされました。
そのためには slog のログハンドラーを自作する必要があります。
一次情報のみを辿って、実装や単体テストに必要な情報を集めましょう。

---

## 設問 1: slog.Handler インタフェースについて調べよう

slog.Handler インタフェースにはメソッドがいくつあり、それぞれ何をするメソッドでしょうか？
どのメソッドがハンドラーの中心になるか考えてみましょう。
また、標準でこのインタフェースを実装している型も探してみましょう。

例えば `logger.Info("login", "user", "alice")` を呼ぶと、ログを受け取って出力する役割と、出力前に共通属性を付ける役割は同じでしょうか。ログがハンドラーへ届く流れを想像してから、4つのメソッドを分類してみましょう。

<details>
<summary>ヒント</summary>

- まずメソッド名と説明の最初の 1 文で雑に予想を立てる（英語が難しければ翻訳する。詳しくはまだ調べなくてよい）
- [Handler のインタフェース定義](https://pkg.go.dev/log/slog#Handler) と各メソッドのコメントを、呼び出し側の `Logger` がどの順で使うかという観点で読む。
- [LevelHandler の Example](https://pkg.go.dev/log/slog#example-Handler-LevelHandler) はラッパーなので簡単に読める。今回の用途はこの方法（ラッパー）でよいか考えてみよう。
- 標準の実装を列挙するときは、`TextHandler` / `JSONHandler` / `MultiHandler` が型なのか、`DiscardHandler` が値なのかも区別する。
- `Enabled` は「このログを通すか」、`Handle` は「通したログをどう出力するか」、`WithAttrs` / `WithGroup` は「後続のログに共通情報をどう持たせるか」と仮置きして、コメントで確かめる。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から標準ライブラリの `log/slog` を開く。
2. [Handler](https://pkg.go.dev/log/slog#Handler) のインタフェース定義とメソッドごとのコメントを読む。
3. [LevelHandler の Example](https://pkg.go.dev/log/slog#example-Handler-LevelHandler) と Index から、Handler を実装している標準の型を探す。
4. [Go 1.27.0 の LevelHandler 実装](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/example_level_handler_test.go;l=13-78) を開き、`Enabled` だけを変え、残る 3 メソッドを内側のハンドラーへ委譲していることを確認する。

**答え**

メソッドは 4 つあります。

- `Enabled(context.Context, Level) bool`: そのレベルのログを処理するかどうかを判定する。
- `Handle(context.Context, Record) error`: ログレコードを受け取って出力する。ハンドラーの中心となるメソッド。
- `WithAttrs([]Attr) Handler`: 属性を追加した新しいハンドラーを返す。
- `WithGroup(string) Handler`: グループ名を付けた新しいハンドラーを返す。

標準には `TextHandler` / `JSONHandler` / `MultiHandler` という `Handler` 実装型があります。また、すべてのログを捨てる `DiscardHandler` は `Handler` 型のパッケージ変数です。

Example の [LevelHandler](https://pkg.go.dev/log/slog#example-Handler-LevelHandler) は、既存ハンドラーを包んでレベル判定だけ差し替えるラッパーです。
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
- Example も参考になる: [DiscardHandler](https://pkg.go.dev/log/slog#example-package-DiscardHandler) / [LevelHandler](https://pkg.go.dev/log/slog#example-Handler-LevelHandler)
- さらに詳しい公式ガイドが [slog handler guide](https://go.dev/s/slog-handler-guide) にある
  - ガイドの IndentHandler の `mu` フィールドが、なぜ `sync.Mutex` ではなく `*sync.Mutex`（ポインタ）なのかにも注目する。
- ガイド冒頭の「埋め込み」を読んだら、Logger が `Enabled`、`WithAttrs`、`WithGroup`、`Handle` をどのように組み合わせて呼ぶかを、メソッドの委譲とコピーの観点から整理する。結論を先に決めず、4 メソッドそれぞれの契約と照合する。
- `LogValuer` の例として、パスワード型が `LogValue` で `slog.StringValue("REDACTED")` を返すケースを考える。`Resolve` を呼ばずに値をそのまま文字列化すると何が失われるか、実際の出力を想像してから確認する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/wXj_QdVwc19) を実行し、panic の有無と `Resolve` 前後の出力を観測する。
2. slog のドキュメントの Overview にある「[Writing a handler](https://pkg.go.dev/log/slog#hdr-Writing_a_handler)」節を読む。
3. [Go 1.27.0 の `Handler.Handle` の契約](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/handler.go;l=41-65) で、属性値の解決など、出力するハンドラーが守る規則を確認する。
4. そこからリンクされている [slog handler guide](https://go.dev/s/slog-handler-guide)（[golang/example の本文](https://github.com/golang/example/blob/master/slog-handler-guide/README.md)）を開き、埋め込み、`WithAttrs` / `WithGroup`、値の解決に関する節を順に読む。

**答え**

主な注意点は 3 つです。

**1. 4 つのメソッドをすべて実装する。**
`slog.Handler` を埋め込んで必要なメソッドだけ実装したくなりますが、Logger と Handler は密結合なのでそれでは動きません。
ガイドは、インタフェースを埋め込んで一部だけ実装する方法ではなく、4 メソッドをすべて実装するよう説明しています。

**2. `WithAttrs` / `WithGroup` は元を変更せず、新しいハンドラーを返す。**
ガイドの `IndentHandler` はこの契約を満たすために構造体をコピーします。その `mu` フィールドが `sync.Mutex` ではなく `*sync.Mutex` なのは、コピー後も同じロックを共有するためです。
コピーされたハンドラー同士は同じ出力先（`io.Writer`）を共有するので、Mutex を値で持つとコピーごとに別のロックになり、排他が効かなくなります。

**3. 属性の値は `Resolve` してから使う。**
`Handle` の中で属性を処理するときは、まず `slog.Value.Resolve` を呼びます。
`slog.LogValuer` を実装した値は、`Resolve` によって初めて `LogValue` メソッドが呼ばれ、本来ログに出したい値に変換されます。
呼ばないと、LogValuer の値（たとえばパスワードを `REDACTED` に置き換えるような型）が意図した形で出力されません。

</details>

---

## 設問 3: slog.Handler のテスト API を使い分けよう

自作ハンドラーが `slog.Handler` の決まりごとを守れているか、標準パッケージで確認する方法を調べましょう。`testing/slogtest` の `TestHandler` と `Run` は、どちらも何を検査し、失敗の報告方法とハンドラーの作り方がどう違うでしょうか。

<details>
<summary>ヒント</summary>

- [slog handler guide](https://go.dev/s/slog-handler-guide) にテストについて書かれた部分がある。
- 標準パッケージに `slog.Handler` のテスト専用パッケージがある。関数一覧で `TestHandler` と `Run` の引数を比べる。
- 1 回のテストで結果をまとめて検査する場合と、ケースごとに新しいハンドラーを作ってサブテストにする場合を分けて考える。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から標準ライブラリの `testing/slogtest` を開く。
2. [Go 1.22 Release Notes](https://go.dev/doc/go1.22#testing/slogtest) で、`Run` がサブテストを使う API として追加された経緯を確認する。
3. [testing/slogtest](https://pkg.go.dev/testing/slogtest) の関数一覧で、[`TestHandler`](https://pkg.go.dev/testing/slogtest#TestHandler) と [`Run`](https://pkg.go.dev/testing/slogtest#Run) のシグネチャと説明を比べる。
4. [Go 1.27.0 の実装](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/testing/slogtest/slogtest.go;l=247-316) を開き、両方が同じテストケースを使い、結果の集約方法が異なることを確認する。
5. 実装例として slog-handler-guide の「[Testing](https://github.com/golang/example/blob/master/slog-handler-guide/README.md#testing)」節を読む。

**答え**

標準パッケージの `testing/slogtest` を使います。`TestHandler` と `Run` は同じテストケースを使い、属性の解決、グループの扱い、空の属性の無視など、Handler が守るべき決まりごとを検査します。

`TestHandler` には、自作ハンドラーと、全ログの出力結果を `[]map[string]any` に変換して返す関数を渡します。すべてのケースを実行した後、違反を `errors.Join` でまとめた `error` として返します。

`Run` には `*testing.T` と、ケースごとに新しいハンドラーを作る関数、1 件分の結果を `map[string]any` で返す関数を渡します。各ケースをサブテストとして実行し、違反をそのサブテストの `t.Error` で報告します。通常の単体テストで失敗したケースを個別に見たい場合は `Run`、1 つのハンドラーへ全ケースを流した結果をまとめて扱いたい場合は `TestHandler` と使い分けられます。

どちらも仕様準拠の共通ケースを提供するため、自分で同じ検査項目を一から列挙する必要はありません。ただし、自作ハンドラー固有の YAML 形式やエラー処理は、別のテストで補います。

</details>

---

## 調査の入り口

- [Go Documentation](https://go.dev/doc/)
- [01-packages の調べ方](../../README.md)
- https://pkg.go.dev/log/slog
- [slog handler guide](https://go.dev/s/slog-handler-guide)
- [package testing/slogtest](https://pkg.go.dev/testing/slogtest)
