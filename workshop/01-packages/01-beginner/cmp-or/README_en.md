# Choose a notification channel with cmp.Or

An internal notification service selects a destination channel in the order of personal settings, team settings, and the company default. Unconfigured fields are empty strings. While looking at code where the team setting `#backend-alerts` was selected even though the personal setting was empty, you noticed `cmp.Or`. Let’s investigate what it does.

Run the following code in the [Go Playground](https://go.dev/play/p/InznZIGNSza) to observe the result when values are selected by priority and when all settings are unconfigured.

```go
package main

import (
	"cmp"
	"fmt"
)

func main() {
	personalChannel := ""
	teamChannel := "#backend-alerts"
	companyDefault := "#general"
	fmt.Println(cmp.Or(personalChannel, teamChannel, companyDefault))
	fmt.Printf("all unavailable: %q\n", cmp.Or("", ""))
}
```

Output:

```text
#backend-alerts
all unavailable: ""
```

---

## Question 1: Why is the team setting selected?

`personalChannel` is empty, so why is the output `#backend-alerts`? Find out in what order `cmp.Or` examines its arguments and which value it returns.

For example, first predict which value is selected by `cmp.Or("", "#backend-alerts", "#general")` after the empty string is skipped, then investigate.

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

`cmp.Or` examines its arguments from left to right and returns the first value that is not the zero value. The zero value of `string` is the empty string, so the empty `personalChannel` is skipped and the next value, `teamChannel` (`#backend-alerts`), is returned.

The argument order expresses the priority. In this code, the order is personal setting, team setting, and company default.

</details>

---

## Question 2: What should you check when everything is unconfigured?

The `""` returned by `cmp.Or("", "")` indicates that no setting was found. Whether to treat that as a configuration error and send no notification, or to add another default, depends on the business rules. Can you determine everything needed for that decision from the information returned by `cmp.Or` alone?

<details>
<summary>Hint</summary>

- Check what happens when all values are zero values in `Or`.
- Consider what information is returned from the function’s return type and number of return values.
- Compare the results of `cmp.Or("", "")` and `cmp.Or("", "#general")`, and consider whether the return value alone distinguishes “no candidate was found”.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read the `cmp` section of the [Go 1.22 Release Notes](https://go.dev/doc/go1.22).
2. Confirm the return value when all arguments are zero values in [cmp.Or](https://pkg.go.dev/cmp#Or).
3. Check the final returned value in the [implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmp/cmp.go;l=69).

**Answer**

When all arguments are zero values, `cmp.Or` returns the zero value of that type. For `string`, that is the empty string. The return value contains only the selected string; it does not say which candidate it came from or whether it represented an unconfigured setting.

Because an empty string means “unconfigured” in this configuration, add an empty-string check and apply the business rules. If the empty string itself must be distinguished as a valid notification destination, you need a data structure that represents presence as well as the value.

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
