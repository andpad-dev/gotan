[進行ガイド・シナリオ一覧](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# Go 1.27 の Faster Memory Allocation は、なんで 80 バイト以下だけなの？

Go 1.27 のリリースノートを読んでいたら、「Faster Memory Allocation」という項目に、こう書いてありました。

> The compiler now generates calls to size-specialized memory allocation routines, reducing the cost of some small (<80 byte) memory allocations by up to 30%.

小さいメモリ割り当てが速くなるのは嬉しい。でも、**なんで「80 バイト以下」に限定されている**のでしょう？
128 バイトでも 256 バイトでも速くしてくれたらいいのに、と思いませんか。この `80` という数字がどこから来て、なぜそこで線を引いたのか、背景を調べてみましょう。

リリースノートの要約は `<80 byte` と書いていますが、Go 1.27.0 の実装は `size > 80` を対象外にするため、**ちょうど 80 バイトの割り当ても対象**です。この問題では実装に合わせて「80 バイト以下」と表記します。

対象になるのは、たとえば次のような「サイズがコンパイル時に確定している小さな割り当て」です。

```go
type Point struct{ X, Y int } // 16 バイト = 80 バイト以下

p := new(Point) // サイズが分かっているので size-specialized の対象になりうる
```

（Go Playground で動かす: https://go.dev/play/p/YnvY6u0c87E ）

Go 1.27.0 での出力:

```text
{1 2}, size=16 bytes
```

---

## 設問 1: そもそも「size-specialized」とは何を専用化している？ なぜコンパイル時にサイズが分かる割り当てにしか効かない？

「size-specialized memory allocation routines を呼ぶ」とは、具体的にコンパイラが何を、どんな関数の呼び出しに変えているのでしょうか。
まずは仕組みを押さえましょう。

<details>
<summary>ヒント</summary>

- リリースノートの文は「The compiler now generates calls to ...」で始まる。変更したのは **コンパイラ** 側。
- `go-review.googlesource.com` で「size specialized malloc」を検索し、`cmd/compile` のコミットを探す。
- 生成先の関数の実体は `src/runtime/_mkmalloc`（`mk...` はコード生成器の慣習）にある。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) の「Faster Memory Allocation」を読む。
2. `cmd/compile` の変更 CL [cmd/compile: call generated size-specialized malloc functions directly](https://go-review.googlesource.com/c/go/+/707856) を読む。
3. コンパイラ側の実装 [`cmd/compile/internal/ssagen/ssa.go` の `specializedMallocSym`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/ssagen/ssa.go;l=804) を読む。
4. 呼び出し先の実体を生成している [`runtime/_mkmalloc/mkmalloc.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/runtime/_mkmalloc/mkmalloc.go) を確認する。

**答え**

従来、`new(T)` のような割り当ては、汎用の `runtime.newobject` → 汎用 `mallocgc` を呼んでいました。汎用 `mallocgc` は実行時に「サイズはいくつか」「ポインタを含むか（スキャンが必要か）」「型ヘッダが要るか」といった分岐を毎回たどります。

Go 1.27 のコンパイラは、**割り当てサイズがコンパイル時に分かっている**とき、その分岐をあらかじめ潰した「サイズクラス専用の `mallocgc`」を直接呼ぶようになりました。生成される専用関数は用途別に分かれています（`runtime/_mkmalloc` が生成）。

- `mallocgcTiny…`: tiny 割り当て用
- `mallocgcSmallNoScanSC<n>`: ポインタを含まない小サイズ用
- `mallocgcSmallScanNoHeaderSC<n>`: ポインタを含むがヘッダ不要な小サイズ用

サイズやスキャン有無が定数として畳み込まれるぶん、分岐と計算が消えて速くなります（対象サイズで最大 30% 削減）。

逆に言うと、この最適化は **サイズがコンパイル時に確定している割り当てにしか効きません**。`make([]T, n)` のように要素数が実行時に決まる割り当ては、どの専用関数を呼ぶべきか静的に決められないため、従来どおり汎用パスを通ります。

</details>

---

## 設問 2: なんで 80 バイト以下に限定したの？（本題）

いよいよ本題です。`80` という数字は、コンパイラとランタイムのどこに、どんな理由で書かれているのでしょうか。
一次情報のコメントを、自分の目で確かめてみましょう。

<details>
<summary>ヒント</summary>

- 設問 1 で読んだ `specializedMallocSym` の中に `80` というリテラルがある。その行のコメントが次の手がかり。
- コメントは「This must match the constant in mkmalloc.」と言っている。つまり同じ定数がもう 1 か所にある。そちらのコメントに理由が書いてある。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) の `<80 byte` という要約を確認し、実装では境界がどう判定されるかを調べる問いを立てる。
2. コンパイラ側 [`ssa.go` の `specializedMallocSym`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/ssagen/ssa.go;l=804) で、上限を決めている定数と `size > specializedMallocMax` の条件を見つける。

   ```go
   const specializedMallocMax = 80 // This must match the constant in mkmalloc.
   if size > specializedMallocMax {
       return nil // 80 バイト超は専用化せず、汎用パスに戻す
   }
   ```

3. コメントの指示どおり、ランタイム側の同名定数 [`runtime/_mkmalloc/constants.go` の `specializedMallocMax`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/runtime/_mkmalloc/constants.go;l=25) を読む。ここに **理由そのもの** が書かれている。

**答え**

`80` の根拠は、`constants.go` のコメントが直接述べています。

> Maximum size to generate size specialized functions for.
> We've seen very limited benefit for specialized functions for larger size classes, and with the wrapper they are sometimes slower than the non-specialized functions.

要約すると、80 バイト超に広げなかった理由は次の 2 点です。

1. **効果が薄い**: 大きいサイズクラスでは専用化の恩恵がごくわずか。割り当てが大きくなるほど、`mallocgc` の分岐処理よりメモリのゼロ化などの実コストが支配的になり、分岐を潰しても全体はほとんど速くならない。
2. **むしろ遅くなることがある**: ラッパー経由になるぶん、大きいサイズでは非専用版より遅くなるケースすらある。

さらにリリースノートにあるとおり、この最適化は **バイナリサイズを約 60KB 増やす**（＝サイズクラスごとに専用関数を生成するコストがある）。効果が薄く時に逆効果な領域まで対象を広げれば、バイナリだけ膨らんで割に合いません。だから「効果がはっきり出る小さいサイズ」に絞り、その境界を `80` に置いた、というわけです。

そして `80` は **コンパイラとランタイム（`mkmalloc`）の両方で一致していなければならない** 定数です。コンパイラが「80 以下なら専用関数を呼ぶ」と判断しても、ランタイム側がその専用関数を生成していなければ辻褄が合わないため、両方のコメントが「must match」とお互いを指し示しています。

</details>

---

## 設問 3: 既定で有効なのに、なぜ `GOEXPERIMENT=nosizespecializedmalloc` という無効化スイッチがあって、しかも Go 1.28 で消える予定なの？

リリースノートには、この最適化を切るための `GOEXPERIMENT=nosizespecializedmalloc` が用意され、「Go 1.28 で削除予定」とあります。
既定で有効な最適化に、わざわざ一時的な opt-out を添える意図は何でしょうか。

<details>
<summary>ヒント</summary>

- リリースノートの当該段落を最後まで読む。「file an issue」「regressions」「expected to be removed in Go 1.28」という語がヒント。
- 設問 2 で見たトレードオフ（速度 +最大30% / 全体 ~1% と、バイナリ +60KB）と合わせて考える。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27)「Faster Memory Allocation」の該当文を読む。

   > Please file an issue if you notice any regressions. You may set `GOEXPERIMENT=nosizespecializedmalloc` at build time to disable it. This opt-out setting is expected to be removed in Go 1.28.

**答え**

これは新しいコード生成に対する **一時的な安全弁（緊急脱出口）** です。

- 専用関数呼び出しは新規のコード生成なので、特定のワークロードやプラットフォームで性能・バイナリサイズのリグレッションを引き起こす可能性がゼロではない。もし踏んでしまった人がいても、`GOEXPERIMENT=nosizespecializedmalloc` を付ければ即座に Go 1.26 相当の挙動へ戻せる。
- 得られる効果は「割り当てヘビーな実プログラムで全体 ~1%」で、代わりにバイナリが約 60KB 増える、というトレードオフ。既定 ON にする価値はあると判断されつつも、影響を受ける人向けの退避手段を 1 バージョンだけ残した。
- `expected to be removed in Go 1.28` なのは、これが experiment（実験フラグ）扱いの経過措置だから。1.27 で広く使われて問題が出ないことが確認できれば、スイッチごと畳んで通常機能に昇格する、という段取りです。

つまり「既定 ON ＋ 1 バージョン限りの opt-out」は、新しい最適化を安全に本流へ取り込むための、Go チームの定番の進め方だと読み取れます。

</details>

---

<details>
<summary>こぼれ話: ランタイムは当初「512 バイトまで」専用関数を作ろうとしていた</summary>

面白いことに、専用関数を生成するランタイム側の最初の CL [runtime: add specialized malloc functions for sizes up to 512 bytes](https://go.googlesource.com/go/+show/411c250d64304033181c46413a6e9381e8fe9b82) は、**512 バイト** まで作ろうとしていました。コミットメッセージにはこうあります。

> That's the limit where it's possible to end up in the no header case when there are scan bits, and where the benefits of the specialized functions significantly diminish according to microbenchmarks.

`512` は「スキャンビットがあってもヘッダ無し（no header）で済む構造上の上限」であり、かつ「マイクロベンチで効果が大きく減り始める境界」でした。

一方、最終的に **実際に使う** カットオフは `specializedMallocMax = 80` に落ち着きました。`512` は「作ろうと思えば作れる上限」、`80` は「作る価値がある上限」——両者の差が、そのまま「専用化できる」と「専用化して得する」の差になっている、というわけです。

</details>

---

## 調査の入り口

- [Go 1.27 リリースノート「Faster Memory Allocation」](https://go.dev/doc/go1.27)
- CL: [cmd/compile: call generated size-specialized malloc functions directly](https://go-review.googlesource.com/c/go/+/707856)
- ソース: [`runtime/_mkmalloc/constants.go`（`specializedMallocMax = 80` の理由コメント）](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/runtime/_mkmalloc/constants.go;l=25)
- ソース: [`cmd/compile/internal/ssagen/ssa.go`（`specializedMallocSym`）](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/ssagen/ssa.go;l=804)
