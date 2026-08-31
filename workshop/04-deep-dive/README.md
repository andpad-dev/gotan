[シナリオ一覧](../SCENARIOS.md) | [ワークショップ進行ガイド](../README.md) | [チームでの進め方](../TEAM_GUIDE.md)

# 04-deep-dive の調べ方（逆引き手順）

このカテゴリでは、Go の言語仕様や標準ライブラリの実装、設計の背景を一次情報の順番に沿って調べます。各シナリオでは、まず共通の入口から観測した現象を言葉にし、根拠を一つずつたどってください。

## 1. まず言語仕様を読む

構文、型、比較可能性、メモリサイズ、エラー処理などのルールを調べるときは、最初に [The Go Programming Language Specification](https://go.dev/ref/spec) を開きます。

1. 目次または `f` キーで、シナリオ本文から得た概念に近い節を探す。
2. 該当する定義、制約、例外条件を読み、主張を仕様の文言に対応付ける。
3. 仕様だけで決まらない実装上の挙動は、次の標準ライブラリのドキュメントとソースへ進む。

## 2. 標準ライブラリの契約を読む

標準パッケージの API や設計上の注意点は、[Go Documentation](https://go.dev/doc) から公式ドキュメントを確認し、必要に応じて [pkg.go.dev/std](https://pkg.go.dev/std) のパッケージ一覧から対象ページを開きます。

- Overview と対象 API のコメントを読む。
- Example と関連する型・メソッドを確認する。
- ドキュメントの宣言から Go 本体のソースへのリンクをたどり、説明と実装を照合する。

## 3. 設計の背景と歴史をたどる

「なぜこの設計なのか」「なぜ別案が採用されなかったのか」を調べる場合は、次の順番で進みます。

1. 対象バージョンのリリースノート（`go.dev/doc/go1.<version>`）を確認する。
2. リリースノートからリンクされた公式ブログ、プロポーザル、Issue を読む。
3. Issue の本文、議論、関連リンクまで確認し、採用された事実・見送られた案・自分の考察を分けて記録する。
4. 最後に実装ソースと実測結果を照合する。

公式ブログを横断して探すときは [Go Blog の一覧](https://go.dev/blog/all) を使います。検索で見つけた外部記事だけを根拠にせず、必ず Go プロジェクトの一次資料へ戻って確認してください。

## 4. 実装ソースをバージョン付きで読む

標準ライブラリやコンパイラの実装を確認するときは、[Go 本体のソース](https://cs.opensource.google/go/go) を開きます。

1. `pkg.go.dev` の宣言リンクまたは公式ドキュメントから対象ファイルを特定する。
2. `refs/tags/go1.27.0` のようなバージョンタグを付け、README に書かれた Go 版と揃える。
3. 対象行だけで結論を出さず、呼び出し元・フィールドコメント・関連する型定義まで読む。

## 5. 手元で挙動を再現する

- 実行可能な例は、README に記載された作業ディレクトリから同じコマンドを実行する。
- `go version`、終了コード、標準出力、標準エラーを記録し、ドキュメントの説明と突き合わせる。
- コンパイルエラーの例も同じツールチェーンで実行し、実際の診断を記録する。
- 小さなコードをブラウザで試す場合は [Go Playground](https://go.dev/play/) を使い、Share したソースと README のコードが一致していること、表示された Go のバージョンを確認する。

## 調査の進め方

観測した現象 → 公式ドキュメント → 言語仕様または API 契約 → 設計議論 → バージョン付き実装ソース → 実測、の順に進めます。一次資料で結論を固定できない場合は、断定せず「不明」または「ここからは推測」と明記してください。

## 設計史・実装を research.swtch.com から逆引きする

Russ Cox は、[2008 年に Go の開発チームへ参加し、2 つのコンパイラと標準ライブラリの構築に携わった](https://go.dev/blog/toward-go2)ソフトウェアエンジニアで、のちに [Go プロジェクトと Google の Go チームで技術面を率いる technical lead を務めました](https://go.dev/blog/open-source)。[research!rsc の目次](https://research.swtch.com/) は彼の個人サイトで、Go の設計判断や実装の経緯を、中心的な開発者の視点からたどるための資料です。

research!rsc は Go プロジェクトの公式ドキュメントではありません。表の go.dev の入口で現在の仕様・対象版・実装を先に固定し、目次を記事の題名で Ctrl+F / Cmd+F して候補を開き、残りの題名・語をシリーズ内または記事内で検索します。最後にもう一度バージョン付きソースと実測へ戻ってください。

| 調べたい課題 | go.dev から先に確認すること | research!rsc の目次で探す題名 → 次に探す題名・語 |
| --- | --- | --- |
| 浮動小数点の出力を変えずに、文字列化の内部実装を置き換えられる理由を追いたい | [`strconv.FormatFloat`](https://go.dev/pkg/strconv/#FormatFloat) で現在の契約を確認し、[Go 1.26.0](https://cs.opensource.google/go/go/+/refs/tags/go1.26.0:src/internal/strconv/ftoa.go) と [Go 1.27.0](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/ftoa.go) のタグ付きソースと変更履歴を比較する。リリースノートに記載がなくても「変更なし」とは結論しない | `Floating Point Formatting` → `Floating-Point Printing and Parsing Can Be Simple And Fast` → `Shortest-Width Printing` → [シリーズ目次](https://research.swtch.com/fp-all) |
| goroutine 間の読み書きを順序付ける規則を、ハードウェア・言語・Go の各層から追いたい | [The Go Memory Model](https://go.dev/ref/mem) で現在の保証を確認し、具体例を race detector と実測で照合する | `Memory Models` → `Hardware Memory Models`、`Programming Language Memory Models`、`Updating the Go Memory Model` → [シリーズ目次](https://research.swtch.com/mm) |
| 言語・標準ライブラリ・ランタイムの設計案が、どの手続きを経て採用または見送りになったのか | [Go proposal process](https://go.dev/s/proposal) から proposal Issue と決定を確認し、採用された変更だけ対象版のリリースノートと実装へ進む | `Go Proposals` → `Enabling Experiments`、`Representation` → [シリーズ目次](https://research.swtch.com/proposals) |
| クリーンなコンパイラソースと、手元のコンパイラをどこまで信頼できるのか | [Go compiler](https://go.dev/cmd/compile/) → [Installing Go from source](https://go.dev/doc/install/source#go14) → [Reproducible Go toolchains](https://go.dev/blog/rebuild) の順に、現在のブートストラップと再現可能性を確認する | `Reflections on Trusting Trust` → `bootstrap`、`reproducible` → [実演記事](https://research.swtch.com/nih)。`Open Source Supply Chain Security` → [講演・参考資料の入口](https://research.swtch.com/acmscored) |
| 依存モジュールへの攻撃に対して、Go の版選択・検証・取得経路がどこを守るのか | [How Go Mitigates Supply Chain Attacks](https://go.dev/blog/supply-chain) → [Go Modules Reference](https://go.dev/ref/mod) の順に、現在の仕組みと限界を確認する | `Open Source Supply Chain Security` → [講演・参考資料の入口](https://research.swtch.com/acmscored)。`Colors Attack` → `dependency`、`latest` → [比較記事](https://research.swtch.com/npm-colors) |
