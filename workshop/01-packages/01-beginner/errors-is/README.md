# 救難信号の正体を照合しよう

救難管制の予約サービスは、利用者に便名を含む詳しいエラーを返しつつ、内部では「空席なし」を数えたい状況です。ログに同じ言葉が見えても `==` が `false` になっていました。`errors.Is` を見かけました。どんなものか調べてみましょう。

次のプログラムを [Go Playground で動かす](https://go.dev/play/p/sQdfdBISMyk) と、表示用の文脈と照合結果が別物だと分かります。

```go
package main

import (
	"errors"
	"fmt"
)

var ErrNoSeat = errors.New("空席がありません")

func reserve(flight string) error {
	return fmt.Errorf("便 %s の予約に失敗: %w", flight, ErrNoSeat)
}

func main() {
	err := reserve("M-17")
	fmt.Println(err)
	fmt.Println("== で照合:", err == ErrNoSeat)
	fmt.Println("errors.Is で照合:", errors.Is(err, ErrNoSeat))
}
```

実行結果:

```text
便 M-17 の予約に失敗: 空席がありません
== で照合: false
errors.Is で照合: true
```

---

## 設問 1: 表示は同じなのに、なぜ `==` は `false`？

`err` の表示には `空席がありません` と出ています。それでも `err == ErrNoSeat` が `false` になる理由と、`errors.Is` が `true` になる理由を説明してください。

<details>
<summary>ヒント</summary>

- `fmt.Errorf` の書式文字列にある `%w` を手掛かりにする。
- `errors.Is` の説明で、どのエラーを順に調べるか探す。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から `errors` パッケージを探す。
2. [errors.Is](https://pkg.go.dev/errors#Is) の error tree の説明を読む。
3. [wrap.go の実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/errors/wrap.go;l=45) を読み、`Is` が包まれたエラーをたどることを確認する。

**答え**

`fmt.Errorf` の `%w` は `ErrNoSeat` を包んだ新しいエラーを作ります。したがって最上位の `err` と元の `ErrNoSeat` は同じ値ではなく、`==` は `false` です。

`errors.Is` は、最上位のエラーから `Unwrap` で得られるエラーをたどり、対象と一致するものがあるかを調べます。その木の中に `ErrNoSeat` があるので `true` になります。

</details>

---

## 設問 2: 管制画面ではどちらで分岐する？

「空席なし」なら待機リストを案内し、それ以外なら再試行を案内する分岐を作ります。`err.Error()` の文字列比較、`err == ErrNoSeat`、`errors.Is(err, ErrNoSeat)` のうち、どれを使うべきでしょうか。また、`%w` を `%v` に変えると何が変わるでしょうか。

<details>
<summary>ヒント</summary>

- 設問 1 のプログラムで `%w` を `%v` に変え、出力と照合結果を予想する。
- `errors.Is` の対象は表示文字列ではなく `error` 値である。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [fmt.Errorf](https://pkg.go.dev/fmt#Errorf) で `%w` の説明を読む。
2. [errors.Is](https://pkg.go.dev/errors#Is) に戻り、照合の対象と探索順を確認する。
3. [fmt のエラー生成実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/fmt/errors.go;l=22) を読む。

**答え**

分岐には `errors.Is(err, ErrNoSeat)` を使います。文字列は文言変更・翻訳・便名の追加で変わり得る表示情報であり、`==` は包んだ最上位のエラーしか比較できません。

`%w` を `%v` に変えると文面は似ていてもエラーを包まなくなります。そのため `errors.Is` は `ErrNoSeat` に到達できず `false` になります。原因として扱いたいエラーには `%w` を使い、表示専用の値には `%v` を使い分けます。

</details>

---

<details>
<summary>こぼれ話</summary>

特定のエラー型に含まれる追加情報を取り出したい場合は、`errors.As` を使います。原因が「この値か」を問うのが `Is`、「この型か」を問うのが `As` です。

</details>

---

## 調査の入り口

- [Go Documentation](https://go.dev/doc/)
- [01-packages の調べ方](../../README.md)
- [package errors](https://pkg.go.dev/errors)
