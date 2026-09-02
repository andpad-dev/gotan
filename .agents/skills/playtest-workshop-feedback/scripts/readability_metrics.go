package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

const (
	surprisalAlpha  = 0.5
	unknownToken    = "<UNK>"
	beginToken      = "<BOS>"
	endToken        = "<EOS>"
	mddMethod       = "kagome-ipa-bunsetsu-forward-heuristic-v1"
	surprisalMethod = "kagome-ipa-add-alpha-word-bigram-v1"
	crsMethod       = "readability-crs-with-estimated-mdd-and-ngram-surprisal-v1"
)

var (
	linkPattern            = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)|\[([^\]]+)\]\([^)]*\)`)
	bareURLPattern         = regexp.MustCompile(`https?://[^\s<>\[\](){}「」『』。、，；：！？]+`)
	boldPattern            = regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	tagPattern             = regexp.MustCompile(`<[^>]+>`)
	inlineCode             = regexp.MustCompile("`([^`]*)`")
	headingPattern         = regexp.MustCompile(`^#{1,6}\s+`)
	listPattern            = regexp.MustCompile(`^\s*(?:[-*+]|\d+\.)\s+`)
	fencePattern           = regexp.MustCompile(`^\s*(` + "`" + `{3,}|~{3,})`)
	detailsTagPattern      = regexp.MustCompile(`(?i)</?details\b[^>]*>`)
	tableSeparator         = regexp.MustCompile(`^\s*\|?\s*:?-{3,}:?\s*(?:\|\s*:?-{3,}:?\s*)+\|?\s*$`)
	spaceBeforePunctuation = regexp.MustCompile(`\s+([。、，；：！？!?])`)
	blockquotePrefix       = regexp.MustCompile(`^\s{0,3}>\s?`)
)

type sentenceMetric struct {
	Text   string `json:"text"`
	Length int    `json:"length"`
}

type dependencyDetail struct {
	From     int    `json:"from"`
	To       int    `json:"to"`
	Distance int    `json:"distance"`
	Rule     string `json:"rule"`
}

type phraseDetail struct {
	Index    int      `json:"index"`
	Text     string   `json:"text"`
	Surfaces []string `json:"surfaces"`
	POS      []string `json:"pos"`
}

type sentenceMDDDetail struct {
	Sentence        string             `json:"sentence"`
	Phrases         []phraseDetail     `json:"phrases"`
	Dependencies    []dependencyDetail `json:"dependencies"`
	RootPhrase      int                `json:"root_phrase"`
	MDD             float64            `json:"mdd_estimate"`
	PhraseCount     int                `json:"phrase_count"`
	DependencyCount int                `json:"dependency_count"`
	Qualification   string             `json:"qualification"`
}

type metrics struct {
	Scope                  string              `json:"scope"`
	TotalTextChars         int                 `json:"total_text_chars"`
	KanjiChars             int                 `json:"kanji_chars"`
	KanjiRatio             float64             `json:"kanji_ratio"`
	DeltaKanjiFrom35       float64             `json:"delta_kanji_from_35"`
	StructuralChars        int                 `json:"structural_chars"`
	StructuralSignalRatio  float64             `json:"structural_signal_ratio"`
	MaxSentenceLength      int                 `json:"max_sentence_length"`
	SentencesOver60        []sentenceMetric    `json:"sentences_over_60"`
	Sentences              []sentenceMetric    `json:"sentences"`
	MDDEstimate            float64             `json:"mdd_estimate"`
	MDDPhraseCount         int                 `json:"mdd_phrase_count"`
	MDDDependencyCount     int                 `json:"mdd_dependency_count"`
	MDDMethod              string              `json:"mdd_method"`
	MDDQualification       string              `json:"mdd_qualification"`
	MDDSentences           []sentenceMDDDetail `json:"mdd_sentences"`
	NgramSurprisal         float64             `json:"ngram_surprisal"`
	SurprisalCorpusFiles   int                 `json:"surprisal_corpus_file_count"`
	SurprisalCorpusTokens  int                 `json:"surprisal_corpus_token_count"`
	SurprisalVocabulary    int                 `json:"surprisal_vocabulary_size"`
	SurprisalAlpha         float64             `json:"surprisal_alpha"`
	SurprisalMethod        string              `json:"surprisal_method"`
	SurprisalQualification string              `json:"surprisal_qualification"`
	CRSEstimate            float64             `json:"crs_estimate"`
	CRSMethod              string              `json:"crs_method"`
	CRSQualification       string              `json:"crs_qualification"`
}

type extractedMarkdown struct {
	totalText      string
	structuralText string
	sentences      []sentenceMetric
	languageText   string
}

type morph struct {
	surface string
	base    string
	pos     string
	posSub1 string
}

type phrase struct {
	morphs []morph
}

type bigramModel struct {
	alpha         float64
	vocabulary    []string
	vocabularySet map[string]struct{}
	bigrams       map[string]map[string]int
	contexts      map[string]int
}

func main() {
	var corpusFlag string
	flag.StringVar(&corpusFlag, "corpus", "", "workshop corpus root (default: discovered repository workshop directory)")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: go run . [-corpus <workshop-root>] <markdown>\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	target, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		exitError(err)
	}
	corpusRoot, err := resolveCorpusRoot(corpusFlag, target)
	if err != nil {
		exitError(err)
	}
	analyzer, err := newAnalyzer()
	if err != nil {
		exitError(err)
	}
	result, err := measure(target, corpusRoot, analyzer)
	if err != nil {
		exitError(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		exitError(err)
	}
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func newAnalyzer() (*tokenizer.Tokenizer, error) {
	return tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
}

func measure(path, corpusRoot string, analyzer *tokenizer.Tokenizer) (metrics, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return metrics{}, err
	}
	extracted := extractMarkdown(string(content))

	totalRunes := []rune(extracted.totalText)
	structuralRunes := []rune(extracted.structuralText)
	kanji := 0
	for _, r := range totalRunes {
		if unicode.Is(unicode.Han, r) {
			kanji++
		}
	}

	maxLength := 0
	over60 := make([]sentenceMetric, 0)
	for _, sentence := range extracted.sentences {
		if sentence.Length > maxLength {
			maxLength = sentence.Length
		}
		if sentence.Length > 60 {
			over60 = append(over60, sentence)
		}
	}

	mdd, phraseCount, dependencyCount, mddSentences := estimateMDD(extracted.sentences, analyzer)
	corpusTexts, corpusFiles, corpusTokens, err := loadCorpus(corpusRoot, path, analyzer)
	if err != nil {
		return metrics{}, err
	}
	model := trainBigram(corpusTexts, analyzer, surprisalAlpha)
	targetTokens := tokenizeWords(analyzer, extracted.languageText)
	surprisal := model.averageSurprisal(targetTokens)

	kanjiRatio := ratio(kanji, len(totalRunes))
	structuralRatio := ratio(len(structuralRunes), len(totalRunes))
	crs := calculateCRS(mdd, maxLength, kanjiRatio, surprisal, structuralRatio)
	return metrics{
		Scope:                  "visible prose excluding details, code blocks, URLs, Markdown, whitespace, punctuation",
		TotalTextChars:         len(totalRunes),
		KanjiChars:             kanji,
		KanjiRatio:             kanjiRatio,
		DeltaKanjiFrom35:       round2(math.Abs(kanjiRatio - 35)),
		StructuralChars:        len(structuralRunes),
		StructuralSignalRatio:  structuralRatio,
		MaxSentenceLength:      maxLength,
		SentencesOver60:        over60,
		Sentences:              extracted.sentences,
		MDDEstimate:            mdd,
		MDDPhraseCount:         phraseCount,
		MDDDependencyCount:     dependencyCount,
		MDDMethod:              mddMethod,
		MDDQualification:       "Estimate from Kagome IPA morphology and documented forward-dependency heuristics; not an exact parser-derived MDD.",
		MDDSentences:           mddSentences,
		NgramSurprisal:         surprisal,
		SurprisalCorpusFiles:   corpusFiles,
		SurprisalCorpusTokens:  corpusTokens,
		SurprisalVocabulary:    len(model.vocabulary),
		SurprisalAlpha:         model.alpha,
		SurprisalMethod:        surprisalMethod,
		SurprisalQualification: "Normalized add-alpha word-bigram conditional surprisal in bits/token; not neural or base-ONNX surprisal.",
		CRSEstimate:            crs,
		CRSMethod:              crsMethod,
		CRSQualification:       "Estimate using heuristic MDD and local n-gram surprisal; do not compare as identical to model-backed CRS.",
	}, nil
}

func extractMarkdown(content string) extractedMarkdown {
	var totalText, structuralText, languageText strings.Builder
	var sentences []sentenceMetric
	var paragraph []string
	var listItem []string
	inCode := false
	var fenceChar byte
	fenceLength := 0
	detailsDepth := 0
	inHTMLComment := false

	appendUnit := func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		languageText.WriteString(text)
		languageText.WriteRune('\n')
		sentences = append(sentences, splitSentences(text)...)
	}
	flushParagraph := func() {
		appendUnit(strings.Join(paragraph, " "))
		paragraph = paragraph[:0]
	}
	flushListItem := func() {
		text := strings.Join(listItem, " ")
		if text != "" {
			structuralText.WriteString(eligible(text))
			appendUnit(text)
		}
		listItem = listItem[:0]
	}
	flushBlocks := func() {
		flushParagraph()
		flushListItem()
	}

	for _, originalLine := range strings.Split(content, "\n") {
		containerLine := stripBlockquotePrefixes(originalLine)
		if inCode {
			if fencePattern.MatchString(containerLine) &&
				isClosingFence(containerLine, fenceChar, fenceLength) {
				inCode = false
				fenceChar = 0
				fenceLength = 0
			}
			continue
		}

		rawLine := stripHTMLComments(containerLine, &inHTMLComment)
		rawLine = stripDetails(rawLine, &detailsDepth)
		line := strings.TrimSpace(rawLine)
		if fence := fencePattern.FindStringSubmatch(rawLine); fence != nil {
			marker := strings.TrimSpace(fence[1])
			if detailsDepth == 0 {
				flushBlocks()
				inCode = true
				fenceChar = marker[0]
				fenceLength = len(marker)
			}
			continue
		}
		if line == "" || line == "---" {
			flushBlocks()
			continue
		}
		isIndented := strings.HasPrefix(rawLine, "\t") || strings.HasPrefix(rawLine, "    ")
		isNestedList := len(listItem) > 0 && listPattern.MatchString(rawLine)
		if isIndented && !isNestedList {
			flushBlocks()
			continue
		}
		if tableSeparator.MatchString(rawLine) {
			flushBlocks()
			continue
		}

		isHeading := headingPattern.MatchString(line)
		isList := listPattern.MatchString(rawLine)
		tableCells, isTableRow := splitTableCells(rawLine)
		plain := plainText(rawLine)
		if plain == "" {
			continue
		}

		totalText.WriteString(eligible(plain))
		if isList {
			flushParagraph()
			flushListItem()
			listItem = append(listItem, plain)
			continue
		}
		if len(listItem) > 0 && !isHeading && !isTableRow {
			listItem = append(listItem, plain)
			continue
		}
		flushListItem()

		if isHeading {
			structuralText.WriteString(eligible(plain))
		} else {
			for _, match := range boldPattern.FindAllStringSubmatch(rawLine, -1) {
				structuralText.WriteString(eligible(plainText(firstNonEmpty(match[1:]...))))
			}
		}

		if isHeading {
			flushParagraph()
			appendUnit(plain)
			continue
		}
		if isTableRow {
			flushParagraph()
			for _, cell := range tableCells {
				appendUnit(plainText(cell))
			}
			continue
		}
		paragraph = append(paragraph, plain)
	}
	flushBlocks()
	return extractedMarkdown{
		totalText:      totalText.String(),
		structuralText: structuralText.String(),
		sentences:      sentences,
		languageText:   languageText.String(),
	}
}

func isClosingFence(line string, fenceChar byte, minimumLength int) bool {
	marker := strings.TrimSpace(line)
	if len(marker) < minimumLength {
		return false
	}
	for i := range marker {
		if marker[i] != fenceChar {
			return false
		}
	}
	return true
}

func stripBlockquotePrefixes(line string) string {
	for blockquotePrefix.MatchString(line) {
		line = blockquotePrefix.ReplaceAllString(line, "")
	}
	return line
}

func stripHTMLComments(line string, inComment *bool) string {
	var visible strings.Builder
	for line != "" {
		if *inComment {
			end := strings.Index(line, "-->")
			if end < 0 {
				return visible.String()
			}
			line = line[end+3:]
			*inComment = false
			continue
		}
		start := strings.Index(line, "<!--")
		if start < 0 {
			visible.WriteString(line)
			break
		}
		visible.WriteString(line[:start])
		visible.WriteRune(' ')
		line = line[start+4:]
		*inComment = true
	}
	return visible.String()
}

func stripDetails(line string, depth *int) string {
	var visible strings.Builder
	last := 0
	for _, location := range detailsTagPattern.FindAllStringIndex(line, -1) {
		if *depth == 0 {
			visible.WriteString(line[last:location[0]])
		}
		tag := strings.ToLower(line[location[0]:location[1]])
		if strings.HasPrefix(tag, "</") {
			if *depth > 0 {
				*depth--
			}
		} else {
			*depth++
		}
		last = location[1]
	}
	if *depth == 0 {
		visible.WriteString(line[last:])
	}
	return visible.String()
}

func splitTableCells(line string) ([]string, bool) {
	var cells []string
	var current strings.Builder
	pipeCount := 0
	codeTicks := 0

	for i := 0; i < len(line); {
		if line[i] == '\\' && i+1 < len(line) && line[i+1] == '|' {
			current.WriteString(`\|`)
			i += 2
			continue
		}
		if line[i] == '`' {
			j := i
			for j < len(line) && line[j] == '`' {
				j++
			}
			count := j - i
			if codeTicks == 0 {
				codeTicks = count
			} else if count == codeTicks {
				codeTicks = 0
			}
			current.WriteString(line[i:j])
			i = j
			continue
		}
		if line[i] == '|' && codeTicks == 0 {
			cells = append(cells, strings.TrimSpace(current.String()))
			current.Reset()
			pipeCount++
			i++
			continue
		}
		current.WriteByte(line[i])
		i++
	}
	cells = append(cells, strings.TrimSpace(current.String()))
	if pipeCount == 0 {
		return nil, false
	}
	if len(cells) > 0 && cells[0] == "" {
		cells = cells[1:]
	}
	if len(cells) > 0 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}
	return cells, true
}

func resolveCorpusRoot(flagValue, target string) (string, error) {
	if flagValue != "" {
		root, err := filepath.Abs(flagValue)
		if err != nil {
			return "", err
		}
		info, err := os.Stat(root)
		if err != nil {
			return "", fmt.Errorf("corpus root: %w", err)
		}
		if !info.IsDir() {
			return "", fmt.Errorf("corpus root is not a directory: %s", root)
		}
		return root, nil
	}
	starts := []string{filepath.Dir(target)}
	if cwd, err := os.Getwd(); err == nil {
		starts = append(starts, cwd)
	}
	for _, start := range starts {
		for current := start; ; current = filepath.Dir(current) {
			candidate := filepath.Join(current, "workshop")
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return candidate, nil
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
		}
	}
	return "", errors.New("could not discover workshop corpus root; pass -corpus")
}

func loadCorpus(root, target string, analyzer *tokenizer.Tokenizer) ([]string, int, int, error) {
	target, err := filepath.Abs(target)
	if err != nil {
		return nil, 0, 0, err
	}
	var paths []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "README.md" {
			return nil
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if absolute != target {
			paths = append(paths, absolute)
		}
		return nil
	})
	if err != nil {
		return nil, 0, 0, err
	}
	sort.Strings(paths)
	texts := make([]string, 0, len(paths))
	tokenCount := 0
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, 0, 0, err
		}
		text := extractMarkdown(string(content)).languageText
		texts = append(texts, text)
		tokenCount += len(tokenizeWords(analyzer, text))
	}
	return texts, len(paths), tokenCount, nil
}

func tokenizeMorphs(analyzer *tokenizer.Tokenizer, text string) []morph {
	tokens := analyzer.Tokenize(text)
	result := make([]morph, 0, len(tokens))
	for _, token := range tokens {
		surface := strings.TrimSpace(token.Surface)
		if surface == "" {
			continue
		}
		pos := token.POS()
		base, ok := token.BaseForm()
		if !ok || base == "*" || base == "" {
			base = surface
		}
		item := morph{surface: surface, base: base}
		if len(pos) > 0 {
			item.pos = pos[0]
		}
		if len(pos) > 1 {
			item.posSub1 = pos[1]
		}
		result = append(result, item)
	}
	return result
}

func tokenizeWords(analyzer *tokenizer.Tokenizer, text string) []string {
	morphs := tokenizeMorphs(analyzer, text)
	words := make([]string, 0, len(morphs))
	for _, item := range morphs {
		if isPunctuation(item) {
			continue
		}
		words = append(words, item.base)
	}
	return words
}

func splitPhrases(morphs []morph) []phrase {
	var phrases []phrase
	var current phrase
	flush := func() {
		if len(current.morphs) != 0 {
			phrases = append(phrases, current)
			current = phrase{}
		}
	}
	for _, item := range morphs {
		current.morphs = append(current.morphs, item)
		switch {
		case isPunctuation(item):
			flush()
		case item.pos == "助詞":
			flush()
		}
	}
	flush()
	return phrases
}

func estimateMDD(sentences []sentenceMetric, analyzer *tokenizer.Tokenizer) (float64, int, int, []sentenceMDDDetail) {
	var details []sentenceMDDDetail
	totalDistance := 0
	totalPhrases := 0
	totalDependencies := 0
	for _, sentence := range sentences {
		phrases := splitPhrases(tokenizeMorphs(analyzer, sentence.Text))
		if len(phrases) == 0 {
			continue
		}
		root := len(phrases) - 1
		detail := sentenceMDDDetail{
			Sentence:      sentence.Text,
			RootPhrase:    root,
			PhraseCount:   len(phrases),
			Qualification: "Forward heuristic; final phrase is root and excluded from the MDD denominator.",
		}
		for i, p := range phrases {
			detail.Phrases = append(detail.Phrases, phraseToDetail(i, p))
		}
		sentenceDistance := 0
		for i := 0; i < root; i++ {
			head, rule := inferHead(phrases, i)
			distance := head - i
			sentenceDistance += distance
			totalDistance += distance
			totalDependencies++
			detail.Dependencies = append(detail.Dependencies, dependencyDetail{
				From: i, To: head, Distance: distance, Rule: rule,
			})
		}
		detail.DependencyCount = len(detail.Dependencies)
		if detail.DependencyCount > 0 {
			detail.MDD = round2(float64(sentenceDistance) / float64(detail.DependencyCount))
		}
		totalPhrases += len(phrases)
		details = append(details, detail)
	}
	if totalDependencies == 0 {
		return 0, totalPhrases, 0, details
	}
	return round2(float64(totalDistance) / float64(totalDependencies)), totalPhrases, totalDependencies, details
}

func inferHead(phrases []phrase, index int) (int, string) {
	last := len(phrases) - 1
	if index >= last {
		return last, "root"
	}
	tail := lastNonPunctuation(phrases[index])
	if tail.surface == "の" && tail.pos == "助詞" {
		for i := index + 1; i <= last; i++ {
			if phraseHasPOS(phrases[i], "名詞") {
				return i, "genitive-no-to-next-noun"
			}
		}
	}
	if tail.pos == "助詞" && isCaseOrTopicParticle(tail) {
		for i := index + 1; i <= last; i++ {
			if phraseIsPredicate(phrases[i]) {
				return i, "case-or-topic-particle-to-next-predicate"
			}
		}
	}
	return index + 1, "nearest-following-phrase"
}

func phraseToDetail(index int, p phrase) phraseDetail {
	detail := phraseDetail{Index: index}
	var text strings.Builder
	seenPOS := make(map[string]struct{})
	for _, item := range p.morphs {
		text.WriteString(item.surface)
		detail.Surfaces = append(detail.Surfaces, item.surface)
		if item.pos != "" {
			if _, ok := seenPOS[item.pos]; !ok {
				detail.POS = append(detail.POS, item.pos)
				seenPOS[item.pos] = struct{}{}
			}
		}
	}
	detail.Text = text.String()
	return detail
}

func lastNonPunctuation(p phrase) morph {
	for i := len(p.morphs) - 1; i >= 0; i-- {
		if !isPunctuation(p.morphs[i]) {
			return p.morphs[i]
		}
	}
	return morph{}
}

func isPunctuation(item morph) bool {
	return item.pos == "記号" || strings.ContainsAny(item.surface, "。！？!?、,.")
}

func isCaseOrTopicParticle(item morph) bool {
	if item.posSub1 == "格助詞" || item.posSub1 == "係助詞" {
		return true
	}
	switch item.surface {
	case "は", "が", "を", "に", "へ", "で", "と", "から", "まで", "より":
		return true
	default:
		return false
	}
}

func phraseHasPOS(p phrase, wanted string) bool {
	for _, item := range p.morphs {
		if item.pos == wanted {
			return true
		}
	}
	return false
}

func phraseIsPredicate(p phrase) bool {
	for _, item := range p.morphs {
		if item.pos == "動詞" || item.pos == "形容詞" {
			return true
		}
	}
	for i, item := range p.morphs {
		if item.pos == "名詞" && i+1 < len(p.morphs) &&
			(p.morphs[i+1].pos == "助動詞" || p.morphs[i+1].surface == "です") {
			return true
		}
	}
	return false
}

func trainBigram(texts []string, analyzer *tokenizer.Tokenizer, alpha float64) bigramModel {
	rawCounts := make(map[string]int)
	tokenized := make([][]string, 0, len(texts))
	for _, text := range texts {
		words := tokenizeWords(analyzer, text)
		tokenized = append(tokenized, words)
		for _, word := range words {
			rawCounts[word]++
		}
	}
	vocabularySet := map[string]struct{}{unknownToken: {}, endToken: {}}
	for word, count := range rawCounts {
		if count >= 2 {
			vocabularySet[word] = struct{}{}
		}
	}
	vocabulary := make([]string, 0, len(vocabularySet))
	for word := range vocabularySet {
		vocabulary = append(vocabulary, word)
	}
	sort.Strings(vocabulary)
	model := bigramModel{
		alpha:         alpha,
		vocabulary:    vocabulary,
		vocabularySet: vocabularySet,
		bigrams:       make(map[string]map[string]int),
		contexts:      make(map[string]int),
	}
	for _, words := range tokenized {
		context := beginToken
		for _, word := range words {
			word = model.normalize(word)
			model.add(context, word)
			context = word
		}
		model.add(context, endToken)
	}
	return model
}

func (m *bigramModel) add(context, word string) {
	if m.bigrams[context] == nil {
		m.bigrams[context] = make(map[string]int)
	}
	m.bigrams[context][word]++
	m.contexts[context]++
}

func (m bigramModel) normalize(word string) string {
	if _, ok := m.vocabularySet[word]; ok {
		return word
	}
	return unknownToken
}

func (m bigramModel) probability(context, word string) float64 {
	word = m.normalize(word)
	vocabularySize := float64(len(m.vocabulary))
	return (float64(m.bigrams[context][word]) + m.alpha) /
		(float64(m.contexts[context]) + m.alpha*vocabularySize)
}

func (m bigramModel) averageSurprisal(words []string) float64 {
	context := beginToken
	total := 0.0
	events := 0
	for _, word := range words {
		word = m.normalize(word)
		total += -math.Log2(m.probability(context, word))
		context = word
		events++
	}
	total += -math.Log2(m.probability(context, endToken))
	events++
	return round2(total / float64(events))
}

func calculateCRS(mdd float64, maxSentenceLength int, kanjiRatio, surprisal, structuralRatio float64) float64 {
	penalty := 15*math.Pow(math.Max(0, mdd-2.5), 2) +
		0.5*math.Max(0, float64(maxSentenceLength-60)) +
		math.Abs(kanjiRatio-35) +
		2*surprisal
	return round2(100 - penalty + 20*(structuralRatio/100))
}

func plainText(line string) string {
	line = linkPattern.ReplaceAllStringFunc(line, func(value string) string {
		match := linkPattern.FindStringSubmatch(value)
		return firstNonEmpty(match[1:]...)
	})
	line = boldPattern.ReplaceAllStringFunc(line, func(value string) string {
		match := boldPattern.FindStringSubmatch(value)
		return firstNonEmpty(match[1:]...)
	})
	line = tagPattern.ReplaceAllString(line, "")
	line = inlineCode.ReplaceAllString(line, "$1")
	line = strings.ReplaceAll(line, `\|`, `|`)
	line = bareURLPattern.ReplaceAllString(line, "")
	line = spaceBeforePunctuation.ReplaceAllString(line, "$1")
	line = headingPattern.ReplaceAllString(line, "")
	line = listPattern.ReplaceAllString(line, "")
	return strings.Join(strings.Fields(line), " ")
}

func eligible(text string) string {
	var result strings.Builder
	for _, r := range text {
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

func splitSentences(text string) []sentenceMetric {
	var result []sentenceMetric
	var sentence strings.Builder
	for _, r := range text {
		sentence.WriteRune(r)
		if strings.ContainsRune("。！？!?", r) {
			result = appendSentence(result, sentence.String())
			sentence.Reset()
		}
	}
	return appendSentence(result, sentence.String())
}

func appendSentence(sentences []sentenceMetric, text string) []sentenceMetric {
	text = strings.TrimSpace(text)
	if text == "" {
		return sentences
	}
	return append(sentences, sentenceMetric{Text: text, Length: len([]rune(text))})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return round2(float64(numerator) / float64(denominator) * 100)
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
