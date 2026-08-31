[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 01-packages](../../README.md)

# Build strings with fmt.Sprintf

You came across the following code in a senior colleague’s work, building a string by inserting values.

```go
package main

import "fmt"

func main() {
	name := "gopher"
	points := 42
	msg := fmt.Sprintf("User %s has %d points", name, points)
	fmt.Println(msg)
	fmt.Printf("%T\n", msg)
}
```

([Run it in the Go Playground](https://go.dev/play/p/u99kEiMPNJA))

Output:

```text
User gopher has 42 points
string
```

Let’s investigate what `fmt.Sprintf` does.

## Question 1: What does `Sprintf` return? What do `%s` and `%d` mean?

`fmt.Println` prints to the screen, but what does this `fmt.Sprintf` return? Also, find out how `name` and `points` correspond to `%s` and `%d`.

<details>
<summary>Hint</summary>

- Press `f` at https://pkg.go.dev/fmt and search for `Sprintf` to see the function signature and return type.
- The meanings of the verbs are listed in the “Printing” section near the beginning of the Overview.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Start at [Go Documentation](https://go.dev/doc/) and open the standard library's `fmt` package.
2. Open https://pkg.go.dev/fmt, open the search dialog with `f`, enter `Sprintf`, and jump to the function.
3. Check the signature `func Sprintf(format string, a ...any) string` and confirm that the return type is `string`.
4. Return to the “Printing” section near the beginning of the Overview and check the meanings of `%s` and `%d` in the verb table.

**Answer**

- `fmt.Sprintf` uses the same formatting directives as `Printf`, but **returns the formatted result as a `string` instead of printing it**. Use it when you want to store the constructed string and use it later.
- `%s`: a verb that inserts an argument as a string.
- `%d`: a verb that inserts an argument as a decimal integer.
- The `%s` and `%d` in `format` correspond to the following arguments, `name` and `points`, from left to right.

```go
name := "gopher"
points := 42
msg := fmt.Sprintf("User %s has %d points", name, points)
fmt.Println(msg)
// Output: User gopher has 42 points
fmt.Printf("%T\n", msg)
// Output: string
```

It is useful to remember the pair this way: `fmt.Printf(...)` formats and prints immediately, while `fmt.Sprintf(...)` formats and returns a string.

</details>

---

## Question 2: Aligning digits and padding with zeros

Logs and reports often need numbers to have the same width, to be zero-padded like `007`, or to have a fixed number of digits after the decimal point. Find out how to specify width and precision for `%d` and `%f`.

Run this code first and relate each requested width or precision to the observed output before looking up the syntax.

```go
package main

import "fmt"

func main() {
	fmt.Println(fmt.Sprintf("[%5d]", 42))
	fmt.Println(fmt.Sprintf("[%-5d]", 42))
	fmt.Println(fmt.Sprintf("[%05d]", 42))
	fmt.Println(fmt.Sprintf("[%.2f]", 3.14159))
	fmt.Println(fmt.Sprintf("[%8.2f]", 3.14159))
}
```

([Run it in the Go Playground](https://go.dev/play/p/SMFWIRANv3B))

Output:

```text
[   42]
[42   ]
[00042]
[3.14]
[    3.14]
```

<details>
<summary>Hint</summary>

- Read the “Width and precision” section in the middle of “Printing” in the Overview.
- Write the number **between** `%` and the verb.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Start at [Go Documentation](https://go.dev/doc/) and open the standard library's `fmt` package.
2. Find the “Width and precision” section in the Overview at https://pkg.go.dev/fmt.
3. Confirm that width follows `%`, precision follows `.`, `0` produces zero padding, and `-` produces left alignment.

**Answer**

- `%5d`: width 5, right-aligned with spaces as needed.
- `%-5d`: width 5, left-aligned.
- `%05d`: width 5, padded with zeros.
- `%.2f`: two digits after the decimal point.
- `%8.2f`: width 8 and two digits after the decimal point.
- `%%`: outputs `%` as a character rather than acting as a verb.

```go
fmt.Println(fmt.Sprintf("[%5d]", 42))        // [   42]
fmt.Println(fmt.Sprintf("[%-5d]", 42))       // [42   ]
fmt.Println(fmt.Sprintf("[%05d]", 42))       // [00042]
fmt.Println(fmt.Sprintf("[%.2f]", 3.14159))  // [3.14]
fmt.Println(fmt.Sprintf("[%8.2f]", 3.14159)) // [    3.14]
fmt.Println(fmt.Sprintf("%d%%", 50))          // 50%
```

(Run it in the Go Playground: https://go.dev/play/p/nUCAOoJdeG6 )

Width and precision work the same way in strings built with `fmt.Sprintf` as they do in Question 1.

</details>

---

<details>
<summary>Trivia: Initial letters indicate a function’s role</summary>

The initial letters of fmt’s output functions indicate their role.

- No initial letter (`Print`, `Printf`, `Println`): writes to standard output.
- `S` (`Sprint`, `Sprintf`, `Sprintln`): returns a string.
- `F` (`Fprint`, `Fprintf`, `Fprintln`): writes to the specified `io.Writer`.

The final `f` means that the function takes a format string, while `ln` means that it adds a newline at the end and spaces between arguments. Knowing this convention lets you infer from the name that `Sprintf` is the string-returning, formatted version even when you see it for the first time.

</details>

---

## Starting points for investigation

- [Go Documentation](https://go.dev/doc/)
- https://pkg.go.dev/fmt
