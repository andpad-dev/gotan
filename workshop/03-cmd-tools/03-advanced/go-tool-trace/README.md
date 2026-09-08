[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [03-cmd-tools の調べ方](../../README.md)

# 100ms の同期待ちを追え: `go tool trace`

![実行環境: 手元の Go](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-%E6%89%8B%E5%85%83%E3%81%AE%20Go-F39C12)

このディレクトリに `go.mod`、`main.go`、`main_test.go` を同梱しています。ローカルではこのディレクトリでコマンドを実行してください。

テストでは、同じ処理を 4 つ並行に実行します。1 回 25ms のはずが、全体は約 100ms かかります。CPU 使用率は低く、CPU profile を開いても「何が待たせたか」は分かりません。

遅い実行を検知した**後**に、直前の状況を調べたい場面もあります。

この仕組みはなんでこうなってるの？背景を調べよう。

**`main.go`**

```go
package main

import (
	"context"
	"fmt"
	"runtime/trace"
	"sync"
	"time"
)

func processWork(ctx context.Context) {
	var lock sync.Mutex
	var wg sync.WaitGroup

	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			trace.WithRegion(ctx, "step", func() {
				lock.Lock()
				defer lock.Unlock()
				time.Sleep(25 * time.Millisecond)
			})
		}()
	}
	wg.Wait()
}

func main() {
	processWork(context.Background())
	fmt.Println("work completed")
}
```

（[Go Playground で動かす](https://go.dev/play/p/Q2woqsOfJ2y)）

```console
$ go run main.go
work completed
```

テストは、4 つの処理を `work` タスクに結びます。

**`main_test.go`**

```go
package main

import (
	"context"
	"runtime/trace"
	"testing"
)

func TestProcessWork(t *testing.T) {
	ctx, task := trace.NewTask(context.Background(), "work")
	defer task.End()

	processWork(ctx)
}
```

Go 1.27.0 / macOS で `go test -trace` を実行すると、テストは成功します。所要時間は実行環境によって変わります。ただし、処理が並列化されているとは限りません。

```console
$ go test -run '^TestProcessWork$' -trace=work.trace
PASS
ok  	example.com/trace-demo	0.853s
```

---

## 設問 1: CPU が暇そうなのに、なぜ trace を採る？

同じ処理を `go` 文で 4 つ起動しています。ならば 25ms 前後で終わりそうなのに、テストは約 100ms です。

最初に CPU profile だけで調べるのが不十分な理由を、一次情報から説明してください。`go test -trace=work.trace` が観測できる事実も説明してください。

まず [Go の診断ツール案内](https://go.dev/doc/diagnostics) で現在の公式な契約を確認します。次に Russ Cox の [How To Build a User-Level CPU Profiler](https://research.swtch.com/pprof) を読みます。記事は pprof のサンプリング設計を 2013 年時点で解説したものです。実装詳細を現在のランタイム仕様として扱わないでください。周期的に得たスタックの標本を数える考え方が、待機時間の調査にどんな限界を持つかを整理してください。

<details>
<summary>ヒント</summary>

- まず [Go の診断ツール案内](https://go.dev/doc/diagnostics) で Profiling と Execution tracer を見比べます。
- research.swtch.com の記事では「Profiling with pprof」と「Interpreting the data」を読み、CPU が動いている瞬間のスタック標本と、runtime event の時系列を比べます。
- `go help testflag` の `-trace` を確認しましょう。
- trace ファイルを作れる経路は [trace の公式ドキュメント](https://go.dev/cmd/trace/) にも載っています。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go の診断ツール案内](https://go.dev/doc/diagnostics) を読み、CPU profile は CPU サイクルを実際に消費している時間を、execution tracer はレイテンシー・利用率・goroutine の動きを調べるものだと区別する。
2. [How To Build a User-Level CPU Profiler](https://research.swtch.com/pprof) の「Profiling with pprof」と「Interpreting the data」を読み、周期的に取得したスタックトレースごとの観測回数が profile の基礎になることを確認する。記事は 2013 年の実装解説なので、固定サイズの表などの詳細は現在実装の根拠にせず、設計の説明として読む。
3. 手元で `go help testflag` を実行し、`-trace trace.out` がテスト終了前に execution trace をファイルへ書くことを確認する。
4. [trace の公式ドキュメント](https://go.dev/cmd/trace/) を読み、`go test -trace`、`runtime/trace.Start`、`net/http/pprof` が trace ファイルの生成経路であることを確認する。続けて [Go 1.27.0 の `cmd/trace` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/trace/doc.go) でも同じ用途と profile type を裏取りする。

**答え**

- CPU profile は、CPU サイクルを使った場所を見つけるのに向いています。公式案内も、sleep や I/O 待ちではなく、CPU を実際に消費している時間を示すものだと区別しています。research.swtch.com の記事が説明するように、CPU profile は実行中に周期的に得たスタックの標本を数えます。そのため、ロックでブロックされて CPU を使っていない goroutine の待機時間は、CPU profile だけでは主役になりません。
- execution trace は goroutine の生成・ブロック・解除、スケジューリング、syscall、GC、ヒープサイズなどの runtime event を時系列で記録します。そのため「4 人がいつ走れ、いつ止まり、どれだけ直列化されたか」を観測できます。
- `go test -trace=work.trace` は、再現テストを走らせながらその時系列の記録を残します。ここでは「100ms だからロック競合だ」と決めつけず、まず trace を採ることが次の検索の手がかりになります。

</details>

---

## 設問 2: trace を「待ち時間の証拠」に変えよう

`work.trace` をブラウザで開く前に、同期待ちだけを pprof 形式へ取り出せます。どのコマンドをつなげればよいでしょうか。

実行した結果から、どの行で処理が待っているかを特定してください。さらに、`trace.NewTask` と `trace.WithRegion` が足す手がかりを説明してください。手がかりがないと、4 本の goroutine は匿名な棒グラフのままです。

<details>
<summary>ヒント</summary>

- [trace の公式ドキュメント](https://go.dev/cmd/trace/) には、trace から複数種類の pprof-like profile を出す書式があります。
- `sync` という profile type が何を表すかを確認しましょう。
- `go tool pprof -h` の `-list` と、`go doc runtime/trace` を続けて調べてください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [trace の公式ドキュメント](https://go.dev/cmd/trace/) を読み、`sync` が synchronization blocking profile であり、`go tool trace -pprof=sync` で pprof-like profile を出せることを確認する。
2. 次を実行する。

    ```console
    go tool trace -pprof=sync work.trace > sync.pprof
    go tool pprof -top sync.pprof
    go tool pprof -list='processWork.func1.1' sync.pprof
    ```

3. `go doc runtime/trace` を読み、[Go 1.27.0 の `runtime/trace` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/runtime/trace/annotation.go) で `NewTask` と `WithRegion` の説明・制約を確認する。

**答え**

今回の trace から抽出した `sync.pprof` の一部です（数値は実行環境により変わります）。

```console
$ go tool pprof -top sync.pprof
Type: delay
Showing nodes accounting for 464.38ms, 100% of 464.38ms total
      flat  flat%   sum%        cum   cum%
  204.65ms 44.07% 44.07%   204.65ms 44.07%  runtime.chanrecv1
  156.27ms 33.65% 77.72%   156.27ms 33.65%  sync.(*Mutex).Lock
  103.46ms 22.28%   100%   103.46ms 22.28%  sync.(*WaitGroup).Wait
```

`-list` で該当箇所まで降りると、待ち時間は鍵を取る行に対応します。

```console
$ go tool pprof -list='processWork.func1.1' sync.pprof
ROUTINE ======================== example.com/trace-demo.processWork.func1.1
         .   156.27ms     19:	trace.WithRegion(ctx, "step", func() {
         .   156.27ms     20:		lock.Lock()
         .          .     21:		defer lock.Unlock()
```

- `sync.(*Mutex).Lock` の待ち時間と、4 回の 25ms がほぼ直列に積み上がる観測から、処理完了まで約 100ms かかった主因は共有 `lock` の競合です。`go` 文が 4 本あることは、同時にロックを取得できることを意味しません。
- `trace.NewTask` は `work` という論理的な処理を `context.Context` に載せます。`trace.WithRegion` は各 goroutine 内の `step` 区間を、その task に結び付けます。実際、`go tool trace -d=parsed work.trace` には `TaskBegin Type="work"` と 4 回の `RegionBegin Type="step"` が出ます。
- task / region の型は無制限に増やす名前ではなく、分析で分類するための少数の種類に保つのが API の意図です。ここでは、runtime の待ち時間を task と region に対応付け、次の設計判断に進めます。

</details>

---

## 設問 3: なぜ「遅くなってから trace を採る」では手遅れなのか？

遅い実行を検知してから trace を取り始める案があります。しかし、検知した時点では、原因となった待ち時間はすでに過去です。

execution tracer が Go 1.21〜1.22 でどう変わったかを追ってください。flight recording がこの課題にどう答えるかを説明してください。あわせて、次の主張が正しいとは言えない理由も答えてください。

> 新しい trace なら、`go tool trace` は巨大なファイルを一切メモリに載せない。

<details>
<summary>ヒント</summary>

- [Go 1.21 リリースノート](https://go.dev/doc/go1.21) で `runtime/trace` を検索します。
- 次に [Go 1.22 リリースノート](https://go.dev/doc/go1.22) の Trace と `runtime/trace` を読みます。
- そこで出てくる「partition」を手がかりに、[execution tracer overhaul の設計文書](https://go.googlesource.com/proposal/+/refs/heads/master/design/60773-execution-tracer-overhaul.md) と [flight recording の Issue #63185](https://github.com/golang/go/issues/63185) の議論を辿ってください。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.21 リリースノート](https://go.dev/doc/go1.21) の `runtime/trace` を読み、amd64 / arm64 での trace 採取コストが大幅に下がったことを確認する。
2. [Go 1.22 リリースノート](https://go.dev/doc/go1.22) の `runtime/trace` と Trace を読み、OS clock、partition、syscall の完全な期間、thread-oriented view、開始・終了レイテンシーの改善を確認する。
3. [execution tracer overhaul の設計文書](https://go.googlesource.com/proposal/+/refs/heads/master/design/60773-execution-tracer-overhaul.md) の Background / Goals を読み、従来は解析時のメモリ要求やストリーミング不能さが課題だったことを確認する。
4. [Issue #63185](https://github.com/golang/go/issues/63185) を冒頭から読み、self-contained partition を直近の移動窓として保持し、必要になった時点で snapshot する flight recording の提案を追う。さらに [Go 公式ブログの解説](https://go.dev/blog/execution-traces-2024) で採用後の説明を確認する。

**答え**

- 遅い実行を検知してから trace を開始しても、`lock.Lock` の競合が起きた時間帯は既に記録されていません。だから「発生後に採る」だけでは原因へ遡れません。
- Go 1.21 では execution trace の採取コストが amd64 / arm64 で大幅に下がり、Go 1.22 では trace 実装が全面的に作り直されました。trace は自己完結した partition に分かれ、開始・終了の影響も減り、ストリームとして処理できる土台ができました。設計文書が目標に置いたのは、解析メモリの削減、ストリーミング、古い実装上の問題の解消です。
- Issue #63185 では、この partition を少なくとも 1 つ保持すれば、直近の時間窓を snapshot できると提案されました。これが flight recording です。低い採取コストと partition が揃ったため、異常を検知した**後**でも直前の証拠を保存できるようになりました。
- ただし「trace 形式が streamable になった」ことと「`go tool trace` がすでに巨大 trace を全く読み込まない」ことは別です。公式ブログは、Go 1.22+ の trace ではその改善が可能になった一方、`go tool trace` 自体はまだ trace 全体をメモリに載せると明記しています。よって運用では、保存する時間窓・ファイルサイズ・解析環境を設計し、無制限な常時採取にはしません。

</details>

---

## 設問 4: trace viewer を誰に見せるか決めよう

trace には goroutine 名、タスク名、ソース位置などの調査情報が入ります。手元で viewer を開くとき、同じネットワークの誰かに意図せず見せたくありません。どの `-http` 指定を選ぶべきでしょうか。

Go 1.27 で変わった `go tool trace -http=:6060` の扱いを、一次情報で確認してください。全アドレスで明示的に公開したい場合の指定も確認します。この問題のコマンドは Go 1.27.0 で実行しているため、listen-address 変更も手元で再現できます。

<details>
<summary>ヒント</summary>

- まず [Go 1.27 リリースノート](https://go.dev/doc/go1.27) の Trace を探します。
- 「ポートだけを渡す」指定と、「アドレスまで渡す」指定を比べてください。
- どちらが便利かではなく、trace に何が含まれうるかから公開範囲を考えます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) の Trace 節を読み、`-http=:6060` の listen address の変更を確認する。
2. 手元で `go tool trace -h` を実行して `-http=addr` の役割を確認する。
3. `go tool trace -http=localhost:0 work.trace` を使い、ローカル専用かつ空いているポートで viewer を起動する。共有が本当に必要なときだけ、ネットワーク到達性・認証・trace の取り扱いを別途確認する。

**答え**

- Go 1.27 では、ポートだけを指定する `-http=:6060` は localhost に制限されます。`go tool pprof -http` と同じ安全寄りの挙動です。
- 全アドレスで待ち受ける必要がある場合は、`-http=0.0.0.0:6060` のようにアドレスまで明示します。これは単なる書式の違いではなく、trace viewer の公開範囲を選ぶ操作です。
- 通常の調査は `-http=localhost:0` を選びます。task や region の名前も trace に載るため、「表示できる」ことと「同じネットワークに公開してよい」ことを混同しません。

</details>

---

<details>
<summary>こぼれ話</summary>

`go tool trace` は trace から `net`、`sync`、`syscall`、`sched` の pprof-like profile を出せます。今回のように「鍵待ち」が仮説なら `sync` から始められますが、スケジューラに載るまでの遅れなら `sched`、ネットワーク待ちなら `net` と、観測したい待ち方に合わせて選びます。まず CPU profile と trace を混ぜずに採り、必要な問いに合う profile だけを読むのが調査を短くするコツです。

</details>

---

## 調査の入り口

1. [Go の診断ツール案内](https://go.dev/doc/diagnostics)
2. [How To Build a User-Level CPU Profiler](https://research.swtch.com/pprof)
3. [trace の公式ドキュメント](https://go.dev/cmd/trace/)
4. 手元の `go help testflag`、`go tool trace -h`、`go tool pprof -h`
5. [Go 1.27.0 の `cmd/trace` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/trace/doc.go)
6. [Go 1.21 リリースノート](https://go.dev/doc/go1.21) と [Go 1.22 リリースノート](https://go.dev/doc/go1.22)
7. [execution tracer overhaul の設計文書](https://go.googlesource.com/proposal/+/refs/heads/master/design/60773-execution-tracer-overhaul.md) と [Issue #63185](https://github.com/golang/go/issues/63185)
8. [Go 1.27 リリースノート](https://go.dev/doc/go1.27)
