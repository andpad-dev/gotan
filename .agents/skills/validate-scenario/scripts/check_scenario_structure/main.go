package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	scenarioPathRE = regexp.MustCompile(
		`^workshop/(01-packages|02-features|03-cmd-tools|04-deep-dive)/` +
			`(01-beginner|02-intermediate|03-advanced)/([^/]+)/README\.md$`,
	)
	questionRE     = regexp.MustCompile(`(?m)^## 設問\s+(\d+)(?::|：|\s)`)
	urlRE          = regexp.MustCompile(`https?://[^\s<>\]})]+`)
	detailRE       = regexp.MustCompile(`(?is)<details\b[^>]*>(.*?)</details>`)
	fenceRE        = regexp.MustCompile(`^\s*(?:>\s*)?(` + "```" + `|~~~)`)
	markdownLinkRE = regexp.MustCompile(`\[[^\]]*\]\(\s*<?([^\s>)]*)>?`)
	goCommandRE    = regexp.MustCompile("`go\\s+(?:run|build|test|vet)\\b[^`\\n]*`")
	titleLineRE    = regexp.MustCompile(`(?m)^#\s+\S.*$`)
)

const (
	routeTrailing = ".,;:!?。 、」』）)]}"
	urlTrailing   = ".,;:!?。 、」』）)]}\"'`"
)

type questionBlock struct {
	number int
	line   int
	text   string
}

type reporter struct {
	writer   io.Writer
	failures int
}

var errHelp = errors.New("help requested")

func parseArgs(args []string) (string, error) {
	parseOptions := true
	var paths []string
	for _, arg := range args {
		switch {
		case parseOptions && arg == "--":
			parseOptions = false
		case parseOptions && (arg == "-h" || arg == "--help"):
			return "", errHelp
		case parseOptions && strings.HasPrefix(arg, "-"):
			return "", fmt.Errorf("unrecognized argument: %s", arg)
		default:
			paths = append(paths, arg)
		}
	}
	if len(paths) != 1 {
		return "", fmt.Errorf("exactly one scenario path is required")
	}
	return paths[0], nil
}

func (r *reporter) emit(level, message string, line int) {
	location := ""
	if line > 0 {
		location = fmt.Sprintf("line %d: ", line)
	}
	fmt.Fprintf(r.writer, "%s\t%s%s\n", level, location, message)
	if level == "FAIL" {
		r.failures++
	}
}

func lineNumber(text string, offset int) int {
	return strings.Count(text[:offset], "\n") + 1
}

func questionBlocks(text string) []questionBlock {
	matches := questionRE.FindAllStringSubmatchIndex(text, -1)
	blocks := make([]questionBlock, 0, len(matches))
	for index, match := range matches {
		end := len(text)
		if index+1 < len(matches) {
			end = matches[index+1][0]
		}
		number, _ := strconv.Atoi(text[match[2]:match[3]])
		blocks = append(blocks, questionBlock{
			number: number,
			line:   lineNumber(text, match[0]),
			text:   text[match[0]:end],
		})
	}
	return blocks
}

func firstRouteURL(block string) string {
	routeIndex := strings.Index(block, "調査ルート")
	if routeIndex < 0 {
		return ""
	}
	route := block[routeIndex+len("調査ルート"):]
	answerRE := regexp.MustCompile(`(?:\*\*答え\*\*|<summary>答え</summary>)`)
	if answer := answerRE.FindStringIndex(route); answer != nil {
		route = route[:answer[0]]
	}
	match := urlRE.FindString(route)
	return strings.TrimRight(match, routeTrailing)
}

func firstURL(text string) string {
	return strings.TrimRight(urlRE.FindString(text), urlTrailing)
}

func resolvedPath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(absolute)
}

func localMarkdownTargets(sourcePath, text string) map[string]struct{} {
	targets := make(map[string]struct{})
	for _, match := range markdownLinkRE.FindAllStringSubmatchIndex(text, -1) {
		if match[0] > 0 && text[match[0]-1] == '!' {
			continue
		}
		value := text[match[2]:match[3]]
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.Path == "" {
			continue
		}
		decodedPath, err := url.PathUnescape(parsed.Path)
		if err != nil {
			continue
		}
		target := decodedPath
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(sourcePath), target)
		}
		info, statErr := os.Stat(target)
		if (statErr == nil && info.IsDir()) || strings.HasSuffix(parsed.Path, "/") {
			target = filepath.Join(target, "README.md")
		}
		targets[resolvedPath(target)] = struct{}{}
	}
	return targets
}

func formatNumbers(numbers []int) string {
	parts := make([]string, len(numbers))
	for index, number := range numbers {
		parts[index] = strconv.Itoa(number)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func containsMarkdownHeading(content string) bool {
	inFence := false
	for _, line := range strings.Split(content, "\n") {
		if fenceRE.MatchString(line) {
			inFence = !inFence
			continue
		}
		if !inFence && regexp.MustCompile(`^#{1,6}\s`).MatchString(line) {
			return true
		}
	}
	return false
}

func localPackageCommands(text string) []string {
	commands := make(map[string]struct{})
	for _, quoted := range goCommandRE.FindAllString(text, -1) {
		command := strings.Trim(quoted, "`")
		fields := strings.Fields(command)
		for _, field := range fields {
			if field == "." || field == "./..." {
				commands[command] = struct{}{}
				break
			}
		}
	}
	result := make([]string, 0, len(commands))
	for command := range commands {
		result = append(result, command)
	}
	sort.Strings(result)
	return result
}

func readText(path string) (string, error) {
	content, err := os.ReadFile(path)
	return string(content), err
}

func host(value string) string {
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func run(args []string, stdout, stderr io.Writer, cwd string) int {
	rawPath, err := parseArgs(args)
	if err != nil {
		if errors.Is(err, errHelp) {
			fmt.Fprintln(stdout, "usage: check_scenario_structure scenario")
			return 0
		}
		fmt.Fprintf(stderr, "check_scenario_structure: %v\n", err)
		return 2
	}

	path := resolvedPath(rawPath)
	relative, err := filepath.Rel(cwd, path)
	if err != nil || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		relative = filepath.ToSlash(rawPath)
	} else {
		relative = filepath.ToSlash(relative)
	}

	report := reporter{writer: stdout}
	match := scenarioPathRE.FindStringSubmatch(relative)
	if match == nil {
		report.emit("FAIL", "path must be workshop/<category>/<difficulty>/<topic>/README.md", 0)
		return 1
	}

	categoryReadme := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(path))), "README.md")
	categoryText := ""
	if info, statErr := os.Stat(categoryReadme); statErr != nil || !info.Mode().IsRegular() {
		report.emit("FAIL", fmt.Sprintf("category README is missing: %s", categoryReadme), 0)
	} else {
		report.emit("PASS", fmt.Sprintf("category README exists: %s", categoryReadme), 0)
		categoryText, err = readText(categoryReadme)
		if err != nil {
			report.emit("FAIL", fmt.Sprintf("cannot read category README: %v", err), 0)
		} else {
			categoryURL := firstURL(categoryText)
			switch {
			case categoryURL == "":
				report.emit("FAIL", "category README has no external starting URL", 0)
			case host(categoryURL) != "go.dev":
				report.emit("FAIL", fmt.Sprintf("category README starts at %s; first URL must be go.dev", categoryURL), 0)
			default:
				report.emit("PASS", fmt.Sprintf("category README starts at go.dev: %s", categoryURL), 0)
			}
			hasSearch := strings.Contains(categoryText, "検索") ||
				strings.Contains(categoryText, "ページ内") ||
				regexp.MustCompile(`\bf\b`).MatchString(categoryText)
			if !strings.Contains(categoryText, "逆引き") || !hasSearch {
				report.emit("FAIL", "category README lacks a usable reverse-search procedure", 0)
			}
		}
	}

	text, err := readText(path)
	if err != nil {
		report.emit("FAIL", fmt.Sprintf("cannot read scenario: %v", err), 0)
		return 1
	}
	titleRE := regexp.MustCompile(`(?m)^#\s+\S`)
	title := titleRE.FindStringIndex(text)
	if title == nil {
		report.emit("FAIL", "scenario title heading is missing", 0)
	} else {
		topTargets := localMarkdownTargets(path, text[:title[0]])
		workshopDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(path))))
		workshopReadme := filepath.Join(workshopDir, "README.md")
		scenarioIndex := filepath.Join(workshopDir, "SCENARIOS.md")
		required := []struct {
			path    string
			failure string
			success string
		}{
			{workshopReadme, "link to workshop/README.md must appear before the scenario title", "top navigation links to workshop/README.md"},
			{scenarioIndex, "link to workshop/SCENARIOS.md must appear before the scenario title", "top navigation links to workshop/SCENARIOS.md"},
			{categoryReadme, "link to the category README must appear before the scenario title", "top navigation links to the category README"},
		}
		for _, item := range required {
			if _, ok := topTargets[resolvedPath(item.path)]; !ok {
				report.emit("FAIL", item.failure, 0)
			} else {
				report.emit("PASS", item.success, 0)
			}
		}

		if info, statErr := os.Stat(scenarioIndex); statErr != nil || !info.Mode().IsRegular() {
			report.emit("FAIL", fmt.Sprintf("scenario index is missing: %s", scenarioIndex), 0)
		} else {
			indexText, readErr := readText(scenarioIndex)
			if readErr != nil {
				report.emit("FAIL", fmt.Sprintf("cannot read scenario index: %v", readErr), 0)
			} else {
				indexTargets := localMarkdownTargets(scenarioIndex, indexText)
				if _, ok := indexTargets[path]; !ok {
					report.emit("FAIL", "scenario is missing from workshop/SCENARIOS.md", 0)
				} else {
					report.emit("PASS", "scenario is linked from workshop/SCENARIOS.md", 0)
				}
			}
		}
	}

	if !strings.Contains(text, "<summary>調査の入り口</summary>") {
		report.emit("FAIL", "調査の入り口 block is missing", 0)
	}

	questionMatches := questionRE.FindAllStringSubmatchIndex(text, -1)
	questionNumbers := make([]int, 0, len(questionMatches))
	for _, match := range questionMatches {
		number, _ := strconv.Atoi(text[match[2]:match[3]])
		questionNumbers = append(questionNumbers, number)
	}
	contiguous := len(questionNumbers) > 0
	for index, number := range questionNumbers {
		if number != index+1 {
			contiguous = false
			break
		}
	}
	switch {
	case len(questionNumbers) == 0:
		report.emit("FAIL", "no ## 設問 N section found", 0)
	case !contiguous:
		report.emit("FAIL", fmt.Sprintf("question numbers are not contiguous from 1: %s", formatNumbers(questionNumbers)), 0)
	default:
		report.emit("PASS", fmt.Sprintf("question sections found: %s", formatNumbers(questionNumbers)), 0)
	}

	introductionEnd := len(text)
	if len(questionMatches) > 0 {
		introductionEnd = questionMatches[0][0]
	}
	introduction := titleLineRE.ReplaceAllString(text[:introductionEnd], "")
	if strings.TrimSpace(introduction) == "" {
		report.emit("FAIL", "scenario introduction is missing", 0)
	} else {
		report.emit("PASS", "scenario introduction is present; review its natural context manually", 0)
	}

	for _, block := range questionBlocks(text) {
		if !strings.Contains(block.text, "<summary>答え</summary>") &&
			!strings.Contains(block.text, "**答え**") {
			report.emit("FAIL", fmt.Sprintf("question %d has no answer section", block.number), block.line)
		}
		if !strings.Contains(block.text, "**調査ルート**") {
			report.emit("FAIL", fmt.Sprintf("question %d has no 調査ルート", block.number), block.line)
		}
		routeURL := firstRouteURL(block.text)
		switch {
		case routeURL == "":
			report.emit("FAIL", fmt.Sprintf("question %d has no URL in 調査ルート", block.number), block.line)
		case host(routeURL) != "go.dev":
			report.emit("FAIL", fmt.Sprintf("question %d route starts at %s; first URL must be go.dev", block.number, routeURL), block.line)
		default:
			report.emit("PASS", fmt.Sprintf("question %d route starts at go.dev: %s", block.number, routeURL), block.line)
		}
	}

	entryStart := strings.Index(text, "<summary>調査の入り口</summary>")
	if entryStart >= 0 {
		entry := text[entryStart:]
		if next := strings.Index(entry, "</details>"); next >= 0 {
			entry = entry[:next]
		}
		entryURL := firstURL(entry)
		switch {
		case entryURL == "":
			report.emit("FAIL", "調査の入り口 has no external starting URL", 0)
		case host(entryURL) != "go.dev":
			report.emit("FAIL", fmt.Sprintf("調査の入り口 starts at %s; first URL must be go.dev", entryURL), 0)
		}
	}

	for _, detail := range detailRE.FindAllStringSubmatchIndex(text, -1) {
		content := text[detail[2]:detail[3]]
		if containsMarkdownHeading(content) {
			report.emit("FAIL", "Markdown heading found inside <details>", lineNumber(text, detail[0]))
		}
	}

	if strings.Contains(text, "```go") && !strings.Contains(text, "https://go.dev/play/p/") {
		report.emit("WARN", "Go code block exists without a Go Playground link; verify whether it is runnable", 0)
	} else if strings.Contains(text, "```go") {
		report.emit("WARN", "Go code blocks found; compare each block with its individual Playground link", 0)
	}

	commands := localPackageCommands(text)
	if len(commands) > 0 {
		goFiles, _ := filepath.Glob(filepath.Join(filepath.Dir(path), "*.go"))
		if len(goFiles) > 0 {
			report.emit("PASS", fmt.Sprintf("local package commands have %d sibling .go file(s)", len(goFiles)), 0)
		} else {
			report.emit(
				"WARN",
				"local package command is shown but the scenario directory has no .go files; "+
					"verify complete save/setup steps: "+strings.Join(commands, ", "),
				0,
			)
		}
	}

	if strings.Contains(text, "<summary>ヒント</summary>") || strings.Contains(text, "### ヒント") {
		report.emit("WARN", "hint leakage still requires manual comparison against the answer", 0)
	}

	if report.failures > 0 {
		return 1
	}
	return 0
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, cwd))
}
