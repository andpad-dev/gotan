[ワークショップ進行ガイド](README.md) | [チームでの進め方](TEAM_GUIDE.md)

# シナリオ一覧

取り組みたいテーマと難易度を選び、シナリオを開いてください。各シナリオの先頭から、この一覧、進行ガイド、カテゴリの調べ方に戻れます。

実行環境の欄は、そのシナリオを進めるのに何が要るかを示します。
水色は [Go Playground](https://go.dev/play/) だけで完結するもの、オレンジは手元に Go が要るもの、グレーはコードを実行しないものです。
オレンジのシナリオにも Playground のリンクが載っていることがあります。これはサンプルコードの挙動を見せるためのもので、設問を解くには手元の Go が要ります。
手元の Go がシナリオの要求より古いと `go run` がツールチェーンのダウンロードを始めるので、会場では先に Playground を試してください。

## 01-packages — [標準パッケージの調べ方](01-packages/README.md)

| 難易度 | シナリオ | 実行環境 |
| --- | --- | --- |
| 初級 | [`cmp.Or`](01-packages/01-beginner/cmp-or/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 初級 | [`fmt.Printf`](01-packages/01-beginner/fmt-printf/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 初級 | [`fmt.Sprintf`](01-packages/01-beginner/fmt-sprintf/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 初級 | [for 文の仕様](01-packages/01-beginner/lang-spec-basics/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 初級 | [`strings.Cut`](01-packages/01-beginner/strings-cut/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 中級 | [`bytes.Buffer`](01-packages/02-intermediate/bytes-buffer-peek/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 中級 | [`crypto/hpke`](01-packages/02-intermediate/crypto-hpke/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 中級 | [`slog.Handler`](01-packages/02-intermediate/slog-handler/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 上級 | [`http.ServeMux`](01-packages/03-advanced/net-http-servemux/README.md) | ![実行環境: Go 1.22 以上](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.22%20%E4%BB%A5%E4%B8%8A-F39C12) |
| 上級 | [`time.Timer`](01-packages/03-advanced/time-timer-channels/README.md) | ![実行環境: Go 1.27 以上](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.27%20%E4%BB%A5%E4%B8%8A-F39C12) |

## 02-features — [言語機能の調べ方](02-features/README.md)

| 難易度 | シナリオ | 実行環境 |
| --- | --- | --- |
| 初級 | [`iota`](02-features/01-beginner/iota-constants/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 中級 | [ジェネリックメソッド](02-features/02-intermediate/generic-methods/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 上級 | [`goroutineleak` プロファイル](02-features/03-advanced/goroutine-leak-profile/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |

## 03-cmd-tools — [Go コマンド・ツールの調べ方](03-cmd-tools/README.md)

| 難易度 | シナリオ | 実行環境 |
| --- | --- | --- |
| 初級 | [`go run`](03-cmd-tools/01-beginner/go-run/README.md) | ![実行環境: Go 1.24 以上](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.24%20%E4%BB%A5%E4%B8%8A-F39C12) |
| 初級 | [`go tool cover`](03-cmd-tools/01-beginner/go-tool-cover/README.md) | ![実行環境: 手元の Go](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-%E6%89%8B%E5%85%83%E3%81%AE%20Go-F39C12) |
| 初級 | [`go vet`](03-cmd-tools/01-beginner/go-vet-basics/README.md) | ![実行環境: Go 1.27 以上](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.27%20%E4%BB%A5%E4%B8%8A-F39C12) |
| 中級 | [`go fix`](03-cmd-tools/02-intermediate/go-fix-modernize/README.md) | ![実行環境: Go 1.27 以上](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.27%20%E4%BB%A5%E4%B8%8A-F39C12) |
| 中級 | [`go generate`](03-cmd-tools/02-intermediate/go-generate/README.md) | ![実行環境: Go 1.21 以上](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.21%20%E4%BB%A5%E4%B8%8A-F39C12) |
| 中級 | [`go tool pprof`](03-cmd-tools/02-intermediate/go-tool-pprof/README.md) | ![実行環境: Go 1.27 以上](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.27%20%E4%BB%A5%E4%B8%8A-F39C12) |
| 上級 | [cmd/go の script tests](03-cmd-tools/03-advanced/go-script-tests/README.md) | ![実行環境: Go 1.27](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%201.27-F39C12) |
| 上級 | [`go tool trace`](03-cmd-tools/03-advanced/go-tool-trace/README.md) | ![実行環境: 手元の Go](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-%E6%89%8B%E5%85%83%E3%81%AE%20Go-F39C12) |
| 上級 | [pkg.go.dev API](03-cmd-tools/03-advanced/pkg-go-dev-api-v1beta/README.md) | ![実行環境: 不要](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-%E4%B8%8D%E8%A6%81-9E9E9E) |

## 04-deep-dive — [仕様・実装・設計背景の調べ方](04-deep-dive/README.md)

| 難易度 | シナリオ | 実行環境 |
| --- | --- | --- |
| 初級 | [`defer` の評価](04-deep-dive/01-beginner/defer-evaluation/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 初級 | [`encoding/json` の nil スライス](04-deep-dive/01-beginner/json-nil-slice/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 中級 | [`context.WithoutCancel`](04-deep-dive/02-intermediate/context-without-cancel/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 中級 | [3 添字スライス](04-deep-dive/02-intermediate/slice-append-aliasing/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 上級 | [error とスタックトレース](04-deep-dive/03-advanced/error-stacktrace/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 上級 | [浮動小数点の文字列変換](04-deep-dive/03-advanced/float-formatting/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 上級 | [Go toolchain の信頼](04-deep-dive/03-advanced/go-toolchain-trust/README.md) | ![実行環境: 不要](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-%E4%B8%8D%E8%A6%81-9E9E9E) |
| 上級 | [size-specialized malloc](04-deep-dive/03-advanced/size-specialized-malloc/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
| 上級 | [`slog.Value` の比較](04-deep-dive/03-advanced/slog-value-comparison/README.md) | ![実行環境: Go Playground](https://img.shields.io/badge/%E5%AE%9F%E8%A1%8C%E7%92%B0%E5%A2%83-Go%20Playground-00ADD8) |
