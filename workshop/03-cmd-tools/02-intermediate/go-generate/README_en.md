[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 03-cmd-tools](../../README.md)

# Investigate What `go generate` Can Do and What It Should Do

We want to automate code generation with `go generate`. `//go:generate` is more flexible than a fixed format limited to code-generation commands. The sample in [`main.go`](./main.go) demonstrates this.

```go
package main

//go:generate echo "hello from go generate"
//go:generate go env GOOS GOARCH
//go:generate sh -c "echo GOFILE=$GOFILE GOLINE=$GOLINE GOPACKAGE=$GOPACKAGE"
//go:generate -command say echo
//go:generate say "An alias defined with -command also works"

func main() {}
```

([Run it in the Go Playground](https://go.dev/play/p/o36nNAzX45u). The Playground does not run `go generate` itself.)

With Go 1.27.0 on macOS arm64, `go generate main.go` produced the following. `GOOS` and `GOARCH` vary by environment.

```console
$ go generate main.go
hello from go generate
darwin
arm64
GOFILE=main.go GOLINE=5 GOPACKAGE=main
An alias defined with -command also works
```

`echo` and `go env` are not code-generation tools, but they run normally. How far does this flexibility go, and where does it come from? Let us investigate the implementation of `go generate`.

---

## Question 1: What commands can `//go:generate` contain?

Run the sample and try explicit, non-destructive commands such as `go version`, `go env GOOS GOARCH`, or `echo sample`. Do not use commands such as `env` that dump all environment variables, and do not display credentials or tokens. Before sharing output in an issue or chat, confirm that it contains no secrets. What do `go generate -n main.go` and `go generate -x main.go` display, and where are variables such as `$GOFILE` and `$GOLINE` listed?

<details>
<summary>Hint</summary>

- Begin with the [Go command documentation for generate](https://go.dev/cmd/go/#hdr-Generate_Go_files_by_processing_source), then compare it with `go help generate` locally.
- Remember that `-n` means “no run” and `-x` means “execute,” a naming convention shared with commands such as `go build -n`.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Open the [Go command documentation for generate](https://go.dev/cmd/go/#hdr-Generate_Go_files_by_processing_source), then read the Usage and variables sections alongside `go help generate`.
2. Run the sample with and without `-n` and `-x`, then compare the output.

**Answer**

The help says: “Generate runs commands described by directives within existing files. Those commands can run any process but the intent is to create or update Go source files.” Thus an executable command can be anything; code generation is the intended use, not an execution restriction. `-n` prints commands without running them, while `-x` runs them and also prints each command. `$GOFILE`, `$GOLINE`, `$GOPACKAGE`, `$GOARCH`, `$GOOS`, `$GOROOT`, `$DOLLAR`, and `$PATH` are listed in the “Go generate sets several variables” section of `go help generate`.

</details>

---

## Question 2: Where does command execution happen?

Find the source location that launches the child process and identify the standard package and function used to do it.

<details>
<summary>Hint</summary>

- The `go` command source is under `cmd/go`, split into `cmd/go/internal/<subcommand>` packages.
- Read [Go 1.27.0 generate.go](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/generate/generate.go).
- There is one standard package designed to launch child processes.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Follow Source Files from the [Go command documentation](https://go.dev/cmd/go/) to the `cmd/go` implementation.
2. Open [Go 1.27.0 generate.go](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/generate/generate.go), search for `exec`, and inspect the command-execution path.

**Answer**

The `(*Generator).exec` method at [generate.go#L487-L512](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/generate/generate.go;l=487-512) is the center of execution:

```go
func (g *Generator) exec(words []string) {
	path := words[0]
	if path != "" && !strings.Contains(path, string(os.PathSeparator)) {
		gorootBinPath, err := pathcache.LookPath(filepath.Join(cfg.GOROOTbin, path))
		if err == nil {
			path = gorootBinPath
		}
	}
	cmd := exec.Command(path, words[1:]...)
	cmd.Args[0] = words[0]

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = g.dir
	cmd.Env = str.StringList(cfg.OrigEnv, g.env)
	err := cmd.Run()
	if err != nil {
		g.errorf("running %q: %s", words[0], err)
	}
}
```

The standard `os/exec` package supplies `exec.Command` and `cmd.Run()`. `go generate` splits a directive into words and starts it as a child process. There is no special sandbox or allow-list. `cmd.Dir = g.dir` runs it in the package source directory. `-n` skips the call and `-x` prints the words before calling it.

</details>

---

## Question 3: Why is `go generate` designed this way?

Read the official blog and proposal and investigate the Goal and Non-goal behind this deliberately simple mechanism.

<details>
<summary>Hint</summary>

- Read the official [Go generate blog post](https://go.dev/blog/generate) by Rob Pike.
- In the [proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/go-generate.md), read “Introduction,” “Discussion,” and “Non-goal.”

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Read the history in the [blog post](https://go.dev/blog/generate).
2. Read the relevant sections of the [proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/go-generate.md).

**Answer**

`go build` automates building Go programs but does not provide preliminary processing such as yacc, protobuf, Unicode-table, HTML-embedding, or binary-data generation. Added in Go 1.4 in December 2014, `go generate` gave package authors a way to run such generators.

The proposal's goals are that package authors, rather than clients, run it; generated `.go` files are committed so clients can simply use `go get` and `go build`; `go build` never invokes `go generate` automatically; and authors may use any generator they choose. Its explicit non-goal is a generalized build system like Unix `make`: it deliberately performs no dependency analysis and “does what is asked of it, nothing more.” The direct `exec.Command` implementation reflects that design.

</details>

---

<details>
<summary>Trivia</summary>

`go generate` suits deterministic, repeatable generation whose results are committed and run by package authors or CI. It is a poor fit for work required on every build, operations with side effects, dependency-aware regeneration, or anything that assumes the end user's environment contains the generator. It does not roll back failed side effects or automatically detect which inputs changed.

</details>

---

## Research starting points

- [Generate Go files by processing source](https://go.dev/cmd/go/#hdr-Generate_Go_files_by_processing_source)
- [Go generate blog post](https://go.dev/blog/generate)
- [go generate proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/go-generate.md)
- [Go 1.27.0 generate.go source](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/internal/generate/generate.go)
