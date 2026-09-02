package main

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestFindOccurrences(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(strings.Join([]string{
		"https://go.dev/play/p/abc_123?v=gotip",
		"https://go.dev/play/p/abc_123?v=gotip https://go.dev/play/p/abc_123?v=gotip",
	}, "\n"))
	got, err := findOccurrences("README.md", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("findOccurrences returned %d items, want 2: %#v", len(got), got)
	}
	if got[0].id != "abc_123" || got[0].line != 1 {
		t.Fatalf("first occurrence = %#v, want id abc_123 on line 1", got[0])
	}
	if got[1].line != 2 {
		t.Fatalf("second occurrence line = %d, want 2", got[1].line)
	}
}

func TestRunWithoutNetworkIsInconclusive(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := tempDir + "/README.md"
	if err := os.WriteFile(path, []byte("https://go.dev/play/p/example\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr strings.Builder
	exitCode := run([]string{path}, &stdout, &stderr, http.DefaultClient)
	if exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr=%q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "SKIP\t"+path+":1") {
		t.Fatalf("stdout = %q, want SKIP result", stdout.String())
	}
}

func TestParseArgsAllowsOptionsAfterPaths(t *testing.T) {
	t.Parallel()

	opts, err := parseArgs([]string{"README.md", "--network"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.network {
		t.Fatalf("options = %#v, want network enabled", opts)
	}
}

func TestFindOccurrencesAcceptsLongLines(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(strings.Repeat("x", 70_000) + " https://go.dev/play/p/example\n")
	occurrences, err := findOccurrences("README.md", input)
	if err != nil {
		t.Fatal(err)
	}
	if len(occurrences) != 1 || occurrences[0].id != "example" {
		t.Fatalf("findOccurrences = %#v, want one Playground URL", occurrences)
	}
}
