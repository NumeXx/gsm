package wordlist

import (
	_ "embed"
	"strings"
	"sync"
)

//go:embed english.txt
var embeddedEnglishTxt string

var (
	parsedWords []string
	once        sync.Once
)

// GetWords returns a slice of strings from the embedded english.txt wordlist.
// It parses the embedded content only once.
func GetWords() []string {
	once.Do(func() {
		lines := strings.Split(strings.ReplaceAll(embeddedEnglishTxt, "\r\n", "\n"), "\n")
		parsedWords = make([]string, 0, len(lines))
		for _, line := range lines {
			word := strings.TrimSpace(line)
			if word != "" {
				parsedWords = append(parsedWords, word)
			}
		}
	})
	return parsedWords
}
