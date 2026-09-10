[Scenario index](../../../SCENARIOS_en.md) | [Workshop guide](../../../README.md) | [How to research 02-features](../../README.md)

# Why Does Go 1.27's `goroutineleak` Profile Report Only Goroutines That Can Never Run Again?

**Execution environment**: Browser only. No Go installation is required.

The number of goroutines in `/debug/pprof/goroutine` is slowly increasing. The stacks show goroutines blocked at `chan send`. The existing `goroutine` profile lists every goroutine that currently exists, so it cannot distinguish a goroutine that can never wake up from one that is simply long-lived.

Go 1.27 adds `/debug/pprof/goroutineleak`. It extracts the same leak pattern and lets you compare both profiles.

```go
package main

import (
    "errors"
    "fmt"
    "os"
    "runtime"
    "runtime/pprof"
    "time"
)

type result struct { value int; err error }

func processWorkItem(id int) (int, error) {
    if id == 2 { return 0, errors.New("boom") }
    time.Sleep(100 * time.Millisecond)
    return id * 10, nil
}

func processWorkItems(ids []int) ([]int, error) {
    ch := make(chan result)
    for _, id := range ids {
        go func(id int) {
            v, err := processWorkItem(id)
            ch <- result{v, err}
        }(id)
    }
    var out []int
    for range ids {
        r := <-ch
        if r.err != nil { return nil, r.err }
        out = append(out, r.value)
    }
    return out, nil
}

func main() {
    fmt.Println("go version:", runtime.Version())
    _, err := processWorkItems([]int{0, 1, 2, 3, 4})
    fmt.Fprintln(os.Stderr, "processWorkItems err:", err)
    time.Sleep(300 * time.Millisecond)
    fmt.Println("=== goroutine profile ===")
    pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
    fmt.Println("=== goroutineleak profile ===")
    pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
}
```

Playground: [Shared code that prints the Go version](https://go.dev/play/p/UwYB3wyRxe9)

This scenario covers a Go 1.27 feature, so the Playground is recommended. Locally, use Go 1.27. Addresses and absolute paths vary by environment.

With Go 1.27.0, the output has the following relevant shape:

```text
go version: go1.27.0
processWorkItems err: boom
=== goroutine profile ===
goroutine profile: total 5
4 @ ... main.processWorkItems.func1 ...
1 @ ... main.main ...
=== goroutineleak profile ===
goroutineleak profile: total 4
4 @ ... main.processWorkItems.func1 ...
```

The addresses and source paths vary by environment. The ordinary profile contains the four blocked workers and the running `main` goroutine; the leak profile contains only the four workers.

<details><summary>Investigation entry points</summary>

Start with the [02-features research guide](../../README.md), then choose one of these primary sources:

- [Go 1.27 release notes](https://go.dev/doc/go1.27) - the new `goroutineleak` profile
- [Go 1.26 release notes](https://go.dev/doc/go1.26) - the experimental version
- [`runtime/pprof`](https://pkg.go.dev/runtime/pprof) - reserved profile names
- [`net/http/pprof`](https://pkg.go.dev/net/http/pprof) - the `/debug/pprof/` endpoints
- [Proposal issue #74609](https://go.dev/issue/74609) and its [design document](https://go.googlesource.com/proposal/+/master/design/74609-goroutine-leak-detection-gc.md)

</details>

---

## Question 1: How do the `goroutine` and `goroutineleak` profiles differ?

Use primary sources to verify what each profile collects.

<details><summary>Hint</summary>

Check the Go 1.27 release notes and the reserved profile names in the `runtime/pprof` documentation. Compare the one-line descriptions of `goroutine` and `goroutineleak`, then find the release note's definition of a leaked goroutine.

</details>

<details><summary>Answer</summary>

**Investigation path**

1. Read the `Goroutine leak profile` section in the [Go 1.27 release notes](https://go.dev/doc/go1.27).
2. Compare the reserved names in the [`runtime/pprof` documentation](https://pkg.go.dev/runtime/pprof).
3. Confirm the definition in the same release-note section.

**Answer**

`goroutine` reports stack traces for all current goroutines. `goroutineleak` reports stack traces for leaked goroutines only. A leaked goroutine is blocked on a concurrency primitive such as a channel, `sync.Mutex`, or `sync.Cond` and cannot possibly become unblocked. In the example, four workers are blocked sending to the abandoned channel; the running `main` goroutine is the fifth entry in the ordinary profile and is not a leak.

</details>

---

## Question 2: How does the runtime determine that a goroutine can never wake up?

The blocked state is visible from the runtime's wait reason. The difficult part is proving that the future cannot unblock it.

<details><summary>Hint</summary>

Follow the Go 1.27 release note to the Go 1.26 experimental feature, proposal issue [#74609](https://go.dev/issue/74609), and its [design document](https://go.googlesource.com/proposal/+/master/design/74609-goroutine-leak-detection-gc.md). Read the numbered steps in `Proposal`.

</details>

<details><summary>Answer</summary>

**Investigation path**

1. Read the current feature in the [Go 1.27 release notes](https://go.dev/doc/go1.27).
2. Read the experimental version in the [Go 1.26 release notes](https://go.dev/doc/go1.26).
3. Follow [issue #74609](https://go.dev/issue/74609) to the [design document](https://go.googlesource.com/proposal/+/master/design/74609-goroutine-leak-detection-gc.md).

**Answer**

The runtime performs a dedicated goroutine-leak-detection GC cycle. Runnable goroutines are initial roots. After reachable memory is marked, a blocked goroutine waiting on a marked synchronization primitive is promoted to a root because it may become runnable. The process repeats to a fixed point. Goroutines that were not promoted are reported as leaks; they are then added as roots for the final mark so their reachable memory is retained.

In the example, the channel becomes unreachable after the early return. The four workers wait to send through it, and no reachable synchronization primitive can wake them. They are therefore reported by `goroutineleak`.

</details>

---

## Question 3: Why are leaks through global variables or runnable goroutine locals not detected?

The release notes warn that reachability-based detection can miss leaks caused by primitives reachable through globals or runnable goroutines.

<details><summary>Hint</summary>

Read the `Rationale` section of the design document and focus on the formal property described there. Consider which error a production diagnostic tool should avoid: false positives or false negatives.

</details>

<details><summary>Answer</summary>

**Investigation path**

1. Read the limitation in the [Go 1.27 release notes](https://go.dev/doc/go1.27).
2. Read `Rationale` in the [proposal design document](https://go.googlesource.com/proposal/+/master/design/74609-goroutine-leak-detection-gc.md).

**Answer**

The design preserves **soundness** at the cost of completeness. If a channel is reachable from a global variable or a runnable goroutine, the runtime cannot prove that no future code will use it. Reporting such a goroutine as a leak could create a false positive. The profile therefore reports only the cases it can prove: a leak report has no false positives, while some real leaks may remain undetected.

</details>

---

