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

## 設問 3: 「ディレクトリの作成」「ビルドの実行」「ビルドしたファイルの扱い」は、それぞれどこで行われている？

設問 1・2 で、`go run` は ①一時ディレクトリを作り → ②その中でビルドし → ③できた実行ファイルを実行し → ④最後にディレクトリごと片付ける、という流れだと分かりました。
では実際に、この 3 つの処理（ディレクトリの作成・ビルドの実行・ビルドしたファイルの扱い）は、`go` コマンド自身のソースコードのどこに書かれているのでしょうか。

**前提知識: `go` のサブコマンドはソースコード上のどこにあるか**

`go` コマンドは 1 つの大きなプログラムですが、内部ではサブコマンドごとに `cmd/go/internal/<サブコマンド名>` というパッケージに実装が分かれています。たとえば `go run` なら `cmd/go/internal/run`、`go build` なら `cmd/go/internal/work` です。各パッケージにはサブコマンド名と同じファイル（`run.go` など）があり、その中で `CmdRun.Run = runRun` のように `Cmd*.Run` フィールドへ登録されている関数が、そのサブコマンドが実行されたときに実際に呼ばれるエントリポイントになります（詳しくは [`03-cmd-tools` のカテゴリ README](../../README.md) を参照）。

つまり `go run` を読み解くなら、まず `cmd/go/internal/run` パッケージのエントリポイントを開くところから始めます。

<details>
<summary>ヒント 1: run.go を読んで、どこを掘り下げればよいか当たりをつける</summary>

`cmd/go/internal/run` パッケージのエントリポイントとなる関数（https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go ）を開き、中で呼ばれている関数・型を上から順に眺めてみましょう。

- 1 つ 1 つの処理が、`run` パッケージ自身の中で定義されたものなのか、それとも別のパッケージのものなのかを見分けてみましょう（呼び出しの前についているパッケージ名がヒントになります）。
- 別パッケージの処理が見つかったら、そのパッケージ名から「何を担当していそうか」を推測してみましょう。ビルドやリンクに関係していそうな名前のパッケージが見つかれば、そこが次に読むべき場所です。
- 「これから何かを準備している」ように見える処理と、「準備したものを使って実行している」ように見える処理を区別できると、設問の 3 つの問い（ディレクトリ作成・ビルド実行・ファイルの扱い）がそれぞれどのあたりに対応するかが見えてきます。

</details>

<details>
<summary>ヒント 2: 見つけたパッケージの中をどう読み解くか</summary>

ヒント 1 で見つかるパッケージは、`go build` とも共通処理を担当するくらい大きく、ファイルも複数に分かれています。全部を上から読もうとせず、次のように的を絞ってみましょう。

- ページ内検索（`f` キー）で、「ディレクトリを作る／消す」といった処理に典型的に出てくる標準ライブラリの関数名（例: `os.MkdirTemp` や `RemoveAll` のような名前）を探してみましょう。
- 「実行ファイルの実体」を表していそうな変数・フィールド名を探し、それがどこで値を設定され、どこで読み出されているかを追ってみましょう。
- 「コンパイラ・リンカを実際に呼び出している」処理は、さらに別の関数に分かれています。ビルドの動詞そのもの（コンパイルする・リンクする）を含むような、短い名前の関数を探すと見つけやすいです。
- 1 つの関数を最後まで読み切ろうとせず、空行で区切られたブロックごとに「これは準備をしているブロックか」「実際に何かを実行しているブロックか」を大づかみするのがコツです（設問 1 のヒントで使った読み方と同じです）。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go の `runRun` 関数（73 行目〜）を読む。
2. `work.NewBuilder` の実体（https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go の 281 行目〜）を読み、一時ディレクトリの作成箇所を確認する。
3. `LinkAction`（同ファイル 922 行目〜）と `CompileAction`（633 行目〜）で、パッケージごとのサブディレクトリ（`Objdir`）と実行ファイルのパス（`Target`）がどう決まるかを確認する。
4. https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/exec.go の `Builder.build`（723 行目〜、コンパイル担当）と `Builder.link`（1591 行目〜、リンク担当）を読み、実際にコンパイラ・リンカを呼び出している場所を確認する。
5. `Builder.Close`（action.go 340 行目〜）を読み、後片付けの実装を確認する。

**答え**

**① ディレクトリの作成はどこで行われるか**

`runRun`（run.go 89 行目）が `work.NewBuilder("", ...)` を呼ぶと、その内部（action.go 281〜310 行目）で

```go
tmp, err := os.MkdirTemp(cfg.Getenv("GOTMPDIR"), "go-build")
...
b.WorkDir = tmp
```

が実行され、`$WORK` にあたる一時ディレクトリが 1 つ作られます。

さらにパッケージごとに、`$WORK` の下へ `b001/`, `b002/`... というサブディレクトリ（`Objdir`）のパスを割り当てる処理が `NewObjdir`（action.go 391 行目〜）にあります。

```go
func (b *Builder) NewObjdir() string {
	b.objdirSeq++
	return str.WithFilePathSeparator(filepath.Join(b.WorkDir, fmt.Sprintf("b%03d", b.objdirSeq)))
}
```

ただしこの時点ではまだパス文字列を組み立てているだけです。実際にディレクトリを作成しているのは、後述する `Builder.build`／`Builder.link`（exec.go）の中にある `sh.Mkdir(a.Objdir)` です。つまり「`$WORK` 本体」と「パッケージごとの `bNNN/` ディレクトリ」は、パスが決まるタイミングと実際に作られるタイミング（コード上の場所）が分かれています。

**② ビルドの実行はどこで行われるか**

`runRun`（run.go 170〜173 行目）に、次の 3 行があります。

```go
a1 := b.LinkAction(moduleLoaderState, work.ModeBuild, work.ModeBuild, p)
a1.CacheExecutable = true
a := &work.Action{Mode: "go run", Actor: work.ActorFunc(buildRunProgram), Args: cmdArgs, Deps: []*work.Action{a1}}
b.Do(ctx, a)
```

`LinkAction`（action.go 922 行目〜）は「まずコンパイルし、それが終わったらリンクする」という依存関係を持った `Action`（`a1`）を組み立てているだけで、実際にコンパイラ・リンカを呼び出す処理はここには書かれていません。実際に外部コマンドを実行しているのは `work/exec.go` の `Builder.build`（723 行目〜、コンパイルを担当）と `Builder.link`（1591 行目〜、リンクを担当）です。`b.Do(ctx, a)` が依存関係をたどりながらこの 2 つの関数を順番に呼び出すことで、初めて `compile`・`link` が実際に動きます。つまり「何をビルドするかを組み立てる場所（`LinkAction`）」と「実際にビルドする場所（`Builder.build`/`Builder.link`）」は、コード上で分かれています。

**③ ビルドしたファイルはどのように扱われるか**

`LinkAction`（action.go 955 行目）で、リンク後にできる実行ファイルのパスが

```go
a.Target = a.Objdir + filepath.Join("exe", name) + cfg.ExeSuffix
a.built = a.Target
```

として `Action` の `built` フィールドに記録されます。`go run` が実際に実行する最後の一手である `buildRunProgram`（run.go 200 行目〜）は、この値を `a.Deps[0].BuiltTarget()` として取り出し、実行ファイルとして起動します。

```go
cmdline := str.StringList(work.FindExecCmd(), a.Deps[0].BuiltTarget(), a.Args)
```

そして実行が終わったあと、`runRun`（run.go 90〜94 行目）が `defer` していた `Builder.Close()`（action.go 340 行目〜）が呼ばれ、

```go
if !cfg.BuildWork {
	if err := robustio.RemoveAll(b.WorkDir); err != nil {
		return err
	}
}
```

という処理で `$WORK` ディレクトリ（＝実行ファイルを含む中間生成物すべて）がまとめて削除されます。つまりビルドされた実行ファイルは、「一時ディレクトリの中に作られる → パスだけを覚えておいて起動される → 実行後にディレクトリごと削除される」という、使い捨てのファイルとして扱われています。

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
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/exec.go
