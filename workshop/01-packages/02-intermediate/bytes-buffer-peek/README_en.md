[Workshop guide and scenario index](../../../README.md) | [How to research 01-packages](../../README.md)

# Peek Ahead at the Header of an Incoming Message

In backend message reception, the first 4 bytes of a message from an external system indicate its type, followed by the body. The receiving handler wants to inspect the type and then pass the message to the appropriate processing logic, but must not consume the body yet. Let's investigate how to use `bytes.Buffer.Peek` to inspect the beginning of a received message.

When you [run the following observation code in the Go Playground](https://go.dev/play/p/FyP4yK-B61w), you can observe three properties: peeking ahead, insufficient data, and sharing of the returned slice.

```go
package main

import (
	"bytes"
	"fmt"
)

func main() {
	packet := bytes.NewBufferString("HDR:payload")
	header, err := packet.Peek(4)
	fmt.Printf("header=%q err=%v remaining=%q\n", header, err, packet.String())

	header[0] = 'h'
	fmt.Printf("after mutation remaining=%q\n", packet.String())

	short := bytes.NewBufferString("OK")
	got, err := short.Peek(4)
	fmt.Printf("short=%q err=%v\n", got, err)
}
```

Output:

```text
header="HDR:" err=<nil> remaining="HDR:payload"
after mutation remaining="hDR:payload"
short="OK" err=EOF
```

---

## Question 1: How can you inspect the type without reading it?

In the first output, `header` is `"HDR:"`, but `remaining` is also `"HDR:payload"`. What does `Peek(4)` do, and how would using `Next(4)` for the same purpose differ?

<details>
<summary>Hint</summary>

- Use [How to investigate a category](../../README.md) as your starting point and open the list of `bytes.Buffer` methods.
- Compare whether `Peek` and `Next` advance the buffer in their descriptions.
- Think of `Peek` as placing a bookmark in a book and looking at the page, while `Next` turns the page and advances the reading position. Trying both from the same position with `packet := bytes.NewBufferString("HDR:payload")` makes the difference easy to see.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open the `bytes` package from the [Go Documentation](https://go.dev/doc/).
2. Compare [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek) and [Buffer.Next](https://pkg.go.dev/bytes#Buffer.Next).
3. Read the [`Peek` implementation in buffer.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/bytes/buffer.go;l=85) to confirm that it does not advance the offset.

**Answer**

`Peek(4)` returns the next 4 bytes but does not advance the buffer's reading position. As a result, even after passing the type to the handler, processing that reads the complete packet, including the body, can receive the same contents.

`Next(4)` returns 4 bytes and advances the buffer as though they had been read. It can be used when you want to classify and consume the type at the same time, but `Peek` is appropriate for the current use case of looking first and then passing the message onward.

</details>

---

## Question 2: What happens when there are fewer than 4 bytes?

Why does `Peek(4)` for `"OK"` return `"OK"` and `EOF` instead of an empty slice? In processing that detects an incomplete reception, how should the return values be handled?

<details>
<summary>Hint</summary>

- Look for the explanation of “fewer than n bytes” in `Peek`.
- Notice that this API returns partial data and an error at the same time.
- `"OK"` does not necessarily mean that the header is corrupt; it may mean that only 2 bytes have arrived in the current buffer. Organize the reasons for handling the partial data and `io.EOF` separately.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read the error description for [Buffer.Peek](https://pkg.go.dev/bytes#Buffer.Peek).
2. In the [implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/bytes/buffer.go;l=85), confirm the branch that returns the remaining bytes and `io.EOF`.

**Answer**

When fewer bytes are available than the requested 4, `Peek` returns the bytes currently available and, at the same time, `io.EOF`. Rather than discarding the information by returning nothing, it represents the state as “reception has progressed this far, but there are not enough bytes for a header.”

The receiving logic should check `err` and, when it is `io.EOF`, receive more data before trying again. It must not process the partial data in `got` as a complete header.

</details>

---

## Question 3: Why does changing `header[0]` also change the body?

When the first byte of the returned `header` is changed to `h`, the entire buffer also becomes `"hDR:payload"`. When you want to pass the type safely to downstream processing and then read or write the buffer, what should you be careful about, and when should you make a copy?

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

If downstream processing needs to retain or modify the type, copy it to another slice first. A copy is unnecessary when the slice is only read and used immediately. The key is to distinguish zero-copy peeking for performance from data whose ownership should be separated.

For example, `safeHeader := bytes.Clone(header)` preserves the original 4 bytes in `safeHeader` even if `header[0]` is changed later. Decide whether a copy is needed based on whether you will pass the value around for a while or only read it in place.

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
