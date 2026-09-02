[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [03-cmd-tools の調べ方](../../README.md)

# go generate で何ができ、何をやるべきなのか調べよう

> **実行環境**: 手元に Go 1.21 以上が必要です。

チームで「コード生成の自動化に `go generate` を使いたいです。どういうふうにやればいいか調べよう」という話になりました。

調べてみると、`//go:generate` ディレクティブに書けるのは決まったフォーマットの「コード生成コマンド」だけだと思っていたのですが、
実際に手元で試してみると、想像以上に自由度が高いことに気づきました。

同じディレクトリに置いてある [`main.go`](./main.go) が実際に試したサンプルコードです。

```go
package main

//go:generate echo "hello from go generate"
//go:generate date "+%Y-%m-%d"
//go:generate sh -c "echo GOFILE=$GOFILE GOLINE=$GOLINE GOPACKAGE=$GOPACKAGE"
//go:generate -command say echo
//go:generate say "-command による別名も使える"

func main() {}
```

手元で `go generate main.go` を実行すると、以下のように出力されました（`go1.26.4` / macOS）。

日付の行は実行日によって変わるため、ここでは `YYYY-MM-DD（実行日）` と表記します。

```
$ go generate main.go
hello from go generate
YYYY-MM-DD（実行日）
GOFILE=main.go GOLINE=5 GOPACKAGE=main
-command による別名も使える
```

`echo` や `date` はコード生成ツールではありませんが、普通に実行されています。
この自由度はどこまであるのか、そしてその自由度はどこから来ているのか、`go generate` の実装まで調べてみましょう。

---

## 設問 1: `//go:generate` に書けるコマンドの範囲を調べよう

上のサンプルを手元で実行し、実際に出力を確認してみましょう。

さらに、コード生成という本来の目的から外れた、もっとトリッキーなコマンドを `//go:generate` に書いて試してみてください。
（例: `pwd` や `whoami`、`git log -1` 、`env` など。手元の環境を壊さない範囲で自由に試してください）

- `go generate -n main.go` と `go generate -x main.go` を実行すると、それぞれ何が表示されるでしょうか。通常実行との違いは何でしょうか。
- コマンドに渡せる `$GOFILE` や `$GOLINE` のような変数は、どこに一覧がありますか。

<details>
<summary>ヒント</summary>

- 一次情報はまず `go help generate` （手元で実行）。オンライン版は https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source 。
- `-n` は "no run"、`-x` は "execute" を意識したフラグ名になっている、他のサブコマンド（`go build -n` など）と共通の命名規則。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. `go help generate` を実行し、Usage と variables の説明を読む。
2. 手元でサンプルコードと `-n` / `-x` フラグを実際に実行して出力を比較する。

**答え**

`go help generate` には次のように書かれています。

> Generate runs commands described by directives within existing files. Those commands can run any process but the intent is to create or update Go source files.

「any process」、つまり **実行可能なものなら何でも実行できる** とドキュメントに明記されています。コード生成という「意図（intent）」はあくまで想定用途であって、実行の仕組み自体はそれに縛られていません。

`-n` は実際にはコマンドを実行せず、実行されるはずのコマンド列だけを表示します。`-x` は実際に実行しつつ、実行したコマンド列も表示します（`-v` はファイル名のみ表示で、実行コマンドは表示しません）。

```
$ go generate -x main.go
echo hello from go generate
hello from go generate
date +%Y-%m-%d
YYYY-MM-DD（実行日）
sh -c echo GOFILE=main.go GOLINE=5 GOPACKAGE=main
GOFILE=main.go GOLINE=5 GOPACKAGE=main
echo -command による別名も使える
-command による別名も使える

$ go generate -n main.go
echo hello from go generate
date +%Y-%m-%d
sh -c echo GOFILE=main.go GOLINE=5 GOPACKAGE=main
echo -command による別名も使える
```

`$GOFILE` / `$GOLINE` / `$GOPACKAGE` などの変数は `go help generate` の "Go generate sets several variables" の節に一覧があります（`$GOARCH`, `$GOOS`, `$GOFILE`, `$GOLINE`, `$GOPACKAGE`, `$GOROOT`, `$DOLLAR`, `$PATH`）。

</details>

---

## 設問 2: コマンドの実行がどこで行われているか、ソースコードで確認しよう

設問 1 で分かったとおり、`go generate` は任意のコマンドを実行できます。では、その実行処理は `go` コマンド自身のソースコードのどこに書かれているのでしょうか。

実際にコマンドを起動している箇所を特定し、どの標準パッケージの、どの関数を使って子プロセスを起動しているか調べましょう。

<details>
<summary>ヒント</summary>

- `go` コマンド本体のソースは `cmd/go` にある。サブコマンドごとに `cmd/go/internal/<サブコマンド名>` というパッケージに分かれている。
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/cmd/go/internal/generate/generate.go を開いて読んでみよう。
- 子プロセスを起動する標準パッケージは 1 つしかない。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/cmd/go/internal/generate/generate.go を開く。
2. ファイル内を `exec` などで検索し、コマンド実行部分を見つける。

**答え**

`generate.go` の末尾にある `(*Generator).exec` メソッド（[generate.go#L487-L512](https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/cmd/go/internal/generate/generate.go;l=487-512)）が実行の中心です。

```go
func (g *Generator) exec(words []string) {
	path := words[0]
	if path != "" && !strings.Contains(path, string(os.PathSeparator)) {
		gorootBinPath, err := pathcache.LookPath(filepath.Join(cfg.GOROOTbin, path))
		if err == nil {
			path = gorootBinPath
		}
	}
	cmd := exec.Command(path, words[1:]...)
	cmd.Args[0] = words[0]

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = g.dir
	cmd.Env = str.StringList(cfg.OrigEnv, g.env)
	err := cmd.Run()
	if err != nil {
		g.errorf("running %q: %s", words[0], err)
	}
}
```

使われているのは標準パッケージ `os/exec` の `exec.Command` と `cmd.Run()` です。つまり `go generate` は、ディレクティブから読み取った 1 行を単語（`words`）に分解し、それをそのまま `os/exec` で子プロセスとして起動しているだけです。特別なサンドボックスや実行できるコマンドの許可リストはありません。

`run()` メソッド（同ファイル内、`//go:generate` 行を 1 行ずつスキャンする処理）から `g.exec(words)` が呼ばれる流れになっており、`-n` のときはこの呼び出し自体をスキップし、`-x` のときは呼び出し前に `words` を標準エラーに出力しています。

また `cmd.Dir = g.dir` により、コマンドは「そのファイルが置かれているディレクトリ」で実行されることも分かります（`go help generate` にも "The generator is run in the package's source directory." と明記されています）。

</details>

---

## 設問 3: なぜ `go generate` はこのような設計になっているのか、背景を調べよう

設問 1・2 で分かったように、`go generate` はコード生成に限らず「任意のコマンドを実行できる」だけの薄い仕組みです。
なぜこのようなシンプルな設計になったのでしょうか。プロポーザルを読み、`go generate` の Goal と Non-goal を調べましょう。

<details>
<summary>ヒント</summary>

- 公式ブログ https://go.dev/blog/generate （Rob Pike, 2014年12月22日）を読む。
- プロポーザル https://go.googlesource.com/proposal/+/refs/heads/master/design/go-generate.md の "Introduction" と "Discussion"、"Non-goal" の節を読む。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. ブログ記事 https://go.dev/blog/generate を読み、`go generate` が作られた経緯を確認する。
2. プロポーザル https://go.googlesource.com/proposal/+/refs/heads/master/design/go-generate.md の各節を読む。

**答え**

**背景（Introduction）**

`go build` は Go プログラムのビルドを自動化しますが、ビルド前に必要な「前処理（preliminary processing）」まではサポートしていません。プロポーザルでは次のような例が挙げられています。

> - yacc: generating .go files from yacc grammar (.y) files
> - protobufs: generating .pb.go files from protocol buffer definition (.proto) files
> - Unicode: generating tables from UnicodeData.txt
> - HTML: embedding .html files into Go source code
> - bindata: translating binary files such as JPEGs into byte arrays in Go source

ブログでも、Go のツールはソースコードから必要なビルド情報をすべて得る設計になっているため、Yacc のような「ソースコードを生成するツール」を `go` コマンドだけでは実行する手段がなかった、という課題が述べられています。`go generate`（Go 1.4, 2014年12月）はこの課題を解決するために追加されました。

**Goal（設計方針）— "Discussion" 節より**

- **パッケージの作者が実行するものであり、利用者（client）が実行するものではない。** 作者が生成済みの `.go` ファイルをリポジトリに含め、利用者は普通に `go get` / `go build` するだけで済むようにする。
- **`go build` が `go generate` を自動実行することは絶対にない。** 明示的に実行されたときだけ動く。
- **作者はどんなジェネレータでも自由に使ってよい。** シェルスクリプトのように利用者の環境では動かないものであっても構わない（利用者が実行しないため）。

**Non-goal（やるべきではないこと）**

> It is not a goal of this proposal to build a generalized build system like the Unix make(1) utility. We deliberately avoid doing any dependency analysis. The tool does what is asked of it, nothing more.

`make` のような依存関係解析付きの汎用ビルドシステムを作ることは明確に非目標とされています。「頼まれたことだけをやる、それ以上は何もしない」という言葉どおり、`go generate` は依存関係を考慮せず、ファイルに書かれたディレクティブを単に順番に実行するだけのシンプルな仕組みとして設計されました。

設問 2 で見た `exec.Command` を直接呼ぶだけの実装は、この「シンプルさを追求する」という設計方針がそのままコードに表れたものだと言えます。

</details>

---

## こぼれ話: `go generate` に向いている処理・向いていない処理

<details>
<summary>こぼれ話</summary>

ここまでの調査から、`go generate` は「任意のコマンドを実行できる薄い仕組み」であり、依存関係の解析や実行結果の安全性は一切保証しないことが分かりました。この特性を踏まえると、向き・不向きは次のように整理できます。

**向いている処理**

- **決定的（idempotent）なコード生成。** 同じ入力から同じ出力を生成する `stringer` や protobuf の `.pb.go` 生成など。何度実行しても結果が変わらない処理は事故が起きにくい。
- **実行結果をリポジトリにコミットする前提の処理。** ブログにも "once the file is generated (and tested!) it must be checked into the source code repository to be available to clients." とあるとおり、生成物はコミットされ、利用者は生成コマンドを実行しなくてよいことが前提。
- **開発者（パッケージの作者）のローカル環境やCIでの実行に限定できる処理。** 利用者の実行環境を考慮しなくてよい。

**向いていない処理**

- **ビルドや実行時に毎回必要な処理。** `go build` からは自動実行されないため、CIやビルドパイプラインに組み込み忘れると生成物が古いまま気づかれない。
- **副作用があり、何度実行するかによって結果が変わる処理。** 外部リソースの変更、削除、ネットワーク越しの一度限りの操作など。`go generate` は失敗時のロールバックや再実行の安全性を保証しない。
- **依存関係に基づいて「変更されたファイルだけ」を再生成したい処理。** Non-goal に明記されているとおり、依存関係解析は意図的に行われていない。ファイルの変更を検知して必要な生成だけ行いたい場合は `make` や専用のタスクランナーの方が向いている。
- **利用者の環境で実行されることを前提にした処理。** ジェネレータが利用者の環境に無い可能性があるため、生成物を必ずコミットしておく必要がある。

</details>

---

## 調査の入り口

- https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source
- https://go.dev/blog/generate
- https://go.googlesource.com/proposal/+/refs/heads/master/design/go-generate.md
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/cmd/go/internal/generate/generate.go
