# Investigate fmt.Printf verbs

You came across format verbs in a senior colleague’s `fmt.Printf` code. Let’s investigate what they do.

```go
value := "gopher"
fmt.Printf("%#[1]v %[1]T\n", value)
```

(Run it in the Go Playground: https://go.dev/play/p/RTNSvn_p2Ai )

Find out what this code does.

This example examines the same `value` in two ways: as a Go-syntax-like representation and as a type name. First predict where `"gopher"` and `string` will appear, then run it to make the roles of the format components easier to follow.

## Question 1: What do `%v`, `%T`, and `#` mean?

Try putting different values, such as a struct, map, or pointer, into `value` and check how the output changes.

<details>
<summary>Hint</summary>

- The “Printing” section near the beginning of the Overview contains a table of verbs such as `%v`.
- The “Other flags” section a little farther down explains the `#` flag.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Starting from [Go Documentation](https://go.dev/doc/), open the standard library’s `fmt` package.
2. Read the “Printing” section near the beginning of the Overview at https://pkg.go.dev/fmt.
3. In the fmt documentation, format components such as `%v` and `%T` are called **verbs**. Find `%v` and `%T` in the verb table.
4. Look for the explanation of the `#` flag in “Other flags” farther down.

**Answer**

- `%v`: prints the value in its default format.
- `%T`: prints the value’s type in Go syntax.
- `#`: requests an alternate format. Combined with `%v` as `%#v`, it prints the value as a Go-syntax representation.

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
2. Search for `[` in the Overview at https://pkg.go.dev/fmt.
3. Find the “Explicit argument indexes” section.

**Answer**

`[n]` is an explicit argument index. It switches the argument used by the next verb to argument number `n`.

Normally, verbs consume arguments from the beginning in order. In `%#[1]v %[1]T`, both verbs specify that they should use the first argument, so the single `value` is printed twice. This avoids passing duplicate arguments when you want to print the same value with multiple formats.

</details>

---

<details>
<summary>Trivia: The order of flags and argument indexes</summary>

`%[1]#v`, which reverses the order of the `#` flag and `[1]`, does not work.

```go
fmt.Printf("%[1]#v %[1]T\n", value)
// Output: %!#(string=gopher)v string
```

A verb must follow the argument index, so `#` is treated as an invalid verb and produces the `%!#(...)` error notation.

Remember to write flags before the argument index.

You can confirm the failure in the Go Playground: https://go.dev/play/p/fYQqEjznC-T

</details>

---

## Starting points for investigation

- https://pkg.go.dev/fmt
