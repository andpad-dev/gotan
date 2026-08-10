# 管制塔のルーティング規約を解読せよ

月面港 API は `/flights/latest` と `/flights/{id}` を同じ管制塔に登録しました。登録順に頼らず、より具体的なルートが選ばれ、誤ったメソッドには 405 を返します。なんでこうなってるの？背景を調べよう。

次のコードを [Go Playground で動かす](https://go.dev/play/p/xgCqZEBJCz3) と、リテラルな `latest`、ワイルドカード、メソッド不一致の振る舞いを観測できます。

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /flights/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "latest")
	})
	mux.HandleFunc("GET /flights/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "flight=%s", r.PathValue("id"))
	})

	for _, path := range []string{"/flights/latest", "/flights/M-17"} {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		fmt.Printf("GET %s -> %d %q\n", path, recorder.Code, recorder.Body.String())
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/flights/M-17", nil))
	fmt.Printf("POST /flights/M-17 -> %d\n", recorder.Code)
}
```

実行結果:

```text
GET /flights/latest -> 200 "latest"
GET /flights/M-17 -> 200 "flight=M-17"
POST /flights/M-17 -> 405
```

---

## 設問 1: どのルートが勝つ？

`GET /flights/latest` が `GET /flights/{id}` より先か後かにかかわらず、`GET /flights/latest` はどちらのハンドラーへ届くでしょうか。また、`POST /flights/M-17` が 405 になる条件を説明してください。

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

`GET /flights/latest` はリテラルな `latest` を要求するため、同じメソッドの `GET /flights/{id}` より一致するリクエスト集合が狭く、より具体的です。したがって登録順によらず `latest` ハンドラーが選ばれます。

`POST /flights/M-17` はどちらの `GET` パターンにも一致しません。一方で同じパスに別メソッドのパターンがあるため、`ServeMux` は 405 Method Not Allowed を返します。

</details>

---

## 設問 2: なぜこの 2 つは登録時に衝突する？

管制官が `GET /flights/{id}` と `/flights/latest` を登録しようとしています。両方とも一部の GET リクエストに一致しますが、なぜどちらも「常により具体的」とは言えず、`HandleFunc` が panic するのでしょうか。

<details>
<summary>ヒント</summary>

- 一方はメソッドが狭くパスが広い、もう一方はメソッドが広くパスが狭い。
- `ServeMux` の conflict の定義を確認する。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [http.ServeMux](https://pkg.go.dev/net/http#ServeMux) の Precedence と conflict の説明を読む。
2. [Go 1.22 Release Notes](https://go.dev/doc/go1.22) で「neither is more specific」の規則を確認する。
3. [pattern.go の `conflictsWith`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/net/http/pattern.go;l=219) を読む。

**答え**

`GET /flights/{id}` は GET に限る代わりに任意の ID を受け、`/flights/latest` は任意のメソッドを受ける代わりに `latest` だけを受けます。`GET /flights/latest` は両方に一致しますが、片方はメソッド、もう片方はパスの方向でしか狭くありません。

片方の一致集合がもう片方の厳密な部分集合ではないため、優先順位を決められません。登録時に panic して曖昧さを早く発見させる設計です。両方を使いたいなら、メソッドまたはパスをそろえて一方を明確に狭くします。

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

提案 Issue と実装をたどり、順序独立の優先順位と登録時の衝突検出が、管制塔の設定を複数チームで保守するときにどんな利点を持つか説明してください。

<details>
<summary>ヒント</summary>

- 提案 Issue の Precedence と Performance を読む。
- `routing_tree.go` のコメントに、より具体的なルートを先に試す理由がある。

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [ServeMux 拡張の提案 Issue #61410](https://github.com/golang/go/issues/61410) の Precedence、Backwards Compatibility、Performance を読む。
2. [Go Blog: Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements) を読む。
3. [routing_tree.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/net/http/routing_tree.go;l=169) と [pattern.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/net/http/pattern.go;l=223) を読む。

**答え**

登録順で結果が変わると、別チームがルートを追加しただけで既存 API の到達先が変わり得ます。集合としてより具体的なパターンを選べば、設定の並び順ではなくルールから到達先を説明できます。

一方で比較できない重なりは、実行時まで隠さず登録時に panic します。実装は具体的なパターンを先に探索しつつ、衝突検出を起動時に行います。管制 API の変更をレビューするときも、「どのリクエスト集合が増減するか」で議論できるようになります。

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
