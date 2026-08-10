# 03-cmd-tools の調べ方（逆引き手順）

このカテゴリでは、`go` コマンドとその配下のサブコマンド・ツール（`go build`, `go run`, `go test`, `go vet`, `go generate`, `go doc`、`go tool trace`、`go tool pprof` など）が実際に何をしているかを、
公式ドキュメントと cmd/go のソースコードから調べます。困ったときは、まずここに戻ってきてください。

## 1. サブコマンドが何をするコマンドか知りたい（公式ヘルプを読む）

1. 手元で `go help <サブコマンド>` を実行する（例: `go help run`, `go help generate`, `go help vet`）。`go` 本体に同梱されているサブコマンドは、これだけで概要と主要フラグが読める。
2. 同じ内容を Web で読むなら [`pkg.go.dev/cmd/go`](https://pkg.go.dev/cmd/go) のページで `f` キーを押し、サブコマンド名を検索してその節（`#hdr-...` で終わる URL）に直接ジャンプできる。
3. `go tool <サブコマンド>` として提供されるツール（`vet`、`pprof`、`trace`、`cover` など）は `go tool <サブコマンド> -h` や `go tool <サブコマンド> help` で使い方を確認できる。

## 2. コマンドの詳しい仕様やフラグを知りたい

- 単体のコマンド／ツールにはそれぞれ専用の pkg.go.dev ページがある: `https://pkg.go.dev/cmd/<サブコマンド>`（例: [cmd/vet](https://pkg.go.dev/cmd/vet), [cmd/go](https://pkg.go.dev/cmd/go), [cmd/pprof](https://pkg.go.dev/cmd/pprof)）。
- ページ冒頭の **Overview** に、コマンドが何をするかと主要なフラグ一覧がまとまっている。CLI の `-h` 出力より情報量が多いことも多い。

## 3. 実際に何が実行されているか、フラグで覗く

- `-n` フラグ: 実行はせず、内部で呼ばれるコマンド列だけを表示する（dry-run）。
- `-x` フラグ: 実際に実行しながら、呼ばれたコマンド列を逐次表示する。
- 使えるフラグはサブコマンドによって違うので、`go help <サブコマンド名>` のオプション一覧で確認する。

## 4. 実装まで踏み込みたい（ソースコードで裏取りする）

1. cmd/go の実装は https://cs.opensource.google/go/go の `src/cmd/go/internal/` 以下に、サブコマンドごとにパッケージが分かれている（例: `go run` → `internal/run`、`go build` → `internal/work`、`go generate` → `internal/generate`）。単体コマンド（`vet`, `pprof` など）は `https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/cmd/<サブコマンド>/` を直接開くか、pkg.go.dev のコマンドページ末尾のソースリンクをたどる。
2. サブコマンド名と同じ名前のファイル（`run.go` など）を開き、`Cmd*.Run` に登録されている関数（`runRun` など）を探す。ここが挙動を追いかけるエントリポイント。単体コマンドの場合は `doc.go` に `go help` で出るヘルプ本文が書かれていることが多く、まずここを読むと全体像がつかめる。
3. `f` キーでページ内検索しながら、その関数が呼んでいる関数へ 1 つずつジャンプする。呼び出し先の実装まで実際に読んで確認する。コメントや関数名の雰囲気だけで判断しない。
4. バージョンタグを手元の環境に合わせる。URL 中の `refs/tags/go1.26.5` のような部分は、`go version` の結果と揃える。

## 5. いつ・なぜ追加された機能か知りたい

- 対象の Go バージョンが分かっている場合、まず `go.dev/doc/go1.<version>` のリリースノートを見る。特に **Tools** セクション（[Go 1.26](https://go.dev/doc/go1.26#tools) / [Go 1.27](https://go.dev/doc/go1.27#tools)）に、`go` コマンドとツール群の変更点がまとまっている。
- リリースノートに Issue 番号やプロポーザルへのリンクがあれば、必ず該当 Issue を開き、議論と関連リンクまで含めて読む（`github.com/golang/go` の Issue、`go.googlesource.com/proposal`）。要約だけから挙動を推測しない。

## 6. 挙動を手元で確かめる

- 十分小さいコード（`package main` + `fmt.Println` 程度）を書き、`-n`（実行せず表示のみ）や `-x`（実行したコマンドを表示）、`-work` などの可視化フラグを付けて実際に出力を見る。
- ドキュメントやソースコードの記述と、手元での実行結果を突き合わせて確認する。
