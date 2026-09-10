[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [03-cmd-tools の調べ方](../../README.md)

# テストカバレッジを確認しよう

![実行環境: 指定なし](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-%E6%8C%87%E5%AE%9A%E3%81%AA%E3%81%97-9E9E9E)

README のコードを保存してから進めます。

CI のテスト結果で、`coverage: 66.7%` を見かけました。どんなものか調べてみましょう。

```
PASS
coverage: 66.7% of statements
```

この例では、乗客数に応じて料金を計算する `fare` 関数をテストしています。`fare` には、乗客数が 10 人以上のときに料金を割り引く分岐があります。その分岐が実行されたかはこの数字だけでは分かりません。
`go tool cover` を使って、どこまで試せていて、どこをまだ通っていないかを確かめましょう。

**`main.go`**

```go
package main

import "fmt"

func fare(passengers int) int {
	if passengers <= 0 {
		return 0
	}
	if passengers >= 10 {
		return passengers * 800
	}
	return passengers * 1000
}

func main() {
	fmt.Println(fare(3))
}
```

（[Go Playground で動かす](https://go.dev/play/p/Ew0R-EaEA2s)）

```console
$ go run main.go
3000
```

**`main_test.go`**

```go
package main

import "testing"

func TestFare(t *testing.T) {
	tests := []struct {
		passengers int
		want       int
	}{
		{passengers: 0, want: 0},
		{passengers: 3, want: 3000},
	}

	for _, tt := range tests {
		if got := fare(tt.passengers); got != tt.want {
			t.Fatalf("fare(%d) = %d, want %d", tt.passengers, got, tt.want)
		}
	}
}
```

Go 1.26.4 で次を実行した結果です。

```console
$ go test -coverprofile=cover.out
PASS
coverage: 66.7% of statements
ok  	example.com/cover-demo	0.226s

$ go tool cover -func=cover.out
example.com/cover-demo/main.go:5:	fare		80.0%
example.com/cover-demo/main.go:15:	main		0.0%
total:					(statements)	66.7%
```

---

## 設問 1: 66.7% と 80.0% は、誰の何を数えている？

`go test` は成功しているのに、全体は 66.7%、運賃計算の `fare` だけ見ると 80.0% です。

1. `-coverprofile=cover.out` は何を作るのでしょうか。`go tool cover -func=cover.out` は何を集計しているのでしょうか。
2. 「66.7% だから、運賃計算も 66.7% しか試せていない」と言い切れないのはなぜでしょうか。

<details>
<summary>ヒント</summary>

- まず [Go のコマンド一覧](https://go.dev/doc/cmd) を開き、`cover` を探してみましょう。
- 手元の `go help testflag` で、`-coverprofile` の説明を確認します。
- 次に `go tool cover -help` を実行し、`-func` が受け取るものを確かめましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go のコマンド一覧](https://go.dev/doc/cmd) で `cover` を見つけ、[cover の公式ドキュメント](https://go.dev/cmd/cover/) を開く。ここで、`cover` は `go test -coverprofile=cover.out` が出力する coverage profile を解析するコマンドだと分かる。
2. 手元で `go help testflag` を実行し、`-coverprofile` がテスト成功後に coverage profile を書き出すことを確認する。
3. `go tool cover -help` を実行し、`-func` を使って関数ごとの到達率を表示する。実装上の注意は、[Go 1.26.4 の `cmd/cover` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/cover/doc.go) の説明も読む。

**答え**

- `go test -coverprofile=cover.out` は、テスト実行で通った文の情報を `cover.out` に書き出します。`go tool cover -func=cover.out` はその profile を関数ごと、さらに全体で集計して表示します。
- この例の **66.7%** は `main.go` にある全対象文の合計です。一方の **80.0%** は `fare` だけの値です。`go test` は `main` 関数を実行しないため、`main` が 0.0% のまま全体値を下げています。
- したがって、総合パーセントは「テストの良さ」の合格印ではありません。まず関数別の表示で、今回守りたい料金計算の分岐に話を絞る必要があります。
- `cover` はソースを解析しておおまかな basic block を計測する方式です。公式ドキュメントどおり、`&&` や `||` の内側までは個別に計測しません。数値は調査の入口であって、仕様の代わりではありません。

</details>

---

## 設問 2: 10 人以上の分岐を確認するために、どのテストを足す？

`fare` は 80.0% でした。`main_test.go` と `go tool cover -func` の出力を突き合わせてください。まだ一度も試されていない料金計算の分岐を特定しましょう。

その分岐を検証するテストケースを 1 つ追加したら、次のコマンドでどのように確認しますか？ また、総合値が 100% にならなくても「料金計算の分岐を検証できた」と言える根拠を説明してください。

<details>
<summary>ヒント</summary>

- `fare` の `if` を上から追い、テスト表の `passengers` と比べてください。
- 端の値（境界値）を 1 つ選ぶと、割引が始まる条件を確認できます。
- `-html` の引数を、設問 1 で作った profile に向けると何が起きるか、ヘルプで調べましょう。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [cover の公式ドキュメント](https://go.dev/cmd/cover/) を読み、coverage が「通った basic block」を表すことを確認する。
2. `main_test.go` の入力 `0` と `3` を `fare` の条件と比較する。次に `go tool cover -func=cover.out` を実行して `fare` の値を再確認する。
3. `go tool cover -help` で `-html` の使い方を確認し、`go tool cover -html=cover.out -o cover.html` を実行して、ブラウザで未通過箇所を確認する。

**答え**

- 未通過なのは `passengers >= 10` の料金割引分岐です。ちょうど割引が始まる境界を確かめるなら、次のケースを表に加えます。

    ```go
    {passengers: 10, want: 8000},
    ```

- 追加後は次の順で確かめます。

    ```console
    $ go test -coverprofile=cover.out
    $ go tool cover -func=cover.out
    $ go tool cover -html=cover.out -o cover.html
    ```

    `-html` は profile を色付きの HTML にして、通った箇所と通っていない箇所をソース上で確認できるようにします。
- `fare` が 100.0% になっても、`main` はテスト実行中に呼ばれないので全体値は 100% になりません。ここで大事なのは「料金割引分岐を含む、`fare` の仕様を入力と期待値で検証した」ことです。判断は総合値だけでなく、守るべきルールに対応したテストで行います。

</details>

---

<details>
<summary>こぼれ話</summary>

複数パッケージの実行をまとめて測りたいときは、`go test` の `-coverpkg` で計測対象を指定できます。まずは `go help testflag` の `-coverpkg` と `-covermode` を読み、どのパッケージを計測対象とみなすかを決めてから広げましょう。数値を大きくするためだけに対象を増やすと、今回のように見るべき分岐が埋もれます。

</details>

---

## 調査の入り口

1. [Go のコマンド一覧](https://go.dev/doc/cmd)
2. [cover の公式ドキュメント](https://go.dev/cmd/cover/)
3. 手元の `go help testflag` と `go tool cover -help`
4. [Go 1.26.4 の `cmd/cover` ソース](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/cover/doc.go)
