package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
provided via --secret flag (format: KEY#tag1,tag2), or from a text file 
containing a list of secret keys (one per line, format: KEY#tag1,tag2).

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
			actualKey, parsedTags := parseKeyAndTags(secretKeyForImport)
			if actualKey == "" {
				fmt.Fprintf(os.Stderr, "%s%sError: Provided secret key via --secret is empty.%s\n", ColorBold, ColorRed, ColorReset)
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
			linesFromFile := []string{}
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				linesFromFile = append(linesFromFile, line) // Store the original line with potential tags
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "%s%sError reading file '%s': %v%s\n", ColorBold, ColorRed, filePathForImport, err, ColorReset)
			}

			if len(linesFromFile) == 0 {
				fmt.Printf("%s[ INFO ]%s No valid lines found in '%s'. Nothing to import.\n", ColorCyan, ColorReset, filePathForImport)
				os.Exit(0)
			}

			uniqueKeysInBatch := make(map[string]bool)      // Tracks keys within the current file batch to avoid duplicate processing
			existingGeneratedNames := make(map[string]bool) // Tracks names already in config or generated in this batch
			for _, existingConn := range existingConfig.Connections {
				existingGeneratedNames[existingConn.Name] = true
			}

			for _, lineWithPotentialTag := range linesFromFile {
				actualKey, parsedTags := parseKeyAndTags(lineWithPotentialTag)

				if actualKey == "" {
					fmt.Fprintf(os.Stdout, "%s[ SKIPPED ]%s Empty key found in file after parsing line: '%s'.\n", ColorYellow, ColorReset, lineWithPotentialTag)
					continue
				}

				if uniqueKeysInBatch[actualKey] {
					fmt.Fprintf(os.Stdout, "%s[ SKIPPED ]%s Duplicate key '%s...' from file (already processed in this batch).\n", ColorYellow, ColorReset, actualKey[:min(len(actualKey), 8)])
					continue
				}
				uniqueKeysInBatch[actualKey] = true

				mnemonicName, err := utils.GenerateMnemonic(actualKey, numWordsForMnemonic, dictionary)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s%sError generating mnemonic for key '%s...' from file: %v. Skipping.%s\n", ColorBold, ColorRed, actualKey[:min(len(actualKey), 8)], err, ColorReset)
					continue
				}

				if existingGeneratedNames[mnemonicName] {
					fmt.Fprintf(os.Stderr, "%s%sError: Auto-generated name '%s' (for key '%s...') already exists. Skipping.%s\n", ColorBold, ColorRed, mnemonicName, actualKey[:min(len(actualKey), 8)], ColorReset)
					continue
				}
				connectionsToAdd = append(connectionsToAdd, config.Connection{Name: mnemonicName, Key: actualKey, Tags: parsedTags, Usage: 0})
				existingGeneratedNames[mnemonicName] = true
				fmt.Printf("%s[ PREPARED ]%s Name > \"%s\" | Key > \"%s...\" | Tags > %v (from file '%s')\n", ColorYellow, ColorReset, mnemonicName, actualKey[:min(len(actualKey), 8)], parsedTags, filePathForImport)
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
	importCmd.Flags().StringVarP(&secretKeyForImport, "secret", "s", "", "Single GSocket secret key to import (format: KEY#tag1,tag2)")
	importCmd.Flags().StringVarP(&filePathForImport, "file", "f", "", "Path to a file with GSocket secret keys (one per line, format: KEY#tag1,tag2)")
	// rootCmd.AddCommand(importCmd) // This should be done in the main/root command setup
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
