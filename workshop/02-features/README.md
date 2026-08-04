[ワークショップに戻る](../README.md)

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

- **最新バージョンの機能を調査したい知りたい**
  - `https://go.dev/doc/go1.xx` (xxはGoのマイナーバージョン)でリリースノートにアクセスできます。
  - Go 1.27 の場合はリリースノート https://go.dev/doc/go1.27 を参照。
  - リリースノート内の内容から議論をたどりたい場合、開発者ツールでHTMLソースコードを参照しましょう。
    - `go.dev/issue/<Issue番号>` のような形で 関連する GitHub Issue番号 が埋め込まれています。
    - `https://go.dev/issues/<Issue番号>` を開き、関連するissueを開きましょう。

- **実際の挙動・実装詳細を知りたい**
  - `pkg.go.dev` のリンクから Go 本体のソースコード検索（ https://cs.opensource.google/go/go ）へ遷移し、標準ライブラリやランタイムの実装コードを読み込む。

- **手元で挙動を試したい**
  * Go Playground（ https://go.dev/play/ ）を利用する。開発版（Dev版）や過去のバージョンに切り替えて挙動を比較・検証する。
