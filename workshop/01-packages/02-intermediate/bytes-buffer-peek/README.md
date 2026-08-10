# 受信メッセージのヘッダーを先読みしよう

バックエンドの受信処理では、外部システムから届くメッセージの先頭 4 バイトが種別、その後ろが本文です。受信担当は種別を見てから適切な処理へ渡したいものの、本文をまだ消費してはいけません。`bytes.Buffer.Peek` を使って受信したメッセージの先頭を確認したいです。どういうふうにやればいいか調べよう。

次の観測コードを [Go Playground で動かす](https://go.dev/play/p/FyP4yK-B61w) と、先読み・不足データ・返されたスライスの共有という 3 つの性質が見えます。

```go
package main

import (
	"bytes"
	"fmt"
)

func main() {
	packet := bytes.NewBufferString("HDR:payload")
	header, err := packet.Peek(4)
	fmt.Printf("header=%q err=%v remaining=%q\n", header, err, packet.String())

	header[0] = 'h'
	fmt.Printf("after mutation remaining=%q\n", packet.String())

	short := bytes.NewBufferString("OK")
	got, err := short.Peek(4)
	fmt.Printf("short=%q err=%v\n", got, err)
}
```

実行結果:

```text
header="HDR:" err=<nil> remaining="HDR:payload"
after mutation remaining="hDR:payload"
short="OK" err=EOF
```

---

## 設問 1: 読み取らずに種別を確認するには？

最初の出力で `header` は `"HDR:"` ですが、`remaining` も同じ `"HDR:payload"` です。`Peek(4)` は何をしており、同じ目的で `Next(4)` を使うと何が違うでしょうか。

<details>
<summary>ヒント</summary>

- [カテゴリの調べ方](../../README.md) を入口に `bytes.Buffer` のメソッド一覧を開く。
- `Peek` と `Next` の説明にある「buffer を進めるか」を比べる。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から `bytes` パッケージを開く。
2. [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek) と [Buffer.Next](https://pkg.go.dev/bytes#Buffer.Next) を比べる。
3. [buffer.go の `Peek` 実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/bytes/buffer.go;l=85) を読んで、オフセットを進めないことを確認する。

**答え**

`Peek(4)` は次の 4 バイトを返しますが、バッファの読み取り位置を進めません。そのため担当へ渡した後でも、本文を含む完全なパケットを読む処理が同じ内容を受け取れます。

`Next(4)` は 4 バイトを返したうえで、読み取ったものとしてバッファを進めます。種別の判定と消費を同時にしたい場合には使えますが、今回の「のぞいてから渡す」用途には `Peek` が合います。

</details>

---

## 設問 2: 4 バイトに足りないときは？

`"OK"` に対する `Peek(4)` は、なぜ空のスライスではなく `"OK"` と `EOF` を返すのでしょうか。受信が途中であることを検知する処理では、返り値をどう扱えばよいでしょうか。

<details>
<summary>ヒント</summary>

- `Peek` の「fewer than n bytes」の説明を探す。
- データの一部とエラーが同時に返る API であることに注目する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek) のエラー時の説明を読む。
2. [実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/bytes/buffer.go;l=85) で、残っているバイト列と `io.EOF` を返す分岐を確認する。

**答え**

要求した 4 バイトより少ないとき、`Peek` は現在あるバイト列を返し、同時に `io.EOF` を返します。空にして情報を失うのではなく、「ここまで受信済みだが、ヘッダーとしては不足」という状態を表せます。

受信処理は `err` を確認し、`io.EOF` ならさらにデータを受信してから再試行します。部分データ `got` を完全なヘッダーとして処理してはいけません。

</details>

---

## 設問 3: なぜ `header[0]` を書き換えると本文まで変わる？

返された `header` の 1 バイトを `h` にすると、バッファ全体も `"hDR:payload"` になりました。後段へ種別を安全に渡してからバッファを読み書きしたいとき、何に注意し、いつコピーを作るべきでしょうか。

<details>
<summary>ヒント</summary>

- `Peek` の説明で `valid until` と `aliases` を探す。
- 返り値を保持・変更するなら、元のバッファと同じ領域かを考える。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek) の返り値の有効期間と alias の説明を読む。
2. [buffer.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/bytes/buffer.go;l=85) のスライス式を確認する。

**答え**

`Peek` が返すスライスはバッファ内容を共有します。読み書きメソッドを呼んだ後は有効ではなく、返ったスライスを書き換えると、次に読む内容も変わり得ます。観測用の返り値を破壊的に変更してはいけません。

後段が種別を保持したり変更したりするなら、先に別のスライスへコピーします。読み取りだけで直後に使うならコピーは不要です。性能のための先読みと、所有権を分けたいデータを区別するのがポイントです。

</details>

---

<details>
<summary>こぼれ話</summary>

`Buffer.Peek` は Go 1.26 で追加されました。古いツールチェーンも対象にするモジュールでは、その最小 Go バージョンと `go.mod` の `go` 行を確認してから採用しましょう。

</details>

---

## 調査の入り口

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [01-packages の調べ方](../../README.md)
- [package bytes](https://pkg.go.dev/bytes)
