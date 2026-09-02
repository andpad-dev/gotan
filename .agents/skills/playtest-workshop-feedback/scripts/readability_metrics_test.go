package main

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ikawaha/kagome/v2/tokenizer"
)

func TestExtractMarkdownExcludesHiddenAndCode(t *testing.T) {
	content := `# 見出し

表示する**文章**です。

<details>
<summary>答え</summary>
隠す文章です。
</details>

` + "```go\nfmt.Println(\"除外\")\n```\n" + `
- 箇条書き
`
	got := extractMarkdown(content)

	if got.totalText != "見出し表示する文章です箇条書き" {
		t.Fatalf("totalText = %q", got.totalText)
	}
	if got.structuralText != "見出し文章箇条書き" {
		t.Fatalf("structuralText = %q", got.structuralText)
	}
	wantSentences := []sentenceMetric{
		{Text: "見出し", Length: 3},
		{Text: "表示する文章です。", Length: 9},
		{Text: "箇条書き", Length: 4},
	}
	if !reflect.DeepEqual(got.sentences, wantSentences) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, wantSentences)
	}
}

func TestExtractMarkdownExcludesBareURLs(t *testing.T) {
	got := extractMarkdown("参照 https://example.com/very/long/path。次へ進む。\n")
	want := []sentenceMetric{
		{Text: "参照。", Length: 3},
		{Text: "次へ進む。", Length: 5},
	}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
	if strings.Contains(got.languageText, "http") || strings.Contains(got.totalText, "example") {
		t.Fatalf("bare URL leaked into extracted text: %#v", got)
	}
}

func TestExtractMarkdownJoinsSoftWrappedSentence(t *testing.T) {
	got := extractMarkdown("この文は次の行へ\n続いて終わります。\n")
	want := []sentenceMetric{{Text: "この文は次の行へ 続いて終わります。", Length: 18}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownJoinsIndentedParagraphContinuation(t *testing.T) {
	got := extractMarkdown("この文は次の行へ\n    続いて終わります。\n")
	want := []sentenceMetric{{Text: "この文は次の行へ 続いて終わります。", Length: 18}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownMeasuresLongListItems(t *testing.T) {
	item := strings.Repeat("長", 61)
	got := extractMarkdown("- " + item + "\n")
	want := []sentenceMetric{{Text: item, Length: 61}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownJoinsListItemContinuation(t *testing.T) {
	got := extractMarkdown("- このリスト項目は次の行へ\n    続いて終わります。\n")
	want := []sentenceMetric{{
		Text:   "このリスト項目は次の行へ 続いて終わります。",
		Length: 22,
	}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
	if got.structuralText != "このリスト項目は次の行へ続いて終わります" {
		t.Fatalf("structuralText = %q", got.structuralText)
	}
}

func TestExtractMarkdownMeasuresNestedListItems(t *testing.T) {
	got := extractMarkdown("- 親項目\n    - 子項目です。\n")
	want := []sentenceMetric{
		{Text: "親項目", Length: 3},
		{Text: "子項目です。", Length: 6},
	}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
	if got.structuralText != "親項目子項目です" {
		t.Fatalf("structuralText = %q", got.structuralText)
	}
}

func TestExtractMarkdownRemovesInlineCodeURL(t *testing.T) {
	got := extractMarkdown("URLは `https://go.dev/play/p/example` の形式です。\n")
	want := []sentenceMetric{{Text: "URLは の形式です。", Length: 11}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
	if strings.ContainsAny(got.languageText, "`") || strings.Contains(got.languageText, "https") {
		t.Fatalf("inline URL leaked into extracted text: %#v", got)
	}
}

func TestExtractMarkdownPreservesHTMLLookingInlineCode(t *testing.T) {
	got := extractMarkdown("値は `<nil>` です。要素は `<details>` です。\n")
	want := []sentenceMetric{
		{Text: "値は <nil> です。", Length: 12},
		{Text: "要素は <details> です。", Length: 17},
	}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownDoesNotSplitInlineCodePunctuation(t *testing.T) {
	got := extractMarkdown("- `?` キーを押します。\n")
	want := []sentenceMetric{{
		Text:    "? キーを押します。",
		Length:  10,
		MDDText: "キーを押します。",
	}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownExcludesIndentedCode(t *testing.T) {
	content := "前です。\n\n    fmt.Println(\"除外\")\n    コードです。\n\n後です。\n"
	got := extractMarkdown(content)
	want := []sentenceMetric{
		{Text: "前です。", Length: 4},
		{Text: "後です。", Length: 4},
	}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownStripsNestedBlockquotePrefixes(t *testing.T) {
	got := extractMarkdown("> > 引用文です。\n")
	want := []sentenceMetric{{Text: "引用文です。", Length: 6}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownExcludesFencedCodeInBlockquote(t *testing.T) {
	content := "> ```go\n> fmt.Println(\"除外\")\n> コードです。\n> ```\n> 表示です。\n"
	got := extractMarkdown(content)
	want := []sentenceMetric{{Text: "表示です。", Length: 5}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownExcludesMultilineHTMLComments(t *testing.T) {
	content := "表示です。\n<!--\nこの説明は表示されません。\n-->\n後です。\n"
	got := extractMarkdown(content)
	want := []sentenceMetric{
		{Text: "表示です。", Length: 5},
		{Text: "後です。", Length: 4},
	}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownExcludesInlineDetails(t *testing.T) {
	content := "前です。 <details><summary>答え</summary>隠し本文</details> 後です。\n"
	got := extractMarkdown(content)
	want := []sentenceMetric{
		{Text: "前です。", Length: 4},
		{Text: "後です。", Length: 4},
	}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownTracksFenceMarkerAndLength(t *testing.T) {
	content := "````markdown\n```go\nコードです。\n```\n<details>\n````go\nまだコードです。\n````\n表示です。\n"
	got := extractMarkdown(content)
	want := []sentenceMetric{{Text: "表示です。", Length: 5}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownTracksTildeFence(t *testing.T) {
	content := "~~~~text\n~~~\nコードです。\n~~~\n~~~~\n表示です。\n"
	got := extractMarkdown(content)
	want := []sentenceMetric{{Text: "表示です。", Length: 5}}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownSkipsTableSeparatorAndMeasuresCells(t *testing.T) {
	content := "項目 | `a\\|b` の説明です。\n--- | ---\n値 | 続きです。\n"
	codeCell := "a|b の説明です。"
	got := extractMarkdown(content)
	want := []sentenceMetric{
		{Text: "項目", Length: 2},
		{Text: codeCell, Length: len([]rune(codeCell))},
		{Text: "値", Length: 1},
		{Text: "続きです。", Length: 5},
	}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
	if got.structuralText != "" {
		t.Fatalf("table text must not count toward STR: %q", got.structuralText)
	}
}

func TestExtractMarkdownSkipsOneColumnTableSeparator(t *testing.T) {
	content := "| 項目 |\n| --- |\n| 値 |\n"
	got := extractMarkdown(content)
	want := []sentenceMetric{
		{Text: "項目", Length: 2},
		{Text: "値", Length: 1},
	}
	if !reflect.DeepEqual(got.sentences, want) {
		t.Fatalf("sentences = %#v, want %#v", got.sentences, want)
	}
}

func TestExtractMarkdownCountsOnlyVisibleBoldLinkText(t *testing.T) {
	got := extractMarkdown("**[公式文書](https://go.dev/doc/)を読む**\n")
	if got.structuralText != "公式文書を読む" {
		t.Fatalf("structuralText = %q", got.structuralText)
	}
	if len([]rune(got.structuralText)) > len([]rune(got.totalText)) {
		t.Fatalf("structural text exceeds total text: %#v", got)
	}
}

func TestRatio(t *testing.T) {
	if got := ratio(1, 3); got != 33.33 {
		t.Fatalf("ratio(1, 3) = %v", got)
	}
	if got := ratio(1, 0); got != 0 {
		t.Fatalf("ratio(1, 0) = %v", got)
	}
}

func TestSplitSentences(t *testing.T) {
	got := splitSentences("一文目です。二文目？ 終端なし")
	want := []sentenceMetric{
		{Text: "一文目です。", Length: 6},
		{Text: "二文目？", Length: 4},
		{Text: "終端なし", Length: 4},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitSentences() = %#v, want %#v", got, want)
	}
}

func TestSplitSentencesRecognizesEnglishFullStops(t *testing.T) {
	got := splitSentences("First sentence. Second sentence. Version 1.22 stays.")
	want := []sentenceMetric{
		{Text: "First sentence.", Length: 15},
		{Text: "Second sentence.", Length: 16},
		{Text: "Version 1.22 stays.", Length: 19},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitSentences() = %#v, want %#v", got, want)
	}
}

func TestBigramProbabilitiesNormalizeAndSurprisalIsDeterministic(t *testing.T) {
	analyzer := mustAnalyzer(t)
	modelA := trainBigram([]string{"猫が走る。猫が眠る。", "犬が走る。"}, analyzer, 0.5)
	modelB := trainBigram([]string{"猫が走る。猫が眠る。", "犬が走る。"}, analyzer, 0.5)

	sum := 0.0
	for _, word := range modelA.vocabulary {
		sum += modelA.probability(beginToken, word)
	}
	if math.Abs(sum-1) > 1e-12 {
		t.Fatalf("probability sum = %.15f", sum)
	}

	target := tokenizeWords(analyzer, "猫が走る。")
	gotA := modelA.averageSurprisal(target)
	gotB := modelB.averageSurprisal(target)
	if gotA != gotB {
		t.Fatalf("surprisal differs: %v != %v", gotA, gotB)
	}
	if gotA <= 0 {
		t.Fatalf("surprisal = %v, want positive", gotA)
	}
}

func TestCalculateCRSEstimate(t *testing.T) {
	got := calculateCRS(3, 70, 40, 4, 20)
	want := 82.25
	if got != want {
		t.Fatalf("calculateCRS() = %v, want %v", got, want)
	}
}

func TestMDDEstimateExcludesRootAndUsesJapaneseParticleRule(t *testing.T) {
	analyzer := mustAnalyzer(t)
	sentences := []sentenceMetric{{Text: "私はGoを使います。", Length: 10}}

	mdd, phrases, dependencies, details := estimateMDD(sentences, analyzer)
	if phrases != 3 {
		t.Fatalf("phrase count = %d, want 3; details=%#v", phrases, details)
	}
	if dependencies != 2 {
		t.Fatalf("dependency count = %d, want 2", dependencies)
	}
	if mdd != 1.5 {
		t.Fatalf("mdd = %v, want 1.5; details=%#v", mdd, details)
	}
	if details[0].RootPhrase != 2 {
		t.Fatalf("root = %d, want 2", details[0].RootPhrase)
	}
	if details[0].Dependencies[0].Rule != "case-or-topic-particle-to-next-predicate" {
		t.Fatalf("first rule = %q", details[0].Dependencies[0].Rule)
	}
}

func TestSplitPhrasesDoesNotCreatePunctuationOnlyRoot(t *testing.T) {
	analyzer := mustAnalyzer(t)
	phrases := splitPhrases(tokenizeMorphs(analyzer, "これは正しいでしょうか。"))
	if len(phrases) == 0 {
		t.Fatal("no phrases")
	}
	last := phraseToDetail(len(phrases)-1, phrases[len(phrases)-1])
	if last.Text == "。" || last.Text == "？" || last.Text == "?" {
		t.Fatalf("punctuation-only root phrase: %#v", last)
	}
}

func TestLoadCorpusSortsAndExcludesTarget(t *testing.T) {
	analyzer := mustAnalyzer(t)
	root := t.TempDir()
	target := filepath.Join(root, "b", "README.md")
	writeFixture(t, filepath.Join(root, "a", "README.md"), "猫が走る。")
	writeFixture(t, target, "除外する。")
	writeFixture(t, filepath.Join(root, "c", "README.md"), "犬が眠る。")

	texts, files, tokens, err := loadCorpus(root, target, analyzer)
	if err != nil {
		t.Fatal(err)
	}
	if files != 2 {
		t.Fatalf("files = %d, want 2", files)
	}
	if len(texts) != 2 || texts[0] != "猫が走る。\n" || texts[1] != "犬が眠る。\n" {
		t.Fatalf("texts = %#v", texts)
	}
	if tokens == 0 {
		t.Fatal("corpus token count is zero")
	}
}

func mustAnalyzer(t *testing.T) *tokenizer.Tokenizer {
	t.Helper()
	analyzer, err := newAnalyzer()
	if err != nil {
		t.Fatal(err)
	}
	return analyzer
}

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
