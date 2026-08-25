[Back to the 03-cmd-tools research guide](../../README.md)

# Track Down the 600 ms Batch: `go tool pprof`

A customer-facing batch issues codes every night. CI shows that the code-issuing test alone takes about 600 ms. Some people suggest reducing the loop count, but the seat codes are already checked against another system. To make it faster without changing output, first identify where the CPU is actually being used.

**`main.go`**

```go
package main

import "fmt"

var result uint64

func seatCode(seed uint64) uint64 {
	for range 4_000_000 {
		seed = seed*2862933555777941757 + 3037000493
	}
	return seed
}

func main() {
	fmt.Println(seatCode(1))
}
```

([Run it in the Go Playground](https://go.dev/play/p/5Qwd1dKuD3I))

```console
$ go run main.go
15662720274509501185
```

**`main_test.go`**

```go
package main

import "testing"

func TestSeatCode(t *testing.T) {
	if got, want := seatCode(1), uint64(15662720274509501185); got != want {
		t.Fatalf("seatCode(1) = %d, want %d", got, want)
	}
}

func TestIssueCodes(t *testing.T) {
	for i := uint64(0); i < 100; i++ {
		result ^= seatCode(i)
	}
}

func BenchmarkSeatCode(b *testing.B) {
	for b.Loop() {
		result = seatCode(1)
	}
}
```

The following was run with Go 1.26.4 on macOS (Apple M1 Max):

```console
$ go test -run '^TestIssueCodes$' -cpuprofile=cpu.out
PASS
ok   	example.com/pprof-demo	1.190s

$ go tool pprof -top pprof-demo.test cpu.out
File: pprof-demo.test
Type: cpu
Duration: 612.42ms, Total samples = 410ms (66.95%)
Showing nodes accounting for 410ms, 100% of 410ms total
      flat  flat%   sum%        cum   cum%
     400ms 97.56% 97.56%      410ms   100%  example.com/pprof-demo.seatCode (inline)
     10ms  2.44%   100%       10ms  2.44%  runtime.asyncPreempt
         0     0%   100%      410ms   100%  example.com/pprof-demo.TestIssueCodes
         0     0%   100%      410ms   100%  testing.tRunner
```

Use `cpu.out` as the magnifying glass and `pprof-demo.test` as the map to find where the CPU time went.

---

## Question 1: What was collected, and what was not?

What does `go test -cpuprofile=cpu.out` create, and why does the test binary `pprof-demo.test` remain afterward? This problem uses a CPU profile. Explain why it should not be the first tool for a case suspected to involve network or lock waiting.

<details>
<summary>Hint</summary>

- Compare Profiling and Tracing in the [Go diagnostics guide](https://go.dev/doc/diagnostics).
- Check `-cpuprofile` and the note immediately after it in `go help testflag`.
- Run `go tool pprof -h` and inspect its accepted arguments.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Confirm in the [diagnostics guide](https://go.dev/doc/diagnostics) that CPU profiles locate places consuming CPU cycles.
2. Use `go help testflag` to confirm that `-cpuprofile cpu.out` writes a CPU profile before exit and that non-coverage profiles leave the test binary.
3. Read the [official pprof documentation](https://go.dev/cmd/pprof/) and `go tool pprof -h` to confirm the `<binary> <profile>` form. The implementation starts in the [Go 1.26.4 `cmd/pprof` source](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/pprof/doc.go).

**Answer**

`-cpuprofile=cpu.out` writes a CPU profile of the test run. `go tool pprof -top pprof-demo.test cpu.out` uses the test binary to map profile addresses back to function names and source locations. The binary remains for that mapping because `go help testflag` says that profile-generating flags other than coverage leave it behind. A CPU profile shows time spent consuming CPU; waiting on a lock or I/O does not become its main signal. Use an execution trace or a profile matching the suspected kind of waiting.

</details>

---

## Question 2: Descend from the function name to the line to change

`seatCode` accounts for more than 97% in `-top`, but a function name alone cannot justify a code change in review. Find the command that reports information by source line and identify the observed center. Then explain why “shorten the loop” cannot be merged immediately.

<details>
<summary>Hint</summary>

- `go tool pprof -h` lists an output format that displays source lines.
- Use the same binary and `cpu.out`.
- `TestSeatCode` provides a clue about compatibility.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Read the [pprof documentation](https://go.dev/cmd/pprof/) about function and source-line views.
2. Confirm that `-list` displays source associated with a function.
3. Run `go tool pprof -list='seatCode' pprof-demo.test cpu.out` and read flat and cumulative time.
4. Compare the result with `TestSeatCode` and the compatibility constraint.

**Answer**

A typical result centers on the loop body:

```console
$ go tool pprof -list='seatCode' pprof-demo.test cpu.out
Total: 410ms
ROUTINE ======================== example.com/pprof-demo.seatCode
     400ms      410ms (flat, cum)   100% of Total
         .          .      7:func seatCode(seed uint64) uint64 {
      40ms       50ms      8:\tfor range 4_000_000 {
     360ms      360ms      9:\t\tseed = seed*2862933555777941757 + 3037000493
```

`-top` narrows the candidate and `-list` connects it to source. However, merely reducing `4_000_000` changes the result. `TestSeatCode` fixes the output for `seed == 1`, so that change breaks the existing compatibility contract. A profile identifies where to investigate; it does not authorize discarding a specification.

</details>

---

## Question 3: How do you report “faster” without breaking compatibility?

Use the existing tests and benchmark appropriately. Run the following command and explain what should be recorded and compared:

```console
go test -run '^$' -bench '^BenchmarkSeatCode$' -benchmem -count=3
```

<details>
<summary>Hint</summary>

- Check `-bench`, `-benchmem`, and `-count` in `go help testflag`.
- Correctness tests and benchmarks have different jobs.
- Think about why the same measurement is repeated instead of relying on one number.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Distinguish profiles and benchmarks in the [diagnostics guide](https://go.dev/doc/diagnostics).
2. Confirm the three flags with `go help testflag`.
3. Record a baseline before changing the code, then run the same command afterward and run the regular tests too.

**Answer**

A baseline from this environment was:

```console
goos: darwin
goarch: arm64
pkg: example.com/pprof-demo
cpu: Apple M1 Max
BenchmarkSeatCode-10     235   5051027 ns/op       0 B/op       0 allocs/op
BenchmarkSeatCode-10     238   5030072 ns/op       0 B/op       0 allocs/op
BenchmarkSeatCode-10     237   5023913 ns/op       0 B/op       0 allocs/op
```

`TestSeatCode` protects output compatibility. `BenchmarkSeatCode` compares `ns/op`, `B/op`, and `allocs/op`; record all three runs, the Go version, OS, CPU, and command. pprof finds an improvement candidate, the benchmark measures its effect, and the unit test protects behavior.

</details>

---

<details>
<summary>Trivia</summary>

`go tool pprof -http=localhost:0 pprof-demo.test cpu.out` opens the Web UI on an available local port. First narrowing the hypothesis with `-top` and `-list` makes the investigation easier to share. For production profiles, also read the representative production-load guidance in the [PGO guide](https://go.dev/doc/pgo).

</details>

---

## Research starting points

1. [Go diagnostics guide](https://go.dev/doc/diagnostics)
2. [Official pprof documentation](https://go.dev/cmd/pprof/)
3. `go help testflag` and `go tool pprof -h`
4. [Go 1.26.4 `cmd/pprof` source](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/pprof/doc.go)
5. [PGO guide](https://go.dev/doc/pgo)
