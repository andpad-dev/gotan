[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# なぜ標準の error にはスタックトレースが含まれていないのか？

![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8)

Go で開発していると、ログに error を出力したものの「どこで発生したエラーなのか分からない。スタックトレースが欲しい」と思う場面によく遭遇します。

Java や Python では、例外（Exception）に標準でスタックトレースが付きます。Go の標準のエラーには含まれていません。

なぜ Go はこのような設計判断をしているのでしょうか？また、エラーをWrapする仕組みはどのような議論を経て導入されたのでしょうか？
背景にある設計思想や歴史的経緯をたどってみましょう。

まず、次のコードを実行して、エラーへの文脈追加・原因の検査・詳細表示を観測してください。[Go Playground で動かす](https://go.dev/play/p/m3PMRBAE_rq) と、`%+v` でもスタックトレースが表示されないことを確認できます。

```go
package main

import (
	"errors"
	"fmt"
)

var errNotFound = errors.New("not found")

func loadConfig() error {
	return errNotFound
}

func handleRequest() error {
	return fmt.Errorf("load config: %w", loadConfig())
}

func main() {
	err := handleRequest()
	fmt.Println("error:", err)
	fmt.Println("is not found:", errors.Is(err, errNotFound))
	fmt.Printf("detailed format: %+v\n", err)
}
```

実行結果:

```text
error: load config: not found
is not found: true
detailed format: load config: not found
```

<details>
<summary>調査の入り口</summary>

1. [04-deep-dive の調べ方](../../README.md) を開き、仕様・実装・設計背景をたどる順番を確かめます。
2. [Go Documentation](https://go.dev/doc/) を開き、バージョン別のリリースノートへ進む入口にします。
3. [Go 1.13 Release Notes](https://go.dev/doc/go1.13) の Error wrapping の項目を読み、そこから設計資料と issue へたどります。
4. [Errors are values](https://go.dev/blog/errors-are-values) を読み、エラーを値として扱う例を確認します。
5. [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) で、提案された API の使い方と背景を確認します。
6. [Error Values — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md) の Problem を読み、エラー生成に求められたコストの性質を確認します。
7. [Error Values proposal](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md) の Stack Frames と Formatting の項目を読みます。
8. [Proposal issue #29934](https://go.dev/issues/29934) の終盤にある決定事項を確認し、採用された仕様と見送られた仕様を切り分けます。

</details>

---

## 設問 1: Go のエラー処理の哲学を調べよう

上のコードでは、`err` に文脈を付けても `errors.Is` で元のエラーを検査できます。
`%+v` の出力にスタックトレースは現れません。
この観測を手がかりに、Go のエラーは例外とどのように違う値として設計されているのか、公式情報から調べましょう。

<details>
<summary>ヒント</summary>

- [Go Blog の一覧](https://go.dev/blog/all) を開き、ブラウザ内検索で `Errors are values` を探しましょう。
- 記事を読んだら、設計資料 [Error Values — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md) の `Problem` で `fixed cost` を探します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/m3PMRBAE_rq) を実行し、エラーの表示と `errors.Is` の結果を観測する。
2. [Go Blog の一覧](https://go.dev/blog/all) を開き、ブラウザ内検索で `Errors are values` を探す。
3. [Errors are values](https://go.dev/blog/errors-are-values) を読み、エラーを値として扱う例を確認する。
4. [Error Values — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md) の Problem を読み、エラー生成に求められたコストの性質を確認する。

**答え**

Go では、エラーは「特別な制御フロー（例外）」ではなく、単なる「値（Value）」として扱われます。

また設計資料は、エラーは例外的にしか起きないものではなく、プログラム中で繰り返し生成・処理・破棄されるものだと整理しています。そのため、エラー生成のコストはスタックの深さなどに左右されない一定のコストである必要があります。標準エラーへ常にスタックを記録する設計は、この制約と緊張関係にあります。

記事には "Errors are values." という有名なフレーズが登場します。

エラーが単なる値であるため、特別な try-catch 構文は存在せず、通常の関数戻り値としてプログラムの通常の制御フローの一部として処理されます。

これにより、開発者はエラーがどこで発生し、どのように処理されるかを明示的に意識して書くことが求められる、というのが Go の基本哲学です。

</details>

---

## 設問 2: エラーの Wrapping の歴史を調べよう

Go 1.13 で Error Wrapping が標準ライブラリに導入されました。書き方は `fmt.Errorf("%w", err)` などです。

この機能が導入される際、設計者たちはどのような問題を解決しようとしていたのでしょうか？

上のコードでは、`fmt.Errorf` がエラーに文脈を追加し、`errors.Is` が元のエラーを見つけています。
この2つの観測を出発点に、導入の背景が書かれた公式情報を調べましょう。

<details>
<summary>ヒント</summary>

- 機能がリリースされたGoバージョンが分かっている場合、リリースノートを見るのが最善です。
- go.dev/doc/go1.xx でそのGoバージョンのリリースノートが開きます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.13 Release Notes](https://go.dev/doc/go1.13) の Error wrapping の項目を読む。
2. リリースノートから次の設計資料と issue をたどる。
   - [Error Values proposal](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md)
   - [the associated issue](https://go.dev/issues/29934)
3. [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) で、提案された API の使い方と背景を確認する。

**答え**

当初の提案には、次の内容が含まれていました。

- エラーを Wrap してチェーン（連鎖）させる仕組み
- エラーチェーンを検査する `errors.Is` と `errors.As`
- エラーの詳細な表示や位置情報を扱う仕組み（`errors.Frame` など）

プロポーザルの issue では
- Genericsの導入後に設計し直す必要が出ないか？ Genericsを使わなくてもインターフェースの活用で十分だ
- `fmt.Errorf("... %w", err)` によるエラーWrapの仕様について、`%w` をフォーマット文字列のどこに置くべきか、出力時の見栄えはどうなるか
- エラーチェーンの中から特定のエラー値や型を見つけ出すためのAPI `errors.Is` や `errors.As` の使用感について

などなど、数々の議題で議論が繰り広げられています。

リリースノートとブログによると、Go 1.13 以前はエラーを `==` 演算子で比較することがありました。
Go 1.13 では、Wrap されたエラーを検査する新しい API として `errors.Is` と `errors.As` が導入されました。

ここまでの調査のサマリーとしてエラーの Wrapping の導入について、以下のような経緯があったことが分かります。

プログラムが複雑になると、下層で発生したエラー（例: io.EOF）を上層に返す際、単にそのまま返すと「どこで起きた EOF なのか」文脈が失われてしまいます。

サードパーティのパッケージ（pkg/errors など）がこの問題を解決するために広く使われていましたが、標準ライブラリ間でエラーの文脈を保ったまま原因を検査する統一的な手法がありませんでした。

そこで、エラーに文脈（Context）を追加しながらも、元のエラーが何であったか（errors.Is や errors.As）をプログラム的に検査できる統一された仕組みを提供するために、Wrapping の仕組みが導入されました。

</details>

---

## 設問 3: スタックトレースの自動付与が見送られた理由を調べよう

Go 1.13 のエラー拡張の設計段階では、もう一つの案がありました。
標準のエラーにスタックトレース（フレーム情報）を持たせる案です。

しかし、この案は標準パッケージには採用されませんでした。

どの仕様が残り、どの仕様が見送られたのかを、実際の出力と一次資料から切り分けましょう。

<details>
<summary>ヒント</summary>

- [Proposal: Go 2 Error Inspection](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md) で `Frame`、`StackTrace`、`Formatting` を検索します。
- 次に [プロポーザルを議論している issue](https://go.dev/issues/29934) の終盤にある決定事項を確認します。
- [Proposal-Accepted](https://github.com/golang/go/issues/29934#issuecomment-489682919) と最終決定のコメントを比較すると、採用された仕様と見送られた仕様を切り分けやすくなります。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/m3PMRBAE_rq) を実行し、`%+v` の出力にスタックトレースが含まれないことをもう一度確認する。
2. [Error Values — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md) の Problem を読み、エラー生成をスタックの深さにかかわらず一定コストにする設計上の制約を確認する。
3. [Error Values proposal](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md) の Stack Frames と Formatting の項目を読む。
4. issue の [Proposal-Accepted のコメント](https://github.com/golang/go/issues/29934#issuecomment-489682919) と [最終決定のコメント](https://github.com/golang/go/issues/29934#issuecomment-521245013) を比較する。
5. [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) で、最終的に導入された Wrapping と検査 API を確認する。

**答え**

最初の提案には、エラーを連鎖させて検査する Wrapping (`fmt.Errorf("... %w", err)`, `errors.Is`, `errors.As`) と、スタックフレームや詳細なフォーマットを扱う案 (`errors.Frame`, `errors.Printer`, `errors.Formatter`) が含まれていました。

その前提となる設計資料では、Go のエラーは頻繁に生成・処理される通常の値なので、生成コストをスタックの深さに依存させないことが要件とされています。これは、すべてのエラー生成時にスタックを採取しない理由を、言語のエラー観から説明する根拠です。

Go 1.13 の採否を決める議論で残ったのは、`Unwrap`、`errors.Is`、`errors.As`、`%w` によるエラーの Wrapping と検査です。一方、`Formatting and Location` に関する議論は合意に至らず、`errors.Printer`、`errors.Formatter`、`errors.Frame` は見送られました。

提案書の Stack Frames は、フレーム情報を保存するだけでなく、Formatting と組み合わせて表示する設計として検討されていました。そのため、Formatting の案が見送られたことと Frame の案が見送られたことは切り離せません。

issue の議論には、スタックトレースを付与した場合のパフォーマンスや API の複雑さを懸念する意見もあります。ただし、これらは議論中に出た懸念であり、最終決定の主な根拠として断定せず、一次資料の採否結果と分けて読み取る必要があります。

したがって、Go の標準エラーは自動的にスタックトレースを保持・表示しません。必要な場合は、独自のエラー型やライブラリが詳細情報を保持・表示する設計を選べます。

</details>
