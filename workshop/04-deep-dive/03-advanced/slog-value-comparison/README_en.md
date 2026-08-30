# Why Can't `slog.Value` Be Compared with `==`?

The following complete program does not compile:

```go
package main

import (
	"fmt"
	"log/slog"
)

func main() {
	v1 := slog.StringValue("gopher")
	v2 := slog.StringValue("gopher")
	fmt.Println(v1 == v2)
}
```

([Verify it in the Go Playground](https://go.dev/play/p/MPeWX_kJChR))

```text
invalid operation: v1 == v2 (struct containing [0]func() cannot be compared)
```

## Question 1: Identify the mechanism that forbids comparison

<details><summary>Answer</summary>

**Investigation route**

1. Start at [Go Documentation](https://go.dev/doc/) and open the standard library's `log/slog` package.
2. Open [Go 1.27.0 `src/log/slog/value.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/value.go;l=21).

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

Distinguish the size of `[0]func()` itself from the size of a struct that places the field first or last. Explain why this mechanism does not increase `slog.Value` in the current implementation.

```go
package main

import (
	"fmt"
	"log/slog"
	"unsafe"
)

type withArrayFirst struct {
	_ [0]func()
	num uint64
	any any
}

type withoutArray struct {
	num uint64
	any any
}

type withArrayLast struct {
	num uint64
	any any
	_ [0]func()
}

func main() {
	fmt.Println(unsafe.Sizeof([0]func(){}))
	fmt.Println(unsafe.Sizeof(withArrayFirst{}))
	fmt.Println(unsafe.Sizeof(withoutArray{}))
	fmt.Println(unsafe.Sizeof(withArrayLast{}))
	fmt.Println(unsafe.Sizeof(slog.Value{}))
}
```

([Run it in the Go Playground](https://go.dev/play/p/w9nhI8d9nnl))

```text
0
24
24
32
24
```

<details><summary>Hint</summary>Read [Size and alignment guarantees](https://go.dev/ref/spec#Size_and_alignment_guarantees) for what the specification guarantees about the zero-length array itself. Then search the Go 1.27.0 compiler source for `zero-sized field` to explain why only the trailing layout grows.</details>

<details><summary>Answer</summary>

**Investigation route**

1. Run the [shared size experiment](https://go.dev/play/p/w9nhI8d9nnl) and observe the difference between first and last placement.
2. Read [Size and alignment guarantees](https://go.dev/ref/spec#Size_and_alignment_guarantees).
3. Read the trailing zero-size-field padding in [Go 1.27.0 `cmd/compile/internal/types/size.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/types/size.go;l=516) and follow its reference to [Issue #9401](https://github.com/golang/go/issues/9401).

**Answer**

The specification guarantees that `[0]func()` itself has size zero. It does not guarantee that adding it anywhere in a struct leaves the struct's total size unchanged. The gc compiler pads a non-zero-size struct ending in a zero-size field so that taking the field's address cannot point into the next heap object. The first-position layout stays at 24 bytes, while the trailing layout grows to 32 bytes. `slog.Value` deliberately puts the field first, importing incomparability while remaining 24 bytes in Go 1.27.0.

</details>

---

## Question 4: Why forbid `==` at all?

Run this example to separate string contents from their data pointers, then observe both the successful and panicking boundaries of `Value.Equal`.

```go
package main

import (
	"fmt"
	"log/slog"
	"strings"
	"unsafe"
)

func main() {
	literal := "X"
	computed := strings.ToUpper("x")
	fmt.Println("strings equal:", literal == computed)
	fmt.Println("data pointers equal:", unsafe.StringData(literal) == unsafe.StringData(computed))
	fmt.Println("Value.Equal string:", slog.StringValue(literal).Equal(slog.StringValue(computed)))

	sliceValue := slog.AnyValue([]int{1})
	func() {
		defer func() {
			fmt.Println("Value.Equal slice panicked:", recover() != nil)
		}()
		fmt.Println(sliceValue.Equal(sliceValue))
	}()
}
```

([Run it in the Go Playground](https://go.dev/play/p/d6m0c3sgC6F))

```text
strings equal: true
data pointers equal: false
Value.Equal string: true
Value.Equal slice panicked: true
```

<details><summary>Hint</summary>Read the comments on the `num` and `any` fields and the interface comparison rule in [Comparison operators](https://go.dev/ref/spec#Comparison_operators). Follow the history of the `disallow ==` line to find the change that states the design reason.</details>

<details><summary>Answer</summary>

**Investigation route**

1. Run the [shared comparison experiment](https://go.dev/play/p/d6m0c3sgC6F).
2. Read the field comments in [Go 1.27.0 `value.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/value.go;l=21).
3. Read the interface rules in [Comparison operators](https://go.dev/ref/spec#Comparison_operators).
4. Follow the file history to [CL 479516](https://go-review.googlesource.com/c/go/+/479516) and [Issue #56345](https://github.com/golang/go/issues/56345).
5. Read [`Value.Equal`](https://pkg.go.dev/log/slog#Value.Equal) and its [Go 1.27.0 implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/value.go;l=420).

**Answer**

Even if `==` compiled, it would not reliably compare `slog.Value` contents. String values store their length in `num` and a pointer-like representation in `any`. The literal `"X"` and runtime-produced `strings.ToUpper("x")` have equal contents but different data pointers, reproducing the condition described in CL 479516. Also, `any` can contain an incomparable dynamic value such as a slice; comparing two interfaces with that dynamic type panics at runtime.

The type therefore rejects an unreliable and potentially panicking comparison. [`Value.Equal`](https://pkg.go.dev/log/slog#Value.Equal) compares ordinary values by content, but it can itself panic when `KindAny` or `KindLogValuer` contains a non-comparable dynamic value. Callers must account for the kind and dynamic type rather than treating `Equal` as an unconditional safe replacement.

</details>

---

<details><summary>Trivia: Why is the field first?</summary>

Putting `_ [0]func()` at the end can add padding (24 bytes at the beginning versus 32 at the end). The compiler adds it so that the address of a trailing zero-size field cannot point into the next object; see [Issue #9401](https://github.com/golang/go/issues/9401) and the [Go 1.27.0 implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/types/size.go;l=516).

</details>

---

## Research starting points

- [Go Documentation](https://go.dev/doc/)
- [Go 1.27.0 `slog.Value` source](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/log/slog/value.go;l=21)
- [Go language specification](https://go.dev/ref/spec)
- [CL 479516](https://go-review.googlesource.com/c/go/+/479516) and [Issue #56345](https://github.com/golang/go/issues/56345)
