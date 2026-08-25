[Back to Workshop](../README.md)

# How to Investigate 02-features (Reverse Lookup Guide)

In this category, you will explore Go features and syntactic structures by reading the Go language specification, official documentation, and official blogs and proposals. When you get stuck, start by consulting primary sources using the following steps.

- **I want to learn about language specifications and grammar rules**
  - Open https://go.dev/ref/spec
  - Because the official documentation is in English, searching the page with the following keywords (in English), based on what you want to investigate (the task, behavior, or concept), makes it easier to find the relevant specification. Use page search (Ctrl+F / Cmd+F).
    - **Variables and constants**
      - Variable declarations, initialization, and the detailed rules for `:=`: `Variable declarations`, `Short variable declarations`
      - Rules for constants and sequential numbering with `iota`: `Constant declarations`, `Iota`
    
    - **Types (data structures)**
      - Defining named types and type aliases: `Type declarations`, `Alias declarations`
      - Defining structs and embedding fields: `Struct types`, `Embedded field`
      - Differences between arrays and slices, and their specifications: `Array types`, `Slice types`
      - The specification for maps (associative arrays): `Map types`
      - The specification for interfaces: `Interface types`
      - The specification for pointer types: `Pointer types`
      - Generics, type parameters, and constraints: `Type parameters`, `Type constraints`, `Core types`

    - **Functions and methods**
      - Defining functions, return values, and variadic arguments: `Function declarations`
      - Defining methods associated with structs and types: `Method declarations`
      - Anonymous functions and closures: `Function literals`

    - **Control syntax**
      - Rules for `if` statements: `If statements`
      - Loops and iteration with `range`: `For statements`, `For statements with range clause`
      - `switch` statements and branching: `Expression switches`
      - Branching and checking by an interface's type: `Type assertions`, `Type switches`
      - Skipping and exiting loops, and jumps using labels: `Break statements`, `Continue statements`, `Labeled statements`
      - Deferred execution immediately before leaving a function (cleanup processing): `Defer statements`

    - **Concurrency (goroutines and channels)**
      - Asynchronous processing (starting goroutines): `Go statements`
      - Defining channels and sending and receiving: `Channel types`, `Send statements`, `Receive operator`
      - Waiting for multiple channels without blocking: `Select statements`

    - **Memory, error handling, and other topics**
      - Memory allocation (the difference between `new` and `make`): `Allocation`, `Making slices, maps and channels`
      - Strict rules for type conversions (casts): `Conversions`
      - Raising and recovering from panics (`panic`, `recover`): `Handling panics`, `Run-time panics`
      - Package import rules: `Import declarations`
      - Lexical rules for source code (comments, automatic semicolon insertion, and so on): `Lexical elements`, `Semicolons`

- **I want to find official documentation from a list**
  - Open https://go.dev/doc
  - To investigate the specifications and usage of standard packages: https://pkg.go.dev/std
  - To learn basic Go syntax and features by running them in a browser: https://go.dev/tour/ (A Tour of Go)
  - To check how to write idiomatic Go code (idioms and best practices): https://go.dev/doc/effective_go (Effective Go)
  - To learn more about Go commands (build, test, mod, and so on): https://go.dev/cmd/go/
  - To learn about modules and dependency management: https://go.dev/doc/modules/managing-dependencies
  - To learn about the Go memory model (such as memory visibility between goroutines): https://go.dev/ref/mem (The Go Memory Model)
  - To learn how to connect to and operate on databases through a tutorial: https://go.dev/doc/tutorial/database-access
  - To learn how to build a Web API (RESTful API) through a tutorial: https://go.dev/doc/tutorial/web-service-gin
  - To check security information and the vulnerability database: https://go.dev/security/

- **I want to find an official blog post from a list**
  - Open https://go.dev/blog/all
  - If you are investigating a topic involving an older language specification, it is faster to search from the list below.
  - Searching by blog author makes it easier to find what you need.
    - Rob Pike
      - The language's internal mechanisms and explanations of specifications (strings, slices, how `append` works, the laws of reflection, constants)
      - Design philosophy and idioms (Go's declaration syntax, Errors are values)
      - Tools and ecosystem (`go generate`, `go fmt`)
      - Cultural background (The Go Gopher (the mascot's background), Go fonts)
    - Russ Cox
      - Long-term vision and governance (Toward Go 2, forward and backward compatibility, and toolchain management)
      - Modules and dependencies (Go Modules, package versioning proposals)
      - Security and core libraries (secure random numbers, command PATH security)
      - Anniversary articles (the annual “Happy Birthday, Go!”)
    - Robert Griesemer
      - In-depth explanations of the type system (how type inference works, comparable types, the removal of core types)
      - Extensions to the language specification and syntax (alias names, syntax support for error handling)
      - Leading proposals for Go 2 and future versions
    - Ian Lance Taylor
      - Generics in general (the proposal to introduce generics, when to use them, deconstructing type parameters)
      - New function features (`range` over function types)
      - Low-level and compiler-related topics (Gccgo)
    - Andrew Gerrand
      - Official announcements of major and minor releases
      - Practical tutorials (JSON and Go, how to use slices, how to write Web applications)
      - Tool introductions (The Go Playground, Godoc)
      - Event reports (Google I/O, GopherCon, and meetup information from around the world)
    - Security-related: Filippo Valsorda, Julie Qiu
    - Tools and package-related: Alan Donovan, Damien Neil
    - New version releases: Go team

- **I want to investigate features in the latest version**
  - Access the release notes at `go.dev/doc/go1.<version>` (`version` is Go's minor version).
  - For Go 1.27, see the release notes at https://go.dev/doc/go1.27
  - If you want to follow the discussions behind the release-note entries, inspect the HTML source code using developer tools.
    - Related GitHub issue numbers are embedded in forms such as `go.dev/issue/<Issue number>`.
    - Open https://go.dev/issues/<Issue number> and then open the related issue.

- **I want to learn about actual behavior and implementation details**
  - Follow links from `pkg.go.dev` to the Go source code search at https://cs.opensource.google/go/go and read the implementation code for the standard library or runtime.

- **I want to try the behavior locally**
  * Use the Go Playground (https://go.dev/play/). You can switch to the development version (Dev version) or past versions to compare and verify behavior.
