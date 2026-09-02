// Package main provides a command that measures the readability of learner-facing
// Markdown in a workshop scenario.
//
// # Processing overview
//
// The command removes content that GitHub does not render as learner-facing prose,
// including code blocks, HTML comments, collapsed details, URLs, and Markdown
// delimiters. It then performs these steps:
//
//  1. Measure sentence length, kanji ratio, and structural signal ratio.
//  2. Tokenize visible Japanese text with Kagome and its IPA dictionary.
//  3. Estimate bunsetsu dependencies with deterministic forward rules.
//  4. Train an add-alpha word-bigram model from the other workshop README files.
//  5. Calculate an estimated cognitive readability score from those metrics.
//
// The corpus paths are sorted, and the target file is excluded. The same repository
// state and analyzer version therefore produce the same result.
//
// # Deterministic text metrics
//
// L is the rune count of the longest visible sentence.
//
// The kanji ratio K is:
//
//	K = kanji characters / visible text characters * 100
//
// The structural signal ratio STR is:
//
//	STR = heading, list-item, and bold characters / visible text characters
//
// Markdown syntax, whitespace, punctuation, and symbols are excluded from the
// ratio denominators. Overlapping structural text is counted once.
//
// # MDD estimate
//
// Kagome provides morphology and part-of-speech data, not dependency parsing.
// This command therefore reports mdd_estimate rather than parser-derived MDD.
// It approximates bunsetsu boundaries at particles and punctuation. Genitive
// phrases seek the next noun, case and topic phrases seek the next predicate,
// and remaining phrases attach to the nearest following phrase.
//
// For the inferred non-root dependencies D, the estimate is:
//
//	mdd_estimate = sum(abs(head(d) - dependent(d))) / len(D)
//
// Each sentence has its own root, which is excluded from D. The JSON output
// includes every inferred phrase, head, distance, and rule for auditing.
//
// # N-gram surprisal
//
// The command builds a word-bigram model from visible prose in every other
// workshop README. Words occurring once are mapped to an unknown token.
// Beginning-of-sequence and end-of-sequence events are included.
//
// With add-alpha smoothing, the conditional probability is:
//
//	P(w_i | w_(i-1)) =
//	    (count(w_(i-1), w_i) + alpha) /
//	    (count(w_(i-1)) + alpha * vocabulary_size)
//
// Average n-gram surprisal in bits per token is:
//
//	S = -sum(log2(P(w_i | w_(i-1)))) / event_count
//
// This is a normalized n-gram conditional surprisal. It is not the probability
// from a neural language model or a user-supplied base ONNX model.
//
// # CRS estimate
//
// The command applies the workshop's heuristic objective function:
//
//	crs_estimate = 100 -
//	    (15 * max(0, mdd_estimate - 2.5)^2 +
//	     0.5 * max(0, L - 60) +
//	     abs(K - 35) +
//	     2 * S) +
//	    20 * STR
//
// STR is converted from a percentage to a ratio before substitution.
// Because this calculation uses an inferred MDD and a local n-gram model, it
// must not be compared as if it were a validated psycholinguistic score.
//
// The thresholds for 60-character sentences, 30-40 percent kanji, 15-30 percent
// STR, and the CRS coefficients are workshop review heuristics. The cited
// research motivates dependency locality, surprisal, and cognitive load, but
// does not validate these exact thresholds or this composite formula.
//
// # References
//
//   - Futrell, Mahowald, and Gibson, "Large-scale evidence of dependency length
//     minimization in 37 languages": https://doi.org/10.1073/pnas.1502134112
//   - Levy, "Expectation-based syntactic comprehension":
//     https://doi.org/10.1016/j.cognition.2007.05.006
//   - Hale, "A Probabilistic Earley Parser as a Psycholinguistic Model":
//     https://aclanthology.org/N01-1021/
//   - Sweller, "Cognitive Load During Problem Solving":
//     https://doi.org/10.1207/s15516709cog1202_4
//   - Shannon, "A Mathematical Theory of Communication":
//     https://doi.org/10.1002/j.1538-7305.1948.tb01338.x
//   - Kagome morphological analyzer: https://github.com/ikawaha/kagome
package main
