[Workshop guide and scenario index](../../../README.md) | [How to research 01-packages](../../README.md)

# Investigate `for` syntax in the specification

You are coding in Go again after a very long break.
You come across a `for` statement. Let’s investigate what it does.

You have also forgotten some details of the language specification, so let’s read the Go language specification.

## Question 1: `for` statements

A counter variable is straightforward, but there was probably syntax like `in`. How is this written in Go?

<details>
<summary>Hint</summary>

- First prepare `values := []string{"a", "b"}` and try `for value in values` and `for _, value := range values`. Compare the one that produces an error with the one that works, then open “For statements” in the specification.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open https://go.dev/ref/spec and select “For statements” from the table of contents.
2. Check the syntax of `for` statements.

**Answer**

- `for` statements have the following three forms.
- There is no syntax like `in`, but there is a form using the `range` keyword, which serves a similar purpose.

```
ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
```

[do it](https://go.dev/play/p/WCZlhaPrgG_p)

</details>

---

## Question 2: `switch` statements

Unlike C’s `switch`, Go’s `switch` does not require `break`. How can you make the next case execute after one case matches, as it does in C?

<details>
<summary>Hint</summary>

- Set `x := 1` and create a short switch that prints `one` for `case 1` and `two` for `case 2`. If you want `two` to print when `x` is 1, find the required syntax in the specification.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open https://go.dev/ref/spec and select “Switch statements” from the table of contents.
2. Check the explanation of the `fallthrough` keyword.

**Answer**

- The `fallthrough` keyword makes the next case execute after a case matches, as in a C `switch` statement.

```go
switch x {
case 1:
    fmt.Println("one")
    fallthrough
case 2:
    fmt.Println("two")
}
```

[just do it](https://go.dev/play/p/HumLQAbZ-eQ)


</details>

---

<details>
<summary>Trivia: </summary>

The [Go language specification](https://go.dev/ref/spec) is unusually short for a programming language specification and is notable for fitting into a single HTML page. While C++ and Java specifications resemble thick dictionaries, Go’s specification was designed with the intention that it could be read through in a single afternoon.

</details>

---

## Starting points for investigation

- https://go.dev/ref/spec
