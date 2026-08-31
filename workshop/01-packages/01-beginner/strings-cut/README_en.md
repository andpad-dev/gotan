[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 01-packages](../../README.md)

# Split integration settings with strings.Cut

Integration settings for an external SaaS arrive from an administration screen as `key=value` entries such as `region=asia-east1`. The implementer wants to split the settings safely first, so an input missing its delimiter is not mistaken for an input whose value is empty. You came across `strings.Cut` in the code. Let’s investigate what it does.

Run the following observation log in the [Go Playground](https://go.dev/play/p/qzzfmNtViQU) to see how the result differs when the delimiter is present or absent.

```go
package main

import (
	"fmt"
	"strings"
)

func main() {
	for _, setting := range []string{"region=asia-east1", "region"} {
		key, value, found := strings.Cut(setting, "=")
		fmt.Printf("%q -> key=%q value=%q found=%t\n", setting, key, value, found)
	}
}
```

Output:

```text
"region=asia-east1" -> key="region" value="asia-east1" found=true
"region" -> key="region" value="" found=false
```

---

## Question 1: What do the three return values communicate?

What are the three return values of `strings.Cut`? Explain why, for `"region"`, `value == ""` alone cannot tell whether the value is empty or whether `=` was missing entirely.

<details>
<summary>Hint</summary>

- First open the standard package documentation from [How to investigate this category](../../README.md).
- Search for `Cut` and read the return value names and the sentence about when the delimiter is absent.
- Compare `"region="` and `"region"`: notice that `value` is empty in both cases, while `found` differs.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Starting from [Go Documentation](https://go.dev/doc/), open the standard library’s `strings` package.
2. Read the explanation of [strings.Cut](https://pkg.go.dev/strings#Cut).
3. Also check the [strings.go implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/strings/strings.go;l=1273) to confirm that `Cut` delegates to an internal implementation.

**Answer**

The return values are, in order, the string `before` the first delimiter, the string `after` it, and `found`, which reports whether the delimiter was found.

When `"region"` is cut with `"="`, `before` is the original `"region"`, `after` is `""`, and `found` is `false`. `after == ""` also occurs when the value is genuinely empty, as in `"region="`, so validate the setting format by checking `found`.

</details>

---

## Question 2: Where does it split?

The integration target sends `callback=https://example.com/a=b`. Where does `strings.Cut(setting, "=")` split? Consider why this behavior is useful for the format in which only the first `=` is a meaningful boundary.

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

The result is `before == "callback"`, `after == "https://example.com/a=b"`, and `found == true`. `Cut` splits only at the first match.

That is useful when a format needs to separate a key from the rest of its value in one operation: an `=` inside the value is preserved. If the value is structured further, you can read it step by step by applying `Cut` to `after` again.

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
