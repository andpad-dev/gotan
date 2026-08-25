# Beginner: Investigate defer argument evaluation and execution order

An import process records its started and completed states in an audit log. Although the process has completed, the cleanup log shows only `started`. You are using `defer` to ensure that the log is written.

You came across the following `defer` code. Let’s investigate what it does.

When you [run it in the Go Playground](https://go.dev/play/p/XEEZ1ax4se2), the two log entries show different values and appear in reverse order.

```go
package main

import "fmt"

func main() {
	status := "started"
	defer fmt.Println("direct:", status)
	defer func() {
		fmt.Println("closure:", status)
	}()

	status = "completed"
}
```

This is the output with Go 1.26.4.

```text
closure: completed
direct: started
```

---

## Question 1: When are the arguments of a deferred function call evaluated?

The assignment of `status` to `completed` comes after the first `defer fmt.Println("direct:", status)`. Even so, explain why `direct` is `started`, distinguishing the evaluation time of the `defer` statement from the execution time of the call.

<details>
<summary>Hint</summary>

Find `defer` statements in the Go language specification. Pay attention to the separate descriptions of the function call itself and the evaluation of the function value and arguments.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open [Defer statements in the Go language specification](https://go.dev/ref/spec#Defer_statements) and confirm what is evaluated by a `defer` statement and when the function call is executed.
2. Compare the evaluation timing with the small example in [Go Blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover).

**Answer**

The function value and arguments written in a `defer` statement are evaluated when that `defer` statement executes. The `status` in `fmt.Println("direct:", status)` is `started` at that point. The function call itself is delayed until `main` returns, but the argument values have already been saved, so it prints `direct: started`.

</details>

---

## Question 2: Why does `closure` show the completed value and print first?

The anonymous function prints `completed`, and it prints before `direct`. Investigate when the anonymous function reads `status` and the execution order of multiple `defer` calls, then explain why the opening audit log recorded the old state.

<details>
<summary>Hint</summary>

Compare the rule about “function values and arguments” from Question 1 with an anonymous function whose body refers to a variable. Then confirm the order of multiple `defer` calls in the same specification section.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Confirm in [Defer statements in the Go language specification](https://go.dev/ref/spec#Defer_statements) that multiple deferred calls execute in reverse order.
2. Compare the delayed-execution example in [Go Blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) with code that reads `status` in the anonymous function body.
3. Compare the output order and values with the [execution example](https://go.dev/play/p/XEEZ1ax4se2).

**Answer**

Because `status` is not passed as an argument to the anonymous function, its body reads the variable when the deferred call executes. The assignment has happened by then, so the value is `completed`. Deferred calls also execute in reverse order from the order in which they were registered, so the anonymous function registered later prints first.

In short, the opening `direct` log is old not because it was printed during cleanup, but because its argument was evaluated and saved when the `defer` statement ran. For an audit log that must record the final state, decide when the value should be fixed and choose an argument or an anonymous function accordingly.

</details>

---

## Starting points for investigation

1. Open [How to investigate 04-deep-dive](../../README.md) and begin with the Go language specification.
2. Use [Defer statements in the Go language specification](https://go.dev/ref/spec#Defer_statements) to investigate evaluation timing and execution order.
3. Compare the observation with [Go Blog: Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover).
