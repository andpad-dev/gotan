[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [03-cmd-tools の調べ方](../../README.md)

# 600ms の計算を追え: `go tool pprof`

**実行環境**: 手元の Go 1.27 以上が必要です。

同じ計算を 100 回呼ぶテストは、CI の計測で**約 600ms**かかっていました。単体テストが固定する**出力は変えられません**。

CPU時間を実際に使っている関数と行を特定し、互換性を保った改善の根拠を作りたいです。何を採取し、どう読めばよいか調べましょう。

**`main.go`**

```go
package main

import "fmt"

var result uint64

func transform(seed uint64) uint64 {
	for range 4_000_000 {
		seed = seed*2862933555777941757 + 3037000493
	}
	return seed
}

func main() {
	fmt.Println(transform(1))
}
```

（[Go Playground で動かす](https://go.dev/play/p/8gkIRfzmhi8)）

同じコードを [main.go](./main.go)、テストを [main_test.go](./main_test.go) として置いてあります。`go 1.27` の [go.mod](./go.mod) もあります。このディレクトリで、そのまま次のコマンドを実行できます。

```console
$ go version
go version go1.27.0 darwin/arm64
$ go run main.go
15662720274509501185
```

**`main_test.go`**

このテストとbenchmarkは、`main.go` を使う複数ファイルの例です。Go Playgroundではなく、同梱ファイルをローカルで実行します。

```go
package main

import "testing"

func TestTransform(t *testing.T) {
	if got, want := transform(1), uint64(15662720274509501185); got != want {
		t.Fatalf("transform(1) = %d, want %d", got, want)
	}
}

func TestManyTransforms(t *testing.T) {
	for i := uint64(0); i < 100; i++ {
		result ^= transform(i)
	}
}

func BenchmarkTransform(b *testing.B) {
	for b.Loop() {
		result = transform(1)
	}
}
```

Go 1.27.0 / macOS（Apple M1 Max）で、テストを profile 付きで実行しました。時間とサンプル数は環境・実行ごとに変わるので、自分の出力も記録してください。

Go 1.25以降の配布物では、ビルドやテストに常用しないtoolは事前ビルドされません。`go tool` が初回に必要なtoolをソースからビルドします（[Go 1.25リリースノート](https://go.dev/doc/go1.25#go-command)）。最初の `go tool pprof` だけ少し待つ場合があります。

```console
$ go test -run '^TestManyTransforms$' -cpuprofile=cpu.out
PASS
ok  	example.com/pprof-demo	1.138s

$ go tool pprof -top cpu.out
File: pprof-demo.test
Type: cpu
Time: 2026-08-31 11:44:23 JST
Duration: 706.48ms, Total samples = 460ms (65.11%)
Showing nodes accounting for 460ms, 100% of 460ms total
      flat  flat%   sum%        cum   cum%
     430ms 93.48% 93.48%      460ms   100%  example.com/pprof-demo.transform (inline)
      30ms  6.52%   100%       30ms  6.52%  runtime.asyncPreempt
         0     0%   100%      460ms   100%  example.com/pprof-demo.TestManyTransforms
         0     0%   100%      460ms   100%  testing.tRunner
```

`cpu.out` と、同時に残った `pprof-demo.test` がそれぞれ何を持つのかを確かめましょう。そのうえで、CPU 時間がどこへ消えたかを追います。

チームで進める場合は、次の3つを分担し、最後にPRへ残す調査手順を統合しましょう。

- 採取するprofileの選択とバイナリ有無の実験
- `top`から`list`へ降りる調査
- テストとbenchmarkによる変更前後の比較

<details>
<summary>調査の入り口</summary>

まず [03-cmd-tools の調べ方](../../README.md) を開き、`go` コマンドとツールの逆引き手順を確かめます。

そのうえで、次のどれかから入ります。

- [Go の診断ツール案内](https://go.dev/doc/diagnostics) — Go の profiling / tracing / debugging ツールを俯瞰する公式ページ
- [pprof の公式ドキュメント](https://go.dev/cmd/pprof/) — `go tool pprof` の使い方の説明
- 手元の `go help testflag` と `go tool pprof -h` — それぞれのヘルプ
- [Go 同梱版 google/pprof の固定版ドキュメント](https://github.com/google/pprof/blob/92041b743c96/doc/README.md) — Go に同梱されている google/pprof の README（版を固定したリンク）
- [Go 1.27.0 の `cmd/pprof` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/pprof/doc.go) と [Go 1.27.0 の `runtime/pprof` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/runtime/pprof/proto.go) — 前者は `go tool pprof` の実装、後者は profile を書き出す側の実装
- [PGO の公式ガイド](https://go.dev/doc/pgo) — profile をコンパイラ最適化に使う PGO の公式ガイド

</details>

---

## 設問 1: まず何を採取し、何を採取していない？

`go test -cpuprofile=cpu.out` は何を作りますか。また、なぜこのコマンドの後に `pprof-demo.test` というテストバイナリも残るのでしょうか。`top` と `list` を読むために、バイナリは本当に必須でしょうか。

次の順でバイナリあり・なしの同じprofileを調べ、出力を比較してください。最後の `mv` で元に戻せます。

```console
$ go tool pprof -top pprof-demo.test cpu.out
$ go tool pprof -list='transform' pprof-demo.test cpu.out
$ mv pprof-demo.test pprof-demo.test.hidden
$ go tool pprof -top cpu.out
$ go tool pprof -list='transform' cpu.out
$ go tool pprof -raw cpu.out
$ go tool pprof -disasm='TestManyTransforms' cpu.out
$ mv pprof-demo.test.hidden pprof-demo.test
```

この問題では CPU profile を選びました。ネットワーク待ちやロック待ちが疑わしいケースでは、同じ道具を最初に使うべきでない理由も説明してください。

<details>
<summary>ヒント</summary>

- 最初は [Go の診断ツール案内](https://go.dev/doc/diagnostics) を読み、Profiling と Tracing の役割を比べます。
- 手元の `go help testflag` で `-cpuprofile` の説明と、その直後の注意書きを確認してください。
- `go tool pprof -h` で `profile.pb.gz`、`Binary`、`-raw`、`-disasm` の説明を比べましょう。
- `-raw` の出力に関数名・ファイル名・行番号が含まれるか確認してください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go の診断ツール案内](https://go.dev/doc/diagnostics) の Profiling を読み、CPU / block / mutex profile と execution trace の対象を比較する。
2. 手元で `go help testflag` を実行し、`-cpuprofile cpu.out` が終了前に CPU profile を書き、profile を生成する testing flag は coverage 以外ではテストバイナリも残すと確認する。
3. [pprof の公式ドキュメント](https://go.dev/cmd/pprof/) と `go tool pprof -h` を読み、`[binary] <source>` のようにバイナリが省略可能であることと、`-raw` / `-disasm` の役割を確認する。実装の入口は [Go 1.27.0 の `cmd/pprof` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/pprof/doc.go) で確認する。
4. 問題文のコマンドを実行し、バイナリを退避しても `-top` / `-list` が読める一方、`-disasm='TestManyTransforms'` はバイナリを開けず失敗することを確かめる。
5. `go tool pprof -raw cpu.out` に関数名・ファイル名・行番号があることを確認し、[Go 1.27.0 の `runtime/pprof.emitLocation`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/runtime/pprof/proto.go) が location と function 情報をprofile protobufへ書く処理を読む。
6. Go同梱版が利用するgoogle/pprofの [固定版ドキュメント](https://github.com/google/pprof/blob/92041b743c96/doc/README.md#details) で、`flat` と `cum` の定義を確認する。

**答え**

- `-cpuprofile=cpu.out` は、テスト実行中の CPU profile を `cpu.out` に書き出します。`go help testflag` は、coverage以外のprofileを生成するflagが、分析に使えるようテストバイナリも `pkg.test` として残すと説明しています。
- ただし、このGo 1.27.0のCPU profileは採取時にシンボル化されており、profile自体に関数名・ファイル名・行番号が入っています。手元で動かすと、バイナリを退避しても `go tool pprof -top cpu.out` と `-list='transform' cpu.out` は読めます。「Goのprofileはバイナリなしでは関数名を出せない」と一般化してはいけません。
- バイナリは不要なのではなく、機械語を読む `-disasm` や、ローカルで追加のシンボル化が必要なprofileで使われます。同梱例ではバイナリ退避後の `-disasm='TestManyTransforms'` が、profileに記録されたビルド時の一時バイナリを開けず終了コード2になります。バイナリの要否は、profileの内容と生成するreportによって変わります。
- `flat` はそのlocation自体の値、`cum` はそのlocationとすべての子孫の合計です。`transform` のように処理本体で時間を使う関数はflatが大きく、呼び出し先で時間を使う上位関数はflatが0でもcumが大きくなります。
- CPU profile は「実行中にCPUを消費した時間」を見るため、sleepやI/O待ちは目立ちません。同期プリミティブでの待機はblock profile、mutex競合はmutex profile、スケジューリング・syscall・ネットワークを含む広いレイテンシー調査はexecution traceというように、観測したい待ちへ道具を合わせます。

</details>

---

## 設問 2: 上位の関数名から、変更すべき行へ降りよう

`-top` の出力では `transform` が90%以上を占めています。しかし、関数名だけではレビューで「どの変更が根拠を持つか」を説明できません。

`transform` の行ごとの情報を出すコマンドを調べ、どの行が観測上の中心なのかを特定してください。そのうえで、「ループを短くする」案を今すぐマージできない理由を答えてください。

<details>
<summary>ヒント</summary>

- `go tool pprof -h` の output format に、ソース行を表示する形式があります。
- `-top` と同じ `cpu.out` を使えます。設問1の実験結果から、今回の `-list` にバイナリが必要かも判断してください。
- 問題文の `TestTransform` は、出力の互換性について小さな手がかりを持っています。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [pprof の公式ドキュメント](https://go.dev/cmd/pprof/) で、関数やソース行ごとのprofile表示を確認する。
2. `go tool pprof -h` で、`-list` が正規表現に一致する関数のannotated sourceを表示すると確認する。
3. Go同梱版が利用するgoogle/pprofの [固定版ドキュメント](https://github.com/google/pprof/blob/92041b743c96/doc/README.md#source-code) で、`-list` がソース行ごとのflat / cumを表示することを確認する。
4. `go tool pprof -list='transform' cpu.out` を実行し、`transform` の各行に対応するflat / cumを読む。
5. [main_test.go](./main_test.go) の `TestTransform` を読み、出力を変えられないという問題文の制約と突き合わせる。

**答え**

実行すると、次のように繰り返し本体が中心でした（時間はマシンや実行ごとに変わります。`ROUTINE` 行の絶対ファイルパスだけ省略しています）。

```console
$ go tool pprof -list='transform' cpu.out
Total: 460ms
ROUTINE ======================== example.com/pprof-demo.transform
     430ms      460ms (flat, cum)   100% of Total
         .          .      7:func transform(seed uint64) uint64 {
      70ms       70ms      8:	for range 4_000_000 {
     360ms      390ms      9:		seed = seed*2862933555777941757 + 3037000493
```

- `-top` はまず候補を狭め、`-list` はその候補をソース行へ結び付けます。この観測から、速くしたい対象は `transform` のループ本体だと説明できます。
- ただし `4_000_000` を小さくするだけでは、計算結果も変わります。`TestTransform` は `seed == 1` の結果を固定しているので、その変更は互換性を破ります。profile は「どこを調べるか」の証拠であって、「どの仕様を捨ててよいか」の許可ではありません。
- 次の調査では、出力を保つアルゴリズム上の改善案を選びます。

</details>

---

## 設問 3: 「速くなった」と、互換性を崩さずに報告するには？

変更前後の比較を、再現できる形で PR に残したいです。

この例にすでにあるテストと benchmark をどう使い分けますか。次のコマンドを実行して、何を比較対象として記録すべきかを説明してください。

```console
go test -run '^$' -bench '^BenchmarkTransform$' -benchmem -count=3
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
3. `go test -run '^$' -bench '^BenchmarkTransform$' -benchmem -count=3` を変更前に実行して基準値を残す。
4. 変更後に同じコマンドを実行し、`TestTransform` と通常のテストも通して、値の互換性と計測値を別々に比較する。

**答え**

Go 1.27.0 / macOS（Apple M1 Max）での基準値の主要部は次のとおりでした。

```console
goos: darwin
goarch: arm64
pkg: example.com/pprof-demo
cpu: Apple M1 Max
BenchmarkTransform-10	     230	   5162745 ns/op	       0 B/op	       0 allocs/op
BenchmarkTransform-10	     232	   5139714 ns/op	       0 B/op	       0 allocs/op
BenchmarkTransform-10	     230	   5173286 ns/op	       0 B/op	       0 allocs/op
```

- `TestTransform` は「同じ入力で同じ結果が出る」という互換性を守る役です。ここが落ちたら、速くてもこの変更は採用できません。
- `BenchmarkTransform` は処理時間（`ns/op`）と割り当て（`B/op`, `allocs/op`）を比較する役です。3 回の値、Go バージョン、OS / CPU、実行コマンドを変更前後で残せば、レビューで差の前提を確認できます。
- `pprof` の結果は改善候補を見つけるため、benchmark は改善の効果を測るため、単体テストは振る舞いを守るために使います。3 つを揃えることで、「600ms の処理を見つけた」だけでなく、「同じ出力のまま、どれだけ改善したか」を報告できます。

</details>

---

<details>
<summary>こぼれ話</summary>

`go tool pprof -http=localhost:0 cpu.out` を使うと、空いているローカルポートで Web UI を開けます。機械語表示も使うならバイナリを第1引数に追加します。ただし、まず `-top` と `-list` だけで仮説を小さくしておくと、CI のログやペア作業でも同じ調査を共有しやすくなります。profile は代表的な負荷で採取することが重要なので、実サービスを測るときは [PGO の公式ガイド](https://go.dev/doc/pgo) の「代表的な本番負荷」の注意も確認してください。

</details>
