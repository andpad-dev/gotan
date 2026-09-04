[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 01-packages](../../README.md)

# Split at the first delimiter with strings.Cut

During review, you find code that splits a string at the first `=`. Investigate how it distinguishes an absent delimiter from an empty remainder.

Run the following observation log in the [Go Playground](https://go.dev/play/p/ieMFoFpSNtz) to see how the result differs when the delimiter is present or absent.

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	for _, input := range []string{"left=right", "left"} {
		before, after, found := strings.Cut(input, "=")
		fmt.Printf("%q -> before=%q after=%q found=%t\n", input, before, after, found)
	}
}
```

Output:

```text
"left=right" -> before="left" after="right" found=true
"left" -> before="left" after="" found=false
```

---

## Question 1: What do the three return values communicate?

What are the three return values of `strings.Cut`? Explain why, for `"left"`, `after == ""` alone cannot tell whether the remainder is empty or whether `=` was missing entirely.

<details>
<summary>Hint</summary>

- First open the standard package documentation from [How to investigate this category](../../README.md).
- Search for `Cut` and read the return value names and the sentence about when the delimiter is absent.
- Compare `"left="` and `"left"`: notice that `after` is empty in both cases, while `found` differs.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Starting from [Go Documentation](https://go.dev/doc/), open the standard library’s `strings` package.
2. Read the explanation of [strings.Cut](https://pkg.go.dev/strings#Cut).
3. Also check the [strings.go implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/strings/strings.go;l=1273) to confirm that `Cut` delegates to an internal implementation.

**Answer**

The return values are, in order, the string `before` the first delimiter, the string `after` it, and `found`, which reports whether the delimiter was found.

When `"left"` is cut with `"="`, `before` is the original `"left"`, `after` is `""`, and `found` is `false`. `after == ""` also occurs for `"left="`, so check `found` to determine whether the delimiter existed.

</details>

---

## Question 2: Where does it split?

Where does `strings.Cut("left=middle=right", "=")` split? Confirm the behavior of splitting only at the first `=`.

<details>
<summary>Hint</summary>

- Focus on `first instance` in the description of `Cut`.
- Consider whether you can pass `before` and `after` to the same function again.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Starting from [Go Documentation](https://go.dev/doc/), open the standard library’s `strings` package.
2. Read the explanation of the “first instance” in [strings.Cut](https://pkg.go.dev/strings#Cut).
3. Open the [relevant implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/strings/strings.go;l=1273) and confirm that `Cut` is responsible for one split.

**Answer**

The result is `before == "left"`, `after == "middle=right"`, and `found == true`. `Cut` splits only at the first match.

The rest is preserved even when it contains another `=`. To split it further, apply `Cut` to `after` again.

</details>

---

<details>
<summary>Trivia</summary>

For labels where you only need to check a prefix or suffix, the same `strings` package also provides `CutPrefix` and `CutSuffix`. Both return the original string and `false` when the requested text is not found.

</details>

---

## Starting points for investigation

- [Go Documentation](https://go.dev/doc/)
- [How to investigate 01-packages](../../README.md)
- [package strings](https://pkg.go.dev/strings)
