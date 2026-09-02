[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [03-cmd-tools の調べ方](../../README.md)

# `go vet` の検査内容を調べよう

> **実行環境**: 手元に Go 1.27 以上が必要です。

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

同じコードを [main.go](./main.go) として、`go 1.27` の [go.mod](./go.mod) と一緒に置いてあります。このディレクトリで `go version`、`go build ./...`、`go run .`、`go vet ./...` の順に実行してください。

`go build` や `go run` は普通に通り、実行結果は次のとおりです。

```
hello, %!d(string=gopher)
```

ところが `go vet ./...` を走らせると、次のような指摘が出ます。

```
main.go:7:21: fmt.Printf format %d has arg name of wrong type string
```

`go vet` はビルドやテストとは別に、いったい何を見ているのでしょうか。

## 設問 1: `go vet` は何を報告するコマンド？ コンパイルとの違いは？

`go vet` の役割と、コンパイラでは検出できない領域について、一次情報から確認してみましょう。

<details>
<summary>ヒント</summary>

- 手元で `go help vet` を読み、そこから案内される次のコマンドを実行してみましょう。
- `go tool vet help` には、検査対象と誤検知の可能性がまとまっています。
- [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) の Overview にも関連する説明があります。CLI と一字一句同じとは限らないので、それぞれの文を確認しましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、`go vet` の公式ドキュメントを探す。
2. Go 1.27.0 で `go help vet` を実行する。`go vet` が既定の `cmd/vet` を呼ぶことと、チェッカーの説明には `go tool vet help` を使うことを確認する。
3. `go tool vet help` を実行する。「vet is a tool for static analysis of Go programs.」に続き、アナライザはヒューリスティックを使うため全報告が本物の問題とは限らない一方、コンパイラが見つけない誤りを発見できると説明されている。
4. [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) の Overview も読む。こちらは「Vet uses heuristics ...」という別の文面で、同じ性質を説明している。
5. `go` コマンド内での位置づけは [`pkg.go.dev/cmd/go` の「Report likely mistakes in packages」節](https://pkg.go.dev/cmd/go#hdr-Report_likely_mistakes_in_packages) で確認する。CLI とWebの文面を同一視せず、役割と実装ツールの関係を照合する。

**答え**

- `go vet` は既定で静的解析ツール `cmd/vet` を実行し、怪しい構文（suspicious constructs）や改善候補（opportunities for improvement）の診断を報告する。
- アナライザは **コンパイラでは検出できない間違い**をヒューリスティックで見つける。`go tool vet help` が明記するとおり、すべての報告が本物の問題である保証はないため、根拠を読んで判断する。
- 冒頭で試した `fmt.Printf("hello, %d\n", name)` は、書式（`%d`）と引数の型（`string`）が食い違っており、コンパイルは通るが実行時に `%!d(string=gopher)` という壊れた出力になる。この種の「型システムをすり抜ける論理エラー」を捕まえるのが `go vet` の代表的な仕事。

</details>

---

## 設問 2: `go vet` にはどんなチェッカーがあり、`printf` の詳細はどこで読める？

さきほどの `printf` 指摘は、`go vet` に組み込まれた多数のアナライザ（analyzer）のひとつです。全体像とアナライザ単位の詳細ドキュメントの在り処を調べましょう。

<details>
<summary>ヒント</summary>

- 手元で `go tool vet help` を実行すると、`Registered analyzers:` の直後に現在のツールチェーンの一覧が出る
- [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) にもアナライザ一覧はあるが、`Registered analyzers` という見出しではない
- `go tool vet help printf` を実行すると個別アナライザの詳細が読める
- `printf` アナライザの公式パッケージ文書は [`golang.org/x/tools`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf) で読める

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) を入口に、`vet` の公式ドキュメントを探す。
2. `go tool vet help` を実行し、`Registered analyzers:` の一覧を読む。Go 1.27.0 では 35 個だが、数や名前は版によって変わるので自分の出力を記録する。
3. [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) の Overview にある「To list the available checks...」以降も読む。CLI と同じ見出しではないため、一覧を探す操作を混同しない。
4. `go tool vet help printf` を実行し、そのアナライザ専用のドキュメントとフラグを確認する。
5. `printf` アナライザのパッケージ文書は [`pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf) にある。`fmt.Printf` / `fmt.Sprintf` などの書式文字列と引数の整合性を検査すると説明されている。

**答え**

- `go vet` は単一のチェッカーではなく、**個別のアナライザの集合体**。実行中の版で正確な一覧を得るには、`go tool vet help` の `Registered analyzers:` を確認する。
- 各アナライザは [`golang.org/x/tools/go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis) フレームワーク上で書かれた独立したモジュールで、実装や詳細ドキュメントはたいてい `passes/<アナライザ名>` パッケージにある。
- 今回の指摘を出したのは `printf` アナライザ。カバー範囲や、追加で検査させたい関数名を指定する `-printf.funcs` フラグの詳細は、`go tool vet help printf` と `pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf` の両方に載っている。

</details>

---

<details>
<summary>こぼれ話: `go test` はこっそり `go vet` を走らせている</summary>

`go test` は、テスト実行の前に **キュレーションされたサブセットの `go vet` を自動で走らせます**。詳しくは [`pkg.go.dev/cmd/go` の「Test packages」節](https://pkg.go.dev/cmd/go#hdr-Test_packages) に書かれています（`atomic`、`bool`、`buildtags`、`directive`、`errorsas`、`ifaceassert`、`nilfunc`、`printf`、`stringintconv`、`tests` が既定で走る、という一覧付き）。

実際、冒頭のコードを `go test ./...` にかけると次のように「ビルドがそもそも失敗した」扱いになります。

    # example.com/vet-demo
    ./main.go:7:21: fmt.Printf format %d has arg name of wrong type string
    FAIL    example.com/vet-demo [build failed]

さらに [Go 1.27 のリリースノート](https://go.dev/doc/go1.27#go-test) には次の 1 行が入っています。

> `go test` now invokes the `stdversion` vet check by default.

つまり Go 1.27 からは、`go.mod` の `go` バージョンや `//go:build` タグで許容されているより **新しすぎる標準ライブラリのシンボル** を使っていないかも、テスト時に自動チェックされるようになります。「`go vet` を CI に足そう」という PR が続くのは、こうした流れが背景にあります。

</details>

---

## 調査の入り口

- https://go.dev/doc/
- https://pkg.go.dev/cmd/vet
- https://pkg.go.dev/cmd/go#hdr-Report_likely_mistakes_in_packages
- https://pkg.go.dev/cmd/go#hdr-Test_packages
- https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf
- https://go.dev/doc/go1.27#go-test
