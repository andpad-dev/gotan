# `go run` の裏側を覗いてみよう

先輩から「動作確認は `go run main.go` でいいよ」と言われました。
Python や Ruby の経験がある同僚は「`go run` ってスクリプト感覚でサクッと動くね。Go ってインタプリタでも動かせるんだ」と言っています。

Go はコンパイル言語のはずですが、この理解で合っているのでしょうか。`go run` を実行したとき、実際には何が起きているのか調べてみましょう。

（手元で試す例: https://go.dev/play/p/_BqTu8xBI9h ）

## 設問 1: `go run` は、ソースコードをインタプリタのように解釈しながら実行している？

同僚が言うように、`go run` はソースコードを逐次解釈しながら実行する「インタプリタ」的な動きをしているのでしょうか。
それとも、何か別の方法でプログラムを動かしているのでしょうか。

<details>
<summary>ヒント</summary>

- まずは `go help run` を実行して、コマンドの説明を読んでみましょう
- `go run -x main.go` のように `-x` フラグを付けて実行すると、実際に裏で呼ばれているコマンド列が表示されます。上記の Go Playground のコードをファイルに保存して手元で試してみましょう

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. `go help run`（= https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program ）を読む。「Run compiles and runs the named main Go package.」とあり、"compiles" という単語が使われている。インタプリタなら「解釈・実行する」という説明になるはずで、まず「コンパイルする」と書かれている時点で疑わしい。
2. 実際に手元で `go run -x main.go` を実行し、裏で呼ばれているコマンド列を確認する。
3. cmd/go のソース（ https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go ）を読むと、`runRun` 関数がパッケージをビルドするアクション（`b.LinkAction`）と、ビルドした実行ファイルを実行するアクション（`buildRunProgram`）を順につないでいることが分かる。

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

- `go run -x main.go` の出力から `WORK=...` の行を控えておき、コマンドの実行が終わった後に、そのパスが実際に存在するか自分で確認してみましょう
- 消えるとしたら、どこでその処理をしているのか、`go help build` のフラグ一覧にヒントがあります

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. `go run -x main.go` で表示された `WORK=...` のパスを、コマンド終了後に `ls` してみる → ディレクトリごと消えている。
2. `go help build` を読むと `-work` フラグの説明に「print the name of the temporary work directory and do not delete it when exiting」とあり、"delete it when exiting" が **デフォルトの挙動である**ことが読み取れる。
3. 実際の削除処理は cmd/go のソース（ https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go ）にある。`NewBuilder` が `os.MkdirTemp(cfg.Getenv("GOTMPDIR"), "go-build")` で一時ディレクトリを作り（207〜行目付近）、`Builder.Close()` が `-work` フラグが指定されていない限り `robustio.RemoveAll(b.WorkDir)` で削除している（343〜行目付近）。

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
