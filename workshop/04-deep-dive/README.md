# 04-deep-dive の調べ方（逆引き手順）

このカテゴリでは、「なんでこうなってるの？」を、言語仕様や標準ライブラリ・ランタイムの **設計判断の背景** まで掘り下げて調べます。
答えは 1 つのドキュメントには載っていないことが多いので、リリースノート → 課題（issue）→ 変更（CL）→ ソースコード、と一次情報をたどるのがコツです。困ったときは、まずここに戻ってきてください。

## 新機能の「なぜ」を知りたい

1. まず [Go 1.26](https://go.dev/doc/go1.26) / [Go 1.27](https://go.dev/doc/go1.27) のリリースノートで該当項目を読む。
2. 本文に issue 番号があれば、その GitHub issue を **議論も含めて丸ごと** 読む。要約だけで結論を出さない。
3. issue から実際の変更（CL）へたどり、コミットメッセージとレビューコメントを読む。

## 変更（CL）を探したい

- コミット全文検索: `gh search commits --repo golang/go "<キーワード>"`。
- Gerrit（コードレビュー）: [go-review.googlesource.com](https://go-review.googlesource.com/) で機能名を検索する。コミットメッセージに設計意図が書かれていることが多い。

## ソースコードの該当箇所を読みたい

- まとめて読むなら [cs.opensource.google/go/go](https://cs.opensource.google/go/go)。
- GitHub で読むなら、挙動が変わらないよう **バージョン付きのref** でリンクする（例: `refs/tags/go1.26.5`、未リリース版は `release-branch.go1.XX`）。
- 定数やしきい値の「なぜその値？」は、多くの場合その宣言のすぐ上のコメントに書かれている。まずコメントを疑ってかかる。

## 挙動を確かめたい

- 手元で `go run` する。GOEXPERIMENT で切り替わる機能なら、`GOEXPERIMENT=<name>` の有無で `go build -gcflags=-S`（アセンブリ出力）を見比べると、生成コードの違いが分かる。
- 手元に環境がなければ [Go Playground](https://go.dev/play/) を使う（ただし GOEXPERIMENT の切り替えはできない点に注意）。

## 一次情報の入り口

- 設計文書: [github.com/golang/proposal](https://github.com/golang/proposal)
- 課題管理: [github.com/golang/go/issues](https://github.com/golang/go/issues)
- 設計解説ブログ: [research.swtch.com](https://research.swtch.com/) 、[golang.design/history](https://golang.design/history/)
