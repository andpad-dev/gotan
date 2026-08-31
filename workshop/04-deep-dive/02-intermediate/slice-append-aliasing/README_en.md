[Workshop guide and scenario index](../../../README.md) | [How to research 04-deep-dive](../../README.md)

# Intermediate: Separate Notification Targets with a Three-Index Slice

A batch notifies the first two people in a shared assignment list and adds one on-call person. The notification view accidentally replaces the third person in the original project data. We must add the on-call person without changing the shared list.

[Run this in the Go Playground](https://go.dev/play/p/-4j2pYJbiJT) and compare an ordinary slice expression with a three-index expression.

```go
package main

import "fmt"

func main() {
	shared := []string{"aya", "bo", "chi"}
	view := shared[:2]
	view = append(view, "on-call")
	fmt.Printf("shared=%q\n", shared)
	fmt.Printf("view=%q\n", view)

	isolatedSource := []string{"aya", "bo", "chi"}
	isolated := isolatedSource[:2:2]
	isolated = append(isolated, "on-call")
	fmt.Printf("isolatedSource=%q\n", isolatedSource)
	fmt.Printf("isolated=%q\n", isolated)
}
```

```text
shared=["aya" "bo" "on-call"]
view=["aya" "bo" "on-call"]
isolatedSource=["aya" "bo" "chi"]
isolated=["aya" "bo" "on-call"]
```

---

## Question 1: Why was the third person replaced?

Immediately after `view := shared[:2]`, its length is 2. Why does appending `on-call` replace `shared`'s third element? Explain using `len`, `cap`, and the `append` rules.

<details><summary>Hint</summary>

Check the capacity produced by a two-index slice expression, then compare append behavior when capacity is sufficient and insufficient.

</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read [Slice expressions](https://go.dev/ref/spec#Slice_expressions).
2. Read [Appending and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices).
3. Compare the [execution example](https://go.dev/play/p/-4j2pYJbiJT).

**Answer**

`shared` has length and capacity 3. `shared[:2]` has length 2 but capacity 3, so appending one element keeps the new length within capacity and may reuse the same backing array. The value is written at index 2, which is also the third element visible through `shared`.

</details>

---

## Question 2: How do you isolate the notification append?

What does the third index in `isolatedSource[:2:2]` limit, and why does only the appended result contain `on-call`?

<details><summary>Hint</summary>

Read the complete three-index slice expression and compare its capacity with the length required after append.

</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read the capacity rule for complete [Slice expressions](https://go.dev/ref/spec#Slice_expressions).
2. Confirm that insufficient capacity makes `append` allocate a new backing array in [Appending and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices).
3. Optionally inspect [`runtime.growslice`](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/runtime/slice.go;l=178).

**Answer**

The third index sets the capacity. `isolatedSource[:2:2]` therefore has length 2 and capacity 2. Appending one element requires length 3, so `append` allocates a new backing array. Only `isolated` receives `on-call`; `isolatedSource` remains unchanged.

</details>

---

## Question 3: How should the intent be made explicit in business code?

How do you test that the source list is unchanged? What intent does a complete slice expression communicate compared with copying into a new slice from the start?

<details><summary>Hint</summary>

Turn the two pairs of output lines into expected values, then compare the specification's `copy` and `append` rules.

</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read [Appending and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices).
2. Return to [Slice expressions](https://go.dev/ref/spec#Slice_expressions).
3. Use the [execution example](https://go.dev/play/p/-4j2pYJbiJT) as the expected behavior.

**Answer**

The test should assert that the source remains `aya, bo, chi` and that only the notification list becomes `aya, bo, on-call`. A complete slice expression communicates “do not let the next append write beyond this range” through capacity. Alternatively, `append([]string(nil), source[:2]...)` explicitly creates an independent backing array. Both are shallow copies; choose based on whether capacity control or immediate independence is the clearest intent.

</details>

---

## Research starting points

1. [How to investigate 04-deep-dive](../../README.md)
2. [Go specification: Slice expressions](https://go.dev/ref/spec#Slice_expressions)
3. [Go specification: Appending and copying slices](https://go.dev/ref/spec#Appending_and_copying_slices)
