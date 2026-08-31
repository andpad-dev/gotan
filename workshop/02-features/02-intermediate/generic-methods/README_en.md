[Workshop guide and scenario index](../../../README.md) | [How to research 02-features](../../README.md)

# Add a Generic Method Helper to a Named Type

You want to add a generic method to a named type. Let's investigate how to do that.

Your internal utility package has a custom container type, `Slice[T]`.
You want to add a conversion helper that maps values to another type as a **method on** the container.

```go
type Slice[T any] struct {
	items []T
}

// This is what we want to write
func (s Slice[T]) Map[F any](f func(T) F) Slice[F] { ... }
```

When you try to compile this code with the Go 1.26.4 you are currently using, the compiler objects:

```
syntax error: method must have no type parameters
```

Apparently, “generic methods” are coming in Go 1.27. Let's first use primary sources to establish:

- what changed in the specification
- why this could not be written through Go 1.26
- what happens with interface methods

Then we'll write and run one for ourselves.

---

## Question 1: Investigate What Changed in the Specification

Identify exactly what became possible to write in a `method` in Go 1.27 using the release notes and [The Go Programming Language Specification](https://go.dev/ref/spec). Let's look as far as how the EBNF in the relevant specification section changed.

<details>
<summary>Hint</summary>

- As described in the category's reverse lookup procedure, “[I want to investigate features in the latest version](../../README.md),” first open the [Go 1.27 release notes](https://go.dev/doc/go1.27).
- There should be a paragraph about “generic methods” in the Changes to the language section.
- The release notes' HTML contains a related issue embedded in the form `go.dev/issue/<number>` (you can find it by viewing the source with developer tools).
- On the specification side, open the `Method declarations` section and compare how its EBNF differs from the older version.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Open the [Changes to the language section of the Go 1.27 release notes](https://go.dev/doc/go1.27#language) and read the paragraph about `generic methods`.
2. Open the embedded `go.dev/issue/77273` and jump to the Proposal section of [proposal #77273 “spec: generic methods for Go”](https://go.dev/issue/77273). The old and new EBNF are shown side by side.
3. Check the adopted EBNF in the specification's [Method declarations](https://go.dev/ref/spec#Method_declarations) section.

**Answer**

The release notes say:

> Go 1.27 now supports generic methods: a method declaration may declare its own type parameters. This widely anticipated change allows adding generic functions within the namespace of a particular data type where before one had to declare such functions with a scope of the entire package. Note that methods of interfaces may not declare type parameters nor can interface methods be implemented by generic methods.

The EBNF in the specification's [Method declarations](https://go.dev/ref/spec#Method_declarations) changed as follows, as contrasted in the Proposal section of proposal #77273:

- **Old (through Go 1.26)**
  ```
  MethodDecl = "func" Receiver MethodName Signature [ FunctionBody ] .
  ```
- **New (Go 1.27 and later)**
  ```
  MethodDecl = "func" Receiver MethodName [ TypeParameters ] Signature [ FunctionBody ] .
  ```

`[ TypeParameters ]` was inserted between `MethodName` and `Signature`. This aligns the position with the type parameters in a function declaration (`FunctionDecl`).

The Go 1.26.4 compiler does not accept the syntax for generic methods, so it rejects it with:

```
syntax error: method must have no type parameters
```

</details>

---

## Question 2: Why Can't Interface Methods Still Use Type Parameters, and Can a Generic Method Implement an Interface?

The release notes contain a significant qualification:

> methods of interfaces may not declare type parameters **nor can interface methods be implemented by generic methods**.

In other words, even in Go 1.27, “interface methods cannot have type parameters,” and “an interface cannot be implemented by a generic method.” Why draw the line here? Try the following code in the Playground and examine the compiler's explanation as well.

```go
package main

import "io"

type Reader struct{}

func (*Reader) Read[E any](p []E) (int, error) { return 0, nil }

func main() {
	var _ io.Reader = (*Reader)(nil)
}
```

(Run it on the Go Playground: https://go.dev/play/p/Lqn-PC8jg9E?v=gotip , and select Dev/gotip before running.)

<details>
<summary>Hint</summary>

- The Background section of proposal #77273 briefly explains why this was historically prohibited.
- The Examples section of the same proposal includes precisely the contrast between `Reader.Read[E any]` and `io.Reader`.
- The Background section links to the [No parameterized methods section of the Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#No-parameterized-methods), which contains a more detailed discussion.

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Read the Background section of [proposal #77273](https://go.dev/issue/77273) to understand the historical reason for the prohibition and the idea behind “A change of view.”
2. Check the relationship between `Reader.Read[E any]` and `io.Reader` in the Examples section of the same proposal.
3. Compile the example locally with `gotip` or in the Playground and inspect the compiler's error message.
4. For a deeper dive: read the historical discussion in the [No parameterized methods section of the Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#No-parameterized-methods).

**Answer**

**1. Why interface methods do not allow type parameters**

The relevant paragraph in the proposal's Background section gives the answer directly:

> Go doesn't support such generic interface methods because we don't know how to implement (calls of) them, or at least we don't know how to implement them efficiently. Specifically, because Go doesn't require a concrete type to declare the interfaces it implements, and instead this is a dynamic property, it cannot be known at compile time which of the infinite possible instantiations of concrete methods will be needed at run time.

The key point is that **interface satisfaction in Go is a dynamic property**. A type does not declare at compile time which interfaces it satisfies, so the compiler cannot enumerate in advance which type arguments will be used when a generic interface method is called at runtime. Because an efficient implementation would be difficult, the decision remains unchanged in Go 1.27: interface methods cannot have type parameters.

**2. Why concrete methods are allowed nevertheless (“A change of view”)**

The proposal reframes the issue as follows:

> concrete methods are not just a means for implementing interfaces. A method is a function associated with a type, and accessed through the namespace of that type. Therefore methods are useful for organizing code even if they don't ever implement an interface.

If methods are viewed only as a way to implement interfaces, allowing them on concrete types would seem to require allowing them on interfaces too. However, methods are also functions in a type's namespace, so introducing generic concrete methods for the value of that namespace alone is a coherent design.

**3. Does `Reader.Read[E any]` implement `io.Reader`?**

The Examples section of the proposal gives the answer directly:

> ```
> type Reader struct{ … }
> func (*Reader) Read[E any]([]E) (int, error) { … }
> ```
> does not implement `io.Reader`, even though it might if there were some way to instantiate the method as `(*Reader).Read[byte]` (which there is not, and we are not proposing it).

When compiled, the gotip compiler says the same thing:

```
./main.go:10:20: cannot use (*Reader)(nil) (value of type *Reader) as io.Reader value in variable declaration: *Reader does not implement io.Reader (wrong type for method Read)
                have Read[E any]([]E) (int, error)
                want Read([]byte) (int, error)
```

The interface's `Read` requires a signature with no type parameters, `Read([]byte) (int, error)`, while this `Read` has the signature `Read[E any]([]E) (int, error)`. The signatures therefore differ. It is accurate to say that none of the existing rules for interface satisfaction have changed.

</details>

---

## Question 3: Write and Run It — Add `Map[F any]` to `Slice[T]`

Now that you understand the behavior from primary sources, complete the `Slice[T].Map[F any]` from the introduction and run it with `gotip` or in the Playground (`?v=gotip`). Also check whether type inference works, so that the caller does not need to specify the type argument explicitly.

<details>
<summary>Hint</summary>

- To try it locally, run `go install golang.org/dl/gotip@latest && gotip download`, followed by `gotip run .`.
- Add `?v=gotip` to the Playground URL, or select “Dev branch” from the Go version selector in the Playground.
- Set `go 1.27` in `go.mod` (`go 1.26` will produce the error from Question 1).
- At the call site, you can write `s.Map(func(n int) string { ... })` (type argument `F` is inferred from the function literal).

</details>

<details>
<summary>Answer</summary>

**Investigation route**

1. Recall the EBNF from Question 1 and put `[F any]` immediately after the method name.
2. Run `gotip run .` locally or run it with Dev/gotip in the Playground, then check the output.
3. Type inference works as described in [Type inference in the specification](https://go.dev/ref/spec#Type_inference). `F` is inferred from the function argument.

**Answer**

```go
package main

import (
	"fmt"
	"strings"
)

type Slice[T any] struct {
	items []T
}

func (s Slice[T]) Map[F any](f func(T) F) Slice[F] {
	out := Slice[F]{items: make([]F, len(s.items))}
	for i, v := range s.items {
		out.items[i] = f(v)
	}
	return out
}

func (s Slice[T]) Items() []T { return s.items }

func main() {
	nums := Slice[int]{items: []int{1, 2, 3}}
	labels := nums.Map(func(n int) string { return fmt.Sprintf("v%d", n) })
	lengths := labels.Map(func(s string) int { return len(s) })

	fmt.Println(strings.Join(labels.Items(), ","))
	fmt.Println(lengths.Items())
}
```

(Run it on the Go Playground: https://go.dev/play/p/ZlBpqqdQJvd?v=gotip )

The output is:

```
v1,v2,v3
[2 2 2]
```

- The receiver's type parameter `T` is already bound when the method is called (`nums` is `Slice[int]`, so `T = int`).
- The method-specific type parameter `F` is inferred from the return type of the passed function literal, `func(int) string`. If you want to be explicit, you can also write `nums.Map[string](...)`.
- The receiver's type parameter list (`Slice[T]`) and the method's type parameter list (`[F any]`) appear in **two stages**, which is the shape of a generic method in Go 1.27.

</details>

---

<details>
<summary>Trivia: <code>Rand.N</code> in the Standard Library</summary>

As a byproduct of the same proposal #77273, a generic method [Rand.N](https://pkg.go.dev/math/rand/v2#Rand.N) was added to the [`Rand` type](https://pkg.go.dev/math/rand/v2#Rand) in math/rand/v2. Previously, only the package-level [rand.N](https://pkg.go.dev/math/rand/v2#N) existed, so calling it from an existing `*Rand` required writing your own wrapper.

```go
package main

import (
	"fmt"
	"math/rand/v2"
)

func main() {
	r := rand.New(rand.NewPCG(1, 2))
	x := r.N(int32(100))
	y := r.N(uint(1000))
	fmt.Printf("int32: %d (type %T)\n", x, x)
	fmt.Printf("uint : %d (type %T)\n", y, y)
}
```

(Run it on the Go Playground: https://go.dev/play/p/HQD1EdC7pL6?v=gotip )

Output:

```
int32: 76 (type int32)
uint : 616 (type uint)
```

The [math/rand/v2 section](https://go.dev/doc/go1.27#minor_library_changes) of the Go 1.27 release notes explicitly says, “`Rand` now supports a generic method `N`, matching the behavior of the top-level `N` function.” If you want to understand “why this was only added in 1.27,” following the thread for proposal #77273 is the quickest route.

</details>

---

## Investigation Starting Points

- Release notes: [Go 1.27 Release Notes #language](https://go.dev/doc/go1.27#language)
- Language specification: [Method declarations](https://go.dev/ref/spec#Method_declarations) / [Type parameter declarations](https://go.dev/ref/spec#Type_parameter_declarations)
- Proposal: [#77273 spec: generic methods for Go](https://go.dev/issue/77273)
- Earlier discussion: [#49085 proposal: spec: allow type parameters in methods](https://go.dev/issue/49085) / [No parameterized methods section of the Type Parameters Proposal](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md#No-parameterized-methods)
- Try it locally: `go install golang.org/dl/gotip@latest && gotip download` / Playground's `?v=gotip`
