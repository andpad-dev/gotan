[Scenario index (Japanese)](../SCENARIOS.md) | [Workshop guide (Japanese)](../README.md) | [Team guide (Japanese)](../TEAM_GUIDE.md)

[Scenario index (Japanese)](../SCENARIOS.md) | [Workshop guide (Japanese)](../README.md) | [Team guide (Japanese)](../TEAM_GUIDE.md)

# How to Explore 03-cmd-tools (Reverse Search Guide)

In this category, we investigate what the `go` command and its subcommands and tools (`go build`, `go run`, `go test`, `go vet`, `go generate`, `go doc`, `go tool trace`, `go tool pprof`, etc.) actually do by consulting official documentation and the cmd/go source code. When you get stuck, come back here first.

## 1. Learning What a Subcommand Does (Reading Official Help)

1. When researching on the web, start by opening [Go Command](https://go.dev/cmd/go/). This is the entry point to the official documentation for the entire `go` command.
2. From your terminal, run `go help <subcommand>` (examples: `go help run`, `go help generate`, `go help vet`). For subcommands included with `go` itself, this gives you an overview and major flags in one place.
3. To read the same content on the web, open [`pkg.go.dev/cmd/go`](https://pkg.go.dev/cmd/go) and press the `f` key to search for a subcommand name, then jump directly to its section (URLs ending with `#hdr-...`).
4. For tools provided as `go tool <subcommand>` (`vet`, `pprof`, `trace`, `cover`, etc.), use `go tool <subcommand> -h` or `go tool <subcommand> help` to check usage.

## 2. Learning Detailed Specifications and Flags

- Each individual command or tool has its own pkg.go.dev page: `https://pkg.go.dev/cmd/<subcommand>` (examples: [cmd/vet](https://pkg.go.dev/cmd/vet), [cmd/go](https://pkg.go.dev/cmd/go), [cmd/pprof](https://pkg.go.dev/cmd/pprof)).
- The **Overview** section at the top of the page summarizes what the command does and lists major flags. It often contains more information than the `-h` CLI output.

## 3. Inspecting What Actually Runs (Using Flags to Peek Inside)

- `-n` flag: Display only the commands that would be executed internally, without actually running them (dry-run).
- `-x` flag: Run the command as usual while displaying the executed commands sequentially.
- Available flags differ by subcommand, so check the options list in `go help <subcommand name>`.

## 4. Going Deep into Implementation (Verifying with Source Code)

1. The cmd/go implementation is organized under `src/cmd/go/internal/` at https://cs.opensource.google/go/go, with separate packages for each subcommand (for example: `go run` -> `internal/run`, `go build` -> `internal/work`, `go generate` -> `internal/generate`). For standalone commands (`vet`, `pprof`, etc.), open `https://cs.opensource.google/go/go/+/refs/tags/go1.26.5:src/cmd/<subcommand>/` directly or follow the source link at the end of the command page on pkg.go.dev.
2. Open the file with the same name as the subcommand (e.g., `run.go`) and look for the function registered with `Cmd*.Run` (e.g., `runRun`). This is your entry point for tracing the behavior. For standalone commands, the help text output by `go help` is usually written in `doc.go`, and reading that first gives you the big picture.
3. Use the `f` key to search the page while jumping one by one to the functions being called. Read the actual implementation of the called functions to confirm. Do not judge based only on comments or the feel of function names.
4. Align the version tag to your environment. Parts like `refs/tags/go1.26.5` in the URL should match the output of `go version`.

## 5. Learning When and Why a Feature Was Added

- If you know the target Go version, first check the release notes at `go.dev/doc/go1.<version>`. In particular, the **Tools** section ([Go 1.26](https://go.dev/doc/go1.26#tools) / [Go 1.27](https://go.dev/doc/go1.27#tools)) summarizes changes to the `go` command and tool suite.
- If the release notes include issue numbers or links to proposals, be sure to open the relevant issue and read the entire discussion and related links (Issues on `github.com/golang/go` and proposals at `go.googlesource.com/proposal`). Do not infer behavior from just the summary.

## 6. Verifying Behavior Locally

- Write sufficiently small code (just `package main` + `fmt.Println`-level) and run it with visualization flags such as `-n` (display only, no execution), `-x` (display executed commands), or `-work` to see the actual output.
- Cross-check the descriptions in documentation and source code with the execution results from your local environment.
