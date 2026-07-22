# `go run` の裏側を覗いてみよう

先輩から「動作確認は `go run main.go` でいいよ」と言われました。
Python や Ruby の経験がある同僚は「`go run` ってスクリプト感覚でサクッと動くね。Go ってインタプリタでも動かせるんだ」と言っています。

Go はコンパイル言語のはずですが、この理解で合っているのでしょうか。`go run` を実行したとき、実際には何が起きているのか調べてみましょう。

（手元で試す例: https://go.dev/play/p/_BqTu8xBI9h 。このディレクトリの [main.go](./main.go) も同じ内容なので、`go run` してそのまま調査に使えます）

## 設問 1: `go run` は、ソースコードをインタプリタのように解釈しながら実行している？

同僚が言うように、`go run` はソースコードを逐次解釈しながら実行する「インタプリタ」的な動きをしているのでしょうか。
それとも、何か別の方法でプログラムを動かしているのでしょうか。

<details>
<summary>ヒント 1</summary>

- `go help run` の説明文を 1 文ずつ読み、動詞に注目してみましょう。「解釈する・実行する」ではなく、別の動詞が使われていないでしょうか。
- この [main.go](./main.go) に対して `go run -x main.go` を実行してみましょう。`-x` は「裏で実際に呼ばれた外部コマンドをそのまま表示する」フラグです。
- 出力される行を 1 行ずつ眺めて、次の 2 種類に仕分けしてみましょう。
  - 何かを新しく作っていそうな行（コマンド名に `compile` や `link` が含まれる行、`mkdir` の行）
  - すでに出来上がった何かを実行していそうな行（パスだけがポツンと書かれた行）
  - この仕分けができると、「何かを作ってから動かしている」のか「ソースコードをその場で読みながら動かしている」のか、判断する材料になります。

</details>

<details>
<summary>ヒント 2</summary>

`-x` の出力だけでも「作ってから実行している」ことは推測できますが、`go` コマンド自身のソースコードで裏取りしてみましょう。
`go` コマンド（`cmd/go`）も Go で書かれた 1 つのプログラムなので、読み方の基本は普段のコードリーディングと同じです。初めての場合は、次の手順で少しずつ読み進めてみましょう。

1. **ファイルを開く**: https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go を開きます。`cmd/go` の中では、サブコマンドごとに `internal/<サブコマンド名>` というパッケージが分かれていて、`go run` の処理はこの `internal/run` パッケージにまとまっています。
   - URL の `refs/tags/go1.26.4` の部分は、手元の `go version` の結果とバージョンを揃えるためのものです。バージョンが違うと実装が変わっている可能性があります。
2. **いきなり全部読まない**: 最初から 1 行ずつ読もうとせず、まずはファイル内の `func` と書かれた行だけを目で追って、どんな関数があるかをざっと把握しましょう。`runRun` という、いかにも中心になりそうな関数が見つかるはずです（73 行目）。
3. **どこから呼ばれる関数か確認する**: `runRun` がいつ呼ばれるかを確認するため、`init()` 関数（65〜71 行目）を見てみましょう。`CmdRun.Run = runRun` という 1 行があります。これは「`go run` が実行されたときに、実際の処理として `runRun` 関数が呼ばれる」という意味です。ここが、これから読み進める出発点になります。
4. **関数の中身を上から読まず、まず末尾を見る**: `runRun` は長い関数ですが、今回の疑問（インタプリタかどうか）に関係するのは末尾の数行（170〜173 行目）です。
   ```go
   a1 := b.LinkAction(moduleLoaderState, work.ModeBuild, work.ModeBuild, p)
   a1.CacheExecutable = true
   a := &work.Action{Mode: "go run", Actor: work.ActorFunc(buildRunProgram), Args: cmdArgs, Deps: []*work.Action{a1}}
   b.Do(ctx, a)
   ```
   Go に詳しくなくても、次の 2 点だけ押さえれば十分です。
   - `a1` という変数に `LinkAction(...)` の結果を入れている（`Link` は「ビルドして実行ファイルを作る」ことを連想させる名前）。
   - その次に作っている `a` という変数は、`Deps: []*work.Action{a1}` として `a1` を「自分より先に終わらせておくもの」として持っている（`Deps` は dependencies＝依存関係の略）。
   - つまり「`a1`（ビルド）が終わってから `a`（実行）が行われる」という、2 段階の処理として書かれています。
5. **「実行」側の中身も覗いてみる**: `a` が実行するときに呼ばれる関数は `Actor: work.ActorFunc(buildRunProgram)` の `buildRunProgram` です。ファイル末尾（198〜201 行目）にあるので開いてみましょう。
   ```go
   func buildRunProgram(b *work.Builder, ctx context.Context, a *work.Action) error {
       cmdline := str.StringList(work.FindExecCmd(), a.Deps[0].BuiltTarget(), a.Args)
   ```
   `a.Deps[0].BuiltTarget()` は「1 つ前の依存アクション（＝ビルドの `a1`）が作った実行ファイルのパスを取り出す」という意味です（`BuiltTarget` という名前からも推測できます）。取り出したパスを、そのままコマンドとして実行しているだけで、ソースコードの中身（`main.go` の文字列）を読んでいるような処理はどこにも出てきません。

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
<summary>ヒント 1</summary>

- `go run -x main.go` の出力の 1 行目 `WORK=/var/folders/.../go-buildXXXXXXXXXX` を控えておきましょう。
- コマンドの実行が終わった後、そのパスに対して `ls` や `find` を実行し、実際にまだ存在するか確認してみましょう。
- `go help build` のフラグ一覧を眺め、一時ディレクトリに言及しているフラグを探してみましょう。見つけたフラグの説明文を読み、それが「デフォルトでは行われないことを追加で行う」フラグなのか、「デフォルトの動作を止める」フラグなのかを見分けてみましょう。

</details>

<details>
<summary>ヒント 2</summary>

設問 1 と同じ要領で、ソースコードでも裏取りしてみましょう。今回のエントリポイントは、`go help build` のフラグ一覧を管理している
https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go です。

1. **キーワードで検索する**: ファイル全体を読む必要はありません。ページを開いたらブラウザの検索（`Cmd`/`Ctrl` + `F`）で `WorkDir` と検索してみましょう。関連する箇所がいくつか見つかります。
2. **「作る」処理を先に見つける**: 検索結果の中に `NewBuilder` という関数（281 行目）があります。「新しい Builder を作る」という名前から、一時ディレクトリを作る処理もこの中にありそうだと当たりをつけられます。実際に中身を読むと、298 行目に次の行があります。
   ```go
   tmp, err := os.MkdirTemp(cfg.Getenv("GOTMPDIR"), "go-build")
   ```
   `os.MkdirTemp` が何をする関数か分からなければ、いったん https://pkg.go.dev/os#MkdirTemp を開いて確認しましょう。「一時ディレクトリを新規作成する関数」だと分かります。
3. **「片付ける」処理を探す**: Go では、後片付け（リソースを閉じる・掃除する）処理はよく `Close` という名前のメソッドに書かれます。同じファイル内で `Close` を検索すると、340 行目に `func (b *Builder) Close() error` が見つかります。中を読むと、352 行目に次の行があります。
   ```go
   if err := robustio.RemoveAll(b.WorkDir); err != nil {
   ```
   これが削除処理です。ただし 1 つ上の行に `if !cfg.BuildWork {` という条件が付いています。`cfg.BuildWork` は `-work` フラグが指定されたかどうかを保持する変数なので、「`-work` フラグが指定されていなければ削除する」と読めます。
4. **その処理が本当に呼ばれるのか確認する**: `Close` メソッドがあるだけでは、実際に呼ばれる保証にはなりません。設問 1 で開いた run.go に戻り、`runRun` 関数の冒頭（90〜94 行目）を見てみましょう。
   ```go
   defer func() {
       if err := b.Close(); err != nil {
           base.Fatal(err)
       }
   }()
   ```
   `defer` は Go の構文で「この関数（`runRun`）の処理が終わるときに、必ず実行する」という意味です。つまり `go run` が実行されるたびに、最後に必ずこの後片付けが呼ばれることになります。「作る処理」と「片付ける処理」が対になっていることを、自分でソースを行き来しながら確認してみましょう。

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

cmd/go のソースコードは 1 つのファイルにサブコマンドの処理がまとまっていることが多いので、迷ったら次の順で読むとつかみやすくなります。

1. `go help <サブコマンド>` でヘルプ文言を読む。
2. `-n` / `-x` フラグを付けて、実際に呼ばれるコマンド列を眺める。
3. `internal/<サブコマンド名>` パッケージの、サブコマンド名と同じファイルを開き、`init()` 内の `Cmd*.Run = ...` から実際の処理関数を特定する。
4. その関数を上から全部読むのではなく、まず `func` の一覧や末尾だけを見て構造をつかみ、疑問に関係ありそうな行から深掘りする。
5. バージョンタグ（`refs/tags/go1.26.4` など）は、手元の `go version` と揃える。
