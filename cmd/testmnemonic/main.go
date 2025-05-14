package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec" // Uncomment this if you want to test actual 'gs-netcat -g'
	"strconv"
	"strings"

	// Sesuaikan path import ini dengan struktur module Go proyek lo
	// Jika go.mod lo mendefinisikan module 'github.com/NumeXx/gsm', maka path ini benar.
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
		// --- Bagian untuk gs-netcat -g (awalnya disimulasi) ---
		runGsNetcat := false // Ganti jadi true untuk coba 'gs-netcat -g' asli

		if runGsNetcat {
			// Pastikan os/exec di-uncomment di atas
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
		// --- Akhir bagian gs-netcat -g ---

		numWords, err = strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: num_words ('%s') must be a number: %v\n", args[0], err)
			os.Exit(1)
		}
		// wordListPath = args[1] // No longer needed

	case 2: // secret and num_words provided
		secret = args[0]
		numWords, err = strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: num_words ('%s') must be a number: %v\n", args[1], err)
			os.Exit(1)
		}
		// wordListPath = args[2] // No longer needed
	default:
		fmt.Fprintf(os.Stderr, "Usage: go run cmd/testmnemonic/main.go [secret] <num_words>\n")
		fmt.Fprintf(os.Stderr, "  Example 1 (provide secret): go run cmd/testmnemonic/main.go \"mySuperSecret\" 2\n")
		fmt.Fprintf(os.Stderr, "  Example 2 (generate/simulate secret): go run cmd/testmnemonic/main.go 2\n")
		os.Exit(1)
	}

	// BARU: Ambil kamus dari pkg/wordlist
	actualDictionary := wordlist.GetWords() // Ini manggil fungsi dari pkg/wordlist
	if len(actualDictionary) == 0 {
		fmt.Fprintf(os.Stderr, "Error: Wordlist from pkg/wordlist is empty. Check embedded file.\n")
		os.Exit(1)
	}

	// Panggil fungsi GenerateMnemonic dari package utils PAKE KAMUS YANG BARU
	mnemonic, err := utils.GenerateMnemonic(secret, numWords, actualDictionary) // wordListPath diganti actualDictionary
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating mnemonic: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(mnemonic) // Pure output
}
