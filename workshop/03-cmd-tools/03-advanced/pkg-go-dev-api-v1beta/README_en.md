[Workshop guide and scenario index](../../../README.md) | [How to research 03-cmd-tools](../../README.md)

# Use the pkg.go.dev API (v1beta)

Our team must choose an HTTP router library. We want to compare and sort candidates by import count and maintenance status, but the pkg.go.dev browser UI has no such feature.

We considered scraping, but the **pkg.go.dev API**, released in beta in [June 2026](https://opensource.googleblog.com/2026/06/a-new-pkggodev-api-for-go.html), provides structured JSON directly. Let us investigate the background behind this API design.

## Question 1: How do I sort and filter search results?

The search API (`/v1beta/search`) sorts a `router` search by relevance by default, but has no parameter for import count or update time. Why is sorting unavailable? The API does provide a `filter` parameter, but its expression is a distinctive “subset of Go expressions,” rather than SQL or a regular expression. Why was it designed this way?

<details>
<summary>Hint</summary>

- The [API documentation](https://pkg.go.dev/v1beta/api) describes `filter` in “Requests.”
- Variables available to a filter are determined by the JSON fields in each endpoint's response type.
- Check which fields `SearchResult` has, including whether it contains popularity metrics such as import count.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Check the response type for `/v1beta/search` in the “Routes” section of https://pkg.go.dev/v1beta/api.
2. Follow the `SearchResult` link and inspect its fields.
3. Inspect the `SearchResult` definition in the pkg.go.dev source at https://cs.opensource.google/go/x/pkgsite.

**Answer**

The `/v1beta/search` response contains only basic information such as `packagePath`, `modulePath`, `version`, and `synopsis`; it does not return metadata such as import count or update time. The definition is at https://cs.opensource.google/go/x/pkgsite/+/refs/tags/v0.3.0:internal/api/types.go;l=115-121:

```go
type SearchResult struct {
    PackagePath string `json:"packagePath"`
    ModulePath  string `json:"modulePath"`
    Version     string `json:"version"`
    Synopsis    string `json:"synopsis"`
}
```

Without those fields, a client cannot sort by them. The API also does not offer a sort parameter because search results are designed to be returned by relevance. To gather other information, query `/v1beta/package/{path}` or `/v1beta/imported-by/{path}` for each result.

The [API documentation](https://pkg.go.dev/v1beta/api) defines filters as a subset of Go expressions. Restricting the operators and functions that the server evaluates makes them safer than accepting arbitrary SQL or regular expressions. A filter narrows results; it does not specify ordering.

For example, to select only paths beginning with `github.com`:

- Go expression: `hasPrefix(packagePath, "github.com")`
- Encoded: `hasPrefix%28packagePath%2C%20%22github.com%22%29`

```bash
curl -L "https://pkg.go.dev/v1beta/search?q=xyzzy&filter=hasPrefix%28packagePath%2C%20%22github.com%22%29" | jq .
```

</details>

---

## Question 2: What happens when a package path is ambiguous?

The API documentation says:

> Package paths are ambiguous: the same path a/b/c could be package c in module a/b or package b/c in module a. If both exist, the API returns an error with a list of possible modules in its candidates field.

Why can a package path theoretically be ambiguous, and how do the browser UI and API behave differently?

```bash
curl -L "https://pkg.go.dev/v1beta/package/golang.org/x/time/rate" | jq .
```

<details>
<summary>Hint</summary>

- Module paths and package paths are independent in Go's module system.
- A `rate` package in the `golang.org/x/time` module has package path `golang.org/x/time/rate`.
- In theory, a separate module could also have that path.
- The API documentation says it does not choose the longest match, unlike the UI.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

- Run `curl -s "https://pkg.go.dev/v1beta/package/golang.org/x/time/rate" | jq .` and observe the response.
- Read the package-path ambiguity section under “Requests” in the API documentation.
- Read the relationship between module and package paths in the [Go Modules reference](https://go.dev/ref/mod).

**Answer**

Module and package paths are independent. The same string can describe package `c` in module `a/b` or package `b/c` in module `a`, so both can theoretically have package path `a/b/c`. In practice this is rare because multiple modules in one repository are discouraged, nested modules are designed to avoid collisions, and modules commonly live at repository roots.

The browser UI chooses the longest matching module path. The API instead returns an error for ambiguity and asks the client to specify the module with a query parameter. This avoids silently returning an unexpected result.

For `golang.org/x/time/rate`, the response is unambiguous:

```json
{
  "modulePath": "golang.org/x/time",
  "version": "v0.15.0",
  "isLatest": true,
  "isStandardLibrary": false,
  "goos": "all",
  "goarch": "all",
  "path": "golang.org/x/time/rate",
  "name": "rate",
  "synopsis": "Package rate provides a rate limiter.",
  "isRedistributable": true
}
```

An ambiguous request may return:

```json
{
  "code": 400,
  "message": "ambiguous package path",
  "candidates": ["module1", "module2"]
}
```

Retry with an explicit query such as `?module=module1`.

</details>

---

## Question 3: Why rate limiting and pagination?

The API has a rate limit of 45 QPS (queries per second) per IP block. Large result sets are returned through pagination. Why are these limits and designs necessary, and what does `nextPageToken` represent?

<details>
<summary>Hint</summary>

- pkg.go.dev is a public service operated by Google.
- Rate limits protect resources and help prevent DoS attacks.
- `nextPageToken` is an opaque string.
- The documentation says to make no request changes other than adding the token.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

- Read the “Rate Limiting” and “Pagination” sections of the API documentation.
- Retrieve search results and inspect `nextPageToken`.
- Inspect the pagination implementation in the pkgsite source at https://cs.opensource.google/go/x/pkgsite.

**Answer**

The 45-QPS-per-IP-block limit protects the free public service from DoS, overload, resource monopolization, and excessive scraping. It is enough for ordinary use but constrains large batch jobs; exceeding it produces `429 Too Many Requests`.

`nextPageToken` is an **opaque token** containing an encoded page position. Its long hexadecimal-looking value is not intended for client interpretation. The documentation warns: “Changing the request in any way other than providing a token may result in an error.” Keep the original request unchanged and pass the token to retrieve the next page.

</details>

---

## Question 4: Why does `imported-by` exclude packages in the same module?

`/v1beta/imported-by/{path}` returns packages that import the specified package, but the documentation says:

> Paths of packages importing the package at {path}, not including packages in the same module.

Why exclude packages in the same module?

<details>
<summary>Hint</summary>

- Dependencies within a module differ from dependencies between modules.
- A common use is understanding the package's external impact.
- Packages in one module are usually managed by the same team and repository.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

- Read the `/v1beta/imported-by/{path}` description in the API documentation.
- Inspect the implementation in the pkgsite source at https://cs.opensource.google/go/x/pkgsite.
- Read about module boundaries in the [Go Modules reference](https://go.dev/ref/mod).

**Answer**

The main use case is understanding impact outside the module. Packages within one module are usually maintained by the same team, released together, and can be changed together. Dependencies between modules represent other projects and teams, where a breaking change has a broader impact. Thus `imported-by` focuses on external users and is useful for questions such as which projects a package change affects and how widely it is used.

</details>

---

## Research starting points

- https://pkg.go.dev/v1beta/api
- https://pkg.go.dev/golang.org/x/pkgsite/internal/api (response type definitions)
- https://cs.opensource.google/go/x/pkgsite (pkgsite source)
- https://go.dev/ref/mod (Go Modules reference)
