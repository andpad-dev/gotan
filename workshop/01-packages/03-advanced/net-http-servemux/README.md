[進行ガイド・シナリオ一覧](../../../README.md) | [01-packages の調べ方](../../README.md)

# http.ServeMux のルーティング規約を解読せよ

管理画面向け API は `/reports/latest` と `/reports/{id}` を同じ `ServeMux` に登録しました。登録順に頼らず、より具体的なルートが選ばれ、誤ったメソッドには 405 を返します。なんでこうなってるの？背景を調べよう。

次のコードを [Go Playground で動かす](https://go.dev/play/p/fbR5kWMHL2Z) と、リテラルな `latest`、ワイルドカード、メソッド不一致の振る舞いを観測できます。

このシナリオには、Go 1.22 以降のルーティング規則を有効にする `go.mod` を同梱しています。ローカルで試すときは、このディレクトリでコードを実行してください。実行前に `go version`、`go env GOMOD`、`go env GODEBUG` も確認しましょう。

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /reports/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "latest")
	})
	mux.HandleFunc("GET /reports/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "report=%s", r.PathValue("id"))
	})

	for _, path := range []string{"/reports/latest", "/reports/2026-08"} {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		fmt.Printf("GET %s -> %d %q\n", path, recorder.Code, recorder.Body.String())
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/reports/2026-08", nil))
	fmt.Printf("POST /reports/2026-08 -> %d\n", recorder.Code)
}
```

実行結果:

```text
GET /reports/latest -> 200 "latest"
GET /reports/2026-08 -> 200 "report=2026-08"
POST /reports/2026-08 -> 405
```

---

## 設問 1: どのルートが勝つ？

`GET /reports/latest` が `GET /reports/{id}` より先か後かにかかわらず、`GET /reports/latest` はどちらのハンドラーへ届くでしょうか。また、`POST /reports/2026-08` が 405 になる条件を説明してください。

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
3. [routing_tree.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/net/http/routing_tree.go;l=12) の探索順を確認する。

**答え**

`GET /reports/latest` はリテラルな `latest` を要求するため、同じメソッドの `GET /reports/{id}` より一致するリクエスト集合が狭く、より具体的です。したがって登録順によらず `latest` ハンドラーが選ばれます。

`POST /reports/2026-08` はどちらの `GET` パターンにも一致しません。一方で同じパスに別メソッドのパターンがあるため、`ServeMux` は 405 Method Not Allowed を返します。

</details>

---

## 設問 2: どの条件でこの 2 つは登録時に衝突する？

このシナリオの `go.mod`（`go 1.22`）で、`httpmuxgo121` を設定していない通常の挙動を前提にします。API 担当者が `GET /reports/{id}` と `/reports/latest` を登録しようとしています。両方とも一部の GET リクエストに一致しますが、なぜどちらも「常により具体的」とは言えず、`HandleFunc` が panic するのでしょうか。

まず、実際の登録処理を次のコードで確認してください。[Go Playground でこのコードを実行する](https://go.dev/play/p/iMHQSYKMgpE) と、Go 1.22 以降の標準設定では 2 つ目の `HandleFunc` の登録時に panic し、`registered` は出力されません。

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /reports/{id}", func(http.ResponseWriter, *http.Request) {})
	mux.HandleFunc("/reports/latest", func(http.ResponseWriter, *http.Request) {})
	fmt.Println("registered")
}
```

ローカルで panic しない場合は、`go version`、`go env GOMOD`、`go env GODEBUG` を確認し、標準設定と `GODEBUG=httpmuxgo121=1` を付けた場合の結果を比較して、その差が生じる条件も説明してください。

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

1. [Go Playground の共有コード](https://go.dev/play/p/iMHQSYKMgpE) を標準設定で実行し、2つ目の登録時の panic と `registered` が出力されないことを観測する。
2. [Go 1.22 Release Notes](https://go.dev/doc/go1.22) の enhanced routing patterns を読み直し、重なるパターンの扱いを確認する。
3. [http.ServeMux](https://pkg.go.dev/net/http#ServeMux) の Precedence と conflict の説明を読む。
4. [Go, Backwards Compatibility, and GODEBUG](https://go.dev/doc/godebug) の `httpmuxgo121` を確認する。
5. 「neither is more specific」の規則を確認し、[Go 1.27.0 の pattern.go にある `conflictsWith`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/pattern.go;l=219) を読む。

**答え**

このシナリオの `go.mod` で `httpmuxgo121` を設定していない場合、`GET /reports/{id}` は GET に限る代わりに任意の ID を受け、`/reports/latest` は任意のメソッドを受ける代わりに `latest` だけを受けます。`GET /reports/latest` は両方に一致しますが、片方はメソッド、もう片方はパスの方向でしか狭くありません。

片方の一致集合がもう片方の厳密な部分集合ではないため、優先順位を決められません。登録時に panic して曖昧さを早く発見させる設計です。両方を使いたいなら、メソッドまたはパスをそろえて一方を明確に狭くします。

ただし、`GODEBUG=httpmuxgo121=1` を設定していると panic しません。Go 1.22 の互換設定が旧来の `ServeMux` の挙動を復元し、メソッドとワイルドカードを含む新しいパターン規則を使わなくなるためです。したがって、panic の有無を再現するときは、Go のバージョンだけでなく、メインモジュールの `go` 行と `GODEBUG` もそろえる必要があります。

</details>

---

## 設問 3: なぜ文字列の書式変更だけで互換性の議論が必要だった？

Go 1.21 では `{id}` を含むパターンは特別なワイルドカードではありませんでした。Go 1.22 の変更で何が変わり、どの互換設定が用意されたでしょうか。プロジェクトの移行時に確認すべき点を挙げてください。

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
3. [servemux121.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/net/http/servemux121.go;l=25) を読む。

**答え**

Go 1.22 から、メソッド付きパターンと `{name}` / `{name...}` のワイルドカードが解釈されるようになりました。以前は `{id}` はただの文字列として扱われたため、同じ登録が別のリクエストへ一致する可能性があります。

移行用には `httpmuxgo121=1` があり旧挙動を復元できます。まず既存の `{`・`}` を含むパターン、エスケープされたパス、登録時の panic をテストし、設定に恒久依存せず新しいパターン規則に合わせるのが安全です。

</details>

---

## 設問 4: なぜ「最後に登録したものが勝つ」ではない？

提案 Issue と実装をたどり、順序独立の優先順位と登録時の衝突検出が、管理 API のルートを複数チームで保守するときにどんな利点を持つか説明してください。

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
4. [routing_tree.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/net/http/routing_tree.go;l=169) と [pattern.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/net/http/pattern.go;l=223) を読む。

**答え**

登録順で結果が変わると、別チームがルートを追加しただけで既存 API の到達先が変わり得ます。集合としてより具体的なパターンを選べば、設定の並び順ではなくルールから到達先を説明できます。

一方で比較できない重なりは、実行時まで隠さず登録時に panic します。実装は具体的なパターンを先に探索しつつ、衝突検出を起動時に行います。管理 API の変更をレビューするときも、「どのリクエスト集合が増減するか」で議論できるようになります。

</details>

---

<details>
<summary>こぼれ話</summary>

`GET` パターンは `HEAD` にも一致します。`GET /` のような広いパターンを登録すると、別メソッドで一致しないリクエストもそちらへ届くため、405 の期待と合わせてルート全体をテストしましょう。

</details>

---

## 調査の入り口

- [Go 1.22 Release Notes](https://go.dev/doc/go1.22)
- [01-packages の調べ方](../../README.md)
- [package net/http](https://pkg.go.dev/net/http)
