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
