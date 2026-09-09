[ワークショップ進行ガイド](README.md) | [チームでの進め方](TEAM_GUIDE.md)

# シナリオ一覧

取り組みたいテーマと難易度を選び、シナリオを開いてください。各シナリオの先頭から、この一覧、進行ガイド、カテゴリの調べ方に戻れます。チュートリアルで使う教材はこの一覧には載せず、[00-tutorial](00-tutorial/README.md) に置いています。

実行環境の欄は、そのシナリオを進めるのに何が要るかを示します。

- **ブラウザだけ** — Go をインストールしていなくても最後まで進められます。
- **手元の Go** — 設問を解くのに手元の Go が要ります。バージョンの指定があるものは、それ以上が必要です。
- **実行しない** — コードを動かさないシナリオです。

「手元の Go」のシナリオにも [Go Playground](https://go.dev/play/) のリンクが載っていることがあります。これはサンプルコードの挙動を見せるためのもので、設問を解くには手元の Go が要ります。
手元の Go がシナリオの要求より古い場合、`go run` はツールチェーンのダウンロードを始めます。回線が細いときは、その時間も見込んで選んでください。

## 01-packages — [標準パッケージの調べ方](01-packages/README.md)

| 難易度 | シナリオ | 実行環境 |
| --- | --- | --- |
| 初級 | [`cmp.Or`](01-packages/01-beginner/cmp-or/README.md) | ブラウザだけ |
| 初級 | [`fmt.Sprintf`](01-packages/01-beginner/fmt-sprintf/README.md) | ブラウザだけ |
| 初級 | [for 文の仕様](01-packages/01-beginner/lang-spec-basics/README.md) | ブラウザだけ |
| 初級 | [`strings.Cut`](01-packages/01-beginner/strings-cut/README.md) | ブラウザだけ |
| 中級 | [`bytes.Buffer`](01-packages/02-intermediate/bytes-buffer-peek/README.md) | ブラウザだけ |
| 中級 | [`crypto/hpke`](01-packages/02-intermediate/crypto-hpke/README.md) | ブラウザだけ |
| 中級 | [`slog.Handler`](01-packages/02-intermediate/slog-handler/README.md) | ブラウザだけ |
| 上級 | [`http.ServeMux`](01-packages/03-advanced/net-http-servemux/README.md) | 手元の Go 1.22 以上 |
| 上級 | [`time.Timer`](01-packages/03-advanced/time-timer-channels/README.md) | 手元の Go 1.27 以上 |

## 02-features — [言語機能の調べ方](02-features/README.md)

| 難易度 | シナリオ | 実行環境 |
| --- | --- | --- |
| 初級 | [`iota`](02-features/01-beginner/iota-constants/README.md) | ブラウザだけ |
| 中級 | [ジェネリックメソッド](02-features/02-intermediate/generic-methods/README.md) | ブラウザだけ |
| 上級 | [`goroutineleak` プロファイル](02-features/03-advanced/goroutine-leak-profile/README.md) | ブラウザだけ |

## 03-cmd-tools — [Go コマンド・ツールの調べ方](03-cmd-tools/README.md)

| 難易度 | シナリオ | 実行環境 |
| --- | --- | --- |
| 初級 | [`go run`](03-cmd-tools/01-beginner/go-run/README.md) | 手元の Go 1.24 以上 |
| 初級 | [`go tool cover`](03-cmd-tools/01-beginner/go-tool-cover/README.md) | 手元の Go |
| 初級 | [`go vet`](03-cmd-tools/01-beginner/go-vet-basics/README.md) | 手元の Go 1.27 以上 |
| 中級 | [`go fix`](03-cmd-tools/02-intermediate/go-fix-modernize/README.md) | 手元の Go 1.27 以上 |
| 中級 | [`go generate`](03-cmd-tools/02-intermediate/go-generate/README.md) | 手元の Go 1.21 以上 |
| 中級 | [`go tool pprof`](03-cmd-tools/02-intermediate/go-tool-pprof/README.md) | 手元の Go 1.27 以上 |
| 上級 | [cmd/go の script tests](03-cmd-tools/03-advanced/go-script-tests/README.md) | 手元の Go 1.27 |
| 上級 | [`go tool trace`](03-cmd-tools/03-advanced/go-tool-trace/README.md) | 手元の Go |
| 上級 | [pkg.go.dev API](03-cmd-tools/03-advanced/pkg-go-dev-api-v1beta/README.md) | 実行しない |

## 04-deep-dive — [仕様・実装・設計背景の調べ方](04-deep-dive/README.md)

| 難易度 | シナリオ | 実行環境 |
| --- | --- | --- |
| 初級 | [`defer` の評価](04-deep-dive/01-beginner/defer-evaluation/README.md) | ブラウザだけ |
| 初級 | [`encoding/json` の nil スライス](04-deep-dive/01-beginner/json-nil-slice/README.md) | ブラウザだけ |
| 中級 | [`context.WithoutCancel`](04-deep-dive/02-intermediate/context-without-cancel/README.md) | ブラウザだけ |
| 中級 | [3 添字スライス](04-deep-dive/02-intermediate/slice-append-aliasing/README.md) | ブラウザだけ |
| 上級 | [error とスタックトレース](04-deep-dive/03-advanced/error-stacktrace/README.md) | ブラウザだけ |
| 上級 | [浮動小数点の文字列変換](04-deep-dive/03-advanced/float-formatting/README.md) | ブラウザだけ |
| 上級 | [Go toolchain の信頼](04-deep-dive/03-advanced/go-toolchain-trust/README.md) | 実行しない |
| 上級 | [size-specialized malloc](04-deep-dive/03-advanced/size-specialized-malloc/README.md) | ブラウザだけ |
| 上級 | [`slog.Value` の比較](04-deep-dive/03-advanced/slog-value-comparison/README.md) | ブラウザだけ |
