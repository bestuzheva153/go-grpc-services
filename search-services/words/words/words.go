package words

import (
	"maps"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/kljensen/snowball"
	"github.com/kljensen/snowball/english"
)

func Norm(phrase string) []string {
	tokens := strings.FieldsFunc(strings.ToLower(phrase), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	words := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if english.IsStopWord(token) {
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
		if english.IsStopWord(stemmed) {
			continue
		}
		words[stemmed] = true
	}
	result := slices.Collect(maps.Keys(words))
	sort.Strings(result)
	return result
}
