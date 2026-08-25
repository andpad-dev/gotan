# Implement the `slog.Handler` Interface

The team has decided that logs should be output in YAML format.
To do that, we need to write a custom `slog` log handler.
Let's follow only primary sources and gather the information needed for the implementation and unit tests.

## Question 1: Investigate the `slog.Handler` interface

How many methods does the `slog.Handler` interface have, and what does each method do?
Consider which method is at the center of a handler.
Also look for the types that implement this interface in the standard library.

For example, when `logger.Info("login", "user", "alice")` is called, are the role of receiving and outputting the log and the role of adding common attributes before output the same? Imagine the flow by which a log reaches a handler, then classify the four methods.

<details>
<summary>Hint</summary>

- First make a rough prediction from the method name and the first sentence of its description. If the English is difficult, translate it; there is no need to investigate the details yet.
- Read the [Handler interface definition](https://pkg.go.dev/log/slog#Handler) and each method's comments from the perspective of the order in which the calling `Logger` uses them.
- The LevelHandler in the [Example](https://pkg.go.dev/log/slog#pkg-examples) is a wrapper, so it is easy to read. Consider whether this approach (a wrapper) is sufficient for the current use case.
- When listing standard implementations, distinguish whether `TextHandler` / `JSONHandler` / `MultiHandler` are types or `DiscardHandler` is a value.
- Tentatively think of `Enabled` as “whether to pass this log,” `Handle` as “how to output the passed log,” and `WithAttrs` / `WithGroup` as “how to carry common information into subsequent logs,” then confirm this in the comments.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open https://pkg.go.dev/log/slog and search for `Handler` to jump to it.
2. Read the interface definition and the comments for each method.
3. Use the Example and the Index to find the standard types that implement `Handler`.

**Answer**

There are four methods.

- `Enabled(context.Context, Level) bool`: determines whether to process a log at that level.
- `Handle(context.Context, Record) error`: receives and outputs a log record. This is the method at the center of the handler.
- `WithAttrs([]Attr) Handler`: returns a new handler with additional attributes.
- `WithGroup(string) Handler`: returns a new handler with a group name.

The standard library has `TextHandler` / `JSONHandler` / `MultiHandler` implementation types for `Handler`. It also has `DiscardHandler`, a package variable of type `Handler` that discards all logs.

The LevelHandler in the [Wrapping](https://pkg.go.dev/log/slog#example-package-Wrapping) Example wraps an existing handler and replaces only the level check.
This is sufficient when you want to change only part of the output, but here we want YAML as the output format itself, so we need to write our own `Handle`; a wrapper is not enough.

</details>

---

## Question 2: Investigate points to note when implementing `slog.Handler`

There are several points to note when writing a custom handler, and they are documented.
Find and read them, then discuss the points as a group.

- Why is it not enough to embed `slog.Handler` and implement only the methods you need?
- What kinds of logs will stop being output correctly if you do not call the `Resolve` method of `slog.Value`?

<details>
<summary>Hint</summary>

- The slog documentation's Overview has a section called “[Writing a handler](https://pkg.go.dev/log/slog#hdr-Writing_a_handler).”
- The Examples are also useful: [DiscardHandler](https://pkg.go.dev/log/slog#example-package-DiscardHandler) / [Wrapping](https://pkg.go.dev/log/slog#example-package-Wrapping)
- There is a more detailed guide at https://golang.org/s/slog-handler-guide
  - Notice why the `mu` field in the guide's IndentHandler is `*sync.Mutex` (a pointer) rather than `sync.Mutex`.
- After reading the section on “embedding” at the beginning of the guide, organize how `Logger` combines calls to `Enabled`, `WithAttrs`, `WithGroup`, and `Handle` from the perspectives of method delegation and copying. Do not decide the conclusion in advance; compare each of the four method contracts.
- Consider a password type whose `LogValue` returns `slog.String("password", "REDACTED")` as an example of `LogValuer`. Before calling `Resolve`, imagine what is lost if the value is converted directly to a string, then confirm it.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read the “Writing a handler” section in the slog documentation's Overview.
2. Read the detailed guide linked from there, https://golang.org/s/slog-handler-guide ([slog-handler-guide in golang/example](https://github.com/golang/example/blob/master/slog-handler-guide/README.md)).

**Answer**

There are three main points to note.

**1. Implement all four methods.**
It is tempting to embed `slog.Handler` and implement only the methods you need, but `Logger` and `Handler` are tightly coupled, so that does not work.
The guide states this explicitly:

> it is tempting to embed slog.Handler in your custom handler and implement only the methods that you need. Loggers and handlers are too tightly coupled for that to work. You should implement all four handler methods.

**2. Copy and return the handler in `WithAttrs` / `WithGroup`.**
This is why the `mu` field in the guide's IndentHandler is `*sync.Mutex` rather than `sync.Mutex`.
The copied handlers share the same output destination (`io.Writer`), so if the Mutex is held by value, each copy gets a different lock and mutual exclusion no longer works.
Using a pointer makes all copies share the same Mutex.

**3. Resolve attribute values before using them.**
When processing attributes inside `Handle`, first call `slog.Value.Resolve`.
For a value that implements `slog.LogValuer`, `Resolve` is what actually calls the `LogValue` method and converts it to the value intended for the log.
Without calling it, a LogValuer value, such as a type that replaces a password with `REDACTED`, will not be output in its intended form.

</details>

---

## Question 3: Investigate how to test a `slog.Handler`

Let's investigate how to use unit tests to verify that a custom handler follows the rules of `slog.Handler`.

<details>
<summary>Hint</summary>

- The https://golang.org/s/slog-handler-guide includes a section about testing.
- The standard library has a package dedicated to testing `slog.Handler`.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read the “[Testing](https://github.com/golang/example/blob/master/slog-handler-guide/README.md#testing)” section of slog-handler-guide.
2. Read the documentation for the [testing/slogtest](https://pkg.go.dev/testing/slogtest) package used there.

**Answer**

Use the standard `testing/slogtest` package.
Pass `slogtest.TestHandler` a custom handler and a function that converts the output into and returns a slice of maps, and it will verify all at once the rules that a `Handler` must follow, including resolving attributes, handling groups, and ignoring empty attributes.
The point is that you can comprehensively verify conformance to the `slog.Handler` specification without enumerating test cases yourself.

</details>

---

## Investigation starting points

- https://pkg.go.dev/log/slog