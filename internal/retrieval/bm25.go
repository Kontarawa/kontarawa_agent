package retrieval

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

type Doc struct {
	DocID string
	Text  string
}

type Retrieved struct {
	DocID string
	Text  string
	Score float64
}

// Retrieve uses a BM25-like scoring over tokens. No external dependencies.
func Retrieve(query string, docs []Doc, k int) []Retrieved {
	if k <= 0 || len(docs) == 0 {
		return nil
	}
	qTokens := tokenize(query)
	if len(qTokens) == 0 {
		return nil
	}

	// document frequencies
	df := map[string]int{}
	docTokens := make([][]string, 0, len(docs))
	for _, d := range docs {
		toks := tokenize(d.Text)
		docTokens = append(docTokens, toks)
		seen := map[string]bool{}
		for _, t := range toks {
			if !seen[t] {
				seen[t] = true
				df[t]++
			}
		}
	}

	N := float64(len(docs))
	avgLen := 0.0
	for _, toks := range docTokens {
		avgLen += float64(len(toks))
	}
	if N > 0 {
		avgLen /= N
	}

	type scored struct {
		idx   int
		score float64
	}
	scoredDocs := make([]scored, 0, len(docs))

	// BM25 params
	k1 := 1.2
	b := 0.75

	for i, toks := range docTokens {
		if len(toks) == 0 {
			continue
		}
		tf := map[string]int{}
		for _, t := range toks {
			tf[t]++
		}

		docLen := float64(len(toks))
		score := 0.0
		for _, t := range qTokens {
			f := float64(tf[t])
			if f == 0 {
				continue
			}
			dft := float64(df[t])
			// idf with smoothing
			idf := math.Log((N-dft+0.5)/(dft+0.5) + 1.0)
			den := f + k1*(1.0-b+b*(docLen/avgLen))
			score += idf * (f * (k1 + 1.0) / den)
		}
		if score > 0 {
			scoredDocs = append(scoredDocs, scored{idx: i, score: score})
		}
	}

	sort.Slice(scoredDocs, func(i, j int) bool { return scoredDocs[i].score > scoredDocs[j].score })
	if len(scoredDocs) > k {
		scoredDocs = scoredDocs[:k]
	}

	out := make([]Retrieved, 0, len(scoredDocs))
	for _, s := range scoredDocs {
		out = append(out, Retrieved{
			DocID: docs[s.idx].DocID,
			Text:  docs[s.idx].Text,
			Score: s.score,
		})
	}
	return out
}

var nonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)

func tokenize(s string) []string {
	s = strings.ToLower(s)
	s = nonWord.ReplaceAllString(s, " ")
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil
	}
	// Tiny stopword list (ru+en) to keep things reasonable.
	stop := map[string]bool{
		"и": true, "в": true, "во": true, "на": true, "не": true, "что": true, "это": true, "я": true, "мы": true,
		"the": true, "a": true, "an": true, "and": true, "or": true, "to": true, "of": true, "in": true, "is": true, "are": true,
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len([]rune(p)) < 2 {
			continue
		}
		if stop[p] {
			continue
		}
		out = append(out, p)
	}
	return out
}

