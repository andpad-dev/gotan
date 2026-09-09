[シナリオ一覧](../../../SCENARIOS.md) | [ワークショップ進行ガイド](../../../README.md) | [03-cmd-tools の調べ方](../../README.md)

# pkg.go.dev API(v1)を活用する

**実行環境**: コードは実行しません。

ブラウザと `curl` で調べます。

pkg.go.dev のブラウザ UI には、パッケージを「インポート数」で比較・ソートする機能がありません。
一方、[2026年6月にベータ公開された](https://opensource.googleblog.com/2026/06/a-new-pkggodev-api-for-go.html) **pkg.go.dev API** は、構造化された JSON データを返します。
この API はなぜ今の形なのか、設計の背景を調べましょう。

<details>
<summary>調査の入り口</summary>

まず [03-cmd-tools の調べ方](../../README.md) を開き、`go` コマンドとツールの逆引き手順を確かめます。

そのうえで、次のどれかから入ります。

- [Go Documentation](https://go.dev/doc/) — pkg.go.dev API ドキュメントと Go Modules Reference を探す入口
- [pkg.go.dev API ドキュメント](https://pkg.go.dev/v1/api) — pkg.go.dev API の公式リファレンス
- [pkgsite internal/api](https://pkg.go.dev/golang.org/x/pkgsite/internal/api) — API のレスポンス型を定義しているパッケージのドキュメント
- [pkgsite のソースコード](https://cs.opensource.google/go/x/pkgsite) — pkg.go.dev 自体（pkgsite）のソース
- [Go Modules Reference](https://go.dev/ref/mod) — Go のモジュールシステムの公式リファレンス

</details>

---

## 設問 1: 検索結果をソート・フィルタリングしたい

検索 API (`/v1/search`) の結果は、デフォルトではマッチ度順にソートされます。
しかし、インポート数や更新日時でソートするパラメータは見当たりません。
なぜソート機能が提供されていないのでしょうか？

また、検索結果を絞り込む `filter` パラメータが用意されています。
このフィルター式は SQL や生の正規表現ではなく、「Go 式のサブセット」です。
なぜこの文法が選ばれたのか、調べてみましょう。

次のリクエストを実行してください。
`filter` を適用した結果に `github.com/` 以外のパッケージが含まれていないことを確認します。
結果件数は将来変わり得るため、条件を満たすかどうかだけを表示します。

```bash
curl -L "https://pkg.go.dev/v1/search?q=xyzzy&filter=hasPrefix%28packagePath%2C%20%22github.com%22%29" \
  | jq -c '{hasItems: ((.items | length) > 0), allGithub: ([.items[].packagePath] | all(startswith("github.com/")))}'
```

実行結果:

```text
{"hasItems":true,"allGithub":true}
```

<details>
<summary>ヒント</summary>

- [API ドキュメント](https://pkg.go.dev/v1/api) の「Requests」セクションに filter の説明がある
- filter で使える変数は、各エンドポイントのレスポンス型(JSON フィールド)によって決まる
- SearchResult 型にどんなフィールドがあるか確認してみよう
- インポート数のような人気度指標が含まれているか見てみよう

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. [Go Documentation](https://go.dev/doc/) から [pkg.go.dev API ドキュメント](https://pkg.go.dev/v1/api) の「Routes」セクションを開き、`/v1/search` のレスポンス型を確認する。
2. レスポンス型のリンク（`SearchResult`）をたどり、どんなフィールドが返されるか確認する。
3. pkg.go.dev のソースコードで [SearchResult の定義](https://cs.opensource.google/go/x/pkgsite/+/v0.4.0:internal/api/types.go;l=121) を確認する。

**答え**

**ソート機能がない理由:**

API の `/v1/search` レスポンスには、`packagePath`, `modulePath`, `version`, `synopsis` といった基本情報しか含まれておらず、**インポート数や更新日時などのメタデータは返されません**。

SearchResult の定義は以下の通りです（[pkgsite v0.4.0 のソース](https://cs.opensource.google/go/x/pkgsite/+/v0.4.0:internal/api/types.go;l=121)）。

```text
type SearchResult struct {
    PackagePath string `json:"packagePath"`
    ModulePath  string `json:"modulePath"`
    Version     string `json:"version"`
    Synopsis    string `json:"synopsis"`
}
```

インポート数や更新日時といった情報がレスポンスに含まれていないため、クライアント側でソートすることができません。
また、API ドキュメントで検索結果について定義されている順序は、クエリへの一致度が高い順です。インポート数や更新日時を指定する別のソートパラメータは定義されていません。
関連度以外でソートしたい場合は、検索結果の各パッケージに対して `/v1/package/{path}` や `/v1/imported-by/{path}` を個別に叩いて情報を集める必要があります。

**フィルターが Go 式のサブセットである理由:**

[API ドキュメント](https://pkg.go.dev/v1/api) では、filter は boolean を返す「Go 式のサブセット」として定義されています。利用できる演算子・関数と各 route のフィールドを限定することで、サーバー側で検証・評価できる入力形式にしています。
これは SQL や生の正規表現を filter 全体として受け付ける設計ではありません。ただし、許可された `matches(string, regexp)` 関数の第2引数には正規表現を書けます。API ドキュメントは、この文法を選んだ目的を「安全のため」とまでは説明していないので、そこは事実と推測を分けてください。
また、filter は主に絞り込み用途で、並び替え（例: インポート数降順）の指定はできません。  

***例：特定のドメインのみに絞り込む***

hasPrefix 関数を使って、packagePath が github.com から始まるものだけを抽出する例です。

- Goの式: `hasPrefix(packagePath, "github.com")`
- エンコード後: `hasPrefix%28packagePath%2C%20%22github.com%22%29`

```bash
curl -L "https://pkg.go.dev/v1/search?q=xyzzy&filter=hasPrefix%28packagePath%2C%20%22github.com%22%29" | jq .
```

</details>

---

## 設問 2: パッケージパスが曖昧なときの挙動を調べよう

API ドキュメントには、次のような記述があります。

> Package paths are ambiguous: the same path a/b/c could be package c in module a/b or package b/c in module a. If both exist, the API returns an error with a list of possible modules in its candidates field.

なぜ pkg.go.dev ではパッケージパスが「理論上曖昧になりうる」のでしょうか？
そして、ブラウザ UI と API で挙動がどう異なるのでしょうか？

```bash
curl -L "https://pkg.go.dev/v1/package/golang.org/x/time/rate" \
  | jq -c '{modulePath, path, name}'
```

実行結果:

```text
{"modulePath":"golang.org/x/time","path":"golang.org/x/time/rate","name":"rate"}
```

次に、実在する曖昧なパスを同じ API へ渡します。

```bash
pkgsite="https://pkg.go.dev"
curl -L "$pkgsite/v1/package/github.com/hashicorp/consul/api" \
  | jq -c '{candidateModules: [.candidates[].modulePath]}'
```

実行結果:

```text
{"candidateModules":["github.com/hashicorp/consul/api","github.com/hashicorp/consul"]}
```

<details> 
<summary>ヒント</summary>

- Go のモジュールシステムでは、モジュールパスとパッケージパスが独立している
- golang.org/x/time というモジュールの中に rate パッケージがある場合、パッケージパスは golang.org/x/time/rate になる
- 理論上、golang.org/x/time/rate というモジュールが別に存在する可能性もある
- API ドキュメントには「UI とは異なり、最長マッチを選ばない」と書かれている

</details> 

<details> 
<summary>答え</summary>

**調査ルート**

- [Go Documentation](https://go.dev/doc/) から [Go Modules Reference](https://go.dev/ref/mod) を開き、モジュールパスとパッケージパスの関係を確認する。
- [pkg.go.dev API ドキュメント](https://pkg.go.dev/v1/api) の「Requests」セクションでパッケージパスの曖昧性について読む。
- `golang.org/x/time/rate` と `github.com/hashicorp/consul/api` に対する上の2コマンドを実行し、一意な応答と `candidates` 応答を比較する。
- `?module=github.com/hashicorp/consul/api` を付けて再実行し、候補を明示すると解決できることを確認する。

**答え**

***曖昧性が理論上起こりうる理由:***

Go のモジュールシステムでは、モジュールパスとパッケージパスは独立した概念です。  
同じパス文字列が、異なるモジュール構成で解釈できる可能性があります：

- モジュール a/b の中のパッケージ c → パッケージパス a/b/c
- モジュール a の中のパッケージ b/c → パッケージパス a/b/c
- どちらも同じパッケージパス a/b/c を持ちますが、所属するモジュールが異なります。

***ブラウザ UI と API の挙動の違い:***

- ブラウザ UI: 最長一致するモジュールパスを自動で選ぶ(ユーザーフレンドリー)
- API: 曖昧な場合はエラーを返し、`module` クエリパラメータで明示することを要求する
- API ドキュメントが明記しているのはこの挙動であり、「明示的で安全」という設計意図はそこからの解釈です。

***実際に golang.org/x/time/rate を叩くと：***

上の実行結果のようにエラーは返らず、`golang.org/x/time` モジュールの `rate` パッケージとして解決されます。これは、この入力では所属モジュールを一意に識別できるためです。

実在する `github.com/hashicorp/consul/api` では、上の実行結果のように2つの候補が `candidates` に返ります。この場合、`?module=github.com/hashicorp/consul/api` のように候補のモジュールをクエリパラメータで明示して再リクエストします。

</details>

---

## 設問 3: レート制限とページネーションの設計を調べよう

API には 45 QPS(queries per second)per IP block というレート制限があります。  
また、結果が多い場合は ページネーション で分割して返されます。

なぜこのような制限・設計になっているのか、背景を考えてみましょう。
また、ページネーションの nextPageToken は何を表しているのか調べてみましょう。

次のコマンドでは、1件に制限した検索結果と、次のページが存在するかを確認できます。検索対象の件数は変わり得るため、出力にはページ内の件数と `nextPageToken` の有無だけを表示します。

```bash
curl -L "https://pkg.go.dev/v1/search?q=xyzzy&limit=1" \
  | jq -c '{items: (.items | length), hasNext: (.nextPageToken != null and .nextPageToken != "")}'
```

実行結果:

```text
{"items":1,"hasNext":true}
```

<details> 
<summary>ヒント</summary>

- nextPageToken は不透明な文字列(opaque token)として設計されている
- API ドキュメントには「リクエストを一切変更せず、token だけ追加せよ」と書かれている
- レート制限の数値とエラー応答は事実として確認し、その目的についての考察とは分ける
</details> 
<details> 
<summary>答え</summary>

**調査ルート**

- [Go Documentation](https://go.dev/doc/) から [pkg.go.dev API ドキュメント](https://pkg.go.dev/v1/api) の「Rate Limiting」と「Pagination」セクションを読む。
- 上のコマンドを実行し、`nextPageToken` がある場合に次のページへ進めることを確認する。
- pkgsite のソースコードでページネーションの実装を確認する。

**答え**

***レート制限の理由:***

pkg.go.dev は公開サービスであり、API ドキュメントには 45 QPS per IP block のレート制限と、超過時に `429 Too Many Requests` を返すことが記載されています。

この制限の目的として、次のようなことが考えられます。

- DoS 攻撃や過負荷によるサービス停止
- 特定のユーザーが大量のリクエストでリソースを独占すること
- スクレイピングボット等による過度な利用
- 45 QPS は、通常の利用には十分な値ですが、大規模なバッチ処理には制約となります。

***ページネーションの設計:***

- `nextPageToken` は**不透明なトークン（opaque token）**で、クライアントが内容を解釈することは想定されていません。

ドキュメントには次のように書かれています.

> Changing the request in any way other than providing a token may result in an error.

これは、「ソート順やフィルターを途中で変更すると、トークンが無効になる可能性がある」ことを意味します。
クライアントは、トークンをそのまま渡すだけで次ページを取得できます。次のように元の `q` と `limit` を維持し、`token` だけを追加して確認できます。

```bash
page_token=$(curl -sS "https://pkg.go.dev/v1/search?q=xyzzy&limit=1" | jq -r .nextPageToken)
pkgsite="https://pkg.go.dev"
curl -sS -G "$pkgsite/v1/search" \
  --data-urlencode "q=xyzzy" \
  --data-urlencode "limit=1" \
  --data-urlencode "token=$page_token" \
  | jq -c '{items: (.items | length)}'
```

```text
{"items":1}
```

</details>

---

## 設問 4: imported-by が同一モジュール内を除外する理由を調べよう

`/v1/imported-by/{path}` は、指定したパッケージをインポートしているパッケージの一覧を返します。
しかし、ドキュメントには次のように書かれています:

> Paths of packages importing the package at {path}, not including packages in the same module.

なぜ同一モジュール内のパッケージを除外するのでしょうか？

次のコマンドは `golang.org/x/time/rate` のインポート元の最初のページを取得します。
あわせて `golang.org/x/time/` で始まるパッケージの件数を表示します。

```bash
curl -L "https://pkg.go.dev/v1/imported-by/golang.org/x/time/rate" \
  | jq -c --arg modulePath "golang.org/x/time" '{pageSize: (.importedBy.items | length), sameModuleItems: ([.importedBy.items[] | select(startswith($modulePath + "/"))] | length), hasNext: (.importedBy.nextPageToken != null and .importedBy.nextPageToken != "")}'
```

実行結果:

```text
{"pageSize":100,"sameModuleItems":0,"hasNext":true}
```

<details> 
<summary>ヒント</summary>

- [Go Modules Reference](https://go.dev/ref/mod) で、同じモジュールに属するパッケージが共有するバージョン境界を確認します。
- コマンドの結果と API ドキュメントが保証する範囲を確認してから、モジュール間の影響調査という用途を考えます。
</details> 
<details> 
<summary>答え</summary>

**調査ルート**

- [Go Documentation](https://go.dev/doc/) から [Go Modules Reference](https://go.dev/ref/mod) を開き、モジュールの境界について確認する。
- [pkg.go.dev API ドキュメント](https://pkg.go.dev/v1/api) の `/v1/imported-by/{path}` の説明を読む。
- 上のコマンドを実行し、少なくとも取得したページでは同一モジュールのパッケージが除外されていることを確認する。

**答え**

API ドキュメントが明示している事実は、`imported-by` の結果に同一モジュール内のパッケージを含めないことです。これは、モジュールをまたいだ外部への影響範囲を調べる用途に焦点を当てた設計だと考えられます。

- モジュール内の依存: 同じリポジトリ・チームで管理されており、一緒にリリースされる。内部実装の依存関係。
- モジュール間の依存: 異なるチーム・プロジェクトが依存している。破壊的変更の影響が広範囲に及ぶ。

imported-by の典型的な用途は:

- 「このパッケージを変更したら、どのプロジェクトに影響するか？」
- 「このパッケージの人気度(外部からの利用度)はどのくらいか？」
- 同一モジュール内のパッケージは、変更時に一緒に修正できるため、外部への影響とは性質が異なります。
- したがって API は「外部への影響範囲」に焦点を当てた設計だと解釈できます。

</details>
