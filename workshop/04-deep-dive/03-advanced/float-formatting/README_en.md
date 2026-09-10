[Scenario index](../../../SCENARIOS_en.md) | [Workshop guide](../../../README.md) | [How to research 04-deep-dive](../../README.md)

# Why Does Go 1.27 Build the Same Floating-Point String Differently?

**Execution environment**: Browser only. No Go installation is required.

Some code uses `strconv.FormatFloat` with `prec = -1`. Snapshot tests still pass after upgrading to Go 1.27, even though the conversion implementation changed substantially.

```go
package main

import (
    "fmt"
    "math"
    "strconv"
)

func main() {
    x := 0.1
    shortest := strconv.FormatFloat(x, 'g', -1, 64)
    fixed17 := strconv.FormatFloat(x, 'g', 17, 64)
    next := math.Nextafter(1, 2)
    nextText := strconv.FormatFloat(next, 'g', -1, 64)
    shorter := nextText[:len(nextText)-1]
    roundTrip, _ := strconv.ParseFloat(nextText, 64)
    shorterValue, _ := strconv.ParseFloat(shorter, 64)
    fmt.Printf("shortest: %s\n", shortest)
    fmt.Printf("17 digits: %s\n", fixed17)
    fmt.Printf("next float: %s\n", nextText)
    fmt.Printf("round-trip: %t\n", math.Float64bits(next) == math.Float64bits(roundTrip))
    fmt.Printf("without last digit: %s (%t)\n", shorter, math.Float64bits(next) == math.Float64bits(shorterValue))
}
```

[Run it in the Go Playground](https://go.dev/play/p/lFFIIE8DDSc)

With `go1.27.0` the output is:

```text
shortest: 0.1
17 digits: 0.10000000000000001
next float: 1.0000000000000002
round-trip: true
without last digit: 1.000000000000000 (false)
```

Follow the public API contract, the Go 1.26 and 1.27 implementations, and the adopted algorithm in that order.

---

## Question 1: What does “shortest” mean?

Read the public documentation for `FormatFloat` and `ParseFloat`. Explain which condition the shortest string must satisfy.

<details><summary>Hint</summary>

Find the `prec` description for negative values. The important test is whether parsing the string returns the original floating-point value, not whether the decimal looks identical.

</details>

<details><summary>Answer</summary>

With `prec = -1`, `FormatFloat` emits the minimum number of digits needed for `ParseFloat` to return the original `f` exactly. Thus the binary `float64` represented by `0.1` can be printed as `0.1`, while the next value after `1` needs the final `2` in `1.0000000000000002`. `ParseFloat` chooses the nearest representable value using IEEE 754 unbiased rounding. The internal algorithm is not part of this public contract.

</details>

---

## Question 2: What changed between Go 1.26 and Go 1.27?

Compare `internal/strconv` at the Go 1.26.0 and Go 1.27.0 tags. Track the paths for shortest output, fixed-precision output, and decimal input.

<details><summary>Answer</summary>

**Investigation path**

1. Check the [Go 1.27 release notes](https://go.dev/doc/go1.27), without assuming that an unmentioned internal change did not happen.
2. Compare Go 1.26.0 [`ftoa.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.0:src/internal/strconv/ftoa.go;l=109) and [`atof.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.0:src/internal/strconv/atof.go;l=578) with the Go 1.27.0 versions ([`ftoa.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/ftoa.go;l=136), [`atof.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/atof.go;l=589)).
3. Confirm the file changes in [commit `71300e8`](https://go.dev/change/71300e80113c6ca56105aac524e9c1b0db43910f).

**Answer**

Go 1.26 used Dragonbox for shortest output, `fixedFtoa` and `bigFtoa` for fixed output, and an Eisel-Lemire path for decimal input. Go 1.27 uses `shortFloat`, `fixedWidthFloat`, and `parseFloat32`/`parseFloat64`, sharing new conversion components. The old implementation files were removed and `uscale.go` was added. The public `FormatFloat` and `ParseFloat` contracts remain unchanged.

</details>

---

## Question 3: What rounding information does the shared component retain?

Read the design article and `uscale.go`. Explain why all discarded digits do not need to be retained.

<details><summary>Answer</summary>

The `unrounded` representation keeps the integer part plus two low bits: one says whether the discarded fraction is at least one half, and the other is a sticky bit saying that further discarded bits existed. Together with the integer's parity, these bits distinguish below-half, exact-half, and above-half cases and implement round-to-even. `uscale` moves the value to the required decimal position while preserving that information. `shortFloat` additionally uses the midpoints to the adjacent floating-point values to find the shortest decimal that rounds back to the original value.

</details>

---

## Question 4: Why replace the implementation if output is unchanged?

Read the adoption commit and review measurements. Separate the motivation from claims that would be too strong.

<details><summary>Answer</summary>

The replacement preserves the public contract while sharing more code and primarily improving printing performance. Commit `71300e8` reports that almost 900 lines were removed. The measurements show substantial improvements for some printing cases, while other cases are unchanged or slower, so it is not a guarantee that every input is faster. The January 2026 article's prediction that the approach would enter Go 1.27 is not evidence by itself; the commit, CL 743860, and the Go 1.27.0 tagged source establish that it did.

</details>

---

## Investigation entry points

- [Go 1.27 release notes](https://go.dev/doc/go1.27)
- [`strconv.FormatFloat`](https://go.dev/pkg/strconv/#FormatFloat)
- [Commit `71300e8`](https://go.dev/change/71300e80113c6ca56105aac524e9c1b0db43910f)
- [Floating Point Formatting](https://research.swtch.com/fp-all)
- [Go 1.27.0 `uscale.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/strconv/uscale.go)
