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

**ヘルプ文言から当たりをつける**

- `go help run` の説明文を 1 文ずつ読み、動詞に注目する。「解釈する・実行する」ではなく、別の動詞が使われていないか確認する。

**フラグを使って手を動かす**

- この [main.go](./main.go) に対して `go run -x main.go` を実行する。`-x` は「裏で実際に呼ばれた外部コマンドをそのまま表示する」フラグ。
- 出力される行を 1 行ずつ眺めて、次の 2 つに仕分けしてみる。
  - 「何かを生成している」行（コマンド名に `compile` や `link` が含まれる行、`mkdir` の行）
  - 「生成済みの何かを実行している」行（パスだけがポツンと書かれた行）
  - この仕分けができると、「作ってから動かしている」のか「その場で読みながら動かしている」のか判断できる。

**ソースコードで裏取りする**

- エントリポイント: https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go の `runRun` 関数（73 行目）。まずここを開く。
- `runRun` を上から読まず、まず末尾（170〜173 行目）だけを見る。`b.LinkAction(...)` の結果を `a1` という変数に入れ、`work.Action{... Actor: work.ActorFunc(buildRunProgram) ...}` の `Deps` に `a1` を渡している。「ビルドを表すアクション」と「実行を表すアクション」が別の変数として分かれ、後者が前者に依存する形でつながれている点に注目する。
- 依存されている側の `buildRunProgram` 関数（200〜201 行目）を読む。何を実行しているか（`a.Deps[0].BuiltTarget()`）を確認し、ソースファイルの中身を読んでいる形跡があるかどうかを見る。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. `go help run`（= https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program ）を読む。「Run compiles and runs the named main Go package.」とあり、動詞は "compiles"（コンパイルする）。インタプリタなら "interprets" や "evaluates" のような動詞になるはずで、この時点で疑わしい。
2. `go run -x main.go` を実行し、出力を「生成している行」と「実行している行」に仕分ける。
3. cmd/go のソース（ https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go ）の `runRun` 関数（73 行目）を読む。170〜173 行目で、パッケージをビルドするアクション（`b.LinkAction`、変数 `a1`）と、ビルド済みの実行ファイルを実行するアクション（`Actor: work.ActorFunc(buildRunProgram)`）を、`Deps: []*work.Action{a1}` という依存関係でつないでいることが分かる。
4. 依存先の `buildRunProgram` 関数（200〜201 行目）を読むと、`a.Deps[0].BuiltTarget()` で「ビルドアクションが作った実行ファイルのパス」を取得し、それをそのまま実行しているだけだと分かる。ソースファイルを 1 行ずつ読んでいる処理はどこにもない。

**答え**

インタプリタではありません。`go run` は「ソースコードを一時的にビルドし、できあがった実行ファイルをその場で実行する」コマンドです。
`go build` のようにカレントディレクトリに実行ファイルを残すのではなく、OS の一時ディレクトリ（後述の `$WORK`）にビルドし、実行が終わったら後片付けをします。

手元で `-x` フラグ付きで実行すると、次のように「一時ディレクトリの作成 → コンパイル → リンク → 実行」という流れが見えます（実際に `go run -x main.go` を実行した出力の抜粋）。

```
WORK=/var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-build1851907018
...
mkdir -p $WORK/b001/exe/
.../link -o $WORK/b001/exe/main2 ...
$WORK/b001/exe/main2
hello, gotan v2
```

つまり `go run` は「`go build` して一時ファイルとして出力 → その一時ファイルを実行 → 実行ファイルを削除」という 3 ステップを 1 コマンドにまとめたものです。ソースコードを 1 行ずつ解釈しているわけではなく、内部的には普通にコンパイル・リンクしてから機械語のバイナリを実行しています。

</details>

---

## 設問 2: ビルドされた実行ファイルや `$WORK` ディレクトリは、実行後どうなる？

設問 1 で見た `$WORK` という一時ディレクトリと、そこに作られた実行ファイルは、実行が終わったらパソコンのどこかに残り続けるのでしょうか？
それとも消えるのでしょうか？

<details>
<summary>ヒント</summary>

**まず手を動かして事実を確認する**

- `go run -x main.go` の出力の 1 行目 `WORK=/var/folders/.../go-buildXXXXXXXXXX` を控える。
- コマンドの実行が終わった後、そのパスに対して `ls` や `find` を実行し、実際にまだ存在するか確認する。

**ヘルプから「デフォルトの挙動」を推測する**

- `go help build` のフラグ一覧を眺め、一時ディレクトリに言及しているフラグを探す。見つけたフラグの説明文を読み、それが「デフォルトでは行われないこと」を指定するフラグなのか、「デフォルトの動作を止める」フラグなのかを見分ける。

**ソースコードで裏取りする**

- エントリポイント: https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go 。ページ内検索（`f` キー）で `WorkDir` を検索すると、関連箇所にジャンプできる。
- まず一時ディレクトリを作る側を読む: `NewBuilder` 関数（281 行目）の中で `os.MkdirTemp(cfg.Getenv("GOTMPDIR"), "go-build")`（298 行目）が呼ばれている箇所を見つける。第一引数・第二引数がそれぞれ何を意味するか（`os.MkdirTemp` のドキュメントも確認する）。
- 次に片付ける側を読む: `Builder` 型の `Close` メソッド（340 行目）を見つける。中で `-work` フラグに対応する `cfg.BuildWork` が `false` のときだけ `robustio.RemoveAll(b.WorkDir)`（352 行目）が呼ばれていることを確認する。
- `Close` がどこから呼ばれているかも遡ってみる（`run.go` の `runRun` 内で `defer func() { ... b.Close() ... }()` となっている）。「一時ディレクトリを作る処理」と「後片付けの処理」が対になっていることを、自分でソースを行き来して確認する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. `go run -x main.go` で表示された `WORK=...` のパスを、コマンド終了後に `ls` してみる → ディレクトリごと消えている。
2. `go help build` を読むと `-work` フラグの説明に「print the name of the temporary work directory and do not delete it when exiting」とあり、"delete it when exiting" が **デフォルトの挙動である**ことが読み取れる。
3. cmd/go のソース（ https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go ）を読む。`NewBuilder` 関数（281 行目）が `os.MkdirTemp(cfg.Getenv("GOTMPDIR"), "go-build")`（298 行目）で一時ディレクトリを作成している。
4. `Builder.Close()`（340 行目）を読むと、`-work` フラグ（`cfg.BuildWork`）が指定されていない限り `robustio.RemoveAll(b.WorkDir)`（352 行目）で削除している。
5. `run.go` の `runRun` 関数冒頭で `b.Close()` が `defer` されており（90〜94 行目）、`go run` の実行が終わるタイミングで必ず後片付けが走ることが分かる。

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

`b001/exe/main2` が、実際に実行された実行ファイルの実体です。

</details>

---

<details>
<summary>こぼれ話: 一時ディレクトリの場所は変えられる</summary>

`os.MkdirTemp(cfg.Getenv("GOTMPDIR"), "go-build")` の第一引数が `GOTMPDIR` 環境変数なので、
`GOTMPDIR` を設定すれば、この一時ディレクトリを作る場所を変更できます（未設定なら OS 標準の一時ディレクトリが使われます）。
`go help environment` の `GOTMPDIR` の説明でも確認できます。

</details>

---

## 調査の入り口

- https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go
