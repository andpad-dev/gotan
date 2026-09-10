[シナリオ一覧](../SCENARIOS.md) | [ワークショップ進行ガイド](../README.md) | [チームでの進め方](../TEAM_GUIDE.md)

# 01-packages の調べ方（逆引き手順）

このカテゴリでは、標準パッケージのドキュメント（pkg.go.dev）を中心に、必要に応じて言語仕様も読み解いて課題に取り組みます。
困ったときは、まずここに戻ってきてください。

調査は [Go Documentation](https://go.dev/doc) を入口にします。標準パッケージの API 契約へ進むときは、そこから package documentation を辿るか、以下の逆引き手順で pkg.go.dev を開いてください。

## 関数・型・メソッドが何か知りたい

1. [Go の公式ドキュメント](https://go.dev/doc/) を入口にし、Packages からパッケージドキュメントへ進みます。
2. Goパッケージ全体から検索したい場合は https://pkg.go.dev/ を開きます。
3. 探すパッケージがわかっている場合は `https://pkg.go.dev/<パッケージ名>` を開きます（例: https://pkg.go.dev/fmt ）。
    - 標準パッケージの場合は、https://pkg.go.dev/std からパッケージ一覧を開くと便利です。
    - また標準パッケージの場合はリダイレクトも充実しています。例えば、`slog` のパッケージは `log/slog` が正解ですが、https://pkg.go.dev/slog でも自動的にリダイレクトされます。
4. `f` キーを押すと検索ダイアログが出るので、関数や型の名前を打ち込んでジャンプします。
5. まず名前と説明の最初の 1 文でざっくりと予想を立てます。英語が難しければ翻訳しても構いません。最初からすべてを読む必要はありません。

## パッケージ全体の使い方・約束事を知りたい

- ページ冒頭の **Overview** を読みます。書式や注意点など、個々の関数には書かれていない説明がここに集まっています。

## 使い方の実例が見たい

- **Examples** セクション（`#pkg-examples`）を見ます。Example はページ上の Run ボタンでその場で実行できます。

## 言語構文・仕様を知りたい

1. [Go 言語仕様](https://go.dev/ref/spec)を開きます。
2. 目次またはブラウザ内検索で、シナリオ本文に出てきた構文名を探します。例えば `range` なら **For statements** から下位の **For statements with range clause** へ進みます。
3. 親の節が構文定義だけで終わる場合は、直後の下位節まで確認します。

## 挙動を確かめたい

- 手元で `go run` します。手元に環境がなければ [Go Playground](https://go.dev/play/) に貼って実行します。班へ渡すときは Share で URL を作ります。
- 実行結果が資料やシナリオ本文と違うときは、まず `go version`、`go env GOMOD`、`go env GODEBUG` を記録します。Go のバージョンに加え、メインモジュールの `go` 行や `GODEBUG` が挙動を切り替える機能があります。
- バージョン依存の挙動は、シナリオ本文に示された Go Playground の実行結果を基準にします。ローカルで比較する場合は一時モジュールの `go` 行を明示し、Playground と同じ条件を作れない場合は無理に結論を出さず Playground で確認します。

## ドキュメントに書いていないことを知りたい

- 関数や型の宣言部分のリンクをクリックすると、実装のソースコードへジャンプできます。
- Go 本体のソースをまとめて読むなら https://cs.opensource.google/go/go を使います。
- https://cs.opensource.google/go/go の調査の仕方は [04-deep-dive](../04-deep-dive/README.md) を参照してください。

## 設計・実装の背景を research.swtch.com から逆引きしたい

Russ Cox は、[2008 年に Go の開発チームへ参加し、2 つのコンパイラと標準ライブラリの構築に携わった](https://go.dev/blog/toward-go2)ソフトウェアエンジニアで、のちに [Go プロジェクトと Google の Go チームで技術面を率いる technical lead を務めました](https://go.dev/blog/open-source)。[research!rsc の目次](https://research.swtch.com/) は彼の個人サイトで、Go の設計判断や実装の経緯を、中心的な開発者の視点からたどるための資料です。

research!rsc は Go プロジェクトの公式ドキュメントではありません。記事が公開された時点の説明を現在の API 契約として扱わないよう、先に表の go.dev の入口から現行ドキュメントと対象バージョンを確認してください。その後、目次を記事の題名で Ctrl+F / Cmd+F して候補を開き、残りの題名・語をシリーズ内または記事内で検索します。

[Go: A Documentary](https://golang.design/history/) は、言語、標準ライブラリ、ツールチェーンなどの歴史を、公開された設計文書・Issue・CL・講演から逆引きする索引として使えます。ただし、サイト自身が本文は公開情報に基づく主観的な理解であり誤りもあり得ると注意しており、項目が追加されても過去時点の役割や状況を述べた本文が残ることがあります。最近の状態を網羅する資料とはみなさず、必ずリンク先の一次資料と対象版のソースへ進んでください。

現在の実装を誰が扱っているか探すときは、[Go Code Owners](https://dev.golang.org/owners) をパッケージ名で検索すると、レビューの primary / secondary owner と Gerrit の変更履歴への入口が得られます。これは探索の手掛かりであって、現在の担当を証明する名簿ではありません。owner や個人の GitHub プロフィールを見つけた後は、直近の Issue・CL・レビューとバージョン付きソースで結論を確認します。

| 調べたい課題 | go.dev から先に確認すること | research!rsc の目次で探す題名 → 次に探す題名・語 |
| --- | --- | --- |
| `fmt` や `strconv` が浮動小数点数をどう文字列化し、どこで丸めるのか | [Go Documentation](https://go.dev/doc) → [`strconv.FormatFloat`](https://go.dev/pkg/strconv/#FormatFloat) で現在の契約を確認し、[Go 1.26.0](https://cs.opensource.google/go/go/+/refs/tags/go1.26.0:src/internal/strconv/ftoa.go) と [Go 1.27.0](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/ftoa.go) のタグ付きソースと変更履歴を比較する | `Floating Point Formatting` → `Floating-Point Printing and Parsing Can Be Simple And Fast` → `Shortest-Width Printing` → [シリーズ目次](https://research.swtch.com/fp-all) |
| Go の版で標準パッケージの挙動が変わり、古い挙動に依存する呼び出し箇所を絞り込みたい | 対象版のリリースノートと、たとえば [Go 1.23 の Timer Channel Changes](https://go.dev/wiki/Go123Timer) で変更条件を固定する | `Hash-Based Bisect Debugging` → `runtime`、`GODEBUG` → [記事](https://research.swtch.com/bisect) |
| パッケージの API 例や境界条件のテストを、小さく読みやすく組み立てたい | [Add a test](https://go.dev/doc/tutorial/add-a-test) → [`testing`](https://pkg.go.dev/testing) の順に、現在のテスト API と実行方法を確認する | `Go Testing By Example` → `testdata`、`test failures` → [記事](https://research.swtch.com/testing) |

## pkg.go.dev の小技

### キーボードショートカット
- `?` キーを押すとキーボードショートカットの一覧が出ます。
- `/` キーを押すと検索ボックスにフォーカスします。
- `f` キーを押すと Jump to のダイアログが出ます。使いたい関数や型の名前が分かっているときは、ここから説明へ直接飛べます。本文や節見出しを探すときは、ブラウザの Ctrl+F / Cmd+F を使います。
- `y` キーを押すと現在のURLを特定のバージョンへのパーマリンクへ変換します。
    - 例えば `@latest` や `@master` のページを閲覧中に `y` を押すと、その時点の正確なバージョン（例: v1.27.0 やコミットの擬似バージョン）のURLに変わり、他者への共有時にドキュメントの内容が変化するのを防ぎます。

### OS/アーキテクチャの指定
- URLの末尾に `?GOOS=<OS名>&GOARCH=<アーキテクチャ名>` を付与します。
- 例えばWindows(amd64)の場合 https://pkg.go.dev/bufio?GOOS=windows&GOARCH=amd64 のように指定します。
- URLに直接指定しなくても `Rendered for` からOS/アーキテクチャを切り替えられます。

### Version 一覧
- ページ上部の Version をクリックすると、利用可能なバージョンの一覧が表示されます。

### 依存パッケージ一覧
- ページ上部の `Imports: n` をクリックすると依存しているパッケージの一覧が表示されます。
- ページ上部の `Imported by: n` をクリックすると 逆にそのパッケージに依存しているパッケージの一覧が表示されます。


---------------------


[Scenario index](../SCENARIOS_en.md) | [Workshop guide](../README.md) | [Team guide](../TEAM_GUIDE_en.md)

# How to Research 01-packages

This category focuses on the standard-library documentation at pkg.go.dev and, when necessary, the Go language specification. Return here whenever you are unsure where to start.

Begin with [Go Documentation](https://go.dev/doc). To reach a standard-package API contract, follow the package documentation from there or use the reverse lookup below.

## I Want to Know What a Function, Type, or Method Is

1. Start at [Go Documentation](https://go.dev/doc/) and follow Packages to the package documentation.
2. For a search across Go packages, open https://pkg.go.dev/.
3. If you know the package, open `https://pkg.go.dev/<package-name>` (for example, https://pkg.go.dev/fmt).
   - For standard packages, [the standard-library list](https://pkg.go.dev/std) is useful.
   - Redirects are also available: `https://pkg.go.dev/slog` leads to `log/slog`.
4. Press `f` to open the search dialog, then enter the function or type name.
5. Make a rough prediction from the name and first sentence. You do not need to read everything at once.

## I Want to Know How to Use a Whole Package

Read the **Overview** near the top of the page. It contains formatting rules and conventions that individual functions may not repeat.

## I Want to See Examples

Open the **Examples** section (`#pkg-examples`). Examples can be run with the page's Run button.

## I Want to Know Language Syntax and the Specification

1. Open the [Go language specification](https://go.dev/ref/spec).
2. Search for the syntax name from the scenario. For example, follow **For statements** to **For statements with range clause** for `range`.
3. If the parent section only defines the syntax, read the immediately following child section too.

## I Want to Verify Behavior

Run `go run` locally, or paste a small `package main` into the [Go Playground](https://go.dev/play/). Share Playground URLs with the team.

When output differs from the material, record `go version`, `go env GOMOD`, and `go env GODEBUG`. The Go version, the module's `go` line, and `GODEBUG` can change behavior. For version-dependent behavior, use the Playground result stated in the scenario as the reference.

## I Want to Know What Is Not in the Documentation

Click the declaration of a function or type to open its implementation source. For the Go source tree, use https://cs.opensource.google/go/go. See [04-deep-dive](../04-deep-dive/README.md) for ways to investigate it.

## I Want to Cross-Reference Design and Implementation from research.swtch.com

[research!rsc](https://research.swtch.com/) is Russ Cox's personal site and is useful for following Go design decisions and implementation history from a core developer's perspective. It is not official Go documentation. Check current go.dev documentation and the target version first, then use the site's table of contents and browser search to find related articles.

[Go: A Documentary](https://golang.design/history/) is an index into public design documents, issues, CLs, and talks. Treat it as an index rather than a complete current-status reference, and follow its primary sources.

When looking for current ownership, search [Go Code Owners](https://dev.golang.org/owners) by package name. Owners are an investigation lead, not proof of current responsibility; verify conclusions with recent issues, CLs, reviews, and versioned source.

| Question | Start at go.dev | Search in research!rsc |
| --- | --- | --- |
| How `fmt` or `strconv` format floating-point numbers | [Go Documentation](https://go.dev/doc) -> [`strconv.FormatFloat`](https://go.dev/pkg/strconv/#FormatFloat), then compare versioned source | `Floating Point Formatting` -> `Floating-Point Printing and Parsing Can Be Simple And Fast` |
| How to narrow down behavior that changed between Go versions | The target release notes and version-specific documentation such as [Go 1.23 Timer Channel Changes](https://go.dev/wiki/Go123Timer) | `Hash-Based Bisect Debugging`, `runtime`, and `GODEBUG` |
| How to build small API and boundary-condition tests | [Add a test](https://go.dev/doc/tutorial/add-a-test) -> [`testing`](https://pkg.go.dev/testing) | `Go Testing By Example`, `testdata`, and `test failures` |

## pkg.go.dev Tips

### Keyboard shortcuts

- Press `?` to see the shortcut list.
- Press `/` to focus the search box.
- Press `f` to open Jump to. Use browser find (Ctrl+F / Cmd+F) for body text and section headings.
- Press `y` to turn the current URL into a version-specific permalink.

### OS and architecture

Add `?GOOS=<OS>&GOARCH=<architecture>` to a URL, for example `https://pkg.go.dev/bufio?GOOS=windows&GOARCH=amd64`. You can also change them from **Rendered for**.

### Version list and dependencies

Click **Version** at the top to see available versions. Click `Imports: n` for dependencies and `Imported by: n` for packages that depend on the package.
