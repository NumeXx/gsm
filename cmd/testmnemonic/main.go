package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/NumeXx/gsm/pkg/utils"
	"github.com/NumeXx/gsm/pkg/wordlist"
)

func main() {
	var secret string
	var numWords int
	// var wordListPath string // No longer needed
	var err error

	args := os.Args[1:] // Exclude program name

	switch len(args) {
	case 1: // Only num_words is provided, generate secret
		runGsNetcat := false

		if runGsNetcat {
			cmd := exec.Command("gs-netcat", "-g")
			output, errCmd := cmd.Output()
			if errCmd != nil {
				fmt.Fprintf(os.Stderr, "Error executing 'gs-netcat -g': %v\n", errCmd)
				fmt.Fprintf(os.Stderr, "Please ensure 'gs-netcat' is installed and in your PATH for this to work.\n")
				os.Exit(1)
			}
			secret = strings.TrimSpace(string(output))
			if secret == "" {
				fmt.Fprintf(os.Stderr, "'gs-netcat -g' did not produce any output.\n")
				os.Exit(1)
			}
		} else {
			// Generate a random secret for simulation
			randomBytes := make([]byte, 16) // 16 bytes will give a 32-char hex string
			_, err = rand.Read(randomBytes) // Assign to existing err
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating random bytes for secret: %v\n", err)
				os.Exit(1)
			}
			secret = hex.EncodeToString(randomBytes)
		}

		numWords, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: num_words ('%s') must be a number: %v\n", args[0], err)
			os.Exit(1)
		}
		// wordListPath = args[1] // No longer needed

	case 2:
		secret = args[0]
		numWords, err = strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: num_words ('%s') must be a number: %v\n", args[1], err)
			os.Exit(1)
		}
		// wordListPath = args[2]
	default:
		fmt.Fprintf(os.Stderr, "Usage: go run cmd/testmnemonic/main.go [secret] <num_words>\n")
		fmt.Fprintf(os.Stderr, "  Example 1 (provide secret): go run cmd/testmnemonic/main.go \"mySuperSecret\" 2\n")
		fmt.Fprintf(os.Stderr, "  Example 2 (generate/simulate secret): go run cmd/testmnemonic/main.go 2\n")
		os.Exit(1)
	}
	actualDictionary := wordlist.GetWords()
	if len(actualDictionary) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Wordlist from pkg/wordlist is empty. Check embedded file.\n")
		os.Exit(1)
	}
	mnemonic, err := utils.GenerateMnemonic(secret, numWords, actualDictionary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating mnemonic: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(mnemonic) // Pure output
}
