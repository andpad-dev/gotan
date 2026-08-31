[進行ガイド・シナリオ一覧](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# なぜ標準の error にはスタックトレースが含まれていないのか？

Go で開発していると、ログに error を出力したものの「どこで発生したエラーなのか分からない。スタックトレースが欲しい」と思う場面によく遭遇します。

Java や Python などの言語では例外（Exception）に標準でスタックトレースが付随しますが、Go の標準のエラーには含まれていません。

なぜ Go はこのような設計判断をしているのでしょうか？また、エラーをWrapする仕組みはどのような議論を経て導入されたのでしょうか？
背景にある設計思想や歴史的経緯をたどってみましょう。

## 設問 1: Go のエラー処理の哲学を調べよう

そもそも Go では、JavaやPythonのようにエラーを「例外(Exception)」のような特別なものとして扱っていません。

Go の公式ブログなどからエラーハンドリングに関する記事を探し、Go におけるエラーの基本的な設計思想を調べましょう。

### ヒント

<details>
<summary>ヒント</summary>

- 公式ブログ https://go.dev/blog/ で "error" という文言で検索してみましょう。
- 言語そのものの "哲学" や "設計思想" のような概念は初期の段階から考えられています。直近のブログではなく一番古いブログから順に探していくのが良いでしょう。

</details>

### 答え

<details>
<summary>答え</summary>

#### 1. https://go.dev/blog/all を開き、検索窓で "error" と検索する。
#### 2. 最新から過去に向かって検索するのではなく、一番過去から最新の順に検索する
#### 3. 以下のような記事が見つかります
   - [Errors are values](https://go.dev/blog/errors-are-values), 12 January 2015 Rob Pike

Go では、エラーは「特別な制御フロー（例外）」ではなく、単なる「値（Value）」として扱われます。

記事には "Errors are values." という有名なフレーズが登場します。

エラーが単なる値であるため、特別な try-catch 構文は存在せず、通常の関数戻り値としてプログラムの通常の制御フローの一部として処理されます。

これにより、開発者はエラーがどこで発生し、どのように処理されるかを明示的に意識して書くことが求められる、というのが Go の基本哲学です。

</details>

---

## 設問 2: エラーの Wrapping の歴史を調べよう

Go 1.13 で、`fmt.Errorf("%w", err)` などによるError Wrappingが標準ライブラリに導入されました。

この機能が導入される際、設計者たちはどのような問題を解決しようとしていたのでしょうか？

機能追加の背景が書かれた公式情報を探してみましょう。

### ヒント

<details>
<summary>ヒント</summary>

- 機能がリリースされたGoバージョンが分かっている場合、リリースノートを見るのが最善です。
- go.dev/doc/go1.xx でそのGoバージョンのリリースノートが開きます。

</details>

### 答え

<details>
<summary>答え</summary>

#### 1. https://go.dev/doc/go1.13 を開き Error Wrapping の項目を見つけます。
#### 2. 以下のようなプロポーザルへのリンクが見つかります。
- [Error Values proposal](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md)
- [the associated issue](https://go.dev/issues/29934)

Go 2 に向けた「エラー値の検査とフォーマット」に関するドラフトデザインに基づきGo 1.13 での実装とフィードバック収集を目的にプロポーザルが記述されました。

Go 2 とは後方互換性の破壊的変更を含む次のGoバージョンという意味合いでの過去の議論で出てくるワードです。
今のところ後方互換性を破壊するGo 2が実装される予定はありません。

提案の内容は以下の通り
- エラーをWrapしてチェーン（連鎖）させる仕組み
- スタックトレースの保持
- およびそれらを展開・検査する標準的なAPIの導入

プロポーザルの issue では
- Genericsの導入後に設計し直す必要が出ないか？ Genericsを使わなくてもインターフェースの活用で十分だ
- `fmt.Errorf("... %w", err)` によるエラーWrapの仕様について、`%w` をフォーマット文字列のどこに置くべきか、出力時の見栄えはどうなるか
- エラーチェーンの中から特定のエラー値や型を見つけ出すためのAPI `errors.Is` や `errors.As` の使用感について

などなど、数々の議題で議論が繰り広げられています。

#### 3. 更にリリースノート中にある [errors package documentation](https://pkg.go.dev/errors) から https://pkg.go.dev/errors へのリンクが見つかります。
#### 4. Wrappingの議論の詳細は https://go.dev/blog/go1.13-errors を見てくださいという記載が見つかると思います。見てみましょう。
#### 5. https://go.dev/blog/go1.13-errors というブログが見つかりました。今までのプロポーザルの議論のサマリーがブログに記載されていることが分かります。

Go 1.13 以前では error を `==` 演算子で比較していました。

Go 1.13 からは Wrapされたエラーでもエラーを検査するための新しいAPIとして `errors.Is` と `errors.As` が導入されています。

エラーのカスタマイズについても触れられているため非常に参考になるドキュメントになっています。

---
ここまでの調査のサマリーとしてエラーの Wrappingの導入について、 以下のようような経緯があったことが分かります。

プログラムが複雑になると、下層で発生したエラー（例: io.EOF）を上層に返す際、単にそのまま返すと「どこで起きた EOF なのか」文脈が失われてしまいます。

サードパーティのパッケージ（pkg/errors など）がこの問題を解決するために広く使われていましたが、標準ライブラリ間でエラーの文脈を保ったまま原因を検査する統一的な手法がありませんでした。

そこで、エラーに文脈（Context）を追加しながらも、元のエラーが何であったか（errors.Is や errors.As）をプログラム的に検査できる統一された仕組みを提供するために、Wrapping の仕組みが導入されました。

</details>

---

## 設問 3: スタックトレースの自動付与が見送られた理由を調べよう

ここまでの調査の中で気づいた方もいらっしゃると思いますが、実は Go 1.13 のエラー拡張の設計段階では

「標準のエラーにスタックトレース（フレーム情報）を持たせる」という案も真剣に検討されていました。

しかし、最終的にそれは標準パッケージには採用されませんでした。

Proposal ドキュメントの議論を読み解き、なぜスタックトレースの自動付与が見送られたのか、考察してみましょう。

### ヒント

<details>
<summary>ヒント</summary>

- 最初のプロポーザルである [Proposal: Go 2 Error Inspection](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md) を読んでみよう。
- "Frame", "StackTrace", "Formatting" というキーワードで探します。
- 次に [プロポーザルを議論しているissue](https://go.dev/issues/29934) を眺めてみましょう。
- 全ての会話を眺めるのは効率が悪いため、まずは "Load more" で 全ての会話情報を取得してから検索するのが良いでしょう。
- issueの下から探すと結論を探しやすいです。
- 当時のGoチームのリーダーである Russ Cox(rsc) の発言を追っていくのが手っ取り早いでしょう。
- もしくはプロポーザルが実際にAcceptedされたタイミングである "Proposal-Accepted" で検索すると前後に決定事項が書かれていることが多いです。


</details>

### 答え

<details>
<summary>答え</summary>

最初のプロポーザル時点では以下の議題がありました。
- Wrapping (`fmt.Errorf("... %w", err)`, `errors.Is`, `errors.As`)
- Stack Frames (`errors.Frame`)
- Formatting (`errors.Printer`, `errors.Formatter`)


Russ Cox(rsc) の発言 を遡っていくと [The accept/decline decision here (also linked in the top comment) is #29934 (comment).](https://github.com/golang/go/issues/29934#issuecomment-521245013) という発言があります。

つまりGo 1.13 で Acceptedされた内容とDeclineされた内容は [#29934 (comment)](https://github.com/golang/go/issues/29934#issuecomment-489682919) にまとまっています。

この中で、Error Wrapping の議論については比較的反対意見もなくすんなりと決まっていますが
`Formatting and Location` の議論について落としどころが付かなかったため延期する旨が記載されています。

`errors.Printer`, `errors.Formatter`, `errors.Frame` がここで削除される決定をされています。

では、 Formatter の機能削除で、何故 Frame も削除されたのか？

[Proposal: Go 2 Error Inspection](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md) の Stack Frames の項を見れば分かる通り、Frameは最初からFormattingありきで検討されていたことが分かります。

他にも探せば [Stack Traceを付与することでのパフォーマンスの悪化を懸念する声](https://github.com/golang/go/issues/29934#issuecomment-486503822) や

[Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) のブログに書かれている通り "errorは値" なのだからGoプログラマーは自由にカスタムエラーが作れる旨など

様々な要因が見つかります。気になる人は更に深堀りしてみるとよいでしょう。

</details>
