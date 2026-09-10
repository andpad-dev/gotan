[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 03-cmd-tools](../../README.md)

# Bring Existing Code Up to Date with go fix Modernizers

We want to bring Go code in line with current idioms. `go fix` was revamped in Go 1.26. Let us investigate how to use it, how it differs from `go vet`, and how it can help migrate an API.

For example, here is code written in an unmistakably old style ([Playground](https://go.dev/play/p/6pcJuZQr7_0)):

```go
package main

import (
	"fmt"
	"sort"
)

func main() {
	xs := []int{3, 1, 2}
	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
	for i := 0; i < len(xs); i++ {
		fmt.Println(i, xs[i])
	}
}
```

It prints:

```
0 1
1 2
2 3
```

Running `go fix -diff ./...` produces this diff:

```diff
 import (
 	"fmt"
-	"sort"
+	"slices"
 )
 
 func main() {
 	xs := []int{3, 1, 2}
-	sort.Slice(xs, func(i, j int) bool { return xs[i] < xs[j] })
+	slices.Sort(xs)
-	for i := 0; i < len(xs); i++ {
+	for i := range xs {
 		fmt.Println(i, xs[i])
 	}
 }
```

After applying it ([Playground](https://go.dev/play/p/cx7XCG7OLqB)):

```go
package main

import (
	"fmt"
	"slices"
)

func main() {
	xs := []int{3, 1, 2}
	slices.Sort(xs)
	for i := range xs {
		fmt.Println(i, xs[i])
	}
}
```

What exactly is this `go fix`, and how far does it go?

---

<details><summary>Investigation entry points</summary>

Start with the [03-cmd-tools research guide](../../README.md), then use the primary sources listed at the end of this scenario.

</details>

## Question 1: Find the role of the new `go fix`

What does the `go fix` revamped in Go 1.26 do, and how does it differ from the old fixers dating back to Go 1.0 and from `go vet`?

<details>
<summary>Hint</summary>

- Search for `go fix` in the Tools section of the [Go 1.26 release notes](https://go.dev/doc/go1.26#go-command).
- Read the Overviews of [cmd/fix](https://pkg.go.dev/cmd/fix) and [cmd/vet](https://pkg.go.dev/cmd/vet) side by side.
- In the [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis) Overview, check the relationship between checkers and fixers.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Read the `go fix` entry in the Tools section of the [Go 1.26 release notes](https://go.dev/doc/go1.26#go-command).
2. Read the Overview of [cmd/fix](https://pkg.go.dev/cmd/fix) and compare it with [cmd/vet](https://pkg.go.dev/cmd/vet).
3. Read the "The Go analysis framework" section of [Using go fix to modernize Go code](https://go.dev/blog/gofix).

**Answer**

Go 1.26 completely rewrote `go fix`. It now runs a collection of fixers built on the same analysis framework as `go vet`; the historical fixers for language changes from the Go 1.0 era were removed. `go vet` runs checkers and reports suspicious constructs, while `go fix` runs fixers that compute safe replacements and applies them to source. `-diff` displays the unified diff without applying it. The main group is the gopls [modernize package](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize), which can replace `sort.Slice` with `slices.Sort`, three-clause loops with `range`, and `interface{}` with `any`. The `inline` analyzer for `//go:fix inline` is also included.

</details>

---

## Question 2: Inspect, select, and apply modernizers

Before running `go fix ./...`, how can you see which modernizers apply, inspect their proposed replacements, and apply only the changes you need?

<details>
<summary>Hint</summary>

- Use `go tool fix help` for the list.
- Use `go tool fix help <analyzer name>` for details.
- Preview with `go fix -diff ./...`.
- Analyzer-specific flags include `-any` and `-slicessort`.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Run `go help fix` and read about `-diff` and analyzer flags.
2. Inspect registered analyzers with `go tool fix help`, then read individual details with `go tool fix help <name>`.
3. Read the [gopls modernize package](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize) and, when necessary, its [source code](https://cs.opensource.google/go/x/tools/+/master:gopls/internal/analysis/modernize/).

**Answer**

1. Use `go tool fix help` to list analyzers such as `any`, `minmax`, `rangeint`, `slicessort`, `stringscut`, and `inline`.
2. Use `go tool fix help slicessort` for the conditions and Go-version requirements of one fixer.
3. Preview with `go fix -diff ./...`.
4. Select one fixer with `go fix -slicessort ./...`, or disable it with `go fix -slicessort=false ./...`.
5. Apply the accepted changes with `go fix ./...`; begin from a clean Git state so the result is easy to review separately.
6. The modernizer analyzers are also used by gopls. `go fix` applies the same kind of editor suggestions in bulk. The implementation is documented in the [gopls modernize package](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize). Generated files containing `// Code generated ...` are not modified; update the generator instead.

</details>

---

## Question 3: Automate a custom API migration with `//go:fix inline`

How can we have `go fix` replace calls to a deprecated internal-library function with its new implementation? What constraints and caveats apply?

<details>
<summary>Hint</summary>

- Start with `go tool fix help inline`.
- Read the [inline analyzer documentation](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inline).
- Look for the interaction with `Deprecated:`, the restrictions for `const` and `type`, and the `var params = args` binding declaration.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Run `go tool fix help inline`.
2. Read the [inline analyzer documentation](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inline), including its flags.
3. Read the self-service section of [Using go fix to modernize Go code](https://go.dev/blog/gofix).

**Answer**

Put `//go:fix inline` immediately before the function, constant, or type alias being migrated:

```go
// Deprecated: use Pow(x, 2) directly.
//go:fix inline
func Square(x int) int { return Pow(x, 2) }
```

Then `go fix -inline ./...` (or `go fix ./...`) expands the body at call sites, including across packages. Pair it with a `Deprecated:` comment and keep the old API as a thin wrapper. For constants, the right-hand side must be another named constant, not a literal. A type alias such as `type A = newpkg.A` can also be inlined. To preserve argument evaluation order, the inliner may insert a `var params = args` binding declaration; `-inline.allow_binding_decl=false` skips fixes requiring one. References in the symbol's own tests are deliberately not inlined, so those tests continue to protect the old API behavior.

</details>

---

<details>
<summary>Trivia</summary>

- The Go team is considering a self-service model in which external module authors distribute analyzers for users' `go fix` and gopls; see [issue #59869](https://go.dev/issue/59869) and the [explanatory article](https://go.dev/blog/gofix).
- When fixes conflict, `go fix` skips the conflict and warns. Running `go fix ./...` again can move the code toward a fixed point.
- The `stringsbuilder` fixer can address near-DoS-level performance problems caused by repeated string concatenation in loops.

</details>

---

## Primary sources

- Local commands: `go help fix` / `go tool fix help` / `go tool fix help <analyzer>`
- Release notes: [Go 1.26 release notes #go-command](https://go.dev/doc/go1.26#go-command)
- Command documentation: [cmd/fix](https://pkg.go.dev/cmd/fix) / [cmd/vet](https://pkg.go.dev/cmd/vet) / [relevant cmd/go section](https://pkg.go.dev/cmd/go#hdr-Apply_fixes_suggested_by_static_checkers)
- Modernizer sources: [gopls modernize package](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize) / [inline analyzer](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inline)
- Foundation: [go/analysis](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- Explanatory article: [Using go fix to modernize Go code](https://go.dev/blog/gofix)
