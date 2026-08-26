# How to Explore 04-deep-dive (Reverse Search Guide)

This category investigates Go language rules, standard-library implementations, and design history by following primary sources. Do not open an individual answer first: begin with the shared entry point, describe the observed phenomenon, and trace each claim to evidence.

## 1. Read the language specification first

For syntax, types, comparability, memory size, and error-handling rules, start with [The Go Programming Language Specification](https://go.dev/ref/spec).

1. Use the table of contents or the `f` key to find the section closest to the concept in the problem.
2. Map each claim to the specification's definitions, restrictions, and exceptions.
3. For behavior not determined by the specification, continue to the standard-library documentation and source.

## 2. Read the standard-library contract

Start at [Go Documentation](https://go.dev/doc), then open the relevant package from [pkg.go.dev/std](https://pkg.go.dev/std). Read the Overview, API comments, examples, and related types and methods. Follow declarations to the Go source and compare the explanation with the implementation.

## 3. Trace design history

For “why this design?” questions:

1. Read the release notes for the target version (`go.dev/doc/go1.<version>`).
2. Follow linked official blogs, proposals, and Issues.
3. Read the full issue discussion and related links, separating accepted facts, rejected alternatives, and your own interpretation.
4. Finally compare the result with versioned source and local measurements.

Use the [Go Blog index](https://go.dev/blog/all) when searching official posts. Do not rely on external articles without returning to Go project primary sources.

## 4. Read versioned implementation source

Open the [Go source](https://cs.opensource.google/go/go), identify the file through pkg.go.dev or the official documentation, and use a tag such as `refs/tags/go1.26.5` matching the README's Go version. Read callers, field comments, and related type definitions, not only the target lines.

## 5. Reproduce behavior locally

Run executable examples from the README's working directory. Record `go version`, exit status, stdout, and stderr, then compare them with the documentation. Run compilation-error examples with the same toolchain. For small browser experiments, use the [Go Playground](https://go.dev/play/) and verify that shared source matches the README.

## Investigation order

Observed phenomenon -> official documentation -> language specification or API contract -> design discussion -> versioned source -> local measurement. If primary sources do not establish a conclusion, say “unknown” or mark the remainder as inference instead of overstating it.
