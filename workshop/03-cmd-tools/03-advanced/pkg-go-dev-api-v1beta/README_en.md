[Scenario index (Japanese)](../../../SCENARIOS.md) | [Workshop guide (Japanese)](../../../README.md) | [How to research 03-cmd-tools](../../README.md)

# Use the pkg.go.dev API (v1)

The pkg.go.dev browser UI cannot compare and sort package candidates by import count and maintenance status.

We considered scraping, but the **pkg.go.dev API**, released in beta in [June 2026](https://opensource.googleblog.com/2026/06/a-new-pkggodev-api-for-go.html), provides structured JSON directly.

<details><summary>Investigation entry points</summary>

Start with the [03-cmd-tools research guide](../../README.md), then use the primary sources listed at the end of this scenario.

</details>

## Question 1: How do I sort and filter search results?

The search API (`/v1/search`) sorts a `router` search by relevance by default, but has no parameter for import count or update time. Why is sorting unavailable? The API does provide a `filter` parameter, but the filter itself is a distinctive “subset of Go expressions,” rather than SQL or a raw regular expression. Why was it designed this way?

Run this request and check the condition rather than relying on a result count that may change over time:

```bash
curl -L "https://pkg.go.dev/v1/search?q=xyzzy&filter=hasPrefix%28packagePath%2C%20%22github.com%22%29" \
  | jq -c '{hasItems: ((.items | length) > 0), allGithub: ([.items[].packagePath] | all(startswith("github.com/")))}'
```

```text
{"hasItems":true,"allGithub":true}
```

<details>
<summary>Hint</summary>

- The [API documentation](https://pkg.go.dev/v1/api) describes `filter` in “Requests.”
- Variables available to a filter are determined by the JSON fields in each endpoint's response type.
- Check which fields `SearchResult` has, including whether it contains popularity metrics such as import count.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Start at [Go Documentation](https://go.dev/doc/) and open the [pkg.go.dev API documentation](https://pkg.go.dev/v1/api).
2. Check the response type for `/v1/search` in “Routes,” then follow the `SearchResult` link and inspect its fields.
3. Inspect the [`SearchResult` definition in pkgsite v0.4.0](https://cs.opensource.google/go/x/pkgsite/+/v0.4.0:internal/api/types.go;l=121).

**Answer**

The `/v1/search` response contains only basic information such as `packagePath`, `modulePath`, `version`, and `synopsis`; it does not return metadata such as import count or update time. The definition is in [pkgsite v0.4.0](https://cs.opensource.google/go/x/pkgsite/+/v0.4.0:internal/api/types.go;l=121):

```go
type SearchResult struct {
    PackagePath string `json:"packagePath"`
    ModulePath  string `json:"modulePath"`
    Version     string `json:"version"`
    Synopsis    string `json:"synopsis"`
}
```

Without those fields, a client cannot sort by them. The documented order is relevance, and no alternate sort parameter is defined. To gather other information, query `/v1/package/{path}` or `/v1/imported-by/{path}` for each result.

The [API documentation](https://pkg.go.dev/v1/api) defines filters as a subset of Go expressions with route-specific variables and approved functions. The filter is not arbitrary SQL or a raw regular expression, although the approved `matches(string, regexp)` function does accept a regular expression as its second argument. The documentation specifies the grammar but does not state that safety was the design motive, so that rationale should be treated as an inference. A filter narrows results; it does not specify ordering.

For example, to select only paths beginning with `github.com`:

- Go expression: `hasPrefix(packagePath, "github.com")`
- Encoded: `hasPrefix%28packagePath%2C%20%22github.com%22%29`

```bash
curl -L "https://pkg.go.dev/v1/search?q=xyzzy&filter=hasPrefix%28packagePath%2C%20%22github.com%22%29" | jq .
```

</details>

---

## Question 2: What happens when a package path is ambiguous?

The API documentation says:

> Package paths are ambiguous: the same path a/b/c could be package c in module a/b or package b/c in module a. If both exist, the API returns an error with a list of possible modules in its candidates field.

Why can a package path theoretically be ambiguous, and how do the browser UI and API behave differently?

First compare an unambiguous path with a real ambiguous path:

```bash
curl -L "https://pkg.go.dev/v1/package/golang.org/x/time/rate" \
  | jq -c '{modulePath, path, name}'
```

```text
{"modulePath":"golang.org/x/time","path":"golang.org/x/time/rate","name":"rate"}
```

```bash
pkgsite="https://pkg.go.dev"
curl -L "$pkgsite/v1/package/github.com/hashicorp/consul/api" \
  | jq -c '{candidateModules: [.candidates[].modulePath]}'
```

```text
{"candidateModules":["github.com/hashicorp/consul/api","github.com/hashicorp/consul"]}
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

1. Start at [Go Documentation](https://go.dev/doc/) and read the relationship between module and package paths in the [Go Modules reference](https://go.dev/ref/mod).
2. Read the package-path ambiguity section under “Requests” in the [API documentation](https://pkg.go.dev/v1/api).
3. Run the two requests above and compare a package response with a `candidates` response.
4. Retry the ambiguous request with `?module=github.com/hashicorp/consul/api` and confirm that the explicit module resolves it.

**Answer**

Module and package paths are independent. The same string can describe package `c` in module `a/b` or package `b/c` in module `a`, so both can have package path `a/b/c`.

The browser UI chooses the longest matching module path. The API instead returns candidates for ambiguity and asks the client to specify the module with a query parameter. The documentation states this behavior; interpreting it as avoiding an implicit choice is an inference.

The observed `golang.org/x/time/rate` response resolves to module `golang.org/x/time`. The real `github.com/hashicorp/consul/api` request returns both `github.com/hashicorp/consul/api` and `github.com/hashicorp/consul` in `candidates`. Retry with an explicit query such as `?module=github.com/hashicorp/consul/api`.

</details>

---

## Question 3: Why rate limiting and pagination?

The API has a rate limit of 45 QPS (queries per second) per IP block. Large result sets are returned through pagination. Why are these limits and designs necessary, and what does `nextPageToken` represent?

```bash
curl -L "https://pkg.go.dev/v1/search?q=xyzzy&limit=1" \
  | jq -c '{items: (.items | length), hasNext: (.nextPageToken != null and .nextPageToken != "")}'
```

```text
{"items":1,"hasNext":true}
```

<details>
<summary>Hint</summary>

- `nextPageToken` is an opaque string.
- The documentation says to make no request changes other than adding the token.
- Separate what the documentation guarantees from your inference about why the limit exists.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Start at [Go Documentation](https://go.dev/doc/) and open the [pkg.go.dev API documentation](https://pkg.go.dev/v1/api).
2. Read “Rate Limiting” and “Pagination.”
3. Run the request above, then pass the returned token without changing the original query to retrieve the next page.

**Answer**

The documentation states a 45-QPS-per-IP-block limit and a `429 Too Many Requests` response when it is exceeded. Protecting service capacity and preventing one client from monopolizing resources are reasonable interpretations, but the API documentation does not state those motives or promise that 45 QPS is sufficient for a particular workload.

`nextPageToken` is an **opaque token** whose contents clients must not interpret. The documentation warns: “Changing the request in any way other than providing a token may result in an error.” Keep the original request unchanged and add only the token:

```bash
page_token=$(curl -sS "https://pkg.go.dev/v1/search?q=xyzzy&limit=1" | jq -r .nextPageToken)
pkgsite="https://pkg.go.dev"
curl -sS -G "$pkgsite/v1/search" \
  --data-urlencode "q=xyzzy" \
  --data-urlencode "limit=1" \
  --data-urlencode "token=$page_token" \
  | jq -c '{items: (.items | length)}'
```

```text
{"items":1}
```

</details>

---

## Question 4: Why does `imported-by` exclude packages in the same module?

`/v1/imported-by/{path}` returns packages that import the specified package, but the documentation says:

> Paths of packages importing the package at {path}, not including packages in the same module.

Why exclude packages in the same module?

```bash
curl -L "https://pkg.go.dev/v1/imported-by/golang.org/x/time/rate" \
  | jq -c --arg modulePath "golang.org/x/time" '{pageSize: (.importedBy.items | length), sameModuleItems: ([.importedBy.items[] | select(startswith($modulePath + "/"))] | length), hasNext: (.importedBy.nextPageToken != null and .importedBy.nextPageToken != "")}'
```

```text
{"pageSize":100,"sameModuleItems":0,"hasNext":true}
```

<details>
<summary>Hint</summary>

- Dependencies within a module differ from dependencies between modules.
- Run the request and inspect only what the returned page demonstrates before inferring the intended use case.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Start at [Go Documentation](https://go.dev/doc/) and read about module boundaries in the [Go Modules reference](https://go.dev/ref/mod).
2. Read the `/v1/imported-by/{path}` description in the [API documentation](https://pkg.go.dev/v1/api).
3. Run the request above and confirm that the returned page contains no path beginning with the target module path.

**Answer**

The documented fact is that same-module packages are excluded. This can be interpreted as focusing the endpoint on impact across module boundaries: packages in one module are versioned together, while other modules are independent consumers. That use-case explanation is an inference from the API behavior and Go's module boundary, not a rationale explicitly stated by the API documentation.

</details>

---

## Primary sources

- [Go Documentation](https://go.dev/doc/)
- [pkg.go.dev API](https://pkg.go.dev/v1/api)
- [pkgsite internal/api](https://pkg.go.dev/golang.org/x/pkgsite/internal/api) (response type definitions)
- [pkgsite source](https://cs.opensource.google/go/x/pkgsite)
- [Go Modules reference](https://go.dev/ref/mod)
