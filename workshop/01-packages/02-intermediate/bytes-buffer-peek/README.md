[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [01-packages の調べ方](../../README.md)

# bytes.Buffer の先頭を読み取らずに確認しよう

![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8)

レビュー中に、`bytes.Buffer` の先頭 4 バイトを消費せず確認するコードを見かけました。`Peek` の返り値とバッファの関係を調べましょう。

次の観測コードを [Go Playground で動かす](https://go.dev/play/p/OdJEVvlBSbf) と、3 つの性質が見えます。先読み・不足データ・返されたスライスの共有です。

```go
package main

import (
	"bytes"
	"fmt"
)

func main() {
	buffer := bytes.NewBufferString("ABCDrest")
	prefix, err := buffer.Peek(4)
	fmt.Printf("prefix=%q err=%v remaining=%q\n", prefix, err, buffer.String())

	prefix[0] = 'a'
	fmt.Printf("after mutation remaining=%q\n", buffer.String())

	short := bytes.NewBufferString("XY")
	got, err := short.Peek(4)
	fmt.Printf("short=%q err=%v\n", got, err)
}
```

実行結果:

```text
prefix="ABCD" err=<nil> remaining="ABCDrest"
after mutation remaining="aBCDrest"
short="XY" err=EOF
```

<details>
<summary>調査の入り口</summary>

1. [01-packages の調べ方](../../README.md) を開き、標準パッケージのメソッド一覧の開き方を確かめます。
2. [Go 1.26 Release Notes](https://go.dev/doc/go1.26) で `Buffer.Peek` が Go 1.26 で追加されたことを確かめます。
3. [package bytes](https://pkg.go.dev/bytes) で `Buffer.Peek` と `Buffer.Next` の説明を比べ、読み取り位置を進めるか、要求より少ないときの返り値、返されたスライスの有効期間を確かめます。

</details>

---

## 設問 1: 読み取らずに先頭を確認するには？

最初の出力で `prefix` は `"ABCD"` ですが、`remaining` も `"ABCDrest"` のままです。`Peek(4)` は何をしており、`Next(4)` を使うと何が違うでしょうか。

<details>
<summary>ヒント</summary>

- [カテゴリの調べ方](../../README.md) を入口に `bytes.Buffer` のメソッド一覧を開く。
- `Peek` と `Next` の説明にある「buffer を進めるか」を比べる。
- `buffer := bytes.NewBufferString("ABCDrest")` で、同じ位置から両方を試す。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から `bytes` パッケージを開く。
2. [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek) と [Buffer.Next](https://pkg.go.dev/bytes#Buffer.Next) を比べる。
3. [buffer.go の `Peek` 実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/bytes/buffer.go;l=85) を読んで、オフセットを進めないことを確認する。

**答え**

`Peek(4)` は次の 4 バイトを返しますが、バッファの読み取り位置を進めません。そのため、後から読む処理も同じ内容を受け取れます。

`Next(4)` は 4 バイトを返したうえで、読み取り位置を進めます。先頭を消費せず確認する用途には `Peek` が合います。

</details>

---

## 設問 2: 4 バイトに足りないときは？

`"XY"` に対する `Peek(4)` は、なぜ空のスライスではなく `"XY"` と `EOF` を返すのでしょうか。返り値をどう扱えばよいでしょうか。

<details>
<summary>ヒント</summary>

- `Peek` の「fewer than n bytes」の説明を探す。
- データの一部とエラーが同時に返る API であることに注目する。
- 今のバッファには 2 バイトしかない。部分データと `io.EOF` を別々に扱う理由を整理する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から `bytes` パッケージを開く。
2. [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek) のエラー時の説明を読む。
3. [実装](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/bytes/buffer.go;l=85) で、残っているバイト列と `io.EOF` を返す分岐を確認する。

**答え**

要求した 4 バイトより少ないとき、`Peek` は現在あるバイト列を返し、同時に `io.EOF` を返します。空にせず、存在する分を返します。

呼び出し側は `err` を確認します。必要な長さに足りない部分データ `got` を、完全な結果として扱ってはいけません。

</details>

---

## 設問 3: なぜ `prefix[0]` を書き換えるとバッファも変わる？

返された `prefix` の 1 バイトを `a` にすると、バッファ全体も `"aBCDrest"` になりました。何に注意し、いつコピーを作るべきでしょうか。

<details>
<summary>ヒント</summary>

- `Peek` の説明で `valid until` と `aliases` を探す。
- 返り値を保持・変更するなら、元のバッファと同じ領域かを考える。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から `bytes` パッケージを開く。
2. [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek) の返り値の有効期間と alias の説明を読む。
3. [buffer.go](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/bytes/buffer.go;l=85) のスライス式を確認する。

**答え**

`Peek` が返すスライスはバッファ内容を共有します。読み書きメソッドを呼んだ後は有効ではなく、返ったスライスを書き換えると、次に読む内容も変わり得ます。観測用の返り値を破壊的に変更してはいけません。

返り値を保持したり変更したりするなら、先に別のスライスへコピーします。読み取りだけで直後に使うならコピーは不要です。

例えば `copyOfPrefix := bytes.Clone(prefix)` としておけば、後から `prefix[0]` を書き換えてもコピーは元の 4 バイトを保持します。返り値の用途でコピーの要否を決めます。

</details>

---

<details>
<summary>こぼれ話</summary>

`Buffer.Peek` は Go 1.26 で追加されました。古いツールチェーンも対象にするモジュールでは、その最小 Go バージョンと `go.mod` の `go` 行を確認してから採用しましょう。

</details>
