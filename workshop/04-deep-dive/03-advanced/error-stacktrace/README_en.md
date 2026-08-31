[Workshop guide and scenario index](../../../README.md) | [How to research 04-deep-dive](../../README.md)

# Why Do Standard Go Errors Not Include Stack Traces?

Go developers often log an error and still wonder where it occurred. Unlike Java or Python exceptions, standard Go errors do not automatically carry a stack trace. Investigate the design philosophy and history of error wrapping.

First run this example and observe context wrapping, cause inspection, and detailed formatting. The `%+v` output confirms that standard errors do not automatically print a stack trace.

```go
package main

import (
	"errors"
	"fmt"
)

var errNotFound = errors.New("not found")

func loadConfig() error {
	return errNotFound
}

func handleRequest() error {
	return fmt.Errorf("load config: %w", loadConfig())
}

func main() {
	err := handleRequest()
	fmt.Println("error:", err)
	fmt.Println("is not found:", errors.Is(err, errNotFound))
	fmt.Printf("detailed format: %+v\n", err)
}
```

([Run it in the Go Playground](https://go.dev/play/p/m3PMRBAE_rq))

```text
error: load config: not found
is not found: true
detailed format: load config: not found
```

## Question 1: What is Go's error-handling philosophy?

The example preserves the original cause for `errors.Is`, while `%+v` still has no stack trace. Use this observation to find the official discussion and explain why errors are values rather than a special exception control flow.

<details><summary>Hint</summary>Use the browser's find command on [all Go Blog posts](https://go.dev/blog/all) to locate `Errors are values`, then search the Problem section of [Error Values — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md) for `fixed cost`.</details>

<details><summary>Answer</summary>

**Investigation route**

1. Open [all Go Blog posts](https://go.dev/blog/all) and use the browser's find command to locate `error`.
2. Read [Errors are values](https://go.dev/blog/errors-are-values), 12 January 2015, Rob Pike.
3. Read the Problem section of [Error Values — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md) and identify the cost constraint placed on error creation.

**Answer**

Go treats errors as ordinary values, not exceptional control flow. There is no special try-catch syntax; errors are returned and handled as part of normal control flow. This makes the point where an error is produced and handled explicit.

The design overview also notes that errors are routinely created, handled, and discarded. Their creation must therefore have a fixed cost independent of stack depth or other context. Automatically recording every stack conflicts with that constraint.

</details>

---

## Question 2: Investigate the history of error wrapping

Go 1.13 introduced wrapping with `fmt.Errorf("%w", err)`. What problem was it intended to solve?

In the example above, `fmt.Errorf` adds context while `errors.Is` still finds the original error. Use those two observations as the starting point for the history.

<details><summary>Hint</summary>Start with the [Go 1.13 release notes](https://go.dev/doc/go1.13).</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read the Error Wrapping section of https://go.dev/doc/go1.13.
2. Follow the [Error Values proposal](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md) and its [associated issue](https://go.dev/issues/29934).
3. Read the [`errors` package documentation](https://pkg.go.dev/errors) and [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors).

**Answer**

As programs grow, returning a low-level error such as `io.EOF` without context loses where it occurred. Packages such as `pkg/errors` filled this gap, but the standard library lacked a common way to retain context while inspecting the cause. Go 1.13 introduced wrapping and the standard `errors.Is` and `errors.As` APIs, allowing context to be added while preserving programmatic inspection of the original error. The proposal and issue also discuss formatting, stack frames, interfaces, and the `%w` design.

</details>

---

## Question 3: Why was automatic stack-trace attachment declined?

Read the proposal discussion and explain why stack frames were not included in the standard package.

<details><summary>Hint</summary>

Search [Proposal: Go 2 Error Inspection](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md) for `Frame`, `StackTrace`, and `Formatting`, then inspect the discussion in the [proposal issue](https://go.dev/issues/29934), especially the decision comments.

</details>

<details><summary>Answer</summary>

**Investigation route**

1. Run the [shared example](https://go.dev/play/p/m3PMRBAE_rq) again and confirm that `%+v` does not add a stack trace.
2. Read the Problem section of [Error Values — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md) for the fixed-cost constraint.
3. Read Stack Frames and Formatting in the [Error Values proposal](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md).
4. Compare the issue's [proposal-acceptance comment](https://github.com/golang/go/issues/29934#issuecomment-489682919) with the [final decision comment](https://github.com/golang/go/issues/29934#issuecomment-521245013).
5. Confirm the APIs that shipped in [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors).

**Answer**

The original proposal considered wrapping, stack frames (`errors.Frame`), and formatting (`errors.Printer` and `errors.Formatter`). The accepted and declined decisions are summarized in [this issue comment](https://github.com/golang/go/issues/29934#issuecomment-489682919), with context in [the decision comment](https://github.com/golang/go/issues/29934#issuecomment-521245013). Wrapping had broad agreement, but formatting and location did not reach a satisfactory design, so those parts were postponed and the related APIs removed. The proposal treated frames largely as part of formatting, and the discussion also raised performance costs and Go's value-oriented, customizable error model. These unresolved tradeoffs explain why automatic stack traces were left to custom error packages rather than added to standard errors.

The broader design requirement is that creating an error remain a fixed-cost operation regardless of stack depth. This explains why capturing a stack for every ordinary error is a poor default, while the issue history explains why the optional frame-and-formatting proposal did not ship in Go 1.13.

</details>

---

## Research starting points

- [Go Blog](https://go.dev/blog/)
- [Go 1.13 release notes](https://go.dev/doc/go1.13)
- [Error Values proposal](https://go.googlesource.com/proposal/+/master/design/29934-error-values.md)
- [Proposal issue](https://go.dev/issues/29934)
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Error Values — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-error-values-overview.md)
