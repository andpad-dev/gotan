[シナリオ一覧](SCENARIOS.md) | [ワークショップ進行ガイド](README.md)

# 画面付きの調査・回答例

各問題の解答にある「画面付きの調査・回答例」から、設問に対応する調査記録へ進めます。この一覧も**答えを含む資料**です。まず自分たちで問題に取り組んでから、調べ方を振り返るために使ってください。

## 何を記録しているか

「どのリンクを開けば答えがあるか」だけではなく、問題文の何を手掛かりにしたか、資料のどの部分を読んだか、それをどう解釈したか、残った疑問からなぜ次の資料へ進んだかを記録しています。ひとつの調査例であり、同じ経路を暗記する必要はありません。

- **出発点**: 問題 README の本文とカテゴリ README。既存解答は、調査経路を決める手掛かりにしません。
- **読む範囲**: 定義・条件・例外と、その前後。検索に一致した一文だけで結論にしません。
- **原文と解釈**: 判断の根拠になる英語の語句・API名・条件式を残し、日本語で意味を説明してから問題へ当てはめます。動作の主体、対象、時点、条件を省いた訳だけで結論にしません。
- **次の一手**: 直前の資料で得た語・リンクと、まだ説明できない点を結び付けます。
- **確認の境界**: 仕様、対象版の実装、手元で測った結果、未確認事項を区別します。
- **画面**: 実際に閲覧した公開ページのキャプチャーです。画像の前後に、読む場所と注目する条件を説明しています。HTTP取得やローカル実行をブラウザーで行った操作としては記録しません。

画像は2026-10-02の撮影記録です。2026-10-13に資料・リンク・実行結果を再確認して本文を整えています。画像の表示内容と、後日の確認を同じ時点の記録として扱わないでください。2026-10-14には、原文の語句と日本語の解釈の対応を一次資料で確認して補足しました。

## 調査が止まったとき

ページが表示できないことを、機能が存在しない証拠にはしません。カテゴリの手順へ戻り、同じ対象・版の別の一次資料を確認できるか調べます。そこで得ていないAPI名、Issue番号、実装ファイル名を、知識や推測で補って経路を完成させないでください。確認できなかった部分はそのまま残します。

調査例を追加・更新する場合は、本文とカテゴリ案内だけで出発し直し、画像とURL、対象版、設問への戻りリンクも更新してください。単なる画面の見た目の変更と、仕様の変更は区別して扱います。

## 問題別の回答例

### チュートリアル

| 教材 | 設問数 | 回答例 |
| --- | ---: | --- |
| [fmt.Printf の書式指定子を調べよう](00-tutorial/README.md) | 2 | [調査の記録](00-tutorial/RESEARCH.md) |

### 01-packages

| 問題 | 設問数 | 回答例 |
| --- | ---: | --- |
| [cmp.Or で通知先チャンネルを決めよう](01-packages/01-beginner/cmp-or/README.md) | 2 | [調査の記録](01-packages/01-beginner/cmp-or/RESEARCH.md) |
| [fmt.Sprintf で文字列を組み立てよう](01-packages/01-beginner/fmt-sprintf/README.md) | 2 | [調査の記録](01-packages/01-beginner/fmt-sprintf/RESEARCH.md) |
| [Go の for 文と switch 文を仕様書で調べよう](01-packages/01-beginner/lang-spec-basics/README.md) | 2 | [調査の記録](01-packages/01-beginner/lang-spec-basics/RESEARCH.md) |
| [strings.Cut で連携設定のキーと値を分けよう](01-packages/01-beginner/strings-cut/README.md) | 2 | [調査の記録](01-packages/01-beginner/strings-cut/RESEARCH.md) |
| [受信メッセージのヘッダーを先読みしよう](01-packages/02-intermediate/bytes-buffer-peek/README.md) | 3 | [調査の記録](01-packages/02-intermediate/bytes-buffer-peek/RESEARCH.md) |
| [監査先へ暗号化ファイルを届けよう](01-packages/02-intermediate/crypto-hpke/README.md) | 3 | [調査の記録](01-packages/02-intermediate/crypto-hpke/RESEARCH.md) |
| [slog.Handler インタフェースを実装しよう](01-packages/02-intermediate/slog-handler/README.md) | 3 | [調査の記録](01-packages/02-intermediate/slog-handler/RESEARCH.md) |
| [http.ServeMux のルーティング規約を解読せよ](01-packages/03-advanced/net-http-servemux/README.md) | 4 | [調査の記録](01-packages/03-advanced/net-http-servemux/RESEARCH.md) |
| [time.Timer の Stop/Reset と古い通知を追え](01-packages/03-advanced/time-timer-channels/README.md) | 4 | [調査の記録](01-packages/03-advanced/time-timer-channels/RESEARCH.md) |

### 02-features

| 問題 | 設問数 | 回答例 |
| --- | ---: | --- |
| [iota を使った定数宣言を読もう](02-features/01-beginner/iota-constants/README.md) | 2 | [調査の記録](02-features/01-beginner/iota-constants/RESEARCH.md) |
| [自作の型にジェネリックメソッドで汎用ヘルパを生やそう](02-features/02-intermediate/generic-methods/README.md) | 3 | [調査の記録](02-features/02-intermediate/generic-methods/RESEARCH.md) |
| [Go 1.27 の `goroutineleak` プロファイルは、なぜ「絶対に起きない goroutine」だけを教えてくれるのか](02-features/03-advanced/goroutine-leak-profile/README.md) | 4 | [調査の記録](02-features/03-advanced/goroutine-leak-profile/RESEARCH.md) |

### 03-cmd-tools

| 問題 | 設問数 | 回答例 |
| --- | ---: | --- |
| [`go run` の裏側を覗いてみよう](03-cmd-tools/01-beginner/go-run/README.md) | 3 | [調査の記録](03-cmd-tools/01-beginner/go-run/RESEARCH.md) |
| [テストカバレッジを確認しよう](03-cmd-tools/01-beginner/go-tool-cover/README.md) | 2 | [調査の記録](03-cmd-tools/01-beginner/go-tool-cover/RESEARCH.md) |
| [`go vet` の検査内容を調べよう](03-cmd-tools/01-beginner/go-vet-basics/README.md) | 2 | [調査の記録](03-cmd-tools/01-beginner/go-vet-basics/RESEARCH.md) |
| [go fix のモダナイザで既存コードを最新イディオムに寄せよう](03-cmd-tools/02-intermediate/go-fix-modernize/README.md) | 3 | [調査の記録](03-cmd-tools/02-intermediate/go-fix-modernize/RESEARCH.md) |
| [go generate で何ができ、何をやるべきなのか調べよう](03-cmd-tools/02-intermediate/go-generate/README.md) | 3 | [調査の記録](03-cmd-tools/02-intermediate/go-generate/RESEARCH.md) |
| [バッチ処理の 600ms を追え: `go tool pprof`](03-cmd-tools/02-intermediate/go-tool-pprof/README.md) | 3 | [調査の記録](03-cmd-tools/02-intermediate/go-tool-pprof/RESEARCH.md) |
| [1 枚のテキストがテストになるまで: cmd/go の script tests](03-cmd-tools/03-advanced/go-script-tests/README.md) | 3 | [調査の記録](03-cmd-tools/03-advanced/go-script-tests/RESEARCH.md) |
| [注文処理の 100ms を追え: `go tool trace`](03-cmd-tools/03-advanced/go-tool-trace/README.md) | 4 | [調査の記録](03-cmd-tools/03-advanced/go-tool-trace/RESEARCH.md) |
| [pkg.go.dev API(v1)を活用する](03-cmd-tools/03-advanced/pkg-go-dev-api-v1beta/README.md) | 4 | [調査の記録](03-cmd-tools/03-advanced/pkg-go-dev-api-v1beta/RESEARCH.md) |

### 04-deep-dive

| 問題 | 設問数 | 回答例 |
| --- | ---: | --- |
| [初級: defer の引数評価と実行順を調べよう](04-deep-dive/01-beginner/defer-evaluation/README.md) | 2 | [調査の記録](04-deep-dive/01-beginner/defer-evaluation/RESEARCH.md) |
| [初級: encoding/json の nil スライスと空スライスを調べよう](04-deep-dive/01-beginner/json-nil-slice/README.md) | 2 | [調査の記録](04-deep-dive/01-beginner/json-nil-slice/RESEARCH.md) |
| [中級: context.WithoutCancel で監査ログを期限付きで残そう](04-deep-dive/02-intermediate/context-without-cancel/README.md) | 3 | [調査の記録](04-deep-dive/02-intermediate/context-without-cancel/RESEARCH.md) |
| [中級: 3添字スライス式で通知対象を分離しよう](04-deep-dive/02-intermediate/slice-append-aliasing/README.md) | 3 | [調査の記録](04-deep-dive/02-intermediate/slice-append-aliasing/RESEARCH.md) |
| [なぜ標準の error にはスタックトレースが含まれていないのか？](04-deep-dive/03-advanced/error-stacktrace/README.md) | 3 | [調査の記録](04-deep-dive/03-advanced/error-stacktrace/RESEARCH.md) |
| [Go 1.27 は、なぜ同じ浮動小数点文字列を別の方法で作るようになったのか](04-deep-dive/03-advanced/float-formatting/README.md) | 4 | [調査の記録](04-deep-dive/03-advanced/float-formatting/RESEARCH.md) |
| [クリーンなソースコードだけで Go コンパイラを信頼できる？](04-deep-dive/03-advanced/go-toolchain-trust/README.md) | 4 | [調査の記録](04-deep-dive/03-advanced/go-toolchain-trust/RESEARCH.md) |
| [Go 1.27 の Faster Memory Allocation は、なんで 80 バイト以下だけなの？](04-deep-dive/03-advanced/size-specialized-malloc/README.md) | 3 | [調査の記録](04-deep-dive/03-advanced/size-specialized-malloc/RESEARCH.md) |
| [slog.Value はなぜ == で比較できないのか](04-deep-dive/03-advanced/slog-value-comparison/README.md) | 4 | [調査の記録](04-deep-dive/03-advanced/slog-value-comparison/RESEARCH.md) |
