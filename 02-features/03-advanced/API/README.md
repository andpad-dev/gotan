# 公式API（v1beta）を活用する

https://pkg.go.dev/ の結果を「インポート数などで比較・ソートしたい！」という場合は、2026年6月にベータ公開された pkg.go.dev API を使うのが現在のベストプラクティスです。

これまではスクレイピングに頼るしかありませんでしたが、現在は [https://pkg.go.dev/v1beta/](https://pkg.go.dev/v1beta/)... のエンドポイントから構造化されたJSONデータを直接GETできるようになりました 🎉

## 設問 1: 

検索結果をJSONで取得してみましょう。

<details>
<summary>答え</summary>

**調査ルート**

1. https://pkg.go.dev/v1beta/api にアクセスして、APIのエンドポイントを確認する。

**答え**

```bash
# 検索結果をJSONで取得する例
curl -s "https://pkg.go.dev/v1beta/search?q=router" | jq .
```

</details>

---

<details>
<summary>こぼれ話: </summary>

</details>

---
