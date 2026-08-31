[Workshop guide and scenario index](../../../README.md) | [How to research 04-deep-dive](../../README.md)

# Investigate nil and empty slices in encoding/json

While reviewing the response from a project-details API with the frontend team, you found that projects without assignees returned only `{"assignees":null}`. The frontend always wants to process an array, so the contract requires `{"assignees":[]}` when the list is empty.

You came across the following `json.Marshal` call in a review. Let’s investigate what it does.

First, observe how two values that appear empty produce different output by [running the example in the Go Playground](https://go.dev/play/p/M_lGwwe4dMX).

```go
package main

import (
	"encoding/json"
	"fmt"
)

type assigneeResponse struct {
	Assignees []string `json:"assignees"`
}

func main() {
	nilJSON, _ := json.Marshal(assigneeResponse{})
	emptyJSON, _ := json.Marshal(assigneeResponse{Assignees: []string{}})
	fmt.Println("nil:", string(nilJSON))
	fmt.Println("empty:", string(emptyJSON))
}
```

This is the output with Go 1.26.4.

```text
nil: {"assignees":null}
empty: {"assignees":[]}
```

---

## Question 1: Why is the JSON different when both lengths are 0?

Both `assigneeResponse{}` and `assigneeResponse{Assignees: []string{}}` have a `len` of 0. Explain why one becomes `null` and the other `[]`, using the nature of slice values and the rules of `json.Marshal`.

<details>
<summary>Hint</summary>

Use the slice type specification to investigate how a merely declared value differs from one created by an empty composite literal. Then search the standard library’s JSON encoding rules for slices.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open [Slice types in the Go language specification](https://go.dev/ref/spec#Slice_types) and confirm that nil slices and initialized empty slices are distinct.
2. Follow the specification page’s standard-library link to [encoding/json `Marshal`](https://pkg.go.dev/encoding/json#Marshal) and read the JSON encoding rule for slices.
3. Follow the documentation to the implementation and confirm in the fixed-version [`sliceEncoder.encode`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/encoding/json/encode.go;l=843) source that `null` is written when `IsNil` is true.

**Answer**

The zero value of `[]string` is a nil slice. In contrast, `[]string{}` is an initialized empty slice even though its length is 0. `encoding/json` encodes a nil slice as JSON `null` and an initialized slice as a JSON array. Therefore, the first value becomes `{"assignees":null}` and the second becomes `{"assignees":[]}`.

</details>

---

## Question 2: How do you return an empty array according to the API contract?

This API’s contract keeps the `assignees` field and represents no assignees as `[]`. How should you construct the response instead of adding `omitempty`? Why does `omitempty` not fit this contract?

<details>
<summary>Hint</summary>

Check the field-tag documentation for `Marshal` and see how a slice of length 0 is handled. Distinguish among `null`, `[]`, and a missing field.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Use [Slice types in the Go language specification](https://go.dev/ref/spec#Slice_types) to confirm how to initialize an empty slice.
2. Read the explanation of `omitempty` in [encoding/json `Marshal`](https://pkg.go.dev/encoding/json#Marshal) and confirm that a slice of length 0 is considered empty.
3. Compare the execution result with the [fixed-version implementation from Question 1](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/encoding/json/encode.go;l=843).

**Answer**

Explicitly initialize the slice when constructing the response, for example with `Assignees: []string{}`. JSON will then contain `[]`, and the field will remain present. `omitempty` omits a slice of length 0 together with the field, so it cannot express this API’s required meaning: there are no assignees, but the list field exists.

Whether an API contract uses `null`, an empty array, or no field at all is a design choice. This frontend contract chooses `[]`, which can always be iterated as an array, so return an initialized empty slice.

</details>

---

## Starting points for investigation

1. Open [How to investigate 04-deep-dive](../../README.md) and begin with the Go language specification.
2. Use [Slice types in the Go language specification](https://go.dev/ref/spec#Slice_types) to investigate nil and empty slices.
3. Follow the specification page to the standard library documentation and check `encoding/json`.
