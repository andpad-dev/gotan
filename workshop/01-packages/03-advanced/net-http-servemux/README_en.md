[Workshop guide and scenario index](../../../README.md) | [How to research 01-packages](../../README.md)

# Decode the Routing Rules of `http.ServeMux`

The administration API registers `/reports/latest` and `/reports/{id}` in the same `ServeMux`. The more specific route is selected without relying on registration order, and a wrong method returns 405. Why does this happen? Let's investigate the background.

When you [run the following code in the Go Playground](https://go.dev/play/p/fbR5kWMHL2Z), you can observe the behavior of the literal `latest`, the wildcard, and a method mismatch.

This scenario includes a `go.mod` that enables the routing rules introduced in Go 1.22 and later, plus a `main.go` containing the code above. To try it locally, run `go run .` in this directory. Before running it, also check `go version`, `go env GOMOD`, and `go env GODEBUG`.

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /reports/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "latest")
	})
	mux.HandleFunc("GET /reports/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "report=%s", r.PathValue("id"))
	})

	for _, path := range []string{"/reports/latest", "/reports/2026-08"} {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		fmt.Printf("GET %s -> %d %q\n", path, recorder.Code, recorder.Body.String())
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/reports/2026-08", nil))
	fmt.Printf("POST /reports/2026-08 -> %d\n", recorder.Code)
}
```

Output:

```text
GET /reports/latest -> 200 "latest"
GET /reports/2026-08 -> 200 "report=2026-08"
POST /reports/2026-08 -> 405
```

---

## Question 1: Which route wins?

Regardless of whether `GET /reports/latest` is registered before or after `GET /reports/{id}`, which handler receives `GET /reports/latest`? Explain the conditions under which `POST /reports/2026-08` becomes 405.

<details>
<summary>Hint</summary>

- Use [How to investigate a category](../../README.md) as your starting point and read `ServeMux`'s Patterns and Precedence.
- Think of “more specific” in terms of which set of requests a pattern matches, rather than the length of its string.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read about enhanced routing patterns in the [Go 1.22 Release Notes](https://go.dev/doc/go1.22).
2. Read Patterns and Precedence in [http.ServeMux](https://pkg.go.dev/net/http#ServeMux).
3. In Go 1.27.0, confirm that [`routing_tree.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/routing_tree.go;l=154-198) tries literal, single-wildcard, and multi-wildcard segments in that order.
4. Read [`findHandler`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/server.go;l=2751-2764) and confirm that a path matching another method produces 405 and an `Allow` header instead of 404.

**Answer**

`GET /reports/latest` requires the literal `latest`, so its set of matching requests is narrower than that of `GET /reports/{id}` for the same method, making it more specific. Therefore, the `latest` handler is selected regardless of registration order.

`POST /reports/2026-08` does not match either `GET` pattern. However, because there is a pattern for another method at the same path, `ServeMux` returns 405 Method Not Allowed.

</details>

---

## Question 2: Why do these two patterns conflict when registered?

Using this scenario's `go.mod` (`go 1.22`) without `httpmuxgo121`, an API developer tries to register `GET /reports/{id}` and `/reports/latest`. Both match some GET requests, but why can neither always be called “more specific,” causing `HandleFunc` to panic?

First, run the actual registration code in the [Go Playground](https://go.dev/play/p/iMHQSYKMgpE). With the standard Go 1.22-or-later routing behavior, the second `HandleFunc` call panics and `registered` is not printed.

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /reports/{id}", func(http.ResponseWriter, *http.Request) {})
	mux.HandleFunc("/reports/latest", func(http.ResponseWriter, *http.Request) {})
	fmt.Println("registered")
}
```

If it does not panic locally, inspect `go version`, `go env GOMOD`, and `go env GODEBUG`. Compare the standard result with `GODEBUG=httpmuxgo121=1`, and explain the condition that causes the difference.

<details>
<summary>Hint</summary>

- One has a narrower method and a broader path, while the other has a broader method and a narrower path.
- Consider sets of method-and-path pairs. Confirm that omitting the method matches all methods and that `GET` also matches `HEAD`.
- Check the definition of a `ServeMux` conflict.
- Check which condition makes `GODEBUG` select the previous routing behavior.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Run the [shared Go Playground code](https://go.dev/play/p/iMHQSYKMgpE) with the standard behavior and observe the panic at the second registration and the absence of `registered`.
2. Re-read the enhanced routing patterns section of the [Go 1.22 Release Notes](https://go.dev/doc/go1.22).
3. Read the explanation of Precedence and conflicts in [http.ServeMux](https://pkg.go.dev/net/http#ServeMux).
4. Check `httpmuxgo121` in [Go, Backwards Compatibility, and GODEBUG](https://go.dev/doc/godebug).
5. Confirm the “neither is more specific” rule in [`conflictsWith` in Go 1.27.0](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/pattern.go;l=219-240).

**Answer**

`GET /reports/{id}` accepts only GET but accepts any ID, while `/reports/latest` accepts any method but only accepts `latest`. `GET /reports/latest` matches both, but one pattern is narrower only in its method and the other only in its path.

Neither matching set is a strict subset of the other, so precedence cannot be determined. The design panics during registration to expose the ambiguity early. If both patterns are needed, align their methods or paths so that one is clearly narrower.

With `GODEBUG=httpmuxgo121=1`, however, it does not panic. That compatibility setting restores the old `ServeMux` behavior, where the newer method and wildcard syntax is not interpreted. Reproducing the panic therefore requires checking not only the toolchain version but also the main module's `go` line and `GODEBUG`.

</details>

---

## Question 3: Why did a change in string syntax require a compatibility discussion?

In Go 1.21, a pattern containing `{id}` was not a special wildcard. What changed in Go 1.22, and what compatibility setting was provided? List the points to check when migrating a project.

<details>
<summary>Hint</summary>

- Find how `{` and `}` were treated in the Go 1.22 release notes.
- Search for `httpmuxgo121`.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read the compatibility-related paragraph in the [Go 1.22 Release Notes](https://go.dev/doc/go1.22).
2. Read the explanation of `httpmuxgo121` in [Go, Backwards Compatibility, and GODEBUG](https://go.dev/doc/godebug).
3. Read [`servemux121.go` in Go 1.27.0](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/servemux121.go;l=25-38) and confirm that `httpmuxgo121=1` is checked once at startup.

**Answer**

Starting in Go 1.22, patterns with methods and the `{name}` / `{name...}` wildcards are interpreted as such. Previously, `{id}` was treated as an ordinary string, so the same registration could match different requests.

The migration setting `httpmuxgo121=1` restores the old behavior. First test existing patterns containing `{` and `}`, escaped paths, and registration-time panics, then adapt to the new pattern rules rather than relying permanently on the compatibility setting.

</details>

---

## Question 4: Why is it not “the last registered pattern wins”?

Trace the proposal issue and implementation, then explain what benefits order-independent precedence and registration-time conflict detection provide when multiple teams maintain administration API routes.

<details>
<summary>Hint</summary>

- Read Precedence and Performance in the proposal issue.
- The comment in `routing_tree.go` explains why more specific routes are tried first.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Re-read the enhanced routing patterns section of the [Go 1.22 Release Notes](https://go.dev/doc/go1.22).
2. Read Precedence, Backwards Compatibility, and Performance in [ServeMux extension proposal Issue #61410](https://github.com/golang/go/issues/61410).
3. Read [Go Blog: Routing Enhancements for Go 1.22](https://go.dev/blog/routing-enhancements).
4. Read [`routing_tree.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/routing_tree.go;l=168-198) and [`pattern.go`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/pattern.go;l=219-240) in Go 1.27.0, and connect the specific-first traversal with registration-time conflict detection.

**Answer**

If the result changed with registration order, simply adding a route by another team could change the destination of an existing API. Selecting the more specific pattern as a set means that the destination can be explained by the rules rather than by the order of configuration.

At the same time, overlapping patterns that cannot be compared are not left hidden until runtime; `ServeMux` panics during registration. The implementation searches more specific patterns first and detects conflicts at startup. When reviewing administration API changes, teams can also discuss which sets of requests are added or removed.

</details>

---

<details>
<summary>Further note</summary>

[`GET` patterns also match `HEAD`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/pattern.go;l=253-279), but they do not match methods such as POST. If only `GET /` is registered, `POST /anything` does not reach that handler; [`findHandler` returns 405 with `Allow: GET, HEAD`](https://cs.opensource.google/go/go/+/refs/tags/go1.27.0:src/net/http/server.go;l=2751-2764). To send every method to the same handler, register `/` without a method. Test both 405 and the `Allow` header without confusing `GET` with a methodless pattern.

</details>

---

## Investigation starting points

- [Go 1.22 Release Notes](https://go.dev/doc/go1.22)
- [How to investigate 01-packages](../../README.md)
- [package net/http](https://pkg.go.dev/net/http)
