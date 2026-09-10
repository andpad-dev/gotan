[Scenario index](../../../SCENARIOS.md) | [Workshop guide](../../../README.md) | [How to research 04-deep-dive](../../README.md)

# Can You Trust a Go Compiler Built from Clean Source?

**Execution environment**: No code is run.

Investigate whether a distributed Go toolchain matches its source. Test the hypothesis that a clean source tree and an executable containing a version are sufficient.

```console
$ GOTOOLCHAIN=go1.27.0 go version -m "$(GOTOOLCHAIN=go1.27.0 go env GOTOOLDIR)/compile"
<GOTOOLDIR>/compile: go1.27.0
	path	cmd/compile
	build	-buildmode=exe
	build	-compiler=gc
	build	-gcflags=cmd/...=-dwarf=false
	build	-pgo=default.pgo
	build	-trimpath=true
	build	CGO_ENABLED=0
	build	GOARCH=arm64
	build	GOARM64=v8.0
	build	GOOS=darwin
```

Follow past compiler attacks through the current Go implementation, then examine Go 1.27 bootstrapping and reproducible builds.

<details><summary>Investigation entry points</summary>

Start with the [04-deep-dive research guide](../../README.md), then choose one of these primary sources:

- [Go Documentation](https://go.dev/doc) - the `go version -m` command
- [Go compiler](https://go.dev/cmd/compile/) - the `cmd/compile` documentation
- [Installing Go from source](https://go.dev/doc/install/source#go14) - bootstrap requirements
- [Perfectly Reproducible, Verified Go Toolchains](https://go.dev/blog/rebuild) - reproducible builds
- [Go Reproducible Build Report](https://go.dev/rebuild) - public rebuild results
- [Running the “Reflections on Trusting Trust” Compiler](https://research.swtch.com/nih) - the compiler-attack demonstration

</details>

---

## Question 1: What can `go version -m` confirm?

Separate what the command reads and displays from what it cannot establish.

<details><summary>Hint</summary>

Use `go help version` to identify what `-m` reads. Distinguish reading embedded build information from executing the binary, rebuilding it from source, or checking a signature or published hash.

</details>

<details><summary>Answer</summary>

`go version -m` does not execute the target binary. It reads embedded module and build information through `debug/buildinfo.ReadFile`, including the Go version, package path, and build settings. It does not rebuild the binary from public source, compare it with a published hash, or verify a signature. A clean source tree and a `go1.27.0` string therefore do not prove that the binary was built from that source without additional changes.

</details>

---

## Question 2: Why can changed behavior survive after the source is restored?

Read Russ Cox's [Running the “Reflections on Trusting Trust” Compiler](https://research.swtch.com/nih) without executing the malicious code. Compare the normal program target with the compiler's own source.

<details><summary>Hint</summary>

Find the boundary immediately before the compiler parses source. Compare a rewrite targeting an application with a rewrite that recognizes and reinjects the compiler's own logic.

</details>

<details><summary>Answer</summary>

A malicious compiler can rewrite a characteristic application such as `hello, world`. It can also recognize the compiler's parser source and inject the same rewrite logic, together with the data needed to regenerate it. When that compiler builds a clean compiler source tree, the resulting compiler contains the logic even though the source on disk is unchanged, so `git diff` is empty and the behavior survives the next rebuild. Go 1.27.0 still has a source-input boundary before parsing: `syntax.Parse` receives an `io.Reader` and passes it to `p.init`. This confirms the relevant boundary exists, not that the official toolchain contains the attack.

</details>

---

## Question 3: Which compiler builds Go 1.27?

Find the bootstrap version and the generation order of `toolchain1`, `go_bootstrap`, `toolchain2`, and `toolchain3`.

<details><summary>Hint</summary>

Read the source-installation rule for the bootstrap version, then search the Go 1.27.0 `cmd/dist` source for all four names. Ask whether starting from one older compiler independently rules out a trust attack.

</details>

<details><summary>Answer</summary>

The source-installation rule gives Go 1.24 as the bootstrap series for Go 1.27, and the Go 1.27.0 source pins `minBootstrap` to `go1.24.6`. In broad terms, the build first uses the old toolchain to build `toolchain1`, then builds `go_bootstrap`; those tools build `toolchain2`, and the final stage uses `go_bootstrap` and `toolchain2` to build `toolchain3` from the same new source. The process can be traced in [`cmd/dist/buildtool.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/dist/buildtool.go) and [`build.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/dist/build.go).

Staged bootstrapping makes it possible to begin from an older released toolchain and, further back, the Go 1.4 C implementation. One build from a Go 1.24 binary does not independently establish the provenance of that binary; independent rebuilds and output comparisons are still needed.

</details>

---

## Question 4: What does a `PASS` from a reproducible rebuild prove?

Read the [Go Reproducible Build Report](https://go.dev/rebuild) and separate its valid conclusion from the claim that all source and local compilers are absolutely safe.

<details><summary>Hint</summary>

Separate reproducible output from source safety. Check what `gorebuild` compares, how Linux, macOS, and Windows distributions are handled, and whether the report's per-target results mean the same thing as its exit status.

</details>

<details><summary>Answer</summary>

`gorebuild` rebuilds Go toolchain distribution files from source and compares them with the public distributions at `go.dev/dl`. Linux/amd64 can use a complete bootstrap from the Go 1.4 C implementation. Most distributions are compared bit-for-bit; macOS executables exclude code signatures, PKG contents are compared as tar.gz, and Windows MSI handling may compare extracted contents with a corresponding zip or skip the MSI when extraction is unavailable.

A matching independent rebuild is strong evidence that the public distribution corresponds to the source without an extra source-external change. It does not prove that the source has no vulnerability, that every build environment is independent, or that a package-manager binary on a local machine is identical to the public distribution. Check the generated JSON or HTML report rather than treating exit status 0 alone as proof that every target matched.

</details>

---

