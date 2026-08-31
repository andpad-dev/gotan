[Workshop guide and scenario index](../../../README.md) | [How to research 01-packages](../../README.md)

# Track Old Notifications from `time.Timer` `Stop` / `Reset`

An old incident report says that a retry job for a payment integration received a notification for the previous deadline immediately after stopping and resetting a timer. Current observations show that the channel capacity is 0. Why is it like this? Let's investigate the background.

When you [run the following code in the Go Playground](https://go.dev/play/p/F4r40BDO4dv), you can confirm the current timer-channel capacity and the reset after stopping.

This scenario includes a `go.mod` that enables the timer-channel behavior introduced in Go 1.23 and later. When trying it locally, run the code in this directory.

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	timer := time.NewTimer(time.Hour)
	fmt.Println("channel capacity:", cap(timer.C))
	fmt.Println("stop before firing:", timer.Stop())

	timer.Reset(time.Millisecond)
	<-timer.C
	fmt.Println("reset timer fired")
}
```

Output:

```text
channel capacity: 0
stop before firing: true
reset timer fired
```

---

## Question 1: How does capacity 0 differ from the incident report?

Before Go 1.23, timer channels had a buffer capacity of 1. With the current synchronous channel of capacity 0, what guarantee is provided about an old retry notification after `Stop` or `Reset` returns?

<details>
<summary>Hint</summary>

- Read the `Before Go 1.23` section in the documentation for `time.NewTimer`.
- In the “Timer changes” section of the Go 1.23 release notes, look for what was called a “stale value.”

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read “Timer changes” in the [Go 1.23 Release Notes](https://go.dev/doc/go1.23).
2. Read the version-specific explanation in [time.NewTimer](https://pkg.go.dev/time#NewTimer).
3. Read the comments in [sleep.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=123).

**Answer**

With the new behavior since Go 1.23, a timer channel is a synchronous channel with capacity 0. It guarantees that after `Stop` or `Reset` returns, a stale time value prepared before that call will not be sent or received.

The old channel with capacity 1 could retain an old notification, which made stopping and resetting more complicated. This is what the retry-job incident report refers to.

</details>

---

## Question 2: Why is the old “drain” dangerous?

Previously, it was common to unconditionally receive from `timer.C` to empty it when `Stop` returned `false`; this was known as draining. With the new semantics since Go 1.23, why should that receive not be left unconditional? For a library that also supports older Go versions, what assumptions should be made explicit?

When running comparison experiments, use the introductory Go Playground result as the baseline, and record `go version`, `go env GOMOD`, and `go env GODEBUG` locally as well. In particular, running without a `go.mod` or using a module below `go 1.23` can produce old timer-channel behavior even with the same Go compiler.

<details>
<summary>Hint</summary>

- Read the explanation of `Stop` and old buffered channels in `NewTimer`.
- Distinguish between “a value remains in the channel” and “the timer has fired but the send has not completed.”
- Confirm that even when `Stop` returns `false`, a value does not necessarily remain buffered. This includes the case where another goroutine received it first; observe unconditional receives with a `select` and a timeout.
- The [minimal experiment](https://go.dev/play/p/X_a-_2PuQoV) receives the notification once, then calls `Stop` and tries a subsequent receive with a timeout. Do not equate `Stop=false` with “a receive from `C` is immediately available.”
- Investigate the main program's `go.mod` `go` line and `GODEBUG=asynctimerchan` as the conditions that switch to the old behavior.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Re-read Timer changes in the [Go 1.23 Release Notes](https://go.dev/doc/go1.23).
2. Read the explanation of `Stop` / `Reset` in [time.NewTimer](https://pkg.go.dev/time#NewTimer).
3. Confirm the background of capacity 0 and stale values in [Issue #37196](https://github.com/golang/go/issues/37196) and the [time package implementation comments](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=133).

**Answer**

With the new semantics, draining to receive an old time value is unnecessary. Even if `Stop` returns `false`, that does not mean that a value is buffered in the channel. With a synchronous channel, another goroutine may already have received the notification, or the old notification may have been invalidated, so an unconditional receive may wait indefinitely.

The behavior differs when older Go versions are also supported. The library should make explicit the minimum supported Go version, the `go` line in the main program's `go.mod`, how to handle `GODEBUG=asynctimerchan`, and which goroutine owns receives from the timer channel. A drain for the old buffered channel is valid only under the old assumption that there is a single receiver. Rather than simply adding or removing the old idiom, confirm as part of the design which semantics the synchronization relies on.

</details>

---

## Question 3: How are `go.mod` and `GODEBUG` related?

In Go 1.23, the module's `go` line was involved in enabling the new behavior. How did `asynctimerchan` change in Go 1.27? Explain why a temporary compatibility setting should not become a permanent solution during migration.

In this question, first observe the “new behavior” through the `cap` / `len` of the channel from `time.NewTimer(0)`. For “compatibility settings,” distinguish the environment variable `GODEBUG`, the `godebug` directive in `go.mod`, and `//go:debug` in source code, and organize which execution conditions each one affects.

<details>
<summary>Hint</summary>

- Find the conditions for enabling the new behavior in the Go 1.23 release notes.
- Search for `asynctimerchan` in the Go 1.26 and Go 1.27 release notes.
- Run the [default experiment](https://go.dev/play/p/zAkeGmN14Q7), the [experiment specifying the old behavior](https://go.dev/play/p/o-pDN23HyaX), and the [experiment specifying the new behavior](https://go.dev/play/p/V8esF3Hmib8) in order, and confirm the difference between `cap=0` and `cap=1`. If Go 1.27 has not been released or cannot be selected in the Playground, treat the release notes as the “planned specification” and do not mix them with execution results.
- These Playgrounds use Go 1.26.5, and the expected outputs are `go1.26.5 0 0`, `go1.26.5 1 0`, and `go1.26.5 0 0`, respectively. If your local or event version differs, record the actual version and compare them.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. In the [Go 1.23 Release Notes](https://go.dev/doc/go1.23), read the explanation of modules at `go 1.23.0` or later and `asynctimerchan=1`.
2. In the [Go 1.26 Release Notes](https://go.dev/doc/go1.26), confirm the planned change starting in Go 1.27.
3. Read the Runtime and GODEBUG explanations in the [Go 1.27 Release Notes](https://go.dev/doc/go1.27).

**Answer**

In Go 1.23, the new timer behavior was enabled when the main program's module had `go 1.23.0` or later, while `asynctimerchan=1` was a setting for restoring the old behavior during investigation or migration.

In Go 1.27, this setting was permanently removed, and `time` timer channels are always synchronous and unbuffered. Therefore, instead of relying on a compatibility setting to hide incidents, fix old drain logic and dependencies on `len` / `cap` so that the code works correctly under the current semantics.

</details>

---

## Question 4: Why was this change discussed for so long?

The problem of receiving an old notification is not merely about API appearance; it concerns the invariants required to use `Stop` / `Reset` correctly. Read the proposal issue and implementation comments, then explain to the team in one minute what difficulty the Go team was trying to reduce.

For example, consider what happens if another goroutine receives from `timer.C` while the sequence “stop the timer -> empty the old notification -> set the deadline again” is in progress. Compare what users previously had to remember to avoid confusing an old notification with a new one against the guarantees of the current API.

<details>
<summary>Hint</summary>

- Read the former explanation at the beginning of the issue that a value might exist after `Stop`.
- Look for `stale time values` in the implementation comments.
- Read the old drain example at the beginning of [Issue #37196](https://github.com/golang/go/issues/37196) alongside the `Stop` / `NewTimer` comments in [Go 1.26.4's sleep.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=105).

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Confirm the overview of the timer changes in the [Go 1.23 Release Notes](https://go.dev/doc/go1.23).
2. Read the problem statement and discussion in [Issue #37196](https://github.com/golang/go/issues/37196).
3. Confirm the final guarantee in the implementation comments of [sleep.go](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/time/sleep.go;l=133).

**Answer**

Previously, an old deadline value could remain buffered after stopping or resetting. Correctness required carefully combining return values, draining, and the possibility of another goroutine receiving the value, which made the situation easy to misuse.

The new design gives the API a strong guarantee that an old value cannot be received later. This reduces the need for users to guess whether a notification just received belongs to before or after the reset, making the timer lifecycle simpler to handle.

</details>

---

<details>
<summary>Further note</summary>

Code that checks the `len` or `cap` of a timer channel to determine whether it can be received from is affected by the migration. To check whether a value has arrived, also review the release note's guidance to use a non-blocking `select`.

</details>

---

## Investigation starting points

- [Go 1.27 Release Notes](https://go.dev/doc/go1.27)
- [How to investigate 01-packages](../../README.md)
- [package time](https://pkg.go.dev/time)
