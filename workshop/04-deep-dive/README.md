[シナリオ一覧](../SCENARIOS.md) | [ワークショップ進行ガイド](../README.md) | [チームでの進め方](../TEAM_GUIDE.md)

# 04-deep-dive の調べ方（逆引き手順）

このカテゴリでは、Go の言語仕様や標準ライブラリの実装、設計の背景を一次情報の順番に沿って調べます。各シナリオでは、まず共通の入口から観測した現象を言葉にし、根拠を一つずつたどってください。

## Deep Dive したい目的から逆引きする

| 知りたいこと | 最初に開く入口 | 判断・経緯・実装までの追い方 |
| --- | --- | --- |
| 新しい言語機能、ツール、標準ライブラリ API や既存関数の変更を知りたい | [Release History](https://go.dev/doc/devel/release) から対象版のリリースノートを開く | 変更の節 → 表示リンクまたは HTML コメントの Issue / CL → Issue の議論と決定 → CL の差分・レビュー → 対象タグの実装 → 実測 |
| Proposal review meeting で最近どの提案が検討され、どう判断されたか知りたい | [Proposal review meeting minutes](https://go.dev/s/proposal-minutes) の最新コメントを開く | 毎週投稿される minutes に並ぶ Issue 番号と状態 → 興味のある Proposal Issue の全議論 → design document → 関連 CL → 対象タグの実装。minutes は議論の索引として使い、判断理由は各 Issue で確認する |
| 関心のあるトピックについて、採用前・見送りを含む Proposal を探したい | `golang/go` Issues の [`Proposal` ラベル](https://github.com/golang/go/issues?q=is%3Aissue%20label%3AProposal)を開く | 機能名・パッケージ名で絞る → Issue の Description・全コメント・状態 → meeting minutes → 必要なら design document とそのレビュー CL → 採用なら実装 CL と対象タグ、見送りなら最終判断コメント |

Proposal の資料は、次のように役割を分けて使います。

- **手続き**: [Go proposal process](https://go.dev/s/proposal) で、提案、議論、Design doc、採否決定の流れを確認する。Proposal は `Proposal` ラベルの付いた GitHub Issue として始まり、必要な場合にだけ Design doc が作られる。
- **週次のレビュー結果**: [Proposal review meeting minutes](https://go.dev/s/proposal-minutes) のコメントから、その週に状態が動いた Issue を見つける。会議はおおむね毎週開かれ、結果は会議後に投稿される。minutes だけで理由を推測せず、リンク先 Issue の全コメントを読む。
- **Design doc の実体**: [Go proposal repository](https://cs.opensource.google/go/proposal/+/master:README.md) でリポジトリと README を確認し、[design ディレクトリ](https://github.com/golang/proposal/tree/master/design) から `design/<Issue番号>-<名前>.md` を探す。
- **Design doc の変更とレビュー**: Gerrit の [`project:proposal` 検索](https://go-review.googlesource.com/q/project:proposal)で、Design doc を変更した CL、patch set、レビューコメント、状態を追う。設計内容の議論は Proposal Issue、文書の修正履歴は Gerrit と役割が異なる。

Proposal Issue、Design doc、その文書を変更した CL、採用後の実装 CL は別の資料です。Issue 番号、Design doc のファイル名、各 CL の commit message にある相互参照を使って接続してください。

## 1. まず言語仕様を読む

構文、型、比較可能性、メモリサイズ、エラー処理などのルールを調べるときは、最初に [The Go Programming Language Specification](https://go.dev/ref/spec) を開きます。

1. 目次または Ctrl+F / Cmd+F で、シナリオ本文から得た概念に近い節を探す。
2. 該当する定義、制約、例外条件を読み、主張を仕様の文言に対応付ける。
3. 仕様だけで決まらない実装上の挙動は、次の標準ライブラリのドキュメントとソースへ進む。

## 2. 標準ライブラリの契約を読む

標準パッケージの API や設計上の注意点は、[Go Documentation](https://go.dev/doc) から公式ドキュメントを確認し、必要に応じて [pkg.go.dev/std](https://pkg.go.dev/std) のパッケージ一覧から対象ページを開きます。

- Overview と対象 API のコメントを読む。
- Example と関連する型・メソッドを確認する。
- ドキュメントの宣言から Go 本体のソースへのリンクをたどり、説明と実装を照合する。

## 3. リリースノートから Issue を見つける

最近の変更は、まず [Release History](https://go.dev/doc/devel/release) から対象版のリリースノート（`go.dev/doc/go1.<version>`）を開きます。リリースノートは変更の索引であり、それだけを設計理由や実装の根拠にはしません。対象の節、変更された機能名、互換性や切り替え条件を記録してから、対応する Issue と CL（Change List）へ進みます。

Issue 番号は本文にリンクされているとは限りません。Go のリリースノートは、対応する Issue を HTML コメントにだけ残すことがあります。ブラウザで次のように探します。

1. リリースノートを開いて Developer Tools を起動する。
2. **Elements** で `go.dev/issue/` を検索する。コメントを見つけにくい場合は、**Network** でページを再読み込みし、HTML document の **Response** を同じ語で検索する。
3. `issue/`、`/issue/`、`CL ` も検索し、対象の説明の直前または直後にある番号を記録する。
4. `<!-- go.dev/issue/77273 -->` なら [go.dev/issue/77273](https://go.dev/issue/77273)、`CL 792780` なら [go.dev/cl/792780](https://go.dev/cl/792780) を開く。`go.dev` の短縮 URL は、それぞれ GitHub Issue と Gerrit の CL へ転送される。

たとえば [Go 1.27 リリースノート](https://go.dev/doc/go1.27#language) の Generic Methods の直前には、HTML ソース上に `<!-- go.dev/issue/77273 -->` があります。また、同じリリースノートには Issue と CL を併記したコメントや、本文から直接開ける [CL 792780](https://go.dev/cl/792780) もあります。この記法は Go 本体の [Release Notes README](https://github.com/golang/go/blob/master/doc/README.md) にも定義されています。

Issue を開いたら、次を順に確認します。

- **Description**: 観測された問題、提案、再現例、対象範囲。
- **Labels / Milestone / State**: proposal なのか bug なのか、対象リリース、accepted・closed の状態。ただし `Closed` だけで対象版への実装完了とは判断しない。
- **全コメントと Timeline**: 反対案、制約、決定、gopherbot が追加した関連 Issue・CL・commit。ページ内で `CL`、`go-review`、`commit`、対象パッケージ名を検索する。
- **リンク先**: proposal/design document、先行 Issue、後続 Issue、CL。提案の採用と、各実装・ツールの追随を分けて記録する。

公式ブログを横断して探すときは [Go Blog の一覧](https://go.dev/blog/all) を使います。検索で見つけた外部記事だけを根拠にせず、必ずリリースノート、Issue、CL、バージョン付きソースへ戻って確認してください。

## 4. cs.opensource.google で実装ソースを読む

[cs.opensource.google](https://cs.opensource.google/) は、Google が公開しているオープンソースプロジェクトを、リポジトリ・版・ファイル・シンボル・参照関係から検索して読める Code Search です。Go 本体は [go/go リポジトリ](https://cs.opensource.google/go/go) から検索します。仕様書や API ドキュメントの代わりではなく、特定の版で実装がどうなっているかを確認するために使います。

1. `pkg.go.dev` の宣言リンク、Issue、CL の **Files** から対象パッケージ・シンボル・ファイル名を得る。
2. 最初は `symbol:FormatFloat`、`function:mallocgc`、`file:ftoa.go`、`comment:TODO`、`usage:mallocgc` のように検索する。詳しい絞り込みは [Code Search の検索構文](https://developers.google.com/code-search/reference) を参照する。
3. 検索結果を開いたら、画面上部の revision が `master` かタグかを確認する。`master` は場所を発見する用途にとどめ、結論には `refs/tags/go1.27.0` のような対象版を選ぶ。
4. 宣言だけでなく、直前の doc comment、関数全体、呼び出す側、関連する型・定数、同じディレクトリのテストを読む。`x` でシンボルの参照先、`o` でファイルの outline を開ける。
5. 実装中のコメントは、前提、例外、TODO、過去の不具合を見つける手掛かりになる。ただしコメントは言語仕様や公開 API の契約ではないため、仕様書・API ドキュメント・テスト・実測と照合する。
6. その行が「いつ・なぜ」変わったかは、`b` で blame、`h` でファイルの revision history を開く。変更した commit、対応する Issue、CL の番号を探し、レビューへ戻る。
7. README に根拠を残すときは、Links メニューの current revision へのリンク（ショートカット `lr`）を使う。HEAD へのリンク（`lh`）だけでは後から内容が変わるため、対象タグと行を含む URL を優先する。

たとえば URL は `https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/<path>` の形になります。異なるタグで同じファイルを開いて比較すると、リリースノートに書かれていない内部変更も確認できます。ただし、リリースノートに記載がないことを「変更がない」証拠にはしません。

## 5. Issue から CL と commit を追う

CL（Change List）は、Go プロジェクトが Gerrit でレビューする一つの変更単位です。[Go Contribution Guide](https://go.dev/doc/contribute) と [go-review.googlesource.com](https://go-review.googlesource.com/) から、実装差分だけでなく、patch set ごとの修正、行コメント、reviewer の判断、テスト結果、最終的な merge 状態を確認できます。

1. Issue の Timeline や gopherbot のコメント、リリースノートの HTML、commit message から `CL <番号>` または `go-review.googlesource.com` のリンクを探す。見つからない場合は Gerrit で Issue 番号やシンボル名を検索する。
2. CL の **Status** と対象 branch を確認する。`MERGED`、`ABANDONED`、未 merge を区別し、提案中の patch set をリリース済み実装として扱わない。
3. commit message の目的、`Fixes #...` / `Updates #...`、変更者、reviewer を読む。`Fixes` は Issue を閉じる変更、`Updates` は途中の変更であることが多いため、後続 CL の有無も確認する。
4. **Files** で最終 patch set の差分と追加・更新されたテストを読む。ソースコード内のコメントと、reviewer が差分に付けたレビューコメントを混同しない。
5. **Change Log / Comments** で patch set 間の修正理由、未解決の指摘、trybot・テスト結果を確認する。必要なら related changes から前提・後続の CL もたどる。
6. merge 後の commit hash を確認し、cs.opensource.google の対象タグにその変更が入っているかを確かめる。`master` に merge 済みでも、調べているリリースタグには未収録の場合がある。

調査結果には「Issue で決まったこと」「CL で実装されたこと」「対象タグで確認できたこと」「実測したこと」を分けて書きます。一つの CL が Issue 全体を実装するとは限らず、コンパイラ、標準ライブラリ、`x/tools` などに複数の CL が分かれる場合があります。

> **小ネタ: go.dev のショートハンド**
>
> - `https://go.dev/issue/<Issue番号>` は `golang/go` の GitHub Issue を開く。例: [go.dev/issue/77273](https://go.dev/issue/77273)
> - `https://go.dev/cl/<CL番号>` は Go の Gerrit CL を開く。例: [go.dev/cl/792780](https://go.dev/cl/792780)
>
> Issue や CL の番号だけ分かっているときは、この形で直接開けます。調査メモや README で共有するときも、GitHub・Gerrit の長い URL より番号との対応が読み取りやすくなります。

## 6. 手元で挙動を再現する

- 実行可能な例は、README に記載された作業ディレクトリから同じコマンドを実行する。
- `go version`、終了コード、標準出力、標準エラーを記録し、ドキュメントの説明と突き合わせる。
- コンパイルエラーの例も同じツールチェーンで実行し、実際の診断を記録する。
- 小さなコードをブラウザで試す場合は [Go Playground](https://go.dev/play/) を使い、Share したソースと README のコードが一致していること、表示された Go のバージョンを確認する。

## 調査の進め方

観測した現象 → リリースノート → HTML に埋め込まれた Issue / CL → Issue の議論と決定 → CL の差分とレビュー → バージョン付き実装ソース → 実測、の順に進めます。言語の規則や公開 API の契約を扱う場合は、途中で言語仕様・API ドキュメントも確認します。一次資料で結論を固定できない場合は、断定せず「不明」または「ここからは推測」と明記してください。

## 設計史・実装を research.swtch.com から逆引きする

Russ Cox は、[2008 年に Go の開発チームへ参加し、2 つのコンパイラと標準ライブラリの構築に携わった](https://go.dev/blog/toward-go2)ソフトウェアエンジニアで、のちに [Go プロジェクトと Google の Go チームで技術面を率いる technical lead を務めました](https://go.dev/blog/open-source)。[research!rsc の目次](https://research.swtch.com/) は彼の個人サイトで、Go の設計判断や実装の経緯を、中心的な開発者の視点からたどるための資料です。

research!rsc は Go プロジェクトの公式ドキュメントではありません。表の go.dev の入口で現在の仕様・対象版・実装を先に固定し、目次を記事の題名で Ctrl+F / Cmd+F して候補を開き、残りの題名・語をシリーズ内または記事内で検索します。最後にもう一度バージョン付きソースと実測へ戻ってください。

[Go: A Documentary](https://golang.design/history/) は、言語設計、コンパイラ、ランタイム、標準ライブラリの歴史を、公開された設計文書・Issue・CL・講演から逆引きする索引として使えます。ただし、サイト自身が本文は公開情報に基づく主観的な理解であり誤りもあり得ると注意しており、項目が追加されても過去時点の役割や状況を述べた本文が残ることがあります。最近の状態を網羅する資料とはみなさず、リンク先の一次資料と対象版のソースへ進んでください。

最近の実装を人から逆引きする場合は、リリースノートや proposal Issue で見つけた著者・実装者・reviewer を起点に、GitHub と Gerrit の直近の活動へ進みます。たとえば [Alan Donovan](https://github.com/adonovan) が起票した [golang/go#77549](https://github.com/golang/go/issues/77549) から、Generic Methods に追随する `x/tools` と外部ツールの作業を追えます。

[Russ Cox](https://github.com/rsc) の活動は過去の設計史を探す入口になります。一方、個人の現在の活動量だけでは、Go プロジェクトの方針や担当を判断できません。[Go Code Owners](https://dev.golang.org/owners) も担当候補と変更履歴の入口として使い、直近の Issue・CL・レビューで確かめます。

| 調べたい課題 | go.dev から先に確認すること | research!rsc の目次で探す題名 → 次に探す題名・語 |
| --- | --- | --- |
| 浮動小数点の出力を変えずに、文字列化の内部実装を置き換えられる理由を追いたい | [`strconv.FormatFloat`](https://go.dev/pkg/strconv/#FormatFloat) で現在の契約を確認し、[Go 1.26.0](https://cs.opensource.google/go/go/+/refs/tags/go1.26.0:src/internal/strconv/ftoa.go) と [Go 1.27.0](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/ftoa.go) のタグ付きソースと変更履歴を比較する。リリースノートに記載がなくても「変更なし」とは結論しない | `Floating Point Formatting` → `Floating-Point Printing and Parsing Can Be Simple And Fast` → `Shortest-Width Printing` → [シリーズ目次](https://research.swtch.com/fp-all) |
| goroutine 間の読み書きを順序付ける規則を、ハードウェア・言語・Go の各層から追いたい | [The Go Memory Model](https://go.dev/ref/mem) で現在の保証を確認し、具体例を race detector と実測で照合する | `Memory Models` → `Hardware Memory Models`、`Programming Language Memory Models`、`Updating the Go Memory Model` → [シリーズ目次](https://research.swtch.com/mm) |
| 言語・標準ライブラリ・ランタイムの設計案が、どの手続きを経て採用または見送りになったのか | [Go proposal process](https://go.dev/s/proposal) から proposal Issue と決定を確認し、採用された変更だけ対象版のリリースノートと実装へ進む | `Go Proposals` → `Enabling Experiments`、`Representation` → [シリーズ目次](https://research.swtch.com/proposals) |
| クリーンなコンパイラソースと、手元のコンパイラをどこまで信頼できるのか | [Go compiler](https://go.dev/cmd/compile/) → [Installing Go from source](https://go.dev/doc/install/source#go14) → [Reproducible Go toolchains](https://go.dev/blog/rebuild) の順に、現在のブートストラップと再現可能性を確認する | `Reflections on Trusting Trust` → `bootstrap`、`reproducible` → [実演記事](https://research.swtch.com/nih)。`Open Source Supply Chain Security` → [講演・参考資料の入口](https://research.swtch.com/acmscored) |
| 依存モジュールへの攻撃に対して、Go の版選択・検証・取得経路がどこを守るのか | [How Go Mitigates Supply Chain Attacks](https://go.dev/blog/supply-chain) → [Go Modules Reference](https://go.dev/ref/mod) の順に、現在の仕組みと限界を確認する | `Open Source Supply Chain Security` → [講演・参考資料の入口](https://research.swtch.com/acmscored)。`Colors Attack` → `dependency`、`latest` → [比較記事](https://research.swtch.com/npm-colors) |
