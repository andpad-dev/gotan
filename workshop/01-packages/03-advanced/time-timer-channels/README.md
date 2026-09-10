[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [01-packages の調べ方](../../README.md)

# time.Timer の Stop/Reset と古い通知を追え

![実行環境: Go 1.27 以上](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.27%20%E4%BB%A5%E4%B8%8A-F39C12)

古い Go では、タイマーを停止して設定し直した後も、以前の期限の値を受け取ることがありました。いま同じコードを動かすと、チャンネル容量は **0** です。なんでこうなってるの？背景を調べよう。

次のコードを [Go Playground で動かす](https://go.dev/play/p/F4r40BDO4dv) と、容量と停止後の再設定を確認できます。

このディレクトリには `main.go` と `go.mod` を同梱しています。ローカルでは `go run .` で動かせます。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	timer := time.NewTimer(time.Hour)
	fmt.Println("channel capacity:", cap(timer.C))
	fmt.Println("stop before firing:", timer.Stop())

	timer.Reset(time.Millisecond)
	<-timer.C
	fmt.Println("reset timer fired")
}
```

実行結果:

```text
channel capacity: 0
stop before firing: true
reset timer fired
```

<details>
<summary>調査の入り口</summary>

まず [01-packages の調べ方](../../README.md) を開き、標準パッケージのドキュメントの開き方と、`go version` などの実行環境の記録手順を確かめます。

そのうえで、次のどれかから入ります。

- [Go 1.27 Release Notes](https://go.dev/doc/go1.27) — 実行環境に指定している Go 1.27 のリリースノート
- [Go Wiki: Go 1.23 Timer Channel Changes](https://go.dev/wiki/Go123Timer) — Go 1.23 のタイマー変更を解説する公式 Wiki
- [Hash-Based Bisect Debugging in Compilers and Runtimes](https://research.swtch.com/bisect) — Russ Cox がタイマー障害の調査を題材に書いた記事
- [package time](https://pkg.go.dev/time) — `Timer` の説明はこのページにある

</details>

---

## 設問 1: 容量 0 は古い挙動と何が違う？

Go 1.23 より前のタイマーチャンネルは容量 1 のバッファ付きでした。現行は容量 0 の同期チャンネルです。`Stop` や `Reset` が返った後、古い通知についてどの保証が得られるでしょうか。

<details>
<summary>ヒント</summary>

- `time.NewTimer` のドキュメントで `Before Go 1.23` を読む。
- Go 1.23 リリースノートの「Timer changes」で、何が「stale value」だったか探す。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.23 Release Notes](https://go.dev/doc/go1.23) の「Timer changes」を読む。
2. [time.NewTimer](https://pkg.go.dev/time#NewTimer) のバージョン間の説明を読む。
3. [Go 1.27.0 の sleep.go のコメント](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/time/sleep.go;l=77) を読む。

**答え**

Go 1.23 以降の新しい挙動では、タイマーチャンネルは容量 0 の同期チャンネルです。`Stop` または `Reset` が返った後に、その呼び出しより前に準備された古い時刻値を送受信しないことが保証されます。

以前の容量 1 のチャンネルには古い通知が残り得たため、停止・再設定後にそれを受け取る複雑さがありました。

</details>

---

## 設問 2: 古い「ドレイン」はなぜ危ない？

`Stop` が `false` のとき、`timer.C` を無条件に受信して空にする。かつてはこの**ドレイン処理**がよく書かれました。Go 1.23 以降の意味論で、なぜその受信を無条件に残せないのでしょうか。古い Go も動作対象にするライブラリでは、どんな前提を明確にすべきでしょうか。

次の最小実験は、通知を一度受信してから `Stop` し、続けて古い値をドレインしようとします。[Go Playground で実行する](https://go.dev/play/p/X_a-_2PuQoV) と分かります。`Stop=false` でも、受信できる値があるとは限りません。

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	timer := time.NewTimer(0)
	<-timer.C
	stopped := timer.Stop()
	fmt.Printf("stop=%v cap=%d\n", stopped, cap(timer.C))
	select {
	case <-timer.C:
		fmt.Println("drain received")
	case <-time.After(20 * time.Millisecond):
		fmt.Println("unconditional drain would block")
	}
}
```

実行結果:

```text
stop=false cap=0
unconditional drain would block
```

比較実験の基準は、冒頭の Go Playground の結果です。手元では `go version`、`go env GOMOD`、`go env GODEBUG` を記録してください。`go.mod` の版や `asynctimerchan` で旧挙動へ切り替えられたのは Go 1.26 までです。Go 1.27 では、旧値を指定しても容量 1 には戻りません。

<details>
<summary>ヒント</summary>

- `NewTimer` の `Stop` と古いバッファ付きチャンネルについての説明を読む。
- 「値がチャンネルに残る」と「タイマーが発火済みだが送信は完了していない」を区別する。
- `Stop` が `false` でも、値がバッファに残っているとは限らないことを確認する。別の goroutine が先に受信した場合も含め、無条件の受信には `select` とタイムアウトを付けて観察する。
- [最小実験](https://go.dev/play/p/X_a-_2PuQoV) の `Stop=false` と「今すぐ `C` から受信できる」を同一視しないでください。
- Go 1.26 までの切り替え条件として、主プログラムの `go.mod` の `go` 行と `GODEBUG=asynctimerchan` を調べる。続けて Go 1.27 で設定が削除されたことを確認する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/X_a-_2PuQoV) を実行し、`Stop=false` の後の受信がタイムアウトすることを観測する。
2. [Go 1.23 Release Notes](https://go.dev/doc/go1.23) の Timer changes を読み直す。
3. [time.NewTimer](https://pkg.go.dev/time#NewTimer) の `Stop` / `Reset` に関する説明を読む。
4. 容量 0 と stale value の背景を、[Issue #37196](https://github.com/golang/go/issues/37196) と [Go 1.27.0 の time パッケージの実装コメント](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/time/sleep.go;l=77) で確認する。

**答え**

新しい意味論では、古い時刻値を受け取るためのドレインは不要です。`Stop` が `false` でも、それは「チャンネルに値がバッファされている」という意味ではありません。同期チャンネルでは、別の goroutine が通知を受信済みであったり、古い通知が無効化されていたりするため、無条件の受信は待ち続ける可能性があります。

一方で、Go 1.26 以前の toolchain も対象なら挙動が異なります。ライブラリは最低対応 Go バージョン、主プログラムの `go.mod` の `go` 行、`GODEBUG=asynctimerchan` の扱い、タイマーチャンネルを受信する goroutine の所有関係を明確にします。古いバッファ付きチャンネル向けのドレインは、受信者を一つに限定できる旧い前提でのみ成立します。単に古い慣用句を削除・追加するのではなく、どの意味論で同期しているかを設計として確認します。

</details>

---

## 設問 3: `go.mod` と `GODEBUG` はどう関係した？

Go 1.23 では、新挙動の有効化にモジュールの `go` 行が関係していました。その後 Go 1.27 では `asynctimerchan` がどう変化したでしょうか。移行調査で一時的な互換設定を恒久策にしない理由を説明してください。

この設問では、次の二つの言葉を切り分けます。

- **新挙動**: `time.NewTimer(0)` のチャンネルの `cap` / `len` で観察する。
- **互換設定**: 環境変数の `GODEBUG`、`go.mod` の `godebug`、ソース中の `//go:debug`。どれがどの実行条件に効くかを整理する。

出発点は次のとおりです。Playground の実行環境は更新されるため、ここでは 2026-09-04 に Go 1.27.1 で再実行した結果を記録しています。なぜ 3 本目だけ、容量を表示する前に失敗するのでしょうか。

| 実験 | Playground (Go 1.27.1) の結果 |
| --- | --- |
| [既定値](https://go.dev/play/p/zAkeGmN14Q7) | `go1.27.1 0 0` |
| [`//go:debug asynctimerchan=0`](https://go.dev/play/p/V8esF3Hmib8) | `go1.27.1 0 0` |
| [`//go:debug asynctimerchan=1`](https://go.dev/play/p/o-pDN23HyaX) | `invalid //go:debug: removed GODEBUG "asynctimerchan" set to old value "1"` |

<details>
<summary>ヒント</summary>

- Go 1.23 のリリースノートで、新挙動が有効になる条件を探す。
- Go 1.26 と Go 1.27 のリリースノートで `asynctimerchan` を検索する。
- 3 つの実験の `//go:debug` の有無と値を比較し、リリースノートの規則に当てはめる。
- [Go 1.27 リリースノートの GODEBUG 節](https://go.dev/doc/go1.27#godebug) で、最終的な既定値は受理され、旧値は拒否されるという規則を確認する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.23 Release Notes](https://go.dev/doc/go1.23) で、`go 1.23.0` 以上のモジュールと `asynctimerchan=1` の説明を読む。
2. [Go 1.26 Release Notes](https://go.dev/doc/go1.26) で、当時予告されていた Go 1.27 での削除を確認する。
3. [Go 1.27 Release Notes の GODEBUG 節](https://go.dev/doc/go1.27#godebug) と [Runtime 節](https://go.dev/doc/go1.27#runtime) を読み、正式に削除された結果を確認する。
4. 3 つの Playground を実行し、既定値・最終値 `0`・旧値 `1` の結果を記録する。
5. [Go 1.27.0 の削除済み GODEBUG 一覧](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/godebugs/table.go;l=100) で、`asynctimerchan` の旧値が `1` と `2` だと確認する。

**答え**

Go 1.23 では、主プログラムのモジュールが `go 1.23.0` 以降なら新しいタイマー挙動が有効で、`asynctimerchan=1` は調査や移行時に旧挙動へ戻すための設定でした。

Go 1.27 ではこの設定は恒久的に削除され、`time` のタイマーチャンネルは同期・非バッファに固定されました。主プログラムの `go` 行が 1.23 未満でも旧挙動には戻りません。`asynctimerchan=0` は最終値として受理されますが意味は変わらず、旧値 `1`（および `2`）は拒否されます。

指定場所によって失敗する段階が異なります。`go.mod` の `godebug` とソースの `//go:debug` は `go` コマンドがビルド前に拒否し、環境変数 `GODEBUG=asynctimerchan=1` は起動時に fatal error になります。どの場合も容量 1 の互換動作は得られません。したがって、互換設定に依存して障害を隠すのではなく、旧来のドレインや `len` / `cap` への依存を直して、現在の意味論で正しく動くようにする必要があります。

| 指定場所 | 再現方法 | Go 1.27.0 の結果 |
| --- | --- | --- |
| 環境変数 | `GODEBUG=asynctimerchan=1 go run .` | 起動時に `fatal error` |
| `go.mod` | `godebug asynctimerchan=1` を追加して `go run .` | `go.mod` 読み込み時にエラー |
| ソース | [`//go:debug asynctimerchan=1` の Playground](https://go.dev/play/p/o-pDN23HyaX) | コンパイルエラー |

</details>

---

## 設問 4: 長く議論された変更の影響箇所をどう突き止める？

古い通知を受信する問題は、`Stop` / `Reset` の**不変条件**に関わります。提案 Issue と実装コメントを読み、Go が減らそうとした困難を説明してください。

「タイマーを止める → 古い通知を空にする → 期限を設定し直す」を考えます。この途中で別の goroutine が `timer.C` を受信すると何が起きるでしょうか。取り違えを防ぐために利用側が覚えておく必要があったことを、現在の API 保証と比べてください。

さらに、Go 1.23〜1.26 への移行中を想定します。テストスイート全体は、旧挙動なら成功し新挙動なら失敗するとします。`GODEBUG` の全体切り替えだけでは、どの `time.NewTimer` 呼び出しが旧挙動に依存するか分かりません。次の二つを辿ってください。二つ目は、実際のタイマー障害を題材にした Russ Cox の記事です。

- [Go 1.23 の Timer Channel Changes](https://go.dev/wiki/Go123Timer)
- [Hash-Based Bisect Debugging in Compilers and Runtimes](https://research.swtch.com/bisect)

`git bisect` と `bisect` はそれぞれ何を二分探索するのでしょうか。揺らぐテストでは何に注意するのでしょうか。この二つも説明してください。

<details>
<summary>ヒント</summary>

- Issue の冒頭にある、`Stop` 後に値があるかもしれないという従来の説明を読む。
- 実装コメントにある `stale time values` を探す。
- [Issue #37196](https://github.com/golang/go/issues/37196) の冒頭にある従来のドレイン例と、[Go 1.27.0 の sleep.go](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/time/sleep.go;l=77) の `Stop` / `NewTimer` のコメントを並べて読む。
- 公式 Wiki の「Debugging」で、互換設定を全体へ切り替えた後に示される診断手順と、その出力例を読む。
- research.swtch.com の記事では、`git bisect` の結果の次に始まる「A New Trick」と、揺らぐ対象を繰り返す理由に注目する。bisect のドキュメントでは、試行の反復に関する flag を探す。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.23 Release Notes](https://go.dev/doc/go1.23) でタイマー変更の概要を確認する。
2. [Issue #37196](https://github.com/golang/go/issues/37196) の問題提起と議論を読む。
3. 最終的な保証を [Go 1.27.0 の sleep.go](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/time/sleep.go;l=77) の実装コメントで確認する。
4. [Go Wiki の Timer Channel Changes](https://go.dev/wiki/Go123Timer) の「Debugging」を読み、旧・新挙動の全体切り替えで原因の種類を確認した後、`bisect` がスタックトレースごとに挙動を切り替えて依存箇所を絞る手順を確認する。
5. [Hash-Based Bisect Debugging in Compilers and Runtimes](https://research.swtch.com/bisect) のタイマー障害の実例を読み、`git bisect` は変更を導入したコミットを、`bisect` は同じプログラム内で変更を有効にする呼び出しスタックを探索する、という探索軸の違いを整理する。揺らぐ失敗に対しては、記事、[`golang.org/x/tools/cmd/bisect` のドキュメント](https://pkg.go.dev/golang.org/x/tools/cmd/bisect)、[v0.47.0 のコマンドソース](https://cs.opensource.google/go/x/tools/+/refs/tags/v0.47.0:cmd/bisect/main.go) を読み、対象コマンドの `go test -count=N` と `bisect -count=N` の役割も区別する。

**答え**

従来は停止や再設定の後に、以前の期限の値がバッファに残る可能性がありました。正しさのために、返り値、ドレイン、他 goroutine の受信を細かく組み合わせる必要があり、誤用しやすい状態でした。

新しい設計は、古い値を後から受け取らないという強い保証を API に持たせます。これにより、利用側が「いま届いた通知は再設定前か」を推測する必要を減らし、タイマーのライフサイクルをより単純に扱えるようにしました。

移行時の大きなテストで失敗した場合、`GODEBUG=asynctimerchan=0` と `=1` をプロセス全体で切り替えれば、タイマー意味論の変更が失敗に関係するかを確認できます。しかし、それだけでは問題のある呼び出し箇所までは特定できません。`git bisect` がリポジトリのコミット履歴を探索するのに対し、`golang.org/x/tools/cmd/bisect` は同じテストを繰り返し、`GODEBUG` のハッシュパターンを使って新旧挙動を適用する呼び出しスタックの集合を絞ります。

この探索は試行結果が一貫していることを前提にします。対象コマンド側の `go test -count=N` は揺らぐ失敗を観測しやすくする一方、`bisect -count=N` は bisect の各試行を複数回実行し、結果の不一致を検出します。後者は単に失敗率を上げる指定ではありません。Go 1.27 では設問 3 のとおり `asynctimerchan` 自体が削除されたため、この切り分けは Go 1.23〜1.26 の移行期間に使う診断手段です。最終的には、旧来のドレインやタイミング依存を、現在の API 保証に合わせて修正します。

</details>

---

<details>
<summary>こぼれ話</summary>

タイマーチャンネルの `len` や `cap` を見て受信可能性を判定するコードは移行の影響を受けます。値が来ているかを確認したい場合は、非ブロッキングの `select` を使うというリリースノートの案内も確認しましょう。

</details>
