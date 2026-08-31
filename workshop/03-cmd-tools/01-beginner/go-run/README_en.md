[Workshop guide and scenario index](../../../README.md) | [How to research 03-cmd-tools](../../README.md)

# Look behind `go run`

You came across `go run`. Let’s investigate what it does.

A senior colleague told you, “For a quick check, `go run main.go` is fine.” A colleague with Python and Ruby experience says, “`go run` is convenient, like running a script. So Go can run as an interpreter too.”

Go is supposed to be a compiled language, but is that understanding correct? Let’s investigate what actually happens when you run `go run`.

(Example to try locally: https://go.dev/play/p/_BqTu8xBI9h . The [main.go](./main.go) in this directory has the same contents, so you can run `go run` and use it for the investigation.)

## Question 1: Does `go run` interpret and execute source code like an interpreter?

As your colleague suggests, does `go run` behave like an “interpreter” that executes source code while interpreting it step by step? Or does it run the program in some other way?

<details>
<summary>Hint</summary>

- Read the description from `go help run` one sentence at a time and focus on the verbs. Does it use verbs other than “interpret” or “execute”?
- Run `go run -x main.go` for this [main.go](./main.go). `-x` prints the external commands actually invoked behind the scenes.
- Look at the output one line at a time and sort it into two groups:
  - Lines that appear to create something new, such as commands containing `compile` or `link`, and `mkdir` commands
  - Lines that appear to execute something already built, such as a line containing only a path
  - This classification provides evidence for deciding whether the command creates something and then runs it, or reads and executes the source code on the spot.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open https://go.dev/cmd/go/ and find the `go run` description in the official documentation for the `go` command.
2. Read `go help run` (= https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program). It says, “Run compiles and runs the named main Go package.” The verb is “compiles”. An interpreter would be expected to use a verb such as “interprets” or “evaluates”, so this is already evidence against that assumption.
3. Run `go run -x main.go` and sort the output into lines that create something and lines that execute something.

**Answer**

It is not an interpreter. When you run it with `-x`, the output looks like this (an excerpt from an actual `go run -x main.go` run).

```
WORK=/var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-buildXXXXXXXXXX
...
mkdir -p $WORK/b001/exe/
.../compile ... # compile the source
.../link -o $WORK/b001/exe/main ... # link the executable
$WORK/b001/exe/main # run the resulting executable
hello, gotan
```

It creates a temporary `$WORK` directory, compiles and links inside it to assemble an executable, and finally runs that executable.

There is no processing that reads `main.go` line by line and evaluates it. Instead, `go run` combines two steps into one command: build an executable as a temporary file, then run that temporary file.

The `$WORK` path and build ID vary by environment, so the output above abbreviates environment-dependent portions with `X` and `...`.

</details>

---

## Question 2: What happens to the built executable and `$WORK` directory after execution?

Does the temporary `$WORK` directory from Question 1 and the executable created there remain somewhere on the computer after execution finishes? Or are they removed?

<details>
<summary>Hint</summary>

- Keep track of the `WORK=/var/folders/.../go-buildXXXXXXXXXX` path on the first line of `go run -x main.go` output.
- After the command finishes, run `ls` or `find` on that path and check whether it still exists.
- Look through the `go help build` flags for one that mentions a temporary directory. Read its description and determine whether it adds behavior that does not happen by default or disables default behavior.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open https://go.dev/cmd/go/ and find the description of the `-work` flag in the `go build` documentation.
2. Run `go run -x main.go`, then run `ls` on the displayed `WORK=...` path after the command exits. The directory has been removed.
3. Read `go help build`. The description of `-work` says “print the name of the temporary work directory and do not delete it when exiting”. The phrase “delete it when exiting” reveals that deletion is the **default behavior**.

**Answer**

The `$WORK` directory is automatically deleted when execution finishes. `go run` compiles and links inside this directory and immediately runs the resulting executable, but its cleanup removes the entire directory, so it normally goes unnoticed.

Run with the `-work` flag to keep the directory instead, which lets you inspect it.

```
$ go run -work main.go
WORK=/var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-buildXXXXXXXXXX
hello, gotan
$ find /var/folders/xx/xxxxxxxxxxxxxxxxxxxxxxxx/T/go-buildXXXXXXXXXX -maxdepth 3
.../go-buildXXXXXXXXXX
.../go-buildXXXXXXXXXX/b001
.../go-buildXXXXXXXXXX/b001/importcfg
.../go-buildXXXXXXXXXX/b001/importcfg.link
.../go-buildXXXXXXXXXX/b001/exe
.../go-buildXXXXXXXXXX/b001/_pkg_.a
.../go-buildXXXXXXXXXX/b001/exe/main
```

`b001/exe/main` is the executable that was actually run. Without `-work`, the temporary directory and everything inside it are removed automatically.

</details>

---

## Question 3: Where are “creating the directory”, “running the build”, and “handling the built file” performed?

Questions 1 and 2 showed that `go run` follows this flow: 1. create a temporary directory, 2. build inside it, 3. run the resulting executable, and 4. clean up the entire directory at the end.
Where in the `go` command’s own source code are these three operations, creating the directory, running the build, and handling the built file, implemented?

**Prerequisite: Where are `go` subcommands located in the source code?**

The `go` command is one large program, but its implementation is divided by subcommand into packages such as `cmd/go/internal/<subcommand>`. For example, `go run` is implemented in `cmd/go/internal/run`, while `go build` is implemented in `cmd/go/internal/work`. Each package has a file named after the subcommand, such as `run.go`, and registers a function in a `Cmd*.Run` field, such as `CmdRun.Run = runRun`. That registered function is the entry point actually called when the subcommand runs (see the [`03-cmd-tools` category README](../../README.md) for details).

Therefore, to understand `go run`, begin by opening the entry point in the `cmd/go/internal/run` package.

<details>
<summary>Hint 1: Read run.go and identify what to investigate next</summary>

Open the entry-point function in the `cmd/go/internal/run` package (https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go) and scan the functions and types it calls from top to bottom.

- Distinguish whether each operation is defined in the `run` package itself or belongs to another package. The package name before a call is a clue.
- When you find another package, infer what it might be responsible for from its name. A package that sounds related to building or linking is the next place to read.
- Separating operations that appear to prepare something from operations that execute what was prepared helps reveal how the three questions, directory creation, build execution, and file handling, map onto the source.

</details>

<details>
<summary>Hint 2: How should you read the package you found?</summary>

The package found in Hint 1 is large enough to share implementation with `go build`, and its code is spread across multiple files. Instead of reading everything from top to bottom, narrow your search as follows.

- Use in-page search (`f`) for standard-library function names typically involved in creating or removing directories, such as `os.MkdirTemp` or `RemoveAll`.
- Find variable or field names that appear to represent the executable itself, then trace where their values are assigned and read.
- The code that actually invokes the compiler and linker is split into further functions. Searching for short function names containing the build verbs, compile and link, helps locate them.
- Rather than reading one function all the way through, get the broad shape of each blank-line-separated block: is it preparing something, or actually running something? This is the same reading approach used in the hint for Question 1.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open https://go.dev/cmd/go/ and confirm that `go run` is implemented as part of `cmd/go`.
2. Read the `runRun` function, from around line 73, in https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go.
3. Read the implementation of `work.NewBuilder`, from around line 281 in https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go, and locate where the temporary directory is created.
4. In `LinkAction`, from around line 922, and `CompileAction`, from around line 633, in the same file, check how each package’s subdirectory (`Objdir`) and executable path (`Target`) are determined.
5. Read `Builder.build`, from around line 723, which handles compilation, and `Builder.link`, from around line 1591, which handles linking, in https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/exec.go. Confirm where the compiler and linker are actually invoked.
6. Read `Builder.Close`, from around line 340 in `action.go`, and confirm how cleanup is implemented.

**Answer**

**1. Where is the directory created?**

When `runRun` (run.go line 89) calls `work.NewBuilder("", ...)`, its implementation in action.go (lines 281–310) executes:

```go
tmp, err := os.MkdirTemp(cfg.Getenv("GOTMPDIR"), "go-build")
...
b.WorkDir = tmp
```

This creates one temporary directory corresponding to `$WORK`.

There is also a `NewObjdir` operation, from around line 391 in action.go, that assigns paths for package-specific subdirectories such as `b001/`, `b002/`, and so on below `$WORK`.

```go
func (b *Builder) NewObjdir() string {
	b.objdirSeq++
	return str.WithFilePathSeparator(filepath.Join(b.WorkDir, fmt.Sprintf("b%03d", b.objdirSeq)))
}
```

At this point, however, it has only assembled a path string. The actual directory creation happens later, in `Builder.build` or `Builder.link` in exec.go, through `sh.Mkdir(a.Objdir)`. In other words, the `$WORK` directory and the package-specific `bNNN/` directories differ in when their paths are determined and when the directories are actually created in the code.

**2. Where is the build run?**

`runRun` (around lines 170–173) contains these three lines:

```go
a1 := b.LinkAction(moduleLoaderState, work.ModeBuild, work.ModeBuild, p)
a1.CacheExecutable = true
a := &work.Action{Mode: "go run", Actor: work.ActorFunc(buildRunProgram), Args: cmdArgs, Deps: []*work.Action{a1}}
b.Do(ctx, a)
```

`LinkAction` (from around line 922 in action.go) only assembles an `Action` (`a1`) with dependencies expressing “compile first, then link”. It does not itself call the compiler or linker. The actual external-command execution is performed by `Builder.build` in `work/exec.go` (compilation, from around line 723) and `Builder.link` (linking, from around line 1591). As `b.Do(ctx, a)` follows the dependency graph, it calls these two functions in order, and only then do `compile` and `link` actually run. Thus, the place that assembles what to build (`LinkAction`) is separate in the source code from the place that performs the build (`Builder.build` and `Builder.link`).

**3. How is the built file handled?**

In `LinkAction` (around line 955 in action.go), the path of the executable produced by linking is recorded in the `Action`’s `built` field:

```go
a.Target = a.Objdir + filepath.Join("exe", name) + cfg.ExeSuffix
a.built = a.Target
```

The final step in `go run`, `buildRunProgram` (from around line 200 in run.go), retrieves this value as `a.Deps[0].BuiltTarget()` and starts it as an executable.

```go
cmdline := str.StringList(work.FindExecCmd(), a.Deps[0].BuiltTarget(), a.Args)
```

After execution finishes, the deferred `Builder.Close()` (from around line 340 in action.go), registered by `runRun` (around lines 90–94), runs:

```go
if !cfg.BuildWork {
	if err := robustio.RemoveAll(b.WorkDir); err != nil {
		return err
	}
}
```

This removes the `$WORK` directory, including the executable and all intermediate files. The built executable is therefore treated as a disposable file: it is created in a temporary directory, its path is retained long enough to start it, and the entire directory is deleted afterward.

</details>

---

<details>
<summary>Trivia: You can change the temporary directory location</summary>

The `$WORK` directory removed by `Builder.Close()` is created by the call `os.MkdirTemp(cfg.Getenv("GOTMPDIR"), "go-build")`.
The first argument is the `GOTMPDIR` environment variable, so setting `GOTMPDIR` changes where this temporary directory is created. If it is unset, the operating system’s standard temporary directory is used.
You can also confirm this in the `GOTMPDIR` description from `go help environment`.

</details>

---

## Starting points for investigation

- https://go.dev/cmd/go/
- https://pkg.go.dev/cmd/go#hdr-Compile_and_run_Go_program
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/run/run.go
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/action.go
- https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/go/internal/work/exec.go
