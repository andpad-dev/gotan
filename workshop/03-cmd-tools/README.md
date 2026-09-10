[シナリオ一覧](../SCENARIOS.md) | [ワークショップ進行ガイド](../README.md) | [チームでの進め方](../TEAM_GUIDE.md)

# 03-cmd-tools の調べ方（逆引き手順）

このカテゴリでは、`go` コマンドとその配下のサブコマンド・ツール（`go build`, `go run`, `go test`, `go vet`, `go generate`, `go doc`、`go tool trace`、`go tool pprof` など）が実際に何をしているかを、
公式ドキュメントと cmd/go のソースコードから調べます。困ったときは、まずここに戻ってきてください。

## 1. サブコマンドが何をするコマンドか知りたい（公式ヘルプを読む）

1. Web で調べるときは、まず [Go Command](https://go.dev/cmd/go/) を開く。`go` コマンド全体の公式ドキュメントへの入口になる。
2. 手元で `go help <サブコマンド>` を実行する（例: `go help run`, `go help generate`, `go help vet`）。`go` 本体に同梱されているサブコマンドは、これだけで概要と主要フラグが読める。
3. 同じ内容を Web で読むなら [`pkg.go.dev/cmd/go`](https://pkg.go.dev/cmd/go) を開き、Ctrl+F / Cmd+F でサブコマンド名を検索してその節（`#hdr-...` で終わる URL）へ移動する。`cmd/go` は関数や型を公開していないため、`f` の Jump to には何も出ない。
4. `go tool <サブコマンド>` として提供されるツール（`vet`、`pprof`、`trace`、`cover` など）は `go tool <サブコマンド> -h` や `go tool <サブコマンド> help` で使い方を確認できる。

## 2. コマンドの詳しい仕様やフラグを知りたい

- 単体のコマンド／ツールにはそれぞれ専用の pkg.go.dev ページがある: `https://pkg.go.dev/cmd/<サブコマンド>`（例: [cmd/vet](https://pkg.go.dev/cmd/vet), [cmd/go](https://pkg.go.dev/cmd/go), [cmd/pprof](https://pkg.go.dev/cmd/pprof)）。
- ページ冒頭の **Overview** に、コマンドが何をするかと主要なフラグ一覧がまとまっている。CLI の `-h` 出力より情報量が多いことも多い。

## 3. 実際に何が実行されているか、フラグで覗く

- `-n` フラグ: 実行はせず、内部で呼ばれるコマンド列だけを表示する（dry-run）。
- `-x` フラグ: 実際に実行しながら、呼ばれたコマンド列を逐次表示する。
- 使えるフラグはサブコマンドによって違うので、`go help <サブコマンド名>` のオプション一覧で確認する。

## 4. 実装まで踏み込みたい（ソースコードで裏取りする）

1. cmd/go の実装は https://cs.opensource.google/go/go の `src/cmd/go/internal/` 以下に、サブコマンドごとにパッケージが分かれている（例: `go run` → `internal/run`、`go build` → `internal/work`、`go generate` → `internal/generate`）。単体コマンド（`vet`, `pprof` など）は `https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/<サブコマンド>/` を直接開くか、pkg.go.dev のコマンドページ末尾のソースリンクをたどる。
2. サブコマンド名と同じ名前のファイル（`run.go` など）を開き、`Cmd*.Run` に登録されている関数（`runRun` など）を探す。ここが挙動を追いかけるエントリポイント。単体コマンドの場合は `doc.go` に `go help` で出るヘルプ本文が書かれていることが多く、まずここを読むと全体像がつかめる。
3. Ctrl+F / Cmd+F でページ内を検索しながら、その関数が呼んでいる関数へ 1 つずつジャンプする。呼び出し先の実装まで実際に読んで確認する。コメントや関数名の雰囲気だけで判断しない。
4. バージョンタグを手元の環境に合わせる。URL 中の `refs/tags/go1.27.0` のような部分は、`go version` の結果と揃える。

## 5. いつ・なぜ追加された機能か知りたい

- 対象の Go バージョンが分かっている場合、まず `go.dev/doc/go1.<version>` のリリースノートを見る。特に **Tools** セクション（[Go 1.26](https://go.dev/doc/go1.26#tools) / [Go 1.27](https://go.dev/doc/go1.27#tools)）に、`go` コマンドとツール群の変更点がまとまっている。
- リリースノートに Issue 番号やプロポーザルへのリンクがあれば、必ず該当 Issue を開き、議論と関連リンクまで含めて読む（`github.com/golang/go` の Issue、`go.googlesource.com/proposal`）。要約だけから挙動を推測しない。

## 6. 挙動を手元で確かめる

- 十分小さいコード（`package main` + `fmt.Println` 程度）を書き、`-n`（実行せず表示のみ）や `-x`（実行したコマンドを表示）、`-work` などの可視化フラグを付けて実際に出力を見る。
- ドキュメントやソースコードの記述と、手元での実行結果を突き合わせて確認する。

## 7. ツールの設計・歴史を research.swtch.com から逆引きする

Russ Cox は、[2008 年に Go の開発チームへ参加し、2 つのコンパイラと標準ライブラリの構築に携わった](https://go.dev/blog/toward-go2)ソフトウェアエンジニアで、のちに [Go プロジェクトと Google の Go チームで技術面を率いる technical lead を務めました](https://go.dev/blog/open-source)。[research!rsc の目次](https://research.swtch.com/) は彼の個人サイトで、Go の設計判断や実装の経緯を、中心的な開発者の視点からたどるための資料です。

research!rsc は Go プロジェクトの公式ドキュメントではありません。まず go.dev で現在のコマンドの契約を確認し、次に目次を記事の題名で Ctrl+F / Cmd+F して候補を開き、残りの題名・語をシリーズ内または記事内で検索してください。記事の公開年と手元の Go の版が違う場合は、考え方の説明と現在の実装を分けて検証します。

[Go: A Documentary](https://golang.design/history/) は、コンパイラ、ツールチェーン、`cmd/go` などの変遷を、公開された設計文書・Issue・CL・講演から逆引きする索引として使えます。ただし、サイト自身が本文は公開情報に基づく主観的な理解であり誤りもあり得ると注意しており、項目が追加されても過去時点の役割や状況を述べた本文が残ることがあります。最近の状態を網羅する資料とはみなさず、リンク先の一次資料と対象版のソースへ進んでください。

ツールの現在の変更を人から探す場合は、まず対象パッケージを [Go Code Owners](https://dev.golang.org/owners) で検索し、owner、Gerrit の変更履歴、関連 Issue をたどります。個人の GitHub プロフィールや owners 一覧は探索の手掛かりであって、現在の担当や完了状態の根拠ではありません。直近の Issue・CL・レビューとバージョン付きソースで確認してください。

| 調べたい課題 | go.dev から先に確認すること | research!rsc の目次で探す題名 → 次に探す題名・語 |
| --- | --- | --- |
| `cmd/go` の複数ファイル・script 形式のテストがどう組み立てられているのか | [Go Command](https://go.dev/cmd/go/) からヘルプ、変更履歴、Go 本体の `cmd/go` ソースをたどり、現在のテスト実装と採用時の根拠を確認する | `Go Testing By Example` → `script`、`testdata` → [記事](https://research.swtch.com/testing)。2023 年の記事は後年の設計観点として読み、採用時の根拠にはしない |
| CPU profile と execution trace が何を観測し、どちらを選ぶべきか | [Diagnostics](https://go.dev/doc/diagnostics) で現在の各プロファイルと trace の用途を確認する | `User-Level CPU Profiler` → `sampling`、`stack` → [記事](https://research.swtch.com/pprof) |
| コンパイラ・ランタイム・標準ライブラリの変更で失敗する箇所を、巨大なプログラムから絞り込みたい | [LoopvarExperiment](https://go.dev/wiki/LoopvarExperiment) または [Timer Channel Changes](https://go.dev/wiki/Go123Timer) の具体例から [`golang.org/x/tools/cmd/bisect`](https://pkg.go.dev/golang.org/x/tools/cmd/bisect) の現在の契約へ進む | `Hash-Based Bisect Debugging` → `compiler`、`runtime` → [記事](https://research.swtch.com/bisect) |
| Go の telemetry が何をローカル保存し、何をどの条件で upload するのか | [Go Telemetry](https://go.dev/doc/telemetry) で現在の mode、counter、report、upload の仕様を確認する | `Transparent Telemetry` → `The Design of Transparent Telemetry`、`Opting In to Transparent Telemetry` → [シリーズ目次](https://research.swtch.com/telemetry) |
| 複数の依存要求から、`go` コマンドがどのモジュール版を選ぶのか | [Go Modules Reference](https://go.dev/ref/mod) で `Minimal version selection`、`Module graph pruning`、`go.sum` を確認する | `Go & Versioning` → `Minimal Version Selection`、`Reproducible, Verifiable, Verified Builds` → [シリーズ目次](https://research.swtch.com/vgo) |
| 依存パッケージの新しい版が公開されたとき、自動更新と明示的な更新のリスクを比べたい | [How Go Mitigates Supply Chain Attacks](https://go.dev/blog/supply-chain) で現在の Go modules の更新・検証方法を確認する | `Colors Attack` → `dependency`、`latest` → [記事](https://research.swtch.com/npm-colors) |


-----------------------


[Scenario index](../SCENARIOS_en.md) | [Workshop guide](../README.md) | [Team guide](../TEAM_GUIDE_en.md)

# How to Explore 03-cmd-tools (Reverse Search Guide)

In this category, we investigate what the `go` command and its subcommands and tools (`go build`, `go run`, `go test`, `go vet`, `go generate`, `go doc`, `go tool trace`, `go tool pprof`, etc.) actually do by consulting official documentation and the cmd/go source code. When you get stuck, come back here first.

## 1. Learning What a Subcommand Does (Reading Official Help)

1. When researching on the web, start by opening [Go Command](https://go.dev/cmd/go/). This is the entry point to the official documentation for the entire `go` command.
2. From your terminal, run `go help <subcommand>` (examples: `go help run`, `go help generate`, `go help vet`). For subcommands included with `go` itself, this gives you an overview and major flags in one place.
3. To read the same content on the web, open [`pkg.go.dev/cmd/go`](https://pkg.go.dev/cmd/go) and press the `f` key to search for a subcommand name, then jump directly to its section (URLs ending with `#hdr-...`).
4. For tools provided as `go tool <subcommand>` (`vet`, `pprof`, `trace`, `cover`, etc.), use `go tool <subcommand> -h` or `go tool <subcommand> help` to check usage.

## 2. Learning Detailed Specifications and Flags

- Each individual command or tool has its own pkg.go.dev page: `https://pkg.go.dev/cmd/<subcommand>` (examples: [cmd/vet](https://pkg.go.dev/cmd/vet), [cmd/go](https://pkg.go.dev/cmd/go), [cmd/pprof](https://pkg.go.dev/cmd/pprof)).
- The **Overview** section at the top of the page summarizes what the command does and lists major flags. It often contains more information than the `-h` CLI output.

## 3. Inspecting What Actually Runs (Using Flags to Peek Inside)

- `-n` flag: Display only the commands that would be executed internally, without actually running them (dry-run).
- `-x` flag: Run the command as usual while displaying the executed commands sequentially.
- Available flags differ by subcommand, so check the options list in `go help <subcommand name>`.

## 4. Going Deep into Implementation (Verifying with Source Code)

1. The cmd/go implementation is organized under `src/cmd/go/internal/` at https://cs.opensource.google/go/go, with separate packages for each subcommand (for example: `go run` -> `internal/run`, `go build` -> `internal/work`, `go generate` -> `internal/generate`). For standalone commands (`vet`, `pprof`, etc.), open `https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/cmd/<subcommand>/` directly or follow the source link at the end of the command page on pkg.go.dev.
2. Open the file with the same name as the subcommand (e.g., `run.go`) and look for the function registered with `Cmd*.Run` (e.g., `runRun`). This is your entry point for tracing the behavior. For standalone commands, the help text output by `go help` is usually written in `doc.go`, and reading that first gives you the big picture.
3. Use the `f` key to search the page while jumping one by one to the functions being called. Read the actual implementation of the called functions to confirm. Do not judge based only on comments or the feel of function names.
4. Align the version tag to your environment. Parts like `refs/tags/go1.26.5` in the URL should match the output of `go version`.

## 5. Learning When and Why a Feature Was Added

- If you know the target Go version, first check the release notes at `go.dev/doc/go1.<version>`. In particular, the **Tools** section ([Go 1.26](https://go.dev/doc/go1.26#tools) / [Go 1.27](https://go.dev/doc/go1.27#tools)) summarizes changes to the `go` command and tool suite.
- If the release notes include issue numbers or links to proposals, be sure to open the relevant issue and read the entire discussion and related links (Issues on `github.com/golang/go` and proposals at `go.googlesource.com/proposal`). Do not infer behavior from just the summary.

## 6. Verifying Behavior Locally

- Write sufficiently small code (just `package main` + `fmt.Println`-level) and run it with visualization flags such as `-n` (display only, no execution), `-x` (display executed commands), or `-work` to see the actual output.
- Cross-check the descriptions in documentation and source code with the execution results from your local environment.

## Researching tool design and history through research.swtch.com

Russ Cox joined the Go team in 2008 and helped build two compilers and the standard library. He later served as a technical lead for the Go project and Google's Go team. His [research!rsc index](https://research.swtch.com/) is a personal site for tracing Go design decisions and implementation history from a core developer's perspective.

It is not official Go documentation. First verify the current command contract on go.dev, then search the index by article title and follow related titles and terms within the series or article. When the article and local Go version differ, verify the historical explanation separately from the current implementation.

[Go: A Documentary](https://golang.design/history/) is an index for tracing changes in the compiler, toolchain, and `cmd/go` through public design documents, Issues, CLs, and talks. Treat it as an index rather than a complete current-status reference, and follow its primary sources.

To trace current tool changes through people, search the target package in [Go Code Owners](https://dev.golang.org/owners), then follow owners, Gerrit history, and related Issues. GitHub profiles and the owners list are investigation leads, not proof of current ownership or completion; verify with recent Issues, CLs, reviews, and versioned source.

| Question | Check first on go.dev | Search in the research!rsc index |
| --- | --- | --- |
| How are multi-file and script-style tests in `cmd/go` assembled? | Start at [Go Command](https://go.dev/cmd/go/), then trace help, history, and the Go source for `cmd/go`. | `Go Testing By Example` -> `script`, `testdata` -> [article](https://research.swtch.com/testing) |
| What do CPU profiles and execution traces observe, and when should each be used? | Read [Diagnostics](https://go.dev/doc/diagnostics) for the current roles of profiles and traces. | `User-Level CPU Profiler` -> `sampling`, `stack` -> [article](https://research.swtch.com/pprof) |
| How can a failing part of a compiler, runtime, or standard-library change be narrowed down from a large program? | Start with [LoopvarExperiment](https://go.dev/wiki/LoopvarExperiment) or [Timer Channel Changes](https://go.dev/wiki/Go123Timer), then read the current [`golang.org/x/tools/cmd/bisect`](https://pkg.go.dev/golang.org/x/tools/cmd/bisect) contract. | `Hash-Based Bisect Debugging` -> `compiler`, `runtime` -> [article](https://research.swtch.com/bisect) |
| What does Go telemetry store locally, and when does it upload? | Read [Go Telemetry](https://go.dev/doc/telemetry) for the current mode, counter, report, and upload rules. | `Transparent Telemetry` -> `The Design of Transparent Telemetry`, `Opting In to Transparent Telemetry` -> [series index](https://research.swtch.com/telemetry) |
| Which module version does `go` select from multiple dependency requirements? | Read `Minimal version selection`, `Module graph pruning`, and `go.sum` in the [Go Modules Reference](https://go.dev/ref/mod). | `Go & Versioning` -> `Minimal Version Selection`, `Reproducible, Verifiable, Verified Builds` -> [series index](https://research.swtch.com/vgo) |
| How should automatic and explicit dependency updates be evaluated against supply-chain risk? | Read [How Go Mitigates Supply Chain Attacks](https://go.dev/blog/supply-chain). | `Colors Attack` -> `dependency`, `latest` -> [article](https://research.swtch.com/npm-colors) |
