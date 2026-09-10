[シナリオ一覧](../SCENARIOS.md) | [ワークショップ進行ガイド](../README.md) | [チームでの進め方](../TEAM_GUIDE.md)

# 02-features の調べ方（逆引き手順）

このカテゴリでは、Goの言語仕様や公式ドキュメント 、公式ブログ・プロポーザルを読み解いてGoの機能や文法構造を探索します。困ったときは、まず以下の手順で一次情報にあたりましょう。

- **言語仕様・文法ルールを知りたい**
  - https://go.dev/ref/spec を開く。
  - 公式ドキュメントは英語であるため、調べたい内容（やりたいこと・概念）に合わせて、以下のキーワード（英語）でページ内検索（Ctrl+F / Cmd+F）すると目的の仕様にたどり着きやすいです。
    - **変数・定数**
      - 変数宣言、初期化、`:=` の詳細ルール : `Variable declarations`, `Short variable declarations`
      - 定数のルール、`iota` を使った連番 : `Constant declarations`, `Iota`
    
    - **型（データ構造）**
      - 独自の型の定義、型エイリアス : `Type declarations`, `Alias declarations`
      - 構造体の定義、フィールドの埋め込み : `Struct types`, `Embedded field`
      - 配列とスライスの違い、仕様 : `Array types`, `Slice types`
      - マップ（連想配列）の仕様 : `Map types`
      - インターフェースの仕様 : `Interface types`
      - ポインタ型の仕様 : `Pointer types`
      - ジェネリクス、型パラメータ、制約 : `Type parameters`, `Type constraints`, `Core types`

    - **関数・メソッド**
      - 関数の定義、戻り値、可変長引数 : `Function declarations`
      - 構造体や型に紐づくメソッドの定義 : `Method declarations`
      - 無名関数、クロージャ : `Function literals`

    - **制御構文**
      - `if` 文のルール : `If statements`
      - ループ処理、`range` を使った反復処理 : `For statements`, `For statements with range clause`
      - `switch` 文、条件分岐 : `Expression switches`
      - インターフェースの型による分岐・チェック : `Type assertions`, `Type switches`
      - ループのスキップ・脱出、ラベル指定のジャンプ : `Break statements`, `Continue statements`, `Labeled statements`
      - 関数を抜ける直前の遅延実行（クリーンアップ処理） : `Defer statements`

    - **並行処理（ゴルーチン・チャネル）**
      - 非同期処理（ゴルーチンの起動） : `Go statements`
      - チャネルの定義、送信・受信処理 : `Channel types`, `Send statements`, `Receive operator`
      - 複数のチャネルのノンブロッキングな待ち受け : `Select statements`

    - **メモリ・エラーハンドリング・その他**
      - メモリ割り当て（`new`, `make` の違い）: `Allocation`, `Making slices, maps and channels`
      - 型変換（キャスト）の厳密なルール: `Conversions`
      - パニックの発生と復帰（`panic`, `recover`）: `Handling panics`, `Run-time panics`
      - パッケージのインポートルール : `Import declarations`
      - ソースコードの字句解析ルール（コメント、セミコロンの自動挿入など）: `Lexical elements`, `Semicolons`

- **公式ドキュメントを一覧から探したい**
  - https://go.dev/doc を開く。
  - 標準パッケージの仕様や使い方を調べたい: https://pkg.go.dev/std
  - Goの基本的な文法や機能をブラウザ上で実際に動かしながら学びたい: https://go.dev/tour/ （A Tour of Go）
  - Goらしいコードの書き方（イディオムやベストプラクティス）を確認したい: https://go.dev/doc/effective_go （Effective Go）
  - Goコマンド（build, test, modなど）の詳しい使い方を知りたい: https://go.dev/cmd/go/
  - モジュールの仕組みや依存関係の管理方法について知りたい: https://go.dev/doc/modules/managing-dependencies
  - Goのメモリモデル（ゴルーチン間のメモリ可視性など）について知りたい: https://go.dev/ref/mem （The Go Memory Model）
  - データベースへの接続や操作方法をチュートリアルで学びたい: https://go.dev/doc/tutorial/database-access
  - Web API（RESTful API）の作り方をチュートリアルで学びたい: https://go.dev/doc/tutorial/web-service-gin
  - セキュリティに関する情報や脆弱性のデータベースを確認したい: https://go.dev/security/

- **公式ブログを一覧から探したい**
  - https://go.dev/blog/all を開く。
  - 調査したい内容が「昔からの言語仕様」の場合は下から探したほうが早い。
  - ブログのAuthorをベースに探すと目的のものを探しやすいです。
    - Rob Pike
      - 言語の内部メカニズムと仕様解説（文字列、スライス、appendの仕組み、リフレクションの法則、定数）
      - 設計思想・イディオム（Goの宣言構文、Errors are values）
      - ツールやエコシステム（go generate、go fmt）
      - 文化的な背景（The Go Gopher（マスコットの背景）、Go fonts）
    - Russ Cox
      - 長期的なビジョンと管理（Toward Go 2、前方・後方互換性とツールチェーン管理）
      - モジュールと依存関係（Go Modules、パッケージバージョニングの提案）
      - セキュリティとコアライブラリ（安全な乱数、コマンドPATHのセキュリティ）
      - アニバーサリー記事（毎年の「Happy Birthday, Go!」）
    - Robert Griesemer
      - 型システムの深い解説（型推論の仕組み、比較可能な型、コアタイプの廃止）
      - 言語仕様の拡張・構文（エイリアス名について、エラーハンドリングの構文サポート）
      - Go 2や新バージョンに向けたプロポーザルの主導
    - Ian Lance Taylor
      - ジェネリクスの全般（ジェネリクス導入の提案、いつ使うべきか、型パラメータの解体）
      - 新しい関数の機能（関数型に対するrange）
      - 低レイヤー・コンパイラ関連（Gccgo）
    - Andrew Gerrand
      - メジャー・マイナーリリースの公式アナウンス
      - 実践的なチュートリアル（JSONとGo、スライスの使い方、Webアプリの書き方）
      - ツールの紹介（The Go Playground、Godoc）
      - イベントレポート（Google I/O、GopherCon、各国のミートアップ情報）
    - セキュリティ関連: Filippo Valsorda, Julie Qiu
    - ツールやパッケージ関連: Alan Donovan, Damien Neil
    - 新しいバージョンのリリース: Go team

- **最新バージョンの機能を調査したい**
  - `go.dev/doc/go1.<version>` (version は Go のマイナーバージョン)でリリースノートにアクセスできます。
  - Go 1.27 の場合はリリースノート https://go.dev/doc/go1.27 を参照。
  - リリースノート内の内容から議論をたどりたい場合、開発者ツールでHTMLソースコードを参照しましょう。
    - `go.dev/issue/<Issue番号>` のような形で 関連する GitHub Issue番号 が埋め込まれています。
    - `https://go.dev/issue/<Issue番号>` を開き、関連する Issue の本文と議論を読みましょう。

- **実際の挙動・実装詳細を知りたい**
  - `pkg.go.dev` のリンクから Go 本体のソースコード検索（ https://cs.opensource.google/go/go ）へ遷移し、標準ライブラリやランタイムの実装コードを読み込む。

- **手元で挙動を試したい**
  * Go Playground（ https://go.dev/play/ ）を利用する。開発版（Dev版）や過去のバージョンに切り替えて挙動を比較・検証する。

## 設計・歴史を research.swtch.com から逆引きしたい

Russ Cox は、[2008 年に Go の開発チームへ参加し、2 つのコンパイラと標準ライブラリの構築に携わった](https://go.dev/blog/toward-go2)ソフトウェアエンジニアで、のちに [Go プロジェクトと Google の Go チームで技術面を率いる technical lead を務めました](https://go.dev/blog/open-source)。[research!rsc の目次](https://research.swtch.com/) は彼の個人サイトで、Go の設計判断や実装の経緯を、中心的な開発者の視点からたどるための資料です。

research!rsc は Go プロジェクトの公式ドキュメントではありません。まず go.dev で現在の仕様と対象バージョンを確認し、次に目次を記事の題名で Ctrl+F / Cmd+F して候補を開き、残りの題名・語をシリーズ内または記事内で検索します。過去の記事の説明と現在の仕様・実装は分けて記録してください。

[Go: A Documentary](https://golang.design/history/) は、Go の言語設計を、公開された設計文書・Issue・CL・講演から年代や論点で逆引きする索引として使えます。ただし、サイト自身が本文は公開情報に基づく主観的な理解であり誤りもあり得ると注意しており、項目が追加されても過去時点の役割や状況を述べた本文が残ることがあります。最近の状態を網羅する資料とはみなさず、リンク先の一次資料と対象版のリリースノート・仕様・実装で結論を更新してください。

最近の変更では、機能名だけでなく関係する開発者から追うと、周辺ツールまで含む作業を見つけられることがあります。たとえば [Alan Donovan のプロフィール](https://github.com/adonovan) から、Generic Methods 採用後の `x/tools`、gopls、vulncheck などの追随作業を整理した [golang/go#77549](https://github.com/golang/go/issues/77549) へ進めます。逆に [Russ Cox のプロフィール](https://github.com/rsc) は過去の設計史を探す入口にはなりますが、過去に Go チームを率いた人物の現在の活動量だけから、Go プロジェクトの現在の方針や担当を判断してはいけません。[Go Code Owners](https://dev.golang.org/owners) も担当候補と Gerrit 履歴を探す手掛かりとして使い、最後は直近の Issue・CL・レビューで確認します。

| 調べたい課題 | go.dev から先に確認すること | research!rsc の目次で探す題名 → 次に探す題名・語 |
| --- | --- | --- |
| goroutine 間の読み書きが、どの同期によって順序付けられるのか | [The Go Memory Model](https://go.dev/ref/mem) で `synchronized before`、`happens before`、`data race` の定義を確認する | `Memory Models` → `Hardware Memory Models`、`Programming Language Memory Models`、`Updating the Go Memory Model` → [シリーズ目次](https://research.swtch.com/mm) |
| 言語機能の変更が、どの提案と議論を経て採用・見送りになったのか | [Go proposal process](https://go.dev/s/proposal) から proposal Issue と決定を確認し、採用された変更だけ対象版のリリースノート・仕様・実装へ進む | `Go Proposals` → `Enabling Experiments`、`Representation` → [シリーズ目次](https://research.swtch.com/proposals) |
| channel を待つ goroutine の扱いが、過去の説明から現在の診断機能までどう変わったのか | [Go 1.27 リリースノート](https://go.dev/doc/go1.27) とリンク先の Issue・実装で現在の機能を固定する | `A Tour of Go` → Q&A の `goroutine`、`channel` → [記事](https://research.swtch.com/gotour) |


-------------


[Scenario index](../SCENARIOS_en.md) | [Workshop guide](../README.md) | [Team guide](../TEAM_GUIDE_en.md)

# How to Investigate 02-features (Reverse Lookup Guide)

In this category, you will explore Go features and syntactic structures by reading the Go language specification, official documentation, and official blogs and proposals. When you get stuck, start by consulting primary sources using the following steps.

- **I want to learn about language specifications and grammar rules**
  - Open https://go.dev/ref/spec
  - Because the official documentation is in English, searching the page with the following keywords (in English), based on what you want to investigate (the task, behavior, or concept), makes it easier to find the relevant specification. Use page search (Ctrl+F / Cmd+F).
    - **Variables and constants**
      - Variable declarations, initialization, and the detailed rules for `:=`: `Variable declarations`, `Short variable declarations`
      - Rules for constants and sequential numbering with `iota`: `Constant declarations`, `Iota`
    
    - **Types (data structures)**
      - Defining named types and type aliases: `Type declarations`, `Alias declarations`
      - Defining structs and embedding fields: `Struct types`, `Embedded field`
      - Differences between arrays and slices, and their specifications: `Array types`, `Slice types`
      - The specification for maps (associative arrays): `Map types`
      - The specification for interfaces: `Interface types`
      - The specification for pointer types: `Pointer types`
      - Generics, type parameters, and constraints: `Type parameters`, `Type constraints`, `Core types`

    - **Functions and methods**
      - Defining functions, return values, and variadic arguments: `Function declarations`
      - Defining methods associated with structs and types: `Method declarations`
      - Anonymous functions and closures: `Function literals`

    - **Control syntax**
      - Rules for `if` statements: `If statements`
      - Loops and iteration with `range`: `For statements`, `For statements with range clause`
      - `switch` statements and branching: `Expression switches`
      - Branching and checking by an interface's type: `Type assertions`, `Type switches`
      - Skipping and exiting loops, and jumps using labels: `Break statements`, `Continue statements`, `Labeled statements`
      - Deferred execution immediately before leaving a function (cleanup processing): `Defer statements`

    - **Concurrency (goroutines and channels)**
      - Asynchronous processing (starting goroutines): `Go statements`
      - Defining channels and sending and receiving: `Channel types`, `Send statements`, `Receive operator`
      - Waiting for multiple channels without blocking: `Select statements`

    - **Memory, error handling, and other topics**
      - Memory allocation (the difference between `new` and `make`): `Allocation`, `Making slices, maps and channels`
      - Strict rules for type conversions (casts): `Conversions`
      - Raising and recovering from panics (`panic`, `recover`): `Handling panics`, `Run-time panics`
      - Package import rules: `Import declarations`
      - Lexical rules for source code (comments, automatic semicolon insertion, and so on): `Lexical elements`, `Semicolons`

- **I want to find official documentation from a list**
  - Open https://go.dev/doc
  - To investigate the specifications and usage of standard packages: https://pkg.go.dev/std
  - To learn basic Go syntax and features by running them in a browser: https://go.dev/tour/ (A Tour of Go)
  - To check how to write idiomatic Go code (idioms and best practices): https://go.dev/doc/effective_go (Effective Go)
  - To learn more about Go commands (build, test, mod, and so on): https://go.dev/cmd/go/
  - To learn about modules and dependency management: https://go.dev/doc/modules/managing-dependencies
  - To learn about the Go memory model (such as memory visibility between goroutines): https://go.dev/ref/mem (The Go Memory Model)
  - To learn how to connect to and operate on databases through a tutorial: https://go.dev/doc/tutorial/database-access
  - To learn how to build a Web API (RESTful API) through a tutorial: https://go.dev/doc/tutorial/web-service-gin
  - To check security information and the vulnerability database: https://go.dev/security/

- **I want to find an official blog post from a list**
  - Open https://go.dev/blog/all
  - If you are investigating a topic involving an older language specification, it is faster to search from the list below.
  - Searching by blog author makes it easier to find what you need.
    - Rob Pike
      - The language's internal mechanisms and explanations of specifications (strings, slices, how `append` works, the laws of reflection, constants)
      - Design philosophy and idioms (Go's declaration syntax, Errors are values)
      - Tools and ecosystem (`go generate`, `go fmt`)
      - Cultural background (The Go Gopher (the mascot's background), Go fonts)
    - Russ Cox
      - Long-term vision and governance (Toward Go 2, forward and backward compatibility, and toolchain management)
      - Modules and dependencies (Go Modules, package versioning proposals)
      - Security and core libraries (secure random numbers, command PATH security)
      - Anniversary articles (the annual “Happy Birthday, Go!”)
    - Robert Griesemer
      - In-depth explanations of the type system (how type inference works, comparable types, the removal of core types)
      - Extensions to the language specification and syntax (alias names, syntax support for error handling)
      - Leading proposals for Go 2 and future versions
    - Ian Lance Taylor
      - Generics in general (the proposal to introduce generics, when to use them, deconstructing type parameters)
      - New function features (`range` over function types)
      - Low-level and compiler-related topics (Gccgo)
    - Andrew Gerrand
      - Official announcements of major and minor releases
      - Practical tutorials (JSON and Go, how to use slices, how to write Web applications)
      - Tool introductions (The Go Playground, Godoc)
      - Event reports (Google I/O, GopherCon, and meetup information from around the world)
    - Security-related: Filippo Valsorda, Julie Qiu
    - Tools and package-related: Alan Donovan, Damien Neil
    - New version releases: Go team

- **I want to investigate features in the latest version**
  - Access the release notes at `go.dev/doc/go1.<version>` (`version` is Go's minor version).
  - For Go 1.27, see the release notes at https://go.dev/doc/go1.27
  - If you want to follow the discussions behind the release-note entries, inspect the HTML source code using developer tools.
    - Related GitHub issue numbers are embedded in forms such as `go.dev/issue/<Issue number>`.
    - Open https://go.dev/issues/<Issue number> and then open the related issue.

- **I want to learn about actual behavior and implementation details**
  - Follow links from `pkg.go.dev` to the Go source code search at https://cs.opensource.google/go/go and read the implementation code for the standard library or runtime.

- **I want to try the behavior locally**
  * Use the Go Playground (https://go.dev/play/). You can switch to the development version (Dev version) or past versions to compare and verify behavior.

## Researching design and history through research.swtch.com

Russ Cox joined the Go team in 2008 and helped build two compilers and the standard library. He later served as a technical lead for the Go project and Google's Go team. His [research!rsc index](https://research.swtch.com/) is a personal site for tracing Go design decisions and implementation history from a core developer's perspective.

It is not official Go project documentation. First verify the current specification and target version on go.dev. Then search the index by article title and follow related titles and terms within the series or article. Keep historical explanations separate from the current specification and implementation.

[Go: A Documentary](https://golang.design/history/) is an index for tracing Go language design through public design documents, Issues, CLs, and talks. Treat it as an index rather than a complete current-status reference, and update conclusions with the target version's release notes, specification, and source.

For recent changes, following the people involved can reveal related work in tools. For example, [Alan Donovan's profile](https://github.com/adonovan) leads to [golang/go#77549](https://github.com/golang/go/issues/77549), which tracks follow-up work in `x/tools`, gopls, and vulncheck after Generic Methods. Do not infer the current direction of Go from a person's activity alone; use [Go Code Owners](https://dev.golang.org/owners) and verify with recent Issues, CLs, reviews, and versioned source.

| Question | Check first on go.dev | Search in the research!rsc index |
| --- | --- | --- |
| Which synchronization orders reads and writes between goroutines? | Read the definitions of `synchronized before`, `happens before`, and `data race` in [The Go Memory Model](https://go.dev/ref/mem). | `Memory Models` -> `Hardware Memory Models`, `Programming Language Memory Models`, and `Updating the Go Memory Model` -> [series index](https://research.swtch.com/mm) |
| Which proposal and discussion led to a language change being accepted or rejected? | Start with the [Go proposal process](https://go.dev/s/proposal), then follow the proposal Issue and decision. For accepted changes, continue to the target release notes, specification, and implementation. | `Go Proposals` -> `Enabling Experiments` and `Representation` -> [series index](https://research.swtch.com/proposals) |
| How has the treatment of goroutines waiting on channels changed from historical explanations to current diagnostics? | Fix the current behavior using the [Go 1.27 release notes](https://go.dev/doc/go1.27), linked Issues, and implementation. | `A Tour of Go` -> the `goroutine` and `channel` Q&A -> [article](https://research.swtch.com/gotour) |
