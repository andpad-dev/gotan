# pkg.go.dev API(v1beta)を活用する

チームで使う HTTP ルーターライブラリを選定することになりました。
「インポート数」や「メンテナンス状況」で比較・ソートしたいのですが、pkg.go.dev のブラウザ UI にはそうした機能がありません。

スクレイピングも検討しましたが、2026年6月にベータ公開された **pkg.go.dev API** を使えば、構造化された JSON データを直接取得できます。
なぜこの API はこのような設計になっているのか、背景を調べましょう。

## 設問 1: 検索結果をソート・フィルタリングしたい

検索 API (`/v1beta/search`) で `router` を検索すると、デフォルトではマッチ度順にソートされます。
しかし、インポート数や更新日時でソートするパラメータは見当たりません。
なぜソート機能が提供されていないのでしょうか？

また、検索結果を絞り込むための `filter` パラメータが用意されています。
このフィルター式は SQL や正規表現ではなく、「Go 式のサブセット」という独特な仕様になっています。
なぜこのような設計になっているのか、調べてみましょう。

<details>
<summary>ヒント</summary>

- [API ドキュメント](https://pkg.go.dev/v1beta/api) の「Requests」セクションに filter の説明がある
- filter で使える変数は、各エンドポイントのレスポンス型(JSON フィールド)によって決まる
- SearchResult 型にどんなフィールドがあるか確認してみよう
- インポート数のような人気度指標が含まれているか見てみよう

</details>

<details>
<summary>答え</summary>

**調査ルート**

1. https://pkg.go.dev/v1beta/api の「Routes」セクションで `/v1beta/search` のレスポンス型を確認する。
2. レスポンス型のリンク(`SearchResult`)をたどり、どんなフィールドが返されるか確認する。
3. pkg.go.dev のソースコード(https://cs.opensource.google/go/x/pkgsite)で SearchResult の定義を見る。

**答え**

**ソート機能がない理由:**

API の `/v1beta/search` レスポンスには、`packagePath`, `modulePath`, `version`, `synopsis` といった基本情報しか含まれておらず、**インポート数や更新日時などのメタデータは返されません**。

SearchResult の定義は以下の通りです(https://cs.opensource.google/go/x/pkgsite/+/refs/tags/v0.3.0:internal/api/types.go;l=115-121):

```go
type SearchResult struct {
    PackagePath string `json:"packagePath"`
    ModulePath  string `json:"modulePath"`
    Version     string `json:"version"`
    Synopsis    string `json:"synopsis"`
}
```

インポート数や更新日時といった情報がレスポンスに含まれていないため、クライアント側でソートすることができません。
また、API 側でソートパラメータを提供していないのは、検索結果が関連度順で返される設計だからです。
関連度以外でソートしたい場合は、検索結果の各パッケージに対して /v1beta/package/{path} や /v1beta/imported-by/{path} を個別に叩いて情報を集める必要があります。

**フィルターが Go 式のサブセットである理由:**

フィルターは、「Go エンジニアにとって馴染みのある文法で、安全に評価できる式」として設計されています。

SQL や正規表現だと、インジェクション攻撃のリスクがある
Go 式のサブセットなら、Go 開発者が直感的に書ける
サーバー側で安全に評価できる範囲に制限されている(演算子・関数を限定)
ただし、== や contains() などの比較しかできないため、「インポート数が多い順」のようなソートには使えません。

***例：特定のドメインのみに絞り込む***

hasPrefix 関数を使って、packagePath が github.com から始まるものだけを抽出する例です。

- Goの式: `hasPrefix(packagePath, "github.com")`
- エンコード後: `hasPrefix%28packagePath%2C%20%22github.com%22%29`

```bash
curl -L "https://pkg.go.dev/v1beta/search?q=xyzzy&filter=hasPrefix%28packagePath%2C%20%22github.com%22%29" | jq .
```

</details>

---

## 設問 2: パッケージパスが曖昧なときの挙動を調べよう

API ドキュメントには、次のような記述があります。

> Package paths are ambiguous: the same path a/b/c could be package c in module a/b or package b/c in module a. If both exist, the API returns an error with a list of possible modules in its candidates field.

なぜ pkg.go.dev ではパッケージパスが「理論上曖昧になりうる」のでしょうか？
そして、ブラウザ UI と API で挙動がどう異なるのでしょうか？

```bash
curl -L "https://pkg.go.dev/v1beta/package/golang.org/x/time/rate" | jq .
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

- curl -s "https://pkg.go.dev/v1beta/package/golang.org/x/time/rate" | jq . を実行して、実際の挙動を確認する。
- API ドキュメントの「Requests」セクションでパッケージパスの曖昧性について読む。
- Go のモジュールシステムのドキュメント(https://go.dev/ref/mod)で、モジュールパスとパッケージパスの関係を確認する。

**答え**

***曖昧性が理論上起こりうる理由:***

Go のモジュールシステムでは、モジュールパスとパッケージパスは独立した概念です。  
同じパス文字列が、異なるモジュール構成で解釈できる可能性があります：

- モジュール a/b の中のパッケージ c → パッケージパス a/b/c
- モジュール a の中のパッケージ b/c → パッケージパス a/b/c
- どちらも同じパッケージパス a/b/c を持ちますが、所属するモジュールが異なります。

***実際には稀な理由:***

実際のGoエコシステムでは、このような曖昧性はほとんど発生しません。なぜなら：

- 1つのリポジトリに複数のモジュールを配置することは推奨されていない
- サブディレクトリに別モジュールを配置する場合、パスが重複しないよう設計される
- Go Modulesの慣習として、モジュールはリポジトリのルートに配置されることが一般的

***ブラウザ UI と API の挙動の違い:***

- ブラウザ UI: 最長一致するモジュールパスを自動で選ぶ(ユーザーフレンドリー)
- API: 曖昧な場合はエラーを返し、module クエリパラメータで明示することを要求する(明示的で安全)
- API の設計は、「暗黙の推測でクライアントに予期しない結果を返すより、明示的にエラーを返して再試行を促す」方針です。

***実際に golang.org/x/time/rate を叩くと：***

```json
{
  "modulePath": "golang.org/x/time",
  "version": "v0.15.0",
  "path": "golang.org/x/time/rate",
  "name": "rate",
  ...
}
```

エラーは返らず正常なレスポンスが返ります。これは golang.org/x/time というモジュールの rate パッケージだと一意に識別できるためです。

曖昧なケースに遭遇した場合、APIは次のようなエラーレスポンスを返します.

```json
{
  "code": 400,
  "message": "ambiguous package path",
  "candidates": ["module1", "module2"]
}
```

この場合、?module=module1 のようにクエリパラメータで明示的にモジュールを指定して再リクエストします。

</details>

---

## 設問 3: レート制限とページネーションの設計を調べよう

API には 45 QPS(queries per second)per IP block というレート制限があります。  
また、結果が多い場合は ページネーション で分割して返されます。

なぜこのような制限・設計になっているのか、背景を考えてみましょう。  
また、ページネーションの nextPageToken は何を表しているのか調べてみましょう。

<details> 
<summary>ヒント</summary>

- pkg.go.dev は Google が運営する公開サービスである
- レート制限は DoS 攻撃の防止やリソース保護のため
- nextPageToken は不透明な文字列(opaque token)として設計されている
- API ドキュメントには「リクエストを一切変更せず、token だけ追加せよ」と書かれている
</details> 
<details> 
<summary>答え</summary>

**調査ルート**

- API ドキュメントの「Rate Limiting」と「Pagination」セクションを読む。
- 実際に検索結果を取得し、nextPageToken がどのような値になっているか確認する。
- pkgsite のソースコード(https://cs.opensource.google/go/x/pkgsite)で、ページネーションの実装を確認する。

**答え**

***レート制限の理由:***

pkg.go.dev は Google が無料で提供する公開サービスです。  
レート制限(45 QPS per IP block)は、以下を防ぐために設定されています。

- DoS 攻撃や過負荷によるサービス停止
- 特定のユーザーが大量のリクエストでリソースを独占すること
- スクレイピングボット等による過度な利用
- 45 QPS は、通常の利用には十分な値ですが、大規模なバッチ処理には制約となります。
- 超過すると 429 Too Many Requests が返されます。

***ページネーションの設計:***

- nextPageToken は**不透明なトークン(opaque token)**で、内部的にはページ位置を示す情報がエンコードされています。
- 実際の値を見ると、長い16進数文字列(例: b689a63f15295533e6320e94470fbf1548acd25c327dff2a858024431056ffe4756658644ee6e2855482af42eb0bd15721a84772854ffd926b97db49479958d4d181f283)になっており、クライアントが解釈することは想定されていません。

ドキュメントには次のように書かれています.

> Changing the request in any way other than providing a token may result in an error.

これは、「ソート順やフィルターを途中で変更すると、トークンが無効になる」ことを意味します。
クライアントは、トークンをそのまま渡すだけで次ページを取得できる設計です。

</details>

---

## 設問 4: imported-by が同一モジュール内を除外する理由を調べよう

/v1beta/imported-by/{path} は、指定したパッケージをインポートしているパッケージの一覧を返します。  
しかし、ドキュメントには次のように書かれています:

> Paths of packages importing the package at {path}, not including packages in the same module.

なぜ同一モジュール内のパッケージを除外するのでしょうか？

<details> 
<summary>ヒント</summary>

- モジュール内の依存関係と、モジュール間の依存関係は性質が異なる
- imported-by の用途として、「このパッケージの影響範囲」を知りたい場合が多い
- 同一モジュール内のパッケージは、通常同じチーム・リポジトリで管理されている
</details> 
<details> 
<summary>答え</summary>

**調査ルート**

- API ドキュメントの `/v1beta/imported-by/{path}` の説明を読む。
- pkgsite のソースコード(https://cs.opensource.google/go/x/pkgsite)で、imported-by の実装を確認する。
- Go のモジュールシステムのドキュメント(https://go.dev/ref/mod)で、モジュールの境界について確認する。

**答え**

同一モジュール内のパッケージを除外する理由は、「外部への影響範囲」を知りたいユースケースが主だからです。

- モジュール内の依存: 同じリポジトリ・チームで管理されており、一緒にリリースされる。内部実装の依存関係。
- モジュール間の依存: 異なるチーム・プロジェクトが依存している。破壊的変更の影響が広範囲に及ぶ。

imported-by の典型的な用途は:

- 「このパッケージを変更したら、どのプロジェクトに影響するか？」
- 「このパッケージの人気度(外部からの利用度)はどのくらいか？」
- 同一モジュール内のパッケージは、変更時に一緒に修正できるため、外部への影響とは性質が異なります。
- API は「外部への影響範囲」に焦点を当てた設計になっています。

</details>

---

## 調査の入り口
- https://pkg.go.dev/v1beta/api
- https://pkg.go.dev/golang.org/x/pkgsite/internal/api (レスポンス型の定義)
- https://cs.opensource.google/go/x/pkgsite (pkgsite のソースコード)
- https://go.dev/ref/mod (Go Modules リファレンス)
