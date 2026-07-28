# 03-cmd-tools の調べ方（逆引き手順）

このカテゴリでは、`go` コマンドとその配下のサブコマンド・ツール（`go vet`、`go doc`、`go tool trace`、`go tool pprof` など）を題材にします。
困ったときは、まずここに戻ってきてください。

## サブコマンドが何をするコマンドか知りたい

1. 手元で `go help <サブコマンド>` を実行する（例: `go help vet`）。`go` 本体に同梱されているサブコマンドは、これだけで概要と主要フラグが読める。
2. 同じ内容を Web で読むなら `https://pkg.go.dev/cmd/go` のページで `f` キーを押し、サブコマンド名を検索してその節に飛ぶ。
3. `go tool <サブコマンド>` として提供されるツール（`vet`、`pprof`、`trace`、`cover` など）は `go tool <サブコマンド> -h` や `go tool <サブコマンド> help` で使い方を確認できる。

## コマンドの詳しい仕様やフラグを知りたい

- 単体のコマンド／ツールにはそれぞれ専用の pkg.go.dev ページがある: `https://pkg.go.dev/cmd/<サブコマンド>`（例: [cmd/vet](https://pkg.go.dev/cmd/vet), [cmd/go](https://pkg.go.dev/cmd/go), [cmd/pprof](https://pkg.go.dev/cmd/pprof)）。
- ページ冒頭の **Overview** に、コマンドが何をするかと主要なフラグ一覧がまとまっている。CLI の `-h` 出力より情報量が多いことも多い。

## 実装まで踏み込みたい

- pkg.go.dev のコマンドページ末尾からソースへのリンクをたどるか、直接 `https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/cmd/<サブコマンド>/` を開く。
- `doc.go` に `go help` で出るヘルプ本文が書かれていることが多い。まずここを読むと全体像がつかめる。

## 最新のリリースで何が変わったか知りたい

- Go のリリースノートの **Tools** セクション（[Go 1.26](https://go.dev/doc/go1.26#tools) / [Go 1.27](https://go.dev/doc/go1.27#tools)）に、`go` コマンドとツール群の変更点がまとまっている。
- リリースノートに Issue 番号が書かれていたら、必ず該当 Issue を開き、議論と関連リンクまで含めて読む。要約だけから挙動を推測しない。
