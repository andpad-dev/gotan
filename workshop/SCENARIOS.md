[ワークショップ進行ガイド](README.md) | [チームでの進め方](TEAM_GUIDE.md)

# シナリオ一覧

取り組みたいテーマと難易度を選び、シナリオを開いてください。各シナリオの先頭から、この一覧、進行ガイド、カテゴリの調べ方に戻れます。

## 実行環境

ほとんどのシナリオは [Go Playground](https://go.dev/play/) だけで進められます。手元に Go が必要なのは次のシナリオです。

| シナリオ | 必要な Go |
| --- | --- |
| [`http.ServeMux`](01-packages/03-advanced/net-http-servemux/README.md) | 1.22 以上 |
| [`time.Timer`](01-packages/03-advanced/time-timer-channels/README.md) | 1.27 以上 |
| [`go run`](03-cmd-tools/01-beginner/go-run/README.md) | 1.24 以上 |
| [`go tool cover`](03-cmd-tools/01-beginner/go-tool-cover/README.md) | 指定なし |
| [`go vet`](03-cmd-tools/01-beginner/go-vet-basics/README.md) | 1.27 以上 |
| [`go fix`](03-cmd-tools/02-intermediate/go-fix-modernize/README.md) | 1.27 以上 |
| [`go generate`](03-cmd-tools/02-intermediate/go-generate/README.md) | 1.21 以上 |
| [`go tool pprof`](03-cmd-tools/02-intermediate/go-tool-pprof/README.md) | 1.27 以上 |
| [cmd/go の script tests](03-cmd-tools/03-advanced/go-script-tests/README.md) | 1.27 |
| [`go tool trace`](03-cmd-tools/03-advanced/go-tool-trace/README.md) | 指定なし |

[`goroutineleak` プロファイル](02-features/03-advanced/goroutine-leak-profile/README.md) は Playground でも実行できます。手元で動かす場合は Go 1.27 が必要です。

手元の Go がシナリオの要求より古いと、`go run` がツールチェーンのダウンロードを始めます。会場では先に Playground を試してください。
同じ情報は各シナリオの先頭にも書いています。

## 01-packages — [標準パッケージの調べ方](01-packages/README.md)

- 初級: [`cmp.Or`](01-packages/01-beginner/cmp-or/README.md) / [`fmt.Printf`](01-packages/01-beginner/fmt-printf/README.md) / [`fmt.Sprintf`](01-packages/01-beginner/fmt-sprintf/README.md) / [for 文の仕様](01-packages/01-beginner/lang-spec-basics/README.md) / [`strings.Cut`](01-packages/01-beginner/strings-cut/README.md)
- 中級: [`bytes.Buffer`](01-packages/02-intermediate/bytes-buffer-peek/README.md) / [`crypto/hpke`](01-packages/02-intermediate/crypto-hpke/README.md) / [`slog.Handler`](01-packages/02-intermediate/slog-handler/README.md)
- 上級: [`http.ServeMux`](01-packages/03-advanced/net-http-servemux/README.md) / [`time.Timer`](01-packages/03-advanced/time-timer-channels/README.md)

## 02-features — [言語機能の調べ方](02-features/README.md)

- 初級: [`iota`](02-features/01-beginner/iota-constants/README.md)
- 中級: [ジェネリックメソッド](02-features/02-intermediate/generic-methods/README.md)
- 上級: [`goroutineleak` プロファイル](02-features/03-advanced/goroutine-leak-profile/README.md)

## 03-cmd-tools — [Go コマンド・ツールの調べ方](03-cmd-tools/README.md)

- 初級: [`go run`](03-cmd-tools/01-beginner/go-run/README.md) / [`go tool cover`](03-cmd-tools/01-beginner/go-tool-cover/README.md) / [`go vet`](03-cmd-tools/01-beginner/go-vet-basics/README.md)
- 中級: [`go fix`](03-cmd-tools/02-intermediate/go-fix-modernize/README.md) / [`go generate`](03-cmd-tools/02-intermediate/go-generate/README.md) / [`go tool pprof`](03-cmd-tools/02-intermediate/go-tool-pprof/README.md)
- 上級: [cmd/go の script tests](03-cmd-tools/03-advanced/go-script-tests/README.md) / [`go tool trace`](03-cmd-tools/03-advanced/go-tool-trace/README.md) / [pkg.go.dev API](03-cmd-tools/03-advanced/pkg-go-dev-api-v1beta/README.md)

## 04-deep-dive — [仕様・実装・設計背景の調べ方](04-deep-dive/README.md)

- 初級: [`defer` の評価](04-deep-dive/01-beginner/defer-evaluation/README.md) / [`encoding/json` の nil スライス](04-deep-dive/01-beginner/json-nil-slice/README.md)
- 中級: [`context.WithoutCancel`](04-deep-dive/02-intermediate/context-without-cancel/README.md) / [3 添字スライス](04-deep-dive/02-intermediate/slice-append-aliasing/README.md)
- 上級: [error とスタックトレース](04-deep-dive/03-advanced/error-stacktrace/README.md) / [浮動小数点の文字列変換](04-deep-dive/03-advanced/float-formatting/README.md) / [Go toolchain の信頼](04-deep-dive/03-advanced/go-toolchain-trust/README.md) / [size-specialized malloc](04-deep-dive/03-advanced/size-specialized-malloc/README.md) / [`slog.Value` の比較](04-deep-dive/03-advanced/slog-value-comparison/README.md)
