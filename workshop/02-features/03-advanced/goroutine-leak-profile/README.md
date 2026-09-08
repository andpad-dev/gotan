[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [02-features の調べ方](../../README.md)

# Go 1.27 の `goroutineleak` プロファイルは、なぜ「絶対に起きない goroutine」だけを教えてくれるのか

![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8)

動かしているプログラムの `/debug/pprof/goroutine` で、goroutine の数がじりじり増えています。
スタックを開くと、`chan send` で止まったものが積み上がっています。  
既存の `goroutine` プロファイルには、今存在する goroutine が全部並びます。  
「永遠に起きられないもの」と「長生きしているだけのもの」を区別できません。

Go 1.27 では `/debug/pprof/goroutineleak` が追加されています。  
同じリークパターンを切り出し、両方のプロファイルを並べて出します。

```go
// main.go
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

type result struct {
	value int
	err   error
}

func processWorkItem(id int) (int, error) {
	if id == 2 {
		return 0, errors.New("boom")
	}
	time.Sleep(100 * time.Millisecond)
	return id * 10, nil
}

func processWorkItems(ids []int) ([]int, error) {
	ch := make(chan result)
	for _, id := range ids {
		go func(id int) {
			v, err := processWorkItem(id)
			ch <- result{v, err}
		}(id)
	}
	var out []int
	for range ids {
		r := <-ch
		if r.err != nil {
			return nil, r.err
		}
		out = append(out, r.value)
	}
	return out, nil
}

func main() {
	fmt.Println("go version:", runtime.Version())
	_, err := processWorkItems([]int{0, 1, 2, 3, 4})
	fmt.Fprintln(os.Stderr, "processWorkItems err:", err)

	time.Sleep(300 * time.Millisecond)

	fmt.Println("=== goroutine profile ===")
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	fmt.Println("=== goroutineleak profile ===")
	pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
}
```

Playground: [Go のバージョンも表示する共有コード](https://go.dev/play/p/UwYB3wyRxe9)

このシナリオは Go 1.27 の機能を扱うため、**Playground で実行するのがおすすめ**です。  
手元で動かす場合は Go 1.27 が必要です。  
`GOTOOLCHAIN=go1.27.0 go run main.go` で実行すると、こう出ます。  
アドレスや絶対パスは環境ごとに変わります。

```
go version: go1.27.0
processWorkItems err: boom
=== goroutine profile ===
goroutine profile: total 5
4 @ 0x... 0x... 0x... 0x... 0x...
#	0x...	main.processWorkItems.func1+0x9b	./main.go:30

1 @ 0x... 0x... 0x... 0x... 0x... 0x... 0x... 0x...
#	0x...	runtime/pprof.writeRuntimeProfile+0xb3	GOROOT/src/runtime/pprof/pprof.go:848
#	0x...	runtime/pprof.writeGoroutine+0x4f	GOROOT/src/runtime/pprof/pprof.go:781
#	0x...	runtime/pprof.(*Profile).WriteTo+0x143	GOROOT/src/runtime/pprof/pprof.go:405
#	0x...	main.main+0x15b	./main.go:52
#	0x...	runtime.main+0x37f	GOROOT/src/runtime/proc.go:302

=== goroutineleak profile ===
goroutineleak profile: total 4
4 @ 0x... 0x... 0x... 0x... 0x...
#	0x...	main.processWorkItems.func1+0x9b	./main.go:30
```

同じ瞬間なのに `goroutine` は total 5、`goroutineleak` は total 4 です。  
この 1 個の差は何を意味するのでしょうか。  
ランタイムは「絶対に起きない」と「そうでない」をどう判別しているのでしょうか。

<details>
<summary>調査の入り口</summary>

1. [02-features の調べ方](../../README.md) を開き、言語仕様・公式ブログ・プロポーザルの逆引き手順を確かめます。
2. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) を開き、`Goroutine leak profile` 節で leaked goroutine の定義と、到達可能性に基づくため検出できないリークがあるという但し書きを読みます。
3. [A Tour of Go](https://research.swtch.com/gotour) で、2012 年の講演 Q&A にある「If a goroutine is stuck reading from a channel ...」という質問を探し、blocked goroutine を回収しない当時の理由を読みます。
4. [Go 1.26 リリースノート](https://go.dev/doc/go1.26) の `Experimental goroutine leak profile` 節で、同じ機能の experiment 版とリークするサンプルコードを確かめ、proposal issue #74609 へのリンクをたどります。
5. [`runtime/pprof` パッケージドキュメント](https://pkg.go.dev/runtime/pprof) の `type Profile` で、予約プロファイル名の一覧から `goroutine` と `goroutineleak` の 1 行説明を照合します。
6. [`net/http/pprof` パッケージドキュメント](https://pkg.go.dev/net/http/pprof) で、`/debug/pprof/goroutineleak` が登録されていることを確かめます。

> ※ 2026-08-31 にローカルの Go 1.27.0 と Go Playground の両方で実行した。  
> 先頭行は `go version: go1.27.0`、プロファイル件数は 5 と 4 だった。  
> 共有コード自身が実行版を表示するため、後日試す場合は先頭行も結果と一緒に記録する。

</details>

---

## 設問 1: `goroutine` プロファイルと `goroutineleak` プロファイル、二つの定義はどう違うのか？

両方が何を集めるプロファイルなのか、定義を一次資料で確認しましょう。

<details>
<summary>ヒント</summary>

- `goroutineleak` は Go 1.27 で新しく追加された「予約プロファイル名」です。まずは Go 1.27 のリリースノートで、この機能がどんな入り口で紹介されているかを確認しましょう。
- `runtime/pprof` パッケージのドキュメントには、既定で用意されているプロファイル名の一覧が並んでいます。`goroutineleak` の 1 行説明と、`goroutine` の 1 行説明を並べて比較してみましょう。
- リリースノートの当該節には、「どんな goroutine を "leaked" と呼ぶか」の定義が 1 文で書かれています。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) を開き、`Goroutine leak profile` 節でこの機能が pprof の新プロファイルとして紹介されている段落を読む。
2. [`runtime/pprof` パッケージドキュメント](https://pkg.go.dev/runtime/pprof) の `type Profile` の docstring に飛び、予約プロファイル名の一覧で `goroutine` と `goroutineleak` の 1 行説明を照合する。
3. リリースノート「Goroutine leak profile」節の中程にある「A leaked goroutine is a goroutine blocked on some concurrency primitive ... that cannot possibly become unblocked.」の一文を確認する。

**答え**

`runtime/pprof.Profile` の docstring では、予約プロファイル名がそれぞれ 1 行で説明されています。

- `goroutine` — `stack traces of all current goroutines`（**現存する** goroutine 全部）
- `goroutineleak` — `stack traces of all leaked goroutines`（**リークした** goroutine だけ）

そして「leaked goroutine」の定義は Go 1.27 リリースノートに明記されています。

> A leaked goroutine is a goroutine blocked on some concurrency primitive (channels, sync.Mutex, sync.Cond, etc) that cannot possibly become unblocked.

つまり `goroutineleak` は次の 2 条件を両方満たす goroutine だけを集めます。

1. 同期プリミティブ（channel、`sync.Mutex`、`sync.Cond` など）でブロックしている
2. **もう二度と起きられない**

冒頭の実行結果を当てはめると、`goroutine` プロファイル total 5 のうち 4 個は `chan send`（`./main.go:30`）でブロックしていて、残りの 1 個はプロファイルを書き出している最中の `main` goroutine です。
`goroutineleak` プロファイルはそこから 4 個だけ抽出しました。
差の 1 個は「まだ生きて動いている main は leak ではない」という当たり前の話です。

問題はここからで、条件 2 の「二度と起きられない」を、ランタイムはどう判定しているのでしょうか。

</details>

---

## 設問 2: ランタイムはどうやって「二度と起きられない」ことを判定しているのか？

「blocked on primitive」はランタイムから見て自明です。  
各 goroutine の `waitreason` などで分かります。  
難しいのは「二度と起きられない」の判定です。  
誰かがまだ channel の送信を握っているかもしれません。  
mutex がこれから解放されるかもしれません。  
この「未来の可能性」を、`goroutineleak` プロファイルはどう見分けているのでしょう？

<details>
<summary>ヒント</summary>

- リリースノートの `Goroutine leak profile` 節には、判定に**ランタイムの既存機能を流用している**ことをほのめかす一文があります。まずはその段落を最後まで読み切って、どの既存機能を頼りにしているのかを掴みましょう。
- Go 1.27 リリースノートで現在の機能を確認したら、そこからリンクされた Go 1.26 リリースノートの `Experimental goroutine leak profile` 節へ戻ります。この節には proposal issue へのリンクがあります。
- proposal issue の Description には「Design document」へのリンクがあります。設計文書は Go 本体のリポジトリではなく `go.googlesource.com/proposal` の下にあります。設計文書の `Proposal` 節には、番号付きで手順が並んでいます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) の `Goroutine leak profile` 節で現在の機能を確認し、そこから [Go 1.26 リリースノート](https://go.dev/doc/go1.26) の experiment 版へ戻る。
2. Go 1.26 リリースノートの `Experimental goroutine leak profile` 節にある proposal issue [#74609](https://go.dev/issue/74609) へのリンクを開く。
3. proposal Description の冒頭にある [Design document (`74609-goroutine-leak-detection-gc.md`)](https://go.googlesource.com/proposal/+/master/design/74609-goroutine-leak-detection-gc.md) を開き、`Proposal` 節の 6 ステップを冒頭の実行結果と突き合わせて理解する。

**答え**

要求のたびに、runtime は **専用の「goroutine leak detection GC cycle」** を回します。設計文書 `Proposal` 節の 6 ステップは次のとおりです。

1. Mark root の準備で、**runnable な goroutine だけ**を初期の root にする（通常の GC は「全 goroutine」を root にする）。
2. その root から reachable なメモリを mark する。
3. 到達可能なメモリを掃き切ったところで、**まだ mark されていない goroutine のうち、mark 済みの同期プリミティブを待っているもの**を探す。
4. 見つかったものは「いずれ起きられる可能性がある」ので新たな root に昇格し、ステップ 2 に戻って再 mark する。
5. 反復の固定点に到達したら、**root に昇格しなかった goroutine を leak として報告**する。安全のため、leak と判定した goroutine も root に加えてもう一度 mark を回し、reachable なメモリの掃き残しを防ぐ。
6. Sweep は通常通り。

冒頭の 4 個の goroutine に当てはめると、次のようになります。

- `processWorkItems` が early return した瞬間、ローカル変数の `ch` は関数の外に露出していないので、runnable な goroutine（`main`）からは **もう到達不可能**。
- 4 個のワーカーは `ch <- result{...}` で `sudog` を作り、その `sudog` は `ch` を指している。
- ステップ 1〜2 で `main` を root にすると、`ch` は mark されない。
- ステップ 3 で「blocked だが、待っている primitive が mark されていない goroutine」を探すと、その 4 個が該当。
- 誰も新たな root に昇格しないので、ステップ 5 で 4 個すべてが leak として報告される。

これが `goroutine` total 5 と `goroutineleak` total 4 の差の実体です。

設計文書は Rationale 節でこう結んでいます。

> It is also theoretically sound, i.e., there are no false positives.

「もう起きられない」と言われた goroutine は、**理論上、本当に永遠に起きられない**――これが Go 1.27 の `goroutineleak` プロファイルの保証です。

</details>

---

## 設問 3: なぜグローバル変数や runnable goroutine から到達できるプリミティブは検出しないのか？

Go 1.27 リリースノートの `Goroutine leak profile` 節には但し書きがあります。  
到達可能性に基づく手法のため、検出できないリークがあると書かれています。  
グローバル変数から辿れる channel を待つだけでも、実質リークではないでしょうか。  
なぜ Go チームはあえてこのパターンを検出範囲から外したのでしょうか。

<details>
<summary>ヒント</summary>

但し書きの原文は次の一文です。

> Because this technique builds on reachability, the runtime may fail to identify leaks caused by blocking on concurrency primitives reachable through global variables or the local variables of runnable goroutines.

- 設計文書の `Rationale` 節を開くと、この手法の性質を短く言い切っている一文があります。そこに書かれている **形式的な性質を表す一語**（形式手法や検証論で使われる語）に注目しましょう。
- 設問 2 の 6 ステップを見直しましょう。ステップ 3〜4 の「reachable なら root に昇格する」は、**未来**を静的に断定できないかわりに使う近似です。「グローバル変数から reachable」も同じ扱いになるとき、この手法は何を優先し、何を犠牲にしていますか？診断ツールが本番運用で嫌われるのはどんな失敗パターンでしょうか。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) の `Goroutine leak profile` 節で、到達可能性に基づくため検出できない例があるという但し書きを確認する。
2. [Go 1.26 リリースノート](https://go.dev/doc/go1.26) の experiment 版から proposal issue [#74609](https://go.dev/issue/74609) を開き、設計文書 [`74609-goroutine-leak-detection-gc.md`](https://go.googlesource.com/proposal/+/master/design/74609-goroutine-leak-detection-gc.md) の `Rationale` 節へ進む。
3. 設計文書 [`74609-goroutine-leak-detection-gc.md`](https://go.googlesource.com/proposal/+/master/design/74609-goroutine-leak-detection-gc.md) の `Rationale` 節を読む。
4. 設計文書 `Background` 節から、Uber Saioc らの学術論文（`10.1145/3676641.3715990`）へのリンクを辿り、"partial deadlock" の定義と検出可能性の議論を確認する（このページはブラウザで開くと 200 が返る一方、`curl` 等では 403 が返るため、参考文献として補助的に扱う）。

**答え**

キーワードは Rationale 節にあります。

> It is also theoretically sound, i.e., there are no false positives. Its primary limitation is that its effectiveness is reduced the more heap resources are over-exposed in memory, i.e., pair-wise reachable.

Go チームは **soundness を保つ代わりに completeness を犠牲にする**、という設計判断をしました。

- グローバル変数や runnable goroutine のローカル変数から reachable な channel を待っている goroutine は、「誰かがまだそこに触れるかもしれない」可能性が残ります。実際に触れるかは静的には決められない（停止性問題に近い）ため、**「リークだ」と断定できません**。
- 「かもしれない」で警告を出すと false positive（本当は将来起きるはずの goroutine を「リーク」と誤報する）を含んでしまいます。本番運用の診断ツールで false positive は致命的で、無害な警告が繰り返し出るとやがて誰も見なくなります。
- そこで Go 1.27 の `goroutineleak` は「**到達不可能な primitive を待っているものだけ**」に対象を絞りました。これは「本当にリークしている場合の一部を漏らす（false negative は許容）」ことと引き換えに、「**リークだと言われたものは必ずリーク**」を保証しています。

冒頭の 4 個は `ch` がローカル変数として消えた瞬間から誰も触れなくなるので、この「sound」な条件にはっきり当てはまるパターン、というわけです。

</details>

---

## 設問 4: なぜ「回収」ではなく「診断」になり、実装まで 10 年以上かかった？

Go 1.26 で experimental、Go 1.27 で GA。  
この裏には 10 年以上前からの前史があります。

[Go 1.27 リリースノート](https://go.dev/doc/go1.27) で現在の機能を確認し、さらに前へ戻ります。  
2012 年、Russ Cox は [A Tour of Go](https://research.swtch.com/gotour) の質疑を受けました。  
「参照されなくなった channel を待つ goroutine は GC されるか」という質問です。  
当時の回答と 2015 年の partial deadlock proposal を読みます。  
そのうえで現在の `goroutineleak` profile と比べてください。  
何が変わり、なぜ今も goroutine を消さず診断情報として残すのかを説明しましょう。

<details>
<summary>ヒント</summary>

- [Go 1.27 リリースノート](https://go.dev/doc/go1.27) で現在の profile を確認し、そこから [Go 1.26 リリースノート](https://go.dev/doc/go1.26) の experiment 版へ戻って proposal issue に進みます。
- research.swtch.com の記事では「If a goroutine is stuck reading from a channel ...」という質問を探し、2012 年当時の実装についての説明と、診断時にスタックを残す理由を分けて読みます。
- proposal issue [#74609](https://go.dev/issue/74609) のコメント欄で、開発者たちが `#13759` という別 issue に言及しています。**その issue を開いてみましょう。**
- issue #13759 の中で、Go チームメンバーの一人が Dec 2015 に「blocked な goroutine を最初は root から外す → mark → 待っている primitive が mark されていれば root に昇格 → 反復」という **アイデア** を短いコメントで書いています。設問 2 で読んだ設計文書の 6 ステップと比べてみましょう。
- proposal #74609 の議論で、提案者本人が「#13759 のあのコメントとほぼ同じアプローチだ」と認めている返信があります。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) の `Goroutine leak profile` で現在の機能を確認し、リンク先の [Go 1.26 リリースノート](https://go.dev/doc/go1.26) の experiment 版から proposal [#74609](https://go.dev/issue/74609) と設計文書へ進む。現在の機能が special GC cycle の結果を profile と traceback に出す診断機能であること、設計文書の手順 5 が報告後の leaked goroutine も mark root に戻すことを確認する。
2. proposal #74609 の [「Collecting dead goroutines and the memory they reference」](https://github.com/golang/go/issues/74609#issuecomment-3119676873) を読む。goroutine はメモリ以外に network connection、file、C memory なども保持し得るため、メモリだけの回収は問題を先送りして診断を難しくし得る、という現在の判断を確認する。
3. [A Tour of Go](https://research.swtch.com/gotour) の 2012 年の質疑を読む。当時は channel の送受信側を GC から区別できないという実装上の説明に加え、blocked goroutine を回収すると deadlock 時に有用な stack を表示できなくなる、という診断上の理由が述べられている。この記述は 2012 年当時の実装説明であり、現在実装の根拠には使わない。
4. proposal #74609 の議論から、`randall77` が言及している [issue #13759](https://github.com/golang/go/issues/13759) を開く。
5. issue #13759 の RLH（Rick Hudson、当時 Go GC 開発者）の Dec 30, 2015 コメントを読む。設問 2 の 6 ステップと同じアイデアを、10 年前に短いコメントで書き切っていることに気付く。
6. issue の下の方で、rsc の Nov 22, 2016 コメントを確認する。proposal は accepted-in-principle にされたが、「we are not ourselves planning to do the work」で `Unplanned` マイルストーンに置かれた。
7. proposal #74609 で [VladSaiocUber の Jul 16, 2025 コメント](https://github.com/golang/go/issues/74609#issuecomment-3077678435) を確認する。RLH の 2015 年コメントと自分たちの実装がほぼ同一だと明言している。

**答え**

`goroutineleak` プロファイルは 2025 年に突然生まれた機能ではありません。**少なくとも 2012 年には「回収するのか」という問いが表に現れ、2015 年から partial deadlock 検出として議論されていた、10 年以上越しの診断機能**です。

- **2012-06** — research.swtch.com の講演 Q&A で、参照されない channel を待つ goroutine を GC するかが質問された。回答は「しない」で、当時の channel 表現の制約だけでなく、回収すると deadlock handler が有用な goroutine stack を失うことも理由に挙げた。これは当時の実装説明であり、現在の channel 表現についての主張ではない。
- **2015-12** — issue [#13759](https://github.com/golang/go/issues/13759) で rfliam が「GC の mark phase を使って partial deadlock を検出できないか」という proposal を投稿。RLH が短いコメントで、**現在の設計とほぼ同じ「blocked な goroutine を最初は root から外して mark を反復する」アプローチ**を提案。
- **2016-11** — rsc が accepted-in-principle でクローズ。「we are not ourselves planning to do the work」で `Unplanned` マイルストーンに。良いアイデアだが誰もやらない、の状態が続く。
- **2025** — Uber の Saioc らが `10.1145/3676641.3715990` で理論と実装を発表。proposal Description によれば、Uber 社内で 3111 のテストスイートと本番サービス（24 時間で 252 件のリーク検出）で検証した実績を伴う。
- **2025-07** — VladSaiocUber が [proposal #74609](https://go.dev/issue/74609) を上げて、設計文書と `goroutineleakprofile` GOEXPERIMENT で提供。「[#13759 の RLH のコメント](https://github.com/golang/go/issues/74609#issuecomment-3077678435) とほぼ同じアプローチだ」と自ら認めている。試作 CL は [go-review 688335](https://go.dev/cl/688335)。
- **2026-02** — Go 1.26 で experiment としてリリース。`GOEXPERIMENT=goroutineleakprofile` でオプトイン。
- **2026-08** — Go 1.27 で既定 ON になり GA。同時に `goroutineleakprofile` フラグは削除された（[CL 774620](https://go.dev/cl/774620) で default-on、[CL 774621](https://go.dev/cl/774621) で experiment 削除）。

2012 年の回答と現在の設計には、blocked goroutine を黙って消さず、原因を調べられる状態を残すという共通点が見えます。ただし、これは歴史資料を並べた考察であり、現在「回収ではなく診断」を選んだ直接の根拠は proposal #74609 の議論にあります。goroutine は Go heap のメモリだけでなく、network connection、file、C memory なども保持し得ます。メモリだけを回収すると他の resource は残り、問題の発覚を遅らせたり、かえって原因を追いにくくしたりする可能性があるため、まず detector として提供するのが安全だと判断されました。

変わったのは、GC の到達可能性を使って「絶対に起きられない」一部を false positive なしで識別し、その結果を `goroutineleak` profile として要求時に取り出せる点です。設計文書の手順 5 は、リークを報告した後、その goroutine も mark root に戻して reachable memory を mark すると定めています。つまり現在の機能も回収ではなく検出です。

冒頭の「本番で goroutine 数がなだらかに増える」現象を思い出すと、10 年前の #13759 で aclements が既にこう書いていました。

> most of the time you know a deadlock has happened and the tricky part is figuring out exactly what's involved in the deadlock.

Go 1.27 で入った `goroutineleak` プロファイルは、まさにその「何が関わっているかを特定する」部分を、**本番サービスの pprof エンドポイントから直接取得できる**形にしたものです。冒頭で見た 1 個の差（`goroutine` total 5 と `goroutineleak` total 4）は、10 年待った結論として「main は runnable なので leak ではない、他の 4 個は誰にも到達されないので leak だ」を、ランタイムが自動で言い切ってくれた結果、というわけです。

</details>

---

<details>
<summary>こぼれ話: `select{}` は「意図的な無限ブロック」だが、実装当初は leak 扱いされていた</summary>

Go でメインループを止める慣用イディオムに `select{}`（case のない select）があります。これに到達した goroutine は永遠にブロックしますが、これは意図した動作で、リークではありません。

`goroutineleak` プロファイル実装の初期版では、`main` goroutine が `select{}` でブロックしているケースも「blocked, かつ待つべき primitive がない」扱いになり、leak として報告されていました。ユーザーからのフィードバックを受けて [CL 770020](https://go.dev/cl/770020) `runtime: exclude main goroutine blocked on select{} from goroutine leak profile` で除外されています。

面白いのは、CL のコミットメッセージにこう書かれていることです。

> The main goroutine is still treated as a leak during the analysis to avoid degrading analysis precision (see test), but has its status changed from leaked back to waiting before the profile is written.

「解析中は leak として扱い続ける（そのほうが検出精度が高いので）、最終的なプロファイル出力の直前だけ status を戻す」――Go の慣用イディオムを尊重するためのちょっとした特別扱いが、Go 1.27 のランタイムの中に隠れています。

</details>
