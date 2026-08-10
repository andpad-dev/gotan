[03-cmd-tools の調べ方に戻る](../../README.md)

# バッチ処理の 600ms を追え: `go tool pprof`

顧客向けのコード発行バッチでは、利用者に渡すコードを毎晩まとめて発行しています。処理開始までの余裕は短く、CI の計測ではコード発行テストだけで約 600ms かかっていました。

「ループ回数を減らせば速そう」と言う人もいますが、座席コードはすでに別システムと照合しています。出力を変えずに速くするには、まず**実際に CPU を使っている場所**を確かめなければなりません。

CPU の使用箇所を調べることをやりたいです。どういうふうにやればいいか調べよう。

**`main.go`**

```go
package main

import "fmt"

var result uint64

func seatCode(seed uint64) uint64 {
	for range 4_000_000 {
		seed = seed*2862933555777941757 + 3037000493
	}
	return seed
}

func main() {
	fmt.Println(seatCode(1))
}
```

（[Go Playground で動かす](https://go.dev/play/p/5Qwd1dKuD3I)）

```console
$ go run main.go
15662720274509501185
```

**`main_test.go`**

```go
package main

import "testing"

func TestSeatCode(t *testing.T) {
	if got, want := seatCode(1), uint64(15662720274509501185); got != want {
		t.Fatalf("seatCode(1) = %d, want %d", got, want)
	}
}

func TestIssueCodes(t *testing.T) {
	for i := uint64(0); i < 100; i++ {
		result ^= seatCode(i)
	}
}

func BenchmarkSeatCode(b *testing.B) {
	for b.Loop() {
		result = seatCode(1)
	}
}
```

Go 1.26.4 / macOS（Apple M1 Max）で、コード発行バッチのテストを profile 付きで実行しました。

```console
$ go test -run '^TestIssueCodes$' -cpuprofile=cpu.out
PASS
ok  	example.com/pprof-demo	1.190s

$ go tool pprof -top pprof-demo.test cpu.out
File: pprof-demo.test
Type: cpu
Duration: 612.42ms, Total samples = 410ms (66.95%)
Showing nodes accounting for 410ms, 100% of 410ms total
      flat  flat%   sum%        cum   cum%
     400ms 97.56% 97.56%      410ms   100%  example.com/pprof-demo.seatCode (inline)
      10ms  2.44%   100%       10ms  2.44%  runtime.asyncPreempt
         0     0%   100%      410ms   100%  example.com/pprof-demo.TestIssueCodes
         0     0%   100%      410ms   100%  testing.tRunner
```

`cpu.out` を虫眼鏡、`pprof-demo.test` を地図にして、コード発行処理の CPU 時間がどこへ消えたかを追いましょう。

---

## 設問 1: まず何を採取し、何を採取していない？

`go test -cpuprofile=cpu.out` は何を作りますか。また、なぜこのコマンドの後に `pprof-demo.test` というテストバイナリも残るのでしょうか。

この問題では CPU profile を選びました。ネットワーク待ちやロック待ちが疑わしいケースでは、同じ道具を最初に使うべきでない理由も説明してください。

<details>
<summary>ヒント</summary>

- 最初は [Go の診断ツール案内](https://go.dev/doc/diagnostics) を読み、Profiling と Tracing の役割を比べます。
- 手元の `go help testflag` で `-cpuprofile` の説明と、その直後の注意書きを確認してください。
- `go tool pprof -h` を実行し、引数に何を渡せるかを見てみましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go の診断ツール案内](https://go.dev/doc/diagnostics) の Profiling を読み、CPU profile は CPU サイクルを実際に消費している箇所を調べるものだと確認する。
2. 手元で `go help testflag` を実行し、`-cpuprofile cpu.out` が終了前に CPU profile を書き、profile を生成する testing flag は coverage 以外ではテストバイナリも残すと確認する。
3. [pprof の公式ドキュメント](https://go.dev/cmd/pprof/) と `go tool pprof -h` を読み、`go tool pprof <binary> <profile>` の形を確認する。実装の入口は [Go 1.26.4 の `cmd/pprof` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/pprof/doc.go) にある。

**答え**

- `-cpuprofile=cpu.out` は、テスト実行中の CPU profile を `cpu.out` に書き出します。`go tool pprof -top pprof-demo.test cpu.out` では、profile のアドレス情報をテストバイナリと対応付けて、関数名やソース行として読めるようにします。
- テストバイナリが残るのは、この対応付けに使えるようにするためです。`go help testflag` は、coverage 以外の profile を出す flag がテストバイナリも残すと説明しています。
- CPU profile は「実行中に CPU を消費した時間」を見るものです。待機が疑わしいのにこれだけで結論を出すと、待っている時間は目立ちません。Go の診断ツール案内も、実行トレースはレイテンシーや利用率、CPU profile は高コストなコードパスを調べる用途として区別しています。ロック・スケジューリング・I/O 待ちを追うなら、次の上級問題で使う execution trace など、現象に合う採取方法へ進みます。

</details>

---

## 設問 2: 上位の関数名から、変更すべき行へ降りよう

`-top` の出力では `seatCode` が 97% 以上を占めています。しかし、関数名だけではレビューで「どの変更が根拠を持つか」を説明できません。

`seatCode` の行ごとの情報を出すコマンドを調べ、どの行が観測上の中心なのかを特定してください。そのうえで、「ループを短くする」案を今すぐマージできない理由を答えてください。

<details>
<summary>ヒント</summary>

- `go tool pprof -h` の output format に、ソース行を表示する形式があります。
- `-top` と同じ `cpu.out`、同じテストバイナリを使えます。
- 問題文の `TestSeatCode` は、座席コードの互換性について小さな手がかりを持っています。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [pprof の公式ドキュメント](https://go.dev/cmd/pprof/) で、関数やソース行ごとの profile 表示を確認する。
2. `go tool pprof -h` で `-list` が関数に対応するソースを表示する形式だと確認する。
3. `go tool pprof -list='seatCode' pprof-demo.test cpu.out` を実行し、`seatCode` の各行に対応する flat / cumulative time を読む。
4. `main_test.go` の `TestSeatCode` を読み、コード値が既存システムとの照合に使われるという問題文の制約と突き合わせる。

**答え**

実測では、次のように繰り返し本体が中心でした（時間はマシンや実行ごとに変わります）。

```console
$ go tool pprof -list='seatCode' pprof-demo.test cpu.out
Total: 410ms
ROUTINE ======================== example.com/pprof-demo.seatCode
     400ms      410ms (flat, cum)   100% of Total
         .          .      7:func seatCode(seed uint64) uint64 {
      40ms       50ms      8:	for range 4_000_000 {
     360ms      360ms      9:		seed = seed*2862933555777941757 + 3037000493
```

- `-top` はまず候補を狭め、`-list` はその候補をソース行へ結び付けます。この観測から、速くしたい対象は `seatCode` のループ本体だと説明できます。
- ただし `4_000_000` を小さくするだけでは、計算結果も変わります。`TestSeatCode` は `seed == 1` の結果を固定しているので、その変更は既存の照合契約を破ります。profile は「どこを調べるか」の証拠であって、「どの仕様を捨ててよいか」の許可ではありません。
- 次の調査では、座席コードに必要な性質（外部システムとの互換性、衝突の扱い、必要な強度）を担当者と確認し、その性質を保つアルゴリズム上の改善案を選びます。

</details>

---

## 設問 3: 「速くなった」と、互換性を崩さずに報告するには？

バッチ処理チームは、変更前後の比較を再現できる形で PR に残したいと考えています。

この例にすでにあるテストと benchmark をどう使い分けますか。次のコマンドを実行して、何を比較対象として記録すべきかを説明してください。

```console
go test -run '^$' -bench '^BenchmarkSeatCode$' -benchmem -count=3
```

<details>
<summary>ヒント</summary>

- `go help testflag` の `-bench`、`-benchmem`、`-count` を確認します。
- correctness を確認するテストと、時間・割り当てを比べる benchmark は役目が違います。
- 単発の数値を結論にせず、同じ条件で複数回測る意図を考えてください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go の診断ツール案内](https://go.dev/doc/diagnostics) で、profile と benchmark の役割を区別する。
2. `go help testflag` で、`-bench` が benchmark を選び、`-benchmem` が割り当て統計を出し、`-count` が各 benchmark を複数回実行することを確認する。
3. `go test -run '^$' -bench '^BenchmarkSeatCode$' -benchmem -count=3` を変更前に実行して基準値を残す。
4. 変更後に同じコマンドを実行し、`TestSeatCode` と通常のテストも通して、値の互換性と計測値を別々に比較する。

**答え**

この環境での基準値は次のとおりでした。

```console
goos: darwin
goarch: arm64
pkg: example.com/pprof-demo
cpu: Apple M1 Max
BenchmarkSeatCode-10	     235	   5051027 ns/op	       0 B/op	       0 allocs/op
BenchmarkSeatCode-10	     238	   5030072 ns/op	       0 B/op	       0 allocs/op
BenchmarkSeatCode-10	     237	   5023913 ns/op	       0 B/op	       0 allocs/op
```

- `TestSeatCode` は「同じ入力で既存と同じコードが出る」という互換性を守る役です。ここが落ちたら、速くてもこの変更は採用できません。
- `BenchmarkSeatCode` は処理時間（`ns/op`）と割り当て（`B/op`, `allocs/op`）を比較する役です。3 回の値、Go バージョン、OS / CPU、実行コマンドを変更前後で残せば、レビューで差の前提を確認できます。
- `pprof` の結果は改善候補を見つけるため、benchmark は改善の効果を測るため、単体テストは振る舞いを守るために使います。3 つを揃えることで、「600ms の犯人を当てた」だけでなく、「同じ座席コードのまま、どれだけ改善したか」を報告できます。

</details>

---

<details>
<summary>こぼれ話</summary>

`go tool pprof -http=localhost:0 pprof-demo.test cpu.out` を使うと、空いているローカルポートで Web UI を開けます。ただし、まず `-top` と `-list` だけで仮説を小さくしておくと、CI のログやペア作業でも同じ調査を共有しやすくなります。profile は代表的な負荷で採取することが重要なので、実サービスを測るときは [PGO の公式ガイド](https://go.dev/doc/pgo) の「代表的な本番負荷」の注意も確認してください。

</details>

---

## 調査の入り口

1. [Go の診断ツール案内](https://go.dev/doc/diagnostics)
2. [pprof の公式ドキュメント](https://go.dev/cmd/pprof/)
3. 手元の `go help testflag` と `go tool pprof -h`
4. [Go 1.26.4 の `cmd/pprof` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/pprof/doc.go)
5. [PGO の公式ガイド](https://go.dev/doc/pgo)
