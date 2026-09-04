[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 04-deep-dive](../../README.md)

# Intermediate: Add an Independent Deadline to context.WithoutCancel

Code must retain a value from its parent context while ignoring parent cancellation. It also needs an independent deadline so the derived operation cannot continue forever. [Run this in the Go Playground](https://go.dev/play/p/ArrX3G2xO82) and observe the error and value after canceling the parent.

```go
package main

import (
	"context"
	"fmt"
)

type markerKey struct{}

func main() {
	parent, cancel := context.WithCancel(
		context.WithValue(context.Background(), markerKey{}, "kept"),
	)
	detached := context.WithoutCancel(parent)
	cancel()

	fmt.Println("parent:", parent.Err())
	fmt.Println("detached:", detached.Err())
	fmt.Println("value:", detached.Value(markerKey{}))
}
```

```text
parent: context canceled
detached: <nil>
value: kept
```

---

## Question 1: What survives detaching from parent cancellation?

What does this derived context inherit, and how do `Deadline`, `Done`, and `Err` change?

<details><summary>Hint</summary>

Start with the release notes that introduced this API. Then read the package documentation and implementation, separating value lookup from cancellation notification.

</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read [Go 1.21 context release notes](https://go.dev/doc/go1.21#context).
2. Read [`context.WithoutCancel`](https://pkg.go.dev/context#WithoutCancel).
3. Check the fixed [context implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/context/context.go;l=592).

**Answer**

`context.WithoutCancel(parent)` can still find values from the parent, but is not canceled when the parent is canceled. It has no deadline, a `nil` `Done` channel, and a `nil` `Err`. Thus the parent reports `context canceled`, while the detached context still returns `kept`.

</details>

---

## Question 2: How do you prevent the derived operation from waiting forever?

How should the operation get its own deadline, and what does a `nil` `Done` channel mean for a `select`?

<details><summary>Hint</summary>

Compare the `Done` behavior with the language specification's rules for `nil` channels in `select`, then inspect `context.WithTimeout` and its cancellation function.

</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read [`Select statements`](https://go.dev/ref/spec#Select_statements).
2. Recheck [`WithoutCancel`](https://pkg.go.dev/context#WithoutCancel).
3. Read [`WithTimeout`](https://pkg.go.dev/context#WithTimeout) and the fixed [implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/context/context.go;l=592).

**Answer**

Detach the parent context with `WithoutCancel`, then wrap it in `WithTimeout` with a short independent deadline, and always call the returned `CancelFunc` when finished. A `nil` channel case is never selected, so `WithoutCancel` alone cannot provide a cancellation wait. The timeout keeps the operation bounded.

</details>

---

## Question 3: How do you test the separation?

Which two observations establish that the value remains available, parent cancellation is ignored, and the independent deadline is obeyed?

<details><summary>Hint</summary>

Turn the three output lines into test observations, then use a short timeout.

</details>

<details><summary>Answer</summary>

**Investigation route**

1. Confirm the purpose in the [Go 1.21 context notes](https://go.dev/doc/go1.21#context).
2. Read [`WithoutCancel`](https://pkg.go.dev/context#WithoutCancel) and [`WithTimeout`](https://pkg.go.dev/context#WithTimeout).
3. Reproduce the [execution example](https://go.dev/play/p/02PmNtj9kUO).

**Answer**

First, after canceling the parent, the derived context must still return the value and must not immediately return `context canceled`, reproducing `detached: <nil>` and `value: kept`. Second, the operation must end at its own deadline. Testing these separately proves that parent cancellation is detached without creating an unbounded wait.

</details>

---

## Research starting points

1. [How to investigate 04-deep-dive](../../README.md)
2. [Go 1.21 release notes: context](https://go.dev/doc/go1.21#context)
3. The [`context` package documentation](https://pkg.go.dev/context)
