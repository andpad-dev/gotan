package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestQuestionBlocks(t *testing.T) {
	t.Parallel()

	text := "intro\n## 設問 1: first\none\n## 設問 2：second\ntwo\n"
	got := questionBlocks(text)
	want := []questionBlock{
		{number: 1, line: 2, text: "## 設問 1: first\none\n"},
		{number: 2, line: 4, text: "## 設問 2：second\ntwo\n"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("questionBlocks = %#v, want %#v", got, want)
	}
}

func TestFirstRouteURLStopsBeforeAnswer(t *testing.T) {
	t.Parallel()

	block := "**調査ルート**\n1. https://go.dev/doc/\n**答え**\nhttps://example.com/"
	if got := firstRouteURL(block); got != "https://go.dev/doc/" {
		t.Fatalf("firstRouteURL = %q, want go.dev URL", got)
	}
}

func TestLocalPackageCommands(t *testing.T) {
	t.Parallel()

	text := "`go test ./...`\n`go run .`\n`go test ./pkg`\n"
	want := []string{"go run .", "go test ./..."}
	if got := localPackageCommands(text); !reflect.DeepEqual(got, want) {
		t.Fatalf("localPackageCommands = %#v, want %#v", got, want)
	}
}

func TestFormatNumbers(t *testing.T) {
	t.Parallel()

	if got := formatNumbers([]int{1, 2, 3}); got != "[1, 2, 3]" {
		t.Fatalf("formatNumbers = %q, want Python-style list", got)
	}
}

func TestTitleAloneIsNotAnIntroduction(t *testing.T) {
	t.Parallel()

	text := "# A complete title\n"
	if got := titleLineRE.ReplaceAllString(text, ""); got != "\n" {
		t.Fatalf("title removal left %q, want only newline", got)
	}
}

func TestParseArgsHelp(t *testing.T) {
	t.Parallel()

	if _, err := parseArgs([]string{"--help"}); !errors.Is(err, errHelp) {
		t.Fatalf("parseArgs error = %v, want errHelp", err)
	}
}
