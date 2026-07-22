# `go run` の裏側を覗いてみよう

先輩から「動作確認は `go run main.go` でいいよ」と言われました。
Python や Ruby の経験がある同僚は「`go run` ってスクリプト感覚でサクッと動くね。Go ってインタプリタでも動かせるんだ」と言っています。

Go はコンパイル言語のはずですが、この理解で合っているのでしょうか。`go run` を実行したとき、実際には何が起きているのか調べてみましょう。

（手元で試す例: https://go.dev/play/p/_BqTu8xBI9h 。このディレクトリの [main.go](./main.go) も同じ内容なので、`go run` してそのまま調査に使えます）

## 設問 1: `go run` は、ソースコードをインタプリタのように解釈しながら実行している？

同僚が言うように、`go run` はソースコードを逐次解釈しながら実行する「インタプリタ」的な動きをしているのでしょうか。
それとも、何か別の方法でプログラムを動かしているのでしょうか。

<details>
<summary>ヒント</summary>

- `go help run` の説明文を 1 文ずつ読み、動詞に注目してみましょう。「解釈する・実行する」ではなく、別の動詞が使われていないでしょうか。
- この [main.go](./main.go) に対して `go run -x main.go` を実行してみましょう。`-x` は「裏で実際に呼ばれた外部コマンドをそのまま表示する」フラグです。
- 出力される行を 1 行ずつ眺めて、次の 2 種類に仕分けしてみましょう。
  - 何かを新しく作っていそうな行（コマンド名に `compile` や `link` が含まれる行、`mkdir` の行）
  - すでに出来上がった何かを実行していそうな行（パスだけがポツンと書かれた行）
  - この仕分けができると、「何かを作ってから動かしている」のか「ソースコードをその場で読みながら動かしている」のか、判断する材料になります。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. `go help run`（= https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program ）を読む。「Run compiles and runs the named main Go package.」とあり、動詞は "compiles"（コンパイルする）。インタプリタなら "interprets" や "evaluates" のような動詞になるはずで、この時点で疑わしい。
2. `go run -x main.go` を実行し、出力を「生成している行」と「実行している行」に仕分ける。

**答え**

インタプリタではありません。実際に `-x` フラグを付けて実行すると、次のような出力になります（実際に `go run -x main.go` を実行した出力の抜粋）。

```
WORK=/var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-build1851907018
...
mkdir -p $WORK/b001/exe/
.../compile ... # ソースをコンパイル
.../link -o $WORK/b001/exe/main2 ... # 実行ファイルにリンク
$WORK/b001/exe/main2 # できた実行ファイルを実行
hello, gotan v2
```

`$WORK` という一時ディレクトリを作り、その中でコンパイル・リンクして実行ファイルを組み立て、最後にその実行ファイルを実行しているだけです。
`main.go` の中身を 1 行ずつ読みながら評価するような処理はどこにもなく、`go build` して一時ファイルとして出力 → その一時ファイルを実行、という 2 段階の処理を 1 コマンドにまとめたものだと分かります。

</details>

---

## 設問 2: ビルドされた実行ファイルや `$WORK` ディレクトリは、実行後どうなる？

設問 1 で見た `$WORK` という一時ディレクトリと、そこに作られた実行ファイルは、実行が終わったらパソコンのどこかに残り続けるのでしょうか？
それとも消えるのでしょうか？

<details>
<summary>ヒント</summary>

- `go run -x main.go` の出力の 1 行目 `WORK=/var/folders/.../go-buildXXXXXXXXXX` を控えておきましょう。
- コマンドの実行が終わった後、そのパスに対して `ls` や `find` を実行し、実際にまだ存在するか確認してみましょう。
- `go help build` のフラグ一覧を眺め、一時ディレクトリに言及しているフラグを探してみましょう。見つけたフラグの説明文を読み、それが「デフォルトでは行われないことを追加で行う」フラグなのか、「デフォルトの動作を止める」フラグなのかを見分けてみましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. `go run -x main.go` で表示された `WORK=...` のパスを、コマンド終了後に `ls` してみる → ディレクトリごと消えている。
2. `go help build` を読むと `-work` フラグの説明に「print the name of the temporary work directory and do not delete it when exiting」とあり、"delete it when exiting" が **デフォルトの挙動である**ことが読み取れる。

**答え**

`$WORK` ディレクトリは実行が終わると自動的に削除されます。`go run` はこのディレクトリの中でコンパイル・リンクを行い、できた実行ファイルをすぐに実行しますが、後片付けとしてディレクトリごと消してしまうため、通常は存在に気づきません。

`-work` フラグを付けて実行すると削除されずに残るので、実際に確認できます。

```
$ go run -work main2.go
WORK=/var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-build1851907018
hello, gotan v2
$ find /var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-build1851907018 -maxdepth 3
.../go-build1851907018
.../go-build1851907018/b001
.../go-build1851907018/b001/importcfg
.../go-build1851907018/b001/importcfg.link
.../go-build1851907018/b001/exe
.../go-build1851907018/b001/_pkg_.a
.../go-build1851907018/b001/exe/main2
```

`b001/exe/main2` が、実際に実行された実行ファイルの実体です。`-work` フラグを付けなければ、この一時ディレクトリごと自動的に削除されます。

</details>

---

## 設問 3: 実際に `go run` を実装しているソースコードを読み解いてみよう

`-x` の出力から「ビルドしてから実行し、後片付けする」という大まかな流れは分かりました。
では実際に、この流れは `go` コマンド自身のソースコードのどこに書かれているのでしょうか。`cmd/go` のソースを実際に読んで確認してみましょう。

<details>
<summary>ヒント 1</summary>

`go` コマンド（`cmd/go`）も Go で書かれた 1 つのプログラムです。読み方の基本は普段のコードリーディングと同じなので、初めての場合は次の手順で少しずつ読み進めてみましょう。

1. **ファイルを開く**: https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go を開きます。`cmd/go` の中では、サブコマンドごとに `internal/<サブコマンド名>` というパッケージが分かれていて、`go run` の処理はこの `internal/run` パッケージにまとまっています。
   - URL の `refs/tags/go1.26.4` の部分は、手元の `go version` の結果とバージョンを揃えるためのものです。バージョンが違うと実装が変わっている可能性があります。
2. **いきなり全部読まない**: 最初から 1 行ずつ読もうとせず、まずはファイル内の `func` と書かれた行だけを目で追って、どんな関数があるかをざっと把握しましょう。`runRun` という、いかにも中心になりそうな関数が見つかるはずです（73 行目）。
3. **どこから呼ばれる関数か確認する**: `runRun` がいつ呼ばれるかを確認するため、`init()` 関数（65〜71 行目）を見てみましょう。`CmdRun.Run = runRun` という 1 行があります。これは「`go run` が実行されたときに、実際の処理として `runRun` 関数が呼ばれる」という意味です。ここが、これから読み進める出発点になります。

</details>

<details>
<summary>ヒント 2</summary>

`runRun` 関数（73〜174 行目）は長いので、上から全部を理解しようとせず、まず空行で区切られたブロックごとに「何をしていそうか」を大づかみしてみましょう。

- 74〜86 行目: `moduleLoaderState` まわりの分岐。何かのモードを判定している？
- 88〜94 行目: `work.BuildInit` と `work.NewBuilder`、それに続く `defer func() { ... b.Close() ... }()`。設問 2 で調べた「後片付け」に関係しそうな行はどこでしょうか。
- 96〜140 行目: `for` ループや `if`/`else if`/`else` が並ぶ、少し長いブロック。`args`（コマンドライン引数）を扱っていそうです。
- 141〜168 行目: `p.Internal...` のように、`p` というパッケージらしき変数を操作している。
- 170〜173 行目: 設問 1 で見た `LinkAction` と `buildRunProgram` が出てくる、最後の 4 行。

それぞれのブロックの直前・直後にあるコメントや、変数名・関数名から「大まかに何のための処理か」を推測してみましょう。すべての行の意味が分からなくても構いません。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. cmd/go のソース（ https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go ）を開く。
2. `init()` 関数（65〜71 行目）を読み、`go run` 実行時に実際の処理を担う関数（`runRun`）がどう登録されているかを確認する。
3. `runRun` 関数（73〜174 行目）を、空行で区切られたブロックごとに読み進める。

**答え**

**`init()` の処理（65〜71 行目）**

```go
func init() {
	CmdRun.Run = runRun // break init loop

	work.AddBuildFlags(CmdRun, work.DefaultBuildFlags)
	work.AddCoverFlags(CmdRun, nil)
	CmdRun.Flag.Var((*base.StringsFlag)(&work.ExecCmd), "exec", "")
}
```

- 1 行目の `CmdRun.Run = runRun` は、22〜63 行目で定義されている `CmdRun`（`go run` コマンドそのものを表す変数）の `Run` フィールドに、実際の処理を行う `runRun` 関数を後から代入しています。`var CmdRun = &base.Command{..., Run: runRun, ...}` のように変数定義の中で直接代入せず、わざわざ `init()` の中で後から結び付けているのは、`// break init loop`（初期化ループを断ち切る）というコメントの通り、何らかの初期化順序の問題を避けるためです。ただし、なぜここで問題になるのかは `run.go` 単体を読むだけでは断定できませんでした。気になる場合は、Go 言語仕様の [Package initialization](https://go.dev/ref/spec#Package_initialization) で `var` の初期化順序のルールを確認したうえで、`git log -p` や https://go-review.googlesource.com でこの行の変更履歴を辿ってみると、実際にどんな問題が起きていたのか調べられます。
- 2〜4 行目は `-race` や `-work` のような、`go build` と共通のビルドフラグ・カバレッジフラグ・`-exec` フラグを `go run` にも登録する処理です。「`go run` にも `go build` と同じフラグが使えるのはなぜか」を確かめたければ、`work.AddBuildFlags` の実装（同じ `cmd/go/internal/work` パッケージ内）を追いかけてみましょう。

**`runRun` のブロックごとの処理内容（73〜174 行目）**

| 行番号 | 処理内容 |
| --- | --- |
| 74〜86 | `go run cmd@version` のようにバージョン付きで指定された場合は、カレントディレクトリの `go.mod` を無視してモジュールを取得するモードに切り替える（`shouldUseOutsideModuleMode` での判定）。それ以外は通常どおり、カレントディレクトリの `go.mod`/ワークスペースを使う。 |
| 88〜94 | `work.BuildInit` でビルド設定を初期化したうえで、`work.NewBuilder("", ...)` を呼び、一時ディレクトリ（`$WORK`）を持つ `Builder` を作成する。直後の `defer func() { ... b.Close() ... }()` が、設問 2 で確認した「後片付け」の予約にあたる。`defer` は Go の構文で「この関数（`runRun`）の処理が終わるときに、必ず実行する」という意味なので、`go run` が実行されるたびに、最後に必ずこの後片付けが呼ばれることになる。 |
| 96〜140 | コマンドライン引数（`args`）を解析し、「`.go` ファイルの並び」なのか「`import path` などのパッケージ指定」なのかを判定して、対象のパッケージ（`p`）をロードする。 |
| 141〜168 | ロードしたパッケージのエラーチェック、カバレッジビルドの準備、デバッグ情報を省く設定（`OmitDebug = true`）、実行ファイル名の決定などを行う。 |
| 170〜173 | 設問 1 で確認した本題の 4 行。`b.LinkAction(...)` でパッケージをビルドするアクション（`a1`）を作り、`buildRunProgram` を実行本体とする実行アクション（`a`）を `Deps: []*work.Action{a1}` で `a1` に依存させたうえで、`b.Do(ctx, a)` でこの依存関係ごと実行する。「`a1`（ビルド）が終わってから `a`（実行）が行われる」という、2 段階の処理として書かれている。 |

つまり `runRun` は、①モードの判定 → ②一時ディレクトリの準備と後片付けの予約 → ③対象パッケージの特定 → ④ビルド前の下ごしらえ → ⑤ビルドと実行、という流れの関数で、設問 1・2 で観察した「ビルドしてから実行し、最後に後片付けする」という挙動が、そのままコードの構造として表れています。

より深く追いたい場合は、`buildRunProgram` 関数（198〜201 行目）で `a.Deps[0].BuiltTarget()`（直前のビルドアクションが作った実行ファイルのパス）を実行していること、`Builder.Close()`（ https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go の 340 行目）が `-work` フラグ（`cfg.BuildWork`）が指定されていない限り一時ディレクトリを `RemoveAll` していることも合わせて確認してみましょう。

</details>

---

<details>
<summary>こぼれ話: 一時ディレクトリの場所は変えられる</summary>

`Builder.Close()` が削除する `$WORK` は、`os.MkdirTemp(cfg.Getenv("GOTMPDIR"), "go-build")` という呼び出しで作られています。
第一引数が `GOTMPDIR` 環境変数なので、`GOTMPDIR` を設定すれば、この一時ディレクトリを作る場所を変更できます（未設定なら OS 標準の一時ディレクトリが使われます）。
`go help environment` の `GOTMPDIR` の説明でも確認できます。

</details>

---

## 調査の入り口

- https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go
