# 再点火タイマーの幽霊信号を追え

夜勤の航行制御では、再点火タイマーを停止して設定し直した直後に、以前の時刻の通知を受け取ったという古い障害報告が残っています。現行の観測ではチャンネル容量が 0 です。なんでこうなってるの？背景を調べよう。

次のコードを [Go Playground で動かす](https://go.dev/play/p/F4r40BDO4dv) と、現行のタイマーチャンネルの容量と、停止後の再設定を確認できます。

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

Go 1.23 より前のタイマーチャンネルは容量 1 のバッファ付きでした。現行の容量 0（同期）のチャンネルでは、`Stop` または `Reset` が返った後の古い通知について、どの保証が得られるでしょうか。

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

以前の容量 1 のチャンネルには古い通知が残り得たため、停止・再設定後にそれを受け取る複雑さがありました。航行制御の障害報告はこの違いを指しています。

</details>

---

## 設問 2: 古い「ドレイン」はなぜ危ない？

以前は `Stop` が `false` のときに `timer.C` を無条件で受信して空にする、いわゆるドレイン処理がよく書かれました。Go 1.23 以降の新しい意味論で、なぜその受信を無条件に残してはいけないのでしょうか。古い Go も動作対象にするライブラリでは、どんな前提を明確にすべきでしょうか。

<details>
<summary>ヒント</summary>

- `NewTimer` の `Stop` と古いバッファ付きチャンネルについての説明を読む。
- 「値がチャンネルに残る」と「タイマーが発火済みだが送信は完了していない」を区別する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [time.NewTimer](https://pkg.go.dev/time#NewTimer) の `Stop` / `Reset` に関する説明を読む。
2. [Go 1.23 Release Notes](https://go.dev/doc/go1.23) の容量 0 と stale value の説明を確認する。
3. 背景の [Issue #37196](https://github.com/golang/go/issues/37196) と [time パッケージの実装コメント](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=133) を読む。

**答え**

新しい意味論では、古い時刻値を受け取るためのドレインは不要です。`Stop` が `false` でも、値がバッファに残っているとは限りません。無条件の受信は待ち続ける可能性があります。

一方で、古い Go も対象なら挙動が異なります。ライブラリは最低対応 Go バージョン、`go.mod` の `go` 行、タイマーチャンネルを受信する goroutine の所有関係を明確にします。単に古い慣用句を削除・追加するのではなく、どの意味論で同期しているかを設計として確認します。

</details>

---

## 設問 3: `go.mod` と `GODEBUG` はどう関係した？

Go 1.23 では、新挙動の有効化にモジュールの `go` 行が関係していました。その後 Go 1.27 では `asynctimerchan` がどう変化したでしょうか。移行調査で一時的な互換設定を恒久策にしない理由を説明してください。

<details>
<summary>ヒント</summary>

- Go 1.23 のリリースノートで、新挙動が有効になる条件を探す。
- Go 1.26 と Go 1.27 のリリースノートで `asynctimerchan` を検索する。

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

## 設問 4: なぜこの変更は長く議論された？

古い通知を受信してしまう問題は、単なる API の見た目ではなく、`Stop` / `Reset` を正しく使う不変条件に関わります。提案 Issue と実装コメントを読んで、Go チームがどのような困難を減らそうとしたのか、1 分でチームへ説明してください。

<details>
<summary>ヒント</summary>

- Issue の冒頭にある、`Stop` 後に値があるかもしれないという従来の説明を読む。
- 実装コメントにある `stale time values` を探す。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Issue #37196](https://github.com/golang/go/issues/37196) の問題提起と議論を読む。
2. [Go 1.23 Release Notes](https://go.dev/doc/go1.23) で最終的な保証を確認する。
3. [sleep.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=133) の実装コメントを読む。

**答え**

従来は停止や再設定の後に、以前の期限の値がバッファに残る可能性がありました。正しさのために、返り値、ドレイン、他 goroutine の受信を細かく組み合わせる必要があり、誤用しやすい状態でした。

新しい設計は、古い値を後から受け取らないという強い保証を API に持たせます。これにより、利用側が「いま届いた通知は再設定前か」を推測する必要を減らし、タイマーのライフサイクルをより単純に扱えるようにしました。

</details>

---

<details>
<summary>こぼれ話</summary>

タイマーチャンネルの `len` や `cap` を見て受信可能性を判定するコードは移行の影響を受けます。値が来ているかを確認したい場合は、非ブロッキングの `select` を使うというリリースノートの案内も確認しましょう。

</details>

---

## 調査の入り口

- [Go 1.27 Release Notes](https://go.dev/doc/go1.27)
- [01-packages の調べ方](../../README.md)
- [package time](https://pkg.go.dev/time)
