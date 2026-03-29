package words

import (
	"regexp"
	"sort"
	"strings"

	"github.com/kljensen/snowball"
)

var tokenRegexp = regexp.MustCompile(`[a-zA-Z0-9]+`)

var stopWords = map[string]struct{}{
	"a": {}, "an": {}, "the": {},
	"of": {}, "to": {}, "in": {}, "on": {}, "at": {}, "by": {}, "for": {}, "with": {}, "from": {},
	"and": {}, "or": {}, "but": {},
	"is": {}, "am": {}, "are": {}, "was": {}, "were": {}, "be": {}, "been": {}, "being": {},
	"do": {}, "does": {}, "did": {},
	"have": {}, "has": {}, "had": {},
	"will": {}, "would": {}, "shall": {}, "should": {}, "can": {}, "could": {}, "may": {}, "might": {}, "must": {},
	"i": {}, "you": {}, "he": {}, "she": {}, "it": {}, "we": {}, "they": {},
	"me": {}, "him": {}, "her": {}, "us": {}, "them": {},
	"my": {}, "your": {}, "his": {}, "its": {}, "our": {}, "their": {},
	"who": {}, "whom": {}, "what": {}, "when": {}, "where": {}, "why": {}, "how": {},
	"this": {}, "that": {}, "these": {}, "those": {},
}

func Normalize(phrase string) []string {
	tokens := tokenRegexp.FindAllString(strings.ToLower(phrase), -1)
	seen := make(map[string]struct{}, len(tokens))
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, skip := stopWords[token]; skip {
			continue
		}
		stemmed, err := snowball.Stem(token, "english", true)
		if err != nil {
			stemmed = token
		}
		stemmed = strings.TrimSpace(stemmed)
		if stemmed == "" {
			continue
		}
		if _, skip := stopWords[stemmed]; skip {
			continue
		}
		if _, exists := seen[stemmed]; exists {
			continue
		}
		seen[stemmed] = struct{}{}
		result = append(result, stemmed)
	}
	sort.Strings(result)
	return result
}
