[Scenario index](../../../SCENARIOS_en.md) | [Workshop guide](../../../README.md) | [How to research 03-cmd-tools](../../README.md)

# From One Text File to a Test: cmd/go Script Tests

**Execution environment**: Go 1.27 is required locally.

Run tests in `GOROOT/src/cmd/go`. CLI integration tests often create temporary directories, write input files, and inspect output. The setup can become longer than the behavior being tested.

Go 1.27.0's [`run_hello.txt`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/testdata/script/run_hello.txt) keeps commands, expected results, and the required Go file together:

```text
env GO111MODULE=off

# hello world
go run hello.go
stderr 'hello world'

-- hello.go --
package main
func main() { println("hello world") }
```

The `hello.go` below the boundary can also run in the [Go Playground](https://go.dev/play/p/4Quk7tidxe8). Investigate how this one text file becomes a test, why the runner is not bash, and what can be reused in your own tests.

<details><summary>Investigation entry points</summary>

Start with the [03-cmd-tools research guide](../../README.md), then choose one of these primary sources:

- [Go command](https://go.dev/cmd/go/) - the command documentation and source links
- [Go 1.27.0 `run_hello.txt`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/testdata/script/run_hello.txt) and [`script_test.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/script_test.go;l=39) - the test data and runner
- [Go Testing By Example](https://research.swtch.com/testing) - compact test-data design
- [2018 script-test commit](https://github.com/golang/go/commit/5890e25b7ccb2d2249b2f8a02ef5dbc36047868b) - the original design change
- [`golang.org/x/tools/txtar`](https://pkg.go.dev/golang.org/x/tools/txtar@v0.47.0) and [`rsc.io/script`](https://pkg.go.dev/rsc.io/script@v0.0.2) - public related modules

</details>

---

## Question 1: How is one text file split and executed?

Why is the text above `-- hello.go --` interpreted as commands rather than written to a file? When and where is the text below the boundary created?

<details><summary>Hint</summary>

Start at `TestScript` in the Go command source. Follow the parser that returns the comment and files, then the code that creates the subtest and temporary working directory.

</details>

<details><summary>Answer</summary>

**Investigation path**

1. Read [Go 1.27.0 `TestScript`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/script_test.go;l=39).
2. Follow `ParseFile` to [Go 1.27.0 `internal/txtar`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/internal/txtar/archive.go;l=5).
3. Follow `NewState`, `ExtractFiles`, and `scripttest.Run`; compare them with [the script test README](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/go/testdata/script/README).

**Answer**

The file uses the `txtar` format. Text before the first `-- FILENAME --` is the archive comment; each section after a boundary is an archive file. `TestScript` enumerates `testdata/script/*.txt`, creates a `testing` subtest named from the file, creates a temporary state, parses the archive, and extracts its files into the working directory. Only the comment is passed to the script engine, which runs `go run hello.go` and checks the expected stderr.

Run one script with:

```console
$ cd "$(GOTOOLCHAIN=go1.27.0 go env GOROOT)/src/cmd/go"
$ GOCACHE=/private/tmp/gotan-go-script-cache GOTOOLCHAIN=go1.27.0 go test . -run='^TestScript/run_hello$' -count=1
ok   cmd/go  1.460s
```

</details>

---

## Question 2: Why is this neither bash nor an ordinary Go test?

The runner does not use the system shell. It interprets a small language made of registered commands and conditions, while `testing` handles file selection, subtests, isolation, and parallel execution.

<details><summary>Hint</summary>

Read the history of `script_test.go`, especially the 2018 change [`cmd/go: add new test script facility`](https://github.com/golang/go/commit/5890e25b7ccb2d2249b2f8a02ef5dbc36047868b), and compare it with the current [script engine](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/cmd/internal/script/engine.go;l=5).

</details>

<details><summary>Answer</summary>

**Answer**

Small shell scripts were easy to write but were difficult to select individually, slow, and unavailable on Windows. The earlier Go test framework supported selection, Windows, parallel execution, and isolation, but its setup and output checks made the test intent hard to read. The script facility combines the strengths: `testing` provides subtests and isolation, while a small shell-like language and `txtar` keep each test case compact. The language is not a security sandbox; registered commands may start real processes.

</details>

---

## Question 3: How much can you reuse in your own tests?

The Go 1.27.0 runner imports `internal/txtar` and `cmd/internal/script`. Can an unrelated module depend on them directly? Separate the public standard-library API, Go's internal implementation, and public external modules.

<details><summary>Hint</summary>

Read `go help importpath` and its `Internal packages` section. Separate archive parsing, command execution, and the connection to `testing`.

</details>

<details><summary>Answer</summary>

`testing.T.Run` and `T.Parallel` are public standard-library APIs. `internal/txtar` and `cmd/internal/script` are Go source-tree implementation details and cannot be imported by an unrelated module. For archive parsing, use a versioned public package such as [`golang.org/x/tools/txtar`](https://pkg.go.dev/golang.org/x/tools/txtar@v0.47.0). For script execution, [`rsc.io/script`](https://pkg.go.dev/rsc.io/script@v0.0.2) and `scripttest` are public options, but their README states that they are published copies without a promise of official support. Pin versions, test supported platforms, and do not treat the engine as a sandbox.

---

