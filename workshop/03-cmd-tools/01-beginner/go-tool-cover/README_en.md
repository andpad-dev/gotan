[Workshop guide and scenario index](../../../README.md) | [How to research 03-cmd-tools](../../README.md)

# Check test coverage

You saw `coverage: 66.7%` in the results of a CI test. Let’s investigate what it means.

```
PASS
coverage: 66.7% of statements
```

This example tests a `fare` function that calculates a fare based on the number of passengers. `fare` has a branch that applies a discount when there are 10 or more passengers, but this number alone does not tell you whether that branch ran.
Use `go tool cover` to find out how much has been exercised and what has not yet been reached.

**`main.go`**

```go
package main

import "fmt"

func fare(passengers int) int {
	if passengers <= 0 {
		return 0
	}
	if passengers >= 10 {
		return passengers * 800
	}
	return passengers * 1000
}

func main() {
	fmt.Println(fare(3))
}
```

(Run it in the [Go Playground](https://go.dev/play/p/Ew0R-EaEA2s))

```console
$ go run main.go
3000
```

**`main_test.go`**

```go
package main

import "testing"

func TestFare(t *testing.T) {
	tests := []struct {
		passengers int
		want       int
	}{
		{passengers: 0, want: 0},
		{passengers: 3, want: 3000},
	}

	for _, tt := range tests {
		if got := fare(tt.passengers); got != tt.want {
			t.Fatalf("fare(%d) = %d, want %d", tt.passengers, got, tt.want)
		}
	}
}
```

These are the results of running the following commands with Go 1.26.4.

```console
$ go test -coverprofile=cover.out
PASS
coverage: 66.7% of statements
ok  	example.com/cover-demo	0.226s

$ go tool cover -func=cover.out
example.com/cover-demo/main.go:5:	fare		80.0%
example.com/cover-demo/main.go:15:	main		0.0%
total:					(statements)	66.7%
```

---

## Question 1: Whose statements do 66.7% and 80.0% count?

`go test` succeeds, yet the total is 66.7%, while `fare` alone is 80.0%.

1. What does `-coverprofile=cover.out` create, and what does `go tool cover -func=cover.out` aggregate?
2. Why can’t you conclude that the fare calculation was exercised only 66.7% because the overall value is 66.7%?

<details>
<summary>Hint</summary>

- First open the [Go command list](https://go.dev/doc/cmd) and look for `cover`.
- Use `go help testflag` locally to check the explanation of `-coverprofile`.
- Then run `go tool cover -help` and determine what `-func` accepts.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Find `cover` in the [Go command list](https://go.dev/doc/cmd) and open the [official cover documentation](https://go.dev/cmd/cover/). You will learn that `cover` analyzes the coverage profile produced by `go test -coverprofile=cover.out`.
2. Run `go help testflag` locally and confirm that `-coverprofile` writes a coverage profile after the tests succeed.
3. Run `go tool cover -help` and use `-func` to display coverage by function. Also read the implementation notes in the [Go 1.26.4 `cmd/cover` source](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/cover/doc.go).

**Answer**

- `go test -coverprofile=cover.out` writes information about statements reached during the test run to `cover.out`. `go tool cover -func=cover.out` aggregates that profile by function and then for the total.
- The **66.7%** in this example is the total for all covered statements in `main.go`. The **80.0%** is the value for `fare` alone. `go test` does not execute `main`, so its 0.0% lowers the total.
- Therefore, the overall percentage is not a pass mark for “test quality”. First use the per-function display to focus on the fare-calculation branch you need to protect.
- `cover` instruments approximate basic blocks after analyzing the source. As the official documentation explains, it does not measure the insides of `&&` and `||` separately. The number is a starting point for investigation, not a substitute for the specification.

</details>

---

## Question 2: Which test should you add to check the 10-or-more branch?

`fare` is at 80.0%. Compare `main_test.go` with the output of `go tool cover -func` and identify the fare-calculation branch that has not been exercised even once.

After adding one test case for that branch, how do you verify it with the following commands? Explain why you can say that the fare-calculation branch is covered even if the overall value does not become 100%.

<details>
<summary>Hint</summary>

- Trace the `if` statements in `fare` from top to bottom and compare them with the test table’s `passengers` values.
- Choose one boundary value to check the condition where the discount begins.
- Use the help to find out what happens when you pass the profile from Question 1 to `-html`.

</details>

<details>
<summary>Answer</summary>

**Investigation path**

1. Read the [official cover documentation](https://go.dev/cmd/cover/) and confirm that coverage represents reached basic blocks.
2. Compare the inputs `0` and `3` in `main_test.go` with the conditions in `fare`. Then rerun `go tool cover -func=cover.out` to confirm the value for `fare`.
3. Check the use of `-html` with `go tool cover -help`, then run `go tool cover -html=cover.out -o cover.html` to inspect unreached code in a browser.

**Answer**

- The unreached branch is the discounted fare for `passengers >= 10`. To check the boundary where the discount begins, add this case to the table:

    ```go
    {passengers: 10, want: 8000},
    ```

- Verify it in this order:

    ```console
    $ go test -coverprofile=cover.out
    $ go tool cover -func=cover.out
    $ go tool cover -html=cover.out -o cover.html
    ```

    `-html` turns the profile into color-coded HTML so you can inspect reached and unreached source lines.
- Even if `fare` reaches 100.0%, `main` is not called during the test run, so the total does not become 100%. The important point is that the `fare` specification, including the discount branch, has been tested with an input and expected value. Judge coverage against the rules that matter, not only against the aggregate number.

</details>

---

<details>
<summary>Trivia</summary>

When you want to measure execution across multiple packages, use `-coverpkg` with `go test` to specify the packages to instrument. First read `-coverpkg` and `-covermode` in `go help testflag`, decide which packages should count, and then expand the scope. Increasing the target only to make the number larger can hide the branch you actually need to inspect, as in this example.

</details>

---

## Starting points for investigation

1. [Go command list](https://go.dev/doc/cmd)
2. [Official cover documentation](https://go.dev/cmd/cover/)
3. `go help testflag` and `go tool cover -help` locally
4. [Go 1.26.4 `cmd/cover` source](https://cs.opensource.google/go/go/+/refs/tags/go1.26.4:src/cmd/cover/doc.go)
