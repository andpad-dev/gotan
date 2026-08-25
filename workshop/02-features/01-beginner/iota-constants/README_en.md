# Read constant declarations using iota

You came across the following constant declaration in a senior colleague’s code.

```go
type Color int

const (
    Red Color = iota
    Green
    Blue
)
```

(Run it in the Go Playground: https://go.dev/play/p/mBjgkLshasU )

`fmt.Println(Red, Green, Blue)` prints `0 1 2`.
What is `iota`, and why do the values change even when the declarations on the second and following lines are omitted? Let’s investigate.

In Question 1, focus first on what `iota` is: a function, keyword, or identifier usable as a constant. Question 2 covers why the expressions for `Green` and `Blue` may be omitted.

## Question 1: What is `iota`?

Use the language specification to investigate the identity of `iota`, which appears only on the first `Red` declaration.

Think of each line in the `const` block as a position that determines the value of `iota`, and predict how the value changes on the first and third lines.

<details>
<summary>Hint</summary>

- Because the Go language specification is one HTML page, the fastest way to find `iota` is often the `f` key for in-page search.
- There is an `Iota` heading inside the “Constant declarations” section.
- Read [Iota](https://go.dev/ref/spec#Iota) and [Predeclared identifiers](https://go.dev/ref/spec#Predeclared_identifiers) in order, and also check whether `iota` appears in the keyword list.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open https://go.dev/ref/spec and search the page for `Iota`.
2. Read the `Iota` section inside “Constant declarations”.

**Answer**

`iota` is a **predeclared constant that can be used only within a constant declaration (`ConstDecl`)**. It is neither a function nor a reserved keyword.

The [Iota](https://go.dev/ref/spec#Iota) section of the specification says:

> Within a constant declaration, the predeclared identifier `iota` represents successive untyped integer constants. Its value is the index of the respective ConstSpec in that constant declaration, starting at zero.

There are two key points:

- Its value is an **untyped integer constant**.
- Its value is the **index of the ConstSpec** in the `const ( ... )` block, here meaning the line’s constant declaration, starting at 0. The first line is 0, the second is 1, the third is 2, and so on.

Therefore, at `Red Color = iota`, `Red = 0` is established and given type `Color`.

</details>

---

## Question 2: Why can `Green` and `Blue` omit their expressions?

`Green` and `Blue` do not even contain `= iota`, so why do they receive the values `1` and `2`? The behavior of `iota` alone is not enough to explain this. There is another rule for constant declarations.

Compare the abbreviated `Red = iota / Green / Blue` with `Red = iota / Green = iota / Blue = iota`. This separates whether omission directly increments a value or merely repeats an expression.

<details>
<summary>Hint</summary>

- The “Constant declarations” section of the specification explains what happens when an expression list is omitted.
- The rule is similar to “repeat the same expression list”.
- Compare `A = 10 / B / C` with `Red = iota / Green / Blue`, and predict the result of repeating the preceding expression.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Open https://go.dev/ref/spec#Constant_declarations
2. Read the paragraph beginning “Within a parenthesized `const` declaration list ...”.

**Answer**

Omission is not syntax for “increase the value by one”. It repeats the previous expression list, and `iota` in that repeated expression takes the value for the current position.

The [Constant declarations](https://go.dev/ref/spec#Constant_declarations) section gives this rule:

> Within a parenthesized `const` declaration list the expression list may be omitted from any but the first ConstSpec. Such an empty list is equivalent to the textual substitution of the first preceding non-empty expression list and its type if any.

This is called **implicit repetition**.

- In a parenthesized `const ( ... )` list, later `ConstSpec` entries may omit the entire expression list and type.
- When omitted, the **most recent non-omitted expression list and type are copied textually as-is**.

Therefore,

```go
const (
    Red Color = iota
    Green
    Blue
)
```

is equivalent from the compiler’s perspective to:

```go
const (
    Red   Color = iota
    Green Color = iota
    Blue  Color = iota
)
```

Since `iota` is the index of the `ConstSpec`, it evaluates to 0 for `Red`, 1 for `Green`, and 2 for `Blue`.
The two rules, implicit repetition and the index behavior of `iota`, combine to form Go’s familiar enum idiom.

</details>

---

<details>
<summary>Trivia: Bit flags with <code>1 &lt;&lt; iota</code></summary>

Because `iota` can be used in an expression, writing `1 << iota` creates bit flags.

```go
type Perm uint

const (
    Read Perm = 1 << iota
    Write
    Execute
)
```

(Run it in the Go Playground: https://go.dev/play/p/QrVksK9QllL )

`fmt.Println(Read, Write, Execute, Read|Write)` outputs `1 2 4 3`.

With implicit repetition, what is copied to the second and following lines is the **expression list itself** (`1 << iota`), not its value.
`iota` is reevaluated on each line, producing `1<<0=1`, `1<<1=2`, and `1<<2=4`.

The [Constants section of Effective Go](https://go.dev/doc/effective_go#constants) also introduces a `ByteSize` example using `1 << iota` for KB, MB, GB, and so on.

</details>

---

## Starting points for investigation

- https://go.dev/ref/spec
- https://go.dev/doc/effective_go
