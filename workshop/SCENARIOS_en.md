[Workshop facilitation guide](README.md) | [How teams work](TEAM_GUIDE_en.md)

# Scenario List

Choose a topic and difficulty that interest you, then open the scenario. From the beginning of each scenario, you can return to this list, the facilitation guide, or the research guide for its category. Materials used in the tutorial are not listed here; they are in [00-tutorial](00-tutorial/README_en.md).

The execution environment column shows what you need to work through each scenario.

- **Browser only** — You can complete the scenario without installing Go.
- **Local Go** — You need Go installed locally to answer the questions. When a version is specified, that version or later is required.
- **No execution** — This scenario does not require running code.

Some scenarios marked **Local Go** also include a [Go Playground](https://go.dev/play/) link. That is for observing sample-code behavior; you still need Go locally to answer the questions.
If your local Go version is older than the version required by a scenario, `go run` will begin downloading a toolchain. Allow extra time for that if your connection is slow.

## 01-packages — [How to research standard packages](01-packages/README.md)

| Difficulty | Scenario | Execution environment |
| --- | --- | --- |
| Beginner | [`cmp.Or`](01-packages/01-beginner/cmp-or/README_en.md) | Browser only |
| Beginner | [`fmt.Sprintf`](01-packages/01-beginner/fmt-sprintf/README_en.md) | Browser only |
| Beginner | [The specification of `for` statements](01-packages/01-beginner/lang-spec-basics/README_en.md) | Browser only |
| Beginner | [`strings.Cut`](01-packages/01-beginner/strings-cut/README_en.md) | Browser only |
| Intermediate | [`bytes.Buffer`](01-packages/02-intermediate/bytes-buffer-peek/README_en.md) | Browser only |
| Intermediate | [`crypto/hpke`](01-packages/02-intermediate/crypto-hpke/README_en.md) | Browser only |
| Intermediate | [`slog.Handler`](01-packages/02-intermediate/slog-handler/README_en.md) | Browser only |
| Advanced | [`http.ServeMux`](01-packages/03-advanced/net-http-servemux/README_en.md) | Local Go 1.22 or later |
| Advanced | [`time.Timer`](01-packages/03-advanced/time-timer-channels/README_en.md) | Local Go 1.27 or later |

## 02-features — [How to research language features](02-features/README.md)

| Difficulty | Scenario | Execution environment |
| --- | --- | --- |
| Beginner | [`iota`](02-features/01-beginner/iota-constants/README_en.md) | Browser only |
| Intermediate | [Generic methods](02-features/02-intermediate/generic-methods/README_en.md) | Browser only |
| Advanced | [`goroutineleak` profile](02-features/03-advanced/goroutine-leak-profile/README_en.md) | Browser only |

## 03-cmd-tools — [How to research Go commands and tools](03-cmd-tools/README.md)

| Difficulty | Scenario | Execution environment |
| --- | --- | --- |
| Beginner | [`go run`](03-cmd-tools/01-beginner/go-run/README_en.md) | Local Go 1.24 or later |
| Beginner | [`go tool cover`](03-cmd-tools/01-beginner/go-tool-cover/README_en.md) | Local Go |
| Beginner | [`go vet`](03-cmd-tools/01-beginner/go-vet-basics/README_en.md) | Local Go 1.27 or later |
| Intermediate | [`go fix`](03-cmd-tools/02-intermediate/go-fix-modernize/README_en.md) | Local Go 1.27 or later |
| Intermediate | [`go generate`](03-cmd-tools/02-intermediate/go-generate/README_en.md) | Local Go 1.21 or later |
| Intermediate | [`go tool pprof`](03-cmd-tools/02-intermediate/go-tool-pprof/README_en.md) | Local Go 1.27 or later |
| Advanced | [cmd/go script tests](03-cmd-tools/03-advanced/go-script-tests/README_en.md) | Local Go 1.27 |
| Advanced | [`go tool trace`](03-cmd-tools/03-advanced/go-tool-trace/README_en.md) | Local Go |
| Advanced | [pkg.go.dev API](03-cmd-tools/03-advanced/pkg-go-dev-api-v1beta/README_en.md) | No execution |

## 04-deep-dive — [How to research specifications, implementation, and design background](04-deep-dive/README.md)

| Difficulty | Scenario | Execution environment |
| --- | --- | --- |
| Beginner | [How `defer` is evaluated](04-deep-dive/01-beginner/defer-evaluation/README_en.md) | Browser only |
| Beginner | [Nil slices in `encoding/json`](04-deep-dive/01-beginner/json-nil-slice/README_en.md) | Browser only |
| Intermediate | [`context.WithoutCancel`](04-deep-dive/02-intermediate/context-without-cancel/README_en.md) | Browser only |
| Intermediate | [Three-index slicing](04-deep-dive/02-intermediate/slice-append-aliasing/README_en.md) | Browser only |
| Advanced | [Errors and stack traces](04-deep-dive/03-advanced/error-stacktrace/README_en.md) | Browser only |
| Advanced | [Floating-point string conversion](04-deep-dive/03-advanced/float-formatting/README_en.md) | Browser only |
| Advanced | [Trust in the Go toolchain](04-deep-dive/03-advanced/go-toolchain-trust/README_en.md) | No execution |
| Advanced | [Size-specialized malloc](04-deep-dive/03-advanced/size-specialized-malloc/README_en.md) | Browser only |
| Advanced | [Comparing `slog.Value`](04-deep-dive/03-advanced/slog-value-comparison/README_en.md) | Browser only |
