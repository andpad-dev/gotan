[Back to the 03-cmd-tools research guide](../../README.md)

# Track Down the 100 ms Order Process: `go tool trace`

A test for an order-processing service runs four validation operations that each take 25 ms. They should run concurrently, yet completion takes about 100 ms. CPU usage is low, and a CPU profile does not make the cause of the wait clear. The operations are also grouped into one `order-processing` task. Let us investigate why execution tracing is shaped this way and build a procedure for the next incident.

**`main.go`**

```go
package main

import (
	"context"
	"fmt"
	"runtime/trace"
	"sync"
	"time"
)

func processOrders(ctx context.Context) {
	var orderLock sync.Mutex
	var wg sync.WaitGroup

	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			trace.WithRegion(ctx, "validate-order", func() {
				orderLock.Lock()
				defer orderLock.Unlock()
				time.Sleep(25 * time.Millisecond)
			})
		}()
	}
	wg.Wait()
}

func main() {
	processOrders(context.Background())
	fmt.Println("orders processed")
}
```

([Run it in the Go Playground](https://go.dev/play/p/8Afw0ygPRUL))

```console
$ go run main.go
orders processed
```

**`main_test.go`**

```go
package main

import (
	"context"
	"runtime/trace"
	"testing"
)

func TestProcessOrders(t *testing.T) {
	ctx, task := trace.NewTask(context.Background(), "order-processing")
	defer task.End()

	processOrders(ctx)
}
```

With Go 1.26.4 on macOS:

```console
$ go test -run '^TestProcessOrders$' -trace=order.trace
PASS
ok   	example.com/trace-demo	0.469s
```

---

## Question 1: Why collect a trace when the CPU appears idle?

Four operations are started with `go`, but the test takes about 100 ms. Explain from primary sources why a CPU profile alone is insufficient and what `go test -trace=order.trace` can observe.

<details>
<summary>Hint</summary>

- Compare Profiling and the Execution tracer in the [Go diagnostics guide](https://go.dev/doc/diagnostics).
- Read the `-trace` entry in `go help testflag`.
- See the trace-file generation paths in the [official trace documentation](https://go.dev/cmd/trace/).

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Use the [diagnostics guide](https://go.dev/doc/diagnostics) to distinguish CPU profiles from execution traces.
2. Use `go help testflag` to confirm that `-trace trace.out` writes an execution trace before the test exits.
3. Read the [trace documentation](https://go.dev/cmd/trace/) and the [Go 1.26.4 `cmd/trace` source](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/trace/doc.go).

**Answer**

A CPU profile is good at finding expensive CPU paths, but time spent waiting for a lock may not be prominent in it. An execution trace records runtime events over time, including goroutine creation, blocking and unblocking, scheduling, syscalls, GC, and heap size. `go test -trace=order.trace` records those events while the reproducible test runs, allowing us to see when the four goroutines ran, stopped, and became serialized instead of assuming the lock was the cause from the duration alone.

</details>

---

## Question 2: Turn a trace into evidence of waiting

Before opening `order.trace` in a browser, extract synchronization waiting in pprof format. Which commands should you run? Identify the line where the wait occurs and explain how `trace.NewTask` and `trace.WithRegion` add useful labels.

<details>
<summary>Hint</summary>

- The [trace documentation](https://go.dev/cmd/trace/) lists several pprof-like profile types.
- Find what the `sync` profile type represents.
- Continue with `go tool pprof -h` and `go doc runtime/trace`.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Confirm that `sync` is a synchronization blocking profile and use `go tool trace -pprof=sync`.
2. Run:

    ```console
    go tool trace -pprof=sync order.trace > sync.pprof
    go tool pprof -top sync.pprof
    go tool pprof -list='processOrders.func1.1' sync.pprof
    ```

3. Read `go doc runtime/trace` and the [Go 1.26.4 runtime/trace source](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/runtime/trace/annotation.go).

**Answer**

A representative profile is:

```console
$ go tool pprof -top sync.pprof
Type: delay
Showing nodes accounting for 466.64ms, 100% of 466.64ms total
      flat  flat%   sum%        cum   cum%
  205.63ms 44.07% 44.07%   205.63ms 44.07%  runtime.chanrecv1
  156.72ms 33.58% 77.65%   156.72ms 33.58%  sync.(*Mutex).Lock
  104.29ms 22.35%   100%   104.29ms 22.35%  sync.(*WaitGroup).Wait
```

Then `-list` connects the wait to the lock line:

```console
$ go tool pprof -list='processOrders.func1.1' sync.pprof
ROUTINE ======================== example.com/trace-demo.processOrders.func1.1
         .   156.72ms     19: trace.WithRegion(ctx, "validate-order", func() {
         .   156.72ms     20:  orderLock.Lock()
         .          .     21:  defer orderLock.Unlock()
```

The mutex wait and the nearly serial sum of four 25 ms operations identify the shared lock as the likely cause. `NewTask` attaches `order-processing` to the context, and `WithRegion` labels each goroutine's `validate-order` interval. These labels let the trace relate runtime waiting to logical work and guide the next design choice: split the lock or move validation outside it.

</details>

---

## Question 3: Why is starting a trace after the slowdown too late?

The operations team proposes starting a trace after detecting a slow request, but the relevant wait has already happened. Trace how the execution tracer changed in Go 1.21 and Go 1.22, and explain how flight recording addresses this. Also explain why a streamable trace does not mean `go tool trace` never loads a large trace into memory.

<details>
<summary>Hint</summary>

- Search for `runtime/trace` in the [Go 1.21 release notes](https://go.dev/doc/go1.21).
- Read Trace and `runtime/trace` in the [Go 1.22 release notes](https://go.dev/doc/go1.22).
- Follow “partition” into the [execution tracer overhaul design](https://go.googlesource.com/proposal/+/refs/heads/master/design/60773-execution-tracer-overhaul.md) and [flight recording issue #63185](https://github.com/golang/go/issues/63185).

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Confirm the lower tracing cost in the [Go 1.21 release notes](https://go.dev/doc/go1.21).
2. Read the partition and runtime changes in the [Go 1.22 release notes](https://go.dev/doc/go1.22).
3. Read the Background and Goals of the [overhaul design](https://go.googlesource.com/proposal/+/refs/heads/master/design/60773-execution-tracer-overhaul.md).
4. Read [Issue #63185](https://github.com/golang/go/issues/63185) and the official [execution traces article](https://go.dev/blog/execution-traces-2024).

**Answer**

Starting after detection cannot record the earlier lock contention. Go 1.21 reduced tracing cost on amd64 and arm64, and Go 1.22 rebuilt the implementation around self-contained partitions, reducing start and stop impact and enabling streaming-oriented processing. Flight recording keeps a recent moving window of partitions and snapshots it when an anomaly is detected, preserving evidence from immediately before detection. However, a streamable trace format does not mean `go tool trace` itself never loads a large trace into memory; the official article says the tool still loads the entire trace. Operations therefore need deliberate limits for the recording window, file size, and analysis environment.

</details>

---

## Question 4: Decide who can see the trace viewer

A trace can contain goroutine names, task names, and source locations. Which `-http` value should a developer use on a laptop to avoid exposing it unintentionally? Confirm the Go 1.27 change to `go tool trace -http=:6060` and the explicit all-addresses form from primary sources. The command measurements above use Go 1.26.4; verify the listen-address change in the release notes.

<details>
<summary>Hint</summary>

- Find Trace in the [Go 1.27 release notes](https://go.dev/doc/go1.27).
- Compare a port-only value with one that includes an address.
- Decide based on what the trace may contain and its intended audience.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Read the Trace section of the [Go 1.27 release notes](https://go.dev/doc/go1.27).
2. Run `go tool trace -h` to confirm `-http=addr`.
3. Use `go tool trace -http=localhost:0 order.trace` for a local-only viewer on an available port.

**Answer**

In Go 1.27, the port-only form `-http=:6060` is restricted to localhost, matching the safer behavior of `go tool pprof -http`. To listen on all addresses, specify one explicitly, such as `-http=0.0.0.0:6060`. For ordinary laptop investigation, use `-http=localhost:0`; trace labels such as `order-processing` and `validate-order` may reveal operational information, so being able to display a trace is not the same as being permitted to publish it.

</details>

---

<details>
<summary>Trivia</summary>

`go tool trace` can extract `net`, `sync`, `syscall`, and `sched` pprof-like profiles. Start with `sync` for lock waiting, `sched` for scheduler delay, and `net` for network waiting. Choose the profile that matches the question instead of mixing CPU profiling and tracing without a hypothesis.

</details>

---

## Research starting points

1. [Go diagnostics guide](https://go.dev/doc/diagnostics)
2. [Official trace documentation](https://go.dev/cmd/trace/)
3. `go help testflag`, `go tool trace -h`, `go tool pprof -h`, and [Go 1.26.4 `cmd/trace` source](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/trace/doc.go)
4. [Go 1.21 release notes](https://go.dev/doc/go1.21) and [Go 1.22 release notes](https://go.dev/doc/go1.22)
5. [Execution tracer overhaul design](https://go.googlesource.com/proposal/+/refs/heads/master/design/60773-execution-tracer-overhaul.md) and [Issue #63185](https://github.com/golang/go/issues/63185)
6. [Go 1.27 release notes](https://go.dev/doc/go1.27)
