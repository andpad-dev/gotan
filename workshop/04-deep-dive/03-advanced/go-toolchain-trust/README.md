[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# クリーンなソースコードだけで Go コンパイラを信頼できる？

リリースを担当するチームで、Go ツールチェーンのサプライチェーン監査をすることになりました。
同僚は「コンパイラのソースツリーに差分はなく、実行ファイルにも `go1.27.0` と書かれているので、配布されたコンパイラはソースどおりだ」と考えています。

まず、コンパイラ実行ファイルをコンパイル処理には使わず、埋め込まれたビルド情報を表示してみます。
次は Go 1.27.0 / macOS (arm64) で実行した結果です。`<GOTOOLDIR>` の部分は環境によって異なります。

```console
$ GOTOOLCHAIN=go1.27.0 go version -m "$(GOTOOLCHAIN=go1.27.0 go env GOTOOLDIR)/compile"
<GOTOOLDIR>/compile: go1.27.0
	path	cmd/compile
	build	-buildmode=exe
	build	-compiler=gc
	build	-gcflags=cmd/...=-dwarf=false
	build	-pgo=default.pgo
	build	-trimpath=true
	build	CGO_ENABLED=0
	build	GOARCH=arm64
	build	GOARM64=v8.0
	build	GOOS=darwin
```

この表示とクリーンなソースツリーは、どこまで信頼の根拠になるのでしょうか。過去のコンパイラ攻撃を現在の Go 実装で追い、Go 1.27 のブートストラップと再現可能なビルドが何を検証しているのか調べます。

---

## 設問 1: `go version -m` の表示だけで何を確認できる？

冒頭のコマンドは、対象ファイルについて何を読み、何を表示しているのでしょうか。
表示されたバージョン名とビルド設定から確認できることと、それだけでは確認できないことを分けて説明してください。

<details>
<summary>ヒント</summary>

- `go help version` で、`-m` の説明に使われている動詞を確認します。
- コマンドが対象ファイルを実行しているのか、ファイル内の情報を読んでいるのかを区別します。
- 出力に、ソースツリーや公開済みハッシュとの比較結果が含まれているかも確認します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc) から Go コマンドのドキュメントへ進み、[`go version` の説明](https://go.dev/cmd/go/#hdr-Print_Go_version)を読む。`-m` は実行ファイルに埋め込まれたモジュール・ビルド情報を表示するオプションだと確認する。
2. Go 1.27.0 の [`cmd/go/internal/version/version.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/version/version.go) で `scanFile` を探し、対象ファイルを [`debug/buildinfo.ReadFile`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/debug/buildinfo/buildinfo.go) に渡していることを確認する。
3. `debug/buildinfo.ReadFile` のコメントと実装を読み、Go バイナリ内の build info blob を読み出す処理であることを確認する。

**答え**

`go version -m` は対象の Go バイナリを実行せず、ファイル内に埋め込まれた Go バージョン、パッケージパス、ビルド設定などを読み出します。したがって、手元のバイナリを棚卸しし、「そのファイルがどのようなビルド情報を持っているか」を確認するには役立ちます。

一方、この処理は公開ソースからバイナリを再構築して比較するものでも、署名や公開済みハッシュで出所を検証するものでもありません。表示されるのはあくまで対象ファイルに埋め込まれた情報です。バージョン名が `go1.27.0` でソースツリーに差分がなくても、「そのバイナリが、そのソースから余計な変更なしに作られた」ことまでは確認できません。

</details>

---

## 設問 2: ソースを元に戻した後も、変更された挙動が残るのはなぜ？

[Go compiler の公式ドキュメント](https://go.dev/cmd/compile/)で現在のコンパイラを確認してから、Russ Cox の [Running the “Reflections on Trusting Trust” Compiler](https://research.swtch.com/nih) を読むと、変更を加えた Go コンパイラを一度インストールした後でソースを元に戻し、クリーンなソースからコンパイラを再インストールしても、`hello, world` が `backdoored!` に変わる現代版の実演があります。

記事の悪性コードを実行せず、説明と Go 1.27.0 のソースだけを読んで、次を説明してください。

- 通常のプログラムとコンパイラ自身のソースは、それぞれどのように書き換えの対象になりますか。
- なぜ `git diff` が空でも、次に作られるコンパイラへ変更を引き継げますか。
- Go 1.27.0 の現在の実装にも、記事が着目した入力境界は残っていますか。

<details>
<summary>ヒント</summary>

- コンパイラがソースを構文として解釈する直前に、入力を受け取る境界を探します。
- 書き換えの対象がアプリケーションだけなら、ソースを元に戻した後のコンパイラ再構築では何が起きるか考えます。
- この記事は攻撃が実在する証拠ではなく、仕組みを示す管理された実演として読みます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go compiler の公式ドキュメント](https://go.dev/cmd/compile/) からコンパイラの役割を確認し、ソースコードを読み込んでコンパイルする処理を調査対象にする。
2. [Running the “Reflections on Trusting Trust” Compiler](https://research.swtch.com/nih) の「A Modern Version」を読み、通常のプログラムを変える処理と、コンパイラ自身へ同じ処理を再注入する処理を分けて追う。
3. Go 1.27.0 の [`cmd/compile/internal/syntax/syntax.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/syntax/syntax.go) で `Parse` を探し、ソース入力を受け取って parser を初期化する現在の境界を記事の説明と照合する。

**答え**

記事の現代版は、コンパイラが構文解析するソース入力の手前に読み取り処理を挟みます。その処理には 2 種類の対象があります。

1. 特徴的な `hello, world` プログラムを見つけたら、文字列を別の出力へ書き換える。
2. コンパイラの構文解析器自身を見つけたら、同じ読み取り処理と、その処理を再生成するためのデータをコンパイラのソース入力へ加える。

最初の悪性コンパイラを使ってクリーンなコンパイラソースをビルドすると、ディスク上のソースを変更する前に 2 番目の処理が入力を変えます。そのため、`git diff` は空のままでも、生成された次のコンパイラには同じ処理が含まれます。そのコンパイラがさらにクリーンなソースをコンパイルしても変更を再生成でき、バイナリだけに状態を残せます。

Go 1.27.0 の `syntax.Parse` もソースを `io.Reader` として受け取り、`p.init(base, src, ...)` へ渡しています。記事の 2023 年時点の完全な差分がそのまま適用できるとまでは、この確認だけでは断定できません。しかし、記事が着目した「構文解析前のソース入力」という境界は現在も存在します。

これは仕組みを示す実演であり、公式 Go ツールチェーンに同じ変更が含まれているという主張ではありません。

</details>

---

## 設問 3: Go 1.27 はどのコンパイラから作られる？

Go のコンパイラ自身も Go で書かれています。現在のコンパイラだけで同じ版のコンパイラを作る閉じた循環にしてしまうと、設問 2 のような変更をソースの外へ隠す余地が残ります。

Go 1.27.0 のビルドを開始するために必要な過去の Go バージョンを求め、`toolchain1`、`go_bootstrap`、`toolchain2`、`toolchain3` がどの順序で作られるかを調べてください。その段階的なビルドで分かることと、それだけではまだ証明できないことも説明してください。

<details>
<summary>ヒント</summary>

- ソースからのインストール手順にある、Go 1.N とブートストラップ版の関係を式として読んでみます。
- Go 1.27.0 の `cmd/dist` ソースで、問題文に出てきた 4 つの名前を検索します。
- すべての段階が同じ祖先のバイナリから始まる場合に、設問 2 の仕組みを完全に排除できるか考えます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Installing Go from source](https://go.dev/doc/install/source#go14) のブートストラップ要件を読み、Go 1.N が必要とする Go 1.M の求め方と、Go 1.4 が C で書かれた最後のツールチェーンであることを確認する。
2. Go 1.27.0 の [`cmd/dist/buildtool.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/dist/buildtool.go) で `bootstrapBuildTools` を読み、過去版のツールチェーンから最初のツール群を作る処理を確認する。
3. 同じタグの [`cmd/dist/build.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/dist/build.go) で `toolchain1`、`go_bootstrap`、`toolchain2`、`toolchain3` を検索し、各段階の入力と出力を追う。
4. [Running the “Reflections on Trusting Trust” Compiler](https://research.swtch.com/nih) の「Bootstrapping Go」を読み、現在版が過去のリリースからビルドできることと、Go 1.4 の C 実装まで遡る意味を確認する。

**答え**

公式手順では、今後の Go 1.N は「N から 2 を引き、偶数へ切り下げた」Go 1.M をブートストラップに要求します。N=27 なら M=24 です。さらに Go 1.27.0 タグの `buildtool.go` は `minBootstrap` を `go1.24.6` と固定しているため、この版の通常のビルドは Go 1.24.6 以降の適切なツールチェーンから開始します。

Go 1.27.0 の `cmd/dist` に書かれた大筋は次のとおりです。

1. 過去版の Go ツールチェーンと `cmd/go` を使い、新しいソースから `toolchain1` を作る。
2. `toolchain1` と `cmd/dist` を使い、新しい `cmd/go` を `go_bootstrap` として作る。
3. `go_bootstrap` と `toolchain1` を使い、ビルド情報を含む `toolchain2` を作る。
4. `go_bootstrap` と `toolchain2` を使い、同じ新しいソースから `toolchain3` を作る。

現在版を現在版だけで最初から作る必要はなく、リリース済みの過去版から段階的に構築できます。さらに必要な版を順に遡れば、Go で書かれたコンパイラより前の Go 1.4 C 実装を出発点にできます。

ただし、普段の `make.bash` を Go 1.24 のバイナリから 1 回実行しただけでは、その Go 1.24 バイナリ自体の出所までは検証していません。また、同じ祖先から作った後続段階が一致しても、祖先が自己再生成する変更を注入する仮説をそれだけで排除できるとは限りません。過去まで遡れる構造を、独立した再構築と出力比較に使う必要があります。

</details>

---

## 設問 4: 再ビルドの `PASS` は何を証明する？

[Go Reproducible Build Report](https://go.dev/rebuild) では、公開された Go ツールチェーンを `gorebuild` で再構築した結果を確認できます。同僚はこのページの `PASS` を見て、「これで Go のソースも、自分の PC にあるコンパイラも、絶対に安全だと証明された」と言っています。

`gorebuild` が比較する対象、ブートストラップの方法、プラットフォームごとの比較方法を調べ、この説明の正しい部分と過剰な部分を分けてください。コマンドの終了コードを見るだけで検証成功と判断できるかも確認します。

<details>
<summary>ヒント</summary>

- 「同じ入力から同じ出力になること」と「入力されたソースが安全であること」を分けます。
- 公開レポートが検証する配布物と、手元のパッケージマネージャが配置したファイルが同じ対象か確認します。
- ツールのドキュメントで、終了ステータスとレポート内の成否が同じ意味かを読みます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Perfectly Reproducible, Verified Go Toolchains](https://go.dev/blog/rebuild) を読み、Go 1.21 以降の「perfectly reproducible」の定義、ブートストラップ段階、配布物を別環境で比較する目的を確認する。
2. [Go Reproducible Build Report](https://go.dev/rebuild) の実行日時、対象バージョン、各配布物の結果を読み、`PASS` までの実際のログを確認する。
3. ブログから [`gorebuild` の公式ドキュメント](https://pkg.go.dev/golang.org/x/build/cmd/gorebuild) へ進み、Linux での完全なブートストラップ、macOS と Windows の例外、終了ステータスと生成されるレポートの関係を読む。続いて、掲載版の [`cmd/gorebuild/main.go`](https://cs.opensource.google/go/x/build/+/f316b62c:cmd/gorebuild/main.go) で同じ説明と処理の入口を照合する。
4. [Running the “Reflections on Trusting Trust” Compiler](https://research.swtch.com/nih) の「Bootstrapping Trust」と「Reproducible Builds」を読み、独立した系統でコンパイラを作って比較する意味と、再現可能性が必要になる理由を整理する。

**答え**

Go 1.21 以降のツールチェーンは、同じソースから対象 OS・アーキテクチャ向けに作れば、ホスト環境やブートストラップ版などを変えても同じツールチェーンを得られるよう設計されています。`gorebuild` はソースからツールチェーン配布物を作り直し、`go.dev/dl` にある公開配布物と比較します。Linux/amd64 では Go 1.4 の C 実装から必要な過去版を順に作る完全なブートストラップも行います。

多くの配布物はビット単位で比較します。ただし、macOS の実行ファイルはコード署名を除いて比較し、PKG は中身を tar.gz と比較します。Windows の MSI も再生成するのではなく、`msiextract` が使える場合に中身を対応する zip と比較し、使えなければその MSI をスキップします。したがって、すべての配布形式を一律に「アーカイブ全体がビット単位で一致した」と説明するのは誤りです。

独立した再ビルドが一致すれば、公開バイナリに対応するソース外の変更が紛れ込んでいないことへの強い根拠になります。設問 1 の埋め込み情報を読むだけの確認や、設問 2 のクリーンなソースツリーを見るだけの確認より一段強く、設問 3 の過去版へ遡れる構造を実際の比較に使っています。

それでも、次のことまでは証明しません。

- 公開ソースそのものに悪意や脆弱性がないこと。
- あらゆる独立ビルド環境が同じ形で侵害されていないこと。
- `go.dev/dl` の配布物ではなく、別のパッケージマネージャが作った手元のバイナリも同一であること。

さらに `gorebuild` は、検証対象がすべて一致したかどうかではなく、レポートを書き出せたときに終了ステータス 0 を返します。自動化では終了コードだけを見ず、生成された JSON / HTML レポート内の各結果を確認する必要があります。

したがって同僚の説明は、「公開配布物とソースの対応を独立再ビルドで強く検証できる」という部分は正しく、「ソースと手元の任意のコンパイラの絶対的な安全まで証明する」という部分は過剰です。

</details>

---

<details>
<summary>こぼれ話: サプライチェーンの別の層</summary>

この問題は、ソースからコンパイラ配布物を作る層に焦点を当てました。依存モジュールの取得には、別の対策があります。[How Go Mitigates Supply Chain Attacks](https://go.dev/blog/supply-chain) は、`go.mod` によるバージョン選択、`go.sum` と checksum database、取得・ビルド時に依存コードを自動実行しない設計を説明しています。

Russ Cox の [Open Source Supply Chain Security at Google](https://research.swtch.com/acmscored) は、再現可能な Go ツールチェーンと依存関係の対策を含む、講演動画・スライド・参考文献への入口です。ページにもあるとおり、言語などに関する講演中の意見は Google の見解ではなく講演者個人のものです。個々の仕組みを断定するときは、そこからリンクされた Go 公式資料や実装へ戻って確認します。

</details>

---

## 調査の入り口

- [Go Documentation](https://go.dev/doc)
- [Go compiler](https://go.dev/cmd/compile/)
- [Installing Go from source](https://go.dev/doc/install/source#go14)
- [Perfectly Reproducible, Verified Go Toolchains](https://go.dev/blog/rebuild)
- [Go Reproducible Build Report](https://go.dev/rebuild)
- [Running the “Reflections on Trusting Trust” Compiler](https://research.swtch.com/nih)
