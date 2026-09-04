[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 01-packages](../../README.md)

# Choose the first non-zero value with cmp.Or

During review, you notice `cmp.Or` in code that selects a value from candidates containing empty strings. Predict which candidate is selected, then run the example.

Run the following code in the [Go Playground](https://go.dev/play/p/WkH9Xink92x) to observe the result when values are selected by priority and when all candidates are zero values.

```go
package main

import (
	"cmp"
	"fmt"
)

func main() {
	first := ""
	second := "second"
	fallback := "fallback"
	fmt.Println(cmp.Or(first, second, fallback))
	fmt.Printf("all unavailable: %q\n", cmp.Or("", ""))
}
```

Output:

```text
second
all unavailable: ""
```

---

## Question 1: Why is the second value selected?

`first` is empty, so why is the output `second`? Find out in what order `cmp.Or` examines its arguments and which value it returns.

For example, first predict which value is selected by `cmp.Or("", "second", "fallback")` after the empty string is skipped, then investigate.

<details>
<summary>Hint</summary>

- Open the standard package documentation from [How to investigate this category](../../README.md).
- Find `Or` in the list of functions in the `cmp` package and look for the phrase “zero value”.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Starting from [Go Documentation](https://go.dev/doc/), open the standard library’s `cmp` package.
2. Read the explanation of [cmp.Or](https://pkg.go.dev/cmp#Or).
3. Open the [implementation in cmp.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmp/cmp.go;l=67) and confirm that the arguments are compared from the beginning.

**Answer**

`cmp.Or` examines its arguments from left to right and returns the first value that is not the zero value. The zero value of `string` is the empty string, so the empty `first` is skipped and `second` is returned.

The argument order expresses the priority.

</details>

---

## Question 2: What should you check when everything is unconfigured?

The `""` returned by `cmp.Or("", "")` indicates that no non-zero candidate was found. Can the return value distinguish that case from an explicitly selected empty string?

<details>
<summary>Hint</summary>

- Check what happens when all values are zero values in `Or`.
- Consider what information is returned from the function’s return type and number of return values.
- Compare the results of `cmp.Or("", "")` and `cmp.Or("", "fallback")`, and consider whether the return value alone distinguishes “no candidate was found”.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read the `cmp` section of the [Go 1.22 Release Notes](https://go.dev/doc/go1.22).
2. Confirm the return value when all arguments are zero values in [cmp.Or](https://pkg.go.dev/cmp#Or).
3. Check the final returned value in the [implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmp/cmp.go;l=69).

**Answer**

When all arguments are zero values, `cmp.Or` returns the zero value of that type. For `string`, that is the empty string. The return value contains only the selected string; it does not say which candidate it came from or whether it represented an unconfigured setting.

If an empty string means “no candidate”, check the returned value separately. If an explicit empty string must be distinguished, use a data structure that represents presence as well as the value.

For example, if you need to distinguish “there is no candidate” from “an explicit empty string is configured”, return both `value string` and `found bool` from the lookup function. `cmp.Or` determines value priority; it does not record whether a setting was present.

</details>

---

<details>
<summary>Trivia</summary>

The type parameter of `cmp.Or` is `comparable`. It is suitable for priority selection using zero values of strings, numbers, pointers, and similar types, but it cannot directly accept values of types that cannot be compared.

</details>

---

## Starting points for investigation

- [Go Documentation](https://go.dev/doc/)
- [How to investigate 01-packages](../../README.md)
- [package cmp](https://pkg.go.dev/cmp)
