package nlp

import (
	"github.com/ArditZubaku/nlp/stemmer"
	"regexp"
	"strings"
)

var (
	wordRe = regexp.MustCompile(`[a-zA-Z]+`)
)

func Tokenize(text string) []string {
	words := wordRe.FindAllString(text, -1)
	var tokens []string
	for _, w := range words {
		token := strings.ToLower(w)
		token = stemmer.Stem(token)
		if len(token) != 0 {
			tokens = append(tokens, token)
		}
	}
	return tokens
}
