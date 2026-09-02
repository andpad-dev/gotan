package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExtractLinks(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(strings.Join([]string{
		"[local](target.md) ![image](image.png) `https://ignored.example`",
		"<https://go.dev/doc/> https://go.dev/doc/",
		"```",
		"https://inside-fence.example",
		"```",
	}, "\n"))
	got, err := extractLinks(input)
	if err != nil {
		t.Fatal(err)
	}

	want := []link{
		{value: "target.md", line: 1, kind: "markdown"},
		{value: "https://ignored.example", line: 1, kind: "bare"},
		{value: "https://go.dev/doc/", line: 2, kind: "autolink"},
		{value: "https://inside-fence.example", line: 4, kind: "bare"},
	}
	if len(got) != len(want) {
		t.Fatalf("extractLinks returned %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("link %d = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func TestCheckLocal(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "README.md")
	target := filepath.Join(tempDir, "target.md")
	if err := os.WriteFile(source, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("# Target Heading\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, detail := checkLocal(source, link{value: "target.md#target-heading"})
	if result != "PASS" {
		t.Fatalf("checkLocal result = %q, detail = %q; want PASS", result, detail)
	}

	result, detail = checkLocal(source, link{value: "missing.md"})
	if result != "FAIL" || !strings.Contains(detail, "missing target") {
		t.Fatalf("missing check = %q, %q; want missing-target FAIL", result, detail)
	}
}

func TestGitHubSlugNormalizesUnicode(t *testing.T) {
	t.Parallel()

	if got := githubSlug("Ｆｏｏ： Bar"); got != "foo-bar" {
		t.Fatalf("githubSlug = %q, want %q", got, "foo-bar")
	}
}

func TestParseArgsAllowsOptionsAfterPaths(t *testing.T) {
	t.Parallel()

	opts, err := parseArgs([]string{"README.md", "--network", "--timeout", "2.5"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.network || opts.timeout != 2500*time.Millisecond {
		t.Fatalf("options = %#v, want network with 2.5 second timeout", opts)
	}
}

func TestMissingAnchorIsInconclusive(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "README.md")
	if err := os.WriteFile(source, []byte("# Existing\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, _ := checkLocal(source, link{value: "#missing"})
	if result != "INCONCLUSIVE" {
		t.Fatalf("checkLocal result = %q, want INCONCLUSIVE", result)
	}
}

func TestExtractLinksAcceptsLongLines(t *testing.T) {
	t.Parallel()

	input := strings.NewReader(strings.Repeat("x", 70_000) + " https://go.dev/doc/\n")
	links, err := extractLinks(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].value != "https://go.dev/doc/" {
		t.Fatalf("extractLinks = %#v, want one go.dev link", links)
	}
}
