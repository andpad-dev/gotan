package main

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var playgroundRE = regexp.MustCompile(`https://go\.dev/play/p/([A-Za-z0-9_-]+)(?:\?[^\s)>]+)?`)

type options struct {
	network bool
	paths   []string
}

type occurrence struct {
	path string
	line int
	url  string
	id   string
}

var errHelp = errors.New("help requested")

func parseArgs(args []string) (options, error) {
	var opts options
	parseOptions := true
	for _, arg := range args {
		switch {
		case parseOptions && arg == "--":
			parseOptions = false
		case parseOptions && (arg == "-h" || arg == "--help"):
			return options{}, errHelp
		case parseOptions && arg == "--network":
			opts.network = true
		case parseOptions && strings.HasPrefix(arg, "-"):
			return options{}, fmt.Errorf("unrecognized argument: %s", arg)
		default:
			opts.paths = append(opts.paths, arg)
		}
	}
	if len(opts.paths) == 0 {
		return options{}, fmt.Errorf("at least one Markdown path is required")
	}
	return opts, nil
}

func findOccurrences(path string, content io.Reader) ([]occurrence, error) {
	var occurrences []occurrence
	seen := make(map[string]struct{})
	data, err := io.ReadAll(content)
	if err != nil {
		return nil, err
	}
	for lineNumber, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSuffix(line, "\r")
		for _, match := range playgroundRE.FindAllStringSubmatch(line, -1) {
			key := fmt.Sprintf("%s\x00%d\x00%s", path, lineNumber, match[0])
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			occurrences = append(occurrences, occurrence{
				path: path,
				line: lineNumber + 1,
				url:  match[0],
				id:   match[1],
			})
		}
	}
	return occurrences, nil
}

func checkSource(ctx context.Context, client *http.Client, item occurrence) (string, error) {
	sourceURL := fmt.Sprintf("https://go.dev/play/p/%s.go", item.id)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "gotan-validate-scenario")

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	source, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d", response.StatusCode)
	}

	digest := sha256.Sum256(source)
	return fmt.Sprintf(
		"source HTTP %d\tsha256=%x\tbytes=%d",
		response.StatusCode,
		digest[:6],
		len(source),
	), nil
}

func run(args []string, stdout, stderr io.Writer, client *http.Client) int {
	opts, err := parseArgs(args)
	if err != nil {
		if errors.Is(err, errHelp) {
			fmt.Fprintln(stdout, "usage: check_playground_sources [--network] paths...")
			return 0
		}
		fmt.Fprintf(stderr, "check_playground_sources: %v\n", err)
		return 2
	}

	failures := 0
	inconclusive := 0
	for _, path := range opts.paths {
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(stderr, "FAIL\t%s: %v\n", path, err)
			failures++
			continue
		}
		occurrences, scanErr := findOccurrences(path, file)
		file.Close()
		if scanErr != nil {
			fmt.Fprintf(stderr, "FAIL\t%s: %v\n", path, scanErr)
			failures++
			continue
		}

		for _, item := range occurrences {
			if !opts.network {
				fmt.Fprintf(stdout, "SKIP\t%s:%d\t%s\tnetwork check disabled\n", item.path, item.line, item.url)
				inconclusive++
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			detail, checkErr := checkSource(ctx, client, item)
			cancel()
			if checkErr != nil {
				fmt.Fprintf(stdout, "FAIL\t%s:%d\t%s\t%s\n", item.path, item.line, item.url, normalizeError(checkErr))
				failures++
				continue
			}
			fmt.Fprintf(stdout, "PASS\t%s:%d\t%s\t%s\n", item.path, item.line, item.url, detail)
		}
	}

	if failures > 0 {
		return 1
	}
	if inconclusive > 0 {
		return 2
	}
	return 0
}

func normalizeError(err error) string {
	return strings.TrimSpace(err.Error())
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, http.DefaultClient))
}
