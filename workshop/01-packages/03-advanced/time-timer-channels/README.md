# time.Timer の Stop/Reset と古い通知を追え

決済連携の再試行ジョブでは、タイマーを停止して設定し直した直後に、以前の期限の通知を受け取ったという古い障害報告が残っています。現行の観測ではチャンネル容量が 0 です。なんでこうなってるの？背景を調べよう。

次のコードを [Go Playground で動かす](https://go.dev/play/p/F4r40BDO4dv) と、現行のタイマーチャンネルの容量と、停止後の再設定を確認できます。

このシナリオには、Go 1.23 以降のタイマーチャンネルの挙動を有効にする `go.mod` を同梱しています。ローカルで試すときは、このディレクトリでコードを実行してください。

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

---

## 設問 1: 容量 0 は障害報告と何が違う？

Go 1.23 より前のタイマーチャンネルは容量 1 のバッファ付きでした。現行の容量 0（同期）のチャンネルでは、`Stop` または `Reset` が返った後の古い再試行通知について、どの保証が得られるでしょうか。

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
3. [sleep.go のコメント](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=123) を読む。

**答え**

Go 1.23 以降の新しい挙動では、タイマーチャンネルは容量 0 の同期チャンネルです。`Stop` または `Reset` が返った後に、その呼び出しより前に準備された古い時刻値を送受信しないことが保証されます。

以前の容量 1 のチャンネルには古い通知が残り得たため、停止・再設定後にそれを受け取る複雑さがありました。再試行ジョブの障害報告はこの違いを指しています。

</details>

---

## 設問 2: 古い「ドレイン」はなぜ危ない？

以前は `Stop` が `false` のときに `timer.C` を無条件で受信して空にする、いわゆるドレイン処理がよく書かれました。Go 1.23 以降の新しい意味論で、なぜその受信を無条件に残してはいけないのでしょうか。古い Go も動作対象にするライブラリでは、どんな前提を明確にすべきでしょうか。

次の最小実験は、タイマーの通知を一度受信してから `Stop` し、その後に古い値をドレインしようとします。[Go Playground で実行する](https://go.dev/play/p/X_a-_2PuQoV) と、`Stop=false` でも受信できる値が残っているとは限らないことを観測できます。

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

比較実験をするときは、導入の Go Playground の結果を基準にし、手元では `go version`、`go env GOMOD`、`go env GODEBUG` も記録してください。特に `go.mod` がない実行や、`go 1.23` 未満のモジュールでは、同じ Go コンパイラでも旧いタイマーチャンネルの挙動になることがあります。

<details>
<summary>ヒント</summary>

- `NewTimer` の `Stop` と古いバッファ付きチャンネルについての説明を読む。
- 「値がチャンネルに残る」と「タイマーが発火済みだが送信は完了していない」を区別する。
- `Stop` が `false` でも、値がバッファに残っているとは限らないことを確認する。別の goroutine が先に受信した場合も含め、無条件の受信には `select` とタイムアウトを付けて観察する。
- [最小実験](https://go.dev/play/p/X_a-_2PuQoV) は通知を一度受信してから `Stop` し、後続の受信をタイムアウト付きで試します。`Stop=false` と「今すぐ `C` から受信できる」を同一視しないでください。
- 古い挙動との切り替え条件として、主プログラムの `go.mod` の `go` 行と `GODEBUG=asynctimerchan` を調べる。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/X_a-_2PuQoV) を実行し、`Stop=false` の後の受信がタイムアウトすることを観測する。
2. [Go 1.23 Release Notes](https://go.dev/doc/go1.23) の Timer changes を読み直す。
3. [time.NewTimer](https://pkg.go.dev/time#NewTimer) の `Stop` / `Reset` に関する説明を読む。
4. 容量 0 と stale value の背景を、[Issue #37196](https://github.com/golang/go/issues/37196) と [time パッケージの実装コメント](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=133) で確認する。

**答え**

新しい意味論では、古い時刻値を受け取るためのドレインは不要です。`Stop` が `false` でも、それは「チャンネルに値がバッファされている」という意味ではありません。同期チャンネルでは、別の goroutine が通知を受信済みであったり、古い通知が無効化されていたりするため、無条件の受信は待ち続ける可能性があります。

一方で、古い Go も対象なら挙動が異なります。ライブラリは最低対応 Go バージョン、主プログラムの `go.mod` の `go` 行、`GODEBUG=asynctimerchan` の扱い、タイマーチャンネルを受信する goroutine の所有関係を明確にします。古いバッファ付きチャンネル向けのドレインは、受信者を一つに限定できる旧い前提でのみ成立します。単に古い慣用句を削除・追加するのではなく、どの意味論で同期しているかを設計として確認します。

</details>

---

## 設問 3: `go.mod` と `GODEBUG` はどう関係した？

Go 1.23 では、新挙動の有効化にモジュールの `go` 行が関係していました。その後 Go 1.27 では `asynctimerchan` がどう変化したでしょうか。移行調査で一時的な互換設定を恒久策にしない理由を説明してください。

この設問でいう「新挙動」は、まず `time.NewTimer(0)` のチャンネルの `cap` / `len` として観察します。「互換設定」は、環境変数の `GODEBUG`、`go.mod` の `godebug`、ソース中の `//go:debug` を混同せず、それぞれがどの実行条件に効くかを整理してください。

<details>
<summary>ヒント</summary>

- Go 1.23 のリリースノートで、新挙動が有効になる条件を探す。
- Go 1.26 と Go 1.27 のリリースノートで `asynctimerchan` を検索する。
- [既定値の実験](https://go.dev/play/p/zAkeGmN14Q7)、[旧挙動を指定した実験](https://go.dev/play/p/o-pDN23HyaX)、[新挙動を指定した実験](https://go.dev/play/p/V8esF3Hmib8) を順に実行し、`cap=0` と `cap=1` の違いを確認する。Go 1.27が未リリースまたはPlaygroundで選べない場合は、リリースノートを「予定仕様」として扱い、実行結果と混ぜない。
- これらのPlaygroundはGo 1.26.5で、期待値は順に `go1.26.5 0 0`、`go1.26.5 1 0`、`go1.26.5 0 0` です。手元や開催時のバージョンが違う場合は、実際のバージョンも記録して比較します。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.23 Release Notes](https://go.dev/doc/go1.23) で、`go 1.23.0` 以上のモジュールと `asynctimerchan=1` の説明を読む。
2. [Go 1.26 Release Notes](https://go.dev/doc/go1.26) で、Go 1.27 からの変更予定を確認する。
3. [Go 1.27 Release Notes](https://go.dev/doc/go1.27) の Runtime と GODEBUG の説明を読む。

**答え**

Go 1.23 では、主プログラムのモジュールが `go 1.23.0` 以降なら新しいタイマー挙動が有効で、`asynctimerchan=1` は調査や移行時に旧挙動へ戻すための設定でした。

Go 1.27 ではこの設定は恒久的に削除され、`time` のタイマーチャンネルは常に同期・非バッファです。したがって、互換設定に依存して障害を隠すのではなく、旧来のドレインや `len` / `cap` への依存を直して、現在の意味論で正しく動くようにする必要があります。

</details>

---

## 設問 4: 長く議論された変更の影響箇所をどう突き止める？

古い通知を受信してしまう問題は、単なる API の見た目ではなく、`Stop` / `Reset` を正しく使う不変条件に関わります。提案 Issue と実装コメントを読んで、Go チームがどのような困難を減らそうとしたのか、1 分でチームへ説明してください。

例えば「タイマーを止める → 古い通知を空にする → 期限を設定し直す」という処理の途中で、別の goroutine が `timer.C` を受信すると何が起きるかを考えます。古い通知と新しい通知を取り違えないために、利用側が何を覚えておく必要があったのかを、現在のAPI保証と比べてください。

さらに Go 1.23〜1.26 への移行中、テストスイート全体では旧挙動なら成功し、新挙動なら失敗することまで分かったとします。`GODEBUG` の全体切り替えだけでは、どの `time.NewTimer` 呼び出しが旧挙動に依存しているかは分かりません。[Go 1.23 の Timer Channel Changes](https://go.dev/wiki/Go123Timer) と、実際のタイマー障害を題材にした Russ Cox の [Hash-Based Bisect Debugging in Compilers and Runtimes](https://research.swtch.com/bisect) を辿り、`git bisect` と `bisect` がそれぞれ何を二分探索するのか、揺らぐテストでは何に注意するのかも説明してください。

<details>
<summary>ヒント</summary>

- Issue の冒頭にある、`Stop` 後に値があるかもしれないという従来の説明を読む。
- 実装コメントにある `stale time values` を探す。
- [Issue #37196](https://github.com/golang/go/issues/37196) の冒頭にある従来のドレイン例と、[Go 1.26.4のsleep.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=105) の `Stop` / `NewTimer` のコメントを並べて読む。
- 公式 Wiki の「Debugging」で、互換設定を全体へ切り替えた後に示される診断手順と、その出力例を読む。
- research.swtch.com の記事では、`git bisect` の結果の次に始まる「A New Trick」と、揺らぐ対象を繰り返す理由に注目する。bisect のドキュメントでは、試行の反復に関する flag を探す。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.23 Release Notes](https://go.dev/doc/go1.23) でタイマー変更の概要を確認する。
2. [Issue #37196](https://github.com/golang/go/issues/37196) の問題提起と議論を読む。
3. 最終的な保証を [sleep.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=133) の実装コメントで確認する。
4. [Go Wiki の Timer Channel Changes](https://go.dev/wiki/Go123Timer) の「Debugging」を読み、旧・新挙動の全体切り替えで原因の種類を確認した後、`bisect` がスタックトレースごとに挙動を切り替えて依存箇所を絞る手順を確認する。
5. [Hash-Based Bisect Debugging in Compilers and Runtimes](https://research.swtch.com/bisect) のタイマー障害の実例を読み、`git bisect` は変更を導入したコミットを、`bisect` は同じプログラム内で変更を有効にする呼び出しスタックを探索する、という探索軸の違いを整理する。揺らぐ失敗に対しては、記事、[`golang.org/x/tools/cmd/bisect` のドキュメント](https://pkg.go.dev/golang.org/x/tools/cmd/bisect)、[v0.47.0 のコマンドソース](https://cs.opensource.google/go/x/tools/+/refs/tags/v0.47.0:cmd/bisect/main.go) を読み、対象コマンドの `go test -count=N` と `bisect -count=N` の役割も区別する。

**答え**

従来は停止や再設定の後に、以前の期限の値がバッファに残る可能性がありました。正しさのために、返り値、ドレイン、他 goroutine の受信を細かく組み合わせる必要があり、誤用しやすい状態でした。

新しい設計は、古い値を後から受け取らないという強い保証を API に持たせます。これにより、利用側が「いま届いた通知は再設定前か」を推測する必要を減らし、タイマーのライフサイクルをより単純に扱えるようにしました。

移行時の大きなテストで失敗した場合、`GODEBUG=asynctimerchan=0` と `=1` をプロセス全体で切り替えれば、タイマー意味論の変更が失敗に関係するかを確認できます。しかし、それだけでは問題のある呼び出し箇所までは特定できません。`git bisect` がリポジトリのコミット履歴を探索するのに対し、`golang.org/x/tools/cmd/bisect` は同じテストを繰り返し、`GODEBUG` のハッシュパターンを使って新旧挙動を適用する呼び出しスタックの集合を絞ります。

この探索は試行結果が一貫していることを前提にします。対象コマンド側の `go test -count=N` は揺らぐ失敗を観測しやすくする一方、`bisect -count=N` は bisect の各試行を複数回実行し、結果の不一致を検出します。後者は単に失敗率を上げる指定ではありません。Go 1.27 では設問 3 のとおり `asynctimerchan` 自体が削除されたため、この切り分けは Go 1.23〜1.26 の移行期間に使う診断手段です。最終的には、冒頭の障害報告にある旧来のドレインやタイミング依存を、現在の API 保証に合わせて修正します。

</details>

---

<details>
<summary>こぼれ話</summary>

タイマーチャンネルの `len` や `cap` を見て受信可能性を判定するコードは移行の影響を受けます。値が来ているかを確認したい場合は、非ブロッキングの `select` を使うというリリースノートの案内も確認しましょう。

</details>

---

## 調査の入り口

- [Go 1.27 Release Notes](https://go.dev/doc/go1.27)
- [Go Wiki: Go 1.23 Timer Channel Changes](https://go.dev/wiki/Go123Timer)
- [Hash-Based Bisect Debugging in Compilers and Runtimes](https://research.swtch.com/bisect)
- [01-packages の調べ方](../../README.md)
- [package time](https://pkg.go.dev/time)
