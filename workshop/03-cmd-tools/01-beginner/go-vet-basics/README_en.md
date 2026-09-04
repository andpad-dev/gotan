[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 03-cmd-tools](../../README.md)

# Investigate what `go vet` checks

You noticed `go vet ./...` in a CI job. Let’s investigate what it does.

For example, run `go vet` on code like this:

```go
package main

import "fmt"

func main() {
	name := "gopher"
	fmt.Printf("hello, %d\n", name)
}
```

(Run it in the Go Playground: https://go.dev/play/p/GOVSRd_FWPU )

`go build` and `go run` succeed normally, and the output is:

```
hello, %!d(string=gopher)
```

However, `go vet ./...` reports something like this:

```
./main.go:7:21: fmt.Printf format %d has arg name of wrong type string
```

What does `go vet` inspect separately from building and testing?

## Question 1: What does `go vet` report, and how does it differ from compilation?

Use primary sources to confirm the role of `go vet` and the areas that the compiler cannot detect.

<details>
<summary>Hint</summary>

- Run `go help vet` locally for a concise explanation.
- The same content appears in the Overview of [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet).
- From the `go` command’s perspective, search [`pkg.go.dev/cmd/go`](https://pkg.go.dev/cmd/go) for “vet” with the `f` key.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Starting at [Go Documentation](https://go.dev/doc/), find the official documentation for `go vet`.
2. Run `go help vet` locally and read the complete CLI help.
3. Confirm the same content in the Overview of [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet). The sentence “Analyzers may use heuristics that do not guarantee all reports are genuine problems, but can find mistakes not caught by the compiler.” is central here.
4. Find the `go` command’s description in the [“Report likely mistakes in packages” section of `pkg.go.dev/cmd/go`](https://pkg.go.dev/cmd/go#hdr-Report_likely_mistakes_in_packages), which has the same wording as `go help vet`.

**Answer**

- `go vet` is a **static analysis tool for Go programs**. It reports “suspicious constructs” and “opportunities for improvement”.
- Its distinguishing feature is that it uses heuristics to find **mistakes that the compiler cannot detect**. Conversely, a `go vet` report is not necessarily a genuine bug, so you must read and evaluate each report.
- In the opening example, the format (`%d`) and argument type (`string`) in `fmt.Printf("hello, %d\n", name)` do not match. The code compiles, but at runtime it produces the broken output `%!d(string=gopher)`. Catching this kind of logical error that passes through the type system is a representative job for `go vet`.

</details>

---

## Question 2: Which checkers does `go vet` have, and where can you read about `printf` in detail?

The `printf` report above came from one of the many analyzers built into `go vet`. Investigate the overall picture and where to find documentation for an individual analyzer.

<details>
<summary>Hint</summary>

- The Overview of [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) contains a table of registered analyzers.
- Run `go tool vet help` locally for the same list, and `go tool vet help printf` for the individual analyzer’s details.
- The `printf` checker itself is in [`golang.org/x/tools`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf), not in the Go repository’s standard library.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Starting at [Go Documentation](https://go.dev/doc/), find the official documentation for `vet`.
2. Scroll through the Overview of [`pkg.go.dev/cmd/vet`](https://pkg.go.dev/cmd/vet) and read the Registered analyzers table, which includes `appends`, `printf`, `slog`, `stdversion`, `waitgroup`, and about 30 others.
3. Run `go tool vet help` to confirm the same list. Add an analyzer name, such as `go tool vet help printf`, to display that analyzer’s documentation and flags.
4. The implementation documentation for the `printf` analyzer is at [`pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf). It describes checking the consistency of format strings and arguments for functions such as `fmt.Printf` and `fmt.Sprintf`.

**Answer**

- `go vet` is not one checker but **a collection of individual analyzers**. The Registered analyzers table in `pkg.go.dev/cmd/vet` and the output of `go tool vet help` are the official catalog.
- Each analyzer is an independent module written on the [`golang.org/x/tools/go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis) framework. Its implementation and detailed documentation are usually in a `passes/<analyzer-name>` package.
- The report in this example came from the `printf` analyzer. Its coverage and the `-printf.funcs` flag for specifying additional functions to check are documented in both `go tool vet help printf` and `pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf`.

</details>

---

<details>
<summary>Trivia: `go test` quietly runs `go vet`</summary>

Before running tests, `go test` automatically runs a **curated subset of `go vet`**. This is documented in the [“Test packages” section of `pkg.go.dev/cmd/go`](https://pkg.go.dev/cmd/go#hdr-Test_packages), including the default list: `atomic`, `bool`, `buildtags`, `directive`, `errorsas`, `ifaceassert`, `nilfunc`, `printf`, `stringintconv`, and `tests`.

When you run `go test ./...` on the opening code, it is treated as a build failure from the outset:

    # example.com/vet-demo
    ./main.go:7:21: fmt.Printf format %d has arg name of wrong type string
    FAIL    example.com/vet-demo [build failed]

The [Go 1.27 release notes](https://go.dev/doc/go1.27#go-test) also add this sentence:

> `go test` now invokes the `stdversion` vet check by default.

In other words, beginning with Go 1.27, tests also automatically check whether you use **standard-library symbols that are newer than allowed** by the `go` version or `//go:build` tags in `go.mod`. This background helps explain why pull requests keep adding “run `go vet` in CI”.

</details>

---

## Starting points for investigation

- https://pkg.go.dev/cmd/vet
- https://pkg.go.dev/cmd/go#hdr-Report_likely_mistakes_in_packages
- https://pkg.go.dev/cmd/go#hdr-Test_packages
- https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/printf
- https://go.dev/doc/go1.27#go-test
