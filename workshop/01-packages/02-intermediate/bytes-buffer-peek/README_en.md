[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 01-packages](../../README.md)

# Inspect the Beginning of bytes.Buffer Without Reading It

During review, you find code that inspects the first four bytes of a `bytes.Buffer` without consuming them. Investigate the relationship between the value returned by `Peek` and the buffer.

When you [run the following observation code in the Go Playground](https://go.dev/play/p/OdJEVvlBSbf), you can observe three properties: peeking ahead, insufficient data, and sharing of the returned slice.

```go
package main

import (
	"bytes"
	"fmt"
)

func main() {
	buffer := bytes.NewBufferString("ABCDrest")
	prefix, err := buffer.Peek(4)
	fmt.Printf("prefix=%q err=%v remaining=%q\n", prefix, err, buffer.String())

	prefix[0] = 'a'
	fmt.Printf("after mutation remaining=%q\n", buffer.String())

	short := bytes.NewBufferString("XY")
	got, err := short.Peek(4)
	fmt.Printf("short=%q err=%v\n", got, err)
}
```

Output:

```text
prefix="ABCD" err=<nil> remaining="ABCDrest"
after mutation remaining="aBCDrest"
short="XY" err=EOF
```

---

## Question 1: How can you inspect the beginning without reading it?

In the first output, `prefix` is `"ABCD"`, but `remaining` is still `"ABCDrest"`. What does `Peek(4)` do, and how would using `Next(4)` differ?

<details>
<summary>Hint</summary>

- Use [How to investigate a category](../../README.md) as your starting point and open the list of `bytes.Buffer` methods.
- Compare whether `Peek` and `Next` advance the buffer in their descriptions.
- Try both from the same position with `buffer := bytes.NewBufferString("ABCDrest")`.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open the `bytes` package from the [Go Documentation](https://go.dev/doc/).
2. Compare [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek) and [Buffer.Next](https://pkg.go.dev/bytes#Buffer.Next).
3. Read the [`Peek` implementation in buffer.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/bytes/buffer.go;l=85) to confirm that it does not advance the offset.

**Answer**

`Peek(4)` returns the next 4 bytes but does not advance the buffer's reading position. A later read therefore receives the same contents.

`Next(4)` returns 4 bytes and advances the buffer as though they had been read. `Peek` is appropriate when the beginning must remain unread.

</details>

---

## Question 2: What happens when there are fewer than 4 bytes?

Why does `Peek(4)` for `"XY"` return `"XY"` and `EOF` instead of an empty slice? How should the return values be handled?

<details>
<summary>Hint</summary>

- Look for the explanation of “fewer than n bytes” in `Peek`.
- Notice that this API returns partial data and an error at the same time.
- The buffer currently contains only 2 bytes. Consider why the partial data and `io.EOF` are returned separately.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read the error description for [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek).
2. In the [implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/bytes/buffer.go;l=85), confirm the branch that returns the remaining bytes and `io.EOF`.

**Answer**

When fewer bytes are available than the requested 4, `Peek` returns the bytes currently available and, at the same time, `io.EOF`.

The caller should check `err`. It must not process the partial data in `got` as a complete result when the required length is unavailable.

</details>

---

## Question 3: Why does changing `prefix[0]` also change the buffer?

When the first byte of the returned `prefix` is changed to `a`, the entire buffer also becomes `"aBCDrest"`. What should you be careful about, and when should you make a copy?

<details>
<summary>Hint</summary>

- In the `Peek` description, look for `valid until` and `aliases`.
- If you retain or modify the returned value, consider whether it occupies the same memory as the original buffer.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read the explanation of the returned value's validity period and aliasing in [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek).
2. Check the slice expression in [buffer.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/bytes/buffer.go;l=85).

**Answer**

The slice returned by `Peek` shares the buffer's contents. It is no longer valid after calling a read or write method, and changing the returned slice can also change what is read next. Do not destructively modify a value returned for observation.

If the returned value must be retained or modified, copy it to another slice first. A copy is unnecessary when the slice is only read and used immediately.

For example, `copyOfPrefix := bytes.Clone(prefix)` preserves the original 4 bytes even if `prefix[0]` is changed later. Decide whether a copy is needed based on how the value is used.

</details>

---

<details>
<summary>Further note</summary>

`Buffer.Peek` was added in Go 1.26. For modules that also target older toolchains, check the minimum Go version and the `go` line in `go.mod` before adopting it.

</details>

---

## Investigation starting points

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [How to investigate 01-packages](../../README.md)
- [package bytes](https://pkg.go.dev/bytes)
