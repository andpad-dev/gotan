[進行ガイド・シナリオ一覧](../../../README.md) | [01-packages の調べ方](../../README.md)

# 監査先へ暗号化ファイルを届けよう

外部監査への顧客データ提出では、送信側が監査先の公開鍵だけを知り、監査先だけが内容を開ける形にしたいです。Go 1.26 で追加された `crypto/hpke` を使って、標準ライブラリの例に沿って一通のメッセージを届けたいです。どういうふうにやればいいか調べよう。

これは API の役割を調べる演習です。実運用のプロトコル選定や鍵管理には、必ず組織のセキュリティレビューを加えてください。

次のコードを [Go Playground で動かす](https://go.dev/play/p/FDMWW7yXLFE) と、同じ `info` なら復号でき、異なる `info` なら拒否されることを確認できます。

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
	info := []byte("audit-export/v1")
	ciphertext, err := hpke.Seal(publicKey, kdf, aead, info, []byte("customer export ready"))
	if err != nil {
		panic(err)
	}

	plaintext, err := hpke.Open(privateKey, kdf, aead, info, ciphertext)
	if err != nil {
		panic(err)
	}
	_, wrongInfoErr := hpke.Open(privateKey, kdf, aead, []byte("audit-export/v2"), ciphertext)
	fmt.Printf("received=%q\n", plaintext)
	fmt.Println("different info rejected:", wrongInfoErr != nil)
}
```

実行結果:

```text
received="customer export ready"
different info rejected: true
```

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
3. [hpke.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/hpke.go;l=183) と [標準ライブラリの例](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/hpke_test.go;l=21) を読む。

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
- たとえば公開鍵を誰でも入れられる郵便受け、秘密鍵を受取人だけが持つ鍵、`info` を「監査エクスポート v1」という用途ラベルだと考える。同じラベルを使わないと、同じ暗号文でも開けない理由を確認する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Seal](https://pkg.go.dev/crypto/hpke#Seal) と [Open](https://pkg.go.dev/crypto/hpke#Open) のシグネチャを確認する。
2. [hpke.go の `Seal`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/hpke.go;l=183) と [`Open`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/hpke.go;l=224) のコメントを読む。

**答え**

送信側は受信者の公開鍵で `Seal` し、受信側は対応する秘密鍵で `Open` します。ここでは `privateKey.PublicKey().Bytes()` を送信側が受け取り、`kem.NewPublicKey` で公開鍵として復元しています。

`info` は両者で同じ値を使う文脈情報です。サンプルでは監査データのエクスポート形式の版を表します。復号側に異なる `info` を渡すと、同じ文脈で作られた暗号文ではないため `Open` がエラーを返します。

</details>

---

## 設問 3: なぜ `kem.NewPublicKey` を使う？

通信で受け取った公開鍵バイト列を、なぜ選んだ `kem` の `NewPublicKey` で復元するのでしょうか。KEM の鍵形式と、KEM/KDF/AEAD からなる暗号スイートを分けて、組み合わせをどのように扱うべきか説明してください。

<details>
<summary>ヒント</summary>

- `KEM.NewPublicKey` の説明で、バイト列を何として解釈するかを探す。
- `MLKEM768X25519` を別の KEM に取り替えた場合を考える。
- `NewPublicKey` に KDF や AEAD を渡していないことにも注目し、暗号スイートが公開鍵バイト列から自動的に復元されるかを確認する。
- 同じバイト列でも、別の KEM で読むと別の形式として扱われます。一方、KDF と AEAD は別途選ぶ部品です。この2つの違いを、鍵の形式と暗号スイートの設定表に分けて整理してみましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [KEM](https://pkg.go.dev/crypto/hpke#KEM) の `NewPublicKey` を読む。
2. [kem.go の実装](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/crypto/hpke/kem.go;l=215) で、KEM が自分の形式で公開鍵を復元することを確認する。
3. [Go 1.26 の `crypto/hpke` 説明](https://go.dev/doc/go1.26) に戻る。

**答え**

公開鍵のバイト列は、単独で意味が決まるものではありません。`KEM.NewPublicKey` は、選択済みの KEM に従ってその列を公開鍵として復元・検証します。ここで指定しているのは KEM の鍵形式であり、KDF や AEAD の選択を復元・交渉しているわけではありません。

送受信者は、公開鍵をどの KEM でエンコードしたかと、KEM/KDF/AEAD の暗号スイートを別々のプロトコル情報として取り決めます。受信したバイト列を別の方式の鍵として推測して扱いません。方式の選択や鍵の配布を場当たり的に混ぜないことが、ライブラリ API を正しく使う前提です。

</details>

---

<details>
<summary>こぼれ話</summary>

一通だけ送るなら `Seal` / `Open` が使えます。複数メッセージを送る `Sender` / `Recipient` では、成功した `Seal` と `Open` の呼び出し順を両側でそろえる必要があります。用途に応じて API を選びましょう。

</details>

---

## 調査の入り口

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [01-packages の調べ方](../../README.md)
- [package crypto/hpke](https://pkg.go.dev/crypto/hpke)
