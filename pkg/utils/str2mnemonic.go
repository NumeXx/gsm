package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	// "os" // No longer needed directly by GenerateMnemonic if dictionary is passed
	// "bufio" // No longer needed if loadWordList is removed or not used by GenerateMnemonic
	"strconv"
	"strings"
	"unicode"
)

// loadWordList is now part of pkg/wordlist or handled externally before calling GenerateMnemonic.
// If it's still needed in this package for other purposes, it can remain,
// but GenerateMnemonic will now expect a []string dictionary.

// GenerateMnemonic generates a mnemonic from a secret string using the provided dictionary.
// It now ensures that each word in the mnemonic starts with a capital letter.
func GenerateMnemonic(secret string, numWords int, dictionary []string) (string, error) {
	if numWords <= 0 {
		return "", fmt.Errorf("number of words must be greater than 0")
	}

	if len(dictionary) == 0 {
		return "", fmt.Errorf("dictionary is empty")
	}
	wordListSize := uint64(len(dictionary))

	hasher := md5.New()
	hasher.Write([]byte(secret))
	hashBytes := hasher.Sum(nil)
	hexHash := hex.EncodeToString(hashBytes)

	if len(hexHash) < 15 {
		return "", fmt.Errorf("MD5 hash is too short (less than 15 characters): %s", hexHash)
	}

	// Take the first 15 hex characters and convert to uint64
	numToConvert := hexHash[:15]
	decimalNum, err := strconv.ParseUint(numToConvert, 16, 64)
	if err != nil {
		return "", fmt.Errorf("failed to convert hex '%s' to decimal: %w", numToConvert, err)
	}

	var mnemonicParts []string
	currentNum := decimalNum

	for i := 0; i < numWords; i++ {
		index := currentNum % wordListSize
		selectedWord := dictionary[index]

		// Capitalize the first letter of the selected word
		if len(selectedWord) > 0 {
			runes := []rune(selectedWord)
			runes[0] = unicode.ToUpper(runes[0])
			selectedWord = string(runes)
		}

		mnemonicParts = append(mnemonicParts, selectedWord)
		currentNum /= wordListSize
		if currentNum == 0 && i < numWords-1 {
			// If the number runs out before all words are selected,
			// we could stop or re-seed from decimalNum (for more variation).
			// The original bash script would produce the same word repeatedly if currentNum became 0.
			// For now, we follow that implicit behavior (the last word might be repeated if entropy runs out early).
		}
	}

	return strings.Join(mnemonicParts, ""), nil
}

// Main function for demonstration is commented out as this is a library package.
/*
func main() {
	// This main function would need to be updated to use wordlist.GetWords()
	// and pass the result to GenerateMnemonic if used for testing.
	// Example:
	// import "github.com/NumeXx/gsm/pkg/wordlist"
	// ...
	// words := wordlist.GetWords()
	// if len(words) == 0 {
	// 	 fmt.Fprintf(os.Stderr, "Wordlist is empty, check pkg/wordlist/wordlist.go and embedded file.\n")
	// 	 os.Exit(1)
	// }
	// mnemonic, err := GenerateMnemonic(secret, numWords, words)
	// ... rest of the main function ...
}
*/
