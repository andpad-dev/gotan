# `go vet` の検査内容を調べよう

先輩が出した PR で、CI のジョブに `go vet ./...` が追加されているのを見かけました。どんなものか調べてみましょう。

たとえば、次のような何気ないコードに `go vet` をかけてみます。

```go
package main

import "fmt"

func main() {
	name := "gopher"
	fmt.Printf("hello, %d\n", name)
}
```

（Go Playground で動かす: https://go.dev/play/p/GOVSRd_FWPU ）

`go build` や `go run` は普通に通り、実行結果は次のとおりです。

```
hello, %!d(string=gopher)
```

ところが `go vet ./...` を走らせると、次のような指摘が出ます。

```
./main.go:7:21: fmt.Printf format %d has arg name of wrong type string
```

`go vet` はビルドやテストとは別に、いったい何を見ているのでしょうか。

## 設問 1: `go vet` は何を報告するコマンド？ コンパイルとの違いは？

`go vet` の役割と、コンパイラでは検出できない領域について、一次情報から確認してみましょう。

<details>
<summary>ヒント</summary>

- 手元で `go help vet` を打つとまとまった説明が読める
- 同じ内容は [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) の Overview にも書かれている
- `go` コマンド側から見た説明は [`pkg.go.dev/cmd/go`](https://pkg.go.dev/cmd/go) を `f` キーで「vet」検索

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. 手元で `go help vet` を実行し、CLI 版のヘルプ全文を読む。
2. 同じ内容を [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) の Overview で確認する。「Analyzers may use heuristics that do not guarantee all reports are genuine problems, but can find mistakes not caught by the compiler.」というくだりが今回の核心。
3. `go` コマンド側の位置づけは [`pkg.go.dev/cmd/go` の「Report likely mistakes in packages」節](https://pkg.go.dev/cmd/go#hdr-Report_likely_mistakes_in_packages) にある（`go help vet` と同じ文面）。

**答え**

- `go vet` は **Go プログラムの静的解析ツール**で、「怪しい構文（suspicious constructs）」や「改善の余地がある箇所（opportunities for improvement）」を報告する。
- 特徴は「**コンパイラでは検出できない間違い**をヒューリスティックで見つける」こと。裏を返せば、`go vet` の指摘は必ずしも真のバグとは限らないため、報告を読んで判断する前提のツール。
- 冒頭で試した `fmt.Printf("hello, %d\n", name)` は、書式（`%d`）と引数の型（`string`）が食い違っており、コンパイルは通るが実行時に `%!d(string=gopher)` という壊れた出力になる。この種の「型システムをすり抜ける論理エラー」を捕まえるのが `go vet` の代表的な仕事。

</details>

---

## 設問 2: `go vet` にはどんなチェッカーがあり、`printf` の詳細はどこで読める？

さきほどの `printf` 指摘は、`go vet` に組み込まれた多数のアナライザ（analyzer）のひとつです。全体像とアナライザ単位の詳細ドキュメントの在り処を調べましょう。

<details>
<summary>ヒント</summary>

- [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) の Overview の下に、登録済みアナライザの一覧表がある
- 手元で `go tool vet help` を打つと同じ一覧が、`go tool vet help printf` を打つと個別アナライザの詳細が読める
- `printf` チェッカーの本体は Go 本体ではなく [`golang.org/x/tools`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf) にある

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) の Overview を下にスクロールし、Registered analyzers の一覧表を読む（`appends`、`printf`、`slog`、`stdversion`、`waitgroup` など約 30 個）。
2. 手元で `go tool vet help` を実行して同じ一覧を確認する。`go tool vet help printf` のようにアナライザ名を続けると、そのアナライザ専用のドキュメントとフラグが表示される。
3. `printf` アナライザの実装ドキュメントは [`pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf) にある。`fmt.Printf` / `fmt.Sprintf` などの書式文字列と引数の整合性を検査する、ということが仕様レベルで書かれている。

**答え**

- `go vet` は単一のチェッカーではなく、**個別のアナライザの集合体**。`pkg.go.dev/cmd/vet` と `go tool vet help` の Registered analyzers 一覧が公式カタログ。
- 各アナライザは [`golang.org/x/tools/go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis) フレームワーク上で書かれた独立したモジュールで、実装や詳細ドキュメントはたいてい `passes/<アナライザ名>` パッケージにある。
- 今回の指摘を出したのは `printf` アナライザ。カバー範囲や、追加で検査させたい関数名を指定する `-printf.funcs` フラグの詳細は、`go tool vet help printf` と `pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf` の両方に載っている。

</details>

---

<details>
<summary>こぼれ話: `go test` はこっそり `go vet` を走らせている</summary>

`go test` は、テスト実行の前に **キュレーションされたサブセットの `go vet` を自動で走らせます**。詳しくは [`pkg.go.dev/cmd/go` の「Test packages」節](https://pkg.go.dev/cmd/go#hdr-Test_packages) に書かれています（`atomic`、`bool`、`buildtags`、`directive`、`errorsas`、`ifaceassert`、`nilfunc`、`printf`、`stringintconv`、`tests` が既定で走る、という一覧付き）。

実際、冒頭のコードを `go test ./...` にかけると次のように「ビルドがそもそも失敗した」扱いになります。

```
# example.com/vet-demo
./main.go:7:21: fmt.Printf format %d has arg name of wrong type string
FAIL    example.com/vet-demo [build failed]
```

さらに [Go 1.27 のリリースノート](https://go.dev/doc/go1.27#go-test) には次の 1 行が入っています。

> `go test` now invokes the `stdversion` vet check by default.

つまり Go 1.27 からは、`go.mod` の `go` バージョンや `//go:build` タグで許容されているより **新しすぎる標準ライブラリのシンボル** を使っていないかも、テスト時に自動チェックされるようになります。「`go vet` を CI に足そう」という PR が続くのは、こうした流れが背景にあります。

</details>

---

## 調査の入り口

- https://pkg.go.dev/cmd/vet
- https://pkg.go.dev/cmd/go#hdr-Report_likely_mistakes_in_packages
- https://pkg.go.dev/cmd/go#hdr-Test_packages
- https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf
- https://go.dev/doc/go1.27#go-test
