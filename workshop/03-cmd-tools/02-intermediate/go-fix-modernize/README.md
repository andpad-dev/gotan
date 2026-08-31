# go fix のモダナイザで既存コードを最新イディオムに寄せよう

チームの Go コードを最新イディオムに寄せていきたいです。
Go 1.26 で `go fix` が刷新されたと聞きました。
どういうふうに使えばいいか、`go vet` との違いや自作 API 移行への応用も含めて調べましょう。

たとえば、次のような「いかにも古い書き方」のコードがあるとします（[Playground](https://go.dev/play/p/6pcJuZQr7_0)）。

```go
package main

import (
	"fmt"
	"sort"
)

func main() {
	xs := []int{3, 1, 2}
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
	for i := 0; i < len(xs); i++ {
		fmt.Println(i, xs[i])
	}
}
```

同じコードを [main.go](./main.go) として、`go 1.27` の [go.mod](./go.mod) と一緒に置いてあります。このディレクトリで次を実行すると、コードを手で保存し直さずに観測できます。

```console
$ go version
go version go1.27.0 darwin/arm64
$ go run .
0 1
1 2
2 3
$ go fix -diff ./...
```

このコードに対して `go fix -diff ./...` を叩くと、こんな diff が返ってきます。

```diff
 import (
 	"fmt"
-	"sort"
+	"slices"
 )
 
 func main() {
 	xs := []int{3, 1, 2}
-	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
+	slices.Sort(xs)
-	for i := 0; i < len(xs); i++ {
+	for i := range xs {
 		fmt.Println(i, xs[i])
 	}
 }
```

適用後はこうなります（[Playground](https://go.dev/play/p/cx7XCG7OLqB)）。

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	xs := []int{3, 1, 2}
	slices.Sort(xs)
	for i := range xs {
		fmt.Println(i, xs[i])
	}
}
```

`-diff` は変更を適用せず差分を表示し、差分が空でないときは終了コードが非 0 になります。この例で終了コード 1 になるのは想定どおりです。

この `go fix` は一体何者で、どこまで面倒を見てくれるのでしょうか。

チームで進める場合は、「Go 1.26/1.27 の変更と `go vet` との違い」「CLIでの一覧・絞り込み・diffの実測」「inline例とGo本体ソース」の3担当に分かれ、最後に安全に適用する手順を一つにまとめてみましょう。

---

## 設問 1: 新しい `go fix` の位置づけを調べよう

Go 1.26 で刷新された `go fix` は何をするコマンドで、旧 `go fix`（Go 1.0 時代からある fixers）や `go vet` とはどう違うのでしょうか？

<details>
<summary>ヒント</summary>

- [Go 1.26 リリースノート](https://go.dev/doc/go1.26#go-command) の Tools 節で `go fix` を検索する
- [cmd/fix](https://pkg.go.dev/cmd/fix) と [cmd/vet](https://pkg.go.dev/cmd/vet) の Overview を並べて読む
- [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis) の Overview で「checker」と「fixer」の関係を確認する

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.26 リリースノート](https://go.dev/doc/go1.26#go-command) の Tools セクションで `go fix` の項目を読む。
2. [cmd/fix](https://pkg.go.dev/cmd/fix) の Overview を読み、[cmd/vet](https://pkg.go.dev/cmd/vet) と比較する。
3. 解説記事「[Using go fix to modernize Go code](https://go.dev/blog/gofix)」の "The Go analysis framework" 節で背景を確認する。

**答え**

- Go 1.26 で `go fix` は完全に書き直され、**`go vet` と同じ analysis framework の上に載る fixer 群**を実行するコマンドになりました。それまで Go 1.0 以前の言語仕様変更に対応するために存在していた historical な fixers はすべて削除されています。
- `go vet` と `go fix` の基盤は同じで、違うのは扱うアナライザの種別と結果の扱いだけです。
  - `go vet`: **checker** を実行し、「怪しい構造（バグの可能性）」を診断として報告する。
  - `go fix`: **fixer** を実行し、「安全に置き換え可能な改善」を計算してそのままソースに適用する（`-diff` を付ければ diff 表示にとどめる）。
- 中心となる fixer 群は **modernizers**（[`go/analysis/passes/modernize`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize)）で、`sort.Slice` → `slices.Sort`、3 節形式の `for` → `for range`、`interface{}` → `any` など、新しい言語機能・標準ライブラリ機能で書き直せる箇所を検出して置き換えます。`go fix` はこのパッケージのうち、安全に一括適用できるものを選んだ suite を実行します。
- 加えて、`//go:fix inline` ディレクティブで印を付けた関数・定数・型エイリアスをインライン展開する **inline** アナライザも含まれています（設問 3 で扱います）。

</details>

---

## 設問 2: modernizer を確認・選択して適用する手順を調べよう

チームのコードにいきなり `go fix ./...` を叩くのは怖いので、
「今どの modernizer が動くのか」「どんな置き換えを提案してくるのか」を確認してから、必要なものだけ適用したいです。
どういう手順を踏めばよいでしょうか？

<details>
<summary>ヒント</summary>

- 一覧は `go tool fix help` で見られる
- 個別 fix の詳細は `go tool fix help <analyzer名>`
- 手元で試すときは `go fix -diff ./...`（適用せずに unified diff を表示）
- 個別の fix だけを試したいときは analyzer 名の付いたフラグを使う（`-any` / `-slicessort` など）

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノートの `go fix` 節](https://go.dev/doc/go1.27#go-fix) を読み、現行版で追加・削除・改名された analyzer 名を確認する。
2. 手元で `go help fix` を実行し、`-diff` の動作と、fixer の説明を見る次のコマンドを確認する。
3. `go tool fix help` を実行し、`Registered analyzers:` に表示された現在の一覧を記録する。Go 1.27.0 では 26 個だが、版によって変わるため自分の出力を正とする。
4. 興味のあるものは `go tool fix help <name>` で詳細を読む。たとえば `go tool fix help slicessort` で、対象コードと必要な Go バージョンを確認する。
5. [`go/analysis/passes/modernize`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize) で modernizer 全体の説明を読み、実際に `go fix` が登録する集合は [Go 1.27.0 の fix suite](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/vendor/golang.org/x/tools/go/analysis/suite/fix/fix.go) とCLI出力で照合する。パッケージが提供する analyzer と `go fix` が登録する analyzer を同一視しない。

**答え**

手順はざっくり次の通りです。

1. **一覧を見る**: `go tool fix help` の `Registered analyzers:` で登録名を確認する。Go 1.27.0 では 26 個で、`any` / `minmax` / `rangeint` / `slicessort` / `stringscut` / `inline` などが含まれる。Go 1.27 では `fmtappendf` が `go fix` の登録対象から削除され、`waitgroup` は `waitgroupgo` に改名されたため、固定した古い一覧を使わない。
2. **個別詳細を見る**: `go tool fix help <name>`。たとえば `go tool fix help slicessort` を叩くと、`sort.Slice` をどんなときに `slices.Sort` に置き換えるか、どの Go バージョン以降のファイルに適用されるかまで説明が出る。
3. **プレビュー**: `go fix -diff ./...` で「適用せず unified diff だけ見る」ことができる。CI やレビュー前の下見はこれで十分。
4. **絞り込み**: すべての analyzer を一気に回さず、たとえば `slicessort` だけプレビューしたいなら `go fix -slicessort -diff ./...`。逆にひとつだけ止めて全体をプレビューするなら `go fix -slicessort=false -diff ./...` とする。`-analyzer=slicessort` という共通フラグは存在しない。
5. **適用**: 問題なさそうなら `go fix ./...`。**クリーンな git 状態から始める**のがブログでも推奨されていて、`go fix` によるコミットとレビュー対象コミットを分けやすくなります。
6. modernizer 群の公式パッケージ文書は [`go/analysis/passes/modernize`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize) にある。ただし、そこに載るすべての analyzer が同じ版の `go fix` に登録されるとは限らない。実行中のコマンドの対象は `go tool fix help` で確定する。

なお `go fix` は generated file（`// Code generated ...` を含むファイル）には触りません。生成コードを更新したいときは generator 側を直します。

</details>

---

## 設問 3: `//go:fix inline` で自作 API の移行を自動化しよう

社内ライブラリで非推奨にした関数の呼び出しを、新しい関数に順次置き換えていきたいです。
呼び出し側に手作業で頼んで回るのは辛いので、`go fix` に自動でやってほしいのですが、どう書けば実現できますか？
どんな制約や注意点があるでしょうか？

特に「自分自身のテストでは置き換えない」という条件を具体的に判別するため、[inline-example/legacy.go](./inline-example/legacy.go)、[inline-example/legacy_test.go](./inline-example/legacy_test.go)、[inline-example/go.mod](./inline-example/go.mod) を用意しました。Go Playground では複数ファイルに対する `go fix` を実行できないため、この例はローカルで確認します。

```console
$ cd inline-example
$ go version
go version go1.27.0 darwin/arm64
$ go test ./...
$ go fix -inline -diff ./...
```

`TestHello`、`BenchmarkHello`、`BenchHello`、`ExampleHello`、`TestSomethingElse` のうち、どの関数内の `Hello()` が残るかを予想してから diff と比較してください。
`BenchHello` は除外判定を観測するための関数名であり、`go test` がベンチマークとして自動実行する標準の `BenchmarkXxx` 形式ではありません。

<details>
<summary>ヒント</summary>

- 手元でまず `go tool fix help inline` を実行して概要をつかむ
- 一次情報は [inline analyzer のドキュメント](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inline)
- `Deprecated:` コメントとの組み合わせ、`const` / `type` の制約、`var params = args` の binding declaration をドキュメントとGo 1.27ソースで確認する

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [解説記事「Using go fix to modernize Go code」](https://go.dev/blog/gofix) の "self-service" パラダイム節を読み、`//go:fix inline` が解決する移行問題を確認する。
2. `go tool fix help inline` を手元で実行し、対象、binding declaration、専用フラグ、専用テストの説明を読む。
3. [inline analyzer のドキュメント](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inline) を読み、関数・定数・型エイリアスごとの制約を確認する。公開パッケージの文書はツールチェーン同梱版より新しい場合があるため、対象版の実装とも照合する。
4. 同梱例で `go fix -inline -diff ./...` を実行し、専用テストの名前ごとの差を記録する。文書だけでは `BenchmarkHello` と `BenchHello` の違いを確定できないため、[Go 1.27.0 の `withinTestOf`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/vendor/golang.org/x/tools/go/analysis/passes/inline/inline.go) を読み、実測と照合する。

**答え**

- 移行元となる関数・定数・型エイリアスの**直前**に `//go:fix inline` コメントを書きます。

    ```go
    // Deprecated: Pow(x, 2) を直接使ってください。
    //go:fix inline
    func Square(x int) int { return Pow(x, 2) }
    ```

    こう書いておくと、`go fix -inline ./...`（あるいは単に `go fix ./...`）で**呼び出し側のコード上で本体にインライン展開**されます。パッケージをまたいでも動きます。
- 用途は「非推奨関数から新関数への移行」「別パッケージへの引っ越し（メジャーバージョンアップ時など）」で、`Deprecated:` コメントと**併記**するのが定番パターンです。旧 API を新 API を呼ぶ薄いラッパーとして残しておけば、呼び出し側は `go fix` するだけで移行できます。
- **定数**に付ける場合は、右辺が「別の名前付き定数」でなければなりません（リテラル値ではダメ）。定数グループの中の 1 つだけに付けることも、グループ全体に付けることもできます。
- **型エイリアス**（`type A = newpkg.A`）にも付けられ、参照側の `A` が `newpkg.A` に置き換わります。Go 1.27.0 では [`HandleAlias` と `inlineAlias`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/vendor/golang.org/x/tools/go/analysis/passes/inline/inline.go) がこの処理を担います。
- inliner は挙動を変えないよう慎重で、引数式の評価順序を保つために `var params = args` という **binding declaration** を差し込むことがあります。スタイル上邪魔なときは `-inline.allow_binding_decl=false` を渡すと、そのケースの fix はスキップされます。
- 関数 `X` の呼び出しを残す「専用テスト」は、同じパッケージ（外部テストパッケージの `_test` suffix も許容）の `_test.go` にある、`TestX` / `ExampleX` / `BenchX`、またはそれらに `_suffix` が続くトップレベル関数です。判定は元ファイルとの対応ではなく、テストファイル・パッケージ・関数名で行われます。
- Go 1.27.0 の実装が確認する接頭辞は `Test` / `Example` / `Bench` です。したがって同梱例では `TestHello`、`TestHello_withSuffix`、`ExampleHello`、`BenchHello` の呼び出しは残りますが、`BenchmarkHello` と `TestSomethingElse` は置き換わります。通常のベンチマーク名 `BenchmarkX` は除外されないため、名前から安全だと推測せず使用中の版で確認してください。

</details>

---

<details>
<summary>こぼれ話</summary>

- Go チームは今後、外部モジュール作者が自作の analyzer を配布し、利用者側の `go fix` / gopls で動的に読み込む "self-service" 方式を検討しています（[go issue #59869](https://go.dev/issue/59869)）。`//go:fix inline` はその方向性の最初の一歩に位置付けられています（出典: [解説記事](https://go.dev/blog/gofix)）。
- `go fix` は synergistic な適用にも配慮されていて、複数の fix が同じ場所で衝突した場合はスキップして警告を出す挙動になっています。「1 回で全部直しきれない」場合は、もう一度 `go fix ./...` を叩くと fixed point に到達しやすい、というのがブログの推奨です。
- `strings.Builder` への書き換えを提案する `stringsbuilder` のように、ループ内の文字列結合による**準・DoS 級のパフォーマンス問題**を潰す fix もあります。旧コードのパフォーマンス改善カード切りにも使えます。

</details>

---

## 調査の入り口

- 手元コマンド: `go help fix` / `go tool fix help` / `go tool fix help <analyzer>`
- リリースノート: [Go 1.26 の刷新](https://go.dev/doc/go1.26#go-command) / [Go 1.27 の analyzer 変更](https://go.dev/doc/go1.27#go-fix)
- コマンドのドキュメント: [cmd/fix](https://pkg.go.dev/cmd/fix) / [cmd/vet](https://pkg.go.dev/cmd/vet) / [cmd/go の該当節](https://pkg.go.dev/cmd/go#hdr-Apply_fixes_suggested_by_static_checkers)
- Modernizer の一次情報: [modernize パッケージ](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/modernize) / [Go 1.27.0 の fix suite](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/vendor/golang.org/x/tools/go/analysis/suite/fix/fix.go) / [inline analyzer](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inline)
- 基盤: [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- 解説記事: [Using go fix to modernize Go code](https://go.dev/blog/gofix)
