[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [04-deep-dive の調べ方](../../README.md)

# Go 1.27 は、なぜ同じ浮動小数点文字列を別の方法で作るようになったのか

![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8)

計測値をテキストへ保存する処理で、`strconv.FormatFloat` の `prec = -1` を使っています。Go 1.27 への更新後も、スナップショットテストはすべて通りました。しかしレビューで「浮動小数点数の変換実装が大きく入れ替わっている」と指摘されました。

まず、次のコードで公開 API から観測できる性質を確認します。

```go
package main

import (
	"fmt"
	"math"
	"strconv"
)

func main() {
	x := 0.1
	shortest := strconv.FormatFloat(x, 'g', -1, 64)
	fixed17 := strconv.FormatFloat(x, 'g', 17, 64)

	next := math.Nextafter(1, 2)
	nextText := strconv.FormatFloat(next, 'g', -1, 64)
	shorter := nextText[:len(nextText)-1]
	roundTrip, _ := strconv.ParseFloat(nextText, 64)
	shorterValue, _ := strconv.ParseFloat(shorter, 64)

	fmt.Printf("shortest: %s\n", shortest)
	fmt.Printf("17 digits: %s\n", fixed17)
	fmt.Printf("next float: %s\n", nextText)
	fmt.Printf("round-trip: %t\n", math.Float64bits(next) == math.Float64bits(roundTrip))
	fmt.Printf("without last digit: %s (%t)\n", shorter, math.Float64bits(next) == math.Float64bits(shorterValue))
}
```

[Go Playground で実行する](https://go.dev/play/p/lFFIIE8DDSc)

`go version go1.27.0 darwin/arm64` で実行した結果です。

```text
shortest: 0.1
17 digits: 0.10000000000000001
next float: 1.0000000000000002
round-trip: true
without last digit: 1.000000000000000 (false)
```

短い `0.1` で十分な値がある一方、隣の浮動小数点数では末尾の `2` を落とすと元の bit 列へ戻れません。公開 API の契約、Go 1.26 と 1.27 の実装、採用されたアルゴリズムの順にたどります。出力を変えずに実装を替えた理由を説明しましょう。

---

## 設問 1: 「最短」は、何を満たす最短なのか？

`0.1` の 17 桁表示は `0.10000000000000001` です。しかし `prec = -1` では `0.1` だけが出力されます。一方、`1` の次に大きい `float64` は `1.0000000000000002` まで必要です。

まず `FormatFloat` と `ParseFloat` の公開ドキュメントを読んでください。「最短」がどの条件を満たす文字列なのか説明してください。

<details>
<summary>ヒント</summary>

- `FormatFloat` の `prec` の説明で、負の値に関する文を探します。
- 見た目の十進小数が元の値と同じかではなく、文字列をもう一度読み込んだ結果に注目します。
- `ParseFloat` が、二つの浮動小数点数のちょうど中間にある値をどう扱うかも確認します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から標準ライブラリの `strconv` を探し、[`FormatFloat`](https://go.dev/pkg/strconv/#FormatFloat) の `prec = -1` の説明を読む。
2. 同じページの [`ParseFloat`](https://go.dev/pkg/strconv/#ParseFloat) で、十進文字列からどの浮動小数点数を選ぶか確認する。
3. 問題文のコードで、`nextText` と、その最後の 1 桁を除いた `shorter` を読み戻した bit 列を比較する。

**答え**

ここでいう「最短」は、元の二進浮動小数点数を十進数で完全に展開した文字列の最短ではありません。`FormatFloat` のドキュメントは `prec = -1` を、**`ParseFloat` で読み戻したときに元の `f` へ正確に戻るために必要な、最小桁数**と定義しています。

そのため、二進数では厳密に `0.1` でない `float64` でも、十進文字列 `"0.1"` が同じ `float64` へ丸め戻されるなら、それ以上の桁は出しません。一方、`math.Nextafter(1, 2)` が返す `1` の隣の値では、`"1.0000000000000002"` の最後の `2` が隣接する値を区別するために必要です。問題文のコードで最後の桁を落とすと `false` になるのは、この契約を直接観測した結果です。

`ParseFloat` は、表現可能な値のうち入力に最も近いものを IEEE 754 の unbiased rounding で選びます。これは公開 API の契約です。一方、どの内部アルゴリズムで文字列を作るかは契約に含まれません。まずこの境界を分けておくことが、Go 1.27 の変更を読む前提になります。

</details>

---

## 設問 2: Go 1.26 と Go 1.27 で、どの実装が入れ替わったのか？

公開契約と実行結果は同じでも、内部では同じ処理を続けているとは限りません。対象は Go 1.26.0 と Go 1.27.0 の `internal/strconv` です。バージョンタグを固定して比較してください。

最短幅の出力、桁数を指定した出力、十進文字列の読み込みが、それぞれどの関数へ進むかを整理しましょう。

<details>
<summary>ヒント</summary>

- まず Go 1.27 のリリースノートで `strconv` と浮動小数点変換を検索し、掲載の有無を確認します。
- 二つのタグの `ftoa.go` で、`prec < 0` の分岐を比較します。
- 二つのタグの `atof.go` で、通常の十進入力を処理する高速経路のコメントを比較します。
- 片方にだけあるファイル名も、実装の役割が統合された手がかりです。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) を開き、`strconv`、`FormatFloat`、`floating` を検索する。この内部変更には個別の記載がないことを確認し、リリースノートに載っていないことだけから「変更なし」と結論しない。
2. [`strconv.FormatFloat` の公式ドキュメント](https://go.dev/pkg/strconv/#FormatFloat) から Go 本体ソースへの導線を確認する。
3. Go 1.26.0 の [`ftoa.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.0:src/internal/strconv/ftoa.go;l=109) と [`atof.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.0:src/internal/strconv/atof.go;l=578) を読む。
4. Go 1.27.0 の [`ftoa.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/ftoa.go;l=136) と [`atof.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/atof.go;l=589) を同じ分岐で比較する。
5. [変更コミット `71300e8`](https://go.dev/change/71300e80113c6ca56105aac524e9c1b0db43910f) の変更ファイル一覧で、追加・削除された実装ファイルを照合する。

**答え**

高速経路は次のように入れ替わりました。

| 処理 | Go 1.26.0 | Go 1.27.0 |
| --- | --- | --- |
| 最短幅の出力 | `dboxFtoa` による Dragonbox | `shortFloat` |
| 桁数指定の出力 | `fixedFtoa`、必要なら多倍長の `bigFtoa` | `fixedWidthFloat`、必要なら多倍長の低速経路 |
| 十進文字列の読み込み | 単純な厳密計算を試し、次に Eisel-Lemire、最後に多倍長へフォールバック | 単純な厳密計算を試し、次に `parseFloat32` / `parseFloat64`、必要な場合だけ多倍長へフォールバック |

Go 1.26.0 の `ftoa.go` には、最短幅で「Use the Dragonbox algorithm」と明記され、`dboxFtoa` を呼ぶ分岐があります。`atof.go` は通常の十進入力に Eisel-Lemire を試します。

Go 1.27.0 では、出力側の `shortFloat` と `fixedWidthFloat`、入力側の `parseFloat32` / `parseFloat64` が、共通の新しい変換部品を使います。コミットでは旧実装の `ftoadbox.go`、`ftoafixed.go`、`atofeisel.go`、`math.go` が削除され、代わりに `uscale.go` が追加されています。

`FormatFloat` と `ParseFloat` の公開コメントは、この入れ替えの前後で同じ契約を示しています。つまりこれは新しい表示形式の追加ではなく、既存の結果を保つ内部実装の交換です。問題文を Go 1.26.6 と Go 1.27.0 で実行して同じ出力になることも確認できますが、数例の一致だけで全入力の互換性を証明したことにはなりません。互換性の根拠は公開契約、テスト、レビュー済みの変更です。

</details>

---

## 設問 3: 新しい共通部品は、丸めに必要な何を残しているのか？

Go 1.27.0 では `uscale.go` が追加されました。`uscale.go` は、浮動小数点数と十進文字列の双方向変換に同じ考え方を使っています。

`research.swtch.com` の記事と実装を照合してください。途中の桁をすべて保持しなくても正しく丸められる理由と、最短文字列を選べる理由を説明してください。

<details>
<summary>ヒント</summary>

- 変更コミットの本文から、実装の背景を説明する記事へ進めます。
- 記事では、整数部分の後ろに残す二つの情報が、それぞれ何を区別するかに注目します。
- `uscale.go` で、小さな整数型に定義された丸め・除算・右シフト用のメソッドを読みます。
- 最短幅については、元の値だけでなく、隣接する二つの浮動小数点数との中間点を探します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [変更コミット `71300e8`](https://go.dev/change/71300e80113c6ca56105aac524e9c1b0db43910f) の本文を読み、そこで参照されている変換方式と記事を確認する。
2. コミット本文から、2026 年 1 月 19 日の [Floating-Point Printing and Parsing Can Be Simple And Fast](https://research.swtch.com/fp) を開く。
3. 記事上部のシリーズ名から [Floating Point Formatting](https://research.swtch.com/fp-all) へ進み、2011 年から 2026 年までの 4 本の位置づけを確認する。対象記事へ戻り、「Unrounded Numbers」「Fast, Accurate Scaling」「Shortest-Width Printing」を順に読む。
4. Go 1.27.0 の [`uscale.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/uscale.go;l=50) で `unrounded`、`shortFloat`、`uscale` を照合する。

**答え**

新しい共通部品の中心は、整数に丸め切る直前の状態を表す `unrounded` です。整数部分に加えて、下位 2 bit が次の情報を保持します。

- 1 bit は、小数部分が半分以上かどうかを表す。
- もう 1 bit は、小数部分がちょうど `0` または `1/2` ではなく、それより下にも捨てた bit があったかを表す。いったん立った後は、除算や右シフトをしても残す sticky bit になる。

この二つがあれば、捨てた全桁を保存しなくても「半分より下」「ちょうど半分」「半分より上」を区別できます。さらに整数部分の偶奇を見れば、ちょうど半分を偶数側へ寄せる丸めも決められます。実装の `round` は、その判定を次の短い式にしています。

```go
func (u unrounded) round() uint64 {
	return uint64((u + 1 + (u>>2)&1) >> 2)
}
```

`uscale` は、二進の仮数へ事前計算した 10 の冪を掛け、丸め前の情報を保ったまま必要な桁位置へ移します。出力側は `fixedWidthFloat` と `shortFloat`、入力側は `parseFloat32` と `parseFloat64` がこの結果を使います。

最短幅では、元の値の前後にある浮動小数点数との中間点から、元の値へ丸め戻される十進数の範囲を求めます。`shortFloat` はその範囲に入る最短の十進整数と 10 の冪を探します。問題文の `1.0000000000000002` は範囲内ですが、末尾を落とした `1.000000000000000` は `1` へ丸められるため範囲外です。この観測が、設問 1 の公開契約と設問 3 の内部表現をつなぎます。

</details>

---

## 設問 4: 出力が同じなのに、なぜ Go 1.27 で置き換えたのか？

ここまでで、公開契約は維持され、複数の変換経路が一つの考え方へまとめられたことが分かりました。採用コミットとコードレビューの計測を読み、チームへ変更理由と注意点を説明してください。

<details>
<summary>ヒント</summary>

- コミットメッセージで、削除行数、出力側、入力側がどう評価されているかを分けて読みます。
- ベンチマーク表は一つの列や一つのケースだけを見ず、ホストごとの差と悪化した行も確認します。
- 2026 年 1 月の記事に書かれた将来予測と、Go 1.27.0 タグに入った事実を区別します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [変更コミット `71300e8`](https://go.dev/change/71300e80113c6ca56105aac524e9c1b0db43910f) で、目的、削除されたコード、ベンチマーク表、変更ファイルを読む。
2. 同じ変更の [CL 743860](https://go.dev/cl/743860) で、レビューされた差分とテストを確認する。
3. [Go 1.27.0 タグの `uscale.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/uscale.go) にコミットの実装が含まれることを確認する。
4. 問題文のコードを Go 1.27.0 で再実行し、設問 1 の契約どおりの出力になることを照合する。

**答え**

置き換えた理由は、表示形式を変えるためではなく、**同じ公開契約を、より共通化された実装で満たしながら、主に出力処理を高速化するため**です。

コミットメッセージは、浮動小数点数の parsing と printing を記事の方式へ変更し、「almost 900 lines」を削除したと説明しています。設問 2 で見たように、従来は最短幅、固定幅、読み込みで別々の高速アルゴリズムと補助ファイルを持っていました。Go 1.27 では、丸め前の情報を保つ scaling を双方向の共通基盤にしたことで、それらをまとめています。

性能面では、コミットは printing を大きく高速化し、parsing は単純になって速度は概ね同程度だと要約しています。掲載された計測では、たとえば `AppendFloat/64Fixed12` が 6 種類のホスト設定すべてで約 23% から 29% 改善しています。ただし、別のケースには変化なしや悪化もあります。したがって「すべての入力・環境で必ず速い」とは結論できず、業務コードでは自分の入力分布と対象環境でも測る必要があります。

記事は 2026 年 1 月の時点で「この Go コードの何らかの形が Go 1.27 に入ると予想する」と述べていました。これは当時の予測であり、それだけでは採用の証拠になりません。`71300e8` の採用コミット、CL 743860、Go 1.27.0 タグのソースを照合して初めて、実際に Go 1.27 へ入ったと確定できます。

最初の違和感もここで説明できます。スナップショットが変わらないのは公開契約を維持した内部実装の交換だからです。`0.1` と `1.0000000000000002` の桁数の違いは、設問 1 の「正確に読み戻せる最短」という契約で決まり、その判定を設問 3 の `unrounded` と隣接値の範囲計算が新しい方法で実現しています。

</details>

---

## 調査の入り口

- [Go 1.27 リリースノート](https://go.dev/doc/go1.27)
- [`strconv.FormatFloat` の公式ドキュメント](https://go.dev/pkg/strconv/#FormatFloat)
- [変更コミット `71300e8`: internal/strconv: use fast unrounded scaling for floating-point](https://go.dev/change/71300e80113c6ca56105aac524e9c1b0db43910f)
- [Floating Point Formatting シリーズ](https://research.swtch.com/fp-all)
- [Go 1.27.0 タグの `internal/strconv/uscale.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/uscale.go)
