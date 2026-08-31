[Workshop guide and scenario index](../../../README.md) | [How to research 04-deep-dive](../../README.md)

# Why Can't `slog.Value` Be Compared with `==`?

The following code does not compile:

```go
v1 := slog.StringValue("gopher")
v2 := slog.StringValue("gopher")
fmt.Println(v1 == v2)
// invalid operation: v1 == v2 (struct containing [0]func() cannot be compared)
```

([Verify it in the Go Playground](https://go.dev/play/p/MPeWX_kJChR))

## Question 1: Identify the mechanism that forbids comparison

<details><summary>Answer</summary>

**Investigation route**

1. Open [Go 1.26.5 `src/log/slog/value.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/log/slog/value.go;l=22).

**Answer**

```go
type Value struct {
	_ [0]func() // disallow ==
	...
}
```

The zero-length array of functions is incomparable. Its blank field name gives the struct the type property without requiring initialization or access.

</details>

---

## Question 2: Why does this mechanism work according to the specification?

<details><summary>Hint</summary>Read the type rules in [Comparison operators](https://go.dev/ref/spec#Comparison_operators) for structs, arrays, and functions.</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read [Comparison operators](https://go.dev/ref/spec#Comparison_operators).
2. Follow the comparability rules for struct, array, and function types.

**Answer**

Function types are not comparable. An array is comparable only when its element type is comparable, so `[0]func()` is not comparable even though its length is zero. A struct is comparable only when all its fields are comparable. Therefore `slog.Value` is not comparable and `v1 == v2` produces the demonstrated error: https://go.dev/play/p/MPeWX_kJChR

</details>

---

## Question 3: Why use a zero-length array?

<details><summary>Hint</summary>Read [Size and alignment guarantees](https://go.dev/ref/spec#Size_and_alignment_guarantees), then verify with `unsafe.Sizeof`.</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read [Size and alignment guarantees](https://go.dev/ref/spec#Size_and_alignment_guarantees).
2. Measure the types with `unsafe.Sizeof`.

**Answer**

The specification guarantees that a struct or array has size zero when it contains no field or element whose size is greater than zero. Thus `[0]func()` imports incomparability without consuming memory. The measurement is available at https://go.dev/play/p/SVs4LjlKXGo:

```go
fmt.Println(unsafe.Sizeof([0]func(){})) // 0
fmt.Println(unsafe.Sizeof(withArray{})) // 24 (with field)
fmt.Println(unsafe.Sizeof(withoutArray{})) // 24 (without field)
```

</details>

---

## Question 4: Why forbid `==` at all?

<details><summary>Hint</summary>Read the comments on the `num` and `any` fields. Also read the interface comparison rule in [Comparison operators](https://go.dev/ref/spec#Comparison_operators).</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read the field comments in `value.go`.
2. Read the interface rules in [Comparison operators](https://go.dev/ref/spec#Comparison_operators).
3. Find the comparison method in [pkg.go.dev/log/slog](https://pkg.go.dev/log/slog).

**Answer**

Even if `==` compiled, it would not reliably compare `slog.Value` contents. String values are stored by splitting their length into `num` and a pointer-like representation into `any`, so comparing the representation could compare pointers rather than string contents. Also, `any` can contain an incomparable dynamic value such as a slice; comparing two interfaces with that dynamic type panics at runtime. The type therefore rejects an unreliable and potentially panicking comparison. Use [`Value.Equal`](https://pkg.go.dev/log/slog#Value.Equal) instead.

</details>

---

<details><summary>Trivia: Why is the field first?</summary>

Putting `_ [0]func()` at the end can add padding (24 bytes at the beginning versus 32 at the end). A zero-size field at the end can also cause its address to point beyond the struct, so placing it first is the established pattern.

</details>

---

## Research starting points

- [slog Value source](https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/log/slog/value.go;l=22)
- [Go language specification](https://go.dev/ref/spec)
