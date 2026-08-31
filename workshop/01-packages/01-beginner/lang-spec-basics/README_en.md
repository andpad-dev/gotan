[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 01-packages](../../README.md)

# Investigate Go `for` and `switch` statements in the specification

You are maintaining Go code again after a long break and encounter `for` and `switch` statements during review. Instead of relying on half-remembered syntax from another language, start from observed output and find the corresponding rules in the Go language specification.

---

## Question 1: `for` statements

How does Go express iteration like `for value in values` in other languages? Observe how the following code prints each element of the `values` slice, then investigate the two values produced by `range` and the role of `_`.

```go
package main

import "fmt"

func main() {
	values := []string{"a", "b"}
	for _, value := range values {
		fmt.Println(value)
	}
}
```

[Run it in the Go Playground](https://go.dev/play/p/bJ7XfrbuNHo).

```text
a
b
```

<details>
<summary>Hint</summary>

- Try both `for value in values` and `for _, value := range values`, and compare the syntax error with the working form.
- After opening “For statements” in the table of contents, continue past the parent grammar to “For statements with range clause.”

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Run the [shared Go Playground code](https://go.dev/play/p/bJ7XfrbuNHo) and observe the slice elements being printed in order.
2. Open the [Go language specification](https://go.dev/ref/spec), select “For statements,” and identify the three `ForStmt` forms.
3. Continue to [For statements with range clause](https://go.dev/ref/spec#For_range), then check the `RangeClause` syntax and the first and second iteration values for a slice.

**Answer**

A `for` statement can contain only a condition, an initializer/condition/post clause, or a `range` clause. Go has no `for value in values` syntax; collection iteration uses `range`.

```
ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
```

For a slice, the first `range` value is the index and the second is the element. The opening `for _, value := range values` discards the unused index with the blank identifier `_` and receives only the element in `value`.

</details>

---

## Question 2: `switch` statements

Unlike C’s `switch`, Go’s `switch` does not require `break`. In the following code, `x == 1` prints only `one`. How can you continue into the next case body?

```go
package main

import "fmt"

func main() {
	x := 1
	switch x {
	case 1:
		fmt.Println("one")
	case 2:
		fmt.Println("two")
	}
}
```

[Run it in the Go Playground](https://go.dev/play/p/FhO9_Qu5-ca).

```text
one
```

<details>
<summary>Hint</summary>

- After observing that `x == 1` prints only `one`, continue from “Switch statements” to its “Expression switches” subsection.
- Check whether the next case expression is evaluated again or control transfers unconditionally.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Run the [shared Go Playground code](https://go.dev/play/p/FhO9_Qu5-ca) and observe that `x == 1` prints only `one`.
2. Open the [Go language specification](https://go.dev/ref/spec) and select “Switch statements.”
3. In [Expression switches](https://go.dev/ref/spec#Expression_switches), check the implicit break and `fallthrough` rules.

**Answer**

Placing `fallthrough` as the final statement of `case 1` transfers control to the first statement of the next case without evaluating that case expression.

```go
package main

import "fmt"

func main() {
	x := 1
	switch x {
	case 1:
		fmt.Println("one")
		fallthrough
	case 2:
		fmt.Println("two")
	}
}
```

[Run it in the Go Playground](https://go.dev/play/p/fOZxxOIZNDA). Both case bodies now execute:

```text
one
two
```


</details>

---

<details>
<summary>Trivia: Staying oriented in the specification</summary>

The [Go language specification](https://go.dev/ref/spec) contains its table of contents and sections in one HTML page. A parent section may contain only the grammar while detailed rules live in subsections. Sharing the exact section link—`#For_range` for this `range` example or `#Expression_switches` for the switch—lets every team member open the same evidence.

</details>

---

## Starting points for investigation

- [Go Documentation](https://go.dev/doc/)
- [How to research 01-packages](../../README.md)
- [Go language specification](https://go.dev/ref/spec)
