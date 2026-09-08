[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 01-packages](../../README.md)

# Investigate fmt.Printf verbs

You came across format verbs in `fmt.Printf` code. Let’s investigate what they do.

```go
package main

import "fmt"

func main() {
	value := "gopher"
	fmt.Printf("%#[1]v %[1]T\n", value)
}
```

[Run it in the Go Playground](https://go.dev/play/p/RTNSvn_p2Ai). With Go 1.27.0, it prints:

```text
"gopher" string
```

Find out what this code does.

This example examines the same `value` in two ways: as a Go-syntax-like representation and as a type name. Investigate which format component produced `"gopher"` and which produced `string`.

---

## Question 1: What do `%v`, `%T`, and `#` mean?

Try putting different values, such as a struct, map, or pointer, into `value` and check how the output changes.

<details>
<summary>Hint</summary>

- In the verb table under “Printing,” compare the `%v`, `%#v`, and `%T` rows directly.
- “Other flags” lists other verb-specific effects of `#`, such as `%#b` and `%#x`. If `%v` is absent there, return to the verb table.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Starting from [Go Documentation](https://go.dev/doc/), open the standard library’s `fmt` package.
2. Open [`fmt`'s “Printing” section](https://pkg.go.dev/fmt#hdr-Printing).
3. In the fmt documentation, format components such as `%v` and `%T` are called **verbs**. Find and compare the `%v`, `%#v`, and `%T` rows in the General table.
4. Read “Other flags” in the same section and confirm that `#` has verb-specific effects and that this list does not define `%#v`.
5. In Go 1.27.0, read [`fmtFlags`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/fmt/format.go;l=26-36) and confirm why `%#v` is handled as `sharpV`, separately from the ordinary `#` flag.

**Answer**

- `%v`: prints the value in its default format.
- `%T`: prints the value’s type in Go syntax.
- `#`: is a formatting flag whose effect depends on the verb. Its combination with `%v` appears in the verb table as the distinct `%#v` form, which prints a Go-syntax representation of the value.

Thus, this code places `%#v` and `%T` side by side to inspect both the value and its type in Go syntax, a common debugging idiom.

```go
value := "gopher"
fmt.Printf("%#[1]v %[1]T\n", value)
// Output: "gopher" string
```

With `%v`, the output would be `gopher`; with `%#v`, it is the Go literal `"gopher"`, including quotation marks. The difference between `%#v` and `%v` becomes even clearer with structs and maps.

</details>

---

## Question 2: What does `[1]` do?

Only one argument, `value`, is passed. Why is it printed twice?

<details>
<summary>Hint</summary>

- Search for `[` in the package Overview.
- For example, how many times does `fmt.Printf("%[1]s / %[1]s\n", "gopher")` use its one argument?

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Starting from [Go Documentation](https://go.dev/doc/), open the standard library’s `fmt` package.
2. Search for `[` in the [`fmt` Overview](https://pkg.go.dev/fmt).
3. Reach [“Explicit argument indexes”](https://pkg.go.dev/fmt#hdr-Explicit_argument_indexes) and read how `[n]` selects an argument.

**Answer**

`[n]` is an explicit argument index. It switches the argument used by the next verb to argument number `n`.

Normally, verbs consume arguments from the beginning in order. In `%#[1]v %[1]T`, both verbs specify that they should use the first argument, so the single `value` is printed twice. This avoids passing duplicate arguments when you want to print the same value with multiple formats.

</details>

---

<details>
<summary>Trivia: The order of flags and argument indexes</summary>

`%[1]#v`, which reverses the order of the `#` flag and `[1]`, does not work.

Run the following complete program in the [Go Playground](https://go.dev/play/p/HczVP8-FaCF) to compare the valid and invalid orders with the same input.

```go
package main

import "fmt"

func main() {
	value := "gopher"
	fmt.Printf("%#[1]v %[1]T\n", value)
	fmt.Printf("%[1]#v %[1]T\n", value)
}
```

```text
"gopher" string
%!#(string=gopher)v string
```

A verb must follow the argument index, so `#` is treated as an invalid verb and produces the `%!#(...)` error notation.

Remember to write flags before the argument index.

</details>

---

## Starting points for investigation

- [Go Documentation](https://go.dev/doc/)
- [How to research 01-packages](../../README.md)
- [package fmt](https://pkg.go.dev/fmt)
