[Workshop guide and scenario index](../../../README.md) | [How to research 04-deep-dive](../../README.md)

# Why Does Go 1.27's Faster Memory Allocation Stop at 80 Bytes?

The Go 1.27 release notes say: “The compiler now generates calls to size-specialized memory allocation routines, reducing the cost of some small (<80 byte) memory allocations by up to 30%.” Why 80 bytes rather than 128 or 256?

The release-note summary says `<80 byte`, but the Go 1.27.0 implementation rejects only `size > 80`; an allocation of exactly 80 bytes is therefore eligible. This scenario uses “80 bytes or less” to match the implementation.

```go
type Point struct{ X, Y int } // 16 bytes = no more than 80 bytes

p := new(Point) // The size is known and may use a size-specialized routine
```

([Run it in the Go Playground](https://go.dev/play/p/YnvY6u0c87E))

Output with Go 1.27.0:

```text
{1 2}, size=16 bytes
```

---

## Question 1: What does “size-specialized” specialize?

<details><summary>Hint</summary>The release-note wording says the compiler generates calls. Search `go-review.googlesource.com` for “size specialized malloc” and inspect `src/runtime/_mkmalloc`.</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read [Go 1.27 release notes](https://go.dev/doc/go1.27).
2. Read [the compiler change](https://go-review.googlesource.com/c/go/+/707856).
3. Inspect [`specializedMallocSym`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/ssagen/ssa.go;l=804) and the generated [mkmalloc implementation](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/runtime/_mkmalloc/mkmalloc.go).

**Answer**

Previously, `new(T)` used generic `runtime.newobject` and `mallocgc`, which repeatedly considered the size, pointer scanning, and header requirements. Go 1.27 can directly call generated functions when allocation size is known at compile time: `mallocgcTiny...`, `mallocgcSmallNoScanSC<n>`, or `mallocgcSmallScanNoHeaderSC<n>`. Runtime-sized allocations such as `make([]T, n)` cannot select one statically and keep the generic path.

</details>

---

## Question 2: Why is the limit 80 bytes?

<details><summary>Hint</summary>Find `80` in `specializedMallocSym`. Its comment points to a matching constant and explanation in mkmalloc.</details>

<details><summary>Answer</summary>

**Investigation route**

1. Start with the `<80 byte` summary in the [Go 1.27 release notes](https://go.dev/doc/go1.27), then ask how the implementation tests the boundary.
2. Read [`specializedMallocSym`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/ssagen/ssa.go;l=804), including the `size > specializedMallocMax` condition.
3. Read [`specializedMallocMax`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/runtime/_mkmalloc/constants.go;l=25).

**Answer**

The runtime comment says larger size classes provide very limited benefit and can be slower because of the wrapper. Larger allocations are dominated by costs such as zeroing memory, so removing generic branches helps little. The generated functions also increase binaries by about 60 KB. `specializedMallocMax = 80` must match in compiler and runtime: it is the boundary where the optimization has demonstrated value, not the largest size that could technically be generated.

</details>

---

## Question 3: Why is the opt-out temporary?

<details><summary>Hint</summary>Read the rest of the release-note paragraph, including “regressions” and “expected to be removed in Go 1.28.”</details>

<details><summary>Answer</summary>

**Investigation route**

1. Read the Faster Memory Allocation section of the [Go 1.27 release notes](https://go.dev/doc/go1.27).

> Please file an issue if you notice any regressions. You may set `GOEXPERIMENT=nosizespecializedmalloc` at build time to disable it. This opt-out setting is expected to be removed in Go 1.28.

**Answer**

`GOEXPERIMENT=nosizespecializedmalloc` is a temporary escape hatch for regressions in a new code-generation optimization. It lets affected workloads return to the generic behavior while the optimization is evaluated. The tradeoff is up to 30% for affected allocations, roughly 1% for allocation-heavy programs overall, and about 60 KB of binary size. Because this is an experiment, the opt-out is expected to disappear once Go 1.27 experience shows that the default implementation is ready to become ordinary functionality.

</details>

---

<details><summary>Trivia</summary>

The first runtime CL, [specialized malloc functions up to 512 bytes](https://go.googlesource.com/go/+show/411c250d64304033181c46413a6e9381e8fe9b82), considered 512 bytes: a technical and microbenchmark boundary. The final 80-byte cutoff is the smaller “worth doing” boundary.

</details>

---

## Research starting points

- [Go 1.27 release notes](https://go.dev/doc/go1.27)
- [Compiler CL](https://go-review.googlesource.com/c/go/+/707856)
- [`runtime/_mkmalloc/constants.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/runtime/_mkmalloc/constants.go;l=25)
- [`cmd/compile/internal/ssagen/ssa.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/compile/internal/ssagen/ssa.go;l=804)
