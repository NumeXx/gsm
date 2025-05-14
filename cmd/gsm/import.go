package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/NumeXx/gsm/pkg/config"
	"github.com/NumeXx/gsm/pkg/utils"
	"github.com/NumeXx/gsm/pkg/wordlist"
	"github.com/spf13/cobra"
)

// ANSI Colors (moved to top for better visibility)
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

var (
	secretKeyForImport string
	filePathForImport  string
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import GSocket connections from a key or file",
	Long: `Import GSocket connections either from a single secret key 
provided via --secret flag (format: KEY[#tag1,tag2]), or from a text file 
containing a list of secret keys (one per line, format: KEY[#tag1,tag2] [optional comments]).

Lines in the file not starting with an alphanumeric character (A-Z, a-z, 0-9) will be skipped.
For lines starting with an alphanumeric character, only the characters up to the first space or tab
will be considered as the KEY[#tag] part. Anything after the first space/tab is ignored.

If a name is not implicitly provided, a mnemonic name will be automatically 
generated based on the secret key. Tags are optional.`,
	Run: func(cmd *cobra.Command, args []string) {
		if secretKeyForImport != "" && filePathForImport != "" {
			fmt.Fprintf(os.Stderr, "%s%sError: --secret and --file flags cannot be used together.%s\n", ColorBold, ColorRed, ColorReset)
			cmd.Usage() //nolint:errcheck
			os.Exit(1)
		}

		if secretKeyForImport == "" && filePathForImport == "" {
			fmt.Fprintf(os.Stderr, "%s%sError: either --secret or --file flag must be provided.%s\n", ColorBold, ColorRed, ColorReset)
			cmd.Usage() //nolint:errcheck
			os.Exit(1)
		}

		dictionary := wordlist.GetWords()
		if len(dictionary) == 0 {
			fmt.Fprintf(os.Stderr, "%s%sError: Wordlist is empty. Cannot generate mnemonic names.%s\n", ColorBold, ColorRed, ColorReset)
			os.Exit(1)
		}
		numWordsForMnemonic := 3

		if err := config.Load(); err != nil {
			fmt.Fprintf(os.Stderr, "%s%sError loading existing configuration: %v%s\n", ColorBold, ColorRed, err, ColorReset)
			os.Exit(1)
		}
		existingConfig := config.GetCurrent()
		connectionsToAdd := []config.Connection{}

		if secretKeyForImport != "" {
			keyCandidate := strings.TrimSpace(secretKeyForImport)

			if keyCandidate == "" || !isBase62(rune(keyCandidate[0])) {
				fmt.Fprintf(os.Stderr, "%s%sError: Secret key provided via --secret must start with an alphanumeric character and not be empty.%s\n", ColorBold, ColorRed, ColorReset)
				os.Exit(1)
			}
			actualKey, parsedTags := parseKeyAndTags(keyCandidate)

			if actualKey == "" {
				fmt.Fprintf(os.Stderr, "%s%sError: Provided secret key via --secret is effectively empty after parsing.%s\n", ColorBold, ColorRed, ColorReset)
				os.Exit(1)
			}

			mnemonicName, err := utils.GenerateMnemonic(actualKey, numWordsForMnemonic, dictionary)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s%sError generating mnemonic for key '%s...': %v%s\n", ColorBold, ColorRed, actualKey[:min(len(actualKey), 8)], err, ColorReset)
				os.Exit(1)
			}

			for _, existingConn := range existingConfig.Connections {
				if existingConn.Name == mnemonicName {
					fmt.Fprintf(os.Stderr, "%s%sError: Auto-generated name '%s' (for key '%s...') already exists.%s\n", ColorBold, ColorRed, mnemonicName, actualKey[:min(len(actualKey), 8)], ColorReset)
					os.Exit(1)
				}
			}
			connectionsToAdd = append(connectionsToAdd, config.Connection{Name: mnemonicName, Key: actualKey, Tags: parsedTags, Usage: 0})
			fmt.Printf("%s[ PREPARED ]%s Name > \"%s\" | Key > \"%s...\" | Tags > %v\n", ColorYellow, ColorReset, mnemonicName, actualKey[:min(len(actualKey), 8)], parsedTags)

		} else if filePathForImport != "" {
			file, err := os.Open(filePathForImport)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s%sError opening file '%s': %v%s\n", ColorBold, ColorRed, filePathForImport, err, ColorReset)
				os.Exit(1)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			linesToProcess := []string{}
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				originalLine := scanner.Text()
				trimmedLine := strings.TrimSpace(originalLine)

				if trimmedLine == "" {
					continue
				}

				if !isBase62(rune(trimmedLine[0])) {
					fmt.Fprintf(os.Stdout, "%s[ SKIPPED ]%s Line %d does not start with alphanumeric char: \"%s\"%s\n", ColorYellow, ColorReset, lineNum, короткий(originalLine, 30), ColorReset)
					continue
				}

				keyCandidateWithPotentialTag := trimmedLine
				endOfKeyIndex := strings.IndexFunc(trimmedLine, func(r rune) bool {
					return r == ' ' || r == '\t'
				})
				if endOfKeyIndex != -1 {
					keyCandidateWithPotentialTag = trimmedLine[:endOfKeyIndex]
				}

				if keyCandidateWithPotentialTag == "" {
					fmt.Fprintf(os.Stdout, "%s[ SKIPPED ]%s Line %d became empty after isolating key part: \"%s\"%s\n", ColorYellow, ColorReset, lineNum, короткий(originalLine, 30), ColorReset)
					continue
				}
				linesToProcess = append(linesToProcess, keyCandidateWithPotentialTag)
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "%s%sError reading file '%s': %v%s\n", ColorBold, ColorRed, filePathForImport, err, ColorReset)
			}

			if len(linesToProcess) == 0 {
				fmt.Printf("%s[ INFO ]%s No processable lines found in '%s'.%s\n", ColorCyan, ColorReset, filePathForImport, ColorReset)
				os.Exit(0)
			}

			uniqueKeysInBatch := make(map[string]bool)
			existingGeneratedNames := make(map[string]bool)
			for _, existingConn := range existingConfig.Connections {
				existingGeneratedNames[existingConn.Name] = true
			}

			for i, keyCandidate := range linesToProcess {
				actualKey, parsedTags := parseKeyAndTags(keyCandidate)

				if actualKey == "" {
					fmt.Fprintf(os.Stdout, "%s[ SKIPPED ]%s Empty key from parsed line (original index %d): '%s'%s\n", ColorYellow, ColorReset, i+1, keyCandidate, ColorReset)
					continue
				}

				if uniqueKeysInBatch[actualKey] {
					fmt.Fprintf(os.Stdout, "%s[ SKIPPED ]%s Duplicate key '%s...' from file batch.%s\n", ColorYellow, ColorReset, actualKey[:min(len(actualKey), 8)], ColorReset)
					continue
				}
				uniqueKeysInBatch[actualKey] = true

				mnemonicName, err := utils.GenerateMnemonic(actualKey, numWordsForMnemonic, dictionary)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s%sError generating mnemonic for key '%s...': %v. Skipping.%s\n", ColorBold, ColorRed, actualKey[:min(len(actualKey), 8)], err, ColorReset)
					continue
				}

				if existingGeneratedNames[mnemonicName] {
					fmt.Fprintf(os.Stderr, "%s%sError: Auto-generated name '%s' (for key '%s...') already exists. Skipping.%s\n", ColorBold, ColorRed, mnemonicName, actualKey[:min(len(actualKey), 8)], ColorReset)
					continue
				}
				connectionsToAdd = append(connectionsToAdd, config.Connection{Name: mnemonicName, Key: actualKey, Tags: parsedTags, Usage: 0})
				existingGeneratedNames[mnemonicName] = true
				fmt.Printf("%s[ PREPARED ]%s Name > \"%s\" | Key > \"%s...\" | Tags > %v\n", ColorYellow, ColorReset, mnemonicName, actualKey[:min(len(actualKey), 8)], parsedTags)
			}
		}

		if len(connectionsToAdd) > 0 {
			for _, newConn := range connectionsToAdd {
				config.AddConnection(newConn)
			}
			if err := config.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "%s%sError saving imported connections: %v%s\n", ColorBold, ColorRed, err, ColorReset)
				os.Exit(1)
			}
			fmt.Printf("%s[ SUCCESS ]%s Successfully imported %s%d%s connection(s).\n", ColorGreen, ColorReset, ColorBold, len(connectionsToAdd), ColorReset)
		} else {
			fmt.Printf("%s[ INFO ]%s No new connections were imported.%s\n", ColorCyan, ColorReset, ColorReset)
		}
	},
}

func init() {
	importCmd.Flags().StringVarP(&secretKeyForImport, "secret", "s", "", "Single GSocket secret key to import (format: KEY[#tag1,tag2])")
	importCmd.Flags().StringVarP(&filePathForImport, "file", "f", "", "Path to a file with GSocket secret keys (one per line, format: KEY[#tag1,tag2] [optional comments])")
	// rootCmd.AddCommand(importCmd) // This should be done in the main/root command setup
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isBase62(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func короткий(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func parseKeyAndTags(input string) (string, []string) {
	parts := strings.SplitN(input, "#", 2)
	key := strings.TrimSpace(parts[0])
	var tags []string
	if len(parts) > 1 {
		rawTags := strings.TrimSpace(parts[1])
		if rawTags != "" {
			tagStrings := strings.Split(rawTags, ",")
			for _, t := range tagStrings {
				trimmedTag := strings.TrimSpace(t)
				if trimmedTag != "" {
					tags = append(tags, trimmedTag)
				}
			}
		}
	}
	return key, tags
}
