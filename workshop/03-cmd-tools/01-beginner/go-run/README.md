[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [03-cmd-tools の調べ方](../../README.md)

# `go run` の裏側を覗いてみよう

![実行環境: Go 1.24 以上](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.24%20%E4%BB%A5%E4%B8%8A-F39C12)

`go run` を見かけました。どんなものか調べてみましょう。

同僚に「`go run` はスクリプト感覚でサクッと動く」と言われました。
「Go もインタプリタで動かせるんだ」とも言っています。

でも Go はコンパイル言語のはずです。
`go run` を実行したとき、**裏では何が起きている**のでしょうか。

手元で試す例: https://go.dev/play/p/_BqTu8xBI9h
このディレクトリの [main.go](./main.go) も同じ内容です。`go run` してそのまま調査に使えます。

## 設問 1: `go run` は、ソースコードをインタプリタのように解釈しながら実行している？

`go run` は、ソースコードを 1 行ずつ解釈しながら動かしているのでしょうか。
それとも、別の方法で動かしているのでしょうか。

<details>
<summary>ヒント</summary>

- `go help run` の説明文を 1 文ずつ読み、動詞に注目してみましょう。「解釈する・実行する」ではなく、別の動詞が使われていないでしょうか。
- この [main.go](./main.go) に対して `go run -a -x main.go` を実行してみましょう。`-x` は裏で呼ばれたコマンドを表示するフラグです。`-a` はキャッシュ済みのパッケージも含めてビルドし直すフラグです。
- 出力される行を 1 行ずつ眺めて、次の 2 種類に仕分けしてみましょう。
  - 何かを新しく作っていそうな行（コマンド名に `compile` や `link` が含まれる行、`mkdir` の行）
  - すでに出来上がった何かを実行していそうな行（パスだけがポツンと書かれた行）
  - この仕分けができると、「何かを作ってから動かしている」のか「ソースコードをその場で読みながら動かしている」のか、判断する材料になります。
- 続けて `-a` を外した `go run -x main.go` も実行し、表示される行が減るか比べてみましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://go.dev/cmd/go/ を開き、`go` コマンドの公式ドキュメントから `go run` の説明を探す。
2. `go help run`（= https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program ）を読む。「Run compiles and runs the named main Go package.」とあり、動詞は "compiles"（コンパイルする）。インタプリタなら "interprets" や "evaluates" のような動詞になるはずで、この時点で疑わしい。
3. `go run -a -x main.go` を実行し、出力を「生成している行」と「実行している行」に仕分ける。`-a` を付けると、直前に `go run` 済みでもコンパイルとリンクを観測できる。
4. 続けて `go run -x main.go` を実行し、2 回目はビルドキャッシュ内の実行ファイルが直接起動されることを確認する。

**答え**

インタプリタではありません。Go 1.27.0 で `-a -x` フラグを付けて実行すると、次のような出力になります（環境依存部分は省略しています）。

```
WORK=/var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-buildXXXXXXXXXX
...
mkdir -p $WORK/b001/exe/
.../compile ... # ソースをコンパイル
.../link -o $WORK/b001/exe/main ... # 実行ファイルにリンク
cp $WORK/b001/exe/main <GOCACHE>/.../main # ビルドキャッシュへ保存
$WORK/b001/exe/main # できた実行ファイルを実行
hello, gotan
```

`$WORK` という一時ディレクトリを作り、その中でコンパイルとリンクを行っています。
できあがった実行ファイルを最後に起動しているだけです。
`main.go` を 1 行ずつ読みながら評価しているわけではありません。

Go 1.24 以降は `go run` の実行ファイルもビルドキャッシュに保存されます。
そのため同じ条件で再実行すると、`compile` と `link` が省略されることがあります。
その場合の `go run -x main.go` は次のように短くなり、キャッシュ内の実行ファイルを直接起動します。
表示が短くても、インタプリタとして動いたという意味ではありません。

```text
WORK=/tmp/go-buildXXXXXXXXXX
<GOCACHE>/.../main
hello, gotan
```

`WORK` のパスやビルドIDは実行環境ごとに変わるため、上の出力では環境依存部分を `X` と `...` で省略しています。

</details>

---

## 設問 2: ビルドされた実行ファイルや `$WORK` ディレクトリは、実行後どうなる？

`$WORK` は、設問 1 の出力に出てきた一時ディレクトリです。
この `$WORK` と、中に作られた実行ファイルは、実行後も**残る**のでしょうか。
それとも消えるのでしょうか。

<details>
<summary>ヒント</summary>

- `go run -a -x main.go` の出力の 1 行目 `WORK=/var/folders/.../go-buildXXXXXXXXXX` と、`cp` 行に出るビルドキャッシュ側のパスを控えておきましょう。
- コマンドの実行が終わった後、そのパスに対して `ls` や `find` を実行し、実際にまだ存在するか確認してみましょう。
- `go help build` のフラグ一覧を眺め、一時ディレクトリに言及しているフラグを探してみましょう。見つけたフラグの説明文を読み、それが「デフォルトでは行われないことを追加で行う」フラグなのか、「デフォルトの動作を止める」フラグなのかを見分けてみましょう。
- 見つけたフラグを試すときも `-a` を併用します。キャッシュが温まった状態では、ビルドを省略して空の `$WORK` だけが残ることがあるためです。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://go.dev/cmd/go/ を開き、`go build` のドキュメントにある `-work` フラグの説明を探す。
2. `go run -a -x main.go` で表示された `WORK=...` のパスを、コマンド終了後に `ls` する。一時ディレクトリは消えているが、`cp` 行が示すビルドキャッシュ側の実行ファイルは残っていることを確認する。
3. `go help build` を読むと `-work` フラグの説明に「print the name of the temporary work directory and do not delete it when exiting」とあり、"delete it when exiting" が **デフォルトの挙動である**ことが読み取れる。
4. `go run -a -work main.go` を実行し、`$WORK/b001/exe/main` を確認する。`-a` を外した場合はキャッシュが使われ、`$WORK` が空になることも比較する。
5. [Go 1.24 リリースノートの Go command 節](https://go.dev/doc/go1.24#go-command)を読み、`go run` の実行ファイルが Go 1.24 からビルドキャッシュに保存されるようになったことを確認する。
6. リリースノートから [提案 Issue #69290](https://go.dev/issue/69290) を開き、繰り返し実行を速くする目的とキャッシュ容量のトレードオフを確認する。

**答え**

`$WORK` ディレクトリと、その中にリンクされた実行ファイルは、通常は実行後に自動削除されます。ただし Go 1.24 以降では、リンク済み実行ファイルのコピーがビルドキャッシュに残ります。削除される一時ファイルと、再実行のために保存されるキャッシュを分けて考える必要があります。

`-a -work` を付けるとビルドをやり直したうえで `$WORK` を残せるので、中間生成物を確認できます。

```
$ go run -a -work main.go
WORK=/var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-buildXXXXXXXXXX
hello, gotan
$ find /var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-buildXXXXXXXXXX -maxdepth 3
.../go-buildXXXXXXXXXX
.../go-buildXXXXXXXXXX/b001
.../go-buildXXXXXXXXXX/b001/importcfg
.../go-buildXXXXXXXXXX/b001/importcfg.link
.../go-buildXXXXXXXXXX/b001/exe
.../go-buildXXXXXXXXXX/b001/_pkg_.a
.../go-buildXXXXXXXXXX/b001/exe/main
```

`b001/exe/main` が、この実行でリンクされて起動した実行ファイルです。
`-work` を付けなければ、`$WORK` とともに削除されます。
ただし `-a -x` の `cp` 行で確認できるビルドキャッシュ側のコピーは残ります。
次回はそのコピーを直接起動できるため、`-work` だけを付けても `$WORK` の中に生成物がない場合があります。

</details>

---

## 設問 3: 「ディレクトリの作成」「ビルドの実行」「ビルドしたファイルの扱い」は、それぞれどこで行われている？

設問 1・2 で分かった `go run` の流れは、次のとおりです。

1. 一時ディレクトリを作る
2. その中でビルドする
3. 実行ファイルをビルドキャッシュにも保存する
4. 実行ファイルを起動する
5. 一時ディレクトリを片付ける

キャッシュに同じ実行ファイルがあれば、1〜3 を省いて直接起動します。

では、この流れは `go` コマンド自身のソースコードのどこに書かれているのでしょうか。
**ディレクトリの作成**・**ビルドの実行**・**ビルドしたファイルの扱い**の 3 つを探します。

**前提知識: `go` のサブコマンドはどこに実装されているか**

サブコマンドの実装は `cmd/go/internal/<サブコマンド名>` に分かれています。
`go run` なら `cmd/go/internal/run` です。`go build` なら `cmd/go/internal/work` です。

各パッケージには `run.go` のように、サブコマンド名と同じ名前のファイルがあります。
そのファイルにある `CmdRun.Run = runRun` という登録に注目しましょう。
ここで指定された関数が、サブコマンドの実行時に最初に呼ばれます。
この関数を**エントリポイント**と呼びます。

詳しくは [`03-cmd-tools` のカテゴリ README](../../README.md) を参照してください。
まずは `cmd/go/internal/run` のエントリポイントを開くところから始めましょう。

<details>
<summary>ヒント 1: run.go を読んで、どこを掘り下げればよいか当たりをつける</summary>

`cmd/go/internal/run` パッケージのエントリポイントとなる関数（https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/run/run.go ）を開き、中で呼ばれている関数・型を上から順に眺めてみましょう。

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

1. https://go.dev/cmd/go/ を開き、`go run` が `cmd/go` の一部として実装されていることを確認する。
2. https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/run/run.go の `runRun` 関数（77 行目〜）を読む。
3. `work.NewBuilder` の実体（https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/work/action.go の 281 行目〜）を読み、一時ディレクトリの作成箇所を確認する。
4. `LinkAction`（同ファイル 922 行目〜）と `CompileAction`（633 行目〜）で、パッケージごとのサブディレクトリ（`Objdir`）と実行ファイルのパス（`Target`）がどう決まるかを確認する。
5. https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/work/exec.go の `Builder.build`（726 行目〜、コンパイル担当）と `Builder.link`（1628 行目〜、リンク担当）を読み、実際にコンパイラ・リンカを呼び出している場所を確認する。
6. https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/work/buildid.go の 756 行目〜を読み、リンク済み実行ファイルをビルドキャッシュへ保存する条件を確認する。
7. `Builder.Close`（action.go 340 行目〜）を読み、後片付けの実装を確認する。

**答え**

**① ディレクトリの作成はどこで行われるか**

`runRun`（run.go 93 行目）が `work.NewBuilder("", ...)` を呼ぶと、その内部（action.go 281〜310 行目）で

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

ただしこの時点では、まだパス文字列を組み立てているだけです。
実際にディレクトリを作成しているのは、後述する `Builder.build`／`Builder.link`（exec.go）の中にある `sh.Mkdir(a.Objdir)` です。
つまり「`$WORK` 本体」と「パッケージごとの `bNNN/` ディレクトリ」では、パスが決まる場所と実際に作られる場所が違います。

**② ビルドの実行はどこで行われるか**

`runRun`（run.go 174〜177 行目）に、次の 4 行があります。

```go
a1 := b.LinkAction(moduleLoader, work.ModeBuild, work.ModeBuild, p)
a1.CacheExecutable = true
a := &work.Action{Mode: "go run", Actor: work.ActorFunc(buildRunProgram), Args: cmdArgs, Deps: []*work.Action{a1}}
b.Do(ctx, a)
```

`LinkAction`（action.go 922 行目〜）は、「まずコンパイルし、それが終わったらリンクする」という依存関係を持つ `Action`（`a1`）を組み立てます。
コンパイラ・リンカを呼び出す処理は、ここには書かれていません。
実際に外部コマンドを実行しているのは `work/exec.go` の `Builder.build`（726 行目〜、コンパイルを担当）と `Builder.link`（1628 行目〜、リンクを担当）です。
`b.Do(ctx, a)` がこの 2 つを順番に呼び出すことで、初めて `compile`・`link` が動きます。
つまりビルドを組み立てる場所（`LinkAction`）と、実際にビルドする場所（`Builder.build`/`Builder.link`）は分かれています。

**③ ビルドしたファイルはどのように扱われるか**

`LinkAction`（action.go 955 行目）で、リンク後にできる一時実行ファイルのパスが

```go
a.Target = a.Objdir + filepath.Join("exe", name) + cfg.ExeSuffix
a.built = a.Target
```

として `Action` の `built` フィールドに記録されます。さらに `runRun` は `CacheExecutable = true` を設定し、`buildid.go` の 756〜770 行目がリンク済み実行ファイルをビルドキャッシュへコピーします。

`go run` が実際に実行する最後の一手である `buildRunProgram`（run.go 204 行目〜）は、`a.Deps[0].BuiltTarget()` を実行ファイルとして起動します。ビルド直後は `$WORK` 側、キャッシュが使える再実行ではビルドキャッシュ側のパスが返ります。

```go
cmdline := str.StringList(work.FindExecCmd(), a.Deps[0].BuiltTarget(), a.Args)
```

そして実行が終わったあと、`runRun`（run.go 94〜98 行目）が `defer` していた `Builder.Close()`（action.go 340 行目〜）が呼ばれ、

```go
if !cfg.BuildWork {
	if err := robustio.RemoveAll(b.WorkDir); err != nil {
		return err
	}
}
```

という処理で `$WORK` ディレクトリ（＝一時実行ファイルを含む中間生成物）がまとめて削除されます。ビルドキャッシュは `$WORK` の外にあるため削除対象ではありません。つまり、一時側は実行後に消えますが、同じプログラムを再実行するためのコピーはキャッシュに残ります。

</details>

---

<details>
<summary>こぼれ話: なぜ実行ファイルまでキャッシュするのか</summary>

[提案 Issue #69290](https://go.dev/issue/69290) では、同じ版のツールを `go run package@version` や `go tool` で繰り返し起動する用途が主に議論されました。リンクを省略できる一方、大きな実行ファイルによってキャッシュ容量が増える点もトレードオフとして扱われています。

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

- https://go.dev/cmd/go/
- https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program
- https://go.dev/doc/go1.24#go-command
- https://go.dev/issue/69290
- https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/run/run.go
- https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/work/action.go
- https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/work/exec.go
- https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/work/buildid.go
