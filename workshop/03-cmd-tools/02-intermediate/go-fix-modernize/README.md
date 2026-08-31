[進行ガイド・シナリオ一覧](../../../README.md) | [03-cmd-tools の調べ方](../../README.md)

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

実行するとこう出ます。

```
0 1
1 2
2 3
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

この `go fix` は一体何者で、どこまで面倒を見てくれるのでしょうか。

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
- 中心となる fixer 群は **modernizers**（[gopls の modernize パッケージ](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize)）で、`sort.Slice` → `slices.Sort`、3 節形式の `for` → `for range`、`interface{}` → `any` など、新しい言語機能・標準ライブラリ機能で書き直せる箇所を検出して置き換えます。
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

1. 手元で `go help fix` を実行し、`-diff` フラグや analyzer 名フラグの説明を読む。
2. `go tool fix help` で登録済み analyzer 一覧を眺め、興味のあるものは `go tool fix help <name>` で詳細を読む。
3. [gopls の modernize パッケージ](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize) のドキュメントで modernizer 全体の狙いを確認し、必要ならリンク先の [ソースコード](https://cs.opensource.google/go/x/tools/+/master:gopls/internal/analysis/modernize/) を追う。

**答え**

手順はざっくり次の通りです。

1. **一覧を見る**: `go tool fix help` で登録された analyzer 名（`any` / `minmax` / `rangeint` / `slicessort` / `stringscut` / `inline` など）を確認する。
2. **個別詳細を見る**: `go tool fix help <name>`。たとえば `go tool fix help slicessort` を叩くと、`sort.Slice` をどんなときに `slices.Sort` に置き換えるか、どの Go バージョン以降のファイルに適用されるかまで説明が出る。
3. **プレビュー**: `go fix -diff ./...` で「適用せず unified diff だけ見る」ことができる。CI やレビュー前の下見はこれで十分。
4. **絞り込み**: すべての analyzer を一気に回さず、たとえば `slicessort` だけ試したいなら `go fix -slicessort ./...`。逆にひとつだけ止めたければ `-slicessort=false` と否定形で指定する（`go build`・`go vet` と同じ流儀）。
5. **適用**: 問題なさそうなら `go fix ./...`。**クリーンな git 状態から始める**のがブログでも推奨されていて、`go fix` によるコミットとレビュー対象コミットを分けやすくなります。
6. modernizer 群の実装本体は [gopls の modernize パッケージ](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize) にあり、gopls（VS Code など）でも同じ analyzer が走ります。エディタ上で電球表示されるヒントは `go fix` で一括適用できる、というのが基本の関係です。

なお `go fix` は generated file（`// Code generated ...` を含むファイル）には触りません。生成コードを更新したいときは generator 側を直します。

</details>

---

## 設問 3: `//go:fix inline` で自作 API の移行を自動化しよう

社内ライブラリで非推奨にした関数の呼び出しを、新しい関数に順次置き換えていきたいです。
呼び出し側に手作業で頼んで回るのは辛いので、`go fix` に自動でやってほしいのですが、どう書けば実現できますか？
どんな制約や注意点があるでしょうか？

<details>
<summary>ヒント</summary>

- 手元でまず `go tool fix help inline` を実行して概要をつかむ
- 一次情報は [inline analyzer のドキュメント](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inline)
- `Deprecated:` コメントと組み合わせる例、`const` / `type` に付ける場合の制約、`var params = args` の binding declaration の話がドキュメントに書かれている

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. `go tool fix help inline` を手元で実行し、`//go:fix inline` の振る舞いを確認する。
2. [inline analyzer のドキュメント](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inline) を読み、対象（関数・定数・型エイリアス）ごとの制約と、`-inline.allow_binding_decl=false` などのフラグまで押さえる。
3. [解説記事](https://go.dev/blog/gofix) の "self-service" パラダイム節で、なぜこのディレクティブが導入されたのかという背景を確認する。

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
- **型エイリアス**（`type A = newpkg.A`）にも付けられ、参照側の `A` が `newpkg.A` に置き換わります。
- inliner は挙動を変えないよう慎重で、引数式の評価順序を保つために `var params = args` という **binding declaration** を差し込むことがあります。スタイル上邪魔なときは `-inline.allow_binding_decl=false` を渡すと、そのケースの fix はスキップされます。
- 対象シンボル `X` の**自分自身のテスト**（`TestX` / `BenchmarkX` / `ExampleX`、または `foo.go` に対する `foo_test.go` 内の参照）はあえて inline されません。旧 API 自体の挙動を守るためです。

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
- リリースノート: [Go 1.26 リリースノート #go-command](https://go.dev/doc/go1.26#go-command)
- コマンドのドキュメント: [cmd/fix](https://pkg.go.dev/cmd/fix) / [cmd/vet](https://pkg.go.dev/cmd/vet) / [cmd/go の該当節](https://pkg.go.dev/cmd/go#hdr-Apply_fixes_suggested_by_static_checkers)
- Modernizer の一次情報: [gopls modernize パッケージ](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize) / [inline analyzer](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inline)
- 基盤: [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- 解説記事: [Using go fix to modernize Go code](https://go.dev/blog/gofix)
