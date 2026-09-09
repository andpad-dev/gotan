[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [01-packages の調べ方](../../README.md)

# crypto/hpke で一通のメッセージを暗号化しよう

**実行環境**: ブラウザだけ。Go のインストールは不要です。

送信側が受信側の公開鍵で一通のメッセージを暗号化し、受信側が秘密鍵で復号します。Go 1.26 で追加された `crypto/hpke` の API を調べましょう。

これは API の役割を調べる演習です。実運用のプロトコル選定や鍵管理には、必ず組織のセキュリティレビューを加えてください。

同じ `info` なら復号でき、異なる `info` なら拒否されます。次のコードを [Go Playground で動かす](https://go.dev/play/p/ONc5q75dM5t) と確認できます。

```go
package main

import (
	"crypto/hpke"
	"fmt"
)

func main() {
	kem, kdf, aead := hpke.MLKEM768X25519(), hpke.HKDFSHA256(), hpke.AES256GCM()
	privateKey, err := kem.GenerateKey()
	if err != nil {
		panic(err)
	}

	publicKey, err := kem.NewPublicKey(privateKey.PublicKey().Bytes())
	if err != nil {
		panic(err)
	}
	info := []byte("example/v1")
	ciphertext, err := hpke.Seal(publicKey, kdf, aead, info, []byte("message"))
	if err != nil {
		panic(err)
	}

	plaintext, err := hpke.Open(privateKey, kdf, aead, info, ciphertext)
	if err != nil {
		panic(err)
	}
	_, wrongInfoErr := hpke.Open(privateKey, kdf, aead, []byte("example/v2"), ciphertext)
	fmt.Printf("received=%q\n", plaintext)
	fmt.Println("different info rejected:", wrongInfoErr != nil)
}
```

実行結果:

```text
received="message"
different info rejected: true
```

<details>
<summary>調査の入り口</summary>

まず [01-packages の調べ方](../../README.md) を開き、標準パッケージの Overview の読み方を確かめます。

そのうえで、次のどれかから入ります。

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26) — `crypto/hpke` が追加されたときのリリースノート
- [package crypto/hpke](https://pkg.go.dev/crypto/hpke) — サンプルで使っている型と関数の説明はこのページにある

</details>

---

## 設問 1: 3 つの部品は何を選んでいる？

`kem`、`kdf`、`aead` は HPKE のどの部品でしょうか。サンプルで選ばれている具体的な実装名も答えてください。

<details>
<summary>ヒント</summary>

- [カテゴリの調べ方](../../README.md) を入口に `crypto/hpke` の Overview を読む。
- `KEM`、`KDF`、`AEAD` の型と、コード中のコンストラクタ名を対応付ける。
- 役割を「誰の公開鍵を使うか」「共有した材料から鍵をどう作るか」「本文をどう暗号化・認証するか」の3つに分けて読むと、似た略語を整理しやすい。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.26 Release Notes](https://go.dev/doc/go1.26) で `crypto/hpke` が新しい標準パッケージであることを確認する。
2. [package crypto/hpke](https://pkg.go.dev/crypto/hpke) の Overview と各型を読む。
3. [Go 1.27.0 のワンショット API](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/crypto/hpke/hpke.go;l=180-193) と [標準ライブラリの例](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/crypto/hpke/hpke_test.go;l=21-66) を読み、KEM・KDF・AEAD がどこで選ばれているか対応付ける。

**答え**

HPKE の暗号スイートは KEM（鍵カプセル化方式）、KDF（鍵導出関数）、AEAD（認証付き暗号）で構成されます。サンプルはそれぞれ `MLKEM768X25519`、`HKDFSHA256`、`AES256GCM` を選んでいます。

自分で暗号処理を組み立てるのではなく、パッケージが定義するこれらの部品と API を一組として扱うのが出発点です。

</details>

---

## 設問 2: 送信側・受信側は何を共有する？

送信側はどの鍵で `Seal` し、受信側はどの鍵で `Open` するでしょうか。また、なぜ両者は同じ `info` を渡す必要があり、異なる値なら失敗するのでしょうか。

<details>
<summary>ヒント</summary>

- `Seal` と `Open` の引数を横に並べる。
- 一回だけ送る API の `Seal` / `Open` のコメントを読む。
- 公開鍵、秘密鍵、`info` のうち、`Seal` と `Open` の両方へ渡す値を確認する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から標準ライブラリの `crypto/hpke` を開く。
2. [Seal](https://pkg.go.dev/crypto/hpke#Seal) と [Open](https://pkg.go.dev/crypto/hpke#Open) のシグネチャを確認する。
3. [Go 1.27.0 の `Seal`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/crypto/hpke/hpke.go;l=180-193) と [`Open`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/crypto/hpke/hpke.go;l=221-235) のコメントと実装を読み、公開鍵・秘密鍵・`info` がどこへ渡るかを追う。

**答え**

送信側は受信者の公開鍵で `Seal` し、受信側は対応する秘密鍵で `Open` します。ここでは `privateKey.PublicKey().Bytes()` を送信側が受け取り、`kem.NewPublicKey` で公開鍵として復元しています。

`info` は両者で同じ値を使う文脈情報です。復号側に異なる `info` を渡すと、同じ文脈で作られた暗号文ではないため `Open` がエラーを返します。

</details>

---

## 設問 3: なぜ `kem.NewPublicKey` を使う？

通信で受け取った公開鍵バイト列を、なぜ選んだ `kem` の `NewPublicKey` で復元するのでしょうか。KEM の鍵形式と、KEM/KDF/AEAD からなる暗号スイートを分けて考えてください。組み合わせをどのように扱うべきか説明してください。

<details>
<summary>ヒント</summary>

- `KEM.NewPublicKey` の説明で、バイト列を何として解釈するかを探す。
- Go のソースで `MLKEM768X25519` を検索し、返している具体的な KEM から `NewPublicKey` の実装を探す。
- `MLKEM768X25519` を別の KEM に取り替えた場合、たどり着く `NewPublicKey` がどう変わるかを考える。
- `NewPublicKey` に KDF や AEAD を渡していないことにも注目し、暗号スイートが公開鍵バイト列から自動的に復元されるかを確認する。
- 同じバイト列でも、別の KEM で読むと別の形式として扱われます。一方、KDF と AEAD は別途選ぶ部品です。この2つの違いを、鍵の形式と暗号スイートの設定表に分けて整理してみましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から標準ライブラリの `crypto/hpke` を開く。
2. [KEM](https://pkg.go.dev/crypto/hpke#KEM) の `NewPublicKey` が公開鍵をバイト列から復元するメソッドであることを確認する。
3. [Go 1.27.0 の `MLKEM768X25519`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/crypto/hpke/pq.go;l=20-43) を開き、返される `mlkem768X25519` が `*hybridKEM` であることを確認する。
4. [その `hybridKEM.NewPublicKey`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/crypto/hpke/pq.go;l=164-180) で、公開鍵全体の長さを検証してから ML-KEM 部と楕円曲線部に分け、それぞれを復元している処理を追う。
5. 対比として [DHKEM の `NewPublicKey`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/crypto/hpke/kem.go;l=215-221) を開き、選んだ KEM によって復元処理が異なることを確認する。

**答え**

公開鍵のバイト列は、単独で意味が決まるものではありません。`KEM.NewPublicKey` は、選択済みの KEM に従ってその列を公開鍵として復元・検証します。

サンプルの `MLKEM768X25519()` が返すのは `*hybridKEM` です。その `NewPublicKey` は、受け取った長さが ML-KEM の公開鍵 1184 バイトと X25519 の公開鍵 32 バイトの合計 1216 バイトかを検証し、前半を ML-KEM、後半を X25519 の処理へ渡します。一方、DHKEM を選べば `dhKEM.NewPublicKey` が楕円曲線の公開鍵として復元します。したがって、サンプルの呼び出しを追う根拠は `dhKEM` ではなく `hybridKEM` の実装です。

ここで指定しているのは KEM の鍵形式であり、KDF や AEAD の選択を復元・交渉しているわけではありません。

送受信者は、公開鍵をどの KEM でエンコードしたかと、KEM/KDF/AEAD の暗号スイートを別々のプロトコル情報として取り決めます。受信したバイト列を別の方式の鍵として推測して扱いません。方式の選択や鍵の配布を場当たり的に混ぜないことが、ライブラリ API を正しく使う前提です。

</details>

---

<details>
<summary>こぼれ話</summary>

一通だけ送るなら `Seal` / `Open` が使えます。複数メッセージを送る `Sender` / `Recipient` では、成功した `Seal` と `Open` の呼び出し順を両側でそろえる必要があります。用途に応じて API を選びましょう。

</details>
