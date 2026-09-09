[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [01-packages の調べ方](../../README.md)

# http.ServeMux のルーティング規約を解読せよ

**実行環境**: 手元の Go 1.22 以上が必要です。

同じ `ServeMux` に `/x/fixed` と `/x/{value}` を登録しました。すると**登録順に頼らず**、より具体的なルートが選ばれます。誤ったメソッドには **405** が返ります。なんでこうなってるの？背景を調べよう。

次のコードは [Go Playground で動かせます](https://go.dev/play/p/DhW53RiIwwY)。リテラル一致、ワイルドカード一致、メソッド不一致の三つを観測できます。

このディレクトリには、上のコードの `main.go` と `go.mod` を同梱しています。`go.mod` は Go 1.22 以降のルーティング規則を有効にします。ローカルでは `go run .` を実行してください。

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /x/fixed", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "fixed")
	})
	mux.HandleFunc("GET /x/{value}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "value=%s", r.PathValue("value"))
	})

	for _, path := range []string{"/x/fixed", "/x/other"} {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		fmt.Printf("GET %s -> %d %q\n", path, recorder.Code, recorder.Body.String())
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/x/other", nil))
	fmt.Printf("POST /x/other -> %d\n", recorder.Code)
}
```

実行結果:

```text
GET /x/fixed -> 200 "fixed"
GET /x/other -> 200 "value=other"
POST /x/other -> 405
```

<details>
<summary>調査の入り口</summary>

まず [01-packages の調べ方](../../README.md) を開き、標準パッケージのドキュメントの開き方を確かめます。

そのうえで、次のどれかから入ります。

- [Go 1.22 Release Notes](https://go.dev/doc/go1.22) — `ServeMux` のルーティング規則が変わったときのリリースノート
- [package net/http](https://pkg.go.dev/net/http) — `ServeMux` 型の説明にルーティング規則が書かれている

</details>

---

## 設問 1: どのルートが勝つ？

登録が先でも後でも、`GET /x/fixed` はどちらのハンドラーへ届くでしょうか。また、`POST /x/other` が 405 になる条件を説明してください。

<details>
<summary>ヒント</summary>

- [カテゴリの調べ方](../../README.md) を入口に `ServeMux` の Patterns と Precedence を読む。
- 「より具体的」の定義を、文字列の長さではなく「どのリクエスト集合に一致するか」で考える。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.22 Release Notes](https://go.dev/doc/go1.22) の enhanced routing patterns を読む。
2. [http.ServeMux](https://pkg.go.dev/net/http#ServeMux) の Patterns と Precedence を読む。
3. [Go 1.27.0 の `routing_tree.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/routing_tree.go;l=154-198) で、リテラル、単一ワイルドカード、複数ワイルドカードの順に探索する処理を確認する。
4. [Go 1.27.0 の `findHandler`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/server.go;l=2751-2764) で、パスに一致する別メソッドがあれば 404 ではなく 405 と `Allow` ヘッダーを返す処理を確認する。

**答え**

`GET /x/fixed` はリテラルな `fixed` を要求するため、同じメソッドの `GET /x/{value}` より一致するリクエスト集合が狭く、より具体的です。したがって登録順によらず `fixed` ハンドラーが選ばれます。

`POST /x/other` はどちらの `GET` パターンにも一致しません。一方で同じパスに別メソッドのパターンがあるため、`ServeMux` は 405 Method Not Allowed を返します。

</details>

---

## 設問 2: どの条件で登録時に衝突する？

**前提**: このシナリオの `go.mod`（`go 1.22`）のまま、`httpmuxgo121` は設定しません。`GET /x/{value}` と `/x/fixed` を登録します。両方とも一部の GET リクエストに一致します。どちらのパターンも、相手より「具体的」とは言い切れません。なぜそうなるのか、そして `HandleFunc` がなぜ panic するのかを調べてください。

まず登録処理を次のコードで確認してください。[Go Playground で実行する](https://go.dev/play/p/-wtRzMO5UHB)と、2 つ目の `HandleFunc` が **panic** します。`registered` は出力されません。

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /x/{value}", func(http.ResponseWriter, *http.Request) {})
	mux.HandleFunc("/x/fixed", func(http.ResponseWriter, *http.Request) {})
	fmt.Println("registered")
}
```

ローカルで panic しない場合は、実行環境を確認してください。見る対象は `go version`、`go env GOMOD`、`go env GODEBUG` の三つです。次に、標準設定と `GODEBUG=httpmuxgo121=1` の結果を比較してください。差が生じる条件も説明してください。

<details>
<summary>ヒント</summary>

- 一方はメソッドが狭くパスが広い、もう一方はメソッドが広くパスが狭い。
- 比較する集合を「メソッドとパスの組」として考える。メソッド省略は全メソッドに一致し、`GET` は `HEAD` にも一致することを確認する。
- `ServeMux` の conflict の定義を確認する。
- `GODEBUG` がルーティング規則を切り替える条件も確認する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Playground の共有コード](https://go.dev/play/p/-wtRzMO5UHB) を標準設定で実行し、2つ目の登録時の panic と `registered` が出力されないことを観測する。
2. [Go 1.22 Release Notes](https://go.dev/doc/go1.22) の enhanced routing patterns を読み直し、重なるパターンの扱いを確認する。
3. [http.ServeMux](https://pkg.go.dev/net/http#ServeMux) の Precedence と conflict の説明を読む。
4. [Go, Backwards Compatibility, and GODEBUG](https://go.dev/doc/godebug) の `httpmuxgo121` を確認する。
5. 「neither is more specific」の規則を確認し、[Go 1.27.0 の pattern.go にある `conflictsWith`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/pattern.go;l=219-240) を読む。

**答え**

このシナリオの `go.mod` で `httpmuxgo121` を設定していない場合、`GET /x/{value}` は GET に限る代わりに任意の値を受け、`/x/fixed` は任意のメソッドを受ける代わりに `fixed` だけを受けます。`GET /x/fixed` は両方に一致しますが、片方はメソッド、もう片方はパスの方向でしか狭くありません。

片方の一致集合がもう片方の厳密な部分集合ではないため、優先順位を決められません。登録時に panic して曖昧さを早く発見させる設計です。両方を使いたいなら、メソッドまたはパスをそろえて一方を明確に狭くします。

ただし、`GODEBUG=httpmuxgo121=1` を設定していると panic しません。Go 1.22 の互換設定が旧来の `ServeMux` の挙動を復元し、メソッドとワイルドカードを含む新しいパターン規則を使わなくなるためです。したがって、panic の有無を再現するときは、Go のバージョンだけでなく、メインモジュールの `go` 行と `GODEBUG` もそろえる必要があります。

</details>

---

## 設問 3: なぜ文字列の書式変更だけで互換性の議論が必要だった？

Go 1.21 では `{id}` を含むパターンは特別なワイルドカードではありませんでした。Go 1.22 の変更で何が変わり、どの互換設定が用意されたでしょうか。移行時に確認すべき点を挙げてください。

<details>
<summary>ヒント</summary>

- Go 1.22 リリースノートで `{` と `}` がどう扱われたかを探す。
- `httpmuxgo121` を検索する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.22 Release Notes](https://go.dev/doc/go1.22) の互換性に関する段落を読む。
2. [Go, Backwards Compatibility, and GODEBUG](https://go.dev/doc/godebug) の `httpmuxgo121` の説明を読む。
3. [Go 1.27.0 の `servemux121.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/servemux121.go;l=25-38) を読み、`httpmuxgo121=1` を起動時に一度だけ判定することを確認する。

**答え**

Go 1.22 から、メソッド付きパターンと `{name}` / `{name...}` のワイルドカードが解釈されるようになりました。以前は `{id}` はただの文字列として扱われたため、同じ登録が別のリクエストへ一致する可能性があります。

移行用には `httpmuxgo121=1` があり旧挙動を復元できます。まず既存の `{`・`}` を含むパターン、エスケープされたパス、登録時の panic をテストし、設定に恒久依存せず新しいパターン規則に合わせるのが安全です。

</details>

---

## 設問 4: なぜ「最後に登録したものが勝つ」ではない？

提案 Issue と実装をたどってください。複数箇所からルートを登録するとき、順序独立の優先順位と登録時の衝突検出にはどんな利点があるでしょうか。

<details>
<summary>ヒント</summary>

- 提案 Issue の Precedence と Performance を読む。
- `routing_tree.go` のコメントに、より具体的なルートを先に試す理由がある。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go 1.22 Release Notes](https://go.dev/doc/go1.22) の enhanced routing patterns を読み直す。
2. [ServeMux 拡張の提案 Issue #61410](https://github.com/golang/go/issues/61410) の Precedence、Backwards Compatibility、Performance を読む。
3. [Go Blog: Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements) を読む。
4. [Go 1.27.0 の `routing_tree.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/routing_tree.go;l=168-198) と [`pattern.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/pattern.go;l=219-240) を読み、具体的なパターンから探索する処理と登録時の衝突判定を対応付ける。

**答え**

登録順で結果が変わると、別の場所でルートを追加しただけで既存の到達先が変わり得ます。集合としてより具体的なパターンを選べば、設定の並び順ではなくルールから到達先を説明できます。

一方で比較できない重なりは、実行時まで隠さず登録時に panic します。実装は具体的なパターンを先に探索しつつ、衝突検出を起動時に行います。変更時も「どのリクエスト集合が増減するか」で確認できます。

</details>

---

<details>
<summary>こぼれ話</summary>

[`GET` パターンは `HEAD` にも一致します](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/pattern.go;l=253-279)が、POST など他のメソッドには一致しません。`GET /` だけを登録した場合、`POST /anything` はそのハンドラーへ届かず、[`Allow: GET, HEAD` を伴う 405](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/server.go;l=2751-2764) になります。すべてのメソッドを同じハンドラーへ届けたい場合は、メソッドを省略した `/` を登録します。`GET` とメソッド省略を混同せず、405 と `Allow` ヘッダーも含めてルート全体をテストしましょう。

</details>
