# Go 1.27 の `goroutineleak` プロファイルは、なぜ「絶対に起きない goroutine」だけを教えてくれるのか

運用しているサービスで、`/debug/pprof/goroutine` を眺めていると goroutine の数が時間経過とともにじりじり増えています。  
スタックを開くと、`chan send` で止まっているものが少しずつ積み上がっているようです。  
ただ、既存の `goroutine` プロファイルには「今この瞬間に存在している goroutine 全部」が並ぶので、「本当に永遠に起きられないやつ」と「単に長生きしているだけのやつ」の区別がつきません。

先輩が「Go 1.27 に上げてみたら `/debug/pprof/goroutineleak` というエンドポイントが増えていた」と教えてくれました。試しに、原因調査のために社内で見つけたリークパターンをそのまま切り出し、`goroutine` と `goroutineleak` の両方を並べて出してみます。

```go
// main.go
package main

import (
	"errors"
	"fmt"
	"os"
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
	_, err := processWorkItems([]int{0, 1, 2, 3, 4})
	fmt.Fprintln(os.Stderr, "processWorkItems err:", err)

	time.Sleep(300 * time.Millisecond)

	fmt.Println("=== goroutine profile ===")
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	fmt.Println("=== goroutineleak profile ===")
	pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
}
```

Playground: <https://go.dev/play/p/baiTvOMwD2C>

`go1.27.0 run main.go` で実測すると、こう出ます（アドレスは環境ごとに変わります）。

```
processWorkItems err: boom
=== goroutine profile ===
goroutine profile: total 5
4 @ 0x... 0x... 0x... 0x... 0x...
#	0x...	main.processWorkItems.func1+0x9b	./main.go:29

1 @ 0x... 0x... 0x... 0x... 0x... 0x... 0x... 0x...
#	0x...	runtime/pprof.writeRuntimeProfile+0xb3	GOROOT/src/runtime/pprof/pprof.go:848
#	0x...	runtime/pprof.writeGoroutine+0x4f	GOROOT/src/runtime/pprof/pprof.go:781
#	0x...	runtime/pprof.(*Profile).WriteTo+0x143	GOROOT/src/runtime/pprof/pprof.go:405
#	0x...	main.main+0xff	./main.go:50
#	0x...	runtime.main+0x37f	GOROOT/src/runtime/proc.go:302

=== goroutineleak profile ===
goroutineleak profile: total 4
4 @ 0x... 0x... 0x... 0x... 0x...
#	0x...	main.processWorkItems.func1+0x9b	./main.go:29
```

同じ瞬間のスナップショットのはずなのに、`goroutine` は total 5、`goroutineleak` は total 4。  
この 1 個の差は何を意味していて、そもそもランタイムはどうやって「これは絶対に起きない」「これは違う」を判別しているのでしょうか？

---

## 設問 1: `goroutine` プロファイルと `goroutineleak` プロファイル、二つの定義はどう違うのか？

まずは両方が「何を集めるプロファイル」なのか、それぞれの定義を一次資料で確認しましょう。

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

冒頭の実測を当てはめると、`goroutine` プロファイル total 5 のうち 4 個は `chan send`（`./main.go:29`）でブロックしていて、残りの 1 個はプロファイルを書き出している最中の `main` goroutine です。  
`goroutineleak` プロファイルはそこから 4 個だけ抽出しました。  
差の 1 個は「まだ生きて動いている main は leak ではない」という当たり前の話です。

問題はここからで、条件 2 の「二度と起きられない」を、ランタイムはどう判定しているのでしょうか。

</details>

---

## 設問 2: ランタイムはどうやって「二度と起きられない」ことを判定しているのか？

「blocked on primitive」はランタイムから見て自明な状態です（各 goroutine の `waitreason` などで分かる）。  
難しいのは「二度と起きられない」の判定です。  
誰かがまだ channel の送信を握っているかもしれない、mutex を解放するかもしれない――そういう「未来の可能性」を、`goroutineleak` プロファイルはどう見分けているのでしょう？

<details>
<summary>ヒント</summary>

- リリースノートの `Goroutine leak profile` 節には、判定に**ランタイムの既存機能を流用している**ことをほのめかす一文があります。まずはその段落を最後まで読み切って、どの既存機能を頼りにしているのかを掴みましょう。
- Go 1.27 リリースノートの各行には `go.dev/issue/<N>` の形で関連 issue のリンクが埋め込まれています（category README 参照）。Goroutine leak profile 節にも該当の proposal issue リンクがあります。
- proposal issue の Description には「Design document」へのリンクがあります。設計文書は Go 本体のリポジトリではなく `go.googlesource.com/proposal` の下にあります。設計文書の `Proposal` 節には、番号付きで手順が並んでいます。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.27 リリースノート](https://go.dev/doc/go1.27) の `Goroutine leak profile` 節から、埋め込まれた proposal issue [#74609](https://go.dev/issue/74609) に飛ぶ。
2. proposal Description の冒頭にある [Design document (`74609-goroutine-leak-detection-gc.md`)](https://go.googlesource.com/proposal/+/master/design/74609-goroutine-leak-detection-gc.md) を開く。
3. 設計文書の `Proposal` 節にある 6 ステップを、冒頭の実行結果と突き合わせて理解する。

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

## 設問 3: なぜ「グローバル変数や runnable goroutine のローカル変数から到達可能なプリミティブ」は検出しないのか？

Go 1.27 リリースノートの `Goroutine leak profile` 節には、実は次の但し書きが添えられています。

> Because this technique builds on reachability, the runtime may fail to identify leaks caused by blocking on concurrency primitives reachable through global variables or the local variables of runnable goroutines.

グローバル変数から辿れる channel を延々と待っているだけでも、「起こしに来る当てがない」なら実質リークですよね？なぜ Go チームはあえてこのパターンを検出範囲から外したのでしょうか。

<details>
<summary>ヒント</summary>

- 設計文書の `Rationale` 節を開くと、この手法の性質を短く言い切っている一文があります。そこに書かれている **形式的な性質を表す一語**（形式手法や検証論で使われる語）に注目しましょう。
- 設問 2 の 6 ステップを見直しましょう。ステップ 3〜4 の「reachable なら root に昇格する」は、**未来**を静的に断定できないかわりに使う近似です。「グローバル変数から reachable」も同じ扱いになるとき、この手法は何を優先し、何を犠牲にしていますか？診断ツールが本番運用で嫌われるのはどんな失敗パターンでしょうか。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. 設計文書 [`74609-goroutine-leak-detection-gc.md`](https://go.googlesource.com/proposal/+/master/design/74609-goroutine-leak-detection-gc.md) の `Rationale` 節を読む。
2. 設計文書 `Background` 節から、Uber Saioc らの学術論文（`10.1145/3676641.3715990`）へのリンクを辿り、"partial deadlock" の定義と検出可能性の議論を確認する（このページはブラウザで開くと 200 が返る一方、`curl` 等では 403 が返るため、参考文献として補助的に扱う）。

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

## 設問 4: この機能はいつから議論されていたのか？

Go 1.26 で experimental、Go 1.27 で GA というスピード感の裏に、実は 10 年前からの前史があります。それを辿ってみましょう。

<details>
<summary>ヒント</summary>

- proposal issue [#74609](https://go.dev/issue/74609) のコメント欄で、開発者たちが `#13759` という別 issue に言及しています。**その issue を開いてみましょう。**
- issue #13759 の中で、Go チームメンバーの一人が Dec 2015 に「blocked な goroutine を最初は root から外す → mark → 待っている primitive が mark されていれば root に昇格 → 反復」という **アイデア** を短いコメントで書いています。設問 2 で読んだ設計文書の 6 ステップと比べてみましょう。
- proposal #74609 の議論で、提案者本人が「#13759 のあのコメントとほぼ同じアプローチだ」と認めている返信があります。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. proposal [#74609](https://go.dev/issue/74609) の議論から、`randall77` が言及している [issue #13759](https://github.com/golang/go/issues/13759) を開く。
2. issue #13759 の RLH（Rick Hudson、当時 Go GC 開発者）の Dec 30, 2015 コメントを読む。設問 2 の 6 ステップと同じアイデアを、10 年前に短いコメントで書き切っていることに気付く。
3. issue の下の方で、rsc の Nov 22, 2016 コメントを確認する。proposal は accepted-in-principle にされたが、「we are not ourselves planning to do the work」で `Unplanned` マイルストーンに置かれた。
4. proposal #74609 で [VladSaiocUber の Jul 16, 2025 コメント](https://github.com/golang/go/issues/74609#issuecomment-3077678435) を確認する。RLH の 2015 年コメントと自分たちの実装がほぼ同一だと明言している。

**答え**

`goroutineleak` プロファイルは 2025 年に突然生まれた機能ではなく、**2015 年から議論されていた partial deadlock 検出の、10 年越しの実装**です。

- **2015-12** — issue [#13759](https://github.com/golang/go/issues/13759) で rfliam が「GC の mark phase を使って partial deadlock を検出できないか」という proposal を投稿。RLH が短いコメントで、**現在の設計とほぼ同じ「blocked な goroutine を最初は root から外して mark を反復する」アプローチ**を提案。
- **2016-11** — rsc が accepted-in-principle でクローズ。「we are not ourselves planning to do the work」で `Unplanned` マイルストーンに。良いアイデアだが誰もやらない、の状態が続く。
- **2025** — Uber の Saioc らが `10.1145/3676641.3715990` で理論と実装を発表。proposal Description によれば、Uber 社内で 3111 のテストスイートと本番サービス（24 時間で 252 件のリーク検出）で検証した実績を伴う。
- **2025-07** — VladSaiocUber が [proposal #74609](https://go.dev/issue/74609) を上げて、設計文書と `goroutineleakprofile` GOEXPERIMENT で提供。「[#13759 の RLH のコメント](https://github.com/golang/go/issues/74609#issuecomment-3077678435) とほぼ同じアプローチだ」と自ら認めている。試作 CL は [go-review 688335](https://go.dev/cl/688335)。
- **2026-02** — Go 1.26 で experiment としてリリース。`GOEXPERIMENT=goroutineleakprofile` でオプトイン。
- **2026-08** — Go 1.27 で既定 ON になり GA。同時に `goroutineleakprofile` フラグは削除された（[CL 774620](https://go.dev/cl/774620) で default-on、[CL 774621](https://go.dev/cl/774621) で experiment 削除）。

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

---

## 調査の入り口

- [Go 1.27 リリースノート](https://go.dev/doc/go1.27)（`Goroutine leak profile` 節。proposal・設計文書・論文へのリンクの起点）
- [Go 1.26 リリースノート](https://go.dev/doc/go1.26)（`Experimental goroutine leak profile` 節。同じ機能の experiment 版と、リークするサンプルコード）
- [`runtime/pprof` パッケージドキュメント](https://pkg.go.dev/runtime/pprof)（`type Profile` docstring の予約プロファイル名一覧）
- [`net/http/pprof` パッケージドキュメント](https://pkg.go.dev/net/http/pprof)（`/debug/pprof/goroutineleak` エンドポイントの登録）

> ※ 実測は `go1.27.0`（`go install golang.org/dl/go1.27.0@latest` で導入）で行った。Go Playground の既定バージョンが Go 1.27 未満の場合、`Lookup("goroutineleak")` が `nil` を返して panic する可能性があるため、Playground 上で試すときはバージョンセレクタで Go 1.27 以降を選ぶ。
