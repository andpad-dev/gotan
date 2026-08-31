[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [03-cmd-tools の調べ方](../../README.md)

# 1 枚のテキストがテストになるまで: cmd/go の script tests

あなたのチームでは、CLI の結合テストごとに一時ディレクトリを作り、複数の入力ファイルを書き出し、コマンドを実行して標準出力・標準エラーを検査しています。テストの準備コードが本題より長く、レビューで「何を試したいのか」が見えにくくなってきました。

同僚が [Go command](https://go.dev/cmd/go/) の Source Files から、Go 1.27.0 の [`run_hello.txt`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/testdata/script/run_hello.txt) を見つけました。1 枚のテキストに、実行手順、期待値、実行時に必要な Go ファイルが同居しています。

```text
env GO111MODULE=off

# hello world
go run hello.go
stderr 'hello world'

-- hello.go --
package main
func main() { println("hello world") }
```

境界線より下の `hello.go` は [Go Playground](https://go.dev/play/p/4Quk7tidxe8) でも実行できます。この 1 枚がどのようにテストへ変換されるのかを起点に、現在の実装、導入時の設計判断、自分たちのプロジェクトで再利用できる境界まで調べましょう。

---

## 設問 1: 1 枚のテキストはどう分解され、実行される？

`-- hello.go --` より上は、なぜファイルとして展開されずにコマンドとして実行されるのでしょうか。反対に、境界線より下はいつ、どこへ作られるのでしょうか。

Go 1.27.0 の `TestScript` で、`run_hello.txt` を読み込んでからテストが終了するまでの流れを追ってください。さらに、この 1 件だけを名前で選んで実行し、実測結果と結び付けて説明してください。

<details>
<summary>ヒント</summary>

- Go command のページから Source Files を開き、テスト用の `.txt` を列挙する関数を探します。
- 境界記号を読む処理が返す値と、その直後に呼ばれる処理へ順番に進みます。
- ファイル名から subtest 名を作る箇所と、一時作業ディレクトリを作る箇所を見比べます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go command](https://go.dev/cmd/go/) を起点に Source Files から `script_test.go` を探し、[Go 1.27.0 の `TestScript`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/script_test.go;l=39) が `testdata/script/*.txt` を列挙するところから読む。
2. 同じソースで `ParseFile` の import 元をたどり、[Go 1.27.0 の `internal/txtar`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/txtar/archive.go;l=5) にあるアーカイブ形式、`Archive`、`Parse` を確認する。
3. [Go Testing By Example](https://research.swtch.com/testing) の Tip 13〜15 を読み、複数ファイルを 1 件のテストデータにし、既存形式への注釈と専用 parser でテストを小さくする、という読み方を得る。
4. `script_test.go` に戻り、`NewState`、`ExtractFiles`、`scripttest.Run` の呼び出し順を追う。[Go 1.27.0 の `State.ExtractFiles`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/internal/script/state.go;l=148) と [script test の README](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/testdata/script/README) で、展開先と実行場所を裏取りする。

**答え**

これは `txtar` 形式のテキストアーカイブです。`internal/txtar.ParseFile` は、最初の `-- FILENAME --` より前を `Archive.Comment` に、各境界線の後を `Archive.Files` に分けます。したがって `run_hello.txt` では、`env` から `stderr` までが comment、`hello.go` が 1 個の file です。アーカイブの parser 自体には構文エラーがありませんが、ファイル名の展開可否や script のコマンドエラーは後段で別に検査されます。

`TestScript` の流れは次の通りです。

1. `testdata/script/*.txt` を列挙し、拡張子を除いた名前で `t.Run` を作る。
2. 各 subtest 用の一時作業ディレクトリと `State` を作る。
3. `txtar.ParseFile` で 1 枚を comment と files に分ける。
4. `ExtractFiles` で files を作業ディレクトリ内へ展開する。`cmd/go` の設定では `$WORK/gopath/src` から始まり、作業ディレクトリ外を指す名前は拒否される。
5. `Archive.Comment` だけを script engine に渡し、`go run hello.go` の後で直前の標準エラーが `hello world` に一致するか検査する。

Go 1.27.0 / macOS で、Go のソースディレクトリから対象 subtest だけを実行した結果です。他の実行と build cache を分けるため、実測では cache も一時ディレクトリへ向けました。所要時間は環境で変わります。

```console
$ cd "$(GOTOOLCHAIN=go1.27.0 go env GOROOT)/src/cmd/go"
$ GOCACHE=/private/tmp/gotan-go-script-cache GOTOOLCHAIN=go1.27.0 go test . -run='^TestScript/run_hello$' -count=1
ok  	cmd/go	1.460s
```

終了コード 0 は、展開された `hello.go` を script 内の `go run` が実行し、`stderr` の検査まで成功したことを示します。subtest 名がファイル名由来なので、巨大な script test 群からこの 1 件だけを `-run` で選べます。

</details>

---

## 設問 2: なぜ bash でも普通の Go テストでもない？

現在の runner は、system shell ではなく、登録されたコマンドと条件だけを解釈する小さな言語を使います。一方、ファイルの列挙、個別選択、並列実行には `testing` の仕組みを使っています。

この二層構造は、どんな旧方式の問題を同時に解こうとして 2018 年に導入されたのでしょうか。導入時の変更と現在の実装を区別し、2023 年の「Go Testing By Example」の Tip 13〜17 がこの設計をどう説明できるかも整理してください。

<details>
<summary>ヒント</summary>

- 現在の `script_test.go` の履歴から、テスト方式を追加した 2018 年の変更を探します。
- その変更の説明では、最初の shell script、次の Go 製 test framework、新方式の 3 段階が比較されています。
- 「書きやすさ」と「選択・移植・分離のしやすさ」を別々の軸にして比較します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go command](https://go.dev/cmd/go/) から現在の [Go 1.27.0 `script_test.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/script_test.go;l=39) を開き、Source の履歴をたどる。
2. [Go Testing By Example](https://research.swtch.com/testing) の Tip 13〜17 を読み、現在の構造を「複数ファイル」「小さな言語」「script」という観点に分ける。ただし記事は 2023 年公開なので、これを 2018 年の採用理由の証拠にはしない。
3. 2018 年の変更 [`cmd/go: add new test script facility`](https://github.com/golang/go/commit/5890e25b7ccb2d2249b2f8a02ef5dbc36047868b) と、その一次レビュー [CL 123577](https://go-review.googlesource.com/c/go/+/123577) を読み、旧方式、新方式、当時の実測を分けて記録する。
4. [Go 1.27.0 の script engine](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/internal/script/engine.go;l=5) を読み、現在も小さく、設定可能で、platform-agnostic な言語として実装されていることを確認する。

**答え**

2018 年の変更説明が示す経緯は次の通りです。

- 最初の小さな shell scripts は書きやすい反面、個別選択ができず、Windows で動かず、遅いという問題があった。
- 次の Go 製 `testgo` framework は、個別選択、Windows、並列実行、テスト間の分離を可能にした。しかし、一時ファイル作成や実行結果の検査が Go の手続きへ膨らみ、テストの意図を流し読みしにくかった。
- 新方式は、外側を `testing` の subtest として選択・並列化し、内側だけを用途専用の shell-like language と txtar で短く記述した。これにより、両方式の長所を組み合わせた。

当時の変更では、同時に移行した 15 件のテストについて `TestScript` が 5.5 秒から 2.5 秒になったと報告されています。これは 2018 年のその変更に対する実測であり、Go 1.27.0 や別環境の性能保証ではありません。

2023 年の記事の Tip 13〜17 は、現在の 1 枚を読むための整理になります。txtar で複数ファイルを束ねる（13）、既存のテキスト形式に注釈して小さな言語にする（14）、専用の parser / printer でテストを単純にする（15）、テスト品質そのものを保つ（16）、操作の列を script として読む（17）という対応です。ただし、記事が 2018 年の変更を引き起こしたという時系列ではありません。資料から言えるのは、同じ著者が後年に設計上の知見としてまとめた、ということまでです。

また、この言語は system shell ではありませんが、セキュリティ sandbox でもありません。実行可能なコマンドと条件を `Engine` に登録でき、`cmd/go` のテストでは Go コマンドなどを実際に起動します。ここでの分離とは、主に一時作業領域、環境、subtest の独立性です。

</details>

---

## 設問 3: 自分たちのテストへ、どこまで持ち帰れる？

チームは同じ形式を採用したくなりました。しかし、Go 1.27.0 の runner は `internal/txtar` と `cmd/internal/script` を import しています。

この実装をそのまま依存先にできるでしょうか。標準の公開 API、Go 本体の internal implementation、公開されている外部 module を区別し、次の二つの場合の採用案を示してください。

- 必要なのは「複数のテキストファイルを 1 枚に束ねること」だけ
- 必要なのは「ファイルの展開に加え、小さな script を実行すること」まで

最後に、冒頭の `run_hello.txt` がなぜ 1 枚で済み、なぜ個別実行でき、同じものを無条件には import できないのかを一続きで説明してください。

<details>
<summary>ヒント</summary>

- `go help importpath` の `Internal packages` を確認します。ソースが読めることと import できることは別です。
- アーカイブの解析、コマンドの実行、`testing` との接続を別々の部品として分類します。
- 2023 年の記事が自作 script tests 向けに案内する module では、README のサポート方針まで確認します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go command の `Internal packages`](https://go.dev/cmd/go/#hdr-Internal_packages) と手元の `go help importpath` を読み、`internal` を含む package の import 可能範囲を確認する。標準のテスト側は [testing の公開 API](https://go.dev/pkg/testing/) も確認する。
2. [Go 1.27.0 の `script_test.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/script_test.go;l=39) で、`internal/txtar` と `cmd/internal/script` が Go 本体内の実装依存であることを確認する。
3. [Go Testing By Example](https://research.swtch.com/testing) の Tip 13 と Tip 18 から公開候補へ進む。
4. [`golang.org/x/tools/txtar@v0.47.0`](https://pkg.go.dev/golang.org/x/tools/txtar@v0.47.0) と [tag 固定の実装](https://cs.opensource.google/go/x/tools/+/refs/tags/v0.47.0:txtar/archive.go;l=5) を読み、公開 package が担当する範囲を確認する。
5. [`rsc.io/script@v0.0.2`](https://pkg.go.dev/rsc.io/script@v0.0.2)、[tag 固定の README](https://github.com/rsc/script/blob/v0.0.2/README.md)、[`scripttest` の実装](https://github.com/rsc/script/blob/v0.0.2/scripttest/scripttest.go) を読み、提供範囲とサポート方針を分けて確認する。

**答え**

依存先は次のように分類します。

| 対象 | 境界 | 採用判断 |
| --- | --- | --- |
| `testing.T.Run`、`T.Parallel` など | 標準ライブラリの公開 API | 自分たちの runner の土台として利用できる |
| `internal/txtar`、`cmd/internal/script` | Go 本体の internal implementation | ソースは読めるが、無関係な module からは import できない。Go 1.27.0 の内部構造を自分たちの API 契約として扱わない |
| `golang.org/x/tools/txtar@v0.47.0` | Go project が公開する外部 module の API | txtar の parse / format が必要な場合に利用できる。script engine は含まないので、複数ファイルを束ねるだけならこの範囲で足りる |
| `rsc.io/script@v0.0.2` と `rsc.io/script/scripttest` | Go 標準外の公開 module | script engine とテスト用の接続を利用できる。ただし README は、試用できるよう公開した copy であり、公式サポートを約束するものではないと明記している |

したがって、複数ファイルだけなら version を固定した `golang.org/x/tools/txtar` と標準の `testing` で小さな runner を作れます。script 実行まで必要なら、`rsc.io/script` の version を固定し、自分たちが使うコマンド、対応 OS、更新時の互換性をテストしたうえで採用する案があります。どちらも Go 本体の internal packages を直接 import する案ではありません。また、外部プログラムを起動できる engine は sandbox ではないため、信頼できない script を実行する用途には使いません。

これで冒頭の違和感を回収できます。`run_hello.txt` が 1 枚で済むのは txtar が command 部と file tree を同居させるからです。1 件だけ選べるのは `TestScript` がファイル名を `testing` の subtest 名へ変えるからです。短い専用言語になったのは、shell scripts の書きやすさと Go 製 test framework の選択・移植・分離を両立するためでした。しかし、それを動かす現在の `cmd/go` 実装は internal 境界の内側です。持ち帰るときは、公開 API と version 固定した外部 module を選び、内部実装の読み解きと依存契約を混同しません。

</details>

---

<details>
<summary>こぼれ話</summary>

`txtar` は一般的な archive 形式を目指していません。Go 1.27.0 の package comment は、手で編集しやすいこと、テキストの file tree を保持できること、Git の履歴や code review で差分を読みやすいことを目標に挙げ、binary data、file mode、symbolic link などを非目標にしています。「本物の filesystem を完全に保存する形式」ではなく、「テストケースを読み書きする形式」だと分かると、1 枚の `.txt` という選択にも納得できます。

</details>

---

## 調査の入り口

1. [Go command](https://go.dev/cmd/go/)
2. [Go 1.27.0 の `run_hello.txt`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/testdata/script/run_hello.txt) と [`script_test.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/script_test.go;l=39)
3. [Go Testing By Example](https://research.swtch.com/testing)
4. [2018 年の script test 導入 commit](https://github.com/golang/go/commit/5890e25b7ccb2d2249b2f8a02ef5dbc36047868b)
5. [`golang.org/x/tools/txtar@v0.47.0`](https://pkg.go.dev/golang.org/x/tools/txtar@v0.47.0) と [`rsc.io/script@v0.0.2`](https://pkg.go.dev/rsc.io/script@v0.0.2)
