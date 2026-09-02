package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

var (
	markdownLinkRE = regexp.MustCompile(`(?m)\[[^\]]*\]\(\s*<?([^\s>)]*)>?`)
	autolinkRE     = regexp.MustCompile(`<((?:https?://)[^>]+)>`)
	htmlHrefRE     = regexp.MustCompile(`(?i)\bhref\s*=\s*["']([^"']+)["']`)
	bareURLRE      = regexp.MustCompile(`https?://[^\s<>\]})]+`)
	inlineCodeRE   = regexp.MustCompile("`+[^`]*`+")
	fenceRE        = regexp.MustCompile(`^\s*(?:>\s*)?(` + "```" + `|~~~)`)
	htmlAnchorRE   = regexp.MustCompile(`(?i)(?:id|name)\s*=\s*["']([^"']+)["']`)
	headingRE      = regexp.MustCompile(`(?m)^#{1,6}\s+(.+?)\s*#*\s*$`)
)

const trailingPunctuation = ".,;:!?。 、」』）)]}\"'`"

var inconclusiveStatuses = map[int]bool{http.StatusForbidden: true, http.StatusTooManyRequests: true}

type link struct {
	value string
	line  int
	kind  string
}

type options struct {
	network bool
	timeout time.Duration
	paths   []string
}

var errHelp = errors.New("help requested")

func parseArgs(args []string) (options, error) {
	opts := options{timeout: 15 * time.Second}
	parseOptions := true
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case parseOptions && arg == "--":
			parseOptions = false
		case parseOptions && (arg == "-h" || arg == "--help"):
			return options{}, errHelp
		case parseOptions && arg == "--network":
			opts.network = true
		case parseOptions && arg == "--timeout":
			if index+1 >= len(args) {
				return options{}, fmt.Errorf("--timeout requires a value")
			}
			index++
			timeout, err := parseTimeout(args[index])
			if err != nil {
				return options{}, err
			}
			opts.timeout = timeout
		case parseOptions && strings.HasPrefix(arg, "--timeout="):
			timeout, err := parseTimeout(strings.TrimPrefix(arg, "--timeout="))
			if err != nil {
				return options{}, err
			}
			opts.timeout = timeout
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

func parseTimeout(value string) (time.Duration, error) {
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid --timeout value %q", value)
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("timeout must be greater than zero")
	}
	return time.Duration(seconds * float64(time.Second)), nil
}

func stripTrailingPunctuation(value string) string {
	return strings.TrimRightFunc(value, func(r rune) bool {
		return strings.ContainsRune(trailingPunctuation, r)
	})
}

func maskInlineCode(line string) string {
	return inlineCodeRE.ReplaceAllStringFunc(line, func(match string) string {
		return strings.Repeat(" ", len(match))
	})
}

func extractLinks(content io.Reader) ([]link, error) {
	var links []link
	seen := make(map[string]struct{})
	inFence := false
	data, err := io.ReadAll(content)
	if err != nil {
		return nil, err
	}
	for index, lineText := range strings.Split(string(data), "\n") {
		lineNumber := index + 1
		lineText = strings.TrimSuffix(lineText, "\r")
		if fenceRE.MatchString(lineText) {
			inFence = !inFence
			continue
		}

		syntaxLine := ""
		if !inFence {
			syntaxLine = maskInlineCode(lineText)
		}
		candidates := appendMarkdownLinks(nil, syntaxLine)
		candidates = appendMatches(candidates, autolinkRE, syntaxLine, "autolink", false)
		candidates = appendMatches(candidates, htmlHrefRE, syntaxLine, "html", false)
		candidates = appendMatches(candidates, bareURLRE, lineText, "bare", true)
		for _, candidate := range candidates {
			if candidate.value == "" {
				continue
			}
			key := fmt.Sprintf("%s\x00%d", candidate.value, lineNumber)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			candidate.line = lineNumber
			links = append(links, candidate)
		}
	}
	return links, nil
}

func appendMarkdownLinks(dst []link, text string) []link {
	for _, match := range markdownLinkRE.FindAllStringSubmatchIndex(text, -1) {
		if match[0] > 0 && text[match[0]-1] == '!' {
			continue
		}
		dst = append(dst, link{value: text[match[2]:match[3]], kind: "markdown"})
	}
	return dst
}

func appendMatches(dst []link, pattern *regexp.Regexp, text, kind string, wholeMatch bool) []link {
	for _, match := range pattern.FindAllStringSubmatch(text, -1) {
		value := match[0]
		if !wholeMatch {
			value = match[1]
		}
		if kind == "bare" {
			value = stripTrailingPunctuation(value)
		}
		dst = append(dst, link{value: value, kind: kind})
	}
	return dst
}

func githubSlug(value string) string {
	value = cases.Fold().String(norm.NFKC.String(value))
	var result strings.Builder
	pendingHyphen := false
	for _, r := range value {
		switch {
		case unicode.IsPunct(r):
			continue
		case unicode.IsSpace(r):
			pendingHyphen = result.Len() > 0
		default:
			if pendingHyphen {
				result.WriteByte('-')
				pendingHyphen = false
			}
			result.WriteRune(r)
		}
	}
	return strings.Trim(result.String(), "-")
}

func localAnchorExists(path, fragment string) (bool, bool) {
	content, err := os.ReadFile(path)
	if err != nil || !utf8.Valid(content) {
		return false, false
	}
	text := string(content)
	for _, match := range htmlAnchorRE.FindAllStringSubmatch(text, -1) {
		if strings.EqualFold(match[1], fragment) {
			return true, true
		}
	}
	for _, match := range headingRE.FindAllStringSubmatch(text, -1) {
		if githubSlug(match[1]) == fragment {
			return true, true
		}
	}
	return false, false
}

func checkLocal(sourcePath string, item link) (string, string) {
	parsed, err := url.Parse(item.value)
	if err != nil {
		return "FAIL", fmt.Sprintf("invalid URL: %v", err)
	}
	if parsed.Scheme != "" || strings.HasPrefix(item.value, "//") {
		return "SKIP", "external"
	}

	target := sourcePath
	if parsed.Path != "" {
		decodedPath, decodeErr := url.PathUnescape(parsed.Path)
		if decodeErr != nil {
			return "FAIL", fmt.Sprintf("invalid path: %v", decodeErr)
		}
		target = decodedPath
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(sourcePath), target)
	}
	target, err = filepath.Abs(target)
	if err != nil {
		return "FAIL", fmt.Sprintf("resolve target: %v", err)
	}
	target = filepath.Clean(target)
	if _, err := os.Stat(target); err != nil {
		if os.IsNotExist(err) {
			return "FAIL", fmt.Sprintf("missing target: %s", target)
		}
		return "FAIL", fmt.Sprintf("inspect target: %v", err)
	}
	if parsed.Fragment == "" {
		return "PASS", fmt.Sprintf("target exists: %s", target)
	}

	exists, conclusive := localAnchorExists(target, parsed.Fragment)
	if exists {
		return "PASS", fmt.Sprintf("target and anchor exist: #%s", parsed.Fragment)
	}
	if !conclusive {
		return "INCONCLUSIVE", fmt.Sprintf("target exists; inspect anchor manually: #%s", parsed.Fragment)
	}
	return "FAIL", fmt.Sprintf("anchor not found: #%s", parsed.Fragment)
}

func checkHTTP(ctx context.Context, client *http.Client, value string) (string, string) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, value, nil)
	if err != nil {
		return "FAIL", fmt.Sprintf("invalid URL: %v", err)
	}
	request.Header.Set("User-Agent", "gotan-validate-scenario/1.0")

	response, err := client.Do(request)
	if err != nil {
		return "INCONCLUSIVE", fmt.Sprintf("network request could not be completed: %v", err)
	}
	defer response.Body.Close()

	status := response.StatusCode
	finalURL := response.Request.URL.String()
	if status >= 200 && status < 300 {
		detail := fmt.Sprintf("HTTP %d", status)
		if finalURL != value {
			detail += " -> " + finalURL
		}
		return "PASS", detail
	}
	if inconclusiveStatuses[status] {
		return "INCONCLUSIVE", fmt.Sprintf("HTTP %d", status)
	}
	return "FAIL", fmt.Sprintf("HTTP %d -> %s", status, finalURL)
}

func run(args []string, stdout, stderr io.Writer, client *http.Client) int {
	opts, err := parseArgs(args)
	if err != nil {
		if errors.Is(err, errHelp) {
			fmt.Fprintln(stdout, "usage: check_markdown_links [--network] [--timeout seconds] paths...")
			return 0
		}
		fmt.Fprintf(stderr, "check_markdown_links: %v\n", err)
		return 2
	}

	failures := 0
	inconclusive := 0
	for _, rawPath := range opts.paths {
		absolutePath, err := filepath.Abs(rawPath)
		if err != nil {
			fmt.Fprintf(stderr, "FAIL\t%s: %v\n", rawPath, err)
			failures++
			continue
		}
		file, err := os.Open(absolutePath)
		if err != nil {
			fmt.Fprintf(stderr, "FAIL\t%s: file does not exist\n", rawPath)
			failures++
			continue
		}
		links, extractErr := extractLinks(file)
		file.Close()
		if extractErr != nil {
			fmt.Fprintf(stderr, "FAIL\t%s: %v\n", rawPath, extractErr)
			failures++
			continue
		}

		for _, item := range links {
			parsed, parseErr := url.Parse(item.value)
			result, detail := "", ""
			if parseErr == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") {
				if !opts.network {
					result, detail = "SKIP", "external; rerun with --network"
				} else {
					ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
					result, detail = checkHTTP(ctx, client, item.value)
					cancel()
				}
			} else {
				result, detail = checkLocal(absolutePath, item)
			}
			fmt.Fprintf(stdout, "%s\t%s:%d\t%s\t%s\t%s\n", result, absolutePath, item.line, item.kind, item.value, detail)
			switch result {
			case "FAIL":
				failures++
			case "INCONCLUSIVE":
				inconclusive++
			}
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

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, http.DefaultClient))
}
